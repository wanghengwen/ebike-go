package com.luopingtech.ebike.ops.data.task

import com.luopingtech.ebike.ops.domain.model.OpsTask
import com.luopingtech.ebike.ops.domain.model.OpsTaskType
import kotlinx.serialization.Serializable

@Serializable
data class ChangeBatteryTaskListDto(
    val list: List<ChangeBatteryTaskDto> = emptyList(),
    val count: Int? = null,
)

@Serializable
data class ChangeBatteryTaskDto(
    val id: String = "",
    val carId: String = "",
    val imei: String = "",
    val address: String = "",
    val areaId: String = "",
    val areaName: String = "",
    val state: Int = 0,
    val checkResult: Int? = null,
    val restBattery: Int? = 0,
    val createdAt: String = "",
    val serviceId: String = "",
    val gps: TaskGpsDto? = null,
    val dragState: Int? = null,
    val source: Int? = null,
    val moveType: Int? = null,
    val recordId: String? = null,
) {
    fun toDomain(fallbackServiceId: String = ""): OpsTask = OpsTask(
        id = id,
        type = OpsTaskType.ChangeBattery,
        carId = carId,
        imei = imei,
        address = address,
        areaId = areaId,
        areaName = areaName,
        state = state,
        checkResult = checkResult,
        restBattery = restBattery ?: 0,
        lat = gps?.lat ?: 0.0,
        lng = gps?.lng ?: 0.0,
        serviceId = serviceId.ifBlank { fallbackServiceId },
        createdAt = createdAt,
        dragState = dragState,
        source = source,
        moveType = moveType,
        recordId = recordId,
    )
}

@Serializable
data class TaskGpsDto(
    val lat: Double = 0.0,
    val lng: Double = 0.0,
)
