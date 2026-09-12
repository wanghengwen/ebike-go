package com.luopingtech.ebike.ops.ui.tools

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
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
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import kotlinx.coroutines.launch

@Composable
fun UnlockedVehiclesScreen(
    app: OpsApp,
    currentArea: ServiceArea?,
    onClose: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.unlockedVehicleFeature.state.collectAsState()
    val scope = rememberCoroutineScope()

    LaunchedEffect(currentArea?.id) {
        app.unlockedVehicleFeature.load(currentArea)
    }

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
            Text(t(Str.UnlockedVehicles), style = MaterialTheme.typography.headlineSmall)
            TextButton(onClick = {
                app.unlockedVehicleFeature.clear()
                onClose()
            }) { Text(t(Str.Back)) }
        }
        Text(
            text = t(Str.UnlockedVehiclesHint),
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
        if (state.canFilterStaff) {
            OutlinedTextField(
                value = state.query,
                onValueChange = { app.unlockedVehicleFeature.setQuery(it) },
                label = { Text(t(Str.UnlockedVehiclesSearch)) },
                modifier = Modifier.fillMaxWidth(),
                singleLine = true,
            )
        } else {
            Text(
                text = t(Str.UnlockedSelfOnlyHint),
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.tertiary,
            )
        }
        Button(
            onClick = { scope.launch { app.unlockedVehicleFeature.load(currentArea) } },
            enabled = !state.loading,
            modifier = Modifier.fillMaxWidth(),
        ) { Text(if (state.loading) t(Str.LoadingEllipsis) else t(Str.Refresh)) }

        state.vehicles.forEach { row ->
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                Column(modifier = Modifier.weight(1f)) {
                    Text(row.carId, style = MaterialTheme.typography.titleSmall)
                    Text(
                        text = "IMEI ${row.imei.ifBlank { "-" }}",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                }
                Button(
                    onClick = { scope.launch { app.unlockedVehicleFeature.lock(row.carId) } },
                    enabled = !state.loading && state.lockingCarId != row.carId,
                ) {
                    Text(
                        if (state.lockingCarId == row.carId) t(Str.LoadingEllipsis)
                        else t(Str.UnlockedLock),
                    )
                }
            }
        }
        if (!state.loading && state.vehicles.isEmpty()) {
            Text(
                text = t(Str.UnlockedVehiclesCount, 0),
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }
        state.message?.let { Text(it) }
        state.errorMessage?.let { Text(it, color = MaterialTheme.colorScheme.error) }
    }
}
