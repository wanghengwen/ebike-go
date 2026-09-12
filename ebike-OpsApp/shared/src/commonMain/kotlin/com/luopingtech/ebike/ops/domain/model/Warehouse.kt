package com.luopingtech.ebike.ops.domain.model

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings

/** Legacy: 1 = 领用出库, 2 = 归还入库. */
enum class WarehouseOperationType(val code: Int) {
    Out(1),
    In(2),
    ;

    val label: String
        get() = when (this) {
            Out -> Strings.t(Str.WarehouseOut)
            In -> Strings.t(Str.WarehouseIn)
        }

    companion object {
        fun fromCode(code: Int?): WarehouseOperationType? =
            entries.firstOrNull { it.code == code }
    }
}

data class WarehouseComponent(
    val id: String = "",
    val componentNo: String = "",
    val componentName: String = "",
    val componentClassify: Int = 1,
    val brand: String = "",
    val existCode: Int = 1,
    val stockQuantity: Int = 0,
    val receivedQuantity: Int = 0,
    val operationType: Int? = null,
    val operationTypeName: String = "",
    val operationNum: Int = 0,
    val receiver: String = "",
    val operator: String = "",
    val operationTime: String = "",
)

data class WarehouseRecord(
    val id: String,
    val componentName: String = "",
    val operationType: WarehouseOperationType? = null,
    val operationTypeName: String = "",
    val operationNum: Int = 0,
    val receiver: String = "",
    val operator: String = "",
    val operationTime: String = "",
    val existCode: Int = 1,
)
