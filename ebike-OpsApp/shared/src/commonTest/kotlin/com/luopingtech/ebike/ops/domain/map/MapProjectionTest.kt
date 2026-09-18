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
    fun cellDegreesShrinksAsZoomGrows() {
        val wide = MapClusterer.cellDegreesForZoom(8f, latitude = 30.0)
        val tight = MapClusterer.cellDegreesForZoom(16f, latitude = 30.0)
        assertTrue(tight < wide / 8.0)
    }

    @Test
    fun highZoomSplitsNearbyPins() {
        val pins = listOf(
            MapPin("1", 28.0, 112.0, "1"),
            MapPin("2", 28.005, 112.005, "2"),
        )
        val low = MapClusterer.cluster(pins, MapClusterer.cellDegreesForZoom(11f, 28.0))
        val high = MapClusterer.cluster(pins, MapClusterer.cellDegreesForZoom(18f, 28.0))
        assertEquals(1, low.size)
        assertTrue(low.first().isCluster)
        assertEquals(2, high.size)
        assertTrue(high.none { it.isCluster })
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
        val pair = clustered.first { it.isCluster }
        assertEquals(28.0001, pair.lat, 1e-9)
        assertEquals(112.0001, pair.lng, 1e-9)
    }

    /** 对齐 legacy DefaultOptionGenerator：聚合模式下单车也画「1」气泡，但点击仍是单车语义。 */
    @Test
    fun clusterMarksSinglePinAsBubble() {
        val pins = listOf(
            MapPin("1", 28.0, 112.0, "1"),
            MapPin("2", 29.0, 113.0, "2"),
        )
        val clustered = MapClusterer.cluster(pins, cellDegrees = 0.01)
        assertEquals(2, clustered.size)
        assertTrue(clustered.all { it.showCluster && it.isClusterBubble })
        assertTrue(clustered.none { it.isCluster })
        assertTrue(clustered.all { it.memberCount == 1 && it.memberIds.size == 1 })
    }

    @Test
    fun clusterInViewportOnlyUsesVisiblePins() {
        val pins = listOf(
            MapPin("in", 28.0, 112.0, "in"),
            MapPin("out", 40.0, 120.0, "out"),
            MapPin("near", 28.001, 112.001, "near"),
        )
        val visible = LatLngBounds(27.9, 28.1, 111.9, 112.1)
        val clustered = MapClusterer.clusterInViewport(
            pins,
            cellDegrees = 0.01,
            visible = visible,
        )
        assertTrue(clustered.all { it.lat in 27.9..28.1 })
        assertTrue(clustered.none { it.id.contains("out") || it.memberIds.contains("out") })
    }

    @Test
    fun filterNearCenterKeepsAllUnder300() {
        val pins = (1..50).map { MapPin("$it", 28.0 + it * 0.01, 112.0, "$it") }
        assertEquals(50, MapClusterer.filterNearCenter(pins, 28.0, 112.0).size)
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
