package com.luopingtech.ebike.ops.ui.vehicle

import androidx.activity.compose.BackHandler
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Surface
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
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.zIndex
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.R
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.control.ControlChannel
import com.luopingtech.ebike.ops.domain.control.VehicleAction
import com.luopingtech.ebike.ops.domain.model.MapPin
import com.luopingtech.ebike.ops.domain.model.Vehicle
import com.luopingtech.ebike.ops.domain.model.batteryMapPinIcon
import com.luopingtech.ebike.ops.domain.permission.OpsPermissions
import com.luopingtech.ebike.ops.domain.vehicle.VehicleAlarmStates
import com.luopingtech.ebike.ops.domain.vehicle.VehicleOperationStates
import com.luopingtech.ebike.ops.domain.vehicle.VehicleRidingStates
import com.luopingtech.ebike.ops.ui.feedback.LocalOpsToast
import com.luopingtech.ebike.ops.ui.feedback.OpsTipDialogHost
import com.luopingtech.ebike.ops.ui.feedback.OpsTipState
import com.luopingtech.ebike.ops.ui.map.TencentMapView
import com.luopingtech.ebike.ops.ui.movecar.FieldMoveCarScreen
import com.luopingtech.ebike.ops.ui.report.FaultReportScreen
import com.luopingtech.ebike.ops.ui.scan.LocalOpsScanPreview
import com.luopingtech.ebike.ops.ui.tag.VehicleTagScreen
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import com.luopingtech.ebike.ops.ui.tools.FieldChangeBatteryScreen
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale

private enum class DetailToolOverlay {
    ChangeBattery,
    MoveCar,
    FaultReport,
    VehicleTag,
}

/**
 * Legacy CarDetailActivity: map + info + menus + action grid.
 */
@Composable
fun VehicleDetailScreen(
    app: OpsApp,
    vehicle: Vehicle,
    serviceAreaId: String?,
    permissions: OpsPermissions,
    mapReady: Boolean,
    mapProviderKind: String,
    onClose: () -> Unit,
    onOpenOrderInfo: () -> Unit = {},
    modifier: Modifier = Modifier,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val scope = rememberCoroutineScope()
    val detailMap by app.vehicleDetailMapFeature.state.collectAsState()
    var actionTip by remember { mutableStateOf<OpsTipState>(OpsTipState.Hidden) }
    var mapExpanded by remember { mutableStateOf(false) }
    var noteOpen by remember { mutableStateOf(false) }
    var moreInfoOpen by remember { mutableStateOf(false) }
    var orderHistoryOpen by remember { mutableStateOf(false) }
    var scanLogOpen by remember { mutableStateOf(false) }
    var switchLockOpen by remember { mutableStateOf(false) }
    var changeBatteryHistoryOpen by remember { mutableStateOf(false) }
    var moveCarHistoryOpen by remember { mutableStateOf(false) }
    var locateNonce by remember { mutableIntStateOf(0) }
    var fitNonce by remember { mutableIntStateOf(0) }
    val homeState by app.homeFeature.state.collectAsState()
    val resolvedServiceId = serviceAreaId?.takeIf { it.isNotBlank() }
        ?: homeState.currentArea?.id?.takeIf { it.isNotBlank() }
        ?: vehicle.serviceId.takeIf { it.isNotBlank() }
    val scanner = LocalOpsScanPreview.current
    val toast = LocalOpsToast.current
    var toolOverlay by remember { mutableStateOf<DetailToolOverlay?>(null) }

    BackHandler {
        when {
            toolOverlay != null -> toolOverlay = null
            orderHistoryOpen -> orderHistoryOpen = false
            moveCarHistoryOpen -> moveCarHistoryOpen = false
            changeBatteryHistoryOpen -> changeBatteryHistoryOpen = false
            switchLockOpen -> switchLockOpen = false
            scanLogOpen -> scanLogOpen = false
            moreInfoOpen -> moreInfoOpen = false
            noteOpen -> noteOpen = false
            else -> onClose()
        }
    }

    LaunchedEffect(vehicle.carId, resolvedServiceId) {
        app.vehicleDetailMapFeature.loadForVehicle(vehicle, resolvedServiceId)
        app.homeFeature.refreshSelectedDetail()
    }

    val fencePolygons = remember(detailMap.fence, detailMap.showFence) {
        if (!detailMap.showFence) emptyList() else detailMap.fence?.all.orEmpty()
    }
    val trackPoints = remember(detailMap.track, detailMap.showTrack) {
        if (detailMap.showTrack) detailMap.track else emptyList()
    }
    val pin = MapPin(
        id = vehicle.carId,
        lat = vehicle.lat,
        lng = vehicle.lng,
        title = vehicle.carId,
        subtitle = vehicle.batteryLabel,
        restBattery = vehicle.restBattery,
        ridingState = vehicle.ridingState,
        memberCount = 1,
        memberIds = listOf(vehicle.carId),
        icon = vehicle.batteryMapPinIcon(),
    )
    val idleBubble = remember(vehicle.lockTimeMs, vehicle.unlockTimeMs, vehicle.ridingState, language) {
        formatNoOrderDuration(app, vehicle)
    }
    val locateTime = remember(vehicle.timestamp) { formatLocateTimestamp(vehicle.timestamp) }

    fun successMessage(action: VehicleAction): String = when (action) {
        VehicleAction.Ring -> t(Str.FindCarSoundOk)
        VehicleAction.Unlock -> t(Str.UnlockSuccess)
        VehicleAction.Lock -> t(Str.LockSuccess)
        VehicleAction.OpenBatteryBox -> t(Str.OpenBatterySuccess)
        VehicleAction.CloseBatteryBox -> t(Str.CloseBatterySuccess)
        VehicleAction.OpenHelmetLock -> t(Str.OpenHelmetSuccess)
        VehicleAction.CloseHelmetLock -> t(Str.CloseHelmetSuccess)
        else -> t(Str.ActionOk, "", "")
    }

    fun failMessage(action: VehicleAction, apiMsg: String): String {
        val trimmed = apiMsg.trim()
        if (trimmed.isNotEmpty()) return trimmed
        return when (action) {
            VehicleAction.Unlock -> t(Str.UnlockFailed)
            VehicleAction.Lock -> t(Str.LockFailed)
            else -> t(Str.ActionFailed, "", "")
        }
    }

    fun runAction(action: VehicleAction) {
        scope.launch {
            // Legacy: WaitDialog「操作车辆中」→ TipDialog SUCCESS/ERROR
            actionTip = OpsTipState.Waiting(t(Str.HandleVehicle))
            var imei = vehicle.imei.trim()
            if (imei.isBlank() &&
                (action == VehicleAction.OpenHelmetLock ||
                    action == VehicleAction.CloseHelmetLock ||
                    action == VehicleAction.Unlock ||
                    action == VehicleAction.Lock)
            ) {
                when (val detail = app.vehicleFeature.refreshDetail(vehicle.carId)) {
                    is OpsResult.Ok -> imei = detail.value.imei.trim()
                    is OpsResult.Err -> Unit
                }
            }
            when (
                val r = app.vehicleControl.execute(
                    vehicleId = vehicle.carId,
                    action = action,
                    channel = ControlChannel.NetworkOnly,
                    imei = imei,
                )
            ) {
                is OpsResult.Ok -> {
                    actionTip = OpsTipState.Success(successMessage(action))
                    delay(1_000)
                    actionTip = OpsTipState.Hidden
                    when (action) {
                        VehicleAction.Unlock,
                        VehicleAction.Lock,
                        VehicleAction.OpenBatteryBox,
                        VehicleAction.CloseBatteryBox,
                        -> {
                            app.homeFeature.refreshSelectedDetail()
                            app.vehicleDetailMapFeature.loadForVehicle(vehicle, resolvedServiceId)
                        }
                        else -> Unit
                    }
                }
                is OpsResult.Err -> {
                    actionTip = OpsTipState.Error(failMessage(action, r.error.message))
                    delay(2_000)
                    actionTip = OpsTipState.Hidden
                }
            }
        }
    }

    Surface(modifier = modifier, color = Color(0xFFF4F6FF)) {
        Box(modifier = Modifier.fillMaxSize()) {
            Column(modifier = Modifier.fillMaxSize()) {
                // Title
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .background(OpsTheme.colors.primary)
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
                        text = t(Str.OrderQueryOpenVehicle),
                        color = Color.White,
                        fontSize = 18.sp,
                        fontWeight = FontWeight.Medium,
                        modifier = Modifier.align(Alignment.Center),
                    )
                    Text(
                        text = "\u21BB",
                        color = Color.White,
                        fontSize = 22.sp,
                        modifier = Modifier
                            .align(Alignment.CenterEnd)
                            .clickable(
                                interactionSource = remember { MutableInteractionSource() },
                                indication = null,
                                onClick = {
                                    scope.launch {
                                        app.homeFeature.refreshSelectedDetail()
                                        app.vehicleDetailMapFeature.loadForVehicle(vehicle, serviceAreaId)
                                        fitNonce++
                                    }
                                },
                            )
                            .padding(horizontal = 16.dp),
                    )
                }

                // Map
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .weight(if (mapExpanded) 1f else 0.38f),
                ) {
                    if (mapReady && mapProviderKind.equals("tencent", ignoreCase = true)) {
                        TencentMapView(
                            pins = listOf(pin),
                            selectedCarId = vehicle.carId,
                            onSelectCarId = {},
                            modifier = Modifier.fillMaxSize(),
                            clusterOverview = false,
                            fencePolygons = fencePolygons,
                            trackPoints = trackPoints,
                            followNonce = locateNonce,
                            followLat = vehicle.lat.takeIf { it != 0.0 },
                            followLng = vehicle.lng.takeIf { it != 0.0 },
                            fitNonce = fitNonce,
                            showStatusOverlay = false,
                        )
                    } else {
                        Box(
                            modifier = Modifier
                                .fillMaxSize()
                                .background(Color(0xFFE8ECF4)),
                            contentAlignment = Alignment.Center,
                        ) {
                            Text(t(Str.MapNotEnabled), color = Color(0xFF666666))
                        }
                    }
                    idleBubble?.let { tip ->
                        Surface(
                            modifier = Modifier
                                .align(Alignment.TopCenter)
                                .padding(top = 48.dp)
                                .padding(horizontal = 24.dp),
                            shape = RoundedCornerShape(8.dp),
                            color = Color.White,
                            shadowElevation = 4.dp,
                        ) {
                            Text(
                                text = tip,
                                color = Color(0xFFE02020),
                                fontSize = 13.sp,
                                modifier = Modifier.padding(horizontal = 12.dp, vertical = 8.dp),
                            )
                        }
                    }
                    Column(
                        modifier = Modifier
                            .align(Alignment.CenterStart)
                            .padding(start = 10.dp)
                            .background(Color.White.copy(alpha = 0.92f), RoundedCornerShape(10.dp))
                            .padding(vertical = 6.dp, horizontal = 4.dp),
                        horizontalAlignment = Alignment.CenterHorizontally,
                    ) {
                        DetailMapTool(t(Str.MapToolFullscreen)) { mapExpanded = !mapExpanded }
                        DetailMapTool(t(Str.VehicleTrack)) {
                            if (!detailMap.showTrack) {
                                scope.launch { app.vehicleDetailMapFeature.loadTrack(vehicle) }
                            } else {
                                app.vehicleDetailMapFeature.setShowTrack(false)
                            }
                        }
                        DetailMapTool(t(Str.MapToolLocate)) { locateNonce++ }
                        DetailMapTool(t(Str.MapToolLegend)) { noteOpen = true }
                    }
                }

                if (!mapExpanded) {
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .weight(0.62f)
                            .verticalScroll(rememberScrollState()),
                    ) {
                        // Header strip: carId + more info
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .background(Color.White)
                                .padding(horizontal = 16.dp, vertical = 14.dp),
                            horizontalArrangement = Arrangement.SpaceBetween,
                            verticalAlignment = Alignment.CenterVertically,
                        ) {
                            Text(
                                text = "${t(Str.VehicleTagCarId)}: ${vehicle.carId}",
                                fontWeight = FontWeight.Bold,
                                fontSize = 16.sp,
                                color = Color(0xFF242936),
                            )
                            Text(
                                text = "${t(Str.MoreInfo)} >",
                                color = Color(0xFF7C87B1),
                                fontSize = 14.sp,
                                modifier = Modifier.clickable(
                                    interactionSource = remember { MutableInteractionSource() },
                                    indication = null,
                                    onClick = { moreInfoOpen = true },
                                ),
                            )
                        }
                        HorizontalDivider(color = Color(0xFFE8E8E8))

                        // Info block matching legacy head
                        Column(
                            modifier = Modifier
                                .fillMaxWidth()
                                .background(Color.White)
                                .padding(horizontal = 16.dp, vertical = 12.dp),
                            verticalArrangement = Arrangement.spacedBy(12.dp),
                        ) {
                            Row(verticalAlignment = Alignment.CenterVertically) {
                                Text(
                                    "${t(Str.VehicleListColStatus)}:",
                                    color = Color(0xFF7C87B1),
                                    fontSize = 14.sp,
                                )
                                Box(modifier = Modifier.width(8.dp))
                                DetailStatusChips(app, vehicle)
                            }
                            InfoValueRow(t(Str.BelongStation), vehicle.forParkName.ifBlank { "-" })
                            InfoValueRow(t(Str.LastLocateTime), locateTime)
                            InfoValueRow(
                                label = t(Str.LastLocatePlace),
                                value = detailMap.address?.ifBlank { null } ?: "-",
                                valueColor = Color(0xFF1180F9),
                            )
                            InfoValueRow(
                                label = t(Str.VehicleTagLabel),
                                value = "--",
                                valueColor = Color(0xFF1180F9),
                                onClick = {
                                    app.vehicleTagFeature.setCarId(vehicle.carId)
                                    toolOverlay = DetailToolOverlay.VehicleTag
                                },
                            )
                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .clickable(
                                        interactionSource = remember { MutableInteractionSource() },
                                        indication = null,
                                        onClick = {
                                            scope.launch {
                                                app.vehicleDetailMapFeature.loadRealtimeTrack(vehicle)
                                                fitNonce++
                                            }
                                        },
                                    ),
                                horizontalArrangement = Arrangement.SpaceBetween,
                            ) {
                                Text(t(Str.RealtimeTrackMenu), color = Color(0xFF1180F9), fontSize = 14.sp)
                                Text(">", color = Color(0xFF1180F9), fontSize = 14.sp)
                            }
                        }

                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(8.dp)
                                .background(Color(0xFFF4F6FF)),
                        )

                        // Menus — align CarDetailActivity secondary pages
                        listOf(
                            t(Str.OrderInfoMenu) to { orderHistoryOpen = true },
                            t(Str.ScanLocateTrackMenu) to { scanLogOpen = true },
                            t(Str.LockRecordMenu) to { switchLockOpen = true },
                            t(Str.ChangeBatteryRecordMenu) to { changeBatteryHistoryOpen = true },
                            t(Str.MoveCarRecordMenu) to { moveCarHistoryOpen = true },
                        ).forEach { (title, onClick) ->
                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .background(Color.White)
                                    .clickable(
                                        interactionSource = remember { MutableInteractionSource() },
                                        indication = null,
                                        onClick = onClick,
                                    )
                                    .padding(horizontal = 16.dp, vertical = 16.dp),
                                horizontalArrangement = Arrangement.SpaceBetween,
                            ) {
                                Text(title, color = Color(0xFF242936), fontSize = 14.sp, fontWeight = FontWeight.Bold)
                                Text(">", color = Color(0xFFCCCCCC), fontSize = 16.sp)
                            }
                            HorizontalDivider(color = Color(0xFFCCCCCC), thickness = 0.5.dp)
                        }

                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(8.dp)
                                .background(Color(0xFFF4F6FF)),
                        )

                        DetailActionGrid(
                            permissions = permissions,
                            onRing = { runAction(VehicleAction.Ring) },
                            onChangeBattery = {
                                // 页内 overlay；勿再调 onOpen*，否则 MainShell 全屏打开会卸掉详情。
                                toolOverlay = DetailToolOverlay.ChangeBattery
                            },
                            onMoveCar = {
                                toolOverlay = DetailToolOverlay.MoveCar
                            },
                            onUnlock = { runAction(VehicleAction.Unlock) },
                            onLock = { runAction(VehicleAction.Lock) },
                            onFault = {
                                app.faultReportFeature.openSubmit()
                                app.faultReportFeature.setCarId(vehicle.carId)
                                toolOverlay = DetailToolOverlay.FaultReport
                            },
                            onVehicleTag = {
                                // 对齐 CarDetail → VehicleTagActivity：关标签页回到详情，不是退出 App。
                                app.vehicleTagFeature.setCarId(vehicle.carId)
                                toolOverlay = DetailToolOverlay.VehicleTag
                            },
                            onOpenBattery = { runAction(VehicleAction.OpenBatteryBox) },
                            onCloseBattery = { runAction(VehicleAction.CloseBatteryBox) },
                            onUnbindBattery = {
                                scope.launch {
                                    actionTip = OpsTipState.Waiting(t(Str.HandleVehicle))
                                    when (val r = app.vehicleFeature.unbindBatterySn(vehicle.carId)) {
                                        is OpsResult.Ok -> {
                                            actionTip = OpsTipState.Success(t(Str.UnbindOk))
                                            delay(1_000)
                                            actionTip = OpsTipState.Hidden
                                        }
                                        is OpsResult.Err -> {
                                            val msg = r.error.message.trim().ifEmpty { t(Str.ActionFailed, t(Str.UnbindBattery), "") }
                                            actionTip = OpsTipState.Error(msg)
                                            delay(2_000)
                                            actionTip = OpsTipState.Hidden
                                        }
                                    }
                                }
                            },
                            onOpenHelmet = { runAction(VehicleAction.OpenHelmetLock) },
                            onCloseHelmet = { runAction(VehicleAction.CloseHelmetLock) },
                            t = { key -> t(key) },
                        )
                        Box(modifier = Modifier.height(24.dp))
                    }
                }
            }

            if (noteOpen) {
                VehicleDetailNoteScreen(
                    app = app,
                    onClose = { noteOpen = false },
                    modifier = Modifier
                        .fillMaxSize()
                        .zIndex(10f),
                )
            }
            if (moreInfoOpen) {
                VehicleMoreInfoScreen(
                    app = app,
                    vehicle = vehicle,
                    address = detailMap.address,
                    onClose = { moreInfoOpen = false },
                    modifier = Modifier
                        .fillMaxSize()
                        .zIndex(10f),
                )
            }
            if (orderHistoryOpen) {
                VehicleOrderHistoryScreen(
                    app = app,
                    carId = vehicle.carId,
                    currentArea = homeState.currentArea,
                    mapReady = mapReady,
                    onClose = { orderHistoryOpen = false },
                    modifier = Modifier
                        .fillMaxSize()
                        .zIndex(10f),
                )
            }
            if (scanLogOpen) {
                VehicleScanLogScreen(
                    app = app,
                    carId = vehicle.carId,
                    mapReady = mapReady,
                    onClose = { scanLogOpen = false },
                    modifier = Modifier
                        .fillMaxSize()
                        .zIndex(10f),
                )
            }
            if (switchLockOpen) {
                VehicleSwitchLockLogScreen(
                    app = app,
                    carId = vehicle.carId,
                    onClose = { switchLockOpen = false },
                    modifier = Modifier
                        .fillMaxSize()
                        .zIndex(10f),
                )
            }
            if (changeBatteryHistoryOpen) {
                VehicleChangeBatteryHistoryScreen(
                    app = app,
                    carId = vehicle.carId,
                    serviceId = resolvedServiceId,
                    onClose = { changeBatteryHistoryOpen = false },
                    modifier = Modifier
                        .fillMaxSize()
                        .zIndex(10f),
                )
            }
            if (moveCarHistoryOpen) {
                VehicleMoveCarHistoryScreen(
                    app = app,
                    carId = vehicle.carId,
                    serviceId = resolvedServiceId,
                    onClose = { moveCarHistoryOpen = false },
                    modifier = Modifier
                        .fillMaxSize()
                        .zIndex(10f),
                )
            }

            when (toolOverlay) {
                DetailToolOverlay.ChangeBattery -> {
                    Box(
                        modifier = Modifier
                            .fillMaxSize()
                            .zIndex(20f),
                    ) {
                        FieldChangeBatteryScreen(
                            app = app,
                            onClose = { toolOverlay = null },
                            onHelp = { toast(t(Str.FieldChangeBatteryHelp)) },
                            initialCarId = vehicle.carId,
                            autoOpenBox = true,
                            scanPreview = { mod, torchOn, enabled, onCode ->
                                scanner.Preview(mod, torchOn, enabled, onCode)
                            },
                        )
                    }
                }
                DetailToolOverlay.MoveCar -> {
                    Box(
                        modifier = Modifier
                            .fillMaxSize()
                            .zIndex(20f),
                    ) {
                        FieldMoveCarScreen(
                            app = app,
                            currentArea = homeState.currentArea,
                            onClose = { toolOverlay = null },
                            onBleSearch = { toast(t(Str.FeatureComingSoon)) },
                            initialCarId = vehicle.carId,
                            scanPreview = { mod, torchOn, enabled, onCode ->
                                scanner.Preview(mod, torchOn, enabled, onCode)
                            },
                        )
                    }
                }
                DetailToolOverlay.FaultReport -> {
                    Box(
                        modifier = Modifier
                            .fillMaxSize()
                            .zIndex(20f),
                    ) {
                        FaultReportScreen(
                            app = app,
                            onClose = { toolOverlay = null },
                        )
                    }
                }
                DetailToolOverlay.VehicleTag -> {
                    Box(
                        modifier = Modifier
                            .fillMaxSize()
                            .zIndex(20f),
                    ) {
                        VehicleTagScreen(
                            app = app,
                            onClose = { toolOverlay = null },
                            scanPreview = { modifier, torchOn, enabled, onCode ->
                                scanner.Preview(modifier, torchOn, enabled, onCode)
                            },
                        )
                    }
                }
                null -> Unit
            }

            OpsTipDialogHost(state = actionTip)
        }
    }
}

@Composable
private fun InfoValueRow(
    label: String,
    value: String,
    valueColor: Color = Color(0xFF242936),
    onClick: (() -> Unit)? = null,
) {
    Row(
        modifier = Modifier
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
            ),
    ) {
        Text(label + ":", color = Color(0xFF7C87B1), fontSize = 14.sp)
        Text(
            text = value,
            color = valueColor,
            fontSize = 14.sp,
            fontWeight = FontWeight.Bold,
            modifier = Modifier
                .weight(1f)
                .padding(start = 8.dp),
        )
    }
}

@Composable
private fun DetailMapTool(label: String, onClick: () -> Unit) {
    Column(
        modifier = Modifier
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            )
            .padding(horizontal = 8.dp, vertical = 6.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text(label, fontSize = 11.sp, color = Color(0xFF555555))
    }
}

@Composable
private fun DetailStatusChips(app: OpsApp, vehicle: Vehicle) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val chips = buildList {
        vehicle.alarmStates.forEach { code ->
            when (code) {
                VehicleAlarmStates.OFFLINE -> add(ChipTone.Error to t(Str.Offline))
                else -> Unit
            }
        }
        vehicle.operationStates.forEach { code ->
            when (code) {
                VehicleOperationStates.OFF -> add(ChipTone.Error to t(Str.OpsSoldOut))
                VehicleOperationStates.LOW_BATTERY -> add(ChipTone.Error to t(Str.FilterLowBattery))
                VehicleOperationStates.REPAIRING -> add(ChipTone.Error to t(Str.FilterRepairing))
                VehicleOperationStates.MOVING_CAR -> add(ChipTone.Warn to t(Str.FilterMoving))
                else -> Unit
            }
        }
        when (vehicle.ridingState) {
            VehicleRidingStates.RIDEABLE -> add(ChipTone.Ok to t(Str.FilterReady))
            VehicleRidingStates.RIDING -> add(ChipTone.Ok to t(Str.FilterRiding))
            VehicleRidingStates.TEMP_PARKING -> add(ChipTone.Warn to t(Str.FilterTempParking))
            VehicleRidingStates.RESERVE -> add(ChipTone.Warn to t(Str.FilterBooking))
            else -> if (vehicle.ridingLabel.isNotBlank()) add(ChipTone.Ok to vehicle.ridingLabel)
        }
        if (!vehicle.isOnline) add(ChipTone.Error to t(Str.Offline))
    }.distinctBy { it.second }.take(4)

    Row(horizontalArrangement = Arrangement.spacedBy(6.dp)) {
        chips.forEach { (tone, label) ->
            val (bg, fg) = when (tone) {
                ChipTone.Ok -> Color(0x1A1DBA4F) to Color(0xFF1DBA4F)
                ChipTone.Warn -> Color(0x1AFFAD00) to Color(0xFFFFAD00)
                ChipTone.Error -> Color(0x1AE02020) to Color(0xFFE02020)
            }
            Box(
                modifier = Modifier
                    .background(bg, RoundedCornerShape(4.dp))
                    .padding(horizontal = 8.dp, vertical = 2.dp),
            ) {
                Text(label, color = fg, fontSize = 12.sp)
            }
        }
    }
}

private enum class ChipTone { Ok, Warn, Error }

@Composable
private fun DetailActionGrid(
    permissions: OpsPermissions,
    onRing: () -> Unit,
    onChangeBattery: () -> Unit,
    onMoveCar: () -> Unit,
    onUnlock: () -> Unit,
    onLock: () -> Unit,
    onFault: () -> Unit,
    onVehicleTag: () -> Unit,
    onOpenBattery: () -> Unit,
    onCloseBattery: () -> Unit,
    onUnbindBattery: () -> Unit,
    onOpenHelmet: () -> Unit,
    onCloseHelmet: () -> Unit,
    t: (Str) -> String,
) {
    data class ActionItem(
        val label: String,
        val iconRes: Int,
        val onClick: () -> Unit,
        val visible: Boolean = true,
    )
    // Order / visibility matches legacy CarDetailActivity.getBottomDataSource().
    val items = listOf(
        ActionItem(t(Str.Ring), R.drawable.btn_vehicle_ring, onRing),
        ActionItem(
            t(Str.ChangeBatteryShort),
            R.drawable.btn_change_battery,
            onChangeBattery,
            permissions.showChangeBatteryTool,
        ),
        ActionItem(
            t(Str.MoveCarShort),
            R.drawable.btn_move_vehicle,
            onMoveCar,
            permissions.showMoveCarTool,
        ),
        ActionItem(
            t(Str.ScanUnlock),
            R.drawable.btn_unlock_vehicle,
            onUnlock,
            permissions.canScanUnlock,
        ),
        ActionItem(
            t(Str.ScanLock),
            R.drawable.btn_lock_vehicle,
            onLock,
            permissions.canScanUnlock,
        ),
        // Legacy always shows fault report (not permission-gated in bottom grid).
        ActionItem(t(Str.FaultReport), R.drawable.btn_breakdown_upload, onFault),
        ActionItem(
            t(Str.VehicleTagTool),
            R.drawable.vehicle_tag_ic,
            onVehicleTag,
            permissions.showVehicleTag,
        ),
        ActionItem(t(Str.OpenBatteryBox), R.drawable.btn_open_battery, onOpenBattery),
        ActionItem(t(Str.CloseBatteryBox), R.drawable.btn_close_battery, onCloseBattery),
        ActionItem(
            t(Str.UnbindBattery),
            R.drawable.btn_unbind_battery,
            onUnbindBattery,
            permissions.canBindBatterySn,
        ),
        ActionItem(t(Str.OpenHelmetLock), R.drawable.vehicle_helmet_open, onOpenHelmet),
        ActionItem(t(Str.CloseHelmetLock), R.drawable.vehicle_helmet_closr, onCloseHelmet),
    ).filter { it.visible }

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(Color.White)
            .padding(top = 4.dp, bottom = 12.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text("˅", color = Color(0xFFCCCCCC), fontSize = 14.sp)
        items.chunked(5).forEach { row ->
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceEvenly,
            ) {
                row.forEach { item ->
                    Column(
                        modifier = Modifier
                            .width(64.dp)
                            .clickable(
                                interactionSource = remember { MutableInteractionSource() },
                                indication = null,
                                onClick = item.onClick,
                            ),
                        horizontalAlignment = Alignment.CenterHorizontally,
                    ) {
                        Image(
                            painter = painterResource(item.iconRes),
                            contentDescription = item.label,
                            modifier = Modifier.size(48.dp),
                            contentScale = ContentScale.Fit,
                        )
                        Box(modifier = Modifier.height(4.dp))
                        Text(
                            text = item.label,
                            fontSize = 12.sp,
                            color = Color(0xFF242F57),
                            textAlign = TextAlign.Center,
                            maxLines = 1,
                            overflow = TextOverflow.Ellipsis,
                        )
                    }
                }
                repeat(5 - row.size) { Box(modifier = Modifier.width(64.dp)) }
            }
        }
    }
}

private fun formatLocateTimestamp(timestamp: Long?): String {
    val ms = timestamp?.takeIf { it > 0 } ?: return "-"
    val normalized = if (ms < 10_000_000_000L) ms * 1000L else ms
    return SimpleDateFormat("yyyy-MM-dd HH:mm:ss", Locale.getDefault()).format(Date(normalized))
}

/**
 * Legacy VehicleInfoModel.showNoOrderTime + FenceDataHelp bubble text.
 * Uses i18n duration units to avoid encoding corruption.
 */
private fun formatNoOrderDuration(app: OpsApp, vehicle: Vehicle): String? {
    if (vehicle.lockTimeMs < 0L) return app.i18n.t(Str.VehicleNoRidingHistory)
    val base = when (vehicle.ridingState) {
        2, 3 -> vehicle.unlockTimeMs
        else -> vehicle.lockTimeMs
    }
    if (base <= 0L) return null
    val start = if (base < 10_000_000_000L) base * 1000L else base
    val elapsed = System.currentTimeMillis() - start
    if (elapsed < 1_000L) return null
    var rem = elapsed
    val day = rem / (24 * 3600_000L)
    rem -= day * 24 * 3600_000L
    val hours = rem / 3600_000L
    rem -= hours * 3600_000L
    val minutes = rem / 60_000L
    rem -= minutes * 60_000L
    val seconds = rem / 1000L
    val duration = buildString {
        if (day > 0) append(app.i18n.t(Str.DurationDays, day.toString()))
        if (hours > 0) append(app.i18n.t(Str.DurationHours, hours.toString()))
        if (minutes > 0) append(app.i18n.t(Str.DurationMinutes, minutes.toString()))
        if (seconds > 0) append(app.i18n.t(Str.DurationSeconds, seconds.toString()))
    }
    if (duration.isBlank()) return null
    return app.i18n.t(Str.NoOrderForDuration, duration)
}
