package com.luopingtech.ebike.ops.ui.map

import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.graphics.nativeCanvas
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.IntSize
import androidx.compose.ui.unit.dp
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.compose.runtime.rememberUpdatedState
import android.graphics.Paint as AndroidPaint
import android.graphics.Typeface
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.domain.map.LatLngBounds
import com.luopingtech.ebike.ops.domain.map.MapClusterer
import com.luopingtech.ebike.ops.domain.map.MapProjection
import com.luopingtech.ebike.ops.domain.model.MapPin
import com.luopingtech.ebike.ops.domain.model.TrackPoint
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import kotlin.math.pow
/**
 * Open-source map surface: projects [MapPin] lat/lng onto a Canvas.
 * Used when [MapProviderKind.SIMULATOR] or vendor SDK is not ready.
 */
@Composable
fun SimulatorMapView(
    pins: List<MapPin>,
    selectedCarId: String?,
    providerLabel: String,
    onSelectCarId: (String) -> Unit,
    onSelectCluster: (List<String>) -> Unit = {},
    onMapTap: ((lat: Double, lng: Double) -> Unit)? = null,
    onCameraIdle: ((lat: Double, lng: Double) -> Unit)? = null,
    onCameraMove: ((lat: Double, lng: Double) -> Unit)? = null,
    trackPoints: List<TrackPoint> = emptyList(),
    modifier: Modifier = Modifier,
    clusterOverview: Boolean = true,
    followNonce: Int = 0,
    followLat: Double? = null,
    followLng: Double? = null,
    @Suppress("UNUSED_PARAMETER") followZoom: Float = 17f,
    screenToLatLngNonce: Int = 0,
    screenPickX: Float = 0f,
    screenPickY: Float = 0f,
    onScreenToLatLng: ((lat: Double, lng: Double) -> Unit)? = null,
    batchScreenToLatLngNonce: Int = 0,
    batchScreenPoints: List<Pair<Float, Float>> = emptyList(),
    onBatchScreenToLatLng: ((List<Pair<Double, Double>>) -> Unit)? = null,
    latLngToScreenNonce: Int = 0,
    latLngToScreenPoints: List<Pair<Double, Double>> = emptyList(),
    onLatLngToScreen: ((List<Pair<Float, Float>>) -> Unit)? = null,
    @Suppress("UNUSED_PARAMETER") autoFitOnPins: Boolean = true,
) {
    val scheme = MaterialTheme.colorScheme
    val clusterColor = OpsTheme.colors.primary
    var size by remember { mutableStateOf(IntSize.Zero) }
    var simZoom by remember { mutableStateOf(12f) }
    var focusLat by remember { mutableStateOf<Double?>(null) }
    var focusLng by remember { mutableStateOf<Double?>(null) }
    val density = LocalDensity.current
    val fullBounds = remember(pins) {
        MapProjection.boundsOf(pins)
    }
    val markers = remember(pins, clusterOverview, simZoom) {
        if (!clusterOverview) {
            pins
        } else {
            val lat = pins.firstOrNull { it.lat != 0.0 || it.lng != 0.0 }?.lat ?: 30.0
            MapClusterer.cluster(pins, MapClusterer.cellDegreesForZoom(simZoom, lat))
        }
    }
    val bounds = remember(fullBounds, simZoom, focusLat, focusLng) {
        val factor = 2.0.pow((12.0 - simZoom).toDouble()).coerceIn(0.05, 4.0)
        val midLat = focusLat ?: ((fullBounds.minLat + fullBounds.maxLat) / 2.0)
        val midLng = focusLng ?: ((fullBounds.minLng + fullBounds.maxLng) / 2.0)
        val halfLat = fullBounds.latSpan / 2.0 * factor
        val halfLng = fullBounds.lngSpan / 2.0 * factor
        LatLngBounds(
            minLat = midLat - halfLat,
            maxLat = midLat + halfLat,
            minLng = midLng - halfLng,
            maxLng = midLng + halfLng,
        )
    }
    val projected = remember(markers, size, bounds) {
        if (size.width == 0 || size.height == 0) {
            emptyList()
        } else {
            markers.map { pin ->
                pin to MapProjection.project(
                    lat = pin.lat,
                    lng = pin.lng,
                    bounds = bounds,
                    width = size.width.toFloat(),
                    height = size.height.toFloat(),
                )
            }
        }
    }
    val cameraIdleUpdated by rememberUpdatedState(onCameraIdle)
    val cameraMoveUpdated by rememberUpdatedState(onCameraMove)
    LaunchedEffect(bounds, cameraIdleUpdated, cameraMoveUpdated) {
        val midLat = (bounds.minLat + bounds.maxLat) / 2.0
        val midLng = (bounds.minLng + bounds.maxLng) / 2.0
        cameraMoveUpdated?.invoke(midLat, midLng)
        cameraIdleUpdated?.invoke(midLat, midLng)
    }
    LaunchedEffect(followNonce, followLat, followLng) {
        if (followNonce > 0 && followLat != null && followLng != null) {
            focusLat = followLat
            focusLng = followLng
        }
    }
    val screenToLatLngUpdated by rememberUpdatedState(onScreenToLatLng)
    LaunchedEffect(screenToLatLngNonce, screenPickX, screenPickY, bounds, size) {
        if (screenToLatLngNonce <= 0 || size.width <= 0 || size.height <= 0) return@LaunchedEffect
        val cb = screenToLatLngUpdated ?: return@LaunchedEffect
        val (lat, lng) = MapProjection.unproject(
            x = screenPickX,
            y = screenPickY,
            bounds = bounds,
            width = size.width.toFloat(),
            height = size.height.toFloat(),
        )
        cb(lat, lng)
    }
    val batchUpdated by rememberUpdatedState(onBatchScreenToLatLng)
    LaunchedEffect(batchScreenToLatLngNonce, batchScreenPoints, bounds, size) {
        if (batchScreenToLatLngNonce <= 0 || batchScreenPoints.isEmpty()) return@LaunchedEffect
        if (size.width <= 0 || size.height <= 0) return@LaunchedEffect
        val cb = batchUpdated ?: return@LaunchedEffect
        val mapped = batchScreenPoints.map { (x, y) ->
            val (lat, lng) = MapProjection.unproject(
                x = x,
                y = y,
                bounds = bounds,
                width = size.width.toFloat(),
                height = size.height.toFloat(),
            )
            lat to lng
        }
        cb(mapped)
    }
    val latLngToScreenUpdated by rememberUpdatedState(onLatLngToScreen)
    LaunchedEffect(latLngToScreenNonce, latLngToScreenPoints, bounds, size) {
        if (latLngToScreenNonce <= 0 || size.width <= 0 || size.height <= 0) return@LaunchedEffect
        val cb = latLngToScreenUpdated ?: return@LaunchedEffect
        val mapped = latLngToScreenPoints.map { (lat, lng) ->
            val pt = MapProjection.project(
                lat = lat,
                lng = lng,
                bounds = bounds,
                width = size.width.toFloat(),
                height = size.height.toFloat(),
            )
            pt.x to pt.y
        }
        cb(mapped)
    }

    Box(
        modifier = modifier
            .background(scheme.surfaceVariant)
            .onSizeChanged { size = it }
            .pointerInput(projected, selectedCarId, simZoom, onMapTap, bounds, size) {
                detectTapGestures { offset ->
                    val hit = MapProjection.hitTest(
                        touchX = offset.x,
                        touchY = offset.y,
                        projected = projected,
                        radiusPx = with(density) { 28.dp.toPx() },
                    )
                    if (hit != null) {
                        if (hit.isCluster) {
                            if (simZoom < 16f) {
                                focusLat = hit.lat
                                focusLng = hit.lng
                                simZoom += 1f
                            } else {
                                onSelectCluster(hit.memberIds)
                            }
                        } else {
                            onSelectCarId(hit.id)
                        }
                    } else if (onMapTap != null && size.width > 0 && size.height > 0) {
                        val (lat, lng) = MapProjection.unproject(
                            x = offset.x,
                            y = offset.y,
                            bounds = bounds,
                            width = size.width.toFloat(),
                            height = size.height.toFloat(),
                        )
                        onMapTap(lat, lng)
                    }
                }
            },
    ) {
        Canvas(modifier = Modifier.fillMaxSize()) {
            val w = this.size.width
            val h = this.size.height
            val step = 48f
            var x = 0f
            while (x < w) {
                drawLine(
                    color = scheme.outlineVariant.copy(alpha = 0.45f),
                    start = Offset(x, 0f),
                    end = Offset(x, h),
                    strokeWidth = 1f,
                )
                x += step
            }
            var y = 0f
            while (y < h) {
                drawLine(
                    color = scheme.outlineVariant.copy(alpha = 0.45f),
                    start = Offset(0f, y),
                    end = Offset(w, y),
                    strokeWidth = 1f,
                )
                y += step
            }

            if (trackPoints.size >= 2 && size.width > 0) {
                val trackPx = trackPoints.map {
                    MapProjection.project(
                        lat = it.lat,
                        lng = it.lng,
                        bounds = bounds,
                        width = w,
                        height = h,
                    )
                }
                for (i in 0 until trackPx.lastIndex) {
                    drawLine(
                        color = Color(0xFF0BB774),
                        start = Offset(trackPx[i].x, trackPx[i].y),
                        end = Offset(trackPx[i + 1].x, trackPx[i + 1].y),
                        strokeWidth = 4f,
                    )
                }
            }

            projected.forEach { (pin, point) ->
                val selected = !pin.isCluster && pin.memberIds.contains(selectedCarId)
                val radius = if (pin.isClusterBubble) 28f else 16f
                val fill = when (pin.icon) {
                    com.luopingtech.ebike.ops.domain.model.MapPinIcon.TrackOrigin -> Color(0xFF00B68A)
                    com.luopingtech.ebike.ops.domain.model.MapPinIcon.TrackEnd -> Color(0xFFEF2E6E)
                    com.luopingtech.ebike.ops.domain.model.MapPinIcon.VehicleRiding -> Color(0xFF00B68A)
                    com.luopingtech.ebike.ops.domain.model.MapPinIcon.UserStart -> Color(0xFF00B68A)
                    com.luopingtech.ebike.ops.domain.model.MapPinIcon.UserEnd -> Color(0xFFEF2E6E)
                    com.luopingtech.ebike.ops.domain.model.MapPinIcon.Parking,
                    com.luopingtech.ebike.ops.domain.model.MapPinIcon.ParkingFunction,
                    -> Color(0xFF1180F9)
                    com.luopingtech.ebike.ops.domain.model.MapPinIcon.ParkingHidden -> Color(0xFFF2A626)
                    com.luopingtech.ebike.ops.domain.model.MapPinIcon.NoParking -> Color(0xFFE02020)
                    else -> when {
                        pin.isClusterBubble -> clusterColor
                        pin.restBattery in 1..30 -> Color(0xFFE67E22)
                        pin.ridingState == 1 -> scheme.primary
                        else -> scheme.secondary
                    }
                }
                drawCircle(
                    color = fill.copy(alpha = 0.92f),
                    radius = radius,
                    center = Offset(point.x, point.y),
                )
                if (pin.isClusterBubble) {
                    val count = pin.memberCount.coerceAtLeast(pin.memberIds.size)
                    val textPaint = AndroidPaint(AndroidPaint.ANTI_ALIAS_FLAG).apply {
                        color = android.graphics.Color.WHITE
                        textAlign = AndroidPaint.Align.CENTER
                        textSize = if (count >= 100) 22f else 28f
                        typeface = Typeface.DEFAULT_BOLD
                    }
                    val y = point.y - (textPaint.descent() + textPaint.ascent()) / 2f
                    drawContext.canvas.nativeCanvas.drawText(
                        count.toString(),
                        point.x,
                        y,
                        textPaint,
                    )
                }
                if (selected) {
                    drawCircle(
                        color = scheme.onSurface,
                        radius = radius + 6f,
                        center = Offset(point.x, point.y),
                        style = Stroke(width = 3f),
                    )
                }
            }
        }

        Text(
            text = when {
                pins.isEmpty() -> Strings.t(Str.NoVehicleCoords)
                else -> Strings.t(Str.SimulatorMapStatus, providerLabel, markers.size, pins.size)
            },
            modifier = Modifier
                .align(Alignment.TopStart)
                .padding(10.dp),
            style = MaterialTheme.typography.labelMedium,
            color = scheme.onSurfaceVariant,
            textAlign = TextAlign.Start,
        )
    }
}
