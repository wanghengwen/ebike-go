package com.luopingtech.ebike.ops.ui.vehicle

import androidx.activity.compose.BackHandler
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
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateMapOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import android.content.Intent
import android.net.Uri
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.domain.model.MapPin
import com.luopingtech.ebike.ops.domain.model.MapPinIcon
import com.luopingtech.ebike.ops.ui.map.TencentMapView
import com.luopingtech.ebike.ops.ui.theme.OpsTheme

@Composable
private fun HistoryTopBar(title: String, onClose: () -> Unit) {
    BackHandler(onBack = onClose)
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
            text = title,
            color = Color.White,
            fontSize = 18.sp,
            fontWeight = FontWeight.Medium,
            modifier = Modifier.align(Alignment.Center),
        )
    }
}

@Composable
fun VehicleScanLogScreen(
    app: OpsApp,
    carId: String,
    mapReady: Boolean,
    onClose: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.vehicleHistoryFeature.state.collectAsState()
    val context = LocalContext.current
    var selectedIndex by remember { mutableIntStateOf(0) }
    var followNonce by remember { mutableIntStateOf(0) }
    val addressCache = remember { mutableStateMapOf<String, String>() }

    DisposableEffect(Unit) {
        onDispose { app.vehicleHistoryFeature.clear() }
    }
    LaunchedEffect(carId) {
        app.vehicleHistoryFeature.loadScanLogs(carId)
    }
    LaunchedEffect(state.scanLogs) {
        selectedIndex = 0
        followNonce++
    }

    val selected = state.scanLogs.getOrNull(selectedIndex)
    LaunchedEffect(selectedIndex, selected?.time, selected?.latitude, selected?.longitude, selected?.address) {
        val log = selected ?: return@LaunchedEffect
        if (log.address.isNotBlank()) {
            addressCache[addressCacheKey(log)] = log.address
            return@LaunchedEffect
        }
        val lat = log.latitude
        val lng = log.longitude
        if (lat == null || lng == null) return@LaunchedEffect
        val key = addressCacheKey(log)
        if (addressCache.containsKey(key)) return@LaunchedEffect
        val result = app.reverseGeocoder.addressOf(lat, lng)
        val text = result.getOrNull()?.takeIf { it.isNotBlank() && !it.equals("Ocean", ignoreCase = true) }
        addressCache[key] = text ?: "--"
    }

    // 对齐遗留：有 lat/lng 就打点（含 0,0）；无坐标则不画。
    val pin = selected?.let { log ->
        val lat = log.latitude
        val lng = log.longitude
        if (lat != null && lng != null) {
            MapPin(
                id = "scan-${log.time}-$selectedIndex",
                lat = lat,
                lng = lng,
                title = "",
                memberCount = 1,
                memberIds = listOf("scan-$selectedIndex"),
                icon = MapPinIcon.VehicleNormal,
            )
        } else null
    }

    Column(modifier = modifier.fillMaxSize().background(Color.White)) {
        HistoryTopBar(title = t(Str.ScanLocateTitle), onClose = onClose)
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(308.dp),
        ) {
            if (mapReady) {
                TencentMapView(
                    pins = listOfNotNull(pin),
                    selectedCarId = null,
                    onSelectCarId = {},
                    clusterOverview = false,
                    autoFitOnPins = false,
                    followNonce = followNonce,
                    followLat = pin?.lat,
                    followLng = pin?.lng,
                    showStatusOverlay = false,
                    modifier = Modifier.fillMaxSize(),
                )
            } else {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .background(Color(0xFFF4F6FF)),
                    contentAlignment = Alignment.Center,
                ) {
                    Text(t(Str.LoadingEllipsis), color = Color(0xFF7C87B1), fontSize = 13.sp)
                }
            }
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
                text = t(Str.ScanLocateTitle),
                color = Color(0xFF242936),
                fontSize = 14.sp,
                fontWeight = FontWeight.Bold,
            )
        }
        HorizontalDivider(color = Color(0xFFD9DCE6), thickness = 1.dp)

        when {
            state.loading && state.scanLogs.isEmpty() -> Text(
                t(Str.LoadingEllipsis),
                modifier = Modifier.padding(16.dp),
                color = Color(0xFF7C87B1),
            )
            state.errorMessage != null && state.scanLogs.isEmpty() -> Text(
                state.errorMessage.orEmpty(),
                modifier = Modifier.padding(16.dp),
                color = Color(0xFFE02020),
            )
            state.scanLogs.isEmpty() -> Text(
                t(Str.NoRecords),
                modifier = Modifier.padding(16.dp),
                color = Color(0xFF7C87B1),
            )
            else -> LazyColumn(modifier = Modifier.fillMaxSize()) {
                itemsIndexed(state.scanLogs) { index, item ->
                    val selectedBg = if (index == selectedIndex) Color(0xFFF4F6FF) else Color.White
                    val phoneText = item.phone.ifBlank { "--" }
                    val locateText = addressCache[addressCacheKey(item)]
                        ?: item.address.takeIf { it.isNotBlank() }
                        ?: "--"
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .background(selectedBg)
                            .clickable(
                                interactionSource = remember { MutableInteractionSource() },
                                indication = null,
                                onClick = {
                                    selectedIndex = index
                                    followNonce++
                                },
                            )
                            .padding(horizontal = 16.dp, vertical = 12.dp),
                        verticalArrangement = Arrangement.spacedBy(8.dp),
                    ) {
                        ScanLogRow(
                            label = t(Str.PhoneColonLabel),
                            value = phoneText,
                            valueBlue = true,
                            onValueClick = {
                                val raw = item.phone.trim()
                                if (raw.isNotBlank() && raw != "--") {
                                    runCatching {
                                        context.startActivity(
                                            Intent(Intent.ACTION_DIAL, Uri.parse("tel:$raw")),
                                        )
                                    }
                                }
                            },
                        )
                        ScanLogRow(
                            label = t(Str.OrderDurationColon),
                            value = item.time.ifBlank { "--" },
                        )
                        ScanLogRow(
                            label = t(Str.LocateColonLabel),
                            value = locateText,
                            valueBlue = true,
                        )
                    }
                    HorizontalDivider(color = Color(0xFFD9DCE6), thickness = 1.dp)
                }
            }
        }
    }
}

@Composable
private fun ScanLogRow(
    label: String,
    value: String,
    valueBlue: Boolean = false,
    onValueClick: (() -> Unit)? = null,
) {
    Row(modifier = Modifier.fillMaxWidth()) {
        Text(label, color = Color(0xFF242936), fontSize = 14.sp)
        Text(
            text = value,
            color = if (valueBlue) Color(0xFF295FCC) else Color(0xFF242936),
            fontSize = 14.sp,
            fontWeight = FontWeight.Bold,
            modifier = Modifier
                .padding(start = 8.dp)
                .then(
                    if (onValueClick != null) {
                        Modifier.clickable(
                            interactionSource = remember { MutableInteractionSource() },
                            indication = null,
                            onClick = onValueClick,
                        )
                    } else {
                        Modifier
                    },
                ),
        )
    }
}

private fun addressCacheKey(log: com.luopingtech.ebike.ops.data.vehicle.VehicleScanLogDto): String {
    if (log.time.isNotBlank()) return "t:${log.time}"
    return "ll:${log.latitude},${log.longitude}"
}

@Composable
fun VehicleSwitchLockLogScreen(
    app: OpsApp,
    carId: String,
    onClose: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.vehicleHistoryFeature.state.collectAsState()

    DisposableEffect(Unit) {
        onDispose { app.vehicleHistoryFeature.clear() }
    }
    LaunchedEffect(carId) {
        app.vehicleHistoryFeature.loadSwitchLocks(carId)
    }

    Column(modifier = modifier.fillMaxSize().background(Color.White)) {
        HistoryTopBar(title = t(Str.LockRecordMenu), onClose = onClose)
        when {
            state.loading -> Text(
                t(Str.LoadingEllipsis),
                modifier = Modifier.padding(16.dp),
                color = Color(0xFF7C87B1),
            )
            state.errorMessage != null && state.switchLocks.isEmpty() -> Text(
                state.errorMessage.orEmpty(),
                modifier = Modifier.padding(16.dp),
                color = Color(0xFFE02020),
            )
            state.switchLocks.isEmpty() -> Text(
                t(Str.NoRecords),
                modifier = Modifier.padding(16.dp),
                color = Color(0xFF7C87B1),
            )
            else -> LazyColumn(modifier = Modifier.fillMaxSize()) {
                itemsIndexed(state.switchLocks) { _, item ->
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 16.dp, vertical = 14.dp),
                        verticalArrangement = Arrangement.spacedBy(10.dp),
                    ) {
                        Row {
                            Text(
                                "${item.name.ifBlank { "--" }}: ",
                                color = Color(0xFF999999),
                                fontSize = 14.sp,
                            )
                            Text(
                                item.phone.ifBlank { "--" },
                                color = Color(0xFF1180F9),
                                fontSize = 14.sp,
                                fontWeight = FontWeight.Bold,
                            )
                        }
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Box(
                                modifier = Modifier
                                    .size(12.dp)
                                    .background(
                                        if (item.isLock()) Color(0xFFEF2F6E) else Color(0xFF04B78A),
                                        CircleShape,
                                    ),
                            )
                            Text(
                                text = item.time.ifBlank { "--" },
                                color = Color(0xFF242936),
                                fontSize = 14.sp,
                                fontWeight = FontWeight.Bold,
                                modifier = Modifier.padding(start = 4.dp),
                            )
                            Text(
                                text = if (item.isLock()) t(Str.ScanLock) else t(Str.ScanUnlock),
                                color = Color(0xFF242936),
                                fontSize = 14.sp,
                                fontWeight = FontWeight.Bold,
                                modifier = Modifier.padding(start = 12.dp),
                            )
                        }
                    }
                    HorizontalDivider(color = Color(0xFFCCCCCC), thickness = 0.5.dp)
                }
            }
        }
    }
}

@Composable
fun VehicleChangeBatteryHistoryScreen(
    app: OpsApp,
    carId: String,
    serviceId: String?,
    onClose: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.vehicleHistoryFeature.state.collectAsState()

    DisposableEffect(Unit) {
        onDispose { app.vehicleHistoryFeature.clear() }
    }
    LaunchedEffect(carId, serviceId) {
        app.vehicleHistoryFeature.loadChangeBattery(carId, serviceId)
    }

    Column(modifier = modifier.fillMaxSize().background(Color.White)) {
        HistoryTopBar(title = t(Str.ChangeBatteryRecordMenu), onClose = onClose)
        when {
            state.loading -> Text(
                t(Str.LoadingEllipsis),
                modifier = Modifier.padding(16.dp),
                color = Color(0xFF7C87B1),
            )
            state.errorMessage != null && state.changeBattery.isEmpty() -> Text(
                state.errorMessage.orEmpty(),
                modifier = Modifier.padding(16.dp),
                color = Color(0xFFE02020),
            )
            state.changeBattery.isEmpty() -> Text(
                t(Str.NoRecords),
                modifier = Modifier.padding(16.dp),
                color = Color(0xFF7C87B1),
            )
            else -> LazyColumn(modifier = Modifier.fillMaxSize()) {
                itemsIndexed(state.changeBattery) { _, item ->
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 16.dp, vertical = 12.dp),
                        verticalArrangement = Arrangement.spacedBy(6.dp),
                    ) {
                        HistoryKv(t(Str.WorkOrderOperator), item.opMan.ifBlank { "--" })
                        HistoryKv(t(Str.Phone), item.phone.ifBlank { "--" }, valueBlue = true)
                        HistoryKv(
                            t(Str.HistoryBattery),
                            "${item.restBatteryBefore ?: "--"}% → ${item.restBatteryAfter ?: "--"}%",
                        )
                        HistoryKv(t(Str.OpenBatteryBox), item.openBatBoxTime.ifBlank { "--" })
                        HistoryKv(t(Str.CloseBatteryBox), item.closeBatBoxTime.ifBlank { "--" })
                        HistoryKv(t(Str.AuditStateLabel), auditLabel(app, item.checkResult))
                    }
                    HorizontalDivider(color = Color(0xFFCCCCCC), thickness = 0.5.dp)
                }
            }
        }
    }
}

@Composable
fun VehicleMoveCarHistoryScreen(
    app: OpsApp,
    carId: String,
    serviceId: String?,
    onClose: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.vehicleHistoryFeature.state.collectAsState()

    DisposableEffect(Unit) {
        onDispose { app.vehicleHistoryFeature.clear() }
    }
    LaunchedEffect(carId, serviceId) {
        app.vehicleHistoryFeature.loadMoveCar(carId, serviceId)
    }

    Column(modifier = modifier.fillMaxSize().background(Color.White)) {
        HistoryTopBar(title = t(Str.MoveCarRecordMenu), onClose = onClose)
        when {
            state.loading -> Text(
                t(Str.LoadingEllipsis),
                modifier = Modifier.padding(16.dp),
                color = Color(0xFF7C87B1),
            )
            state.errorMessage != null && state.moveCar.isEmpty() -> Text(
                state.errorMessage.orEmpty(),
                modifier = Modifier.padding(16.dp),
                color = Color(0xFFE02020),
            )
            state.moveCar.isEmpty() -> Text(
                t(Str.NoRecords),
                modifier = Modifier.padding(16.dp),
                color = Color(0xFF7C87B1),
            )
            else -> LazyColumn(modifier = Modifier.fillMaxSize()) {
                itemsIndexed(state.moveCar) { _, item ->
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 16.dp, vertical = 12.dp),
                        verticalArrangement = Arrangement.spacedBy(6.dp),
                    ) {
                        HistoryKv(t(Str.WorkOrderOperator), item.operator.ifBlank { "--" })
                        HistoryKv(t(Str.Phone), item.phone.ifBlank { "--" }, valueBlue = true)
                        HistoryKv(t(Str.Start), item.startTime.ifBlank { "--" })
                        HistoryKv(t(Str.HistoryEnd), item.endTime.ifBlank { "--" })
                        HistoryKv(t(Str.HistoryDuration), item.lastTime.ifBlank { "--" })
                        HistoryKv(t(Str.AuditStateLabel), auditLabel(app, item.checkResult))
                    }
                    HorizontalDivider(color = Color(0xFFCCCCCC), thickness = 0.5.dp)
                }
            }
        }
    }
}

@Composable
private fun HistoryKv(label: String, value: String, valueBlue: Boolean = false) {
    Row(modifier = Modifier.fillMaxWidth()) {
        Text("$label:", color = Color(0xFF7C87B1), fontSize = 14.sp)
        Text(
            text = value,
            color = if (valueBlue) Color(0xFF1180F9) else Color(0xFF242936),
            fontSize = 14.sp,
            fontWeight = FontWeight.Bold,
            modifier = Modifier.padding(start = 8.dp),
        )
    }
}

private fun auditLabel(app: OpsApp, checkResult: Int?): String {
    return when (checkResult) {
        null -> "--"
        0 -> app.i18n.t(Str.AuditResultPending)
        1 -> app.i18n.t(Str.AuditReject)
        2, 3 -> app.i18n.t(Str.AuditPass)
        4 -> app.i18n.t(Str.AuditReject)
        else -> checkResult.toString()
    }
}
