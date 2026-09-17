package com.luopingtech.ebike.ops.ui.fence

import androidx.compose.foundation.Canvas
import androidx.compose.foundation.gestures.awaitEachGesture
import androidx.compose.foundation.gestures.awaitFirstDown
import androidx.compose.foundation.gestures.drag
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.size
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.graphics.nativeCanvas
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.input.pointer.positionChange
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.unit.IntOffset
import androidx.compose.ui.unit.dp
import android.graphics.Paint as AndroidPaint
import android.graphics.Typeface
import com.luopingtech.ebike.ops.core.time.nowEpochMillis
import kotlin.math.roundToInt

/**
 * 对齐 FenceClipSelectPointView：
 * - 顶点热区 ±16dp；>300ms 才拖
 * - 框外不拦截地图
 * - 白字序号 + 边长 Xm
 */
@Composable
fun FenceSelectPointOverlay(
    isParking: Boolean,
    screenPoints: List<Pair<Float, Float>>,
    /** 与边一一对应：边 i 为点 i → 点 (i+1)%n 的米数；可空则不画 */
    edgeMeters: List<Int> = emptyList(),
    onMovePointScreen: (index: Int, screenX: Float, screenY: Float) -> Unit,
    onMoveFinished: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val density = LocalDensity.current
    val radiusPx = with(density) { 8.dp.toPx() }
    val hitRadiusPx = with(density) { 16.dp.toPx() }
    val fillColor = if (isParking) Color(0x80242F57) else Color(0x8066E02020)
    val borderColor = if (isParking) Color(0xFF242F57) else Color(0xFFFF3366)
    val textBg = Color(0xFF242F57)

    var draggingIndex by remember { mutableIntStateOf(-1) }
    var dragPos by remember { mutableStateOf<Pair<Float, Float>?>(null) }
    val moveUpdated by rememberUpdatedState(onMovePointScreen)
    val finishedUpdated by rememberUpdatedState(onMoveFinished)
    val edgesUpdated by rememberUpdatedState(edgeMeters)

    val display = screenPoints.mapIndexed { i, p ->
        if (i == draggingIndex) dragPos ?: p else p
    }

    Box(modifier = modifier.fillMaxSize()) {
        Canvas(modifier = Modifier.fillMaxSize()) {
            if (display.size >= 2) {
                val path = Path().apply {
                    moveTo(display[0].first, display[0].second)
                    for (i in 1 until display.size) {
                        lineTo(display[i].first, display[i].second)
                    }
                    if (display.size >= 3) close()
                }
                drawPath(path, color = fillColor)
                drawPath(path, color = borderColor, style = Stroke(width = 4f))
            }
            val labelPaint = AndroidPaint().apply {
                color = android.graphics.Color.WHITE
                textSize = radiusPx * 1.2f
                isAntiAlias = true
                typeface = Typeface.DEFAULT_BOLD
                textAlign = AndroidPaint.Align.CENTER
            }
            val metrePaint = AndroidPaint().apply {
                color = android.graphics.Color.parseColor("#242F57")
                textSize = radiusPx * 1.1f
                isAntiAlias = true
                textAlign = AndroidPaint.Align.CENTER
            }
            val nc = drawContext.canvas.nativeCanvas
            val edges = edgesUpdated
            if (display.size >= 2) {
                val n = display.size
                for (i in 0 until n) {
                    if (n < 3 && i == n - 1) break
                    val j = (i + 1) % n
                    if (n < 3 && j == 0) break
                    val ax = display[i].first
                    val ay = display[i].second
                    val bx = display[j].first
                    val by = display[j].second
                    val mx = (ax + bx) / 2f
                    val my = (ay + by) / 2f
                    val metres = edges.getOrNull(i)
                    if (metres != null && metres >= 0) {
                        nc.drawText("${metres}m", mx, my - 4f, metrePaint)
                    }
                }
            }
            display.forEachIndexed { index, (x, y) ->
                val c = Offset(x, y)
                drawCircle(color = textBg, radius = radiusPx, center = c)
                val fm = labelPaint.fontMetrics
                val baseline = y - (fm.ascent + fm.descent) / 2f
                nc.drawText("${index + 1}", x, baseline, labelPaint)
            }
        }

        screenPoints.forEachIndexed { index, (sx, sy) ->
            val centerXUpdated by rememberUpdatedState(sx)
            val centerYUpdated by rememberUpdatedState(sy)
            Box(
                modifier = Modifier
                    .offset {
                        IntOffset(
                            (sx - hitRadiusPx).roundToInt(),
                            (sy - hitRadiusPx).roundToInt(),
                        )
                    }
                    .size(with(density) { (hitRadiusPx * 2).toDp() })
                    .pointerInput(index) {
                        awaitEachGesture {
                            val down = awaitFirstDown(requireUnconsumed = false)
                            down.consume()
                            val downTime = nowEpochMillis()
                            var absX = centerXUpdated
                            var absY = centerYUpdated
                            var armed = false
                            drag(down.id) { change ->
                                val delta = change.positionChange()
                                absX += delta.x
                                absY += delta.y
                                if (nowEpochMillis() - downTime <= 300L) return@drag
                                change.consume()
                                if (!armed) {
                                    armed = true
                                    draggingIndex = index
                                }
                                dragPos = absX to absY
                                moveUpdated(index, absX, absY)
                            }
                            draggingIndex = -1
                            dragPos = null
                            if (armed) finishedUpdated()
                        }
                    },
            )
        }
    }
}
