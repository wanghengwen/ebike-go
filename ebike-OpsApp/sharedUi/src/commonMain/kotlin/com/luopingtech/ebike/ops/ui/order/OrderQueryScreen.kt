package com.luopingtech.ebike.ops.ui.order

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.order.OrderFormat
import com.luopingtech.ebike.ops.domain.order.OrderRecord
import com.luopingtech.ebike.ops.domain.order.OrderUserDetail
import com.luopingtech.ebike.ops.domain.order.OrderUserPageItem
import com.luopingtech.ebike.ops.feature.order.OrderQueryNav
import com.luopingtech.ebike.ops.ui.theme.OpsColors
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.launch

/**
 * 订单查询（CMP）。可跳原生车辆详情；不再走 H5。
 */
@Composable
fun OrderQueryScreen(
    app: OpsApp,
    currentArea: ServiceArea?,
    onClose: () -> Unit,
    onOpenVehicleDetail: (String) -> Unit = {},
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.orderQueryFeature.state.collectAsState()
    val colors = OpsTheme.colors
    val scope = rememberCoroutineScope()
    val feature = app.orderQueryFeature

    DisposableEffect(Unit) {
        onDispose { feature.clear() }
    }

    LaunchedEffect(currentArea?.id) {
        feature.open(currentArea)
    }

    Box(modifier = Modifier.fillMaxSize().background(colors.pageBackground)) {
        when (val nav = state.nav) {
            OrderQueryNav.Search -> SearchPage(
                app = app,
                t = { key, args -> t(key, *args) },
                onClose = onClose,
                onSearch = { scope.launch { feature.search() } },
                onLoadMore = { scope.launch { feature.loadMoreHome() } },
                onOpenOrderVehicle = { carId ->
                    scope.launch { feature.openVehicleFromHome(carId) }
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
                t = { key, args -> t(key, *args) },
                title = state.scopeUser?.authName?.ifBlank { nav.pin } ?: nav.pin,
                user = state.scopeUser,
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
            )
            is OrderQueryNav.VehicleOrders -> ScopePage(
                t = { key, args -> t(key, *args) },
                title = nav.carId?.ifBlank { null }
                    ?: nav.imei?.ifBlank { null }
                    ?: t(Str.OrderQueryScreen),
                user = null,
                carId = nav.carId,
                imei = nav.imei,
                lastOrder = state.lastOrder,
                history = state.history,
                loading = state.scopeLoading || state.historyLoading,
                finished = state.historyFinished,
                error = state.historyError,
                onBack = { feature.navigateBack() },
                onLoadMore = { scope.launch { feature.loadMoreHistory() } },
                onOpenVehicle = onOpenVehicleDetail,
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

@Composable
private fun SearchPage(
    app: OpsApp,
    t: Tr,
    onClose: () -> Unit,
    onSearch: () -> Unit,
    onLoadMore: () -> Unit,
    onOpenOrderVehicle: (String) -> Unit,
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

    LaunchedEffect(listState, state.loading, state.finished) {
        snapshotFlow {
            val info = listState.layoutInfo
            val last = info.visibleItemsInfo.lastOrNull()?.index ?: 0
            last >= info.totalItemsCount - 2 && info.totalItemsCount > 0
        }.distinctUntilChanged().collect { nearEnd ->
            if (nearEnd && !state.loading && !state.finished) onLoadMore()
        }
    }

    Column(modifier = Modifier.fillMaxSize()) {
        TopBar(title = t(Str.OrderQueryScreen, emptyArray()), onBack = onClose, colors = colors)
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color.White)
                .padding(12.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            BasicTextField(
                value = state.keyword,
                onValueChange = { app.orderQueryFeature.setKeyword(it) },
                singleLine = true,
                enabled = placeholder.isNotEmpty(),
                textStyle = TextStyle(fontSize = 15.sp, color = colors.textPrimary),
                cursorBrush = SolidColor(colors.primary),
                keyboardOptions = KeyboardOptions(imeAction = ImeAction.Search),
                keyboardActions = KeyboardActions(onSearch = { onSearch() }),
                modifier = Modifier
                    .weight(1f)
                    .background(Color(0xFFF5F6F8), RoundedCornerShape(20.dp))
                    .padding(horizontal = 14.dp, vertical = 10.dp),
                decorationBox = { inner ->
                    if (state.keyword.isEmpty()) {
                        Text(
                            placeholder.ifEmpty { t(Str.OrderQueryNoPerm, emptyArray()) },
                            color = colors.textTertiary,
                            fontSize = 14.sp,
                        )
                    }
                    inner()
                },
            )
            Box(
                modifier = Modifier
                    .background(colors.primary, RoundedCornerShape(6.dp))
                    .clickable(enabled = placeholder.isNotEmpty() && !state.searching, onClick = onSearch)
                    .padding(horizontal = 14.dp, vertical = 10.dp),
            ) {
                Text(
                    if (state.searching) {
                        t(Str.LoadingEllipsis, emptyArray())
                    } else {
                        t(Str.OrderQuerySearchAction, emptyArray())
                    },
                    color = colors.onPrimary,
                    fontSize = 14.sp,
                )
            }
        }

        if (placeholder.isEmpty()) {
            Text(
                t(Str.OrderQueryNoPerm, emptyArray()),
                modifier = Modifier.padding(40.dp).fillMaxWidth(),
                color = colors.textTertiary,
                fontSize = 14.sp,
            )
        } else {
            Text(
                text = buildString {
                    append(t(Str.OrderQueryRecentHint, emptyArray()))
                    if (state.areaName.isNotBlank()) append(" · ${state.areaName}")
                },
                modifier = Modifier.padding(horizontal = 14.dp, vertical = 8.dp),
                color = colors.textTertiary,
                fontSize = 12.sp,
            )

            LazyColumn(
                state = listState,
                modifier = Modifier
                    .weight(1f)
                    .fillMaxWidth(),
            ) {
                items(state.items, key = { it.id.ifBlank { "${it.carId}-${it.startTime}" } }) { order ->
                    OrderCard(
                        order = order,
                        t = t,
                        clickable = state.canQueryVehicle && order.carId.isNotBlank(),
                        onClick = { onOpenOrderVehicle(order.carId) },
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
                            color = colors.textTertiary,
                            fontSize = 14.sp,
                        )
                        state.finished -> Text(
                            t(Str.VehicleListEnd, emptyArray()),
                            modifier = Modifier.padding(16.dp).fillMaxWidth(),
                            color = colors.textTertiary,
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

    Column(modifier = Modifier.fillMaxSize()) {
        TopBar(title = t(Str.OrderQueryPickUser, emptyArray()), onBack = onBack, colors = colors)
        Text(
            "「$name」 · ${users.size}",
            modifier = Modifier.padding(horizontal = 14.dp, vertical = 10.dp),
            color = colors.textTertiary,
            fontSize = 12.sp,
        )
        LazyColumn(state = listState, modifier = Modifier.weight(1f)) {
            items(users, key = { it.pin }) { user ->
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 12.dp, vertical = 6.dp)
                        .background(Color.White, RoundedCornerShape(8.dp))
                        .clickable { onPick(user.pin) }
                        .padding(12.dp),
                ) {
                    Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                        Text(OrderFormat.orDash(user.authName), fontWeight = FontWeight.SemiBold, fontSize = 16.sp)
                        Text(OrderFormat.ridingStateLabel(user.ridingState), color = colors.textTertiary, fontSize = 12.sp)
                    }
                    Spacer(Modifier.height(6.dp))
                    InfoRow(t(Str.Phone, emptyArray()), OrderFormat.orDash(user.phone))
                    InfoRow(t(Str.OrderQueryAmount, emptyArray()), OrderFormat.yuanWithUnit(user.balance))
                    InfoRow(t(Str.OrderQueryStart, emptyArray()), OrderFormat.isoDateTime(user.createdAt))
                }
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
                        color = colors.textTertiary,
                    )
                }
            }
        }
    }
}

@Composable
private fun ScopePage(
    t: Tr,
    title: String,
    user: OrderUserDetail?,
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

    Column(modifier = Modifier.fillMaxSize()) {
        TopBar(title = title, onBack = onBack, colors = colors)
        LazyColumn(state = listState, modifier = Modifier.weight(1f)) {
            if (user != null) {
                item {
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(12.dp)
                            .background(Color.White, RoundedCornerShape(8.dp))
                            .padding(12.dp),
                    ) {
                        Text(OrderFormat.orDash(user.authName), fontWeight = FontWeight.SemiBold, fontSize = 17.sp)
                        Spacer(Modifier.height(8.dp))
                        InfoRow(t(Str.Phone, emptyArray()), OrderFormat.orDash(user.phone))
                        InfoRow(t(Str.OrderQueryUser, emptyArray()), OrderFormat.ridingStateLabel(user.ridingState))
                        InfoRow(t(Str.OrderQueryAmount, emptyArray()), OrderFormat.yuanWithUnit(user.balance))
                        if (user.serviceName.isNotBlank()) {
                            InfoRow(t(Str.SelectServiceArea, emptyArray()), user.serviceName)
                        }
                    }
                }
            } else {
                item {
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(12.dp)
                            .background(Color.White, RoundedCornerShape(8.dp))
                            .padding(12.dp),
                    ) {
                        if (!carId.isNullOrBlank()) {
                            InfoRow(t(Str.OrderQueryCarLabel, emptyArray()), carId)
                        }
                        if (!imei.isNullOrBlank()) {
                            InfoRow(t(Str.OrderQueryImeiLabel, emptyArray()), imei)
                        }
                        val openId = carId?.takeIf { it.isNotBlank() }
                            ?: lastOrder?.carId?.takeIf { it.isNotBlank() }
                        if (openId != null) {
                            Spacer(Modifier.height(8.dp))
                            Text(
                                t(Str.OrderQueryOpenVehicle, emptyArray()),
                                color = colors.primary,
                                fontSize = 14.sp,
                                modifier = Modifier.clickable { onOpenVehicle(openId) },
                            )
                        }
                    }
                }
            }

            if (lastOrder != null) {
                item {
                    Text(
                        t(Str.OrderQueryLastOrder, emptyArray()),
                        modifier = Modifier.padding(horizontal = 14.dp, vertical = 6.dp),
                        color = colors.textSecondary,
                        fontSize = 13.sp,
                    )
                }
                item {
                    OrderCard(
                        order = lastOrder,
                        t = t,
                        clickable = lastOrder.carId.isNotBlank(),
                        onClick = { onOpenVehicle(lastOrder.carId) },
                    )
                }
            }

            item {
                Text(
                    t(Str.OrderQueryHistory, emptyArray()),
                    modifier = Modifier.padding(horizontal = 14.dp, vertical = 8.dp),
                    color = colors.textSecondary,
                    fontSize = 13.sp,
                )
            }
            items(history, key = { it.id.ifBlank { "${it.carId}-${it.startTime}-h" } }) { order ->
                OrderCard(
                    order = order,
                    t = t,
                    clickable = order.carId.isNotBlank(),
                    onClick = { onOpenVehicle(order.carId) },
                )
            }
            item {
                when {
                    loading -> Box(Modifier.fillMaxWidth().padding(16.dp), contentAlignment = Alignment.Center) {
                        CircularProgressIndicator(color = colors.primary)
                    }
                    error != null -> Text(error, modifier = Modifier.padding(16.dp), color = Color(0xFFE53935))
                    finished && history.isEmpty() && lastOrder == null -> Text(
                        t(Str.AdminEmptyList, emptyArray()),
                        modifier = Modifier.padding(40.dp),
                        color = colors.textTertiary,
                    )
                }
            }
        }
    }
}

@Composable
private fun OrderCard(
    order: OrderRecord,
    t: Tr,
    clickable: Boolean,
    onClick: () -> Unit,
) {
    val settled = OrderFormat.isSettled(order.izPaid)
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 12.dp, vertical = 6.dp)
            .background(Color.White, RoundedCornerShape(8.dp))
            .then(if (clickable) Modifier.clickable(onClick = onClick) else Modifier)
            .padding(12.dp),
    ) {
        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
            Text(OrderFormat.orDash(order.carId), fontWeight = FontWeight.SemiBold, fontSize = 16.sp)
            Text(
                OrderFormat.payStateLabel(order.izPaid),
                color = if (settled) Color(0xFF999999) else Color(0xFFEE0A24),
                fontWeight = if (settled) FontWeight.Normal else FontWeight.SemiBold,
                fontSize = 12.sp,
            )
        }
        Spacer(Modifier.height(8.dp))
        Row(modifier = Modifier.fillMaxWidth()) {
            Column(modifier = Modifier.weight(1f)) {
                InfoRow(t(Str.OrderQueryAmount, emptyArray()), OrderFormat.yuanWithUnit(order.payCost))
                InfoRow(t(Str.OrderQueryDistance, emptyArray()), OrderFormat.distance(order.mile))
            }
            Spacer(Modifier.width(12.dp))
            Column(modifier = Modifier.weight(1f)) {
                InfoRow(t(Str.OrderQueryDuration, emptyArray()), OrderFormat.duration(order.ridingTimeRaw))
                InfoRow(t(Str.OrderQueryUser, emptyArray()), OrderFormat.orDash(order.phone))
            }
        }
        HorizontalDivider(modifier = Modifier.padding(vertical = 8.dp), color = Color(0xFFF2F2F2))
        Text(
            "${t(Str.OrderQueryStart, emptyArray())} ${OrderFormat.orDash(order.startTime)}",
            color = Color(0xFF646464),
            fontSize = 12.sp,
        )
        Text(
            "${t(Str.OrderQueryEnd, emptyArray())} ${OrderFormat.orDash(order.endTime)}",
            color = Color(0xFF646464),
            fontSize = 12.sp,
        )
        Text(
            "${t(Str.OrderQueryId, emptyArray())} ${OrderFormat.orDash(order.id)}",
            color = Color(0xFF999999),
            fontSize = 12.sp,
            modifier = Modifier.padding(top = 4.dp),
        )
    }
}

@Composable
private fun InfoRow(label: String, value: String) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 2.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
    ) {
        Text(label, color = Color(0xFF999999), fontSize = 13.sp)
        Text(value, color = Color(0xFF282828), fontSize = 13.sp, fontWeight = FontWeight.Medium)
    }
}

@Composable
private fun TopBar(
    title: String,
    onBack: () -> Unit,
    colors: OpsColors,
) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .background(colors.primary)
            .statusBarsPadding()
            .padding(horizontal = 8.dp, vertical = 10.dp),
    ) {
        Text(
            text = "‹",
            color = colors.onPrimary,
            fontSize = 28.sp,
            modifier = Modifier
                .align(Alignment.CenterStart)
                .clickable(onClick = onBack)
                .padding(4.dp),
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
