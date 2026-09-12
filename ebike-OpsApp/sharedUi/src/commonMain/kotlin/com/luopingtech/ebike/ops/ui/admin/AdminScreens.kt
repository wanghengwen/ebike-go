package com.luopingtech.ebike.ops.ui.admin

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.domain.admin.BlacklistItem
import com.luopingtech.ebike.ops.domain.admin.CareerAuditItem
import com.luopingtech.ebike.ops.domain.admin.IdBindAuditItem
import com.luopingtech.ebike.ops.domain.admin.ObjectionOrderItem
import com.luopingtech.ebike.ops.domain.admin.OperationLogItem
import com.luopingtech.ebike.ops.ui.analysis.VcdTopBar
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import kotlinx.coroutines.launch

private fun auditStateLabel(translate: (Str) -> String, state: Int): String = when (state) {
    0 -> translate(Str.AuditResultPending)
    1 -> translate(Str.AuditStatePassed)
    2 -> translate(Str.AuditStateRejected)
    else -> state.toString()
}

@Composable
fun ProfessionAuditScreen(app: OpsApp, onClose: () -> Unit) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val home by app.homeFeature.state.collectAsState()
    val state by app.professionAuditFeature.state.collectAsState()
    val scope = rememberCoroutineScope()

    LaunchedEffect(home.currentArea?.id) { app.professionAuditFeature.load(home.currentArea) }

    AdminShell(title = t(Str.ProfessionAudit), onClose = {
        app.professionAuditFeature.clear()
        onClose()
    }, loading = state.loading, error = state.errorMessage, message = state.message) {
        if (state.items.isEmpty()) {
            Text(t(Str.AdminEmptyList))
        } else {
            state.items.forEach { item ->
                AuditRow(
                    title = item.name.ifBlank { item.phone },
                    subtitle = t(Str.AuditStateLabel, auditStateLabel({ t(it) }, item.auditState)),
                    onPass = { scope.launch { app.professionAuditFeature.audit(home.currentArea, item.id, true) } },
                    onReject = { scope.launch { app.professionAuditFeature.audit(home.currentArea, item.id, false) } },
                    passLabel = t(Str.AuditPass),
                    rejectLabel = t(Str.AuditReject),
                    showActions = item.auditState == 0,
                )
            }
        }
    }
}

@Composable
fun ObjectionOrderScreen(app: OpsApp, onClose: () -> Unit) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val home by app.homeFeature.state.collectAsState()
    val state by app.objectionOrderFeature.state.collectAsState()
    val scope = rememberCoroutineScope()

    LaunchedEffect(home.currentArea?.id) { app.objectionOrderFeature.load(home.currentArea) }

    val selected = state.selected
    if (selected != null) {
        AdminShell(title = t(Str.ObjectionOrderDetail), onClose = {
            app.objectionOrderFeature.closeDetail()
        }, loading = state.loading, error = state.errorMessage, message = state.message) {
            Text("${selected.userName} · ${selected.phone}")
            Text("${selected.carId} · ${selected.orderId}")
            Text(selected.userReason.ifBlank { "-" })
            if (selected.state == 0) {
                Button(
                    onClick = { scope.launch { app.objectionOrderFeature.deal(home.currentArea) } },
                    modifier = Modifier.fillMaxWidth(),
                ) { Text(t(Str.Confirm)) }
            }
        }
        return
    }

    AdminShell(title = t(Str.ObjectionOrder), onClose = {
        app.objectionOrderFeature.clear()
        onClose()
    }, loading = state.loading, error = state.errorMessage, message = state.message) {
        if (state.items.isEmpty()) {
            Text(t(Str.AdminEmptyList))
        } else {
            state.items.forEach { item ->
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .clickable { scope.launch { app.objectionOrderFeature.openDetail(home.currentArea, item) } }
                        .background(Color.White, MaterialTheme.shapes.medium)
                        .padding(12.dp),
                ) {
                    Text(item.userName.ifBlank { item.phone }, style = MaterialTheme.typography.titleSmall)
                    Text("${item.carId} · ${auditStateLabel({ t(it) }, item.state)}")
                }
            }
        }
    }
}

@Composable
fun BlacklistScreen(app: OpsApp, onClose: () -> Unit) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val home by app.homeFeature.state.collectAsState()
    val state by app.blacklistFeature.state.collectAsState()
    val scope = rememberCoroutineScope()

    LaunchedEffect(home.currentArea?.id) { app.blacklistFeature.load(home.currentArea) }

    AdminShell(title = t(Str.Blacklist), onClose = {
        app.blacklistFeature.clear()
        onClose()
    }, loading = state.loading, error = state.errorMessage, message = state.message) {
        if (state.items.isEmpty()) {
            Text(t(Str.AdminEmptyList))
        } else {
            state.items.forEach { item ->
                BlacklistRow(item, cancelLabel = t(Str.BlacklistCancel)) {
                    scope.launch { app.blacklistFeature.cancel(home.currentArea, item.id) }
                }
            }
        }
    }
}

@Composable
fun IdBindAuditScreen(app: OpsApp, onClose: () -> Unit) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val home by app.homeFeature.state.collectAsState()
    val state by app.idBindAuditFeature.state.collectAsState()
    val scope = rememberCoroutineScope()

    LaunchedEffect(home.currentArea?.id) { app.idBindAuditFeature.load(home.currentArea) }

    AdminShell(title = t(Str.IdBindAudit), onClose = {
        app.idBindAuditFeature.clear()
        onClose()
    }, loading = state.loading, error = state.errorMessage, message = state.message) {
        if (state.items.isEmpty()) {
            Text(t(Str.AdminEmptyList))
        } else {
            state.items.forEach { item ->
                AuditRow(
                    title = item.authName.ifBlank { item.applyPhone },
                    subtitle = "${item.originPhone} → ${item.applyPhone} · ${auditStateLabel({ t(it) }, item.auditState)}",
                    onPass = { scope.launch { app.idBindAuditFeature.audit(home.currentArea, item.id, true) } },
                    onReject = { scope.launch { app.idBindAuditFeature.audit(home.currentArea, item.id, false) } },
                    passLabel = t(Str.AuditPass),
                    rejectLabel = t(Str.AuditReject),
                    showActions = item.auditState == 0,
                )
            }
        }
    }
}

@Composable
fun OperationLogScreen(app: OpsApp, onClose: () -> Unit) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val home by app.homeFeature.state.collectAsState()
    val state by app.operationLogFeature.state.collectAsState()

    LaunchedEffect(home.currentArea?.id) { app.operationLogFeature.load(home.currentArea) }

    AdminShell(title = t(Str.OpLogTitle), onClose = {
        app.operationLogFeature.clear()
        onClose()
    }, loading = state.loading, error = state.errorMessage, message = null) {
        if (state.items.isEmpty()) {
            Text(t(Str.AdminEmptyList))
        } else {
            state.items.forEach { item -> OperationLogRow(item) }
        }
    }
}

@Composable
private fun AdminShell(
    title: String,
    onClose: () -> Unit,
    loading: Boolean,
    error: String?,
    message: String?,
    content: @Composable () -> Unit,
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(Color(0xFFF6F7F9)),
    ) {
        VcdTopBar(title = title, onBack = onClose)
        Column(
            modifier = Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState())
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            if (loading) CircularProgressIndicator(color = OpsTheme.colors.primary)
            error?.let { Text(it, color = MaterialTheme.colorScheme.error) }
            message?.let { Text(it, color = MaterialTheme.colorScheme.primary) }
            content()
        }
    }
}

@Composable
private fun AuditRow(
    title: String,
    subtitle: String,
    passLabel: String,
    rejectLabel: String,
    showActions: Boolean,
    onPass: () -> Unit,
    onReject: () -> Unit,
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(Color.White, MaterialTheme.shapes.medium)
            .padding(12.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        Text(title, style = MaterialTheme.typography.titleSmall)
        Text(subtitle, style = MaterialTheme.typography.bodySmall)
        if (showActions) {
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                Button(onClick = onPass) { Text(passLabel) }
                TextButton(onClick = onReject) { Text(rejectLabel) }
            }
        }
    }
}

@Composable
private fun BlacklistRow(item: BlacklistItem, cancelLabel: String, onCancel: () -> Unit) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(Color.White, MaterialTheme.shapes.medium)
            .padding(12.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        Text(item.authName.ifBlank { item.phone }, style = MaterialTheme.typography.titleSmall)
        Text(item.reason.ifBlank { "-" }, style = MaterialTheme.typography.bodySmall)
        if (item.state == 1) {
            TextButton(onClick = onCancel) { Text(cancelLabel) }
        }
    }
}

@Composable
private fun OperationLogRow(item: OperationLogItem) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(Color.White, MaterialTheme.shapes.medium)
            .padding(12.dp),
    ) {
        Text(item.time, style = MaterialTheme.typography.labelMedium)
        Text(item.operatorName, style = MaterialTheme.typography.titleSmall)
        Text(item.content, style = MaterialTheme.typography.bodyMedium)
        if (item.carId.isNotBlank()) {
            Text(item.carId, style = MaterialTheme.typography.bodySmall)
        }
    }
}
