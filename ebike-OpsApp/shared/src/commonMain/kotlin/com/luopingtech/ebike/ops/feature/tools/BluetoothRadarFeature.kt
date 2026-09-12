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
    val ringing: Boolean = false,
    val bleAvailable: Boolean = true,
    val devices: List<BleDevice> = emptyList(),
    val selectedId: String? = null,
    val message: String? = null,
    val errorMessage: String? = null,
)

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

    fun selectDevice(id: String) {
        _state.value = _state.value.copy(selectedId = id, errorMessage = null)
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
                selectedId = result.value.firstOrNull()?.id,
            )
            is OpsResult.Err -> _state.value = _state.value.copy(scanning = false, errorMessage = result.error.message)
        }
    }

    suspend fun ring() {
        val deviceId = _state.value.selectedId
        if (deviceId.isNullOrBlank()) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.BleSelectDevice))
            return
        }
        _state.value = _state.value.copy(ringing = true, errorMessage = null, message = null)
        when (
            val result = control.execute(
                vehicleId = deviceId,
                action = VehicleAction.Ring,
                channel = ControlChannel.NetworkOnly,
            )
        ) {
            is OpsResult.Ok -> _state.value = _state.value.copy(
                ringing = false,
                message = Strings.t(Str.ActionOk, Strings.t(Str.BleRing), deviceId),
            )
            is OpsResult.Err -> _state.value = _state.value.copy(ringing = false, errorMessage = result.error.message)
        }
    }
}
