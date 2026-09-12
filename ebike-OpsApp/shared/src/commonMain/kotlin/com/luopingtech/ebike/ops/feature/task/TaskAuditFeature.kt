package com.luopingtech.ebike.ops.feature.task

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.task.TaskAuditRepository
import com.luopingtech.ebike.ops.domain.model.OpsTask
import com.luopingtech.ebike.ops.domain.model.OpsTaskType
import com.luopingtech.ebike.ops.domain.model.TaskAuditResult
import com.luopingtech.ebike.ops.domain.model.canViewAuditResult
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class TaskAuditUiState(
    val loading: Boolean = false,
    val taskId: String? = null,
    val taskType: OpsTaskType? = null,
    val result: TaskAuditResult? = null,
    val errorMessage: String? = null,
)

class TaskAuditFeature(
    private val repository: TaskAuditRepository,
) {
    private val _state = MutableStateFlow(TaskAuditUiState())
    val state: StateFlow<TaskAuditUiState> = _state.asStateFlow()

    fun clear() {
        _state.value = TaskAuditUiState()
    }

    suspend fun load(task: OpsTask?): OpsResult<TaskAuditResult> {
        if (task == null || task.id.isBlank()) {
            val err = OpsResult.Err(OpsError.business("AUDIT", Strings.t(Str.AuditNeedTask)))
            _state.value = TaskAuditUiState(errorMessage = err.error.message)
            return err
        }
        if (!task.canViewAuditResult()) {
            val err = OpsResult.Err(OpsError.business("AUDIT", Strings.t(Str.AuditNotAvailable)))
            _state.value = TaskAuditUiState(errorMessage = err.error.message)
            return err
        }
        _state.value = TaskAuditUiState(
            loading = true,
            taskId = task.id,
            taskType = task.type,
        )
        return when (val result = repository.load(task.id, task.type)) {
            is OpsResult.Ok -> {
                _state.value = TaskAuditUiState(
                    loading = false,
                    taskId = task.id,
                    taskType = task.type,
                    result = result.value,
                )
                result
            }
            is OpsResult.Err -> {
                _state.value = TaskAuditUiState(
                    loading = false,
                    taskId = task.id,
                    taskType = task.type,
                    errorMessage = result.error.message,
                )
                result
            }
        }
    }
}
