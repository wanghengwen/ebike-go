package com.luopingtech.ebike.rider.feature.ble

import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.core.i18n.Strings
import com.luopingtech.ebike.rider.core.logging.NoOpLogger
import com.luopingtech.ebike.rider.core.logging.RiderLogger
import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.data.ble.BleTokenApi
import com.luopingtech.ebike.rider.domain.ble.BleAckFrame
import com.luopingtech.ebike.rider.domain.ble.BleCommands
import com.luopingtech.ebike.rider.domain.ble.BleFrame
import com.luopingtech.ebike.rider.domain.ble.BleFrameParser
import com.luopingtech.ebike.rider.domain.ble.BleNotifyDecoded
import com.luopingtech.ebike.rider.domain.ble.BleVoice
import com.luopingtech.ebike.rider.platform.BleAdvertDevice
import com.luopingtech.ebike.rider.platform.BleTransport
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock

data class BleAck(
    val success: Boolean,
    val code: Int,
    val hex: String,
    val decoded: BleNotifyDecoded = BleNotifyDecoded(cmd = 0),
)

data class BleHistoryItem(
    val imei: String,
    val device: BleAdvertDevice,
)

/**
 * UniApp `BleSession`: history, connect-by-IMEI, writeAndAwait, `code === 0` success.
 * Network ride-report fallback is P4 — this layer only returns a clear [RiderResult].
 */
class BleSessionFeature(
    private val transport: BleTransport,
    private val tokenApi: BleTokenApi,
    private val logger: RiderLogger = NoOpLogger,
) {
    private val mutex = Mutex()
    private val history = ArrayList<BleHistoryItem>()
    private var currentImei: String? = null

    val connectionHistory: List<BleHistoryItem> get() = history.toList()

    suspend fun unlock(imei: String, mute: Boolean = true): RiderResult<BleAck> {
        val payload = if (mute) BleFrame.MUTE_PAYLOAD else intArrayOf()
        return send(imei, BleCommands.UNLOCK, payload)
    }

    suspend fun lock(imei: String, mute: Boolean = true): RiderResult<BleAck> {
        val payload = if (mute) BleFrame.MUTE_PAYLOAD else intArrayOf()
        return send(imei, BleCommands.LOCK, payload)
    }

    suspend fun tempLock(imei: String): RiderResult<BleAck> =
        send(imei, BleCommands.LOCK, intArrayOf(BleVoice.TEMP_LOCK, BleVoice.DEFAULT_VOLUME))

    suspend fun playVoice(
        imei: String,
        voiceIdx: Int,
        volume: Int = BleVoice.DEFAULT_VOLUME,
    ): RiderResult<BleAck> = send(imei, BleCommands.PLAY_VOICE, intArrayOf(voiceIdx, volume))

    suspend fun queryDeviceInfo(imei: String): RiderResult<BleAck> =
        send(imei, BleCommands.GET_DEVICE_INFO)

    suspend fun queryGps(imei: String): RiderResult<BleAck> =
        send(imei, BleCommands.GET_GPS)

    suspend fun checkBeacon(imei: String): RiderResult<BleAck> =
        send(imei, BleCommands.GET_LAST_BEACON)

    suspend fun queryRfid(imei: String): RiderResult<BleAck> =
        send(imei, BleCommands.GET_RFID)

    suspend fun connectByImei(imei: String): RiderResult<BleAdvertDevice> = mutex.withLock {
        connectLocked(imei)
    }

    suspend fun disconnect() {
        mutex.withLock {
            transport.disconnect()
            currentImei = null
        }
    }

    suspend fun writeAndAwait(imei: String, cmd: Int, payload: IntArray): RiderResult<BleAck> =
        send(imei, cmd, payload)

    private suspend fun send(imei: String, cmd: Int, payload: IntArray = intArrayOf()): RiderResult<BleAck> =
        mutex.withLock {
            val linked = connectLocked(imei)
            if (linked is RiderResult.Err) return linked
            val token = fetchTokenBytes(imei)
            val hex = BleFrame.build(cmd, payload, token)
            logger.d(TAG, "write cmd=0x${cmd.toString(16)} hex=$hex")
            when (val written = transport.writeAndAwait(hex, BleCommands.SEND_TIMEOUT_MS)) {
                is RiderResult.Err -> written
                is RiderResult.Ok -> {
                    val ack: BleAckFrame = BleFrameParser.resolveAck(written.value.hex)
                    val decoded = BleFrameParser.parseNotify(written.value.hex)
                    // Legacy: only code === 0 counts as success
                    if (ack.code == 0) {
                        RiderResult.Ok(
                            BleAck(
                                success = true,
                                code = ack.code,
                                hex = written.value.hex,
                                decoded = decoded,
                            ),
                        )
                    } else {
                        RiderResult.Err(
                            RiderError.business(
                                "BLE_NACK",
                                Strings.t(Str.BleNack, ack.code),
                            ),
                        )
                    }
                }
            }
        }

    private suspend fun connectLocked(imei: String): RiderResult<BleAdvertDevice> {
        val opened = transport.openAdapter()
        if (opened is RiderResult.Err) return opened
        if (currentImei == imei) {
            history.firstOrNull { it.imei == imei }?.let { return RiderResult.Ok(it.device) }
        }
        val fromHistory = history.lastOrNull { it.imei == imei }
        if (fromHistory != null) {
            when (val reused = transport.connect(fromHistory.device.id, BleCommands.CONNECT_TIMEOUT_MS)) {
                is RiderResult.Ok -> {
                    currentImei = imei
                    return RiderResult.Ok(fromHistory.device)
                }
                is RiderResult.Err -> {
                    logger.w(TAG, "history reconnect failed ${fromHistory.device.id}")
                }
            }
        }
        val found = when (val scan = transport.findByImei(imei, BleCommands.SEARCH_TIMEOUT_MS)) {
            is RiderResult.Ok -> scan.value
            is RiderResult.Err -> return scan
        }
        val connected = transport.connect(found.id, BleCommands.CONNECT_TIMEOUT_MS)
        if (connected is RiderResult.Err) return connected
        history.removeAll { it.device.id == found.id }
        history.add(BleHistoryItem(imei, found))
        currentImei = imei
        return RiderResult.Ok(found)
    }

    private suspend fun fetchTokenBytes(imei: String): ByteArray {
        return when (val result = tokenApi.getToken(imei)) {
            is RiderResult.Ok -> BleFrame.tokenBytes(result.value)
            is RiderResult.Err -> {
                logger.w(TAG, "ble token fail ${result.error.code}")
                BleFrame.DEFAULT_TOKEN
            }
        }
    }

    companion object {
        private const val TAG = "BleSession"
    }
}
