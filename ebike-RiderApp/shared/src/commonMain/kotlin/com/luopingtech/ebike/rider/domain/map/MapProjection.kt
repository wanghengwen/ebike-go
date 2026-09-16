package com.luopingtech.ebike.rider.domain.map

import com.luopingtech.ebike.rider.domain.model.MapPin
import kotlin.math.PI
import kotlin.math.abs
import kotlin.math.cos
import kotlin.math.max
import kotlin.math.min

data class LatLngBounds(
    val minLat: Double,
    val maxLat: Double,
    val minLng: Double,
    val maxLng: Double,
) {
    val latSpan: Double get() = max(maxLat - minLat, 1e-9)
    val lngSpan: Double get() = max(maxLng - minLng, 1e-9)

    fun padded(fraction: Double = 0.12): LatLngBounds {
        val f = fraction.coerceIn(0.0, 0.45)
        val dLat = latSpan * f
        val dLng = lngSpan * f
        return LatLngBounds(
            minLat = minLat - dLat,
            maxLat = maxLat + dLat,
            minLng = minLng - dLng,
            maxLng = maxLng + dLng,
        )
    }
}

data class ScreenPoint(val x: Float, val y: Float)

/** Equirectangular projection for simulator / host-canvas maps. */
object MapProjection {
    fun boundsOf(pins: List<MapPin>, fallbackLat: Double = 28.22, fallbackLng: Double = 112.94): LatLngBounds {
        val valid = pins.filter { it.lat != 0.0 || it.lng != 0.0 }
        if (valid.isEmpty()) {
            return LatLngBounds(
                minLat = fallbackLat - 0.01,
                maxLat = fallbackLat + 0.01,
                minLng = fallbackLng - 0.01,
                maxLng = fallbackLng + 0.01,
            ).padded()
        }
        var minLat = valid.first().lat
        var maxLat = minLat
        var minLng = valid.first().lng
        var maxLng = minLng
        valid.forEach { pin ->
            minLat = min(minLat, pin.lat)
            maxLat = max(maxLat, pin.lat)
            minLng = min(minLng, pin.lng)
            maxLng = max(maxLng, pin.lng)
        }
        if (abs(maxLat - minLat) < 1e-5) {
            minLat -= 0.002
            maxLat += 0.002
        }
        if (abs(maxLng - minLng) < 1e-5) {
            minLng -= 0.002
            maxLng += 0.002
        }
        return LatLngBounds(minLat, maxLat, minLng, maxLng).padded()
    }

    fun project(
        lat: Double,
        lng: Double,
        bounds: LatLngBounds,
        width: Float,
        height: Float,
        margin: Float = 16f,
    ): ScreenPoint {
        val w = (width - margin * 2).coerceAtLeast(1f)
        val h = (height - margin * 2).coerceAtLeast(1f)
        val x = margin + ((lng - bounds.minLng) / bounds.lngSpan).toFloat() * w
        val y = margin + (1f - ((lat - bounds.minLat) / bounds.latSpan).toFloat()) * h
        return ScreenPoint(x, y)
    }

    fun hitTest(
        touchX: Float,
        touchY: Float,
        projected: List<Pair<MapPin, ScreenPoint>>,
        radiusPx: Float = 36f,
    ): MapPin? {
        var best: MapPin? = null
        var bestDist = radiusPx * radiusPx
        projected.forEach { (pin, point) ->
            val dx = point.x - touchX
            val dy = point.y - touchY
            val d = dx * dx + dy * dy
            if (d <= bestDist) {
                bestDist = d
                best = pin
            }
        }
        return best
    }
}

object MapClusterer {
    fun cluster(pins: List<MapPin>, cellDegrees: Double): List<MapPin> {
        if (pins.isEmpty()) return emptyList()
        val cell = cellDegrees.coerceAtLeast(1e-6)
        val buckets = linkedMapOf<String, MutableList<MapPin>>()
        pins.forEach { pin ->
            val keyLat = kotlin.math.floor(pin.lat / cell).toLong()
            val keyLng = kotlin.math.floor(pin.lng / cell).toLong()
            val key = "$keyLat:$keyLng"
            buckets.getOrPut(key) { mutableListOf() }.add(pin)
        }
        return buckets.values.map { group ->
            if (group.size == 1) {
                group.first().copy(memberCount = 1, memberIds = listOf(group.first().id))
            } else {
                val lat = group.map { it.lat }.average()
                val lng = group.map { it.lng }.average()
                val ids = group.map { it.id }
                MapPin(
                    id = "cluster:${ids.sorted().joinToString(",")}",
                    lat = lat,
                    lng = lng,
                    title = "${group.size}",
                    subtitle = "vehicles",
                    restBattery = group.map { it.restBattery }.average().toInt(),
                    memberCount = group.size,
                    memberIds = ids,
                )
            }
        }
    }

    fun suggestedCellDegrees(bounds: LatLngBounds, targetCells: Int = 6): Double {
        val span = max(
            bounds.latSpan,
            bounds.lngSpan * abs(cos(bounds.minLat * PI / 180.0)).coerceAtLeast(0.5),
        )
        return (span / targetCells.coerceIn(3, 12)).coerceAtLeast(1e-5)
    }
}
