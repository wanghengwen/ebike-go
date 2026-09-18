package com.luopingtech.ebike.ops.domain.map

import com.luopingtech.ebike.ops.domain.model.MapPin
import kotlin.math.PI
import kotlin.math.abs
import kotlin.math.cos
import kotlin.math.max
import kotlin.math.min
import kotlin.math.pow

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

/**
 * Equirectangular projection for simulator / host-canvas maps.
 * Y grows downward (Android Canvas / Compose).
 */
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
        // Avoid zero-span when all pins coincide.
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

    fun unproject(
        x: Float,
        y: Float,
        bounds: LatLngBounds,
        width: Float,
        height: Float,
        margin: Float = 16f,
    ): Pair<Double, Double> {
        val w = (width - margin * 2).coerceAtLeast(1f)
        val h = (height - margin * 2).coerceAtLeast(1f)
        val lng = bounds.minLng + ((x - margin) / w).toDouble() * bounds.lngSpan
        val lat = bounds.minLat + (1.0 - ((y - margin) / h).toDouble()) * bounds.latSpan
        return lat to lng
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

/**
 * Grid cluster for overview zoom. Pure domain — no Android types.
 * [cellDegrees] ≈ screen cell size in degrees; larger → fewer markers.
 */
object MapClusterer {
    /** Legacy AMapClusterManagerV2/V3: at most 100 cluster/vehicle markers. */
    const val MAX_MARKERS = 100

    /** 对齐原版：只聚当前视野内的点，再截前 [maxMarkers]。 */
    fun clusterInViewport(
        pins: List<MapPin>,
        cellDegrees: Double,
        visible: LatLngBounds?,
        maxMarkers: Int = MAX_MARKERS,
    ): List<MapPin> {
        val source = if (visible == null) pins else pinsInBounds(pins, visible.padded(0.05))
        return cluster(source, cellDegrees, maxMarkers)
    }

    fun pinsInBounds(pins: List<MapPin>, bounds: LatLngBounds): List<MapPin> =
        pins.filter { pin ->
            pin.lat in bounds.minLat..bounds.maxLat && pin.lng in bounds.minLng..bounds.maxLng
        }

    /**
     * 非聚合时按数量做中心距离裁剪（对齐原版 LoadMarkerHelp）：
     * &lt;300 全显；300–1000 约 500m；≥1000 约 300m。
     */
    fun filterNearCenter(
        pins: List<MapPin>,
        centerLat: Double,
        centerLng: Double,
    ): List<MapPin> {
        if (pins.size < 300) return pins
        val radiusMeters = if (pins.size >= 1000) 300.0 else 500.0
        val latDelta = radiusMeters / 111_320.0
        val lngDelta = radiusMeters / (111_320.0 * abs(cos(centerLat * PI / 180.0)).coerceAtLeast(0.2))
        return pins.filter {
            abs(it.lat - centerLat) <= latDelta && abs(it.lng - centerLng) <= lngDelta
        }
    }

    /**
     * 聚合模式建簇。对齐 legacy AMapClusterManagerV3：簇的 showCluster 恒为 true，
     * 所以只含 1 台车的簇同样画「1」数字气泡（非聚合模式走 [filterNearCenter]，不进这里）。
     */
    fun cluster(pins: List<MapPin>, cellDegrees: Double, maxMarkers: Int = MAX_MARKERS): List<MapPin> {
        if (pins.isEmpty()) return emptyList()
        val cell = cellDegrees.coerceAtLeast(1e-6)
        val buckets = linkedMapOf<String, MutableList<MapPin>>()
        pins.forEach { pin ->
            val keyLat = kotlin.math.floor(pin.lat / cell).toLong()
            val keyLng = kotlin.math.floor(pin.lng / cell).toLong()
            val key = "$keyLat:$keyLng"
            buckets.getOrPut(key) { mutableListOf() }.add(pin)
        }
        val clustered = buckets.values.map { group ->
            if (group.size == 1) {
                group.first().copy(
                    memberCount = 1,
                    memberIds = listOf(group.first().id),
                    showCluster = true,
                )
            } else {
                // Legacy DefaultOptionGenerator: clusterItems.getLast() so the bubble
                // stays on a real vehicle instead of the geographic average (can sit outside the service area).
                val last = group.last()
                val ids = group.map { it.id }
                MapPin(
                    id = "cluster:${ids.sorted().joinToString(",")}",
                    lat = last.lat,
                    lng = last.lng,
                    title = "${group.size}",
                    subtitle = "vehicles",
                    restBattery = last.restBattery,
                    ridingState = last.ridingState,
                    memberCount = group.size,
                    memberIds = ids,
                    icon = last.icon,
                    badgeDrawableName = last.badgeDrawableName,
                    showCluster = true,
                )
            }
        }
        return if (clustered.size > maxMarkers) clustered.take(maxMarkers) else clustered
    }

    /**
     * Legacy ClusterPixelEnum × metersPerPixel: pixel threshold stays ~40–60dp,
     * so geographic cell shrinks as zoom grows and clusters break apart.
     */
    fun cellDegreesForZoom(zoom: Float, latitude: Double = 30.0): Double {
        val dp = when {
            zoom < 10f -> 60.0
            zoom < 13f -> 50.0
            zoom < 14f -> 40.0
            else -> 60.0
        }
        val px = dp * 3.0
        val z = zoom.toDouble().coerceIn(3.0, 22.0)
        val metersPerPixel =
            156_543.03392 * abs(cos(latitude * PI / 180.0)).coerceAtLeast(0.2) / 2.0.pow(z)
        return ((px * metersPerPixel) / 111_320.0).coerceAtLeast(1e-6)
    }
}
