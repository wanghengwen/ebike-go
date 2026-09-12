package com.luopingtech.ebike.ops.feature.analysis

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.analysis.OfflineOpsRepository
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsDashboard
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsPeriod
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsTab
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsTrendGrain
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class OfflineOpsUiState(
    val loading: Boolean = false,
    val tab: OfflineOpsTab = OfflineOpsTab.ChangeBattery,
    val period: OfflineOpsPeriod = OfflineOpsPeriod.Today,
    val trendGrain: OfflineOpsTrendGrain = OfflineOpsTrendGrain.Daily,
    val dashboard: OfflineOpsDashboard = OfflineOpsDashboard(),
    val errorMessage: String? = null,
)

class OfflineOpsFeature(
    private val repository: OfflineOpsRepository,
) {
    private val _state = MutableStateFlow(OfflineOpsUiState())
    val state: StateFlow<OfflineOpsUiState> = _state.asStateFlow()

    fun open() {
        _state.value = OfflineOpsUiState()
    }

    fun clear() {
        _state.value = OfflineOpsUiState()
    }

    fun selectTab(tab: OfflineOpsTab) {
        _state.value = _state.value.copy(
            tab = tab,
            period = OfflineOpsPeriod.Today,
            trendGrain = OfflineOpsTrendGrain.Daily,
            errorMessage = null,
        )
    }

    fun selectPeriod(period: OfflineOpsPeriod) {
        _state.value = _state.value.copy(period = period, errorMessage = null)
    }

    fun selectTrendGrain(grain: OfflineOpsTrendGrain) {
        _state.value = _state.value.copy(trendGrain = grain, errorMessage = null)
    }

    suspend fun load(
        area: ServiceArea?,
        mineOnly: Boolean = false,
        opPin: String? = null,
    ) {
        if (area == null) {
            _state.value = _state.value.copy(
                loading = false,
                errorMessage = Strings.t(Str.SelectServiceAreaFirst),
            )
            return
        }
        val snap = _state.value
        _state.value = snap.copy(loading = true, errorMessage = null)
        when (
            val result = repository.load(
                area = area,
                tab = snap.tab,
                period = snap.period,
                grain = snap.trendGrain,
                mineOnly = mineOnly,
                opPin = opPin,
            )
        ) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    dashboard = result.value,
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
