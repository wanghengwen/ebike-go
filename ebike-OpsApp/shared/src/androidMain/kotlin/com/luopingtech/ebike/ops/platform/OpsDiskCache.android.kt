package com.luopingtech.ebike.ops.platform

import android.content.Context

private var appContext: Context? = null

/** 由 Android 宿主在 Application.onCreate 绑定。 */
fun bindOpsDiskCacheContext(context: Context) {
    appContext = context.applicationContext
}

actual fun clearOpsDiskCache(): Boolean {
    val ctx = appContext ?: return true
    return runCatching {
        ctx.cacheDir.deleteRecursively()
        ctx.externalCacheDir?.deleteRecursively()
        true
    }.getOrDefault(false)
}
