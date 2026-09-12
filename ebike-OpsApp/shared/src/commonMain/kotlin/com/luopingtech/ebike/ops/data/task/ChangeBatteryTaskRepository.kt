package com.luopingtech.ebike.ops.data.task

import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.BatteryRange
import com.luopingtech.ebike.ops.domain.model.OpsTask
import com.luopingtech.ebike.ops.domain.model.OpsTaskType
import com.luopingtech.ebike.ops.domain.model.ServiceArea

interface ChangeBatteryTaskRepository {
    suspend fun batteryRange(area: ServiceArea): OpsResult<BatteryRange>
    suspend fun list(area: ServiceArea, maxBattery: Int = 30): OpsResult<List<OpsTask>>
    suspend fun detail(taskId: String): OpsResult<OpsTask>
    suspend fun openBatteryBox(
        task: OpsTask,
        pin: String,
        izBlue: Boolean = false,
    ): OpsResult<Unit>

    suspend fun closeBatteryBox(
        task: OpsTask,
        pin: String,
        izBlue: Boolean = false,
    ): OpsResult<Unit>
}

class ChangeBatteryTaskRepositoryImpl(
    private val demoMode: Boolean,
    private val api: ChangeBatteryTaskApi? = null,
) : ChangeBatteryTaskRepository {
    private val demoTasks = linkedMapOf<String, OpsTask>()

    override suspend fun batteryRange(area: ServiceArea): OpsResult<BatteryRange> {
        if (demoMode || api == null) {
            return OpsResult.Ok(BatteryRange(min = 5, max = 30))
        }
        return api.batteryRange(area.id)
    }

    override suspend fun list(area: ServiceArea, maxBattery: Int): OpsResult<List<OpsTask>> {
        if (demoMode || api == null) {
            val tasks = demoCatalog(area).filter { it.restBattery <= maxBattery }
            tasks.forEach { demoTasks[it.id] = it }
            return OpsResult.Ok(tasks)
        }
        return api.list(serviceId = area.id, maxBattery = maxBattery)
    }

    override suspend fun detail(taskId: String): OpsResult<OpsTask> {
        if (demoMode || api == null) {
            demoTasks[taskId]?.let { return OpsResult.Ok(it) }
            return OpsResult.Err(OpsError.business("TASK_NOT_FOUND", "task $taskId not found"))
        }
        return api.detail(taskId)
    }

    override suspend fun openBatteryBox(
        task: OpsTask,
        pin: String,
        izBlue: Boolean,
    ): OpsResult<Unit> {
        if (demoMode || api == null) {
            val next = task.copy(state = 1)
            demoTasks[task.id] = next
            return OpsResult.Ok(Unit)
        }
        return api.start(
            taskId = task.id,
            serviceId = task.serviceId.ifBlank { task.areaId },
            pin = pin,
            izBlue = izBlue,
        )
    }

    override suspend fun closeBatteryBox(
        task: OpsTask,
        pin: String,
        izBlue: Boolean,
    ): OpsResult<Unit> {
        if (demoMode || api == null) {
            val next = task.copy(state = 2, checkResult = 2)
            demoTasks[task.id] = next
            return OpsResult.Ok(Unit)
        }
        return api.finish(
            taskId = task.id,
            serviceId = task.serviceId.ifBlank { task.areaId },
            pin = pin,
            izBlue = izBlue,
        )
    }

    companion object {
        fun demoCatalog(area: ServiceArea): List<OpsTask> {
            val baseLat = if (area.centerLat != 0.0) area.centerLat else 28.22
            val baseLng = if (area.centerLng != 0.0) area.centerLng else 112.94
            return listOf(
                OpsTask(
                    id = "cb-${area.id}-1",
                    type = OpsTaskType.ChangeBattery,
                    carId = "D${area.id}-001",
                    imei = "860000000000001",
                    address = "${area.name} · gate A",
                    areaId = area.id,
                    areaName = area.name,
                    state = 0,
                    restBattery = 18,
                    lat = baseLat + 0.001,
                    lng = baseLng + 0.001,
                    serviceId = area.id,
                    createdAt = "2026-09-10T10:00:00Z",
                ),
                OpsTask(
                    id = "cb-${area.id}-2",
                    type = OpsTaskType.ChangeBattery,
                    carId = "D${area.id}-003",
                    imei = "860000000000003",
                    address = "${area.name} · station B",
                    areaId = area.id,
                    areaName = area.name,
                    state = 1,
                    restBattery = 9,
                    lat = baseLat - 0.0008,
                    lng = baseLng + 0.0005,
                    serviceId = area.id,
                    createdAt = "2026-09-10T11:30:00Z",
                ),
                OpsTask(
                    id = "cb-${area.id}-3",
                    type = OpsTaskType.ChangeBattery,
                    carId = "D${area.id}-004",
                    imei = "860000000000004",
                    address = "${area.name} · parking C",
                    areaId = area.id,
                    areaName = area.name,
                    state = 0,
                    restBattery = 42,
                    lat = baseLat + 0.0003,
                    lng = baseLng - 0.001,
                    serviceId = area.id,
                    createdAt = "2026-09-10T12:00:00Z",
                ),
            )
        }
    }
}
