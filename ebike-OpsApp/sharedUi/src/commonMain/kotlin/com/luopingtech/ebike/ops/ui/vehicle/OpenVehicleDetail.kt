package com.luopingtech.ebike.ops.ui.vehicle

import androidx.compose.runtime.staticCompositionLocalOf
import com.luopingtech.ebike.ops.domain.model.Vehicle

/**
 * 打开遗留 CarDetailActivity 对齐的整页车辆详情。
 * Android 宿主在 [com.luopingtech.ebike.ops.MainActivity] 注入并叠加 [VehicleDetailScreen]；
 * 未注入时为 no-op（iOS 等可后续补齐）。
 */
fun interface OpenVehicleDetail {
    fun open(vehicle: Vehicle, serviceAreaId: String?)
}

val LocalOpenVehicleDetail = staticCompositionLocalOf<OpenVehicleDetail> {
    OpenVehicleDetail { _, _ -> }
}
