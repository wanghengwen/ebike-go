package com.luopingtech.ebike.ops.ui.task

import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.material3.Button
import androidx.compose.material3.FilterChip
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
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
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.domain.model.MapPin
import com.luopingtech.ebike.ops.domain.model.OpsTask
import com.luopingtech.ebike.ops.domain.permission.OpsPermissions
import com.luopingtech.ebike.ops.ui.map.SimulatorMapView
import com.luopingtech.ebike.ops.ui.map.TencentMapView
import kotlinx.coroutines.launch

enum class TaskMapKind {
    ChangeBattery,
    MoveCar,
}

/**
 * Independent task map (legacy TaskMapActivity):
 * fence overlay, battery filter, cluster toggle, pin actions (ring / open / finish).
 * List workflow on Tasks tab remains available.
 */
@Composable
fun TaskMapScreen(
    app: OpsApp,
    permissions: OpsPermissions,
    initialKind: TaskMapKind,
    onClose: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val home by app.homeFeature.state.collectAsState()
    val changeState by app.changeBatteryTaskFeature.state.collectAsState()
    val moveState by app.moveCarTaskFeature.state.collectAsState()
    val detailMap by app.vehicleDetailMapFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    var kind by remember { mutableStateOf(initialKind) }
    var selectedId by remember { mutableStateOf<String?>(null) }
    var clusterIds by remember { mutableStateOf<List<String>?>(null) }
    var clusterOverview by remember { mutableStateOf(true) }
    var showFence by remember { mutableStateOf(true) }
    var actionMessage by remember { mutableStateOf<String?>(null) }

    LaunchedEffect(kind, home.currentArea?.id) {
        val area = home.currentArea ?: return@LaunchedEffect
        when (kind) {
            TaskMapKind.ChangeBattery -> app.changeBatteryTaskFeature.load(area)
            TaskMapKind.MoveCar -> app.moveCarTaskFeature.load(area)
        }
        selectedId = null
        clusterIds = null
        actionMessage = null
        val sid = area.id
        if (sid.isNotBlank()) {
            app.vehicleDetailMapFeature.loadFence(sid)
        }
    }

    val tasks = when (kind) {
        TaskMapKind.ChangeBattery -> changeState.tasks
        TaskMapKind.MoveCar -> moveState.tasks
    }
    val pins = remember(tasks) {
        tasks
            .filter { it.lat != 0.0 || it.lng != 0.0 }
            .map {
                MapPin(
                    id = it.carId.ifBlank { it.id },
                    lat = it.lat,
                    lng = it.lng,
                    title = it.carId.ifBlank { it.id },
                    subtitle = it.address.ifBlank { "${it.restBattery}%" },
                    restBattery = it.restBattery,
                    memberCount = 1,
                    memberIds = listOf(it.carId.ifBlank { it.id }),
                )
            }
    }
    val selectedTask: OpsTask? = remember(tasks, selectedId) {
        val id = selectedId ?: return@remember null
        tasks.firstOrNull { it.carId == id || it.id == id }
    }
    val batteryPresets = remember(changeState.rangeMin, changeState.rangeMax) {
        listOf(changeState.rangeMin, 20, 30, changeState.rangeMax)
            .filter { it in changeState.rangeMin..changeState.rangeMax }
            .distinct()
            .sorted()
    }
    val fencePolygons = if (showFence) detailMap.fence?.all.orEmpty() else emptyList()

    fun selectCar(id: String) {
        selectedId = id
        clusterIds = null
        actionMessage = null
        val task = tasks.firstOrNull { it.carId == id || it.id == id }
        when (kind) {
            TaskMapKind.ChangeBattery -> app.changeBatteryTaskFeature.selectTask(task?.id)
            TaskMapKind.MoveCar -> app.moveCarTaskFeature.selectTask(task?.id)
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(12.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
        ) {
            Text(t(Str.TaskMap), style = MaterialTheme.typography.titleMedium)
            TextButton(onClick = onClose) { Text(t(Str.Close)) }
        }
        Text(
            t(Str.TaskMapHint),
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .horizontalScroll(rememberScrollState()),
            horizontalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            if (permissions.showChangeBattery) {
                FilterChip(
                    selected = kind == TaskMapKind.ChangeBattery,
                    onClick = { kind = TaskMapKind.ChangeBattery },
                    label = { Text(t(Str.ChangeBatteryShort)) },
                )
            }
            if (permissions.showMoveCar) {
                FilterChip(
                    selected = kind == TaskMapKind.MoveCar,
                    onClick = { kind = TaskMapKind.MoveCar },
                    label = { Text(t(Str.MoveCarShort)) },
                )
            }
            FilterChip(
                selected = showFence,
                onClick = { showFence = !showFence },
                label = { Text(t(Str.VehicleFence)) },
            )
            FilterChip(
                selected = clusterOverview,
                onClick = { clusterOverview = !clusterOverview },
                label = {
                    Text(if (clusterOverview) t(Str.MapClusterOn) else t(Str.MapClusterOff))
                },
            )
            TextButton(onClick = {
                scope.launch {
                    val area = home.currentArea ?: return@launch
                    when (kind) {
                        TaskMapKind.ChangeBattery -> app.changeBatteryTaskFeature.load(area)
                        TaskMapKind.MoveCar -> app.moveCarTaskFeature.load(area)
                    }
                    if (area.id.isNotBlank()) {
                        app.vehicleDetailMapFeature.loadFence(area.id)
                    }
                }
            }) { Text(t(Str.Refresh)) }
        }

        if (kind == TaskMapKind.ChangeBattery) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .horizontalScroll(rememberScrollState()),
                horizontalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                batteryPresets.forEach { value ->
                    FilterChip(
                        selected = changeState.maxBattery == value,
                        onClick = {
                            scope.launch {
                                app.changeBatteryTaskFeature.applyMaxBattery(home.currentArea, value)
                            }
                        },
                        label = { Text("≤$value%") },
                    )
                }
            }
        }

        Text(t(Str.ShowCountOfTotal, pins.size, tasks.size))

        val isTencent = home.mapProviderKind.equals("tencent", ignoreCase = true) && home.mapReady
        if (isTencent) {
            TencentMapView(
                pins = pins,
                selectedCarId = selectedId,
                onSelectCarId = { selectCar(it) },
                onSelectCluster = { ids ->
                    if (ids.size <= 1) {
                        ids.firstOrNull()?.let { selectCar(it) }
                    } else {
                        clusterOverview = false
                        clusterIds = ids
                        selectedId = null
                    }
                },
                clusterOverview = clusterOverview,
                fencePolygons = fencePolygons,
                modifier = Modifier
                    .fillMaxWidth()
                    .weight(1f),
            )
        } else {
            SimulatorMapView(
                pins = pins,
                selectedCarId = selectedId,
                providerLabel = t(Str.TaskMap),
                onSelectCarId = { selectCar(it) },
                onSelectCluster = { ids ->
                    if (ids.size <= 1) {
                        ids.firstOrNull()?.let { selectCar(it) }
                    } else {
                        clusterOverview = false
                        clusterIds = ids
                        selectedId = null
                    }
                },
                clusterOverview = clusterOverview,
                modifier = Modifier
                    .fillMaxWidth()
                    .weight(1f),
            )
        }

        clusterIds?.takeIf { it.size > 1 }?.let { ids ->
            Surface(tonalElevation = 2.dp, modifier = Modifier.fillMaxWidth()) {
                Column(
                    modifier = Modifier.padding(12.dp),
                    verticalArrangement = Arrangement.spacedBy(6.dp),
                ) {
                    Text(t(Str.ClusterVehicles, ids.size), style = MaterialTheme.typography.titleSmall)
                    ids.take(8).forEach { id ->
                        TextButton(onClick = { selectCar(id) }) { Text(id) }
                    }
                    TextButton(onClick = { clusterIds = null }) { Text(t(Str.Close)) }
                }
            }
        }

        selectedTask?.let { task ->
            TaskMapVehicleSheet(
                app = app,
                kind = kind,
                task = task,
                loading = when (kind) {
                    TaskMapKind.ChangeBattery -> changeState.loading
                    TaskMapKind.MoveCar -> moveState.loading
                },
                message = actionMessage
                    ?: when (kind) {
                        TaskMapKind.ChangeBattery -> changeState.message ?: changeState.errorMessage
                        TaskMapKind.MoveCar -> moveState.message ?: moveState.errorMessage
                    },
                onRing = {
                    scope.launch {
                        actionMessage = null
                        when (kind) {
                            TaskMapKind.ChangeBattery -> app.changeBatteryTaskFeature.ringSelected()
                            TaskMapKind.MoveCar -> app.moveCarTaskFeature.ringSelected()
                        }
                    }
                },
                onOpenBox = {
                    scope.launch {
                        actionMessage = null
                        app.changeBatteryTaskFeature.openBox()
                    }
                },
                onFinishSwap = {
                    scope.launch {
                        actionMessage = null
                        app.changeBatteryTaskFeature.closeBox()
                    }
                },
                onDismiss = {
                    selectedId = null
                    when (kind) {
                        TaskMapKind.ChangeBattery -> app.changeBatteryTaskFeature.selectTask(null)
                        TaskMapKind.MoveCar -> app.moveCarTaskFeature.selectTask(null)
                    }
                },
            )
        }
    }
}

@Composable
private fun TaskMapVehicleSheet(
    app: OpsApp,
    kind: TaskMapKind,
    task: OpsTask,
    loading: Boolean,
    message: String?,
    onRing: () -> Unit,
    onOpenBox: () -> Unit,
    onFinishSwap: () -> Unit,
    onDismiss: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    Surface(tonalElevation = 3.dp, modifier = Modifier.fillMaxWidth()) {
        Column(
            modifier = Modifier.padding(12.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
            ) {
                Text(
                    text = "${task.carId} · ${task.restBattery}%",
                    style = MaterialTheme.typography.titleSmall,
                )
                TextButton(onClick = onDismiss) { Text(t(Str.Close)) }
            }
            if (task.address.isNotBlank()) {
                Text(
                    task.address,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            }
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                Button(
                    onClick = onRing,
                    enabled = !loading,
                    modifier = Modifier.weight(1f),
                ) { Text(t(Str.Ring)) }
                if (kind == TaskMapKind.ChangeBattery) {
                    Button(
                        onClick = onOpenBox,
                        enabled = !loading,
                        modifier = Modifier.weight(1f),
                    ) { Text(t(Str.OpenBatteryBox)) }
                    Button(
                        onClick = onFinishSwap,
                        enabled = !loading,
                        modifier = Modifier.weight(1f),
                    ) { Text(t(Str.FinishChangeBattery)) }
                }
            }
            message?.let {
                Text(
                    it,
                    style = MaterialTheme.typography.bodySmall,
                    color = if (it.contains("fail", ignoreCase = true) ||
                        it.contains("失败")
                    ) {
                        MaterialTheme.colorScheme.error
                    } else {
                        MaterialTheme.colorScheme.onSurface
                    },
                )
            }
        }
    }
}
