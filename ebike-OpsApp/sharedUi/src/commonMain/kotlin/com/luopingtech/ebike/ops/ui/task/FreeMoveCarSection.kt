package com.luopingtech.ebike.ops.ui.task

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.FilterChip
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
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
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import kotlinx.coroutines.launch

@OptIn(ExperimentalLayoutApi::class)
@Composable
fun FreeMoveCarSection(
    app: OpsApp,
    currentArea: ServiceArea?,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.freeMoveCarFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    var photoHint by remember { mutableStateOf<String?>(null) }

    LaunchedEffect(currentArea?.id) {
        app.freeMoveCarFeature.load(currentArea)
    }

    LaunchedEffect(state.needPhotograph, state.photoUrls.isNotEmpty()) {
        if (state.needPhotograph || state.photoUrls.isNotEmpty()) {
            app.freeMoveCarFeature.ensureTeamWorkersLoaded()
        }
    }

    /** 相机与相册只差取图那一步，权限与取消都由 [PhotoCapture] 实现方吞掉。 */
    fun addPhoto(take: suspend () -> OpsResult<String>) {
        scope.launch {
            when (val shot = take()) {
                is OpsResult.Ok -> {
                    app.freeMoveCarFeature.addPhotoUrl(shot.value)
                    photoHint = null
                }
                is OpsResult.Err -> photoHint = shot.error.message
            }
        }
    }

    Text(
        text = t(Str.FreeMoveHint),
        style = MaterialTheme.typography.bodySmall,
        color = MaterialTheme.colorScheme.onSurfaceVariant,
    )

    OutlinedTextField(
        value = state.carInput,
        onValueChange = { app.freeMoveCarFeature.setCarInput(it) },
        label = { Text(t(Str.VehicleId)) },
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
                            val hosts = app.authFeature.state.value.runtimeConfig?.qrHosts.orEmpty()
                            app.freeMoveCarFeature.addByScanRaw(scan.value, currentArea, hosts)
                        }
                        is OpsResult.Err -> photoHint = scan.error.message
                    }
                }
            },
            enabled = !state.loading,
            modifier = Modifier.weight(1f),
        ) { Text(t(Str.FreeMoveJoinScan)) }
        Button(
            onClick = {
                scope.launch {
                    app.freeMoveCarFeature.addByCarId(state.carInput, currentArea)
                }
            },
            enabled = !state.loading,
            modifier = Modifier.weight(1f),
        ) { Text(if (state.loading) t(Str.LoadingEllipsis) else t(Str.FreeMoveJoin)) }
    }

    Button(
        onClick = { scope.launch { app.freeMoveCarFeature.load(currentArea) } },
        enabled = !state.loading,
        modifier = Modifier.fillMaxWidth(),
    ) { Text(t(Str.FreeMoveRefresh)) }

    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        TextButton(onClick = { app.freeMoveCarFeature.selectAll() }) { Text(t(Str.FreeMoveSelectAll)) }
        TextButton(onClick = { app.freeMoveCarFeature.clearSelection() }) { Text(t(Str.FreeMoveClearSelect)) }
    }

    state.cars.forEach { car ->
        val selected = car.carId in state.selectedCarIds
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .clickable { app.freeMoveCarFeature.toggleSelect(car.carId) }
                .padding(vertical = 4.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
        ) {
            Text(
                text = "${if (selected) "●" else "○"} ${car.carId} · ${t(Str.BatteryPercent, car.restBattery)}",
                color = if (selected) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.onSurface,
                modifier = Modifier.weight(1f),
            )
            TextButton(
                onClick = { scope.launch { app.freeMoveCarFeature.removeCar(car.carId) } },
                enabled = !state.loading,
            ) { Text(t(Str.FreeMoveRemove)) }
        }
    }

    Button(
        onClick = { scope.launch { app.freeMoveCarFeature.finishSelected() } },
        enabled = !state.loading &&
            state.selectedCarIds.isNotEmpty() &&
            (
                !state.needPhotograph ||
                    state.photoUrls.isNotEmpty()
                ),
        modifier = Modifier.fillMaxWidth(),
    ) {
        Text(
            when {
                state.loading -> t(Str.LoadingEllipsis)
                state.needPhotograph -> t(Str.FreeMoveFinishPhoto, state.selectedCarIds.size)
                else -> t(Str.FreeMoveFinish, state.selectedCarIds.size)
            },
        )
    }

    if (state.selectedCarIds.isNotEmpty() || state.needPhotograph || state.photoUrls.isNotEmpty()) {
        Text(
            text = when {
                state.needPhotograph -> t(Str.PhotoAuditCount, state.photoUrls.size)
                else -> t(Str.PhotoOptionalHint)
            },
            style = MaterialTheme.typography.bodySmall,
            color = if (state.needPhotograph) {
                MaterialTheme.colorScheme.tertiary
            } else {
                MaterialTheme.colorScheme.onSurfaceVariant
            },
        )
        if (state.needPhotograph || state.photoUrls.isNotEmpty()) {
            OutlinedTextField(
                value = state.remark,
                onValueChange = { app.freeMoveCarFeature.setRemark(it) },
                label = {
                    Text(
                        t(Str.PhotoRemarkOptional),
                    )
                },
                modifier = Modifier.fillMaxWidth(),
                minLines = 2,
            )
            Text(
                text = "${t(Str.TeamWorkerLabel)} · ${t(Str.TeamWorkerHint)}",
                style = MaterialTheme.typography.labelLarge,
            )
            if (state.teamCandidates.isEmpty()) {
                Text(
                    t(Str.TeamWorkerNone),
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            } else {
                FlowRow(
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                    verticalArrangement = Arrangement.spacedBy(4.dp),
                ) {
                    state.teamCandidates.forEach { worker ->
                        FilterChip(
                            selected = worker.selectionKey in state.selectedTeamKeys,
                            onClick = { app.freeMoveCarFeature.toggleTeamWorker(worker.selectionKey) },
                            label = { Text(worker.name) },
                        )
                    }
                }
                if (state.selectedTeamKeys.isNotEmpty()) {
                    Text(
                        t(Str.TeamWorkerSelected, state.selectedTeamKeys.size),
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.primary,
                    )
                }
            }
        }
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Button(
                onClick = { addPhoto { app.photoCapture.takePhoto("freemove") } },
                modifier = Modifier.weight(1f),
            ) { Text(t(Str.Camera)) }
            Button(
                onClick = { addPhoto { app.photoCapture.pickFromGallery() } },
                modifier = Modifier.weight(1f),
            ) { Text(t(Str.Album)) }
            if (app.isDemoMode) {
                Button(
                    onClick = { app.freeMoveCarFeature.addDemoPhoto() },
                    modifier = Modifier.weight(1f),
                ) { Text(t(Str.DemoPhoto)) }
            }
        }
        if (state.photoUrls.isNotEmpty()) {
            Button(
                onClick = { app.freeMoveCarFeature.clearPhotos() },
                modifier = Modifier.fillMaxWidth(),
            ) { Text(t(Str.ClearPhotos)) }
        }
        photoHint?.let {
            Text(it, color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.bodySmall)
        }
        state.photoUrls.forEach { url ->
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
            ) {
                Text(
                    text = url.substringAfterLast('/').ifBlank { url }.take(48),
                    style = MaterialTheme.typography.bodySmall,
                    modifier = Modifier.weight(1f),
                )
                TextButton(onClick = { app.freeMoveCarFeature.removePhoto(url) }) { Text(t(Str.Remove)) }
            }
        }
    }

    state.message?.let { Text(it) }
    state.errorMessage?.let { Text(it, color = MaterialTheme.colorScheme.error) }
}
