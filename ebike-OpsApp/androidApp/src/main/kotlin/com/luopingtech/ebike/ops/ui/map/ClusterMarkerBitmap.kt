package com.luopingtech.ebike.ops.ui.map

import android.graphics.Bitmap
import android.graphics.Canvas
import android.graphics.Color
import android.graphics.Paint
import android.graphics.Typeface
import android.util.LruCache
import android.util.TypedValue
import android.content.res.Resources

/**
 * Legacy-aligned cluster marker: theme-colored circle with vehicle count text
 * ([DefaultOptionGenerator.getNumCluster] / [TMapClusterRenderer]).
 */
object ClusterMarkerBitmap {
    private val cache = LruCache<String, Bitmap>(128)

    fun obtain(count: Int, fillColorArgb: Int, density: Float = Resources.getSystem().displayMetrics.density): Bitmap {
        val key = "$count@$fillColorArgb@${density.toInt()}"
        cache.get(key)?.let { return it }
        val bitmap = create(count, fillColorArgb, density)
        cache.put(key, bitmap)
        return bitmap
    }

    private fun create(count: Int, fillColorArgb: Int, density: Float): Bitmap {
        val text = formatCount(count)
        val sizePx = (20f * density).toInt().coerceAtLeast(20)
        val bitmap = Bitmap.createBitmap(sizePx, sizePx, Bitmap.Config.ARGB_8888)
        val canvas = Canvas(bitmap)

        val fill = Paint(Paint.ANTI_ALIAS_FLAG).apply {
            style = Paint.Style.FILL
            color = fillColorArgb
        }
        val cx = sizePx / 2f
        canvas.drawCircle(cx, cx, sizePx / 2f, fill)

        val textPaint = Paint(Paint.ANTI_ALIAS_FLAG).apply {
            color = Color.WHITE
            typeface = Typeface.DEFAULT_BOLD
            textAlign = Paint.Align.CENTER
            // Legacy DefaultOptionGenerator: always 15sp bold.
            textSize = TypedValue.applyDimension(
                TypedValue.COMPLEX_UNIT_SP,
                15f,
                Resources.getSystem().displayMetrics,
            )
        }
        val y = cx - (textPaint.descent() + textPaint.ascent()) / 2f
        canvas.drawText(text, cx, y, textPaint)
        return bitmap
    }

    /** Matches legacy DefaultOptionGenerator.getCountText (non-vague). */
    fun formatCount(count: Int): String {
        val n = count.coerceAtLeast(0)
        return if (n > 10) "$n" else " $n "
    }
}
