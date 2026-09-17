package com.luopingtech.ebike.ops.domain.staff

/**
 * Staff management domain models — aligned with Merchant-Android employee APIs.
 */

data class StaffEmployee(
    val id: String,
    val name: String,
    val phone: String = "",
    val roleName: String = "",
    val createdAt: String = "",
    val enabled: Boolean = true,
    val pin: String = "",
    val profession: String = "",
    val izRoot: Boolean = false,
    val agentIds: List<String> = emptyList(),
) {
    val userPin: String get() = pin
    val dialPhone: String
        get() = phone.removePrefix("+86-").removePrefix("+86").trim()
}

data class StaffRole(
    val id: String,
    val name: String,
    val status: Int = 1,
    val remark: String = "",
    val serviceName: String = "",
    val tabCodeList: List<String> = emptyList(),
) {
    val isAdmin: Boolean get() = name.equals("admin", ignoreCase = false)
    val permissionSummary: String
        get() = if (tabCodeList.isEmpty()) {
            ""
        } else {
            val labels = tabCodeList.mapNotNull { code ->
                HomeTabCatalog.allOptions.find { it.code == code }?.labelZh
            }.take(2)
            if (labels.isEmpty()) {
                ""
            } else {
                labels.joinToString("、") + " 等${tabCodeList.size}个"
            }
        }
}

data class StaffRoleOption(
    val id: String,
    val name: String,
)

data class StaffEmployeePageQuery(
    val pageNum: Int = 1,
    val pageSize: Int = 10,
    /** 1 enabled / 0 disabled */
    val status: Int = 1,
    val nameOrPhone: String? = null,
    val roleIds: List<String>? = null,
    val agentIds: List<String>? = null,
)

data class AddEmployeeRequest(
    val name: String,
    val phone: String,
    val roleIds: List<String>,
    val profession: String,
    val izRoot: Boolean = false,
)

data class HomeTabOption(
    val code: String,
    val labelZh: String,
    val labelEn: String,
    val groupCode: String,
    val forced: Boolean = false,
)

data class HomeTabGroup(
    val groupCode: String,
    val labelZh: String,
    val labelEn: String,
    val options: List<HomeTabOption>,
)

object HomeTabCatalog {
    const val GROUP_BASIC = "1230"
    const val GROUP_MAINTENANCE = "maintenanceModule"
    const val GROUP_OPERATION = "operationModule"
    const val GROUP_PRODUCTION = "productionModule"

    const val CODE_CHANGE_BATTERY_MAP = "123001"
    const val CODE_MOVE_CAR_MAP = "123002"

    val allOptions: List<HomeTabOption> = listOf(
        HomeTabOption(CODE_CHANGE_BATTERY_MAP, "换电地图", "Battery map", GROUP_BASIC, forced = true),
        HomeTabOption(CODE_MOVE_CAR_MAP, "挪车地图", "Move-car map", GROUP_BASIC, forced = true),
        HomeTabOption("1201", "车辆列表", "Vehicle list", GROUP_MAINTENANCE),
        HomeTabOption("1202", "换电", "Battery swap", GROUP_MAINTENANCE),
        HomeTabOption("1203", "挪车", "Move car", GROUP_MAINTENANCE),
        HomeTabOption("1209", "维修工单", "Repair orders", GROUP_MAINTENANCE),
        HomeTabOption("1208", "巡检工单", "Inspection orders", GROUP_MAINTENANCE),
        HomeTabOption("1210", "报修", "Fault report", GROUP_MAINTENANCE),
        HomeTabOption("1224", "换电阈值", "Battery threshold", GROUP_MAINTENANCE),
        HomeTabOption("1211", "停车区", "Parking zone", GROUP_MAINTENANCE),
        HomeTabOption("1218", "员工管理", "Staff manage", GROUP_OPERATION),
        HomeTabOption("1225", "拉黑名单", "Blacklist", GROUP_OPERATION),
        HomeTabOption("1205", "中控绑定", "Device bind", GROUP_PRODUCTION),
        HomeTabOption("1207", "上下架", "Shelf on/off", GROUP_PRODUCTION),
    )

    val groups: List<HomeTabGroup> = listOf(
        HomeTabGroup(GROUP_BASIC, "基本权限", "Basic", allOptions.filter { it.groupCode == GROUP_BASIC }),
        HomeTabGroup(GROUP_MAINTENANCE, "运维模块", "Ops", allOptions.filter { it.groupCode == GROUP_MAINTENANCE }),
        HomeTabGroup(GROUP_OPERATION, "运营模块", "Operation", allOptions.filter { it.groupCode == GROUP_OPERATION }),
        HomeTabGroup(GROUP_PRODUCTION, "生产模块", "Production", allOptions.filter { it.groupCode == GROUP_PRODUCTION }),
    )

    fun withForcedBasics(selected: Collection<String>): List<String> {
        val set = selected.toMutableSet()
        set.add(CODE_CHANGE_BATTERY_MAP)
        set.add(CODE_MOVE_CAR_MAP)
        return allOptions.map { it.code }.filter { it in set }
    }
}

data class MoveCarConfig(
    val carStatus: List<Long> = emptyList(),
    val carIdStart: String = "0",
    val carIdEnd: String = "0",
)

data class BatteryConfig(
    val voltagePlanIds: List<Long> = emptyList(),
    val carIdStart: String = "0",
    val carIdEnd: String = "0",
)

data class ViewConfig(
    val roleId: String,
    val moveCar: MoveCarConfig = MoveCarConfig(),
    val battery: BatteryConfig = BatteryConfig(),
)

data class VoltagePlanOption(
    val id: Long,
    val name: String,
)

data class CarStatusOption(
    val filterCode: Long,
    val labelZh: String,
    val labelEn: String,
)

object CarStatusCatalog {
    val options: List<CarStatusOption> = listOf(
        CarStatusOption(1, "全部", "All"),
        CarStatusOption(5, "可使用", "Available"),
        CarStatusOption(6, "骑行中", "Riding"),
        CarStatusOption(7, "被预约", "Reserved"),
        CarStatusOption(8, "临时停车", "Temp park"),
        CarStatusOption(9, "报修", "Repair"),
        CarStatusOption(10, "拖回", "Tow back"),
        CarStatusOption(11, "低电量", "Low battery"),
        CarStatusOption(12, "换电中", "Swapping"),
        CarStatusOption(13, "站点外", "Out of station"),
        CarStatusOption(14, "异常移动", "Abnormal move"),
        CarStatusOption(15, "出服务区", "Out of area"),
        CarStatusOption(16, "电瓶移除", "Battery removed"),
        CarStatusOption(17, "异常离线", "Abnormal offline"),
        CarStatusOption(18, "有单无程", "Order no trip"),
        CarStatusOption(20, "车辆报失", "Lost"),
        CarStatusOption(21, "订单超长", "Long order"),
        CarStatusOption(22, "短时订单", "Short order"),
        CarStatusOption(23, "开锁异常", "Unlock anomaly"),
        CarStatusOption(24, "禁停区", "No-parking"),
        CarStatusOption(28, "调度中", "Dispatching"),
        CarStatusOption(29, "已下架", "Off shelf"),
        CarStatusOption(31, "头盔丢失", "Helmet lost"),
        CarStatusOption(32, "头盔故障", "Helmet fault"),
    )
}

data class TrackPoint(
    val createdAt: String = "",
    val lng: Double = 0.0,
    val lat: Double = 0.0,
)

data class TrackOperation(
    val id: String = "",
    val operationType: String = "",
    val createdAt: String = "",
    val carId: String = "",
    val opMan: String = "",
    val phone: String = "",
)

data class EmployeeTrackResult(
    val points: List<TrackPoint> = emptyList(),
    val operations: List<TrackOperation> = emptyList(),
)

/** 员工实时位置（realTimeLocation 接口）。 */
data class StaffRealtimeLocation(
    val id: String = "",
    val pin: String = "",
    val name: String = "",
    val phone: String = "",
    val points: List<TrackPoint> = emptyList(),
)

enum class ViewConfigEditKind {
    MoveCar,
    Battery,
}
