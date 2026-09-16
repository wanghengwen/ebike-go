package com.luopingtech.ebike.rider.platform

import com.luopingtech.ebike.rider.core.config.TenantConfig
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.core.i18n.Strings
import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.domain.ble.BleCommands
import com.luopingtech.ebike.rider.domain.ble.BleConvert
import com.luopingtech.ebike.rider.domain.ble.BleFrame
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow

/**
 * C-end BLE port. Scan-by-IMEI, GATT write, one notify. No Ops maintenance cmds.
 */
interface BleTransport {
    val isAvailable: Boolean

    fun connectionState(): Flow<BleConnectionState>

    suspend fun openAdapter(): RiderResult<Unit>

    suspend fun closeAdapter(): RiderResult<Unit>

    suspend fun findByImei(
        imei: String,
        timeoutMs: Long = BleCommands.SEARCH_TIMEOUT_MS,
    ): RiderResult<BleAdvertDevice>

    suspend fun connect(
        deviceId: String,
        timeoutMs: Long = BleCommands.CONNECT_TIMEOUT_MS,
    ): RiderResult<Unit>

    suspend fun disconnect(): RiderResult<Unit>

    suspend fun writeAndAwait(
        hex: String,
        timeoutMs: Long = BleCommands.SEND_TIMEOUT_MS,
    ): RiderResult<BleNotifyFrame>
}

enum class BleConnectionState {
    Disconnected,
    Connecting,
    Connected,
    Failed,
}

data class BleAdvertDevice(
    val id: String,
    val name: String? = null,
    val rssi: Int? = null,
    val advertHex: String = "",
)

data class BleNotifyFrame(
    val hex: String,
)

object BleTransportFactory {
    fun fromConfig(
        config: TenantConfig,
        override: BleTransport? = null,
        native: BleTransport? = null,
    ): BleTransport {
        if (override != null) return override
        return when (config.features.bleTransport.lowercase()) {
            "native", "real", "hardware" -> native ?: UnavailableBleTransport()
            "simulator", "sim", "" -> SimulatorBleTransport()
            else -> SimulatorBleTransport()
        }
    }
}

class UnavailableBleTransport : BleTransport {
    override val isAvailable: Boolean = false
    private val state = MutableStateFlow(BleConnectionState.Disconnected)

    override fun connectionState(): Flow<BleConnectionState> = state.asStateFlow()

    override suspend fun openAdapter(): RiderResult<Unit> =
        RiderResult.Err(RiderError.unsupported(Strings.t(Str.BleUnavailable)))

    override suspend fun closeAdapter(): RiderResult<Unit> {
        state.value = BleConnectionState.Disconnected
        return RiderResult.Ok(Unit)
    }

    override suspend fun findByImei(imei: String, timeoutMs: Long): RiderResult<BleAdvertDevice> =
        RiderResult.Err(RiderError.unsupported(Strings.t(Str.BleUnavailable)))

    override suspend fun connect(deviceId: String, timeoutMs: Long): RiderResult<Unit> =
        RiderResult.Err(RiderError.unsupported(Strings.t(Str.BleUnavailable)))

    override suspend fun disconnect(): RiderResult<Unit> {
        state.value = BleConnectionState.Disconnected
        return RiderResult.Ok(Unit)
    }

    override suspend fun writeAndAwait(hex: String, timeoutMs: Long): RiderResult<BleNotifyFrame> =
        RiderResult.Err(RiderError.unsupported(Strings.t(Str.BleUnavailable)))
}

/**
 * In-process stand-in. [findByImei] always hits; [writeAndAwait] returns a
 * 1-byte `code==0` ACK so session success matches UniApp (`code === 0`).
 */
class SimulatorBleTransport : BleTransport {
    override val isAvailable: Boolean = true

    private val state = MutableStateFlow(BleConnectionState.Disconnected)

    override fun connectionState(): Flow<BleConnectionState> = state.asStateFlow()

    override suspend fun openAdapter(): RiderResult<Unit> = RiderResult.Ok(Unit)

    override suspend fun closeAdapter(): RiderResult<Unit> {
        state.value = BleConnectionState.Disconnected
        return RiderResult.Ok(Unit)
    }

    override suspend fun findByImei(imei: String, timeoutMs: Long): RiderResult<BleAdvertDevice> {
        if (imei.isBlank()) {
            return RiderResult.Err(RiderError.business("BLE_NOT_FOUND", Strings.t(Str.BleNotFound)))
        }
        val advert = if (imei.length >= 15) imei.substring(3, 15) else imei.padEnd(12, '0').take(12)
        return RiderResult.Ok(
            BleAdvertDevice(
                id = "SIM-$imei",
                name = "Simulator Bike",
                rssi = -55,
                advertHex = advert,
            ),
        )
    }

    override suspend fun connect(deviceId: String, timeoutMs: Long): RiderResult<Unit> {
        if (deviceId.isBlank()) {
            state.value = BleConnectionState.Failed
            return RiderResult.Err(RiderError.business("BLE_NOT_FOUND", Strings.t(Str.BleNotFound)))
        }
        state.value = BleConnectionState.Connecting
        state.value = BleConnectionState.Connected
        return RiderResult.Ok(Unit)
    }

    override suspend fun disconnect(): RiderResult<Unit> {
        state.value = BleConnectionState.Disconnected
        return RiderResult.Ok(Unit)
    }

    override suspend fun writeAndAwait(hex: String, timeoutMs: Long): RiderResult<BleNotifyFrame> {
        if (state.value != BleConnectionState.Connected) {
            return RiderResult.Err(RiderError.business("BLE_NOT_CONNECTED", Strings.t(Str.BleUnavailable)))
        }
        val cmd = hex.take(2).toIntOrNull(16) ?: 0
        val crc = (cmd + 1) and 0xff
        val ack = BleConvert.toPad2HexStr(cmd) + "01" + "00" + BleConvert.toPad2HexStr(crc)
        return RiderResult.Ok(BleNotifyFrame(ack))
    }
}

/** Android `BluetoothGatt` / iOS `CoreBluetooth`. Hosts construct the actual. */
expect class NativeBleTransport : BleTransport {
    override val isAvailable: Boolean
    override fun connectionState(): Flow<BleConnectionState>
    override suspend fun openAdapter(): RiderResult<Unit>
    override suspend fun closeAdapter(): RiderResult<Unit>
    override suspend fun findByImei(imei: String, timeoutMs: Long): RiderResult<BleAdvertDevice>
    override suspend fun connect(deviceId: String, timeoutMs: Long): RiderResult<Unit>
    override suspend fun disconnect(): RiderResult<Unit>
    override suspend fun writeAndAwait(hex: String, timeoutMs: Long): RiderResult<BleNotifyFrame>
}
