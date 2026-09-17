package com.luopingtech.ebike.ops.ui.workorder

import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
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
import com.luopingtech.ebike.ops.domain.model.AlarmTypeLabels
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.model.WorkOrder
import com.luopingtech.ebike.ops.domain.model.WorkOrderKind
import com.luopingtech.ebike.ops.domain.model.WorkOrderStateLabels
import com.luopingtech.ebike.ops.feature.workorder.WorkOrderFeature
import kotlinx.coroutines.launch

@Composable
fun WorkOrderScreen(
    app: OpsApp,
    kind: WorkOrderKind,
    currentArea: ServiceArea?,
    onClose: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val feature: WorkOrderFeature = when (kind) {
        WorkOrderKind.Inspection -> app.inspectionOrderFeature
        WorkOrderKind.Repair -> app.repairOrderFeature
    }
    val state by feature.state.collectAsState()
    val scope = rememberCoroutineScope()
    val title = when (kind) {
        WorkOrderKind.Inspection -> t(Str.InspectionOrder)
        WorkOrderKind.Repair -> t(Str.RepairOrder)
    }

    LaunchedEffect(currentArea?.id, kind) {
        feature.load(currentArea)
    }

    fun closePage() {
        feature.clear()
        onClose()
    }

    com.luopingtech.ebike.ops.ui.navigation.OpsBackHandler(onBack = ::closePage)

    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState())
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
        ) {
            Text(title, style = MaterialTheme.typography.headlineSmall)
            TextButton(onClick = ::closePage) { Text(t(Str.Back)) }
        }
        Text(
            text = t(Str.WorkOrderHint),
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )

        when (kind) {
            WorkOrderKind.Inspection -> {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .horizontalScroll(rememberScrollState()),
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    FilterChip(
                        selected = state.selectedAlarmType == null,
                        onClick = {
                            feature.setAlarmType(null)
                            scope.launch { feature.load(currentArea) }
                        },
                        label = { Text(t(Str.FilterAll)) },
                    )
                    AlarmTypeLabels.ALL_CODES.forEach { code ->
                        FilterChip(
                            selected = state.selectedAlarmType == code,
                            onClick = {
                                feature.setAlarmType(code)
                                scope.launch { feature.load(currentArea) }
                            },
                            label = { Text(t(AlarmTypeLabels.strKey(code))) },
                        )
                    }
                }
            }
            WorkOrderKind.Repair -> {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .horizontalScroll(rememberScrollState()),
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    FilterChip(
                        selected = state.selectedFixName == null,
                        onClick = {
                            feature.setFixName(null)
                            scope.launch { feature.load(currentArea) }
                        },
                        label = { Text(t(Str.FilterAll)) },
                    )
                    state.fixNameChips.forEach { name ->
                        FilterChip(
                            selected = state.selectedFixName == name,
                            onClick = {
                                feature.setFixName(name)
                                scope.launch { feature.load(currentArea) }
                            },
                            label = { Text(name) },
                        )
                    }
                }
            }
        }

        OutlinedTextField(
            value = state.carIdQuery,
            onValueChange = { feature.setCarIdQuery(it) },
            label = { Text(t(Str.WorkOrderCarIdFilter)) },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true,
        )
        Button(
            onClick = { scope.launch { feature.load(currentArea) } },
            enabled = !state.loading,
            modifier = Modifier.fillMaxWidth(),
        ) { Text(if (state.loading) t(Str.LoadingEllipsis) else t(Str.Refresh)) }

        state.orders.forEach { order ->
            WorkOrderRow(
                order = order,
                canTake = state.canTake,
                acting = state.actingId == order.id,
                loading = state.loading,
                t = { key -> t(key) },
                onAccept = { scope.launch { feature.accept(order.id) } },
                onFinish = { scope.launch { feature.finish(order.id) } },
            )
        }
        if (!state.loading && state.orders.isEmpty()) {
            Text(
                text = t(Str.WorkOrderCount, 0),
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }
        state.message?.let { Text(it) }
        state.errorMessage?.let { Text(it, color = MaterialTheme.colorScheme.error) }
    }
}

@Composable
private fun WorkOrderRow(
    order: WorkOrder,
    canTake: Boolean,
    acting: Boolean,
    loading: Boolean,
    t: (Str) -> String,
    onAccept: () -> Unit,
    onFinish: () -> Unit,
) {
    Column(
        modifier = Modifier.fillMaxWidth(),
        verticalArrangement = Arrangement.spacedBy(4.dp),
    ) {
        Text(
            text = "${t(Str.WorkOrderId)} ${order.id}",
            style = MaterialTheme.typography.titleSmall,
        )
        Text(order.carId, style = MaterialTheme.typography.bodyLarge)
        val typeLabel = when (order.kind) {
            WorkOrderKind.Inspection ->
                order.alarmType?.let { t(AlarmTypeLabels.strKey(it)) } ?: t(Str.AlarmTypeUnknown)
            WorkOrderKind.Repair ->
                order.partNames.joinToString(" · ").ifBlank { t(Str.WorkOrderParts) }
        }
        Text(
            text = "$typeLabel · ${t(WorkOrderStateLabels.strKey(order.state))}",
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
        if (order.fixReason.isNotBlank()) {
            Text(
                text = "${t(Str.WorkOrderReason)} · ${order.fixReason}",
                style = MaterialTheme.typography.bodySmall,
            )
        }
        if (order.opManName.isNotBlank() || order.opManPhone.isNotBlank()) {
            Text(
                text = "${t(Str.WorkOrderOperator)} · ${order.opManName} ${order.opManPhone}".trim(),
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }
        if (order.carStopped) {
            Text(
                text = t(Str.WorkOrderStopped),
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.error,
            )
        }
        if (order.createdAt.isNotBlank()) {
            Text(
                text = order.createdAt,
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }
        if (canTake) {
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                if (order.canAccept) {
                    Button(
                        onClick = onAccept,
                        enabled = !loading && !acting,
                    ) {
                        Text(if (acting) t(Str.LoadingEllipsis) else t(Str.WorkOrderAccept))
                    }
                }
                if (order.canFinish) {
                    Button(
                        onClick = onFinish,
                        enabled = !loading && !acting,
                    ) {
                        Text(if (acting) t(Str.LoadingEllipsis) else t(Str.WorkOrderFinish))
                    }
                }
            }
        }
    }
}
