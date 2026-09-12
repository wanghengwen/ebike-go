package com.luopingtech.ebike.ops.data.warehouse

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.WarehouseComponent
import com.luopingtech.ebike.ops.domain.model.WarehouseOperationType
import com.luopingtech.ebike.ops.domain.model.WarehouseRecord
import com.luopingtech.ebike.ops.domain.warehouse.WarehouseComponentFilter
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.builtins.ListSerializer
import kotlinx.serialization.builtins.serializer
import kotlinx.serialization.json.addJsonObject
import kotlinx.serialization.json.buildJsonArray
import kotlinx.serialization.json.put

class WarehouseApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    suspend fun queryByComponentNo(componentNo: String): OpsResult<WarehouseComponent> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/component/queryByComponentNo",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("componentNo", componentNo)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-management/component/queryByComponentNo",
                bodyJson = body,
                deserializer = WarehouseComponentDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toComponent())
            is OpsResult.Err -> result
        }
    }

    suspend fun operateByCode(
        components: List<WarehouseComponent>,
        operationType: WarehouseOperationType,
        receiver: String,
        pin: String,
    ): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/component/record/operateByCode",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("operationType", operationType.code)
            put("receiver", receiver)
            put("pin", pin)
            put(
                "detailDTOList",
                buildJsonArray {
                    components.forEach { c ->
                        addJsonObject {
                            put("id", c.id)
                            put("componentNo", c.componentNo)
                            put("componentName", c.componentName)
                            put("componentClassify", c.componentClassify)
                            put("brand", c.brand)
                            put("existCode", c.existCode)
                        }
                    }
                },
            )
        }
        return signedApi.postUnit(
            path = "business/ebike-management/component/record/operateByCode",
            bodyJson = body,
        )
    }

    suspend fun operateWithoutCode(
        componentName: String,
        componentClassify: Int,
        operationType: WarehouseOperationType,
        operationNum: Int,
        receiver: String,
    ): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/component/record/operateWithOutCode",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("componentName", componentName)
            put("componentClassify", componentClassify)
            put("operationType", operationType.code)
            put("operationNum", operationNum)
            put("receiver", receiver)
        }
        return signedApi.postUnit(
            path = "business/ebike-management/component/record/operateWithOutCode",
            bodyJson = body,
        )
    }

    suspend fun pageRecords(
        pageNum: Int = 1,
        pageSize: Int = 20,
        operationType: WarehouseOperationType? = null,
        componentName: String = "",
        keyWord: String = "",
    ): OpsResult<List<WarehouseRecord>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/component/record/pageList",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("pageNum", pageNum)
            put("pageSize", pageSize)
            if (operationType != null) put("operationType", operationType.code)
            val filterName = WarehouseComponentFilter.toApiFilter(componentName)
            if (filterName.isNotEmpty()) {
                put("componentName", filterName)
            }
            if (keyWord.isNotBlank()) put("keyWord", keyWord)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-management/component/record/pageList",
                bodyJson = body,
                deserializer = WarehousePageDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.list.map { it.toRecord() })
            is OpsResult.Err -> result
        }
    }

    suspend fun detailList(
        recordId: String,
        pageNum: Int = 1,
        pageSize: Int = 50,
    ): OpsResult<List<WarehouseComponent>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/component/record/detailList",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("recordId", recordId)
            put("pageNum", pageNum)
            put("pageSize", pageSize)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-management/component/record/detailList",
                bodyJson = body,
                deserializer = WarehousePageDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.list.map { it.toComponent() })
            is OpsResult.Err -> result
        }
    }

    suspend fun queryAllComponentNames(existCode: Int? = null): OpsResult<List<String>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/component/configuration/queryAllComponentNameList",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            if (existCode != null) put("existCode", existCode)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-management/component/configuration/queryAllComponentNameList",
                bodyJson = body,
                deserializer = ListSerializer(String.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value)
            is OpsResult.Err -> result
        }
    }

    suspend fun queryQuantity(componentName: String): OpsResult<WarehouseComponent> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/component/configuration/queryQuantity",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("componentName", componentName)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-management/component/configuration/queryQuantity",
                bodyJson = body,
                deserializer = WarehouseComponentDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toComponent())
            is OpsResult.Err -> result
        }
    }
}
