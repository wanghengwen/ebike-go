package com.luopingtech.ebike.rider.platform

import android.Manifest
import android.os.Build
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.platform.LocalContext
import com.luopingtech.ebike.rider.RiderApp
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.ui.permission.RiderBlePermissionGate

@Composable
fun rememberBlePermissionGate(app: RiderApp): RiderBlePermissionGate {
    val context = LocalContext.current
    var pending by remember { mutableStateOf<((String?) -> Unit)?>(null) }

    val launcher = rememberLauncherForActivityResult(
        ActivityResultContracts.RequestMultiplePermissions(),
    ) { result ->
        val onResult = pending
        pending = null
        val requested = blePermissionIds()
        val ok = app.isDemoMode || blePermissionsGranted(context) ||
            requested.all { result[it] == true }
        onResult?.invoke(if (ok) null else app.i18n.t(Str.BlePermissionRequired))
    }

    return remember(app) {
        RiderBlePermissionGate { onResult ->
            if (app.isDemoMode || blePermissionsGranted(context)) {
                onResult(null)
            } else {
                pending = onResult
                launcher.launch(blePermissionIds())
            }
        }
    }
}

private fun blePermissionIds(): Array<String> {
    return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
        arrayOf(
            Manifest.permission.BLUETOOTH_SCAN,
            Manifest.permission.BLUETOOTH_CONNECT,
        )
    } else {
        arrayOf(
            Manifest.permission.ACCESS_FINE_LOCATION,
            Manifest.permission.ACCESS_COARSE_LOCATION,
        )
    }
}

private fun blePermissionsGranted(context: android.content.Context): Boolean {
    val ids = blePermissionIds()
    return ids.all { id ->
        context.checkSelfPermission(id) == android.content.pm.PackageManager.PERMISSION_GRANTED
    } || (Build.VERSION.SDK_INT < Build.VERSION_CODES.S &&
        (context.checkSelfPermission(Manifest.permission.ACCESS_FINE_LOCATION) ==
            android.content.pm.PackageManager.PERMISSION_GRANTED ||
            context.checkSelfPermission(Manifest.permission.ACCESS_COARSE_LOCATION) ==
            android.content.pm.PackageManager.PERMISSION_GRANTED))
}
