package com.luopingtech.ebike.ops.ui.production

import androidx.compose.foundation.background
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
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.DetectStepStatus
import com.luopingtech.ebike.ops.domain.permission.OpsPermissions
import com.luopingtech.ebike.ops.feature.production.DetectChannelTab
import com.luopingtech.ebike.ops.feature.production.ProductionPage
import com.luopingtech.ebike.ops.feature.production.ShelfMode
import kotlinx.coroutines.launch

@Composable
fun ProductionScreen(
    app: OpsApp,
    permissions: OpsPermissions,
    onClose: () -> Unit,
    onLocateOnMap: () -> Unit = {},
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.productionFeature.state.collectAsState()
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
            val title = when (state.page) {
                ProductionPage.Detect, ProductionPage.Overload -> t(Str.ProductionDetect)
                ProductionPage.Bind -> t(Str.ProductionBind)
                ProductionPage.Shelves -> t(Str.ProductionShelves)
                ProductionPage.Hub -> t(Str.ProductionField)
            }
            Text(title, style = MaterialTheme.typography.headlineSmall)
            TextButton(onClick = {
                app.productionFeature.clear()
                onClose()
            }) {
                Text(t(Str.Close))
            }
        }

        when (state.page) {
            ProductionPage.Hub -> {
                Column(
                    modifier = Modifier
                        .fillMaxSize()
                        .verticalScroll(rememberScrollState()),
                    verticalArrangement = Arrangement.spacedBy(10.dp),
                ) {
                    Text(
                        text = t(Str.ProductionHint),
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                    state.message?.let {
                        Text(it, color = MaterialTheme.colorScheme.primary)
                    }
                    if (permissions.showProductionDetect) {
                        Button(
                            onClick = { app.productionFeature.openDetect() },
                            modifier = Modifier.fillMaxWidth(),
                        ) { Text(t(Str.ProductionDetect)) }
                    }
                    if (permissions.showProductionBind) {
                        Button(
                            onClick = { app.productionFeature.openBind() },
                            modifier = Modifier.fillMaxWidth(),
                        ) { Text(t(Str.ProductionBind)) }
                    }
                    if (permissions.showProductionShelves) {
                        Button(
                            onClick = { app.productionFeature.openShelves(ShelfMode.PutOn) },
                            modifier = Modifier.fillMaxWidth(),
                        ) { Text(t(Str.ShelfPutOn)) }
                        Button(
                            onClick = { app.productionFeature.openShelves(ShelfMode.PullOff) },
                            modifier = Modifier.fillMaxWidth(),
                        ) { Text(t(Str.ShelfPullOff)) }
                    }
                }
            }
            ProductionPage.Detect -> DetectPane(app, onBack = onClose, onLocateOnMap = onLocateOnMap)
            ProductionPage.Bind -> BindPane(app, onBack = onClose)
            ProductionPage.Shelves -> ShelvesPane(app, onBack = onClose)
            ProductionPage.Overload -> OverloadPane(app)
        }

        state.errorMessage?.let {
            Spacer(modifier = Modifier.height(8.dp))
            Text(it, color = MaterialTheme.colorScheme.error)
        }
    }
}

@Composable
private fun DetectPane(
    app: OpsApp,
    onBack: () -> Unit,
    onLocateOnMap: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.productionFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState()),
        verticalArrangement = Arrangement.spacedBy(10.dp),
    ) {
        TextButton(onClick = {
            app.productionFeature.clear()
            onBack()
        }) { Text("← ${t(Str.Back)}") }
        Text(t(Str.ProductionDetect), style = MaterialTheme.typography.titleMedium)
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            FilterChip(
                selected = state.detectChannel == DetectChannelTab.Network,
                onClick = { app.productionFeature.setDetectChannel(DetectChannelTab.Network) },
                label = { Text(t(Str.DetectTabNetwork)) },
            )
            FilterChip(
                selected = state.detectChannel == DetectChannelTab.Bluetooth,
                onClick = { app.productionFeature.setDetectChannel(DetectChannelTab.Bluetooth) },
                label = { Text(t(Str.DetectTabBluetooth)) },
            )
        }
        if (state.detectChannel == DetectChannelTab.Bluetooth && !state.bleAvailable) {
            Text(
                t(Str.DetectBleUnavailableHint),
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.tertiary,
            )
        }
        OutlinedTextField(
            value = state.carId,
            onValueChange = { app.productionFeature.setCarId(it) },
            label = { Text(t(Str.VehicleId)) },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true,
        )
        OutlinedTextField(
            value = state.imei,
            onValueChange = { app.productionFeature.setImei(it) },
            label = { Text(t(Str.ImeiLabel)) },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true,
        )
        Button(
            onClick = {
                scope.launch {
                    when (val scan = app.codeScanner.scanOnce()) {
                        is OpsResult.Ok -> {
                            val hosts = app.authFeature.state.value.runtimeConfig?.qrHosts.orEmpty()
                            app.productionFeature.applyScanRaw(scan.value, hosts)
                        }
                        is OpsResult.Err -> Unit
                    }
                }
            },
            modifier = Modifier.fillMaxWidth(),
        ) { Text(t(Str.ScanFillCarOrImei)) }
        Button(
            onClick = {
                scope.launch {
                    when (val result = app.productionFeature.locateVehicleOnMap()) {
                        is OpsResult.Ok -> {
                            app.vehicleFeature.upsertAndSelect(result.value)
                            app.productionFeature.clear()
                            onLocateOnMap()
                        }
                        is OpsResult.Err -> Unit
                    }
                }
            },
            enabled = !state.loading,
            modifier = Modifier.fillMaxWidth(),
        ) { Text(t(Str.DetectVehicleLocation)) }
        Button(
            onClick = {
                scope.launch { app.productionFeature.openOverloadCheck() }
            },
            enabled = !state.loading,
            modifier = Modifier.fillMaxWidth(),
        ) { Text(t(Str.DetectOverloadCheck)) }
        state.message?.let { Text(it, color = MaterialTheme.colorScheme.primary) }
        Text(t(Str.DetectSwitchesTitle), style = MaterialTheme.typography.labelLarge)
        state.detectSwitches.forEach { sw ->
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Text(
                    text = "${sw.label} · ${if (sw.on) t(Str.DetectSwitchOn) else t(Str.DetectSwitchOff)}",
                    style = MaterialTheme.typography.bodyMedium,
                )
                Switch(
                    checked = sw.on,
                    onCheckedChange = { checked ->
                        scope.launch {
                            app.productionFeature.toggleDetectSwitch(sw.kind, checked)
                        }
                    },
                    enabled = !sw.busy,
                )
            }
        }
        state.detectSteps.forEach { step ->
            val statusText = when (step.status) {
                DetectStepStatus.Idle -> t(Str.DetectPending)
                DetectStepStatus.Running -> t(Str.DetectRunning)
                DetectStepStatus.Ok -> t(Str.DetectPass)
                DetectStepStatus.Failed -> t(Str.DetectFail)
            }
            Button(
                onClick = { scope.launch { app.productionFeature.runDetectStep(step.kind) } },
                enabled = step.status != DetectStepStatus.Running,
                modifier = Modifier.fillMaxWidth(),
            ) {
                Text("${step.label} · $statusText")
            }
            if (step.detail.isNotBlank()) {
                Text(step.detail, style = MaterialTheme.typography.bodySmall)
            }
        }
    }
}

@Composable
private fun OverloadPane(app: OpsApp) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.productionFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    val contacts = state.overloadContacts
    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState()),
        verticalArrangement = Arrangement.spacedBy(10.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        TextButton(
            onClick = { app.productionFeature.backToDetect() },
            modifier = Modifier.align(Alignment.Start),
        ) { Text("← ${t(Str.Back)}") }
        Text(t(Str.DetectOverloadCheck), style = MaterialTheme.typography.titleMedium)
        if (state.overloadRunning && state.overloadCountdown > 0) {
            Text(
                text = "${state.overloadCountdown}s",
                style = MaterialTheme.typography.headlineMedium,
            )
        }
        Text(
            t(Str.DetectOverloadTips),
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
        Row(horizontalArrangement = Arrangement.spacedBy(16.dp)) {
            OverloadContactChip(label = t(Str.DetectOverloadFront), on = contacts.frontOn)
            OverloadContactChip(label = t(Str.DetectOverloadCenter), on = contacts.centerOn)
            OverloadContactChip(label = t(Str.DetectOverloadBack), on = contacts.backOn)
        }
        Button(
            onClick = {
                scope.launch { app.productionFeature.startOverloadCheck() }
            },
            enabled = !state.loading,
            modifier = Modifier.fillMaxWidth(),
        ) {
            Text(
                if (state.overloadRunning) t(Str.DetectOverloadRestart) else t(Str.DetectOverloadStart),
            )
        }
        state.message?.let { Text(it, color = MaterialTheme.colorScheme.primary) }
    }
}

@Composable
private fun OverloadContactChip(label: String, on: Boolean) {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Text(
            text = if (on) "●" else "○",
            style = MaterialTheme.typography.headlineSmall,
            color = if (on) {
                MaterialTheme.colorScheme.primary
            } else {
                MaterialTheme.colorScheme.onSurfaceVariant
            },
        )
        Text(label, style = MaterialTheme.typography.labelMedium)
    }
}

@Composable
private fun BindPane(app: OpsApp, onBack: () -> Unit) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.productionFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState()),
        verticalArrangement = Arrangement.spacedBy(10.dp),
    ) {
        TextButton(onClick = {
            app.productionFeature.clear()
            onBack()
        }) { Text("← ${t(Str.Back)}") }
        Text(t(Str.ProductionBind), style = MaterialTheme.typography.titleMedium)
        OutlinedTextField(
            value = state.carId,
            onValueChange = { app.productionFeature.setCarId(it) },
            label = { Text(t(Str.VehicleId)) },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true,
        )
        Button(
            onClick = {
                scope.launch {
                    when (val scan = app.codeScanner.scanOnce()) {
                        is OpsResult.Ok -> {
                            val hosts = app.authFeature.state.value.runtimeConfig?.qrHosts.orEmpty()
                            app.productionFeature.applyScanRaw(scan.value, hosts)
                        }
                        is OpsResult.Err -> Unit
                    }
                }
            },
            modifier = Modifier.fillMaxWidth(),
        ) { Text(t(Str.ScanFillCarOrImei)) }
        Button(
            onClick = { scope.launch { app.productionFeature.loadBindInfo() } },
            enabled = !state.loading,
            modifier = Modifier.fillMaxWidth(),
        ) { Text(t(Str.QueryBind)) }
        OutlinedTextField(
            value = state.imei,
            onValueChange = { app.productionFeature.setImei(it) },
            label = { Text(t(Str.ImeiLabel)) },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true,
        )
        Button(
            onClick = {
                scope.launch {
                    when (val scan = app.codeScanner.scanOnce()) {
                        is OpsResult.Ok -> {
                            val hosts = app.authFeature.state.value.runtimeConfig?.qrHosts.orEmpty()
                            app.productionFeature.applyScanRaw(scan.value, hosts)
                        }
                        is OpsResult.Err -> Unit
                    }
                }
            },
            modifier = Modifier.fillMaxWidth(),
        ) { Text("${t(Str.WarehouseScan)} ${t(Str.ImeiLabel)}") }
        OutlinedTextField(
            value = state.helmet,
            onValueChange = { app.productionFeature.setHelmet(it) },
            label = { Text(t(Str.HelmetOptional)) },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true,
        )
        state.bindInfo?.let { info ->
            Text(
                t(Str.BindCurrentInfo, info.imei.ifBlank { "-" }, info.brand, info.model),
                style = MaterialTheme.typography.bodySmall,
            )
        }
        state.message?.let { Text(it, color = MaterialTheme.colorScheme.primary) }
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.fillMaxWidth()) {
            Button(
                onClick = { scope.launch { app.productionFeature.bindCenter() } },
                enabled = !state.loading,
                modifier = Modifier.weight(1f),
            ) { Text(t(Str.BindAction)) }
            Button(
                onClick = { scope.launch { app.productionFeature.unbindCenter() } },
                enabled = !state.loading,
                modifier = Modifier.weight(1f),
            ) { Text(t(Str.UnbindAction)) }
        }
    }
}

@Composable
private fun ShelvesPane(app: OpsApp, onBack: () -> Unit) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.productionFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    val shelfAction = if (state.shelfMode == ShelfMode.PutOn) t(Str.ShelfPutOn) else t(Str.ShelfPullOff)
    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState()),
        verticalArrangement = Arrangement.spacedBy(10.dp),
    ) {
        TextButton(onClick = {
            app.productionFeature.clear()
            onBack()
        }) { Text("← ${t(Str.Back)}") }
        Text(
            shelfAction,
            style = MaterialTheme.typography.titleMedium,
        )
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            FilterChip(
                selected = state.shelfMode == ShelfMode.PutOn,
                onClick = { app.productionFeature.openShelves(ShelfMode.PutOn) },
                label = { Text(t(Str.ShelfPutOn)) },
            )
            FilterChip(
                selected = state.shelfMode == ShelfMode.PullOff,
                onClick = { app.productionFeature.openShelves(ShelfMode.PullOff) },
                label = { Text(t(Str.ShelfPullOff)) },
            )
        }
        if (state.shelfMode == ShelfMode.PutOn) {
            OutlinedTextField(
                value = state.shelfServiceId,
                onValueChange = { app.productionFeature.setShelfServiceId(it) },
                label = { Text(t(Str.TargetServiceAreaId)) },
                modifier = Modifier.fillMaxWidth(),
                singleLine = true,
            )
        }
        OutlinedTextField(
            value = state.carId,
            onValueChange = { app.productionFeature.setCarId(it) },
            label = { Text(t(Str.VehicleId)) },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true,
        )
        Button(
            onClick = {
                scope.launch {
                    when (val scan = app.codeScanner.scanOnce()) {
                        is OpsResult.Ok -> {
                            val hosts = app.authFeature.state.value.runtimeConfig?.qrHosts.orEmpty()
                            app.productionFeature.applyScanRaw(scan.value, hosts)
                            if (app.productionFeature.state.value.carId.isNotBlank()) {
                                app.productionFeature.addShelfCar()
                            }
                        }
                        is OpsResult.Err -> Unit
                    }
                }
            },
            enabled = !state.loading,
            modifier = Modifier.fillMaxWidth(),
        ) { Text(t(Str.ScanAddToList)) }
        Button(
            onClick = { scope.launch { app.productionFeature.addShelfCar() } },
            enabled = !state.loading,
            modifier = Modifier.fillMaxWidth(),
        ) { Text(if (state.loading) t(Str.Querying) else t(Str.ValidateAndAdd)) }
        state.shelfQueue.forEach { item ->
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(MaterialTheme.colorScheme.surfaceVariant)
                    .padding(10.dp),
                horizontalArrangement = Arrangement.SpaceBetween,
            ) {
                Column(modifier = Modifier.weight(1f)) {
                    Text(item.carId, style = MaterialTheme.typography.titleSmall)
                    Text(
                        listOfNotNull(
                            item.brand.takeIf { it.isNotBlank() },
                            item.serviceName.takeIf { it.isNotBlank() },
                        ).joinToString(" · ").ifBlank { item.carNo },
                        style = MaterialTheme.typography.bodySmall,
                    )
                }
                TextButton(onClick = { app.productionFeature.removeShelfItem(item.carId) }) {
                    Text(t(Str.Remove))
                }
            }
        }
        state.message?.let { Text(it, color = MaterialTheme.colorScheme.primary) }
        Button(
            onClick = { scope.launch { app.productionFeature.submitShelfQueue() } },
            enabled = !state.loading && state.shelfQueue.isNotEmpty(),
            modifier = Modifier.fillMaxWidth(),
        ) {
            Text(
                if (state.loading) t(Str.Submitting)
                else t(Str.ConfirmShelf, shelfAction),
            )
        }
    }
}
