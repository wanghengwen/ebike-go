package com.luopingtech.ebike.ops.feature.order

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.core.time.nowEpochMillis
import com.luopingtech.ebike.ops.data.order.OrderRepository
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.order.OrderDepositRecord
import com.luopingtech.ebike.ops.domain.order.OrderListQuery
import com.luopingtech.ebike.ops.domain.order.OrderPayStates
import com.luopingtech.ebike.ops.domain.order.OrderRecord
import com.luopingtech.ebike.ops.domain.order.OrderRideCard
import com.luopingtech.ebike.ops.domain.order.OrderRideCardRecord
import com.luopingtech.ebike.ops.domain.order.OrderSearchClassifier
import com.luopingtech.ebike.ops.domain.order.OrderSearchKind
import com.luopingtech.ebike.ops.domain.order.OrderUserAssets
import com.luopingtech.ebike.ops.domain.order.OrderUserDetail
import com.luopingtech.ebike.ops.domain.order.OrderUserPageItem
import com.luopingtech.ebike.ops.domain.order.OrderWalletInfo
import com.luopingtech.ebike.ops.domain.order.OrderWalletRecord
import com.luopingtech.ebike.ops.domain.permission.OpsPermissionCodes
import com.luopingtech.ebike.ops.domain.permission.OpsPermissions
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

sealed class OrderQueryNav {
    data object Search : OrderQueryNav()
    data class UserPicker(val name: String) : OrderQueryNav()
    data class UserOrders(val pin: String) : OrderQueryNav()
    data class UserWallet(val pin: String) : OrderQueryNav()
    data class UserDeposit(val pin: String) : OrderQueryNav()
    data class UserRideCard(val pin: String) : OrderQueryNav()
    data class UserOrderHistory(val pin: String) : OrderQueryNav()
    data class VehicleOrders(val carId: String? = null, val imei: String? = null) : OrderQueryNav()
    /** 对齐 ModifyAmountActivity：结束行程改金额 / 未支付改金额。 */
    data class ModifyAmount(val fromEndTrip: Boolean) : OrderQueryNav()
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
    val userAssets: OrderUserAssets? = null,
    val lastOrder: OrderRecord? = null,
    val history: List<OrderRecord> = emptyList(),
    val historyLoading: Boolean = false,
    val historyFinished: Boolean = false,
    val historyError: String? = null,
    val scopeLoading: Boolean = false,
    /** 订单记录页当前选中项（对齐 OrderHistoryActivity 点选后画轨迹）。 */
    val selectedOrder: OrderRecord? = null,
    val trackLoading: Boolean = false,
    val trackFitNonce: Int = 0,
    val walletInfo: OrderWalletInfo? = null,
    val walletRecords: List<OrderWalletRecord> = emptyList(),
    val depositRecords: List<OrderDepositRecord> = emptyList(),
    val rideCards: List<OrderRideCard> = emptyList(),
    val rideRecords: List<OrderRideCardRecord> = emptyList(),
    val assetLoading: Boolean = false,
    /** 结束行程 / 启动 / 临时通电等操作中。 */
    val actionLoading: Boolean = false,
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
        when (val nav = _state.value.nav) {
            is OrderQueryNav.Search -> Unit
            is OrderQueryNav.UserPicker -> {
                _state.value = _state.value.copy(nav = OrderQueryNav.Search)
            }
            is OrderQueryNav.UserWallet,
            is OrderQueryNav.UserDeposit,
            is OrderQueryNav.UserRideCard,
            is OrderQueryNav.UserOrderHistory,
            is OrderQueryNav.ModifyAmount,
            -> {
                val pin = when (nav) {
                    is OrderQueryNav.UserWallet -> nav.pin
                    is OrderQueryNav.UserDeposit -> nav.pin
                    is OrderQueryNav.UserRideCard -> nav.pin
                    is OrderQueryNav.UserOrderHistory -> nav.pin
                    is OrderQueryNav.ModifyAmount -> scopeUserPin.orEmpty()
                }
                _state.value = _state.value.copy(nav = OrderQueryNav.UserOrders(pin))
            }
            is OrderQueryNav.UserOrders, is OrderQueryNav.VehicleOrders -> {
                _state.value = _state.value.copy(
                    nav = OrderQueryNav.Search,
                    scopeUser = null,
                    userAssets = null,
                    lastOrder = null,
                    history = emptyList(),
                    historyError = null,
                    selectedOrder = null,
                    trackLoading = false,
                    walletInfo = null,
                    walletRecords = emptyList(),
                    depositRecords = emptyList(),
                    rideCards = emptyList(),
                    rideRecords = emptyList(),
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

    /**
     * 仅刷新服务区与权限，不强制回到搜索首页。
     * 用于「车辆详情 → 订单信息」等已指定车号入口，避免冲掉 [OrderQueryNav.VehicleOrders]。
     */
    suspend fun bindArea(area: ServiceArea?) {
        currentArea = area
        refreshPermissions()
        _state.value = _state.value.copy(areaName = area?.name.orEmpty())
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
        // 对齐原版 onlyOneWeek：近 7 天，请求体 endTime=[from, now]
        val from = end - 7L * 24 * 60 * 60 * 1000
        val baseQuery = OrderListQuery(
            pageNum = listPage,
            pageSize = PAGE_SIZE,
            serviceId = serviceId,
            endTimeMs = from to end,
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

    /**
     * 首页点击完结订单 → 用户详情（对齐 SearchOrderUserDetailActivity）。
     */
    suspend fun openUserFromHome(pin: String) {
        refreshPermissions()
        if (!_state.value.canQueryPerson) {
            _state.value = _state.value.copy(toastMessage = Strings.t(Str.OrderQueryNoPerm))
            return
        }
        val id = pin.trim()
        if (id.isBlank()) {
            _state.value = _state.value.copy(toastMessage = Strings.t(Str.OrderQueryUserNotFound))
            return
        }
        _state.value = _state.value.copy(nav = OrderQueryNav.UserOrders(id))
        openUserScope(id)
    }

    suspend fun openVehicleFromHome(carId: String) {
        refreshPermissions()
        if (!_state.value.canQueryVehicle) {
            _state.value = _state.value.copy(toastMessage = Strings.t(Str.OrderQueryNoVehiclePerm))
            return
        }
        val id = carId.trim()
        if (id.isBlank()) {
            _state.value = _state.value.copy(toastMessage = Strings.t(Str.EnterCarId))
            return
        }
        _state.value = _state.value.copy(nav = OrderQueryNav.VehicleOrders(carId = id))
        openVehicleScope(carId = id, imei = null)
    }

    private suspend fun openUserScope(pin: String) {
        scopeUserPin = pin
        scopeCarId = null
        scopeImei = null
        historyPage = 1
        _state.value = _state.value.copy(
            scopeLoading = true,
            scopeUser = null,
            userAssets = null,
            lastOrder = null,
            history = emptyList(),
            historyFinished = false,
            historyError = null,
            selectedOrder = null,
            trackLoading = false,
            trackFitNonce = 0,
        )
        when (val user = repository.getUserByPin(pin)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(scopeUser = user.value)
            is OpsResult.Err -> _state.value = _state.value.copy(historyError = user.error.message)
        }
        val serviceId = currentArea?.id.orEmpty()
        when (val assets = repository.getUserAssets(pin, serviceId)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(userAssets = assets.value)
            is OpsResult.Err -> Unit
        }
        when (val last = repository.detailLastRecord(userPin = pin)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    lastOrder = last.value,
                    selectedOrder = last.value,
                    trackFitNonce = if (last.value.trajectory.isNotEmpty()) 1 else 0,
                )
                if (last.value.trajectory.isEmpty()) {
                    selectHistoryOrder(last.value)
                }
            }
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
            userAssets = null,
            lastOrder = null,
            history = emptyList(),
            historyFinished = false,
            historyError = null,
            selectedOrder = null,
            trackLoading = false,
            trackFitNonce = 0,
        )
        when (val last = repository.detailLastRecord(carId = carId, imei = imei)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    lastOrder = last.value,
                    selectedOrder = last.value,
                    trackFitNonce = if (last.value.trajectory.isNotEmpty()) 1 else 0,
                )
                if (last.value.trajectory.isEmpty()) {
                    selectHistoryOrder(last.value)
                }
            }
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
                val nextSelected = when {
                    _state.value.selectedOrder != null -> _state.value.selectedOrder
                    merged.isNotEmpty() -> merged.first()
                    _state.value.lastOrder != null -> _state.value.lastOrder
                    else -> null
                }
                _state.value = _state.value.copy(
                    historyLoading = false,
                    history = merged,
                    historyFinished = batch.isEmpty() || batch.size < PAGE_SIZE,
                    selectedOrder = nextSelected,
                )
                // 首屏自动取第一单轨迹（对齐 autoTakeFristItem）
                if (_state.value.selectedOrder != null &&
                    _state.value.selectedOrder!!.trajectory.isEmpty() &&
                    historyPage == 2
                ) {
                    selectHistoryOrder(_state.value.selectedOrder!!)
                } else if (nextSelected != null &&
                    _state.value.trackFitNonce == 0 &&
                    nextSelected.trajectory.isNotEmpty()
                ) {
                    _state.value = _state.value.copy(trackFitNonce = 1)
                }
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

    /**
     * 对齐 OrderHistoryActivity 点选：
     * 列表超 6 个月常无轨迹 → 调 orderDetail 补全，再画轨迹。
     */
    suspend fun selectHistoryOrder(order: OrderRecord) {
        val key = order.id.ifBlank { "${order.carId}|${order.startTime}" }
        _state.value = _state.value.copy(
            selectedOrder = order,
            trackFitNonce = _state.value.trackFitNonce + 1,
        )
        if (order.trajectory.size >= 2) return
        val orderId = order.id.trim()
        if (orderId.isEmpty() || orderId == "0") return
        _state.value = _state.value.copy(trackLoading = true)
        when (val detail = repository.orderDetail(orderId)) {
            is OpsResult.Ok -> {
                val merged = order.copy(
                    trajectory = detail.value.trajectory.ifEmpty { order.trajectory },
                    startLat = detail.value.startLat ?: order.startLat,
                    startLng = detail.value.startLng ?: order.startLng,
                    endLat = detail.value.endLat ?: order.endLat,
                    endLng = detail.value.endLng ?: order.endLng,
                )
                replaceOrderInLists(merged, key)
                _state.value = _state.value.copy(
                    selectedOrder = merged,
                    trackLoading = false,
                    trackFitNonce = _state.value.trackFitNonce + 1,
                )
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    trackLoading = false,
                    toastMessage = detail.error.message,
                )
            }
        }
    }

    private fun replaceOrderInLists(merged: OrderRecord, key: String) {
        fun same(o: OrderRecord): Boolean =
            o.id.ifBlank { "${o.carId}|${o.startTime}" } == key
        val last = _state.value.lastOrder
        val history = _state.value.history.map { if (same(it)) merged else it }
        _state.value = _state.value.copy(
            lastOrder = if (last != null && same(last)) merged else last,
            history = history,
        )
    }

    fun openWallet() {
        val pin = scopeUserPin ?: return
        _state.value = _state.value.copy(nav = OrderQueryNav.UserWallet(pin))
    }

    fun openDeposit() {
        val pin = scopeUserPin ?: return
        _state.value = _state.value.copy(nav = OrderQueryNav.UserDeposit(pin))
    }

    fun openRideCard() {
        val pin = scopeUserPin ?: return
        _state.value = _state.value.copy(nav = OrderQueryNav.UserRideCard(pin))
    }

    fun openUserHistory() {
        val pin = scopeUserPin ?: return
        _state.value = _state.value.copy(nav = OrderQueryNav.UserOrderHistory(pin))
    }

    suspend fun loadWallet() {
        val pin = scopeUserPin ?: return
        _state.value = _state.value.copy(assetLoading = true)
        when (val info = repository.getWalletInfo(pin)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(walletInfo = info.value)
            is OpsResult.Err -> _state.value = _state.value.copy(toastMessage = info.error.message)
        }
        when (val rec = repository.getWalletRecords(pin)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(walletRecords = rec.value)
            is OpsResult.Err -> _state.value = _state.value.copy(toastMessage = rec.error.message)
        }
        _state.value = _state.value.copy(assetLoading = false)
    }

    suspend fun loadDeposit() {
        val pin = scopeUserPin ?: return
        _state.value = _state.value.copy(assetLoading = true)
        val serviceId = currentArea?.id.orEmpty()
        when (val assets = repository.getUserAssets(pin, serviceId)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(userAssets = assets.value)
            is OpsResult.Err -> Unit
        }
        when (val rec = repository.getDepositRecords(pin)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(depositRecords = rec.value)
            is OpsResult.Err -> _state.value = _state.value.copy(toastMessage = rec.error.message)
        }
        _state.value = _state.value.copy(assetLoading = false)
    }

    suspend fun loadRideCards() {
        val pin = scopeUserPin ?: return
        _state.value = _state.value.copy(assetLoading = true)
        when (val cards = repository.getRideCards(pin)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(rideCards = cards.value)
            is OpsResult.Err -> _state.value = _state.value.copy(toastMessage = cards.error.message)
        }
        _state.value = _state.value.copy(assetLoading = false)
    }

    suspend fun loadRideRecords() {
        val pin = scopeUserPin ?: return
        _state.value = _state.value.copy(assetLoading = true)
        when (val rec = repository.getRideCardRecords(pin)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(rideRecords = rec.value)
            is OpsResult.Err -> _state.value = _state.value.copy(toastMessage = rec.error.message)
        }
        _state.value = _state.value.copy(assetLoading = false)
    }

    suspend fun editWalletPresentYuan(yuan: Double): Boolean {
        val pin = scopeUserPin ?: return false
        val fen = (yuan * 100).toInt()
        return when (val result = repository.editWalletPresent(pin, fen)) {
            is OpsResult.Ok -> {
                loadWallet()
                true
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(toastMessage = result.error.message)
                false
            }
        }
    }

    suspend fun refundableFen(tradeNo: String, paidAt: String): Int? {
        return when (val result = repository.refundableAmount(tradeNo, paidAt)) {
            is OpsResult.Ok -> {
                if (result.value == 0) {
                    _state.value = _state.value.copy(toastMessage = Strings.t(Str.OrderQueryNoRefundAmount))
                    null
                } else {
                    result.value
                }
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(toastMessage = result.error.message)
                null
            }
        }
    }

    suspend fun refundWallet(record: OrderWalletRecord, yuan: Double): Boolean {
        val pin = scopeUserPin ?: return false
        val fen = (yuan * 100).toInt()
        return when (
            val result = repository.refundWallet(pin, fen, record.merchantTradeNo, record.paidAt)
        ) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(toastMessage = Strings.t(Str.OrderQueryRefundSuccess))
                loadWallet()
                true
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(toastMessage = result.error.message)
                false
            }
        }
    }

    suspend fun refundRide(record: OrderRideCardRecord, yuan: Double): Boolean {
        val pin = scopeUserPin ?: return false
        val fen = (yuan * 100).toInt()
        return when (val result = repository.ridingRefund(pin, fen, record.merchantTradeNo)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(toastMessage = Strings.t(Str.OrderQueryRefundSuccess))
                loadRideRecords()
                true
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(toastMessage = result.error.message)
                false
            }
        }
    }

    suspend fun resetCarStatus() {
        val pin = scopeUserPin ?: return
        when (val result = repository.resetCarStatus(pin)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(toastMessage = Strings.t(Str.OrderQueryResetSuccess))
                loadDeposit()
            }
            is OpsResult.Err -> _state.value = _state.value.copy(toastMessage = result.error.message)
        }
    }

    suspend fun manualReturnDeposit() {
        val pin = scopeUserPin ?: return
        val fen = (_state.value.userAssets?.depositedMountFen ?: 0L).toInt()
        when (val result = repository.manualReturnDeposit(pin, fen)) {
            is OpsResult.Ok -> {
                val serviceId = currentArea?.id.orEmpty()
                when (val user = repository.getUserByPin(pin)) {
                    is OpsResult.Ok -> _state.value = _state.value.copy(scopeUser = user.value)
                    is OpsResult.Err -> Unit
                }
                loadDeposit()
            }
            is OpsResult.Err -> _state.value = _state.value.copy(toastMessage = result.error.message)
        }
    }

    /** 对齐 UserDetailInfoViewModel.endTrip。 */
    suspend fun endTrip() {
        val order = _state.value.lastOrder ?: return
        val pin = order.userPin.ifBlank { scopeUserPin.orEmpty() }
        val carId = order.carId.trim()
        if (pin.isBlank() || carId.isBlank()) {
            _state.value = _state.value.copy(toastMessage = Strings.t(Str.OrderQueryEndTripMissing))
            return
        }
        _state.value = _state.value.copy(actionLoading = true)
        when (val result = repository.focusReturn(pin, carId)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    actionLoading = false,
                    toastMessage = Strings.t(Str.OrderQueryEndTripSuccess),
                )
                refreshUserScope()
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    actionLoading = false,
                    toastMessage = result.error.message,
                )
            }
        }
    }

    /**
     * 结束行程并修改金额（分）。
     * 对齐 ModifyAmountViewModel.focusReturnWithCost。
     */
    suspend fun endTripWithCost(
        modifyPayCostFen: Int,
        modifyDispatchCostFen: Int,
        modifyHelmetPenaltyFen: Int,
    ): Boolean {
        val order = _state.value.lastOrder ?: return false
        val pin = order.userPin.ifBlank { scopeUserPin.orEmpty() }
        val carId = order.carId.trim()
        if (pin.isBlank() || carId.isBlank()) {
            _state.value = _state.value.copy(toastMessage = Strings.t(Str.OrderQueryEndTripMissing))
            return false
        }
        _state.value = _state.value.copy(actionLoading = true)
        return when (
            val result = repository.focusReturnWithCost(
                pin,
                carId,
                modifyPayCostFen,
                modifyDispatchCostFen,
                modifyHelmetPenaltyFen,
            )
        ) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    actionLoading = false,
                    toastMessage = Strings.t(Str.OrderQueryEndTripSuccess),
                    nav = OrderQueryNav.UserOrders(pin),
                )
                refreshUserScope()
                true
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    actionLoading = false,
                    toastMessage = result.error.message,
                )
                false
            }
        }
    }

    /**
     * 未支付订单修改金额（分）。
     * 对齐 UserDetailInfoViewModel.updateAmount → createUpdateCostTicket。
     */
    suspend fun modifyCost(
        modifyPayCostFen: Int,
        modifyDispatchCostFen: Int,
        modifyHelmetPenaltyFen: Int,
    ): Boolean {
        val orderId = _state.value.lastOrder?.id?.trim().orEmpty()
        if (orderId.isEmpty() || orderId == "0") {
            _state.value = _state.value.copy(toastMessage = Strings.t(Str.OrderQueryEndTripMissing))
            return false
        }
        _state.value = _state.value.copy(actionLoading = true)
        return when (
            val result = repository.modifyOrderCost(
                orderId,
                modifyPayCostFen,
                modifyDispatchCostFen,
                modifyHelmetPenaltyFen,
            )
        ) {
            is OpsResult.Ok -> {
                val pin = scopeUserPin.orEmpty()
                _state.value = _state.value.copy(
                    actionLoading = false,
                    toastMessage = Strings.t(Str.OrderQueryModifyAmountSuccess),
                    nav = if (pin.isNotBlank()) OrderQueryNav.UserOrders(pin) else _state.value.nav,
                )
                refreshUserScope()
                true
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    actionLoading = false,
                    toastMessage = result.error.message,
                )
                false
            }
        }
    }

    /** [minutes] 分钟，内部转秒；对齐 temporaryPowerOn(it.toInt() * 60)。 */
    suspend fun tempUnlock(minutes: Int) {
        val carId = _state.value.lastOrder?.carId?.trim().orEmpty()
        if (carId.isBlank()) {
            _state.value = _state.value.copy(toastMessage = Strings.t(Str.OrderQueryEndTripMissing))
            return
        }
        if (minutes <= 0) {
            _state.value = _state.value.copy(toastMessage = Strings.t(Str.OrderQueryPleaseEnterAmount))
            return
        }
        _state.value = _state.value.copy(actionLoading = true)
        when (val result = repository.temporaryUnlock(carId, minutes * 60)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    actionLoading = false,
                    toastMessage = Strings.t(Str.OrderQueryTempUnlockSuccess),
                )
                refreshUserScope()
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    actionLoading = false,
                    toastMessage = result.error.message,
                )
            }
        }
    }

    /** 临停「启动」→ tools/start。 */
    suspend fun startVehicle() {
        val carId = _state.value.lastOrder?.carId?.trim().orEmpty()
        if (carId.isBlank()) {
            _state.value = _state.value.copy(toastMessage = Strings.t(Str.OrderQueryEndTripMissing))
            return
        }
        _state.value = _state.value.copy(actionLoading = true)
        when (val result = repository.startVehicle(carId)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    actionLoading = false,
                    toastMessage = Strings.t(Str.OrderQueryStartVehicleSuccess),
                )
                refreshUserScope()
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    actionLoading = false,
                    toastMessage = result.error.message,
                )
            }
        }
    }

    fun openModifyAmount(fromEndTrip: Boolean) {
        if (scopeUserPin.isNullOrBlank() && _state.value.lastOrder == null) return
        _state.value = _state.value.copy(nav = OrderQueryNav.ModifyAmount(fromEndTrip))
    }

    /** 结束后刷新用户详情 + 最近一单 + 历史列表。 */
    private suspend fun refreshUserScope() {
        val pin = scopeUserPin ?: return
        historyPage = 1
        when (val user = repository.getUserByPin(pin)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(scopeUser = user.value)
            is OpsResult.Err -> Unit
        }
        val serviceId = currentArea?.id.orEmpty()
        when (val assets = repository.getUserAssets(pin, serviceId)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(userAssets = assets.value)
            is OpsResult.Err -> Unit
        }
        when (val last = repository.detailLastRecord(userPin = pin)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    lastOrder = last.value,
                    selectedOrder = last.value,
                    history = emptyList(),
                    historyFinished = false,
                    historyError = null,
                )
            }
            is OpsResult.Err -> Unit
        }
        loadMoreHistory()
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
