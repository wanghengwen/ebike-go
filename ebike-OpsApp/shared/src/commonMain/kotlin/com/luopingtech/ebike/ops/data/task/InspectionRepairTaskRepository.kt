package com.luopingtech.ebike.ops.data.task

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.OpsTask
import com.luopingtech.ebike.ops.domain.model.OpsTaskType
import com.luopingtech.ebike.ops.domain.model.ServiceArea

interface InspectionTaskRepository {
    suspend fun list(area: ServiceArea, opPin: String): OpsResult<List<OpsTask>>
    suspend fun claim(task: OpsTask, signPin: String): OpsResult<Unit>
    suspend fun assign(tasks: List<OpsTask>, signPin: String): OpsResult<Unit>
    suspend fun start(task: OpsTask, pin: String): OpsResult<Unit>
    suspend fun finish(task: OpsTask, pictures: List<String> = emptyList(), remark: String? = null): OpsResult<Unit>
}

class InspectionTaskRepositoryImpl(
    private val demoMode: Boolean,
    private val api: InspectionTaskApi? = null,
) : InspectionTaskRepository {
    private val demoTasks = linkedMapOf<String, OpsTask>()

    override suspend fun list(area: ServiceArea, opPin: String): OpsResult<List<OpsTask>> {
        if (demoMode || api == null) {
            val tasks = demoCatalog(area)
            tasks.forEach { demoTasks[it.id] = it }
            return OpsResult.Ok(tasks)
        }
        val pending = api.listPage(serviceId = area.id, state = 0)
        if (pending is OpsResult.Err) return pending
        val processing = api.listPage(serviceId = area.id, state = 1, opPin = opPin)
        if (processing is OpsResult.Err) return processing
        // Finished tickets so field ops can open audit result (legacy showAuditResultView).
        val finished = api.listPage(serviceId = area.id, state = 2, opPin = opPin)
        if (finished is OpsResult.Err) return finished
        val merged = (
            pending.getOrNull().orEmpty() +
                processing.getOrNull().orEmpty() +
                finished.getOrNull().orEmpty()
            )
            .distinctBy { it.id }
        return OpsResult.Ok(merged)
    }

    override suspend fun claim(task: OpsTask, signPin: String): OpsResult<Unit> {
        if (demoMode || api == null) {
            demoTasks[task.id] = task.copy(state = 1)
            return OpsResult.Ok(Unit)
        }
        return api.claim(listOf(task.id), signPin)
    }

    override suspend fun assign(tasks: List<OpsTask>, signPin: String): OpsResult<Unit> {
        if (tasks.isEmpty()) {
            return OpsResult.Err(OpsError.business("NO_SEL", Strings.t(Str.SelectTaskFirst)))
        }
        if (demoMode || api == null) {
            tasks.forEach { demoTasks[it.id] = it.copy(state = 1) }
            return OpsResult.Ok(Unit)
        }
        return api.assign(type = 2, taskIds = tasks.map { it.id }, signPin = signPin)
    }

    override suspend fun start(task: OpsTask, pin: String): OpsResult<Unit> {
        if (demoMode || api == null) {
            demoTasks[task.id] = task.copy(state = 1)
            return OpsResult.Ok(Unit)
        }
        return api.start(task.id, pin)
    }

    override suspend fun finish(task: OpsTask, pictures: List<String>, remark: String?): OpsResult<Unit> {
        if (demoMode || api == null) {
            if (pictures.isEmpty()) {
                return OpsResult.Err(OpsError.business("23326", Strings.t(Str.NeedPhotoAudit)))
            }
            demoTasks[task.id] = task.copy(state = 2, checkResult = 2)
            return OpsResult.Ok(Unit)
        }
        return api.finish(task.id, pictures, remark)
    }

    companion object {
        fun demoCatalog(area: ServiceArea): List<OpsTask> = listOf(
            OpsTask(
                id = "insp-${area.id}-1",
                type = OpsTaskType.Inspection,
                carId = "D${area.id}-001",
                address = "异常离线",
                areaId = area.id,
                areaName = area.name,
                state = 0,
                restBattery = 60,
                serviceId = area.id,
                createdAt = "2026-09-04 18:00:00",
            ),
            OpsTask(
                id = "insp-${area.id}-2",
                type = OpsTaskType.Inspection,
                carId = "D${area.id}-002",
                address = "长时未还",
                areaId = area.id,
                areaName = area.name,
                state = 1,
                restBattery = 40,
                serviceId = area.id,
                createdAt = "2026-09-10 09:00:00",
                startTime = "2026-09-11 10:00:00",
            ),
            OpsTask(
                id = "insp-${area.id}-done",
                type = OpsTaskType.Inspection,
                carId = "D${area.id}-003",
                address = "${area.name} · audit rejected",
                areaId = area.id,
                areaName = area.name,
                state = 2,
                checkResult = 4,
                restBattery = 55,
                serviceId = area.id,
            ),
        )
    }
}

interface RepairTaskRepository {
    suspend fun list(area: ServiceArea, opPin: String): OpsResult<List<OpsTask>>
    suspend fun claim(task: OpsTask, signPin: String): OpsResult<Unit>
    suspend fun assign(tasks: List<OpsTask>, signPin: String): OpsResult<Unit>
    suspend fun start(task: OpsTask, pin: String): OpsResult<Unit>
    suspend fun finish(task: OpsTask, pictures: List<String> = emptyList(), remark: String? = null): OpsResult<Unit>
    suspend fun createDrag(task: OpsTask, dragReason: String, dragAddress: String): OpsResult<Unit>
}

class RepairTaskRepositoryImpl(
    private val demoMode: Boolean,
    private val api: RepairTaskApi? = null,
) : RepairTaskRepository {
    private val demoTasks = linkedMapOf<String, OpsTask>()

    override suspend fun list(area: ServiceArea, opPin: String): OpsResult<List<OpsTask>> {
        if (demoMode || api == null) {
            val tasks = demoCatalog(area)
            tasks.forEach { demoTasks[it.id] = it }
            return OpsResult.Ok(tasks)
        }
        val pending = api.listPage(serviceId = area.id, state = 0)
        if (pending is OpsResult.Err) return pending
        val processing = api.listPage(serviceId = area.id, state = 1, opPin = opPin)
        if (processing is OpsResult.Err) return processing
        val finished = api.listPage(serviceId = area.id, state = 2, opPin = opPin)
        if (finished is OpsResult.Err) return finished
        val merged = (
            pending.getOrNull().orEmpty() +
                processing.getOrNull().orEmpty() +
                finished.getOrNull().orEmpty()
            )
            .distinctBy { it.id }
        return OpsResult.Ok(merged)
    }

    override suspend fun claim(task: OpsTask, signPin: String): OpsResult<Unit> {
        if (demoMode || api == null) {
            demoTasks[task.id] = task.copy(state = 1)
            return OpsResult.Ok(Unit)
        }
        return api.claim(task.id, signPin)
    }

    override suspend fun assign(tasks: List<OpsTask>, signPin: String): OpsResult<Unit> {
        val task = tasks.firstOrNull()
            ?: return OpsResult.Err(OpsError.business("NO_SEL", Strings.t(Str.SelectTaskFirst)))
        if (demoMode || api == null) {
            demoTasks[task.id] = task.copy(state = 1)
            return OpsResult.Ok(Unit)
        }
        return api.assign(type = 2, taskId = task.id, signPin = signPin)
    }

    override suspend fun start(task: OpsTask, pin: String): OpsResult<Unit> {
        if (demoMode || api == null) {
            demoTasks[task.id] = task.copy(state = 1)
            return OpsResult.Ok(Unit)
        }
        return api.start(task.id, pin)
    }

    override suspend fun finish(task: OpsTask, pictures: List<String>, remark: String?): OpsResult<Unit> {
        if (demoMode || api == null) {
            if (task.dragState == 2) {
                return OpsResult.Err(
                    OpsError.business("DRAG_BLOCK", Strings.t(Str.DragBackBlockFinish)),
                )
            }
            if (pictures.isEmpty()) {
                return OpsResult.Err(OpsError.business("23326", Strings.t(Str.NeedPhotoAudit)))
            }
            demoTasks[task.id] = task.copy(state = 2, checkResult = 2)
            return OpsResult.Ok(Unit)
        }
        return api.finish(task.id, pictures, remark)
    }

    override suspend fun createDrag(
        task: OpsTask,
        dragReason: String,
        dragAddress: String,
    ): OpsResult<Unit> {
        if (demoMode || api == null) {
            if (dragReason.isBlank()) {
                return OpsResult.Err(OpsError.business("DRAG_REASON", Strings.t(Str.DragBackReasonRequired)))
            }
            if (dragAddress.isBlank()) {
                return OpsResult.Err(OpsError.business("DRAG_ADDR", Strings.t(Str.DragBackAddressRequired)))
            }
            if (task.dragState == 2) {
                return OpsResult.Err(
                    OpsError.business("DRAG_BUSY", Strings.t(Str.DragBackInProgress)),
                )
            }
            if (task.dragState == 3) {
                return OpsResult.Err(
                    OpsError.business("DRAG_DONE", Strings.t(Str.DragBackDone)),
                )
            }
            demoTasks[task.id] = task.copy(dragState = 2, state = 1)
            return OpsResult.Ok(Unit)
        }
        return api.createDrag(
            taskId = task.id,
            carId = task.carId,
            areaId = task.areaId.ifBlank { task.serviceId },
            dragReason = dragReason,
            dragAddress = dragAddress,
        )
    }

    companion object {
        fun demoCatalog(area: ServiceArea): List<OpsTask> = listOf(
            OpsTask(
                id = "rp-${area.id}-1",
                type = OpsTaskType.Repair,
                carId = "D${area.id}-003",
                address = "${area.name} · brake fault",
                areaId = area.id,
                areaName = area.name,
                state = 0,
                restBattery = 22,
                serviceId = area.id,
                dragState = 1,
                izStop = true,
                createdAt = "2026-09-05 08:00:00",
            ),
            OpsTask(
                id = "rp-${area.id}-1b",
                type = OpsTaskType.Repair,
                carId = "D${area.id}-006",
                address = "${area.name} · tire",
                areaId = area.id,
                areaName = area.name,
                state = 0,
                restBattery = 40,
                serviceId = area.id,
                dragState = 1,
                izStop = false,
                createdAt = "2026-09-11 12:00:00",
            ),
            OpsTask(
                id = "rp-${area.id}-2",
                type = OpsTaskType.Repair,
                carId = "D${area.id}-004",
                address = "${area.name} · screen damage",
                areaId = area.id,
                areaName = area.name,
                state = 1,
                restBattery = 15,
                serviceId = area.id,
                dragState = 1,
                izStop = true,
            ),
            OpsTask(
                id = "rp-${area.id}-done",
                type = OpsTaskType.Repair,
                carId = "D${area.id}-005",
                address = "${area.name} · audit rejected",
                areaId = area.id,
                areaName = area.name,
                state = 2,
                checkResult = 4,
                restBattery = 18,
                serviceId = area.id,
                dragState = 1,
            ),
        )
    }
}
