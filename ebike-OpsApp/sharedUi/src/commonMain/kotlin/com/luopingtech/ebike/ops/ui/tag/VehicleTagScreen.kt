package com.luopingtech.ebike.ops.ui.tag

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.Checkbox
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.compose.ui.window.Dialog
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.ui.analysis.VcdTopBar
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import kotlinx.coroutines.launch

/**
 * CMP 车辆标记：对齐遗留 VehicleTagActivity（本地待标记列表 → 选类型 → multipart 提交）。
 */
@Composable
fun VehicleTagScreen(
    app: OpsApp,
    onClose: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val home by app.homeFeature.state.collectAsState()
    val state by app.vehicleTagFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    val feature = app.vehicleTagFeature

    LaunchedEffect(home.currentArea?.id) {
        feature.load(home.currentArea)
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(Color(0xFFF6F7F9)),
    ) {
        VcdTopBar(title = t(Str.VehicleTagTool), onBack = {
            feature.clear()
            onClose()
        })
        Column(
            modifier = Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState())
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            if (state.loading && state.types.isEmpty()) {
                CircularProgressIndicator(color = OpsTheme.colors.primary)
            } else {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    OutlinedTextField(
                        value = state.carId,
                        onValueChange = { feature.setCarId(it) },
                        label = { Text(t(Str.VehicleTagCarId)) },
                        modifier = Modifier.weight(1f),
                        singleLine = true,
                    )
                    Button(
                        onClick = {
                            scope.launch {
                                when (val scan = app.codeScanner.scanOnce()) {
                                    is OpsResult.Ok -> {
                                        feature.applyScan(scan.value)
                                        feature.addToPending(home.currentArea)
                                    }
                                    is OpsResult.Err -> Unit
                                }
                            }
                        },
                    ) { Text(t(Str.VehicleTagScan)) }
                    Button(
                        onClick = { scope.launch { feature.addToPending(home.currentArea) } },
                        enabled = !state.loading,
                    ) { Text(t(Str.VehicleTagAdd)) }
                }

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text(t(Str.VehicleTagPendingList), style = MaterialTheme.typography.titleSmall)
                    TextButton(onClick = {
                        val allOn = state.pending.all { it.selected }
                        feature.selectAllPending(!allOn)
                    }) { Text(t(Str.VehicleTagSelectAll)) }
                }

                if (state.pending.isEmpty()) {
                    Text(t(Str.AdminEmptyList), color = MaterialTheme.colorScheme.onSurfaceVariant)
                } else {
                    state.pending.forEach { item ->
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .background(Color.White, MaterialTheme.shapes.medium)
                                .padding(horizontal = 8.dp, vertical = 4.dp),
                            verticalAlignment = Alignment.CenterVertically,
                        ) {
                            Checkbox(
                                checked = item.selected,
                                onCheckedChange = { feature.togglePending(item.carId) },
                            )
                            Text(item.carId, modifier = Modifier.weight(1f))
                            TextButton(onClick = { feature.removePending(item.carId) }) {
                                Text(t(Str.VehicleTagDelete))
                            }
                        }
                    }
                }

                Button(
                    onClick = { feature.openTypePicker(true) },
                    enabled = !state.submitting && state.pending.any { it.selected },
                    modifier = Modifier.fillMaxWidth(),
                ) {
                    Text(if (state.submitting) t(Str.LoadingEllipsis) else t(Str.VehicleTagSubmit))
                }
            }
            state.message?.let { Text(it, color = MaterialTheme.colorScheme.primary) }
            state.errorMessage?.let { Text(it, color = MaterialTheme.colorScheme.error) }
        }
    }

    if (state.showTypePicker) {
        Dialog(onDismissRequest = { feature.openTypePicker(false) }) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(Color.White, MaterialTheme.shapes.medium)
                    .padding(16.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                Text(t(Str.VehicleTagPickType), style = MaterialTheme.typography.titleMedium)
                if (state.types.isEmpty()) {
                    Text(t(Str.AdminEmptyList), color = MaterialTheme.colorScheme.onSurfaceVariant)
                } else {
                    state.types.forEach { type ->
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .clickable {
                                    scope.launch {
                                        feature.submitSelected(home.currentArea, type.id)
                                    }
                                }
                                .padding(vertical = 12.dp),
                        ) {
                            Text(type.name, style = MaterialTheme.typography.bodyLarge)
                        }
                    }
                }
                TextButton(
                    onClick = { feature.openTypePicker(false) },
                    modifier = Modifier.align(Alignment.End),
                ) { Text(t(Str.Cancel)) }
            }
        }
    }
}
