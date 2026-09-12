package com.luopingtech.ebike.ops.data.warehouse

import com.luopingtech.ebike.ops.domain.model.WarehouseComponent
import com.luopingtech.ebike.ops.domain.model.WarehouseOperationType
import com.luopingtech.ebike.ops.domain.model.WarehouseRecord
import kotlinx.serialization.Serializable

@Serializable
data class WarehouseComponentDto(
    val id: String = "",
    val brand: String = "",
    val componentClassify: Int = 1,
    val componentName: String = "",
    val componentNo: String = "",
    val existCode: Int = 1,
    val stockQuantity: Int = 0,
    val receivedQuantity: Int = 0,
    val operationType: Int? = null,
    val operationTypeName: String = "",
    val operationNum: Int = 0,
    val receiver: String = "",
    val operator: String = "",
    val operationTime: String = "",
) {
    fun toComponent(): WarehouseComponent = WarehouseComponent(
        id = id,
        componentNo = componentNo,
        componentName = componentName,
        componentClassify = componentClassify,
        brand = brand,
        existCode = existCode,
        stockQuantity = stockQuantity,
        receivedQuantity = receivedQuantity,
        operationType = operationType,
        operationTypeName = operationTypeName,
        operationNum = operationNum,
        receiver = receiver,
        operator = operator,
        operationTime = operationTime,
    )

    fun toRecord(): WarehouseRecord = WarehouseRecord(
        id = id.ifBlank { componentNo },
        componentName = componentName,
        operationType = WarehouseOperationType.fromCode(operationType),
        operationTypeName = operationTypeName.ifBlank {
            WarehouseOperationType.fromCode(operationType)?.label.orEmpty()
        },
        operationNum = operationNum,
        receiver = receiver,
        operator = operator,
        operationTime = operationTime,
        existCode = existCode,
    )
}

@Serializable
data class WarehousePageDto(
    val count: String = "",
    val pageNum: Int = 1,
    val pageSize: Int = 10,
    val list: List<WarehouseComponentDto> = emptyList(),
)

@Serializable
data class WarehouseNameListDto(
    val list: List<String> = emptyList(),
)
