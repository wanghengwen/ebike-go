package com.luopingtech.ebike.ops.domain.geo

import com.luopingtech.ebike.ops.platform.GeoPoint
import kotlin.math.PI
import kotlin.math.atan2
import kotlin.math.cos
import kotlin.math.sin
import kotlin.math.sqrt

object GeoMath {
    private const val EARTH_RADIUS_M = 6_371_000.0

    /** Great-circle distance in meters (WGS84 sphere approximation). */
    fun distanceMeters(a: GeoPoint, b: GeoPoint): Double {
        val lat1 = toRadians(a.latitude)
        val lat2 = toRadians(b.latitude)
        val dLat = toRadians(b.latitude - a.latitude)
        val dLng = toRadians(b.longitude - a.longitude)
        val h = sin(dLat / 2) * sin(dLat / 2) +
            cos(lat1) * cos(lat2) * sin(dLng / 2) * sin(dLng / 2)
        return 2 * EARTH_RADIUS_M * atan2(sqrt(h), sqrt(1 - h))
    }

    private fun toRadians(degrees: Double): Double = degrees * PI / 180.0
}
