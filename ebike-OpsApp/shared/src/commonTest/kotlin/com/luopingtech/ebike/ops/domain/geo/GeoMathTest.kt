package com.luopingtech.ebike.ops.domain.geo

import com.luopingtech.ebike.ops.platform.GeoPoint
import kotlin.test.Test
import kotlin.test.assertTrue

class GeoMathTest {
    @Test
    fun nearbyPoints_aboutTwentyMeters() {
        // ~0.00018 deg lat ≈ 20m
        val a = GeoPoint(30.0, 104.0)
        val b = GeoPoint(30.00018, 104.0)
        val d = GeoMath.distanceMeters(a, b)
        assertTrue(d in 18.0..22.0, "distance=$d")
    }
}
