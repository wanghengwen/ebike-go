package com.luopingtech.ebike.rider.platform

import android.Manifest
import android.content.pm.PackageManager
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.platform.LocalContext
import androidx.core.content.ContextCompat
import com.luopingtech.ebike.rider.RiderApp
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.ui.permission.RiderLocationPermissionGate

@Composable
fun rememberLocationPermissionGate(app: RiderApp): RiderLocationPermissionGate {
    val context = LocalContext.current
    var pending by remember { mutableStateOf<((String?) -> Unit)?>(null) }

    val launcher = rememberLauncherForActivityResult(
        ActivityResultContracts.RequestMultiplePermissions(),
    ) { result ->
        val onResult = pending
        pending = null
        val androidTracker = app.locationTracker as? AndroidLocationTracker
        val ok = app.isDemoMode ||
            androidTracker == null ||
            androidTracker.hasPermission() ||
            result[Manifest.permission.ACCESS_FINE_LOCATION] == true ||
            result[Manifest.permission.ACCESS_COARSE_LOCATION] == true
        onResult?.invoke(if (ok) null else app.i18n.t(Str.LocationPermissionRequired))
    }

    return remember(app) {
        RiderLocationPermissionGate { onResult ->
            val androidTracker = app.locationTracker as? AndroidLocationTracker
            val need = androidTracker != null && !androidTracker.hasPermission() && !app.isDemoMode
            if (!need) {
                onResult(null)
            } else {
                pending = onResult
                launcher.launch(
                    arrayOf(
                        Manifest.permission.ACCESS_FINE_LOCATION,
                        Manifest.permission.ACCESS_COARSE_LOCATION,
                    ),
                )
            }
        }
    }
}
