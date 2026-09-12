package com.luopingtech.ebike.ops.data.auth

import com.luopingtech.ebike.ops.domain.model.BusinessTenant
import com.luopingtech.ebike.ops.domain.model.TenantRuntimeConfig
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonNames

@Serializable
data class OauthTokenDto(
    @SerialName("accessToken")
    @JsonNames("access_token")
    val accessToken: String = "",
    @SerialName("refreshToken")
    @JsonNames("refresh_token")
    val refreshToken: String = "",
    @SerialName("tokenType")
    @JsonNames("token_type")
    val tokenType: String = "Bearer",
    @SerialName("expiresIn")
    @JsonNames("expires_in")
    val expiresIn: Long = 0,
    val phone: String = "",
    val pin: String = "",
    val nickname: String = "",
    val avatar: String = "",
    val grantType: String = "",
    val jti: String = "",
)

@Serializable
data class UserInfoDto(
    val id: String = "",
    val userId: String = "",
    val pin: String = "",
    val phone: String = "",
    val nickname: String = "",
    val userName: String = "",
    val realName: String = "",
    val avatar: String = "",
    val tenantId: String = "",
    val tenantName: String = "",
    val serviceAreaId: String = "",
    val codes: List<String> = emptyList(),
    val hasPassword: Boolean = true,
    /** HQ account may pick operator; department accounts skip the picker (legacy `izRoot`). */
    val izRoot: Boolean = false,
    val roleName: String = "",
)

@Serializable
data class AppConfigDto(
    val tenantId: String = "",
    val tenantName: String = "",
    val alias: String = "",
    val clientDomain: String = "",
    val businessDomain: String = "",
    val qrCodes: List<String> = emptyList(),
) {
    fun toDomain(): TenantRuntimeConfig =
        TenantRuntimeConfig(
            tenantId = tenantId,
            tenantName = tenantName,
            alias = alias,
            clientDomain = clientDomain,
            businessDomain = businessDomain,
            qrHosts = qrCodes.filter { it.isNotBlank() },
        )
}

@Serializable
data class TenantListDto(
    val tenantList: List<TenantItemDto> = emptyList(),
    val secret: String = "",
)

@Serializable
data class TenantItemDto(
    val tenantId: String = "",
    val tenantName: String = "",
    val companyName: String = "",
    val aliasName: String = "",
    val alias: String = "",
) {
    fun toDomain(): BusinessTenant = BusinessTenant(
        tenantId = tenantId,
        tenantName = tenantName,
        companyName = companyName,
        aliasName = aliasName.ifBlank { alias },
    )
}
