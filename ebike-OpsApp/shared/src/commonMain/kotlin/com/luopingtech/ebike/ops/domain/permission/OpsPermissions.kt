package com.luopingtech.ebike.ops.domain.permission

/**
 * Legacy Merchant-Android [PermissionsCode] subset used by the field-ops shell.
 * Keep string values identical so remote `/getUserByToken` codes work unchanged.
 */
object OpsPermissionCodes {
    const val SCAN: String = "1223"
    const val SCAN_DETAILS: String = "122301"
    const val SCAN_SWITCH_LOCK: String = "122302"

    const val HOME_MAP: String = "123004"
    /** 首页优先聚合模式。遗留 PCODE_HOME_CLUSTER_FIRST。 */
    const val HOME_CLUSTER_FIRST: String = "12300401"
    /** 首页底部状态统计栏。遗留 PCODE_HOME_STATISTICS_SHOW。 */
    const val HOME_STATISTICS_SHOW: String = "12300402"
    /** 首页右侧告警筛选。遗留 PCODE_HOME_STATISTICS_FILTER_SHOW。 */
    const val HOME_STATISTICS_FILTER_SHOW: String = "12300403"

    const val TASK_CENTER: String = "1234"
    const val TASK_CHANGE_BATTERY: String = "123401"
    const val TASK_MOVE_CAR: String = "123402"
    const val TASK_INSPECTION: String = "123403"
    const val TASK_REPAIR: String = "123404"
    /** 任务指派。遗留 PCODE_TASK_ASSIGN。 */
    const val TASK_ASSIGN: String = "123405"

    /** 运维模块父节点。遗留 PCODE_MAINTENANCE。 */
    const val MAINTENANCE_MODULE: String = "maintenanceModule"
    /** 运营模块父节点。遗留 PCODE_OPERATION。 */
    const val OPERATION_MODULE: String = "operationModule"

    const val WAREHOUSE_MODULE: String = "warehouseModule"
    const val WAREHOUSE_IN: String = "1231"
    const val WAREHOUSE_OUT: String = "1232"
    const val WAREHOUSE_RECORD: String = "1233"

    const val PRODUCTION_MODULE: String = "productionModule"
    const val PRODUCTION_DETECT: String = "1204"
    const val PRODUCTION_BIND: String = "1205"
    const val PRODUCTION_SHELVES: String = "1207"

    /** 运维报修上报（审核侧不在 App）。遗留 PermissionsCode: 1210=报修。 */
    const val FAULT_REPORT: String = "1210"

    /** 运维实名举报（sneak）。遗留 PCODE_MAINTENANCE_REPORT。与 1210 报修独立。 */
    const val SNEAK_REPORT: String = "1214"

    /** 详情绑定电池 SN。遗留 PermissionsCode.PCODE_BIND_BATTERY_VIEW。 */
    const val BIND_BATTERY_SN: String = "120201"

    /** 未关锁车辆列表（工作台工具）。 */
    const val UNLOCKED_VEHICLES: String = "1228"
    /** 未关锁 · 员工记录筛选（姓名/手机号）。无无此码则强制按本人手机号筛。 */
    const val UNLOCKED_STAFF_RECORDS: String = "122802"

    /** 更改车辆定位（扫码 + GPS 回写）。遗留 PCODE_MAINTENANCE_RELOCATION。 */
    const val RELOCATION: String = "1239"

    /** 车辆列表。遗留 PCODE_MAINTENANCE_VEHICLE_LIST。 */
    const val VEHICLE_LIST: String = "1201"
    /** 换电。遗留 PCODE_MAINTENANCE_CHANGE_BATTERY。 */
    const val CHANGE_BATTERY: String = "1202"
    /** 挪车。遗留 PCODE_MAINTENANCE_MOVE_CAR。 */
    const val MOVE_CAR: String = "1203"
    /** 蓝牙雷达。遗留 PCODE_MAINTENANCE_BLUETOOTH_RADAR。 */
    const val BLUETOOTH_RADAR: String = "1229"
    /** 运维设置阈值。遗留 PCODE_MAINTENANCE_SETTING_THRESHOLD。 */
    const val OPS_SETTING: String = "1224"
    /** 我的任务。遗留 PCODE_MAINTENANCE_MY_TASK。 */
    const val MY_TASK: String = "1238"
    /** 车辆标记。遗留 PCODE_MAINTENANCE_VEHICLE_TAG。 */
    const val VEHICLE_TAG: String = "1236"
    /** 站点围栏。遗留 PCODE_MAINTENANCE_PARKING_LOT。 */
    const val PARKING_FENCE: String = "1211"
    /** 员工管理。遗留 PCODE_OPERATION_STAFF_MANAGE。 */
    const val STAFF_MANAGE: String = "1218"
    /** 职业认证审核。 */
    const val PROFESSION_AUDIT: String = "1221"
    /** 异议工单。 */
    const val OBJECTION_ORDER: String = "1226"
    /** 拉黑名单。 */
    const val BLACKLIST: String = "1225"
    /** 实名换绑审核。 */
    const val ID_BIND_AUDIT: String = "1222"
    /** 操作日志。 */
    const val OPERATION_LOG: String = "1227"

    /** 数据分析模块入口。 */
    const val ANALYSIS_MODULE: String = "dataAnalysisModule"
    /** 线下运维。 */
    const val ANALYSIS_OFFLINE_OPS: String = "1217"
    /** 站点监控。 */
    const val ANALYSIS_STATION: String = "1240"
    /** 还车分布。 */
    const val ANALYSIS_RETURN_CAR: String = "1243"
    /** 车况分布。 */
    const val ANALYSIS_VEHICLE_DIST: String = "1251"

    /** 旧巡检工单台账（非任务中心 123403）。遗留 PCODE_MAINTENANCE_INSPECTION_WORK_ORDER。 */
    const val INSPECTION_ORDER: String = "1208"
    /** 巡检工单接单/完单。 */
    const val INSPECTION_ORDER_TAKE: String = "120801"
    /** 旧维修工单台账（非任务中心 123404）。遗留 PCODE_MAINTENANCE_REPAIR_WORK_ORDER。 */
    const val REPAIR_ORDER: String = "1209"
    /** 维修工单接单/完单。 */
    const val REPAIR_ORDER_TAKE: String = "120901"

    /** 运营数据大屏（WebView H5）。 */
    const val OPERATION_DATA: String = "0221"
    /** 营收数据大屏（WebView H5）。 */
    const val REVENUE_DATA: String = "LargeRevenueScreen"

    /** 订单查询（WebView H5）。遗留 PCODE_OPERATION_ORDER_QUERY。 */
    const val ORDER_QUERY: String = "1212"
    /** 订单查询 · 按车号 / IMEI 检索。无此码则搜索框不受理车辆类输入。 */
    const val ORDER_QUERY_VEHICLE: String = "121201"
    /** 订单查询 · 按手机号 / 姓名检索。无此码则搜索框不受理人员类输入。 */
    const val ORDER_QUERY_PERSONAL: String = "121202"

    /**
     * PC 后台「用户订单 / 骑行订单」菜单码。同一个功能 PC 侧发这个、App 侧发 [ORDER_QUERY]，
     * 两边发码习惯不统一，入口按两者取其一放行。
     */
    const val PC_ORDER_MENU: String = "0204"

    /** Demo full field-ops set (includes optional H5 dashboards). */
    val FIELD_OPS_ALL: Set<String> = setOf(
        SCAN,
        SCAN_DETAILS,
        SCAN_SWITCH_LOCK,
        HOME_MAP,
        HOME_CLUSTER_FIRST,
        HOME_STATISTICS_SHOW,
        HOME_STATISTICS_FILTER_SHOW,
        TASK_CENTER,
        TASK_CHANGE_BATTERY,
        TASK_MOVE_CAR,
        TASK_INSPECTION,
        TASK_REPAIR,
        TASK_ASSIGN,
        MAINTENANCE_MODULE,
        OPERATION_MODULE,
        WAREHOUSE_MODULE,
        WAREHOUSE_IN,
        WAREHOUSE_OUT,
        WAREHOUSE_RECORD,
        PRODUCTION_MODULE,
        PRODUCTION_DETECT,
        PRODUCTION_BIND,
        PRODUCTION_SHELVES,
        FAULT_REPORT,
        SNEAK_REPORT,
        BIND_BATTERY_SN,
        UNLOCKED_VEHICLES,
        UNLOCKED_STAFF_RECORDS,
        RELOCATION,
        VEHICLE_LIST,
        CHANGE_BATTERY,
        MOVE_CAR,
        BLUETOOTH_RADAR,
        OPS_SETTING,
        MY_TASK,
        VEHICLE_TAG,
        PARKING_FENCE,
        STAFF_MANAGE,
        PROFESSION_AUDIT,
        OBJECTION_ORDER,
        BLACKLIST,
        ID_BIND_AUDIT,
        OPERATION_LOG,
        ANALYSIS_MODULE,
        ANALYSIS_OFFLINE_OPS,
        ANALYSIS_STATION,
        ANALYSIS_RETURN_CAR,
        ANALYSIS_VEHICLE_DIST,
        INSPECTION_ORDER,
        INSPECTION_ORDER_TAKE,
        REPAIR_ORDER,
        REPAIR_ORDER_TAKE,
        OPERATION_DATA,
        REVENUE_DATA,
        ORDER_QUERY,
        ORDER_QUERY_VEHICLE,
        ORDER_QUERY_PERSONAL,
    )

    /** Map + mine only — for demo account `limited`. */
    val FIELD_OPS_LIMITED: Set<String> = setOf(HOME_MAP)
}

/**
 * Thin gate over permission codes. Empty remote codes hide gated entries
 * (same spirit as legacy providers).
 */
class OpsPermissions(
    private val codes: Set<String>,
) {
    fun has(code: String): Boolean = codes.contains(code)

    fun hasAny(vararg code: String): Boolean = code.any { codes.contains(it) }

    val showMap: Boolean get() = has(OpsPermissionCodes.HOME_MAP)
    /** 有码则默认开聚合；无码仍可手动切「详情」。 */
    val homeClusterFirst: Boolean get() = has(OpsPermissionCodes.HOME_CLUSTER_FIRST)
    /** 底部 8 格状态统计。 */
    val showHomeStatistics: Boolean get() = has(OpsPermissionCodes.HOME_STATISTICS_SHOW)
    /** 右侧告警筛选把手。 */
    val showHomeAlarmFilter: Boolean get() = has(OpsPermissionCodes.HOME_STATISTICS_FILTER_SHOW)
    val showScan: Boolean get() = has(OpsPermissionCodes.SCAN)
    val showTaskCenter: Boolean get() = has(OpsPermissionCodes.TASK_CENTER)

    val showChangeBattery: Boolean
        get() = showTaskCenter && has(OpsPermissionCodes.TASK_CHANGE_BATTERY)
    val showMoveCar: Boolean
        get() = showTaskCenter && has(OpsPermissionCodes.TASK_MOVE_CAR)
    val showInspection: Boolean
        get() = showTaskCenter && has(OpsPermissionCodes.TASK_INSPECTION)
    val showRepair: Boolean
        get() = showTaskCenter && has(OpsPermissionCodes.TASK_REPAIR)
    /** 巡检/维修指派。 */
    val canAssignTask: Boolean
        get() = showTaskCenter && has(OpsPermissionCodes.TASK_ASSIGN)

    /** Legacy: parent code must exist or section is GONE. */
    val showMaintainModule: Boolean get() = has(OpsPermissionCodes.MAINTENANCE_MODULE)
    val showOperationModule: Boolean get() = has(OpsPermissionCodes.OPERATION_MODULE)

    val showWarehouse: Boolean
        get() = has(OpsPermissionCodes.WAREHOUSE_MODULE) ||
            hasAny(
                OpsPermissionCodes.WAREHOUSE_IN,
                OpsPermissionCodes.WAREHOUSE_OUT,
                OpsPermissionCodes.WAREHOUSE_RECORD,
            )
    /** Child tiles require their own codes (parent alone hides the section via empty list). */
    val showWarehouseIn: Boolean get() = has(OpsPermissionCodes.WAREHOUSE_IN)
    val showWarehouseOut: Boolean get() = has(OpsPermissionCodes.WAREHOUSE_OUT)
    val showWarehouseRecord: Boolean get() = has(OpsPermissionCodes.WAREHOUSE_RECORD)

    val showProduction: Boolean
        get() = has(OpsPermissionCodes.PRODUCTION_MODULE) ||
            hasAny(
                OpsPermissionCodes.PRODUCTION_DETECT,
                OpsPermissionCodes.PRODUCTION_BIND,
                OpsPermissionCodes.PRODUCTION_SHELVES,
            )
    val showProductionDetect: Boolean get() = has(OpsPermissionCodes.PRODUCTION_DETECT)
    val showProductionBind: Boolean get() = has(OpsPermissionCodes.PRODUCTION_BIND)
    val showProductionShelves: Boolean get() = has(OpsPermissionCodes.PRODUCTION_SHELVES)

    val showFaultReport: Boolean get() = has(OpsPermissionCodes.FAULT_REPORT)

    /** Legacy 1214 — field sneak / violation report (not FaultReport 1210). */
    val showSneakReport: Boolean get() = has(OpsPermissionCodes.SNEAK_REPORT)

    /** Legacy PCODE_BIND_BATTERY_VIEW — bind battery SN on vehicle detail. */
    val canBindBatterySn: Boolean get() = has(OpsPermissionCodes.BIND_BATTERY_SN)

    val showUnlockedVehicles: Boolean get() = has(OpsPermissionCodes.UNLOCKED_VEHICLES)
    /** Legacy hasStaffRecords — show name/phone search; else lock query to self phone. */
    val canFilterUnlockedStaff: Boolean get() = has(OpsPermissionCodes.UNLOCKED_STAFF_RECORDS)

    /** Legacy PCODE_MAINTENANCE_RELOCATION — relocate offline vehicles via GPS. */
    val showRelocation: Boolean get() = has(OpsPermissionCodes.RELOCATION)

    val showVehicleList: Boolean get() = has(OpsPermissionCodes.VEHICLE_LIST)
    val showChangeBatteryTool: Boolean get() = has(OpsPermissionCodes.CHANGE_BATTERY)
    val showMoveCarTool: Boolean get() = has(OpsPermissionCodes.MOVE_CAR)
    val showBluetoothRadar: Boolean get() = has(OpsPermissionCodes.BLUETOOTH_RADAR)
    val showOpsSetting: Boolean get() = has(OpsPermissionCodes.OPS_SETTING)
    val showMyTask: Boolean get() = has(OpsPermissionCodes.MY_TASK)
    val showVehicleTag: Boolean get() = has(OpsPermissionCodes.VEHICLE_TAG)
    val showParkingFence: Boolean get() = has(OpsPermissionCodes.PARKING_FENCE)
    val showStaffManage: Boolean get() = has(OpsPermissionCodes.STAFF_MANAGE)
    val showProfessionAudit: Boolean get() = has(OpsPermissionCodes.PROFESSION_AUDIT)
    val showObjectionOrder: Boolean get() = has(OpsPermissionCodes.OBJECTION_ORDER)
    val showBlacklist: Boolean get() = has(OpsPermissionCodes.BLACKLIST)
    val showIdBindAudit: Boolean get() = has(OpsPermissionCodes.ID_BIND_AUDIT)
    val showOperationLog: Boolean get() = has(OpsPermissionCodes.OPERATION_LOG)

    val showAnalysisTab: Boolean
        get() = has(OpsPermissionCodes.ANALYSIS_MODULE) ||
            hasAny(
                OpsPermissionCodes.ANALYSIS_OFFLINE_OPS,
                OpsPermissionCodes.ANALYSIS_STATION,
                OpsPermissionCodes.ANALYSIS_RETURN_CAR,
                OpsPermissionCodes.ANALYSIS_VEHICLE_DIST,
            )
    val showAnalysisOfflineOps: Boolean get() = has(OpsPermissionCodes.ANALYSIS_OFFLINE_OPS)
    val showAnalysisStation: Boolean get() = has(OpsPermissionCodes.ANALYSIS_STATION)
    val showAnalysisReturnCar: Boolean get() = has(OpsPermissionCodes.ANALYSIS_RETURN_CAR)
    val showAnalysisVehicleDist: Boolean get() = has(OpsPermissionCodes.ANALYSIS_VEHICLE_DIST)

    /** Legacy 1208 — inspection work-order ledger (not task-center). */
    val showInspectionOrder: Boolean get() = has(OpsPermissionCodes.INSPECTION_ORDER)
    val canTakeInspectionOrder: Boolean get() = has(OpsPermissionCodes.INSPECTION_ORDER_TAKE)
    /** Legacy 1209 — repair work-order ledger (not task-center). */
    val showRepairOrder: Boolean get() = has(OpsPermissionCodes.REPAIR_ORDER)
    val canTakeRepairOrder: Boolean get() = has(OpsPermissionCodes.REPAIR_ORDER_TAKE)

    val showOperationScreen: Boolean get() = has(OpsPermissionCodes.OPERATION_DATA)
    val showRevenueScreen: Boolean get() = has(OpsPermissionCodes.REVENUE_DATA)
    /** 订单查询入口：父码 / PC 码 / 叶子码任一即可。 */
    val showOrderQuery: Boolean
        get() = hasAny(
            OpsPermissionCodes.ORDER_QUERY,
            OpsPermissionCodes.PC_ORDER_MENU,
            OpsPermissionCodes.ORDER_QUERY_VEHICLE,
            OpsPermissionCodes.ORDER_QUERY_PERSONAL,
        )

    val canScanUnlock: Boolean get() = has(OpsPermissionCodes.SCAN_SWITCH_LOCK)
    val canScanDetails: Boolean get() = has(OpsPermissionCodes.SCAN_DETAILS)

    val rawCodes: Set<String> get() = codes

    companion object {
        fun fromCodes(codes: Collection<String>): OpsPermissions =
            OpsPermissions(codes.toSet())

        fun demoFull(): OpsPermissions = OpsPermissions(OpsPermissionCodes.FIELD_OPS_ALL)

        fun demoLimited(): OpsPermissions = OpsPermissions(OpsPermissionCodes.FIELD_OPS_LIMITED)

        fun forDemoAccount(account: String): OpsPermissions =
            if (account.trim().equals("limited", ignoreCase = true)) {
                demoLimited()
            } else {
                demoFull()
            }
    }
}
