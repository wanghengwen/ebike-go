package com.luopingtech.ebike.ops.domain.analysis

import com.luopingtech.ebike.ops.domain.model.OpsTaskType

enum class TaskStatisticsKind {
    ChangeBattery,
    MoveCar,
    Inspection,
    Repair,
}

data class TaskStatisticsItem(
    val id: String,
    val carId: String,
    val kind: TaskStatisticsKind,
    val startTime: String = "",
    val finishTime: String = "",
    val durationText: String = "",
    val restBatteryBefore: Int? = null,
    val restBatteryAfter: Int? = null,
    val openBatBoxTime: String = "",
    val closeBatBoxTime: String = "",
    val distance: String = "",
    val sourceLabel: String = "",
)

data class TaskStatisticsPage(
    val items: List<TaskStatisticsItem> = emptyList(),
    val total: Int = 0,
    val pageNum: Int = 1,
) {
    val hasMore: Boolean get() = items.size < total
}

fun TaskStatisticsKind.toOpsTaskType(): OpsTaskType = when (this) {
    TaskStatisticsKind.ChangeBattery -> OpsTaskType.ChangeBattery
    TaskStatisticsKind.MoveCar -> OpsTaskType.MoveCar
    TaskStatisticsKind.Inspection -> OpsTaskType.Inspection
    TaskStatisticsKind.Repair -> OpsTaskType.Repair
}
