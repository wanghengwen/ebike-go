package com.luopingtech.ebike.ops.data.task

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.OpsTask
import com.luopingtech.ebike.ops.domain.model.OpsTaskType
import com.luopingtech.ebike.ops.domain.model.ServiceArea

interface MoveCarTaskRepository {
    suspend fun list(area: ServiceArea, pageNum: Int = 1, pageSize: Int = 50): OpsResult<List<OpsTask>>
    suspend fun detail(taskId: String): OpsResult<OpsTask>
    suspend fun claim(task: OpsTask, signPin: String): OpsResult<Unit>
    suspend fun start(task: OpsTask): OpsResult<Unit>
    suspend fun finish(task: OpsTask, pictures: List<String> = emptyList(), remark: String? = null): OpsResult<Unit>
}

class MoveCarTaskRepositoryImpl(
    private val demoMode: Boolean,
    private val api: MoveCarTaskApi? = null,
) : MoveCarTaskRepository {
    private val demoTasks = linkedMapOf<String, OpsTask>()

    override suspend fun list(
        area: ServiceArea,
        pageNum: Int,
        pageSize: Int,
    ): OpsResult<List<OpsTask>> {
        if (demoMode || api == null) {
            val tasks = demoCatalog(area)
            tasks.forEach { demoTasks[it.id] = it }
            return OpsResult.Ok(tasks)
        }
        return api.list(serviceId = area.id, pageNum = pageNum, pageSize = pageSize)
    }

    override suspend fun detail(taskId: String): OpsResult<OpsTask> {
        if (demoMode || api == null) {
            demoTasks[taskId]?.let { return OpsResult.Ok(it) }
            return OpsResult.Err(OpsError.business("TASK_NOT_FOUND", "task $taskId not found"))
        }
        return api.detail(taskId)
    }

    override suspend fun claim(task: OpsTask, signPin: String): OpsResult<Unit> {
        if (demoMode || api == null) {
            demoTasks[task.id] = task.copy(state = 1)
            return OpsResult.Ok(Unit)
        }
        return api.claim(taskId = task.id, signPin = signPin)
    }

    override suspend fun start(task: OpsTask): OpsResult<Unit> {
        if (demoMode || api == null) {
            demoTasks[task.id] = task.copy(state = 1)
            return OpsResult.Ok(Unit)
        }
        return api.start(task.id)
    }

    override suspend fun finish(task: OpsTask, pictures: List<String>, remark: String?): OpsResult<Unit> {
        if (demoMode || api == null) {
            // Align with legacy: first finish without photos → 23326 need photograph.
            if (pictures.isEmpty()) {
                return OpsResult.Err(
                    OpsError.business("23326", Strings.t(Str.NeedPhotoAudit)),
                )
            }
            demoTasks[task.id] = task.copy(state = 2, checkResult = 2)
            return OpsResult.Ok(Unit)
        }
        return api.finish(task.id, pictures, remark)
    }

    companion object {
        fun demoCatalog(area: ServiceArea): List<OpsTask> {
            val baseLat = if (area.centerLat != 0.0) area.centerLat else 28.22
            val baseLng = if (area.centerLng != 0.0) area.centerLng else 112.94
            return listOf(
                OpsTask(
                    id = "mc-${area.id}-1",
                    type = OpsTaskType.MoveCar,
                    carId = "D${area.id}-002",
                    imei = "860000000000002",
                    address = "${area.name} · pile-up zone",
                    areaId = area.id,
                    areaName = area.name,
                    state = 0,
                    restBattery = 55,
                    lat = baseLat - 0.0012,
                    lng = baseLng + 0.0008,
                    serviceId = area.id,
                    createdAt = "2026-09-10T09:00:00Z",
                ),
                OpsTask(
                    id = "mc-${area.id}-2",
                    type = OpsTaskType.MoveCar,
                    carId = "D${area.id}-004",
                    imei = "860000000000004",
                    address = "${area.name} · no-parking",
                    areaId = area.id,
                    areaName = area.name,
                    state = 1,
                    restBattery = 70,
                    lat = baseLat + 0.0004,
                    lng = baseLng - 0.0009,
                    serviceId = area.id,
                    createdAt = "2026-09-10T09:45:00Z",
                ),
                OpsTask(
                    id = "mc-${area.id}-done",
                    type = OpsTaskType.MoveCar,
                    carId = "D${area.id}-006",
                    imei = "860000000000006",
                    address = "${area.name} · audit rejected",
                    areaId = area.id,
                    areaName = area.name,
                    state = 2,
                    checkResult = 4,
                    restBattery = 48,
                    lat = baseLat,
                    lng = baseLng,
                    serviceId = area.id,
                    createdAt = "2026-09-10T10:00:00Z",
                ),
                OpsTask(
                    id = "mc-${area.id}-batch",
                    type = OpsTaskType.MoveCar,
                    carId = "",
                    imei = "",
                    address = "${area.name} · man-made batch",
                    areaId = area.id,
                    areaName = area.name,
                    state = 1,
                    restBattery = 0,
                    lat = baseLat - 0.0006,
                    lng = baseLng + 0.0011,
                    serviceId = area.id,
                    createdAt = "2026-09-10T10:00:00Z",
                    source = 2,
                    moveType = 2,
                    recordId = "mc-${area.id}-batch",
                ),
            )
        }
    }
}
