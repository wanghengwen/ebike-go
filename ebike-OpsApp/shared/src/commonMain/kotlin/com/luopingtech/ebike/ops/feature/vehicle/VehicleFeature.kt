package com.luopingtech.ebike.ops.feature.vehicle

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.vehicle.VehicleRepository
import com.luopingtech.ebike.ops.data.vehicle.toMapPin
import com.luopingtech.ebike.ops.domain.model.MapPin
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.model.Vehicle
import com.luopingtech.ebike.ops.domain.vehicle.BatterySnNormalizer
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class VehicleUiState(
    val loading: Boolean = false,
    val detailLoading: Boolean = false,
    val vehicles: List<Vehicle> = emptyList(),
    val pins: List<MapPin> = emptyList(),
    val serviceAreaId: String? = null,
    val errorMessage: String? = null,
    val selectedCarId: String? = null,
)

class VehicleFeature(
    private val repository: VehicleRepository,
) {
    private val _state = MutableStateFlow(VehicleUiState())
    val state: StateFlow<VehicleUiState> = _state.asStateFlow()

    suspend fun loadForArea(area: ServiceArea?) {
        if (area == null) {
            _state.value = VehicleUiState(errorMessage = Strings.t(Str.SelectServiceAreaFirst))
            return
        }
        _state.value = _state.value.copy(
            loading = true,
            errorMessage = null,
            serviceAreaId = area.id,
        )
        when (val result = repository.loadByServiceArea(area)) {
            is OpsResult.Ok -> {
                val vehicles = result.value
                _state.value = VehicleUiState(
                    loading = false,
                    vehicles = vehicles,
                    pins = vehicles.map { it.toMapPin() },
                    serviceAreaId = area.id,
                    selectedCarId = _state.value.selectedCarId
                        ?.takeIf { id -> vehicles.any { it.carId == id } },
                )
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    errorMessage = result.error.message,
                )
            }
        }
    }

    fun selectVehicle(carId: String?) {
        _state.value = _state.value.copy(selectedCarId = carId)
    }

    /** Merge a known detail into the list and select it (e.g. detect → map jump). */
    fun upsertAndSelect(vehicle: Vehicle) {
        val id = vehicle.carId.trim()
        if (id.isBlank()) return
        val merged = _state.value.vehicles
            .map { if (it.carId.equals(id, ignoreCase = true)) vehicle else it }
            .let { list ->
                if (list.any { it.carId.equals(id, ignoreCase = true) }) list else list + vehicle
            }
        _state.value = _state.value.copy(
            vehicles = merged,
            pins = merged.map { it.toMapPin() },
            selectedCarId = vehicle.carId,
            errorMessage = null,
        )
    }

    /**
     * Pull device detail and merge into the in-memory list (keeps map pin selection).
     */
    suspend fun refreshDetail(carId: String): OpsResult<Vehicle> {
        val trimmed = carId.trim()
        if (trimmed.isBlank()) {
            return OpsResult.Err(
                com.luopingtech.ebike.ops.core.result.OpsError.business(
                    "CAR",
                    Strings.t(Str.EnterCarId),
                ),
            )
        }
        _state.value = _state.value.copy(detailLoading = true, errorMessage = null)
        return when (val result = repository.getDetail(trimmed)) {
            is OpsResult.Ok -> {
                val merged = _state.value.vehicles
                    .map { if (it.carId.equals(trimmed, ignoreCase = true)) result.value else it }
                    .let { list ->
                        if (list.any { it.carId.equals(trimmed, ignoreCase = true) }) {
                            list
                        } else {
                            list + result.value
                        }
                    }
                _state.value = _state.value.copy(
                    detailLoading = false,
                    vehicles = merged,
                    pins = merged.map { it.toMapPin() },
                    selectedCarId = result.value.carId,
                )
                result
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    detailLoading = false,
                    errorMessage = result.error.message,
                )
                result
            }
        }
    }

    /**
     * Legacy bindBatterySn — normalize QR then POST carInfo/bindBatterySn and refresh detail.
     */
    suspend fun bindBatterySn(carId: String, rawBatterySn: String): OpsResult<Vehicle> {
        val id = carId.trim()
        val sn = BatterySnNormalizer.normalize(rawBatterySn)
        if (id.isBlank()) {
            return OpsResult.Err(
                com.luopingtech.ebike.ops.core.result.OpsError.business(
                    "CAR",
                    Strings.t(Str.EnterCarId),
                ),
            )
        }
        if (!BatterySnNormalizer.isValid(rawBatterySn)) {
            return OpsResult.Err(
                com.luopingtech.ebike.ops.core.result.OpsError.business(
                    "BATTERY_SN",
                    Strings.t(Str.BindBatterySnInvalid),
                ),
            )
        }
        _state.value = _state.value.copy(detailLoading = true, errorMessage = null)
        return when (val bind = repository.bindBatterySn(id, sn)) {
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    detailLoading = false,
                    errorMessage = bind.error.message,
                )
                bind
            }
            is OpsResult.Ok -> refreshDetail(id)
        }
    }

    /**
     * Legacy CarDetailViewModel.unBindBattery: bindBatterySn(carId, "").
     */
    suspend fun unbindBatterySn(carId: String): OpsResult<Vehicle> {
        val id = carId.trim()
        if (id.isBlank()) {
            return OpsResult.Err(
                com.luopingtech.ebike.ops.core.result.OpsError.business(
                    "CAR",
                    Strings.t(Str.EnterCarId),
                ),
            )
        }
        _state.value = _state.value.copy(detailLoading = true, errorMessage = null)
        return when (val bind = repository.bindBatterySn(id, "")) {
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    detailLoading = false,
                    errorMessage = bind.error.message,
                )
                bind
            }
            is OpsResult.Ok -> refreshDetail(id)
        }
    }

    fun clear() {
        _state.value = VehicleUiState()
    }
}
