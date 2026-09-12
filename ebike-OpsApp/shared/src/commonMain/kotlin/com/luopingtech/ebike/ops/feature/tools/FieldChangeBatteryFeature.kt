package com.luopingtech.ebike.ops.feature.tools

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.vehicle.VehicleRepository
import com.luopingtech.ebike.ops.domain.control.ControlChannel
import com.luopingtech.ebike.ops.domain.control.VehicleAction
import com.luopingtech.ebike.ops.domain.control.VehicleControlPolicy
import com.luopingtech.ebike.ops.domain.scan.ScanCodeParser
import com.luopingtech.ebike.ops.domain.scan.ScanTarget
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class FieldChangeBatteryUiState(
    val loading: Boolean = false,
    val carInput: String = "",
    val resolvedCarId: String? = null,
    val resolvedImei: String = "",
    val message: String? = null,
    val errorMessage: String? = null,
)

/**
 * 工作台现场换电：扫码 / 输入车号 → 开电池仓（permission + paas batteryCompartment）。
 * 与任务中心 [ChangeBatteryTaskFeature] 无关。
 */
class FieldChangeBatteryFeature(
    private val repository: VehicleRepository,
    private val control: VehicleControlPolicy,
    private val qrHostsProvider: () -> List<String> = { emptyList() },
) {
    private val _state = MutableStateFlow(FieldChangeBatteryUiState())
    val state: StateFlow<FieldChangeBatteryUiState> = _state.asStateFlow()

    fun setCarInput(value: String) {
        _state.value = _state.value.copy(
            carInput = value,
            errorMessage = null,
            message = null,
        )
    }

    fun clear() {
        _state.value = FieldChangeBatteryUiState()
    }

    suspend fun applyScanRaw(raw: String): Boolean {
        val hosts = qrHostsProvider()
        val target = ScanCodeParser.parse(raw, hosts)
            ?: run {
                _state.value = _state.value.copy(errorMessage = Strings.t(Str.CannotParseScan))
                return false
            }
        val carId = when (target) {
            is ScanTarget.CarId -> target.value
            is ScanTarget.Imei -> target.value
        }
        _state.value = _state.value.copy(carInput = carId, errorMessage = null)
        return resolveAndRemember(carId)
    }

    suspend fun openBatteryBox(channel: ControlChannel = ControlChannel.NetworkOnly): OpsResult<Unit> {
        val input = _state.value.carInput.trim().ifBlank { _state.value.resolvedCarId.orEmpty() }
        if (input.isBlank()) {
            val err = OpsResult.Err(OpsError.business("BAT_CAR", Strings.t(Str.OpenBoxNeedCarId)))
            _state.value = _state.value.copy(errorMessage = err.error.message)
            return err
        }
        if (!resolveAndRemember(input)) {
            return OpsResult.Err(
                OpsError.business("BAT_RESOLVE", _state.value.errorMessage ?: Strings.t(Str.CannotParseScan)),
            )
        }
        val carId = _state.value.resolvedCarId.orEmpty()
        val imei = _state.value.resolvedImei
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        return when (
            val result = control.execute(
                vehicleId = carId,
                action = VehicleAction.OpenBatteryBox,
                channel = channel,
                imei = imei,
            )
        ) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    message = Strings.t(Str.OpenBoxOk, carId),
                )
                result
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    errorMessage = Strings.t(Str.OpenBoxFailed, result.error.message),
                )
                result
            }
        }
    }

    private suspend fun resolveAndRemember(raw: String): Boolean {
        val hosts = qrHostsProvider()
        val target = ScanCodeParser.parse(raw, hosts)
            ?: run {
                // 纯数字车号允许直接用
                val trimmed = raw.trim()
                if (trimmed.isBlank()) {
                    _state.value = _state.value.copy(errorMessage = Strings.t(Str.EnterCarId))
                    return false
                }
                _state.value = _state.value.copy(
                    resolvedCarId = trimmed,
                    resolvedImei = "",
                    carInput = trimmed,
                )
                return true
            }
        return when (val result = repository.findByScanTarget(target)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    resolvedCarId = result.value.carId,
                    resolvedImei = result.value.imei,
                    carInput = result.value.carId,
                    errorMessage = null,
                )
                true
            }
            is OpsResult.Err -> {
                // 详情接口失败时仍可用扫到的车号开仓（遗留现场换电也允许）
                val fallback = when (target) {
                    is ScanTarget.CarId -> target.value
                    is ScanTarget.Imei -> target.value
                }
                _state.value = _state.value.copy(
                    resolvedCarId = fallback,
                    resolvedImei = if (target is ScanTarget.Imei) target.value else "",
                    carInput = fallback,
                    message = result.error.message,
                    errorMessage = null,
                )
                true
            }
        }
    }
}
