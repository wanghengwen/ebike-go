package com.luopingtech.ebike.ops.feature.movecar

import com.luopingtech.ebike.ops.data.movecar.BatchMoveCarRepositoryImpl
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking

class BatchMoveCarFeatureTest {
    @Test
    fun loadStartFinishWithPhoto() = runBlocking {
        val parentId = "mc-1001-batch"
        val feature = BatchMoveCarFeature(BatchMoveCarRepositoryImpl(demoMode = true))
        feature.load(parentId)
        assertEquals(3, feature.state.value.children.size)

        feature.selectPending()
        assertTrue(feature.state.value.selectedTaskIds.isNotEmpty())

        assertTrue(feature.startSelected().isOk)
        assertTrue(feature.state.value.children.any { it.state == 1 })

        feature.selectPending()
        assertTrue(feature.finishSelected().isErr)
        assertTrue(feature.state.value.needPhotograph)

        feature.addDemoPhoto()
        assertTrue(feature.finishSelected().isErr)
        feature.setRemark("batch parked")
        assertTrue(feature.finishSelected().isOk)
        assertTrue(feature.state.value.children.all { it.isFinished })
    }
}
