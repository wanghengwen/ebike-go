package com.luopingtech.ebike.ops.ui.movecar

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.LocalContentColor
import androidx.compose.material3.LocalTextStyle
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clipToBounds
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.FreeMoveCar
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.scan.ScanCodeParser
import com.luopingtech.ebike.ops.domain.scan.ScanTarget
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import com.luopingtech.ebike.ops.ui.tools.CheckboxMark
import com.luopingtech.ebike.ops.ui.tools.ScanCornerBrackets
import kotlinx.coroutines.launch

private enum class FieldMoveTab { Unlock, Lock }

/**
 * 工作台现场挪车（自主挪车 UI）。相机预览由宿主注入。
 */
@Composable
fun FieldMoveCarScreen(
    app: OpsApp,
    currentArea: ServiceArea?,
    onClose: () -> Unit,
    onBleSearch: () -> Unit = {},
    /** Legacy MoveCarActivity.start(carId): preload then addCar. */
    initialCarId: String? = null,
    scanPreview: @Composable (
        modifier: Modifier,
        torchOn: Boolean,
        enabled: Boolean,
        onCode: (String) -> Unit,
    ) -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.freeMoveCarFeature.state.collectAsState()
    val colors = OpsTheme.colors
    val scope = rememberCoroutineScope()
    var torchOn by remember { mutableStateOf(false) }
    var scanEnabled by remember { mutableStateOf(true) }
    var lastCode by remember { mutableStateOf<String?>(null) }
    var tab by remember { mutableStateOf(FieldMoveTab.Unlock) }
    var actionHint by remember { mutableStateOf<String?>(null) }

    LaunchedEffect(currentArea?.id, initialCarId) {
        // Warm last-known GPS so move_car requests carry latitude/longitude (legacy ParamsUtil).
        app.locationTracker.currentLocation()
        app.freeMoveCarFeature.load(currentArea)
        val seed = initialCarId?.trim().orEmpty()
        if (seed.isNotBlank()) {
            app.freeMoveCarFeature.setCarInput(seed)
            app.freeMoveCarFeature.addByCarId(seed, currentArea)
        }
    }

    fun carIdFromRaw(raw: String): String? {
        val hosts = app.authFeature.state.value.runtimeConfig?.qrHosts.orEmpty()
        return when (val target = ScanCodeParser.parse(raw, hosts)) {
            is ScanTarget.CarId -> target.value
            // Legacy DecodeUtil.decodeCarId on this page does not accept bare IMEI as carId.
            is ScanTarget.Imei -> null
            null -> {
                val trimmed = raw.trim()
                // Plain numeric car id typed/scanned without URL wrapper.
                trimmed.takeIf { it.length in 1..10 && it.all { ch -> ch.isDigit() } }
            }
        }
    }

    fun handleUnlock(carId: String) {
        scope.launch {
            val ok = app.freeMoveCarFeature.addByCarId(carId, currentArea)
            if (!ok) lastCode = null
        }
    }

    fun handleLock(carId: String) {
        scope.launch {
            // Legacy addToCloseCarList: only finish if already on open list; else「车子未开锁」.
            when (val result = app.freeMoveCarFeature.finishOneCar(carId)) {
                is OpsResult.Ok -> {
                    actionHint = null
                    lastCode = null
                }
                is OpsResult.Err -> {
                    actionHint = result.error.message
                    lastCode = null
                }
            }
        }
    }

    fun onScannedOrSubmit(raw: String) {
        val carId = carIdFromRaw(raw) ?: run {
            lastCode = null
            actionHint = t(Str.InvalidVehicleId)
            return
        }
        app.freeMoveCarFeature.setCarInput(carId)
        when (tab) {
            FieldMoveTab.Unlock -> handleUnlock(carId)
            FieldMoveTab.Lock -> handleLock(carId)
        }
    }

    com.luopingtech.ebike.ops.ui.navigation.OpsBackHandler(onBack = onClose)

    Column(modifier = Modifier.fillMaxSize().background(Color.White)) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .background(colors.primary)
                .statusBarsPadding()
                .padding(horizontal = 8.dp, vertical = 10.dp),
        ) {
            Text(
                text = "‹ ${t(Str.Back)}",
                color = colors.onPrimary,
                fontSize = 16.sp,
                modifier = Modifier
                    .align(Alignment.CenterStart)
                    .clickable(onClick = onClose)
                    .padding(4.dp),
            )
            Text(
                text = t(Str.MoveCarTool),
                color = colors.onPrimary,
                fontSize = 18.sp,
                fontWeight = FontWeight.Medium,
                modifier = Modifier.align(Alignment.Center),
            )
        }

        Box(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color(0xFFFFF4E0))
                .clickable(onClick = onBleSearch)
                .padding(vertical = 8.dp),
            contentAlignment = Alignment.Center,
        ) {
            Text(t(Str.FreeMoveBleSearch), color = Color(0xFFE67E22), fontSize = 13.sp)
        }

        val scanBg = Color(0xFF242936)
        // 预览区：只放相机，裁剪越界画面；操作文案一律放预览外。
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(220.dp)
                .background(scanBg),
            contentAlignment = Alignment.Center,
        ) {
            Box(
                modifier = Modifier
                    .fillMaxWidth(0.55f)
                    .aspectRatio(1f)
                    .clipToBounds(),
            ) {
                scanPreview(
                    Modifier.fillMaxSize(),
                    torchOn,
                    scanEnabled && !state.loading,
                ) { raw ->
                    if (!scanEnabled || state.loading) return@scanPreview
                    if (raw == lastCode) return@scanPreview
                    lastCode = raw
                    scanEnabled = false
                    onScannedOrSubmit(raw)
                    scanEnabled = true
                }
                ScanCornerBrackets(color = Color.White)
            }
        }

        // 不透明操作条：手电筒 / 推车模式，与预览分区。
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .background(scanBg)
                .padding(vertical = 12.dp),
            horizontalArrangement = Arrangement.spacedBy(28.dp, Alignment.CenterHorizontally),
        ) {
            Row(
                verticalAlignment = Alignment.CenterVertically,
                modifier = Modifier.clickable { torchOn = !torchOn },
            ) {
                CheckboxMark(checked = torchOn)
                Spacer(modifier = Modifier.width(6.dp))
                Text(t(Str.Torch), color = Color.White, fontSize = 14.sp)
            }
            Row(
                verticalAlignment = Alignment.CenterVertically,
                modifier = Modifier.clickable {
                    app.freeMoveCarFeature.setPushMode(!state.pushMode)
                },
            ) {
                CheckboxMark(checked = state.pushMode)
                Spacer(modifier = Modifier.width(6.dp))
                Text(t(Str.PushCarMode), color = Color.White, fontSize = 14.sp)
            }
        }

        Row(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color.White)
                .padding(12.dp),
            horizontalArrangement = Arrangement.spacedBy(10.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            // 车号输入：深底 + 白色输入文字，避免黑字看不清。
            val inputTextStyle = TextStyle(
                fontSize = 16.sp,
                color = Color.White,
                fontWeight = FontWeight.Medium,
            )
            CompositionLocalProvider(
                LocalContentColor provides Color.White,
                LocalTextStyle provides inputTextStyle,
            ) {
                BasicTextField(
                    value = state.carInput,
                    onValueChange = { app.freeMoveCarFeature.setCarInput(it) },
                    singleLine = true,
                    textStyle = inputTextStyle,
                    cursorBrush = SolidColor(Color.White),
                    modifier = Modifier
                        .weight(1f)
                        .background(Color(0xFF242936), RoundedCornerShape(6.dp))
                        .border(1.dp, Color(0xFF3A4158), RoundedCornerShape(6.dp))
                        .padding(horizontal = 12.dp, vertical = 12.dp),
                    decorationBox = { inner ->
                        Box(modifier = Modifier.fillMaxWidth()) {
                            if (state.carInput.isEmpty()) {
                                Text(
                                    t(Str.FieldChangeBatteryHint),
                                    color = Color(0xFF9AA0B5),
                                    fontSize = 15.sp,
                                )
                            }
                            inner()
                        }
                    },
                )
            }
            Box(
                modifier = Modifier
                    .background(colors.primary, RoundedCornerShape(6.dp))
                    .clickable(enabled = !state.loading) {
                        onScannedOrSubmit(state.carInput)
                    }
                    .padding(horizontal = 18.dp, vertical = 12.dp),
            ) {
                Text(
                    text = when (tab) {
                        FieldMoveTab.Unlock -> t(Str.ScanUnlock)
                        FieldMoveTab.Lock -> t(Str.ScanLock)
                    },
                    color = colors.onPrimary,
                    fontSize = 15.sp,
                )
            }
        }

        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.Center,
        ) {
            MoveTab(
                label = t(Str.ScanUnlockTab),
                selected = tab == FieldMoveTab.Unlock,
                onClick = { tab = FieldMoveTab.Unlock },
                color = colors.primary,
            )
            Spacer(modifier = Modifier.width(32.dp))
            MoveTab(
                label = t(Str.ScanLockTab),
                selected = tab == FieldMoveTab.Lock,
                onClick = { tab = FieldMoveTab.Lock },
                color = colors.primary,
            )
        }

        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp, vertical = 8.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(t(Str.VehicleId), modifier = Modifier.weight(1.2f), fontSize = 13.sp, color = colors.textSecondary)
            Text(
                t(Str.VehicleListColStatus),
                modifier = Modifier.weight(1f),
                textAlign = TextAlign.Center,
                fontSize = 13.sp,
                color = colors.textSecondary,
            )
            Text(
                t(Str.VehicleListColBattery),
                modifier = Modifier.weight(0.7f),
                textAlign = TextAlign.Center,
                fontSize = 13.sp,
                color = colors.textSecondary,
            )
            Text(
                t(Str.FreeMoveRemove),
                modifier = Modifier.weight(0.7f),
                textAlign = TextAlign.End,
                fontSize = 13.sp,
                color = colors.textSecondary,
            )
            SelectBox(
                checked = state.cars.isNotEmpty() && state.selectedCarIds.size == state.cars.size,
                onClick = {
                    if (state.selectedCarIds.size == state.cars.size) {
                        app.freeMoveCarFeature.clearSelection()
                    } else {
                        app.freeMoveCarFeature.selectAll()
                    }
                },
            )
        }
        HorizontalDivider(color = colors.divider)

        LazyColumn(
            modifier = Modifier
                .weight(1f)
                .fillMaxWidth(),
        ) {
            items(state.cars, key = { it.carId }) { car ->
                val selected = car.carId in state.selectedCarIds
                val statusText = when (car.state) {
                    FreeMoveCar.STATE_OPERATION -> t(Str.RidingOperation)
                    0 -> t(Str.FreeMoveUnlocking)
                    else -> t(Str.FreeMoveMoving)
                }
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .clickable { app.freeMoveCarFeature.toggleSelect(car.carId) }
                        .padding(horizontal = 16.dp, vertical = 12.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text(car.carId, modifier = Modifier.weight(1.2f), fontSize = 14.sp)
                    Text(statusText, modifier = Modifier.weight(1f), textAlign = TextAlign.Center, fontSize = 14.sp)
                    Text("${car.restBattery}%", modifier = Modifier.weight(0.7f), textAlign = TextAlign.Center, fontSize = 14.sp)
                    Text(
                        t(Str.FreeMoveRemove),
                        modifier = Modifier
                            .weight(0.7f)
                            .clickable {
                                scope.launch { app.freeMoveCarFeature.removeCar(car.carId) }
                            },
                        textAlign = TextAlign.End,
                        color = colors.primary,
                        fontSize = 14.sp,
                    )
                    SelectBox(
                        checked = selected,
                        onClick = { app.freeMoveCarFeature.toggleSelect(car.carId) },
                    )
                }
                HorizontalDivider(color = colors.divider)
            }
        }

        state.message?.let {
            Text(it, modifier = Modifier.padding(horizontal = 16.dp), color = colors.primary, fontSize = 12.sp)
        }
        (state.errorMessage ?: actionHint)?.let {
            Text(it, modifier = Modifier.padding(horizontal = 16.dp), color = Color(0xFFE53935), fontSize = 12.sp)
        }

        Box(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp)
                .background(colors.primary, RoundedCornerShape(6.dp))
                .clickable(enabled = !state.loading && state.selectedCarIds.isNotEmpty()) {
                    scope.launch { app.freeMoveCarFeature.finishSelected() }
                }
                .padding(vertical = 14.dp),
            contentAlignment = Alignment.Center,
        ) {
            Text(
                text = if (state.loading) t(Str.LoadingEllipsis) else t(Str.FreeMoveComplete),
                color = colors.onPrimary,
                fontSize = 16.sp,
                fontWeight = FontWeight.Medium,
            )
        }
    }
}

@Composable
private fun MoveTab(
    label: String,
    selected: Boolean,
    onClick: () -> Unit,
    color: Color,
) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        modifier = Modifier
            .clickable(onClick = onClick)
            .padding(vertical = 8.dp),
    ) {
        Text(
            label,
            color = if (selected) color else Color(0xFF888888),
            fontWeight = if (selected) FontWeight.SemiBold else FontWeight.Normal,
            fontSize = 15.sp,
        )
        Spacer(modifier = Modifier.height(4.dp))
        Box(
            modifier = Modifier
                .width(48.dp)
                .height(2.dp)
                .background(if (selected) color else Color.Transparent),
        )
    }
}

@Composable
private fun SelectBox(checked: Boolean, onClick: () -> Unit) {
    Box(
        modifier = Modifier
            .padding(start = 4.dp)
            .size(18.dp)
            .border(1.dp, Color(0xFFAAAAAA), RoundedCornerShape(2.dp))
            .background(if (checked) Color(0xFF1887F8) else Color.Transparent)
            .clickable(onClick = onClick),
    )
}
