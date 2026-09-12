package com.luopingtech.ebike.ops.feature.tag

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.tag.VehicleTagRepository
import com.luopingtech.ebike.ops.data.vehicle.VehicleRepository
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.scan.ScanCodeParser
import com.luopingtech.ebike.ops.domain.scan.ScanTarget
import com.luopingtech.ebike.ops.domain.tag.VehicleTagType
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class PendingTagCar(
    val carId: String,
    val selected: Boolean = true,
)

data class VehicleTagUiState(
    val loading: Boolean = false,
    val submitting: Boolean = false,
    val carId: String = "",
    val types: List<VehicleTagType> = emptyList(),
    val pending: List<PendingTagCar> = emptyList(),
    val showTypePicker: Boolean = false,
    val message: String? = null,
    val errorMessage: String? = null,
)

/**
 * Aligns with legacy VehicleTagActivity: local pending list → pick type → multipart add.
 * Does not depend on record/page (legacy marks that API unused).
 */
class VehicleTagFeature(
    private val repository: VehicleTagRepository,
    private val vehicleRepository: VehicleRepository? = null,
    private val qrHostsProvider: () -> List<String> = { emptyList() },
) {
    private val _state = MutableStateFlow(VehicleTagUiState())
    val state: StateFlow<VehicleTagUiState> = _state.asStateFlow()

    fun clear() {
        _state.value = VehicleTagUiState()
    }

    fun setCarId(value: String) {
        _state.value = _state.value.copy(carId = value.take(20), errorMessage = null)
    }

    fun applyScan(raw: String) {
        val parsed = ScanCodeParser.parse(raw, qrHostsProvider())
        val carId = when (parsed) {
            is ScanTarget.CarId -> parsed.value
            is ScanTarget.Imei -> parsed.value
            null -> raw.trim()
        }
        if (carId.isNotBlank()) {
            _state.value = _state.value.copy(carId = carId, errorMessage = null)
        }
    }

    fun togglePending(carId: String) {
        _state.value = _state.value.copy(
            pending = _state.value.pending.map {
                if (it.carId == carId) it.copy(selected = !it.selected) else it
            },
        )
    }

    fun selectAllPending(selected: Boolean) {
        _state.value = _state.value.copy(
            pending = _state.value.pending.map { it.copy(selected = selected) },
        )
    }

    fun removePending(carId: String) {
        _state.value = _state.value.copy(
            pending = _state.value.pending.filterNot { it.carId == carId },
        )
    }

    fun openTypePicker(open: Boolean) {
        _state.value = _state.value.copy(showTypePicker = open, errorMessage = null)
    }

    suspend fun load(area: ServiceArea?) {
        if (area == null) {
            _state.value = _state.value.copy(
                loading = false,
                errorMessage = Strings.t(Str.SelectServiceAreaFirst),
            )
            return
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null)
        when (val r = repository.listTypes(area.id)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(
                loading = false,
                types = r.value,
            )
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                errorMessage = r.error.message,
            )
        }
    }

    /** Legacy: validate car exists then append to local list. */
    suspend fun addToPending(area: ServiceArea?) {
        val snap = _state.value
        val carId = snap.carId.trim()
        if (area == null) {
            _state.value = snap.copy(errorMessage = Strings.t(Str.SelectServiceAreaFirst))
            return
        }
        if (carId.isBlank()) {
            _state.value = snap.copy(errorMessage = Strings.t(Str.EnterCarIdOrScan))
            return
        }
        if (snap.pending.any { it.carId.equals(carId, ignoreCase = true) }) {
            _state.value = snap.copy(errorMessage = Strings.t(Str.VehicleTagAlreadyAdded), carId = "")
            return
        }
        _state.value = snap.copy(loading = true, errorMessage = null, message = null)
        val resolved = resolveCarId(carId)
        when (resolved) {
            is OpsResult.Err -> {
                _state.value = _state.value.copy(loading = false, errorMessage = resolved.error.message)
                return
            }
            is OpsResult.Ok -> {
                val id = resolved.value
                _state.value = _state.value.copy(
                    loading = false,
                    carId = "",
                    pending = snap.pending + PendingTagCar(carId = id, selected = true),
                    message = Strings.t(Str.VehicleTagPendingAdded, id),
                )
            }
        }
    }

    suspend fun submitSelected(area: ServiceArea?, typeId: String) {
        val snap = _state.value
        val selected = snap.pending.filter { it.selected }.map { it.carId }
        if (area == null) {
            _state.value = snap.copy(errorMessage = Strings.t(Str.SelectServiceAreaFirst))
            return
        }
        if (selected.isEmpty()) {
            _state.value = snap.copy(errorMessage = Strings.t(Str.VehicleTagNeedSelect))
            return
        }
        if (typeId.isBlank()) {
            _state.value = snap.copy(errorMessage = Strings.t(Str.VehicleTagType))
            return
        }
        _state.value = snap.copy(submitting = true, errorMessage = null, message = null, showTypePicker = false)
        when (val result = repository.addRecords(area.id, selected, typeId)) {
            is OpsResult.Ok -> {
                val remain = snap.pending.filterNot { it.selected }
                _state.value = _state.value.copy(
                    submitting = false,
                    pending = remain,
                    message = Strings.t(Str.VehicleTagSubmitOk, selected.size),
                )
            }
            is OpsResult.Err -> _state.value = _state.value.copy(
                submitting = false,
                errorMessage = result.error.message,
            )
        }
    }

    private suspend fun resolveCarId(raw: String): OpsResult<String> {
        val repo = vehicleRepository ?: return OpsResult.Ok(raw)
        val target = if (raw.all { it.isDigit() } && raw.length >= 15) {
            ScanTarget.Imei(raw)
        } else {
            ScanTarget.CarId(raw)
        }
        return when (val detail = repo.findByScanTarget(target)) {
            is OpsResult.Ok -> {
                val id = detail.value.carId.ifBlank { raw }
                OpsResult.Ok(id)
            }
            is OpsResult.Err -> detail
        }
    }
}
