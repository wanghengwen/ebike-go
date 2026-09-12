package com.luopingtech.ebike.ops.data.auth

import com.luopingtech.ebike.ops.core.network.AuthHeaderMode
import com.luopingtech.ebike.ops.core.network.NetworkSession
import com.luopingtech.ebike.ops.core.network.RequestAuth
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.core.signing.RequestSigner
import com.luopingtech.ebike.ops.core.time.nowEpochMillis
import com.luopingtech.ebike.ops.core.util.Ids
import com.luopingtech.ebike.ops.data.api.ApiEnvelope
import com.luopingtech.ebike.ops.domain.model.BusinessTenantList
import com.luopingtech.ebike.ops.domain.model.TenantRuntimeConfig
import com.luopingtech.ebike.ops.platform.DeviceInfo
import io.ktor.client.HttpClient
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.client.statement.bodyAsText
import io.ktor.http.HttpStatusCode
import io.ktor.http.isSuccess
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.put

class AuthApi(
    private val client: HttpClient,
    private val requestAuth: RequestAuth,
    private val json: Json,
    private val tenantIdProvider: () -> String,
    private val signSecretProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
    private val areaCodeProvider: () -> String = { PhoneNormalizer.DEFAULT_AREA_CODE },
) {
    private fun loginPhone(phone: String): String =
        PhoneNormalizer.toLoginPhone(phone, areaCodeProvider())

    suspend fun sendSmsCode(phone: String, scene: Int = SmsScene.LOGIN): OpsResult<Unit> {
        val bodyJson = buildJsonObject {
            putCommonFields("/business/ebike-management/code/send")
            put("phone", loginPhone(phone))
            put("scene", scene)
        }.toString()
        return try {
            val response = client.post("business/ebike-management/code/send") {
                requestAuth.applyPostJson(this, bodyJson, authMode = AuthHeaderMode.None)
                setBody(bodyJson)
            }
            val raw = response.bodyAsText()
            if (raw.isBlank()) {
                return if (response.status.isSuccess()) {
                    OpsResult.Ok(Unit)
                } else {
                    OpsResult.Err(OpsError.network("empty sms response"))
                }
            }
            val envelope = json.decodeFromString(
                ApiEnvelope.serializer(JsonElement.serializer()),
                raw,
            )
            if (envelope.isSuccessful || response.status.isSuccess()) {
                OpsResult.Ok(Unit)
            } else {
                OpsResult.Err(
                    OpsError.business(
                        envelope.code.ifBlank { "SMS" },
                        envelope.msg ?: "send sms failed",
                    ),
                )
            }
        } catch (t: Throwable) {
            OpsResult.Err(OpsError.network(t.message ?: "send sms failed", t))
        }
    }

    suspend fun forgetPassword(
        phone: String,
        validCode: String,
        newPassword: String,
    ): OpsResult<Unit> {
        val bodyJson = buildJsonObject {
            putCommonFields("/business/ebike-management/user/forget-password")
            // Legacy forget-password uses 11-digit phone (no +86- prefix).
            put("phone", PhoneNormalizer.toLocalMobile(phone))
            put("validCode", validCode)
            put("newPwd", newPassword)
        }.toString()
        return postUnitUnauthenticated("business/ebike-management/user/forget-password", bodyJson)
    }

    suspend fun setPassword(session: NetworkSession, newPassword: String): OpsResult<Unit> {
        val bodyJson = buildJsonObject {
            putCommonFields("/business/ebike-management/user/set-password")
            put("newPwd", newPassword)
        }.toString()
        return postUnitAuthenticated(
            "business/ebike-management/user/set-password",
            bodyJson,
            session,
        )
    }

    suspend fun updatePassword(
        session: NetworkSession,
        oldPassword: String,
        newPassword: String,
    ): OpsResult<Unit> {
        val bodyJson = buildJsonObject {
            putCommonFields("/business/ebike-management/user/update-password")
            put("oldPwd", oldPassword)
            put("newPwd", newPassword)
        }.toString()
        return postUnitAuthenticated(
            "business/ebike-management/user/update-password",
            bodyJson,
            session,
        )
    }

    suspend fun selectAppConfig(
        session: NetworkSession,
    ): OpsResult<TenantRuntimeConfig> {
        val timestamp = nowEpochMillis().toString()
        val tenantId = tenantIdProvider()
        val signParams = mapOf("tenantId" to tenantId)
        val sign = RequestSigner.signFormParams(
            params = signParams,
            timestamp = timestamp,
            secret = signSecretProvider(),
        )
        val bodyJson = buildJsonObject {
            putCommonFields("/business/tenant/selectAppConfig")
            put("tenantId", tenantId)
            put("time", timestamp.toLongOrNull() ?: 0L)
            put("sign", sign)
        }.toString()
        return try {
            val response = client.post("business/tenant/selectAppConfig") {
                requestAuth.applyPostJson(this, bodyJson, session, AuthHeaderMode.Bearer)
                setBody(bodyJson)
            }
            when (
                val parsed = parseEnvelope(
                    response.status,
                    response.bodyAsText(),
                    AppConfigDto.serializer(),
                )
            ) {
                is OpsResult.Ok -> OpsResult.Ok(parsed.value.toDomain())
                is OpsResult.Err -> parsed
            }
        } catch (t: Throwable) {
            OpsResult.Err(OpsError.network(t.message ?: "selectAppConfig failed", t))
        }
    }

    suspend fun loginWithPassword(phone: String, password: String): OpsResult<OauthTokenDto> {
        val bodyJson = buildOauthBody {
            put("grant_type", "phone_password")
            put("phone", loginPhone(phone))
            put("password", password)
        }
        return postOauth(bodyJson)
    }

    suspend fun loginWithSms(phone: String, messageCode: String): OpsResult<OauthTokenDto> {
        val bodyJson = buildOauthBody {
            put("grant_type", "phone_code")
            put("phone", loginPhone(phone))
            put("messageCode", messageCode)
        }
        return postOauth(bodyJson)
    }

    suspend fun loginWithPhoneSecret(phone: String, secret: String): OpsResult<OauthTokenDto> {
        val bodyJson = buildOauthBody {
            put("grant_type", "phone_secret")
            put("phone", loginPhone(phone))
            put("secret", secret)
        }
        return postOauth(bodyJson)
    }

    suspend fun refreshToken(refreshToken: String): OpsResult<OauthTokenDto> {
        val bodyJson = buildOauthBody {
            put("grant_type", "refresh_token")
            put("refresh_token", refreshToken)
        }
        return postOauth(bodyJson)
    }

    suspend fun fetchUserInfo(session: NetworkSession): OpsResult<UserInfoDto> {
        val bodyJson = buildCommonBody("/business/ebike-management/user/getUserByToken")
        return try {
            val response = client.post("business/ebike-management/user/getUserByToken") {
                requestAuth.applyPostJson(this, bodyJson, session, AuthHeaderMode.Bearer)
                setBody(bodyJson)
            }
            parseUserInfo(response.status, response.bodyAsText())
        } catch (t: Throwable) {
            OpsResult.Err(OpsError.network(t.message ?: "user info request failed", t))
        }
    }

    suspend fun listTenants(phone: String, session: NetworkSession): OpsResult<BusinessTenantList> {
        val timestamp = nowEpochMillis().toString()
        val traceId = Ids.uuidV4()
        val normalized = loginPhone(phone)
        val signParams = mapOf(
            "izFilterRootTenant" to "true",
            "traceId" to traceId,
            "phone" to normalized,
        )
        val sign = RequestSigner.signFormParams(
            params = signParams,
            timestamp = timestamp,
            secret = signSecretProvider(),
        )
        val bodyJson = buildJsonObject {
            putCommonFields("/business/tenant/queryList")
            put("izFilterRootTenant", true)
            put("phone", normalized)
            put("traceId", traceId)
            put("time", timestamp.toLongOrNull() ?: 0L)
            put("sign", sign)
        }.toString()
        return try {
            val response = client.post("business/tenant/queryList") {
                requestAuth.applyPostJson(this, bodyJson, session, AuthHeaderMode.Bearer)
                setBody(bodyJson)
            }
            when (
                val parsed = parseEnvelope(
                    response.status,
                    response.bodyAsText(),
                    TenantListDto.serializer(),
                )
            ) {
                is OpsResult.Ok -> OpsResult.Ok(
                    BusinessTenantList(
                        tenants = parsed.value.tenantList.map { it.toDomain() },
                        secret = parsed.value.secret,
                    ),
                )
                is OpsResult.Err -> parsed
            }
        } catch (t: Throwable) {
            OpsResult.Err(OpsError.network(t.message ?: "tenant list failed", t))
        }
    }

    suspend fun logout(session: NetworkSession): OpsResult<Unit> {
        val bodyJson = buildCommonBody("/oauth/logout")
        return try {
            val response = client.post("oauth/logout") {
                requestAuth.applyPostJson(this, bodyJson, session, AuthHeaderMode.Bearer)
                setBody(bodyJson)
            }
            if (response.status.isSuccess() || response.status == HttpStatusCode.Unauthorized) {
                OpsResult.Ok(Unit)
            } else {
                OpsResult.Err(OpsError.network("logout failed: HTTP ${response.status.value}"))
            }
        } catch (t: Throwable) {
            OpsResult.Err(OpsError.network(t.message ?: "logout failed", t))
        }
    }

    private suspend fun postUnitUnauthenticated(path: String, bodyJson: String): OpsResult<Unit> {
        return try {
            val response = client.post(path) {
                requestAuth.applyPostJson(this, bodyJson, authMode = AuthHeaderMode.None)
                setBody(bodyJson)
            }
            decodeUnit(response.status, response.bodyAsText())
        } catch (t: Throwable) {
            OpsResult.Err(OpsError.network(t.message ?: "request failed", t))
        }
    }

    private suspend fun postUnitAuthenticated(
        path: String,
        bodyJson: String,
        session: NetworkSession,
    ): OpsResult<Unit> {
        return try {
            val response = client.post(path) {
                requestAuth.applyPostJson(this, bodyJson, session, AuthHeaderMode.Bearer)
                setBody(bodyJson)
            }
            decodeUnit(response.status, response.bodyAsText())
        } catch (t: Throwable) {
            OpsResult.Err(OpsError.network(t.message ?: "request failed", t))
        }
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
            val envelope = json.decodeFromString(
                ApiEnvelope.serializer(JsonElement.serializer()),
                raw,
            )
            if (envelope.isSuccessful || status.isSuccess()) {
                OpsResult.Ok(Unit)
            } else {
                OpsResult.Err(
                    OpsError.business(
                        envelope.code.ifBlank { "BUSINESS" },
                        envelope.msg ?: "request failed",
                    ),
                )
            }
        } catch (t: Throwable) {
            if (status.isSuccess()) {
                OpsResult.Ok(Unit)
            } else {
                OpsResult.Err(OpsError.network("parse failed: ${t.message}", t))
            }
        }
    }

    private suspend fun postOauth(bodyJson: String): OpsResult<OauthTokenDto> {
        return try {
            val response = client.post("oauth/token") {
                requestAuth.applyPostJson(this, bodyJson, authMode = AuthHeaderMode.BasicClient)
                setBody(bodyJson)
            }
            parseEnvelope(response.status, response.bodyAsText(), OauthTokenDto.serializer())
        } catch (t: Throwable) {
            OpsResult.Err(OpsError.network(t.message ?: "oauth request failed", t))
        }
    }

    private fun parseUserInfo(status: HttpStatusCode, raw: String): OpsResult<UserInfoDto> =
        parseEnvelope(status, raw, UserInfoDto.serializer())

    private fun <T> parseEnvelope(
        status: HttpStatusCode,
        raw: String,
        deserializer: kotlinx.serialization.KSerializer<T>,
    ): OpsResult<T> {
        if (raw.isBlank()) {
            return OpsResult.Err(OpsError.network("empty response (HTTP ${status.value})"))
        }
        return try {
            val envelope = json.decodeFromString(ApiEnvelope.serializer(deserializer), raw)
            unwrap(status, envelope)
        } catch (t: Throwable) {
            OpsResult.Err(OpsError.network("parse failed: ${t.message}", t))
        }
    }

    private fun <T> unwrap(status: HttpStatusCode, envelope: ApiEnvelope<T>): OpsResult<T> {
        val data = envelope.data
        return when {
            data != null && (envelope.isSuccessful || status.isSuccess()) -> OpsResult.Ok(data)
            data != null -> OpsResult.Err(
                OpsError.business(
                    envelope.code.ifBlank { "BUSINESS" },
                    envelope.msg ?: "request failed",
                ),
            )
            !status.isSuccess() ->
                OpsResult.Err(OpsError.network("HTTP ${status.value}: ${envelope.msg ?: "error"}"))
            else -> OpsResult.Err(
                OpsError.business(
                    envelope.code.ifBlank { "EMPTY" },
                    envelope.msg ?: "empty data",
                ),
            )
        }
    }

    private inline fun buildOauthBody(
        block: kotlinx.serialization.json.JsonObjectBuilder.() -> Unit,
    ): String {
        val obj = buildJsonObject {
            putCommonFields("oauth/token")
            block()
        }
        return obj.toString()
    }

    private fun buildCommonBody(source: String): String =
        buildJsonObject { putCommonFields(source) }.toString()

    private fun kotlinx.serialization.json.JsonObjectBuilder.putCommonFields(source: String) {
        put("traceId", Ids.uuidV4())
        put("platform", deviceInfo.platform)
        put("version", deviceInfo.appVersion)
        put("source", source)
        put("stressTesting", false)
        put("tenantId", tenantIdProvider())
        put("Accept-Language", com.luopingtech.ebike.ops.core.i18n.LocaleContext.acceptLanguage)
        put("deviceId", deviceIdProvider())
    }
}
