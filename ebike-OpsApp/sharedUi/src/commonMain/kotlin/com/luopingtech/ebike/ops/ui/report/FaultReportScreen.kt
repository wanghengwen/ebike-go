package com.luopingtech.ebike.ops.ui.report

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
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
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.RepairType
import com.luopingtech.ebike.ops.domain.report.RepairBodyParts
import com.luopingtech.ebike.ops.feature.report.ReportPage
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
            .background(MaterialTheme.colorScheme.background)
            .padding(16.dp),
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
        ) {
            Text(t(Str.FaultReportTitle), style = MaterialTheme.typography.headlineSmall)
            TextButton(onClick = {
                app.faultReportFeature.clear()
                onClose()
            }) { Text(t(Str.Close)) }
        }

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
        TextButton(onClick = { app.faultReportFeature.openHub() }) { Text("← ${t(Str.Back)}") }
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
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Button(
                onClick = { addPhoto { app.photoCapture.takePhoto("fault") } },
                modifier = Modifier.weight(1f),
            ) { Text(t(Str.Camera)) }
            Button(
                onClick = { addPhoto { app.photoCapture.pickFromGallery() } },
                modifier = Modifier.weight(1f),
            ) { Text(t(Str.Album)) }
            if (app.isDemoMode) {
                Button(
                    onClick = { app.faultReportFeature.addDemoPhoto() },
                    modifier = Modifier.weight(1f),
                ) { Text(t(Str.DemoPhoto)) }
            }
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
                TextButton(onClick = { app.faultReportFeature.removePhoto(url) }) {
                    Text(t(Str.Remove))
                }
            }
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

    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(8.dp),
        verticalAlignment = Alignment.Top,
    ) {
        Column(
            modifier = Modifier.weight(1f),
            verticalArrangement = Arrangement.spacedBy(4.dp),
        ) {
            leftTypes.forEach { type ->
                FilterChip(
                    selected = type.id in selectedIds,
                    onClick = { onToggle(type.id) },
                    label = { Text(type.name) },
                )
            }
        }
        Column(
            modifier = Modifier
                .width(56.dp)
                .padding(top = 24.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Text(
                text = "VEH",
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }
        Column(
            modifier = Modifier.weight(1f),
            verticalArrangement = Arrangement.spacedBy(4.dp),
            horizontalAlignment = Alignment.End,
        ) {
            rightTypes.forEach { type ->
                FilterChip(
                    selected = type.id in selectedIds,
                    onClick = { onToggle(type.id) },
                    label = { Text(type.name) },
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

