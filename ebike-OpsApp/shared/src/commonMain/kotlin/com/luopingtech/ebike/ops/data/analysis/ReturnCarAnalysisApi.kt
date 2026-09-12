package com.luopingtech.ebike.ops.data.analysis

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.analysis.ReturnCarAnalyzeResult
import com.luopingtech.ebike.ops.domain.analysis.TimeRange
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.json.put

class ReturnCarAnalysisApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    suspend fun analyze(
        serviceId: String,
        range: TimeRange,
    ): OpsResult<ReturnCarAnalyzeResult> {
        val path = "/business/orderAnalyze/getOrderAnalyze"
        val body = CommonRequestBody.toJsonString(
            source = path,
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            val asLong = serviceId.toLongOrNull()
            if (asLong != null) put("serviceId", asLong) else put("serviceId", serviceId)
            put("start", range.startText)
            put("end", range.endText)
        }
        return when (
            val result = signedApi.post(
                path = path.removePrefix("/"),
                bodyJson = body,
                deserializer = ReturnCarAnalyzeDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain())
            is OpsResult.Err -> result
        }
    }
}
