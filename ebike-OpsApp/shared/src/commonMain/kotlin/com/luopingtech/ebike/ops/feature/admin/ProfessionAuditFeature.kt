package com.luopingtech.ebike.ops.feature.admin

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.admin.AdminRepository
import com.luopingtech.ebike.ops.domain.admin.CareerAuditItem
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class ProfessionAuditUiState(
    val loading: Boolean = false,
    val items: List<CareerAuditItem> = emptyList(),
    val message: String? = null,
    val errorMessage: String? = null,
)

class ProfessionAuditFeature(
    private val repository: AdminRepository,
) {
    private val _state = MutableStateFlow(ProfessionAuditUiState())
    val state: StateFlow<ProfessionAuditUiState> = _state.asStateFlow()

    fun clear() {
        _state.value = ProfessionAuditUiState()
    }

    suspend fun load(area: ServiceArea?) {
        if (area == null) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.SelectServiceAreaFirst))
            return
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null)
        when (val result = repository.careerPage(area.id)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(loading = false, items = result.value)
            is OpsResult.Err -> _state.value = _state.value.copy(loading = false, errorMessage = result.error.message)
        }
    }

    suspend fun audit(area: ServiceArea?, id: String, pass: Boolean) {
        if (area == null) return
        _state.value = _state.value.copy(loading = true, message = null, errorMessage = null)
        when (val result = repository.careerAudit(id, pass)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(message = Strings.t(Str.OperationOk))
                load(area)
            }
            is OpsResult.Err -> _state.value = _state.value.copy(loading = false, errorMessage = result.error.message)
        }
    }
}
