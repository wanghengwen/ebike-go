package com.luopingtech.ebike.ops.platform

import android.Manifest
import android.content.pm.PackageManager
import android.os.Build
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.platform.LocalContext
import androidx.core.content.ContextCompat
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.ui.permission.OpsTrackPermissionGate

/**
 * Android 侧的轨迹权限流程：先要前台定位（外加 13+ 的通知权限，前台服务要挂通知），
 * 拿到之后再顺手引导「始终允许」——后台定位是分次弹的，一起弹系统会直接拒。
 *
 * 共享层只看得到 [OpsTrackPermissionGate]：null 表示可以开始上报，字符串是给用户看的提示。
 */
@Composable
internal fun rememberTrackPermissionGate(app: OpsApp): OpsTrackPermissionGate {
    val context = LocalContext.current
    var pending by remember { mutableStateOf<((String?) -> Unit)?>(null) }

    val backgroundLocationLauncher = rememberLauncherForActivityResult(
        ActivityResultContracts.RequestPermission(),
    ) { /* optional; FGS still works with foreground-only */ }

    val locationPermissionLauncher = rememberLauncherForActivityResult(
        ActivityResultContracts.RequestMultiplePermissions(),
    ) { result ->
        val onResult = pending
        pending = null
        val androidTracker = app.locationTracker as? AndroidLocationTracker
        val locationOk = app.isDemoMode ||
            androidTracker == null ||
            androidTracker.hasPermission() ||
            result[Manifest.permission.ACCESS_FINE_LOCATION] == true ||
            result[Manifest.permission.ACCESS_COARSE_LOCATION] == true
        if (locationOk) {
            onResult?.invoke(null)
            // Android 10+：进程被杀/退到后台还要接着跑，得再要一次「始终允许」。
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q &&
                !app.isDemoMode &&
                ContextCompat.checkSelfPermission(
                    context,
                    Manifest.permission.ACCESS_BACKGROUND_LOCATION,
                ) != PackageManager.PERMISSION_GRANTED
            ) {
                backgroundLocationLauncher.launch(Manifest.permission.ACCESS_BACKGROUND_LOCATION)
            }
        } else {
            onResult?.invoke(app.i18n.t(Str.TrackNeedLocation))
        }
    }

    return remember(app, context) {
        OpsTrackPermissionGate { onResult ->
            val androidTracker = app.locationTracker as? AndroidLocationTracker
            val needLocation = androidTracker != null &&
                !androidTracker.hasPermission() &&
                !app.isDemoMode
            val needNotify = Build.VERSION.SDK_INT >= 33 &&
                ContextCompat.checkSelfPermission(
                    context,
                    Manifest.permission.POST_NOTIFICATIONS,
                ) != PackageManager.PERMISSION_GRANTED
            when {
                app.isDemoMode || androidTracker == null || (!needLocation && !needNotify) ->
                    onResult(null)
                else -> {
                    pending = onResult
                    val perms = buildList {
                        if (needLocation) {
                            add(Manifest.permission.ACCESS_FINE_LOCATION)
                            add(Manifest.permission.ACCESS_COARSE_LOCATION)
                        }
                        if (needNotify) add(Manifest.permission.POST_NOTIFICATIONS)
                    }.toTypedArray()
                    locationPermissionLauncher.launch(perms)
                }
            }
        }
    }
}
