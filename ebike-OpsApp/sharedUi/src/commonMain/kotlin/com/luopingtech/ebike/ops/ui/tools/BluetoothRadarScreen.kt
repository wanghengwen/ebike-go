package com.luopingtech.ebike.ops.ui.tools

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.ui.analysis.VcdTopBar
import kotlinx.coroutines.launch

@Composable
fun BluetoothRadarScreen(
    app: OpsApp,
    onClose: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.bluetoothRadarFeature.state.collectAsState()
    val scope = rememberCoroutineScope()

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(Color(0xFFF6F7F9)),
    ) {
        VcdTopBar(title = t(Str.BluetoothRadar), onBack = {
            app.bluetoothRadarFeature.clear()
            onClose()
        })
        Column(
            modifier = Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState())
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            if (!state.bleAvailable) {
                Text(t(Str.BleUnavailable), color = MaterialTheme.colorScheme.error)
            }
            Button(
                onClick = { scope.launch { app.bluetoothRadarFeature.scan() } },
                enabled = state.bleAvailable && !state.scanning,
                modifier = Modifier.fillMaxWidth(),
            ) {
                Text(if (state.scanning) t(Str.LoadingEllipsis) else t(Str.BleScan))
            }
            state.devices.forEach { device ->
                val selected = state.selectedId == device.id
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .clickable { app.bluetoothRadarFeature.selectDevice(device.id) }
                        .background(
                            if (selected) Color(0xFFE8F0FF) else Color.White,
                            MaterialTheme.shapes.medium,
                        )
                        .padding(12.dp),
                    horizontalArrangement = Arrangement.SpaceBetween,
                ) {
                    Column {
                        Text(device.name ?: device.id, style = MaterialTheme.typography.titleSmall)
                        Text(device.id, style = MaterialTheme.typography.bodySmall)
                    }
                    Text("RSSI ${device.rssi ?: "-"}", style = MaterialTheme.typography.bodySmall)
                }
            }
            Button(
                onClick = { scope.launch { app.bluetoothRadarFeature.ring() } },
                enabled = !state.ringing && state.selectedId != null,
                modifier = Modifier.fillMaxWidth(),
            ) {
                Text(if (state.ringing) t(Str.LoadingEllipsis) else t(Str.BleRing))
            }
            state.message?.let { Text(it) }
            state.errorMessage?.let { Text(it, color = MaterialTheme.colorScheme.error) }
        }
    }
}
