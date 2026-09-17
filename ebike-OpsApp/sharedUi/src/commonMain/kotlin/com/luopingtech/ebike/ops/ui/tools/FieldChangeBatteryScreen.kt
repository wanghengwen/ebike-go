package com.luopingtech.ebike.ops.ui.tools

import androidx.compose.foundation.Canvas
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
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
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
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.ui.icons.OpsBackChevron
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import kotlinx.coroutines.launch

/**
 * 工作台现场换电页。相机预览由宿主注入（Android: OpsCameraScanPreview）。
 */
@Composable
fun FieldChangeBatteryScreen(
    app: OpsApp,
    onClose: () -> Unit,
    onHelp: () -> Unit = {},
    /** Legacy ReplaceBatteryActivity1 intent carId. */
    initialCarId: String? = null,
    /** Legacy: enter page then openBatteryBox() immediately. */
    autoOpenBox: Boolean = false,
    scanPreview: @Composable (
        modifier: Modifier,
        torchOn: Boolean,
        enabled: Boolean,
        onCode: (String) -> Unit,
    ) -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.fieldChangeBatteryFeature.state.collectAsState()
    val colors = OpsTheme.colors
    val scope = rememberCoroutineScope()
    var torchOn by remember { mutableStateOf(false) }
    var scanEnabled by remember { mutableStateOf(true) }
    var lastCode by remember { mutableStateOf<String?>(null) }
    val scanBg = Color(0xFF242936)

    DisposableEffect(Unit) {
        onDispose { app.fieldChangeBatteryFeature.clear() }
    }

    com.luopingtech.ebike.ops.ui.navigation.OpsBackHandler(onBack = onClose)

    LaunchedEffect(initialCarId, autoOpenBox) {
        val carId = initialCarId?.trim().orEmpty()
        if (carId.isBlank()) return@LaunchedEffect
        app.fieldChangeBatteryFeature.setCarInput(carId)
        if (autoOpenBox) {
            app.fieldChangeBatteryFeature.openBatteryBox()
        }
    }

    Column(modifier = Modifier.fillMaxSize().background(Color.White)) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .background(colors.primary)
                .statusBarsPadding()
                .padding(horizontal = 8.dp, vertical = 10.dp),
        ) {
            OpsBackChevron(
                onClick = onClose,
                modifier = Modifier.align(Alignment.CenterStart),
            )
            Text(
                text = t(Str.ChangeBatteryTool),
                color = colors.onPrimary,
                fontSize = 18.sp,
                fontWeight = FontWeight.Medium,
                modifier = Modifier.align(Alignment.Center),
            )
            Text(
                text = "?",
                color = colors.onPrimary,
                fontSize = 18.sp,
                modifier = Modifier
                    .align(Alignment.CenterEnd)
                    .clickable(onClick = onHelp)
                    .border(1.dp, colors.onPrimary, RoundedCornerShape(50))
                    .padding(horizontal = 8.dp, vertical = 2.dp),
            )
        }

        // 上半：仅预览框（裁剪边界），不在此区域叠操作文案。
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .weight(1f)
                .background(scanBg),
            contentAlignment = Alignment.Center,
        ) {
            Box(
                modifier = Modifier
                    .fillMaxWidth(0.72f)
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
                    scope.launch {
                        app.fieldChangeBatteryFeature.applyScanRaw(raw)
                        scanEnabled = true
                    }
                }
                ScanCornerBrackets(color = colors.primary)
            }
        }

        // 不透明操作条：与预览分区，避免字压在画面上。
        Row(
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.Center,
            modifier = Modifier
                .fillMaxWidth()
                .background(scanBg)
                .clickable { torchOn = !torchOn }
                .padding(vertical = 14.dp),
        ) {
            CheckboxMark(checked = torchOn)
            Spacer(modifier = Modifier.width(8.dp))
            Text(t(Str.Torch), color = Color.White, fontSize = 15.sp)
        }

        Column(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color.White)
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(10.dp),
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(10.dp),
            ) {
                BasicTextField(
                    value = state.carInput,
                    onValueChange = { app.fieldChangeBatteryFeature.setCarInput(it) },
                    singleLine = true,
                    textStyle = TextStyle(fontSize = 15.sp, color = colors.textPrimary),
                    cursorBrush = SolidColor(colors.primary),
                    modifier = Modifier
                        .weight(1f)
                        .border(1.dp, colors.divider, RoundedCornerShape(6.dp))
                        .padding(horizontal = 12.dp, vertical = 12.dp),
                    decorationBox = { inner ->
                        if (state.carInput.isEmpty()) {
                            Text(t(Str.FieldChangeBatteryHint), color = colors.textTertiary, fontSize = 15.sp)
                        }
                        inner()
                    },
                )
                Box(
                    modifier = Modifier
                        .background(colors.primary, RoundedCornerShape(6.dp))
                        .clickable(enabled = !state.loading) {
                            scope.launch { app.fieldChangeBatteryFeature.openBatteryBox() }
                        }
                        .padding(horizontal = 16.dp, vertical = 12.dp),
                ) {
                    Text(
                        text = if (state.loading) t(Str.LoadingEllipsis) else t(Str.OpenBatteryBox),
                        color = colors.onPrimary,
                        fontSize = 15.sp,
                    )
                }
            }
            state.message?.let { Text(it, color = colors.primary, fontSize = 13.sp) }
            state.errorMessage?.let { Text(it, color = Color(0xFFE53935), fontSize = 13.sp) }
        }
    }
}

@Composable
internal fun ScanCornerBrackets(color: Color, modifier: Modifier = Modifier) {
    Canvas(modifier = modifier.fillMaxSize()) {
        val len = size.minDimension * 0.14f
        val stroke = 4.dp.toPx()
        val inset = 2.dp.toPx()
        fun corner(x0: Float, y0: Float, dx: Float, dy: Float) {
            drawLine(color, Offset(x0, y0), Offset(x0 + dx * len, y0), stroke, StrokeCap.Square)
            drawLine(color, Offset(x0, y0), Offset(x0, y0 + dy * len), stroke, StrokeCap.Square)
        }
        corner(inset, inset, 1f, 1f)
        corner(size.width - inset, inset, -1f, 1f)
        corner(inset, size.height - inset, 1f, -1f)
        corner(size.width - inset, size.height - inset, -1f, -1f)
    }
}

@Composable
internal fun CheckboxMark(checked: Boolean) {
    Box(
        modifier = Modifier
            .size(18.dp)
            .border(1.5.dp, Color.White, RoundedCornerShape(2.dp))
            .background(if (checked) Color.White else Color.Transparent),
        contentAlignment = Alignment.Center,
    ) {
        if (checked) {
            Text("✓", color = Color(0xFF242936), fontSize = 12.sp, fontWeight = FontWeight.Bold)
        }
    }
}

