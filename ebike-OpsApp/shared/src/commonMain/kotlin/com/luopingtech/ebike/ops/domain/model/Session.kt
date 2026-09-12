package com.luopingtech.ebike.ops.domain.model

data class UserSession(
    val userId: String = "",
    val displayName: String = "",
    /** Login / profile mobile — used when staff-record filter (122802) is absent. */
    val phone: String = "",
    val accessToken: String = "",
    val refreshToken: String = "",
    val tenantId: String = "",
    val serviceAreaId: String = "",
    val serviceAreaName: String = "",
    /** Legacy permission codes from getUserByToken (`codes`). */
    val permissionCodes: List<String> = emptyList(),
    /** From getUserByToken; false → force set-password before shell. */
    val hasPassword: Boolean = true,
    /** Role display name from getUserByToken (`roleName`). */
    val roleName: String = "",
)

data class ServiceArea(
    val id: String,
    val name: String,
    val centerLat: Double = 0.0,
    val centerLng: Double = 0.0,
    val agentId: String = "",
    val minDistance: Double = 0.0,
)
