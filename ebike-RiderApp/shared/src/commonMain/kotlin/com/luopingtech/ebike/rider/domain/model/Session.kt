package com.luopingtech.ebike.rider.domain.model

/**
 * C 端登录态。token 落 [com.luopingtech.ebike.rider.platform.SecureStore]，
 * 资料字段来自 `/oauth/token` + `/client/user/user/personInfo`。
 */
data class UserSession(
    val userId: String = "",
    val pin: String = "",
    val displayName: String = "",
    val phone: String = "",
    val accessToken: String = "",
    val refreshToken: String = "",
    val tenantId: String = "",
    val serviceAreaId: String = "",
    val avatar: String = "",
    val balance: Long = 0,
)
