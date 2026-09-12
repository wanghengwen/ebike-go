package com.luopingtech.ebike.ops.ui.vehicle

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
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
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.Vehicle
import kotlinx.coroutines.launch

@Composable
fun VehicleDetailSection(
    app: OpsApp,
    vehicle: Vehicle,
    serviceAreaId: String? = null,
    canBindBattery: Boolean = false,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val mapState by app.vehicleDetailMapFeature.state.collectAsState()
    val vehicleState by app.vehicleFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    val displayVehicle = remember(vehicleState.vehicles, vehicle.carId, vehicle) {
        vehicleState.vehicles.firstOrNull { it.carId.equals(vehicle.carId, ignoreCase = true) }
            ?: vehicle
    }
    var snInput by remember(displayVehicle.carId) { mutableStateOf("") }
    var bindMessage by remember(displayVehicle.carId) { mutableStateOf<String?>(null) }
    var bindError by remember(displayVehicle.carId) { mutableStateOf<String?>(null) }

    LaunchedEffect(vehicle.carId, vehicle.lat, vehicle.lng, serviceAreaId) {
        app.vehicleDetailMapFeature.loadForVehicle(vehicle, serviceAreaId)
    }

    Column(
        modifier = Modifier.fillMaxWidth(),
        verticalArrangement = Arrangement.spacedBy(4.dp),
    ) {
        DetailLine(
            "${t(Str.Voltage)} ${displayVehicle.voltageLabel} · ${t(Str.BatterySn)} ${displayVehicle.batterySn.ifBlank { "-" }}",
        )
        if (canBindBattery) {
            OutlinedTextField(
                value = snInput,
                onValueChange = {
                    snInput = it
                    bindMessage = null
                    bindError = null
                },
                label = { Text(t(Str.BindBatterySnHint)) },
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
                                is OpsResult.Ok -> {
                                    snInput = scan.value
                                    bindMessage = null
                                    bindError = null
                                }
                                is OpsResult.Err -> bindError = scan.error.message
                            }
                        }
                    },
                    enabled = !vehicleState.detailLoading,
                    modifier = Modifier.weight(1f),
                ) { Text(t(Str.WarehouseScan)) }
                Button(
                    onClick = {
                        scope.launch {
                            bindMessage = null
                            bindError = null
                            when (
                                val result = app.vehicleFeature.bindBatterySn(
                                    displayVehicle.carId,
                                    snInput,
                                )
                            ) {
                                is OpsResult.Ok -> {
                                    snInput = ""
                                    bindMessage = t(Str.BindBatterySnOk)
                                }
                                is OpsResult.Err -> bindError = result.error.message
                            }
                        }
                    },
                    enabled = !vehicleState.detailLoading && snInput.isNotBlank(),
                    modifier = Modifier.weight(1f),
                ) { Text(t(Str.BindBatterySn)) }
            }
            bindMessage?.let {
                Text(it, color = MaterialTheme.colorScheme.primary, style = MaterialTheme.typography.bodySmall)
            }
            bindError?.let {
                Text(it, color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.bodySmall)
            }
        }
        DetailLine(
            "${t(Str.BatteryLock)} ${displayVehicle.batteryLockLabel} · ${t(Str.HelmetLock)} ${displayVehicle.helmetLockLabel}",
        )
        if (displayVehicle.helmetMac.isNotBlank()) {
            DetailLine("${t(Str.HelmetMac)} ${displayVehicle.helmetMac}")
        }
        DetailLine("${t(Str.SiteName)} ${displayVehicle.siteLabel}")
        mapState.address?.let { addr ->
            DetailLine("${t(Str.VehicleAddress)} $addr")
        }
        DetailLine("${t(Str.Mileage)} ${displayVehicle.mileageLabel}")
        DetailLine("IMEI ${displayVehicle.imei.ifBlank { "-" }}")
        DetailLine(
            "${t(Str.ServiceAreaName)} ${displayVehicle.serviceName.ifBlank { displayVehicle.serviceId.ifBlank { "-" } }}",
        )
        DetailLine("${t(Str.OpsStateLabel)} ${displayVehicle.operationLabels}")
        DetailLine("${t(Str.AlarmLabel)} ${displayVehicle.alarmLabels}")

        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            FilterChip(
                selected = mapState.showFence,
                onClick = {
                    val next = !mapState.showFence
                    app.vehicleDetailMapFeature.setShowFence(next)
                    if (next && mapState.fence == null) {
                        scope.launch {
                            val sid = (serviceAreaId ?: displayVehicle.serviceId).ifBlank { return@launch }
                            val near = mapState.lastOrder?.nearFenceLocations().orEmpty()
                            if (near.isNotEmpty()) {
                                app.vehicleDetailMapFeature.loadNearFence(sid, near)
                            } else {
                                app.vehicleDetailMapFeature.loadFence(sid)
                            }
                        }
                    }
                },
                label = { Text(t(Str.VehicleFence)) },
            )
            FilterChip(
                selected = mapState.showTrack,
                onClick = {
                    if (!mapState.showTrack) {
                        scope.launch { app.vehicleDetailMapFeature.loadTrack(displayVehicle) }
                    } else {
                        app.vehicleDetailMapFeature.setShowTrack(false)
                    }
                },
                label = {
                    Text(
                        if (mapState.loadingTrack) t(Str.VehicleTrackLoad) else t(Str.VehicleTrack),
                    )
                },
            )
            TextButton(
                onClick = {
                    scope.launch { app.vehicleDetailMapFeature.loadRealtimeTrack(displayVehicle) }
                },
            ) {
                Text(t(Str.VehicleTrackRealtime))
            }
        }
        mapState.errorMessage?.let {
            Text(it, color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.bodySmall)
        }
        if (mapState.showTrack && mapState.track.isNotEmpty()) {
            DetailLine("${t(Str.VehicleTrack)} · ${mapState.track.size}")
        }
    }
}

@Composable
private fun DetailLine(text: String) {
    Text(
        text = text,
        style = MaterialTheme.typography.bodySmall,
        color = MaterialTheme.colorScheme.onSurfaceVariant,
    )
}
