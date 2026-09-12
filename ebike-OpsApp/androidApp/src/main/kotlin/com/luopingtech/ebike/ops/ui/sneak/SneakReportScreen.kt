package com.luopingtech.ebike.ops.ui.sneak

import android.Manifest
import android.content.pm.PackageManager
import android.net.Uri
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
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
import com.luopingtech.ebike.ops.domain.model.SneakCheckResultLabels
import com.luopingtech.ebike.ops.feature.sneak.SneakPage
import com.luopingtech.ebike.ops.ui.common.createCachePhotoUri
import kotlinx.coroutines.launch

@Composable
fun SneakReportScreen(
    app: OpsApp,
    onClose: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.sneakReportFeature.state.collectAsState()
    val scope = rememberCoroutineScope()

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
            .padding(16.dp),
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
        ) {
            Text(t(Str.SneakReportTitle), style = MaterialTheme.typography.headlineSmall)
            TextButton(onClick = {
                app.sneakReportFeature.clear()
                onClose()
            }) { Text(t(Str.Close)) }
        }

        when (state.page) {
            SneakPage.Hub -> {
                Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                    Text(
                        text = t(Str.SneakReportHint),
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                    state.message?.let {
                        Text(it, color = MaterialTheme.colorScheme.primary)
                    }
                    Button(
                        onClick = {
                            app.sneakReportFeature.openSubmit()
                            scope.launch { app.sneakReportFeature.ensureTypes() }
                        },
                        modifier = Modifier.fillMaxWidth(),
                    ) { Text(t(Str.NewSneakReport)) }
                    Button(
                        onClick = { scope.launch { app.sneakReportFeature.openHistory() } },
                        modifier = Modifier.fillMaxWidth(),
                    ) { Text(t(Str.MySneakReports)) }
                }
            }
            SneakPage.Submit -> SubmitPane(app)
            SneakPage.History -> HistoryPane(app)
            SneakPage.Detail -> DetailPane(app)
        }

        state.errorMessage?.let {
            Spacer(modifier = Modifier.height(8.dp))
            Text(it, color = MaterialTheme.colorScheme.error)
        }
    }
}

@Composable
private fun SubmitPane(app: OpsApp) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.sneakReportFeature.state.collectAsState()
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
            app.sneakReportFeature.addPhotoUrl(uri.toString())
            photoHint = null
        } else {
            photoHint = t(Str.NoPhotoTaken)
        }
    }

    val galleryLauncher = rememberLauncherForActivityResult(
        ActivityResultContracts.GetContent(),
    ) { uri ->
        if (uri != null) {
            app.sneakReportFeature.addPhotoUrl(uri.toString())
            photoHint = null
        }
    }

    val cameraPermissionLauncher = rememberLauncherForActivityResult(
        ActivityResultContracts.RequestPermission(),
    ) { granted ->
        if (granted) {
            val uri = createCachePhotoUri(context, prefix = "sneak")
                ?: return@rememberLauncherForActivityResult
            pendingCameraUri = uri
            takePictureLauncher.launch(uri)
        } else {
            photoHint = t(Str.CameraPermissionRequired)
        }
    }

    LaunchedEffect(Unit) {
        app.sneakReportFeature.ensureTypes()
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState()),
        verticalArrangement = Arrangement.spacedBy(10.dp),
    ) {
        TextButton(onClick = { app.sneakReportFeature.openHub() }) { Text("← ${t(Str.Back)}") }
        OutlinedTextField(
            value = state.carId,
            onValueChange = {
                app.sneakReportFeature.setCarId(it)
                scope.launch { app.sneakReportFeature.lookupLastOrderIfReady() }
            },
            label = { Text(t(Str.VehicleId)) },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true,
        )
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            Button(
                onClick = {
                    scope.launch {
                        when (val scan = app.codeScanner.scanOnce()) {
                            is OpsResult.Ok -> {
                                val hosts = app.authFeature.state.value.runtimeConfig?.qrHosts.orEmpty()
                                if (app.sneakReportFeature.applyScanRaw(scan.value, hosts)) {
                                    app.sneakReportFeature.lookupLastOrderIfReady()
                                }
                            }
                            is OpsResult.Err -> photoHint = scan.error.message
                        }
                    }
                },
                modifier = Modifier.weight(1f),
            ) { Text(t(Str.ScanFillCarId)) }
        }
        state.lastOrder?.let { order ->
            Text(
                text = "${t(Str.SneakOrderId)} ${order.id}",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.primary,
            )
            Text(
                text = "${t(Str.SneakReportedUser)} ${order.userPhone.ifBlank { order.userPin }}",
                style = MaterialTheme.typography.bodySmall,
            )
        }
        state.message?.let { Text(it, color = MaterialTheme.colorScheme.primary) }

        Text(t(Str.SneakSelectType), style = MaterialTheme.typography.labelLarge)
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            state.types.forEach { type ->
                FilterChip(
                    selected = type.id in state.selectedTypeIds,
                    onClick = { app.sneakReportFeature.toggleType(type.id) },
                    label = { Text(type.name) },
                )
            }
        }
        if (state.types.any { it.isOther && it.id in state.selectedTypeIds }) {
            OutlinedTextField(
                value = state.otherType,
                onValueChange = { app.sneakReportFeature.setOtherType(it) },
                label = { Text(t(Str.SneakOtherType)) },
                modifier = Modifier.fillMaxWidth(),
                singleLine = true,
            )
        }
        OutlinedTextField(
            value = state.description,
            onValueChange = { app.sneakReportFeature.setDescription(it) },
            label = { Text(t(Str.SneakDescription)) },
            modifier = Modifier.fillMaxWidth(),
            minLines = 2,
        )
        Text(
            text = t(Str.SneakPhotosOptional),
            style = MaterialTheme.typography.labelMedium,
        )
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            Button(
                onClick = {
                    val granted = ContextCompat.checkSelfPermission(
                        context,
                        Manifest.permission.CAMERA,
                    ) == PackageManager.PERMISSION_GRANTED
                    if (granted) {
                        val uri = createCachePhotoUri(context, prefix = "sneak") ?: return@Button
                        pendingCameraUri = uri
                        takePictureLauncher.launch(uri)
                    } else {
                        cameraPermissionLauncher.launch(Manifest.permission.CAMERA)
                    }
                },
                enabled = state.photoUrls.size < 3,
                modifier = Modifier.weight(1f),
            ) { Text(t(Str.Camera)) }
            Button(
                onClick = { galleryLauncher.launch("image/*") },
                enabled = state.photoUrls.size < 3,
                modifier = Modifier.weight(1f),
            ) { Text(t(Str.Album)) }
            if (app.isDemoMode) {
                Button(
                    onClick = { app.sneakReportFeature.addDemoPhoto() },
                    enabled = state.photoUrls.size < 3,
                    modifier = Modifier.weight(1f),
                ) { Text(t(Str.DemoPhoto)) }
            }
        }
        state.photoUrls.forEach { url ->
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
            ) {
                Text(url.takeLast(28), style = MaterialTheme.typography.bodySmall)
                TextButton(onClick = { app.sneakReportFeature.removePhoto(url) }) {
                    Text(t(Str.Remove))
                }
            }
        }
        photoHint?.let { Text(it, color = MaterialTheme.colorScheme.error) }
        Button(
            onClick = { scope.launch { app.sneakReportFeature.submit() } },
            enabled = !state.loading,
            modifier = Modifier.fillMaxWidth(),
        ) { Text(if (state.loading) t(Str.Submitting) else t(Str.SubmitFault)) }
    }
}

@Composable
private fun HistoryPane(app: OpsApp) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.sneakReportFeature.state.collectAsState()

    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState()),
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        TextButton(onClick = { app.sneakReportFeature.openHub() }) { Text("← ${t(Str.Back)}") }
        if (state.loading) {
            Text(t(Str.LoadingEllipsis))
        }
        state.records.forEach { record ->
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .clickable { app.sneakReportFeature.openDetail(record) }
                    .padding(vertical = 6.dp),
            ) {
                Text(
                    text = "${record.id} · ${record.carId}",
                    style = MaterialTheme.typography.titleSmall,
                )
                Text(
                    text = "${t(Str.SneakReportedUser)} ${record.reportedUserPhone.ifBlank { record.reportedUserName }}",
                    style = MaterialTheme.typography.bodySmall,
                )
                Text(
                    text = "${t(SneakCheckResultLabels.strKey(record.checkResult))} · ${record.createdAt}",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
                if (record.remark.isNotBlank()) {
                    Text(
                        text = "${t(Str.SneakAuditRemark)} · ${record.remark}",
                        style = MaterialTheme.typography.bodySmall,
                    )
                }
            }
        }
        if (!state.loading && state.records.isEmpty()) {
            Text(t(Str.MySneakReports), color = MaterialTheme.colorScheme.onSurfaceVariant)
        }
    }
}

@Composable
private fun DetailPane(app: OpsApp) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.sneakReportFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    val record = state.selectedRecord ?: return

    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState()),
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        TextButton(onClick = { scope.launch { app.sneakReportFeature.openHistory() } }) {
            Text("← ${t(Str.Back)}")
        }
        Text(record.id, style = MaterialTheme.typography.titleMedium)
        Text("${t(Str.VehicleId)} · ${record.carId}")
        Text("${t(Str.SneakOrderId)} · ${record.itinId.ifBlank { "-" }}")
        Text(
            "${t(Str.SneakReportedUser)} · ${
                record.reportedUserPhone.ifBlank { record.reportedUserName.ifBlank { record.reportedUserPin } }
            }",
        )
        Text(t(SneakCheckResultLabels.strKey(record.checkResult)))
        if (record.typeLabels.isNotEmpty() || record.otherType.isNotBlank()) {
            Text(
                (record.typeLabels + listOfNotNull(record.otherType.takeIf { it.isNotBlank() }))
                    .joinToString(" · "),
            )
        }
        if (record.description.isNotBlank()) {
            Text(record.description)
        }
        if (record.remark.isNotBlank()) {
            Text("${t(Str.SneakAuditRemark)} · ${record.remark}")
        }
        record.photoUrls.forEach { url ->
            Text(url, style = MaterialTheme.typography.bodySmall)
        }
        state.message?.let { Text(it, color = MaterialTheme.colorScheme.primary) }
        if (record.canCancel) {
            Button(
                onClick = { scope.launch { app.sneakReportFeature.cancelSelected() } },
                enabled = !state.loading,
                modifier = Modifier.fillMaxWidth(),
            ) { Text(t(Str.SneakCancel)) }
        }
    }
}
