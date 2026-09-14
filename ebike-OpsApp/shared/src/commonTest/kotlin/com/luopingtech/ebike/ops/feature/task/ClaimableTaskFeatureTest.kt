package com.luopingtech.ebike.ops.feature.task

import com.luopingtech.ebike.ops.data.task.InspectionTaskRepositoryImpl
import com.luopingtech.ebike.ops.data.task.RepairTaskRepositoryImpl
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking

class ClaimableTaskFeatureTest {
    private val area = ServiceArea(id = "1001", name = "Demo")

    @Test
    fun inspection_claimStartFinishWithPhoto() = runBlocking {
        val feature = InspectionTaskFeature(
            repository = InspectionTaskRepositoryImpl(demoMode = true),
            pinProvider = { "demo-pin" },
        )
        feature.load(area)
        assertEquals(3, feature.state.value.tasks.size)
        val pending = feature.state.value.tasks.first { it.state == 0 }
        feature.selectTask(pending.id)
        assertTrue(feature.claimSelected().isOk)
        assertTrue(feature.startSelected().isOk)
        assertTrue(feature.finishSelected().isErr)
        assertTrue(feature.state.value.needPhotograph)
        feature.addDemoPhoto()
        assertTrue(feature.finishSelected().isErr)
        feature.setRemark("inspection ok")
        assertTrue(feature.finishSelected().isOk)
        assertEquals(2, feature.state.value.selected?.state)
    }

    @Test
    fun repair_claimFinishWithPhoto() = runBlocking {
        val feature = RepairTaskFeature(
            repository = RepairTaskRepositoryImpl(demoMode = true),
            pinProvider = { "demo-pin" },
        )
        feature.load(area)
        assertEquals(4, feature.state.value.tasks.size)
        feature.selectTask(feature.state.value.tasks.first { it.state == 0 }.id)
        assertTrue(feature.claimSelected().isOk)
        assertTrue(feature.finishSelected().isErr)
        feature.addDemoPhoto()
        assertTrue(feature.finishSelected().isErr)
        feature.setRemark("repaired")
        assertTrue(feature.finishSelected().isOk)
    }

    @Test
    fun repair_createDragBlocksFinishUntilCleared() = runBlocking {
        val feature = RepairTaskFeature(
            repository = RepairTaskRepositoryImpl(demoMode = true),
            pinProvider = { "demo-pin" },
        )
        feature.load(area)
        feature.selectTask(feature.state.value.tasks.first { it.state == 0 }.id)
        assertTrue(feature.createDragSelected().isErr)
        feature.setDragReason("broken brake")
        assertTrue(feature.createDragSelected().isErr)
        feature.setDragAddress("warehouse A")
        assertTrue(feature.createDragSelected().isOk)
        assertEquals(2, feature.state.value.selected?.dragState)
        assertTrue(feature.finishSelected().isErr)
        assertTrue(feature.createDragSelected().isErr)
    }
}
