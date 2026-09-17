package com.luopingtech.ebike.ops.ui.report

import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.Checkbox
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
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.unit.dp
import androidx.compose.ui.window.Dialog
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.RepairType
import com.luopingtech.ebike.ops.domain.report.RepairBodyParts
import com.luopingtech.ebike.ops.feature.report.ReportPage
import com.luopingtech.ebike.ops.ui.analysis.VcdTopBar
import com.luopingtech.ebike.ops.ui.icons.OpsIcon
import com.luopingtech.ebike.ops.ui.icons.painterResource
import com.luopingtech.ebike.ops.ui.media.LocalPathImage
import kotlinx.coroutines.launch

@Composable
fun FaultReportScreen(
    app: OpsApp,
    onClose: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.faultReportFeature.state.collectAsState()
    val scope = rememberCoroutineScope()

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.White),
    ) {
        VcdTopBar(
            title = t(Str.FaultReportTitle),
            onBack = {
                when (state.page) {
                    ReportPage.Submit, ReportPage.Hub -> {
                        app.faultReportFeature.clear()
                        onClose()
                    }
                    ReportPage.History -> app.faultReportFeature.openHub()
                    ReportPage.Detail -> scope.launch { app.faultReportFeature.openHistory() }
                }
            },
        )
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(16.dp),
        ) {
            when (state.page) {
                ReportPage.Hub -> {
                    Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                        Text(
                            text = t(Str.FaultReportHint),
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                        )
                        state.message?.let {
                            Text(it, color = MaterialTheme.colorScheme.primary)
                        }
                        Button(
                            onClick = {
                                app.faultReportFeature.openSubmit()
                                scope.launch { app.faultReportFeature.ensureRepairTypes() }
                            },
                            modifier = Modifier.fillMaxWidth(),
                        ) { Text(t(Str.NewFaultReport)) }
                        Button(
                            onClick = { scope.launch { app.faultReportFeature.openHistory() } },
                            modifier = Modifier.fillMaxWidth(),
                        ) { Text(t(Str.MyReports)) }
                    }
                }
                ReportPage.Submit -> SubmitPane(app)
                ReportPage.History -> HistoryPane(app)
                ReportPage.Detail -> DetailPane(app)
            }

            state.errorMessage?.let {
                Spacer(modifier = Modifier.height(8.dp))
                Text(it, color = MaterialTheme.colorScheme.error)
            }
        }
    }
}

@Composable
private fun SubmitPane(app: OpsApp) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.faultReportFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    var photoHint by remember { mutableStateOf<String?>(null) }

    /** 相机与相册只差取图那一步，权限和取消都由 [PhotoCapture] 实现方吞掉。 */
    fun addPhoto(take: suspend () -> OpsResult<String>) {
        scope.launch {
            when (val shot = take()) {
                is OpsResult.Ok -> {
                    app.faultReportFeature.addPhotoUrl(shot.value)
                    photoHint = null
                }
                is OpsResult.Err -> photoHint = shot.error.message
            }
        }
    }

    LaunchedEffect(Unit) {
        app.faultReportFeature.ensureRepairTypes()
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState()),
        verticalArrangement = Arrangement.spacedBy(10.dp),
    ) {
        OutlinedTextField(
            value = state.carId,
            onValueChange = {
                app.faultReportFeature.setCarId(it)
                scope.launch { app.faultReportFeature.lookupVehicleIfReady() }
            },
            label = { Text(t(Str.VehicleId)) },
            supportingText = { Text(t(Str.FaultCarIdHint)) },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true,
        )
        state.vehicleSummary?.let {
            Text(it, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.primary)
        }
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            Button(
                onClick = {
                    scope.launch {
                        when (val scan = app.codeScanner.scanOnce()) {
                            is OpsResult.Ok -> {
                                val hosts = app.authFeature.state.value.runtimeConfig?.qrHosts.orEmpty()
                                if (app.faultReportFeature.applyScanRaw(scan.value, hosts)) {
                                    app.faultReportFeature.lookupVehicleIfReady()
                                }
                            }
                            is OpsResult.Err -> photoHint = scan.error.message
                        }
                    }
                },
                modifier = Modifier.weight(1f),
            ) { Text(t(Str.ScanFillCarId)) }
        }
        Text(t(Str.FaultType), style = MaterialTheme.typography.labelLarge)
        RepairTypePartPicker(
            types = state.repairTypes,
            selectedIds = state.selectedTypeIds,
            vehicleModel = state.vehicleModel,
            onToggle = { app.faultReportFeature.toggleType(it) },
            otherTitle = t(Str.FaultOtherTypes),
        )
        OutlinedTextField(
            value = state.fixReason,
            onValueChange = { app.faultReportFeature.setFixReason(it) },
            label = { Text(t(Str.FaultDesc)) },
            supportingText = { Text("${state.fixReason.length}/100") },
            modifier = Modifier.fillMaxWidth(),
        )
        Text(t(Str.IzStop), style = MaterialTheme.typography.labelLarge)
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            FilterChip(
                selected = state.izStop == true,
                onClick = { app.faultReportFeature.setIzStop(true) },
                label = { Text(t(Str.StopYes)) },
            )
            FilterChip(
                selected = state.izStop == false,
                onClick = { app.faultReportFeature.setIzStop(false) },
                label = { Text(t(Str.StopNo)) },
            )
        }
        Text(
            text = t(Str.PhotosRequired),
            style = MaterialTheme.typography.labelLarge,
        )
        FaultPhotoGrid(
            photos = state.photoUrls,
            canAdd = state.photoUrls.size < 3,
            onAddCamera = { addPhoto { app.photoCapture.takePhoto("fault") } },
            onAddAlbum = { addPhoto { app.photoCapture.pickFromGallery() } },
            onAddDemo = if (app.isDemoMode) {
                { app.faultReportFeature.addDemoPhoto() }
            } else {
                null
            },
            onRemove = { app.faultReportFeature.removePhoto(it) },
            demoLabel = t(Str.DemoPhoto),
            cameraLabel = t(Str.Camera),
            albumLabel = t(Str.Album),
        )
        photoHint?.let {
            Text(it, color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.bodySmall)
        }
        Button(
            onClick = { scope.launch { app.faultReportFeature.submit() } },
            enabled = !state.loading,
            modifier = Modifier.fillMaxWidth(),
        ) {
            Text(if (state.loading) t(Str.Submitting) else t(Str.SubmitFault))
        }
    }
}

@Composable
private fun HistoryPane(app: OpsApp) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.faultReportFeature.state.collectAsState()
    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState()),
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        TextButton(onClick = { app.faultReportFeature.openHub() }) { Text("← ${t(Str.Back)}") }
        Text(t(Str.MyReports), style = MaterialTheme.typography.titleMedium)
        if (state.loading) Text(t(Str.Loading))
        state.records.forEach { record ->
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .clickable { app.faultReportFeature.openDetail(record) }
                    .background(MaterialTheme.colorScheme.surfaceVariant)
                    .padding(12.dp),
            ) {
                Text("${record.carId} · ${record.statusLabel}", style = MaterialTheme.typography.titleSmall)
                Text(
                    "${record.typeNames.joinToString()} · ${record.fixReason}",
                    style = MaterialTheme.typography.bodySmall,
                )
                Text(record.createdAt, style = MaterialTheme.typography.labelSmall)
            }
        }
        if (!state.loading && state.records.isEmpty()) {
            Text(t(Str.NoRecords), color = MaterialTheme.colorScheme.onSurfaceVariant)
        }
    }
}

@Composable
private fun DetailPane(app: OpsApp) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.faultReportFeature.state.collectAsState()
    val record = state.selectedRecord
    val scope = rememberCoroutineScope()
    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState()),
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        TextButton(
            onClick = {
                scope.launch { app.faultReportFeature.openHistory() }
            },
        ) { Text("← ${t(Str.Back)}") }
        if (record == null) {
            Text(t(Str.NoDetail))
            return
        }
        Text(record.carId, style = MaterialTheme.typography.titleMedium)
        Text(record.statusLabel)
        Text(t(Str.TypeColon, record.typeNames.joinToString()))
        Text(t(Str.DescColon, record.fixReason))
        Text(t(Str.StopColon, if (record.izStop) t(Str.Yes) else t(Str.No)))
        Text(t(Str.TimeColon, record.createdAt))
        Text(t(Str.PhotosColon))
        record.photoUrls.forEach { Text("· $it", style = MaterialTheme.typography.bodySmall) }
    }
}

@OptIn(ExperimentalLayoutApi::class)
@Composable
private fun RepairTypePartPicker(
    types: List<RepairType>,
    selectedIds: Set<String>,
    vehicleModel: String,
    onToggle: (String) -> Unit,
    otherTitle: String,
) {
    val byId = remember(types) { types.associateBy { it.id } }
    val leftTypes = RepairBodyParts.left.mapNotNull { part ->
        val matched = byId[part.id] ?: types.firstOrNull { it.name == part.titleZh }
        matched?.copy(name = matched.name.ifBlank { part.titleZh })
    }
    val rightTypes = RepairBodyParts.right.mapNotNull { part ->
        val matched = byId[part.id] ?: types.firstOrNull { it.name == part.titleZh }
        matched?.copy(name = matched.name.ifBlank { part.titleZh })
    }
    val bodyIds = (leftTypes + rightTypes).map { it.id }.toSet()
    val others = types.filter { it.id !in bodyIds && it.id !in RepairBodyParts.allIds }
    val bodyIcon = when (vehicleModel.trim()) {
        "0" -> OpsIcon.VehicleRepairV1
        "2" -> OpsIcon.VehicleRepairV3
        else -> OpsIcon.VehicleRepairV2
    }
    // Legacy VehicleRepairPartView: image 214×264dp, cellHeight = 264/7
    val cellHeight = (264f / 7f).dp

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(top = 15.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Column(
            modifier = Modifier.weight(1f),
            horizontalAlignment = Alignment.End,
        ) {
            leftTypes.forEach { type ->
                RepairPartRow(
                    name = type.name,
                    selected = type.id in selectedIds,
                    checkboxOnEnd = true,
                    modifier = Modifier
                        .height(cellHeight)
                        .padding(end = 4.dp),
                    onToggle = { onToggle(type.id) },
                )
            }
        }
        Image(
            painter = painterResource(bodyIcon),
            contentDescription = null,
            modifier = Modifier
                .width(214.dp)
                .height(264.dp),
            contentScale = ContentScale.Fit,
        )
        Column(
            modifier = Modifier.weight(1f),
            horizontalAlignment = Alignment.Start,
        ) {
            rightTypes.forEach { type ->
                RepairPartRow(
                    name = type.name,
                    selected = type.id in selectedIds,
                    checkboxOnEnd = false,
                    modifier = Modifier
                        .height(cellHeight)
                        .padding(start = 4.dp),
                    onToggle = { onToggle(type.id) },
                )
            }
        }
    }

    if (others.isNotEmpty()) {
        Text(otherTitle, style = MaterialTheme.typography.labelMedium)
        FlowRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            others.forEach { type ->
                FilterChip(
                    selected = type.id in selectedIds,
                    onClick = { onToggle(type.id) },
                    label = { Text(type.name) },
                )
            }
        }
    }

    // Fallback when API type ids do not match legacy body-part codes.
    if (leftTypes.isEmpty() && rightTypes.isEmpty() && types.isNotEmpty()) {
        FlowRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            types.forEach { type ->
                FilterChip(
                    selected = type.id in selectedIds,
                    onClick = { onToggle(type.id) },
                    label = { Text(type.name) },
                )
            }
        }
    }
}

/** Aligns with adapter_vehicle_part_left/right_item: label + 24dp checkbox. */
@Composable
private fun RepairPartRow(
    name: String,
    selected: Boolean,
    checkboxOnEnd: Boolean,
    modifier: Modifier = Modifier,
    onToggle: () -> Unit,
) {
    Row(
        modifier = modifier.clickable(onClick = onToggle),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = if (checkboxOnEnd) Arrangement.End else Arrangement.Start,
    ) {
        if (!checkboxOnEnd) {
            Checkbox(
                checked = selected,
                onCheckedChange = null,
                modifier = Modifier.size(24.dp),
            )
        }
        Text(
            text = name,
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurface,
        )
        if (checkboxOnEnd) {
            Checkbox(
                checked = selected,
                onCheckedChange = null,
                modifier = Modifier.size(24.dp),
            )
        }
    }
}

@Composable
private fun FaultPhotoGrid(
    photos: List<String>,
    canAdd: Boolean,
    onAddCamera: () -> Unit,
    onAddAlbum: () -> Unit,
    onAddDemo: (() -> Unit)?,
    onRemove: (String) -> Unit,
    demoLabel: String,
    cameraLabel: String,
    albumLabel: String,
) {
    var showPicker by remember { mutableStateOf(false) }
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        if (canAdd) {
            Box(
                modifier = Modifier
                    .size(80.dp)
                    .clickable { showPicker = true },
            ) {
                Image(
                    painter = painterResource(OpsIcon.PhotoReplaceHolder),
                    contentDescription = null,
                    modifier = Modifier.fillMaxSize(),
                    contentScale = ContentScale.Crop,
                )
            }
        }
        photos.forEach { path ->
            Box(modifier = Modifier.size(80.dp)) {
                LocalPathImage(path = path, modifier = Modifier.fillMaxSize())
                Image(
                    painter = painterResource(OpsIcon.PhotoDelete),
                    contentDescription = null,
                    modifier = Modifier
                        .align(Alignment.TopEnd)
                        .size(18.dp)
                        .clickable { onRemove(path) },
                )
            }
        }
    }
    if (showPicker) {
        Dialog(onDismissRequest = { showPicker = false }) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(Color.White, MaterialTheme.shapes.medium)
                    .padding(16.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                TextButton(onClick = {
                    showPicker = false
                    onAddCamera()
                }) { Text(cameraLabel) }
                TextButton(onClick = {
                    showPicker = false
                    onAddAlbum()
                }) { Text(albumLabel) }
                onAddDemo?.let { demo ->
                    TextButton(onClick = {
                        showPicker = false
                        demo()
                    }) { Text(demoLabel) }
                }
            }
        }
    }
}

