package com.luopingtech.ebike.ops.feature.task

import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.task.TaskAuditRepository
import com.luopingtech.ebike.ops.domain.model.OpsTask
import com.luopingtech.ebike.ops.domain.model.OpsTaskType
import com.luopingtech.ebike.ops.domain.model.TaskAuditResult
import kotlinx.coroutines.runBlocking
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class TaskAuditFeatureTest {
    @Test
    fun loadsRejectedAudit() = runBlocking {
        val feature = TaskAuditFeature(
            repository = object : TaskAuditRepository {
                override suspend fun load(taskId: String, type: OpsTaskType) =
                    OpsResult.Ok(
                        TaskAuditResult(
                            checkResult = 4,
                            checkView = "retake",
                            remark = "field",
                            photo = listOf("https://x/a.jpg"),
                        ),
                    )
            },
        )
        val task = OpsTask(
            id = "t1",
            type = OpsTaskType.Repair,
            state = 2,
            checkResult = 4,
        )
        assertTrue(feature.load(task).isOk)
        assertEquals(4, feature.state.value.result?.checkResult)
        assertEquals("retake", feature.state.value.result?.checkView)
    }
}
