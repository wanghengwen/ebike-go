package com.luopingtech.ebike.ops.ui.shell

import com.luopingtech.ebike.ops.OpsApp
import androidx.compose.foundation.layout.size
import androidx.compose.material3.Text
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
import com.luopingtech.ebike.ops.ui.feedback.LocalOpsToast
import com.luopingtech.ebike.ops.ui.icons.OpsIcon
import com.luopingtech.ebike.ops.ui.workbench.WORKBENCH_COMMON_MAX
import com.luopingtech.ebike.ops.ui.workbench.WorkbenchEditBadge
import com.luopingtech.ebike.ops.ui.workbench.WorkbenchModuleItem
import com.luopingtech.ebike.ops.ui.workbench.WorkbenchModuleSection
import com.luopingtech.ebike.ops.ui.workbench.WorkbenchScaffold
import com.luopingtech.ebike.ops.platform.SecureStore
import com.luopingtech.ebike.ops.core.config.H5ScreenKind
import com.luopingtech.ebike.ops.core.config.H5ScreenUrls
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.BusinessTenant
import com.luopingtech.ebike.ops.domain.model.WarehouseOperationType
import com.luopingtech.ebike.ops.domain.permission.OpsPermissions
import com.luopingtech.ebike.ops.feature.home.HomeUiState
import com.luopingtech.ebike.ops.feature.production.ShelfMode
import com.luopingtech.ebike.ops.ui.auth.BusinessPickerScaffold
import com.luopingtech.ebike.ops.ui.auth.UpdatePasswordScaffold
import kotlinx.coroutines.launch

@Composable
internal fun WorkbenchTab(
    app: OpsApp,
    homeState: HomeUiState,
    permissions: OpsPermissions,
    onOpenWarehouse: () -> Unit,
    onOpenProduction: () -> Unit,
    onOpenFaultReport: () -> Unit,
    onOpenSneakReport: () -> Unit,
    onOpenUnlockedVehicles: () -> Unit,
    onOpenVehicleList: () -> Unit,
    onOpenFieldChangeBattery: () -> Unit,
    onOpenFieldMoveCar: () -> Unit,
    onOpenOrderQuery: () -> Unit,
    onOpenRelocation: () -> Unit,
    onOpenInspectionOrder: () -> Unit,
    onOpenRepairOrder: () -> Unit,
    onOpenH5: (H5ScreenKind) -> Unit,
    onOpenMyTask: () -> Unit,
    onOpenFenceBrowse: () -> Unit,
    onOpenOpsSetting: () -> Unit,
    onOpenVehicleTag: () -> Unit,
    onOpenBluetoothRadar: () -> Unit,
    onOpenStaffDirectory: () -> Unit,
    onOpenProfessionAudit: () -> Unit,
    onOpenObjectionOrder: () -> Unit,
    onOpenBlacklist: () -> Unit,
    onOpenIdBindAudit: () -> Unit,
    onOpenOperationLog: () -> Unit,
    onChangeArea: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val authState by app.authFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    var settingsOpen by remember { mutableStateOf(false) }
    var switchBusinessOpen by remember { mutableStateOf(false) }
    var switchTenants by remember { mutableStateOf<List<BusinessTenant>>(emptyList()) }
    var switchQuery by remember { mutableStateOf("") }
    var updatePasswordOpen by remember { mutableStateOf(false) }
    var oldPassword by remember { mutableStateOf("") }
    var newPassword by remember { mutableStateOf("") }
    var oldPasswordVisible by remember { mutableStateOf(false) }
    var newPasswordVisible by remember { mutableStateOf(false) }

    val toast = LocalOpsToast.current
    fun comingSoon() = toast(t(Str.FeatureComingSoon))

    val opsItems = buildList {
        if (!permissions.showMaintainModule && !app.isDemoMode) return@buildList
        fun addModule(
            visible: Boolean,
            id: String,
            title: String,
            icon: OpsIcon,
            onClick: () -> Unit,
        ) {
            if (visible) add(WorkbenchModuleItem(id, title, icon, onClick))
        }
        addModule(permissions.showVehicleList, "vlist", t(Str.VehicleList), OpsIcon.VehicleList, onOpenVehicleList)
        addModule(permissions.showChangeBatteryTool, "batt", t(Str.ChangeBatteryTool), OpsIcon.ReplaceBattery, onOpenFieldChangeBattery)
        addModule(permissions.showMoveCarTool, "move", t(Str.MoveCarTool), OpsIcon.MoveVehicle, onOpenFieldMoveCar)
        addModule(permissions.showFaultReport, "fault", t(Str.FaultReport), OpsIcon.Repair) {
            app.faultReportFeature.openSubmit()
            onOpenFaultReport()
        }
        addModule(permissions.showUnlockedVehicles, "unlocked", t(Str.UnlockedVehicles), OpsIcon.UnlockedVehicle, onOpenUnlockedVehicles)
        addModule(permissions.showBluetoothRadar, "ble", t(Str.BluetoothRadar), OpsIcon.BluetoothRadar, onOpenBluetoothRadar)
        addModule(permissions.showOpsSetting, "ops-set", t(Str.OpsSettingTool), OpsIcon.OperationSetting, onOpenOpsSetting)
        addModule(permissions.showMyTask, "my-task", t(Str.MyTaskTool), OpsIcon.MyTask, onOpenMyTask)
        addModule(permissions.showVehicleTag, "vtag", t(Str.VehicleTagTool), OpsIcon.VehicleTag, onOpenVehicleTag)
        addModule(permissions.showRelocation, "reloc", t(Str.Relocation), OpsIcon.Relocation, onOpenRelocation)
    }

    val operationItems = buildList {
        if (!permissions.showOperationModule && !app.isDemoMode) return@buildList
        fun addModule(
            visible: Boolean,
            id: String,
            title: String,
            icon: OpsIcon,
            onClick: () -> Unit,
        ) {
            if (visible) add(WorkbenchModuleItem(id, title, icon, onClick))
        }
        addModule(permissions.showParkingFence, "fence", t(Str.ParkingFence), OpsIcon.ParkingArea, onOpenFenceBrowse)
        addModule(
            permissions.showOrderQuery,
            "order-query",
            t(Str.OrderQueryScreen),
            OpsIcon.OrderQuery,
            onOpenOrderQuery,
        )
        addModule(permissions.showOperationScreen, "h5-op", t(Str.OperationScreen), OpsIcon.OperationScreen) {
            if (H5ScreenUrls.isConfigured(app.config, H5ScreenKind.Operation)) {
                onOpenH5(H5ScreenKind.Operation)
            } else {
                comingSoon()
            }
        }
        addModule(permissions.showRevenueScreen, "h5-rev", t(Str.RevenueScreen), OpsIcon.RevenueScreen) {
            if (H5ScreenUrls.isConfigured(app.config, H5ScreenKind.Revenue)) {
                onOpenH5(H5ScreenKind.Revenue)
            } else {
                comingSoon()
            }
        }
        addModule(permissions.showStaffManage, "staff", t(Str.StaffManage), OpsIcon.EmployeeManager, onOpenStaffDirectory)
        addModule(permissions.showProfessionAudit, "prof", t(Str.ProfessionAudit), OpsIcon.ProfessionAudit, onOpenProfessionAudit)
        addModule(permissions.showObjectionOrder, "obj", t(Str.ObjectionOrder), OpsIcon.ObjectionOrder, onOpenObjectionOrder)
        addModule(permissions.showBlacklist, "black", t(Str.Blacklist), OpsIcon.BlackList, onOpenBlacklist)
        addModule(permissions.showIdBindAudit, "idbind", t(Str.IdBindAudit), OpsIcon.IdAudit, onOpenIdBindAudit)
        addModule(permissions.showOperationLog, "oplog", t(Str.OperationLog), OpsIcon.OperationLog, onOpenOperationLog)
    }

    val productionItems = buildList {
        if (permissions.showProductionDetect) {
            add(
                WorkbenchModuleItem(
                    id = "prod-detect",
                    title = t(Str.ProductionDetect),
                    icon = OpsIcon.VehicleInspection,
                    onClick = {
                        app.productionFeature.openDetect()
                        onOpenProduction()
                    },
                ),
            )
        }
        if (permissions.showProductionBind) {
            add(
                WorkbenchModuleItem(
                    id = "prod-bind",
                    title = t(Str.ProductionBind),
                    icon = OpsIcon.CenterControlBind,
                    onClick = {
                        app.productionFeature.openBind()
                        onOpenProduction()
                    },
                ),
            )
        }
        if (permissions.showProductionShelves) {
            add(
                WorkbenchModuleItem(
                    id = "prod-shelves",
                    title = t(Str.ProductionShelves),
                    icon = OpsIcon.PutPullShelves,
                    onClick = {
                        // Legacy defaults to put-on tab inside the same activity.
                        app.productionFeature.openShelves(ShelfMode.PutOn)
                        onOpenProduction()
                    },
                ),
            )
        }
    }

    val warehouseItems = buildList {
        if (permissions.showWarehouseIn) {
            add(
                WorkbenchModuleItem(
                    id = "wh-in",
                    title = t(Str.WarehouseIn),
                    icon = OpsIcon.InWarehouse,
                    onClick = {
                        app.warehouseFeature.openKind(WarehouseOperationType.In)
                        onOpenWarehouse()
                    },
                ),
            )
        }
        if (permissions.showWarehouseOut) {
            add(
                WorkbenchModuleItem(
                    id = "wh-out",
                    title = t(Str.WarehouseOut),
                    icon = OpsIcon.OutWarehouse,
                    onClick = {
                        app.warehouseFeature.openKind(WarehouseOperationType.Out)
                        onOpenWarehouse()
                    },
                ),
            )
        }
        if (permissions.showWarehouseRecord) {
            add(
                WorkbenchModuleItem(
                    id = "wh-rec",
                    title = t(Str.WarehouseRecords),
                    icon = OpsIcon.WarehouseRecord,
                    onClick = {
                        scope.launch {
                            app.warehouseFeature.openRecords()
                            onOpenWarehouse()
                        }
                    },
                ),
            )
        }
    }

    val catalogItems = opsItems + operationItems + productionItems + warehouseItems
    val catalogById = catalogItems.associateBy { it.id }

    fun loadCommonIds(): List<String> =
        app.secureStore.getString(SecureStore.KEY_COMMON_MODULE_IDS)
            ?.split('|')
            ?.map { it.trim() }
            ?.filter { it.isNotBlank() }
            ?.distinct()
            .orEmpty()

    fun persistCommonIds(ids: List<String>) {
        app.secureStore.putString(
            SecureStore.KEY_COMMON_MODULE_IDS,
            ids.take(WORKBENCH_COMMON_MAX).joinToString("|"),
        )
    }

    var commonIds by remember {
        mutableStateOf(loadCommonIds().filter { catalogById.containsKey(it) }.take(WORKBENCH_COMMON_MAX))
    }
    var editingCommon by remember { mutableStateOf(false) }
    var draftCommonIds by remember { mutableStateOf(commonIds) }

    LaunchedEffect(catalogById.keys.joinToString()) {
        val filtered = commonIds.filter { catalogById.containsKey(it) }.take(WORKBENCH_COMMON_MAX)
        if (filtered != commonIds) {
            commonIds = filtered
            persistCommonIds(filtered)
        }
        if (!editingCommon) {
            draftCommonIds = filtered
        }
    }

    fun reorderIds(ids: List<String>, from: Int, to: Int): List<String> {
        if (from !in ids.indices || to !in ids.indices || from == to) return ids
        val mutable = ids.toMutableList()
        val item = mutable.removeAt(from)
        mutable.add(to, item)
        return mutable
    }

    val activeCommonIds = if (editingCommon) draftCommonIds else commonIds
    val commonItems = activeCommonIds.mapNotNull { id ->
        val base = catalogById[id] ?: return@mapNotNull null
        WorkbenchModuleItem(
            id = base.id,
            title = base.title,
            icon = base.icon,
            onClick = {
                if (!editingCommon) base.onClick()
            },
            editBadge = if (editingCommon) WorkbenchEditBadge.Remove else WorkbenchEditBadge.None,
            onBadgeClick = {
                draftCommonIds = draftCommonIds.filterNot { it == id }
            },
        )
    }

    fun decorateSection(items: List<WorkbenchModuleItem>): List<WorkbenchModuleItem> {
        if (!editingCommon) return items
        return items.map { item ->
            val alreadyPinned = draftCommonIds.contains(item.id)
            WorkbenchModuleItem(
                id = item.id,
                title = item.title,
                icon = item.icon,
                onClick = {},
                editBadge = if (alreadyPinned) WorkbenchEditBadge.None else WorkbenchEditBadge.Add,
                onBadgeClick = {
                    if (draftCommonIds.size >= WORKBENCH_COMMON_MAX) {
                        toast(t(Str.MaxCommonModules))
                    } else if (!draftCommonIds.contains(item.id)) {
                        draftCommonIds = draftCommonIds + item.id
                    }
                },
            )
        }
    }

    val sections = listOf(
        WorkbenchModuleSection(title = t(Str.OpsModule), items = decorateSection(opsItems)),
        WorkbenchModuleSection(title = t(Str.OperationModule), items = decorateSection(operationItems)),
        WorkbenchModuleSection(title = t(Str.ProductionModule), items = decorateSection(productionItems)),
        WorkbenchModuleSection(title = t(Str.WarehouseModule), items = decorateSection(warehouseItems)),
    )

    val companyName = authState.runtimeConfig?.tenantName
        ?.takeIf { it.isNotBlank() }
        ?: app.config.app.displayName.ifBlank { app.config.name }
    val userName = homeState.session?.displayName
        ?.takeIf { it.isNotBlank() }
        ?: homeState.session?.phone.orEmpty()
    val roleLabel = homeState.session?.roleName
        ?.takeIf { it.isNotBlank() }
        ?: t(Str.AdminRole)

    if (updatePasswordOpen) {
        UpdatePasswordScaffold(
            title = t(Str.ChangePassword),
            originalTitle = t(Str.OldPassword),
            originalHint = t(Str.OriginalPasswordHint),
            originalValue = oldPassword,
            onOriginalChange = { oldPassword = it },
            originalVisible = oldPasswordVisible,
            onToggleOriginalVisible = { oldPasswordVisible = !oldPasswordVisible },
            newTitle = t(Str.NewPassword),
            newHint = t(Str.UpdatePasswordNewHint),
            newValue = newPassword,
            onNewChange = { newPassword = it },
            newVisible = newPasswordVisible,
            onToggleNewVisible = { newPasswordVisible = !newPasswordVisible },
            confirmLabel = t(Str.Confirm),
            loading = authState.loading,
            errorMessage = authState.errorMessage,
            backLabel = t(Str.Back),
            onBack = { updatePasswordOpen = false },
            onConfirm = {
                scope.launch {
                    when (app.authFeature.updatePassword(oldPassword, newPassword)) {
                        is OpsResult.Ok -> {
                            toast(t(Str.PasswordChangeSuccess))
                            updatePasswordOpen = false
                            oldPassword = ""
                            newPassword = ""
                        }
                        is OpsResult.Err -> Unit
                    }
                }
            },
        )
        return
    }

    if (switchBusinessOpen) {
        BusinessPickerScaffold(
            title = t(Str.SelectBusiness),
            searchHint = t(Str.Search),
            query = switchQuery,
            onQueryChange = { switchQuery = it },
            businesses = switchTenants,
            loading = authState.loading,
            errorMessage = authState.errorMessage,
            backLabel = t(Str.Back),
            onBack = { switchBusinessOpen = false },
            onSelect = { tenantId ->
                scope.launch {
                    when (val r = app.authFeature.switchBusiness(tenantId)) {
                        is OpsResult.Ok -> {
                            switchBusinessOpen = false
                            settingsOpen = false
                        }
                        is OpsResult.Err -> toast(r.error.message)
                    }
                }
            },
        )
        return
    }

    WorkbenchScaffold(
        areaName = homeState.currentArea?.name?.takeIf { it.isNotBlank() }
            ?: t(Str.SelectServiceArea),
        companyName = companyName,
        userName = userName,
        roleLabel = roleLabel,
        commonTitle = t(Str.CommonModules),
        editLabel = t(Str.Edit),
        settingsTitle = t(Str.Settings),
        backLabel = t(Str.Back),
        saveLabel = t(Str.Save),
        editCommonTitle = t(Str.EditCommonModules),
        dragHint = t(Str.DragCommonHint),
        commonItems = commonItems,
        sections = sections,
        settingsOpen = settingsOpen,
        editingCommon = editingCommon,
        onOpenSettings = { settingsOpen = true },
        onCloseSettings = { settingsOpen = false },
        onChangeArea = onChangeArea,
        onStartEditCommon = {
            draftCommonIds = commonIds
            editingCommon = true
        },
        onCancelEditCommon = {
            draftCommonIds = commonIds
            editingCommon = false
        },
        onSaveEditCommon = {
            commonIds = draftCommonIds.take(WORKBENCH_COMMON_MAX)
            persistCommonIds(commonIds)
            editingCommon = false
            toast(t(Str.Save))
        },
        onReorderCommon = { from, to ->
            if (editingCommon) {
                draftCommonIds = reorderIds(draftCommonIds, from, to)
            } else {
                commonIds = reorderIds(commonIds, from, to)
                persistCommonIds(commonIds)
            }
        },
        settingsContent = {
            WorkbenchSettingsContent(
                app = app,
                authState = authState,
                onOpenSwitchBusiness = { tenants ->
                    switchTenants = tenants
                    switchQuery = ""
                    switchBusinessOpen = true
                },
                onOpenUpdatePassword = {
                    oldPassword = ""
                    newPassword = ""
                    oldPasswordVisible = false
                    newPasswordVisible = false
                    updatePasswordOpen = true
                },
            )
        },
    )
}
