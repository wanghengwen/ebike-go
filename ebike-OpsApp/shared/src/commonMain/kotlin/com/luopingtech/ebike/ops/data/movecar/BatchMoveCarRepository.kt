package com.luopingtech.ebike.ops.data.movecar

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.BatchMoveChild

interface BatchMoveCarRepository {
    suspend fun listChildren(parentTaskId: String): OpsResult<List<BatchMoveChild>>
    suspend fun startBatch(childTaskIds: List<String>, pin: String): OpsResult<Unit>
    suspend fun finishBatch(
        childTaskIds: List<String>,
        pin: String,
        pictures: List<String> = emptyList(),
        remark: String? = null,
    ): OpsResult<Unit>
}

class BatchMoveCarRepositoryImpl(
    private val demoMode: Boolean,
    private val api: BatchMoveCarApi? = null,
) : BatchMoveCarRepository {
    private val demoByParent = linkedMapOf<String, MutableList<BatchMoveChild>>()

    override suspend fun listChildren(parentTaskId: String): OpsResult<List<BatchMoveChild>> {
        if (demoMode || api == null) {
            val list = demoByParent.getOrPut(parentTaskId) {
                demoCatalog(parentTaskId).toMutableList()
            }
            return OpsResult.Ok(list.toList())
        }
        return api.listChildren(parentTaskId)
    }

    override suspend fun startBatch(childTaskIds: List<String>, pin: String): OpsResult<Unit> {
        if (demoMode || api == null) {
            if (childTaskIds.isEmpty()) {
                return OpsResult.Err(OpsError.business("BATCH_NONE", Strings.t(Str.BatchMoveNeedSelect)))
            }
            demoByParent.keys.toList().forEach { parentId ->
                val children = demoByParent[parentId] ?: return@forEach
                demoByParent[parentId] = children.map { child ->
                    if (child.taskId in childTaskIds && !child.isFinished) {
                        child.copy(state = 1)
                    } else {
                        child
                    }
                }.toMutableList()
            }
            return OpsResult.Ok(Unit)
        }
        return api.startBatch(childTaskIds, pin)
    }

    override suspend fun finishBatch(
        childTaskIds: List<String>,
        pin: String,
        pictures: List<String>,
        remark: String?,
    ): OpsResult<Unit> {
        if (demoMode || api == null) {
            if (childTaskIds.isEmpty()) {
                return OpsResult.Err(OpsError.business("BATCH_NONE", Strings.t(Str.BatchMoveNeedSelect)))
            }
            // Align with legacy VehicleBatchMoveCar: finish without photo → 23011.
            if (pictures.isEmpty()) {
                return OpsResult.Err(
                    OpsError.business(CODE_NEED_PHOTOGRAPH, Strings.t(Str.NeedPhotoAudit)),
                )
            }
            demoByParent.keys.toList().forEach { parentId ->
                val children = demoByParent[parentId] ?: return@forEach
                demoByParent[parentId] = children.map { child ->
                    if (child.taskId in childTaskIds) {
                        child.copy(state = 2, izFinish = true)
                    } else {
                        child
                    }
                }.toMutableList()
            }
            return OpsResult.Ok(Unit)
        }
        return api.finishBatch(childTaskIds, pin, pictures, remark)
    }

    companion object {
        const val CODE_NEED_PHOTOGRAPH = "23011"

        fun demoCatalog(parentTaskId: String): List<BatchMoveChild> {
            val suffix = parentTaskId.takeLast(4).ifBlank { "batch" }
            return listOf(
                BatchMoveChild(
                    taskId = "$parentTaskId-c1",
                    carId = "D-$suffix-A",
                    imei = "860000000000101",
                    restBattery = 42,
                    state = 0,
                ),
                BatchMoveChild(
                    taskId = "$parentTaskId-c2",
                    carId = "D-$suffix-B",
                    imei = "860000000000102",
                    restBattery = 38,
                    state = 1,
                ),
                BatchMoveChild(
                    taskId = "$parentTaskId-c3",
                    carId = "D-$suffix-C",
                    imei = "860000000000103",
                    restBattery = 51,
                    state = 0,
                ),
            )
        }
    }
}
