package com.luopingtech.ebike.ops.feature.tools

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.tools.UnlockedVehicleRepository
import com.luopingtech.ebike.ops.domain.control.ControlChannel
import com.luopingtech.ebike.ops.domain.control.VehicleAction
import com.luopingtech.ebike.ops.domain.control.VehicleControlPolicy
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.model.UnlockedVehicle
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class UnlockedVehicleUiState(
    val loading: Boolean = false,
    val lockingCarId: String? = null,
    val query: String = "",
    /** Legacy 122802 — when false, search is hidden and query is forced to self phone. */
    val canFilterStaff: Boolean = true,
    val serviceAreaId: String? = null,
    val vehicles: List<UnlockedVehicle> = emptyList(),
    val message: String? = null,
    val errorMessage: String? = null,
)

/**
 * Unlocked-vehicle report (legacy UnLockedVehiclesActivity): list + network lock.
 */
class UnlockedVehicleFeature(
    private val repository: UnlockedVehicleRepository,
    private val control: VehicleControlPolicy,
    private val canFilterStaffProvider: () -> Boolean = { true },
    private val selfPhoneProvider: () -> String = { "" },
) {
    private val _state = MutableStateFlow(
        UnlockedVehicleUiState(canFilterStaff = canFilterStaffProvider()),
    )
    val state: StateFlow<UnlockedVehicleUiState> = _state.asStateFlow()

    fun setQuery(value: String) {
        if (!_state.value.canFilterStaff) return
        _state.value = _state.value.copy(query = value, errorMessage = null)
    }

    fun clear() {
        _state.value = UnlockedVehicleUiState(canFilterStaff = canFilterStaffProvider())
    }

    suspend fun load(area: ServiceArea?) {
        if (area == null) {
            _state.value = UnlockedVehicleUiState(
                canFilterStaff = canFilterStaffProvider(),
                errorMessage = Strings.t(Str.SelectServiceAreaFirst),
            )
            return
        }
        val canFilter = canFilterStaffProvider()
        val effectiveQuery = if (canFilter) {
            _state.value.query
        } else {
            selfPhoneProvider().trim()
        }
        _state.value = _state.value.copy(
            loading = true,
            errorMessage = null,
            message = null,
            serviceAreaId = area.id,
            canFilterStaff = canFilter,
            query = if (canFilter) _state.value.query else effectiveQuery,
        )
        when (val result = repository.list(area.id, effectiveQuery)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    vehicles = result.value,
                    message = Strings.t(Str.UnlockedVehiclesCount, result.value.size),
                )
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(loading = false, errorMessage = result.error.message)
            }
        }
    }

    suspend fun lock(carId: String): OpsResult<Unit> {
        val trimmed = carId.trim()
        if (trimmed.isBlank()) {
            val err = OpsResult.Err(OpsError.business("CAR", Strings.t(Str.EnterCarId)))
            _state.value = _state.value.copy(errorMessage = err.error.message)
            return err
        }
        _state.value = _state.value.copy(
            lockingCarId = trimmed,
            loading = true,
            errorMessage = null,
            message = null,
        )
        return when (
            val result = control.execute(
                vehicleId = trimmed,
                action = VehicleAction.Lock,
                channel = ControlChannel.NetworkOnly,
            )
        ) {
            is OpsResult.Ok -> {
                repository.noteLocked(trimmed)
                val remaining = _state.value.vehicles.filterNot { it.carId.equals(trimmed, ignoreCase = true) }
                _state.value = _state.value.copy(
                    loading = false,
                    lockingCarId = null,
                    vehicles = remaining,
                    message = Strings.t(Str.UnlockedLockOk, trimmed),
                )
                result
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    lockingCarId = null,
                    errorMessage = Strings.t(Str.UnlockedLockFailed, result.error.message),
                )
                result
            }
        }
    }
}
