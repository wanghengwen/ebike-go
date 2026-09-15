package com.luopingtech.ebike.ops.ui.task

import androidx.compose.foundation.background
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
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.RangeSlider
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
import com.luopingtech.ebike.ops.domain.model.MapPin
import com.luopingtech.ebike.ops.domain.vehicle.VehicleRidingStates
import com.luopingtech.ebike.ops.ui.home.HomeFilterHandle
import com.luopingtech.ebike.ops.ui.map.OpsMapSpec
import com.luopingtech.ebike.ops.ui.map.OpsMapView
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import kotlinx.coroutines.launch

/**
 * 换电地图 —— 对齐遗留 TaskMapActivity + HomeChangeBatteryFragment。
 */
@Composable
fun ChangeBatteryMapScreen(
    app: OpsApp,
    onClose: () -> Unit,
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

    var batteryMin by remember { mutableFloatStateOf(0f) }
    var batteryMax by remember { mutableFloatStateOf(100f) }
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

    LaunchedEffect(home.currentArea?.id) {
        app.homeFeature.reloadVehicles()
        home.currentArea?.id?.takeIf { it.isNotBlank() }?.let {
            app.vehicleDetailMapFeature.loadFence(it)
        }
    }

    val filtered = remember(home.vehicles, batteryMin, batteryMax, hideRiding) {
        home.vehicles.filter { v ->
            val b = v.restBattery
            val inBattery = b in batteryMin.toInt()..batteryMax.toInt()
            val hasCoord = v.lat != 0.0 || v.lng != 0.0
            val ridingOk = !hideRiding || v.ridingState != VehicleRidingStates.RIDING
            inBattery && hasCoord && ridingOk
        }
    }
    val high = filtered.count { it.restBattery >= 60 }
    val mid = filtered.count { it.restBattery in 30..59 }
    val low = filtered.count { it.restBattery < 30 }
    val pins = remember(filtered) {
        filtered.map {
            MapPin(
                id = it.carId,
                lat = it.lat,
                lng = it.lng,
                title = it.carId,
                subtitle = "${it.restBattery}%",
                restBattery = it.restBattery,
                ridingState = it.ridingState,
                memberIds = listOf(it.carId),
            )
        }
    }
    val fencePolygons = if (showFence) detailMap.fence?.all.orEmpty() else emptyList()

    Column(modifier = Modifier.fillMaxSize().background(Color.White)) {
        TaskFullscreenTopBar(
            title = t(Str.ChangeBatteryMapTitle),
            primary = colors.primary,
            onBack = onClose,
            trailingLabel = t(Str.ChangeBatteryStats),
            onTrailing = onOpenStats,
        )
        Box(modifier = Modifier.weight(1f).fillMaxWidth()) {
            OpsMapView(
                spec = OpsMapSpec(
                    pins = pins,
                    selectedCarId = selectedCarId,
                    providerLabel = t(Str.ChangeBatteryMapTitle),
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
                ),
                modifier = Modifier.fillMaxSize(),
            )
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
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Column {
                    Text(
                        "${filtered.size}",
                        fontSize = 22.sp,
                        fontWeight = FontWeight.Bold,
                        color = Color(0xFF333333),
                    )
                    Text(t(Str.FilterResultVehicles), fontSize = 11.sp, color = Color(0xFF888888))
                }
                BatteryStat(high.toString(), t(Str.BatteryHigh), Color(0xFF4CAF50))
                BatteryStat(mid.toString(), t(Str.BatteryMid), Color(0xFFFFC107))
                BatteryStat(low.toString(), t(Str.BatteryLow), Color(0xFFE53935))
            }
            Spacer(modifier = Modifier.height(8.dp))
            RangeSlider(
                value = batteryMin..batteryMax,
                onValueChange = {
                    batteryMin = it.start
                    batteryMax = it.endInclusive
                },
                valueRange = 0f..100f,
                colors = SliderDefaults.colors(
                    thumbColor = Color(0xFF4CAF50),
                    activeTrackColor = Color(0xFF4CAF50),
                ),
            )
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                Text("0%", fontSize = 11.sp, color = Color(0xFF999999))
                Text("100%", fontSize = 11.sp, color = Color(0xFF999999))
            }
        }
    }
}

@Composable
private fun BatteryStat(value: String, label: String, color: Color) {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Text(value, fontSize = 20.sp, fontWeight = FontWeight.Bold, color = color)
        Text(label, fontSize = 11.sp, color = color)
    }
}
