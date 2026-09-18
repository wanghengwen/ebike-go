package com.luopingtech.ebike.ops.ui.relocation

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Checkbox
import androidx.compose.material3.OutlinedButton
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
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.domain.model.MapPin
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.ui.analysis.VcdTopBar
import com.luopingtech.ebike.ops.ui.map.OpsMapSpec
import com.luopingtech.ebike.ops.ui.map.OpsMapView
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import com.luopingtech.ebike.ops.ui.tools.LegacyScanPreviewBlock
import kotlinx.coroutines.launch

/**
 * 对齐 Flutter RelocationScanPage → RelocateMapPage：页内连续扫码后进地图确认。
 */
@Composable
fun RelocationScreen(
    app: OpsApp,
    currentArea: ServiceArea?,
    onClose: () -> Unit,
    scanPreview: @Composable (
        modifier: Modifier,
        torchOn: Boolean,
        enabled: Boolean,
        onCode: (String) -> Unit,
    ) -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.relocationFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    val colors = OpsTheme.colors
    var mapStep by remember { mutableStateOf(false) }
    var torchOn by remember { mutableStateOf(false) }
    var scanEnabled by remember { mutableStateOf(true) }
    var lastCode by remember { mutableStateOf<String?>(null) }

    LaunchedEffect(Unit) {
        if (state.pinLat == null || state.pinLng == null) {
            app.relocationFeature.refreshPinFromGps()
        }
    }

    fun onScanned(raw: String) {
        if (!scanEnabled || state.loading || mapStep) return
        if (raw == lastCode) return
        lastCode = raw
        scanEnabled = false
        scope.launch {
            app.relocationFeature.addByRaw(raw, currentArea)
            scanEnabled = true
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.White),
    ) {
        VcdTopBar(
            title = if (mapStep) t(Str.RelocationMapTitle) else t(Str.Relocation),
            onBack = {
                if (mapStep) {
                    mapStep = false
                } else {
                    app.relocationFeature.clear()
                    onClose()
                }
            },
        )

        if (!mapStep) {
            LegacyScanPreviewBlock(
                torchOn = torchOn,
                torchLabel = t(Str.Torch),
                scanEnabled = scanEnabled && !state.loading,
                onTorchChange = { torchOn = it },
                scanPreview = scanPreview,
                onCode = ::onScanned,
            )
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .verticalScroll(rememberScrollState())
                    .padding(16.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                Text(t(Str.RelocationHint), color = Color(0xFF666666), fontSize = 13.sp)
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(10.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    OutlinedTextField(
                        value = state.carInput,
                        onValueChange = { app.relocationFeature.setCarInput(it) },
                        placeholder = { Text(t(Str.EnterCarIdOrScan)) },
                        modifier = Modifier.weight(1f),
                        singleLine = true,
                    )
                    Button(
                        onClick = { scope.launch { app.relocationFeature.addManual(currentArea) } },
                        enabled = !state.loading && state.carInput.isNotBlank(),
                        shape = RoundedCornerShape(4.dp),
                        colors = ButtonDefaults.buttonColors(
                            containerColor = colors.primary,
                            contentColor = colors.onPrimary,
                        ),
                    ) { Text(t(Str.Confirm)) }
                }
                if (state.devices.isNotEmpty()) {
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Text(t(Str.VehicleTagCarId), color = Color(0xFF666666), fontSize = 14.sp, modifier = Modifier.weight(1f))
                        Text(t(Str.VehicleListColStatus), color = Color(0xFF666666), fontSize = 14.sp, modifier = Modifier.weight(1f))
                        Text(t(Str.VehicleTagActionCol), color = Color(0xFF666666), fontSize = 14.sp, modifier = Modifier.weight(1f))
                        Checkbox(
                            checked = state.allSelected,
                            onCheckedChange = { app.relocationFeature.setAllSelected(it) },
                        )
                    }
                    Box(modifier = Modifier.fillMaxWidth().height(1.dp).background(Color(0xFFCCCCCC)))
                    state.devices.forEach { device ->
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(44.dp)
                                .clickable { app.relocationFeature.toggleSelected(device.carId) },
                            verticalAlignment = Alignment.CenterVertically,
                        ) {
                            Text(device.carId, modifier = Modifier.weight(1f), color = Color(0xFF333333), fontSize = 15.sp)
                            Text(
                                if (device.isOffline) t(Str.Offline) else t(Str.Online),
                                modifier = Modifier.weight(1f),
                                color = if (device.isOffline) Color(0xFFE02020) else Color(0xFF63D144),
                                fontSize = 14.sp,
                            )
                            TextButton(
                                onClick = { app.relocationFeature.remove(device.carId) },
                                modifier = Modifier.weight(1f),
                            ) {
                                Text(t(Str.VehicleTagDelete), color = Color(0xFFE02020))
                            }
                            Checkbox(
                                checked = device.selected,
                                onCheckedChange = { app.relocationFeature.toggleSelected(device.carId) },
                            )
                        }
                        Box(modifier = Modifier.fillMaxWidth().height(1.dp).background(Color(0xFFEEEEEE)))
                    }
                    Button(
                        onClick = {
                            scope.launch {
                                app.relocationFeature.refreshPinFromGps()
                                mapStep = true
                            }
                        },
                        enabled = state.selected.isNotEmpty(),
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(48.dp),
                        colors = ButtonDefaults.buttonColors(
                            containerColor = colors.primary,
                            contentColor = colors.onPrimary,
                        ),
                    ) {
                        Text(t(Str.RelocationNextMap), fontWeight = FontWeight.Bold)
                    }
                }
                state.message?.let { Text(it, color = colors.primary) }
                state.errorMessage?.let { Text(it, color = Color(0xFFE02020)) }
            }
        } else {
            val lat = state.pinLat ?: 0.0
            val lng = state.pinLng ?: 0.0
            val pin = MapPin(
                id = "relocation-pin",
                lat = lat,
                lng = lng,
                title = t(Str.Relocation),
            )
            Box(modifier = Modifier.weight(1f)) {
                OpsMapView(
                    spec = OpsMapSpec(
                        pins = listOf(pin),
                        selectedCarId = pin.id,
                        clusterOverview = false,
                        followLat = lat,
                        followLng = lng,
                        followNonce = 1,
                        showStatusOverlay = false,
                    ),
                    modifier = Modifier.fillMaxSize(),
                )
                Column(
                    modifier = Modifier
                        .align(Alignment.BottomCenter)
                        .fillMaxWidth()
                        .background(Color.White)
                        .padding(16.dp),
                    verticalArrangement = Arrangement.spacedBy(10.dp),
                ) {
                    Text("${t(Str.RelocationPinLat)}: ${state.pinLat ?: "-"}", color = Color(0xFF333333), fontSize = 14.sp)
                    Text("${t(Str.RelocationPinLng)}: ${state.pinLng ?: "-"}", color = Color(0xFF333333), fontSize = 14.sp)
                    OutlinedButton(
                        onClick = { scope.launch { app.relocationFeature.refreshPinFromGps() } },
                        enabled = !state.loading,
                        modifier = Modifier.fillMaxWidth(),
                        border = BorderStroke(1.dp, colors.primary),
                        colors = ButtonDefaults.outlinedButtonColors(contentColor = colors.primary),
                    ) { Text(t(Str.RelocationRefreshGps)) }
                    Button(
                        onClick = { scope.launch { app.relocationFeature.confirmLocation() } },
                        enabled = !state.loading && state.selected.isNotEmpty(),
                        modifier = Modifier.fillMaxWidth().height(48.dp),
                        colors = ButtonDefaults.buttonColors(
                            containerColor = colors.primary,
                            contentColor = colors.onPrimary,
                        ),
                    ) {
                        Text(
                            if (state.reporting) t(Str.LoadingEllipsis) else t(Str.RelocationConfirm),
                            fontWeight = FontWeight.Bold,
                        )
                    }
                    state.message?.let { Text(it, color = colors.primary) }
                    state.errorMessage?.let { Text(it, color = Color(0xFFE02020)) }
                }
            }
        }
    }
}
