package com.luopingtech.ebike.ops.ui.task

import android.widget.Toast
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.MapPin
import com.luopingtech.ebike.ops.domain.model.OpsTask
import com.luopingtech.ebike.ops.domain.model.TeamWorker
import com.luopingtech.ebike.ops.domain.permission.OpsPermissions
import com.luopingtech.ebike.ops.domain.task.TaskWaitingTime
import com.luopingtech.ebike.ops.feature.task.ClaimableTaskFeature
import com.luopingtech.ebike.ops.ui.map.SimulatorMapView
import com.luopingtech.ebike.ops.ui.map.TencentMapView
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import kotlinx.coroutines.launch

/**
 * 巡检任务 —— 对齐 Flutter TaskPage：待领取/处理中 + 半图 + 勾选列表 + 领取/指派。
 */
@Composable
fun InspectionTaskScreen(
    app: OpsApp,
    onClose: () -> Unit,
    onOpenStats: () -> Unit,
) {
    ClaimableTaskMapListScreen(
        app = app,
        feature = app.inspectionTaskFeature,
        title = app.i18n.t(Str.InspectionTaskTitle),
        statsLabel = app.i18n.t(Str.InspectionStats),
        listMode = true,
        onClose = onClose,
        onOpenStats = onOpenStats,
    )
}

/**
 * 维修任务 —— 待领取以地图为主，点车再看详情；空态提示对齐截图。
 */
@Composable
fun RepairTaskScreen(
    app: OpsApp,
    onClose: () -> Unit,
    onOpenStats: () -> Unit,
) {
    ClaimableTaskMapListScreen(
        app = app,
        feature = app.repairTaskFeature,
        title = app.i18n.t(Str.RepairTaskTitle),
        statsLabel = app.i18n.t(Str.RepairStats),
        listMode = false,
        onClose = onClose,
        onOpenStats = onOpenStats,
    )
}

@Composable
private fun ClaimableTaskMapListScreen(
    app: OpsApp,
    feature: ClaimableTaskFeature,
    title: String,
    statsLabel: String,
    listMode: Boolean,
    onClose: () -> Unit,
    onOpenStats: () -> Unit,
) {
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val home by app.homeFeature.state.collectAsState()
    val state by feature.state.collectAsState()
    val staff by app.staffDirectoryFeature.state.collectAsState()
    val detailMap by app.vehicleDetailMapFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    val context = LocalContext.current
    val colors = OpsTheme.colors

    val permissions = remember(home.session?.permissionCodes, app.isDemoMode) {
        if (app.isDemoMode) {
            OpsPermissions.demoFull()
        } else {
            OpsPermissions.fromCodes(home.session?.permissionCodes.orEmpty())
        }
    }

    var tab by remember { mutableIntStateOf(0) } // 0 pending 1 processing
    var checked by remember { mutableStateOf(setOf<String>()) }
    var fitNonce by remember { mutableIntStateOf(0) }
    var onlyStopped by remember { mutableStateOf(false) }
    var assignOpen by remember { mutableStateOf(false) }
    var selectedWorker by remember { mutableStateOf<TeamWorker?>(null) }

    LaunchedEffect(home.currentArea?.id) {
        feature.load(home.currentArea)
        home.currentArea?.id?.takeIf { it.isNotBlank() }?.let {
            app.vehicleDetailMapFeature.loadFence(it)
        }
    }

    LaunchedEffect(assignOpen, home.currentArea?.id) {
        if (assignOpen) {
            app.staffDirectoryFeature.load(home.currentArea)
        }
    }

    val pending = state.tasks.filter { it.state == 0 }
    val processing = state.tasks.filter { it.state == 1 }
    val rawTabTasks = if (tab == 0) pending else processing
    val tabTasks = if (!listMode && onlyStopped && tab == 0) {
        rawTabTasks.filter { it.izStop }
    } else {
        rawTabTasks
    }
    val pins = remember(tabTasks) {
        tabTasks
            .filter { it.lat != 0.0 || it.lng != 0.0 }
            .map {
                MapPin(
                    id = it.id,
                    lat = it.lat,
                    lng = it.lng,
                    title = it.carId.ifBlank { it.id },
                    subtitle = it.address,
                    restBattery = it.restBattery,
                    memberIds = listOf(it.id),
                )
            }
    }
    val isTencent = home.mapProviderKind.equals("tencent", ignoreCase = true) && home.mapReady
    val fencePolygons = detailMap.fence?.all.orEmpty()
    val selected = state.selected

    Box(modifier = Modifier.fillMaxSize()) {
        Column(modifier = Modifier.fillMaxSize().background(Color.White)) {
            TaskFullscreenTopBar(
                title = title,
                primary = colors.primary,
                onBack = onClose,
                trailingLabel = statsLabel,
                onTrailing = onOpenStats,
            )
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(44.dp)
                    .background(Color.White),
            ) {
                listOf(
                    t(Str.TaskTabPending, pending.size),
                    t(Str.TaskTabProcessing, processing.size),
                ).forEachIndexed { i, label ->
                    Column(
                        modifier = Modifier
                            .weight(1f)
                            .fillMaxSize()
                            .clickable {
                                tab = i
                                checked = emptySet()
                                feature.selectTask(null)
                            },
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.Center,
                    ) {
                        Text(
                            label,
                            color = if (tab == i) colors.primary else Color(0xFF333333),
                            fontWeight = FontWeight.SemiBold,
                            fontSize = 14.sp,
                        )
                        Spacer(modifier = Modifier.height(4.dp))
                        Box(
                            modifier = Modifier
                                .width(if (tab == i) 36.dp else 0.dp)
                                .height(3.dp)
                                .background(if (tab == i) colors.primary else Color.Transparent),
                        )
                    }
                }
            }

            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(if (listMode) 220.dp else 280.dp),
            ) {
                if (isTencent) {
                    TencentMapView(
                        pins = pins,
                        selectedCarId = state.selectedTaskId,
                        onSelectCarId = { id ->
                            feature.selectTask(id)
                            if (tab == 0 && listMode) {
                                checked = if (id in checked) checked - id else checked + id
                            }
                        },
                        clusterOverview = true,
                        fencePolygons = fencePolygons,
                        fitNonce = fitNonce,
                        showStatusOverlay = false,
                        modifier = Modifier.fillMaxSize(),
                    )
                } else {
                    SimulatorMapView(
                        pins = pins,
                        selectedCarId = state.selectedTaskId,
                        providerLabel = title,
                        onSelectCarId = { id ->
                            feature.selectTask(id)
                            if (tab == 0 && listMode) {
                                checked = if (id in checked) checked - id else checked + id
                            }
                        },
                        clusterOverview = true,
                        modifier = Modifier.fillMaxSize(),
                    )
                }
            TaskMapSimpleSideTools(
                refreshLabel = t(Str.Refresh),
                locateLabel = t(Str.MapToolLocate),
                onRefresh = { scope.launch { feature.load(home.currentArea) } },
                onLocate = { fitNonce += 1 },
                modifier = Modifier
                    .align(Alignment.CenterStart)
                    .padding(start = 12.dp),
            )
                if (!listMode) {
                    Text(
                        text = t(Str.OnlyStoppedVehicles),
                        fontSize = 12.sp,
                        color = if (onlyStopped) colors.primary else Color(0xFF333333),
                        modifier = Modifier
                            .align(Alignment.BottomEnd)
                            .padding(12.dp)
                            .background(Color.White, RoundedCornerShape(8.dp))
                            .clickable { onlyStopped = !onlyStopped }
                            .padding(horizontal = 10.dp, vertical = 8.dp),
                    )
                }
            }

            when {
                state.loading && state.tasks.isEmpty() -> {
                    Column(
                        modifier = Modifier.weight(1f).fillMaxWidth(),
                        verticalArrangement = Arrangement.Center,
                        horizontalAlignment = Alignment.CenterHorizontally,
                    ) {
                        CircularProgressIndicator(color = colors.primary)
                    }
                }
                listMode -> {
                    LazyColumn(modifier = Modifier.weight(1f).fillMaxWidth()) {
                        items(tabTasks, key = { it.id }) { task ->
                            InspectionTaskRow(
                                app = app,
                                task = task,
                                processing = tab == 1,
                                checked = task.id in checked,
                                onToggle = {
                                    checked = if (task.id in checked) checked - task.id else checked + task.id
                                    feature.selectTask(task.id)
                                },
                            )
                        }
                    }
                }
                else -> {
                    Column(
                        modifier = Modifier
                            .weight(1f)
                            .fillMaxWidth()
                            .padding(16.dp),
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.Center,
                    ) {
                        if (selected == null || (tab == 0 && selected.state != 0)) {
                            Text(
                                t(Str.RepairMapEmptyHint),
                                color = Color(0xFFAAAAAA),
                                fontSize = 14.sp,
                            )
                        } else {
                            Text(
                                "${t(Str.VehicleId)}: ${selected.carId}",
                                color = colors.primary,
                                fontWeight = FontWeight.SemiBold,
                            )
                            Spacer(modifier = Modifier.height(8.dp))
                            Text(selected.address, color = Color(0xFF666666), fontSize = 13.sp)
                            Spacer(modifier = Modifier.height(6.dp))
                            Text(
                                "${t(Str.PendingDuration)}: ${waitingText(selected, tab == 1)}",
                                color = Color(0xFF666666),
                                fontSize = 12.sp,
                            )
                            if (tab == 0) {
                                Spacer(modifier = Modifier.height(4.dp))
                                Text(
                                    if (selected.izStop) t(Str.VehicleStopped) else t(Str.VehicleNotStopped),
                                    color = if (selected.izStop) Color(0xFFE53935) else Color(0xFF999999),
                                    fontSize = 12.sp,
                                )
                            }
                        }
                    }
                }
            }

            if (tab == 0) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(12.dp),
                    horizontalArrangement = Arrangement.spacedBy(12.dp),
                ) {
                    val count = if (listMode) checked.size else if (selected?.state == 0) 1 else 0
                    OutlinedButton(
                        onClick = {
                            scope.launch {
                                val ids = if (listMode) checked.toList() else listOfNotNull(selected?.id)
                                feature.claimMany(ids)
                                checked = emptySet()
                            }
                        },
                        modifier = Modifier.weight(1f).height(48.dp),
                        enabled = count > 0 && !state.loading,
                        colors = ButtonDefaults.outlinedButtonColors(contentColor = colors.primary),
                    ) {
                        Text("${t(Str.Claim)}($count)")
                    }
                    if (permissions.canAssignTask || app.isDemoMode) {
                        Button(
                            onClick = {
                                if (count <= 0) {
                                    Toast.makeText(context, t(Str.SelectTaskFirst), Toast.LENGTH_SHORT).show()
                                } else {
                                    selectedWorker = null
                                    assignOpen = true
                                }
                            },
                            modifier = Modifier.weight(1f).height(48.dp),
                            colors = ButtonDefaults.buttonColors(containerColor = colors.primary),
                        ) {
                            Text(t(Str.Assign))
                        }
                    }
                }
            }
            state.errorMessage?.let {
                Text(it, color = Color(0xFFE53935), modifier = Modifier.padding(12.dp))
            }
            state.message?.let {
                Text(it, color = colors.primary, modifier = Modifier.padding(12.dp))
            }
        }

        if (assignOpen) {
            AssignStaffSheet(
                app = app,
                loading = staff.loading,
                workers = staff.workers,
                error = staff.errorMessage,
                selected = selectedWorker,
                onSelect = { selectedWorker = it },
                onDismiss = { assignOpen = false },
                onConfirm = {
                    val pin = selectedWorker?.pin?.takeIf { it.isNotBlank() }
                        ?: selectedWorker?.selectionKey.orEmpty()
                    if (pin.isBlank()) {
                        Toast.makeText(context, t(Str.SelectStaffFirst), Toast.LENGTH_SHORT).show()
                        return@AssignStaffSheet
                    }
                    scope.launch {
                        val ids = if (listMode) checked.toList() else listOfNotNull(selected?.id)
                        when (feature.assignMany(ids, pin)) {
                            is OpsResult.Ok -> {
                                Toast.makeText(context, t(Str.AssignOk), Toast.LENGTH_SHORT).show()
                                checked = emptySet()
                                assignOpen = false
                            }
                            is OpsResult.Err -> Unit
                        }
                    }
                },
            )
        }
    }
}

private fun waitingText(task: OpsTask, processing: Boolean): String =
    if (processing) {
        TaskWaitingTime.format(task.createdAt, task.startTime.takeIf { it.isNotBlank() })
    } else {
        TaskWaitingTime.format(task.createdAt)
    }

@Composable
private fun AssignStaffSheet(
    app: OpsApp,
    loading: Boolean,
    workers: List<TeamWorker>,
    error: String?,
    selected: TeamWorker?,
    onSelect: (TeamWorker) -> Unit,
    onDismiss: () -> Unit,
    onConfirm: () -> Unit,
) {
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val colors = OpsTheme.colors
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color(0x99000000))
            .clickable(onClick = onDismiss),
    ) {
        Column(
            modifier = Modifier
                .align(Alignment.BottomCenter)
                .fillMaxWidth()
                .background(Color.White, RoundedCornerShape(topStart = 16.dp, topEnd = 16.dp))
                .clickable(enabled = false, onClick = {})
                .padding(bottom = 16.dp),
        ) {
            Text(
                t(Str.Assign),
                fontWeight = FontWeight.SemiBold,
                fontSize = 16.sp,
                modifier = Modifier.padding(16.dp),
            )
            when {
                loading -> {
                    Row(
                        modifier = Modifier.fillMaxWidth().padding(24.dp),
                        horizontalArrangement = Arrangement.Center,
                    ) {
                        CircularProgressIndicator(color = colors.primary)
                    }
                }
                error != null -> Text(error, color = Color(0xFFE53935), modifier = Modifier.padding(16.dp))
                workers.isEmpty() -> Text(t(Str.TeamWorkerNone), color = Color(0xFF999999), modifier = Modifier.padding(16.dp))
                else -> {
                    LazyColumn(modifier = Modifier.height(280.dp)) {
                        items(workers, key = { it.selectionKey }) { worker ->
                            val isSel = selected?.selectionKey == worker.selectionKey
                            Text(
                                worker.label,
                                color = if (isSel) colors.primary else Color(0xFF333333),
                                fontWeight = if (isSel) FontWeight.SemiBold else FontWeight.Normal,
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .clickable { onSelect(worker) }
                                    .padding(horizontal = 16.dp, vertical = 14.dp),
                            )
                            Box(modifier = Modifier.fillMaxWidth().height(1.dp).background(Color(0xFFE8E8E8)))
                        }
                    }
                }
            }
            Row(
                modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 8.dp),
                horizontalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                OutlinedButton(
                    onClick = onDismiss,
                    modifier = Modifier.weight(1f).height(44.dp),
                ) {
                    Text(t(Str.Cancel))
                }
                Button(
                    onClick = onConfirm,
                    modifier = Modifier.weight(1f).height(44.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = colors.primary),
                    enabled = selected != null && !loading,
                ) {
                    Text(t(Str.Confirm))
                }
            }
        }
    }
}

@Composable
private fun InspectionTaskRow(
    app: OpsApp,
    task: OpsTask,
    processing: Boolean,
    checked: Boolean,
    onToggle: () -> Unit,
) {
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val colors = OpsTheme.colors
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onToggle)
            .padding(horizontal = 12.dp, vertical = 14.dp),
        verticalAlignment = Alignment.Top,
    ) {
        Box(
            modifier = Modifier
                .padding(top = 2.dp)
                .size(18.dp)
                .border(1.dp, if (checked) colors.primary else Color(0xFFCCCCCC), RoundedCornerShape(3.dp))
                .background(if (checked) colors.primary else Color.Transparent, RoundedCornerShape(3.dp)),
        )
        Spacer(modifier = Modifier.width(10.dp))
        Column(modifier = Modifier.weight(1f)) {
            Row(modifier = Modifier.fillMaxWidth()) {
                Text(
                    "${t(Str.VehicleId)}: ",
                    color = Color(0xFF333333),
                    fontSize = 13.sp,
                )
                Text(task.carId, color = colors.primary, fontSize = 13.sp, fontWeight = FontWeight.Medium)
                Spacer(modifier = Modifier.weight(1f))
                Text(
                    "${t(Str.PendingDuration)}: ${waitingText(task, processing)}",
                    color = Color(0xFF666666),
                    fontSize = 12.sp,
                )
            }
            Spacer(modifier = Modifier.height(6.dp))
            Text(
                "${t(Str.TaskSource)}: ${t(Str.OpsRulesSource)}",
                color = Color(0xFF666666),
                fontSize = 12.sp,
            )
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                "${t(Str.AlarmStatus)}: ${task.address.ifBlank { "-" }}",
                color = Color(0xFF999999),
                fontSize = 12.sp,
            )
        }
    }
    Box(modifier = Modifier.fillMaxWidth().height(1.dp).background(Color(0xFFE8E8E8)))
}
