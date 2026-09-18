package com.luopingtech.ebike.ops.platform

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent

/**
 * 原版无开机自启保活服务。轨迹上报在下次打开 App、MainShell 登录后由
 * TrackUploadFeature 重新开启；退后台时再由 MainActivity 拉起 FGS。
 */
class TrackBootReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent?) {
        // 对齐原版：不做开机拉起 FGS。
    }
}
