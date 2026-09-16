package com.luopingtech.ebike.rider.ui.permission

import androidx.compose.runtime.staticCompositionLocalOf

val LocalRiderBlePermissionGate =
    staticCompositionLocalOf<RiderBlePermissionGate> { NoOpBlePermissionGate }
