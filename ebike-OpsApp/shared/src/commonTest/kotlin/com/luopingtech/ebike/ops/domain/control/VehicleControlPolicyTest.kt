package com.luopingtech.ebike.ops.domain.control

import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.control.DemoNetworkVehicleControl
import com.luopingtech.ebike.ops.platform.SimulatorBleTransport
import kotlin.test.Test
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking

class VehicleControlPolicyTest {
    @Test
    fun blePreferred_usesSimulatorSuccessfully() = runBlocking {
        val ble = SimulatorBleTransport()
        val policy = VehicleControlPolicy(
            ble = ble,
            network = object : NetworkVehicleControl {
                override suspend fun execute(
                    vehicleId: String,
                    action: VehicleAction,
                    imei: String,
                ): OpsResult<Unit> = OpsResult.Err(
                    com.luopingtech.ebike.ops.core.result.OpsError.business("NET", "should not call"),
                )
            },
        )
        val result = policy.execute(
            vehicleId = "SIM-VEHICLE-001",
            action = VehicleAction.Ring,
            channel = ControlChannel.BlePreferred,
        )
        assertTrue(result.isOk)
    }

    @Test
    fun networkOnly_accOn_demoOk() = runBlocking {
        val policy = VehicleControlPolicy(
            ble = SimulatorBleTransport(),
            network = DemoNetworkVehicleControl(),
        )
        val result = policy.execute(
            vehicleId = "D1001-001",
            action = VehicleAction.AccOn,
            channel = ControlChannel.NetworkOnly,
            imei = "860000000000001",
        )
        assertTrue(result.isOk)
    }

    @Test
    fun networkPreferred_fallsBackToBleWhenNetworkFails() = runBlocking {
        val ble = SimulatorBleTransport()
        ble.connect("SIM-VEHICLE-001")
        val policy = VehicleControlPolicy(
            ble = ble,
            network = object : NetworkVehicleControl {
                override suspend fun execute(
                    vehicleId: String,
                    action: VehicleAction,
                    imei: String,
                ): OpsResult<Unit> = OpsResult.Err(
                    com.luopingtech.ebike.ops.core.result.OpsError.business("NET", "down"),
                )
            },
        )
        val result = policy.execute(
            vehicleId = "SIM-VEHICLE-001",
            action = VehicleAction.Restart,
            channel = ControlChannel.NetworkPreferred,
        )
        assertTrue(result.isOk)
    }
}
