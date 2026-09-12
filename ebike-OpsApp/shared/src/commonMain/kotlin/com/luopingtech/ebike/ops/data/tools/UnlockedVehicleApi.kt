package com.luopingtech.ebike.ops.data.tools

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.UnlockedVehicle
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.Serializable
import kotlinx.serialization.builtins.ListSerializer
import kotlinx.serialization.json.put

@Serializable
data class UnlockedVehicleDto(
    val carId: String = "",
    val imei: String = "",
) {
    fun toDomain(): UnlockedVehicle = UnlockedVehicle(carId = carId, imei = imei)
}

/**
 * Legacy: POST /business/ebike-operation/tools/unlock_car_list
 * body: serviceId + optional name (non-numeric) or phone (numeric).
 */
class UnlockedVehicleApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    suspend fun list(serviceId: String, query: String = ""): OpsResult<List<UnlockedVehicle>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/tools/unlock_car_list",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("serviceId", serviceId)
            val trimmed = query.trim()
            if (trimmed.isNotEmpty()) {
                if (trimmed.toLongOrNull() == null) {
                    put("name", trimmed)
                } else {
                    put("phone", trimmed)
                }
            }
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-operation/tools/unlock_car_list",
                bodyJson = body,
                deserializer = ListSerializer(UnlockedVehicleDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.map { it.toDomain() })
            is OpsResult.Err -> result
        }
    }
}
