package com.luopingtech.ebike.rider.ui.permission

import androidx.compose.runtime.staticCompositionLocalOf

val LocalRiderLocationPermissionGate =
    staticCompositionLocalOf<RiderLocationPermissionGate> { NoOpLocationPermissionGate }
