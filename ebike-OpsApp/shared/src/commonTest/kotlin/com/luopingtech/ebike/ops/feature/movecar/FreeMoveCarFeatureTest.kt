package com.luopingtech.ebike.ops.feature.movecar

import com.luopingtech.ebike.ops.data.movecar.FreeMoveCarRepositoryImpl
import com.luopingtech.ebike.ops.data.staff.ServiceUserRepositoryImpl
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking

class FreeMoveCarFeatureTest {
    private val area = ServiceArea(id = "1001", name = "Demo", centerLat = 28.22, centerLng = 112.94)

    private fun feature() = FreeMoveCarFeature(
        repository = FreeMoveCarRepositoryImpl(demoMode = true),
        serviceUserRepository = ServiceUserRepositoryImpl(demoMode = true),
    )

    @Test
    fun addFinishWithPhoto() = runBlocking {
        val feature = feature()
        feature.load(area)
        assertTrue(feature.state.value.cars.isNotEmpty())

        assertTrue(feature.addByCarId("D1001-099", area))
        assertTrue(feature.state.value.cars.any { it.carId == "D1001-099" })
        // addByCarId 已勾选；确保只完成这一台
        feature.clearSelection()
        feature.toggleSelect("D1001-099")

        assertTrue(feature.finishSelected().isErr)
        assertTrue(feature.state.value.needPhotograph)
        assertTrue(feature.state.value.teamCandidates.isNotEmpty())

        feature.addDemoPhoto()
        // Remark optional for free-move photograph audit (legacy PHOTOGRAPH_MOVE_CAR).
        val first = feature.state.value.teamCandidates.first()
        feature.toggleTeamWorker(first.selectionKey)
        assertEquals(1, feature.state.value.selectedTeamWorkers.size)
        assertTrue(feature.finishSelected().isOk)
        assertTrue(feature.state.value.cars.none { it.carId == "D1001-099" })
        assertTrue(feature.state.value.selectedTeamKeys.isEmpty())
    }

    @Test
    fun scanAddAndRemove() = runBlocking {
        val feature = feature()
        feature.load(area)
        assertTrue(
            feature.addByScanRaw(
                "https://ops.example/qr?carId=SCAN-1",
                area,
                listOf("ops.example"),
            ),
        )
        assertEquals(true, feature.state.value.cars.any { it.carId == "SCAN-1" })
        feature.removeCar("SCAN-1")
        assertTrue(feature.state.value.cars.none { it.carId == "SCAN-1" })
    }
}
