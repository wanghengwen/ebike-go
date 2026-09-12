package com.luopingtech.ebike.ops.data.analysis

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.analysis.StationAnalyzeDetail
import com.luopingtech.ebike.ops.domain.analysis.StationAnalyzePage
import com.luopingtech.ebike.ops.domain.analysis.StationOptStateFilter
import com.luopingtech.ebike.ops.domain.analysis.StationSortOrder
import com.luopingtech.ebike.ops.domain.analysis.StationTag
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.builtins.ListSerializer
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.buildJsonArray
import kotlinx.serialization.json.put

class StationAnalysisApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    suspend fun list(
        serviceId: String,
        pageNum: Int,
        pageSize: Int,
        optState: StationOptStateFilter,
        order: StationSortOrder,
        tagIds: List<String>,
    ): OpsResult<StationAnalyzePage> {
        val path = "/business/fence/parking/site/analyze"
        val body = CommonRequestBody.toJsonString(
            source = path,
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            val asLong = serviceId.toLongOrNull()
            if (asLong != null) put("id", asLong) else put("id", serviceId)
            put("pageNum", pageNum)
            put("pageSize", pageSize)
            order.apiValue?.let { put("order", it) }
            when (optState) {
                StationOptStateFilter.All -> Unit
                StationOptStateFilter.Operating -> put("optState", "1")
                StationOptStateFilter.Stopped -> put("optState", "0")
            }
            if (tagIds.isNotEmpty()) {
                put("tags", buildJsonArray { tagIds.forEach { add(JsonPrimitive(it)) } })
            }
        }
        return when (
            val result = signedApi.post(
                path = path.removePrefix("/"),
                bodyJson = body,
                deserializer = StationAnalyzePageDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain())
            is OpsResult.Err -> result
        }
    }

    suspend fun analyzeOne(
        parkingId: String,
        serviceId: String,
    ): OpsResult<StationAnalyzeDetail> {
        val path = "/business/fence/parking/site/analyzeOne"
        val body = CommonRequestBody.toJsonString(
            source = path,
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            val pLong = parkingId.toLongOrNull()
            val sLong = serviceId.toLongOrNull()
            if (pLong != null) put("parkingId", pLong) else put("parkingId", parkingId)
            if (sLong != null) put("serviceId", sLong) else put("serviceId", serviceId)
        }
        return when (
            val result = signedApi.post(
                path = path.removePrefix("/"),
                bodyJson = body,
                deserializer = StationAnalyzeOneDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain())
            is OpsResult.Err -> result
        }
    }

    suspend fun tags(): OpsResult<List<StationTag>> {
        val path = "/business/fence/tags/getAll"
        val body = CommonRequestBody.toJsonString(
            source = path,
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) { }
        return when (
            val result = signedApi.post(
                path = path.removePrefix("/"),
                bodyJson = body,
                deserializer = ListSerializer(StationTagDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.map { it.toDomain() }.filter { it.id.isNotBlank() })
            is OpsResult.Err -> result
        }
    }
}
