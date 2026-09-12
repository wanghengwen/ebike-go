package com.luopingtech.ebike.ops.platform

import com.luopingtech.ebike.ops.core.config.TenantConfig
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow

/**
 * Private BLE SDK adapter. Real vendors plug in later; scaffold ships [SimulatorBleTransport].
 */
interface BleTransport {
    val isAvailable: Boolean

    suspend fun scan(timeoutMs: Long = 8_000): OpsResult<List<BleDevice>>

    suspend fun connect(deviceId: String): OpsResult<Unit>

    suspend fun disconnect(): OpsResult<Unit>

    suspend fun sendCommand(command: BleCommand): OpsResult<BleCommandResult>

    fun connectionState(): Flow<BleConnectionState>
}

/**
 * Resolves BLE from tenant config. `native` without a vendor SDK → [UnavailableBleTransport].
 */
object BleTransportFactory {
    fun fromConfig(config: TenantConfig, override: BleTransport? = null): BleTransport {
        if (override != null) return override
        return when (config.features.bleTransport.lowercase()) {
            "simulator", "sim", "" -> SimulatorBleTransport()
            "native", "real", "hardware" -> UnavailableBleTransport()
            else -> SimulatorBleTransport()
        }
    }
}

/**
 * Explicit “no private SDK” transport. Network channel still works via [VehicleControlPolicy].
 */
class UnavailableBleTransport : BleTransport {
    override val isAvailable: Boolean = false

    private val state = MutableStateFlow(BleConnectionState.Disconnected)

    override suspend fun scan(timeoutMs: Long): OpsResult<List<BleDevice>> =
        OpsResult.Err(OpsError.unsupported(Strings.t(Str.BleUnavailable)))

    override suspend fun connect(deviceId: String): OpsResult<Unit> =
        OpsResult.Err(OpsError.unsupported(Strings.t(Str.BleUnavailable)))

    override suspend fun disconnect(): OpsResult<Unit> {
        state.value = BleConnectionState.Disconnected
        return OpsResult.Ok(Unit)
    }

    override suspend fun sendCommand(command: BleCommand): OpsResult<BleCommandResult> =
        OpsResult.Err(OpsError.unsupported(Strings.t(Str.BleUnavailable)))

    override fun connectionState(): Flow<BleConnectionState> = state.asStateFlow()
}

data class BleDevice(
    val id: String,
    val name: String? = null,
    val rssi: Int? = null,
)

enum class BleConnectionState {
    Disconnected,
    Connecting,
    Connected,
    Failed,
}

sealed class BleCommand {
    data class Unlock(val vehicleId: String) : BleCommand()
    data class Lock(val vehicleId: String) : BleCommand()
    data class Ring(val vehicleId: String) : BleCommand()
    data class OpenBatteryBox(val vehicleId: String) : BleCommand()
    data class CloseBatteryBox(val vehicleId: String) : BleCommand()
    data class AccOn(val vehicleId: String) : BleCommand()
    data class AccOff(val vehicleId: String) : BleCommand()
    data class DefendOn(val vehicleId: String) : BleCommand()
    data class DefendOff(val vehicleId: String) : BleCommand()
    data class OpenHelmetLock(val vehicleId: String) : BleCommand()
    data class CloseHelmetLock(val vehicleId: String) : BleCommand()
    data class OpenBackWheelLock(val vehicleId: String) : BleCommand()
    data class CloseBackWheelLock(val vehicleId: String) : BleCommand()
    data class Raw(val payload: ByteArray) : BleCommand() {
        override fun equals(other: Any?): Boolean =
            other is Raw && payload.contentEquals(other.payload)

        override fun hashCode(): Int = payload.contentHashCode()
    }
}

data class BleCommandResult(
    val success: Boolean,
    val message: String = "",
)

/**
 * Default stub for open-source builds. Does not talk to real hardware.
 */
class SimulatorBleTransport : BleTransport {
    override val isAvailable: Boolean = true

    private val devices = listOf(
        BleDevice(id = "SIM-VEHICLE-001", name = "Simulator Bike", rssi = -55),
        BleDevice(id = "SIM-VEHICLE-002", name = "Simulator Bike 2", rssi = -70),
    )

    private val state = kotlinx.coroutines.flow.MutableStateFlow(BleConnectionState.Disconnected)

    override suspend fun scan(timeoutMs: Long): OpsResult<List<BleDevice>> =
        OpsResult.Ok(devices)

    override suspend fun connect(deviceId: String): OpsResult<Unit> {
        if (deviceId.isBlank()) {
            state.value = BleConnectionState.Failed
            return OpsResult.Err(OpsError.business("BLE_NOT_FOUND", "device id empty"))
        }
        // Scaffold simulator accepts any carId; real SDK will validate advertising ids.
        state.value = BleConnectionState.Connecting
        state.value = BleConnectionState.Connected
        return OpsResult.Ok(Unit)
    }

    override suspend fun disconnect(): OpsResult<Unit> {
        state.value = BleConnectionState.Disconnected
        return OpsResult.Ok(Unit)
    }

    override suspend fun sendCommand(command: BleCommand): OpsResult<BleCommandResult> {
        if (state.value != BleConnectionState.Connected) {
            return OpsResult.Err(OpsError.business("BLE_NOT_CONNECTED", "not connected"))
        }
        return OpsResult.Ok(
            BleCommandResult(
                success = true,
                message = "simulated ${command::class.simpleName}",
            ),
        )
    }

    override fun connectionState(): Flow<BleConnectionState> = state
}
