package com.luopingtech.ebike.ops.feature.task

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.task.MoveCarTaskRepository
import com.luopingtech.ebike.ops.domain.control.ControlChannel
import com.luopingtech.ebike.ops.domain.control.VehicleAction
import com.luopingtech.ebike.ops.domain.control.VehicleControlPolicy
import com.luopingtech.ebike.ops.domain.geo.GeoMath
import com.luopingtech.ebike.ops.domain.model.OpsTask
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.scan.ScanCodeParser
import com.luopingtech.ebike.ops.domain.scan.ScanTarget
import com.luopingtech.ebike.ops.platform.DemoMediaUploader
import com.luopingtech.ebike.ops.platform.GeoPoint
import com.luopingtech.ebike.ops.platform.LocationTracker
import com.luopingtech.ebike.ops.platform.MediaUploader
import com.luopingtech.ebike.ops.platform.UnsupportedLocationTracker
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class MoveCarTaskUiState(
    val loading: Boolean = false,
    val tasks: List<OpsTask> = emptyList(),
    val selectedTaskId: String? = null,
    val serviceAreaId: String? = null,
    val photoUrls: List<String> = emptyList(),
    /** Legacy business code 23326 — finish requires photograph audit. */
    val needPhotograph: Boolean = false,
    val remark: String = "",
    val distanceMeters: Double? = null,
    val arrived: Boolean = false,
    val arrivalMaxMeters: Double = DEFAULT_ARRIVAL_METERS,
    val message: String? = null,
    val errorMessage: String? = null,
) {
    val selected: OpsTask?
        get() = tasks.firstOrNull { it.id == selectedTaskId }

    companion object {
        const val DEFAULT_ARRIVAL_METERS = 200.0
        const val CODE_NEED_PHOTOGRAPH = "23326"
    }
}

/**
 * Move-car task flow: list → claim → start → 到点 → finish (photo if 23326).
 */
class MoveCarTaskFeature(
    private val repository: MoveCarTaskRepository,
    private val control: VehicleControlPolicy,
    private val pinProvider: () -> String,
    private val locationTracker: LocationTracker = UnsupportedLocationTracker(),
    private val mediaUploader: MediaUploader = DemoMediaUploader(),
) {
    private val _state = MutableStateFlow(MoveCarTaskUiState())
    val state: StateFlow<MoveCarTaskUiState> = _state.asStateFlow()

    fun selectTask(taskId: String?) {
        _state.value = _state.value.copy(
            selectedTaskId = taskId,
            errorMessage = null,
            needPhotograph = false,
            remark = "",
            distanceMeters = null,
            arrived = false,
        )
    }

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
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.VehicleNotInTaskList))
            return false
        }
        _state.value = _state.value.copy(
            selectedTaskId = task.id,
            message = Strings.t(Str.SelectedScanVehicle, task.carId),
            errorMessage = null,
            needPhotograph = false,
            remark = "",
        )
        return true
    }

    fun setRemark(value: String) {
        _state.value = _state.value.copy(remark = value, errorMessage = null)
    }

    fun addPhotoUrl(url: String) {
        val trimmed = url.trim()
        if (trimmed.isBlank()) return
        _state.value = _state.value.copy(
            photoUrls = _state.value.photoUrls + trimmed,
            errorMessage = null,
        )
    }

    fun addDemoPhoto() {
        addPhotoUrl("demo://move/${_state.value.photoUrls.size + 1}")
    }

    fun removePhoto(url: String) {
        _state.value = _state.value.copy(
            photoUrls = _state.value.photoUrls.filterNot { it == url },
        )
    }

    fun clearPhotos() {
        _state.value = _state.value.copy(photoUrls = emptyList(), needPhotograph = false, remark = "")
    }

    suspend fun load(area: ServiceArea?) {
        if (area == null) {
            _state.value = MoveCarTaskUiState(errorMessage = Strings.t(Str.SelectServiceAreaFirst))
            return
        }
        _state.value = _state.value.copy(
            loading = true,
            errorMessage = null,
            message = null,
            serviceAreaId = area.id,
            needPhotograph = false,
        )
        when (val result = repository.list(area)) {
            is OpsResult.Ok -> {
                val tasks = result.value
                val keep = _state.value.selectedTaskId
                    ?.takeIf { id -> tasks.any { it.id == id } }
                _state.value = _state.value.copy(
                    loading = false,
                    tasks = tasks,
                    selectedTaskId = keep,
                    message = Strings.t(Str.MoveCarTasksCount, tasks.size),
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

    /**
     * Compare operator GPS with task pin. Unsupported location → skip gate (arrived=true).
     */
    suspend fun refreshArrival(maxMeters: Double = _state.value.arrivalMaxMeters): Boolean {
        val task = _state.value.selected
            ?: run {
                _state.value = _state.value.copy(
                    errorMessage = Strings.t(Str.SelectTaskFirst),
                    arrived = false,
                )
                return false
            }
        if (task.lat == 0.0 && task.lng == 0.0) {
            _state.value = _state.value.copy(
                arrived = true,
                distanceMeters = null,
                arrivalMaxMeters = maxMeters,
                message = Strings.t(Str.MoveSkipArrivalNoCoord),
                errorMessage = null,
            )
            return true
        }
        return when (val loc = locationTracker.currentLocation()) {
            is OpsResult.Err -> {
                if (loc.error.code == "UNSUPPORTED") {
                    _state.value = _state.value.copy(
                        arrived = true,
                        distanceMeters = null,
                        arrivalMaxMeters = maxMeters,
                        message = Strings.t(Str.MoveSkipArrivalNoLoc),
                        errorMessage = null,
                    )
                    true
                } else {
                    _state.value = _state.value.copy(
                        arrived = false,
                        distanceMeters = null,
                        errorMessage = Strings.t(Str.MoveLocFailed, loc.error.message),
                    )
                    false
                }
            }
            is OpsResult.Ok -> {
                val dist = GeoMath.distanceMeters(
                    loc.value,
                    GeoPoint(latitude = task.lat, longitude = task.lng),
                )
                val ok = dist <= maxMeters
                _state.value = _state.value.copy(
                    arrived = ok,
                    distanceMeters = dist,
                    arrivalMaxMeters = maxMeters,
                    message = if (ok) {
                        Strings.t(Str.MoveArrived, dist.toInt())
                    } else {
                        Strings.t(Str.MoveNotArrived, dist.toInt(), maxMeters.toInt())
                    },
                    errorMessage = if (ok) null else Strings.t(Str.MovePleaseApproach),
                )
                ok
            }
        }
    }

    suspend fun claimSelected(): OpsResult<Unit> {
        val task = requireSelected() ?: return missingSelection()
        val pin = requirePin() ?: return missingPin()
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        return when (val result = repository.claim(task, pin)) {
            is OpsResult.Ok -> {
                patchLocalState(task.id) { it.copy(state = 1) }
                _state.value = _state.value.copy(
                    loading = false,
                    message = Strings.t(Str.ActionOk, Strings.t(Str.Claim), task.carId),
                )
                result
            }
            is OpsResult.Err -> fail(result, Str.Claim)
        }
    }

    suspend fun startSelected(): OpsResult<Unit> {
        val task = requireSelected() ?: return missingSelection()
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        return when (val result = repository.start(task)) {
            is OpsResult.Ok -> {
                patchLocalState(task.id) { it.copy(state = 1) }
                _state.value = _state.value.copy(
                    loading = false,
                    message = Strings.t(Str.ActionOk, Strings.t(Str.Start), task.carId),
                )
                result
            }
            is OpsResult.Err -> fail(result, Str.Start)
        }
    }

    /**
     * Finish move-car. Legacy does not gate on arrival distance.
     * If server returns 23326, sets [needPhotograph].
     * When photos exist, uploads then retries finish with remote URLs.
     */
    suspend fun finishSelected(): OpsResult<Unit> {
        val task = requireSelected() ?: return missingSelection()
        if (_state.value.needPhotograph && _state.value.remark.trim().isEmpty()) {
            val err = OpsResult.Err(OpsError.business("REMARK", Strings.t(Str.PhotoRemarkRequired)))
            _state.value = _state.value.copy(errorMessage = err.error.message)
            return err
        }
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
        return when (val result = repository.finish(task, pictures, remark)) {
            is OpsResult.Ok -> {
                patchLocalState(task.id) { it.copy(state = 2, checkResult = 2) }
                _state.value = _state.value.copy(
                    loading = false,
                    needPhotograph = false,
                    photoUrls = emptyList(),
                    remark = "",
                    message = Strings.t(Str.MoveCarFinishedOne, task.carId),
                )
                result
            }
            is OpsResult.Err -> {
                if (result.error.code == MoveCarTaskUiState.CODE_NEED_PHOTOGRAPH) {
                    _state.value = _state.value.copy(
                        loading = false,
                        needPhotograph = true,
                        errorMessage = null,
                        message = Strings.t(Str.NeedPhotoAudit),
                    )
                    result
                } else {
                    fail(result, Str.Finish)
                }
            }
        }
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
            is OpsResult.Err -> fail(result, Str.Ring)
        }
    }

    fun clear() {
        _state.value = MoveCarTaskUiState()
    }

    private fun requireSelected(): OpsTask? = _state.value.selected

    private fun requirePin(): String? = pinProvider().takeIf { it.isNotBlank() }

    private fun missingSelection(): OpsResult.Err {
        val err = OpsResult.Err(OpsError.business("TASK_NONE", Strings.t(Str.SelectTaskFirst)))
        _state.value = _state.value.copy(errorMessage = err.error.message)
        return err
    }

    private fun missingPin(): OpsResult.Err {
        val err = OpsResult.Err(OpsError.unauthorized(Strings.t(Str.MissingOperatorPin)))
        _state.value = _state.value.copy(errorMessage = err.error.message)
        return err
    }

    private fun fail(result: OpsResult.Err, actionKey: Str): OpsResult.Err {
        _state.value = _state.value.copy(
            loading = false,
            errorMessage = Strings.t(Str.ActionFailed, Strings.t(actionKey), result.error.message),
        )
        return result
    }

    private fun patchLocalState(taskId: String, transform: (OpsTask) -> OpsTask) {
        val tasks = _state.value.tasks.map { if (it.id == taskId) transform(it) else it }
        _state.value = _state.value.copy(tasks = tasks)
    }
}
