package com.luopingtech.ebike.ops.data.task

import com.luopingtech.ebike.ops.domain.analysis.TaskStatisticsItem
import com.luopingtech.ebike.ops.domain.analysis.TaskStatisticsKind
import com.luopingtech.ebike.ops.domain.model.OpsTask
import com.luopingtech.ebike.ops.domain.model.OpsTaskType
import com.luopingtech.ebike.ops.domain.task.TaskWaitingTime
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
    val startTime: String = "",
    val finishTime: String = "",
    val startFixTime: String = "",
    val serviceId: String = "",
    val gps: TaskGpsDto? = null,
    val dragState: Int? = null,
    val source: Int? = null,
    val moveType: Int? = null,
    val recordId: String? = null,
    val izStop: Boolean? = null,
    val restBatteryBefore: Int? = null,
    val restBatteryAfter: Int? = null,
    val openBatBoxTime: String = "",
    val closeBatBoxTime: String = "",
    val lastTime: Long? = null,
    val distance: Double? = null,
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
        startTime = startTime.ifBlank { startFixTime },
        finishTime = finishTime,
        dragState = dragState,
        source = source,
        moveType = moveType,
        recordId = recordId,
        izStop = izStop == true,
    )

    fun toStatisticsItem(kind: TaskStatisticsKind): TaskStatisticsItem {
        val start = when (kind) {
            TaskStatisticsKind.ChangeBattery -> openBatBoxTime.ifBlank { startTime }
            TaskStatisticsKind.Repair -> startFixTime.ifBlank { startTime }
            else -> startTime
        }
        val end = when (kind) {
            TaskStatisticsKind.ChangeBattery -> closeBatBoxTime.ifBlank { finishTime }
            else -> finishTime
        }
        val duration = when {
            lastTime != null && lastTime > 0 -> TaskWaitingTime.formatSeconds(lastTime)
            else -> TaskWaitingTime.format(start, end)
        }
        val sourceLabel = when (source) {
            2 -> "人工任务"
            1 -> "运维规则"
            else -> ""
        }
        return TaskStatisticsItem(
            id = id,
            carId = carId,
            kind = kind,
            startTime = start,
            finishTime = end,
            durationText = duration,
            restBatteryBefore = restBatteryBefore,
            restBatteryAfter = restBatteryAfter,
            openBatBoxTime = openBatBoxTime,
            closeBatBoxTime = closeBatBoxTime,
            distance = distance?.let { "${it}km" }.orEmpty(),
            sourceLabel = sourceLabel,
        )
    }
}

@Serializable
data class TaskGpsDto(
    val lat: Double = 0.0,
    val lng: Double = 0.0,
)
