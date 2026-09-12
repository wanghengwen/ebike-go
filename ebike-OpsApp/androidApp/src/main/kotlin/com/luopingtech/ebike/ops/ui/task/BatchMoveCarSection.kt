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
import androidx.compose.runtime.LaunchedEffect
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
import com.luopingtech.ebike.ops.ui.common.createCachePhotoUri
import kotlinx.coroutines.launch

@Composable
fun BatchMoveCarSection(
    app: OpsApp,
    parentTaskId: String?,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.batchMoveCarFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    val context = LocalContext.current
    var pendingCameraUri by remember { mutableStateOf<Uri?>(null) }
    var photoHint by remember { mutableStateOf<String?>(null) }

    LaunchedEffect(parentTaskId) {
        if (!parentTaskId.isNullOrBlank()) {
            app.batchMoveCarFeature.load(parentTaskId)
        }
    }

    val takePictureLauncher = rememberLauncherForActivityResult(
        ActivityResultContracts.TakePicture(),
    ) { ok ->
        val uri = pendingCameraUri
        pendingCameraUri = null
        if (ok && uri != null) {
            app.batchMoveCarFeature.addPhotoUrl(uri.toString())
            photoHint = null
        } else {
            photoHint = t(Str.NoPhotoTaken)
        }
    }

    val galleryLauncher = rememberLauncherForActivityResult(
        ActivityResultContracts.GetContent(),
    ) { uri ->
        if (uri != null) {
            app.batchMoveCarFeature.addPhotoUrl(uri.toString())
            photoHint = null
        }
    }

    val cameraPermissionLauncher = rememberLauncherForActivityResult(
        ActivityResultContracts.RequestPermission(),
    ) { granted ->
        if (granted) {
            val uri = createCachePhotoUri(context, prefix = "batchmove")
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
            val uri = createCachePhotoUri(context, prefix = "batchmove") ?: return
            pendingCameraUri = uri
            takePictureLauncher.launch(uri)
        } else {
            cameraPermissionLauncher.launch(Manifest.permission.CAMERA)
        }
    }

    val selectedCount = state.selectedTaskIds.size

    Text(
        text = t(Str.BatchMoveHint),
        style = MaterialTheme.typography.bodySmall,
        color = MaterialTheme.colorScheme.onSurfaceVariant,
    )
    Text(
        text = parentTaskId?.let { t(Str.BatchMoveParentLabel, it) } ?: t(Str.BatchMoveNeedParent),
        style = MaterialTheme.typography.bodyMedium,
    )
    Button(
        onClick = {
            scope.launch {
                parentTaskId?.let { app.batchMoveCarFeature.load(it) }
            }
        },
        enabled = !state.loading && !parentTaskId.isNullOrBlank(),
        modifier = Modifier.fillMaxWidth(),
    ) { Text(if (state.loading) t(Str.LoadingEllipsis) else t(Str.Refresh)) }

    Row(horizontalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.fillMaxWidth()) {
        TextButton(
            onClick = { app.batchMoveCarFeature.selectPending() },
            enabled = state.children.any { !it.isFinished },
        ) { Text(t(Str.BatchMoveSelectPending)) }
        TextButton(
            onClick = { app.batchMoveCarFeature.clearSelection() },
            enabled = state.selectedTaskIds.isNotEmpty(),
        ) { Text(t(Str.FreeMoveClearSelect)) }
    }

    state.children.forEach { child ->
        val selected = child.taskId in state.selectedTaskIds
        val mark = when {
            child.isFinished -> "✓"
            selected -> "●"
            else -> "○"
        }
        Text(
            text = "$mark ${child.carId} · ${child.restBattery}% · ${child.taskId}",
            modifier = Modifier
                .fillMaxWidth()
                .clickable(enabled = !child.isFinished) {
                    app.batchMoveCarFeature.toggleSelect(child.taskId)
                }
                .padding(vertical = 4.dp),
            color = when {
                child.isFinished -> MaterialTheme.colorScheme.onSurfaceVariant
                selected -> MaterialTheme.colorScheme.primary
                else -> MaterialTheme.colorScheme.onSurface
            },
        )
    }

    Row(horizontalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.fillMaxWidth()) {
        Button(
            onClick = { scope.launch { app.batchMoveCarFeature.startSelected() } },
            enabled = !state.loading && selectedCount > 0,
            modifier = Modifier.weight(1f),
        ) { Text(t(Str.BatchMoveStart, selectedCount)) }
        Button(
            onClick = { scope.launch { app.batchMoveCarFeature.finishSelected() } },
            enabled = !state.loading && selectedCount > 0 &&
                (
                    !state.needPhotograph ||
                        (state.photoUrls.isNotEmpty() && state.remark.isNotBlank())
                    ),
            modifier = Modifier.weight(1f),
        ) {
            Text(
                when {
                    state.needPhotograph -> t(Str.BatchMoveFinishPhoto, selectedCount)
                    else -> t(Str.BatchMoveFinish, selectedCount)
                },
            )
        }
    }

    Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Text(
            text = when {
                state.needPhotograph -> t(Str.PhotoAuditCount, state.photoUrls.size)
                state.photoUrls.isNotEmpty() -> t(Str.PhotoAuditCount, state.photoUrls.size)
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
                onValueChange = { app.batchMoveCarFeature.setRemark(it) },
                label = {
                    Text(
                        if (state.needPhotograph) t(Str.PhotoRemarkHint) else t(Str.PhotoRemarkOptional),
                    )
                },
                modifier = Modifier.fillMaxWidth(),
                minLines = 2,
            )
        }
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Button(onClick = { launchCamera() }, modifier = Modifier.weight(1f)) {
                Text(t(Str.Camera))
            }
            Button(onClick = { galleryLauncher.launch("image/*") }, modifier = Modifier.weight(1f)) {
                Text(t(Str.Album))
            }
            if (app.isDemoMode) {
                Button(
                    onClick = { app.batchMoveCarFeature.addDemoPhoto() },
                    modifier = Modifier.weight(1f),
                ) { Text(t(Str.DemoPhoto)) }
            }
        }
        if (state.photoUrls.isNotEmpty()) {
            Button(
                onClick = { app.batchMoveCarFeature.clearPhotos() },
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
                TextButton(onClick = { app.batchMoveCarFeature.removePhoto(url) }) {
                    Text(t(Str.Remove))
                }
            }
        }
    }

    state.message?.let { Text(it) }
    state.errorMessage?.let { Text(it, color = MaterialTheme.colorScheme.error) }
}
