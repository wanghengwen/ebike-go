package com.luopingtech.ebike.ops.feature.production

import com.luopingtech.ebike.ops.core.i18n.OpsI18n
import com.luopingtech.ebike.ops.core.i18n.OpsLanguage
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.control.DemoNetworkVehicleControl
import com.luopingtech.ebike.ops.data.production.ProductionRepositoryImpl
import com.luopingtech.ebike.ops.data.vehicle.VehicleRepositoryImpl
import com.luopingtech.ebike.ops.domain.control.VehicleControlPolicy
import com.luopingtech.ebike.ops.domain.model.DetectStepKind
import com.luopingtech.ebike.ops.domain.model.DetectStepStatus
import com.luopingtech.ebike.ops.domain.scan.ScanTarget
import com.luopingtech.ebike.ops.platform.SimulatorBleTransport
import com.luopingtech.ebike.ops.platform.UnavailableBleTransport
import kotlin.test.BeforeTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNull
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking

class ProductionFeatureTest {
    @BeforeTest
    fun installStrings() {
        Strings.install(OpsI18n.fallback(OpsLanguage.ZH_CN))
    }
    private fun feature(
        bleAvailable: Boolean = true,
        refreshAfterControlDelayMs: Long = -1L,
    ): ProductionFeature = ProductionFeature(
        repository = ProductionRepositoryImpl(demoMode = true),
        control = VehicleControlPolicy(
            ble = if (bleAvailable) SimulatorBleTransport() else UnavailableBleTransport(),
            network = DemoNetworkVehicleControl(),
        ),
        serviceAreaIdProvider = { "1001" },
        vehicleRepository = VehicleRepositoryImpl(demoMode = true),
        bleAvailableProvider = { bleAvailable },
        refreshAfterControlDelayMs = refreshAfterControlDelayMs,
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
        assertEquals(1, f.state.value.putShelfQueue.size)
        f.setShelfItemSelected("100600007", true)
        f.submitShelfQueue(serviceId = "1001")
        assertEquals(ProductionPage.Shelves, f.state.value.page)
        assertEquals(0, f.state.value.putShelfQueue.size)
        assertEquals(
            Strings.t(Str.ShelfSubmitOk, Strings.t(Str.ShelfPutOn), 1),
            f.state.value.message,
        )
    }

    @Test
    fun demo_queryDetectAppliesSwitchPolarity() = runBlocking {
        val f = feature()
        f.openDetect()
        f.setSearchInput("D1001-001")
        f.queryDetectVehicle()
        val state = f.state.value
        assertEquals("D1001-001", state.detectVehicle?.carId)
        assertEquals("860000000000001", state.imei)
        assertEquals(true, state.hasOverloadDevice)
        assertEquals(-71, state.detectVehicle?.gsmSignal)
        assertEquals(false, state.detectSwitches.first { it.kind == DetectSwitchKind.Acc }.on)
        assertEquals(true, state.detectSwitches.first { it.kind == DetectSwitchKind.Defend }.on)
        assertEquals(false, state.detectSwitches.first { it.kind == DetectSwitchKind.Battery }.on)
        assertEquals(false, state.detectSwitches.first { it.kind == DetectSwitchKind.Helmet }.on)
    }

    @Test
    fun demo_queryDetectRejectsInvalidInput() = runBlocking {
        val f = feature()
        f.openDetect()
        f.setSearchInput("not-a-valid-code-at-all")
        f.queryDetectVehicle()
        assertEquals(Strings.t(Str.DetectInvalidCarOrImei), f.state.value.errorMessage)
        assertNull(f.state.value.detectVehicle)
    }

    @Test
    fun demo_resolveDetectInputCarIdAndImei() {
        val f = feature()
        assertEquals(ScanTarget.CarId("D1001-001"), f.resolveDetectInput("D1001-001"))
        assertEquals(ScanTarget.Imei("860000000000001"), f.resolveDetectInput("860000000000001"))
        assertNull(f.resolveDetectInput("too-long-to-be-car-or-imei-token"))
    }

    @Test
    fun demo_detectRingUsesNetworkFirst() = runBlocking {
        val f = feature()
        f.openDetect()
        f.setSearchInput("D1001-001")
        f.queryDetectVehicle()
        f.runDetectStep(DetectStepKind.Ring)
        val step = f.state.value.detectSteps.first { it.kind == DetectStepKind.Ring }
        assertEquals(DetectStepStatus.Ok, step.status)
        assertEquals(Strings.t(Str.DetectPlayFindSound), f.state.value.message)
    }

    @Test
    fun demo_detectReboot() = runBlocking {
        val f = feature()
        f.openDetect()
        f.setSearchInput("D1001-001")
        f.queryDetectVehicle()
        f.runDetectStep(DetectStepKind.Reboot)
        val step = f.state.value.detectSteps.first { it.kind == DetectStepKind.Reboot }
        assertEquals(DetectStepStatus.Ok, step.status)
        assertEquals(Strings.t(Str.DetectRebootOk), f.state.value.message)
    }

    @Test
    fun demo_detectBluetoothTabSwitchUsesBleOnly() = runBlocking {
        val f = feature()
        f.openDetect()
        f.setSearchInput("D1001-001")
        f.queryDetectVehicle()
        f.setDetectChannel(DetectChannelTab.Bluetooth)
        assertEquals(DetectChannelTab.Bluetooth, f.state.value.detectChannel)
        f.toggleDetectSwitch(DetectSwitchKind.Acc, on = true)
        val sw = f.state.value.detectSwitches.first { it.kind == DetectSwitchKind.Acc }
        assertEquals(true, sw.on)
        assertTrue(f.state.value.message!!.contains(Strings.t(Str.DetectChannelBle)))
    }

    @Test
    fun demo_detectBluetoothUnavailableSwitchFails_ringStillWorks() = runBlocking {
        val f = feature(bleAvailable = false)
        f.openDetect()
        f.setSearchInput("D1001-001")
        f.queryDetectVehicle()
        f.setDetectChannel(DetectChannelTab.Bluetooth)
        f.toggleDetectSwitch(DetectSwitchKind.Acc, on = true)
        assertEquals(Strings.t(Str.DetectBleUnavailableHint), f.state.value.errorMessage)

        f.runDetectStep(DetectStepKind.Ring)
        val step = f.state.value.detectSteps.first { it.kind == DetectStepKind.Ring }
        assertEquals(DetectStepStatus.Ok, step.status)
    }

    @Test
    fun demo_detectSwitchAccOnNetwork() = runBlocking {
        val f = feature()
        f.openDetect()
        f.setSearchInput("D1001-001")
        f.queryDetectVehicle()
        f.toggleDetectSwitch(DetectSwitchKind.Acc, on = true)
        val sw = f.state.value.detectSwitches.first { it.kind == DetectSwitchKind.Acc }
        assertEquals(true, sw.on)
        assertTrue(f.state.value.message!!.contains(Strings.t(Str.DetectChannelNetwork)))
    }

    @Test
    fun demo_switchingTabKeepsVehicleSwitches() = runBlocking {
        val f = feature()
        f.openDetect()
        f.setSearchInput("D1001-001")
        f.queryDetectVehicle()
        f.setDetectChannel(DetectChannelTab.Bluetooth)
        assertEquals(true, f.state.value.detectSwitches.first { it.kind == DetectSwitchKind.Defend }.on)
        assertEquals("D1001-001", f.state.value.detectVehicle?.carId)
    }

    @Test
    fun demo_footerRequiresVehicle() = runBlocking {
        val f = feature()
        f.openDetect()
        f.runDetectStep(DetectStepKind.Ring)
        assertEquals(Strings.t(Str.DetectNoVehicleInfo), f.state.value.errorMessage)
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
    fun bind_scanRoutesImeiMacAndCarId() = runBlocking {
        val f = feature()
        f.openBind()
        assertTrue(f.applyBindScan("IMEI:860000000000001 SN:BB0631490221"))
        assertEquals("860000000000001", f.state.value.imei)

        assertTrue(f.applyBindScan("MAC:AA:BB:CC:DD:EE:FF SN:HELMET01"))
        assertEquals("AABBCCDDEEFF", f.state.value.helmet)

        assertTrue(f.applyBindScan("https://ops.example/qr?carId=100600007", listOf("ops.example")))
        assertEquals("100600007", f.state.value.carId)
        assertTrue(f.canBindOrUnbind())
    }

    @Test
    fun bind_requiresCarAndDeviceOrHelmet() = runBlocking {
        val f = feature()
        f.openBind()
        f.setCarId("100600007")
        f.bindCenter()
        assertEquals(Strings.t(Str.BindNeedDeviceOrHelmet), f.state.value.errorMessage)
        f.setImei("860000000000001")
        assertTrue(f.canBindOrUnbind())
        f.bindCenter()
        assertEquals(Strings.t(Str.BoundImei, "860000000000001"), f.state.value.message)
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
            refreshAfterControlDelayMs = -1L,
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
