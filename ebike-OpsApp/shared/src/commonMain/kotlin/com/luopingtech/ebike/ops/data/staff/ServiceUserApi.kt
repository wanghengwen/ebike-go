package com.luopingtech.ebike.ops.data.staff

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.TeamWorker
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.Serializable
import kotlinx.serialization.builtins.ListSerializer
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.add
import kotlinx.serialization.json.buildJsonArray
import kotlinx.serialization.json.put

@Serializable
data class ServiceUserDto(
    val name: String? = null,
    val userName: String? = null,
    val phone: String? = null,
    val mobile: String? = null,
    val pin: String? = null,
) {
    fun toDomain(): TeamWorker? {
        val displayName = name?.trim().orEmpty().ifBlank { userName?.trim().orEmpty() }
        val phoneNumber = phone?.trim().orEmpty().ifBlank { mobile?.trim().orEmpty() }
        if (displayName.isBlank() && phoneNumber.isBlank()) return null
        return TeamWorker(
            name = displayName.ifBlank { phoneNumber },
            phone = phoneNumber,
        )
    }
}

/**
 * Legacy: POST /business/ebike-management/user/listByServiceIds
 * body: ids = [serviceId]
 */
class ServiceUserApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    suspend fun listByServiceId(serviceId: String): OpsResult<List<TeamWorker>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/user/listByServiceIds",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("ids", buildJsonArray { add(JsonPrimitive(serviceId)) })
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-management/user/listByServiceIds",
                bodyJson = body,
                deserializer = ListSerializer(ServiceUserDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(
                result.value.mapNotNull { it.toDomain() }
                    .distinctBy { it.selectionKey },
            )
            is OpsResult.Err -> result
        }
    }
}
