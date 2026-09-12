package com.luopingtech.ebike.ops.data.task

import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.OpsTaskType
import com.luopingtech.ebike.ops.domain.model.TaskAuditResult

interface TaskAuditRepository {
    suspend fun load(taskId: String, type: OpsTaskType): OpsResult<TaskAuditResult>
}

class TaskAuditRepositoryImpl(
    private val demoMode: Boolean,
    private val api: TaskAuditApi? = null,
) : TaskAuditRepository {
    override suspend fun load(taskId: String, type: OpsTaskType): OpsResult<TaskAuditResult> {
        if (demoMode || api == null) {
            return OpsResult.Ok(
                TaskAuditResult(
                    checkResult = 4,
                    checkView = "照片不清晰，请补拍车辆与站点全景",
                    remark = "现场完成备注（demo）",
                    photo = listOf("https://demo.cdn.ops/audit/$taskId.jpg"),
                    checkTime = "2026-09-11 12:00:00",
                    checkManName = "审核员",
                ),
            )
        }
        return api.load(taskId, type)
    }
}
