package com.luopingtech.ebike.ops.core.network

import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.core.signing.RequestSigner
import com.luopingtech.ebike.ops.data.api.ApiEnvelope
import io.ktor.client.HttpClient
import io.ktor.client.request.forms.MultiPartFormDataContent
import io.ktor.client.request.forms.formData
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.client.statement.bodyAsText
import io.ktor.http.HttpStatusCode
import io.ktor.http.isSuccess
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import kotlinx.serialization.KSerializer
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.contentOrNull
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive

/**
 * Signed JSON POST helper with one-shot token refresh on 00005 / 00006.
 */
class SignedApiClient(
    private val client: HttpClient,
    private val requestAuth: RequestAuth,
    private val json: Json,
    private val baseUrl: String,
    private val sessionProvider: () -> NetworkSession,
    private val refreshAccessToken: suspend () -> OpsResult<String>,
    private val onSessionInvalid: (code: String, message: String) -> Unit = { _, _ -> },
) {
    private val refreshMutex = Mutex()

    suspend fun <T> post(
        path: String,
        bodyJson: String,
        deserializer: KSerializer<T>,
    ): OpsResult<T> = withAuthRetry(path, bodyJson) { status, raw ->
        decode(status, raw, deserializer)
    }

    /**
     * For tool APIs that return success with null/empty `data` (unlock/lock/ring).
     */
    suspend fun postUnit(path: String, bodyJson: String): OpsResult<Unit> =
        withAuthRetry(path, bodyJson) { status, raw ->
            decodeUnit(status, raw)
        }

    /**
     * Multipart form POST (legacy carTag/record/add etc.).
     * Body is not JSON-signed; headers reuse empty-object sign like [FileUploadApi].
     */
    suspend fun postMultipartUnit(
        path: String,
        fields: Map<String, String>,
    ): OpsResult<Unit> = withAuthRetryMultipart(path, fields) { status, raw ->
        decodeUnit(status, raw)
    }

    private fun resolveUrl(path: String): String {
        val absolutePath = "/" + path.trimStart('/')
        val base = baseUrl.trim().trimEnd('/')
        return if (base.isBlank()) absolutePath else base + absolutePath
    }

    private suspend fun <T> withAuthRetry(
        path: String,
        bodyJson: String,
        decode: (HttpStatusCode, String) -> OpsResult<T>,
    ): OpsResult<T> {
        val first = execute(path, bodyJson, sessionProvider())
        val code = peekCode(first.raw)
        if (ApiCodes.shouldForceRelogin(code)) {
            onSessionInvalid(code, peekMessage(first.raw))
            return OpsResult.Err(OpsError.unauthorized(peekMessage(first.raw).ifBlank { code }))
        }
        if (!ApiCodes.shouldRefresh(code)) {
            return decode(first.status, first.raw)
        }

        val refreshed = refreshMutex.withLock { refreshAccessToken() }
        return when (refreshed) {
            is OpsResult.Err -> {
                onSessionInvalid(refreshed.error.code, refreshed.error.message)
                refreshed
            }
            is OpsResult.Ok -> {
                val retrySession = NetworkSession(accessToken = refreshed.value)
                val second = execute(path, bodyJson, retrySession)
                val retryCode = peekCode(second.raw)
                if (ApiCodes.shouldForceRelogin(retryCode) || ApiCodes.shouldRefresh(retryCode)) {
                    onSessionInvalid(retryCode, peekMessage(second.raw))
                    OpsResult.Err(OpsError.unauthorized(peekMessage(second.raw).ifBlank { retryCode }))
                } else {
                    decode(second.status, second.raw)
                }
            }
        }
    }

    private suspend fun execute(
        path: String,
        bodyJson: String,
        session: NetworkSession,
    ): RawResponse {
        // Full absolute URL — matches Retrofit baseUrl + @POST("/…") and avoids any
        // relative merge against defaultRequest / sticky OkHttp paths.
        val fullUrl = resolveUrl(path)
        val response = client.post(fullUrl) {
            requestAuth.applyPostJson(this, bodyJson, session, AuthHeaderMode.Bearer)
            // TextContent bypasses ContentNegotiation so the body is not JSON-string-wrapped.
            setBody(io.ktor.http.content.TextContent(bodyJson, io.ktor.http.ContentType.Application.Json))
        }
        return RawResponse(response.status, response.bodyAsText())
    }

    private suspend fun <T> withAuthRetryMultipart(
        path: String,
        fields: Map<String, String>,
        decode: (HttpStatusCode, String) -> OpsResult<T>,
    ): OpsResult<T> {
        val first = executeMultipart(path, fields, sessionProvider())
        val code = peekCode(first.raw)
        if (ApiCodes.shouldForceRelogin(code)) {
            onSessionInvalid(code, peekMessage(first.raw))
            return OpsResult.Err(OpsError.unauthorized(peekMessage(first.raw).ifBlank { code }))
        }
        if (!ApiCodes.shouldRefresh(code)) {
            return decode(first.status, first.raw)
        }

        val refreshed = refreshMutex.withLock { refreshAccessToken() }
        return when (refreshed) {
            is OpsResult.Err -> {
                onSessionInvalid(refreshed.error.code, refreshed.error.message)
                refreshed
            }
            is OpsResult.Ok -> {
                val retrySession = NetworkSession(accessToken = refreshed.value)
                val second = executeMultipart(path, fields, retrySession)
                val retryCode = peekCode(second.raw)
                if (ApiCodes.shouldForceRelogin(retryCode) || ApiCodes.shouldRefresh(retryCode)) {
                    onSessionInvalid(retryCode, peekMessage(second.raw))
                    OpsResult.Err(OpsError.unauthorized(peekMessage(second.raw).ifBlank { retryCode }))
                } else {
                    decode(second.status, second.raw)
                }
            }
        }
    }

    private suspend fun executeMultipart(
        path: String,
        fields: Map<String, String>,
        session: NetworkSession,
    ): RawResponse {
        val fullUrl = resolveUrl(path)
        val signBody = "{}"
        val response = client.post(fullUrl) {
            requestAuth.applyPostJson(this, signBody, session, AuthHeaderMode.Bearer)
            setBody(
                MultiPartFormDataContent(
                    formData {
                        fields.forEach { (key, value) ->
                            append(key, value)
                        }
                    },
                ),
            )
            if (session.accessToken.isNotBlank()) {
                headers.set(
                    RequestSigner.HEADER_AUTHORIZATION,
                    "Bearer ${session.accessToken.trim()}",
                )
            }
        }
        return RawResponse(response.status, response.bodyAsText())
    }

    private fun <T> decode(
        status: HttpStatusCode,
        raw: String,
        deserializer: KSerializer<T>,
    ): OpsResult<T> {
        if (raw.isBlank()) {
            return OpsResult.Err(OpsError.network("empty response (HTTP ${status.value})"))
        }
        return try {
            val envelope = json.decodeFromString(ApiEnvelope.serializer(deserializer), raw)
            val data = envelope.data
            when {
                data != null && (envelope.isSuccessful || status.isSuccess()) -> OpsResult.Ok(data)
                data != null -> OpsResult.Err(
                    OpsError.business(
                        envelope.code.ifBlank { "BUSINESS" },
                        envelope.msg ?: "request failed",
                    ),
                )
                // 业务成功但 data=null（常见空列表）：按空数组/空对象落成 Ok，避免把「成功！」当错误展示
                envelope.isSuccessful || status.isSuccess() -> {
                    emptyDataFallback(deserializer)?.let { OpsResult.Ok(it) }
                        ?: OpsResult.Err(
                            OpsError.business(
                                envelope.code.ifBlank { "EMPTY" },
                                envelope.msg?.takeUnless { it.isBlank() || it.contains("成功") }
                                    ?: "empty data",
                            ),
                        )
                }
                !status.isSuccess() ->
                    OpsResult.Err(OpsError.network("HTTP ${status.value}: ${envelope.msg ?: "error"}"))
                else -> OpsResult.Err(
                    OpsError.business(
                        envelope.code.ifBlank { "EMPTY" },
                        envelope.msg ?: "empty data",
                    ),
                )
            }
        } catch (t: Throwable) {
            OpsResult.Err(OpsError.network("parse failed: ${t.message}", t))
        }
    }

    /** data=null 时尝试 [] / {}，覆盖 List / 部分对象 DTO。 */
    private fun <T> emptyDataFallback(deserializer: KSerializer<T>): T? {
        listOf("[]", "{}").forEach { raw ->
            try {
                return json.decodeFromString(deserializer, raw)
            } catch (_: Throwable) {
                // try next
            }
        }
        return null
    }

    private fun decodeUnit(status: HttpStatusCode, raw: String): OpsResult<Unit> {
        if (raw.isBlank()) {
            return if (status.isSuccess()) {
                OpsResult.Ok(Unit)
            } else {
                OpsResult.Err(OpsError.network("empty response (HTTP ${status.value})"))
            }
        }
        return try {
            val envelope = json.decodeFromString(ApiEnvelope.serializer(JsonElement.serializer()), raw)
            when {
                envelope.isSuccessful -> OpsResult.Ok(Unit)
                status.isSuccess() && envelope.code.isBlank() -> OpsResult.Ok(Unit)
                !status.isSuccess() ->
                    OpsResult.Err(OpsError.network("HTTP ${status.value}: ${envelope.msg ?: "error"}"))
                else -> OpsResult.Err(
                    OpsError.business(
                        envelope.code.ifBlank { "BUSINESS" },
                        envelope.msg ?: "request failed",
                    ),
                )
            }
        } catch (t: Throwable) {
            OpsResult.Err(OpsError.network("parse failed: ${t.message}", t))
        }
    }

    private fun peekCode(raw: String): String {
        if (raw.isBlank()) return ""
        return runCatching {
            val element: JsonElement = json.parseToJsonElement(raw)
            element.jsonObject["code"]?.jsonPrimitive?.contentOrNull.orEmpty()
        }.getOrDefault("")
    }

    private fun peekMessage(raw: String): String {
        if (raw.isBlank()) return ""
        return runCatching {
            val obj: JsonObject = json.parseToJsonElement(raw).jsonObject
            obj["msg"]?.jsonPrimitive?.contentOrNull
                ?: (obj["message"] as? JsonPrimitive)?.contentOrNull
                ?: ""
        }.getOrDefault("")
    }

    private data class RawResponse(val status: HttpStatusCode, val raw: String)
}
