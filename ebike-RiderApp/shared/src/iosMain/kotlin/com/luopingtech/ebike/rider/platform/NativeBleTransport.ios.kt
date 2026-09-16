package com.luopingtech.ebike.rider.platform

import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.core.i18n.Strings
import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.domain.ble.BleCommands
import com.luopingtech.ebike.rider.domain.ble.BleConvert
import com.luopingtech.ebike.rider.domain.ble.BleFrame
import kotlinx.cinterop.BetaInteropApi
import kotlinx.cinterop.ExperimentalForeignApi
import kotlinx.cinterop.ObjCSignatureOverride
import kotlinx.cinterop.addressOf
import kotlinx.cinterop.usePinned
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.withTimeoutOrNull
import platform.CoreBluetooth.CBAdvertisementDataManufacturerDataKey
import platform.CoreBluetooth.CBAdvertisementDataServiceDataKey
import platform.CoreBluetooth.CBCentralManager
import platform.CoreBluetooth.CBCentralManagerDelegateProtocol
import platform.CoreBluetooth.CBCharacteristic
import platform.CoreBluetooth.CBCharacteristicWriteWithResponse
import platform.CoreBluetooth.CBManagerStatePoweredOn
import platform.CoreBluetooth.CBPeripheral
import platform.CoreBluetooth.CBPeripheralDelegateProtocol
import platform.CoreBluetooth.CBService
import platform.CoreBluetooth.CBUUID
import platform.Foundation.NSData
import platform.Foundation.NSError
import platform.Foundation.NSNumber
import platform.Foundation.create
import platform.darwin.NSObject
import platform.darwin.dispatch_get_main_queue
import platform.posix.memcpy

/**
 * iOS CoreBluetooth central. Windows only needs this source to exist; link is Mac-only.
 */
@OptIn(ExperimentalForeignApi::class, BetaInteropApi::class)
actual class NativeBleTransport : BleTransport {
    private val state = MutableStateFlow(BleConnectionState.Disconnected)
    private val bridge = BleBridge()
    private val manager = CBCentralManager(bridge, dispatch_get_main_queue())

    actual override val isAvailable: Boolean
        get() = true

    actual override fun connectionState(): Flow<BleConnectionState> = state.asStateFlow()

    actual override suspend fun openAdapter(): RiderResult<Unit> {
        return if (manager.state == CBManagerStatePoweredOn) {
            RiderResult.Ok(Unit)
        } else {
            RiderResult.Err(RiderError.business("BLE_ADAPTER_OFF", Strings.t(Str.BleAdapterOff)))
        }
    }

    actual override suspend fun closeAdapter(): RiderResult<Unit> = disconnect()

    actual override suspend fun findByImei(imei: String, timeoutMs: Long): RiderResult<BleAdvertDevice> {
        val opened = openAdapter()
        if (opened is RiderResult.Err) return opened
        val waiter = CompletableDeferred<BleAdvertDevice>()
        bridge.scanImei = imei
        bridge.scanWaiter = waiter
        val filter = listOf(CBUUID.UUIDWithString(BleCommands.SERVICE_FILTER))
        manager.scanForPeripheralsWithServices(filter, null)
        val hit = withTimeoutOrNull(timeoutMs) { waiter.await() }
        manager.stopScan()
        bridge.scanWaiter = null
        bridge.scanImei = null
        return if (hit == null) {
            RiderResult.Err(RiderError.business("BLE_NOT_FOUND", Strings.t(Str.BleNotFound)))
        } else {
            RiderResult.Ok(hit)
        }
    }

    actual override suspend fun connect(deviceId: String, timeoutMs: Long): RiderResult<Unit> {
        val opened = openAdapter()
        if (opened is RiderResult.Err) return opened
        val peripheral = bridge.peripheralById(deviceId)
            ?: return RiderResult.Err(RiderError.business("BLE_NOT_FOUND", Strings.t(Str.BleNotFound)))
        if (state.value == BleConnectionState.Connected && bridge.readyToWrite()) {
            return RiderResult.Ok(Unit)
        }
        state.value = BleConnectionState.Connecting
        val waiter = CompletableDeferred<RiderResult<Unit>>()
        bridge.connectWaiter = waiter
        peripheral.delegate = bridge
        manager.connectPeripheral(peripheral, null)
        val result = withTimeoutOrNull(timeoutMs) { waiter.await() }
            ?: RiderResult.Err(RiderError.business("BLE_TIMEOUT", Strings.t(Str.BleTimeout)))
        if (result is RiderResult.Err) {
            state.value = BleConnectionState.Failed
        } else {
            state.value = BleConnectionState.Connected
        }
        return result
    }

    actual override suspend fun disconnect(): RiderResult<Unit> {
        bridge.activePeripheral?.let { manager.cancelPeripheralConnection(it) }
        bridge.resetLink()
        state.value = BleConnectionState.Disconnected
        return RiderResult.Ok(Unit)
    }

    actual override suspend fun writeAndAwait(hex: String, timeoutMs: Long): RiderResult<BleNotifyFrame> {
        val peripheral = bridge.activePeripheral
        val write = bridge.writeCharacteristic
        if (peripheral == null || write == null || state.value != BleConnectionState.Connected) {
            return RiderResult.Err(RiderError.business("BLE_NOT_CONNECTED", Strings.t(Str.BleUnavailable)))
        }
        val waiter = CompletableDeferred<String>()
        bridge.notifyWaiter = waiter
        val data = BleConvert.hexToByteArray(hex).toNSData()
        peripheral.writeValue(data, write, CBCharacteristicWriteWithResponse)
        val hexAck = withTimeoutOrNull(timeoutMs) { waiter.await() }
        bridge.notifyWaiter = null
        return if (hexAck == null) {
            RiderResult.Err(RiderError.business("BLE_TIMEOUT", Strings.t(Str.BleTimeout)))
        } else {
            RiderResult.Ok(BleNotifyFrame(hexAck))
        }
    }

    private inner class BleBridge : NSObject(), CBCentralManagerDelegateProtocol, CBPeripheralDelegateProtocol {
        var scanImei: String? = null
        var scanWaiter: CompletableDeferred<BleAdvertDevice>? = null
        var connectWaiter: CompletableDeferred<RiderResult<Unit>>? = null
        var notifyWaiter: CompletableDeferred<String>? = null
        var activePeripheral: CBPeripheral? = null
        var writeCharacteristic: CBCharacteristic? = null
        private val retained = ArrayList<CBPeripheral>()

        fun peripheralById(id: String): CBPeripheral? =
            retained.firstOrNull { it.identifier.UUIDString.equals(id, ignoreCase = true) }

        fun readyToWrite(): Boolean = writeCharacteristic != null

        fun resetLink() {
            writeCharacteristic = null
            activePeripheral = null
            connectWaiter?.cancel()
            connectWaiter = null
            notifyWaiter?.cancel()
            notifyWaiter = null
        }

        override fun centralManagerDidUpdateState(central: CBCentralManager) = Unit

        @ObjCSignatureOverride
        override fun centralManager(
            central: CBCentralManager,
            didDiscoverPeripheral: CBPeripheral,
            advertisementData: Map<Any?, *>,
            RSSI: NSNumber,
        ) {
            val imei = scanImei ?: return
            val candidates = advertHexCandidates(advertisementData)
            val advert = candidates.firstOrNull { BleFrame.matchImeiAdvert(imei, it) } ?: return
            if (retained.none { it.identifier.UUIDString == didDiscoverPeripheral.identifier.UUIDString }) {
                retained.add(didDiscoverPeripheral)
            }
            scanWaiter?.complete(
                BleAdvertDevice(
                    id = didDiscoverPeripheral.identifier.UUIDString,
                    name = didDiscoverPeripheral.name,
                    rssi = RSSI.intValue,
                    advertHex = advert,
                ),
            )
        }

        override fun centralManager(central: CBCentralManager, didConnectPeripheral: CBPeripheral) {
            activePeripheral = didConnectPeripheral
            didConnectPeripheral.delegate = this
            didConnectPeripheral.discoverServices(null)
        }

        @ObjCSignatureOverride
        override fun centralManager(
            central: CBCentralManager,
            didFailToConnectPeripheral: CBPeripheral,
            error: NSError?,
        ) {
            connectWaiter?.complete(
                RiderResult.Err(
                    RiderError.business("BLE_CONNECT_FAILED", error?.localizedDescription ?: "fail"),
                ),
            )
            connectWaiter = null
        }

        @ObjCSignatureOverride
        override fun centralManager(
            central: CBCentralManager,
            didDisconnectPeripheral: CBPeripheral,
            error: NSError?,
        ) {
            if (activePeripheral?.identifier?.UUIDString == didDisconnectPeripheral.identifier.UUIDString) {
                state.value = BleConnectionState.Disconnected
                resetLink()
            }
        }

        override fun peripheral(peripheral: CBPeripheral, didDiscoverServices: NSError?) {
            if (didDiscoverServices != null) {
                failConnect(didDiscoverServices.localizedDescription)
                return
            }
            val services = peripheral.services.orEmpty().filterIsInstance<CBService>()
            val service = services.firstOrNull { svc ->
                val uuid = svc.UUID.UUIDString
                uuid.equals(BleCommands.SERVICE_UUID, ignoreCase = true) ||
                    uuid.uppercase().contains(BleCommands.SERVICE_UUID_TAIL)
            } ?: services.firstOrNull()
            if (service == null) {
                failConnect("no service")
                return
            }
            peripheral.discoverCharacteristics(null, service)
        }

        @ObjCSignatureOverride
        override fun peripheral(
            peripheral: CBPeripheral,
            didDiscoverCharacteristicsForService: CBService,
            error: NSError?,
        ) {
            if (error != null) {
                failConnect(error.localizedDescription)
                return
            }
            val chars = didDiscoverCharacteristicsForService.characteristics
                .orEmpty()
                .filterIsInstance<CBCharacteristic>()
            val write = chars.firstOrNull { it.UUID.UUIDString.equals(BleCommands.RX_UUID, ignoreCase = true) }
                ?: chars.firstOrNull()
            val notify = chars.firstOrNull { it.UUID.UUIDString.equals(BleCommands.TX_UUID, ignoreCase = true) }
                ?: chars.lastOrNull()
            if (write == null || notify == null) {
                failConnect("missing rx/tx")
                return
            }
            writeCharacteristic = write
            peripheral.setNotifyValue(true, notify)
        }

        @ObjCSignatureOverride
        override fun peripheral(
            peripheral: CBPeripheral,
            didUpdateNotificationStateForCharacteristic: CBCharacteristic,
            error: NSError?,
        ) {
            connectWaiter?.complete(
                if (error == null) {
                    RiderResult.Ok(Unit)
                } else {
                    RiderResult.Err(RiderError.business("BLE_CONNECT_FAILED", error.localizedDescription))
                },
            )
            connectWaiter = null
        }

        @ObjCSignatureOverride
        override fun peripheral(
            peripheral: CBPeripheral,
            didUpdateValueForCharacteristic: CBCharacteristic,
            error: NSError?,
        ) {
            val data = didUpdateValueForCharacteristic.value ?: return
            val hex = BleConvert.toHexString(data.toByteArray() ?: return)
            notifyWaiter?.complete(hex)
        }

        @ObjCSignatureOverride
        override fun peripheral(
            peripheral: CBPeripheral,
            didWriteValueForCharacteristic: CBCharacteristic,
            error: NSError?,
        ) {
            if (error != null) {
                notifyWaiter?.completeExceptionally(IllegalStateException(error.localizedDescription))
            }
        }

        private fun failConnect(reason: String) {
            connectWaiter?.complete(RiderResult.Err(RiderError.business("BLE_CONNECT_FAILED", reason)))
            connectWaiter = null
        }
    }
}

@OptIn(ExperimentalForeignApi::class)
private fun advertHexCandidates(advertisementData: Map<Any?, *>): List<String> {
    val out = ArrayList<String>()
    val mfg = advertisementData[CBAdvertisementDataManufacturerDataKey] as? NSData
    mfg?.toByteArray()?.let { bytes ->
        out.add(BleConvert.toHexString(bytes))
        if (bytes.size > 6) {
            out.add(BleConvert.toHexString(bytes.copyOfRange(2, bytes.size)))
        }
    }
    val serviceData = advertisementData[CBAdvertisementDataServiceDataKey] as? Map<*, *>
    serviceData?.values?.forEach { value ->
        (value as? NSData)?.toByteArray()?.let { out.add(BleConvert.toHexString(it)) }
    }
    return out
}

@OptIn(ExperimentalForeignApi::class, BetaInteropApi::class)
private fun ByteArray.toNSData(): NSData {
    if (isEmpty()) return NSData()
    return usePinned { pinned ->
        NSData.create(bytes = pinned.addressOf(0), length = size.toULong())
    }
}

@OptIn(ExperimentalForeignApi::class)
private fun NSData.toByteArray(): ByteArray? {
    val size = length.toInt()
    if (size == 0) return ByteArray(0)
    val out = ByteArray(size)
    out.usePinned { pinned -> memcpy(pinned.addressOf(0), bytes, length) }
    return out
}
