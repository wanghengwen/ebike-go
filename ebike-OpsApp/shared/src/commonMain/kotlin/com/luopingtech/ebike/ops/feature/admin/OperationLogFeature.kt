package com.luopingtech.ebike.ops.feature.admin

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.admin.AdminRepository
import com.luopingtech.ebike.ops.domain.admin.OperationLogItem
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class OperationLogUiState(
    val loading: Boolean = false,
    val items: List<OperationLogItem> = emptyList(),
    val errorMessage: String? = null,
)

class OperationLogFeature(
    private val repository: AdminRepository,
) {
    private val _state = MutableStateFlow(OperationLogUiState())
    val state: StateFlow<OperationLogUiState> = _state.asStateFlow()

    fun clear() {
        _state.value = OperationLogUiState()
    }

    suspend fun load(area: ServiceArea?) {
        if (area == null) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.SelectServiceAreaFirst))
            return
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null)
        when (val result = repository.operationLogList(area.id)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(loading = false, items = result.value)
            is OpsResult.Err -> _state.value = _state.value.copy(loading = false, errorMessage = result.error.message)
        }
    }
}
