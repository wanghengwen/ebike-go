package com.luopingtech.ebike.ops.feature.tools

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.control.ControlChannel
import com.luopingtech.ebike.ops.domain.control.VehicleAction
import com.luopingtech.ebike.ops.domain.control.VehicleControlPolicy
import com.luopingtech.ebike.ops.platform.BleDevice
import com.luopingtech.ebike.ops.platform.BleTransport
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class BluetoothRadarUiState(
    val scanning: Boolean = false,
    val ringingId: String? = null,
    val bleAvailable: Boolean = true,
    val filter: String = "",
    val devices: List<BleDevice> = emptyList(),
    val message: String? = null,
    val errorMessage: String? = null,
) {
    val filteredDevices: List<BleDevice>
        get() {
            val q = filter.trim()
            if (q.isEmpty()) return devices
            return devices.filter {
                (it.name ?: it.id).contains(q, ignoreCase = true) ||
                    it.id.contains(q, ignoreCase = true)
            }
        }
}

class BluetoothRadarFeature(
    private val bleTransport: BleTransport,
    private val control: VehicleControlPolicy,
) {
    private val _state = MutableStateFlow(
        BluetoothRadarUiState(bleAvailable = bleTransport.isAvailable),
    )
    val state: StateFlow<BluetoothRadarUiState> = _state.asStateFlow()

    fun clear() {
        _state.value = BluetoothRadarUiState(bleAvailable = bleTransport.isAvailable)
    }

    fun setFilter(value: String) {
        _state.value = _state.value.copy(filter = value.filter { it.isDigit() }.take(9))
    }

    suspend fun scan() {
        if (!bleTransport.isAvailable) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.BleUnavailable))
            return
        }
        _state.value = _state.value.copy(scanning = true, errorMessage = null, message = null)
        when (val result = bleTransport.scan()) {
            is OpsResult.Ok -> _state.value = _state.value.copy(
                scanning = false,
                devices = result.value,
            )
            is OpsResult.Err -> _state.value = _state.value.copy(
                scanning = false,
                errorMessage = result.error.message,
            )
        }
    }

    suspend fun ring(deviceId: String) {
        val id = deviceId.trim()
        if (id.isEmpty()) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.BleSelectDevice))
            return
        }
        _state.value = _state.value.copy(ringingId = id, errorMessage = null, message = null)
        when (
            val result = control.execute(
                vehicleId = id,
                action = VehicleAction.Ring,
                channel = ControlChannel.NetworkOnly,
            )
        ) {
            is OpsResult.Ok -> _state.value = _state.value.copy(
                ringingId = null,
                message = Strings.t(Str.ActionOk, Strings.t(Str.BleRing), id),
            )
            is OpsResult.Err -> _state.value = _state.value.copy(
                ringingId = null,
                errorMessage = result.error.message,
            )
        }
    }
}
