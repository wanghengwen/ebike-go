package com.luopingtech.ebike.ops.data.production

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.BindVehicleInfo
import com.luopingtech.ebike.ops.domain.model.SaddleOverloadContact
import com.luopingtech.ebike.ops.domain.model.ShelfCheckResult
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.json.add
import kotlinx.serialization.json.buildJsonArray
import kotlinx.serialization.json.put

class ProductionApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    suspend fun getBind(carId: String): OpsResult<BindVehicleInfo> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/carInfo/getBind",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("carId", carId)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-management/carInfo/getBind",
                bodyJson = body,
                deserializer = BindVehicleInfoDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain())
            is OpsResult.Err -> result
        }
    }

    suspend fun bind(carId: String, imei: String?, helmet: String?): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/carInfo/bind",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("carId", carId)
            if (!imei.isNullOrBlank()) put("imei", imei)
            if (!helmet.isNullOrBlank()) put("helmet", helmet)
        }
        return signedApi.postUnit("business/ebike-management/carInfo/bind", body)
    }

    suspend fun unbind(carId: String, imei: String?, helmet: String?): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/carInfo/unbind",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("carId", carId)
            if (!imei.isNullOrBlank()) put("imei", imei)
            if (!helmet.isNullOrBlank()) put("helmet", helmet)
        }
        return signedApi.postUnit("business/ebike-management/carInfo/unbind", body)
    }

    suspend fun onlineCheck(carId: String): OpsResult<ShelfCheckResult> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/carInfo/onlineCheck",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("carId", carId)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-management/carInfo/onlineCheck",
                bodyJson = body,
                deserializer = ShelfCheckDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain())
            is OpsResult.Err -> result
        }
    }

    suspend fun offlineCheck(carId: String): OpsResult<ShelfCheckResult> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/carInfo/offlineCheck",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("carId", carId)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-management/carInfo/offlineCheck",
                bodyJson = body,
                deserializer = ShelfCheckDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain())
            is OpsResult.Err -> result
        }
    }

    suspend fun onlineByCarList(serviceId: String, carIds: List<String>): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/carInfo/onlineByCarList",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("serviceId", serviceId)
            put("carList", buildJsonArray { carIds.forEach { add(it) } })
        }
        return signedApi.postUnit("business/ebike-management/carInfo/onlineByCarList", body)
    }

    suspend fun offlineByCarList(carIds: List<String>): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/carInfo/offlineByCarList",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("carList", buildJsonArray { carIds.forEach { add(it) } })
        }
        return signedApi.postUnit("business/ebike-management/carInfo/offlineByCarList", body)
    }

    /** Legacy Flutter: start overload check window (timeout seconds). */
    suspend fun triggerOverloadCheck(
        carId: String,
        imei: String,
        timeoutSec: Int = 60,
    ): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/paas/device/overSaddleTriggerTempState",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("carId", carId)
            put("imei", imei)
            put("timeout", timeoutSec)
            put("izSync", true)
        }
        return signedApi.postUnit("business/paas/device/overSaddleTriggerTempState", body)
    }

    /** Legacy Flutter: poll saddle pressure contacts while check window is open. */
    suspend fun queryOverloadContact(imei: String): OpsResult<SaddleOverloadContact> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/paas/device/querySaddleOverloadContact",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("imei", imei)
        }
        return when (
            val result = signedApi.post(
                path = "business/paas/device/querySaddleOverloadContact",
                bodyJson = body,
                deserializer = SaddleOverloadContactDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain())
            is OpsResult.Err -> result
        }
    }
}
