package com.luopingtech.ebike.ops.domain.map

import com.luopingtech.ebike.ops.domain.model.MapPin
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class MapProjectionTest {
    @Test
    fun boundsPadsNonEmptyPins() {
        val pins = listOf(
            MapPin("a", 28.0, 112.0, "a"),
            MapPin("b", 28.01, 112.02, "b"),
        )
        val bounds = MapProjection.boundsOf(pins)
        assertTrue(bounds.minLat < 28.0)
        assertTrue(bounds.maxLat > 28.01)
        assertTrue(bounds.minLng < 112.0)
        assertTrue(bounds.maxLng > 112.02)
    }

    @Test
    fun projectCorners() {
        val bounds = LatLngBounds(0.0, 10.0, 0.0, 10.0)
        val topLeft = MapProjection.project(10.0, 0.0, bounds, 100f, 100f, margin = 0f)
        assertEquals(0f, topLeft.x, 0.01f)
        assertEquals(0f, topLeft.y, 0.01f)
        val bottomRight = MapProjection.project(0.0, 10.0, bounds, 100f, 100f, margin = 0f)
        assertEquals(100f, bottomRight.x, 0.01f)
        assertEquals(100f, bottomRight.y, 0.01f)
    }

    @Test
    fun clusterMergesNearbyPins() {
        val pins = listOf(
            MapPin("1", 28.0, 112.0, "1"),
            MapPin("2", 28.0001, 112.0001, "2"),
            MapPin("3", 29.0, 113.0, "3"),
        )
        val clustered = MapClusterer.cluster(pins, cellDegrees = 0.01)
        assertEquals(2, clustered.size)
        assertTrue(clustered.any { it.isCluster && it.memberCount == 2 })
        assertTrue(clustered.any { !it.isCluster })
    }

    @Test
    fun hitTestPicksNearest() {
        val pinA = MapPin("a", 0.0, 0.0, "a")
        val pinB = MapPin("b", 0.0, 0.0, "b")
        val hit = MapProjection.hitTest(
            touchX = 12f,
            touchY = 12f,
            projected = listOf(
                pinA to ScreenPoint(10f, 10f),
                pinB to ScreenPoint(80f, 80f),
            ),
            radiusPx = 40f,
        )
        assertEquals("a", hit?.id)
    }
}
