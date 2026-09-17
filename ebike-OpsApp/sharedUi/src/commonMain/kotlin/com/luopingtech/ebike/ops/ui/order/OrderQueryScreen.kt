package com.luopingtech.ebike.ops.ui.order

import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.Slider
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateMapOf
import androidx.compose.runtime.mutableFloatStateOf
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalUriHandler
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.domain.model.TrackPoint
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.order.OrderFormat
import com.luopingtech.ebike.ops.domain.order.OrderPayStates
import com.luopingtech.ebike.ops.domain.order.buildOrderHistoryPins
import com.luopingtech.ebike.ops.domain.order.formatOrderTrackClock
import com.luopingtech.ebike.ops.ui.icons.OpsIcon
import com.luopingtech.ebike.ops.ui.icons.painterResource
import com.luopingtech.ebike.ops.domain.order.OrderRecord
import com.luopingtech.ebike.ops.domain.order.OrderUserAssets
import com.luopingtech.ebike.ops.domain.order.OrderUserDetail
import com.luopingtech.ebike.ops.domain.order.OrderUserPageItem
import com.luopingtech.ebike.ops.feature.order.OrderQueryFeature
import com.luopingtech.ebike.ops.feature.order.OrderQueryNav
import com.luopingtech.ebike.ops.ui.icons.OpsBackChevron
import com.luopingtech.ebike.ops.ui.map.OpsMapSpec
import com.luopingtech.ebike.ops.ui.map.OpsMapView
import com.luopingtech.ebike.ops.ui.theme.OpsColors
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.launch

/**
 * 订单查询（CMP）。可跳原生车辆详情；不再走 H5。
 *
 * [initialCarId] 对齐遗留 OrderHistoryActivity(carId)：从车辆详情「订单信息」进入时直达该车订单。
 */
@Composable
fun OrderQueryScreen(
    app: OpsApp,
    currentArea: ServiceArea?,
    onClose: () -> Unit,
    onOpenVehicleDetail: (String) -> Unit = {},
    initialCarId: String? = null,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.orderQueryFeature.state.collectAsState()
    val colors = OpsTheme.colors
    val scope = rememberCoroutineScope()
    val feature = app.orderQueryFeature
    val uriHandler = LocalUriHandler.current
    val entryCarId = initialCarId?.trim()?.takeIf { it.isNotBlank() }

    DisposableEffect(Unit) {
        onDispose { feature.clear() }
    }

    LaunchedEffect(currentArea?.id, entryCarId) {
        if (entryCarId != null) {
            // 车辆详情入口：不要走 open() 重置回搜索首页。
            feature.bindArea(currentArea)
            feature.openVehicleFromHome(entryCarId)
        } else {
            feature.open(currentArea)
        }
    }

    fun onScopeBack() {
        if (entryCarId != null) {
            // 对齐遗留：从详情进订单，返回应关掉订单页回到详情，而不是订单搜索首页。
            onClose()
        } else {
            feature.navigateBack()
        }
    }

    Box(modifier = Modifier.fillMaxSize().background(Color.White)) {
        when (val nav = state.nav) {
            OrderQueryNav.Search -> SearchPage(
                app = app,
                t = { key, args -> t(key, *args) },
                onClose = onClose,
                onSearch = { scope.launch { feature.search() } },
                onLoadMore = { scope.launch { feature.loadMoreHome() } },
                onOpenHomeOrder = { order ->
                    // 对齐原版：骑行/临停 → 车辆详情；已完结 → 用户详情
                    if (OrderFormat.isRidingLike(order.carState)) {
                        val carId = order.carId.trim()
                        if (carId.isNotBlank()) onOpenVehicleDetail(carId)
                    } else {
                        val pin = order.userPin.trim()
                        if (pin.isNotBlank()) {
                            scope.launch { feature.openUserFromHome(pin) }
                        }
                    }
                },
            )
            is OrderQueryNav.UserPicker -> UserPickerPage(
                t = { key, args -> t(key, *args) },
                name = nav.name,
                users = state.pickerUsers,
                loading = state.pickerLoading,
                finished = state.pickerFinished,
                error = state.pickerError,
                onBack = { feature.navigateBack() },
                onLoadMore = { scope.launch { feature.loadMorePicker() } },
                onPick = { pin -> scope.launch { feature.openUserFromPicker(pin) } },
            )
            is OrderQueryNav.UserOrders -> ScopePage(
                app = app,
                t = { key, args -> t(key, *args) },
                title = t(Str.OrderQueryUserDetailTitle),
                user = state.scopeUser,
                userAssets = state.userAssets,
                carId = null,
                imei = null,
                lastOrder = state.lastOrder,
                history = state.history,
                loading = state.scopeLoading || state.historyLoading,
                finished = state.historyFinished,
                error = state.historyError,
                onBack = { feature.navigateBack() },
                onLoadMore = { scope.launch { feature.loadMoreHistory() } },
                onOpenVehicle = onOpenVehicleDetail,
                onCallPhone = { phone ->
                    val raw = phone.trim().removePrefix("+86-").removePrefix("+86")
                    if (raw.isNotBlank()) {
                        runCatching { uriHandler.openUri("tel:$raw") }
                    }
                },
                onOpenDeposit = { feature.openDeposit() },
                onOpenRideCard = { feature.openRideCard() },
                onOpenWallet = { feature.openWallet() },
                onLookMore = {
                    feature.openUserHistory()
                    scope.launch { feature.loadMoreHistory() }
                },
                showInProgressActions = true,
                actionLoading = state.actionLoading,
                onStartVehicle = { scope.launch { feature.startVehicle() } },
                onTempUnlock = { minutes -> scope.launch { feature.tempUnlock(minutes) } },
                onEndTrip = { scope.launch { feature.endTrip() } },
                onEndTripAndModify = { feature.openModifyAmount(fromEndTrip = true) },
                onModifyAmount = { feature.openModifyAmount(fromEndTrip = false) },
            )
            is OrderQueryNav.ModifyAmount -> ModifyAmountPage(
                t = { key, args -> t(key, *args) },
                feature = feature,
                order = state.lastOrder,
                fromEndTrip = nav.fromEndTrip,
                loading = state.actionLoading,
                onBack = { feature.navigateBack() },
            )
            is OrderQueryNav.UserWallet -> UserWalletPage(
                t = { key, args -> t(key, *args) },
                feature = feature,
                wallet = state.walletInfo,
                records = state.walletRecords,
                loading = state.assetLoading,
                onBack = { feature.navigateBack() },
            )
            is OrderQueryNav.UserDeposit -> UserDepositPage(
                t = { key, args -> t(key, *args) },
                feature = feature,
                user = state.scopeUser,
                assets = state.userAssets,
                records = state.depositRecords,
                loading = state.assetLoading,
                onBack = { feature.navigateBack() },
            )
            is OrderQueryNav.UserRideCard -> UserRideCardPage(
                t = { key, args -> t(key, *args) },
                feature = feature,
                cards = state.rideCards,
                records = state.rideRecords,
                loading = state.assetLoading,
                onBack = { feature.navigateBack() },
            )
            is OrderQueryNav.UserOrderHistory -> ScopePage(
                app = app,
                t = { key, args -> t(key, *args) },
                title = t(Str.OrderRecordTitle),
                user = null,
                userAssets = null,
                carId = null,
                imei = null,
                lastOrder = state.lastOrder,
                history = state.history,
                loading = state.scopeLoading || state.historyLoading,
                finished = state.historyFinished,
                error = state.historyError,
                onBack = { feature.navigateBack() },
                onLoadMore = { scope.launch { feature.loadMoreHistory() } },
                onOpenVehicle = onOpenVehicleDetail,
                legacyVehicleStyle = true,
                listSectionTitle = t(Str.OrderRecordTitle),
                selectedOrder = state.selectedOrder,
                trackLoading = state.trackLoading,
                trackFitNonce = state.trackFitNonce,
                onSelectOrder = { order -> scope.launch { feature.selectHistoryOrder(order) } },
            )
            is OrderQueryNav.VehicleOrders -> ScopePage(
                app = app,
                t = { key, args -> t(key, *args) },
                title = nav.carId?.ifBlank { null }
                    ?: nav.imei?.ifBlank { null }
                    ?: t(Str.OrderQueryScreen),
                user = null,
                userAssets = null,
                carId = nav.carId,
                imei = nav.imei,
                lastOrder = state.lastOrder,
                history = state.history,
                loading = state.scopeLoading || state.historyLoading,
                finished = state.historyFinished,
                error = state.historyError,
                onBack = { onScopeBack() },
                onLoadMore = { scope.launch { feature.loadMoreHistory() } },
                onOpenVehicle = onOpenVehicleDetail,
                legacyVehicleStyle = true,
                selectedOrder = state.selectedOrder,
                trackLoading = state.trackLoading,
                trackFitNonce = state.trackFitNonce,
                onSelectOrder = { order -> scope.launch { feature.selectHistoryOrder(order) } },
            )
        }

        state.toastMessage?.let { msg ->
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(bottom = 48.dp),
                contentAlignment = Alignment.BottomCenter,
            ) {
                Text(
                    text = msg,
                    color = Color.White,
                    fontSize = 13.sp,
                    modifier = Modifier
                        .background(Color(0xCC333333), RoundedCornerShape(8.dp))
                        .padding(horizontal = 14.dp, vertical = 10.dp)
                        .clickable { feature.consumeToast() },
                )
            }
            LaunchedEffect(msg) {
                delay(2200)
                feature.consumeToast()
            }
        }
    }
}

private typealias Tr = (Str, Array<out Any?>) -> String

private val PageWhite = Color.White
private val TipGray = Color(0xFF666666)
private val SearchStroke = Color(0xFFD7D7D7)
private val QueryBlue = Color(0xFF1180F9)
private val CellLabel = Color(0xFF7C87B1)
private val CellValue = Color(0xFF333333)
private val TextPrimary = Color(0xFF242936)
private val SectionGap = Color(0xFFEEF1FA)
private val DividerColor = Color(0xFFD9DCE6)
private val LookMoreGray = Color(0xFF999999)

@Composable
private fun SearchPage(
    app: OpsApp,
    t: Tr,
    onClose: () -> Unit,
    onSearch: () -> Unit,
    onLoadMore: () -> Unit,
    onOpenHomeOrder: (OrderRecord) -> Unit,
) {
    val state by app.orderQueryFeature.state.collectAsState()
    val colors = OpsTheme.colors
    val placeholder = when {
        state.canQueryPerson && state.canQueryVehicle -> t(Str.OrderQuerySearchHintBoth, emptyArray())
        state.canQueryPerson -> t(Str.OrderQuerySearchHintPerson, emptyArray())
        state.canQueryVehicle -> t(Str.OrderQuerySearchHintVehicle, emptyArray())
        else -> ""
    }
    val listState = rememberLazyListState()
    var selectedIndex by remember { mutableIntStateOf(-1) }
    val canSearch = placeholder.isNotEmpty() && state.keyword.isNotBlank() && !state.searching

    LaunchedEffect(listState, state.loading, state.finished) {
        snapshotFlow {
            val info = listState.layoutInfo
            val last = info.visibleItemsInfo.lastOrNull()?.index ?: 0
            last >= info.totalItemsCount - 2 && info.totalItemsCount > 0
        }.distinctUntilChanged().collect { nearEnd ->
            if (nearEnd && !state.loading && !state.finished) onLoadMore()
        }
    }

    Column(modifier = Modifier.fillMaxSize().background(PageWhite)) {
        TopBar(title = t(Str.OrderQueryScreen, emptyArray()), onBack = onClose, colors = colors)

        // 对齐 activity_order_search：描边圆角搜索框 + 查询按钮
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp)
                .padding(top = 24.dp)
                .height(40.dp)
                .border(1.dp, SearchStroke, RoundedCornerShape(4.dp)),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            BasicTextField(
                value = state.keyword,
                onValueChange = { app.orderQueryFeature.setKeyword(it) },
                singleLine = true,
                enabled = placeholder.isNotEmpty(),
                textStyle = TextStyle(fontSize = 14.sp, color = TextPrimary),
                cursorBrush = SolidColor(QueryBlue),
                keyboardOptions = KeyboardOptions(imeAction = ImeAction.Search),
                keyboardActions = KeyboardActions(onSearch = { if (canSearch) onSearch() }),
                modifier = Modifier
                    .weight(1f)
                    .padding(horizontal = 10.dp),
                decorationBox = { inner ->
                    if (state.keyword.isEmpty()) {
                        Text(
                            placeholder.ifEmpty { t(Str.OrderQueryNoPerm, emptyArray()) },
                            color = Color(0xFF999999),
                            fontSize = 14.sp,
                        )
                    }
                    inner()
                },
            )
            Box(
                modifier = Modifier
                    .width(70.dp)
                    .fillMaxSize()
                    .clickable(
                        enabled = canSearch,
                        interactionSource = remember { MutableInteractionSource() },
                        indication = null,
                        onClick = onSearch,
                    ),
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    if (state.searching) {
                        t(Str.LoadingEllipsis, emptyArray())
                    } else {
                        t(Str.OrderQuerySearchAction, emptyArray())
                    },
                    color = if (canSearch || state.searching) QueryBlue else Color(0xFFCCCCCC),
                    fontSize = 16.sp,
                )
            }
        }

        if (placeholder.isEmpty()) {
            Text(
                t(Str.OrderQueryNoPerm, emptyArray()),
                modifier = Modifier.padding(40.dp).fillMaxWidth(),
                color = Color(0xFF999999),
                fontSize = 14.sp,
            )
        } else {
            LazyColumn(
                state = listState,
                modifier = Modifier
                    .weight(1f)
                    .fillMaxWidth()
                    .padding(top = 12.dp),
            ) {
                // 对齐 AppBar 可滚走的搜索说明
                item(key = "tips") {
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(bottom = 12.dp),
                    ) {
                        Text(
                            t(Str.OrderQuerySearchInstructions, emptyArray()),
                            color = TipGray,
                            fontSize = 14.sp,
                            modifier = Modifier.padding(start = 28.dp, top = 24.dp),
                        )
                        Text(
                            t(Str.OrderQueryTip1, emptyArray()),
                            color = TipGray,
                            fontSize = 14.sp,
                            modifier = Modifier.padding(start = 28.dp, top = 8.dp),
                        )
                        Text(
                            t(Str.OrderQueryTip2, emptyArray()),
                            color = TipGray,
                            fontSize = 14.sp,
                            modifier = Modifier.padding(start = 28.dp, top = 8.dp),
                        )
                        Text(
                            t(Str.OrderQueryTip3, emptyArray()),
                            color = TipGray,
                            fontSize = 14.sp,
                            modifier = Modifier.padding(start = 28.dp, top = 8.dp),
                        )
                    }
                }

                itemsIndexed(
                    state.items,
                    key = { _, it -> it.id.ifBlank { "${it.carId}-${it.startTime}" } },
                ) { index, order ->
                    val riding = OrderFormat.isRidingLike(order.carState)
                    val clickable = when {
                        riding -> state.canQueryVehicle && order.carId.isNotBlank()
                        else -> state.canQueryPerson && order.userPin.isNotBlank()
                    }
                    OrderHistoryListItem(
                        order = order,
                        t = t,
                        selected = index == selectedIndex,
                        startAddress = "--",
                        endAddress = "--",
                        onClick = {
                            selectedIndex = index
                            if (clickable) onOpenHomeOrder(order)
                        },
                    )
                }
                item {
                    when {
                        state.loading -> Box(
                            Modifier.fillMaxWidth().padding(16.dp),
                            contentAlignment = Alignment.Center,
                        ) { CircularProgressIndicator(color = colors.primary) }
                        state.errorMessage != null -> Text(
                            state.errorMessage!!,
                            modifier = Modifier.padding(16.dp),
                            color = Color(0xFFE53935),
                            fontSize = 13.sp,
                        )
                        state.finished && state.items.isEmpty() -> Text(
                            t(Str.AdminEmptyList, emptyArray()),
                            modifier = Modifier.padding(40.dp).fillMaxWidth(),
                            color = Color(0xFF999999),
                            fontSize = 14.sp,
                        )
                        state.finished -> Text(
                            t(Str.VehicleListEnd, emptyArray()),
                            modifier = Modifier.padding(16.dp).fillMaxWidth(),
                            color = Color(0xFF999999),
                            fontSize = 12.sp,
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun UserPickerPage(
    t: Tr,
    name: String,
    users: List<OrderUserPageItem>,
    loading: Boolean,
    finished: Boolean,
    error: String?,
    onBack: () -> Unit,
    onLoadMore: () -> Unit,
    onPick: (String) -> Unit,
) {
    val colors = OpsTheme.colors
    val listState = rememberLazyListState()
    LaunchedEffect(listState, loading, finished) {
        snapshotFlow {
            val info = listState.layoutInfo
            val last = info.visibleItemsInfo.lastOrNull()?.index ?: 0
            last >= info.totalItemsCount - 2 && info.totalItemsCount > 0
        }.distinctUntilChanged().collect { nearEnd ->
            if (nearEnd && !loading && !finished) onLoadMore()
        }
    }

    Column(modifier = Modifier.fillMaxSize().background(PageWhite)) {
        TopBar(title = t(Str.OrderQueryPickUser, emptyArray()), onBack = onBack, colors = colors)
        Text(
            "「$name」",
            modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp),
            color = TipGray,
            fontSize = 13.sp,
        )
        LazyColumn(state = listState, modifier = Modifier.weight(1f)) {
            items(users, key = { it.pin }) { user ->
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .clickable { onPick(user.pin) }
                        .padding(horizontal = 16.dp, vertical = 12.dp),
                ) {
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Text(
                            OrderFormat.orDash(user.authName),
                            color = TextPrimary,
                            fontWeight = FontWeight.Bold,
                            fontSize = 14.sp,
                        )
                        Text(
                            OrderFormat.ridingStateLabel(user.ridingState),
                            color = TipGray,
                            fontSize = 12.sp,
                        )
                    }
                    Spacer(Modifier.height(8.dp))
                    UserInfoCell(
                        label = t(Str.Phone, emptyArray()),
                        value = OrderFormat.orDash(user.phone),
                        valueColor = QueryBlue,
                    )
                    UserInfoCell(
                        label = t(Str.OrderQueryWalletBalance, emptyArray()),
                        value = OrderFormat.yuanWithUnit(user.balance),
                    )
                }
                HorizontalDivider(color = DividerColor, thickness = 1.dp)
            }
            item {
                when {
                    loading -> Box(Modifier.fillMaxWidth().padding(16.dp), contentAlignment = Alignment.Center) {
                        CircularProgressIndicator(color = colors.primary)
                    }
                    error != null -> Text(error, modifier = Modifier.padding(16.dp), color = Color(0xFFE53935))
                    finished && users.isEmpty() -> Text(
                        t(Str.AdminEmptyList, emptyArray()),
                        modifier = Modifier.padding(40.dp),
                        color = Color(0xFF999999),
                    )
                }
            }
        }
    }
}

@Composable
private fun ScopePage(
    app: OpsApp,
    t: Tr,
    title: String,
    user: OrderUserDetail?,
    userAssets: OrderUserAssets? = null,
    carId: String?,
    imei: String?,
    lastOrder: OrderRecord?,
    history: List<OrderRecord>,
    loading: Boolean,
    finished: Boolean,
    error: String?,
    onBack: () -> Unit,
    onLoadMore: () -> Unit,
    onOpenVehicle: (String) -> Unit,
    onCallPhone: (String) -> Unit = {},
    onOpenDeposit: () -> Unit = {},
    onOpenRideCard: () -> Unit = {},
    onOpenWallet: () -> Unit = {},
    onLookMore: () -> Unit = {},
    legacyVehicleStyle: Boolean = false,
    listSectionTitle: String? = null,
    selectedOrder: OrderRecord? = null,
    trackLoading: Boolean = false,
    trackFitNonce: Int = 0,
    onSelectOrder: (OrderRecord) -> Unit = {},
    showInProgressActions: Boolean = false,
    actionLoading: Boolean = false,
    onStartVehicle: () -> Unit = {},
    onTempUnlock: (Int) -> Unit = {},
    onEndTrip: () -> Unit = {},
    onEndTripAndModify: () -> Unit = {},
    onModifyAmount: () -> Unit = {},
) {
    val colors = OpsTheme.colors
    val listState = rememberLazyListState()
    var showEndTripSheet by remember { mutableStateOf(false) }
    var showTempUnlockDialog by remember { mutableStateOf(false) }
    LaunchedEffect(listState, loading, finished) {
        snapshotFlow {
            val info = listState.layoutInfo
            val last = info.visibleItemsInfo.lastOrNull()?.index ?: 0
            last >= info.totalItemsCount - 2 && info.totalItemsCount > 0
        }.distinctUntilChanged().collect { nearEnd ->
            if (nearEnd && !loading && !finished) onLoadMore()
        }
    }

    val orders = remember(lastOrder, history) {
        buildList {
            lastOrder?.let { add(it) }
            addAll(history.filter { it.id != lastOrder?.id || it.id.isBlank() })
        }.distinctBy { it.id.ifBlank { "${it.carId}-${it.startTime}" } }
    }

    fun orderKey(o: OrderRecord) = o.id.ifBlank { "${o.carId}|${o.startTime}" }
    val selectedKey = selectedOrder?.let { orderKey(it) }
    val historySelectedIndex = orders.indexOfFirst { orderKey(it) == selectedKey }.coerceAtLeast(0)
    val track = selectedOrder?.trajectory.orEmpty()
    var playbackProgress by remember(selectedKey) { mutableFloatStateOf(0f) }
    var previewSelected by remember { mutableIntStateOf(0) }
    var showLegend by remember { mutableStateOf(false) }
    var locateNonce by remember { mutableIntStateOf(0) }
    val addressCache = remember { mutableStateMapOf<String, String>() }
    val mapPins = remember(selectedOrder, track, playbackProgress) {
        buildOrderHistoryPins(selectedOrder, track, playbackProgress)
    }

    val inProgress = showInProgressActions &&
        OrderFormat.showInProgressActions(user?.ridingState, lastOrder?.carState)
    val showStart = inProgress &&
        OrderFormat.isTempParking(user?.ridingState, lastOrder?.carState)
    val showModifyAmount = showInProgressActions &&
        !inProgress &&
        lastOrder?.izPaid == OrderPayStates.ToPay

    LaunchedEffect(selectedOrder?.id, selectedOrder?.startLat, selectedOrder?.endLat) {
        val order = selectedOrder ?: return@LaunchedEffect
        resolveOrderAddress(app, order.startLat, order.startLng, addressCache)
        resolveOrderAddress(app, order.endLat, order.endLng, addressCache)
    }

    Column(modifier = Modifier.fillMaxSize().background(PageWhite)) {
        TopBar(
            title = if (legacyVehicleStyle) t(Str.OrderRecordTitle, emptyArray()) else title,
            onBack = onBack,
            colors = colors,
        )
        if (legacyVehicleStyle) {
            // 对齐 activity_order_history：顶部地图 + 轨迹回放条 + 列表
            Box(
                Modifier
                    .fillMaxWidth()
                    .height(308.dp)
                    .background(Color(0xFFE8ECF4)),
            ) {
                OpsMapView(
                    spec = OpsMapSpec(
                        pins = mapPins,
                        trackPoints = track,
                        clusterOverview = false,
                        autoFitOnPins = true,
                        showStatusOverlay = false,
                        followNonce = locateNonce,
                        followLat = selectedOrder?.endLat ?: selectedOrder?.startLat
                            ?: track.lastOrNull()?.lat,
                        followLng = selectedOrder?.endLng ?: selectedOrder?.startLng
                            ?: track.lastOrNull()?.lng,
                        fitNonce = if (track.size >= 2) trackFitNonce + locateNonce else locateNonce,
                    ),
                    modifier = Modifier.fillMaxSize(),
                )
                Column(
                    modifier = Modifier
                        .align(Alignment.CenterStart)
                        .padding(start = 16.dp)
                        .background(Color.White, RoundedCornerShape(8.dp))
                        .width(40.dp),
                ) {
                    OrderMapSideTool(
                        icon = OpsIcon.MapLocation,
                        label = t(Str.MapToolLocate, emptyArray()),
                        onClick = { locateNonce++ },
                    )
                    HorizontalDivider(color = Color(0xFFE5E5E5), thickness = 1.dp)
                    OrderMapSideTool(
                        icon = OpsIcon.MapExplain,
                        label = t(Str.MapToolLegend, emptyArray()),
                        onClick = { showLegend = true },
                    )
                }
                if (trackLoading) {
                    CircularProgressIndicator(
                        color = colors.primary,
                        modifier = Modifier.align(Alignment.Center),
                    )
                }
            }
            if (track.size >= 2) {
                OrderTrackPlaybackBar(
                    points = track,
                    progress = playbackProgress,
                    onProgress = { playbackProgress = it },
                )
            }
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(44.dp)
                    .background(PageWhite)
                    .padding(horizontal = 16.dp),
                contentAlignment = Alignment.CenterStart,
            ) {
                Text(
                    listSectionTitle ?: t(Str.VehicleOrderHistory, emptyArray()),
                    color = TextPrimary,
                    fontSize = 14.sp,
                    fontWeight = FontWeight.Bold,
                )
            }
            HorizontalDivider(color = DividerColor, thickness = 1.dp)
            LazyColumn(state = listState, modifier = Modifier.weight(1f)) {
                items(orders.size) { index ->
                    val order = orders[index]
                    OrderHistoryListItem(
                        order = order,
                        t = t,
                        selected = index == historySelectedIndex,
                        startAddress = addressOfOrder(order.startLat, order.startLng, addressCache),
                        endAddress = addressOfOrder(order.endLat, order.endLng, addressCache),
                        onClick = { onSelectOrder(order) },
                    )
                }
                item { ScopeFooter(loading, error, finished && orders.isEmpty(), t, colors) }
            }
        } else {
            // 对齐 activity_user_info：地图 + 用户信息 cells + 订单记录
            LazyColumn(state = listState, modifier = Modifier.weight(1f)) {
                item(key = "map") {
                    val track = lastOrder?.trajectory.orEmpty()
                    val followLat = lastOrder?.endLat ?: lastOrder?.startLat
                    val followLng = lastOrder?.endLng ?: lastOrder?.startLng
                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(288.dp)
                            .background(Color(0xFFE8ECF4)),
                    ) {
                        OpsMapView(
                            spec = OpsMapSpec(
                                pins = buildOrderHistoryPins(lastOrder, track, 0f),
                                trackPoints = track,
                                clusterOverview = false,
                                autoFitOnPins = true,
                                showStatusOverlay = false,
                                followLat = followLat,
                                followLng = followLng,
                                followNonce = if (followLat != null && followLng != null) 1 else 0,
                                fitNonce = if (track.isNotEmpty()) 1 else 0,
                            ),
                            modifier = Modifier.fillMaxSize(),
                        )
                    }
                }

                if (user != null) {
                    item(key = "user-info") {
                        Column(modifier = Modifier.fillMaxWidth()) {
                            Text(
                                t(Str.OrderQueryUserInfoSection, emptyArray()),
                                color = TextPrimary,
                                fontSize = 14.sp,
                                fontWeight = FontWeight.Bold,
                                modifier = Modifier.padding(start = 16.dp, top = 24.dp),
                            )
                            Spacer(Modifier.height(8.dp))
                            UserInfoCell(
                                label = t(Str.OrderQueryNameLabel, emptyArray()),
                                value = OrderFormat.orDash(user.authName),
                            )
                            UserInfoCell(
                                label = t(Str.Phone, emptyArray()),
                                value = OrderFormat.orDash(user.phone),
                                valueColor = QueryBlue,
                                onClick = { onCallPhone(user.phone) },
                            )
                            UserInfoCell(
                                label = t(Str.OrderQueryIdCardLabel, emptyArray()),
                                value = OrderFormat.orDash(user.authNo),
                                valueColor = QueryBlue,
                            )
                            UserInfoCell(
                                label = t(Str.OrderQueryRegisterTime, emptyArray()),
                                value = OrderFormat.isoDateTime(user.createdAt).ifBlank {
                                    OrderFormat.orDash(user.createdAt)
                                },
                            )
                            UserInfoCell(
                                label = t(Str.OrderQueryDepositStatus, emptyArray()),
                                value = OrderFormat.depositStatus(user, userAssets),
                                showChevron = true,
                                onClick = onOpenDeposit,
                            )
                            UserInfoCell(
                                label = t(Str.OrderQueryRideCard, emptyArray()),
                                value = OrderFormat.ridingCardText(userAssets),
                                showChevron = true,
                                onClick = onOpenRideCard,
                            )
                            UserInfoCell(
                                label = t(Str.OrderQueryWalletBalance, emptyArray()),
                                value = OrderFormat.walletBalanceText(userAssets, user.balance),
                                showChevron = true,
                                onClick = onOpenWallet,
                            )
                            Spacer(Modifier.height(24.dp))
                        }
                    }
                } else {
                    item(key = "vehicle-meta") {
                        Column(
                            Modifier
                                .fillMaxWidth()
                                .padding(horizontal = 16.dp, vertical = 16.dp),
                        ) {
                            if (!carId.isNullOrBlank()) {
                                UserInfoCell(
                                    label = t(Str.OrderQueryCarLabel, emptyArray()),
                                    value = carId,
                                )
                            }
                            if (!imei.isNullOrBlank()) {
                                UserInfoCell(
                                    label = t(Str.OrderQueryImeiLabel, emptyArray()),
                                    value = imei,
                                )
                            }
                            val openId = carId?.takeIf { it.isNotBlank() }
                                ?: lastOrder?.carId?.takeIf { it.isNotBlank() }
                            if (openId != null) {
                                Spacer(Modifier.height(8.dp))
                                Text(
                                    t(Str.OrderQueryOpenVehicle, emptyArray()),
                                    color = QueryBlue,
                                    fontSize = 14.sp,
                                    modifier = Modifier.clickable { onOpenVehicle(openId) },
                                )
                            }
                        }
                    }
                }

                item(key = "order-section-gap") {
                    Box(
                        Modifier
                            .fillMaxWidth()
                            .height(8.dp)
                            .background(SectionGap),
                    )
                }

                item(key = "order-section-title") {
                    Row(
                        Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 16.dp)
                            .padding(top = 24.dp, bottom = 8.dp)
                            .clickable(
                                interactionSource = remember { MutableInteractionSource() },
                                indication = null,
                                onClick = onLookMore,
                            ),
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Text(
                            t(Str.OrderRecordTitle, emptyArray()),
                            color = TextPrimary,
                            fontSize = 14.sp,
                            fontWeight = FontWeight.Bold,
                            modifier = Modifier.weight(1f),
                        )
                        Text(
                            t(Str.OrderQueryLookMore, emptyArray()),
                            color = LookMoreGray,
                            fontSize = 14.sp,
                        )
                        Text("›", color = LookMoreGray, fontSize = 18.sp)
                    }
                }

                val preview = lastOrder ?: orders.firstOrNull()
                if (preview != null) {
                    item(key = "preview-order") {
                        OrderHistoryListItem(
                            order = preview,
                            t = t,
                            selected = previewSelected == 0,
                            startAddress = "--",
                            endAddress = "--",
                            onClick = {
                                previewSelected = 0
                                if (preview.carId.isNotBlank()) onOpenVehicle(preview.carId)
                            },
                        )
                    }
                }

                if (showModifyAmount) {
                    item(key = "modify-amount") {
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(horizontal = 32.dp, vertical = 20.dp)
                                .height(45.dp)
                                .border(1.dp, Color(0xFF999999), RoundedCornerShape(4.dp))
                                .clickable(
                                    enabled = !actionLoading,
                                    interactionSource = remember { MutableInteractionSource() },
                                    indication = null,
                                    onClick = onModifyAmount,
                                ),
                            contentAlignment = Alignment.Center,
                        ) {
                            Text(
                                t(Str.OrderQueryModifyAmount, emptyArray()),
                                color = TextPrimary,
                                fontSize = 14.sp,
                                fontWeight = FontWeight.Bold,
                            )
                        }
                    }
                }

                if (inProgress) {
                    item(key = "riding-actions") {
                        RidingActionBar(
                            t = t,
                            showStart = showStart,
                            enabled = !actionLoading,
                            onStart = onStartVehicle,
                            onTempUnlock = { showTempUnlockDialog = true },
                            onEndTrip = { showEndTripSheet = true },
                        )
                    }
                }

                item {
                    ScopeFooter(
                        loading = loading || actionLoading,
                        error = error,
                        empty = finished && orders.isEmpty(),
                        t = t,
                        colors = colors,
                    )
                }
            }
        }
    }
    if (showLegend) {
        OrderHistoryLegendOverlay(t = t, onClose = { showLegend = false })
    }
    if (showEndTripSheet) {
        EndTripBottomSheet(
            t = t,
            onDismiss = { showEndTripSheet = false },
            onEndOnly = {
                showEndTripSheet = false
                onEndTrip()
            },
            onEndAndModify = {
                showEndTripSheet = false
                onEndTripAndModify()
            },
        )
    }
    if (showTempUnlockDialog) {
        TempUnlockDialog(
            t = t,
            onDismiss = { showTempUnlockDialog = false },
            onConfirm = { minutes ->
                showTempUnlockDialog = false
                onTempUnlock(minutes)
            },
        )
    }
}

@Composable
fun OrderTrackPlaybackBar(
    points: List<TrackPoint>,
    progress: Float,
    onProgress: (Float) -> Unit,
) {
    val minTs = points.first().timestamp
    val maxTs = points.last().timestamp
    val span = (maxTs - minTs).coerceAtLeast(1L)
    val elapsed = (progress * span).toLong()
    val timeText = formatOrderTrackClock(minTs + elapsed)
    Column(
        Modifier
            .fillMaxWidth()
            .height(56.dp)
            .background(Color.White)
            .padding(horizontal = 12.dp),
    ) {
        Text(
            "时间: $timeText",
            color = Color(0xFF666666),
            fontSize = 10.sp,
            modifier = Modifier.padding(start = 8.dp, top = 4.dp),
        )
        Slider(
            value = progress,
            onValueChange = onProgress,
            modifier = Modifier.fillMaxWidth(),
        )
    }
}

@Composable
private fun OrderMapSideTool(
    icon: OpsIcon,
    label: String,
    onClick: () -> Unit,
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .height(58.dp)
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            ),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
    ) {
        Image(
            painter = painterResource(icon),
            contentDescription = label,
            modifier = Modifier.size(22.dp),
        )
        Text(label, color = Color(0xFF242936), fontSize = 10.sp)
    }
}

@Composable
private fun OrderHistoryLegendOverlay(
    t: Tr,
    onClose: () -> Unit,
) {
    val notes = listOf(
        "车辆定位" to "显示该车辆的实时定位位置，会存在少量定位偏差。",
        "最近行程轨迹" to "骑行中则显示骑行中轨迹；非骑行中则显示上一单骑行轨迹。",
        "车的行程起点" to "根据车辆定位，显示骑行轨迹的起点。",
        "车的行程终点" to "根据车辆定位，显示骑行轨迹的终点。",
        "用车人的起点" to "显示最近一单，用车时的手机 GPS 定位。",
        "还车人的终点" to "显示最近一单，还车时的手机 GPS 定位。",
    )
    Column(Modifier.fillMaxSize().background(Color.White)) {
        TopBar(title = t(Str.FunctionDescription, emptyArray()), onBack = onClose, colors = OpsTheme.colors)
        notes.forEach { (title, tip) ->
            Column(Modifier.fillMaxWidth().padding(horizontal = 20.dp, vertical = 12.dp)) {
                Text(title, fontWeight = FontWeight.Bold, fontSize = 15.sp, color = Color(0xFF242936))
                Text(tip, fontSize = 13.sp, color = Color(0xFF7C87B1), modifier = Modifier.padding(top = 4.dp))
            }
            HorizontalDivider(color = Color(0xFFEEF0F6))
        }
    }
}

private fun addressKeyOf(lat: Double?, lng: Double?): String? {
    if (lat == null || lng == null || (lat == 0.0 && lng == 0.0)) return null
    return "${lat},${lng}"
}

private fun addressOfOrder(
    lat: Double?,
    lng: Double?,
    cache: Map<String, String>,
): String {
    val key = addressKeyOf(lat, lng) ?: return "--"
    return cache[key] ?: "--"
}

private suspend fun resolveOrderAddress(
    app: OpsApp,
    lat: Double?,
    lng: Double?,
    cache: MutableMap<String, String>,
) {
    val key = addressKeyOf(lat, lng) ?: return
    if (cache.containsKey(key)) return
    val text = app.reverseGeocoder.addressOf(lat!!, lng!!).getOrNull()
        ?.takeIf { it.isNotBlank() && !it.equals("Ocean", ignoreCase = true) }
    cache[key] = text ?: "--"
}

@Composable
private fun ScopeFooter(
    loading: Boolean,
    error: String?,
    empty: Boolean,
    t: Tr,
    colors: OpsColors,
) {
    when {
        loading -> Box(Modifier.fillMaxWidth().padding(16.dp), contentAlignment = Alignment.Center) {
            CircularProgressIndicator(color = colors.primary)
        }
        error != null -> Text(error, modifier = Modifier.padding(16.dp), color = Color(0xFFE53935))
        empty -> Text(
            t(Str.AdminEmptyList, emptyArray()),
            modifier = Modifier.padding(40.dp),
            color = Color(0xFF999999),
        )
    }
}

/** 对齐 item_order_user ItemCellView：左灰标签、右值；诚信金/骑行卡/钱包带箭头。 */
@Composable
private fun UserInfoCell(
    label: String,
    value: String,
    valueColor: Color = CellValue,
    showChevron: Boolean = false,
    onClick: (() -> Unit)? = null,
) {
    Row(
        Modifier
            .fillMaxWidth()
            .then(
                if (onClick != null) {
                    Modifier.clickable(
                        interactionSource = remember { MutableInteractionSource() },
                        indication = null,
                        onClick = onClick,
                    )
                } else {
                    Modifier
                },
            )
            .padding(horizontal = 16.dp, vertical = 8.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(label, color = CellLabel, fontSize = 14.sp)
        Row(verticalAlignment = Alignment.CenterVertically) {
            Text(value, color = valueColor, fontSize = 14.sp)
            if (showChevron) {
                Text(" ›", color = LookMoreGray, fontSize = 16.sp)
            }
        }
    }
}

@Composable
private fun RidingActionBar(
    t: Tr,
    showStart: Boolean,
    enabled: Boolean,
    onStart: () -> Unit,
    onTempUnlock: () -> Unit,
    onEndTrip: () -> Unit,
) {
    val blue = Color(0xFF1180F9)
    val red = Color(0xFFFF0808)
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp)
            .padding(top = 10.dp, bottom = 20.dp),
        horizontalArrangement = Arrangement.SpaceEvenly,
        verticalAlignment = Alignment.CenterVertically,
    ) {
        if (showStart) {
            ActionOutlineButton(
                text = t(Str.OrderQueryStartVehicle, emptyArray()),
                color = blue,
                enabled = enabled,
                onClick = onStart,
            )
        }
        ActionOutlineButton(
            text = t(Str.OrderQueryTempUnlock, emptyArray()),
            color = blue,
            enabled = enabled,
            onClick = onTempUnlock,
        )
        ActionOutlineButton(
            text = t(Str.OrderQueryEndTrip, emptyArray()),
            color = red,
            enabled = enabled,
            onClick = onEndTrip,
        )
    }
}

@Composable
private fun ActionOutlineButton(
    text: String,
    color: Color,
    enabled: Boolean,
    onClick: () -> Unit,
) {
    Box(
        modifier = Modifier
            .width(80.dp)
            .height(32.dp)
            .border(1.dp, color, RoundedCornerShape(4.dp))
            .clickable(
                enabled = enabled,
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            ),
        contentAlignment = Alignment.Center,
    ) {
        Text(text, color = color, fontSize = 13.sp, fontWeight = FontWeight.Bold)
    }
}

/** 对齐 CarDetailBottomPop：仅结束行程 / 结束行程并修改金额 / 取消。 */
@Composable
private fun EndTripBottomSheet(
    t: Tr,
    onDismiss: () -> Unit,
    onEndOnly: () -> Unit,
    onEndAndModify: () -> Unit,
) {
    val blue = Color(0xFF295FCC)
    Dialog(onDismissRequest = onDismiss) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 8.dp),
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(Color.White, RoundedCornerShape(14.dp)),
            ) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(56.dp)
                        .clickable(
                            interactionSource = remember { MutableInteractionSource() },
                            indication = null,
                            onClick = onEndOnly,
                        ),
                    contentAlignment = Alignment.Center,
                ) {
                    Text(
                        t(Str.OrderQueryEndTripOnly, emptyArray()),
                        color = blue,
                        fontSize = 18.sp,
                        fontWeight = FontWeight.Bold,
                    )
                }
                HorizontalDivider(color = Color(0xFFEFECED), thickness = 1.dp)
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(56.dp)
                        .clickable(
                            interactionSource = remember { MutableInteractionSource() },
                            indication = null,
                            onClick = onEndAndModify,
                        ),
                    contentAlignment = Alignment.Center,
                ) {
                    Text(
                        t(Str.OrderQueryEndTripAndModify, emptyArray()),
                        color = blue,
                        fontSize = 18.sp,
                        fontWeight = FontWeight.Bold,
                    )
                }
            }
            Spacer(Modifier.height(7.dp))
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(56.dp)
                    .background(Color.White, RoundedCornerShape(14.dp))
                    .clickable(
                        interactionSource = remember { MutableInteractionSource() },
                        indication = null,
                        onClick = onDismiss,
                    ),
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    t(Str.Cancel, emptyArray()),
                    color = Color.Black,
                    fontSize = 18.sp,
                    fontWeight = FontWeight.Bold,
                )
            }
            Spacer(Modifier.height(24.dp))
        }
    }
}
@Composable
private fun TempUnlockDialog(
    t: Tr,
    onDismiss: () -> Unit,
    onConfirm: (Int) -> Unit,
) {
    var text by remember { mutableStateOf("") }
    Dialog(onDismissRequest = onDismiss) {
        Column(
            Modifier
                .fillMaxWidth()
                .background(Color.White, RoundedCornerShape(8.dp))
                .padding(16.dp),
        ) {
            Text(
                text = t(Str.OrderQueryTempUnlock, emptyArray()),
                color = TextPrimary,
                fontSize = 16.sp,
                fontWeight = FontWeight.Bold,
            )
            Text(
                text = t(Str.OrderQueryTempUnlockTip, emptyArray()),
                color = Color(0xFF666666),
                fontSize = 13.sp,
                modifier = Modifier.padding(top = 8.dp),
            )
            Row(
                Modifier.fillMaxWidth().padding(top = 12.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                BasicTextField(
                    value = text,
                    onValueChange = { text = it.filter { ch -> ch.isDigit() } },
                    singleLine = true,
                    keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
                    textStyle = TextStyle(color = TextPrimary, fontSize = 14.sp),
                    modifier = Modifier
                        .weight(1f)
                        .height(40.dp)
                        .border(1.dp, Color(0xFFD7D7D7), RoundedCornerShape(4.dp))
                        .padding(horizontal = 10.dp, vertical = 10.dp),
                )
                Spacer(Modifier.width(8.dp))
                Text(
                    text = t(Str.OrderQueryTempUnlockUnit, emptyArray()),
                    color = TextPrimary,
                    fontSize = 14.sp,
                )
            }
            Row(
                Modifier.fillMaxWidth().padding(top = 12.dp),
                horizontalArrangement = Arrangement.End,
            ) {
                Text(
                    text = t(Str.Cancel, emptyArray()),
                    color = Color(0xFF999999),
                    modifier = Modifier.clickable(onClick = onDismiss).padding(8.dp),
                )
                Spacer(Modifier.width(16.dp))
                Text(
                    text = t(Str.Confirm, emptyArray()),
                    color = QueryBlue,
                    modifier = Modifier
                        .clickable {
                            text.toIntOrNull()?.takeIf { it > 0 }?.let(onConfirm)
                        }
                        .padding(8.dp),
                )
            }
        }
    }
}

/** 对齐 ModifyAmountActivity。 */
@Composable
private fun ModifyAmountPage(
    t: Tr,
    feature: OrderQueryFeature,
    order: OrderRecord?,
    fromEndTrip: Boolean,
    loading: Boolean,
    onBack: () -> Unit,
) {
    val scope = rememberCoroutineScope()
    var rideYuan by remember { mutableStateOf("") }
    var dispatchYuan by remember { mutableStateOf("") }
    var helmetYuan by remember { mutableStateOf("") }
    val canSubmit = rideYuan.isNotBlank() && dispatchYuan.isNotBlank() &&
        helmetYuan.isNotBlank() && !loading
    val colors = OpsTheme.colors

    Column(Modifier.fillMaxSize().background(PageWhite)) {
        Box(
            Modifier
                .fillMaxWidth()
                .background(colors.primary)
                .statusBarsPadding()
                .padding(horizontal = 8.dp, vertical = 10.dp),
        ) {
            Text(
                text = t(Str.Cancel, emptyArray()),
                color = colors.onPrimary,
                fontSize = 15.sp,
                modifier = Modifier
                    .align(Alignment.CenterStart)
                    .clickable(onClick = onBack)
                    .padding(8.dp),
            )
            Text(
                text = t(Str.OrderQueryModifyAmountTitle, emptyArray()),
                color = colors.onPrimary,
                fontSize = 18.sp,
                fontWeight = FontWeight.Medium,
                modifier = Modifier.align(Alignment.Center),
            )
            Text(
                text = t(Str.OrderQueryDone, emptyArray()),
                color = if (canSubmit) colors.onPrimary else Color(0x66FFFFFF),
                fontSize = 15.sp,
                modifier = Modifier
                    .align(Alignment.CenterEnd)
                    .clickable(enabled = canSubmit) {
                        scope.launch {
                            val pay = ((rideYuan.toDoubleOrNull() ?: 0.0) * 100).toInt()
                            val dispatch = ((dispatchYuan.toDoubleOrNull() ?: 0.0) * 100).toInt()
                            val helmet = ((helmetYuan.toDoubleOrNull() ?: 0.0) * 100).toInt()
                            if (fromEndTrip) {
                                feature.endTripWithCost(pay, dispatch, helmet)
                            } else {
                                feature.modifyCost(pay, dispatch, helmet)
                            }
                        }
                    }
                    .padding(8.dp),
            )
        }
        if (order != null) {
            ModifyAmountInfoRow(
                label = t(Str.OrderQueryAmount, emptyArray()),
                value = OrderFormat.yuanWithUnit(order.payCost),
            )
            ModifyAmountInfoRow(
                label = t(Str.OrderQueryRideCostLabel, emptyArray()),
                value = OrderFormat.yuanWithUnit(order.originCost),
            )
            ModifyAmountInfoRow(
                label = t(Str.OrderQueryDispatchCostLabel, emptyArray()),
                value = OrderFormat.yuanWithUnit(order.dispatchCost),
            )
            ModifyAmountInfoRow(
                label = t(Str.OrderQueryHelmetCostLabel, emptyArray()),
                value = OrderFormat.yuanWithUnit(order.helmetPenalty),
            )
            Box(Modifier.fillMaxWidth().height(8.dp).background(SectionGap))
        }
        ModifyAmountInputRow(
            label = t(Str.OrderQueryRideCostLabel, emptyArray()),
            value = rideYuan,
            onValueChange = { rideYuan = it },
            unit = t(Str.OrderQueryCny, emptyArray()),
            hint = t(Str.OrderQueryPleaseInput, emptyArray()),
        )
        ModifyAmountInputRow(
            label = t(Str.OrderQueryDispatchCostLabel, emptyArray()),
            value = dispatchYuan,
            onValueChange = { dispatchYuan = it },
            unit = t(Str.OrderQueryCny, emptyArray()),
            hint = t(Str.OrderQueryPleaseInput, emptyArray()),
        )
        ModifyAmountInputRow(
            label = t(Str.OrderQueryHelmetCostLabel, emptyArray()),
            value = helmetYuan,
            onValueChange = { helmetYuan = it },
            unit = t(Str.OrderQueryCny, emptyArray()),
            hint = t(Str.OrderQueryPleaseInput, emptyArray()),
        )
        if (loading) {
            Box(Modifier.fillMaxWidth().padding(24.dp), contentAlignment = Alignment.Center) {
                CircularProgressIndicator(color = colors.primary)
            }
        }
    }
}

@Composable
private fun ModifyAmountInfoRow(label: String, value: String) {
    Row(
        Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 10.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
    ) {
        Text(text = label, color = Color(0xFF333333), fontSize = 14.sp)
        Text(text = value, color = Color(0xFF333333), fontSize = 14.sp)
    }
}

@Composable
private fun ModifyAmountInputRow(
    label: String,
    value: String,
    onValueChange: (String) -> Unit,
    unit: String,
    hint: String,
) {
    Row(
        Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 16.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.SpaceBetween,
    ) {
        Text(
            text = label,
            color = Color(0xFF333333),
            fontSize = 16.sp,
            modifier = Modifier.weight(1f),
        )
        BasicTextField(
            value = value,
            onValueChange = { raw ->
                onValueChange(raw.filter { it.isDigit() || it == '.' })
            },
            singleLine = true,
            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Decimal),
            textStyle = TextStyle(color = Color(0xFF333333), fontSize = 12.sp),
            modifier = Modifier
                .width(100.dp)
                .height(34.dp)
                .border(1.dp, Color(0xFFD7D7D7), RoundedCornerShape(5.dp))
                .padding(horizontal = 10.dp, vertical = 8.dp),
            decorationBox = { inner ->
                if (value.isEmpty()) {
                    Text(text = hint, color = Color(0xFF999999), fontSize = 12.sp)
                }
                inner()
            },
        )
        Spacer(Modifier.width(8.dp))
        Text(text = unit, color = Color.Black, fontSize = 16.sp)
    }
}


@Composable
private fun TopBar(
    title: String,
    onBack: () -> Unit,
    colors: OpsColors,
) {
    com.luopingtech.ebike.ops.ui.navigation.OpsBackHandler(onBack = onBack)
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .background(colors.primary)
            .statusBarsPadding()
            .padding(horizontal = 8.dp, vertical = 10.dp),
    ) {
        OpsBackChevron(
            onClick = onBack,
            modifier = Modifier.align(Alignment.CenterStart),
        )
        Text(
            text = title,
            color = colors.onPrimary,
            fontSize = 18.sp,
            fontWeight = FontWeight.Medium,
            modifier = Modifier.align(Alignment.Center),
        )
    }
}
