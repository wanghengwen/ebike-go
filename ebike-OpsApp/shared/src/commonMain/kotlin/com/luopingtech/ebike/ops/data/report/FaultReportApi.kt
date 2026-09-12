package com.luopingtech.ebike.ops.data.report

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.FaultReportRecord
import com.luopingtech.ebike.ops.domain.model.RepairType
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.builtins.ListSerializer
import kotlinx.serialization.json.add
import kotlinx.serialization.json.buildJsonArray
import kotlinx.serialization.json.put

class FaultReportApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    suspend fun repairTypeList(model: String = ""): OpsResult<List<RepairType>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/repairConfig/list",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            if (model.isNotBlank()) put("carModel", model)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-management/repairConfig/list",
                bodyJson = body,
                deserializer = ListSerializer(RepairTypeDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.map { it.toDomain() })
            is OpsResult.Err -> result
        }
    }

    suspend fun submitRepair(
        carId: String,
        fixReason: String,
        typeIds: List<Long>,
        typeNames: List<String>,
        photoUrls: List<String>,
        izStop: Boolean?,
    ): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/repair/add",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("carId", carId)
            put("fixReason", fixReason)
            put("types", buildJsonArray { typeIds.forEach { add(it) } })
            put("typeNames", buildJsonArray { typeNames.forEach { add(it) } })
            put("photo", buildJsonArray { photoUrls.forEach { add(it) } })
            if (izStop != null) put("izStop", izStop)
        }
        return signedApi.postUnit("business/ebike-operation/repair/add", body)
    }

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

    /**
     * History of artificial repair tickets (repair/add → TaskFixTicket).
     * Legacy merchant app had no "my reports" list; OpsApp uses fix/task/page
     * with taskSource=2. Cannot filter by reporter pin on current query DTO.
     */
    suspend fun myReports(
        serviceId: String,
        pageNum: Int = 1,
        pageSize: Int = 50,
    ): OpsResult<List<FaultReportRecord>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/fix/task/page",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("serviceId", serviceId)
            put("pageNum", pageNum)
            put("pageSize", pageSize)
            put("taskSource", 2)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-operation/fix/task/page",
                bodyJson = body,
                deserializer = RepairHistoryPageDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.list.map { it.toDomain() })
            is OpsResult.Err -> result
        }
    }
}
