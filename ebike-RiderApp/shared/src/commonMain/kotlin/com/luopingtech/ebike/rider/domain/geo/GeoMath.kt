package com.luopingtech.ebike.rider.domain.geo

import com.luopingtech.ebike.rider.domain.model.GeoLatLng
import com.luopingtech.ebike.rider.platform.GeoPoint
import kotlin.math.PI
import kotlin.math.abs
import kotlin.math.asin
import kotlin.math.cos
import kotlin.math.min
import kotlin.math.sin
import kotlin.math.sqrt

/**
 * 球面距离与点在多边形内判定。与 OpsApp `domain/geo/GeoMath.kt` 同语义（只读拷贝）。
 */
object GeoMath {
    private const val EARTH_RADIUS_METERS = 6_371_000.0
    private const val DEG_TO_RAD = PI / 180.0

    fun distanceMeters(a: GeoPoint, b: GeoPoint): Double =
        distanceMeters(a.latitude, a.longitude, b.latitude, b.longitude)

    fun distanceMeters(a: GeoLatLng, b: GeoLatLng): Double =
        distanceMeters(a.lat, a.lng, b.lat, b.lng)

    /** Haversine。短距离下比等距圆柱投影稳，且不会在跨经线时炸掉。 */
    fun distanceMeters(lat1: Double, lng1: Double, lat2: Double, lng2: Double): Double {
        val dLat = (lat2 - lat1) * DEG_TO_RAD
        val dLng = (lng2 - lng1) * DEG_TO_RAD
        val sinLat = sin(dLat / 2)
        val sinLng = sin(dLng / 2)
        val h = sinLat * sinLat +
            cos(lat1 * DEG_TO_RAD) * cos(lat2 * DEG_TO_RAD) * sinLng * sinLng
        return 2 * EARTH_RADIUS_METERS * asin(min(1.0, sqrt(h)))
    }

    /** 折线总长（米）。轨迹里程用。 */
    fun polylineLengthMeters(points: List<GeoLatLng>): Double {
        if (points.size < 2) return 0.0
        var total = 0.0
        for (i in 1 until points.size) {
            total += distanceMeters(points[i - 1], points[i])
        }
        return total
    }

    /**
     * 射线法判定点是否在多边形内。用于「还车前本地预判是否在站点内」这类不值得
     * 打一次网络的场景 —— 真正的判定始终以 `returnPermission` 为准。
     */
    fun isInsidePolygon(point: GeoLatLng, polygon: List<GeoLatLng>): Boolean {
        if (polygon.size < 3) return false
        var inside = false
        var j = polygon.size - 1
        for (i in polygon.indices) {
            val pi = polygon[i]
            val pj = polygon[j]
            // 顶点重合时 (pj.lat - pi.lat) 为 0，下面的除法会产生 NaN；先跳过退化边。
            if (abs(pj.lat - pi.lat) > 1e-12) {
                val crosses = (pi.lat > point.lat) != (pj.lat > point.lat)
                if (crosses) {
                    val x = (pj.lng - pi.lng) * (point.lat - pi.lat) / (pj.lat - pi.lat) + pi.lng
                    if (point.lng < x) inside = !inside
                }
            }
            j = i
        }
        return inside
    }
}
