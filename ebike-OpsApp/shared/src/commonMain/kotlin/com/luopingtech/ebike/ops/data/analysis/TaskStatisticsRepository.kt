package com.luopingtech.ebike.ops.data.analysis

import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.task.ChangeBatteryTaskApi
import com.luopingtech.ebike.ops.data.task.ChangeBatteryTaskDto
import com.luopingtech.ebike.ops.data.task.ChangeBatteryTaskListDto
import com.luopingtech.ebike.ops.data.task.InspectionTaskApi
import com.luopingtech.ebike.ops.data.task.MoveCarTaskApi
import com.luopingtech.ebike.ops.data.task.RepairTaskApi
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsPeriod
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsTimeRanges
import com.luopingtech.ebike.ops.domain.analysis.TaskStatisticsItem
import com.luopingtech.ebike.ops.domain.analysis.TaskStatisticsKind
import com.luopingtech.ebike.ops.domain.analysis.TaskStatisticsPage
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.task.TaskWaitingTime

interface TaskStatisticsRepository {
    suspend fun load(
        area: ServiceArea,
        kind: TaskStatisticsKind,
        period: OfflineOpsPeriod,
        valid: Boolean,
        pageNum: Int,
        pageSize: Int = 10,
        opPin: String,
    ): OpsResult<TaskStatisticsPage>
}

class TaskStatisticsRepositoryImpl(
    private val demoMode: Boolean,
    private val changeBatteryApi: ChangeBatteryTaskApi? = null,
    private val moveCarApi: MoveCarTaskApi? = null,
    private val inspectionApi: InspectionTaskApi? = null,
    private val repairApi: RepairTaskApi? = null,
) : TaskStatisticsRepository {
    override suspend fun load(
        area: ServiceArea,
        kind: TaskStatisticsKind,
        period: OfflineOpsPeriod,
        valid: Boolean,
        pageNum: Int,
        pageSize: Int,
        opPin: String,
    ): OpsResult<TaskStatisticsPage> {
        if (demoMode || changeBatteryApi == null) {
            return OpsResult.Ok(demoPage(kind, valid, pageNum, pageSize))
        }
        val range = OfflineOpsTimeRanges.period(period)
        val state = if (valid) 2 else 3
        val result = when (kind) {
            TaskStatisticsKind.ChangeBattery -> changeBatteryApi.pageStats(
                serviceId = area.id,
                state = state,
                pageNum = pageNum,
                pageSize = pageSize,
                opPin = opPin,
                finishTimeBegin = range.startText,
                finishTimeEnd = range.endText,
            )
            TaskStatisticsKind.MoveCar -> moveCarApi?.pageStats(
                serviceId = area.id,
                state = state,
                pageNum = pageNum,
                pageSize = pageSize,
                opPin = opPin,
                taskEndTimeBegin = range.startText,
                taskEndTimeEnd = range.endText,
            ) ?: return OpsResult.Ok(demoPage(kind, valid, pageNum, pageSize))
            TaskStatisticsKind.Inspection -> inspectionApi?.pageStats(
                serviceId = area.id,
                state = state,
                pageNum = pageNum,
                pageSize = pageSize,
                opPin = opPin,
                finishTimeStart = range.startText,
                finishTimeEnd = range.endText,
            ) ?: return OpsResult.Ok(demoPage(kind, valid, pageNum, pageSize))
            TaskStatisticsKind.Repair -> repairApi?.pageStats(
                serviceId = area.id,
                state = state,
                pageNum = pageNum,
                pageSize = pageSize,
                opPin = opPin,
                finishTimeBegin = range.startText,
                finishTimeEnd = range.endText,
            ) ?: return OpsResult.Ok(demoPage(kind, valid, pageNum, pageSize))
        }
        return when (result) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toPage(kind, pageNum))
            is OpsResult.Err -> result
        }
    }

    private fun ChangeBatteryTaskListDto.toPage(
        kind: TaskStatisticsKind,
        pageNum: Int,
    ): TaskStatisticsPage = TaskStatisticsPage(
        items = list.map { it.toStatisticsItem(kind) },
        total = count ?: list.size,
        pageNum = pageNum,
    )

    companion object {
        fun demoPage(
            kind: TaskStatisticsKind,
            valid: Boolean,
            pageNum: Int,
            pageSize: Int,
        ): TaskStatisticsPage {
            val all = (1..6).map { i ->
                ChangeBatteryTaskDto(
                    id = "stat-$kind-$i",
                    carId = "10060${i}23",
                    startTime = "2026-09-12 09:0$i:00",
                    finishTime = "2026-09-12 10:1$i:00",
                    openBatBoxTime = "2026-09-12 09:0$i:00",
                    closeBatBoxTime = "2026-09-12 09:1$i:00",
                    restBatteryBefore = 15 + i,
                    restBatteryAfter = 80 + i,
                    lastTime = (300L + i * 40),
                    distance = 0.5 + i * 0.1,
                    source = if (i % 2 == 0) 2 else 1,
                    startFixTime = "2026-09-12 09:0$i:00",
                ).toStatisticsItem(kind).let {
                    if (!valid && i > 3) it.copy(durationText = TaskWaitingTime.formatSeconds(120)) else it
                }
            }.filterIndexed { index, _ -> if (valid) index < 4 else index >= 4 }
            val from = ((pageNum - 1) * pageSize).coerceAtLeast(0)
            val slice = all.drop(from).take(pageSize)
            return TaskStatisticsPage(items = slice, total = all.size, pageNum = pageNum)
        }
    }
}
