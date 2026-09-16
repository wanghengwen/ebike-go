package com.luopingtech.ebike.rider.feature.ble

import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.data.ble.DemoBleTokenApi
import com.luopingtech.ebike.rider.domain.ble.BleCommands
import com.luopingtech.ebike.rider.domain.ble.BleConvert
import com.luopingtech.ebike.rider.domain.ble.BleFrame
import com.luopingtech.ebike.rider.platform.BleAdvertDevice
import com.luopingtech.ebike.rider.platform.BleConnectionState
import com.luopingtech.ebike.rider.platform.BleNotifyFrame
import com.luopingtech.ebike.rider.platform.BleTransport
import com.luopingtech.ebike.rider.platform.SimulatorBleTransport
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.test.runTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class BleSessionFeatureTest {
    @Test
    fun simulatorUnlock_codeZeroIsSuccess() = runTest {
        val session = BleSessionFeature(SimulatorBleTransport(), DemoBleTokenApi())
        val result = session.unlock("867567046128534", mute = true)
        assertTrue(result is RiderResult.Ok)
        val ack = result.getOrNull()!!
        assertTrue(ack.success)
        assertEquals(0, ack.code)
        assertEquals(1, session.connectionHistory.size)
    }

    @Test
    fun history_reusesDevice() = runTest {
        val session = BleSessionFeature(SimulatorBleTransport(), DemoBleTokenApi())
        session.unlock("867567046128534")
        session.unlock("867567046128534")
        assertEquals(1, session.connectionHistory.size)
    }

    @Test
    fun writeAndAwait_nackWhenCodeNotZero() = runTest {
        val transport = ScriptedBleTransport(ackHex = "2c01012e")
        val session = BleSessionFeature(transport, DemoBleTokenApi())
        val result = session.unlock("867567046128534")
        assertTrue(result is RiderResult.Err)
        assertEquals("BLE_NACK", result.error.code)
    }

    @Test
    fun muteUnlock_usesMutePayload() = runTest {
        val transport = ScriptedBleTransport()
        val session = BleSessionFeature(transport, DemoBleTokenApi())
        session.unlock("867000000000001", mute = true)
        assertEquals(
            BleFrame.build(BleCommands.UNLOCK, BleFrame.MUTE_PAYLOAD),
            transport.lastWrittenHex,
        )
    }
}

private class ScriptedBleTransport(
    private val ackHex: String? = null,
) : BleTransport {
    override val isAvailable: Boolean = true
    var lastWrittenHex: String = ""
    private val state = MutableStateFlow(BleConnectionState.Disconnected)

    override fun connectionState(): Flow<BleConnectionState> = state.asStateFlow()

    override suspend fun openAdapter(): RiderResult<Unit> = RiderResult.Ok(Unit)

    override suspend fun closeAdapter(): RiderResult<Unit> {
        state.value = BleConnectionState.Disconnected
        return RiderResult.Ok(Unit)
    }

    override suspend fun findByImei(imei: String, timeoutMs: Long): RiderResult<BleAdvertDevice> =
        RiderResult.Ok(BleAdvertDevice(id = "DEV", advertHex = imei.drop(3).take(12)))

    override suspend fun connect(deviceId: String, timeoutMs: Long): RiderResult<Unit> {
        state.value = BleConnectionState.Connected
        return RiderResult.Ok(Unit)
    }

    override suspend fun disconnect(): RiderResult<Unit> {
        state.value = BleConnectionState.Disconnected
        return RiderResult.Ok(Unit)
    }

    override suspend fun writeAndAwait(hex: String, timeoutMs: Long): RiderResult<BleNotifyFrame> {
        lastWrittenHex = hex
        val cmd = hex.take(2).toIntOrNull(16) ?: 0
        val crc = (cmd + 1) and 0xff
        val fallback = BleConvert.toPad2HexStr(cmd) + "01" + "00" + BleConvert.toPad2HexStr(crc)
        return RiderResult.Ok(BleNotifyFrame(ackHex ?: fallback))
    }
}
