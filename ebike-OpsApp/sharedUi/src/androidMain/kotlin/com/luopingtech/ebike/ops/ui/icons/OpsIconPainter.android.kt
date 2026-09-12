package com.luopingtech.ebike.ops.ui.icons

import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.painter.ColorPainter
import androidx.compose.ui.graphics.painter.Painter
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.painterResource

/**
 * 通过宿主 package 的 drawable 名解析资源，避免 sharedUi 依赖 androidApp 的 R 类。
 */
@Composable
actual fun painterResource(icon: OpsIcon): Painter {
    val context = LocalContext.current
    val name = icon.drawableName()
    val id = context.resources.getIdentifier(name, "drawable", context.packageName)
    if (id == 0) {
        return ColorPainter(Color(0xFFB0B0B0))
    }
    return painterResource(id)
}

private fun OpsIcon.drawableName(): String = when (this) {
    OpsIcon.TenantLogo -> "tenant_logo"
    OpsIcon.ArrowDown -> "arrow_down"
    OpsIcon.Setting -> "setting"
    OpsIcon.MineModuleAddGray -> "mine_module_add_gray"
    OpsIcon.MineModuleEditAdd -> "mine_module_edit_add"
    OpsIcon.MineModuleEditDelete -> "mine_module_edit_delete"
    OpsIcon.BgIconTaskCenter -> "bg_icon_task_center"
    OpsIcon.BgAnalysis -> "bg_analysis"
    OpsIcon.VehicleList -> "vehicle_list"
    OpsIcon.ReplaceBattery -> "replace_battery"
    OpsIcon.MoveVehicle -> "move_vehicle"
    OpsIcon.Repair -> "repair"
    OpsIcon.UnlockedVehicle -> "unlocked_vehicle_icon"
    OpsIcon.BluetoothRadar -> "btn_bluetooth_radar"
    OpsIcon.OperationSetting -> "ic_operation_setting"
    OpsIcon.MyTask -> "mine_my_task"
    OpsIcon.VehicleTag -> "mine_vehicle_tag"
    OpsIcon.Relocation -> "mine_relocation"
    OpsIcon.ParkingArea -> "btn_parking_area"
    OpsIcon.OrderQuery -> "ic_order_query"
    OpsIcon.OperationScreen -> "btn_operation"
    OpsIcon.RevenueScreen -> "btn_revenue"
    OpsIcon.EmployeeManager -> "ic_employee_manager"
    OpsIcon.ProfessionAudit -> "mine_profession_audit_ic"
    OpsIcon.ObjectionOrder -> "mine_objection_order"
    OpsIcon.BlackList -> "black_list"
    OpsIcon.IdAudit -> "mine_id_audit_ic"
    OpsIcon.OperationLog -> "ic_operation_log"
    OpsIcon.VehicleInspection -> "ic_vehicle_inspection"
    OpsIcon.CenterControlBind -> "btn_center_control_bind"
    OpsIcon.PutPullShelves -> "btn_put_pull_shelves"
    OpsIcon.InWarehouse -> "ic_in_warehouse"
    OpsIcon.OutWarehouse -> "ic_out_warehouse"
    OpsIcon.WarehouseRecord -> "ic_warehouse_record"
    OpsIcon.OfflineOperation -> "ic_offline_operation"
    OpsIcon.AnalysisStation -> "icon_analysis_station"
    OpsIcon.AnalysisReturnBike -> "icon_analysis_return_bike"
    OpsIcon.AnalysisVehicleDistribution -> "icon_analysis_vehicle_distribution"
    OpsIcon.TaskChangeBattery -> "icon_task_change_battery"
    OpsIcon.TaskMoveBike -> "icon_task_move_bike"
    OpsIcon.TaskInspection -> "icon_task_inspection"
    OpsIcon.TaskRepair -> "icon_task_repair"
}
