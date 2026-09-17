package com.luopingtech.ebike.ops.ui.production

import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.FilterChip
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.DetectStepKind
import com.luopingtech.ebike.ops.domain.model.DetectStepStatus
import com.luopingtech.ebike.ops.domain.permission.OpsPermissions
import com.luopingtech.ebike.ops.feature.production.DetectChannelTab
import com.luopingtech.ebike.ops.feature.production.DetectSwitchKind
import com.luopingtech.ebike.ops.feature.production.ProductionFeature
import com.luopingtech.ebike.ops.feature.production.ProductionPage
import com.luopingtech.ebike.ops.feature.production.ShelfMode
import com.luopingtech.ebike.ops.ui.feedback.OpsTipDialogHost
import com.luopingtech.ebike.ops.ui.feedback.OpsTipState
import com.luopingtech.ebike.ops.ui.icons.OpsBackChevron
import com.luopingtech.ebike.ops.ui.icons.OpsIcon
import com.luopingtech.ebike.ops.ui.icons.painterResource
import com.luopingtech.ebike.ops.ui.navigation.OpsBackHandler
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import com.luopingtech.ebike.ops.ui.tools.LegacyScanPreviewBlock
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

/** Legacy activity_vehicle_detect / item_xitem / VehicleSearchView / SwitchView colors. */
private val DetectLabelGray = Color(0xFFACB3CE)
private val DetectValueBlack = Color(0xFF242936)
private val DetectOptionBlack = Color(0xFF040404)
private val DetectSplitBg = Color(0xFFF0F4FB)
private val DetectRowDivider = Color(0xFFE6E6E6)
private val DetectSearchStroke = Color(0xFFC0C7E1)
private val DetectPagerLine = Color(0xFFE1E7FF)
private val DetectSwitchOff = Color(0xFFE3E3E3)
private val DetectSwitchOffBorder = Color(0xFFBFBFBF)
private val DetectSearchBarWidth = 345.dp
private val DetectSwitchWidth = 51.dp
private val DetectSwitchHeight = 31.dp
/** Legacy color123366CD — light blue wash card. */
private val BindInputBg = Color(0x123366CD)
private val BindInputErrorStroke = Color(0xFFE02020)
private val BindUnbindDisabledText = Color(0xFF9FA7C7)
private val BindUnbindDisabledStroke = Color(0xFFDCDCDC)
private val BindHintGray = Color(0xFF999999)

@Composable
fun ProductionScreen(
    app: OpsApp,
    permissions: OpsPermissions,
    onClose: () -> Unit,
    onLocateOnMap: () -> Unit = {},
    scanPreview: @Composable (
        modifier: Modifier,
        torchOn: Boolean,
        enabled: Boolean,
        onCode: (String) -> Unit,
    ) -> Unit = { _, _, _, _ -> },
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.productionFeature.state.collectAsState()
    val primary = OpsTheme.colors.primary
    val isDetectChrome = state.page == ProductionPage.Detect || state.page == ProductionPage.Overload
    val isBindChrome = state.page == ProductionPage.Bind
    val isShelvesChrome = state.page == ProductionPage.Shelves

    if (isDetectChrome) {
        val onBack = {
            when (state.page) {
                ProductionPage.Overload -> app.productionFeature.backToDetect()
                else -> {
                    app.productionFeature.clear()
                    onClose()
                }
            }
        }
        OpsBackHandler(onBack = onBack)
        Column(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.White),
        ) {
            DetectToolbar(
                title = when (state.page) {
                    ProductionPage.Overload -> t(Str.DetectOverloadCheck)
                    else -> t(Str.ProductionDetect)
                },
                primary = primary,
                onBack = onBack,
            )
            when (state.page) {
                ProductionPage.Detect -> DetectPane(
                    app = app,
                    onLocateOnMap = onLocateOnMap,
                    modifier = Modifier.weight(1f),
                )
                ProductionPage.Overload -> OverloadPane(
                    app = app,
                    modifier = Modifier.weight(1f),
                )
                else -> Unit
            }
            state.errorMessage?.let {
                Text(
                    it,
                    color = MaterialTheme.colorScheme.error,
                    modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp),
                )
            }
        }
        return
    }

    if (isBindChrome) {
        val onBack = {
            app.productionFeature.clear()
            onClose()
        }
        OpsBackHandler(onBack = onBack)
        Column(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.White),
        ) {
            DetectToolbar(
                title = t(Str.ProductionBind),
                primary = primary,
                onBack = onBack,
            )
            BindPane(
                app = app,
                scanPreview = scanPreview,
                modifier = Modifier.weight(1f),
            )
            state.errorMessage?.let {
                Text(
                    it,
                    color = MaterialTheme.colorScheme.error,
                    modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp),
                )
            }
        }
        return
    }

    if (isShelvesChrome) {
        val onBack = {
            app.productionFeature.clear()
            onClose()
        }
        OpsBackHandler(onBack = onBack)
        Column(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.White),
        ) {
            DetectToolbar(
                title = t(Str.ProductionShelves),
                primary = primary,
                onBack = onBack,
            )
            ShelvesPane(
                app = app,
                scanPreview = scanPreview,
                modifier = Modifier.weight(1f),
            )
            state.errorMessage?.let {
                Text(
                    it,
                    color = MaterialTheme.colorScheme.error,
                    modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp),
                )
            }
        }
        return
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
            .padding(horizontal = 16.dp, vertical = 16.dp),
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(bottom = 8.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            val title = when (state.page) {
                else -> t(Str.ProductionField)
            }
            Text(
                title,
                color = DetectValueBlack,
                fontSize = 18.sp,
                fontWeight = FontWeight.Bold,
            )
            TextButton(onClick = {
                app.productionFeature.clear()
                onClose()
            }) {
                Text(t(Str.Close), color = OpsTheme.colors.primary)
            }
        }

        when (state.page) {
            ProductionPage.Hub -> {
                Column(
                    modifier = Modifier
                        .fillMaxSize()
                        .verticalScroll(rememberScrollState()),
                    verticalArrangement = Arrangement.spacedBy(10.dp),
                ) {
                    Text(
                        text = t(Str.ProductionHint),
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                    state.message?.let {
                        Text(it, color = MaterialTheme.colorScheme.primary)
                    }
                    if (permissions.showProductionDetect) {
                        Button(
                            onClick = { app.productionFeature.openDetect() },
                            modifier = Modifier.fillMaxWidth(),
                        ) { Text(t(Str.ProductionDetect)) }
                    }
                    if (permissions.showProductionBind) {
                        Button(
                            onClick = { app.productionFeature.openBind() },
                            modifier = Modifier.fillMaxWidth(),
                        ) { Text(t(Str.ProductionBind)) }
                    }
                    if (permissions.showProductionShelves) {
                        Button(
                            onClick = { app.productionFeature.openShelves(ShelfMode.PutOn) },
                            modifier = Modifier.fillMaxWidth(),
                        ) { Text(t(Str.ShelfPutOn)) }
                        Button(
                            onClick = { app.productionFeature.openShelves(ShelfMode.PullOff) },
                            modifier = Modifier.fillMaxWidth(),
                        ) { Text(t(Str.ShelfPullOff)) }
                    }
                }
            }
            else -> Unit
        }

        state.errorMessage?.let {
            Spacer(modifier = Modifier.height(8.dp))
            Text(it, color = MaterialTheme.colorScheme.error)
        }
    }
}

@Composable
private fun DetectToolbar(
    title: String,
    primary: Color,
    onBack: () -> Unit,
) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .background(primary)
            .statusBarsPadding()
            .height(48.dp)
            .padding(horizontal = 8.dp),
    ) {
        OpsBackChevron(
            onClick = onBack,
            modifier = Modifier.align(Alignment.CenterStart),
        )
        Text(
            text = title,
            color = Color.White,
            fontSize = 17.sp,
            fontWeight = FontWeight.SemiBold,
            modifier = Modifier.align(Alignment.Center),
        )
    }
}

@Composable
private fun DetectPane(
    app: OpsApp,
    onLocateOnMap: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.productionFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    val vehicle = state.detectVehicle
    val showOverload = state.detectChannel == DetectChannelTab.Network &&
        state.hasOverloadDevice == true
    val primary = OpsTheme.colors.primary
    val queryEnabled = !state.loading &&
        app.productionFeature.resolveDetectInput(state.searchInput) != null

    fun query(raw: String? = null) {
        scope.launch {
            val hosts = app.authFeature.state.value.runtimeConfig?.qrHosts.orEmpty()
            app.productionFeature.queryDetectVehicle(raw = raw, qrHosts = hosts)
        }
    }

    Column(
        modifier = modifier
            .fillMaxSize()
            .background(Color.White),
    ) {
        BoxWithConstraints(
            modifier = Modifier
                .fillMaxWidth()
                .padding(top = 16.dp),
            contentAlignment = Alignment.Center,
        ) {
            val barWidth = minOf(DetectSearchBarWidth, maxWidth - 30.dp)
            DetectSearchBar(
                value = state.searchInput,
                onValueChange = { raw ->
                    app.productionFeature.setSearchInput(raw.filter { it.isDigit() }.take(20))
                },
                hint = t(Str.DetectHintCarIdImei),
                queryLabel = if (state.loading) t(Str.Querying) else t(Str.DetectQuery),
                queryEnabled = queryEnabled,
                onScan = {
                    scope.launch {
                        when (val scan = app.codeScanner.scanOnce()) {
                            is OpsResult.Ok -> query(scan.value)
                            is OpsResult.Err -> Unit
                        }
                    }
                },
                onQuery = { query() },
                primary = primary,
                modifier = Modifier.width(barWidth),
            )
        }

        Column(
            modifier = Modifier
                .weight(1f)
                .fillMaxWidth()
                .verticalScroll(rememberScrollState()),
        ) {
            if (vehicle != null) {
                DetectInfoRow(
                    label = t(Str.VehicleId),
                    value = vehicle.carId.ifBlank { "-" },
                    topPadding = 24.dp,
                )
                DetectInfoRow(label = t(Str.OrderQueryImeiLabel), value = vehicle.imei.ifBlank { "-" })
                DetectInfoRow(label = t(Str.DetectSignal), value = vehicle.signalLabel)
                DetectInfoRow(label = t(Str.DetectVoltageValue), value = vehicle.voltageLabel)
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(top = 24.dp)
                        .height(8.dp)
                        .background(DetectSplitBg),
                )
            }

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 10.dp)
                    .height(48.dp),
                horizontalArrangement = Arrangement.Center,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                DetectTabLabel(
                    text = t(Str.DetectTabNetwork),
                    selected = state.detectChannel == DetectChannelTab.Network,
                    primary = primary,
                    onClick = { app.productionFeature.setDetectChannel(DetectChannelTab.Network) },
                )
                Spacer(modifier = Modifier.width(36.dp))
                DetectTabLabel(
                    text = t(Str.DetectTabBluetooth),
                    selected = state.detectChannel == DetectChannelTab.Bluetooth,
                    primary = primary,
                    onClick = { app.productionFeature.setDetectChannel(DetectChannelTab.Bluetooth) },
                )
            }

            if (state.detectChannel == DetectChannelTab.Bluetooth && !state.bleAvailable) {
                Text(
                    t(Str.DetectBleUnavailableHint),
                    modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp),
                    fontSize = 13.sp,
                    color = OpsTheme.colors.textTertiary,
                )
            }

            state.message?.takeIf { it.isNotBlank() }?.let {
                Text(
                    it,
                    modifier = Modifier.padding(horizontal = 16.dp, vertical = 6.dp),
                    fontSize = 13.sp,
                    color = primary,
                )
            }

            if (vehicle == null) {
                Text(
                    t(Str.DetectNoVehicleInfo),
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(vertical = 64.dp),
                    fontSize = 16.sp,
                    color = DetectLabelGray,
                    textAlign = TextAlign.Center,
                )
            } else {
                Spacer(modifier = Modifier.height(10.dp))
                state.detectSwitches.forEach { sw ->
                    DetectOptionRow(
                        icon = detectSwitchIcon(sw.kind),
                        label = sw.label,
                    ) {
                        DetectLegacySwitch(
                            checked = sw.on,
                            enabled = !sw.busy && !state.loading,
                            primary = primary,
                            onCheckedChange = { checked ->
                                scope.launch {
                                    app.productionFeature.toggleDetectSwitch(sw.kind, checked)
                                }
                            },
                        )
                    }
                }
                DetectOptionRow(
                    icon = OpsIcon.DetectLocation,
                    label = t(Str.DetectVehicleLocation),
                    onClick = {
                        scope.launch {
                            when (val result = app.productionFeature.locateVehicleOnMap()) {
                                is OpsResult.Ok -> {
                                    app.vehicleFeature.upsertAndSelect(result.value)
                                    app.productionFeature.clear()
                                    onLocateOnMap()
                                }
                                is OpsResult.Err -> Unit
                            }
                        }
                    },
                ) {
                    Image(
                        painter = painterResource(OpsIcon.DetectArrow),
                        contentDescription = null,
                        modifier = Modifier.size(width = 12.dp, height = 20.dp),
                    )
                }
                if (showOverload) {
                    DetectOptionRow(
                        icon = OpsIcon.DetectOverload,
                        label = t(Str.DetectOverloadCheck),
                        onClick = {
                            scope.launch { app.productionFeature.openOverloadCheck() }
                        },
                    ) {
                        Image(
                            painter = painterResource(OpsIcon.DetectArrow),
                            contentDescription = null,
                            modifier = Modifier.size(width = 12.dp, height = 20.dp),
                        )
                    }
                }
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(0.5.dp)
                        .background(DetectPagerLine),
                )
            }
        }

        if (vehicle != null) {
            HorizontalDivider(color = DetectPagerLine, thickness = 0.5.dp)
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(104.dp)
                    .background(Color.White),
                horizontalArrangement = Arrangement.SpaceEvenly,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                state.detectSteps.forEach { step ->
                    val running = step.status == DetectStepStatus.Running || state.loading
                    DetectFooterAction(
                        icon = detectActionIcon(step.kind),
                        label = step.label,
                        enabled = !running,
                        onClick = { scope.launch { app.productionFeature.runDetectStep(step.kind) } },
                    )
                }
            }
        }
    }
}

@Composable
private fun DetectSearchBar(
    value: String,
    onValueChange: (String) -> Unit,
    hint: String,
    queryLabel: String,
    queryEnabled: Boolean,
    onScan: () -> Unit,
    onQuery: () -> Unit,
    primary: Color,
    modifier: Modifier = Modifier,
) {
    Row(
        modifier = modifier
            .height(40.dp)
            .border(1.dp, DetectSearchStroke, RoundedCornerShape(5.dp)),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Box(
            modifier = Modifier
                .width(45.dp)
                .fillMaxSize()
                .clickable(onClick = onScan),
            contentAlignment = Alignment.Center,
        ) {
            Image(
                painter = painterResource(OpsIcon.DetectScan),
                contentDescription = null,
                modifier = Modifier.size(22.dp),
            )
        }
        BasicTextField(
            value = value,
            onValueChange = onValueChange,
            singleLine = true,
            textStyle = TextStyle(color = DetectValueBlack, fontSize = 16.sp),
            cursorBrush = SolidColor(primary),
            modifier = Modifier
                .weight(1f)
                .padding(end = 10.dp),
            decorationBox = { inner ->
                Box(contentAlignment = Alignment.CenterStart) {
                    if (value.isEmpty()) {
                        Text(hint, color = DetectLabelGray, fontSize = 15.sp)
                    }
                    inner()
                }
            },
        )
        Box(
            modifier = Modifier
                .width(70.dp)
                .fillMaxSize()
                .background(
                    if (queryEnabled) primary else Color.Transparent,
                    RoundedCornerShape(topEnd = 5.dp, bottomEnd = 5.dp),
                )
                .clickable(enabled = queryEnabled, onClick = onQuery),
            contentAlignment = Alignment.Center,
        ) {
            Text(
                text = queryLabel,
                fontSize = 16.sp,
                color = if (queryEnabled) Color.White else OpsTheme.colors.disabled,
            )
        }
    }
}

@Composable
private fun DetectInfoRow(
    label: String,
    value: String,
    topPadding: Dp = 14.dp,
) {
    Row(
        modifier = Modifier.padding(start = 24.dp, top = topPadding),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            text = "$label: ",
            fontSize = 14.sp,
            color = DetectLabelGray,
        )
        Text(
            text = value,
            fontSize = 16.sp,
            color = DetectValueBlack,
        )
    }
}

@Composable
private fun DetectTabLabel(
    text: String,
    selected: Boolean,
    primary: Color,
    onClick: () -> Unit,
) {
    Column(
        modifier = Modifier
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            )
            .padding(horizontal = 12.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text(
            text = text,
            fontSize = 19.sp,
            fontWeight = if (selected) FontWeight.Bold else FontWeight.Normal,
            color = if (selected) primary else DetectValueBlack,
        )
        Spacer(modifier = Modifier.height(6.dp))
        Box(
            modifier = Modifier
                .width(22.dp)
                .height(3.dp)
                .background(
                    if (selected) primary else Color.Transparent,
                    RoundedCornerShape(1.5.dp),
                ),
        )
    }
}

/** 对齐遗留 SwitchView：51×31，开=主题色，关=#E3E3E3，白圆钮。 */
@Composable
private fun DetectLegacySwitch(
    checked: Boolean,
    enabled: Boolean,
    primary: Color,
    onCheckedChange: (Boolean) -> Unit,
) {
    val progress by animateFloatAsState(
        targetValue = if (checked) 1f else 0f,
        animationSpec = tween(durationMillis = 180),
    )
    val thumbSize = 27.dp
    val inset = 2.dp
    val travel = DetectSwitchWidth - inset * 2 - thumbSize
    Box(
        modifier = Modifier
            .size(width = DetectSwitchWidth, height = DetectSwitchHeight)
            .alpha(if (enabled) 1f else 0.45f)
            .clip(RoundedCornerShape(percent = 50))
            .background(if (checked) primary else DetectSwitchOff)
            .clickable(
                enabled = enabled,
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = { onCheckedChange(!checked) },
            ),
    ) {
        Box(
            modifier = Modifier
                .align(Alignment.CenterStart)
                .padding(start = inset)
                .offset(x = travel * progress)
                .size(thumbSize)
                .border(
                    width = if (checked) 0.dp else 1.dp,
                    color = DetectSwitchOffBorder,
                    shape = CircleShape,
                )
                .background(Color.White, CircleShape),
        )
    }
}

@Composable
private fun DetectOptionRow(
    icon: OpsIcon,
    label: String,
    onClick: (() -> Unit)? = null,
    trailing: @Composable () -> Unit,
) {
    Column {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .height(47.dp)
                .then(if (onClick != null) Modifier.clickable(onClick = onClick) else Modifier)
                .padding(horizontal = 12.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Image(
                painter = painterResource(icon),
                contentDescription = null,
                modifier = Modifier.size(24.dp),
            )
            Spacer(modifier = Modifier.width(11.dp))
            Text(
                text = label,
                modifier = Modifier.weight(1f),
                fontSize = 16.sp,
                fontWeight = FontWeight.Bold,
                color = DetectOptionBlack,
            )
            trailing()
        }
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .padding(start = 12.dp)
                .height(1.dp)
                .background(DetectRowDivider),
        )
    }
}

@Composable
private fun DetectFooterAction(
    icon: OpsIcon,
    label: String,
    enabled: Boolean,
    onClick: () -> Unit,
) {
    Column(
        modifier = Modifier
            .clickable(
                enabled = enabled,
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            )
            .padding(horizontal = 12.dp, vertical = 8.dp)
            .alpha(if (enabled) 1f else 0.4f),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Image(
            painter = painterResource(icon),
            contentDescription = null,
            modifier = Modifier.size(46.dp),
        )
        Spacer(modifier = Modifier.height(4.dp))
        Text(
            text = label,
            fontSize = 14.sp,
            fontWeight = FontWeight.Bold,
            color = DetectOptionBlack,
        )
    }
}

private fun detectSwitchIcon(kind: DetectSwitchKind): OpsIcon = when (kind) {
    DetectSwitchKind.Acc -> OpsIcon.DetectSwitchAcc
    DetectSwitchKind.Defend -> OpsIcon.DetectSwitchDefend
    DetectSwitchKind.Battery -> OpsIcon.DetectSwitchBattery
    DetectSwitchKind.Helmet -> OpsIcon.DetectSwitchHelmet
    DetectSwitchKind.Wheel -> OpsIcon.DetectSwitchWheel
}

private fun detectActionIcon(kind: DetectStepKind): OpsIcon = when (kind) {
    DetectStepKind.Ring -> OpsIcon.DetectRing
    DetectStepKind.Refresh -> OpsIcon.DetectRefresh
    DetectStepKind.Reboot -> OpsIcon.DetectReboot
}

@Composable
private fun OverloadPane(
    app: OpsApp,
    modifier: Modifier = Modifier,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.productionFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    val contacts = state.overloadContacts
    Column(
        modifier = modifier
            .fillMaxSize()
            .background(Color.White)
            .verticalScroll(rememberScrollState())
            .padding(horizontal = 16.dp, vertical = 20.dp),
        verticalArrangement = Arrangement.spacedBy(14.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        if (state.overloadRunning && state.overloadCountdown > 0) {
            Text(
                text = "${state.overloadCountdown}s",
                fontSize = 32.sp,
                fontWeight = FontWeight.Bold,
                color = DetectValueBlack,
            )
        }
        Text(
            t(Str.DetectOverloadTips),
            fontSize = 14.sp,
            color = DetectLabelGray,
            textAlign = TextAlign.Center,
        )
        Row(horizontalArrangement = Arrangement.spacedBy(24.dp)) {
            OverloadContactChip(label = t(Str.DetectOverloadFront), on = contacts.frontOn)
            OverloadContactChip(label = t(Str.DetectOverloadCenter), on = contacts.centerOn)
            OverloadContactChip(label = t(Str.DetectOverloadBack), on = contacts.backOn)
        }
        Button(
            onClick = {
                scope.launch { app.productionFeature.startOverloadCheck() }
            },
            enabled = !state.loading,
            modifier = Modifier
                .fillMaxWidth()
                .height(44.dp),
        ) {
            Text(
                if (state.overloadRunning) t(Str.DetectOverloadRestart) else t(Str.DetectOverloadStart),
            )
        }
        state.message?.let {
            Text(it, color = OpsTheme.colors.primary, fontSize = 13.sp)
        }
    }
}

@Composable
private fun OverloadContactChip(label: String, on: Boolean) {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Text(
            text = if (on) "●" else "○",
            style = MaterialTheme.typography.headlineSmall,
            color = if (on) {
                MaterialTheme.colorScheme.primary
            } else {
                MaterialTheme.colorScheme.onSurfaceVariant
            },
        )
        Text(label, style = MaterialTheme.typography.labelMedium)
    }
}

@Composable
private fun BindPane(
    app: OpsApp,
    scanPreview: @Composable (
        modifier: Modifier,
        torchOn: Boolean,
        enabled: Boolean,
        onCode: (String) -> Unit,
    ) -> Unit,
    modifier: Modifier = Modifier,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.productionFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    val primary = OpsTheme.colors.primary
    var torchOn by remember { mutableStateOf(false) }
    var lastCode by remember { mutableStateOf<String?>(null) }
    var showDeviceError by remember { mutableStateOf(false) }
    var showHelmetError by remember { mutableStateOf(false) }
    val actionsEnabled = app.productionFeature.canBindOrUnbind() && !state.loading
    val qrHosts = app.authFeature.state.value.runtimeConfig?.qrHosts.orEmpty()

    LaunchedEffect(state.carId) {
        app.productionFeature.onBindCarIdChanged()
    }
    LaunchedEffect(state.imei) {
        app.productionFeature.onBindImeiChanged()
    }

    Column(
        modifier = modifier
            .fillMaxSize()
            .background(Color.White)
            .verticalScroll(rememberScrollState()),
    ) {
        LegacyScanPreviewBlock(
            torchOn = torchOn,
            torchLabel = t(Str.Torch),
            scanEnabled = !state.loading,
            onTorchChange = { torchOn = it },
            scanPreview = scanPreview,
            onCode = { raw ->
                if (state.loading) return@LegacyScanPreviewBlock
                if (raw == lastCode) return@LegacyScanPreviewBlock
                lastCode = raw
                scope.launch {
                    app.productionFeature.applyBindScan(raw, qrHosts)
                }
            },
        )

        Spacer(modifier = Modifier.height(12.dp))

        BindInputCard(
            title = t(Str.BindCarLabel),
            value = state.carId,
            hint = t(Str.BindScanCarHint),
            showClear = state.carId.isNotEmpty(),
            showError = false,
            errorText = t(Str.BindInputWrongNumber),
            trailingTitle = null,
            onTrailingTitleClick = null,
            onClear = { app.productionFeature.setCarId("") },
            onValueChange = { raw ->
                app.productionFeature.setCarId(raw.take(ProductionFeature.MAX_CAR_ID_LEN))
            },
            onFocusLost = { /* carId error suppressed in legacy */ },
            primary = primary,
        )
        Spacer(modifier = Modifier.height(12.dp))
        BindInputCard(
            title = t(Str.BindDeviceLabel),
            value = state.imei,
            hint = t(Str.BindScanDeviceHint),
            showClear = state.imei.isNotEmpty(),
            showError = showDeviceError,
            errorText = t(Str.BindInputWrongNumber),
            trailingTitle = null,
            onTrailingTitleClick = null,
            onClear = {
                showDeviceError = false
                app.productionFeature.setImei("")
            },
            onValueChange = { raw ->
                app.productionFeature.setImei(raw.filter { it.isDigit() }.take(ProductionFeature.BIND_IMEI_EXACT_LEN))
            },
            onFocusLost = {
                val imei = app.productionFeature.state.value.imei.trim()
                showDeviceError = imei.isNotEmpty() && imei.length < 14
            },
            primary = primary,
        )
        Spacer(modifier = Modifier.height(12.dp))
        BindInputCard(
            title = t(Str.BindHelmetLabel),
            value = state.helmet,
            hint = t(Str.BindScanHelmetHint),
            showClear = state.helmet.isNotEmpty(),
            showError = showHelmetError,
            errorText = t(Str.BindInputWrongNumber),
            trailingTitle = t(Str.BindBluetoothIdentify),
            onTrailingTitleClick = {
                // Legacy BluetoothIdentificationActivity — not ported yet.
                app.productionFeature.setTransientMessage(app.i18n.t(Str.FeatureComingSoon))
            },
            onClear = {
                showHelmetError = false
                app.productionFeature.setHelmet("")
            },
            onValueChange = { raw ->
                app.productionFeature.setHelmet(
                    raw.filter { it.isLetterOrDigit() }.take(ProductionFeature.BIND_HELMET_MAX_LEN),
                )
            },
            onFocusLost = {
                val helmet = app.productionFeature.state.value.helmet.trim()
                showHelmetError = helmet.isNotEmpty() && helmet.length < 11
            },
            primary = primary,
        )

        state.message?.takeIf { it.isNotBlank() }?.let {
            Text(
                it,
                color = primary,
                fontSize = 13.sp,
                modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp),
            )
        }

        Spacer(modifier = Modifier.height(20.dp))
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp)
                .height(48.dp)
                .clip(RoundedCornerShape(8.dp))
                .background(if (actionsEnabled) primary else primary.copy(alpha = 0.3f))
                .clickable(enabled = actionsEnabled) {
                    scope.launch { app.productionFeature.bindCenter() }
                },
            contentAlignment = Alignment.Center,
        ) {
            Text(t(Str.BindAction), color = Color.White, fontSize = 16.sp)
        }
        Spacer(modifier = Modifier.height(10.dp))
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp)
                .height(48.dp)
                .clip(RoundedCornerShape(8.dp))
                .border(
                    1.dp,
                    if (actionsEnabled) primary else BindUnbindDisabledStroke,
                    RoundedCornerShape(8.dp),
                )
                .background(Color.White)
                .clickable(enabled = actionsEnabled) {
                    scope.launch { app.productionFeature.unbindCenter() }
                },
            contentAlignment = Alignment.Center,
        ) {
            Text(
                t(Str.UnbindAction),
                color = if (actionsEnabled) primary else BindUnbindDisabledText,
                fontSize = 16.sp,
            )
        }
        Spacer(modifier = Modifier.height(40.dp))
    }
}

@Composable
private fun BindInputCard(
    title: String,
    value: String,
    hint: String,
    showClear: Boolean,
    showError: Boolean,
    errorText: String,
    trailingTitle: String?,
    onTrailingTitleClick: (() -> Unit)?,
    onClear: () -> Unit,
    onValueChange: (String) -> Unit,
    onFocusLost: () -> Unit,
    primary: Color,
) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp)
            .height(85.dp)
            .border(
                width = if (showError) 1.dp else 0.dp,
                color = if (showError) BindInputErrorStroke else Color.Transparent,
                shape = RoundedCornerShape(8.dp),
            )
            .background(BindInputBg, RoundedCornerShape(8.dp)),
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(start = 16.dp, top = 12.dp, end = 12.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(title, color = DetectValueBlack, fontSize = 14.sp)
            if (trailingTitle != null && onTrailingTitleClick != null) {
                Spacer(modifier = Modifier.width(10.dp))
                Text(
                    text = trailingTitle,
                    color = primary,
                    fontSize = 12.sp,
                    modifier = Modifier
                        .border(1.dp, primary, RoundedCornerShape(4.dp))
                        .clickable(onClick = onTrailingTitleClick)
                        .padding(horizontal = 8.dp, vertical = 2.dp),
                )
            }
            if (showError) {
                Spacer(modifier = Modifier.width(8.dp))
                Text(errorText, color = BindInputErrorStroke, fontSize = 14.sp)
            }
        }
        BasicTextField(
            value = value,
            onValueChange = onValueChange,
            singleLine = true,
            textStyle = TextStyle(color = DetectValueBlack, fontSize = 16.sp),
            cursorBrush = SolidColor(primary),
            modifier = Modifier
                .align(Alignment.BottomStart)
                .padding(start = 16.dp, end = 48.dp, bottom = 14.dp)
                .fillMaxWidth()
                .onFocusChanged { focus ->
                    if (!focus.isFocused) onFocusLost()
                },
            decorationBox = { inner ->
                Box {
                    if (value.isEmpty()) {
                        Text(hint, color = BindHintGray, fontSize = 15.sp)
                    }
                    inner()
                }
            },
        )
        if (showClear) {
            Image(
                painter = painterResource(OpsIcon.BindClear),
                contentDescription = null,
                modifier = Modifier
                    .align(Alignment.CenterEnd)
                    .padding(end = 20.dp)
                    .size(20.dp)
                    .clickable(onClick = onClear),
            )
        }
    }
}

@Composable
private fun ShelvesPane(
    app: OpsApp,
    scanPreview: @Composable (
        modifier: Modifier,
        torchOn: Boolean,
        enabled: Boolean,
        onCode: (String) -> Unit,
    ) -> Unit,
    modifier: Modifier = Modifier,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.productionFeature.state.collectAsState()
    val homeState by app.homeFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    val primary = OpsTheme.colors.primary
    val putOn = state.shelfMode == ShelfMode.PutOn
    var torchOn by remember { mutableStateOf(false) }
    var lastCode by remember { mutableStateOf<String?>(null) }
    var showAreaPicker by remember { mutableStateOf(false) }
    var tip by remember { mutableStateOf<OpsTipState>(OpsTipState.Hidden) }
    val addEnabled = app.productionFeature.canAddShelfCar()
    val submitEnabled = app.productionFeature.hasShelfSelection() && !state.loading
    val allSelected = app.productionFeature.isShelfAllSelected()
    val qrHosts = app.authFeature.state.value.runtimeConfig?.qrHosts.orEmpty()
    val areas = homeState.serviceAreas

    suspend fun showShelfFeedbackTip() {
        val s = app.productionFeature.state.value
        val success = s.message?.takeIf { it.isNotBlank() }
        val error = s.errorMessage?.takeIf { it.isNotBlank() }
        when {
            success != null -> {
                tip = OpsTipState.Success(success)
                delay(1_000)
            }
            error != null -> {
                tip = OpsTipState.Error(error)
                delay(2_000)
            }
            else -> return
        }
        tip = OpsTipState.Hidden
    }

    Box(modifier = modifier.fillMaxSize()) {
        Column(modifier = Modifier.fillMaxSize().background(Color.White)) {
        // Segmented 上架 / 下架 under theme strip (legacy llTopTab).
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .background(primary)
                .padding(bottom = 10.dp),
            contentAlignment = Alignment.Center,
        ) {
            Row(
                modifier = Modifier
                    .width(208.dp)
                    .height(36.dp)
                    .border(1.dp, Color.White, RoundedCornerShape(8.dp))
                    .padding(2.dp),
            ) {
                ShelfTabChip(
                    text = t(Str.ShelfPutOn),
                    selected = putOn,
                    primary = primary,
                    modifier = Modifier.weight(1f),
                    onClick = {
                        lastCode = null
                        app.productionFeature.setShelfMode(ShelfMode.PutOn)
                    },
                )
                ShelfTabChip(
                    text = t(Str.ShelfPullOff),
                    selected = !putOn,
                    primary = primary,
                    modifier = Modifier.weight(1f),
                    onClick = {
                        lastCode = null
                        app.productionFeature.setShelfMode(ShelfMode.PullOff)
                    },
                )
            }
        }

        LegacyScanPreviewBlock(
            torchOn = torchOn,
            torchLabel = t(Str.Torch),
            scanEnabled = !state.loading,
            onTorchChange = { torchOn = it },
            scanPreview = scanPreview,
            onCode = { raw ->
                if (state.loading) return@LegacyScanPreviewBlock
                if (raw == lastCode) return@LegacyScanPreviewBlock
                scope.launch {
                    val ok = app.productionFeature.applyShelvesScan(raw, qrHosts)
                    // Only debounce after success so failed scans can be retried immediately.
                    if (ok) lastCode = raw
                    showShelfFeedbackTip()
                }
            },
        )

        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp, vertical = 10.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            BasicTextField(
                value = state.carId,
                onValueChange = { app.productionFeature.setCarId(it.take(ProductionFeature.MAX_CAR_ID_LEN)) },
                singleLine = true,
                textStyle = TextStyle(color = DetectValueBlack, fontSize = 16.sp),
                cursorBrush = SolidColor(primary),
                modifier = Modifier
                    .weight(1f)
                    .height(40.dp)
                    .background(BindInputBg, RoundedCornerShape(4.dp))
                    .padding(horizontal = 12.dp, vertical = 10.dp),
                decorationBox = { inner ->
                    Box {
                        if (state.carId.isEmpty()) {
                            Text(t(Str.ShelfEnterCarHint), color = BindHintGray, fontSize = 15.sp)
                        }
                        inner()
                    }
                },
            )
            Spacer(modifier = Modifier.width(12.dp))
            Box(
                modifier = Modifier
                    .width(82.dp)
                    .height(40.dp)
                    .clip(RoundedCornerShape(4.dp))
                    .background(if (addEnabled) primary else primary.copy(alpha = 0.3f))
                    .clickable(enabled = addEnabled) {
                        scope.launch {
                            app.productionFeature.addShelfCar()
                            showShelfFeedbackTip()
                        }
                    },
                contentAlignment = Alignment.Center,
            ) {
                Text(t(Str.ShelfAdd), color = Color.White, fontSize = 16.sp)
            }
        }

        // Header
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .height(40.dp)
                .padding(horizontal = 16.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(
                t(Str.ShelfCarNumberCol),
                modifier = Modifier.width(90.dp),
                fontSize = 16.sp,
                color = DetectValueBlack,
            )
            if (!putOn) {
                Text(
                    t(Str.ServiceAreaLabel),
                    modifier = Modifier.weight(2f),
                    fontSize = 16.sp,
                    color = DetectValueBlack,
                )
            }
            Text(
                t(Str.ShelfOperateCol),
                modifier = Modifier.weight(1f),
                fontSize = 16.sp,
                color = DetectValueBlack,
                textAlign = TextAlign.Center,
            )
            androidx.compose.material3.Checkbox(
                checked = allSelected,
                onCheckedChange = { app.productionFeature.setShelfSelectAll(it) },
                enabled = state.shelfQueue.isNotEmpty(),
                colors = androidx.compose.material3.CheckboxDefaults.colors(checkedColor = primary),
            )
        }

        Column(
            modifier = Modifier
                .weight(1f)
                .fillMaxWidth()
                .verticalScroll(rememberScrollState()),
        ) {
            state.shelfQueue.forEach { item ->
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(45.dp)
                        .padding(horizontal = 16.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text(
                        item.carId,
                        modifier = Modifier.width(90.dp),
                        fontSize = 16.sp,
                        color = DetectValueBlack,
                        maxLines = 1,
                    )
                    if (!putOn) {
                        Text(
                            item.serviceName.ifBlank { "-" },
                            modifier = Modifier.weight(2f),
                            fontSize = 16.sp,
                            color = DetectValueBlack,
                            maxLines = 1,
                        )
                    }
                    Text(
                        t(Str.ShelfDelete),
                        modifier = Modifier
                            .weight(1f)
                            .clickable { app.productionFeature.removeShelfItem(item.carId) },
                        fontSize = 16.sp,
                        color = BindInputErrorStroke,
                        textAlign = TextAlign.Center,
                    )
                    androidx.compose.material3.Checkbox(
                        checked = item.selected,
                        onCheckedChange = {
                            app.productionFeature.setShelfItemSelected(item.carId, it)
                        },
                        colors = androidx.compose.material3.CheckboxDefaults.colors(checkedColor = primary),
                    )
                }
            }
        }

        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(60.dp),
            contentAlignment = Alignment.Center,
        ) {
            Box(
                modifier = Modifier
                    .width(300.dp)
                    .height(50.dp)
                    .clip(RoundedCornerShape(4.dp))
                    .background(if (submitEnabled) primary else primary.copy(alpha = 0.3f))
                    .clickable(enabled = submitEnabled) {
                        if (putOn) {
                            showAreaPicker = true
                        } else {
                            scope.launch {
                                app.productionFeature.submitShelfQueue()
                                showShelfFeedbackTip()
                            }
                        }
                    },
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    if (putOn) t(Str.ShelfSelectPutArea) else t(Str.ShelfConfirm),
                    color = Color.White,
                    fontSize = 16.sp,
                )
            }
        }
        }

        OpsTipDialogHost(state = tip)
    }

    if (showAreaPicker) {
        androidx.compose.material3.AlertDialog(
            onDismissRequest = { showAreaPicker = false },
            title = { Text(t(Str.ShelfSelectPutArea)) },
            text = {
                Column(modifier = Modifier.verticalScroll(rememberScrollState())) {
                    if (areas.isEmpty()) {
                        Text(t(Str.NoServiceArea), color = DetectLabelGray)
                    } else {
                        areas.forEach { area ->
                            Text(
                                text = area.name.ifBlank { area.id },
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .clickable {
                                        showAreaPicker = false
                                        scope.launch {
                                            app.productionFeature.submitShelfQueue(serviceId = area.id)
                                            showShelfFeedbackTip()
                                        }
                                    }
                                    .padding(vertical = 12.dp),
                                fontSize = 16.sp,
                                color = DetectValueBlack,
                            )
                        }
                    }
                }
            },
            confirmButton = {
                TextButton(onClick = { showAreaPicker = false }) {
                    Text(t(Str.Close), color = primary)
                }
            },
        )
    }
}

@Composable
private fun ShelfTabChip(
    text: String,
    selected: Boolean,
    primary: Color,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Box(
        modifier = modifier
            .fillMaxSize()
            .clip(RoundedCornerShape(6.dp))
            .background(if (selected) Color.White else Color.Transparent)
            .clickable(onClick = onClick),
        contentAlignment = Alignment.Center,
    ) {
        Text(
            text = text,
            fontSize = 15.sp,
            fontWeight = if (selected) FontWeight.Bold else FontWeight.Normal,
            color = if (selected) primary else Color.White,
        )
    }
}
