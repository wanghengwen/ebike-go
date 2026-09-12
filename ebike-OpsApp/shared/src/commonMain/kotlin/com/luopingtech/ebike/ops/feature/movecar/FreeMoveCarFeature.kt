package com.luopingtech.ebike.ops.feature.movecar

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.movecar.FreeMoveCarRepository
import com.luopingtech.ebike.ops.data.staff.ServiceUserRepository
import com.luopingtech.ebike.ops.domain.model.FreeMoveCar
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.model.TeamWorker
import com.luopingtech.ebike.ops.domain.scan.ScanCodeParser
import com.luopingtech.ebike.ops.domain.scan.ScanTarget
import com.luopingtech.ebike.ops.platform.DemoMediaUploader
import com.luopingtech.ebike.ops.platform.MediaUploader
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class FreeMoveCarUiState(
    val loading: Boolean = false,
    val cars: List<FreeMoveCar> = emptyList(),
    val selectedCarIds: Set<String> = emptySet(),
    val carInput: String = "",
    val pushMode: Boolean = false,
    val serviceAreaId: String? = null,
    val photoUrls: List<String> = emptyList(),
    val needPhotograph: Boolean = false,
    val remark: String = "",
    /** Candidates from listByServiceIds (optional multi-select). */
    val teamCandidates: List<TeamWorker> = emptyList(),
    val selectedTeamKeys: Set<String> = emptySet(),
    val message: String? = null,
    val errorMessage: String? = null,
) {
    companion object {
        const val CODE_NEED_PHOTOGRAPH = "23326"
    }

    val selectedTeamWorkers: List<TeamWorker>
        get() = teamCandidates.filter { it.selectionKey in selectedTeamKeys }
}

/**
 * Free move-car: scan/add → in-progress list → multi-select finish (photo if 23326).
 * PhotographAudit for free-move also allows optional [TeamWorker] co-operators.
 * Remark is optional (legacy PHOTOGRAPH_MOVE_CAR); phone is required on finish body.
 */
class FreeMoveCarFeature(
    private val repository: FreeMoveCarRepository,
    private val serviceUserRepository: ServiceUserRepository? = null,
    private val mediaUploader: MediaUploader = DemoMediaUploader(),
    private val phoneProvider: () -> String = { "" },
) {
    private val _state = MutableStateFlow(FreeMoveCarUiState())
    val state: StateFlow<FreeMoveCarUiState> = _state.asStateFlow()

    fun setCarInput(value: String) {
        _state.value = _state.value.copy(carInput = value, errorMessage = null)
    }

    fun setPushMode(enabled: Boolean) {
        _state.value = _state.value.copy(pushMode = enabled)
    }

    fun toggleSelect(carId: String) {
        val set = _state.value.selectedCarIds.toMutableSet()
        if (!set.add(carId)) set.remove(carId)
        _state.value = _state.value.copy(selectedCarIds = set)
    }

    fun selectAll() {
        _state.value = _state.value.copy(
            selectedCarIds = _state.value.cars.map { it.carId }.toSet(),
        )
    }

    fun clearSelection() {
        _state.value = _state.value.copy(selectedCarIds = emptySet())
    }

    fun setRemark(value: String) {
        _state.value = _state.value.copy(remark = value, errorMessage = null)
    }

    fun toggleTeamWorker(key: String) {
        val set = _state.value.selectedTeamKeys.toMutableSet()
        if (!set.add(key)) set.remove(key)
        _state.value = _state.value.copy(selectedTeamKeys = set, errorMessage = null)
    }

    fun addPhotoUrl(url: String) {
        val trimmed = url.trim()
        if (trimmed.isBlank()) return
        _state.value = _state.value.copy(photoUrls = _state.value.photoUrls + trimmed, errorMessage = null)
    }

    fun addDemoPhoto() {
        addPhotoUrl("demo://freemove/${_state.value.photoUrls.size + 1}")
    }

    fun removePhoto(url: String) {
        _state.value = _state.value.copy(photoUrls = _state.value.photoUrls.filterNot { it == url })
    }

    fun clearPhotos() {
        _state.value = _state.value.copy(
            photoUrls = emptyList(),
            needPhotograph = false,
            remark = "",
            selectedTeamKeys = emptySet(),
        )
    }

    suspend fun load(area: ServiceArea?) {
        if (area == null) {
            _state.value = FreeMoveCarUiState(errorMessage = Strings.t(Str.SelectServiceAreaFirst))
            return
        }
        _state.value = _state.value.copy(
            loading = true,
            errorMessage = null,
            message = null,
            serviceAreaId = area.id,
            needPhotograph = false,
            teamCandidates = emptyList(),
            selectedTeamKeys = emptySet(),
        )
        when (val result = repository.list(area.id)) {
            is OpsResult.Ok -> {
                val keep = _state.value.selectedCarIds.filter { id -> result.value.any { it.carId == id } }.toSet()
                _state.value = _state.value.copy(
                    loading = false,
                    cars = result.value,
                    selectedCarIds = keep,
                    message = Strings.t(Str.FreeMoveInProgress, result.value.size),
                )
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(loading = false, errorMessage = result.error.message)
            }
        }
    }

    /** Load service-area staff for optional teamWorker multi-select (legacy getHelperStaffList). */
    suspend fun ensureTeamWorkersLoaded() {
        val serviceId = _state.value.serviceAreaId ?: return
        val repo = serviceUserRepository ?: return
        if (_state.value.teamCandidates.isNotEmpty()) return
        when (val result = repo.listTeamWorkers(serviceId)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(teamCandidates = result.value)
            }
            is OpsResult.Err -> {
                // Optional UI — keep finish usable; surface soft message.
                _state.value = _state.value.copy(
                    message = Strings.t(Str.TeamWorkerLoadFailed, result.error.message),
                )
            }
        }
    }

    /**
     * Parse QR / carId / IMEI then start move for that vehicle.
     */
    suspend fun addByScanRaw(raw: String, area: ServiceArea?, qrHosts: List<String> = emptyList()): Boolean {
        val target = ScanCodeParser.parse(raw, qrHosts)
            ?: run {
                _state.value = _state.value.copy(errorMessage = Strings.t(Str.CannotParseScan))
                return false
            }
        val carId = when (target) {
            is ScanTarget.CarId -> target.value
            is ScanTarget.Imei -> target.value
        }
        return addByCarId(carId, area)
    }

    suspend fun addByCarId(carId: String, area: ServiceArea?): Boolean {
        val trimmed = carId.trim()
        if (trimmed.isBlank()) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.EnterCarId))
            return false
        }
        if (area == null) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.SelectServiceAreaFirst))
            return false
        }
        if (_state.value.cars.any { it.carId.equals(trimmed, ignoreCase = true) }) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.FreeMoveAlreadyInList), carInput = "")
            return false
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        when (val perm = repository.checkPermission(trimmed, area.id)) {
            is OpsResult.Err -> {
                _state.value = _state.value.copy(loading = false, errorMessage = perm.error.message)
                return false
            }
            is OpsResult.Ok -> Unit
        }
        return when (val started = repository.start(listOf(trimmed), area.id, _state.value.pushMode)) {
            is OpsResult.Ok -> {
                val merged = (_state.value.cars + started.value)
                    .distinctBy { it.carId }
                _state.value = _state.value.copy(
                    loading = false,
                    cars = merged,
                    carInput = "",
                    selectedCarIds = _state.value.selectedCarIds + trimmed,
                    message = Strings.t(Str.FreeMoveJoined, trimmed),
                )
                true
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(loading = false, errorMessage = started.error.message)
                false
            }
        }
    }

    suspend fun removeCar(carId: String) {
        val serviceId = _state.value.serviceAreaId
        if (serviceId.isNullOrBlank()) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.SelectServiceAreaFirst))
            return
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        when (val result = repository.remove(carId, serviceId)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    cars = _state.value.cars.filterNot { it.carId == carId },
                    selectedCarIds = _state.value.selectedCarIds - carId,
                    message = Strings.t(Str.FreeMoveRemoved, carId),
                )
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(loading = false, errorMessage = result.error.message)
            }
        }
    }

    suspend fun finishSelected(): OpsResult<Unit> {
        val serviceId = _state.value.serviceAreaId
        if (serviceId.isNullOrBlank()) {
            val err = OpsResult.Err(OpsError.business("AREA", Strings.t(Str.SelectServiceAreaFirst)))
            _state.value = _state.value.copy(errorMessage = err.error.message)
            return err
        }
        val carIds = _state.value.selectedCarIds.toList()
        if (carIds.isEmpty()) {
            val err = OpsResult.Err(OpsError.business("MOVE_NONE", Strings.t(Str.FreeMoveNeedSelect)))
            _state.value = _state.value.copy(errorMessage = err.error.message)
            return err
        }
        // Legacy PHOTOGRAPH_MOVE_CAR: photos required after 23326; remark is optional.
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)

        val pictures = if (_state.value.photoUrls.isEmpty()) {
            emptyList()
        } else {
            when (val uploaded = mediaUploader.upload(_state.value.photoUrls)) {
                is OpsResult.Ok -> uploaded.value
                is OpsResult.Err -> {
                    _state.value = _state.value.copy(
                        loading = false,
                        errorMessage = Strings.t(Str.PhotoUploadFailed, uploaded.error.message),
                    )
                    return uploaded
                }
            }
        }

        val remark = _state.value.remark.trim().takeIf { it.isNotEmpty() }
        val teamWorkers = _state.value.selectedTeamWorkers
        val phone = phoneProvider()
        return when (val result = repository.finish(carIds, phone, pictures, remark, teamWorkers)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    cars = _state.value.cars.filterNot { it.carId in carIds },
                    selectedCarIds = emptySet(),
                    photoUrls = emptyList(),
                    remark = "",
                    needPhotograph = false,
                    selectedTeamKeys = emptySet(),
                    message = Strings.t(Str.FreeMoveFinished, carIds.size),
                )
                result
            }
            is OpsResult.Err -> {
                if (result.error.code == FreeMoveCarUiState.CODE_NEED_PHOTOGRAPH) {
                    _state.value = _state.value.copy(
                        loading = false,
                        needPhotograph = true,
                        errorMessage = null,
                        message = Strings.t(Str.NeedPhotoAudit),
                    )
                    ensureTeamWorkersLoaded()
                    result
                } else {
                    _state.value = _state.value.copy(
                        loading = false,
                        errorMessage = Strings.t(
                            Str.ActionFailed,
                            Strings.t(Str.Finish),
                            result.error.message,
                        ),
                    )
                    result
                }
            }
        }
    }

    fun clear() {
        _state.value = FreeMoveCarUiState()
    }
}
