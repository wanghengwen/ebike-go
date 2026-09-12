package com.luopingtech.ebike.ops.feature.tenant

import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.tenant.ServiceAreaRepository
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class ServiceAreaUiState(
    val loading: Boolean = false,
    val areas: List<ServiceArea> = emptyList(),
    val selected: ServiceArea? = null,
    val errorMessage: String? = null,
)

class ServiceAreaFeature(
    private val repository: ServiceAreaRepository,
) {
    private val _state = MutableStateFlow(
        ServiceAreaUiState(selected = repository.currentArea()),
    )
    val state: StateFlow<ServiceAreaUiState> = _state.asStateFlow()

    suspend fun load() {
        _state.value = _state.value.copy(loading = true, errorMessage = null)
        when (val result = repository.loadAreas()) {
            is OpsResult.Ok -> {
                _state.value = ServiceAreaUiState(
                    loading = false,
                    areas = result.value,
                    selected = repository.currentArea(),
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

    fun select(area: ServiceArea) {
        repository.selectArea(area)
        _state.value = _state.value.copy(selected = area)
    }

    fun clear() {
        repository.clearSelection()
        _state.value = ServiceAreaUiState()
    }
}
