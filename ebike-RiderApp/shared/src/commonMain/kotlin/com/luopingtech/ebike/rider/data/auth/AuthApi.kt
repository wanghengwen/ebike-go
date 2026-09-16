package com.luopingtech.ebike.rider.data.auth

import com.luopingtech.ebike.rider.core.network.AuthHeaderMode
import com.luopingtech.ebike.rider.core.network.CommonRequestBody
import com.luopingtech.ebike.rider.core.network.NetworkSession
import com.luopingtech.ebike.rider.core.network.RequestAuth
import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.data.api.ApiEnvelope
import com.luopingtech.ebike.rider.platform.DeviceInfo
import io.ktor.client.HttpClient
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.client.statement.bodyAsText
import io.ktor.http.HttpStatusCode
import io.ktor.http.isSuccess
import kotlinx.serialization.KSerializer
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.put

/**
 * Unauthenticated auth endpoints. `/oauth/token` uses Basic
 * `base64(tenantId:businessSecret)` and `auth: false` (UniApp).
 */
interface AuthRemote {
    suspend fun sendSmsCode(phone: String, scene: Int = SmsScene.LOGIN): RiderResult<Unit>
    suspend fun loginWithSms(phone: String, messageCode: String): RiderResult<OauthTokenDto>
    suspend fun refreshToken(refreshToken: String): RiderResult<OauthTokenDto>
}

class AuthApi(
    private val client: HttpClient,
    private val requestAuth: RequestAuth,
    private val json: Json,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
    private val areaCodeProvider: () -> String = { PhoneNormalizer.DEFAULT_AREA_CODE },
) : AuthRemote {
    override suspend fun sendSmsCode(phone: String, scene: Int): RiderResult<Unit> {
        val bodyJson = commonBody {
            put("phone", PhoneNormalizer.toLoginPhone(phone, areaCodeProvider()))
            put("scene", scene)
        }
        // 登录前没有 token；对齐 OpsApp / UniApp（未登录发码不带 Bearer）。
        return postUnit("client/code/send", bodyJson, AuthHeaderMode.None)
    }

    override suspend fun loginWithSms(phone: String, messageCode: String): RiderResult<OauthTokenDto> {
        val bodyJson = commonBody {
            put("grant_type", "phone_code")
            put("phone", PhoneNormalizer.toLoginPhone(phone, areaCodeProvider()))
            put("messageCode", messageCode)
        }
        return postOauth(bodyJson)
    }

    override suspend fun refreshToken(refreshToken: String): RiderResult<OauthTokenDto> {
        val bodyJson = commonBody {
            put("grant_type", "refresh_token")
            put("refresh_token", refreshToken)
        }
        return postOauth(bodyJson)
    }

    private suspend fun postOauth(bodyJson: String): RiderResult<OauthTokenDto> = try {
        val response = client.post("oauth/token") {
            requestAuth.applyPostJson(this, bodyJson, NetworkSession(), AuthHeaderMode.BasicClient)
            setBody(bodyJson)
        }
        parseEnvelope(response.status, response.bodyAsText(), OauthTokenDto.serializer())
    } catch (t: Throwable) {
        RiderResult.Err(RiderError.network(t.message ?: "oauth request failed", t))
    }

    private suspend fun postUnit(
        path: String,
        bodyJson: String,
        authMode: AuthHeaderMode,
    ): RiderResult<Unit> = try {
        val response = client.post(path) {
            requestAuth.applyPostJson(this, bodyJson, NetworkSession(), authMode)
            setBody(bodyJson)
        }
        decodeUnit(response.status, response.bodyAsText())
    } catch (t: Throwable) {
        RiderResult.Err(RiderError.network(t.message ?: "request failed", t))
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
            val envelope = json.decodeFromString(
                ApiEnvelope.serializer(JsonElement.serializer()),
                raw,
            )
            if (envelope.isSuccessful || status.isSuccess()) {
                RiderResult.Ok(Unit)
            } else {
                RiderResult.Err(
                    RiderError.business(
                        envelope.code.ifBlank { "BUSINESS" },
                        envelope.msg ?: "request failed",
                    ),
                )
            }
        } catch (t: Throwable) {
            if (status.isSuccess()) {
                RiderResult.Ok(Unit)
            } else {
                RiderResult.Err(RiderError.network("parse failed: ${t.message}", t))
            }
        }
    }

    private fun <T> parseEnvelope(
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

    private fun commonBody(
        block: kotlinx.serialization.json.JsonObjectBuilder.() -> Unit,
    ): String = CommonRequestBody.toJsonString(
        tenantId = tenantIdProvider(),
        deviceInfo = deviceInfo,
        deviceId = deviceIdProvider(),
        block = block,
    )
}
