package com.luopingtech.ebike.ops.feature.workorder

import com.luopingtech.ebike.ops.data.workorder.InspectionOrderRepositoryImpl
import com.luopingtech.ebike.ops.data.workorder.RepairOrderRepositoryImpl
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.model.WorkOrderKind
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking

class WorkOrderFeatureTest {
    private val area = ServiceArea(id = "1001", name = "Demo")

    @Test
    fun inspection_listAcceptFinish() = runBlocking {
        val feature = WorkOrderFeature(
            kind = WorkOrderKind.Inspection,
            inspectionRepository = InspectionOrderRepositoryImpl(demoMode = true),
            canTakeProvider = { true },
        )
        feature.load(area)
        assertEquals(2, feature.state.value.orders.size)
        val pending = feature.state.value.orders.first { it.state == 0 }
        assertTrue(feature.accept(pending.id).isOk)
        assertTrue(feature.state.value.orders.any { it.id == pending.id && it.state == 1 })
        assertTrue(feature.finish(pending.id).isOk)
        assertFalse(feature.state.value.orders.any { it.id == pending.id })
    }

    @Test
    fun inspection_filterByAlarmType() = runBlocking {
        val feature = WorkOrderFeature(
            kind = WorkOrderKind.Inspection,
            inspectionRepository = InspectionOrderRepositoryImpl(demoMode = true),
        )
        feature.setAlarmType(3)
        feature.load(area)
        assertEquals(1, feature.state.value.orders.size)
        assertEquals(3, feature.state.value.orders.first().alarmType)
    }

    @Test
    fun repair_listAcceptFinishAndFilter() = runBlocking {
        val feature = WorkOrderFeature(
            kind = WorkOrderKind.Repair,
            repairRepository = RepairOrderRepositoryImpl(demoMode = true),
            canTakeProvider = { true },
        )
        feature.load(area)
        assertEquals(2, feature.state.value.orders.size)
        assertTrue(feature.state.value.fixNameChips.isNotEmpty())

        val brake = feature.state.value.fixNameChips.first()
        feature.setFixName(brake)
        feature.load(area)
        assertTrue(feature.state.value.orders.isNotEmpty())
        assertTrue(feature.state.value.orders.all { it.partNames.any { n -> n.contains(brake) } })

        feature.setFixName(null)
        feature.load(area)
        val pending = feature.state.value.orders.first { it.state == 0 }
        assertTrue(feature.accept(pending.id).isOk)
        assertTrue(feature.finish(pending.id).isOk)
        assertFalse(feature.state.value.orders.any { it.id == pending.id })
    }

    @Test
    fun withoutTakePerm_hidesActionsAndBlocks() = runBlocking {
        val feature = WorkOrderFeature(
            kind = WorkOrderKind.Inspection,
            inspectionRepository = InspectionOrderRepositoryImpl(demoMode = true),
            canTakeProvider = { false },
        )
        feature.load(area)
        assertFalse(feature.state.value.canTake)
        val pending = feature.state.value.orders.first { it.state == 0 }
        assertTrue(feature.accept(pending.id).isErr)
    }
}
