package com.luopingtech.ebike.ops.feature.admin

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.admin.AdminRepository
import com.luopingtech.ebike.ops.domain.admin.ObjectionOrderItem
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class ObjectionOrderUiState(
    val loading: Boolean = false,
    val items: List<ObjectionOrderItem> = emptyList(),
    val selected: ObjectionOrderItem? = null,
    val message: String? = null,
    val errorMessage: String? = null,
)

class ObjectionOrderFeature(
    private val repository: AdminRepository,
) {
    private val _state = MutableStateFlow(ObjectionOrderUiState())
    val state: StateFlow<ObjectionOrderUiState> = _state.asStateFlow()

    fun clear() {
        _state.value = ObjectionOrderUiState()
    }

    fun closeDetail() {
        _state.value = _state.value.copy(selected = null)
    }

    suspend fun load(area: ServiceArea?) {
        if (area == null) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.SelectServiceAreaFirst))
            return
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null)
        when (val result = repository.objectionPage(area.id)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(loading = false, items = result.value)
            is OpsResult.Err -> _state.value = _state.value.copy(loading = false, errorMessage = result.error.message)
        }
    }

    suspend fun openDetail(area: ServiceArea?, item: ObjectionOrderItem) {
        if (area == null) return
        _state.value = _state.value.copy(loading = true, errorMessage = null)
        when (val result = repository.objectionDetail(item.id, item.orderId)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(loading = false, selected = result.value)
            is OpsResult.Err -> _state.value = _state.value.copy(loading = false, selected = item)
        }
    }

    suspend fun deal(area: ServiceArea?) {
        val selected = _state.value.selected ?: return
        if (area == null) return
        when (val result = repository.dealObjection(selected.id, selected.orderId)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    message = Strings.t(Str.ObjectionDealOk),
                    selected = null,
                )
                load(area)
            }
            is OpsResult.Err -> _state.value = _state.value.copy(errorMessage = result.error.message)
        }
    }
}
