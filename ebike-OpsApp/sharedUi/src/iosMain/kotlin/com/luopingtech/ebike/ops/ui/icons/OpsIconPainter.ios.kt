package com.luopingtech.ebike.ops.ui.icons

import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.ImageBitmap
import androidx.compose.ui.graphics.painter.BitmapPainter
import androidx.compose.ui.graphics.painter.ColorPainter
import androidx.compose.ui.graphics.painter.Painter
import androidx.compose.ui.graphics.vector.rememberVectorPainter
import kotlinx.cinterop.ExperimentalForeignApi
import kotlinx.cinterop.addressOf
import kotlinx.cinterop.usePinned
import androidx.compose.ui.graphics.toComposeImageBitmap
import org.jetbrains.skia.Image
import platform.Foundation.NSBundle
import platform.Foundation.NSData
import platform.Foundation.dataWithContentsOfFile
import platform.posix.memcpy

/**
 * iOS 图标解析，跟 Android 的 actual 共用一套 drawable 文件名：
 * 宿主把 `androidApp/src/main/res/drawable-xxhdpi` 里的位图原样放进 app bundle
 * （见 iosApp/project.yml 的 OpsAppHost/Media），这里按名字取。
 *
 * 三个矢量图标（扫码页的铅笔、手电开关）Android 侧是 vector XML，iOS 读不了，
 * 在 [OpsVectorIcons] 里用同一份 path 数据重建，两端长得一样。
 */
@Composable
actual fun painterResource(icon: OpsIcon): Painter {
    OpsVectorIcons.of(icon)?.let { return rememberVectorPainter(it) }
    val bitmap = remember(icon) { bitmapCache.getOrPut(icon) { loadBitmap(icon.resourceName()) } }
    return bitmap?.let { BitmapPainter(it) } ?: ColorPainter(FALLBACK)
}

private val bitmapCache = mutableMapOf<OpsIcon, ImageBitmap?>()

private val FALLBACK = Color(0xFFB0B0B0)

/** webp 是 Android 那边打包时压过的格式，Skia 能直接解，不必再转一遍 png。 */
private val EXTENSIONS = listOf("webp", "png", "jpg")

@OptIn(ExperimentalForeignApi::class)
private fun loadBitmap(name: String): ImageBitmap? {
    val path = EXTENSIONS.firstNotNullOfOrNull { ext ->
        NSBundle.mainBundle.pathForResource(name, ofType = ext)
    } ?: return null
    val data = NSData.dataWithContentsOfFile(path) ?: return null
    val size = data.length.toInt()
    if (size == 0) return null
    val bytes = ByteArray(size)
    val source = data.bytes ?: return null
    bytes.usePinned { pinned -> memcpy(pinned.addressOf(0), source, data.length) }
    return runCatching { Image.makeFromEncoded(bytes).toComposeImageBitmap() }.getOrNull()
}

private fun OpsIcon.resourceName(): String = when (this) {
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
    OpsIcon.ScanManual -> "ic_scan_manual"
    OpsIcon.TorchOn -> "ic_torch_on"
    OpsIcon.TorchOff -> "ic_torch_off"
    OpsIcon.TaskChangeBattery -> "icon_task_change_battery"
    OpsIcon.TaskMoveBike -> "icon_task_move_bike"
    OpsIcon.TaskInspection -> "icon_task_inspection"
    OpsIcon.TaskRepair -> "icon_task_repair"
    OpsIcon.FilterDrawerOpen -> "btn_filter_drawer_open"
    OpsIcon.FilterDrawerClose -> "btn_filter_drawer_close"
    OpsIcon.VehicleRepairV1 -> "vehicle_repair_version_one"
    OpsIcon.VehicleRepairV2 -> "vehicle_repair_version_two"
    OpsIcon.VehicleRepairV3 -> "vehicle_repair_version_three"
    OpsIcon.PhotoReplaceHolder -> "bg_photo_replace_holder"
    OpsIcon.PhotoDelete -> "close_red"
    OpsIcon.VehicleAddIc -> "vehicle_add_ic"
    OpsIcon.ChevronLeft -> "chevron_left"
    OpsIcon.CommonSwitch -> "common_switch_icon"
    OpsIcon.MapMore -> "btn_map_more"
    OpsIcon.MapMoreSelected -> "theme_map_more_selected"
    OpsIcon.MapLocation -> "btn_map_location"
    OpsIcon.MapLocationSelected -> "theme_map_location_selected"
    OpsIcon.MapExplain -> "btn_explain"
    OpsIcon.MapSatellite -> "theme_map_satellite_un_selected"
    OpsIcon.MapSatelliteSelected -> "theme_map_satellite_selected"
    OpsIcon.IconParking -> "icon_parking"
    OpsIcon.IconNoParking -> "icon_no_parking"
    OpsIcon.ArrowDownBlack -> "arrow_down_black"
    OpsIcon.ArrowUpBlue -> "arrow_up_blue"
    OpsIcon.CommonAble -> "common_able_ic"
    OpsIcon.CommonUnable -> "common_unable_ic"
    OpsIcon.CommonDelete -> "common_delect_ic"
    OpsIcon.BtnClose -> "btn_close"
    OpsIcon.FenceUndoAble -> "theme_fence_back_able"
    OpsIcon.FenceUndoDisable -> "theme_fence_back_disable"
    OpsIcon.SelectMapPoint -> "btn_select_map_point"
    OpsIcon.MapCenterPoint -> "icon_map_center_point"
    OpsIcon.FenceAngleAble -> "theme_fence_angle_able"
    OpsIcon.FenceAngleDisable -> "theme_fence_angle_un_disable"
    OpsIcon.FenceEditParams -> "theme_fence_edit"
    OpsIcon.FenceModifySize -> "theme_fence_modify_size"
    OpsIcon.FenceTabPatch -> "theme_fence_patch_selected"
    OpsIcon.FenceTabPoint -> "theme_fence_point_selected"
    OpsIcon.HomeRefresh -> "icon_refresh_home"
    OpsIcon.HomeDetail -> "icon_detail_home"
    OpsIcon.HomeDetailSelected -> "icon_detail_home_selected"
    OpsIcon.HomeFence -> "icon_fence_home"
    OpsIcon.HomeFenceSelected -> "icon_fence_home_selected"
    OpsIcon.HomeSwitch -> "icon_switch_home"
    OpsIcon.HomeSwitchSelected -> "icon_switch_home_selected"
    OpsIcon.TabMap -> "ic_map_unsel"
    OpsIcon.TabMapSelected -> "theme_home_map_selected"
    OpsIcon.TabTask -> "ic_task_unsel"
    OpsIcon.TabTaskSelected -> "theme_home_task_selected"
    OpsIcon.TabScan -> "ic_scan_special"
    OpsIcon.TabAnalysis -> "ic_analysis_unsel"
    OpsIcon.TabAnalysisSelected -> "theme_home_analysis_selected"
    OpsIcon.TabMine -> "ic_mine_unsel"
    OpsIcon.TabMineSelected -> "theme_home_me_selected"
    OpsIcon.DetectScan -> "ic_scan_unsel"
    OpsIcon.DetectSwitchAcc -> "vehicle_switch_ic"
    OpsIcon.DetectSwitchDefend -> "vehicle_defend_ic"
    OpsIcon.DetectSwitchBattery -> "vehicle_battery_ic"
    OpsIcon.DetectSwitchHelmet -> "vehicle_helmet_ic"
    OpsIcon.DetectSwitchWheel -> "vehicle_wheel_ic"
    OpsIcon.DetectLocation -> "vehicle_location_ic"
    OpsIcon.DetectOverload -> "vehicle_overload_check"
    OpsIcon.DetectArrow -> "vehicle_right_arrow"
    OpsIcon.DetectRing -> "vehicle_ring_ic"
    OpsIcon.DetectRefresh -> "vehicle_refresh_ic"
    OpsIcon.DetectReboot -> "vehicle_restart_ic"
    OpsIcon.BindClear -> "btn_bind_close"
}
