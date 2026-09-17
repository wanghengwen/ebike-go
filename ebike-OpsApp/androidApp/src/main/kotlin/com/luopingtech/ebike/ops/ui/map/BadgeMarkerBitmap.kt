package com.luopingtech.ebike.ops.ui.map

import android.content.Context
import android.graphics.Bitmap
import android.graphics.BitmapFactory
import android.graphics.Canvas
import android.util.LruCache

/**
 * Legacy MapOptionProvide.combineBitmap：车标底图 + tips_* 角标叠右下。
 */
object BadgeMarkerBitmap {
    private val cache = object : LruCache<String, Bitmap>(64) {
        override fun sizeOf(key: String, value: Bitmap): Int = value.byteCount / 1024
    }

    fun obtain(context: Context, vehicleResId: Int, badgeResId: Int): Bitmap? {
        if (vehicleResId == 0 || badgeResId == 0) return null
        val key = "$vehicleResId|$badgeResId"
        cache.get(key)?.let { return it }
        return runCatching {
            val bg = BitmapFactory.decodeResource(context.resources, vehicleResId) ?: return null
            val fg = BitmapFactory.decodeResource(context.resources, badgeResId) ?: return null
            val out = Bitmap.createBitmap(bg.width, bg.height, Bitmap.Config.ARGB_8888)
            val canvas = Canvas(out)
            canvas.drawBitmap(bg, 0f, 0f, null)
            val left = (bg.width - fg.width).toFloat()
            val top = (bg.height - fg.height * 1.5f)
            canvas.drawBitmap(fg, left, top, null)
            cache.put(key, out)
            out
        }.getOrNull()
    }

    fun drawableId(context: Context, name: String?): Int {
        if (name.isNullOrBlank()) return 0
        return context.resources.getIdentifier(name, "drawable", context.packageName)
    }
}
