package com.luopingtech.ebike.ops.feature.staff

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.staff.ServiceUserRepository
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.model.TeamWorker
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class StaffDirectoryUiState(
    val loading: Boolean = false,
    val workers: List<TeamWorker> = emptyList(),
    val errorMessage: String? = null,
)

class StaffDirectoryFeature(
    private val repository: ServiceUserRepository,
) {
    private val _state = MutableStateFlow(StaffDirectoryUiState())
    val state: StateFlow<StaffDirectoryUiState> = _state.asStateFlow()

    fun clear() {
        _state.value = StaffDirectoryUiState()
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
        when (val result = repository.listTeamWorkers(area.id)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(loading = false, workers = result.value)
            is OpsResult.Err -> _state.value = _state.value.copy(loading = false, errorMessage = result.error.message)
        }
    }
}
