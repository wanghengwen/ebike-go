package com.luopingtech.ebike.ops.feature.fence

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.fence.FenceRepository
import com.luopingtech.ebike.ops.domain.model.FenceBundle
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class FenceBrowseUiState(
    val loading: Boolean = false,
    val bundle: FenceBundle = FenceBundle(),
    val errorMessage: String? = null,
)

class FenceBrowseFeature(
    private val repository: FenceRepository,
) {
    private val _state = MutableStateFlow(FenceBrowseUiState())
    val state: StateFlow<FenceBrowseUiState> = _state.asStateFlow()

    fun clear() {
        _state.value = FenceBrowseUiState()
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
        when (val result = repository.loadByServiceId(area.id)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(loading = false, bundle = result.value)
            is OpsResult.Err -> _state.value = _state.value.copy(loading = false, errorMessage = result.error.message)
        }
    }
}
