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
import androidx.compose.runtime.LaunchedEffect
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
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.unit.IntOffset
import androidx.compose.ui.unit.IntSize
import androidx.compose.ui.unit.dp
import android.graphics.Paint as AndroidPaint
import android.graphics.Typeface
import kotlin.math.abs
import kotlin.math.atan2
import kotlin.math.cos
import kotlin.math.hypot
import kotlin.math.max
import kotlin.math.min
import kotlin.math.roundToInt
import kotlin.math.sin

/**
 * 对齐 FenceClipPatchView：
 * - 框钉屏幕；框外无命中 → 地图可拖（原版 return touchPos != -1）
 * - 平移：跟手，不强制夹在画布内
 * - 拉伸：对边角固定（resetLeft/Right/Top/Bottom）；旋转后按 checkHandleMove 重映射轴向
 * - 旋转：相对按下累加圆心角
 * - 顶/侧米标：widthPx/heightPx × metersPerPixel
 */
data class FencePatchRectState(
    val centerX: Float = Float.NaN,
    val centerY: Float = Float.NaN,
    val widthPx: Float = 200f,
    val heightPx: Float = 100f,
    val rotationDeg: Float = 0f,
) {
    val isReady: Boolean get() = !centerX.isNaN() && !centerY.isNaN()

    fun corners(): List<Pair<Float, Float>> {
        if (!isReady) return emptyList()
        val hw = widthPx / 2f
        val hh = heightPx / 2f
        val rad = Math.toRadians(rotationDeg.toDouble())
        val c = cos(rad).toFloat()
        val s = sin(rad).toFloat()
        fun rot(dx: Float, dy: Float): Pair<Float, Float> =
            (centerX + dx * c - dy * s) to (centerY + dx * s + dy * c)
        return listOf(rot(-hw, -hh), rot(hw, -hh), rot(hw, hh), rot(-hw, hh))
    }

    companion object {
        fun fromScreenCorners(corners: List<Pair<Float, Float>>): FencePatchRectState? {
            if (corners.size < 4) return null
            val cx = (corners[0].first + corners[1].first + corners[2].first + corners[3].first) / 4f
            val cy = (corners[0].second + corners[1].second + corners[2].second + corners[3].second) / 4f
            if (!cx.isFinite() || !cy.isFinite()) return null
            if (abs(cx) > 20000f || abs(cy) > 20000f) return null
            val lt = corners[0]
            val rb = corners[2]
            var w = abs(rb.first - lt.first)
            var h = abs(rb.second - lt.second)
            if (w < 8f || h < 8f) {
                w = corners.maxOf { it.first } - corners.minOf { it.first }
                h = corners.maxOf { it.second } - corners.minOf { it.second }
            }
            if (w < 8f || h < 8f || w > 20000f || h > 20000f) return null
            val rt = corners[1]
            val rot = Math.toDegrees(
                atan2((rt.second - lt.second).toDouble(), (rt.first - lt.first).toDouble()),
            ).toFloat()
            return FencePatchRectState(
                centerX = cx,
                centerY = cy,
                widthPx = w,
                heightPx = h,
                rotationDeg = if (rot.isFinite()) rot else 0f,
            )
        }
    }
}

private enum class PatchDragMode { Move, Resize, Rotate }

/** 对齐 checkHandleMove：把「右下手柄」映射到未旋转矩形的哪条边组合。 */
private enum class PatchResizeAxes { BottomRight, BottomLeft, TopLeft, TopRight }

private fun checkHandleMove(rotationDeg: Float): PatchResizeAxes {
    var r = rotationDeg % 360f
    if (r < 0) r += 360f
    return when {
        r > 45f && r <= 135f -> PatchResizeAxes.BottomLeft
        r > 135f && r <= 225f -> PatchResizeAxes.TopLeft
        r > 225f && r <= 315f -> PatchResizeAxes.TopRight
        else -> PatchResizeAxes.BottomRight
    }
}

/** 对边角固定拉伸（对齐 resetLeft/Right/Top/Bottom）。 */
private fun resizeKeepOpposite(
    rect: FencePatchRectState,
    deltaX: Float,
    deltaY: Float,
    minSide: Float,
): FencePatchRectState {
    val axes = checkHandleMove(rect.rotationDeg)
    // 屏幕增量转到矩形本地坐标
    val rad = Math.toRadians(-rect.rotationDeg.toDouble())
    val c = cos(rad).toFloat()
    val s = sin(rad).toFloat()
    val ldx = deltaX * c - deltaY * s
    val ldy = deltaX * s + deltaY * c
    var left = -rect.widthPx / 2f
    var right = rect.widthPx / 2f
    var top = -rect.heightPx / 2f
    var bottom = rect.heightPx / 2f
    when (axes) {
        PatchResizeAxes.BottomRight -> {
            right += ldx
            bottom += ldy
        }
        PatchResizeAxes.BottomLeft -> {
            // 原版 POS_BOTTOM_LEFT: resetBottom(-dx); resetLeft(-dy)
            left += -ldy
            bottom += -ldx
        }
        PatchResizeAxes.TopLeft -> {
            left += ldx
            top += ldy
        }
        PatchResizeAxes.TopRight -> {
            right += ldx
            top += ldy
        }
    }
    var w = right - left
    var h = bottom - top
    if (w < minSide) {
        if (axes == PatchResizeAxes.BottomLeft || axes == PatchResizeAxes.TopLeft) {
            left = right - minSide
        } else {
            right = left + minSide
        }
        w = minSide
    }
    if (h < minSide) {
        if (axes == PatchResizeAxes.TopLeft || axes == PatchResizeAxes.TopRight) {
            top = bottom - minSide
        } else {
            bottom = top + minSide
        }
        h = minSide
    }
    val localCx = (left + right) / 2f
    val localCy = (top + bottom) / 2f
    val rRad = Math.toRadians(rect.rotationDeg.toDouble())
    val rc = cos(rRad).toFloat()
    val rs = sin(rRad).toFloat()
    val worldCx = rect.centerX + localCx * rc - localCy * rs
    val worldCy = rect.centerY + localCx * rs + localCy * rc
    return rect.copy(centerX = worldCx, centerY = worldCy, widthPx = w, heightPx = h)
}

@Composable
fun FencePatchOverlay(
    isParking: Boolean,
    onRectChanged: (FencePatchRectState) -> Unit,
    modifier: Modifier = Modifier,
    forcedSizePx: Pair<Float, Float>? = null,
    seedRect: FencePatchRectState? = null,
    seedNonce: Int = 0,
    allowDefaultInit: Boolean = true,
    /** 当前缩放下米/像素；用于米标与调尺寸。 */
    metersPerPixel: Double = 0.55,
) {
    val density = LocalDensity.current
    var canvasSize by remember { mutableStateOf(IntSize.Zero) }
    var rect by remember { mutableStateOf(FencePatchRectState()) }
    var initialized by remember { mutableStateOf(false) }
    var appliedSeedNonce by remember { mutableIntStateOf(0) }
    var dragging by remember { mutableStateOf(false) }
    val handlePx = with(density) { 30.dp.toPx() }
    val minSidePx = with(density) { 40.dp.toPx() }
    val fillColor = if (isParking) Color(0x80242F57) else Color(0x80E02020)
    val strokeColor = if (isParking) Color(0xFF242F57) else Color(0xFFFF0808)
    val metreColor = Color(0xFF666666)

    LaunchedEffect(forcedSizePx) {
        val size = forcedSizePx ?: return@LaunchedEffect
        if (!initialized) return@LaunchedEffect
        rect = rect.copy(
            widthPx = max(minSidePx, size.first),
            heightPx = max(minSidePx, size.second),
        )
        onRectChanged(rect)
    }

    LaunchedEffect(seedNonce, seedRect) {
        val seed = seedRect ?: return@LaunchedEffect
        if (seedNonce <= 0 || seedNonce == appliedSeedNonce) return@LaunchedEffect
        if (!seed.isReady) return@LaunchedEffect
        appliedSeedNonce = seedNonce
        rect = if (initialized && rect.isReady) {
            rect.copy(
                widthPx = max(minSidePx, seed.widthPx),
                heightPx = max(minSidePx, seed.heightPx),
                rotationDeg = seed.rotationDeg,
                centerX = seed.centerX,
                centerY = seed.centerY,
            )
        } else {
            seed
        }
        initialized = true
        onRectChanged(rect)
    }

    val rectUpdated by rememberUpdatedState(rect)
    val onChangedUpdated by rememberUpdatedState(onRectChanged)
    val mppUpdated by rememberUpdatedState(metersPerPixel)

    Box(
        modifier = modifier
            .fillMaxSize()
            .onSizeChanged { size: IntSize ->
                canvasSize = size
                if (!initialized && allowDefaultInit && size.width > 0 && size.height > 0 && seedNonce <= 0) {
                    initialized = true
                    rect = FencePatchRectState(
                        centerX = size.width / 2f,
                        centerY = size.height / 2f,
                        widthPx = min(200f, size.width * 0.8f),
                        heightPx = min(100f, size.height * 0.6f),
                    )
                    onRectChanged(rect)
                }
            },
    ) {
        Canvas(modifier = Modifier.fillMaxSize()) {
            val corners = rect.corners()
            if (corners.size < 4) return@Canvas
            val path = Path().apply {
                moveTo(corners[0].first, corners[0].second)
                lineTo(corners[1].first, corners[1].second)
                lineTo(corners[2].first, corners[2].second)
                lineTo(corners[3].first, corners[3].second)
                close()
            }
            drawPath(path, color = fillColor)
            drawPath(path, color = strokeColor, style = Stroke(width = 6f))

            val mpp = mppUpdated.coerceAtLeast(1e-6)
            val widthM = max(10, (rect.widthPx * mpp).roundToInt())
            val heightM = max(10, (rect.heightPx * mpp).roundToInt())
            val paint = AndroidPaint().apply {
                color = android.graphics.Color.parseColor("#666666")
                textSize = 28f
                isAntiAlias = true
                typeface = Typeface.DEFAULT_BOLD
                textAlign = AndroidPaint.Align.CENTER
            }
            val nc = drawContext.canvas.nativeCanvas
            // 顶边中点宽、左边中点高（对齐原版画米）
            val topMid = Offset(
                (corners[0].first + corners[1].first) / 2f,
                (corners[0].second + corners[1].second) / 2f,
            )
            val leftMid = Offset(
                (corners[0].first + corners[3].first) / 2f,
                (corners[0].second + corners[3].second) / 2f,
            )
            nc.drawText("${widthM}m", topMid.x, topMid.y - 12f, paint)
            nc.save()
            nc.rotate(-90f, leftMid.x, leftMid.y)
            nc.drawText("${heightM}m", leftMid.x, leftMid.y - 12f, paint)
            nc.restore()

            // 拖动中隐藏手柄（对齐 isMove）
            if (!dragging) {
                val br = Offset(corners[2].first, corners[2].second)
                val tl = Offset(corners[0].first, corners[0].second)
                drawCircle(color = Color.White, radius = handlePx / 2f, center = br)
                drawCircle(color = strokeColor, radius = handlePx / 2f, center = br, style = Stroke(3f))
                drawCircle(color = Color.White, radius = handlePx / 2f, center = tl)
                drawCircle(color = strokeColor, radius = handlePx / 2f, center = tl, style = Stroke(3f))
            }
        }

        if (rect.isReady) {
            val corners = rect.corners()
            val br = corners[2]
            val tl = corners[0]
            val bodyMinX = corners.minOf { it.first }
            val bodyMaxX = corners.maxOf { it.first }
            val bodyMinY = corners.minOf { it.second }
            val bodyMaxY = corners.maxOf { it.second }
            val bodyW = (bodyMaxX - bodyMinX).coerceAtLeast(1f)
            val bodyH = (bodyMaxY - bodyMinY).coerceAtLeast(1f)
            val bodyMinXUpdated by rememberUpdatedState(bodyMinX)
            val bodyMinYUpdated by rememberUpdatedState(bodyMinY)
            val brUpdated by rememberUpdatedState(br)
            val tlUpdated by rememberUpdatedState(tl)

            Box(
                modifier = Modifier
                    .offset { IntOffset(bodyMinX.roundToInt(), bodyMinY.roundToInt()) }
                    .size(
                        width = with(density) { bodyW.toDp() },
                        height = with(density) { bodyH.toDp() },
                    )
                    .pointerInput(minSidePx) {
                        awaitEachGesture {
                            val down = awaitFirstDown(requireUnconsumed = false)
                            val start = rectUpdated
                            if (!start.isReady) return@awaitEachGesture
                            val ox = down.position.x + bodyMinXUpdated
                            val oy = down.position.y + bodyMinYUpdated
                            val brNow = brUpdated
                            val tlNow = tlUpdated
                            if (hypot((ox - brNow.first).toDouble(), (oy - brNow.second).toDouble()) <= handlePx ||
                                hypot((ox - tlNow.first).toDouble(), (oy - tlNow.second).toDouble()) <= handlePx
                            ) {
                                return@awaitEachGesture
                            }
                            if (!pointInRotatedRect(ox, oy, start)) return@awaitEachGesture
                            down.consume()
                            dragging = true
                            var current = start
                            drag(down.id) { change ->
                                val amount = change.positionChange()
                                change.consume()
                                // 原版平移不夹画布
                                current = current.copy(
                                    centerX = current.centerX + amount.x,
                                    centerY = current.centerY + amount.y,
                                )
                                rect = current
                                onChangedUpdated(current)
                            }
                            dragging = false
                        }
                    },
            )

            PatchHandle(
                centerX = br.first,
                centerY = br.second,
                handlePx = handlePx,
                density = density,
                mode = PatchDragMode.Resize,
                minSidePx = minSidePx,
                rectUpdated = { rectUpdated },
                onDragging = { dragging = it },
                onRect = { next ->
                    rect = next
                    onChangedUpdated(next)
                },
            )
            PatchHandle(
                centerX = tl.first,
                centerY = tl.second,
                handlePx = handlePx,
                density = density,
                mode = PatchDragMode.Rotate,
                minSidePx = minSidePx,
                rectUpdated = { rectUpdated },
                onDragging = { dragging = it },
                onRect = { next ->
                    rect = next
                    onChangedUpdated(next)
                },
            )
        }
    }
}

@Composable
private fun PatchHandle(
    centerX: Float,
    centerY: Float,
    handlePx: Float,
    density: androidx.compose.ui.unit.Density,
    mode: PatchDragMode,
    minSidePx: Float,
    rectUpdated: () -> FencePatchRectState,
    onDragging: (Boolean) -> Unit,
    onRect: (FencePatchRectState) -> Unit,
) {
    val centerXUpdated by rememberUpdatedState(centerX)
    val centerYUpdated by rememberUpdatedState(centerY)
    Box(
        modifier = Modifier
            .offset {
                IntOffset(
                    (centerX - handlePx).roundToInt(),
                    (centerY - handlePx).roundToInt(),
                )
            }
            .size(with(density) { (handlePx * 2).toDp() })
            .pointerInput(mode, minSidePx) {
                awaitEachGesture {
                    val down = awaitFirstDown(requireUnconsumed = false)
                    down.consume()
                    var current = rectUpdated()
                    if (!current.isReady) return@awaitEachGesture
                    onDragging(true)
                    var fingerX = centerXUpdated
                    var fingerY = centerYUpdated
                    val startAngle = atan2(
                        (fingerY - current.centerY).toDouble(),
                        (fingerX - current.centerX).toDouble(),
                    )
                    val baseRot = current.rotationDeg
                    drag(down.id) { change ->
                        val amount = change.positionChange()
                        change.consume()
                        fingerX += amount.x
                        fingerY += amount.y
                        current = when (mode) {
                            PatchDragMode.Resize -> resizeKeepOpposite(
                                current,
                                amount.x,
                                amount.y,
                                minSidePx,
                            )
                            PatchDragMode.Rotate -> {
                                // 对齐 angle(cen, first, second)：相对按下累加
                                val now = atan2(
                                    (fingerY - current.centerY).toDouble(),
                                    (fingerX - current.centerX).toDouble(),
                                )
                                val deltaDeg = Math.toDegrees(now - startAngle).toFloat()
                                current.copy(rotationDeg = (baseRot + deltaDeg) % 360f)
                            }
                            PatchDragMode.Move -> current
                        }
                        onRect(current)
                    }
                    onDragging(false)
                }
            },
    )
}

private fun pointInRotatedRect(px: Float, py: Float, rect: FencePatchRectState): Boolean {
    if (!rect.isReady) return false
    val rad = Math.toRadians(-rect.rotationDeg.toDouble())
    val c = cos(rad).toFloat()
    val s = sin(rad).toFloat()
    val dx = px - rect.centerX
    val dy = py - rect.centerY
    val lx = dx * c - dy * s
    val ly = dx * s + dy * c
    return abs(lx) <= rect.widthPx / 2f && abs(ly) <= rect.heightPx / 2f
}
