package com.luopingtech.ebike.rider.core.network

import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.core.signing.RequestSigner
import com.luopingtech.ebike.rider.data.api.ApiEnvelope
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
 * 一次业务响应的原样结果：业务码、文案、`data` 都留着。
 *
 * [SignedApiClient.post] 在业务失败时只剩 code + msg，而骑行域的还车矩阵恰恰依赖
 * 「失败但带 data」——`returnByNet` 返回 code 0 时会把围栏判定塞在 `data` 里，
 * 界面要靠它重开弹层。
 */
data class ApiOutcome(
    val success: Boolean,
    val code: String,
    val message: String,
    val data: JsonElement?,
)

/**
 * Signed JSON POST helper.
 *
 * Aligns with UniApp `request.ts` WHITE_CODES:
 * `00005` / `00013` single-flight refresh + retry; `00015` force re-login;
 * `00006` clears session only when a token was actually sent.
 */
class SignedApiClient(
    private val client: HttpClient,
    private val requestAuth: RequestAuth,
    private val json: Json,
    private val sessionProvider: () -> NetworkSession,
    private val refreshAccessToken: suspend () -> RiderResult<String>,
    private val onSessionInvalid: (code: String, message: String) -> Unit = { _, _ -> },
) {
    private val refreshMutex = Mutex()

    suspend fun <T> post(
        path: String,
        bodyJson: String,
        deserializer: KSerializer<T>,
    ): RiderResult<T> = withAuthRetry(path, bodyJson) { status, raw ->
        decode(status, raw, deserializer)
    }

    /**
     * H5 桥代理：走与业务接口相同的签名 / 刷新，但把网关原文原样交回 WebView。
     * 不在这里二次解码 `data`，避免长尾页各自的 envelope 被 typed decode 丢掉。
     */
    suspend fun postRaw(path: String, bodyJson: String): RiderResult<String> =
        withAuthRetry(path, bodyJson) { _, raw -> RiderResult.Ok(raw) }

    /**
     * 只把「登录态失效」当 [RiderResult.Err]，业务失败原样返回。
     * 调用方自己按业务码分支（骑行域的 17012 降级 / 15002 冻结单 / code 0 重开弹层）。
     */
    suspend fun postOutcome(path: String, bodyJson: String): RiderResult<ApiOutcome> =
        withAuthRetry(path, bodyJson) { status, raw -> decodeOutcome(status, raw) }

    /**
     * For tool APIs that return success with null/empty `data` (unlock/lock/ring).
     */
    suspend fun postUnit(path: String, bodyJson: String): RiderResult<Unit> =
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
    ): RiderResult<Unit> = withAuthRetryMultipart(path, fields) { status, raw ->
        decodeUnit(status, raw)
    }

    private suspend fun <T> withAuthRetry(
        path: String,
        bodyJson: String,
        decode: (HttpStatusCode, String) -> RiderResult<T>,
    ): RiderResult<T> {
        val session = sessionProvider()
        val first = execute(path, bodyJson, session)
        return handleAuthRetry(session, first, decode) { retrySession ->
            execute(path, bodyJson, retrySession)
        }
    }

    private suspend fun execute(
        path: String,
        bodyJson: String,
        session: NetworkSession,
    ): RawResponse {
        val clean = path.trimStart('/')
        val response = client.post(clean) {
            requestAuth.applyPostJson(this, bodyJson, session, AuthHeaderMode.Bearer)
            setBody(bodyJson)
        }
        return RawResponse(response.status, response.bodyAsText())
    }

    private suspend fun <T> withAuthRetryMultipart(
        path: String,
        fields: Map<String, String>,
        decode: (HttpStatusCode, String) -> RiderResult<T>,
    ): RiderResult<T> {
        val session = sessionProvider()
        val first = executeMultipart(path, fields, session)
        return handleAuthRetry(session, first, decode) { retrySession ->
            executeMultipart(path, fields, retrySession)
        }
    }

    private suspend fun <T> handleAuthRetry(
        session: NetworkSession,
        first: RawResponse,
        decode: (HttpStatusCode, String) -> RiderResult<T>,
        retry: suspend (NetworkSession) -> RawResponse,
    ): RiderResult<T> {
        val code = peekCode(first.raw)
        val message = peekMessage(first.raw)
        when {
            ApiCodes.shouldForceRelogin(code) -> {
                onSessionInvalid(code, message)
                return RiderResult.Err(RiderError.unauthorized(message.ifBlank { code }))
            }
            ApiCodes.isGuestUnauthorized(code) -> {
                if (session.accessToken.isNotBlank()) {
                    onSessionInvalid(code, message)
                    return RiderResult.Err(RiderError.unauthorized(message.ifBlank { code }))
                }
                return decode(first.status, first.raw)
            }
            !ApiCodes.shouldRefresh(code) -> return decode(first.status, first.raw)
        }

        val refreshed = singleFlightRefresh(session.accessToken)
        return when (refreshed) {
            is RiderResult.Err -> {
                onSessionInvalid(refreshed.error.code, refreshed.error.message)
                refreshed
            }
            is RiderResult.Ok -> {
                val second = retry(NetworkSession(accessToken = refreshed.value))
                val retryCode = peekCode(second.raw)
                val retryMessage = peekMessage(second.raw)
                val retryClears = ApiCodes.shouldForceRelogin(retryCode) ||
                    ApiCodes.shouldRefresh(retryCode) ||
                    (ApiCodes.isGuestUnauthorized(retryCode) && refreshed.value.isNotBlank())
                if (retryClears) {
                    onSessionInvalid(retryCode, retryMessage)
                    RiderResult.Err(RiderError.unauthorized(retryMessage.ifBlank { retryCode }))
                } else {
                    decode(second.status, second.raw)
                }
            }
        }
    }

    /**
     * Concurrent 00005/00013 share one refresh. Waiters that arrive after a
     * successful refresh reuse the new token instead of calling oauth again.
     */
    private suspend fun singleFlightRefresh(staleToken: String): RiderResult<String> =
        refreshMutex.withLock {
            val current = sessionProvider().accessToken.trim()
            val stale = staleToken.trim()
            if (current.isNotBlank() && current != stale) {
                RiderResult.Ok(current)
            } else {
                refreshAccessToken()
            }
        }

    private suspend fun executeMultipart(
        path: String,
        fields: Map<String, String>,
        session: NetworkSession,
    ): RawResponse {
        val clean = path.trimStart('/')
        val signBody = "{}"
        val response = client.post(clean) {
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
    ): RiderResult<T> {
        if (raw.isBlank()) {
            return RiderResult.Err(RiderError.network("empty response (HTTP ${status.value})"))
        }
        return try {
            val envelope = json.decodeFromString(ApiEnvelope.serializer(deserializer), raw)
            val data = envelope.data
            when {
                data != null && (envelope.isSuccessful || status.isSuccess()) -> RiderResult.Ok(data)
                data != null -> RiderResult.Err(
                    RiderError.business(
                        envelope.code.ifBlank { "BUSINESS" },
                        envelope.msg ?: "request failed",
                    ),
                )
                !status.isSuccess() ->
                    RiderResult.Err(RiderError.network("HTTP ${status.value}: ${envelope.msg ?: "error"}"))
                else -> RiderResult.Err(
                    RiderError.business(
                        envelope.code.ifBlank { "EMPTY" },
                        envelope.msg ?: "empty data",
                    ),
                )
            }
        } catch (t: Throwable) {
            RiderResult.Err(RiderError.network("parse failed: ${t.message}", t))
        }
    }

    private fun decodeOutcome(status: HttpStatusCode, raw: String): RiderResult<ApiOutcome> {
        if (raw.isBlank()) {
            return if (status.isSuccess()) {
                RiderResult.Ok(ApiOutcome(success = true, code = "", message = "", data = null))
            } else {
                RiderResult.Err(RiderError.network("empty response (HTTP ${status.value})"))
            }
        }
        return try {
            val envelope = json.decodeFromString(ApiEnvelope.serializer(JsonElement.serializer()), raw)
            RiderResult.Ok(
                ApiOutcome(
                    success = envelope.isSuccessful,
                    code = envelope.code,
                    message = envelope.msg.orEmpty(),
                    data = envelope.data,
                ),
            )
        } catch (t: Throwable) {
            RiderResult.Err(RiderError.network("parse failed: ${t.message}", t))
        }
    }

    private fun decodeUnit(status: HttpStatusCode, raw: String): RiderResult<Unit> {
        if (raw.isBlank()) {
            return if (status.isSuccess()) {
                RiderResult.Ok(Unit)
            } else {
                RiderResult.Err(RiderError.network("empty response (HTTP ${status.value})"))
            }
        }
        return try {
            val envelope = json.decodeFromString(ApiEnvelope.serializer(JsonElement.serializer()), raw)
            when {
                envelope.isSuccessful -> RiderResult.Ok(Unit)
                status.isSuccess() && envelope.code.isBlank() -> RiderResult.Ok(Unit)
                !status.isSuccess() ->
                    RiderResult.Err(RiderError.network("HTTP ${status.value}: ${envelope.msg ?: "error"}"))
                else -> RiderResult.Err(
                    RiderError.business(
                        envelope.code.ifBlank { "BUSINESS" },
                        envelope.msg ?: "request failed",
                    ),
                )
            }
        } catch (t: Throwable) {
            RiderResult.Err(RiderError.network("parse failed: ${t.message}", t))
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
