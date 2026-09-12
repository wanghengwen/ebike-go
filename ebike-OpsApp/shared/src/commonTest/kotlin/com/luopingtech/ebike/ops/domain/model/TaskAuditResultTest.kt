package com.luopingtech.ebike.ops.domain.model

import kotlin.test.Test
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class TaskAuditResultTest {
    @Test
    fun statusFlags() {
        assertTrue(TaskAuditResult(checkResult = 4).isRejected)
        assertTrue(TaskAuditResult(checkResult = 1).isRejected)
        assertTrue(TaskAuditResult(checkResult = 2).isPassed)
        assertTrue(TaskAuditResult(checkResult = 3).isPassed)
        assertFalse(TaskAuditResult(checkResult = 0).isRejected)
    }

    @Test
    fun canViewAuditResultRules() {
        assertTrue(
            OpsTask(
                id = "1",
                type = OpsTaskType.Inspection,
                state = 2,
                checkResult = 4,
            ).canViewAuditResult(),
        )
        assertFalse(
            OpsTask(
                id = "2",
                type = OpsTaskType.Inspection,
                state = 2,
                checkResult = 2,
            ).canViewAuditResult(),
        )
        assertFalse(
            OpsTask(
                id = "3",
                type = OpsTaskType.ChangeBattery,
                state = 2,
                checkResult = 4,
            ).canViewAuditResult(),
        )
        assertTrue(
            OpsTask(
                id = "4",
                type = OpsTaskType.MoveCar,
                state = 2,
                checkResult = null,
            ).canViewAuditResult(),
        )
    }
}
