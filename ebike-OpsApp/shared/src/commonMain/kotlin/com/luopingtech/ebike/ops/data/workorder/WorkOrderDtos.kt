package com.luopingtech.ebike.ops.data.workorder

import com.luopingtech.ebike.ops.domain.model.WorkOrder
import com.luopingtech.ebike.ops.domain.model.WorkOrderKind
import kotlinx.serialization.Serializable

@Serializable
data class AlarmTicketPageDto(
    val list: List<AlarmTicketDto> = emptyList(),
    val count: Int? = null,
    val pageNum: Int? = null,
    val pageSize: Int? = null,
)

@Serializable
data class AlarmTicketDto(
    val id: Long? = null,
    val serviceId: Long? = null,
    val imei: String? = null,
    val carId: String? = null,
    val opManName: String? = null,
    val opManPhone: String? = null,
    val type: Int? = null,
    val state: Int? = null,
    val fixedTime: String? = null,
    val createdAt: String? = null,
) {
    fun toDomain(fallbackServiceId: String = ""): WorkOrder = WorkOrder(
        id = id?.toString().orEmpty(),
        kind = WorkOrderKind.Inspection,
        carId = carId.orEmpty(),
        imei = imei.orEmpty(),
        state = state ?: 0,
        alarmType = type,
        opManName = opManName.orEmpty(),
        opManPhone = opManPhone.orEmpty(),
        createdAt = createdAt.orEmpty(),
        fixedTime = fixedTime.orEmpty(),
        serviceId = serviceId?.toString()?.ifBlank { fallbackServiceId } ?: fallbackServiceId,
    )
}

@Serializable
data class FixTicketPageDto(
    val list: List<FixTicketDto> = emptyList(),
    val count: Int? = null,
    val pageNum: Int? = null,
    val pageSize: Int? = null,
)

@Serializable
data class FixTicketDto(
    val id: Long? = null,
    val createdAt: String? = null,
    val carId: String? = null,
    val imei: String? = null,
    val type: Long? = null,
    val name: String? = null,
    val opManName: String? = null,
    val opManPhone: String? = null,
    val state: Int? = null,
    val fixedTime: String? = null,
    val carModel: String? = null,
    val fixReason: String? = null,
    val nameExtraInfo: List<String>? = null,
    val carState: Boolean? = null,
    val source: Int? = null,
    val lat: Double? = null,
    val lng: Double? = null,
    val serviceId: Long? = null,
) {
    fun toDomain(fallbackServiceId: String = ""): WorkOrder = WorkOrder(
        id = id?.toString().orEmpty(),
        kind = WorkOrderKind.Repair,
        carId = carId.orEmpty(),
        imei = imei.orEmpty(),
        state = state ?: 0,
        partNames = nameExtraInfo.orEmpty().filter { it.isNotBlank() }.ifEmpty {
            listOfNotNull(name?.takeIf { it.isNotBlank() })
        },
        fixReason = fixReason.orEmpty(),
        opManName = opManName.orEmpty(),
        opManPhone = opManPhone.orEmpty(),
        createdAt = createdAt.orEmpty(),
        fixedTime = fixedTime.orEmpty(),
        carStopped = carState == true,
        serviceId = serviceId?.toString()?.ifBlank { fallbackServiceId } ?: fallbackServiceId,
    )
}
