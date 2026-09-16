package com.luopingtech.ebike.rider.core.h5

import com.luopingtech.ebike.rider.domain.model.UserSession
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.put

/**
 * 给 H5 的展示资料。不含 access/refresh token，也不含 signSecret。
 */
object H5ProfileView {
    val FORBIDDEN_KEYS: Set<String> = setOf(
        "accesstoken",
        "refreshtoken",
        "token",
        "tokentype",
        "sign",
        "signsecret",
        "businesssecret",
        "tenantsecret",
        "authorization",
    )

    fun from(
        session: UserSession?,
        tenantId: String,
        /** 首页定位得到的服务区，优先于登录资料里可能为空的字段。 */
        serviceAreaId: String = "",
    ): JsonObject {
        val s = session
        val sid = serviceAreaId.ifBlank { s?.serviceAreaId.orEmpty() }
        return buildJsonObject {
            put("loggedIn", s != null)
            put("userId", s?.userId.orEmpty())
            put("pin", s?.pin.orEmpty())
            put("displayName", s?.displayName.orEmpty())
            put("phone", s?.phone.orEmpty())
            put("avatar", s?.avatar.orEmpty())
            put("balance", s?.balance ?: 0L)
            put("tenantId", s?.tenantId?.ifBlank { tenantId }.orEmpty().ifBlank { tenantId })
            put("serviceAreaId", sid)
        }
    }

    fun containsSecrets(obj: JsonObject): Boolean =
        obj.keys.any { it.lowercase() in FORBIDDEN_KEYS }
}
