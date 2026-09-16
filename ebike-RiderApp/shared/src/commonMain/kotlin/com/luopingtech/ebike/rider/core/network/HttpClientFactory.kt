package com.luopingtech.ebike.rider.core.network

import com.luopingtech.ebike.rider.core.config.TenantConfig
import kotlin.concurrent.Volatile
import com.luopingtech.ebike.rider.core.crypto.Base64Text
import com.luopingtech.ebike.rider.core.logging.RiderLogger
import com.luopingtech.ebike.rider.core.signing.RequestSigner
import com.luopingtech.ebike.rider.core.time.nowEpochMillis
import com.luopingtech.ebike.rider.platform.SecureStore
import io.ktor.client.HttpClient
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.plugins.defaultRequest
import io.ktor.client.plugins.logging.LogLevel
import io.ktor.client.plugins.logging.Logger
import io.ktor.client.plugins.logging.Logging
import io.ktor.client.request.HttpRequestBuilder
import io.ktor.client.request.header
import io.ktor.http.ContentType
import io.ktor.http.HttpHeaders
import io.ktor.http.contentType
import io.ktor.serialization.kotlinx.json.json
import kotlinx.serialization.json.Json

expect fun createPlatformHttpClient(): HttpClient

data class NetworkSession(
    val accessToken: String = "",
    /**
     * OAuth `tokenType` from the server is typically lowercase `bearer`.
     * Authorization scheme must stay legacy-compatible `Bearer` — see [RequestAuth].
     */
    val tokenType: String = "Bearer",
)

enum class AuthHeaderMode {
    /** `Bearer {accessToken}` from session / secure store. */
    Bearer,

    /** `Basic base64(tenantId:businessSecret)` for `/oauth/token`. */
    BasicClient,

    /** No Authorization header. */
    None,
}

/**
 * Builds signed request headers matching the legacy merchant app contract.
 */
class RequestAuth(
    private val config: TenantConfig,
    private val secureStore: SecureStore,
    private val clock: () -> Long = { nowEpochMillis() },
) {
    /** Optional override for `/oauth/token` Basic after分部选择 (`tenantId:secret`). */
    @Volatile
    var oauthBasicTenantId: String? = null

    @Volatile
    var oauthBasicSecret: String? = null

    fun clearOauthBasicOverride() {
        oauthBasicTenantId = null
        oauthBasicSecret = null
    }

    fun basicClientAuthorization(): String {
        val id = oauthBasicTenantId?.takeIf { it.isNotBlank() }
            ?: secureStore.getString(SecureStore.KEY_TENANT_ID)?.takeIf { it.isNotBlank() }
            ?: config.tenantId
        val secret = oauthBasicSecret?.takeIf { it.isNotBlank() }
            ?: config.auth.businessSecret
        val raw = "$id:$secret"
        return "Basic ${Base64Text.encodeUtf8(raw)}"
    }

    fun applyGet(
        builder: HttpRequestBuilder,
        queryParams: Map<String, String> = emptyMap(),
        session: NetworkSession = NetworkSession(),
        authMode: AuthHeaderMode = AuthHeaderMode.Bearer,
    ) {
        val timestamp = clock().toString()
        val sign = RequestSigner.signGet(queryParams, timestamp, config.auth.signSecret)
        applyCommon(builder, timestamp, sign, session, authMode)
    }

    fun applyPostJson(
        builder: HttpRequestBuilder,
        bodyJson: String,
        session: NetworkSession = NetworkSession(),
        authMode: AuthHeaderMode = AuthHeaderMode.Bearer,
    ) {
        val timestamp = clock().toString()
        val sign = RequestSigner.signPostJson(bodyJson, timestamp, config.auth.signSecret)
        applyCommon(builder, timestamp, sign, session, authMode)
        builder.contentType(ContentType.Application.Json)
    }

    private fun applyCommon(
        builder: HttpRequestBuilder,
        timestamp: String,
        sign: String,
        session: NetworkSession,
        authMode: AuthHeaderMode,
    ) {
        builder.header(RequestSigner.HEADER_TIMESTAMP, timestamp)
        builder.header(RequestSigner.HEADER_SIGN, sign)
        builder.header(HttpHeaders.AcceptLanguage, com.luopingtech.ebike.rider.core.i18n.LocaleContext.acceptLanguage)

        when (authMode) {
            AuthHeaderMode.None -> Unit
            AuthHeaderMode.BasicClient -> {
                // Replace (not append) — OkHttp/gateway take the first Authorization value.
                builder.headers.set(
                    RequestSigner.HEADER_AUTHORIZATION,
                    basicClientAuthorization(),
                )
            }
            AuthHeaderMode.Bearer -> {
                val stored = secureStore.getString(SecureStore.KEY_ACCESS_TOKEN).orEmpty().trim()
                val token = session.accessToken.trim().ifBlank { stored }
                if (token.isNotBlank()) {
                    // Legacy ManagerApp always uses capital "Bearer ", never server tokenType.
                    builder.headers.set(
                        RequestSigner.HEADER_AUTHORIZATION,
                        "Bearer $token",
                    )
                }
            }
        }
    }
}

class HttpClientFactory(
    private val config: TenantConfig,
    private val riderLogger: RiderLogger,
) {
    val json: Json = Json {
        ignoreUnknownKeys = true
        isLenient = true
        encodeDefaults = true
        explicitNulls = false
        coerceInputValues = true
    }

    fun create(): HttpClient = createPlatformHttpClient().config {
        expectSuccess = false
        install(ContentNegotiation) {
            json(json)
        }
        install(Logging) {
            level = LogLevel.HEADERS
            this.logger = object : Logger {
                override fun log(message: String) {
                    // Headers only — avoid dumping bodies that may contain secrets.
                    riderLogger.d("Http", message)
                }
            }
        }
        val base = config.api.baseUrl.trim()
        if (base.isNotEmpty()) {
            defaultRequest {
                url(base.trimEnd('/') + "/")
            }
        }
    }
}
