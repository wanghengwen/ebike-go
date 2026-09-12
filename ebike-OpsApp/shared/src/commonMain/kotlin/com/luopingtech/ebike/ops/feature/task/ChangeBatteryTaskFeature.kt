package com.luopingtech.ebike.ops.feature.task

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.task.ChangeBatteryTaskRepository
import com.luopingtech.ebike.ops.domain.control.ControlChannel
import com.luopingtech.ebike.ops.domain.control.VehicleAction
import com.luopingtech.ebike.ops.domain.control.VehicleControlPolicy
import com.luopingtech.ebike.ops.domain.model.OpsTask
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.scan.ScanCodeParser
import com.luopingtech.ebike.ops.domain.scan.ScanTarget
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class ChangeBatteryTaskUiState(
    val loading: Boolean = false,
    val tasks: List<OpsTask> = emptyList(),
    val selectedTaskId: String? = null,
    val maxBattery: Int = 30,
    val rangeMin: Int = 5,
    val rangeMax: Int = 30,
    val serviceAreaId: String? = null,
    val message: String? = null,
    val errorMessage: String? = null,
) {
    val selected: OpsTask?
        get() = tasks.firstOrNull { it.id == selectedTaskId }
}

/**
 * Change-battery task flow: list → select → ring → open box (start) → close box (finish).
 * Loads [task_rules/battery_range] then filters list by maxBattery.
 */
class ChangeBatteryTaskFeature(
    private val repository: ChangeBatteryTaskRepository,
    private val control: VehicleControlPolicy,
    private val pinProvider: () -> String,
) {
    private val _state = MutableStateFlow(ChangeBatteryTaskUiState())
    val state: StateFlow<ChangeBatteryTaskUiState> = _state.asStateFlow()

    fun setMaxBattery(maxBattery: Int) {
        val min = _state.value.rangeMin
        val max = _state.value.rangeMax
        _state.value = _state.value.copy(maxBattery = maxBattery.coerceIn(min, max))
    }

    fun selectTask(taskId: String?) {
        _state.value = _state.value.copy(selectedTaskId = taskId, errorMessage = null)
    }

    /**
     * Field flow: scan car QR/IMEI then select matching task in the loaded list.
     * @return true if a task was selected.
     */
    fun selectByScanRaw(raw: String, qrHosts: List<String> = emptyList()): Boolean {
        val target = ScanCodeParser.parse(raw, qrHosts)
            ?: run {
                _state.value = _state.value.copy(errorMessage = Strings.t(Str.CannotParseScan))
                return false
            }
        val task = when (target) {
            is ScanTarget.CarId ->
                _state.value.tasks.firstOrNull { it.carId.equals(target.value, ignoreCase = true) }
            is ScanTarget.Imei ->
                _state.value.tasks.firstOrNull { it.imei == target.value }
        }
        if (task == null) {
            _state.value = _state.value.copy(
                errorMessage = Strings.t(Str.VehicleNotInListOrThreshold),
            )
            return false
        }
        _state.value = _state.value.copy(
            selectedTaskId = task.id,
            message = Strings.t(Str.SelectedScanVehicle, task.carId),
            errorMessage = null,
        )
        return true
    }

    /**
     * Fetch battery range (if needed) then list. Pass [maxBattery] to override slider value.
     */
    suspend fun load(
        area: ServiceArea?,
        maxBattery: Int? = null,
        refreshRange: Boolean = true,
    ) {
        if (area == null) {
            _state.value = ChangeBatteryTaskUiState(
                errorMessage = Strings.t(Str.SelectServiceAreaFirst),
            )
            return
        }
        val needRange = refreshRange || _state.value.serviceAreaId != area.id
        _state.value = _state.value.copy(
            loading = true,
            errorMessage = null,
            message = null,
            serviceAreaId = area.id,
        )

        var rangeMin = _state.value.rangeMin
        var rangeMax = _state.value.rangeMax
        if (needRange) {
            when (val rangeResult = repository.batteryRange(area)) {
                is OpsResult.Ok -> {
                    rangeMin = rangeResult.value.min
                    rangeMax = rangeResult.value.max.coerceAtLeast(rangeMin)
                }
                is OpsResult.Err -> {
                    // Keep previous / defaults; still try to load list.
                }
            }
        }

        val threshold = (maxBattery ?: _state.value.maxBattery).coerceIn(rangeMin, rangeMax)
        when (val result = repository.list(area, threshold)) {
            is OpsResult.Ok -> {
                val tasks = result.value
                val keep = _state.value.selectedTaskId
                    ?.takeIf { id -> tasks.any { it.id == id } }
                _state.value = _state.value.copy(
                    loading = false,
                    tasks = tasks,
                    selectedTaskId = keep,
                    maxBattery = threshold,
                    rangeMin = rangeMin,
                    rangeMax = rangeMax,
                    message = Strings.t(Str.ChangeBatteryTasksCount, tasks.size, threshold),
                )
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    rangeMin = rangeMin,
                    rangeMax = rangeMax,
                    maxBattery = threshold,
                    errorMessage = result.error.message,
                )
            }
        }
    }

    suspend fun applyMaxBattery(area: ServiceArea?, maxBattery: Int) {
        setMaxBattery(maxBattery)
        load(area, maxBattery = _state.value.maxBattery, refreshRange = false)
    }

    suspend fun ringSelected(
        channel: ControlChannel = ControlChannel.BlePreferred,
    ): OpsResult<Unit> {
        val task = requireSelected() ?: return missingSelection()
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        return when (
            val result = control.execute(
                vehicleId = task.carId,
                action = VehicleAction.Ring,
                channel = channel,
            )
        ) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    message = Strings.t(Str.RingOk, task.carId),
                )
                result
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    errorMessage = Strings.t(Str.RingFailed, result.error.message),
                )
                result
            }
        }
    }

    suspend fun openBox(izBlue: Boolean = false): OpsResult<Unit> =
        controlBatteryBox(open = true, izBlue = izBlue)

    suspend fun closeBox(izBlue: Boolean = false): OpsResult<Unit> =
        controlBatteryBox(open = false, izBlue = izBlue)

    /**
     * Legacy ReplaceBatteryRepositoryV2.taskControlBatteryBox:
     * network fail + izBlue=false → BLE switch → resubmit same API with izBlue=true.
     */
    private suspend fun controlBatteryBox(open: Boolean, izBlue: Boolean): OpsResult<Unit> {
        val task = requireSelected() ?: return missingSelection()
        val pin = pinProvider().ifBlank {
            val err = OpsResult.Err(OpsError.unauthorized(Strings.t(Str.MissingOperatorPin)))
            _state.value = _state.value.copy(errorMessage = err.error.message)
            return err
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        val networkResult = if (open) {
            repository.openBatteryBox(task, pin, izBlue)
        } else {
            repository.closeBatteryBox(task, pin, izBlue)
        }
        return when (networkResult) {
            is OpsResult.Ok -> {
                if (open) {
                    patchLocalState(task.id) { it.copy(state = 1) }
                    _state.value = _state.value.copy(
                        loading = false,
                        message = Strings.t(Str.OpenBoxOk, task.carId),
                    )
                } else {
                    patchLocalState(task.id) { it.copy(state = 2, checkResult = 2) }
                    _state.value = _state.value.copy(
                        loading = false,
                        message = Strings.t(Str.FinishSwapOk, task.carId),
                    )
                }
                networkResult
            }
            is OpsResult.Err -> {
                if (!izBlue) {
                    val action = if (open) {
                        VehicleAction.OpenBatteryBox
                    } else {
                        VehicleAction.CloseBatteryBox
                    }
                    when (
                        val ble = control.execute(
                            vehicleId = task.carId,
                            action = action,
                            channel = ControlChannel.BleOnly,
                            imei = task.imei,
                        )
                    ) {
                        is OpsResult.Ok -> return controlBatteryBox(open = open, izBlue = true)
                        is OpsResult.Err -> {
                            val msg = ble.error.message.ifBlank { networkResult.error.message }
                            _state.value = _state.value.copy(
                                loading = false,
                                errorMessage = if (open) {
                                    Strings.t(Str.OpenBoxFailed, msg)
                                } else {
                                    Strings.t(Str.FinishSwapFailed, msg)
                                },
                            )
                            return ble
                        }
                    }
                }
                _state.value = _state.value.copy(
                    loading = false,
                    errorMessage = if (open) {
                        Strings.t(Str.OpenBoxFailed, networkResult.error.message)
                    } else {
                        Strings.t(Str.FinishSwapFailed, networkResult.error.message)
                    },
                )
                networkResult
            }
        }
    }

    fun clear() {
        _state.value = ChangeBatteryTaskUiState()
    }

    private fun requireSelected(): OpsTask? = _state.value.selected

    private fun missingSelection(): OpsResult.Err {
        val err = OpsResult.Err(OpsError.business("TASK_NONE", Strings.t(Str.SelectTaskFirst)))
        _state.value = _state.value.copy(errorMessage = err.error.message)
        return err
    }

    private fun patchLocalState(taskId: String, transform: (OpsTask) -> OpsTask) {
        val tasks = _state.value.tasks.map { if (it.id == taskId) transform(it) else it }
        _state.value = _state.value.copy(tasks = tasks)
    }
}
