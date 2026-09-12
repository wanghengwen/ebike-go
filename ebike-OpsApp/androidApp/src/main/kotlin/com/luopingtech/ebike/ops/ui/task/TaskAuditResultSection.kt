package com.luopingtech.ebike.ops.ui.task

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.domain.model.OpsTask
import com.luopingtech.ebike.ops.domain.model.canViewAuditResult
import kotlinx.coroutines.launch

@Composable
fun TaskAuditResultSection(
    app: OpsApp,
    task: OpsTask?,
) {
    if (task == null || !task.canViewAuditResult()) return
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val auditState by app.taskAuditFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    val showingThis = auditState.taskId == task.id && auditState.result != null

    Column(
        modifier = Modifier.fillMaxWidth(),
        verticalArrangement = Arrangement.spacedBy(6.dp),
    ) {
        Button(
            onClick = {
                scope.launch {
                    if (showingThis) {
                        app.taskAuditFeature.clear()
                    } else {
                        app.taskAuditFeature.load(task)
                    }
                }
            },
            enabled = !auditState.loading,
            modifier = Modifier.fillMaxWidth(),
        ) {
            Text(
                when {
                    auditState.loading && auditState.taskId == task.id -> t(Str.LoadingEllipsis)
                    showingThis -> t(Str.Collapse)
                    else -> t(Str.AuditResultView)
                },
            )
        }
        if (auditState.taskId == task.id) {
            auditState.errorMessage?.let {
                Text(it, color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.bodySmall)
            }
            auditState.result?.let { result ->
                Text(
                    text = "${t(Str.AuditResultTitle)} · ${result.statusLabel}",
                    style = MaterialTheme.typography.titleSmall,
                    color = when {
                        result.isRejected -> MaterialTheme.colorScheme.error
                        result.isPassed -> MaterialTheme.colorScheme.primary
                        else -> MaterialTheme.colorScheme.onSurface
                    },
                )
                if (result.checkView.isNotBlank()) {
                    Text(
                        text = "${t(Str.AuditCheckView)}：${result.checkView}",
                        style = MaterialTheme.typography.bodySmall,
                    )
                }
                if (result.remark.isNotBlank()) {
                    Text(
                        text = "${t(Str.AuditFieldRemark)}：${result.remark}",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                }
                if (result.checkTime.isNotBlank()) {
                    Text(
                        text = "${t(Str.AuditCheckTime)}：${result.checkTime}",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                }
                if (result.checkManName.isNotBlank()) {
                    Text(
                        text = "${t(Str.AuditCheckMan)}：${result.checkManName}",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                }
                if (result.photo.isNotEmpty()) {
                    Text(
                        text = "${t(Str.AuditPhotos)} · ${result.photo.size}",
                        style = MaterialTheme.typography.bodySmall,
                    )
                    result.photo.take(6).forEach { url ->
                        Text(
                            text = url,
                            style = MaterialTheme.typography.labelSmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                        )
                    }
                }
            }
        }
    }
}
