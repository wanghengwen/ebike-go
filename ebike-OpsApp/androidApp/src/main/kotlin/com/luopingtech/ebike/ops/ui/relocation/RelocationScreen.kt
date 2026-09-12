package com.luopingtech.ebike.ops.ui.relocation

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
import androidx.compose.material3.FilterChip
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import kotlinx.coroutines.launch

@Composable
fun RelocationScreen(
    app: OpsApp,
    currentArea: ServiceArea?,
    onClose: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.relocationFeature.state.collectAsState()
    val scope = rememberCoroutineScope()

    LaunchedEffect(Unit) {
        if (state.pinLat == null || state.pinLng == null) {
            app.relocationFeature.refreshPinFromGps()
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState())
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
        ) {
            Text(t(Str.Relocation), style = MaterialTheme.typography.headlineSmall)
            TextButton(onClick = {
                app.relocationFeature.clear()
                onClose()
            }) { Text(t(Str.Back)) }
        }
        Text(
            text = t(Str.RelocationHint),
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
        OutlinedTextField(
            value = state.carInput,
            onValueChange = { app.relocationFeature.setCarInput(it) },
            label = { Text(t(Str.EnterCarIdOrScan)) },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true,
        )
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Button(
                onClick = {
                    scope.launch {
                        when (val scan = app.codeScanner.scanOnce()) {
                            is OpsResult.Ok -> app.relocationFeature.addByRaw(scan.value, currentArea)
                            is OpsResult.Err -> Unit
                        }
                    }
                },
                enabled = !state.loading,
                modifier = Modifier.weight(1f),
            ) { Text(t(Str.WarehouseScan)) }
            Button(
                onClick = { scope.launch { app.relocationFeature.addManual(currentArea) } },
                enabled = !state.loading && state.carInput.isNotBlank(),
                modifier = Modifier.weight(1f),
            ) { Text(t(Str.Confirm)) }
        }

        if (state.devices.isNotEmpty()) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                FilterChip(
                    selected = state.allSelected,
                    onClick = { app.relocationFeature.setAllSelected(!state.allSelected) },
                    label = { Text(t(Str.RelocationSelectAll)) },
                )
            }
            state.devices.forEach { device ->
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .clickable { app.relocationFeature.toggleSelected(device.carId) }
                        .padding(vertical = 4.dp),
                    horizontalArrangement = Arrangement.SpaceBetween,
                ) {
                    Text(
                        text = "${if (device.selected) "●" else "○"} ${device.carId} · ${device.imei} · ${device.restBattery}%",
                        modifier = Modifier.weight(1f),
                        color = if (device.selected) {
                            MaterialTheme.colorScheme.primary
                        } else {
                            MaterialTheme.colorScheme.onSurface
                        },
                    )
                    TextButton(onClick = { app.relocationFeature.remove(device.carId) }) {
                        Text(t(Str.Close))
                    }
                }
            }
        }

        OutlinedTextField(
            value = state.pinLat?.toString().orEmpty(),
            onValueChange = {
                app.relocationFeature.setPin(it.toDoubleOrNull(), state.pinLng)
            },
            label = { Text(t(Str.RelocationPinLat)) },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true,
        )
        OutlinedTextField(
            value = state.pinLng?.toString().orEmpty(),
            onValueChange = {
                app.relocationFeature.setPin(state.pinLat, it.toDoubleOrNull())
            },
            label = { Text(t(Str.RelocationPinLng)) },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true,
        )
        Button(
            onClick = { scope.launch { app.relocationFeature.refreshPinFromGps() } },
            enabled = !state.loading,
            modifier = Modifier.fillMaxWidth(),
        ) { Text(t(Str.RelocationRefreshGps)) }
        Button(
            onClick = { scope.launch { app.relocationFeature.confirmLocation() } },
            enabled = !state.loading && state.selected.isNotEmpty(),
            modifier = Modifier.fillMaxWidth(),
        ) {
            Text(
                if (state.reporting) t(Str.LoadingEllipsis) else t(Str.RelocationConfirm),
            )
        }
        state.message?.let {
            Text(it, color = MaterialTheme.colorScheme.primary)
        }
        state.errorMessage?.let {
            Text(it, color = MaterialTheme.colorScheme.error)
        }
    }
}
