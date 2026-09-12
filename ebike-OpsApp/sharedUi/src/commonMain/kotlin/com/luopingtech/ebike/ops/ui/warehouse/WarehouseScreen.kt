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
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.WarehouseOperationType
import com.luopingtech.ebike.ops.domain.permission.OpsPermissions
import com.luopingtech.ebike.ops.feature.warehouse.WarehousePage
import kotlinx.coroutines.launch

/**
 * 仓库：工作台分入口直达下级（对齐遗留 WarehouseMainActivity / Record）。
 * - 归还入库 / 领用出库 → 有编码 / 无编码 二选一
 * - 出入库记录 → 记录列表
 */
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

    fun close() {
        app.warehouseFeature.clear()
        onClose()
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(Color(0xFFF8F8F8))
            .padding(16.dp),
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
        ) {
            val title = when (state.page) {
                WarehousePage.Records, WarehousePage.Detail -> t(Str.WarehouseRecords)
                WarehousePage.KindMenu, WarehousePage.Operate -> state.operationType.label
            }
            Text(title, style = MaterialTheme.typography.headlineSmall)
            TextButton(onClick = { close() }) {
                Text(t(Str.Close))
            }
        }

        when (state.page) {
            WarehousePage.KindMenu -> WarehouseKindMenu(
                operationType = state.operationType,
                message = state.message,
                onCoded = {
                    app.warehouseFeature.openOperate(state.operationType, withCode = true)
                    scope.launch { app.warehouseFeature.ensureComponentNames(1) }
                },
                onUncoded = {
                    app.warehouseFeature.openOperate(state.operationType, withCode = false)
                    scope.launch {
                        app.warehouseFeature.ensureComponentNames(0)
                        app.warehouseFeature.refreshStockHint()
                    }
                },
                translate = { key, args -> t(key, *args) },
            )
            WarehousePage.Operate -> WarehouseOperatePane(
                app = app,
                loading = state.loading,
                onBack = { app.warehouseFeature.backToKindMenu() },
            )
            WarehousePage.Records -> WarehouseRecordsPane(app = app, onClose = { close() })
            WarehousePage.Detail -> WarehouseDetailPane(app = app)
        }

        state.errorMessage?.let {
            Spacer(modifier = Modifier.height(8.dp))
            Text(it, color = MaterialTheme.colorScheme.error)
        }
    }
}

@Composable
private fun WarehouseKindMenu(
    operationType: WarehouseOperationType,
    message: String?,
    onCoded: () -> Unit,
    onUncoded: () -> Unit,
    translate: (Str, Array<out Any?>) -> String,
) {
    val accent = if (operationType == WarehouseOperationType.In) {
        Color(0xFF00A763)
    } else {
        Color(0xFF196CFF)
    }
    val uncodedAccent = if (operationType == WarehouseOperationType.In) {
        Color(0xFFFF7319)
    } else {
        Color(0xFF009BC5)
    }
    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState()),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        message?.let { Text(it, color = MaterialTheme.colorScheme.primary) }
        WarehouseModeCard(
            title = translate(Str.WarehouseCodedTitle, emptyArray()),
            subtitle = translate(Str.WarehouseCodedSubtitle, emptyArray()),
            accent = accent,
            onClick = onCoded,
        )
        WarehouseModeCard(
            title = translate(Str.WarehouseUncodedTitle, emptyArray()),
            subtitle = translate(Str.WarehouseUncodedSubtitle, emptyArray()),
            accent = uncodedAccent,
            onClick = onUncoded,
        )
    }
}

@Composable
private fun WarehouseModeCard(
    title: String,
    subtitle: String,
    accent: Color,
    onClick: () -> Unit,
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(Color.White, MaterialTheme.shapes.medium)
            .clickable(onClick = onClick)
            .padding(horizontal = 20.dp, vertical = 18.dp),
        verticalArrangement = Arrangement.spacedBy(4.dp),
    ) {
        Text(title, style = MaterialTheme.typography.titleMedium, color = accent)
        Text(subtitle, style = MaterialTheme.typography.bodySmall, color = accent.copy(alpha = 0.85f))
    }
}

@Composable
private fun WarehouseOperatePane(
    app: OpsApp,
    loading: Boolean,
    onBack: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.warehouseFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    val modeTitle = when {
        state.operationType == WarehouseOperationType.In && state.withCode -> t(Str.WarehouseScanInTitle)
        state.operationType == WarehouseOperationType.In && !state.withCode -> t(Str.WarehouseBulkInTitle)
        state.operationType == WarehouseOperationType.Out && state.withCode -> t(Str.WarehouseScanOutTitle)
        else -> t(Str.WarehouseBulkOutTitle)
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState()),
        verticalArrangement = Arrangement.spacedBy(10.dp),
    ) {
        TextButton(onClick = onBack) {
            Text("← ${t(Str.Back)}")
        }
        Text(modeTitle, style = MaterialTheme.typography.titleMedium)
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
                ) { Text(t(Str.VehicleTagAdd)) }
            }
            state.scanned.forEach { item ->
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                ) {
                    Text("${item.componentNo} · ${item.componentName}")
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
                Text(if (loading) t(Str.Submitting) else t(Str.Confirm))
            }
        } else {
            Text(t(Str.ComponentName), style = MaterialTheme.typography.labelLarge)
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
                enabled = !loading,
                modifier = Modifier.fillMaxWidth(),
            ) {
                Text(if (loading) t(Str.Submitting) else t(Str.Confirm))
            }
        }
        state.message?.let { Text(it, color = MaterialTheme.colorScheme.primary) }
    }
}

@Composable
private fun WarehouseRecordsPane(app: OpsApp, onClose: () -> Unit) {
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
        TextButton(onClick = onClose) {
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
                    .background(Color.White, MaterialTheme.shapes.medium)
                    .padding(12.dp),
            ) {
                Text(
                    "${record.componentName} · ${record.operationTypeName.ifBlank { record.operationType?.label.orEmpty() }}",
                    style = MaterialTheme.typography.titleSmall,
                )
                Text(
                    "${record.operationNum} · ${record.operator} · ${record.operationTime}",
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
        if (state.loading) Text(t(Str.Loading))
        state.details.forEach { item ->
            Text(
                "${item.componentNo} · ${item.componentName} × ${item.operationNum}",
                style = MaterialTheme.typography.bodyMedium,
            )
        }
        if (!state.loading && state.details.isEmpty()) {
            Text(t(Str.NoRecords), color = MaterialTheme.colorScheme.onSurfaceVariant)
        }
    }
}
