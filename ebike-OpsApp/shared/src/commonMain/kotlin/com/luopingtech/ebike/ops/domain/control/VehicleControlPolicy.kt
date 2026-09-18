package com.luopingtech.ebike.ops.domain.control

import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.platform.BleCommand
import com.luopingtech.ebike.ops.platform.BleTransport

enum class ControlChannel {
    /** Prefer BLE, fall back to network. */
    BlePreferred,

    /** BLE only. */
    BleOnly,

    /** Network / MQTT-style remote command only. */
    NetworkOnly,

    /** Legacy NETWORK_FIRST: try network, fall back to BLE. */
    NetworkPreferred,
}

enum class VehicleAction {
    Unlock,
    Lock,
    Ring,
    OpenBatteryBox,
    CloseBatteryBox,
    AccOn,
    AccOff,
    DefendOn,
    DefendOff,
    OpenHelmetLock,
    CloseHelmetLock,
    OpenBackWheelLock,
    CloseBackWheelLock,
    Restart,
}

/**
 * Network remote control port. Implemented by data layer.
 */
interface NetworkVehicleControl {
    suspend fun execute(
        vehicleId: String,
        action: VehicleAction,
        imei: String = "",
    ): OpsResult<Unit>

    /**
     * Legacy openBatBoxPermission before opening the battery compartment.
     * Demo / unsupported channels default to OK.
     */
    suspend fun checkOpenBatteryBoxPermission(vehicleId: String): OpsResult<Unit> =
        OpsResult.Ok(Unit)
}

/**
 * Single decision point for vehicle control. Both hosts must use this instead of
 * duplicating BLE-vs-network branching in UI code.
 */
class VehicleControlPolicy(
    private val ble: BleTransport,
    private val network: NetworkVehicleControl,
) {
    suspend fun execute(
        vehicleId: String,
        action: VehicleAction,
        channel: ControlChannel = ControlChannel.BlePreferred,
        imei: String = "",
    ): OpsResult<Unit> {
        if (action == VehicleAction.OpenBatteryBox) {
            when (val perm = network.checkOpenBatteryBoxPermission(vehicleId)) {
                is OpsResult.Err -> return perm
                is OpsResult.Ok -> Unit
            }
        }
        return when (channel) {
            ControlChannel.NetworkOnly -> network.execute(vehicleId, action, imei)
            ControlChannel.BleOnly -> executeBle(vehicleId, action)
            ControlChannel.BlePreferred -> {
                val bleResult = executeBle(vehicleId, action)
                if (bleResult.isOk) bleResult else network.execute(vehicleId, action, imei)
            }
            ControlChannel.NetworkPreferred -> {
                val netResult = network.execute(vehicleId, action, imei)
                if (netResult.isOk) netResult else executeBle(vehicleId, action)
            }
        }
    }

    private suspend fun executeBle(vehicleId: String, action: VehicleAction): OpsResult<Unit> {
        if (!ble.isAvailable) {
            return OpsResult.Err(OpsError.unsupported("BleTransport"))
        }
        val connect = ble.connect(vehicleId)
        if (connect is OpsResult.Err) return connect

        val command = when (action) {
            VehicleAction.Unlock -> BleCommand.Unlock(vehicleId)
            VehicleAction.Lock -> BleCommand.Lock(vehicleId)
            VehicleAction.Ring -> BleCommand.Ring(vehicleId)
            VehicleAction.OpenBatteryBox -> BleCommand.OpenBatteryBox(vehicleId)
            VehicleAction.CloseBatteryBox -> BleCommand.CloseBatteryBox(vehicleId)
            VehicleAction.AccOn -> BleCommand.AccOn(vehicleId)
            VehicleAction.AccOff -> BleCommand.AccOff(vehicleId)
            VehicleAction.DefendOn -> BleCommand.DefendOn(vehicleId)
            VehicleAction.DefendOff -> BleCommand.DefendOff(vehicleId)
            VehicleAction.OpenHelmetLock -> BleCommand.OpenHelmetLock(vehicleId)
            VehicleAction.CloseHelmetLock -> BleCommand.CloseHelmetLock(vehicleId)
            VehicleAction.OpenBackWheelLock -> BleCommand.OpenBackWheelLock(vehicleId)
            VehicleAction.CloseBackWheelLock -> BleCommand.CloseBackWheelLock(vehicleId)
            VehicleAction.Restart -> BleCommand.Restart(vehicleId)
        }
        val sent = ble.sendCommand(command)
        return when (sent) {
            is OpsResult.Ok -> if (sent.value.success) {
                OpsResult.Ok(Unit)
            } else {
                OpsResult.Err(OpsError.business("BLE_CMD", sent.value.message))
            }
            is OpsResult.Err -> sent
        }
    }
}
