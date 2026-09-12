package com.luopingtech.ebike.ops.data.vehicle

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.Vehicle
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.builtins.ListSerializer
import kotlinx.serialization.json.add
import kotlinx.serialization.json.put
import kotlinx.serialization.json.putJsonArray

class VehicleApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    suspend fun listByServiceIds(serviceIds: List<String>): OpsResult<List<Vehicle>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/paas/device/list",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            putJsonArray("serviceIdList") {
                serviceIds.forEach { id ->
                    val asLong = id.toLongOrNull()
                    if (asLong != null) add(asLong) else add(id)
                }
            }
        }
        return when (
            val result = signedApi.post(
                path = "business/paas/device/list",
                bodyJson = body,
                deserializer = ListSerializer(VehicleDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.map { it.toDomain() })
            is OpsResult.Err -> result
        }
    }

    suspend fun getDetail(carId: String? = null, imei: String? = null): OpsResult<Vehicle> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/paas/device/detail",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            if (!carId.isNullOrBlank()) put("carId", carId)
            if (!imei.isNullOrBlank()) put("imei", imei)
        }
        return when (
            val result = signedApi.post(
                path = "business/paas/device/detail",
                bodyJson = body,
                deserializer = VehicleDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain())
            is OpsResult.Err -> result
        }
    }

    /** Legacy carPermissionCheck — car must belong to the selected service area. */
    suspend fun checkServicePermission(carId: String, serviceId: String): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/carInfo/carPermissionCheck",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("carId", carId)
            put("serviceId", serviceId)
        }
        return signedApi.postUnit("business/ebike-management/carInfo/carPermissionCheck", body)
    }

    /** Legacy ReplaceBatteryRepositoryV2.bindBatterySn. */
    suspend fun bindBatterySn(carId: String, batterySn: String): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/carInfo/bindBatterySn",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("carId", carId)
            put("batterySn", batterySn)
        }
        return signedApi.postUnit("business/ebike-management/carInfo/bindBatterySn", body)
    }
}
