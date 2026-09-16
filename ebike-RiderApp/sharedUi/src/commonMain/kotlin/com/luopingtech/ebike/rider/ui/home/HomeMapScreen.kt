package com.luopingtech.ebike.rider.ui.home

import androidx.compose.foundation.Canvas
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
import androidx.compose.foundation.layout.safeDrawingPadding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.rider.RiderApp
import com.luopingtech.ebike.rider.core.config.H5ScreenKind
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.domain.ble.BleFrame
import com.luopingtech.ebike.rider.domain.home.HomeNavItem
import com.luopingtech.ebike.rider.domain.home.HomeNavJump
import com.luopingtech.ebike.rider.domain.home.HomeNavTarget
import com.luopingtech.ebike.rider.domain.model.MapPin
import com.luopingtech.ebike.rider.domain.scan.ScanCodeParser
import com.luopingtech.ebike.rider.domain.scan.ScanTarget
import com.luopingtech.ebike.rider.platform.GeoPoint
import com.luopingtech.ebike.rider.platform.MapProviderKind
import com.luopingtech.ebike.rider.ui.feedback.LocalRiderToast
import com.luopingtech.ebike.rider.ui.map.RiderMapSpec
import com.luopingtech.ebike.rider.ui.map.RiderMapView
import com.luopingtech.ebike.rider.ui.media.RiderNetworkImage
import com.luopingtech.ebike.rider.ui.permission.LocalRiderBlePermissionGate
import com.luopingtech.ebike.rider.ui.permission.LocalRiderLocationPermissionGate
import com.luopingtech.ebike.rider.ui.theme.RiderTheme
import kotlinx.coroutines.launch

/** 扫码 / 头像：优先租户 branding，否则非凡皮肤。 */
private const val FALLBACK_SCAN_ICON =
    "https://feifan-frontend.oss-cn-shenzhen.aliyuncs.com/miniapp/project/feifan/icon_scan.png"
private const val FALLBACK_AVATAR =
    "https://feifan-frontend.oss-cn-shenzhen.aliyuncs.com/miniapp/project/feifan/default_avatar.png"

private val HomePageBg = Color(0xFFF4F5F7)
private val ScanBtnText = Color(0xFF1E4A38)
private val ShortcutLabel = Color(0xFF666666)

/**
 * 首页 —— 对齐 UniApp `pages/home/home.vue`：
 * 白顶栏（头像 + 居中品牌）→ 圆角地图 + 快捷入口 → 固定底栏「立即用车」。
 *
 * 快捷入口优先用 `getHomeNavByServiceId`；现网常返回 `data:[]`，此时用带 CDN 图标的默认四宫格。
 */
@Composable
fun HomeMapScreen(
    app: RiderApp,
    onOpenH5: (H5ScreenKind?, String?) -> Unit = { _, _ -> },
    onStartRide: ((carId: String, imei: String) -> Unit)? = null,
    onOpenScan: (() -> Unit)? = null,
    modifier: Modifier = Modifier,
) {
    val scope = rememberCoroutineScope()
    val toast = LocalRiderToast.current
    val permissionGate = LocalRiderLocationPermissionGate.current
    val blePermissionGate = LocalRiderBlePermissionGate.current
    var pins by remember { mutableStateOf<List<MapPin>>(emptyList()) }
    var shortcuts by remember(app.config.branding) {
        mutableStateOf(HomeNavJump.defaultItems(app.config.branding))
    }
    val profileEntryUrl = app.config.branding.logo.ifBlank {
        app.config.branding.defaultAvatar.ifBlank { FALLBACK_AVATAR }
    }
    val scanIconUrl = app.config.branding.iconScan.ifBlank { FALLBACK_SCAN_ICON }
    var selectedCarId by remember { mutableStateOf<String?>(null) }
    var selectedImei by remember { mutableStateOf<String?>(null) }
    var unlocking by remember { mutableStateOf(false) }
    var loading by remember { mutableStateOf(true) }
    var followNonce by remember { mutableIntStateOf(0) }
    var followLat by remember { mutableStateOf<Double?>(null) }
    var followLng by remember { mutableStateOf<Double?>(null) }
    var fitNonce by remember { mutableIntStateOf(0) }
    var myLocation by remember { mutableStateOf<GeoPoint?>(null) }

    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val brand = RiderTheme.colors.primary

    fun loadHomeNav(serviceId: String) {
        scope.launch {
            when (val result = app.homeNavRemote.load(serviceId)) {
                is RiderResult.Ok -> {
                    if (result.value.isNotEmpty()) {
                        shortcuts = result.value
                    }
                }
                is RiderResult.Err -> Unit
            }
        }
    }

    fun refreshNearby(at: GeoPoint?) {
        scope.launch {
            loading = true
            val lat = at?.latitude ?: myLocation?.latitude ?: 28.22
            val lng = at?.longitude ?: myLocation?.longitude ?: 112.94
            val point = GeoPoint(lat, lng)
            var sid = app.rideSessionStore.serviceAreaId
            when (val area = app.fenceRemote.serviceAreaIdAt(point)) {
                is RiderResult.Ok -> {
                    if (area.value.isNotBlank()) {
                        app.rideSessionStore.serviceAreaId = area.value
                        sid = area.value
                    }
                }
                is RiderResult.Err -> Unit
            }
            loadHomeNav(sid)
            when (val result = app.nearbyVehicleRemote.nearby(lat, lng, sid)) {
                is RiderResult.Ok -> {
                    pins = result.value
                    fitNonce++
                }
                is RiderResult.Err -> Unit
            }
            loading = false
        }
    }

    fun centerOn(at: GeoPoint) {
        myLocation = at
        followLat = at.latitude
        followLng = at.longitude
        followNonce++
    }

    fun requestLocateAndRefresh(showErrorToast: Boolean = true) {
        permissionGate.ensure { msg ->
            if (msg != null) {
                if (showErrorToast) toast(msg)
                // 无权限时不要用长沙默认点刷地图，避免区域错乱
                loadHomeNav(app.rideSessionStore.serviceAreaId)
                loading = false
                return@ensure
            }
            scope.launch {
                when (val loc = app.locationTracker.currentLocation()) {
                    is RiderResult.Ok -> {
                        centerOn(loc.value)
                        refreshNearby(loc.value)
                    }
                    is RiderResult.Err -> {
                        if (showErrorToast) toast(t(Str.LocationUnavailable))
                        loading = false
                    }
                }
            }
        }
    }

    fun unlockSelected() {
        val carId = selectedCarId ?: return
        val imei = selectedImei ?: BleFrame.demoImeiForCarId(carId)
        if (onStartRide != null) {
            onStartRide(carId, selectedImei.orEmpty())
            return
        }
        blePermissionGate.ensure { msg ->
            if (msg != null) {
                toast(msg)
                return@ensure
            }
            scope.launch {
                unlocking = true
                when (val result = app.bleSession.unlock(imei, mute = true)) {
                    is RiderResult.Ok -> toast(t(Str.BleUnlockOk))
                    is RiderResult.Err -> toast(t(Str.BleUnlockFailed, result.error.message))
                }
                unlocking = false
            }
        }
    }

    fun openShortcut(item: HomeNavItem) {
        when (val target = HomeNavJump.resolve(item)) {
            is HomeNavTarget.Kind -> onOpenH5(target.kind, null)
            is HomeNavTarget.Hash -> onOpenH5(null, target.route)
            HomeNavTarget.Unsupported -> toast(t(Str.HomeNavUnsupported))
        }
    }

    fun startScan() {
        if (onOpenScan != null) {
            onOpenScan()
            return
        }
        scope.launch {
            when (val scan = app.codeScanner.scanOnce()) {
                is RiderResult.Ok -> {
                    when (val target = ScanCodeParser.parse(scan.value)) {
                        is ScanTarget.CarId -> {
                            selectedCarId = target.value
                            selectedImei = null
                            if (onStartRide != null) {
                                onStartRide(target.value, "")
                            } else {
                                toast(t(Str.ScannedVehicle, target.value))
                            }
                        }
                        is ScanTarget.Imei -> {
                            selectedImei = target.value
                            selectedCarId = target.value
                            if (onStartRide != null) {
                                onStartRide(target.value, target.value)
                            } else {
                                toast(t(Str.ScannedVehicle, target.value))
                            }
                        }
                        null -> toast(t(Str.ScanUnrecognized))
                    }
                }
                is RiderResult.Err -> {
                    if (scan.error.code != "SCAN_CANCELLED") {
                        toast(scan.error.message)
                    }
                }
            }
        }
    }

    LaunchedEffect(Unit) {
        // 对齐 UniApp home.onShow：已登录且服务端仍在骑 → 抬到骑行页
        scope.launch { app.ridingFeature.resumeActiveRideIfNeeded() }
        loading = true
        loadHomeNav(app.rideSessionStore.serviceAreaId)
        requestLocateAndRefresh(showErrorToast = false)
    }

    val providerLabel = when (app.mapCapability.kind) {
        MapProviderKind.TENCENT -> t(Str.TencentMap)
        MapProviderKind.SIMULATOR -> t(Str.SimulatorMap)
        else -> app.mapCapability.kind.name
    }

    Column(
        modifier = modifier
            .fillMaxSize()
            .background(Color.White)
            .safeDrawingPadding(),
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(60.dp)
                .background(Color.White),
        ) {
            // 原 52dp 圆 / 29dp 图 → 80%：42dp / 23dp；图优先租户 logo
            Box(
                modifier = Modifier
                    .align(Alignment.CenterStart)
                    .padding(start = 10.dp)
                    .size(42.dp)
                    .shadow(3.dp, CircleShape)
                    .clip(CircleShape)
                    .background(Color.White)
                    .clickable { onOpenH5(H5ScreenKind.Profile, null) },
                contentAlignment = Alignment.Center,
            ) {
                RiderNetworkImage(
                    url = profileEntryUrl,
                    contentDescription = t(Str.ProfileEntry),
                    modifier = Modifier.size(23.dp),
                )
            }
            Text(
                text = app.config.app.displayName.ifBlank { t(Str.AppName) },
                modifier = Modifier.align(Alignment.Center),
                style = MaterialTheme.typography.titleMedium,
                color = Color(0xFF333333),
                fontWeight = FontWeight.SemiBold,
                fontSize = 18.sp,
            )
            if (loading) {
                CircularProgressIndicator(
                    modifier = Modifier
                        .align(Alignment.CenterEnd)
                        .padding(end = 12.dp)
                        .size(18.dp),
                    strokeWidth = 2.dp,
                )
            }
        }

        Column(
            modifier = Modifier
                .weight(1f)
                .fillMaxWidth()
                .background(HomePageBg),
        ) {
            Box(
                modifier = Modifier
                    .weight(1f)
                    .fillMaxWidth()
                    .padding(10.dp)
                    .clip(RoundedCornerShape(15.dp))
                    .background(HomePageBg),
            ) {
                RiderMapView(
                    spec = RiderMapSpec(
                        pins = pins,
                        selectedCarId = selectedCarId,
                        providerLabel = providerLabel,
                        onSelectCarId = {
                            selectedCarId = it
                            selectedImei = null
                        },
                        clusterOverview = true,
                        fitNonce = fitNonce,
                        followNonce = followNonce,
                        followLat = followLat,
                        followLng = followLng,
                    ),
                    modifier = Modifier.fillMaxSize(),
                )

                selectedCarId?.let { carId ->
                    val pin = pins.firstOrNull { it.id == carId || it.memberIds.contains(carId) }
                    Surface(
                        modifier = Modifier
                            .align(Alignment.TopCenter)
                            .padding(10.dp)
                            .fillMaxWidth(),
                        tonalElevation = 2.dp,
                        shape = MaterialTheme.shapes.medium,
                    ) {
                        Column(Modifier.padding(12.dp)) {
                            Text(
                                text = t(Str.ScannedVehicle, carId),
                                style = MaterialTheme.typography.titleSmall,
                            )
                            pin?.let {
                                Text(
                                    text = t(Str.VehicleBattery, it.restBattery),
                                    style = MaterialTheme.typography.bodySmall,
                                    color = RiderTheme.colors.textSecondary,
                                )
                            }
                            Button(
                                onClick = { unlockSelected() },
                                enabled = !unlocking,
                                modifier = Modifier.padding(top = 8.dp),
                            ) {
                                Text(if (onStartRide != null) t(Str.RideUnlock) else t(Str.BleUnlock))
                            }
                        }
                    }
                }

                // 圆形定位钮：十字准星图标（对齐 UniApp mapCfg.sideGetLocation）
                Box(
                    modifier = Modifier
                        .align(Alignment.BottomEnd)
                        .padding(12.dp)
                        .size(44.dp)
                        .shadow(4.dp, CircleShape)
                        .clip(CircleShape)
                        .background(Color.White)
                        .clickable { requestLocateAndRefresh(showErrorToast = true) },
                    contentAlignment = Alignment.Center,
                ) {
                    LocateCrosshairIcon(
                        color = Color(0xFF333333),
                        modifier = Modifier.size(22.dp),
                    )
                }
            }

            if (shortcuts.isNotEmpty()) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .background(Color.White)
                        .padding(horizontal = 28.dp, vertical = 14.dp),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    shortcuts.take(5).forEach { item ->
                        HomeShortcutCell(
                            item = item,
                            onClick = { openShortcut(item) },
                            modifier = Modifier.weight(1f),
                        )
                    }
                }
            }
        }

        Surface(
            modifier = Modifier.fillMaxWidth(),
            color = Color.White,
            shadowElevation = 0.dp,
            shape = RoundedCornerShape(topStart = 16.dp, topEnd = 16.dp),
        ) {
            Button(
                onClick = { startScan() },
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 12.dp, vertical = 16.dp)
                    .height(48.dp),
                shape = RoundedCornerShape(24.dp),
                colors = ButtonDefaults.buttonColors(
                    containerColor = brand,
                    contentColor = ScanBtnText,
                ),
            ) {
                RiderNetworkImage(
                    url = scanIconUrl,
                    contentDescription = null,
                    modifier = Modifier.size(20.dp),
                )
                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    text = t(Str.UseBikeNow),
                    fontSize = 17.sp,
                    fontWeight = FontWeight.SemiBold,
                    color = ScanBtnText,
                )
            }
        }
    }
}

@Composable
private fun HomeShortcutCell(
    item: HomeNavItem,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Column(
        modifier = modifier
            .clickable(onClick = onClick)
            .padding(horizontal = 4.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        RiderNetworkImage(
            url = item.iconUrl,
            contentDescription = item.name,
            modifier = Modifier.size(53.dp),
        )
        Spacer(modifier = Modifier.height(8.dp))
        Text(
            text = item.name,
            fontSize = 12.sp,
            color = ShortcutLabel,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
            textAlign = TextAlign.Center,
        )
    }
}

/** 地图「定位到我」准星图标。 */
@Composable
private fun LocateCrosshairIcon(
    color: Color,
    modifier: Modifier = Modifier,
) {
    Canvas(modifier = modifier) {
        val stroke = Stroke(width = size.minDimension * 0.09f, cap = StrokeCap.Round)
        val cx = size.width / 2f
        val cy = size.height / 2f
        val r = size.minDimension * 0.32f
        val arm = size.minDimension * 0.18f
        drawCircle(color = color, radius = r, style = stroke)
        drawCircle(color = color, radius = size.minDimension * 0.08f)
        drawLine(color, Offset(cx, cy - r - arm), Offset(cx, cy - r * 0.55f), stroke.width, StrokeCap.Round)
        drawLine(color, Offset(cx, cy + r * 0.55f), Offset(cx, cy + r + arm), stroke.width, StrokeCap.Round)
        drawLine(color, Offset(cx - r - arm, cy), Offset(cx - r * 0.55f, cy), stroke.width, StrokeCap.Round)
        drawLine(color, Offset(cx + r * 0.55f, cy), Offset(cx + r + arm, cy), stroke.width, StrokeCap.Round)
    }
}
