package com.luopingtech.ebike.rider.data.media

import com.luopingtech.ebike.rider.core.i18n.LocaleContext
import com.luopingtech.ebike.rider.core.network.NetworkSession
import com.luopingtech.ebike.rider.core.network.RequestAuth
import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.core.signing.RequestSigner
import com.luopingtech.ebike.rider.core.util.Ids
import com.luopingtech.ebike.rider.data.api.ApiEnvelope
import com.luopingtech.ebike.rider.platform.DeviceInfo
import io.ktor.client.HttpClient
import io.ktor.client.request.forms.MultiPartFormDataContent
import io.ktor.client.request.forms.formData
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.client.statement.bodyAsText
import io.ktor.http.Headers
import io.ktor.http.HttpHeaders
import io.ktor.http.HttpStatusCode
import io.ktor.http.isSuccess
import kotlinx.serialization.json.Json
import kotlinx.serialization.serializer

/**
 * C 端上传：`POST /client/file/upload`（multipart `file` + 公共字段）。
 * 对齐 UniApp `src/shared/upload.ts`。
 */
class FileUploadApi(
    private val client: HttpClient,
    private val requestAuth: RequestAuth,
    private val json: Json,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
    private val sessionProvider: () -> NetworkSession,
) {
    suspend fun uploadBytes(
        bytes: ByteArray,
        fileName: String,
    ): RiderResult<String> {
        if (bytes.isEmpty()) {
            return RiderResult.Err(RiderError.business("MEDIA_EMPTY", "empty file"))
        }
        val source = "/client/file/upload"
        val session = sessionProvider()
        val fields = linkedMapOf(
            "traceId" to Ids.uuidV4(),
            "platform" to deviceInfo.platform,
            "version" to deviceInfo.appVersion,
            "source" to source,
            "stressTesting" to "false",
            "tenantId" to tenantIdProvider(),
            "Accept-Language" to LocaleContext.acceptLanguage,
            "deviceId" to deviceIdProvider(),
        )
        val signBody = "{}"
        return try {
            val response = client.post("client/file/upload") {
                requestAuth.applyPostJson(this, signBody, session)
                setBody(
                    MultiPartFormDataContent(
                        formData {
                            fields.forEach { (key, value) ->
                                append(key, value)
                            }
                            append(
                                "file",
                                bytes,
                                Headers.build {
                                    append(HttpHeaders.ContentType, "image/jpeg")
                                    append(
                                        HttpHeaders.ContentDisposition,
                                        "filename=\"${fileName.ifBlank { "photo_${Ids.uuidV4()}.jpg" }}\"",
                                    )
                                },
                            )
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
            decodeUrl(response.status, response.bodyAsText())
        } catch (t: Throwable) {
            RiderResult.Err(RiderError.network("upload failed: ${t.message}", t))
        }
    }

    private fun decodeUrl(status: HttpStatusCode, raw: String): RiderResult<String> {
        if (raw.isBlank()) {
            return RiderResult.Err(RiderError.network("empty upload response (HTTP ${status.value})"))
        }
        return try {
            val envelope = json.decodeFromString(ApiEnvelope.serializer(serializer<String>()), raw)
            val data = envelope.data
            when {
                !data.isNullOrBlank() && (envelope.isSuccessful || status.isSuccess()) ->
                    RiderResult.Ok(data)
                !status.isSuccess() ->
                    RiderResult.Err(RiderError.network("HTTP ${status.value}: ${envelope.msg ?: "error"}"))
                else -> RiderResult.Err(
                    RiderError.business(
                        envelope.code.ifBlank { "UPLOAD" },
                        envelope.msg ?: "upload failed",
                    ),
                )
            }
        } catch (t: Throwable) {
            RiderResult.Err(RiderError.network("upload parse failed: ${t.message}", t))
        }
    }
}
