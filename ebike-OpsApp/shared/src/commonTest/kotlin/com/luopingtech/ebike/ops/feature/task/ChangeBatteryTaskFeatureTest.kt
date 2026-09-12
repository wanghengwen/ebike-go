package com.luopingtech.ebike.ops.feature.task

import com.luopingtech.ebike.ops.data.control.DemoNetworkVehicleControl
import com.luopingtech.ebike.ops.data.task.ChangeBatteryTaskRepositoryImpl
import com.luopingtech.ebike.ops.domain.control.VehicleControlPolicy
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.platform.SimulatorBleTransport
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking

class ChangeBatteryTaskFeatureTest {
    private fun feature(): ChangeBatteryTaskFeature = ChangeBatteryTaskFeature(
        repository = ChangeBatteryTaskRepositoryImpl(demoMode = true),
        control = VehicleControlPolicy(
            ble = SimulatorBleTransport(),
            network = DemoNetworkVehicleControl(),
        ),
        pinProvider = { "demo-pin" },
    )

    @Test
    fun loadFiltersByMaxBattery_thenOpenAndClose() = runBlocking {
        val feature = feature()
        val area = ServiceArea(id = "1001", name = "Demo", centerLat = 28.22, centerLng = 112.94)

        feature.load(area, maxBattery = 30)
        assertTrue(feature.state.value.tasks.isNotEmpty())
        assertTrue(feature.state.value.tasks.all { it.restBattery <= 30 })
        // demo catalog: 18% and 9% pass; 42% filtered out
        assertEquals(2, feature.state.value.tasks.size)

        val firstId = feature.state.value.tasks.first().id
        feature.selectTask(firstId)

        assertTrue(feature.openBox().isOk)
        assertEquals(1, feature.state.value.selected?.state)

        assertTrue(feature.closeBox().isOk)
        assertEquals(2, feature.state.value.selected?.state)
    }

    @Test
    fun load_fetchesDemoBatteryRange() = runBlocking {
        val feature = feature()
        val area = ServiceArea(id = "1001", name = "Demo")
        feature.load(area)
        assertEquals(5, feature.state.value.rangeMin)
        assertEquals(30, feature.state.value.rangeMax)
        assertEquals(30, feature.state.value.maxBattery)
    }

    @Test
    fun openWithoutSelection_fails() = runBlocking {
        val feature = feature()
        assertTrue(feature.openBox().isErr)
    }
}
