package com.luopingtech.ebike.ops.ui.staff

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Switch
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
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.platform.LocalUriHandler
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import androidx.compose.ui.window.DialogProperties
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsTimeRanges
import com.luopingtech.ebike.ops.domain.model.MapPin
import com.luopingtech.ebike.ops.domain.staff.CarStatusCatalog
import com.luopingtech.ebike.ops.domain.staff.HomeTabCatalog
import com.luopingtech.ebike.ops.domain.staff.StaffEmployee
import com.luopingtech.ebike.ops.domain.staff.StaffRole
import com.luopingtech.ebike.ops.domain.staff.ViewConfigEditKind
import com.luopingtech.ebike.ops.feature.staff.StaffManageUiState
import com.luopingtech.ebike.ops.feature.staff.StaffNav
import com.luopingtech.ebike.ops.feature.staff.StaffTrackPreset
import com.luopingtech.ebike.ops.ui.analysis.VcdTopBar
import com.luopingtech.ebike.ops.ui.map.OpsMapSpec
import com.luopingtech.ebike.ops.ui.map.OpsMapView
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.launch
import com.luopingtech.ebike.ops.domain.model.TrackPoint as MapTrackPoint
import com.luopingtech.ebike.ops.domain.staff.TrackPoint as StaffTrackPoint

private val PageBg = Color(0xFFF6F7F9)
private val TextPrimary = Color(0xFF242936)
private val LabelGray = Color(0xFF999999)
private val LabelMuted = Color(0xFF7C87B1)
private val LinkBlue = Color(0xFF3366CD)
private val RoleActionBlue = Color(0xFF1180F9)
private val RoleGreen = Color(0xFF00A763)
private val RoleGreenBg = Color(0xFFDBF7F0)
private val DividerColor = Color(0xFFDAE0F5)
private val FilterBarBg = Color(0x1A295FCC)
private val ActionBlue = Color(0xFF295FCC)
private val Danger = Color(0xFFE02020)
private val SearchHintWhite = Color(0x99FFFFFF)
private val SearchFieldBg = Color(0x33FFFFFF)
private val TextDark = Color(0xFF333333)

@Composable
fun StaffManageScreen(app: OpsApp, onClose: () -> Unit) {
    @Suppress("UNUSED_VARIABLE")
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.staffManageFeature.state.collectAsState()
    val feature = app.staffManageFeature
    val scope = rememberCoroutineScope()
    val primary = OpsTheme.colors.primary
    val isEn = language.name.contains("EN", ignoreCase = true)

    fun handleBack() {
        if (!feature.navigateBack()) {
            feature.clear()
            onClose()
        }
    }

    LaunchedEffect(state.nav) {
        when (state.nav) {
            StaffNav.EmployeeList -> feature.loadEmployeeList()
            StaffNav.RoleList -> feature.loadRoles()
            StaffNav.ViewConfig -> feature.loadViewConfigRoles()
            StaffNav.Track -> feature.loadTrackEmployees(state.trackSelected)
            else -> Unit
        }
    }

    when (state.nav) {
        StaffNav.Entry -> StaffEntryPage(
            title = t(Str.StaffManage),
            labels = listOf(
                t(Str.StaffEntryEmployeeAccount) to StaffNav.EmployeeList,
                t(Str.StaffEntryRoleTable) to StaffNav.RoleList,
                t(Str.StaffEntryViewConfig) to StaffNav.ViewConfig,
                t(Str.StaffEntryTrack) to StaffNav.Track,
            ),
            primary = primary,
            onBack = {
                feature.clear()
                onClose()
            },
            onOpen = { nav -> feature.openEntry(nav) },
        )
        StaffNav.EmployeeList -> StaffEmployeeListPage(
            t = { key -> t(key) },
            state = state,
            primary = primary,
            onBack = ::handleBack,
            onKeyword = { value ->
                feature.setEmployeeKeyword(value)
                if (value.isBlank()) scope.launch { feature.searchEmployees() }
            },
            onSearch = { scope.launch { feature.searchEmployees() } },
            onRoleFilter = { id ->
                feature.setRoleFilterId(id)
                scope.launch { feature.searchEmployees() }
            },
            onEnabledFilter = { enabled ->
                feature.setEnabledFilter(enabled)
                scope.launch { feature.searchEmployees() }
            },
            onLoadMore = { scope.launch { feature.loadMoreEmployees() } },
            onAdd = { scope.launch { feature.prepareAddEmployee() } },
            onToggleStatus = { emp -> scope.launch { feature.toggleEmployeeStatus(emp) } },
            onDelete = { emp -> scope.launch { feature.deleteEmployee(emp) } },
            onViewTrack = { emp -> feature.openTrackFor(emp) },
            onClearMessage = feature::clearMessage,
        )
        StaffNav.EmployeeAdd -> StaffEmployeeAddPage(
            t = { key -> t(key) },
            state = state,
            primary = primary,
            showIzRoot = feature.isIzRoot(),
            onBack = ::handleBack,
            onName = feature::setAddName,
            onPhone = feature::setAddPhone,
            onRole = feature::setAddRoleId,
            onProfession = feature::setAddProfession,
            onIzRoot = feature::setAddIzRoot,
            onSave = { scope.launch { feature.saveEmployee() } },
        )
        StaffNav.RoleList -> StaffRoleListPage(
            t = { key -> t(key) },
            state = state,
            primary = primary,
            onBack = ::handleBack,
            onTab = { status ->
                feature.setRoleStatusTab(status)
                scope.launch { feature.loadRoles() }
            },
            onKeyword = { value ->
                feature.setRoleKeyword(value)
                if (value.isBlank()) scope.launch { feature.loadRoles() }
            },
            onSearch = { scope.launch { feature.loadRoles() } },
            onToggle = { role -> scope.launch { feature.toggleRoleStatus(role) } },
            onDelete = { role -> scope.launch { feature.deleteRole(role) } },
            onSetPerm = { role -> scope.launch { feature.openHomePermission(role) } },
            onClearMessage = feature::clearMessage,
        )
        StaffNav.HomePermission -> StaffHomePermissionPage(
            t = { key -> t(key) },
            state = state,
            primary = primary,
            isEn = isEn,
            onBack = ::handleBack,
            onToggle = feature::toggleTabCode,
            onDone = { scope.launch { feature.saveHomePermission() } },
        )
        StaffNav.ViewConfig -> StaffViewConfigPage(
            t = { key -> t(key) },
            state = state,
            primary = primary,
            onBack = ::handleBack,
            onToggleExpand = feature::toggleViewRoleExpanded,
            onEdit = { role, kind -> scope.launch { feature.openViewConfigEdit(role, kind) } },
        )
        StaffNav.ViewConfigEdit -> StaffViewConfigEditPage(
            t = { key -> t(key) },
            state = state,
            primary = primary,
            isEn = isEn,
            onBack = ::handleBack,
            onToggleCarStatus = feature::toggleCarStatus,
            onTogglePlan = feature::toggleVoltagePlan,
            onSelectAllPlans = feature::selectAllVoltagePlans,
            onStart = feature::setCarIdStart,
            onEnd = feature::setCarIdEnd,
            onSubmit = { scope.launch { feature.saveViewConfigEdit() } },
        )
        StaffNav.Track -> StaffTrackPage(
            t = { key -> t(key) },
            state = state,
            primary = primary,
            onBack = ::handleBack,
            onSelect = { emp ->
                feature.setTrackSelected(emp)
                scope.launch { feature.queryTrack() }
            },
            onPreset = { preset ->
                feature.setTrackPreset(preset)
                scope.launch { feature.queryTrack() }
            },
            onStart = feature::setTrackStartTime,
            onEnd = feature::setTrackEndTime,
            onQuery = { scope.launch { feature.queryTrack() } },
        )
    }
}

@Composable
private fun StaffEntryPage(
    title: String,
    labels: List<Pair<String, StaffNav>>,
    primary: Color,
    onBack: () -> Unit,
    onOpen: (StaffNav) -> Unit,
) {
    Column(
        Modifier
            .fillMaxSize()
            .background(PageBg),
    ) {
        VcdTopBar(title = title, onBack = onBack)
        Column(
            Modifier
                .fillMaxSize()
                .padding(horizontal = 28.dp)
                .padding(top = 24.dp),
            verticalArrangement = Arrangement.spacedBy(31.dp),
        ) {
            labels.forEach { (label, nav) ->
                Box(
                    Modifier
                        .fillMaxWidth()
                        .height(48.dp)
                        .background(primary, RoundedCornerShape(4.dp))
                        .clickable { onOpen(nav) },
                    contentAlignment = Alignment.Center,
                ) {
                    Text(text = label, color = Color.White, fontSize = 16.sp, fontWeight = FontWeight.Medium)
                }
            }
        }
    }
}

@Composable
private fun StaffEmployeeListPage(
    t: (Str) -> String,
    state: StaffManageUiState,
    primary: Color,
    onBack: () -> Unit,
    onKeyword: (String) -> Unit,
    onSearch: () -> Unit,
    onRoleFilter: (String?) -> Unit,
    onEnabledFilter: (Boolean) -> Unit,
    onLoadMore: () -> Unit,
    onAdd: () -> Unit,
    onToggleStatus: (StaffEmployee) -> Unit,
    onDelete: (StaffEmployee) -> Unit,
    onViewTrack: (StaffEmployee) -> Unit,
    onClearMessage: () -> Unit,
) {
    val uriHandler = LocalUriHandler.current
    var actionEmp by remember { mutableStateOf<StaffEmployee?>(null) }
    var confirmDelete by remember { mutableStateOf<StaffEmployee?>(null) }
    var showRolePicker by remember { mutableStateOf(false) }
    var showStatusPicker by remember { mutableStateOf(false) }
    val listState = rememberLazyListState()

    LaunchedEffect(listState) {
        snapshotFlow {
            val info = listState.layoutInfo
            val last = info.visibleItemsInfo.lastOrNull()?.index ?: 0
            last >= info.totalItemsCount - 2
        }.distinctUntilChanged().collect { nearEnd ->
            if (nearEnd) onLoadMore()
        }
    }

    Box(Modifier.fillMaxSize().background(PageBg)) {
        Column(Modifier.fillMaxSize()) {
            Box(Modifier.fillMaxWidth()) {
                VcdTopBar(title = t(Str.StaffEmployeeAccount), onBack = onBack)
                TextButton(
                    onClick = onAdd,
                    modifier = Modifier.align(Alignment.CenterEnd).padding(top = 8.dp, end = 4.dp),
                ) {
                    Text(t(Str.StaffAdd), color = Color.White)
                }
            }
            Box(
                Modifier
                    .fillMaxWidth()
                    .height(68.dp)
                    .background(primary)
                    .padding(horizontal = 16.dp),
                contentAlignment = Alignment.Center,
            ) {
                ThemeSearchField(
                    value = state.employeeKeyword,
                    hint = t(Str.StaffSearchUserHint),
                    onValue = onKeyword,
                    onSearch = onSearch,
                    onClear = {
                        onKeyword("")
                        onSearch()
                    },
                )
            }
            Box(Modifier.fillMaxWidth()) {
                Row(
                    Modifier
                        .fillMaxWidth()
                        .height(52.dp)
                        .background(FilterBarBg),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    FilterBarItem(
                        label = state.roleOptions.find { it.id == state.roleFilterId }?.name
                            ?: t(Str.StaffAllRoles),
                        expanded = showRolePicker,
                        primary = primary,
                        onClick = {
                            showStatusPicker = false
                            showRolePicker = !showRolePicker
                        },
                        modifier = Modifier.weight(1f),
                    )
                    FilterBarItem(
                        label = if (state.enabledFilter) t(Str.StaffEnabled) else t(Str.StaffDisabled),
                        expanded = showStatusPicker,
                        primary = primary,
                        onClick = {
                            showRolePicker = false
                            showStatusPicker = !showStatusPicker
                        },
                        modifier = Modifier.weight(1f),
                    )
                }
                if (showRolePicker) {
                    FilterDropdownPanel(
                        options = listOf(null to t(Str.StaffAllRoles)) +
                            state.roleOptions.map { it.id to it.name },
                        onSelect = {
                            showRolePicker = false
                            onRoleFilter(it)
                        },
                        modifier = Modifier
                            .align(Alignment.TopCenter)
                            .padding(top = 52.dp)
                            .fillMaxWidth(),
                    )
                }
                if (showStatusPicker) {
                    FilterDropdownPanel(
                        options = listOf("1" to t(Str.StaffEnabled), "0" to t(Str.StaffDisabled)),
                        onSelect = {
                            showStatusPicker = false
                            onEnabledFilter(it == "1")
                        },
                        modifier = Modifier
                            .align(Alignment.TopCenter)
                            .padding(top = 52.dp)
                            .fillMaxWidth(),
                    )
                }
            }
            if (state.loading && state.employees.isEmpty()) {
                Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    CircularProgressIndicator(color = primary)
                }
            } else {
                LazyColumn(state = listState, modifier = Modifier.fillMaxSize()) {
                    if (!state.errorMessage.isNullOrBlank()) {
                        item {
                            Text(
                                state.errorMessage.orEmpty(),
                                color = Danger,
                                fontSize = 13.sp,
                                modifier = Modifier.padding(12.dp),
                            )
                        }
                    }
                    if (!state.message.isNullOrBlank()) {
                        item {
                            Text(
                                state.message.orEmpty(),
                                color = RoleGreen,
                                fontSize = 13.sp,
                                modifier = Modifier.padding(12.dp),
                            )
                            LaunchedEffect(state.message) { onClearMessage() }
                        }
                    }
                    if (state.employees.isEmpty() && !state.loading) {
                        item {
                            Text(
                                t(Str.AdminEmptyList),
                                color = LabelMuted,
                                modifier = Modifier.padding(16.dp),
                            )
                        }
                    }
                    items(state.employees, key = { it.id }) { emp ->
                        EmployeeRow(
                            emp = emp,
                            actionLabel = t(Str.StaffAction),
                            phoneLabel = t(Str.StaffPhoneLabel),
                            createdLabel = t(Str.StaffCreatedAtLabel),
                            trackLabel = t(Str.StaffViewTrack),
                            onAction = { actionEmp = emp },
                            onDial = {
                                val phone = emp.dialPhone
                                if (phone.isNotBlank()) uriHandler.openUri("tel:$phone")
                            },
                            onTrack = { onViewTrack(emp) },
                        )
                    }
                    if (state.loadingMore) {
                        item {
                            Box(
                                modifier = Modifier.fillMaxWidth().padding(12.dp),
                                contentAlignment = Alignment.Center,
                            ) {
                                CircularProgressIndicator(
                                    modifier = Modifier.size(28.dp),
                                    color = primary,
                                    strokeWidth = 3.dp,
                                )
                            }
                        }
                    }
                    item { Spacer(Modifier.height(16.dp)) }
                }
            }
        }
        if (showRolePicker || showStatusPicker) {
            Box(
                Modifier
                    .fillMaxSize()
                    .padding(top = 68.dp + 52.dp + 48.dp)
                    .clickable {
                        showRolePicker = false
                        showStatusPicker = false
                    },
            )
        }
    }

    actionEmp?.let { emp ->
        BottomActionSheet(
            primaryLabel = if (emp.enabled) t(Str.StaffDisabled) else t(Str.StaffEnabled),
            dangerLabel = t(Str.StaffDelete),
            cancelLabel = t(Str.Cancel),
            onPrimary = {
                actionEmp = null
                onToggleStatus(emp)
            },
            onDanger = {
                actionEmp = null
                if (emp.enabled) onDelete(emp) else confirmDelete = emp
            },
            onCancel = { actionEmp = null },
        )
    }
    confirmDelete?.let { emp ->
        ConfirmDialog(
            title = t(Str.StaffDeleteEmployeeConfirm),
            confirm = t(Str.Confirm),
            cancel = t(Str.Cancel),
            onConfirm = {
                confirmDelete = null
                onDelete(emp)
            },
            onCancel = { confirmDelete = null },
        )
    }
}

@Composable
private fun EmployeeRow(
    emp: StaffEmployee,
    actionLabel: String,
    phoneLabel: String,
    createdLabel: String,
    trackLabel: String,
    onAction: () -> Unit,
    onDial: () -> Unit,
    onTrack: () -> Unit,
) {
    Column(Modifier.fillMaxWidth().background(Color.White)) {
        Column(Modifier.fillMaxWidth().padding(horizontal = 16.dp).padding(top = 17.dp, bottom = 16.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(emp.name, fontWeight = FontWeight.Bold, color = TextPrimary, fontSize = 16.sp)
                if (emp.roleName.isNotBlank()) {
                    Spacer(Modifier.width(12.dp))
                    Text(
                        text = emp.roleName,
                        color = RoleGreen,
                        fontSize = 14.sp,
                        modifier = Modifier
                            .background(RoleGreenBg, RoundedCornerShape(12.dp))
                            .padding(horizontal = 10.dp, vertical = 2.dp),
                    )
                }
                Spacer(Modifier.weight(1f))
                Text(
                    text = actionLabel,
                    color = LinkBlue,
                    fontSize = 14.sp,
                    modifier = Modifier.clickable(onClick = onAction).padding(16.dp),
                )
            }
            Spacer(Modifier.height(11.dp))
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(phoneLabel, color = LabelGray, fontSize = 14.sp)
                Text(
                    text = emp.phone,
                    color = LinkBlue,
                    fontSize = 14.sp,
                    fontWeight = FontWeight.Bold,
                    modifier = Modifier.clickable(onClick = onDial),
                )
            }
            Spacer(Modifier.height(8.dp))
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(createdLabel, color = LabelGray, fontSize = 14.sp)
                Text(emp.createdAt, color = TextPrimary, fontSize = 14.sp, fontWeight = FontWeight.Bold)
                Spacer(Modifier.weight(1f))
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    modifier = Modifier.clickable(onClick = onTrack),
                ) {
                    Text(trackLabel, color = LabelGray, fontSize = 14.sp)
                    Text(" >", color = LabelGray, fontSize = 14.sp)
                }
            }
        }
        HorizontalDivider(thickness = 0.5.dp, color = DividerColor)
    }
}

@Composable
private fun StaffEmployeeAddPage(
    t: (Str) -> String,
    state: StaffManageUiState,
    primary: Color,
    showIzRoot: Boolean,
    onBack: () -> Unit,
    onName: (String) -> Unit,
    onPhone: (String) -> Unit,
    onRole: (String) -> Unit,
    onProfession: (String) -> Unit,
    onIzRoot: (Boolean) -> Unit,
    onSave: () -> Unit,
) {
    var showRolePicker by remember { mutableStateOf(false) }
    val canSave = state.canSaveAdd
    Column(Modifier.fillMaxSize().background(PageBg)) {
        VcdTopBar(title = t(Str.StaffAddAccount), onBack = onBack)
        Column(
            Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState())
                .padding(horizontal = 16.dp, vertical = 24.dp),
            verticalArrangement = Arrangement.spacedBy(20.dp),
        ) {
            AddFormRow(t(Str.StaffNameLabel)) {
                BorderedInput(state.addName, t(Str.StaffNameLabel), onName)
            }
            AddFormRow(t(Str.StaffPhoneFieldLabel)) {
                BorderedInput(
                    value = state.addPhone,
                    hint = t(Str.StaffPhoneFieldLabel),
                    onValue = onPhone,
                    keyboardType = KeyboardType.Phone,
                )
            }
            AddFormRow(t(Str.StaffRoleLabel)) {
                val roleName = state.roleOptions.find { it.id == state.addRoleId }?.name
                    ?: t(Str.StaffPleaseSelect)
                Box(
                    Modifier
                        .fillMaxWidth()
                        .height(40.dp)
                        .border(1.dp, Color(0xFFD8D8D8), RoundedCornerShape(4.dp))
                        .clickable { showRolePicker = true }
                        .padding(horizontal = 8.dp),
                    contentAlignment = Alignment.CenterStart,
                ) {
                    Text(
                        text = roleName,
                        color = if (state.addRoleId.isBlank()) LabelGray else TextPrimary,
                        fontSize = 14.sp,
                    )
                }
            }
            AddFormRow(t(Str.StaffProfessionLabel)) {
                BorderedInput(state.addProfession, t(Str.StaffProfessionLabel), onProfession)
            }
            if (showIzRoot) {
                Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
                    Text(t(Str.StaffIzRootLabel), color = LabelGray, fontSize = 14.sp)
                    Spacer(Modifier.weight(1f))
                    Switch(checked = state.addIzRoot, onCheckedChange = onIzRoot)
                }
            }
            if (!state.errorMessage.isNullOrBlank()) {
                Text(state.errorMessage.orEmpty(), color = Danger)
            }
            Spacer(Modifier.height(12.dp))
            Button(
                onClick = onSave,
                enabled = canSave && !state.loading,
                colors = ButtonDefaults.buttonColors(
                    containerColor = primary,
                    disabledContainerColor = Color(0xFFCCCCCC),
                ),
                modifier = Modifier.fillMaxWidth().height(48.dp),
                shape = RoundedCornerShape(4.dp),
            ) {
                if (state.loading) {
                    CircularProgressIndicator(
                        modifier = Modifier.size(22.dp),
                        color = Color.White,
                        strokeWidth = 2.dp,
                    )
                } else {
                    Text(t(Str.StaffSave), color = Color.White, fontSize = 16.sp)
                }
            }
        }
    }
    if (showRolePicker) {
        OptionPickerDialog(
            title = t(Str.StaffRoleLabel),
            options = state.roleOptions.map { it.id to it.name },
            onSelect = {
                showRolePicker = false
                if (it != null) onRole(it)
            },
            onDismiss = { showRolePicker = false },
        )
    }
}

@Composable
private fun StaffRoleListPage(
    t: (Str) -> String,
    state: StaffManageUiState,
    primary: Color,
    onBack: () -> Unit,
    onTab: (Int) -> Unit,
    onKeyword: (String) -> Unit,
    onSearch: () -> Unit,
    onToggle: (StaffRole) -> Unit,
    onDelete: (StaffRole) -> Unit,
    onSetPerm: (StaffRole) -> Unit,
    onClearMessage: () -> Unit,
) {
    var actionRole by remember { mutableStateOf<StaffRole?>(null) }
    var confirmDelete by remember { mutableStateOf<StaffRole?>(null) }
    var confirmDisable by remember { mutableStateOf<StaffRole?>(null) }
    var serviceRole by remember { mutableStateOf<StaffRole?>(null) }

    Column(Modifier.fillMaxSize().background(PageBg)) {
        VcdTopBar(title = t(Str.StaffRoleTable), onBack = onBack)
        Row(
            Modifier
                .fillMaxWidth()
                .background(primary.copy(alpha = 0.1f))
                .padding(start = 16.dp, end = 16.dp, top = 8.dp, bottom = 16.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                RoleTab(
                    label = t(Str.StaffEnabled),
                    selected = state.roleStatusTab == 1,
                    primary = primary,
                    onClick = { onTab(1) },
                )
                Spacer(Modifier.width(16.dp))
                RoleTab(
                    label = t(Str.StaffDisabled),
                    selected = state.roleStatusTab == 0,
                    primary = primary,
                    onClick = { onTab(0) },
                )
            }
            Spacer(Modifier.width(16.dp))
            RoleSearchField(
                value = state.roleKeyword,
                hint = t(Str.StaffSearchRoleHint),
                onValue = onKeyword,
                onSearch = onSearch,
                onClear = {
                    onKeyword("")
                    onSearch()
                },
                modifier = Modifier.weight(1f),
            )
        }
        if (state.loading && state.roles.isEmpty()) {
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                CircularProgressIndicator(color = primary)
            }
        } else {
            LazyColumn(Modifier.fillMaxSize()) {
                if (!state.errorMessage.isNullOrBlank()) {
                    item {
                        Text(
                            state.errorMessage.orEmpty(),
                            color = Danger,
                            modifier = Modifier.padding(12.dp),
                        )
                    }
                }
                if (!state.message.isNullOrBlank()) {
                    item {
                        Text(
                            state.message.orEmpty(),
                            color = RoleGreen,
                            modifier = Modifier.padding(12.dp),
                        )
                        LaunchedEffect(state.message) { onClearMessage() }
                    }
                }
                if (state.roles.isEmpty()) {
                    item {
                        Text(
                            t(Str.AdminEmptyList),
                            color = LabelMuted,
                            modifier = Modifier.padding(16.dp),
                        )
                    }
                }
                items(state.roles, key = { it.id }) { role ->
                    RoleRow(
                        role = role,
                        roleNameLabel = t(Str.StaffRoleNameLabel),
                        homePermLabel = t(Str.StaffHomePermLabel),
                        unsetLabel = t(Str.StaffHomePermUnset),
                        goSetLabel = t(Str.StaffGoSet),
                        serviceLabel = t(Str.StaffServiceAreaLabel),
                        remarkLabel = t(Str.StaffRemarkLabel),
                        actionLabel = t(Str.StaffAction),
                        onAction = { actionRole = role },
                        onSetPerm = { onSetPerm(role) },
                        onService = { serviceRole = role },
                    )
                }
            }
        }
    }

    actionRole?.let { role ->
        BottomActionSheet(
            primaryLabel = if (state.roleStatusTab == 1) t(Str.StaffDisabled) else t(Str.StaffEnabled),
            dangerLabel = t(Str.StaffDelete),
            cancelLabel = t(Str.Cancel),
            onPrimary = {
                actionRole = null
                if (state.roleStatusTab == 1) confirmDisable = role else onToggle(role)
            },
            onDanger = {
                actionRole = null
                confirmDelete = role
            },
            onCancel = { actionRole = null },
        )
    }
    confirmDisable?.let { role ->
        ConfirmDialog(
            title = t(Str.StaffConfirmDisable),
            confirm = t(Str.Confirm),
            cancel = t(Str.Cancel),
            onConfirm = {
                confirmDisable = null
                onToggle(role)
            },
            onCancel = { confirmDisable = null },
        )
    }
    confirmDelete?.let { role ->
        ConfirmDialog(
            title = t(Str.StaffConfirmDelete),
            confirm = t(Str.Confirm),
            cancel = t(Str.Cancel),
            onConfirm = {
                confirmDelete = null
                onDelete(role)
            },
            onCancel = { confirmDelete = null },
        )
    }
    serviceRole?.let { role ->
        ServiceAreaDialog(
            title = t(Str.StaffServiceAreaTitle),
            content = role.serviceName.ifBlank { "-" },
            gotIt = t(Str.StaffGotIt),
            onDismiss = { serviceRole = null },
        )
    }
}

@Composable
private fun RoleRow(
    role: StaffRole,
    roleNameLabel: String,
    homePermLabel: String,
    unsetLabel: String,
    goSetLabel: String,
    serviceLabel: String,
    remarkLabel: String,
    actionLabel: String,
    onAction: () -> Unit,
    onSetPerm: () -> Unit,
    onService: () -> Unit,
) {
    val summary = role.permissionSummary.ifBlank { unsetLabel }
    Column(Modifier.fillMaxWidth().background(Color.White)) {
        Column(Modifier.fillMaxWidth().padding(16.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(roleNameLabel, color = LabelMuted, fontSize = 14.sp)
                Spacer(Modifier.width(8.dp))
                Text(
                    role.name,
                    color = TextPrimary,
                    fontSize = 14.sp,
                    fontWeight = FontWeight.Bold,
                    modifier = Modifier.weight(1f),
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
                if (!role.isAdmin) {
                    Text(
                        text = actionLabel,
                        color = RoleActionBlue,
                        fontSize = 14.sp,
                        fontWeight = FontWeight.Bold,
                        modifier = Modifier.clickable(onClick = onAction),
                    )
                }
            }
            Spacer(Modifier.height(8.dp))
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(homePermLabel, color = LabelMuted, fontSize = 14.sp)
                Spacer(Modifier.width(8.dp))
                Text(summary, color = TextPrimary, fontSize = 14.sp, fontWeight = FontWeight.Bold)
                Spacer(Modifier.width(8.dp))
                Text(
                    text = goSetLabel,
                    color = RoleActionBlue,
                    fontSize = 14.sp,
                    fontWeight = FontWeight.Bold,
                    modifier = Modifier.clickable(onClick = onSetPerm),
                )
            }
            Spacer(Modifier.height(8.dp))
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(serviceLabel, color = LabelMuted, fontSize = 14.sp)
                Spacer(Modifier.width(8.dp))
                Text(
                    text = role.serviceName.ifBlank { "-" },
                    color = RoleActionBlue,
                    fontSize = 14.sp,
                    fontWeight = FontWeight.Bold,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                    modifier = Modifier.clickable(onClick = onService),
                )
            }
            Spacer(Modifier.height(8.dp))
            Row(verticalAlignment = Alignment.Top) {
                Text(remarkLabel, color = LabelMuted, fontSize = 14.sp)
                Spacer(Modifier.width(8.dp))
                Text(
                    text = role.remark.ifBlank { "-" },
                    color = TextPrimary,
                    fontSize = 14.sp,
                )
            }
        }
        HorizontalDivider(thickness = 0.5.dp, color = DividerColor)
    }
}

@OptIn(ExperimentalLayoutApi::class)
@Composable
private fun StaffHomePermissionPage(
    t: (Str) -> String,
    state: StaffManageUiState,
    primary: Color,
    isEn: Boolean,
    onBack: () -> Unit,
    onToggle: (String) -> Unit,
    onDone: () -> Unit,
) {
    Column(Modifier.fillMaxSize().background(PageBg)) {
        VcdTopBar(title = t(Str.StaffHomePermission), onBack = onBack)
        Column(
            Modifier
                .weight(1f)
                .verticalScroll(rememberScrollState()),
        ) {
            if (state.loading) {
                Box(Modifier.fillMaxWidth().padding(24.dp), contentAlignment = Alignment.Center) {
                    CircularProgressIndicator(color = primary)
                }
            }
            HomeTabCatalog.groups.forEach { group ->
                val groupTitle = when (group.groupCode) {
                    HomeTabCatalog.GROUP_BASIC -> t(Str.StaffBasicPermGroup)
                    HomeTabCatalog.GROUP_MAINTENANCE -> t(Str.StaffMaintModule)
                    HomeTabCatalog.GROUP_OPERATION -> t(Str.StaffOpModule)
                    else -> t(Str.StaffProdModule)
                }
                Text(
                    groupTitle,
                    fontWeight = FontWeight.Bold,
                    color = TextPrimary,
                    fontSize = 15.sp,
                    modifier = Modifier.padding(start = 16.dp, top = 32.dp, bottom = 12.dp),
                )
                FlowRow(
                    modifier = Modifier.padding(horizontal = 12.dp),
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    group.options.forEach { opt ->
                        val selected = opt.code in state.selectedTabCodes
                        SelectCapsule(
                            label = if (isEn) opt.labelEn else opt.labelZh,
                            selected = selected,
                            primary = primary,
                            enabled = !opt.forced,
                            onClick = { onToggle(opt.code) },
                        )
                    }
                }
            }
            if (!state.errorMessage.isNullOrBlank()) {
                Text(
                    state.errorMessage.orEmpty(),
                    color = Danger,
                    modifier = Modifier.padding(16.dp),
                )
            }
            Spacer(Modifier.height(24.dp))
        }
        Button(
            onClick = onDone,
            enabled = !state.loading,
            colors = ButtonDefaults.buttonColors(containerColor = primary),
            modifier = Modifier.fillMaxWidth().padding(16.dp).height(48.dp),
            shape = RoundedCornerShape(4.dp),
        ) {
            Text(t(Str.StaffDone), color = Color.White, fontSize = 16.sp)
        }
    }
}

@Composable
private fun StaffViewConfigPage(
    t: (Str) -> String,
    state: StaffManageUiState,
    primary: Color,
    onBack: () -> Unit,
    onToggleExpand: (String) -> Unit,
    onEdit: (StaffRole, ViewConfigEditKind) -> Unit,
) {
    Column(Modifier.fillMaxSize().background(PageBg)) {
        VcdTopBar(title = t(Str.StaffPermissionConfig), onBack = onBack)
        if (state.loading && state.viewConfigRoles.isEmpty()) {
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                CircularProgressIndicator(color = primary)
            }
        } else {
            LazyColumn(Modifier.fillMaxSize().background(Color.White)) {
                if (!state.errorMessage.isNullOrBlank()) {
                    item {
                        Text(
                            state.errorMessage.orEmpty(),
                            color = Danger,
                            modifier = Modifier.padding(12.dp),
                        )
                    }
                }
                items(state.viewConfigRoles, key = { it.id }) { role ->
                    val expanded = role.id in state.expandedViewRoleIds
                    Column(Modifier.fillMaxWidth()) {
                        Row(
                            Modifier
                                .fillMaxWidth()
                                .height(55.dp)
                                .clickable { onToggleExpand(role.id) }
                                .padding(horizontal = 16.dp),
                            verticalAlignment = Alignment.CenterVertically,
                        ) {
                            Text(role.name, fontWeight = FontWeight.Medium, color = TextPrimary, fontSize = 15.sp)
                            Spacer(Modifier.weight(1f))
                            Text(
                                text = if (expanded) "?" else "?",
                                color = primary,
                                fontSize = 12.sp,
                            )
                        }
                        if (expanded) {
                            Row(
                                Modifier.padding(start = 16.dp, end = 16.dp, bottom = 16.dp),
                                horizontalArrangement = Arrangement.spacedBy(12.dp),
                            ) {
                                OutlineCapsuleBtn(
                                    label = t(Str.StaffMoveCarViewBtn),
                                    primary = primary,
                                    onClick = { onEdit(role, ViewConfigEditKind.MoveCar) },
                                )
                                OutlineCapsuleBtn(
                                    label = t(Str.StaffBatteryViewBtn),
                                    primary = primary,
                                    onClick = { onEdit(role, ViewConfigEditKind.Battery) },
                                )
                            }
                        } else {
                            HorizontalDivider(thickness = 0.5.dp, color = DividerColor)
                        }
                    }
                }
            }
        }
    }
}

@OptIn(ExperimentalLayoutApi::class)
@Composable
private fun StaffViewConfigEditPage(
    t: (Str) -> String,
    state: StaffManageUiState,
    primary: Color,
    isEn: Boolean,
    onBack: () -> Unit,
    onToggleCarStatus: (Long) -> Unit,
    onTogglePlan: (Long) -> Unit,
    onSelectAllPlans: () -> Unit,
    onStart: (String) -> Unit,
    onEnd: (String) -> Unit,
    onSubmit: () -> Unit,
) {
    val title = if (state.viewEditKind == ViewConfigEditKind.MoveCar) {
        t(Str.StaffMoveCarView)
    } else {
        t(Str.StaffBatteryView)
    }
    val allSelected = state.voltagePlans.isNotEmpty() &&
        state.selectedVoltagePlanIds.containsAll(state.voltagePlans.map { it.id })
    Column(Modifier.fillMaxSize().background(PageBg)) {
        VcdTopBar(title = title, onBack = onBack)
        Column(
            Modifier
                .weight(1f)
                .verticalScroll(rememberScrollState()),
        ) {
            if (state.loading) {
                Box(Modifier.fillMaxWidth().padding(24.dp), contentAlignment = Alignment.Center) {
                    CircularProgressIndicator(color = primary)
                }
            }
            Text(
                text = if (state.viewEditKind == ViewConfigEditKind.MoveCar) {
                    t(Str.StaffCarStatusSection)
                } else {
                    t(Str.StaffBatteryPlanSection)
                },
                fontWeight = FontWeight.Bold,
                color = Color.Black,
                fontSize = 16.sp,
                modifier = Modifier.padding(16.dp),
            )
            FlowRow(
                modifier = Modifier.padding(horizontal = 8.dp),
                horizontalArrangement = Arrangement.spacedBy(8.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                if (state.viewEditKind == ViewConfigEditKind.MoveCar) {
                    CarStatusCatalog.options.forEach { opt ->
                        val selected = opt.filterCode in state.selectedCarStatus
                        val label = if (isEn) opt.labelEn else opt.labelZh
                        SelectCapsule(
                            label = "$label\n(${opt.filterCode})",
                            selected = selected,
                            primary = primary,
                            onClick = { onToggleCarStatus(opt.filterCode) },
                        )
                    }
                } else {
                    SelectCapsule(
                        label = t(Str.StaffAllPlans),
                        selected = allSelected,
                        primary = primary,
                        onClick = onSelectAllPlans,
                    )
                    state.voltagePlans.forEach { plan ->
                        val selected = plan.id in state.selectedVoltagePlanIds && !allSelected
                        SelectCapsule(
                            label = "${plan.name}\n(${plan.id})",
                            selected = selected,
                            primary = primary,
                            onClick = { onTogglePlan(plan.id) },
                        )
                    }
                }
            }
            Text(
                t(Str.StaffCarIdFilter),
                fontWeight = FontWeight.Bold,
                color = Color.Black,
                fontSize = 16.sp,
                modifier = Modifier.padding(16.dp),
            )
            Row(
                Modifier.padding(start = 16.dp, end = 16.dp, bottom = 16.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                NumberBox(
                    value = state.carIdStart,
                    hint = t(Str.StaffCarIdStart),
                    onValue = onStart,
                )
                Box(
                    Modifier
                        .padding(horizontal = 8.dp)
                        .width(11.dp)
                        .height(2.dp)
                        .background(Color(0xFFD8D8D8)),
                )
                NumberBox(
                    value = state.carIdEnd,
                    hint = t(Str.StaffCarIdEnd),
                    onValue = onEnd,
                )
            }
            if (!state.errorMessage.isNullOrBlank()) {
                Text(
                    state.errorMessage.orEmpty(),
                    color = Danger,
                    modifier = Modifier.padding(16.dp),
                )
            }
        }
        Button(
            onClick = onSubmit,
            enabled = !state.loading,
            colors = ButtonDefaults.buttonColors(containerColor = primary),
            modifier = Modifier.fillMaxWidth().padding(16.dp).height(48.dp),
            shape = RoundedCornerShape(4.dp),
        ) {
            Text(t(Str.StaffSubmit), color = Color.White, fontSize = 16.sp)
        }
    }
}

@OptIn(ExperimentalLayoutApi::class)
@Composable
private fun StaffTrackPage(
    t: (Str) -> String,
    state: StaffManageUiState,
    primary: Color,
    onBack: () -> Unit,
    onSelect: (StaffEmployee?) -> Unit,
    onPreset: (StaffTrackPreset) -> Unit,
    onStart: (String) -> Unit,
    onEnd: (String) -> Unit,
    onQuery: () -> Unit,
) {
    var expandEmployees by remember { mutableStateOf(false) }
    var zoomInNonce by remember { mutableStateOf(0) }
    var zoomOutNonce by remember { mutableStateOf(0) }
    var fitNonce by remember { mutableStateOf(0) }

    val mapTrackPoints = remember(state.trackPoints) {
        state.trackPoints.map { it.toMapTrackPoint() }
    }
    val pins = remember(state.trackPoints, state.trackSelected) {
        state.trackPoints.mapIndexed { idx, p ->
            MapPin(
                id = "tp-$idx",
                lat = p.lat,
                lng = p.lng,
                title = state.trackSelected?.name.orEmpty(),
            )
        }.takeLast(1)
    }

    Column(Modifier.fillMaxSize().background(PageBg)) {
        VcdTopBar(title = t(Str.StaffEmployeeTrack), onBack = onBack)
        Box(Modifier.weight(1f).fillMaxWidth()) {
            OpsMapView(
                spec = OpsMapSpec(
                    pins = pins,
                    trackPoints = mapTrackPoints,
                    clusterOverview = false,
                    autoFitOnPins = true,
                    zoomInNonce = zoomInNonce,
                    zoomOutNonce = zoomOutNonce,
                    fitNonce = fitNonce,
                ),
                modifier = Modifier.fillMaxSize(),
            )
            Column(
                Modifier
                    .align(Alignment.BottomStart)
                    .padding(12.dp),
            ) {
                MapToolBtn("+") { zoomInNonce++ }
                Spacer(Modifier.height(8.dp))
                MapToolBtn("-") { zoomOutNonce++ }
            }
            Box(
                Modifier
                    .align(Alignment.BottomEnd)
                    .padding(12.dp)
                    .size(40.dp)
                    .background(Color.White, RoundedCornerShape(4.dp))
                    .clickable {
                        fitNonce++
                        onQuery()
                    },
                contentAlignment = Alignment.Center,
            ) {
                Text("?", color = TextPrimary, fontSize = 18.sp)
            }
        }
        Column(
            Modifier
                .fillMaxWidth()
                .heightIn(min = 250.dp)
                .background(Color.White)
                .padding(bottom = 20.dp),
        ) {
            if (!state.trackHidePicker) {
                Row(
                    Modifier
                        .fillMaxWidth()
                        .clickable { expandEmployees = !expandEmployees }
                        .padding(top = 16.dp),
                    horizontalArrangement = Arrangement.Center,
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text(
                        text = state.trackSelected?.name ?: t(Str.StaffAllEmployees),
                        color = Color.Black,
                        fontSize = 16.sp,
                        fontWeight = FontWeight.Bold,
                    )
                    Text(
                        text = if (expandEmployees) " ?" else " ?",
                        color = LabelGray,
                        fontSize = 12.sp,
                    )
                }
            }
            if (expandEmployees && !state.trackHidePicker) {
                LazyColumn(Modifier.fillMaxWidth().heightIn(max = 280.dp)) {
                    item {
                        TrackEmployeeRow(
                            name = t(Str.StaffAllEmployees),
                            selected = state.trackSelected == null,
                            primary = primary,
                            onClick = {
                                expandEmployees = false
                                onSelect(null)
                            },
                        )
                    }
                    items(state.trackEmployees, key = { it.id }) { emp ->
                        TrackEmployeeRow(
                            name = emp.name,
                            selected = state.trackSelected?.id == emp.id,
                            primary = primary,
                            onClick = {
                                expandEmployees = false
                                onSelect(emp)
                            },
                        )
                    }
                }
            } else if (state.trackSelected == null && !state.trackHidePicker) {
                Box(
                    Modifier
                        .fillMaxWidth()
                        .height(180.dp),
                    contentAlignment = Alignment.Center,
                ) {
                    Text(
                        t(Str.StaffTrackSelectHint),
                        color = LabelMuted,
                        fontSize = 14.sp,
                        textAlign = TextAlign.Center,
                        modifier = Modifier.padding(horizontal = 24.dp),
                    )
                }
            } else {
                val presets = listOf(
                    StaffTrackPreset.Realtime to t(Str.StaffTrackRealtime),
                    StaffTrackPreset.M30 to t(Str.StaffTrackM30),
                    StaffTrackPreset.H2 to t(Str.StaffTrackH2),
                    StaffTrackPreset.H4 to t(Str.StaffTrackH4),
                    StaffTrackPreset.H8 to t(Str.StaffTrackH8),
                    StaffTrackPreset.H12 to t(Str.StaffTrackH12),
                )
                FlowRow(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(start = 8.dp, top = 16.dp, end = 8.dp),
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    presets.forEach { (preset, label) ->
                        SelectCapsule(
                            label = label,
                            selected = state.trackPreset == preset,
                            primary = primary,
                            onClick = { onPreset(preset) },
                        )
                    }
                }
                Row(
                    Modifier
                        .fillMaxWidth()
                        .padding(start = 13.dp, end = 12.dp, top = 16.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text(
                        t(Str.StaffCustomTime),
                        color = Color.Black,
                        fontSize = 14.sp,
                        fontWeight = FontWeight.Bold,
                    )
                    Spacer(Modifier.width(10.dp))
                    BorderedInput(
                        value = state.trackStartTime,
                        hint = t(Str.StaffStartTime),
                        onValue = onStart,
                        modifier = Modifier.weight(1f).height(39.dp),
                        textAlign = TextAlign.Center,
                    )
                    Box(
                        Modifier
                            .padding(horizontal = 4.dp)
                            .width(11.dp)
                            .height(2.dp)
                            .background(Color(0xFFD8D8D8)),
                    )
                    BorderedInput(
                        value = state.trackEndTime,
                        hint = t(Str.StaffEndTime),
                        onValue = onEnd,
                        modifier = Modifier.weight(1f).height(39.dp),
                        textAlign = TextAlign.Center,
                    )
                }
                if (state.trackPreset == StaffTrackPreset.Custom) {
                    TextButton(
                        onClick = onQuery,
                        modifier = Modifier.align(Alignment.End).padding(end = 12.dp),
                    ) {
                        Text(t(Str.StaffQuery), color = primary)
                    }
                }
                if (state.trackOperations.isNotEmpty()) {
                    Text(
                        t(Str.StaffTrackOperations),
                        fontWeight = FontWeight.Bold,
                        color = TextPrimary,
                        modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp),
                    )
                    state.trackOperations.take(5).forEach { op ->
                        Column(Modifier.padding(horizontal = 16.dp, vertical = 4.dp)) {
                            Text(
                                op.operationType.ifBlank { "-" },
                                fontWeight = FontWeight.Medium,
                                color = TextPrimary,
                                fontSize = 13.sp,
                            )
                            Text(
                                "${op.createdAt} ? ${op.carId}",
                                color = LabelMuted,
                                fontSize = 12.sp,
                            )
                        }
                    }
                }
            }
            if (!state.errorMessage.isNullOrBlank()) {
                Text(
                    state.errorMessage.orEmpty(),
                    color = Danger,
                    fontSize = 13.sp,
                    modifier = Modifier.padding(horizontal = 16.dp, vertical = 4.dp),
                )
            }
            if (state.loading) {
                Box(Modifier.fillMaxWidth().padding(8.dp), contentAlignment = Alignment.Center) {
                    CircularProgressIndicator(modifier = Modifier.size(24.dp), color = primary, strokeWidth = 2.dp)
                }
            }
        }
    }
}

// region shared widgets

@Composable
private fun ThemeSearchField(
    value: String,
    hint: String,
    onValue: (String) -> Unit,
    onSearch: () -> Unit,
    onClear: () -> Unit,
) {
    Row(
        Modifier
            .fillMaxWidth()
            .height(36.dp)
            .background(SearchFieldBg, RoundedCornerShape(4.dp))
            .padding(horizontal = 12.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        BasicTextField(
            value = value,
            onValueChange = onValue,
            singleLine = true,
            textStyle = TextStyle(color = Color.White, fontSize = 14.sp),
            cursorBrush = SolidColor(Color.White),
            keyboardOptions = KeyboardOptions(imeAction = ImeAction.Search),
            keyboardActions = KeyboardActions(onSearch = { onSearch() }),
            modifier = Modifier.weight(1f),
            decorationBox = { inner ->
                if (value.isEmpty()) Text(hint, color = SearchHintWhite, fontSize = 14.sp)
                inner()
            },
        )
        Text(
            text = if (value.isEmpty()) "?" else "?",
            color = Color.White,
            fontSize = 16.sp,
            modifier = Modifier
                .clickable { if (value.isEmpty()) onSearch() else onClear() }
                .padding(start = 8.dp),
        )
    }
}

@Composable
private fun RoleSearchField(
    value: String,
    hint: String,
    onValue: (String) -> Unit,
    onSearch: () -> Unit,
    onClear: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Row(
        modifier
            .height(40.dp)
            .background(Color.White, RoundedCornerShape(4.dp))
            .border(1.dp, Color(0xFFE0E0E0), RoundedCornerShape(4.dp))
            .padding(horizontal = 10.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        BasicTextField(
            value = value,
            onValueChange = onValue,
            singleLine = true,
            textStyle = TextStyle(color = TextPrimary, fontSize = 14.sp),
            cursorBrush = SolidColor(TextPrimary),
            keyboardOptions = KeyboardOptions(imeAction = ImeAction.Search),
            keyboardActions = KeyboardActions(onSearch = { onSearch() }),
            modifier = Modifier.weight(1f),
            decorationBox = { inner ->
                if (value.isEmpty()) Text(hint, color = LabelGray, fontSize = 14.sp)
                inner()
            },
        )
        if (value.isNotEmpty()) {
            Text(
                "?",
                color = LabelGray,
                modifier = Modifier.clickable(onClick = onClear).padding(start = 4.dp),
            )
        }
    }
}

@Composable
private fun FilterBarItem(
    label: String,
    expanded: Boolean,
    primary: Color,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Row(
        modifier.clickable(onClick = onClick),
        horizontalArrangement = Arrangement.Center,
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            text = label,
            color = if (expanded) primary else TextDark,
            fontSize = 15.sp,
            fontWeight = FontWeight.Bold,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
        )
        Spacer(Modifier.width(6.dp))
        Text(
            text = if (expanded) "?" else "?",
            color = if (expanded) primary else TextDark,
            fontSize = 10.sp,
        )
    }
}

@Composable
private fun FilterDropdownPanel(
    options: List<Pair<String?, String>>,
    onSelect: (String?) -> Unit,
    modifier: Modifier = Modifier,
) {
    Column(
        modifier
            .background(Color.White)
            .border(0.5.dp, DividerColor),
    ) {
        options.forEach { (id, label) ->
            Text(
                text = label,
                color = TextPrimary,
                fontSize = 14.sp,
                modifier = Modifier
                    .fillMaxWidth()
                    .clickable { onSelect(id) }
                    .padding(horizontal = 16.dp, vertical = 14.dp),
            )
            HorizontalDivider(thickness = 0.5.dp, color = DividerColor)
        }
    }
}

@Composable
private fun RoleTab(
    label: String,
    selected: Boolean,
    primary: Color,
    onClick: () -> Unit,
) {
    Column(
        Modifier
            .clickable(onClick = onClick)
            .padding(vertical = 8.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text(
            text = label,
            color = if (selected) primary else TextDark,
            fontSize = 16.sp,
            fontWeight = if (selected) FontWeight.Bold else FontWeight.Normal,
        )
        Spacer(Modifier.height(4.dp))
        Box(
            Modifier
                .width(28.dp)
                .height(2.dp)
                .background(if (selected) primary else Color.Transparent),
        )
    }
}

@Composable
private fun AddFormRow(label: String, content: @Composable () -> Unit) {
    Row(
        Modifier.fillMaxWidth(),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            label,
            color = LabelGray,
            fontSize = 14.sp,
            modifier = Modifier.widthIn(min = 72.dp),
        )
        Spacer(Modifier.width(8.dp))
        Box(Modifier.weight(1f)) { content() }
    }
}

@Composable
private fun BorderedInput(
    value: String,
    hint: String,
    onValue: (String) -> Unit,
    modifier: Modifier = Modifier,
    keyboardType: KeyboardType = KeyboardType.Text,
    textAlign: TextAlign = TextAlign.Start,
) {
    Box(
        modifier
            .fillMaxWidth()
            .height(40.dp)
            .border(1.dp, Color(0xFFD8D8D8), RoundedCornerShape(4.dp))
            .padding(horizontal = 8.dp),
        contentAlignment = Alignment.CenterStart,
    ) {
        BasicTextField(
            value = value,
            onValueChange = onValue,
            singleLine = true,
            textStyle = TextStyle(color = TextPrimary, fontSize = 14.sp, textAlign = textAlign),
            cursorBrush = SolidColor(TextPrimary),
            keyboardOptions = KeyboardOptions(keyboardType = keyboardType),
            modifier = Modifier.fillMaxWidth(),
            decorationBox = { inner ->
                if (value.isEmpty()) {
                    Text(
                        hint,
                        color = LabelGray,
                        fontSize = 14.sp,
                        textAlign = textAlign,
                        modifier = Modifier.fillMaxWidth(),
                    )
                }
                inner()
            },
        )
    }
}

@Composable
private fun NumberBox(
    value: String,
    hint: String,
    onValue: (String) -> Unit,
) {
    Box(
        Modifier
            .width(120.dp)
            .height(40.dp)
            .border(1.dp, Color(0xFFD8D8D8), RoundedCornerShape(4.dp))
            .padding(horizontal = 6.dp),
        contentAlignment = Alignment.Center,
    ) {
        BasicTextField(
            value = value,
            onValueChange = onValue,
            singleLine = true,
            textStyle = TextStyle(color = Color.Black, fontSize = 14.sp, textAlign = TextAlign.Center),
            cursorBrush = SolidColor(TextPrimary),
            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
            modifier = Modifier.fillMaxWidth(),
            decorationBox = { inner ->
                if (value.isEmpty()) {
                    Text(hint, color = LabelMuted, fontSize = 14.sp, textAlign = TextAlign.Center, modifier = Modifier.fillMaxWidth())
                }
                inner()
            },
        )
    }
}

@Composable
private fun SelectCapsule(
    label: String,
    selected: Boolean,
    primary: Color,
    enabled: Boolean = true,
    onClick: () -> Unit,
) {
    Box(
        Modifier
            .width(103.dp)
            .height(40.dp)
            .then(
                if (selected) {
                    Modifier.background(primary, RoundedCornerShape(20.dp))
                } else {
                    Modifier.border(1.dp, Color(0xFFD0D5E0), RoundedCornerShape(20.dp))
                },
            )
            .clickable(enabled = enabled, onClick = onClick),
        contentAlignment = Alignment.Center,
    ) {
        Text(
            text = label,
            color = when {
                selected -> Color.White
                else -> TextPrimary
            },
            fontSize = 12.sp,
            textAlign = TextAlign.Center,
            lineHeight = 14.sp,
            maxLines = 2,
        )
    }
}

@Composable
private fun OutlineCapsuleBtn(
    label: String,
    primary: Color,
    onClick: () -> Unit,
) {
    Box(
        Modifier
            .border(1.dp, primary, RoundedCornerShape(20.dp))
            .clickable(onClick = onClick)
            .padding(horizontal = 22.dp, vertical = 9.dp),
        contentAlignment = Alignment.Center,
    ) {
        Text(label, color = primary, fontSize = 13.sp)
    }
}

@Composable
private fun TrackEmployeeRow(
    name: String,
    selected: Boolean,
    primary: Color,
    onClick: () -> Unit,
) {
    Row(
        Modifier
            .fillMaxWidth()
            .height(40.dp)
            .clickable(onClick = onClick)
            .padding(horizontal = 16.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            name,
            color = if (selected) primary else TextPrimary,
            fontSize = 14.sp,
            modifier = Modifier.weight(1f),
        )
        if (selected) {
            Text("?", color = primary, fontSize = 16.sp)
        }
    }
}

@Composable
private fun MapToolBtn(label: String, onClick: () -> Unit) {
    Box(
        Modifier
            .size(40.dp)
            .background(Color.White, RoundedCornerShape(4.dp))
            .clickable(onClick = onClick),
        contentAlignment = Alignment.Center,
    ) {
        Text(label, color = TextPrimary, fontSize = 20.sp)
    }
}

@Composable
private fun OptionPickerDialog(
    title: String,
    options: List<Pair<String?, String>>,
    onSelect: (String?) -> Unit,
    onDismiss: () -> Unit,
) {
    Dialog(onDismissRequest = onDismiss) {
        Column(
            Modifier
                .fillMaxWidth()
                .background(Color.White, RoundedCornerShape(12.dp))
                .padding(16.dp),
        ) {
            Text(title, fontWeight = FontWeight.Bold, color = TextPrimary)
            Spacer(Modifier.height(8.dp))
            options.forEach { (id, label) ->
                Text(
                    text = label,
                    color = TextPrimary,
                    modifier = Modifier.fillMaxWidth().clickable { onSelect(id) }.padding(vertical = 12.dp),
                )
                HorizontalDivider(color = DividerColor)
            }
        }
    }
}

@Composable
private fun ConfirmDialog(
    title: String,
    confirm: String,
    cancel: String,
    onConfirm: () -> Unit,
    onCancel: () -> Unit,
) {
    Dialog(onDismissRequest = onCancel) {
        Column(
            Modifier
                .fillMaxWidth()
                .background(Color.White, RoundedCornerShape(12.dp))
                .padding(20.dp),
        ) {
            Text(title, color = TextPrimary, fontSize = 16.sp)
            Spacer(Modifier.height(16.dp))
            Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.End) {
                TextButton(onClick = onCancel) { Text(cancel) }
                TextButton(onClick = onConfirm) { Text(confirm, color = Danger) }
            }
        }
    }
}

@Composable
private fun ServiceAreaDialog(
    title: String,
    content: String,
    gotIt: String,
    onDismiss: () -> Unit,
) {
    Dialog(onDismissRequest = onDismiss) {
        Column(
            Modifier
                .fillMaxWidth()
                .background(Color.White, RoundedCornerShape(12.dp))
                .padding(20.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Text(title, fontWeight = FontWeight.Bold, color = TextPrimary, fontSize = 16.sp)
            Spacer(Modifier.height(12.dp))
            Text(content, color = TextPrimary, fontSize = 14.sp, textAlign = TextAlign.Center)
            Spacer(Modifier.height(16.dp))
            TextButton(onClick = onDismiss) { Text(gotIt, color = RoleActionBlue) }
        }
    }
}

@Composable
private fun BottomActionSheet(
    primaryLabel: String,
    dangerLabel: String,
    cancelLabel: String,
    onPrimary: () -> Unit,
    onDanger: () -> Unit,
    onCancel: () -> Unit,
) {
    Dialog(
        onDismissRequest = onCancel,
        properties = DialogProperties(usePlatformDefaultWidth = false),
    ) {
        Box(
            Modifier
                .fillMaxSize()
                .background(Color.Black.copy(alpha = 0.35f))
                .clickable(onClick = onCancel),
            contentAlignment = Alignment.BottomCenter,
        ) {
            Column(
                Modifier
                    .fillMaxWidth()
                    .background(Color.White)
                    .clickable(enabled = false) {},
                horizontalAlignment = Alignment.CenterHorizontally,
            ) {
                Text(
                    text = primaryLabel,
                    color = ActionBlue,
                    fontSize = 16.sp,
                    modifier = Modifier
                        .fillMaxWidth()
                        .clickable(onClick = onPrimary)
                        .padding(vertical = 16.dp),
                    textAlign = TextAlign.Center,
                )
                HorizontalDivider(thickness = 0.5.dp, color = Color(0xFFE8E8E8))
                Text(
                    text = dangerLabel,
                    color = Danger,
                    fontSize = 16.sp,
                    modifier = Modifier
                        .fillMaxWidth()
                        .clickable(onClick = onDanger)
                        .padding(vertical = 16.dp),
                    textAlign = TextAlign.Center,
                )
                Box(Modifier.fillMaxWidth().height(8.dp).background(Color(0xFFF2F2F2)))
                Text(
                    text = cancelLabel,
                    color = Color.Black,
                    fontSize = 16.sp,
                    modifier = Modifier
                        .fillMaxWidth()
                        .clickable(onClick = onCancel)
                        .padding(vertical = 16.dp),
                    textAlign = TextAlign.Center,
                )
            }
        }
    }
}

private fun StaffTrackPoint.toMapTrackPoint(): MapTrackPoint {
    val ts = createdAt.takeIf { it.isNotBlank() }?.let { text ->
        val parts = text.trim().split(' ')
        if (parts.size >= 2) {
            val day = OfflineOpsTimeRanges.parseDateStart(parts[0])
            val hm = parts[1].split(':').mapNotNull { it.toIntOrNull() }
            if (day != null && hm.size >= 2) day + hm[0] * 3_600_000L + hm[1] * 60_000L else null
        } else {
            null
        }
    } ?: 0L
    return MapTrackPoint(lat = lat, lng = lng, timestamp = ts)
}

// endregion
