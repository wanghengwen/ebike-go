package com.luopingtech.ebike.ops.feature.fence

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.fence.FenceRepository
import com.luopingtech.ebike.ops.domain.analysis.StationOptStateFilter
import com.luopingtech.ebike.ops.domain.analysis.StationTag
import com.luopingtech.ebike.ops.domain.model.FenceBundle
import com.luopingtech.ebike.ops.domain.model.FenceKind
import com.luopingtech.ebike.ops.domain.model.FencePolygon
import com.luopingtech.ebike.ops.domain.model.GeoLatLng
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.model.metersBetween
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

/** 对齐 FenceActivity：地图 / 列表。 */
enum class FenceBrowseMode { Map, List }

/** 对齐 FenceListFragment SegmentTab：停车区 / 禁停区。 */
enum class FenceListTab { Parking, NoParking }

data class FenceBrowseUiState(
    val loading: Boolean = false,
    val bundle: FenceBundle = FenceBundle(),
    val errorMessage: String? = null,
    val mode: FenceBrowseMode = FenceBrowseMode.Map,
    val listTab: FenceListTab = FenceListTab.Parking,
    /** false：只显示服务区中心约 1km 内停车/禁停（对齐 FenceMapFragment）。 */
    val showAllFences: Boolean = false,
    val optState: StationOptStateFilter = StationOptStateFilter.All,
    val selectedTagIds: Set<String> = emptySet(),
    val availableTags: List<StationTag> = emptyList(),
    val selectedFenceId: String? = null,
    /** 列表编辑模式（对齐工具栏「编辑」）。 */
    val editMode: Boolean = false,
    val selectedIds: Set<String> = emptySet(),
) {
    val parkingCount: Int get() = bundle.parkings.size
    val noParkingCount: Int get() = bundle.noParkings.size

    val focusCenter: GeoLatLng?
        get() = bundle.serviceAreas.firstOrNull()?.centerOrCentroid()
            ?: bundle.parkings.firstOrNull()?.centerOrCentroid()

    fun matchesFilters(fence: FencePolygon): Boolean {
        val statusOk = when (optState) {
            StationOptStateFilter.All -> true
            StationOptStateFilter.Operating -> fence.izEnable == true
            StationOptStateFilter.Stopped -> fence.izEnable == false
        }
        if (!statusOk) return false
        if (selectedTagIds.isEmpty()) return true
        if (fence.kind != FenceKind.Parking) return true
        return fence.tags.any { it.id in selectedTagIds }
    }

    fun filteredParkings(): List<FencePolygon> =
        bundle.parkings.filter { matchesFilters(it) }

    fun filteredNoParkings(): List<FencePolygon> =
        bundle.noParkings.filter { matchesFilters(it) }

    fun listItems(): List<FencePolygon> = when (listTab) {
        FenceListTab.Parking -> filteredParkings()
        FenceListTab.NoParking -> filteredNoParkings()
    }

    /** 地图多边形：服务区始终 + 距离裁剪后的停车/禁停。 */
    fun mapPolygons(): List<FencePolygon> {
        val center = focusCenter
        val parks = if (showAllFences || center == null) {
            bundle.parkings
        } else {
            bundle.parkings.filter { withinKm(it, center) }
        }
        val noParks = if (showAllFences || center == null) {
            bundle.noParkings
        } else {
            bundle.noParkings.filter { withinKm(it, center) }
        }
        return bundle.serviceAreas + parks + noParks
    }

    fun selectedFence(): FencePolygon? {
        val id = selectedFenceId ?: return null
        return bundle.all.firstOrNull { it.id == id }
    }

    private fun withinKm(fence: FencePolygon, center: GeoLatLng): Boolean {
        val c = fence.centerOrCentroid() ?: return false
        return metersBetween(center, c) <= NEAR_METERS
    }

    companion object {
        const val NEAR_METERS = 1000.0
    }
}

class FenceBrowseFeature(
    private val repository: FenceRepository,
    private val loadTags: suspend () -> OpsResult<List<StationTag>>,
) {
    private val _state = MutableStateFlow(FenceBrowseUiState())
    val state: StateFlow<FenceBrowseUiState> = _state.asStateFlow()

    fun clear() {
        _state.value = FenceBrowseUiState()
    }

    fun setMode(mode: FenceBrowseMode) {
        _state.value = _state.value.copy(mode = mode, selectedFenceId = null)
    }

    fun toggleMode() {
        val next = if (_state.value.mode == FenceBrowseMode.Map) {
            FenceBrowseMode.List
        } else {
            FenceBrowseMode.Map
        }
        setMode(next)
    }

    fun setListTab(tab: FenceListTab) {
        _state.value = _state.value.copy(listTab = tab, selectedFenceId = null)
    }

    fun setShowAllFences(showAll: Boolean) {
        _state.value = _state.value.copy(showAllFences = showAll)
    }

    fun setOptState(filter: StationOptStateFilter) {
        _state.value = _state.value.copy(optState = filter)
    }

    fun toggleTag(tagId: String) {
        val cur = _state.value.selectedTagIds.toMutableSet()
        if (!cur.add(tagId)) cur.remove(tagId)
        _state.value = _state.value.copy(selectedTagIds = cur)
    }

    fun clearTags() {
        _state.value = _state.value.copy(selectedTagIds = emptySet())
    }

    fun selectFence(id: String?) {
        _state.value = _state.value.copy(selectedFenceId = id)
    }

    fun setEditMode(enabled: Boolean) {
        _state.value = _state.value.copy(
            editMode = enabled,
            selectedIds = emptySet(),
            selectedFenceId = null,
        )
    }

    fun toggleSelected(id: String) {
        val cur = _state.value.selectedIds.toMutableSet()
        if (!cur.add(id)) cur.remove(id)
        _state.value = _state.value.copy(selectedIds = cur)
    }

    suspend fun deleteSelected(): OpsResult<Unit> {
        val ids = _state.value.selectedIds.toList()
        if (ids.isEmpty()) {
            return OpsResult.Err(com.luopingtech.ebike.ops.core.result.OpsError.business("FENCE", "empty"))
        }
        val result = when (_state.value.listTab) {
            FenceListTab.Parking -> repository.deleteParkingBatch(ids)
            FenceListTab.NoParking -> repository.deleteNoParkingBatch(ids)
        }
        if (result is OpsResult.Ok) {
            _state.value = _state.value.copy(selectedIds = emptySet(), editMode = false)
        }
        return result
    }

    suspend fun enableSelected(): OpsResult<Unit> {
        val ids = _state.value.selectedIds.toList()
        if (ids.isEmpty() || _state.value.listTab != FenceListTab.Parking) {
            return OpsResult.Err(com.luopingtech.ebike.ops.core.result.OpsError.business("FENCE", "empty"))
        }
        val result = repository.enableParkingBatch(ids)
        if (result is OpsResult.Ok) {
            _state.value = _state.value.copy(selectedIds = emptySet(), editMode = false)
        }
        return result
    }

    suspend fun disableSelected(): OpsResult<Unit> {
        val ids = _state.value.selectedIds.toList()
        if (ids.isEmpty() || _state.value.listTab != FenceListTab.Parking) {
            return OpsResult.Err(com.luopingtech.ebike.ops.core.result.OpsError.business("FENCE", "empty"))
        }
        val result = repository.disableParkingBatch(ids)
        if (result is OpsResult.Ok) {
            _state.value = _state.value.copy(selectedIds = emptySet(), editMode = false)
        }
        return result
    }

    suspend fun deleteFence(id: String, kind: FenceKind): OpsResult<Unit> {
        val result = when (kind) {
            FenceKind.NoParking -> repository.deleteNoParking(id)
            else -> repository.deleteParking(id)
        }
        if (result is OpsResult.Ok) {
            _state.value = _state.value.copy(selectedFenceId = null)
        }
        return result
    }

    suspend fun enableFence(id: String): OpsResult<Unit> {
        val result = repository.enableParkingBatch(listOf(id))
        if (result is OpsResult.Ok) {
            _state.value = _state.value.copy(selectedFenceId = null)
        }
        return result
    }

    suspend fun disableFence(id: String): OpsResult<Unit> {
        val result = repository.disableParkingBatch(listOf(id))
        if (result is OpsResult.Ok) {
            _state.value = _state.value.copy(selectedFenceId = null)
        }
        return result
    }

    suspend fun load(area: ServiceArea?) {
        if (area == null) {
            _state.value = _state.value.copy(
                loading = false,
                errorMessage = Strings.t(Str.SelectServiceAreaFirst),
            )
            return
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null, selectedFenceId = null)
        when (val tags = loadTags()) {
            is OpsResult.Ok -> _state.value = _state.value.copy(availableTags = tags.value)
            is OpsResult.Err -> Unit
        }
        when (val result = repository.loadByServiceId(area.id)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(loading = false, bundle = result.value)
            is OpsResult.Err -> _state.value = _state.value.copy(loading = false, errorMessage = result.error.message)
        }
    }
}
