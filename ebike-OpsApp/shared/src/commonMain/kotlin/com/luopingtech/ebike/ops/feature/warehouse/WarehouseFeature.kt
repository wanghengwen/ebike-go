package com.luopingtech.ebike.ops.feature.warehouse

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.warehouse.WarehouseRepository
import com.luopingtech.ebike.ops.domain.model.WarehouseComponent
import com.luopingtech.ebike.ops.domain.model.WarehouseOperationType
import com.luopingtech.ebike.ops.domain.model.WarehouseRecord
import com.luopingtech.ebike.ops.domain.warehouse.WarehouseComponentFilter
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

enum class WarehousePage {
    /** Legacy WarehouseMainActivity: 有编码 / 无编码 for a single In or Out kind. */
    KindMenu,
    Operate,
    Records,
    Detail,
}

data class WarehouseUiState(
    val page: WarehousePage = WarehousePage.KindMenu,
    val loading: Boolean = false,
    /** Required for KindMenu / Operate — set by workbench 归还入库 / 领用出库. */
    val operationType: WarehouseOperationType = WarehouseOperationType.Out,
    val withCode: Boolean = true,
    val componentNames: List<String> = emptyList(),
    val selectedComponentName: String = "",
    val quantity: Int = 1,
    val codeInput: String = "",
    val scanned: List<WarehouseComponent> = emptyList(),
    val stockHint: String? = null,
    val records: List<WarehouseRecord> = emptyList(),
    val selectedRecordId: String? = null,
    val details: List<WarehouseComponent> = emptyList(),
    val message: String? = null,
    val errorMessage: String? = null,
)

class WarehouseFeature(
    private val repository: WarehouseRepository,
    private val pinProvider: () -> String,
) {
    private val _state = MutableStateFlow(WarehouseUiState())
    val state: StateFlow<WarehouseUiState> = _state.asStateFlow()

    /** Workbench 归还入库 / 领用出库 → 遗留 WarehouseMainActivity. */
    fun openKind(type: WarehouseOperationType) {
        _state.value = WarehouseUiState(
            page = WarehousePage.KindMenu,
            operationType = type,
        )
    }

    fun openOperate(type: WarehouseOperationType, withCode: Boolean) {
        _state.value = WarehouseUiState(
            page = WarehousePage.Operate,
            operationType = type,
            withCode = withCode,
        )
    }

    fun backToKindMenu() {
        val type = _state.value.operationType
        openKind(type)
    }

    fun setCodeInput(value: String) {
        _state.value = _state.value.copy(codeInput = value, errorMessage = null)
    }

    fun setQuantity(value: Int) {
        _state.value = _state.value.copy(quantity = value.coerceAtLeast(1), errorMessage = null)
    }

    fun setSelectedComponentName(name: String) {
        _state.value = _state.value.copy(selectedComponentName = name, errorMessage = null)
    }

    fun removeScanned(componentNo: String) {
        _state.value = _state.value.copy(
            scanned = _state.value.scanned.filterNot { it.componentNo == componentNo },
        )
    }

    suspend fun ensureComponentNames(existCode: Int? = if (_state.value.withCode) 1 else 0) {
        _state.value = _state.value.copy(loading = true, errorMessage = null)
        when (val result = repository.queryAllComponentNames(existCode)) {
            is OpsResult.Ok -> {
                val names = result.value
                val preferred = _state.value.selectedComponentName
                    .ifBlank { names.firstOrNull { !WarehouseComponentFilter.isAll(it) }.orEmpty() }
                    .ifBlank { names.firstOrNull().orEmpty() }
                _state.value = _state.value.copy(
                    loading = false,
                    componentNames = names,
                    selectedComponentName = preferred,
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

    suspend fun refreshStockHint() {
        val name = _state.value.selectedComponentName
        if (WarehouseComponentFilter.isAll(name)) {
            _state.value = _state.value.copy(stockHint = null)
            return
        }
        if (name.isBlank()) return
        when (val result = repository.queryQuantity(name)) {
            is OpsResult.Ok -> {
                val c = result.value
                _state.value = _state.value.copy(
                    stockHint = Strings.t(Str.StockHint, c.stockQuantity, c.receivedQuantity),
                )
            }
            is OpsResult.Err -> Unit
        }
    }

    suspend fun resolveCode(raw: String = _state.value.codeInput) {
        val code = raw.trim()
        if (code.isBlank()) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.EnterOrScanComponent))
            return
        }
        if (_state.value.scanned.any { it.componentNo.equals(code, ignoreCase = true) }) {
            _state.value = _state.value.copy(
                errorMessage = Strings.t(Str.AlreadyAddedCode, code),
                codeInput = "",
            )
            return
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        when (val result = repository.queryByComponentNo(code)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    scanned = _state.value.scanned + result.value,
                    codeInput = "",
                    message = Strings.t(Str.JoinedComponent, result.value.componentNo),
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

    suspend fun submitOperate() {
        val pin = pinProvider()
        val current = _state.value
        _state.value = current.copy(loading = true, errorMessage = null, message = null)
        val result = if (current.withCode) {
            repository.operateByCode(
                components = current.scanned,
                operationType = current.operationType,
                receiver = pin,
                pin = pin,
            )
        } else {
            if (WarehouseComponentFilter.isAll(current.selectedComponentName)) {
                _state.value = current.copy(
                    loading = false,
                    errorMessage = Strings.t(Str.SelectComponentFirst),
                )
                return
            }
            repository.operateWithoutCode(
                componentName = current.selectedComponentName,
                componentClassify = 1,
                operationType = current.operationType,
                operationNum = current.quantity,
                receiver = pin,
            )
        }
        when (result) {
            is OpsResult.Ok -> {
                _state.value = WarehouseUiState(
                    page = WarehousePage.KindMenu,
                    operationType = current.operationType,
                    message = Strings.t(Str.OpSuccess, current.operationType.label),
                )
            }
            is OpsResult.Err -> {
                _state.value = current.copy(
                    loading = false,
                    errorMessage = result.error.message,
                )
            }
        }
    }

    suspend fun openRecords(operationType: WarehouseOperationType? = null) {
        _state.value = WarehouseUiState(
            page = WarehousePage.Records,
            loading = true,
            operationType = operationType ?: WarehouseOperationType.Out,
        )
        when (val result = repository.pageRecords(operationType = operationType)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    records = result.value,
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

    suspend fun openDetail(recordId: String) {
        _state.value = _state.value.copy(
            page = WarehousePage.Detail,
            selectedRecordId = recordId,
            loading = true,
            errorMessage = null,
        )
        when (val result = repository.detailList(recordId)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    details = result.value,
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

    fun clear() {
        _state.value = WarehouseUiState()
    }
}
