package com.luopingtech.ebike.ops.feature.admin

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.admin.AdminRepository
import com.luopingtech.ebike.ops.domain.admin.IdBindAuditItem
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class IdBindAuditUiState(
    val loading: Boolean = false,
    val items: List<IdBindAuditItem> = emptyList(),
    val message: String? = null,
    val errorMessage: String? = null,
)

class IdBindAuditFeature(
    private val repository: AdminRepository,
) {
    private val _state = MutableStateFlow(IdBindAuditUiState())
    val state: StateFlow<IdBindAuditUiState> = _state.asStateFlow()

    fun clear() {
        _state.value = IdBindAuditUiState()
    }

    suspend fun load(area: ServiceArea?) {
        if (area == null) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.SelectServiceAreaFirst))
            return
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null)
        when (val result = repository.changeBindPage(area.id)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(loading = false, items = result.value)
            is OpsResult.Err -> _state.value = _state.value.copy(loading = false, errorMessage = result.error.message)
        }
    }

    suspend fun audit(area: ServiceArea?, id: String, pass: Boolean) {
        if (area == null) return
        _state.value = _state.value.copy(loading = true, message = null, errorMessage = null)
        when (val result = repository.changeBindAudit(id, pass)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(message = Strings.t(Str.OperationOk))
                load(area)
            }
            is OpsResult.Err -> _state.value = _state.value.copy(loading = false, errorMessage = result.error.message)
        }
    }
}
