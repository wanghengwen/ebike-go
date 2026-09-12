package com.luopingtech.ebike.ops.feature.analysis

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.analysis.ReturnCarAnalysisRepository
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsPeriod
import com.luopingtech.ebike.ops.domain.analysis.ReturnCarAnalyzeResult
import com.luopingtech.ebike.ops.domain.analysis.ReturnCarStatusFilter
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class ReturnCarAnalysisUiState(
    val loading: Boolean = false,
    val period: OfflineOpsPeriod = OfflineOpsPeriod.Today,
    val statusFilter: ReturnCarStatusFilter = ReturnCarStatusFilter.Normal,
    val result: ReturnCarAnalyzeResult = ReturnCarAnalyzeResult(),
    val errorMessage: String? = null,
) {
    val visiblePoints get() = result.visible(statusFilter)
}

class ReturnCarAnalysisFeature(
    private val repository: ReturnCarAnalysisRepository,
) {
    private val _state = MutableStateFlow(ReturnCarAnalysisUiState())
    val state: StateFlow<ReturnCarAnalysisUiState> = _state.asStateFlow()

    fun open() {
        _state.value = ReturnCarAnalysisUiState()
    }

    fun clear() {
        _state.value = ReturnCarAnalysisUiState()
    }

    fun selectPeriod(period: OfflineOpsPeriod) {
        _state.value = _state.value.copy(period = period, errorMessage = null)
    }

    fun selectStatus(filter: ReturnCarStatusFilter) {
        _state.value = _state.value.copy(statusFilter = filter)
    }

    suspend fun load(area: ServiceArea?) {
        if (area == null) {
            _state.value = _state.value.copy(
                loading = false,
                errorMessage = Strings.t(Str.SelectServiceAreaFirst),
            )
            return
        }
        val snap = _state.value
        _state.value = snap.copy(loading = true, errorMessage = null)
        when (val result = repository.load(area, snap.period)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    result = result.value,
                    errorMessage = null,
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
}
