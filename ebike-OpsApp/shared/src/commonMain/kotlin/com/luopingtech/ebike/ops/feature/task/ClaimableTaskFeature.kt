package com.luopingtech.ebike.ops.feature.task

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.task.InspectionTaskRepository
import com.luopingtech.ebike.ops.data.task.RepairTaskRepository
import com.luopingtech.ebike.ops.domain.model.OpsTask
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.scan.ScanCodeParser
import com.luopingtech.ebike.ops.domain.scan.ScanTarget
import com.luopingtech.ebike.ops.platform.DemoMediaUploader
import com.luopingtech.ebike.ops.platform.MediaUploader
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

enum class ClaimableKind {
    Inspection,
    Repair,
}

data class ClaimableTaskUiState(
    val loading: Boolean = false,
    val tasks: List<OpsTask> = emptyList(),
    val selectedTaskId: String? = null,
    val serviceAreaId: String? = null,
    val photoUrls: List<String> = emptyList(),
    /** Align with legacy photograph audit (business 23326). */
    val needPhotograph: Boolean = false,
    /** PhotographAuditActivity description — required when [needPhotograph]. */
    val remark: String = "",
    /** DragBackTaskActivity fields (repair only). */
    val dragReason: String = "",
    val dragAddress: String = "",
    val message: String? = null,
    val errorMessage: String? = null,
) {
    val selected: OpsTask?
        get() = tasks.firstOrNull { it.id == selectedTaskId }

    companion object {
        const val CODE_NEED_PHOTOGRAPH = "23326"
    }
}

/**
 * Shared claim → start → finish shell for inspection and repair (with photo audit).
 */
class ClaimableTaskFeature(
    private val list: suspend (ServiceArea, String) -> OpsResult<List<OpsTask>>,
    private val claim: suspend (OpsTask, String) -> OpsResult<Unit>,
    private val start: suspend (OpsTask, String) -> OpsResult<Unit>,
    private val finish: suspend (OpsTask, List<String>, String?) -> OpsResult<Unit>,
    private val pinProvider: () -> String,
    private val kind: ClaimableKind,
    private val mediaUploader: MediaUploader = DemoMediaUploader(),
    private val createDrag: (suspend (OpsTask, String, String) -> OpsResult<Unit>)? = null,
    private val assign: (suspend (List<OpsTask>, String) -> OpsResult<Unit>)? = null,
) {
    private val _state = MutableStateFlow(ClaimableTaskUiState())
    val state: StateFlow<ClaimableTaskUiState> = _state.asStateFlow()

    fun selectTask(taskId: String?) {
        _state.value = _state.value.copy(
            selectedTaskId = taskId,
            errorMessage = null,
            needPhotograph = false,
            remark = "",
            dragReason = "",
            dragAddress = "",
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

    fun setDragReason(value: String) {
        _state.value = _state.value.copy(dragReason = value, errorMessage = null)
    }

    fun setDragAddress(value: String) {
        _state.value = _state.value.copy(dragAddress = value, errorMessage = null)
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
        addPhotoUrl("demo://${kind.name.lowercase()}/${_state.value.photoUrls.size + 1}")
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
            _state.value = ClaimableTaskUiState(errorMessage = Strings.t(Str.SelectServiceAreaFirst))
            return
        }
        val pin = pinProvider()
        _state.value = _state.value.copy(
            loading = true,
            errorMessage = null,
            message = null,
            serviceAreaId = area.id,
            needPhotograph = false,
        )
        when (val result = list(area, pin)) {
            is OpsResult.Ok -> {
                val tasks = result.value
                val keep = _state.value.selectedTaskId
                    ?.takeIf { id -> tasks.any { it.id == id } }
                val countMsg = when (kind) {
                    ClaimableKind.Inspection -> Strings.t(Str.InspectionTasksCount, tasks.size)
                    ClaimableKind.Repair -> Strings.t(Str.RepairTasksCount, tasks.size)
                }
                _state.value = _state.value.copy(
                    loading = false,
                    tasks = tasks,
                    selectedTaskId = keep,
                    message = countMsg,
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

    suspend fun claimSelected(): OpsResult<Unit> = act(Str.Claim) { task, pin ->
        claim(task, pin).also {
            if (it.isOk) patchLocalState(task.id) { t -> t.copy(state = 1) }
        }
    }

    /** 批量领取（巡检待领取勾选底栏）。 */
    suspend fun claimMany(taskIds: List<String>): OpsResult<Unit> {
        if (taskIds.isEmpty()) {
            return OpsResult.Err(OpsError.business("NO_SEL", Strings.t(Str.SelectTaskFirst)))
        }
        val pin = pinProvider()
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        var lastErr: OpsResult.Err? = null
        for (id in taskIds) {
            val task = _state.value.tasks.firstOrNull { it.id == id } ?: continue
            when (val r = claim(task, pin)) {
                is OpsResult.Ok -> patchLocalState(id) { t -> t.copy(state = 1) }
                is OpsResult.Err -> lastErr = r
            }
        }
        _state.value = _state.value.copy(loading = false)
        return lastErr ?: OpsResult.Ok(Unit).also {
            _state.value = _state.value.copy(message = Strings.t(Str.ClaimOk, taskIds.size))
        }
    }

    /** 指派给员工 pin（type=2）。 */
    suspend fun assignMany(taskIds: List<String>, assigneePin: String): OpsResult<Unit> {
        val assignFn = assign
            ?: return OpsResult.Err(OpsError.business("NO_ASSIGN", Strings.t(Str.FeatureComingSoon)))
        if (taskIds.isEmpty()) {
            return OpsResult.Err(OpsError.business("NO_SEL", Strings.t(Str.SelectTaskFirst)))
        }
        if (assigneePin.isBlank()) {
            return OpsResult.Err(OpsError.business("NO_PIN", Strings.t(Str.SelectStaffFirst)))
        }
        val tasks = taskIds.mapNotNull { id -> _state.value.tasks.firstOrNull { it.id == id } }
        if (tasks.isEmpty()) {
            return OpsResult.Err(OpsError.business("NO_SEL", Strings.t(Str.SelectTaskFirst)))
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        return when (val r = assignFn(tasks, assigneePin)) {
            is OpsResult.Ok -> {
                tasks.forEach { patchLocalState(it.id) { t -> t.copy(state = 1) } }
                _state.value = _state.value.copy(
                    loading = false,
                    message = Strings.t(Str.AssignOk),
                )
                r
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(loading = false, errorMessage = r.error.message)
                r
            }
        }
    }

    suspend fun startSelected(): OpsResult<Unit> = act(Str.Start) { task, pin ->
        start(task, pin).also {
            if (it.isOk) patchLocalState(task.id) { t -> t.copy(state = 1) }
        }
    }

    suspend fun finishSelected(): OpsResult<Unit> {
        val task = requireSelected() ?: return missingSelection()
        if (kind == ClaimableKind.Repair && task.isDragBacking) {
            val err = OpsResult.Err(
                OpsError.business("DRAG_BLOCK", Strings.t(Str.DragBackBlockFinish)),
            )
            _state.value = _state.value.copy(errorMessage = err.error.message)
            return err
        }
        requireAuditRemark()?.let { return it }
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
        return when (val result = finish(task, pictures, remark)) {
            is OpsResult.Ok -> {
                patchLocalState(task.id) { it.copy(state = 2, checkResult = 2) }
                _state.value = _state.value.copy(
                    loading = false,
                    needPhotograph = false,
                    photoUrls = emptyList(),
                    remark = "",
                    message = Strings.t(Str.FinishOk, task.carId),
                )
                result
            }
            is OpsResult.Err -> {
                if (result.error.code == ClaimableTaskUiState.CODE_NEED_PHOTOGRAPH) {
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

    /** Legacy PhotographAudit: description required once photo audit is active. */
    private fun requireAuditRemark(): OpsResult.Err? {
        if (!_state.value.needPhotograph) return null
        if (_state.value.remark.trim().isNotEmpty()) return null
        val err = OpsResult.Err(OpsError.business("REMARK", Strings.t(Str.PhotoRemarkRequired)))
        _state.value = _state.value.copy(errorMessage = err.error.message)
        return err
    }

    suspend fun createDragSelected(): OpsResult<Unit> {
        val drag = createDrag
        if (kind != ClaimableKind.Repair || drag == null) {
            val err = OpsResult.Err(
                OpsError.unsupported("createDrag"),
            )
            _state.value = _state.value.copy(errorMessage = err.error.message)
            return err
        }
        val task = requireSelected() ?: return missingSelection()
        when (task.dragState) {
            2 -> {
                val err = OpsResult.Err(
                    OpsError.business("DRAG_BUSY", Strings.t(Str.DragBackInProgress)),
                )
                _state.value = _state.value.copy(errorMessage = err.error.message)
                return err
            }
            3 -> {
                val err = OpsResult.Err(
                    OpsError.business("DRAG_DONE", Strings.t(Str.DragBackDone)),
                )
                _state.value = _state.value.copy(errorMessage = err.error.message)
                return err
            }
        }
        val reason = _state.value.dragReason.trim()
        val address = _state.value.dragAddress.trim()
        if (reason.isEmpty()) {
            val err = OpsResult.Err(OpsError.business("DRAG_REASON", Strings.t(Str.DragBackReasonRequired)))
            _state.value = _state.value.copy(errorMessage = err.error.message)
            return err
        }
        if (address.isEmpty()) {
            val err = OpsResult.Err(OpsError.business("DRAG_ADDR", Strings.t(Str.DragBackAddressRequired)))
            _state.value = _state.value.copy(errorMessage = err.error.message)
            return err
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        return when (val result = drag(task, reason, address)) {
            is OpsResult.Ok -> {
                patchLocalState(task.id) { it.copy(dragState = 2, state = 1) }
                _state.value = _state.value.copy(
                    loading = false,
                    dragReason = "",
                    dragAddress = "",
                    message = Strings.t(Str.DragBackOk, task.carId),
                )
                result
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    errorMessage = Strings.t(
                        Str.ActionFailed,
                        Strings.t(Str.DragBack),
                        result.error.message,
                    ),
                )
                result
            }
        }
    }

    fun clear() {
        _state.value = ClaimableTaskUiState()
    }

    private suspend fun act(
        actionKey: Str,
        block: suspend (OpsTask, String) -> OpsResult<Unit>,
    ): OpsResult<Unit> {
        val task = requireSelected() ?: return missingSelection()
        val pin = pinProvider().takeIf { it.isNotBlank() } ?: return missingPin()
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        val label = Strings.t(actionKey)
        return when (val result = block(task, pin)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    message = Strings.t(Str.ActionOk, label, task.carId),
                )
                result
            }
            is OpsResult.Err -> fail(result, actionKey)
        }
    }

    private fun requireSelected(): OpsTask? = _state.value.selected

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

fun InspectionTaskFeature(
    repository: InspectionTaskRepository,
    pinProvider: () -> String,
    mediaUploader: MediaUploader = DemoMediaUploader(),
): ClaimableTaskFeature = ClaimableTaskFeature(
    list = { area, pin -> repository.list(area, pin) },
    claim = { task, pin -> repository.claim(task, pin) },
    start = { task, pin -> repository.start(task, pin) },
    finish = { task, pictures, remark -> repository.finish(task, pictures, remark) },
    pinProvider = pinProvider,
    kind = ClaimableKind.Inspection,
    mediaUploader = mediaUploader,
    assign = { tasks, pin -> repository.assign(tasks, pin) },
)

fun RepairTaskFeature(
    repository: RepairTaskRepository,
    pinProvider: () -> String,
    mediaUploader: MediaUploader = DemoMediaUploader(),
): ClaimableTaskFeature = ClaimableTaskFeature(
    list = { area, pin -> repository.list(area, pin) },
    claim = { task, pin -> repository.claim(task, pin) },
    start = { task, pin -> repository.start(task, pin) },
    finish = { task, pictures, remark -> repository.finish(task, pictures, remark) },
    pinProvider = pinProvider,
    kind = ClaimableKind.Repair,
    mediaUploader = mediaUploader,
    createDrag = { task, reason, address -> repository.createDrag(task, reason, address) },
    assign = { tasks, pin -> repository.assign(tasks, pin) },
)
