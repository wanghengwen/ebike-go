package com.luopingtech.ebike.ops.ui.shell

import com.luopingtech.ebike.ops.OpsApp
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.material3.Button
import androidx.compose.material3.FilterChip
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import com.luopingtech.ebike.ops.ui.home.HomeAreaTitleBar
import com.luopingtech.ebike.ops.ui.home.HomeFilterHandle
import com.luopingtech.ebike.ops.ui.home.HomeMapToolsRail
import com.luopingtech.ebike.ops.ui.home.HomeStatItem
import com.luopingtech.ebike.ops.ui.home.HomeStatisticsPanel
import com.luopingtech.ebike.ops.ui.home.homeStatColor
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.foundation.layout.size
import androidx.compose.ui.zIndex
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.control.ControlChannel
import com.luopingtech.ebike.ops.domain.control.VehicleAction
import com.luopingtech.ebike.ops.domain.model.MapPin
import com.luopingtech.ebike.ops.domain.model.Vehicle
import com.luopingtech.ebike.ops.domain.permission.OpsPermissions
import com.luopingtech.ebike.ops.domain.vehicle.VehicleAlarmFilter
import com.luopingtech.ebike.ops.domain.vehicle.VehicleAlarmFilterLogic
import com.luopingtech.ebike.ops.domain.vehicle.VehicleMapFilter
import com.luopingtech.ebike.ops.domain.vehicle.VehicleMapFilterLogic
import com.luopingtech.ebike.ops.feature.home.HomeUiState
import com.luopingtech.ebike.ops.ui.map.OpsMapSpec
import com.luopingtech.ebike.ops.ui.map.OpsMapView
import com.luopingtech.ebike.ops.ui.vehicle.VehicleDetailSection
import kotlinx.coroutines.launch

@Composable
internal fun MapTab(
    app: OpsApp,
    homeState: HomeUiState,
    permissions: OpsPermissions,
    onChangeArea: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val scope = rememberCoroutineScope()
    var filter by remember { mutableStateOf(VehicleMapFilter.All) }
    var selectedAlarms by remember { mutableStateOf(setOf<Int>()) }
    var alarmPanelOpen by remember { mutableStateOf(false) }
    var cardMessage by remember { mutableStateOf<String?>(null) }
    var detailOpen by remember { mutableStateOf(false) }
    var showFence by remember { mutableStateOf(false) }
    var mapTypeSatellite by remember { mutableStateOf(false) }
    var clusterOverview by remember { mutableStateOf(true) }
    var clusterExpandIds by remember { mutableStateOf<List<String>?>(null) }
    val selected = homeState.vehicles.firstOrNull { it.carId == homeState.selectedCarId }
    val detailMap by app.vehicleDetailMapFeature.state.collectAsState()
    val vehicles = homeState.vehicles
    val counts = remember(vehicles) {
        VehicleMapFilterLogic.counts(vehicles)
    }
    val alarmCounts = remember(vehicles) {
        VehicleAlarmFilterLogic.counts(vehicles)
    }
    val filtered = remember(vehicles, filter, selectedAlarms) {
        val alarmActive = selectedAlarms.isNotEmpty()
        vehicles.filter {
            (!alarmActive || !VehicleAlarmFilterLogic.isSoldOut(it.operationStates)) &&
                VehicleMapFilterLogic.matches(
                    ridingState = it.ridingState,
                    operationStates = it.operationStates,
                    restBattery = it.restBattery,
                    filter = filter,
                ) && VehicleAlarmFilterLogic.matchesAlarms(
                alarmStates = it.alarmStates,
                isOnline = it.isOnline,
                selected = selectedAlarms,
            )
        }
    }
    val stateFilters = listOf(
        VehicleMapFilter.Warehouse,
        VehicleMapFilter.Ready,
        VehicleMapFilter.Booking,
        VehicleMapFilter.Riding,
        VehicleMapFilter.TempParking,
        VehicleMapFilter.LowBattery,
        VehicleMapFilter.Repairing,
        VehicleMapFilter.Moving,
    )
    val statItems = stateFilters.map { item ->
        HomeStatItem(
            filter = item,
            label = when (item) {
                VehicleMapFilter.Warehouse -> t(Str.FilterWarehouse)
                VehicleMapFilter.Ready -> t(Str.FilterReady)
                VehicleMapFilter.Booking -> t(Str.FilterBooking)
                VehicleMapFilter.Riding -> t(Str.FilterRiding)
                VehicleMapFilter.TempParking -> t(Str.FilterTempParking)
                VehicleMapFilter.LowBattery -> t(Str.FilterLowBattery)
                VehicleMapFilter.Repairing -> t(Str.FilterRepairing)
                VehicleMapFilter.Moving -> t(Str.FilterMoving)
                VehicleMapFilter.All -> t(Str.FilterAll)
            },
            count = counts[item] ?: 0,
            valueColor = homeStatColor(item),
        )
    }

    Box(modifier = Modifier.fillMaxSize()) {
        MapSurface(
            app = app,
            pins = filtered.map {
                MapPin(
                    id = it.carId,
                    lat = it.lat,
                    lng = it.lng,
                    title = it.carId,
                    subtitle = it.batteryLabel,
                    restBattery = it.restBattery,
                    ridingState = it.ridingState,
                    memberCount = 1,
                    memberIds = listOf(it.carId),
                )
            },
            selectedCarId = homeState.selectedCarId,
            mapReady = homeState.mapReady,
            mapProviderKind = homeState.mapProviderKind,
            clusterOverview = clusterOverview,
            fencePolygons = if (showFence || (detailOpen && detailMap.showFence)) {
                detailMap.fence?.all.orEmpty()
            } else {
                emptyList()
            },
            trackPoints = if (detailOpen && detailMap.showTrack) detailMap.track else emptyList(),
            onPinClick = {
                detailOpen = false
                clusterExpandIds = null
                cardMessage = null
                app.homeFeature.selectVehicle(it)
            },
            onClusterClick = { ids ->
                detailOpen = false
                if (ids.size <= 1) {
                    clusterExpandIds = null
                    app.homeFeature.selectVehicle(ids.firstOrNull())
                } else {
                    clusterOverview = false
                    clusterExpandIds = ids
                }
            },
            modifier = Modifier.fillMaxSize(),
        )

        HomeAreaTitleBar(
            areaName = homeState.currentArea?.name ?: t(Str.NoServiceArea),
            onClick = onChangeArea,
            modifier = Modifier
                .align(Alignment.TopCenter)
                .statusBarsPadding()
                .zIndex(2f),
        )

        HomeMapToolsRail(
            refreshLabel = t(Str.Refresh),
            detailLabel = t(Str.Detail),
            fenceLabel = t(Str.MapToolFence),
            switchLabel = t(Str.MapToolSwitch),
            detailSelected = detailOpen,
            fenceSelected = showFence,
            switchSelected = mapTypeSatellite,
            onRefresh = { scope.launch { app.homeFeature.reloadVehicles() } },
            onDetail = {
                val opening = !detailOpen
                detailOpen = opening
                if (opening && selected != null) {
                    scope.launch { app.homeFeature.refreshSelectedDetail() }
                }
            },
            onFence = {
                showFence = !showFence
                if (showFence && selected != null) {
                    scope.launch { app.homeFeature.refreshSelectedDetail() }
                }
            },
            onSwitch = { mapTypeSatellite = !mapTypeSatellite },
            modifier = Modifier
                .align(Alignment.CenterStart)
                .padding(start = 12.dp)
                .zIndex(2f),
        )

        HomeFilterHandle(
            label = t(Str.FilterHandle),
            open = alarmPanelOpen,
            onClick = { alarmPanelOpen = !alarmPanelOpen },
            modifier = Modifier
                .align(Alignment.CenterEnd)
                .zIndex(2f),
        )

        if (alarmPanelOpen) {
            Surface(
                modifier = Modifier
                    .align(Alignment.CenterEnd)
                    .fillMaxWidth(0.82f)
                    .fillMaxSize()
                    .zIndex(3f),
                color = Color.White,
                shadowElevation = 8.dp,
            ) {
                Column(modifier = Modifier.padding(12.dp)) {
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Text(t(Str.Alarms), fontWeight = FontWeight.Medium)
                        TextButton(onClick = { alarmPanelOpen = false }) { Text(t(Str.Close)) }
                    }
                    AlarmFilterPanel(
                        app = app,
                        selected = selectedAlarms,
                        counts = alarmCounts,
                        onToggle = { code ->
                            selectedAlarms = if (code in selectedAlarms) {
                                selectedAlarms - code
                            } else {
                                selectedAlarms + code
                            }
                        },
                        onClear = { selectedAlarms = emptySet() },
                    )
                }
            }
        }

        clusterExpandIds?.takeIf { it.size > 1 }?.let { ids ->
            Surface(
                tonalElevation = 2.dp,
                modifier = Modifier
                    .align(Alignment.Center)
                    .padding(horizontal = 24.dp)
                    .zIndex(2f),
            ) {
                Column(
                    modifier = Modifier.padding(12.dp),
                    verticalArrangement = Arrangement.spacedBy(4.dp),
                ) {
                    Text(
                        t(Str.ClusterVehicles, ids.size),
                        style = MaterialTheme.typography.titleSmall,
                    )
                    ids.take(10).forEach { id ->
                        TextButton(
                            onClick = {
                                clusterExpandIds = null
                                app.homeFeature.selectVehicle(id)
                            },
                        ) { Text(id) }
                    }
                    TextButton(onClick = { clusterExpandIds = null }) { Text(t(Str.Close)) }
                }
            }
        }

        Column(
            modifier = Modifier
                .align(Alignment.BottomCenter)
                .fillMaxWidth()
                .zIndex(2f),
        ) {
            if (selected != null || cardMessage != null) {
                SelectedVehicleCard(
                    app = app,
                    vehicle = selected,
                    message = cardMessage,
                    detailOpen = detailOpen,
                    serviceAreaId = homeState.currentArea?.id,
                    showChangeBattery = permissions.showChangeBattery,
                    showDetails = permissions.canScanDetails || permissions.showMap,
                    canBindBattery = permissions.canBindBatterySn,
                    onRing = {
                        scope.launch {
                            val carId = selected?.carId ?: return@launch
                            cardMessage = when (
                                val r = app.vehicleControl.execute(
                                    carId,
                                    VehicleAction.Ring,
                                    ControlChannel.BlePreferred,
                                )
                            ) {
                                is OpsResult.Ok -> t(Str.RingOk, carId)
                                is OpsResult.Err -> t(Str.RingFailed, r.error.message)
                            }
                        }
                    },
                    onOpenBattery = {
                        scope.launch {
                            val carId = selected?.carId ?: return@launch
                            cardMessage = when (
                                val r = app.vehicleControl.execute(
                                    carId,
                                    VehicleAction.OpenBatteryBox,
                                    ControlChannel.BlePreferred,
                                )
                            ) {
                                is OpsResult.Ok -> t(Str.OpenBoxOkShort, carId)
                                is OpsResult.Err -> t(Str.OpenBoxFailedShort, r.error.message)
                            }
                        }
                    },
                    onFinishSwap = {
                        scope.launch {
                            val carId = selected?.carId ?: return@launch
                            cardMessage = when (
                                val r = app.vehicleControl.execute(
                                    carId,
                                    VehicleAction.CloseBatteryBox,
                                    ControlChannel.BlePreferred,
                                )
                            ) {
                                is OpsResult.Ok -> {
                                    app.homeFeature.reloadVehicles()
                                    t(Str.FinishSwapOk, carId)
                                }
                                is OpsResult.Err -> t(Str.FinishSwapFailed, r.error.message)
                            }
                        }
                    },
                    onDetails = {
                        val opening = !detailOpen
                        detailOpen = opening
                        if (opening) {
                            scope.launch { app.homeFeature.refreshSelectedDetail() }
                        }
                    },
                )
            }
            homeState.errorMessage?.let {
                Text(
                    text = it,
                    color = MaterialTheme.colorScheme.error,
                    modifier = Modifier
                        .fillMaxWidth()
                        .background(Color.White.copy(alpha = 0.9f))
                        .padding(horizontal = 16.dp, vertical = 4.dp),
                )
            }
            HomeStatisticsPanel(
                items = statItems,
                selected = filter,
                onSelect = { filter = it },
            )
        }
    }
}

@Composable
private fun SelectedVehicleCard(
    app: OpsApp,
    vehicle: Vehicle?,
    message: String?,
    detailOpen: Boolean,
    serviceAreaId: String? = null,
    showChangeBattery: Boolean,
    showDetails: Boolean,
    canBindBattery: Boolean = false,
    onRing: () -> Unit,
    onOpenBattery: () -> Unit,
    onFinishSwap: () -> Unit,
    onDetails: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    Surface(
        tonalElevation = 2.dp,
        modifier = Modifier.fillMaxWidth(),
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            if (vehicle == null) {
                Text(
                    text = t(Str.MapPickVehicle),
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            } else {
                Text(
                    text = "${vehicle.carId} · ${vehicle.batteryLabel} · ${vehicle.ridingLabel}" +
                        if (vehicle.isOnline) " · ${t(Str.Online)}" else " · ${t(Str.Offline)}",
                    style = MaterialTheme.typography.titleSmall,
                )
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    Button(onClick = onRing, modifier = Modifier.weight(1f)) { Text(t(Str.Ring)) }
                    if (showChangeBattery) {
                        Button(onClick = onOpenBattery, modifier = Modifier.weight(1f)) {
                            Text(t(Str.OpenBatteryBox))
                        }
                        Button(onClick = onFinishSwap, modifier = Modifier.weight(1f)) {
                            Text(t(Str.FinishChangeBattery))
                        }
                    }
                    if (showDetails) {
                        TextButton(onClick = onDetails) {
                            Text(if (detailOpen) t(Str.Collapse) else t(Str.Detail))
                        }
                    }
                }
                if (detailOpen) {
                    VehicleDetailSection(
                        app = app,
                        vehicle = vehicle,
                        serviceAreaId = serviceAreaId,
                        canBindBattery = canBindBattery,
                    )
                }
            }
            message?.let { Text(text = it, style = MaterialTheme.typography.bodySmall) }
        }
    }
}

@Composable
private fun AlarmFilterPanel(
    app: OpsApp,
    selected: Set<Int>,
    counts: Map<VehicleAlarmFilter, Int>,
    onToggle: (Int) -> Unit,
    onClear: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    Surface(
        tonalElevation = 2.dp,
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 12.dp),
    ) {
        Column(
            modifier = Modifier.padding(12.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Text(t(Str.Alarms), style = MaterialTheme.typography.titleSmall)
                TextButton(onClick = onClear, enabled = selected.isNotEmpty()) {
                    Text(t(Str.Clear))
                }
            }
            Text(
                text = t(Str.AlarmFilterHint),
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
            VehicleAlarmFilter.entries.chunked(3).forEach { row ->
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(6.dp),
                ) {
                    row.forEach { item ->
                        FilterChip(
                            selected = item.code in selected,
                            onClick = { onToggle(item.code) },
                            label = {
                                Text("${item.label} ${counts[item] ?: 0}")
                            },
                            modifier = Modifier.weight(1f),
                        )
                    }
                    repeat(3 - row.size) {
                        Spacer(modifier = Modifier.weight(1f))
                    }
                }
            }
        }
    }
}

@Composable
private fun MapSurface(
    app: OpsApp,
    pins: List<MapPin>,
    selectedCarId: String?,
    mapReady: Boolean,
    mapProviderKind: String,
    onPinClick: (String) -> Unit,
    onClusterClick: (List<String>) -> Unit,
    modifier: Modifier = Modifier,
    clusterOverview: Boolean = true,
    fencePolygons: List<com.luopingtech.ebike.ops.domain.model.FencePolygon> = emptyList(),
    trackPoints: List<com.luopingtech.ebike.ops.domain.model.TrackPoint> = emptyList(),
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    when {
        mapProviderKind.equals("none", ignoreCase = true) && !mapReady -> {
            Box(
                modifier = modifier
                    .background(MaterialTheme.colorScheme.surfaceVariant)
                    .border(1.dp, MaterialTheme.colorScheme.outlineVariant)
                    .padding(12.dp),
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    text = t(Str.MapNotEnabled),
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    textAlign = TextAlign.Center,
                )
            }
        }
        else -> {
            val label = when {
                mapProviderKind.equals("simulator", ignoreCase = true) -> t(Str.SimulatorMap)
                mapProviderKind.equals("tencent", ignoreCase = true) -> t(Str.MapSimNoTencentKey)
                mapProviderKind.equals("google", ignoreCase = true) -> t(Str.MapSimNoGoogle)
                else -> "${t(Str.SimulatorMap)} · $mapProviderKind"
            }
            OpsMapView(
                spec = OpsMapSpec(
                    pins = pins,
                    selectedCarId = selectedCarId,
                    providerLabel = label,
                    onSelectCarId = onPinClick,
                    onSelectCluster = onClusterClick,
                    clusterOverview = clusterOverview,
                    fencePolygons = fencePolygons,
                    trackPoints = trackPoints,
                ),
                modifier = modifier,
            )
        }
    }
}
