package com.luopingtech.ebike.ops.data.order

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.LastOrder
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.json.put

class OrderApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    /** Legacy: /business/order/detailLast �?last order + embedded deviceTrajectory. */
    suspend fun detailLast(carId: String): OpsResult<LastOrder> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/order/detailLast",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("carId", carId)
        }
        return when (
            val result = signedApi.post(
                path = "business/order/detailLast",
                bodyJson = body,
                deserializer = LastOrderDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain())
            is OpsResult.Err -> result
        }
    }
}
