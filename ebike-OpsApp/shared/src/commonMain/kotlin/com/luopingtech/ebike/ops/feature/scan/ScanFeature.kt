package com.luopingtech.ebike.ops.feature.scan

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.vehicle.VehicleRepository
import com.luopingtech.ebike.ops.domain.control.ControlChannel
import com.luopingtech.ebike.ops.domain.control.VehicleAction
import com.luopingtech.ebike.ops.domain.control.VehicleControlPolicy
import com.luopingtech.ebike.ops.domain.model.Vehicle
import com.luopingtech.ebike.ops.domain.scan.ScanCodeParser
import com.luopingtech.ebike.ops.platform.CodeScanner
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class ScanUiState(
    val loading: Boolean = false,
    val rawInput: String = "",
    val vehicle: Vehicle? = null,
    val message: String? = null,
    val errorMessage: String? = null,
)

class ScanFeature(
    private val repository: VehicleRepository,
    private val control: VehicleControlPolicy,
    private val codeScanner: CodeScanner,
    private val qrHostsProvider: () -> List<String> = { emptyList() },
    private val serviceAreaIdProvider: () -> String = { "" },
) {
    private val _state = MutableStateFlow(ScanUiState())
    val state: StateFlow<ScanUiState> = _state.asStateFlow()

    fun updateRawInput(raw: String) {
        _state.value = _state.value.copy(rawInput = raw, errorMessage = null)
    }

    suspend fun resolveManual(raw: String = _state.value.rawInput): OpsResult<Vehicle> {
        val trimmed = raw.trim()
        _state.value = _state.value.copy(
            loading = true,
            rawInput = trimmed,
            errorMessage = null,
            message = null,
            vehicle = null,
        )
        val hosts = qrHostsProvider()
        val target = ScanCodeParser.parse(trimmed, allowedHosts = hosts)
        if (target == null) {
            val msg = if (trimmed.startsWith("http", ignoreCase = true)) {
                Strings.t(Str.QrHostNotAllowed)
            } else {
                Strings.t(Str.EnterCarIdOrScan)
            }
            val err = OpsResult.Err(OpsError.business("SCAN_EMPTY", msg))
            _state.value = _state.value.copy(loading = false, errorMessage = err.error.message)
            return err
        }
        return when (val result = repository.findByScanTarget(target)) {
            is OpsResult.Ok -> enforceServiceArea(result.value)
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    errorMessage = result.error.message,
                )
                result
            }
        }
    }

    suspend fun resolveFromScanner(): OpsResult<Vehicle> {
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        return when (val scanned = codeScanner.scanOnce()) {
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    errorMessage = scanned.error.message,
                )
                scanned
            }
            is OpsResult.Ok -> resolveManual(scanned.value)
        }
    }

    suspend fun unlock(channel: ControlChannel = ControlChannel.NetworkOnly): OpsResult<Unit> =
        act(VehicleAction.Unlock, channel, Strings.t(Str.ScanUnlock))

    suspend fun lock(channel: ControlChannel = ControlChannel.NetworkOnly): OpsResult<Unit> =
        act(VehicleAction.Lock, channel, Strings.t(Str.ScanLock))

    suspend fun ring(channel: ControlChannel = ControlChannel.NetworkOnly): OpsResult<Unit> =
        act(VehicleAction.Ring, channel, Strings.t(Str.ScanRing))

    fun clear() {
        _state.value = ScanUiState()
    }

    private suspend fun enforceServiceArea(vehicle: Vehicle): OpsResult<Vehicle> {
        val serviceId = serviceAreaIdProvider().ifBlank { vehicle.serviceId }
        if (serviceId.isBlank()) {
            _state.value = _state.value.copy(
                loading = false,
                vehicle = vehicle,
                message = Strings.t(Str.ResolvedVehicle, vehicle.carId),
            )
            return OpsResult.Ok(vehicle)
        }
        return when (val check = repository.checkServicePermission(vehicle.carId, serviceId)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    vehicle = vehicle,
                    message = Strings.t(Str.ResolvedVehicle, vehicle.carId),
                )
                OpsResult.Ok(vehicle)
            }
            is OpsResult.Err -> {
                val msg = check.error.message.ifBlank { Strings.t(Str.VehicleNotInServiceArea) }
                _state.value = _state.value.copy(
                    loading = false,
                    vehicle = null,
                    errorMessage = msg,
                )
                OpsResult.Err(OpsError.business(check.error.code.ifBlank { "SCAN_AREA" }, msg))
            }
        }
    }

    private suspend fun act(
        action: VehicleAction,
        channel: ControlChannel,
        label: String,
    ): OpsResult<Unit> {
        val carId = _state.value.vehicle?.carId
        if (carId.isNullOrBlank()) {
            val err = OpsResult.Err(
                OpsError.business("SCAN_NO_VEHICLE", Strings.t(Str.ResolveVehicleFirst)),
            )
            _state.value = _state.value.copy(errorMessage = err.error.message)
            return err
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        return when (val result = control.execute(carId, action, channel)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    message = Strings.t(Str.ActionOk, label, carId),
                )
                result
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    errorMessage = Strings.t(Str.ActionFailed, label, result.error.message),
                )
                result
            }
        }
    }
}
