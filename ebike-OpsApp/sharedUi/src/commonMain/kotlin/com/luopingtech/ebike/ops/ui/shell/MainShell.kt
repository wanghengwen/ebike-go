package com.luopingtech.ebike.ops.ui.shell

import com.luopingtech.ebike.ops.OpsApp
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.ui.analysis.OfflineOpsScreen
import com.luopingtech.ebike.ops.ui.analysis.ReturnCarAnalysisMapScreen
import com.luopingtech.ebike.ops.ui.analysis.StationAnalysisScreen
import com.luopingtech.ebike.ops.ui.analysis.VehicleConditionDistributionMapScreen
import com.luopingtech.ebike.ops.ui.analysis.VehicleConditionDistributionScreen
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.ui.graphics.Color
import androidx.compose.foundation.layout.size
import androidx.compose.material3.NavigationBarItemDefaults
import com.luopingtech.ebike.ops.core.config.H5ScreenKind
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.domain.permission.OpsPermissions
import com.luopingtech.ebike.ops.feature.home.HomeUiState
import com.luopingtech.ebike.ops.ui.feedback.LocalOpsToast
import com.luopingtech.ebike.ops.ui.h5.H5Screen
import com.luopingtech.ebike.ops.ui.permission.LocalOpsTrackPermissionGate
import com.luopingtech.ebike.ops.ui.scan.LocalOpsScanPreview
import com.luopingtech.ebike.ops.ui.movecar.FieldMoveCarScreen
import com.luopingtech.ebike.ops.ui.order.OrderQueryScreen
import com.luopingtech.ebike.ops.ui.production.ProductionScreen
import com.luopingtech.ebike.ops.ui.report.FaultReportScreen
import com.luopingtech.ebike.ops.ui.task.BatchMoveCarSection
import com.luopingtech.ebike.ops.ui.task.ChangeBatteryMapScreen
import com.luopingtech.ebike.ops.ui.task.FreeMoveCarSection
import com.luopingtech.ebike.ops.ui.task.InspectionTaskScreen
import com.luopingtech.ebike.ops.ui.task.MoveCarMapScreen
import com.luopingtech.ebike.ops.ui.task.RepairTaskScreen
import com.luopingtech.ebike.ops.ui.task.TaskFullscreenTopBar
import com.luopingtech.ebike.ops.ui.task.TaskStatisticsScreen
import com.luopingtech.ebike.ops.domain.analysis.TaskStatisticsKind
import com.luopingtech.ebike.ops.ui.relocation.RelocationScreen
import com.luopingtech.ebike.ops.ui.tools.FieldChangeBatteryScreen
import com.luopingtech.ebike.ops.ui.tools.UnlockedVehiclesScreen
import com.luopingtech.ebike.ops.ui.tools.OpsSettingScreen
import com.luopingtech.ebike.ops.ui.tools.BluetoothRadarScreen
import com.luopingtech.ebike.ops.ui.fence.FenceBrowseScreen
import com.luopingtech.ebike.ops.ui.staff.StaffDirectoryScreen
import com.luopingtech.ebike.ops.ui.tag.VehicleTagScreen
import com.luopingtech.ebike.ops.ui.admin.ProfessionAuditScreen
import com.luopingtech.ebike.ops.ui.admin.ObjectionOrderScreen
import com.luopingtech.ebike.ops.ui.admin.BlacklistScreen
import com.luopingtech.ebike.ops.ui.admin.IdBindAuditScreen
import com.luopingtech.ebike.ops.ui.admin.OperationLogScreen
import com.luopingtech.ebike.ops.ui.sneak.SneakReportScreen
import com.luopingtech.ebike.ops.ui.workorder.WorkOrderScreen
import com.luopingtech.ebike.ops.domain.model.WorkOrderKind
import com.luopingtech.ebike.ops.ui.vehicle.VehicleListScreen
import com.luopingtech.ebike.ops.ui.warehouse.WarehouseScreen
import kotlinx.coroutines.launch

private enum class MainTab { Map, Tasks, Analysis, Workbench }

@Composable
internal fun MainShell(
    app: OpsApp,
    homeState: HomeUiState,
    onChangeArea: () -> Unit,
) {
    val permissions = remember(homeState.session?.permissionCodes, app.isDemoMode) {
        when {
            app.isDemoMode && homeState.session?.permissionCodes.isNullOrEmpty() ->
                OpsPermissions.demoFull()
            else -> OpsPermissions.fromCodes(homeState.session?.permissionCodes.orEmpty())
        }
    }
    var tab by remember { mutableStateOf(MainTab.Map) }
    var warehouseOpen by remember { mutableStateOf(false) }
    var changeBatteryMapOpen by remember { mutableStateOf(false) }
    var moveCarMapOpen by remember { mutableStateOf(false) }
    var inspectionTaskOpen by remember { mutableStateOf(false) }
    var repairTaskOpen by remember { mutableStateOf(false) }
    var taskStatisticsKind by remember { mutableStateOf<TaskStatisticsKind?>(null) }
    var taskChangeAreaOpen by remember { mutableStateOf(false) }
    var freeMoveFromTaskOpen by remember { mutableStateOf(false) }
    var batchMoveFromTaskOpen by remember { mutableStateOf(false) }
    var batchParentTaskId by remember { mutableStateOf<String?>(null) }
    var vehicleConditionDistOpen by remember { mutableStateOf(false) }
    var vehicleConditionDistMapOpen by remember { mutableStateOf(false) }
    var offlineOpsOpen by remember { mutableStateOf(false) }
    var stationAnalysisOpen by remember { mutableStateOf(false) }
    var returnCarAnalysisOpen by remember { mutableStateOf(false) }
    var productionOpen by remember { mutableStateOf(false) }
    var faultReportOpen by remember { mutableStateOf(false) }
    var sneakReportOpen by remember { mutableStateOf(false) }
    var unlockedVehiclesOpen by remember { mutableStateOf(false) }
    var myTaskOpen by remember { mutableStateOf(false) }
    var fenceBrowseOpen by remember { mutableStateOf(false) }
    var opsSettingOpen by remember { mutableStateOf(false) }
    var vehicleTagOpen by remember { mutableStateOf(false) }
    var bluetoothRadarOpen by remember { mutableStateOf(false) }
    var staffDirectoryOpen by remember { mutableStateOf(false) }
    var professionAuditOpen by remember { mutableStateOf(false) }
    var objectionOrderOpen by remember { mutableStateOf(false) }
    var blacklistOpen by remember { mutableStateOf(false) }
    var idBindAuditOpen by remember { mutableStateOf(false) }
    var operationLogOpen by remember { mutableStateOf(false) }
    var vehicleListOpen by remember { mutableStateOf(false) }
    var fieldChangeBatteryOpen by remember { mutableStateOf(false) }
    var fieldMoveCarOpen by remember { mutableStateOf(false) }
    var orderQueryOpen by remember { mutableStateOf(false) }
    var relocationOpen by remember { mutableStateOf(false) }
    var inspectionOrderOpen by remember { mutableStateOf(false) }
    var repairOrderOpen by remember { mutableStateOf(false) }
    var h5Screen by remember { mutableStateOf<H5ScreenKind?>(null) }
    var scanOpen by remember { mutableStateOf(false) }
    var trackPermissionHint by remember { mutableStateOf<String?>(null) }
    val trackPermissionGate = LocalOpsTrackPermissionGate.current
    val toast = LocalOpsToast.current
    val scanner = LocalOpsScanPreview.current

    // 登录完成就接着上报轨迹（进程被杀 / 重启由宿主的常驻服务兜）。
    // 「拿到许可就开上报」是业务规则，留在共享层；怎么向用户要权限是平台的事。
    LaunchedEffect(homeState.session?.userId) {
        if (homeState.session == null) return@LaunchedEffect
        if (app.trackUploadFeature.state.value.enabled) return@LaunchedEffect
        trackPermissionGate.request { denied ->
            trackPermissionHint = denied
            if (denied == null) app.trackUploadFeature.setEnabled(true)
        }
    }

    if (myTaskOpen) {
        OfflineOpsScreen(
            app = app,
            mineOnly = true,
            onClose = {
                app.offlineOpsFeature.clear()
                myTaskOpen = false
            },
        )
        return
    }
    if (fenceBrowseOpen) {
        FenceBrowseScreen(app = app, onClose = { fenceBrowseOpen = false })
        return
    }
    if (opsSettingOpen) {
        OpsSettingScreen(app = app, onClose = { opsSettingOpen = false })
        return
    }
    if (vehicleTagOpen) {
        VehicleTagScreen(app = app, onClose = { vehicleTagOpen = false })
        return
    }
    if (bluetoothRadarOpen) {
        BluetoothRadarScreen(app = app, onClose = { bluetoothRadarOpen = false })
        return
    }
    if (staffDirectoryOpen) {
        StaffDirectoryScreen(app = app, onClose = { staffDirectoryOpen = false })
        return
    }
    if (professionAuditOpen) {
        ProfessionAuditScreen(app = app, onClose = { professionAuditOpen = false })
        return
    }
    if (objectionOrderOpen) {
        ObjectionOrderScreen(app = app, onClose = { objectionOrderOpen = false })
        return
    }
    if (blacklistOpen) {
        BlacklistScreen(app = app, onClose = { blacklistOpen = false })
        return
    }
    if (idBindAuditOpen) {
        IdBindAuditScreen(app = app, onClose = { idBindAuditOpen = false })
        return
    }
    if (operationLogOpen) {
        OperationLogScreen(app = app, onClose = { operationLogOpen = false })
        return
    }
    if (taskStatisticsKind != null) {
        TaskStatisticsScreen(
            app = app,
            kind = taskStatisticsKind!!,
            onClose = { taskStatisticsKind = null },
        )
        return
    }
    if (changeBatteryMapOpen) {
        Box(modifier = Modifier.fillMaxSize()) {
            ChangeBatteryMapScreen(
                app = app,
                onClose = {
                    taskChangeAreaOpen = false
                    changeBatteryMapOpen = false
                },
                onOpenStats = { taskStatisticsKind = TaskStatisticsKind.ChangeBattery },
                onChangeArea = { taskChangeAreaOpen = true },
            )
            if (taskChangeAreaOpen) {
                AreaGateScreen(
                    app = app,
                    homeState = homeState,
                    allowCancel = true,
                    onCancel = { taskChangeAreaOpen = false },
                    onSelected = { taskChangeAreaOpen = false },
                )
            }
        }
        return
    }
    if (moveCarMapOpen) {
        Box(modifier = Modifier.fillMaxSize()) {
            MoveCarMapScreen(
                app = app,
                onClose = {
                    taskChangeAreaOpen = false
                    moveCarMapOpen = false
                },
                onOpenFreeMove = {
                    freeMoveFromTaskOpen = true
                },
                onOpenStats = { taskStatisticsKind = TaskStatisticsKind.MoveCar },
                onChangeArea = { taskChangeAreaOpen = true },
            )
            if (taskChangeAreaOpen) {
                AreaGateScreen(
                    app = app,
                    homeState = homeState,
                    allowCancel = true,
                    onCancel = { taskChangeAreaOpen = false },
                    onSelected = { taskChangeAreaOpen = false },
                )
            }
        }
        return
    }
    if (freeMoveFromTaskOpen) {
        Column(modifier = Modifier.fillMaxSize().background(Color.White)) {
            TaskFullscreenTopBar(
                title = app.i18n.t(Str.FreeMoveTaskTitle),
                primary = OpsTheme.colors.primary,
                onBack = { freeMoveFromTaskOpen = false },
            )
            FreeMoveCarSection(app = app, currentArea = homeState.currentArea)
        }
        return
    }
    if (batchMoveFromTaskOpen) {
        Column(modifier = Modifier.fillMaxSize().background(Color.White)) {
            TaskFullscreenTopBar(
                title = app.i18n.t(Str.BatchMoveTaskTitle),
                primary = OpsTheme.colors.primary,
                onBack = {
                    app.batchMoveCarFeature.clear()
                    batchParentTaskId = null
                    batchMoveFromTaskOpen = false
                },
            )
            BatchMoveCarSection(app = app, parentTaskId = batchParentTaskId)
        }
        return
    }
    if (inspectionTaskOpen) {
        InspectionTaskScreen(
            app = app,
            onClose = { inspectionTaskOpen = false },
            onOpenStats = { taskStatisticsKind = TaskStatisticsKind.Inspection },
        )
        return
    }
    if (repairTaskOpen) {
        RepairTaskScreen(
            app = app,
            onClose = { repairTaskOpen = false },
            onOpenStats = { taskStatisticsKind = TaskStatisticsKind.Repair },
        )
        return
    }
    if (offlineOpsOpen) {
        OfflineOpsScreen(
            app = app,
            onClose = { offlineOpsOpen = false },
        )
        return
    }
    if (stationAnalysisOpen) {
        StationAnalysisScreen(
            app = app,
            onClose = { stationAnalysisOpen = false },
        )
        return
    }
    if (returnCarAnalysisOpen) {
        ReturnCarAnalysisMapScreen(
            app = app,
            onClose = { returnCarAnalysisOpen = false },
        )
        return
    }
    if (vehicleConditionDistMapOpen) {
        VehicleConditionDistributionMapScreen(
            app = app,
            onClose = { vehicleConditionDistMapOpen = false },
        )
        return
    }
    if (vehicleConditionDistOpen) {
        VehicleConditionDistributionScreen(
            app = app,
            onClose = { vehicleConditionDistOpen = false },
            onOpenMap = { vehicleConditionDistMapOpen = true },
        )
        return
    }
    if (warehouseOpen) {
        WarehouseScreen(
            app = app,
            permissions = permissions,
            onClose = {
                app.warehouseFeature.clear()
                warehouseOpen = false
            },
        )
        return
    }
    if (productionOpen) {
        ProductionScreen(
            app = app,
            permissions = permissions,
            onClose = { productionOpen = false },
            onLocateOnMap = {
                productionOpen = false
                tab = MainTab.Map
            },
        )
        return
    }
    if (faultReportOpen) {
        FaultReportScreen(
            app = app,
            onClose = { faultReportOpen = false },
        )
        return
    }
    if (sneakReportOpen) {
        SneakReportScreen(
            app = app,
            onClose = { sneakReportOpen = false },
        )
        return
    }
    if (unlockedVehiclesOpen) {
        UnlockedVehiclesScreen(
            app = app,
            currentArea = homeState.currentArea,
            onClose = { unlockedVehiclesOpen = false },
        )
        return
    }
    if (vehicleListOpen) {
        VehicleListScreen(
            app = app,
            currentArea = homeState.currentArea,
            onClose = { vehicleListOpen = false },
            onChangeArea = onChangeArea,
            onVehicleClick = { carId ->
                app.vehicleFeature.selectVehicle(carId)
                vehicleListOpen = false
                tab = MainTab.Map
            },
        )
        return
    }
    if (fieldChangeBatteryOpen) {
        FieldChangeBatteryScreen(
            app = app,
            onClose = { fieldChangeBatteryOpen = false },
            onHelp = { toast(app.i18n.t(Str.FieldChangeBatteryHelp)) },
            scanPreview = { modifier, torchOn, enabled, onCode ->
                scanner.Preview(modifier, torchOn, enabled, onCode)
            },
        )
        return
    }
    if (fieldMoveCarOpen) {
        FieldMoveCarScreen(
            app = app,
            currentArea = homeState.currentArea,
            onClose = { fieldMoveCarOpen = false },
            onBleSearch = { toast(app.i18n.t(Str.FeatureComingSoon)) },
            scanPreview = { modifier, torchOn, enabled, onCode ->
                scanner.Preview(modifier, torchOn, enabled, onCode)
            },
        )
        return
    }
    if (orderQueryOpen) {
        OrderQueryScreen(
            app = app,
            currentArea = homeState.currentArea,
            onClose = { orderQueryOpen = false },
            onOpenVehicleDetail = { carId ->
                app.vehicleFeature.selectVehicle(carId)
                orderQueryOpen = false
                tab = MainTab.Map
            },
        )
        return
    }
    if (relocationOpen) {
        RelocationScreen(
            app = app,
            currentArea = homeState.currentArea,
            onClose = { relocationOpen = false },
        )
        return
    }
    if (inspectionOrderOpen) {
        WorkOrderScreen(
            app = app,
            kind = WorkOrderKind.Inspection,
            currentArea = homeState.currentArea,
            onClose = { inspectionOrderOpen = false },
        )
        return
    }
    if (repairOrderOpen) {
        WorkOrderScreen(
            app = app,
            kind = WorkOrderKind.Repair,
            currentArea = homeState.currentArea,
            onClose = { repairOrderOpen = false },
        )
        return
    }
    if (h5Screen != null) {
        H5Screen(
            app = app,
            kind = h5Screen!!,
            onClose = { h5Screen = null },
        )
        return
    }
    if (scanOpen) {
        ScanOverlay(
            app = app,
            permissions = permissions,
            onClose = {
                app.scanFeature.clear()
                scanOpen = false
            },
        )
        return
    }

    LaunchedEffect(permissions.showMap, permissions.showTaskCenter, tab) {
        val allowed = buildList {
            if (permissions.showMap) add(MainTab.Map)
            if (permissions.showTaskCenter) add(MainTab.Tasks)
            add(MainTab.Analysis)
            add(MainTab.Workbench)
        }
        if (tab !in allowed) {
            tab = allowed.first()
        }
    }

    Scaffold(
        containerColor = Color.White,
        bottomBar = {
            NavigationBar(
                containerColor = Color.White,
                tonalElevation = 0.dp,
            ) {
                val itemColors = NavigationBarItemDefaults.colors(
                    selectedIconColor = OpsTheme.colors.primary,
                    selectedTextColor = OpsTheme.colors.primary,
                    indicatorColor = Color.Transparent,
                    unselectedIconColor = Color(0xFF666666),
                    unselectedTextColor = Color(0xFF666666),
                )
                if (permissions.showMap) {
                    NavigationBarItem(
                        selected = tab == MainTab.Map,
                        onClick = { tab = MainTab.Map },
                        colors = itemColors,
                        icon = { Text("◉", fontSize = 16.sp) },
                        label = { Text(app.i18n.t(Str.TabMap), fontSize = 11.sp) },
                    )
                }
                if (permissions.showTaskCenter) {
                    NavigationBarItem(
                        selected = tab == MainTab.Tasks,
                        onClick = {
                            tab = MainTab.Tasks
                        },
                        colors = itemColors,
                        icon = { Text("☑", fontSize = 16.sp) },
                        label = { Text(app.i18n.t(Str.TabTasks), fontSize = 11.sp) },
                    )
                }
                if (permissions.showScan) {
                    NavigationBarItem(
                        selected = false,
                        onClick = { scanOpen = true },
                        colors = itemColors,
                        icon = {
                            Box(
                                modifier = Modifier
                                    .size(42.dp)
                                    .background(OpsTheme.colors.primary, CircleShape),
                                contentAlignment = Alignment.Center,
                            ) {
                                Text("⬚", color = Color.White, fontSize = 18.sp)
                            }
                        },
                        label = { Text("") },
                    )
                }
                NavigationBarItem(
                    selected = tab == MainTab.Analysis,
                    onClick = { tab = MainTab.Analysis },
                    colors = itemColors,
                    icon = { Text("📈", fontSize = 14.sp) },
                    label = { Text(app.i18n.t(Str.TabAnalysis), fontSize = 11.sp) },
                )
                NavigationBarItem(
                    selected = tab == MainTab.Workbench,
                    onClick = { tab = MainTab.Workbench },
                    colors = itemColors,
                    icon = { Text("▦", fontSize = 16.sp) },
                    label = { Text(app.i18n.t(Str.TabWorkbench), fontSize = 11.sp) },
                )
            }
        },
    ) { padding ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding),
        ) {
            when (tab) {
                MainTab.Map -> MapTab(
                    app = app,
                    homeState = homeState,
                    permissions = permissions,
                    onChangeArea = onChangeArea,
                )
                MainTab.Tasks -> TasksTab(
                    app = app,
                    homeState = homeState,
                    permissions = permissions,
                    onChangeArea = onChangeArea,
                    onOpenChangeBatteryMap = { changeBatteryMapOpen = true },
                    onOpenMoveCarMap = { moveCarMapOpen = true },
                    onOpenInspectionTask = { inspectionTaskOpen = true },
                    onOpenRepairTask = { repairTaskOpen = true },
                )
                MainTab.Analysis -> AnalysisTab(
                    app = app,
                    homeState = homeState,
                    permissions = permissions,
                    onChangeArea = onChangeArea,
                    onOpenOfflineOps = {
                        app.offlineOpsFeature.open()
                        offlineOpsOpen = true
                    },
                    onOpenStationAnalysis = {
                        app.stationAnalysisFeature.open()
                        stationAnalysisOpen = true
                    },
                    onOpenReturnCarAnalysis = {
                        app.returnCarAnalysisFeature.open()
                        returnCarAnalysisOpen = true
                    },
                    onOpenVehicleDist = {
                        app.vehicleConditionDistributionFeature.open()
                        vehicleConditionDistOpen = true
                    },
                )
                MainTab.Workbench -> WorkbenchTab(
                    app = app,
                    homeState = homeState,
                    permissions = permissions,
                    trackPermissionHint = trackPermissionHint,
                    onOpenWarehouse = { warehouseOpen = true },
                    onOpenProduction = { productionOpen = true },
                    onOpenFaultReport = { faultReportOpen = true },
                    onOpenSneakReport = { sneakReportOpen = true },
                    onOpenUnlockedVehicles = { unlockedVehiclesOpen = true },
                    onOpenVehicleList = { vehicleListOpen = true },
                    onOpenFieldChangeBattery = { fieldChangeBatteryOpen = true },
                    onOpenFieldMoveCar = { fieldMoveCarOpen = true },
                    onOpenOrderQuery = { orderQueryOpen = true },
                    onOpenRelocation = { relocationOpen = true },
                    onOpenInspectionOrder = { inspectionOrderOpen = true },
                    onOpenRepairOrder = { repairOrderOpen = true },
                    onOpenH5 = { h5Screen = it },
                    onOpenMyTask = {
                        app.offlineOpsFeature.open()
                        myTaskOpen = true
                    },
                    onOpenFenceBrowse = { fenceBrowseOpen = true },
                    onOpenOpsSetting = { opsSettingOpen = true },
                    onOpenVehicleTag = { vehicleTagOpen = true },
                    onOpenBluetoothRadar = { bluetoothRadarOpen = true },
                    onOpenStaffDirectory = { staffDirectoryOpen = true },
                    onOpenProfessionAudit = { professionAuditOpen = true },
                    onOpenObjectionOrder = { objectionOrderOpen = true },
                    onOpenBlacklist = { blacklistOpen = true },
                    onOpenIdBindAudit = { idBindAuditOpen = true },
                    onOpenOperationLog = { operationLogOpen = true },
                    onChangeArea = onChangeArea,
                )
            }
        }
    }
}
