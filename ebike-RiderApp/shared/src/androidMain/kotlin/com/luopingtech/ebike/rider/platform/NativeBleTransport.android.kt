package com.luopingtech.ebike.rider.platform

import android.Manifest
import android.annotation.SuppressLint
import android.bluetooth.BluetoothAdapter
import android.bluetooth.BluetoothDevice
import android.bluetooth.BluetoothGatt
import android.bluetooth.BluetoothGattCallback
import android.bluetooth.BluetoothGattCharacteristic
import android.bluetooth.BluetoothGattDescriptor
import android.bluetooth.BluetoothManager
import android.bluetooth.BluetoothProfile
import android.bluetooth.le.ScanCallback
import android.bluetooth.le.ScanFilter
import android.bluetooth.le.ScanResult
import android.bluetooth.le.ScanSettings
import android.content.Context
import android.content.pm.PackageManager
import android.os.Build
import android.os.ParcelUuid
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.core.i18n.Strings
import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.domain.ble.BleCommands
import com.luopingtech.ebike.rider.domain.ble.BleConvert
import com.luopingtech.ebike.rider.domain.ble.BleFrame
import java.util.UUID
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.withContext
import kotlinx.coroutines.withTimeoutOrNull

/**
 * Android GATT central. Scan filter `5841`, write RX / notify TX as UniApp.
 *
 * Vendor stacks (Xiaomi / OPPO / Vivo / Harmony) may delay CCCD writes or drop
 * the first notify — documented, not papered over here.
 */
@SuppressLint("MissingPermission")
actual class NativeBleTransport(
    context: Context,
) : BleTransport {
    private val appContext = context.applicationContext
    private val state = MutableStateFlow(BleConnectionState.Disconnected)

    @Volatile
    private var gatt: BluetoothGatt? = null

    @Volatile
    private var writeChar: BluetoothGattCharacteristic? = null

    @Volatile
    private var notifyWaiter: CompletableDeferred<String>? = null

    @Volatile
    private var connectWaiter: CompletableDeferred<RiderResult<Unit>>? = null

    actual override val isAvailable: Boolean
        get() = appContext.packageManager.hasSystemFeature(PackageManager.FEATURE_BLUETOOTH_LE)

    actual override fun connectionState(): Flow<BleConnectionState> = state.asStateFlow()

    actual override suspend fun openAdapter(): RiderResult<Unit> {
        if (!isAvailable) {
            return RiderResult.Err(RiderError.unsupported(Strings.t(Str.BleUnavailable)))
        }
        if (!hasPermissions()) {
            return RiderResult.Err(RiderError.business("BLE_PERMISSION", Strings.t(Str.BlePermissionRequired)))
        }
        val adapter = adapterOrNull()
            ?: return RiderResult.Err(RiderError.business("BLE_ADAPTER_OFF", Strings.t(Str.BleAdapterOff)))
        if (!adapter.isEnabled) {
            return RiderResult.Err(RiderError.business("BLE_ADAPTER_OFF", Strings.t(Str.BleAdapterOff)))
        }
        return RiderResult.Ok(Unit)
    }

    actual override suspend fun closeAdapter(): RiderResult<Unit> = disconnect()

    actual override suspend fun findByImei(imei: String, timeoutMs: Long): RiderResult<BleAdvertDevice> {
        val opened = openAdapter()
        if (opened is RiderResult.Err) return opened
        val adapter = adapterOrNull()
            ?: return RiderResult.Err(RiderError.business("BLE_ADAPTER_OFF", Strings.t(Str.BleAdapterOff)))
        val scanner = adapter.bluetoothLeScanner
            ?: return RiderResult.Err(RiderError.business("BLE_ADAPTER_OFF", Strings.t(Str.BleAdapterOff)))

        return withContext(Dispatchers.Main.immediate) {
            val found = CompletableDeferred<BleAdvertDevice>()
            val callback = object : ScanCallback() {
                override fun onScanResult(callbackType: Int, result: ScanResult) {
                    val advert = advertHexCandidates(result).firstOrNull { BleFrame.matchImeiAdvert(imei, it) }
                        ?: return
                    val device = BleAdvertDevice(
                        id = result.device.address,
                        name = result.scanRecord?.deviceName ?: result.device.name,
                        rssi = result.rssi,
                        advertHex = advert,
                    )
                    found.complete(device)
                }

                override fun onScanFailed(errorCode: Int) {
                    found.completeExceptionally(IllegalStateException("scan failed $errorCode"))
                }
            }
            val filter = ScanFilter.Builder()
                .setServiceUuid(ParcelUuid.fromString(BleCommands.SERVICE_FILTER_UUID))
                .build()
            val settings = ScanSettings.Builder()
                .setScanMode(ScanSettings.SCAN_MODE_LOW_LATENCY)
                .build()
            try {
                scanner.startScan(listOf(filter), settings, callback)
                val hit = withTimeoutOrNull(timeoutMs) { found.await() }
                if (hit == null) {
                    RiderResult.Err(RiderError.business("BLE_NOT_FOUND", Strings.t(Str.BleNotFound)))
                } else {
                    RiderResult.Ok(hit)
                }
            } catch (t: Throwable) {
                RiderResult.Err(RiderError.business("BLE_NOT_FOUND", t.message ?: Strings.t(Str.BleNotFound)))
            } finally {
                runCatching { scanner.stopScan(callback) }
            }
        }
    }

    actual override suspend fun connect(deviceId: String, timeoutMs: Long): RiderResult<Unit> {
        val opened = openAdapter()
        if (opened is RiderResult.Err) return opened
        val adapter = adapterOrNull()
            ?: return RiderResult.Err(RiderError.business("BLE_ADAPTER_OFF", Strings.t(Str.BleAdapterOff)))
        if (deviceId.isBlank()) {
            return RiderResult.Err(RiderError.business("BLE_NOT_FOUND", Strings.t(Str.BleNotFound)))
        }
        val existing = gatt
        if (existing != null && existing.device?.address.equals(deviceId, ignoreCase = true) &&
            state.value == BleConnectionState.Connected && writeChar != null
        ) {
            return RiderResult.Ok(Unit)
        }
        disconnectSilent()
        state.value = BleConnectionState.Connecting
        val device: BluetoothDevice = try {
            adapter.getRemoteDevice(deviceId)
        } catch (t: Throwable) {
            state.value = BleConnectionState.Failed
            return RiderResult.Err(RiderError.business("BLE_CONNECT_FAILED", t.message ?: deviceId))
        }
        val waiter = CompletableDeferred<RiderResult<Unit>>()
        connectWaiter = waiter
        val callback = GattCallback()
        val created = withContext(Dispatchers.Main.immediate) {
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
                device.connectGatt(appContext, false, callback, BluetoothDevice.TRANSPORT_LE)
            } else {
                device.connectGatt(appContext, false, callback)
            }
        }
        if (created == null) {
            state.value = BleConnectionState.Failed
            return RiderResult.Err(RiderError.business("BLE_CONNECT_FAILED", Strings.t(Str.BleUnavailable)))
        }
        gatt = created
        val result = withTimeoutOrNull(timeoutMs) { waiter.await() }
            ?: RiderResult.Err(RiderError.business("BLE_TIMEOUT", Strings.t(Str.BleTimeout)))
        if (result is RiderResult.Err) {
            disconnectSilent()
            state.value = BleConnectionState.Failed
        }
        return result
    }

    actual override suspend fun disconnect(): RiderResult<Unit> {
        disconnectSilent()
        state.value = BleConnectionState.Disconnected
        return RiderResult.Ok(Unit)
    }

    actual override suspend fun writeAndAwait(hex: String, timeoutMs: Long): RiderResult<BleNotifyFrame> {
        val characteristic = writeChar
        val connected = gatt
        if (characteristic == null || connected == null || state.value != BleConnectionState.Connected) {
            return RiderResult.Err(RiderError.business("BLE_NOT_CONNECTED", Strings.t(Str.BleUnavailable)))
        }
        val bytes = BleConvert.hexToByteArray(hex)
        val waiter = CompletableDeferred<String>()
        notifyWaiter = waiter
        characteristic.value = bytes
        val writeType = if (characteristic.properties and BluetoothGattCharacteristic.PROPERTY_WRITE != 0) {
            BluetoothGattCharacteristic.WRITE_TYPE_DEFAULT
        } else {
            BluetoothGattCharacteristic.WRITE_TYPE_NO_RESPONSE
        }
        characteristic.writeType = writeType
        val queued = connected.writeCharacteristic(characteristic)
        if (!queued) {
            notifyWaiter = null
            return RiderResult.Err(RiderError.business("BLE_WRITE_FAILED", Strings.t(Str.OperationFailed)))
        }
        val hexAck = withTimeoutOrNull(timeoutMs) { waiter.await() }
        notifyWaiter = null
        return if (hexAck == null) {
            RiderResult.Err(RiderError.business("BLE_TIMEOUT", Strings.t(Str.BleTimeout)))
        } else {
            RiderResult.Ok(BleNotifyFrame(hexAck))
        }
    }

    private fun adapterOrNull(): BluetoothAdapter? {
        val manager = appContext.getSystemService(Context.BLUETOOTH_SERVICE) as? BluetoothManager
        return manager?.adapter
    }

    private fun hasPermissions(): Boolean {
        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
            granted(Manifest.permission.BLUETOOTH_SCAN) && granted(Manifest.permission.BLUETOOTH_CONNECT)
        } else {
            granted(Manifest.permission.ACCESS_FINE_LOCATION) ||
                granted(Manifest.permission.ACCESS_COARSE_LOCATION)
        }
    }

    private fun granted(permission: String): Boolean =
        appContext.checkSelfPermission(permission) == PackageManager.PERMISSION_GRANTED

    private fun disconnectSilent() {
        notifyWaiter?.cancel()
        notifyWaiter = null
        connectWaiter?.cancel()
        connectWaiter = null
        writeChar = null
        runCatching { gatt?.disconnect() }
        runCatching { gatt?.close() }
        gatt = null
    }

    private inner class GattCallback : BluetoothGattCallback() {
        override fun onConnectionStateChange(gatt: BluetoothGatt, status: Int, newState: Int) {
            if (newState == BluetoothProfile.STATE_CONNECTED) {
                gatt.discoverServices()
            } else if (newState == BluetoothProfile.STATE_DISCONNECTED) {
                state.value = BleConnectionState.Disconnected
                connectWaiter?.complete(
                    RiderResult.Err(RiderError.business("BLE_CONNECT_FAILED", "status=$status")),
                )
                connectWaiter = null
            }
        }

        override fun onServicesDiscovered(gatt: BluetoothGatt, status: Int) {
            if (status != BluetoothGatt.GATT_SUCCESS) {
                failConnect("discover $status")
                return
            }
            val service = gatt.services.firstOrNull { svc ->
                uuidMatches(svc.uuid, BleCommands.SERVICE_UUID) ||
                    svc.uuid.toString().uppercase().contains(BleCommands.SERVICE_UUID_TAIL)
            } ?: gatt.services.firstOrNull()
            if (service == null) {
                failConnect("no service")
                return
            }
            val write = service.characteristics.firstOrNull { uuidMatches(it.uuid, BleCommands.RX_UUID) }
                ?: service.characteristics.firstOrNull {
                    val p = it.properties
                    (p and BluetoothGattCharacteristic.PROPERTY_WRITE) != 0 ||
                        (p and BluetoothGattCharacteristic.PROPERTY_WRITE_NO_RESPONSE) != 0
                }
            val notify = service.characteristics.firstOrNull { uuidMatches(it.uuid, BleCommands.TX_UUID) }
                ?: service.characteristics.firstOrNull {
                    val p = it.properties
                    (p and BluetoothGattCharacteristic.PROPERTY_NOTIFY) != 0 ||
                        (p and BluetoothGattCharacteristic.PROPERTY_INDICATE) != 0
                }
            if (write == null || notify == null) {
                failConnect("missing rx/tx")
                return
            }
            writeChar = write
            gatt.setCharacteristicNotification(notify, true)
            val cccd = notify.getDescriptor(CCCD)
            if (cccd != null) {
                cccd.value = BluetoothGattDescriptor.ENABLE_NOTIFICATION_VALUE
                gatt.writeDescriptor(cccd)
            } else {
                markConnected()
            }
        }

        override fun onDescriptorWrite(
            gatt: BluetoothGatt,
            descriptor: BluetoothGattDescriptor,
            status: Int,
        ) {
            markConnected()
        }

        override fun onCharacteristicChanged(
            gatt: BluetoothGatt,
            characteristic: BluetoothGattCharacteristic,
        ) {
            val hex = BleConvert.toHexString(characteristic.value ?: ByteArray(0))
            notifyWaiter?.complete(hex)
        }

        override fun onCharacteristicWrite(
            gatt: BluetoothGatt,
            characteristic: BluetoothGattCharacteristic,
            status: Int,
        ) {
            if (status != BluetoothGatt.GATT_SUCCESS) {
                notifyWaiter?.completeExceptionally(IllegalStateException("write status=$status"))
            }
        }

        private fun failConnect(reason: String) {
            state.value = BleConnectionState.Failed
            connectWaiter?.complete(
                RiderResult.Err(RiderError.business("BLE_CONNECT_FAILED", reason)),
            )
            connectWaiter = null
        }

        private fun markConnected() {
            state.value = BleConnectionState.Connected
            connectWaiter?.complete(RiderResult.Ok(Unit))
            connectWaiter = null
        }
    }

    companion object {
        private val CCCD: UUID = UUID.fromString("00002902-0000-1000-8000-00805f9b34fb")

        private fun uuidMatches(uuid: UUID, expected: String): Boolean =
            uuid.toString().equals(expected, ignoreCase = true)

        private fun advertHexCandidates(result: ScanResult): List<String> {
            val record = result.scanRecord ?: return emptyList()
            val out = ArrayList<String>()
            val mfg = record.manufacturerSpecificData
            if (mfg != null) {
                for (i in 0 until mfg.size()) {
                    val value = mfg.valueAt(i) ?: continue
                    val hex = BleConvert.toHexString(value)
                    out.add(hex)
                    if (value.size > 6) {
                        out.add(BleConvert.toHexString(value.copyOfRange(2, value.size)))
                    }
                }
            }
            record.serviceData?.values?.forEach { value ->
                out.add(BleConvert.toHexString(value))
            }
            return out
        }
    }
}
