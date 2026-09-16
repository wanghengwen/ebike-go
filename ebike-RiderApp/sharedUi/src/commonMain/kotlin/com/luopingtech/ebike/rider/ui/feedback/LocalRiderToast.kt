package com.luopingtech.ebike.rider.ui.feedback

import androidx.compose.runtime.staticCompositionLocalOf

val LocalRiderToast = staticCompositionLocalOf<(String) -> Unit> { { _ -> } }
