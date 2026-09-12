package com.luopingtech.ebike.ops.feature.analysis

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.analysis.TaskStatisticsRepository
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsPeriod
import com.luopingtech.ebike.ops.domain.analysis.TaskStatisticsItem
import com.luopingtech.ebike.ops.domain.analysis.TaskStatisticsKind
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class TaskStatisticsUiState(
    val loading: Boolean = false,
    val loadingMore: Boolean = false,
    val kind: TaskStatisticsKind = TaskStatisticsKind.ChangeBattery,
    val period: OfflineOpsPeriod = OfflineOpsPeriod.Today,
    val validTab: Boolean = true,
    val items: List<TaskStatisticsItem> = emptyList(),
    val total: Int = 0,
    val pageNum: Int = 1,
    val hasMore: Boolean = false,
    val errorMessage: String? = null,
)

class TaskStatisticsFeature(
    private val repository: TaskStatisticsRepository,
    private val opPinProvider: () -> String,
) {
    private val _state = MutableStateFlow(TaskStatisticsUiState())
    val state: StateFlow<TaskStatisticsUiState> = _state.asStateFlow()

    fun open(kind: TaskStatisticsKind) {
        _state.value = TaskStatisticsUiState(kind = kind)
    }

    fun clear() {
        _state.value = TaskStatisticsUiState()
    }

    fun selectPeriod(period: OfflineOpsPeriod) {
        _state.value = _state.value.copy(period = period, errorMessage = null)
    }

    fun selectValidTab(valid: Boolean) {
        _state.value = _state.value.copy(validTab = valid, errorMessage = null)
    }

    suspend fun refresh(area: ServiceArea?) = load(area, pageNum = 1, append = false)

    suspend fun loadMore(area: ServiceArea?) {
        val snap = _state.value
        if (!snap.hasMore || snap.loading || snap.loadingMore) return
        load(area, pageNum = snap.pageNum + 1, append = true)
    }

    private suspend fun load(area: ServiceArea?, pageNum: Int, append: Boolean) {
        if (area == null) {
            _state.value = _state.value.copy(
                loading = false,
                loadingMore = false,
                errorMessage = Strings.t(Str.SelectServiceAreaFirst),
            )
            return
        }
        val snap = _state.value
        _state.value = snap.copy(loading = !append, loadingMore = append, errorMessage = null)
        when (
            val result = repository.load(
                area = area,
                kind = snap.kind,
                period = snap.period,
                valid = snap.validTab,
                pageNum = pageNum,
                opPin = opPinProvider(),
            )
        ) {
            is OpsResult.Ok -> {
                val merged = if (append) snap.items + result.value.items else result.value.items
                _state.value = _state.value.copy(
                    loading = false,
                    loadingMore = false,
                    items = merged,
                    total = result.value.total,
                    pageNum = result.value.pageNum,
                    hasMore = merged.size < result.value.total,
                )
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    loadingMore = false,
                    errorMessage = result.error.message,
                )
            }
        }
    }
}
