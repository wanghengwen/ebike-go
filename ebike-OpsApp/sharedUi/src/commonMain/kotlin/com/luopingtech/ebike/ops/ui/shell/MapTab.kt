package com.luopingtech.ebike.ops.ui.shell

import com.luopingtech.ebike.ops.OpsApp
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.BiasAlignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.zIndex
import com.luopingtech.ebike.ops.ui.home.HomeAlarmFilterSlide
import com.luopingtech.ebike.ops.ui.home.HomeAreaTitleBar
import com.luopingtech.ebike.ops.ui.home.HomeFilterHandle
import com.luopingtech.ebike.ops.ui.home.HomeMapToolsRail
import com.luopingtech.ebike.ops.ui.home.HomeStatItem
import com.luopingtech.ebike.ops.ui.home.HomeStatisticsPanel
import com.luopingtech.ebike.ops.ui.home.HomeVehiclePopupCard
import com.luopingtech.ebike.ops.ui.home.buildHomeVehicleStatusChips
import com.luopingtech.ebike.ops.ui.home.homeStatColor
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.control.ControlChannel
import com.luopingtech.ebike.ops.domain.control.VehicleAction
import com.luopingtech.ebike.ops.domain.model.MapPin
import com.luopingtech.ebike.ops.domain.model.Vehicle
import com.luopingtech.ebike.ops.domain.model.homeMapBadgeDrawableName
import com.luopingtech.ebike.ops.domain.model.homeMapPinIcon
import com.luopingtech.ebike.ops.ui.home.SimpleCarListScreen
import com.luopingtech.ebike.ops.domain.permission.OpsPermissions
import com.luopingtech.ebike.ops.domain.vehicle.VehicleAlarmFilterLogic
import com.luopingtech.ebike.ops.domain.vehicle.VehicleMapFilter
import com.luopingtech.ebike.ops.domain.vehicle.VehicleMapFilterLogic
import com.luopingtech.ebike.ops.feature.home.HomeUiState
import com.luopingtech.ebike.ops.ui.map.OpsMapSpec
import com.luopingtech.ebike.ops.ui.map.OpsMapView
import com.luopingtech.ebike.ops.ui.vehicle.LocalOpenVehicleDetail
import androidx.compose.ui.platform.LocalClipboardManager
import androidx.compose.ui.text.AnnotatedString
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
    val openVehicleDetail = LocalOpenVehicleDetail.current
    val clipboard = LocalClipboardManager.current
    val scope = rememberCoroutineScope()
    var filter by remember { mutableStateOf(VehicleMapFilter.All) }
    var selectedAlarms by remember { mutableStateOf(setOf<Int>()) }
    var alarmPanelOpen by remember { mutableStateOf(false) }
    var cardMessage by remember { mutableStateOf<String?>(null) }
    /** 左侧围栏：停车区/禁停区显隐（服务区轮廓始终可显示）。 */
    var showFence by remember { mutableStateOf(false) }
    var mapTypeSatellite by remember { mutableStateOf(false) }
    /** true=聚合打点；左侧「详情」选中时为 false。默认跟 12300401（有码优先聚合）。 */
    var clusterOverview by remember(permissions.homeClusterFirst) {
        mutableStateOf(permissions.homeClusterFirst)
    }
    var clusterList by remember { mutableStateOf<List<Vehicle>?>(null) }
    /** 对齐 locationToServiceByCalculate：切服务区 / 围栏到位后 fit 服务区，而不是 fit 全部车点。 */
    var serviceFitNonce by remember { mutableIntStateOf(0) }
    val selected = homeState.vehicles.firstOrNull { it.carId == homeState.selectedCarId }
    val detailMap by app.vehicleDetailMapFeature.state.collectAsState()
    val vehicles = homeState.vehicles
    val areaId = homeState.currentArea?.id?.takeIf { it.isNotBlank() }
    val showStatistics = permissions.showHomeStatistics
    val showAlarmFilter = permissions.showHomeAlarmFilter

    // 进页/切服务区时预拉围栏（对齐原版进页就拉，按钮只切显隐）
    LaunchedEffect(areaId) {
        showFence = false
        clusterOverview = permissions.homeClusterFirst
        mapTypeSatellite = false
        clusterList = null
        if (areaId != null) {
            app.vehicleDetailMapFeature.loadFence(areaId)
        }
    }

    val fencePolygons = remember(detailMap.fence, showFence) {
        val bundle = detailMap.fence
        val service = bundle?.serviceAreas.orEmpty()
        val parking = if (showFence) {
            bundle?.parkings.orEmpty() + bundle?.noParkings.orEmpty()
        } else {
            emptyList()
        }
        service + parking
    }

    // 服务区轮廓到位后缩放到服务区（对齐 BaseHomeMapFragment.locationToServiceByCalculate）
    LaunchedEffect(areaId, detailMap.fence?.serviceAreas) {
        if (!detailMap.fence?.serviceAreas.isNullOrEmpty()) {
            serviceFitNonce += 1
        }
    }
    val counts = remember(vehicles) {
        VehicleMapFilterLogic.counts(vehicles)
    }
    val filtered = remember(vehicles, filter, selectedAlarms) {
        vehicles.filter {
            !VehicleAlarmFilterLogic.isSoldOut(it.operationStates) &&
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
                    icon = it.homeMapPinIcon(),
                    badgeDrawableName = it.homeMapBadgeDrawableName(),
                )
            },
            selectedCarId = homeState.selectedCarId,
            mapReady = homeState.mapReady,
            mapProviderKind = homeState.mapProviderKind,
            clusterOverview = clusterOverview,
            fencePolygons = fencePolygons,
            trackPoints = emptyList(),
            mapTypeSatellite = mapTypeSatellite,
            // 首页不刷 debug 条；不因车点变化反复 fit（原版 fit 服务区）
            showStatusOverlay = false,
            autoFitOnPins = false,
            animateToSelection = false,
            fitNonce = serviceFitNonce,
            onPinClick = {
                clusterList = null
                cardMessage = null
                app.homeFeature.selectVehicle(it)
                // 对齐点车后 fetchCarStatus：刷新该车电量等到列表缓存
                scope.launch {
                    app.vehicleFeature.refreshDetail(it)
                }
            },
            onClusterClick = { ids ->
                val idSet = ids.toSet()
                val list = filtered.filter { it.carId in idSet }
                if (list.size <= 1) {
                    clusterList = null
                    app.homeFeature.selectVehicle(list.firstOrNull()?.carId ?: ids.firstOrNull())
                } else {
                    // Legacy handleClusterClick at zoom≥16: SimpleCarListActivity, keep cluster mode.
                    clusterList = list
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

        if (selected != null || cardMessage != null) {
            val chips = selected?.let { v ->
                buildHomeVehicleStatusChips(
                    vehicle = v,
                    canUseLabel = t(Str.VehicleCanUse),
                    helmetNotClosedLabel = t(Str.HelmetNotClosed),
                    offlineLabel = t(Str.Offline),
                    soldOutLabel = t(Str.OpsSoldOut),
                    lowBatteryLabel = t(Str.FilterLowBattery),
                    repairingLabel = t(Str.FilterRepairing),
                    movingLabel = t(Str.FilterMoving),
                    ridingLabel = t(Str.FilterRiding),
                    tempParkingLabel = t(Str.FilterTempParking),
                    bookingLabel = t(Str.FilterBooking),
                )
            }.orEmpty()
            HomeVehiclePopupCard(
                vehicle = selected,
                message = cardMessage,
                detailOpen = false,
                showUnlock = permissions.canScanUnlock,
                showDetails = permissions.canScanDetails || permissions.showMap,
                carNumberLabel = t(Str.VehicleCarNumberColon).let { s ->
                    if (s.endsWith(":") || s.endsWith("：")) s else "$s："
                },
                deviceNoLabel = t(Str.VehicleDeviceNoColon),
                statusLabel = t(Str.VehicleStatusColon),
                ringLabel = t(Str.Ring),
                unlockLabel = t(Str.ScanUnlock),
                lockLabel = t(Str.ScanLock),
                detailLabel = t(Str.Detail),
                pickHint = t(Str.MapPickVehicle),
                statusChips = chips,
                detailContent = null,
                onCopyCarId = {
                    val id = selected?.carId?.takeIf { it.isNotBlank() } ?: return@HomeVehiclePopupCard
                    clipboard.setText(AnnotatedString(id))
                    cardMessage = t(Str.Copied)
                },
                onCopyImei = {
                    val imei = selected?.imei?.takeIf { it.isNotBlank() } ?: return@HomeVehiclePopupCard
                    clipboard.setText(AnnotatedString(imei))
                    cardMessage = t(Str.Copied)
                },
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
                onUnlock = {
                    scope.launch {
                        val carId = selected?.carId ?: return@launch
                        if (!permissions.canScanUnlock) {
                            cardMessage = t(Str.NoUnlockPermission)
                            return@launch
                        }
                        cardMessage = when (
                            val r = app.vehicleControl.execute(
                                carId,
                                VehicleAction.Unlock,
                                ControlChannel.BlePreferred,
                            )
                        ) {
                            is OpsResult.Ok -> t(Str.ActionOk, t(Str.ScanUnlock), carId)
                            is OpsResult.Err -> t(Str.ActionFailed, t(Str.ScanUnlock), r.error.message)
                        }
                    }
                },
                onLock = {
                    scope.launch {
                        val carId = selected?.carId ?: return@launch
                        if (!permissions.canScanUnlock) {
                            cardMessage = t(Str.NoUnlockPermission)
                            return@launch
                        }
                        cardMessage = when (
                            val r = app.vehicleControl.execute(
                                carId,
                                VehicleAction.Lock,
                                ControlChannel.BlePreferred,
                            )
                        ) {
                            is OpsResult.Ok -> t(Str.ActionOk, t(Str.ScanLock), carId)
                            is OpsResult.Err -> t(Str.ActionFailed, t(Str.ScanLock), r.error.message)
                        }
                    }
                },
                onDetails = {
                    val v = selected ?: return@HomeVehiclePopupCard
                    // 对齐 HomeVehiclePop → CarDetailActivity：先气泡，再整页详情。
                    scope.launch {
                        val fresh = when (val r = app.vehicleFeature.refreshDetail(v.carId)) {
                            is OpsResult.Ok -> r.value
                            is OpsResult.Err -> v
                        }
                        openVehicleDetail.open(fresh, homeState.currentArea?.id)
                    }
                },
                cardModifier = Modifier
                    .align(Alignment.TopCenter)
                    // Legacy HomeVehiclePop: Gravity.TOP + 85dp
                    .statusBarsPadding()
                    .padding(top = 52.dp)
                    .zIndex(2f),
            )
        }

        HomeMapToolsRail(
            refreshLabel = t(Str.Refresh),
            detailLabel = t(Str.Detail),
            fenceLabel = t(Str.MapToolFence),
            switchLabel = t(Str.MapToolSwitch),
            // 选中=单车点模式（非聚合），对齐 obsVehicleDetailSelect
            detailSelected = !clusterOverview,
            fenceSelected = showFence,
            switchSelected = mapTypeSatellite,
            onRefresh = {
                scope.launch {
                    app.homeFeature.reloadVehicles()
                    areaId?.let { app.vehicleDetailMapFeature.loadFence(it) }
                }
            },
            onDetail = {
                // 对齐 onDetailSwitch：选中详情 → 关聚合；取消 → 开聚合
                clusterOverview = !clusterOverview
                clusterList = null
            },
            onFence = {
                val next = !showFence
                showFence = next
                if (next && detailMap.fence == null) {
                    scope.launch {
                        areaId?.let { app.vehicleDetailMapFeature.loadFence(it) }
                    }
                }
            },
            onSwitch = { mapTypeSatellite = !mapTypeSatellite },
            modifier = Modifier
                .align(Alignment.CenterStart)
                // Legacy home_map_control_layout_v3：底栏上方留白，避免压住统计
                .padding(start = 12.dp, bottom = if (showStatistics) 148.dp else 24.dp)
                .zIndex(4f),
        )

        if (showAlarmFilter && !alarmPanelOpen) {
            HomeFilterHandle(
                label = t(Str.FilterHandle),
                open = false,
                onClick = { alarmPanelOpen = true },
                modifier = Modifier
                    // Legacy fragment_home_v3 viewHomeFilterSwitch vertical_bias=0.55
                    .align(BiasAlignment(horizontalBias = 1f, verticalBias = 0.55f))
                    .zIndex(2f),
            )
        }

        if (showAlarmFilter) {
            HomeAlarmFilterSlide(
                visible = alarmPanelOpen,
                applied = selectedAlarms,
                title = t(Str.AlarmFilterTitle),
                resetLabel = t(Str.StationFilterReset),
                sureLabel = t(Str.FilterSure),
                onDismiss = { alarmPanelOpen = false },
                onApply = { selectedAlarms = it },
            )
        }

        clusterList?.takeIf { it.size > 1 }?.let { list ->
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .zIndex(20f),
            ) {
                SimpleCarListScreen(
                    vehicles = list,
                    title = t(Str.VehicleList),
                    backLabel = t(Str.Back),
                    colId = t(Str.VehicleTagCarId),
                    colStatus = t(Str.VehicleListColStatus),
                    colBattery = t(Str.VehicleListColBattery),
                    onBack = { clusterList = null },
                    onVehicleClick = { vehicle ->
                        clusterList = null
                        scope.launch {
                            val fresh = when (val r = app.vehicleFeature.refreshDetail(vehicle.carId)) {
                                is OpsResult.Ok -> r.value
                                is OpsResult.Err -> vehicle
                            }
                            openVehicleDetail.open(fresh, homeState.currentArea?.id)
                        }
                    },
                )
            }
        }

        Column(
            modifier = Modifier
                .align(Alignment.BottomCenter)
                .fillMaxWidth()
                .zIndex(2f),
        ) {
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
            if (showStatistics) {
                HomeStatisticsPanel(
                    items = statItems,
                    selected = filter,
                    onSelect = { filter = it },
                )
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
    mapTypeSatellite: Boolean = false,
    showStatusOverlay: Boolean = false,
    autoFitOnPins: Boolean = false,
    animateToSelection: Boolean = false,
    fitNonce: Int = 0,
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
                    mapTypeSatellite = mapTypeSatellite,
                    showStatusOverlay = showStatusOverlay,
                    autoFitOnPins = autoFitOnPins,
                    animateToSelection = animateToSelection,
                    fitNonce = fitNonce,
                ),
                modifier = modifier,
            )
        }
    }
}
