package com.luopingtech.ebike.ops.feature.task

import com.luopingtech.ebike.ops.data.control.DemoNetworkVehicleControl
import com.luopingtech.ebike.ops.data.task.MoveCarTaskRepositoryImpl
import com.luopingtech.ebike.ops.domain.control.VehicleControlPolicy
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.platform.GeoPoint
import com.luopingtech.ebike.ops.platform.SimulatorBleTransport
import com.luopingtech.ebike.ops.platform.SimulatorLocationTracker
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking

class MoveCarTaskFeatureTest {
    private fun feature(
        origin: GeoPoint = GeoPoint(28.22, 112.94),
    ): MoveCarTaskFeature = MoveCarTaskFeature(
        repository = MoveCarTaskRepositoryImpl(demoMode = true),
        control = VehicleControlPolicy(
            ble = SimulatorBleTransport(),
            network = DemoNetworkVehicleControl(),
        ),
        pinProvider = { "demo-pin" },
        locationTracker = SimulatorLocationTracker(origin = origin),
    )

    @Test
    fun loadClaimStartFinishWithPhoto() = runBlocking {
        val feature = feature()
        val area = ServiceArea(id = "1001", name = "Demo", centerLat = 28.22, centerLng = 112.94)
        feature.load(area)
        assertEquals(4, feature.state.value.tasks.size)
        assertTrue(feature.state.value.tasks.any { it.isManMadeBatch })

        val pending = feature.state.value.tasks.first { it.state == 0 && !it.isManMadeBatch }
        feature.selectTask(pending.id)

        assertTrue(feature.claimSelected().isOk)
        assertEquals(1, feature.state.value.selected?.state)

        assertTrue(feature.startSelected().isOk)

        // First finish without photo → 23326
        assertTrue(feature.finishSelected().isErr)
        assertTrue(feature.state.value.needPhotograph)

        feature.addDemoPhoto()
        assertTrue(feature.finishSelected().isErr)
        feature.setRemark("parked at station A")
        assertTrue(feature.finishSelected().isOk)
        assertEquals(2, feature.state.value.selected?.state)
        assertTrue(!feature.state.value.needPhotograph)
    }

    @Test
    fun arrivalFarAway_doesNotBlockFinish() = runBlocking {
        val feature = feature(origin = GeoPoint(30.0, 104.0))
        val area = ServiceArea(id = "1001", name = "Demo", centerLat = 28.22, centerLng = 112.94)
        feature.load(area)
        feature.selectTask(feature.state.value.tasks.first { it.state == 0 }.id)
        assertTrue(!feature.refreshArrival())
        assertTrue(feature.claimSelected().isOk)
        assertTrue(feature.startSelected().isOk)
        // Legacy: no 200m gate — far GPS still allows finish (photo/remark may still apply).
        assertTrue(feature.finishSelected().isErr)
        assertTrue(feature.state.value.needPhotograph)
        feature.addDemoPhoto()
        feature.setRemark("far but allowed")
        assertTrue(feature.finishSelected().isOk)
    }

    @Test
    fun claimWithoutSelection_fails() = runBlocking {
        assertTrue(feature().claimSelected().isErr)
    }
}
