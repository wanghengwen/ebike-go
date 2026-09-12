package com.luopingtech.ebike.ops.ui.staff

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.ui.analysis.VcdTopBar
import com.luopingtech.ebike.ops.ui.theme.OpsTheme

@Composable
fun StaffDirectoryScreen(
    app: OpsApp,
    onClose: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val home by app.homeFeature.state.collectAsState()
    val state by app.staffDirectoryFeature.state.collectAsState()

    LaunchedEffect(home.currentArea?.id) {
        app.staffDirectoryFeature.load(home.currentArea)
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(Color(0xFFF6F7F9)),
    ) {
        VcdTopBar(title = t(Str.StaffManage), onBack = {
            app.staffDirectoryFeature.clear()
            onClose()
        })
        Column(
            modifier = Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState())
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            Text(
                text = t(Str.StaffPcHint),
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
            if (state.loading) {
                CircularProgressIndicator(color = OpsTheme.colors.primary)
            } else if (state.errorMessage != null) {
                Text(text = state.errorMessage.orEmpty(), color = MaterialTheme.colorScheme.error)
            } else if (state.workers.isEmpty()) {
                Text(t(Str.AdminEmptyList))
            } else {
                state.workers.forEach { worker ->
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .background(Color.White, MaterialTheme.shapes.medium)
                            .padding(12.dp),
                    ) {
                        Text(worker.name, style = MaterialTheme.typography.titleSmall)
                        Text(worker.phone, style = MaterialTheme.typography.bodyMedium)
                    }
                }
            }
        }
    }
}
