package com.luopingtech.ebike.ops.data.warehouse

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.WarehouseComponent
import com.luopingtech.ebike.ops.domain.model.WarehouseOperationType
import com.luopingtech.ebike.ops.domain.model.WarehouseRecord
import com.luopingtech.ebike.ops.domain.warehouse.WarehouseComponentFilter

interface WarehouseRepository {
    suspend fun queryByComponentNo(componentNo: String): OpsResult<WarehouseComponent>
    suspend fun operateByCode(
        components: List<WarehouseComponent>,
        operationType: WarehouseOperationType,
        receiver: String,
        pin: String,
    ): OpsResult<Unit>

    suspend fun operateWithoutCode(
        componentName: String,
        componentClassify: Int,
        operationType: WarehouseOperationType,
        operationNum: Int,
        receiver: String,
    ): OpsResult<Unit>

    suspend fun pageRecords(
        operationType: WarehouseOperationType? = null,
        componentName: String = "",
        keyWord: String = "",
    ): OpsResult<List<WarehouseRecord>>

    suspend fun detailList(recordId: String): OpsResult<List<WarehouseComponent>>
    suspend fun queryAllComponentNames(existCode: Int? = null): OpsResult<List<String>>
    suspend fun queryQuantity(componentName: String): OpsResult<WarehouseComponent>
}

class WarehouseRepositoryImpl(
    private val demoMode: Boolean,
    private val api: WarehouseApi? = null,
) : WarehouseRepository {
    private val demoRecords = mutableListOf(
        WarehouseRecord(
            id = "WR-1001",
            componentName = Strings.t(Str.RepairTypeBattery),
            operationType = WarehouseOperationType.Out,
            operationTypeName = WarehouseOperationType.Out.label,
            operationNum = 2,
            receiver = "demo-user",
            operator = "demo-user",
            operationTime = "2026-09-10 10:00",
            existCode = 1,
        ),
        WarehouseRecord(
            id = "WR-1002",
            componentName = Strings.t(Str.RepairTypeBattery),
            operationType = WarehouseOperationType.In,
            operationTypeName = WarehouseOperationType.In.label,
            operationNum = 1,
            receiver = "demo-user",
            operator = "demo-user",
            operationTime = "2026-09-10 12:00",
            existCode = 0,
        ),
    )

    override suspend fun queryByComponentNo(componentNo: String): OpsResult<WarehouseComponent> {
        if (componentNo.isBlank()) {
            return OpsResult.Err(OpsError.business("WH_EMPTY", "componentNo empty"))
        }
        if (demoMode || api == null) {
            return OpsResult.Ok(
                WarehouseComponent(
                    id = "C-$componentNo",
                    componentNo = componentNo.trim(),
                    componentName = Strings.t(Str.RepairTypeBattery),
                    componentClassify = 1,
                    brand = "Demo",
                    existCode = 1,
                    stockQuantity = 12,
                ),
            )
        }
        return api.queryByComponentNo(componentNo.trim())
    }

    override suspend fun operateByCode(
        components: List<WarehouseComponent>,
        operationType: WarehouseOperationType,
        receiver: String,
        pin: String,
    ): OpsResult<Unit> {
        if (components.isEmpty()) {
            return OpsResult.Err(OpsError.business("WH_EMPTY", "no components"))
        }
        if (demoMode || api == null) {
            demoRecords.add(
                0,
                WarehouseRecord(
                    id = "WR-D${demoRecords.size + 1}",
                    componentName = components.first().componentName.ifBlank {
                        Strings.t(Str.ComponentPart)
                    },
                    operationType = operationType,
                    operationTypeName = operationType.label,
                    operationNum = components.size,
                    receiver = receiver.ifBlank { pin },
                    operator = pin,
                    operationTime = "demo-now",
                    existCode = 1,
                ),
            )
            return OpsResult.Ok(Unit)
        }
        return api.operateByCode(components, operationType, receiver, pin)
    }

    override suspend fun operateWithoutCode(
        componentName: String,
        componentClassify: Int,
        operationType: WarehouseOperationType,
        operationNum: Int,
        receiver: String,
    ): OpsResult<Unit> {
        if (componentName.isBlank() || operationNum <= 0) {
            return OpsResult.Err(OpsError.business("WH_INVALID", "name or quantity invalid"))
        }
        if (demoMode || api == null) {
            demoRecords.add(
                0,
                WarehouseRecord(
                    id = "WR-D${demoRecords.size + 1}",
                    componentName = componentName,
                    operationType = operationType,
                    operationTypeName = operationType.label,
                    operationNum = operationNum,
                    receiver = receiver,
                    operator = receiver,
                    operationTime = "demo-now",
                    existCode = 0,
                ),
            )
            return OpsResult.Ok(Unit)
        }
        return api.operateWithoutCode(
            componentName = componentName,
            componentClassify = componentClassify,
            operationType = operationType,
            operationNum = operationNum,
            receiver = receiver,
        )
    }

    override suspend fun pageRecords(
        operationType: WarehouseOperationType?,
        componentName: String,
        keyWord: String,
    ): OpsResult<List<WarehouseRecord>> {
        if (demoMode || api == null) {
            var list = demoRecords.toList()
            if (operationType != null) {
                list = list.filter { it.operationType == operationType }
            }
            if (componentName.isNotBlank() && !WarehouseComponentFilter.isAll(componentName)) {
                list = list.filter { it.componentName.contains(componentName) }
            }
            if (keyWord.isNotBlank()) {
                list = list.filter {
                    it.id.contains(keyWord, ignoreCase = true) ||
                        it.receiver.contains(keyWord, ignoreCase = true) ||
                        it.componentName.contains(keyWord, ignoreCase = true)
                }
            }
            return OpsResult.Ok(list)
        }
        return api.pageRecords(
            operationType = operationType,
            componentName = componentName,
            keyWord = keyWord,
        )
    }

    override suspend fun detailList(recordId: String): OpsResult<List<WarehouseComponent>> {
        if (demoMode || api == null) {
            val record = demoRecords.firstOrNull { it.id == recordId }
                ?: return OpsResult.Err(OpsError.business("WH_NOT_FOUND", "record not found"))
            return OpsResult.Ok(
                List(record.operationNum.coerceAtLeast(1)) { index ->
                    WarehouseComponent(
                        id = "$recordId-$index",
                        componentNo = if (record.existCode == 1) "BAT-${1000 + index}" else "",
                        componentName = record.componentName,
                        componentClassify = 1,
                        operationType = record.operationType?.code,
                        operationTypeName = record.operationTypeName,
                        operationNum = 1,
                        receiver = record.receiver,
                        operator = record.operator,
                        operationTime = record.operationTime,
                        existCode = record.existCode,
                    )
                },
            )
        }
        return api.detailList(recordId)
    }

    override suspend fun queryAllComponentNames(existCode: Int?): OpsResult<List<String>> {
        if (demoMode || api == null) {
            return OpsResult.Ok(
                WarehouseComponentFilter.toDisplayList(
                    listOf(
                        WarehouseComponentFilter.API_ALL,
                        Strings.t(Str.RepairTypeBattery),
                        Strings.t(Str.ComponentHelmet),
                        Strings.t(Str.RepairTypeEcu),
                    ),
                ),
            )
        }
        return when (val result = api.queryAllComponentNames(existCode)) {
            is OpsResult.Ok -> OpsResult.Ok(WarehouseComponentFilter.toDisplayList(result.value))
            is OpsResult.Err -> result
        }
    }

    override suspend fun queryQuantity(componentName: String): OpsResult<WarehouseComponent> {
        if (demoMode || api == null) {
            return OpsResult.Ok(
                WarehouseComponent(
                    componentName = componentName,
                    stockQuantity = 24,
                    receivedQuantity = 3,
                ),
            )
        }
        return api.queryQuantity(componentName)
    }
}
