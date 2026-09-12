package com.luopingtech.ebike.ops.platform

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import com.luopingtech.ebike.ops.OpsApplication

/**
 * After reboot, if track upload was left on, bring up the location FGS.
 * Collect/upload resumes when MainShell re-enables [TrackUploadFeature].
 */
class TrackBootReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent?) {
        if (intent?.action != Intent.ACTION_BOOT_COMPLETED) return
        val app = context.applicationContext as? OpsApplication ?: return
        if (!app.isOpsAppReady) return
        if (!app.opsApp.trackUploadFeature.wasEnabledPersisted()) return
        if (app.opsApp.isDemoMode) return
        val tracker = app.opsApp.locationTracker as? AndroidLocationTracker ?: return
        if (!tracker.hasPermission()) return
        TrackLocationService.start(context.applicationContext)
    }
}
