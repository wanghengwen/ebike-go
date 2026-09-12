package com.luopingtech.ebike.ops.feature.production

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.control.DemoNetworkVehicleControl
import com.luopingtech.ebike.ops.data.production.ProductionRepositoryImpl
import com.luopingtech.ebike.ops.data.vehicle.VehicleRepositoryImpl
import com.luopingtech.ebike.ops.domain.control.VehicleControlPolicy
import com.luopingtech.ebike.ops.domain.model.DetectStepKind
import com.luopingtech.ebike.ops.domain.model.DetectStepStatus
import com.luopingtech.ebike.ops.platform.SimulatorBleTransport
import com.luopingtech.ebike.ops.platform.UnavailableBleTransport
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking

class ProductionFeatureTest {
    private fun feature(): ProductionFeature = ProductionFeature(
        repository = ProductionRepositoryImpl(demoMode = true),
        control = VehicleControlPolicy(
            ble = SimulatorBleTransport(),
            network = DemoNetworkVehicleControl(),
        ),
        serviceAreaIdProvider = { "1001" },
        vehicleRepository = VehicleRepositoryImpl(demoMode = true),
        bleAvailableProvider = { true },
    )

    @Test
    fun demo_bindAndUnbind() = runBlocking {
        val f = feature()
        f.openBind()
        f.setCarId("D1001-001")
        f.setImei("860000000000001")
        f.bindCenter()
        assertEquals(Strings.t(Str.BoundImei, "860000000000001"), f.state.value.message)
        assertEquals("860000000000001", f.state.value.bindInfo?.imei)
        f.unbindCenter()
        assertEquals(Strings.t(Str.UnbindOk), f.state.value.message)
    }

    @Test
    fun demo_putOnShelves() = runBlocking {
        val f = feature()
        f.openShelves(ShelfMode.PutOn)
        f.setCarId("100600007")
        f.addShelfCar()
        assertEquals(1, f.state.value.shelfQueue.size)
        f.submitShelfQueue()
        assertEquals(ProductionPage.Hub, f.state.value.page)
        assertEquals(
            Strings.t(Str.ShelfSubmitOk, Strings.t(Str.ShelfPutOn), 1),
            f.state.value.message,
        )
    }

    @Test
    fun demo_detectRing() = runBlocking {
        val f = feature()
        f.openDetect()
        assertEquals(DetectChannelTab.Network, f.state.value.detectChannel)
        f.setCarId("D1001-001")
        f.runDetectStep(DetectStepKind.Ring)
        val step = f.state.value.detectSteps.first { it.kind == DetectStepKind.Ring }
        assertEquals(DetectStepStatus.Ok, step.status)
        assertTrue(step.detail.contains(Strings.t(Str.DetectChannelNetwork)))
    }

    @Test
    fun demo_detectBluetoothTabUsesBleOnly() = runBlocking {
        val f = feature()
        f.openDetect()
        f.setCarId("D1001-001")
        f.setDetectChannel(DetectChannelTab.Bluetooth)
        assertEquals(DetectChannelTab.Bluetooth, f.state.value.detectChannel)
        f.runDetectStep(DetectStepKind.Unlock)
        val step = f.state.value.detectSteps.first { it.kind == DetectStepKind.Unlock }
        assertEquals(DetectStepStatus.Ok, step.status)
        assertTrue(step.detail.contains(Strings.t(Str.DetectChannelBle)))
    }

    @Test
    fun demo_detectBluetoothUnavailableFails() = runBlocking {
        val f = ProductionFeature(
            repository = ProductionRepositoryImpl(demoMode = true),
            control = VehicleControlPolicy(
                ble = UnavailableBleTransport(),
                network = DemoNetworkVehicleControl(),
            ),
            serviceAreaIdProvider = { "1001" },
            vehicleRepository = VehicleRepositoryImpl(demoMode = true),
            bleAvailableProvider = { false },
        )
        f.openDetect()
        f.setCarId("D1001-001")
        f.setDetectChannel(DetectChannelTab.Bluetooth)
        f.runDetectStep(DetectStepKind.Ring)
        val step = f.state.value.detectSteps.first { it.kind == DetectStepKind.Ring }
        assertEquals(DetectStepStatus.Failed, step.status)
        assertEquals(Strings.t(Str.DetectBleUnavailableHint), step.detail)

        f.setDetectChannel(DetectChannelTab.Network)
        f.runDetectStep(DetectStepKind.Ring)
        val net = f.state.value.detectSteps.first { it.kind == DetectStepKind.Ring }
        assertEquals(DetectStepStatus.Ok, net.status)
    }

    @Test
    fun demo_detectSwitchAccOnNetwork() = runBlocking {
        val f = feature()
        f.openDetect()
        f.setCarId("D1001-001")
        f.setImei("860000000000001")
        f.toggleDetectSwitch(DetectSwitchKind.Acc, on = true)
        val sw = f.state.value.detectSwitches.first { it.kind == DetectSwitchKind.Acc }
        assertEquals(true, sw.on)
        assertTrue(f.state.value.message!!.contains(Strings.t(Str.DetectChannelNetwork)))
    }

    @Test
    fun demo_offlineCheckRequiresServiceNamePath() = runBlocking {
        val repo = ProductionRepositoryImpl(demoMode = true)
        val result = repo.offlineCheck("100600008")
        assertTrue(result is OpsResult.Ok)
        assertTrue((result as OpsResult.Ok).value.serviceName.isNotBlank())
    }

    @Test
    fun demo_scanFillsCarId() {
        val f = feature()
        f.openDetect()
        assertTrue(f.applyScanRaw("https://ops.example/qr?carId=D1001-009", listOf("ops.example")))
        assertEquals("D1001-009", f.state.value.carId)
    }

    @Test
    fun demo_scanImeiFillsImei() {
        val f = feature()
        f.openBind()
        assertTrue(f.applyScanRaw("860000000000099"))
        assertEquals("860000000000099", f.state.value.imei)
    }

    @Test
    fun demo_locateVehicleOnMapByCarId() = runBlocking {
        val f = feature()
        f.openDetect()
        f.setCarId("D1001-003")
        val result = f.locateVehicleOnMap()
        assertTrue(result is OpsResult.Ok)
        assertEquals("D1001-003", (result as OpsResult.Ok).value.carId)
        assertEquals(Strings.t(Str.DetectLocationOk, "D1001-003"), f.state.value.message)
    }

    @Test
    fun demo_locateVehicleOnMapRequiresId() = runBlocking {
        val f = feature()
        f.openDetect()
        val result = f.locateVehicleOnMap()
        assertTrue(result is OpsResult.Err)
        assertEquals(Strings.t(Str.EnterCarIdOrScan), f.state.value.errorMessage)
    }

    @Test
    fun demo_openOverloadAndPollContacts() = runBlocking {
        val f = ProductionFeature(
            repository = ProductionRepositoryImpl(demoMode = true),
            control = VehicleControlPolicy(
                ble = SimulatorBleTransport(),
                network = DemoNetworkVehicleControl(),
            ),
            serviceAreaIdProvider = { "1001" },
            vehicleRepository = VehicleRepositoryImpl(demoMode = true),
            bleAvailableProvider = { true },
            overloadTimeoutSec = 3,
            overloadPollDelayMs = 1L,
        )
        f.openDetect()
        f.setCarId("D1001-001")
        assertTrue(f.openOverloadCheck().isOk)
        assertEquals(ProductionPage.Overload, f.state.value.page)
        assertEquals("860000000000001", f.state.value.imei)
        f.startOverloadCheck()
        assertEquals(true, f.state.value.overloadContacts.frontOn)
        assertEquals(true, f.state.value.overloadContacts.centerOn)
        assertEquals(true, f.state.value.overloadContacts.backOn)
        assertEquals(false, f.state.value.overloadRunning)
    }

    @Test
    fun demo_openOverloadWithoutIdFails() = runBlocking {
        val f = feature()
        f.openDetect()
        assertTrue(f.openOverloadCheck().isErr)
    }
}
