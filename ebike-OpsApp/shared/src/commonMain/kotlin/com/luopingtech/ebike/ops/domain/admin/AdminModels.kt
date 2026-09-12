package com.luopingtech.ebike.ops.domain.admin

data class CareerAuditItem(
    val id: String,
    val name: String,
    val phone: String,
    val auditState: Int,
    val createdAt: String = "",
)

data class ObjectionOrderItem(
    val id: String,
    val orderId: String,
    val userName: String,
    val phone: String,
    val carId: String,
    val state: Int,
    val userReason: String = "",
    val createdAt: String = "",
)

data class BlacklistItem(
    val id: String,
    val authName: String,
    val phone: String,
    val reason: String,
    val state: Int,
    val createdAt: String = "",
)

data class IdBindAuditItem(
    val id: String,
    val authName: String,
    val applyPhone: String,
    val originPhone: String,
    val auditState: Int,
    val createdAt: String = "",
)

data class OperationLogItem(
    val time: String,
    val operatorName: String,
    val content: String,
    val carId: String = "",
)
