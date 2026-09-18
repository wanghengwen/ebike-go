package com.luopingtech.ebike.ops.platform

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.Service
import android.content.Context
import android.content.Intent
import android.content.pm.ServiceInfo
import android.os.Build
import android.os.IBinder
import androidx.core.app.NotificationCompat
import com.luopingtech.ebike.ops.R

/**
 * 对齐原版 [com.xyytech.lite.service.APPKeepService]：
 * - 仅在 App 退到后台时拉起，回前台即停（见 [MainActivity]）
 * - 通知无标题/正文，渠道名「服务常驻通知」
 * - 自身不采点；采点/上报仍由 [AndroidLocationTracker] + TrackUploadFeature 负责
 *
 * 现代系统要求声明 FGS type；因后台仍要收定位，使用 location 类型（原版未声明 type）。
 */
class TrackLocationService : Service() {
    override fun onBind(intent: Intent?): IBinder? = null

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        when (intent?.action) {
            ACTION_STOP -> {
                stopForeground(STOP_FOREGROUND_REMOVE)
                stopSelf()
                return START_NOT_STICKY
            }
            else -> promoteToForeground()
        }
        return START_STICKY
    }

    private fun promoteToForeground() {
        ensureChannel()
        val notification = buildNotification()
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            startForeground(
                NOTIFICATION_ID,
                notification,
                ServiceInfo.FOREGROUND_SERVICE_TYPE_LOCATION,
            )
        } else {
            @Suppress("DEPRECATION")
            startForeground(NOTIFICATION_ID, notification)
        }
    }

    private fun ensureChannel() {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.O) return
        val manager = getSystemService(NotificationManager::class.java) ?: return
        if (manager.getNotificationChannel(CHANNEL_ID) != null) return
        val channel = NotificationChannel(
            CHANNEL_ID,
            CHANNEL_NAME,
            NotificationManager.IMPORTANCE_HIGH,
        ).apply {
            setLockscreenVisibility(Notification.VISIBILITY_PUBLIC)
            setShowBadge(false)
        }
        manager.createNotificationChannel(channel)
    }

    /** 对齐 APPKeepService.getNotification：不设 title/text，仅常驻。 */
    private fun buildNotification(): Notification {
        return NotificationCompat.Builder(this, CHANNEL_ID)
            .setSmallIcon(R.drawable.qit)
            .setVisibility(NotificationCompat.VISIBILITY_PUBLIC)
            .setOngoing(true)
            .setCategory(Notification.CATEGORY_SERVICE)
            .setPriority(NotificationCompat.PRIORITY_MAX)
            .setSilent(true)
            .build()
    }

    companion object {
        /** 对齐原版 channelId=Service_Id / channelName=服务常驻通知 */
        const val CHANNEL_ID = "Service_Id"
        const val CHANNEL_NAME = "服务常驻通知"
        const val NOTIFICATION_ID = 1001
        const val ACTION_STOP = "com.luopingtech.ebike.ops.TRACK_FG_STOP"

        fun start(context: Context) {
            val intent = Intent(context, TrackLocationService::class.java)
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                context.startForegroundService(intent)
            } else {
                context.startService(intent)
            }
        }

        fun stop(context: Context) {
            context.stopService(Intent(context, TrackLocationService::class.java))
        }
    }
}
