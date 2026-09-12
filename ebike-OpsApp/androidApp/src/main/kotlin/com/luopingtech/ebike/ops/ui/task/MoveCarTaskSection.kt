package com.luopingtech.ebike.ops.ui.task

import android.Manifest
import android.content.pm.PackageManager
import android.net.Uri
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.core.content.ContextCompat
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.OpsTask
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.ui.common.createCachePhotoUri
import kotlinx.coroutines.launch

@Composable
fun MoveCarTaskSection(
    app: OpsApp,
    currentArea: ServiceArea?,
    selectedCard: @Composable (OpsTask?) -> Unit,
    onOpenBatchMove: (OpsTask) -> Unit = {},
    onOpenFreeMove: () -> Unit = {},
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val taskState by app.moveCarTaskFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    val context = LocalContext.current
    var pendingCameraUri by remember { mutableStateOf<Uri?>(null) }
    var photoHint by remember { mutableStateOf<String?>(null) }

    val takePictureLauncher = rememberLauncherForActivityResult(
        ActivityResultContracts.TakePicture(),
    ) { ok ->
        val uri = pendingCameraUri
        pendingCameraUri = null
        if (ok && uri != null) {
            app.moveCarTaskFeature.addPhotoUrl(uri.toString())
            photoHint = null
        } else {
            photoHint = t(Str.NoPhotoTaken)
        }
    }

    val galleryLauncher = rememberLauncherForActivityResult(
        ActivityResultContracts.GetContent(),
    ) { uri ->
        if (uri != null) {
            app.moveCarTaskFeature.addPhotoUrl(uri.toString())
            photoHint = null
        }
    }

    val cameraPermissionLauncher = rememberLauncherForActivityResult(
        ActivityResultContracts.RequestPermission(),
    ) { granted ->
        if (granted) {
            val uri = createCachePhotoUri(context, prefix = "movecar")
                ?: return@rememberLauncherForActivityResult
            pendingCameraUri = uri
            takePictureLauncher.launch(uri)
        } else {
            photoHint = t(Str.CameraPermissionRequired)
        }
    }

    fun launchCamera() {
        val granted = ContextCompat.checkSelfPermission(
            context,
            Manifest.permission.CAMERA,
        ) == PackageManager.PERMISSION_GRANTED
        if (granted) {
            val uri = createCachePhotoUri(context, prefix = "movecar") ?: return
            pendingCameraUri = uri
            takePictureLauncher.launch(uri)
        } else {
            cameraPermissionLauncher.launch(Manifest.permission.CAMERA)
        }
    }

    Text(
        text = t(Str.MoveCarFlowHint),
        style = MaterialTheme.typography.bodySmall,
        color = MaterialTheme.colorScheme.onSurfaceVariant,
    )
    TextButton(
        onClick = onOpenFreeMove,
        modifier = Modifier.fillMaxWidth(),
    ) { Text(t(Str.FreeMoveShort)) }
    Button(
        onClick = { scope.launch { app.moveCarTaskFeature.load(currentArea) } },
        enabled = !taskState.loading,
        modifier = Modifier.fillMaxWidth(),
    ) { Text(if (taskState.loading) t(Str.LoadingEllipsis) else t(Str.Refresh)) }

    Button(
        onClick = {
            scope.launch {
                when (val scan = app.codeScanner.scanOnce()) {
                    is OpsResult.Ok -> {
                        val hosts = app.authFeature.state.value.runtimeConfig?.qrHosts.orEmpty()
                        app.moveCarTaskFeature.selectByScanRaw(scan.value, hosts)
                    }
                    is OpsResult.Err -> photoHint = scan.error.message
                }
            }
        },
        enabled = !taskState.loading && taskState.tasks.isNotEmpty(),
        modifier = Modifier.fillMaxWidth(),
    ) { Text(t(Str.ScanSelectCar)) }

    selectedCard(taskState.selected)

    if (taskState.selected?.isManMadeBatch == true) {
        Button(
            onClick = { taskState.selected?.let(onOpenBatchMove) },
            enabled = !taskState.loading,
            modifier = Modifier.fillMaxWidth(),
        ) { Text(t(Str.BatchMoveOpen)) }
    }

    var taskQuery by remember { mutableStateOf("") }
    OutlinedTextField(
        value = taskQuery,
        onValueChange = { taskQuery = it },
        label = { Text(t(Str.ListSearchHint)) },
        modifier = Modifier.fillMaxWidth(),
        singleLine = true,
    )
    val filteredTasks = remember(taskState.tasks, taskQuery) {
        val q = taskQuery.trim()
        if (q.isEmpty()) taskState.tasks
        else taskState.tasks.filter {
            it.carId.contains(q, ignoreCase = true) ||
                it.imei.contains(q, ignoreCase = true) ||
                it.batchParentId.contains(q, ignoreCase = true)
        }
    }
    filteredTasks.forEach { task ->
        val selected = task.id == taskState.selectedTaskId
        val label = if (task.isManMadeBatch) {
            t(Str.BatchMoveParentLabel, task.batchParentId)
        } else {
            task.carId
        }
        Text(
            text = "${if (selected) "●" else "○"} $label · ${task.stateLabel}",
            modifier = Modifier
                .fillMaxWidth()
                .clickable { app.moveCarTaskFeature.selectTask(task.id) }
                .padding(vertical = 4.dp),
            color = if (selected) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.onSurface,
        )
    }
    val isBatch = taskState.selected?.isManMadeBatch == true
    Row(horizontalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.fillMaxWidth()) {
        Button(
            onClick = { scope.launch { app.moveCarTaskFeature.claimSelected() } },
            enabled = !taskState.loading && !isBatch && taskState.selected?.state == 0,
            modifier = Modifier.weight(1f),
        ) { Text(t(Str.Claim)) }
        Button(
            onClick = { scope.launch { app.moveCarTaskFeature.startSelected() } },
            enabled = !taskState.loading && !isBatch &&
                taskState.selected != null &&
                taskState.selected!!.state != 2,
            modifier = Modifier.weight(1f),
        ) { Text(t(Str.Start)) }
        Button(
            onClick = { scope.launch { app.moveCarTaskFeature.finishSelected() } },
            enabled = !taskState.loading && !isBatch && taskState.selected != null &&
                taskState.selected!!.state != 2 &&
                (
                    !taskState.needPhotograph ||
                        (taskState.photoUrls.isNotEmpty() && taskState.remark.isNotBlank())
                    ),
            modifier = Modifier.weight(1f),
        ) {
            Text(
                when {
                    taskState.loading -> t(Str.LoadingEllipsis)
                    taskState.needPhotograph -> t(Str.SubmitPhotoFinish)
                    else -> t(Str.Finish)
                },
            )
        }
    }
    TaskAuditResultSection(app = app, task = taskState.selected)
    Button(
        onClick = { scope.launch { app.moveCarTaskFeature.refreshArrival() } },
        enabled = !taskState.loading && taskState.selected != null && taskState.selected?.state != 2,
        modifier = Modifier.fillMaxWidth(),
    ) {
        val dist = taskState.distanceMeters?.toInt()
        Text(
            when {
                taskState.arrived && dist != null -> t(Str.ArrivalCheckOk, dist)
                dist != null -> t(Str.ArrivalCheckDist, dist)
                else -> t(Str.ArrivalCheck)
            },
        )
    }

    if (taskState.selected != null) {
        Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
            Text(
                text = when {
                    taskState.needPhotograph ->
                        t(Str.PhotoAuditCount, taskState.photoUrls.size)
                    taskState.photoUrls.isNotEmpty() ->
                        t(Str.PhotoAuditCount, taskState.photoUrls.size)
                    else -> t(Str.PhotoOptionalHint)
                },
                style = MaterialTheme.typography.bodySmall,
                color = if (taskState.needPhotograph) {
                    MaterialTheme.colorScheme.tertiary
                } else {
                    MaterialTheme.colorScheme.onSurfaceVariant
                },
            )
            if (taskState.needPhotograph || taskState.photoUrls.isNotEmpty()) {
                OutlinedTextField(
                    value = taskState.remark,
                    onValueChange = { app.moveCarTaskFeature.setRemark(it) },
                    label = {
                        Text(
                            if (taskState.needPhotograph) t(Str.PhotoRemarkHint) else t(Str.PhotoRemarkOptional),
                        )
                    },
                    modifier = Modifier.fillMaxWidth(),
                    singleLine = false,
                    minLines = 2,
                )
            }
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                Button(
                    onClick = { launchCamera() },
                    modifier = Modifier.weight(1f),
                ) { Text(t(Str.Camera)) }
                Button(
                    onClick = { galleryLauncher.launch("image/*") },
                    modifier = Modifier.weight(1f),
                ) { Text(t(Str.Album)) }
                if (app.isDemoMode) {
                    Button(
                        onClick = { app.moveCarTaskFeature.addDemoPhoto() },
                        modifier = Modifier.weight(1f),
                    ) { Text(t(Str.DemoPhoto)) }
                }
            }
            if (taskState.photoUrls.isNotEmpty()) {
                Button(
                    onClick = { app.moveCarTaskFeature.clearPhotos() },
                    modifier = Modifier.fillMaxWidth(),
                ) { Text(t(Str.ClearPhotos)) }
            }
            photoHint?.let {
                Text(it, color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.bodySmall)
            }
            taskState.photoUrls.forEach { url ->
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                ) {
                    Text(
                        text = url.substringAfterLast('/').ifBlank { url }.take(48),
                        style = MaterialTheme.typography.bodySmall,
                        modifier = Modifier.weight(1f),
                    )
                    TextButton(onClick = { app.moveCarTaskFeature.removePhoto(url) }) {
                        Text(t(Str.Remove))
                    }
                }
            }
        }
    }

    Button(
        onClick = { scope.launch { app.moveCarTaskFeature.ringSelected() } },
        enabled = !taskState.loading && taskState.selected != null,
        modifier = Modifier.fillMaxWidth(),
    ) { Text(t(Str.Ring)) }
    taskState.message?.let { Text(it) }
    taskState.errorMessage?.let { Text(it, color = MaterialTheme.colorScheme.error) }
}
