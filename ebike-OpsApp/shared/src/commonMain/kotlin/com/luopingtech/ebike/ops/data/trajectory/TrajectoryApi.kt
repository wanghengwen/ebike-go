package com.luopingtech.ebike.ops.data.trajectory

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.TrackPoint
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.builtins.ListSerializer
import kotlinx.serialization.json.put

class TrajectoryApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    /** Legacy: /business/paas/device/trajectory/realTime */
    suspend fun realTime(
        imei: String,
        startTimeMs: Long,
        endTimeMs: Long,
    ): OpsResult<List<TrackPoint>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/paas/device/trajectory/realTime",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("imei", imei)
            put("startTime", startTimeMs)
            put("endTime", endTimeMs)
        }
        return when (
            val result = signedApi.post(
                path = "business/paas/device/trajectory/realTime",
                bodyJson = body,
                deserializer = ListSerializer(TrackPointDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.mapNotNull { it.toDomain() })
            is OpsResult.Err -> result
        }
    }
}
