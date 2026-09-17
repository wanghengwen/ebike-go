package com.luopingtech.ebike.ops.ui.tools

import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.CornerRadius
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.platform.BleDevice
import com.luopingtech.ebike.ops.ui.analysis.VcdTopBar
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch

/**
 * 对齐 NewBluetoothRadarActivity：车号过滤、黄底提示、自动扫描、行内寻车。
 */
@Composable
fun BluetoothRadarScreen(
    app: OpsApp,
    onClose: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.bluetoothRadarFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    val colors = OpsTheme.colors

    LaunchedEffect(Unit) {
        while (isActive) {
            app.bluetoothRadarFeature.scan()
            delay(10_000)
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.White),
    ) {
        VcdTopBar(title = t(Str.BluetoothRadar), onBack = {
            app.bluetoothRadarFeature.clear()
            onClose()
        })
        OutlinedTextField(
            value = state.filter,
            onValueChange = { app.bluetoothRadarFeature.setFilter(it) },
            placeholder = { Text(t(Str.BleFilterHint), color = Color(0xFF999999)) },
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp, vertical = 12.dp),
            singleLine = true,
        )
        Text(
            text = t(Str.BleRadarTip),
            color = Color(0xFFFA6400),
            fontSize = 14.sp,
            modifier = Modifier
                .fillMaxWidth()
                .background(Color(0xFFFFF8E5))
                .padding(horizontal = 16.dp, vertical = 12.dp),
        )
        if (!state.bleAvailable) {
            Text(
                t(Str.BleUnavailable),
                color = Color(0xFFE02020),
                modifier = Modifier.padding(16.dp),
            )
        } else if (state.filteredDevices.isEmpty()) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 48.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
            ) {
                Text(
                    if (state.scanning) t(Str.BleSearching) else t(Str.BleSearching),
                    color = Color(0xFF242936),
                    fontSize = 16.sp,
                )
                Spacer(modifier = Modifier.height(12.dp))
                Text(t(Str.BleMoveHint), color = Color(0xFF9FA7C7), fontSize = 14.sp)
            }
        } else {
            LazyColumn(modifier = Modifier.fillMaxSize()) {
                items(state.filteredDevices, key = { it.id }) { device ->
                    BleRadarRow(
                        device = device,
                        ringing = state.ringingId == device.id,
                        ringLabel = t(Str.BleRing),
                        imeiLabel = t(Str.ImeiLabel),
                        onRing = {
                            scope.launch { app.bluetoothRadarFeature.ring(device.id) }
                        },
                    )
                }
            }
        }
        state.message?.let {
            Text(it, color = colors.primary, modifier = Modifier.padding(16.dp))
        }
        state.errorMessage?.let {
            Text(it, color = Color(0xFFE02020), modifier = Modifier.padding(16.dp))
        }
    }
}

@Composable
private fun BleRadarRow(
    device: BleDevice,
    ringing: Boolean,
    ringLabel: String,
    imeiLabel: String,
    onRing: () -> Unit,
) {
    val title = device.name?.takeIf { it.isNotBlank() } ?: device.id
    val rssi = device.rssi
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .height(87.dp)
            .padding(horizontal = 16.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Column(modifier = Modifier.weight(1f)) {
            Text(title, color = Color.Black, fontSize = 20.sp, fontWeight = FontWeight.Bold)
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                "$imeiLabel ${device.id}",
                color = Color.Black,
                fontSize = 14.sp,
            )
        }
        Column(horizontalAlignment = Alignment.End) {
            Row(verticalAlignment = Alignment.Bottom) {
                BleSignalBars(rssi = rssi)
                Spacer(modifier = Modifier.width(6.dp))
                Text("${rssi ?: "-"}", color = Color.Black, fontSize = 14.sp)
            }
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                "${rssi ?: "-"} dBm",
                color = signalStateColor(rssi),
                fontSize = 14.sp,
            )
        }
        Spacer(modifier = Modifier.width(8.dp))
        Box(
            modifier = Modifier
                .size(46.dp)
                .clickable(enabled = !ringing, onClick = onRing)
                .background(OpsTheme.colors.primary.copy(alpha = 0.12f)),
            contentAlignment = Alignment.Center,
        ) {
            Text(
                if (ringing) "…" else "♪",
                color = OpsTheme.colors.primary,
                fontSize = 20.sp,
            )
        }
    }
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .height(1.dp)
            .background(Color(0xFFDAE0F5)),
    )
}

@Composable
private fun BleSignalBars(rssi: Int?) {
    val level = when {
        rssi == null -> 0
        rssi >= -55 -> 4
        rssi >= -65 -> 3
        rssi >= -75 -> 2
        rssi >= -85 -> 1
        else -> 0
    }
    Canvas(modifier = Modifier.size(width = 22.dp, height = 16.dp)) {
        val barW = size.width / 5f
        val gap = barW / 3f
        for (i in 0 until 4) {
            val h = size.height * (i + 1) / 4f
            val color = if (i < level) Color(0xFF63D144) else Color(0xFFD0D0D0)
            drawRoundRect(
                color = color,
                topLeft = Offset(i * (barW + gap), size.height - h),
                size = Size(barW, h),
                cornerRadius = CornerRadius(1.dp.toPx(), 1.dp.toPx()),
            )
        }
    }
}

private fun signalStateColor(rssi: Int?): Color = when {
    rssi == null -> Color(0xFF999999)
    rssi >= -65 -> Color(0xFF63D144)
    rssi >= -80 -> Color(0xFFFA6400)
    else -> Color(0xFFE02020)
}
