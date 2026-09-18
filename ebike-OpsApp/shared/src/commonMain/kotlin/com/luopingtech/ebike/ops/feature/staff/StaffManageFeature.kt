package com.luopingtech.ebike.ops.feature.staff

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.staff.StaffManagementRepository
import com.luopingtech.ebike.ops.domain.staff.AddEmployeeRequest
import com.luopingtech.ebike.ops.domain.staff.BatteryConfig
import com.luopingtech.ebike.ops.domain.staff.CarStatusCatalog
import com.luopingtech.ebike.ops.domain.staff.HomeTabCatalog
import com.luopingtech.ebike.ops.domain.staff.MoveCarConfig
import com.luopingtech.ebike.ops.domain.staff.StaffEmployee
import com.luopingtech.ebike.ops.domain.staff.StaffEmployeePageQuery
import com.luopingtech.ebike.ops.domain.staff.StaffRole
import com.luopingtech.ebike.ops.domain.staff.StaffRoleOption
import com.luopingtech.ebike.ops.domain.staff.TrackOperation
import com.luopingtech.ebike.ops.domain.staff.TrackPoint
import com.luopingtech.ebike.ops.domain.staff.ViewConfigEditKind
import com.luopingtech.ebike.ops.domain.staff.VoltagePlanOption
import com.luopingtech.ebike.ops.core.time.nowEpochMillis
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsTimeRanges
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

enum class StaffNav {
    Entry,
    EmployeeList,
    EmployeeAdd,
    RoleList,
    HomePermission,
    ViewConfig,
    ViewConfigEdit,
    Track,
}

enum class StaffTrackPreset {
    Realtime,
    M30,
    H2,
    H4,
    H8,
    H12,
    Custom,
}

data class StaffManageUiState(
    val nav: StaffNav = StaffNav.Entry,
    val loading: Boolean = false,
    val loadingMore: Boolean = false,
    val finished: Boolean = false,
    val pageNum: Int = 1,
    val message: String? = null,
    val errorMessage: String? = null,

    // employee list
    val employees: List<StaffEmployee> = emptyList(),
    val employeeKeyword: String = "",
    val roleFilterId: String? = null,
    val enabledFilter: Boolean = true,
    val roleOptions: List<StaffRoleOption> = emptyList(),

    // add employee
    val addName: String = "",
    val addPhone: String = "",
    val addRoleId: String = "",
    val addProfession: String = "",
    val addIzRoot: Boolean = false,

    // roles
    val roles: List<StaffRole> = emptyList(),
    val roleStatusTab: Int = 1,
    val roleKeyword: String = "",

    // home permission
    val editingRole: StaffRole? = null,
    val selectedTabCodes: Set<String> = emptySet(),

    // view config
    val viewConfigRoles: List<StaffRole> = emptyList(),
    val expandedViewRoleIds: Set<String> = emptySet(),
    val viewEditKind: ViewConfigEditKind = ViewConfigEditKind.MoveCar,
    val viewEditRoleId: String = "",
    val viewEditRoleName: String = "",
    val selectedCarStatus: Set<Long> = emptySet(),
    val selectedVoltagePlanIds: Set<Long> = emptySet(),
    val voltagePlans: List<VoltagePlanOption> = emptyList(),
    val carIdStart: String = "",
    val carIdEnd: String = "",

    // track
    val trackEmployees: List<StaffEmployee> = emptyList(),
    val trackSelected: StaffEmployee? = null,
    /** 从账号「查看轨迹」进入时隐藏选人区 */
    val trackHidePicker: Boolean = false,
    val trackPreset: StaffTrackPreset = StaffTrackPreset.Realtime,
    val trackStartTime: String = "",
    val trackEndTime: String = "",
    val trackPoints: List<TrackPoint> = emptyList(),
    val trackOperations: List<TrackOperation> = emptyList(),
) {
    val canSaveAdd: Boolean
        get() = addName.isNotBlank() &&
            addPhone.length == 11 &&
            addRoleId.isNotBlank() &&
            addProfession.isNotBlank()
}

class StaffManageFeature(
    private val repository: StaffManagementRepository,
    private val tenantIdProvider: () -> String,
    private val currentUserIdProvider: () -> String,
    private val izRootProvider: () -> Boolean,
) {
    private val _state = MutableStateFlow(StaffManageUiState())
    val state: StateFlow<StaffManageUiState> = _state.asStateFlow()

    fun clear() {
        _state.value = StaffManageUiState()
    }

    fun clearMessage() {
        _state.value = _state.value.copy(message = null, errorMessage = null)
    }

    fun isIzRoot(): Boolean = izRootProvider()

    fun navigateBack(): Boolean {
        val s = _state.value
        return when (s.nav) {
            StaffNav.Entry -> false
            StaffNav.EmployeeAdd -> {
                _state.value = s.copy(nav = StaffNav.EmployeeList, errorMessage = null, message = null)
                true
            }
            StaffNav.HomePermission -> {
                _state.value = s.copy(
                    nav = StaffNav.RoleList,
                    editingRole = null,
                    selectedTabCodes = emptySet(),
                    errorMessage = null,
                    message = null,
                )
                true
            }
            StaffNav.ViewConfigEdit -> {
                _state.value = s.copy(nav = StaffNav.ViewConfig, errorMessage = null, message = null)
                true
            }
            StaffNav.EmployeeList,
            StaffNav.RoleList,
            StaffNav.ViewConfig,
            StaffNav.Track,
            -> {
                _state.value = StaffManageUiState(nav = StaffNav.Entry)
                true
            }
        }
    }

    fun openEntry(target: StaffNav) {
        _state.value = StaffManageUiState(nav = target)
    }

    // region employee list
    fun setEmployeeKeyword(value: String) {
        _state.value = _state.value.copy(employeeKeyword = value)
    }

    fun setRoleFilterId(id: String?) {
        _state.value = _state.value.copy(roleFilterId = id)
    }

    fun setEnabledFilter(enabled: Boolean) {
        _state.value = _state.value.copy(enabledFilter = enabled)
    }

    suspend fun loadEmployeeList() {
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null, pageNum = 1, finished = false)
        loadRoleOptions()
        when (val result = repository.pageEmployees(buildEmployeeQuery(1))) {
            is OpsResult.Ok -> _state.value = _state.value.copy(
                loading = false,
                employees = result.value,
                finished = result.value.size < PAGE_SIZE,
                pageNum = 1,
            )
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                errorMessage = result.error.message,
            )
        }
    }

    suspend fun searchEmployees() {
        loadEmployeeList()
    }

    suspend fun loadMoreEmployees() {
        val s = _state.value
        if (s.loading || s.loadingMore || s.finished || s.nav != StaffNav.EmployeeList) return
        val next = s.pageNum + 1
        _state.value = s.copy(loadingMore = true, errorMessage = null)
        when (val result = repository.pageEmployees(buildEmployeeQuery(next))) {
            is OpsResult.Ok -> _state.value = _state.value.copy(
                loadingMore = false,
                employees = s.employees + result.value,
                finished = result.value.size < PAGE_SIZE,
                pageNum = next,
            )
            is OpsResult.Err -> _state.value = _state.value.copy(
                loadingMore = false,
                errorMessage = result.error.message,
            )
        }
    }

    private suspend fun loadRoleOptions() {
        when (val result = repository.roleListForFilter(tenantIdProvider())) {
            is OpsResult.Ok -> _state.value = _state.value.copy(roleOptions = result.value)
            is OpsResult.Err -> Unit
        }
    }

    private fun buildEmployeeQuery(pageNum: Int): StaffEmployeePageQuery {
        val s = _state.value
        return StaffEmployeePageQuery(
            pageNum = pageNum,
            pageSize = PAGE_SIZE,
            status = if (s.enabledFilter) 1 else 0,
            nameOrPhone = s.employeeKeyword.trim().ifBlank { null },
            roleIds = s.roleFilterId?.let { listOf(it) },
        )
    }

    suspend fun toggleEmployeeStatus(employee: StaffEmployee) {
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        when (val result = repository.setEmployeeStatus(employee.id, !employee.enabled)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(loading = false, message = Strings.t(Str.StaffActionSuccess))
                loadEmployeeList()
            }
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                errorMessage = result.error.message,
            )
        }
    }

    suspend fun deleteEmployee(employee: StaffEmployee) {
        if (employee.enabled) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.StaffDisableBeforeDelete))
            return
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        when (val result = repository.deleteEmployee(employee.id)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(loading = false, message = Strings.t(Str.StaffActionSuccess))
                loadEmployeeList()
            }
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                errorMessage = result.error.message,
            )
        }
    }

    fun openAddEmployee() {
        _state.value = _state.value.copy(
            nav = StaffNav.EmployeeAdd,
            addName = "",
            addPhone = "",
            addRoleId = "",
            addProfession = "",
            addIzRoot = false,
            errorMessage = null,
            message = null,
        )
    }

    suspend fun prepareAddEmployee() {
        if (_state.value.roleOptions.isEmpty()) {
            loadRoleOptions()
        }
        openAddEmployee()
    }

    fun setAddName(v: String) {
        _state.value = _state.value.copy(addName = v.take(20))
    }

    fun setAddPhone(v: String) {
        _state.value = _state.value.copy(addPhone = v.filter { it.isDigit() }.take(11))
    }

    fun setAddRoleId(v: String) {
        _state.value = _state.value.copy(addRoleId = v)
    }

    fun setAddProfession(v: String) {
        _state.value = _state.value.copy(addProfession = v.take(20))
    }

    fun setAddIzRoot(v: Boolean) {
        _state.value = _state.value.copy(addIzRoot = v)
    }

    fun canSaveEmployee(): Boolean = _state.value.canSaveAdd

    suspend fun saveEmployee() {
        if (!_state.value.canSaveAdd) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.StaffFormIncomplete))
            return
        }
        val s = _state.value
        _state.value = s.copy(loading = true, errorMessage = null, message = null)
        val req = AddEmployeeRequest(
            name = s.addName.trim(),
            phone = "+86-${s.addPhone}",
            roleIds = listOf(s.addRoleId),
            profession = s.addProfession.trim(),
            izRoot = if (izRootProvider()) s.addIzRoot else false,
        )
        when (val result = repository.addEmployee(req)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    nav = StaffNav.EmployeeList,
                    message = Strings.t(Str.StaffActionSuccess),
                )
                loadEmployeeList()
            }
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                errorMessage = result.error.message,
            )
        }
    }
    // endregion

    // region roles
    fun setRoleStatusTab(status: Int) {
        _state.value = _state.value.copy(roleStatusTab = status)
    }

    fun setRoleKeyword(value: String) {
        _state.value = _state.value.copy(roleKeyword = value.take(20))
    }

    suspend fun loadRoles() {
        val s = _state.value
        _state.value = s.copy(loading = true, errorMessage = null, message = null)
        when (
            val result = repository.roleAppList(
                userId = currentUserIdProvider(),
                status = s.roleStatusTab,
                subTenantId = tenantIdProvider(),
                name = s.roleKeyword.trim().ifBlank { null },
            )
        ) {
            is OpsResult.Ok -> _state.value = _state.value.copy(loading = false, roles = result.value)
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                errorMessage = result.error.message,
            )
        }
    }

    suspend fun toggleRoleStatus(role: StaffRole) {
        if (role.isAdmin) return
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        when (
            val result = repository.roleDisableEnabled(
                id = role.id,
                name = role.name,
                status = _state.value.roleStatusTab,
            )
        ) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(loading = false, message = Strings.t(Str.StaffActionSuccess))
                loadRoles()
            }
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                errorMessage = result.error.message,
            )
        }
    }

    suspend fun deleteRole(role: StaffRole) {
        if (role.isAdmin) return
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        when (
            val result = repository.roleDelete(
                id = role.id,
                name = role.name,
                status = _state.value.roleStatusTab,
            )
        ) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(loading = false, message = Strings.t(Str.StaffActionSuccess))
                loadRoles()
            }
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                errorMessage = result.error.message,
            )
        }
    }

    suspend fun openHomePermission(role: StaffRole) {
        _state.value = _state.value.copy(
            loading = true,
            errorMessage = null,
            message = null,
            editingRole = role,
            nav = StaffNav.HomePermission,
        )
        when (val result = repository.getHomeTabsByRoleId(role.id)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(
                loading = false,
                selectedTabCodes = HomeTabCatalog.withForcedBasics(result.value).toSet(),
            )
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                selectedTabCodes = setOf(
                    HomeTabCatalog.CODE_CHANGE_BATTERY_MAP,
                    HomeTabCatalog.CODE_MOVE_CAR_MAP,
                ),
                errorMessage = result.error.message,
            )
        }
    }

    fun toggleTabCode(code: String) {
        if (code == HomeTabCatalog.CODE_CHANGE_BATTERY_MAP ||
            code == HomeTabCatalog.CODE_MOVE_CAR_MAP
        ) {
            return
        }
        val cur = _state.value.selectedTabCodes.toMutableSet()
        if (code in cur) cur.remove(code) else cur.add(code)
        _state.value = _state.value.copy(selectedTabCodes = cur)
    }

    suspend fun saveHomePermission() {
        val role = _state.value.editingRole ?: return
        val codes = HomeTabCatalog.withForcedBasics(_state.value.selectedTabCodes)
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        when (val result = repository.saveHomeTabs(role.id, codes)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    nav = StaffNav.RoleList,
                    editingRole = null,
                    message = Strings.t(Str.StaffActionSuccess),
                )
                loadRoles()
            }
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                errorMessage = result.error.message,
            )
        }
    }
    // endregion

    // region view config
    suspend fun loadViewConfigRoles() {
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        when (
            val result = repository.roleAppList(
                userId = currentUserIdProvider(),
                status = 1,
                subTenantId = tenantIdProvider(),
            )
        ) {
            is OpsResult.Ok -> _state.value = _state.value.copy(
                loading = false,
                viewConfigRoles = result.value,
            )
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                errorMessage = result.error.message,
            )
        }
    }

    fun toggleViewRoleExpanded(roleId: String) {
        val set = _state.value.expandedViewRoleIds.toMutableSet()
        if (roleId in set) set.remove(roleId) else set.add(roleId)
        _state.value = _state.value.copy(expandedViewRoleIds = set)
    }

    suspend fun openViewConfigEdit(role: StaffRole, kind: ViewConfigEditKind) {
        _state.value = _state.value.copy(
            loading = true,
            errorMessage = null,
            message = null,
            nav = StaffNav.ViewConfigEdit,
            viewEditKind = kind,
            viewEditRoleId = role.id,
            viewEditRoleName = role.name,
            selectedCarStatus = emptySet(),
            selectedVoltagePlanIds = emptySet(),
            carIdStart = "",
            carIdEnd = "",
        )
        if (kind == ViewConfigEditKind.Battery) {
            when (val plans = repository.voltagePlanList()) {
                is OpsResult.Ok -> _state.value = _state.value.copy(voltagePlans = plans.value)
                is OpsResult.Err -> Unit
            }
        }
        when (val result = repository.getViewConfig(role.id)) {
            is OpsResult.Ok -> {
                val cfg = result.value
                _state.value = if (kind == ViewConfigEditKind.MoveCar) {
                    _state.value.copy(
                        loading = false,
                        selectedCarStatus = cfg.moveCar.carStatus.toSet(),
                        carIdStart = cfg.moveCar.carIdStart.takeIf { it != "0" }.orEmpty(),
                        carIdEnd = cfg.moveCar.carIdEnd.takeIf { it != "0" }.orEmpty(),
                    )
                } else {
                    _state.value.copy(
                        loading = false,
                        selectedVoltagePlanIds = cfg.battery.voltagePlanIds.toSet(),
                        carIdStart = cfg.battery.carIdStart.takeIf { it != "0" }.orEmpty(),
                        carIdEnd = cfg.battery.carIdEnd.takeIf { it != "0" }.orEmpty(),
                    )
                }
            }
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                errorMessage = result.error.message,
            )
        }
    }

    fun toggleCarStatus(code: Long) {
        val allCode = CarStatusCatalog.options.first().filterCode
        val set = _state.value.selectedCarStatus.toMutableSet()
        if (code == allCode) {
            if (allCode in set) set.clear() else {
                set.clear()
                set.add(allCode)
            }
        } else {
            set.remove(allCode)
            if (code in set) set.remove(code) else set.add(code)
        }
        _state.value = _state.value.copy(selectedCarStatus = set)
    }

    fun toggleVoltagePlan(id: Long) {
        val allIds = _state.value.voltagePlans.map { it.id }.toSet()
        val set = _state.value.selectedVoltagePlanIds.toMutableSet()
        // 与「全部」互斥：点单项时若当前是全选，则先清空再只选当前
        if (allIds.isNotEmpty() && set.containsAll(allIds)) {
            set.clear()
            set.add(id)
        } else {
            if (id in set) set.remove(id) else set.add(id)
        }
        _state.value = _state.value.copy(selectedVoltagePlanIds = set)
    }

    fun selectAllVoltagePlans() {
        val allIds = _state.value.voltagePlans.map { it.id }.toSet()
        val cur = _state.value.selectedVoltagePlanIds
        _state.value = _state.value.copy(
            selectedVoltagePlanIds = if (allIds.isNotEmpty() && cur.containsAll(allIds)) {
                emptySet()
            } else {
                allIds
            },
        )
    }

    fun setCarIdStart(v: String) {
        _state.value = _state.value.copy(carIdStart = v.filter { it.isDigit() }.take(9))
    }

    fun setCarIdEnd(v: String) {
        _state.value = _state.value.copy(carIdEnd = v.filter { it.isDigit() }.take(9))
    }

    suspend fun saveViewConfigEdit() {
        val s = _state.value
        val start = s.carIdStart
        val end = s.carIdEnd
        if ((start.isNotBlank() && start.length != 9) || (end.isNotBlank() && end.length != 9)) {
            _state.value = s.copy(errorMessage = Strings.t(Str.StaffInvalidCarId))
            return
        }
        val startVal = start.ifBlank { "0" }
        val endVal = end.ifBlank { "0" }
        _state.value = s.copy(loading = true, errorMessage = null, message = null)
        val result = when (s.viewEditKind) {
            ViewConfigEditKind.MoveCar -> repository.saveMoveCarConfig(
                s.viewEditRoleId,
                MoveCarConfig(
                    carStatus = s.selectedCarStatus.toList().ifEmpty {
                        listOf(CarStatusCatalog.options.first().filterCode)
                    },
                    carIdStart = startVal,
                    carIdEnd = endVal,
                ),
            )
            ViewConfigEditKind.Battery -> repository.saveBatteryConfig(
                s.viewEditRoleId,
                BatteryConfig(
                    voltagePlanIds = s.selectedVoltagePlanIds.toList(),
                    carIdStart = startVal,
                    carIdEnd = endVal,
                ),
            )
        }
        when (result) {
            is OpsResult.Ok -> _state.value = _state.value.copy(
                loading = false,
                nav = StaffNav.ViewConfig,
                message = Strings.t(Str.StaffActionSuccess),
            )
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                errorMessage = result.error.message,
            )
        }
    }
    // endregion

    // region track
    fun setTrackSelected(employee: StaffEmployee?) {
        _state.value = _state.value.copy(
            trackSelected = employee,
            trackPoints = emptyList(),
            trackOperations = emptyList(),
        )
    }

    fun setTrackStartTime(v: String) {
        _state.value = _state.value.copy(trackStartTime = v, trackPreset = StaffTrackPreset.Custom)
    }

    fun setTrackEndTime(v: String) {
        _state.value = _state.value.copy(trackEndTime = v, trackPreset = StaffTrackPreset.Custom)
    }

    fun setTrackPreset(preset: StaffTrackPreset) {
        val now = nowEpochMillis()
        val end = OfflineOpsTimeRanges.formatDateTime(now).take(16)
        val durationMs = when (preset) {
            StaffTrackPreset.Realtime -> 30L * 60_000L
            StaffTrackPreset.M30 -> 30L * 60_000L
            StaffTrackPreset.H2 -> 2L * 3600_000L
            StaffTrackPreset.H4 -> 4L * 3600_000L
            StaffTrackPreset.H8 -> 8L * 3600_000L
            StaffTrackPreset.H12 -> 12L * 3600_000L - 60_000L
            StaffTrackPreset.Custom -> null
        }
        if (durationMs != null) {
            val start = OfflineOpsTimeRanges.formatDateTime(now - durationMs).take(16)
            _state.value = _state.value.copy(
                trackPreset = preset,
                trackStartTime = start,
                trackEndTime = end,
            )
        } else {
            _state.value = _state.value.copy(trackPreset = preset)
        }
    }

    fun ensureDefaultTrackTimes() {
        if (_state.value.trackPreset != StaffTrackPreset.Realtime &&
            _state.value.trackStartTime.isNotBlank()
        ) {
            return
        }
        setTrackPreset(_state.value.trackPreset)
    }

    suspend fun loadTrackEmployees(preselect: StaffEmployee? = null) {
        ensureDefaultTrackTimes()
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        when (
            val result = repository.pageEmployees(
                StaffEmployeePageQuery(pageNum = 1, pageSize = 1000, status = 1),
            )
        ) {
            is OpsResult.Ok -> {
                val selected = when {
                    preselect != null -> result.value.find { it.id == preselect.id } ?: preselect
                    _state.value.trackHidePicker -> _state.value.trackSelected
                    else -> {
                        val cur = _state.value.trackSelected
                        if (cur != null) result.value.find { it.id == cur.id } ?: cur else null
                    }
                }
                _state.value = _state.value.copy(
                    loading = false,
                    trackEmployees = result.value,
                    trackSelected = selected,
                )
                if (selected != null || _state.value.trackPreset == StaffTrackPreset.Realtime) {
                    queryTrack()
                }
            }
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                errorMessage = result.error.message,
            )
        }
    }

    fun openTrackFor(employee: StaffEmployee) {
        _state.value = StaffManageUiState(
            nav = StaffNav.Track,
            trackSelected = employee,
            trackEmployees = listOf(employee),
            trackHidePicker = true,
            trackPreset = StaffTrackPreset.Realtime,
        )
        setTrackPreset(StaffTrackPreset.Realtime)
    }

    suspend fun queryTrack() {
        val s = _state.value
        if (s.trackPreset == StaffTrackPreset.Realtime) {
            queryRealtimeTrack()
            return
        }
        val emp = s.trackSelected
        if (emp == null) {
            _state.value = s.copy(
                trackPoints = emptyList(),
                trackOperations = emptyList(),
                errorMessage = null,
            )
            return
        }
        if (s.trackStartTime.isBlank() || s.trackEndTime.isBlank()) {
            _state.value = s.copy(errorMessage = Strings.t(Str.StaffSelectTimeRange))
            return
        }
        if (!isTrackRangeValid(s.trackStartTime, s.trackEndTime)) {
            _state.value = s.copy(errorMessage = Strings.t(Str.StaffTrackRangeTooLong))
            return
        }
        val pin = emp.pin.ifBlank { emp.id }
        _state.value = s.copy(loading = true, errorMessage = null, message = null)
        val trackResult = repository.getTrack(pin, s.trackStartTime, s.trackEndTime)
        val opsResult = repository.getAllOperations(pin, s.trackStartTime, s.trackEndTime)
        when (trackResult) {
            is OpsResult.Ok -> {
                val ops = when (opsResult) {
                    is OpsResult.Ok -> opsResult.value
                    is OpsResult.Err -> trackResult.value.operations
                }
                _state.value = _state.value.copy(
                    loading = false,
                    trackPoints = trackResult.value.points,
                    trackOperations = ops.ifEmpty { trackResult.value.operations },
                )
            }
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                errorMessage = trackResult.error.message,
            )
        }
    }

    private suspend fun queryRealtimeTrack() {
        val s = _state.value
        val emp = s.trackSelected
        _state.value = s.copy(loading = true, errorMessage = null, message = null)
        val pin = emp?.pin?.ifBlank { emp.id }
        when (val realtime = repository.realTimeLocation(pin)) {
            is OpsResult.Ok -> {
                val points = if (emp != null) {
                    realtime.value.filter {
                        it.pin == emp.pin || it.id == emp.id || it.phone == emp.phone
                    }.ifEmpty { realtime.value }
                } else {
                    realtime.value
                }.flatMap { it.points }
                val ops = if (emp != null) {
                    val userPin = emp.pin.ifBlank { emp.id }
                    when (val opsResult = repository.getAllOperations(userPin, null, null)) {
                        is OpsResult.Ok -> opsResult.value
                        is OpsResult.Err -> emptyList()
                    }
                } else {
                    emptyList()
                }
                if (points.isNotEmpty()) {
                    _state.value = _state.value.copy(
                        loading = false,
                        trackPoints = points,
                        trackOperations = ops,
                    )
                } else {
                    // API 无数据时兜底：近 30 分钟 getTrack
                    fallbackRealtimeTrack(emp)
                }
            }
            is OpsResult.Err -> fallbackRealtimeTrack(emp)
        }
    }

    private suspend fun fallbackRealtimeTrack(emp: StaffEmployee?) {
        if (emp == null) {
            _state.value = _state.value.copy(
                loading = false,
                trackPoints = emptyList(),
                trackOperations = emptyList(),
            )
            return
        }
        val now = nowEpochMillis()
        val end = OfflineOpsTimeRanges.formatDateTime(now).take(16)
        val start = OfflineOpsTimeRanges.formatDateTime(now - 30L * 60_000L).take(16)
        val pin = emp.pin.ifBlank { emp.id }
        when (val trackResult = repository.getTrack(pin, start, end)) {
            is OpsResult.Ok -> {
                val ops = when (val opsResult = repository.getAllOperations(pin, start, end)) {
                    is OpsResult.Ok -> opsResult.value
                    is OpsResult.Err -> trackResult.value.operations
                }
                _state.value = _state.value.copy(
                    loading = false,
                    trackStartTime = start,
                    trackEndTime = end,
                    trackPoints = trackResult.value.points,
                    trackOperations = ops.ifEmpty { trackResult.value.operations },
                )
            }
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                errorMessage = trackResult.error.message,
            )
        }
    }
    // endregion

    companion object {
        const val PAGE_SIZE = 10

        /** Interval must be < 12h; start <= end. */
        fun isTrackRangeValid(start: String, end: String): Boolean {
            val s = parseMinute(start) ?: return false
            val e = parseMinute(end) ?: return false
            if (e < s) return false
            return (e - s) < 12L * 60L * 60L * 1000L
        }

        private fun parseMinute(value: String): Long? {
            // yyyy-MM-dd HH:mm → epoch via OfflineOpsTimeRanges day start + time-of-day
            val parts = value.trim().split(' ')
            if (parts.size != 2) return null
            val dayMs = OfflineOpsTimeRanges.parseDateStart(parts[0]) ?: return null
            val t = parts[1].split(':').mapNotNull { it.toIntOrNull() }
            if (t.size < 2) return null
            val h = t[0]
            val mi = t[1]
            if (h !in 0..23 || mi !in 0..59) return null
            return dayMs + h * 3_600_000L + mi * 60_000L
        }
    }
}
