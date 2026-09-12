package com.luopingtech.ebike.ops.feature.order

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.core.time.nowEpochMillis
import com.luopingtech.ebike.ops.data.order.OrderRepository
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.order.OrderListQuery
import com.luopingtech.ebike.ops.domain.order.OrderPayStates
import com.luopingtech.ebike.ops.domain.order.OrderRecord
import com.luopingtech.ebike.ops.domain.order.OrderSearchClassifier
import com.luopingtech.ebike.ops.domain.order.OrderSearchKind
import com.luopingtech.ebike.ops.domain.order.OrderUserDetail
import com.luopingtech.ebike.ops.domain.order.OrderUserPageItem
import com.luopingtech.ebike.ops.domain.permission.OpsPermissionCodes
import com.luopingtech.ebike.ops.domain.permission.OpsPermissions
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

sealed class OrderQueryNav {
    data object Search : OrderQueryNav()
    data class UserPicker(val name: String) : OrderQueryNav()
    data class UserOrders(val pin: String) : OrderQueryNav()
    data class VehicleOrders(val carId: String? = null, val imei: String? = null) : OrderQueryNav()
}

data class OrderQueryUiState(
    val keyword: String = "",
    val searching: Boolean = false,
    val items: List<OrderRecord> = emptyList(),
    val loading: Boolean = false,
    val finished: Boolean = false,
    val errorMessage: String? = null,
    val toastMessage: String? = null,
    val areaName: String = "",
    val canQueryPerson: Boolean = false,
    val canQueryVehicle: Boolean = false,
    val nav: OrderQueryNav = OrderQueryNav.Search,
    val pickerUsers: List<OrderUserPageItem> = emptyList(),
    val pickerLoading: Boolean = false,
    val pickerFinished: Boolean = false,
    val pickerError: String? = null,
    val scopeUser: OrderUserDetail? = null,
    val lastOrder: OrderRecord? = null,
    val history: List<OrderRecord> = emptyList(),
    val historyLoading: Boolean = false,
    val historyFinished: Boolean = false,
    val historyError: String? = null,
    val scopeLoading: Boolean = false,
)

/**
 * 订单查询：搜索首页 + 人选 + 人/车订单范围页。
 */
class OrderQueryFeature(
    private val repository: OrderRepository,
    private val permissionsProvider: () -> OpsPermissions,
) {
    private val _state = MutableStateFlow(OrderQueryUiState())
    val state: StateFlow<OrderQueryUiState> = _state.asStateFlow()

    private var listPage = 1
    private var pickerPage = 1
    private var historyPage = 1
    private var currentArea: ServiceArea? = null
    private var pickerName: String = ""
    private var scopeCarId: String? = null
    private var scopeImei: String? = null
    private var scopeUserPin: String? = null

    fun clear() {
        listPage = 1
        pickerPage = 1
        historyPage = 1
        pickerName = ""
        scopeCarId = null
        scopeImei = null
        scopeUserPin = null
        _state.value = OrderQueryUiState()
    }

    fun consumeToast() {
        _state.value = _state.value.copy(toastMessage = null)
    }

    fun setKeyword(value: String) {
        _state.value = _state.value.copy(keyword = value)
    }

    fun refreshPermissions() {
        val p = permissionsProvider()
        val parent = p.has(OpsPermissionCodes.ORDER_QUERY) || p.has(OpsPermissionCodes.PC_ORDER_MENU)
        _state.value = _state.value.copy(
            canQueryPerson = parent || p.has(OpsPermissionCodes.ORDER_QUERY_PERSONAL),
            canQueryVehicle = parent || p.has(OpsPermissionCodes.ORDER_QUERY_VEHICLE),
        )
    }

    fun navigateBack() {
        when (_state.value.nav) {
            is OrderQueryNav.Search -> Unit
            is OrderQueryNav.UserPicker -> {
                _state.value = _state.value.copy(nav = OrderQueryNav.Search)
            }
            is OrderQueryNav.UserOrders, is OrderQueryNav.VehicleOrders -> {
                _state.value = _state.value.copy(
                    nav = OrderQueryNav.Search,
                    scopeUser = null,
                    lastOrder = null,
                    history = emptyList(),
                    historyError = null,
                )
            }
        }
    }

    suspend fun open(area: ServiceArea?) {
        currentArea = area
        refreshPermissions()
        _state.value = _state.value.copy(
            nav = OrderQueryNav.Search,
            areaName = area?.name.orEmpty(),
            errorMessage = null,
        )
        refreshHomeList()
    }

    suspend fun refreshHomeList() {
        listPage = 1
        _state.value = _state.value.copy(
            items = emptyList(),
            finished = false,
            errorMessage = null,
            loading = true,
        )
        loadMoreHome()
    }

    suspend fun loadMoreHome() {
        if (_state.value.finished) return
        _state.value = _state.value.copy(loading = true, errorMessage = null)
        val serviceId = currentArea?.id?.toLongOrNull()
        val end = nowEpochMillis()
        val from = end - 3L * 24 * 60 * 60 * 1000
        val baseQuery = OrderListQuery(
            pageNum = listPage,
            pageSize = PAGE_SIZE,
            serviceId = serviceId,
            startTimeMs = from to end,
        )
        when (val result = repository.listOrders(baseQuery)) {
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    finished = true,
                    errorMessage = result.error.message,
                )
                return
            }
            is OpsResult.Ok -> {
                var batch = result.value
                if (listPage == 1) {
                    val riding = repository.listOrders(
                        OrderListQuery(
                            pageNum = 1,
                            pageSize = PAGE_SIZE,
                            serviceId = serviceId,
                            izPaid = OrderPayStates.Riding,
                        ),
                    )
                    if (riding is OpsResult.Ok) {
                        batch = dedupe(riding.value + batch)
                    }
                }
                val merged = dedupe(_state.value.items + batch)
                val capped = if (merged.size > MAX_HOME) merged.take(MAX_HOME) else merged
                listPage += 1
                _state.value = _state.value.copy(
                    loading = false,
                    items = capped,
                    finished = batch.isEmpty() || batch.size < PAGE_SIZE || capped.size >= MAX_HOME,
                )
            }
        }
    }

    suspend fun search() {
        val raw = _state.value.keyword.trim()
        if (raw.isEmpty()) {
            _state.value = _state.value.copy(toastMessage = Strings.t(Str.OrderQueryInputHint))
            return
        }
        val kind = OrderSearchClassifier.classify(raw)
        if (!canSearch(kind)) {
            _state.value = _state.value.copy(toastMessage = Strings.t(Str.OrderQueryNoPerm))
            return
        }
        _state.value = _state.value.copy(searching = true, toastMessage = null)
        try {
            when (kind) {
                OrderSearchKind.CarId -> {
                    _state.value = _state.value.copy(nav = OrderQueryNav.VehicleOrders(carId = raw))
                    openVehicleScope(carId = raw, imei = null)
                }
                OrderSearchKind.Imei -> {
                    _state.value = _state.value.copy(nav = OrderQueryNav.VehicleOrders(imei = raw))
                    openVehicleScope(carId = null, imei = raw)
                }
                OrderSearchKind.Phone -> {
                    val area = currentArea
                    if (area == null || area.id.isBlank()) {
                        _state.value = _state.value.copy(toastMessage = Strings.t(Str.SelectServiceAreaFirst))
                        return
                    }
                    when (val result = repository.getUserByPhone(raw, listOf(area.id))) {
                        is OpsResult.Ok -> {
                            if (result.value.pin.isBlank()) {
                                _state.value = _state.value.copy(toastMessage = Strings.t(Str.OrderQueryUserNotFound))
                            } else {
                                _state.value = _state.value.copy(nav = OrderQueryNav.UserOrders(result.value.pin))
                                openUserScope(result.value.pin)
                            }
                        }
                        is OpsResult.Err -> {
                            _state.value = _state.value.copy(
                                toastMessage = result.error.message.ifBlank {
                                    Strings.t(Str.OrderQueryUserNotFound)
                                },
                            )
                        }
                    }
                }
                OrderSearchKind.Name -> {
                    val area = currentArea
                    val sid = area?.id?.toLongOrNull()
                    if (sid == null) {
                        _state.value = _state.value.copy(toastMessage = Strings.t(Str.SelectServiceAreaFirst))
                        return
                    }
                    when (val probe = repository.listUsersByName(raw, listOf(sid), 1, 1)) {
                        is OpsResult.Ok -> {
                            if (probe.value.first.isEmpty()) {
                                _state.value = _state.value.copy(toastMessage = Strings.t(Str.OrderQueryUserNotFound))
                            } else {
                                pickerName = raw
                                pickerPage = 1
                                _state.value = _state.value.copy(
                                    nav = OrderQueryNav.UserPicker(raw),
                                    pickerUsers = emptyList(),
                                    pickerFinished = false,
                                    pickerError = null,
                                )
                                loadMorePicker()
                            }
                        }
                        is OpsResult.Err -> {
                            _state.value = _state.value.copy(toastMessage = probe.error.message)
                        }
                    }
                }
            }
        } finally {
            _state.value = _state.value.copy(searching = false)
        }
    }

    suspend fun loadMorePicker() {
        if (_state.value.pickerFinished) return
        val sid = currentArea?.id?.toLongOrNull()
        if (sid == null) {
            _state.value = _state.value.copy(
                pickerFinished = true,
                pickerError = Strings.t(Str.SelectServiceAreaFirst),
            )
            return
        }
        _state.value = _state.value.copy(pickerLoading = true, pickerError = null)
        when (val result = repository.listUsersByName(pickerName, listOf(sid), pickerPage, PICKER_PAGE)) {
            is OpsResult.Ok -> {
                val (batch, count) = result.value
                val merged = _state.value.pickerUsers + batch
                pickerPage += 1
                _state.value = _state.value.copy(
                    pickerLoading = false,
                    pickerUsers = merged,
                    pickerFinished = batch.size < PICKER_PAGE || merged.size >= count,
                )
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    pickerLoading = false,
                    pickerFinished = true,
                    pickerError = result.error.message,
                )
            }
        }
    }

    suspend fun openUserFromPicker(pin: String) {
        _state.value = _state.value.copy(nav = OrderQueryNav.UserOrders(pin))
        openUserScope(pin)
    }

    suspend fun openVehicleFromHome(carId: String) {
        if (!_state.value.canQueryVehicle) {
            _state.value = _state.value.copy(toastMessage = Strings.t(Str.OrderQueryNoVehiclePerm))
            return
        }
        _state.value = _state.value.copy(nav = OrderQueryNav.VehicleOrders(carId = carId))
        openVehicleScope(carId = carId, imei = null)
    }

    private suspend fun openUserScope(pin: String) {
        scopeUserPin = pin
        scopeCarId = null
        scopeImei = null
        historyPage = 1
        _state.value = _state.value.copy(
            scopeLoading = true,
            scopeUser = null,
            lastOrder = null,
            history = emptyList(),
            historyFinished = false,
            historyError = null,
        )
        when (val user = repository.getUserByPin(pin)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(scopeUser = user.value)
            is OpsResult.Err -> _state.value = _state.value.copy(historyError = user.error.message)
        }
        when (val last = repository.detailLastRecord(userPin = pin)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(lastOrder = last.value)
            is OpsResult.Err -> Unit
        }
        _state.value = _state.value.copy(scopeLoading = false)
        loadMoreHistory()
    }

    private suspend fun openVehicleScope(carId: String?, imei: String?) {
        scopeUserPin = null
        scopeCarId = carId
        scopeImei = imei
        historyPage = 1
        _state.value = _state.value.copy(
            scopeLoading = true,
            scopeUser = null,
            lastOrder = null,
            history = emptyList(),
            historyFinished = false,
            historyError = null,
        )
        when (val last = repository.detailLastRecord(carId = carId, imei = imei)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(lastOrder = last.value)
            is OpsResult.Err -> _state.value = _state.value.copy(historyError = last.error.message)
        }
        _state.value = _state.value.copy(scopeLoading = false)
        loadMoreHistory()
    }

    suspend fun loadMoreHistory() {
        if (_state.value.historyFinished) return
        _state.value = _state.value.copy(historyLoading = true, historyError = null)
        val query = OrderListQuery(
            pageNum = historyPage,
            pageSize = PAGE_SIZE,
            userPin = scopeUserPin,
            carId = scopeCarId,
            imei = scopeImei,
        )
        when (val result = repository.listOrders(query)) {
            is OpsResult.Ok -> {
                val batch = result.value
                val merged = dedupe(_state.value.history + batch)
                historyPage += 1
                _state.value = _state.value.copy(
                    historyLoading = false,
                    history = merged,
                    historyFinished = batch.isEmpty() || batch.size < PAGE_SIZE,
                )
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    historyLoading = false,
                    historyFinished = true,
                    historyError = result.error.message,
                )
            }
        }
    }

    private fun canSearch(kind: OrderSearchKind): Boolean {
        val s = _state.value
        return when (kind) {
            OrderSearchKind.Phone, OrderSearchKind.Name -> s.canQueryPerson
            OrderSearchKind.CarId, OrderSearchKind.Imei -> s.canQueryVehicle
        }
    }

    private fun dedupe(rows: List<OrderRecord>): List<OrderRecord> {
        val seen = linkedSetOf<String>()
        return rows.filter { row ->
            val key = row.id.ifBlank { "${row.carId}|${row.startTime}|${row.phone}" }
            seen.add(key)
        }
    }

    companion object {
        private const val PAGE_SIZE = 10
        private const val PICKER_PAGE = 15
        private const val MAX_HOME = 200
    }
}
