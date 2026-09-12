package com.luopingtech.ebike.ops.feature.scan

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.data.control.DemoNetworkVehicleControl
import com.luopingtech.ebike.ops.data.vehicle.VehicleRepositoryImpl
import com.luopingtech.ebike.ops.domain.control.VehicleControlPolicy
import com.luopingtech.ebike.ops.platform.SimulatorBleTransport
import com.luopingtech.ebike.ops.platform.UnsupportedCodeScanner
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking

class ScanFeatureTest {
    private fun feature(): ScanFeature = ScanFeature(
        repository = VehicleRepositoryImpl(demoMode = true),
        control = VehicleControlPolicy(
            ble = SimulatorBleTransport(),
            network = DemoNetworkVehicleControl(),
        ),
        codeScanner = UnsupportedCodeScanner(),
    )

    @Test
    fun resolveDemoCarId_thenUnlock() = runBlocking {
        val scan = feature()
        val resolved = scan.resolveManual("D1001-001")
        assertTrue(resolved.isOk)
        assertEquals("D1001-001", resolved.getOrNull()?.carId)

        val unlocked = scan.unlock()
        assertTrue(unlocked.isOk)
        assertEquals(
            Strings.t(Str.ActionOk, Strings.t(Str.ScanUnlock), "D1001-001"),
            scan.state.value.message,
        )
    }

    @Test
    fun resolveImei() = runBlocking {
        val scan = feature()
        val resolved = scan.resolveManual("IMEI:860000000000002")
        assertTrue(resolved.isOk)
        assertEquals("D1001-002", resolved.getOrNull()?.carId)
    }

    @Test
    fun unlockWithoutResolve_fails() = runBlocking {
        val scan = feature()
        val result = scan.unlock()
        assertTrue(result.isErr)
    }
}
