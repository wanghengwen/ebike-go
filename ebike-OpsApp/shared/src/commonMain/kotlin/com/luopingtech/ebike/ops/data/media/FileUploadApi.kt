package com.luopingtech.ebike.ops.data.media

import com.luopingtech.ebike.ops.core.network.NetworkSession
import com.luopingtech.ebike.ops.core.network.RequestAuth
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.core.signing.RequestSigner
import com.luopingtech.ebike.ops.core.util.Ids
import com.luopingtech.ebike.ops.data.api.ApiEnvelope
import com.luopingtech.ebike.ops.platform.DeviceInfo
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
 * Legacy path: `POST /business/ebike-management/file/upload` (multipart `file` + common fields).
 * Returns the CDN/OSS URL string in envelope `data`.
 */
class FileUploadApi(
    private val client: HttpClient,
    private val requestAuth: RequestAuth,
    private val json: Json,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
    private val pinProvider: () -> String,
    private val sessionProvider: () -> NetworkSession,
) {
    suspend fun uploadBytes(
        bytes: ByteArray,
        fileName: String,
    ): OpsResult<String> {
        if (bytes.isEmpty()) {
            return OpsResult.Err(OpsError.business("MEDIA_EMPTY", "empty file"))
        }
        val source = "/business/ebike-management/file/upload"
        val session = sessionProvider()
        val pin = pinProvider()
        val fields = linkedMapOf(
            "traceId" to Ids.uuidV4(),
            "platform" to deviceInfo.platform,
            "version" to deviceInfo.appVersion,
            "source" to source,
            "stressTesting" to "false",
            "tenantId" to tenantIdProvider(),
            "Accept-Language" to com.luopingtech.ebike.ops.core.i18n.LocaleContext.acceptLanguage,
            "deviceId" to deviceIdProvider(),
        )
        if (pin.isNotBlank()) {
            fields["pin"] = pin
        }
        // Multipart bodies are not JSON-signed the same way; reuse empty-object sign + bearer.
        val signBody = "{}"
        return try {
            val response = client.post("business/ebike-management/file/upload") {
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
            OpsResult.Err(OpsError.network("upload failed: ${t.message}", t))
        }
    }

    private fun decodeUrl(status: HttpStatusCode, raw: String): OpsResult<String> {
        if (raw.isBlank()) {
            return OpsResult.Err(OpsError.network("empty upload response (HTTP ${status.value})"))
        }
        return try {
            val envelope = json.decodeFromString(ApiEnvelope.serializer(serializer<String>()), raw)
            val data = envelope.data
            when {
                !data.isNullOrBlank() && (envelope.isSuccessful || status.isSuccess()) ->
                    OpsResult.Ok(data)
                !status.isSuccess() ->
                    OpsResult.Err(OpsError.network("HTTP ${status.value}: ${envelope.msg ?: "error"}"))
                else -> OpsResult.Err(
                    OpsError.business(
                        envelope.code.ifBlank { "UPLOAD" },
                        envelope.msg ?: "upload failed",
                    ),
                )
            }
        } catch (t: Throwable) {
            OpsResult.Err(OpsError.network("upload parse failed: ${t.message}", t))
        }
    }
}
