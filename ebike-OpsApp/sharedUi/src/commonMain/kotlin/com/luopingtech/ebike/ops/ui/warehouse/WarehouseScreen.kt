package com.luopingtech.ebike.ops.ui.warehouse

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
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.WarehouseOperationType
import com.luopingtech.ebike.ops.domain.permission.OpsPermissions
import com.luopingtech.ebike.ops.feature.warehouse.WarehousePage
import kotlinx.coroutines.launch

@Composable
fun WarehouseScreen(
    app: OpsApp,
    permissions: OpsPermissions,
    onClose: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.warehouseFeature.state.collectAsState()
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
            Text(t(Str.WarehouseTitle), style = MaterialTheme.typography.headlineSmall)
            TextButton(onClick = {
                app.warehouseFeature.clear()
                onClose()
            }) {
                Text(t(Str.Close))
            }
        }

        when (state.page) {
            WarehousePage.Hub -> WarehouseHub(
                app = app,
                permissions = permissions,
                message = state.message,
                onOutCode = {
                    app.warehouseFeature.openOperate(WarehouseOperationType.Out, withCode = true)
                    scope.launch { app.warehouseFeature.ensureComponentNames(1) }
                },
                onOutBulk = {
                    app.warehouseFeature.openOperate(WarehouseOperationType.Out, withCode = false)
                    scope.launch {
                        app.warehouseFeature.ensureComponentNames(0)
                        app.warehouseFeature.refreshStockHint()
                    }
                },
                onInCode = {
                    app.warehouseFeature.openOperate(WarehouseOperationType.In, withCode = true)
                    scope.launch { app.warehouseFeature.ensureComponentNames(1) }
                },
                onInBulk = {
                    app.warehouseFeature.openOperate(WarehouseOperationType.In, withCode = false)
                    scope.launch {
                        app.warehouseFeature.ensureComponentNames(0)
                        app.warehouseFeature.refreshStockHint()
                    }
                },
                onRecords = {
                    scope.launch { app.warehouseFeature.openRecords() }
                },
            )
            WarehousePage.Operate -> WarehouseOperatePane(
                app = app,
                loading = state.loading,
            )
            WarehousePage.Records -> WarehouseRecordsPane(app = app)
            WarehousePage.Detail -> WarehouseDetailPane(app = app)
        }

        state.errorMessage?.let {
            Spacer(modifier = Modifier.height(8.dp))
            Text(it, color = MaterialTheme.colorScheme.error)
        }
    }
}

@Composable
private fun WarehouseHub(
    app: OpsApp,
    permissions: OpsPermissions,
    message: String?,
    onOutCode: () -> Unit,
    onOutBulk: () -> Unit,
    onInCode: () -> Unit,
    onInBulk: () -> Unit,
    onRecords: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState()),
        verticalArrangement = Arrangement.spacedBy(10.dp),
    ) {
        Text(
            text = t(Str.WarehouseHint),
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
        message?.let { Text(it, color = MaterialTheme.colorScheme.primary) }
        if (permissions.showWarehouseOut) {
            Text(t(Str.WarehouseOut), style = MaterialTheme.typography.titleMedium)
            Button(onClick = onOutCode, modifier = Modifier.fillMaxWidth()) {
                Text(t(Str.WarehouseOutScan))
            }
            Button(onClick = onOutBulk, modifier = Modifier.fillMaxWidth()) {
                Text(t(Str.WarehouseOutBulk))
            }
        }
        if (permissions.showWarehouseIn) {
            Text(t(Str.WarehouseIn), style = MaterialTheme.typography.titleMedium)
            Button(onClick = onInCode, modifier = Modifier.fillMaxWidth()) {
                Text(t(Str.WarehouseInScan))
            }
            Button(onClick = onInBulk, modifier = Modifier.fillMaxWidth()) {
                Text(t(Str.WarehouseInBulk))
            }
        }
        if (permissions.showWarehouseRecord) {
            Button(onClick = onRecords, modifier = Modifier.fillMaxWidth()) {
                Text(t(Str.WarehouseRecords))
            }
        }
    }
}

@Composable
private fun WarehouseOperatePane(app: OpsApp, loading: Boolean) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.warehouseFeature.state.collectAsState()
    val scope = rememberCoroutineScope()

    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState()),
        verticalArrangement = Arrangement.spacedBy(10.dp),
    ) {
        TextButton(onClick = { app.warehouseFeature.openHub() }) {
            Text("← ${t(Str.Back)}")
        }
        Text(
            text = "${state.operationType.label} · ${if (state.withCode) t(Str.WithCode) else t(Str.WithoutCode)}",
            style = MaterialTheme.typography.titleMedium,
        )
        if (state.withCode) {
            OutlinedTextField(
                value = state.codeInput,
                onValueChange = { app.warehouseFeature.setCodeInput(it) },
                label = { Text(t(Str.ComponentCode)) },
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
                                    app.warehouseFeature.setCodeInput(scan.value)
                                    app.warehouseFeature.resolveCode()
                                }
                                is OpsResult.Err -> Unit
                            }
                        }
                    },
                    enabled = !loading,
                    modifier = Modifier.weight(1f),
                ) { Text(t(Str.WarehouseScan)) }
                Button(
                    onClick = { scope.launch { app.warehouseFeature.resolveCode() } },
                    enabled = !loading,
                    modifier = Modifier.weight(1f),
                ) {
                    Text(if (loading) t(Str.Querying) else t(Str.FreeMoveJoin))
                }
            }
            state.scanned.forEach { item ->
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .background(MaterialTheme.colorScheme.surfaceVariant)
                        .padding(10.dp),
                    horizontalArrangement = Arrangement.SpaceBetween,
                ) {
                    Column(modifier = Modifier.weight(1f)) {
                        Text(item.componentNo, style = MaterialTheme.typography.titleSmall)
                        Text(
                            "${item.componentName} · ${item.brand}",
                            style = MaterialTheme.typography.bodySmall,
                        )
                    }
                    TextButton(onClick = { app.warehouseFeature.removeScanned(item.componentNo) }) {
                        Text(t(Str.Remove))
                    }
                }
            }
            Button(
                onClick = { scope.launch { app.warehouseFeature.submitOperate() } },
                enabled = !loading && state.scanned.isNotEmpty(),
                modifier = Modifier.fillMaxWidth(),
            ) {
                Text(if (loading) t(Str.Submitting) else t(Str.ConfirmOp, state.operationType.label))
            }
        } else {
            Text(t(Str.ComponentName), style = MaterialTheme.typography.labelLarge)
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                state.componentNames.forEach { name ->
                    FilterChip(
                        selected = state.selectedComponentName == name,
                        onClick = {
                            app.warehouseFeature.setSelectedComponentName(name)
                            scope.launch { app.warehouseFeature.refreshStockHint() }
                        },
                        label = { Text(name) },
                    )
                }
            }
            state.stockHint?.let {
                Text(it, style = MaterialTheme.typography.bodySmall)
            }
            OutlinedTextField(
                value = state.quantity.toString(),
                onValueChange = {
                    app.warehouseFeature.setQuantity(it.filter { ch -> ch.isDigit() }.toIntOrNull() ?: 1)
                },
                label = { Text(t(Str.Quantity)) },
                modifier = Modifier.fillMaxWidth(),
                singleLine = true,
            )
            Button(
                onClick = { scope.launch { app.warehouseFeature.submitOperate() } },
                enabled = !loading && state.selectedComponentName.isNotBlank(),
                modifier = Modifier.fillMaxWidth(),
            ) {
                Text(if (loading) t(Str.Submitting) else t(Str.ConfirmOp, state.operationType.label))
            }
        }
    }
}

@Composable
private fun WarehouseRecordsPane(app: OpsApp) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.warehouseFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState()),
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        TextButton(onClick = { app.warehouseFeature.openHub() }) {
            Text("← ${t(Str.Back)}")
        }
        Text(t(Str.WarehouseRecords), style = MaterialTheme.typography.titleMedium)
        if (state.loading) Text(t(Str.Loading))
        state.records.forEach { record ->
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .clickable {
                        scope.launch { app.warehouseFeature.openDetail(record.id) }
                    }
                    .background(MaterialTheme.colorScheme.surfaceVariant)
                    .padding(12.dp),
            ) {
                Text(
                    "${record.operationTypeName.ifBlank { record.operationType?.label }} · ${record.componentName}",
                    style = MaterialTheme.typography.titleSmall,
                )
                Text(
                    "${t(Str.Quantity)} ${record.operationNum} · ${record.receiver} · ${record.operationTime}",
                    style = MaterialTheme.typography.bodySmall,
                )
            }
        }
        if (!state.loading && state.records.isEmpty()) {
            Text(t(Str.NoRecords), color = MaterialTheme.colorScheme.onSurfaceVariant)
        }
    }
}

@Composable
private fun WarehouseDetailPane(app: OpsApp) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.warehouseFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState()),
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        TextButton(onClick = { scope.launch { app.warehouseFeature.openRecords() } }) {
            Text("← ${t(Str.WarehouseRecords)}")
        }
        Text(t(Str.RecordTitle, state.selectedRecordId), style = MaterialTheme.typography.titleMedium)
        if (state.loading) Text(t(Str.Loading))
        state.details.forEach { item ->
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(MaterialTheme.colorScheme.surfaceVariant)
                    .padding(12.dp),
            ) {
                Text(
                    item.componentNo.ifBlank { item.componentName },
                    style = MaterialTheme.typography.titleSmall,
                )
                Text(
                    listOfNotNull(
                        item.componentName.takeIf { it.isNotBlank() },
                        item.operationTypeName.takeIf { it.isNotBlank() },
                        item.receiver.takeIf { it.isNotBlank() },
                    ).joinToString(" · "),
                    style = MaterialTheme.typography.bodySmall,
                )
            }
        }
    }
}
