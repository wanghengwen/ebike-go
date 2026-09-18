package com.luopingtech.ebike.ops.ui.vehicle

import androidx.activity.compose.BackHandler
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableFloatStateOf
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateMapOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.model.TrackPoint
import com.luopingtech.ebike.ops.domain.order.OrderRecord
import com.luopingtech.ebike.ops.domain.order.buildOrderHistoryPins
import com.luopingtech.ebike.ops.ui.map.TencentMapView
import com.luopingtech.ebike.ops.ui.order.OrderHistoryListItem
import com.luopingtech.ebike.ops.ui.order.OrderTrackPlaybackBar
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.launch

/**
 * 对齐遗留 OrderHistoryActivity：顶栏「订单记录」+ 地图起终点/轨迹 +「车辆订单记录」列表。
 */
@Composable
fun VehicleOrderHistoryScreen(
    app: OpsApp,
    carId: String,
    currentArea: ServiceArea?,
    mapReady: Boolean,
    onClose: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.orderQueryFeature.state.collectAsState()
    val feature = app.orderQueryFeature
    val scope = rememberCoroutineScope()
    val colors = OpsTheme.colors

    var selectedIndex by remember { mutableIntStateOf(0) }
    var fitNonce by remember { mutableIntStateOf(0) }
    var showNote by remember { mutableStateOf(false) }
    var trackLoading by remember { mutableStateOf(false) }
    val addressCache = remember { mutableStateMapOf<String, String>() }
    /** orderId → 已补全轨迹/坐标的订单（对齐遗留点选后 queryOrderDetail）。 */
    val orderEnrichment = remember { mutableStateMapOf<String, OrderRecord>() }

    val enrichmentSnapshot = orderEnrichment.toMap()
    val orders = buildList {
        state.lastOrder?.let { add(it) }
        addAll(state.history.filter { it.id != state.lastOrder?.id || it.id.isBlank() })
    }.distinctBy { it.id.ifBlank { "${it.carId}-${it.startTime}" } }
        .map { base -> enrichmentSnapshot[orderKey(base)] ?: base }

    DisposableEffect(Unit) {
        onDispose { feature.clear() }
    }

    LaunchedEffect(carId, currentArea?.id) {
        orderEnrichment.clear()
        feature.bindArea(currentArea)
        feature.openVehicleFromHome(carId)
        selectedIndex = 0
        fitNonce++
    }

    val selected = orders.getOrNull(selectedIndex)

    // 点选订单：补轨迹（列表常无 deviceTrajectory），再 fit 地图。
    LaunchedEffect(selectedIndex, selected?.id, selected?.startTime) {
        val order = selected ?: return@LaunchedEffect
        resolveAddress(app, order.startLat, order.startLng, addressCache)
        resolveAddress(app, order.endLat, order.endLng, addressCache)

        val key = orderKey(order)
        if (order.trajectory.size >= 2) {
            fitNonce++
            return@LaunchedEffect
        }
        val id = order.id.trim()
        if (id.isEmpty() || id == "0") {
            fitNonce++
            return@LaunchedEffect
        }
        trackLoading = true
        when (val result = app.orderRepository.orderDetail(id)) {
            is OpsResult.Ok -> {
                val detail = result.value
                val merged = order.copy(
                    startLat = detail.startLat ?: order.startLat,
                    startLng = detail.startLng ?: order.startLng,
                    endLat = detail.endLat ?: order.endLat,
                    endLng = detail.endLng ?: order.endLng,
                    trajectory = detail.trajectory.ifEmpty { order.trajectory },
                )
                orderEnrichment[key] = merged
                resolveAddress(app, merged.startLat, merged.startLng, addressCache)
                resolveAddress(app, merged.endLat, merged.endLng, addressCache)
            }
            is OpsResult.Err -> Unit
        }
        trackLoading = false
        fitNonce++
    }

    val listState = rememberLazyListState()
    LaunchedEffect(listState, state.historyLoading, state.historyFinished) {
        snapshotFlow {
            val info = listState.layoutInfo
            val last = info.visibleItemsInfo.lastOrNull()?.index ?: 0
            last >= info.totalItemsCount - 2 && info.totalItemsCount > 0
        }.distinctUntilChanged().collect { nearEnd ->
            if (nearEnd && !state.historyLoading && !state.historyFinished) {
                scope.launch { feature.loadMoreHistory() }
            }
        }
    }

    val trackPoints: List<TrackPoint> = selected?.trajectory.orEmpty()
    var playbackProgress by remember(selected?.id, selected?.startTime) { mutableFloatStateOf(0f) }
    val pins = remember(selected, trackPoints, playbackProgress) {
        buildOrderHistoryPins(selected, trackPoints, playbackProgress)
    }

    BackHandler(onBack = {
        if (showNote) showNote = false else onClose()
    })

    Box(modifier = modifier.fillMaxSize().background(Color.White)) {
        Column(modifier = Modifier.fillMaxSize()) {
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(colors.primary)
                    .statusBarsPadding()
                    .height(48.dp),
            ) {
                Text(
                    text = "\u2039",
                    color = Color.White,
                    fontSize = 28.sp,
                    modifier = Modifier
                        .align(Alignment.CenterStart)
                        .clickable(
                            interactionSource = remember { MutableInteractionSource() },
                            indication = null,
                            onClick = onClose,
                        )
                        .padding(horizontal = 16.dp),
                )
                Text(
                    text = t(Str.OrderRecordTitle),
                    color = Color.White,
                    fontSize = 18.sp,
                    fontWeight = FontWeight.Medium,
                    modifier = Modifier.align(Alignment.Center),
                )
            }

            if (mapReady) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(308.dp),
                ) {
                    TencentMapView(
                        pins = pins,
                        selectedCarId = null,
                        onSelectCarId = {},
                        clusterOverview = false,
                        trackPoints = trackPoints,
                        fitNonce = fitNonce,
                        showStatusOverlay = false,
                        modifier = Modifier.fillMaxSize(),
                    )
                    if (trackLoading) {
                        CircularProgressIndicator(
                            color = colors.primary,
                            modifier = Modifier
                                .align(Alignment.TopEnd)
                                .padding(12.dp),
                        )
                    }
                    Column(
                        modifier = Modifier
                            .align(Alignment.CenterStart)
                            .padding(start = 16.dp)
                            .background(Color.White, RoundedCornerShape(8.dp))
                            .width(40.dp),
                    ) {
                        MapSideTool(
                            label = t(Str.MapToolLocate),
                            onClick = { fitNonce++ },
                        )
                        HorizontalDivider(color = Color(0xFFE5E5E5), thickness = 1.dp)
                        MapSideTool(
                            label = t(Str.MapToolLegend),
                            onClick = { showNote = true },
                        )
                    }
                }
            }

            if (trackPoints.size >= 2) {
                OrderTrackPlaybackBar(
                    points = trackPoints,
                    progress = playbackProgress,
                    onProgress = { playbackProgress = it },
                )
            }

            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(44.dp)
                    .background(Color.White)
                    .padding(horizontal = 16.dp),
                contentAlignment = Alignment.CenterStart,
            ) {
                Text(
                    text = t(Str.VehicleOrderHistory),
                    color = Color(0xFF242936),
                    fontSize = 14.sp,
                    fontWeight = FontWeight.Bold,
                )
            }
            HorizontalDivider(color = Color(0xFFD9DCE6), thickness = 1.dp)

            when {
                (state.scopeLoading || state.historyLoading) && orders.isEmpty() -> {
                    Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                        CircularProgressIndicator(color = colors.primary)
                    }
                }
                state.historyError != null && orders.isEmpty() -> {
                    Text(
                        state.historyError.orEmpty(),
                        modifier = Modifier.padding(16.dp),
                        color = Color(0xFFE02020),
                    )
                }
                orders.isEmpty() -> {
                    Text(
                        t(Str.AdminEmptyList),
                        modifier = Modifier.padding(40.dp).fillMaxWidth(),
                        color = Color(0xFF7C87B1),
                        textAlign = TextAlign.Center,
                    )
                }
                else -> LazyColumn(state = listState, modifier = Modifier.fillMaxSize()) {
                    itemsIndexed(
                        orders,
                        key = { _, it -> orderKey(it) },
                    ) { index, order ->
                        OrderHistoryListItem(
                            order = order,
                            t = { key, args -> t(key, *args) },
                            selected = index == selectedIndex,
                            startAddress = addressOf(order.startLat, order.startLng, addressCache),
                            endAddress = addressOf(order.endLat, order.endLng, addressCache),
                            onClick = {
                                if (selectedIndex != index) {
                                    selectedIndex = index
                                } else {
                                    fitNonce++
                                }
                            },
                        )
                    }
                    item {
                        when {
                            state.historyLoading -> Box(
                                Modifier.fillMaxWidth().padding(16.dp),
                                contentAlignment = Alignment.Center,
                            ) { CircularProgressIndicator(color = colors.primary) }
                            state.historyFinished -> Text(
                                t(Str.VehicleListEnd),
                                modifier = Modifier.padding(16.dp).fillMaxWidth(),
                                color = Color(0xFF7C87B1),
                                fontSize = 12.sp,
                                textAlign = TextAlign.Center,
                            )
                        }
                    }
                }
            }
        }

        if (showNote) {
            VehicleDetailNoteScreen(
                app = app,
                onClose = { showNote = false },
            )
        }
    }
}

private fun orderKey(order: OrderRecord): String =
    order.id.ifBlank { "${order.carId}-${order.startTime}" }

@Composable
private fun MapSideTool(label: String, onClick: () -> Unit) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .height(58.dp)
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            ),
        contentAlignment = Alignment.Center,
    ) {
        Text(
            text = label,
            color = Color(0xFF242936),
            fontSize = 10.sp,
            textAlign = TextAlign.Center,
        )
    }
}

private fun addressKey(lat: Double?, lng: Double?): String? {
    if (lat == null || lng == null || lat == 0.0 || lng == 0.0) return null
    return "${"%.5f".format(lat)},${"%.5f".format(lng)}"
}

private fun addressOf(
    lat: Double?,
    lng: Double?,
    cache: Map<String, String>,
): String {
    val key = addressKey(lat, lng) ?: return "--"
    return cache[key] ?: "--"
}

private suspend fun resolveAddress(
    app: OpsApp,
    lat: Double?,
    lng: Double?,
    cache: MutableMap<String, String>,
) {
    val key = addressKey(lat, lng) ?: return
    if (cache.containsKey(key)) return
    val result = app.reverseGeocoder.addressOf(lat!!, lng!!)
    val text = result.getOrNull()?.takeIf { it.isNotBlank() && !it.equals("Ocean", ignoreCase = true) }
    cache[key] = text ?: "--"
}
