package com.luopingtech.ebike.ops.feature.admin

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.admin.AdminRepository
import com.luopingtech.ebike.ops.domain.admin.OperationLogEventNode
import com.luopingtech.ebike.ops.domain.admin.OperationLogItem
import com.luopingtech.ebike.ops.domain.admin.OperationLogKind
import com.luopingtech.ebike.ops.domain.admin.OperationLogListQuery
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsTimeRanges
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

sealed class OperationLogNav {
    data object Filter : OperationLogNav()
    data object List : OperationLogNav()
    data class Detail(val result: String) : OperationLogNav()
}

data class OperationLogUiState(
    val nav: OperationLogNav = OperationLogNav.Filter,
    val kind: OperationLogKind = OperationLogKind.Car,
    val topInput: String = "",
    val bottomPhone: String = "",
    val startDate: String = OfflineOpsTimeRanges.yesterdayDateText(),
    val endDate: String = OfflineOpsTimeRanges.todayDateText(),
    val eventLabel: String = "",
    val eventTypeIds: String? = null,
    val eventTree: List<OperationLogEventNode> = emptyList(),
    val loading: Boolean = false,
    val loadingMore: Boolean = false,
    val items: List<OperationLogItem> = emptyList(),
    val pageNum: Int = 1,
    val finished: Boolean = false,
    val activeQuery: OperationLogListQuery? = null,
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

    fun setTopInput(value: String) {
        val s = _state.value
        val limited = when (s.kind) {
            OperationLogKind.Car -> value.filter { it.isDigit() }.take(10)
            OperationLogKind.Device -> value.filterNot { it.isWhitespace() }.take(24)
            OperationLogKind.Operator -> value.filter { it.isDigit() }.take(11)
        }
        _state.value = s.copy(topInput = limited)
    }

    fun setBottomPhone(value: String) {
        _state.value = _state.value.copy(bottomPhone = value.filter { it.isDigit() }.take(11))
    }

    fun setStartDate(date: String) {
        _state.value = _state.value.copy(startDate = date, errorMessage = null)
    }

    fun setEndDate(date: String) {
        _state.value = _state.value.copy(endDate = date, errorMessage = null)
    }

    fun setEventSelection(label: String?, typeIds: String?) {
        _state.value = _state.value.copy(
            eventLabel = label.orEmpty(),
            eventTypeIds = typeIds,
        )
    }

    fun selectKind(kind: OperationLogKind) {
        _state.value = _state.value.copy(
            kind = kind,
            topInput = "",
            bottomPhone = "",
            startDate = OfflineOpsTimeRanges.yesterdayDateText(),
            endDate = OfflineOpsTimeRanges.todayDateText(),
            eventLabel = "",
            eventTypeIds = null,
            errorMessage = null,
        )
    }

    fun resetFilter() {
        selectKind(_state.value.kind)
    }

    fun navigateBack() {
        when (val nav = _state.value.nav) {
            is OperationLogNav.Detail -> _state.value = _state.value.copy(nav = OperationLogNav.List)
            OperationLogNav.List -> _state.value = _state.value.copy(
                nav = OperationLogNav.Filter,
                items = emptyList(),
                activeQuery = null,
                pageNum = 1,
                finished = false,
            )
            OperationLogNav.Filter -> Unit
        }
    }

    fun openDetail(item: OperationLogItem) {
        _state.value = _state.value.copy(nav = OperationLogNav.Detail(item.result.ifBlank { item.content }))
    }

    suspend fun loadEventTree() {
        val type = _state.value.kind.eventTreeType
        when (val result = repository.operationLogEventTree(type)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(eventTree = result.value)
            is OpsResult.Err -> _state.value = _state.value.copy(eventTree = emptyList())
        }
    }

    fun confirmFilter(): Boolean {
        val s = _state.value
        val startMs = OfflineOpsTimeRanges.parseDateStart(s.startDate)
        val endMs = OfflineOpsTimeRanges.parseDateStart(s.endDate)
        if (startMs == null || endMs == null || startMs > endMs) {
            _state.value = s.copy(errorMessage = Strings.t(Str.ObjectionDateInvalid))
            return false
        }
        val startSec = OfflineOpsTimeRanges.dateStartEpochSec(s.startDate)?.toString()
        val endSec = OfflineOpsTimeRanges.dateEndEpochSec(s.endDate)?.toString()
        val query = when (s.kind) {
            OperationLogKind.Car -> OperationLogListQuery(
                kind = s.kind,
                carId = s.topInput.ifBlank { "*" },
                phone = s.bottomPhone.ifBlank { null },
                startTimeSec = startSec,
                endTimeSec = endSec,
                eventTypeIds = s.eventTypeIds,
            )
            OperationLogKind.Device -> OperationLogListQuery(
                kind = s.kind,
                imei = s.topInput.ifBlank { "*" },
                startTimeSec = startSec,
                endTimeSec = endSec,
                eventTypeIds = s.eventTypeIds,
            )
            OperationLogKind.Operator -> OperationLogListQuery(
                kind = s.kind,
                phone = s.topInput.ifBlank { null },
                startTimeSec = startSec,
                endTimeSec = endSec,
                eventTypeIds = s.eventTypeIds,
            )
        }
        _state.value = s.copy(
            activeQuery = query,
            nav = OperationLogNav.List,
            errorMessage = null,
            items = emptyList(),
            pageNum = 1,
            finished = false,
        )
        return true
    }

    suspend fun loadList(refresh: Boolean = true) {
        val base = _state.value.activeQuery ?: return
        val page = if (refresh) 1 else _state.value.pageNum + 1
        if (!refresh) {
            val s = _state.value
            if (s.loading || s.loadingMore || s.finished) return
            _state.value = s.copy(loadingMore = true, errorMessage = null)
        } else {
            _state.value = _state.value.copy(loading = true, errorMessage = null, pageNum = 1, finished = false)
        }
        val query = base.copy(pageNum = page, pageSize = 10)
        when (val result = repository.operationLogList(query)) {
            is OpsResult.Ok -> {
                val merged = if (refresh) result.value else _state.value.items + result.value
                _state.value = _state.value.copy(
                    loading = false,
                    loadingMore = false,
                    items = merged,
                    pageNum = page,
                    finished = result.value.size < query.pageSize,
                )
            }
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                loadingMore = false,
                errorMessage = result.error.message,
            )
        }
    }
}
