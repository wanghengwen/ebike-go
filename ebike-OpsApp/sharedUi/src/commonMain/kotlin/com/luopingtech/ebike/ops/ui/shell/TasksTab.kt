package com.luopingtech.ebike.ops.ui.shell

import com.luopingtech.ebike.ops.OpsApp
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.setValue
import com.luopingtech.ebike.ops.ui.icons.OpsIcon
import androidx.compose.foundation.layout.size
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.domain.model.OpsTask
import com.luopingtech.ebike.ops.domain.permission.OpsPermissions
import com.luopingtech.ebike.ops.feature.home.HomeUiState
import com.luopingtech.ebike.ops.ui.task.TaskCenterCardItem
import com.luopingtech.ebike.ops.ui.task.TaskScaffold

@Composable
internal fun TasksTab(
    app: OpsApp,
    homeState: HomeUiState,
    permissions: OpsPermissions,
    onChangeArea: () -> Unit,
    onOpenChangeBatteryMap: () -> Unit,
    onOpenMoveCarMap: () -> Unit,
    onOpenInspectionTask: () -> Unit,
    onOpenRepairTask: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val battery by app.changeBatteryTaskFeature.state.collectAsState()
    val move by app.moveCarTaskFeature.state.collectAsState()
    val inspection by app.inspectionTaskFeature.state.collectAsState()
    val repair by app.repairTaskFeature.state.collectAsState()

    LaunchedEffect(homeState.currentArea?.id) {
        val area = homeState.currentArea ?: return@LaunchedEffect
        if (permissions.showChangeBattery) app.changeBatteryTaskFeature.load(area)
        if (permissions.showMoveCar) app.moveCarTaskFeature.load(area)
        if (permissions.showInspection) app.inspectionTaskFeature.load(area)
        if (permissions.showRepair) app.repairTaskFeature.load(area)
    }

    val cards = buildList {
        if (permissions.showChangeBattery || app.isDemoMode) {
            add(
                TaskCenterCardItem(
                    id = "battery",
                    title = t(Str.ChangeBatteryTaskTitle),
                    icon = OpsIcon.TaskChangeBattery,
                    badgeText = badgeTotalCount(battery.tasks.size),
                    onClick = onOpenChangeBatteryMap,
                ),
            )
        }
        if (permissions.showMoveCar || app.isDemoMode) {
            add(
                TaskCenterCardItem(
                    id = "move",
                    title = t(Str.MoveCarTaskTitle),
                    icon = OpsIcon.TaskMoveBike,
                    badgeText = badgeTotalCount(move.tasks.size),
                    onClick = onOpenMoveCarMap,
                ),
            )
        }
        if (permissions.showInspection || app.isDemoMode) {
            add(
                TaskCenterCardItem(
                    id = "inspection",
                    title = t(Str.InspectionTaskTitle),
                    icon = OpsIcon.TaskInspection,
                    badgeText = badgeClaimed(inspection.tasks),
                    onClick = onOpenInspectionTask,
                ),
            )
        }
        if (permissions.showRepair || app.isDemoMode) {
            add(
                TaskCenterCardItem(
                    id = "repair",
                    title = t(Str.RepairTaskTitle),
                    icon = OpsIcon.TaskRepair,
                    badgeText = badgeClaimed(repair.tasks),
                    onClick = onOpenRepairTask,
                ),
            )
        }
    }
    TaskScaffold(
        areaName = homeState.currentArea?.name?.takeIf { it.isNotBlank() }
            ?: t(Str.SelectServiceArea),
        pageTitle = t(Str.TaskCenter),
        cards = cards,
        onChangeArea = onChangeArea,
        emptyHint = t(Str.NoTaskPermission),
    )
}

/** Legacy: hide tip when total is 0. */
private fun badgeTotalCount(total: Int): String? =
    if (total <= 0) null else total.toString()

/** Legacy inspection/repair tip: received/total among pending+processing; hide when 0. */
private fun badgeClaimed(tasks: List<com.luopingtech.ebike.ops.domain.model.OpsTask>): String? {
    val active = tasks.filter { it.state == 0 || it.state == 1 }
    if (active.isEmpty()) return null
    val claimed = active.count { it.state == 1 }
    return "$claimed/${active.size}"
}
