package com.luopingtech.ebike.ops.feature.analysis

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.analysis.StationAnalysisRepository
import com.luopingtech.ebike.ops.domain.analysis.StationAnalyzeDetail
import com.luopingtech.ebike.ops.domain.analysis.StationAnalyzeItem
import com.luopingtech.ebike.ops.domain.analysis.StationOptStateFilter
import com.luopingtech.ebike.ops.domain.analysis.StationSortOrder
import com.luopingtech.ebike.ops.domain.analysis.StationTag
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

enum class StationAnalysisPage {
    List,
    Detail,
}

data class StationAnalysisUiState(
    val loading: Boolean = false,
    val loadingMore: Boolean = false,
    val detailLoading: Boolean = false,
    val page: StationAnalysisPage = StationAnalysisPage.List,
    val items: List<StationAnalyzeItem> = emptyList(),
    val total: Int = 0,
    val pageNum: Int = 1,
    val hasMore: Boolean = false,
    val tags: List<StationTag> = emptyList(),
    val selectedTagIds: Set<String> = emptySet(),
    val optState: StationOptStateFilter = StationOptStateFilter.All,
    val sort: StationSortOrder = StationSortOrder.None,
    val selected: StationAnalyzeItem? = null,
    val detail: StationAnalyzeDetail? = null,
    val errorMessage: String? = null,
)

class StationAnalysisFeature(
    private val repository: StationAnalysisRepository,
) {
    private val _state = MutableStateFlow(StationAnalysisUiState())
    val state: StateFlow<StationAnalysisUiState> = _state.asStateFlow()

    fun open() {
        _state.value = StationAnalysisUiState()
    }

    fun clear() {
        _state.value = StationAnalysisUiState()
    }

    fun setOptState(filter: StationOptStateFilter) {
        _state.value = _state.value.copy(optState = filter, errorMessage = null)
    }

    fun setSort(order: StationSortOrder) {
        _state.value = _state.value.copy(sort = order, errorMessage = null)
    }

    fun setSelectedTags(ids: Set<String>) {
        _state.value = _state.value.copy(selectedTagIds = ids, errorMessage = null)
    }

    fun toggleTag(id: String) {
        val cur = _state.value.selectedTagIds.toMutableSet()
        if (!cur.add(id)) cur.remove(id)
        _state.value = _state.value.copy(selectedTagIds = cur, errorMessage = null)
    }

    fun openDetail(item: StationAnalyzeItem) {
        _state.value = _state.value.copy(
            page = StationAnalysisPage.Detail,
            selected = item,
            detail = null,
            errorMessage = null,
        )
    }

    fun closeDetail() {
        _state.value = _state.value.copy(
            page = StationAnalysisPage.List,
            selected = null,
            detail = null,
            detailLoading = false,
        )
    }

    suspend fun loadTags() {
        when (val r = repository.tags()) {
            is OpsResult.Ok -> _state.value = _state.value.copy(tags = r.value)
            is OpsResult.Err -> Unit
        }
    }

    suspend fun refresh(area: ServiceArea?) {
        load(area, pageNum = 1, append = false)
    }

    suspend fun loadMore(area: ServiceArea?) {
        val snap = _state.value
        if (!snap.hasMore || snap.loading || snap.loadingMore) return
        load(area, pageNum = snap.pageNum + 1, append = true)
    }

    suspend fun loadDetail(area: ServiceArea?) {
        val item = _state.value.selected ?: return
        val serviceId = item.serviceId.ifBlank { area?.id.orEmpty() }
        if (serviceId.isBlank()) {
            _state.value = _state.value.copy(
                detailLoading = false,
                errorMessage = Strings.t(Str.SelectServiceAreaFirst),
            )
            return
        }
        _state.value = _state.value.copy(detailLoading = true, errorMessage = null)
        when (val r = repository.analyzeOne(item.parkingId, serviceId)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(detailLoading = false, detail = r.value, errorMessage = null)
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(detailLoading = false, errorMessage = r.error.message)
            }
        }
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
        _state.value = snap.copy(
            loading = !append,
            loadingMore = append,
            errorMessage = null,
        )
        when (
            val result = repository.list(
                area = area,
                pageNum = pageNum,
                optState = snap.optState,
                order = snap.sort,
                tagIds = snap.selectedTagIds.toList(),
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
                    errorMessage = null,
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
