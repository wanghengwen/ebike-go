package com.luopingtech.ebike.ops.data.staff

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.staff.AddEmployeeRequest
import com.luopingtech.ebike.ops.domain.staff.BatteryConfig
import com.luopingtech.ebike.ops.domain.staff.EmployeeTrackResult
import com.luopingtech.ebike.ops.domain.staff.HomeTabCatalog
import com.luopingtech.ebike.ops.domain.staff.MoveCarConfig
import com.luopingtech.ebike.ops.domain.staff.StaffEmployee
import com.luopingtech.ebike.ops.domain.staff.StaffEmployeePageQuery
import com.luopingtech.ebike.ops.domain.staff.StaffRealtimeLocation
import com.luopingtech.ebike.ops.domain.staff.StaffRole
import com.luopingtech.ebike.ops.domain.staff.StaffRoleOption
import com.luopingtech.ebike.ops.domain.staff.TrackOperation
import com.luopingtech.ebike.ops.domain.staff.TrackPoint
import com.luopingtech.ebike.ops.domain.staff.ViewConfig
import com.luopingtech.ebike.ops.domain.staff.VoltagePlanOption

interface StaffManagementRepository {
    suspend fun pageEmployees(query: StaffEmployeePageQuery): OpsResult<List<StaffEmployee>>
    suspend fun addEmployee(req: AddEmployeeRequest): OpsResult<Unit>
    suspend fun setEmployeeStatus(id: String, enabled: Boolean): OpsResult<Unit>
    suspend fun deleteEmployee(id: String): OpsResult<Unit>
    suspend fun roleListForFilter(subTenantId: String): OpsResult<List<StaffRoleOption>>
    suspend fun roleAppList(
        userId: String,
        status: Int,
        subTenantId: String,
        name: String? = null,
    ): OpsResult<List<StaffRole>>
    suspend fun roleDisableEnabled(id: String, name: String, status: Int): OpsResult<Unit>
    suspend fun roleDelete(id: String, name: String, status: Int): OpsResult<Unit>
    suspend fun getHomeTabsByRoleId(roleId: String): OpsResult<List<String>>
    suspend fun saveHomeTabs(roleId: String, tabCodeList: List<String>): OpsResult<Unit>
    suspend fun getViewConfig(roleId: String): OpsResult<ViewConfig>
    suspend fun saveMoveCarConfig(roleId: String, config: MoveCarConfig): OpsResult<Unit>
    suspend fun saveBatteryConfig(roleId: String, config: BatteryConfig): OpsResult<Unit>
    suspend fun voltagePlanList(): OpsResult<List<VoltagePlanOption>>
    suspend fun getTrack(userPin: String, startTime: String, endTime: String): OpsResult<EmployeeTrackResult>
    suspend fun realTimeLocation(userPin: String? = null): OpsResult<List<StaffRealtimeLocation>>
    suspend fun getAllOperations(
        userPin: String,
        startTime: String? = null,
        endTime: String? = null,
    ): OpsResult<List<TrackOperation>>
}

class StaffManagementRepositoryImpl(
    private val demoMode: Boolean,
    private val api: StaffManagementApi? = null,
) : StaffManagementRepository {

    private val demoEmployees = mutableListOf(
        StaffEmployee(
            id = "e1",
            name = Strings.t(Str.DemoTeamWorkerA),
            phone = "+86-13800001111",
            roleName = "换电工",
            createdAt = "2025-01-10 09:00:00",
            enabled = true,
            pin = "demo-a",
            profession = "换电",
        ),
        StaffEmployee(
            id = "e2",
            name = Strings.t(Str.DemoTeamWorkerB),
            phone = "+86-13900002222",
            roleName = "挪车员",
            createdAt = "2025-02-12 14:30:00",
            enabled = true,
            pin = "demo-b",
            profession = "挪车",
        ),
        StaffEmployee(
            id = "e3",
            name = Strings.t(Str.DemoTeamWorkerC),
            phone = "+86-13700003333",
            roleName = "巡检员",
            createdAt = "2025-03-01 11:20:00",
            enabled = false,
            pin = "demo-c",
            profession = "巡检",
        ),
        StaffEmployee(
            id = "e4",
            name = "张三",
            phone = "+86-13600004444",
            roleName = "admin",
            createdAt = "2024-12-01 08:00:00",
            enabled = true,
            pin = "demo-admin",
            profession = "管理",
            izRoot = true,
        ),
    )

    private val demoRoles = mutableListOf(
        StaffRole(
            id = "r1",
            name = "admin",
            status = 1,
            remark = "系统管理员",
            serviceName = "全部服务区",
            tabCodeList = HomeTabCatalog.allOptions.map { it.code },
        ),
        StaffRole(
            id = "r2",
            name = "换电工",
            status = 1,
            remark = "负责换电",
            serviceName = "城东服务区",
            tabCodeList = listOf("123001", "123002", "1202", "1224"),
        ),
        StaffRole(
            id = "r3",
            name = "挪车员",
            status = 1,
            remark = "负责挪车",
            serviceName = "城西服务区",
            tabCodeList = listOf("123001", "123002", "1203", "1211"),
        ),
        StaffRole(
            id = "r4",
            name = "停用角色",
            status = 0,
            remark = "已停用",
            serviceName = "-",
            tabCodeList = emptyList(),
        ),
    )

    private val demoViewConfigs = mutableMapOf(
        "r1" to ViewConfig(
            roleId = "r1",
            moveCar = MoveCarConfig(carStatus = listOf(1), carIdStart = "0", carIdEnd = "0"),
            battery = BatteryConfig(voltagePlanIds = listOf(101, 102), carIdStart = "0", carIdEnd = "0"),
        ),
        "r2" to ViewConfig(
            roleId = "r2",
            moveCar = MoveCarConfig(carStatus = listOf(5, 11), carIdStart = "100000001", carIdEnd = "100000099"),
            battery = BatteryConfig(voltagePlanIds = listOf(101), carIdStart = "0", carIdEnd = "0"),
        ),
        "r3" to ViewConfig(
            roleId = "r3",
            moveCar = MoveCarConfig(carStatus = listOf(6, 8, 13), carIdStart = "0", carIdEnd = "0"),
            battery = BatteryConfig(),
        ),
    )

    private val demoHomeTabs = mutableMapOf(
        "r1" to HomeTabCatalog.allOptions.map { it.code },
        "r2" to listOf("123001", "123002", "1202", "1224"),
        "r3" to listOf("123001", "123002", "1203", "1211"),
        "r4" to listOf("123001", "123002"),
    )

    private val demoVoltagePlans = listOf(
        VoltagePlanOption(101, "48V标准方案"),
        VoltagePlanOption(102, "60V长续航"),
        VoltagePlanOption(103, "72V商用"),
    )

    override suspend fun pageEmployees(query: StaffEmployeePageQuery): OpsResult<List<StaffEmployee>> {
        if (demoMode || api == null) {
            val wantEnabled = query.status == 1
            var list = demoEmployees.filter { it.enabled == wantEnabled }
            query.nameOrPhone?.takeIf { it.isNotBlank() }?.let { kw ->
                list = list.filter {
                    it.name.contains(kw, ignoreCase = true) ||
                        it.phone.contains(kw) ||
                        it.dialPhone.contains(kw)
                }
            }
            query.roleIds?.takeIf { it.isNotEmpty() }?.let { ids ->
                val names = demoRoles.filter { it.id in ids }.map { it.name }.toSet()
                list = list.filter { it.roleName in names || it.id in ids }
            }
            val from = ((query.pageNum - 1) * query.pageSize).coerceAtLeast(0)
            return OpsResult.Ok(list.drop(from).take(query.pageSize))
        }
        return api.pageEmployees(query)
    }

    override suspend fun addEmployee(req: AddEmployeeRequest): OpsResult<Unit> {
        if (demoMode || api == null) {
            val roleName = demoRoles.find { it.id in req.roleIds }?.name.orEmpty()
            demoEmployees.add(
                0,
                StaffEmployee(
                    id = "e${demoEmployees.size + 10}",
                    name = req.name,
                    phone = req.phone,
                    roleName = roleName,
                    createdAt = "2026-09-17 12:00:00",
                    enabled = true,
                    pin = "demo-${demoEmployees.size + 10}",
                    profession = req.profession,
                    izRoot = req.izRoot,
                ),
            )
            return OpsResult.Ok(Unit)
        }
        return api.addEmployee(req)
    }

    override suspend fun setEmployeeStatus(id: String, enabled: Boolean): OpsResult<Unit> {
        if (demoMode || api == null) {
            val idx = demoEmployees.indexOfFirst { it.id == id }
            if (idx < 0) return OpsResult.Err(OpsError.business("STAFF", "not found"))
            demoEmployees[idx] = demoEmployees[idx].copy(enabled = enabled)
            return OpsResult.Ok(Unit)
        }
        return api.setEmployeeStatus(id, enabled)
    }

    override suspend fun deleteEmployee(id: String): OpsResult<Unit> {
        if (demoMode || api == null) {
            val emp = demoEmployees.find { it.id == id }
                ?: return OpsResult.Err(OpsError.business("STAFF", "not found"))
            if (emp.enabled) {
                return OpsResult.Err(OpsError.business("STAFF", Strings.t(Str.StaffDisableBeforeDelete)))
            }
            demoEmployees.removeAll { it.id == id }
            return OpsResult.Ok(Unit)
        }
        return api.deleteEmployee(id)
    }

    override suspend fun roleListForFilter(subTenantId: String): OpsResult<List<StaffRoleOption>> {
        if (demoMode || api == null) {
            return OpsResult.Ok(demoRoles.filter { it.status == 1 }.map { StaffRoleOption(it.id, it.name) })
        }
        return api.roleListForFilter(subTenantId)
    }

    override suspend fun roleAppList(
        userId: String,
        status: Int,
        subTenantId: String,
        name: String?,
    ): OpsResult<List<StaffRole>> {
        if (demoMode || api == null) {
            var list = demoRoles.filter { it.status == status }.map { role ->
                role.copy(tabCodeList = demoHomeTabs[role.id] ?: role.tabCodeList)
            }
            name?.takeIf { it.isNotBlank() }?.let { kw ->
                list = list.filter { it.name.contains(kw, ignoreCase = true) }
            }
            return OpsResult.Ok(list)
        }
        return api.roleAppList(userId, status, subTenantId, name)
    }

    override suspend fun roleDisableEnabled(id: String, name: String, status: Int): OpsResult<Unit> {
        if (demoMode || api == null) {
            val idx = demoRoles.indexOfFirst { it.id == id }
            if (idx < 0) return OpsResult.Err(OpsError.business("STAFF", "not found"))
            val target = if (status == 1) 0 else 1
            demoRoles[idx] = demoRoles[idx].copy(status = target)
            return OpsResult.Ok(Unit)
        }
        return api.roleDisableEnabled(id, name, status)
    }

    override suspend fun roleDelete(id: String, name: String, status: Int): OpsResult<Unit> {
        if (demoMode || api == null) {
            demoRoles.removeAll { it.id == id }
            demoHomeTabs.remove(id)
            demoViewConfigs.remove(id)
            return OpsResult.Ok(Unit)
        }
        return api.roleDelete(id, name, status)
    }

    override suspend fun getHomeTabsByRoleId(roleId: String): OpsResult<List<String>> {
        if (demoMode || api == null) {
            return OpsResult.Ok(
                HomeTabCatalog.withForcedBasics(demoHomeTabs[roleId].orEmpty()),
            )
        }
        return api.getHomeTabsByRoleId(roleId)
    }

    override suspend fun saveHomeTabs(roleId: String, tabCodeList: List<String>): OpsResult<Unit> {
        if (demoMode || api == null) {
            val codes = HomeTabCatalog.withForcedBasics(tabCodeList)
            demoHomeTabs[roleId] = codes
            val idx = demoRoles.indexOfFirst { it.id == roleId }
            if (idx >= 0) demoRoles[idx] = demoRoles[idx].copy(tabCodeList = codes)
            return OpsResult.Ok(Unit)
        }
        return api.saveHomeTabs(roleId, HomeTabCatalog.withForcedBasics(tabCodeList))
    }

    override suspend fun getViewConfig(roleId: String): OpsResult<ViewConfig> {
        if (demoMode || api == null) {
            return OpsResult.Ok(demoViewConfigs[roleId] ?: ViewConfig(roleId = roleId))
        }
        return api.getViewConfig(roleId)
    }

    override suspend fun saveMoveCarConfig(roleId: String, config: MoveCarConfig): OpsResult<Unit> {
        if (demoMode || api == null) {
            val cur = demoViewConfigs[roleId] ?: ViewConfig(roleId = roleId)
            demoViewConfigs[roleId] = cur.copy(moveCar = config)
            return OpsResult.Ok(Unit)
        }
        return api.saveMoveCarConfig(roleId, config)
    }

    override suspend fun saveBatteryConfig(roleId: String, config: BatteryConfig): OpsResult<Unit> {
        if (demoMode || api == null) {
            val cur = demoViewConfigs[roleId] ?: ViewConfig(roleId = roleId)
            demoViewConfigs[roleId] = cur.copy(battery = config)
            return OpsResult.Ok(Unit)
        }
        return api.saveBatteryConfig(roleId, config)
    }

    override suspend fun voltagePlanList(): OpsResult<List<VoltagePlanOption>> {
        if (demoMode || api == null) return OpsResult.Ok(demoVoltagePlans)
        return api.voltagePlanList()
    }

    override suspend fun getTrack(
        userPin: String,
        startTime: String,
        endTime: String,
    ): OpsResult<EmployeeTrackResult> {
        if (demoMode || api == null) {
            return OpsResult.Ok(
                EmployeeTrackResult(
                    points = listOf(
                        TrackPoint(createdAt = startTime, lng = 120.15, lat = 30.28),
                        TrackPoint(createdAt = startTime, lng = 120.16, lat = 30.285),
                        TrackPoint(createdAt = endTime, lng = 120.17, lat = 30.29),
                    ),
                    operations = listOf(
                        TrackOperation(
                            id = "op1",
                            operationType = "换电",
                            createdAt = startTime,
                            carId = "100000001",
                            opMan = userPin,
                        ),
                        TrackOperation(
                            id = "op2",
                            operationType = "挪车",
                            createdAt = endTime,
                            carId = "100000002",
                            opMan = userPin,
                        ),
                    ),
                ),
            )
        }
        return api.getTrack(userPin, startTime, endTime)
    }

    override suspend fun realTimeLocation(userPin: String?): OpsResult<List<StaffRealtimeLocation>> {
        if (demoMode || api == null) {
            return OpsResult.Ok(
                demoEmployees.take(2).mapIndexed { idx, emp ->
                    StaffRealtimeLocation(
                        id = emp.id,
                        pin = emp.pin,
                        name = emp.name,
                        phone = emp.phone,
                        points = listOf(
                            TrackPoint(
                                createdAt = "2026-09-17 12:00:00",
                                lng = 120.15 + idx * 0.01,
                                lat = 30.28 + idx * 0.01,
                            ),
                        ),
                    )
                },
            )
        }
        return api.realTimeLocation(userPin)
    }

    override suspend fun getAllOperations(
        userPin: String,
        startTime: String?,
        endTime: String?,
    ): OpsResult<List<TrackOperation>> {
        if (demoMode || api == null) {
            return OpsResult.Ok(
                listOf(
                    TrackOperation(
                        id = "op1",
                        operationType = "换电",
                        createdAt = startTime.orEmpty().ifBlank { "2026-09-17 10:00:00" },
                        carId = "100000001",
                        opMan = userPin,
                    ),
                    TrackOperation(
                        id = "op2",
                        operationType = "开锁",
                        createdAt = endTime.orEmpty().ifBlank { "2026-09-17 11:00:00" },
                        carId = "100000003",
                        opMan = userPin,
                    ),
                ),
            )
        }
        return api.getAllOperations(userPin, startTime, endTime)
    }
}
