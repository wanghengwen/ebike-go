package com.luopingtech.ebike.rider.domain.ble

/**
 * C-end BLE command / voice / UUID constants.
 * Ported from UniApp `commands.ts` + `config.ts` (no Ops maintenance set).
 */
object BleCommands {
    const val PLAY_VOICE: Int = 0x28
    const val LOCK: Int = 0x2b
    const val UNLOCK: Int = 0x2c
    const val GET_GPS: Int = 0x32
    const val GET_DEVICE_INFO: Int = 0x41
    const val GET_LAST_BEACON: Int = 0x42
    const val GET_RFID: Int = 0x54

    const val SERVICE_FILTER: String = "5841"
    const val SERVICE_UUID: String = "0783B03E-8535-B5A0-7140-A304D2495CB7"
    const val RX_UUID: String = "0783B03E-8535-B5A0-7140-A304D2495CBA"
    const val TX_UUID: String = "0783B03E-8535-B5A0-7140-A304D2495CB8"

    /** 16-bit filter `5841` expanded to Bluetooth base UUID. */
    const val SERVICE_FILTER_UUID: String = "00005841-0000-1000-8000-00805F9B34FB"

    const val SEARCH_TIMEOUT_MS: Long = 60_000L
    const val CONNECT_TIMEOUT_MS: Long = 60_000L
    const val SEND_TIMEOUT_MS: Long = 60_000L

    /** UniApp `BLE_SERVICE_UUID.slice(-12)`. */
    const val SERVICE_UUID_TAIL: String = "A304D2495CB7"
}

object BleVoice {
    const val START: Int = 3
    const val LOCK: Int = 1
    const val TEMP_LOCK: Int = 7
    const val DEFAULT_VOLUME: Int = 100
}
