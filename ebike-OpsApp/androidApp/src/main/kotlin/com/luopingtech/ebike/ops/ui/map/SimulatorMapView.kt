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
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.IntSize
import androidx.compose.ui.unit.dp
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.domain.map.MapClusterer
import com.luopingtech.ebike.ops.domain.map.MapProjection
import com.luopingtech.ebike.ops.domain.model.MapPin

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
    modifier: Modifier = Modifier,
    clusterOverview: Boolean = true,
) {
    val scheme = MaterialTheme.colorScheme
    var size by remember { mutableStateOf(IntSize.Zero) }
    val density = LocalDensity.current
    val bounds = remember(pins) {
        MapProjection.boundsOf(pins)
    }
    val markers = remember(pins, clusterOverview, bounds) {
        if (!clusterOverview || pins.size <= 4) {
            pins
        } else {
            val cell = MapClusterer.suggestedCellDegrees(bounds, targetCells = 5)
            MapClusterer.cluster(pins, cell)
        }
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

    Box(
        modifier = modifier
            .background(scheme.surfaceVariant)
            .onSizeChanged { size = it }
            .pointerInput(projected, selectedCarId) {
                detectTapGestures { offset ->
                    val hit = MapProjection.hitTest(
                        touchX = offset.x,
                        touchY = offset.y,
                        projected = projected,
                        radiusPx = with(density) { 28.dp.toPx() },
                    ) ?: return@detectTapGestures
                    if (hit.isCluster) {
                        onSelectCluster(hit.memberIds)
                    } else {
                        onSelectCarId(hit.id)
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

            projected.forEach { (pin, point) ->
                val selected = !pin.isCluster && pin.memberIds.contains(selectedCarId)
                val radius = if (pin.isCluster) 28f else 16f
                val fill = when {
                    pin.isCluster -> scheme.tertiary
                    pin.restBattery in 1..30 -> Color(0xFFE67E22)
                    pin.ridingState == 1 -> scheme.primary
                    else -> scheme.secondary
                }
                drawCircle(
                    color = fill.copy(alpha = 0.92f),
                    radius = radius,
                    center = Offset(point.x, point.y),
                )
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
