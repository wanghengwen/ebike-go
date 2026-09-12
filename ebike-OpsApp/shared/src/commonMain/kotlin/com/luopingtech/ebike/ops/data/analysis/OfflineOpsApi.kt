package com.luopingtech.ebike.ops.data.analysis

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsAnalyze
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsRankRow
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsTab
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsTrendPoint
import com.luopingtech.ebike.ops.domain.analysis.TimeRange
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.builtins.ListSerializer
import kotlinx.serialization.json.put

class OfflineOpsApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    suspend fun analyze(
        tab: OfflineOpsTab,
        serviceId: String,
        range: TimeRange,
        opPin: String? = null,
    ): OpsResult<OfflineOpsAnalyze> {
        val path = path(tab, "analyze_data")
        val body = bodyJson(path, serviceId, range, opPin)
        return when (
            val result = signedApi.post(
                path = path.removePrefix("/"),
                bodyJson = body,
                deserializer = OfflineOpsAnalyzeDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain())
            is OpsResult.Err -> result
        }
    }

    suspend fun trend(
        tab: OfflineOpsTab,
        serviceId: String,
        range: TimeRange,
        opPin: String? = null,
    ): OpsResult<List<OfflineOpsTrendPoint>> {
        val path = path(tab, "trend_poly_lines")
        val body = bodyJson(path, serviceId, range, opPin)
        return when (
            val result = signedApi.post(
                path = path.removePrefix("/"),
                bodyJson = body,
                deserializer = ListSerializer(OfflineOpsSeriesDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.map { it.toTrend() })
            is OpsResult.Err -> result
        }
    }

    suspend fun rank(
        tab: OfflineOpsTab,
        serviceId: String,
        range: TimeRange,
        opPin: String? = null,
    ): OpsResult<List<OfflineOpsRankRow>> {
        val path = path(tab, "num_rank")
        val body = bodyJson(path, serviceId, range, opPin)
        return when (
            val result = signedApi.post(
                path = path.removePrefix("/"),
                bodyJson = body,
                deserializer = ListSerializer(OfflineOpsSeriesDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.map { it.toRank() })
            is OpsResult.Err -> result
        }
    }

    private fun bodyJson(
        source: String,
        serviceId: String,
        range: TimeRange,
        opPin: String? = null,
    ): String =
        CommonRequestBody.toJsonString(
            source = source,
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            val asLong = serviceId.toLongOrNull()
            if (asLong != null) put("serviceId", asLong) else put("serviceId", serviceId)
            put("startTime", range.startText)
            put("endTime", range.endText)
            if (!opPin.isNullOrBlank()) put("opPin", opPin)
        }

    private fun path(tab: OfflineOpsTab, action: String): String {
        val type = when (tab) {
            OfflineOpsTab.ChangeBattery -> "change_battery"
            OfflineOpsTab.MoveCar -> "move_car"
            OfflineOpsTab.Inspection -> "alarm"
            OfflineOpsTab.Repair -> "fix"
        }
        return "/business/ebike-operation/$type/task/$action"
    }
}
