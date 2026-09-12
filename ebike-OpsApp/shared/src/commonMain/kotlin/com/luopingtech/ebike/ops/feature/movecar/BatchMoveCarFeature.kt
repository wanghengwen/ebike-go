package com.luopingtech.ebike.ops.feature.movecar

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.movecar.BatchMoveCarRepository
import com.luopingtech.ebike.ops.data.movecar.BatchMoveCarRepositoryImpl
import com.luopingtech.ebike.ops.domain.model.BatchMoveChild
import com.luopingtech.ebike.ops.platform.DemoMediaUploader
import com.luopingtech.ebike.ops.platform.MediaUploader
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class BatchMoveCarUiState(
    val loading: Boolean = false,
    val parentTaskId: String? = null,
    val children: List<BatchMoveChild> = emptyList(),
    val selectedTaskIds: Set<String> = emptySet(),
    val photoUrls: List<String> = emptyList(),
    val needPhotograph: Boolean = false,
    val remark: String = "",
    val message: String? = null,
    val errorMessage: String? = null,
)

/**
 * Man-made batch move: load children → select → start batch unlock → finish (photo if 23011).
 */
class BatchMoveCarFeature(
    private val repository: BatchMoveCarRepository,
    private val mediaUploader: MediaUploader = DemoMediaUploader(),
    private val pinProvider: () -> String = { "" },
) {
    private val _state = MutableStateFlow(BatchMoveCarUiState())
    val state: StateFlow<BatchMoveCarUiState> = _state.asStateFlow()

    fun toggleSelect(taskId: String) {
        val set = _state.value.selectedTaskIds.toMutableSet()
        if (!set.add(taskId)) set.remove(taskId)
        _state.value = _state.value.copy(selectedTaskIds = set)
    }

    fun selectPending() {
        _state.value = _state.value.copy(
            selectedTaskIds = _state.value.children
                .filterNot { it.isFinished }
                .map { it.taskId }
                .toSet(),
        )
    }

    fun clearSelection() {
        _state.value = _state.value.copy(selectedTaskIds = emptySet())
    }

    fun setRemark(value: String) {
        _state.value = _state.value.copy(remark = value, errorMessage = null)
    }

    fun addPhotoUrl(url: String) {
        val trimmed = url.trim()
        if (trimmed.isBlank()) return
        _state.value = _state.value.copy(photoUrls = _state.value.photoUrls + trimmed, errorMessage = null)
    }

    fun addDemoPhoto() {
        addPhotoUrl("demo://batchmove/${_state.value.photoUrls.size + 1}")
    }

    fun removePhoto(url: String) {
        _state.value = _state.value.copy(photoUrls = _state.value.photoUrls.filterNot { it == url })
    }

    fun clearPhotos() {
        _state.value = _state.value.copy(photoUrls = emptyList(), needPhotograph = false, remark = "")
    }

    suspend fun load(parentTaskId: String) {
        val id = parentTaskId.trim()
        if (id.isBlank()) {
            _state.value = BatchMoveCarUiState(errorMessage = Strings.t(Str.BatchMoveNeedParent))
            return
        }
        _state.value = _state.value.copy(
            loading = true,
            errorMessage = null,
            message = null,
            parentTaskId = id,
            needPhotograph = false,
        )
        when (val result = repository.listChildren(id)) {
            is OpsResult.Ok -> {
                val keep = _state.value.selectedTaskIds.filter { tid ->
                    result.value.any { it.taskId == tid && !it.isFinished }
                }.toSet()
                _state.value = _state.value.copy(
                    loading = false,
                    children = result.value,
                    selectedTaskIds = keep,
                    message = Strings.t(Str.BatchMoveChildrenCount, result.value.size),
                )
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(loading = false, errorMessage = result.error.message)
            }
        }
    }

    suspend fun startSelected(): OpsResult<Unit> {
        val ids = selectedActionableIds()
        if (ids.isEmpty()) {
            val err = OpsResult.Err(OpsError.business("BATCH_NONE", Strings.t(Str.BatchMoveNeedSelect)))
            _state.value = _state.value.copy(errorMessage = err.error.message)
            return err
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        return when (val result = repository.startBatch(ids, pinProvider())) {
            is OpsResult.Ok -> {
                val parent = _state.value.parentTaskId
                if (parent != null) load(parent) else {
                    _state.value = _state.value.copy(loading = false)
                }
                _state.value = _state.value.copy(
                    message = Strings.t(Str.BatchMoveStarted, ids.size),
                    errorMessage = null,
                )
                result
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    errorMessage = Strings.t(Str.ActionFailed, Strings.t(Str.Start), result.error.message),
                )
                result
            }
        }
    }

    suspend fun finishSelected(): OpsResult<Unit> {
        val ids = selectedActionableIds()
        if (ids.isEmpty()) {
            val err = OpsResult.Err(OpsError.business("BATCH_NONE", Strings.t(Str.BatchMoveNeedSelect)))
            _state.value = _state.value.copy(errorMessage = err.error.message)
            return err
        }
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
        return when (val result = repository.finishBatch(ids, pinProvider(), pictures, remark)) {
            is OpsResult.Ok -> {
                val parent = _state.value.parentTaskId
                _state.value = _state.value.copy(
                    selectedTaskIds = emptySet(),
                    photoUrls = emptyList(),
                    needPhotograph = false,
                    remark = "",
                )
                if (parent != null) {
                    load(parent)
                } else {
                    _state.value = _state.value.copy(loading = false)
                }
                _state.value = _state.value.copy(
                    message = Strings.t(Str.BatchMoveFinished, ids.size),
                    errorMessage = null,
                )
                result
            }
            is OpsResult.Err -> {
                if (result.error.code == BatchMoveCarRepositoryImpl.CODE_NEED_PHOTOGRAPH ||
                    result.error.code == FreeMoveCarUiState.CODE_NEED_PHOTOGRAPH
                ) {
                    _state.value = _state.value.copy(
                        loading = false,
                        needPhotograph = true,
                        errorMessage = null,
                        message = Strings.t(Str.NeedPhotoAudit),
                    )
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
        _state.value = BatchMoveCarUiState()
    }

    private fun selectedActionableIds(): List<String> {
        val finished = _state.value.children.filter { it.isFinished }.map { it.taskId }.toSet()
        return _state.value.selectedTaskIds.filterNot { it in finished }
    }
}
