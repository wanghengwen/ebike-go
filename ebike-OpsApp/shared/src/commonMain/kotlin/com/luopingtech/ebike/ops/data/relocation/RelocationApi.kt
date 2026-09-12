package com.luopingtech.ebike.ops.data.relocation

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.RelocationDevice
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.add
import kotlinx.serialization.json.put
import kotlinx.serialization.json.putJsonArray

@Serializable
data class RelocationDeviceDto(
    val carId: String? = null,
    val imei: String? = null,
    val lat: Double? = null,
    val lng: Double? = null,
    val restBattery: Int? = null,
    val isOnline: Int? = null,
) {
    fun toDomain(): RelocationDevice = RelocationDevice(
        carId = carId.orEmpty(),
        imei = imei.orEmpty(),
        lat = lat ?: 0.0,
        lng = lng ?: 0.0,
        restBattery = restBattery ?: 0,
        isOnline = isOnline ?: 0,
        selected = true,
    )
}

/**
 * Legacy Flutter RelocationRepository:
 * - POST /business/paas/device/deviceScan
 * - POST /business/paas/device/ble/report/locationReport
 */
class RelocationApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    suspend fun deviceScan(
        serviceId: String,
        carId: String? = null,
        imei: String? = null,
    ): OpsResult<RelocationDevice> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/paas/device/deviceScan",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("serviceId", serviceId)
            if (!carId.isNullOrBlank()) put("carId", carId)
            if (!imei.isNullOrBlank()) put("imei", imei)
        }
        return when (
            val result = signedApi.post(
                path = "business/paas/device/deviceScan",
                bodyJson = body,
                deserializer = RelocationDeviceDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain())
            is OpsResult.Err -> result
        }
    }

    suspend fun locationReport(
        imeiList: List<String>,
        latitude: Double,
        longitude: Double,
    ): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/paas/device/ble/report/locationReport",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            putJsonArray("imeiList") {
                imeiList.forEach { add(it) }
            }
            put("lat", latitude)
            put("lng", longitude)
        }
        return signedApi.postUnit("business/paas/device/ble/report/locationReport", body)
    }
}
