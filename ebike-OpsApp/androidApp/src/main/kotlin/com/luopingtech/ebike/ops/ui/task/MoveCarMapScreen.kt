package com.luopingtech.ebike.ops.ui.task

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Slider
import androidx.compose.material3.SliderDefaults
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableFloatStateOf
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.zIndex
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.core.time.nowEpochMillis
import com.luopingtech.ebike.ops.domain.model.MapPin
import com.luopingtech.ebike.ops.domain.vehicle.VehicleOperationStates
import com.luopingtech.ebike.ops.domain.vehicle.VehicleRidingStates
import com.luopingtech.ebike.ops.ui.home.HomeFilterHandle
import com.luopingtech.ebike.ops.ui.map.SimulatorMapView
import com.luopingtech.ebike.ops.ui.map.TencentMapView
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import kotlinx.coroutines.launch

/**
 * 挪车地图 —— 对齐遗留 TaskMapActivity + HomeMoveCarFragment。
 */
@Composable
fun MoveCarMapScreen(
    app: OpsApp,
    onClose: () -> Unit,
    onOpenFreeMove: () -> Unit,
    onOpenStats: () -> Unit,
    onChangeArea: () -> Unit,
) {
    @Suppress("UNUSED_VARIABLE")
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val home by app.homeFeature.state.collectAsState()
    val detailMap by app.vehicleDetailMapFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    val colors = OpsTheme.colors

    var idleHours by remember { mutableFloatStateOf(0f) }
    var fitNonce by remember { mutableIntStateOf(0) }
    var zoomInNonce by remember { mutableIntStateOf(0) }
    var zoomOutNonce by remember { mutableIntStateOf(0) }
    var followNonce by remember { mutableIntStateOf(0) }
    var followLat by remember { mutableStateOf<Double?>(null) }
    var followLng by remember { mutableStateOf<Double?>(null) }
    var selectedCarId by remember { mutableStateOf<String?>(null) }
    var showFence by remember { mutableStateOf(true) }
    var mapTypeSatellite by remember { mutableStateOf(false) }
    var clusterOverview by remember { mutableStateOf(true) }
    var filterOpen by remember { mutableStateOf(false) }
    var hideRiding by remember { mutableStateOf(false) }
    var hideMoving by remember { mutableStateOf(true) }
    val now = remember { nowEpochMillis() }

    LaunchedEffect(home.currentArea?.id) {
        app.homeFeature.reloadVehicles()
        home.currentArea?.id?.takeIf { it.isNotBlank() }?.let {
            app.vehicleDetailMapFeature.loadFence(it)
        }
    }

    fun idleH(lockMs: Long): Float {
        if (lockMs <= 0L) return 0f
        return ((now - lockMs).coerceAtLeast(0L) / 3_600_000f)
    }

    val filtered = remember(home.vehicles, idleHours, now, hideRiding, hideMoving) {
        home.vehicles.filter { v ->
            val hasCoord = v.lat != 0.0 || v.lng != 0.0
            val idleOk = idleH(v.lockTimeMs) >= idleHours
            val ridingOk = !hideRiding || v.ridingState != VehicleRidingStates.RIDING
            val movingOk = !hideMoving || !v.operationStates.contains(VehicleOperationStates.MOVING_CAR)
            hasCoord && idleOk && ridingOk && movingOk
        }
    }
    val ridingCount = filtered.count {
        it.ridingState == VehicleRidingStates.RIDING ||
            it.ridingState == VehicleRidingStates.TEMP_PARKING
    }
    val pins = remember(filtered) {
        filtered.map {
            MapPin(
                id = it.carId,
                lat = it.lat,
                lng = it.lng,
                title = it.carId,
                subtitle = t(Str.HoursUnit, idleH(it.lockTimeMs).toInt()),
                restBattery = it.restBattery,
                ridingState = it.ridingState,
                memberIds = listOf(it.carId),
            )
        }
    }
    val isTencent = home.mapProviderKind.equals("tencent", ignoreCase = true) && home.mapReady
    val fencePolygons = if (showFence) detailMap.fence?.all.orEmpty() else emptyList()

    Column(modifier = Modifier.fillMaxSize().background(Color.White)) {
        TaskFullscreenTopBar(
            title = t(Str.MoveCarMapTitle),
            primary = colors.primary,
            onBack = onClose,
            trailingLabel = t(Str.MoveCarStats),
            onTrailing = onOpenStats,
        )
        Box(modifier = Modifier.weight(1f).fillMaxWidth()) {
            if (isTencent) {
                TencentMapView(
                    pins = pins,
                    selectedCarId = selectedCarId,
                    onSelectCarId = { selectedCarId = it },
                    clusterOverview = clusterOverview,
                    fencePolygons = fencePolygons,
                    fitNonce = fitNonce,
                    zoomInNonce = zoomInNonce,
                    zoomOutNonce = zoomOutNonce,
                    followNonce = followNonce,
                    followLat = followLat,
                    followLng = followLng,
                    mapTypeSatellite = mapTypeSatellite,
                    showStatusOverlay = false,
                    modifier = Modifier.fillMaxSize(),
                )
            } else {
                SimulatorMapView(
                    pins = pins,
                    selectedCarId = selectedCarId,
                    providerLabel = t(Str.MoveCarMapTitle),
                    onSelectCarId = { selectedCarId = it },
                    clusterOverview = clusterOverview,
                    modifier = Modifier.fillMaxSize(),
                )
            }
            TaskMapSideTools(
                refreshLabel = t(Str.Refresh),
                locateLabel = t(Str.MapToolLocate),
                moreLabel = t(Str.MapToolMore),
                parkingLabel = t(Str.MapToolParking),
                satelliteLabel = t(Str.MapToolSatellite),
                detailLabel = t(Str.Detail),
                onZoomIn = { zoomInNonce += 1 },
                onZoomOut = { zoomOutNonce += 1 },
                onRefresh = {
                    scope.launch {
                        app.homeFeature.reloadVehicles()
                        home.currentArea?.id?.let { app.vehicleDetailMapFeature.loadFence(it) }
                    }
                },
                onLocate = {
                    scope.launch {
                        when (val loc = app.locationTracker.currentLocation()) {
                            is OpsResult.Ok -> {
                                followLat = loc.value.latitude
                                followLng = loc.value.longitude
                                followNonce += 1
                            }
                            is OpsResult.Err -> fitNonce += 1
                        }
                    }
                },
                parkingSelected = showFence,
                satelliteSelected = mapTypeSatellite,
                detailSelected = !clusterOverview,
                onToggleParking = { showFence = !showFence },
                onToggleSatellite = { mapTypeSatellite = !mapTypeSatellite },
                onToggleDetail = { clusterOverview = !clusterOverview },
                modifier = Modifier
                    .align(Alignment.CenterStart)
                    .padding(start = 12.dp)
                    .zIndex(2f),
            )
            HomeFilterHandle(
                label = t(Str.FilterHandle),
                open = filterOpen,
                onClick = { filterOpen = !filterOpen },
                modifier = Modifier
                    .align(Alignment.CenterEnd)
                    .zIndex(2f),
            )
            TaskMapSwitchAreaButton(
                label = t(Str.SwitchAreaShort),
                onClick = onChangeArea,
                modifier = Modifier
                    .align(Alignment.BottomEnd)
                    .padding(end = 16.dp, bottom = 16.dp)
                    .zIndex(2f),
            )
            if (filterOpen) {
                Surface(
                    modifier = Modifier
                        .align(Alignment.CenterEnd)
                        .fillMaxWidth(0.72f)
                        .fillMaxHeight()
                        .zIndex(3f),
                    color = Color.White,
                    shadowElevation = 8.dp,
                ) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.SpaceBetween,
                            verticalAlignment = Alignment.CenterVertically,
                        ) {
                            Text(t(Str.FilterHandle), fontWeight = FontWeight.SemiBold, fontSize = 16.sp)
                            TextButton(onClick = { filterOpen = false }) { Text(t(Str.Close)) }
                        }
                        Spacer(modifier = Modifier.height(8.dp))
                        TaskMapFilterToggle(
                            label = t(Str.FilterHideRiding),
                            checked = hideRiding,
                            onToggle = { hideRiding = !hideRiding },
                        )
                        TaskMapFilterToggle(
                            label = t(Str.FilterHideMoving),
                            checked = hideMoving,
                            onToggle = { hideMoving = !hideMoving },
                        )
                    }
                }
            }
        }
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color.White, RoundedCornerShape(topStart = 16.dp, topEnd = 16.dp))
                .padding(horizontal = 16.dp, vertical = 12.dp),
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Column(modifier = Modifier.weight(1f)) {
                    Text(
                        "${filtered.size}",
                        fontSize = 22.sp,
                        fontWeight = FontWeight.Bold,
                        color = Color(0xFF333333),
                    )
                    Text(t(Str.FilterResultVehicles), fontSize = 11.sp, color = Color(0xFF888888))
                }
                Text(
                    text = t(Str.MoveCarShort),
                    color = Color.White,
                    fontWeight = FontWeight.SemiBold,
                    fontSize = 16.sp,
                    modifier = Modifier
                        .background(colors.primary, RoundedCornerShape(24.dp))
                        .clickable(onClick = onOpenFreeMove)
                        .padding(horizontal = 28.dp, vertical = 12.dp),
                )
                Column(
                    modifier = Modifier.weight(1f),
                    horizontalAlignment = Alignment.End,
                ) {
                    Text(
                        "$ridingCount",
                        fontSize = 22.sp,
                        fontWeight = FontWeight.Bold,
                        color = colors.primary,
                    )
                    Text(t(Str.RidingCountLabel), fontSize = 11.sp, color = Color(0xFF888888))
                }
            }
            Spacer(modifier = Modifier.height(8.dp))
            Row(verticalAlignment = Alignment.CenterVertically) {
                Slider(
                    value = idleHours,
                    onValueChange = { idleHours = it },
                    valueRange = 0f..72f,
                    modifier = Modifier.weight(1f),
                    colors = SliderDefaults.colors(
                        thumbColor = Color(0xFF4CAF50),
                        activeTrackColor = Color(0xFF4CAF50),
                    ),
                )
                Text(
                    t(Str.HoursUnit, idleHours.toInt()),
                    fontSize = 13.sp,
                    color = Color(0xFF333333),
                    modifier = Modifier.padding(start = 8.dp),
                )
            }
            Text(
                t(Str.NoOrderFilter),
                fontSize = 11.sp,
                color = Color(0xFF999999),
                modifier = Modifier.align(Alignment.End),
            )
        }
    }
}
