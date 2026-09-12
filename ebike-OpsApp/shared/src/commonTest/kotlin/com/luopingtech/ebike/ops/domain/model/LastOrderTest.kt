package com.luopingtech.ebike.ops.domain.model

import kotlin.test.Test
import kotlin.test.assertEquals

class LastOrderTest {
    @Test
    fun nearFenceLocations_usesStartAndEnd() {
        val order = LastOrder(
            carId = "c1",
            startLat = 28.1,
            startLng = 112.9,
            endLat = 28.2,
            endLng = 113.0,
            trajectory = emptyList(),
        )
        assertEquals(
            listOf(GeoLatLng(28.1, 112.9), GeoLatLng(28.2, 113.0)),
            order.nearFenceLocations(),
        )
    }

    @Test
    fun nearFenceLocations_fallsBackToLastTrackPoint() {
        val order = LastOrder(
            carId = "c1",
            startLat = 28.1,
            startLng = 112.9,
            endLat = null,
            endLng = null,
            trajectory = listOf(
                TrackPoint(28.11, 112.91),
                TrackPoint(28.12, 112.92),
            ),
        )
        assertEquals(
            listOf(GeoLatLng(28.1, 112.9), GeoLatLng(28.12, 112.92)),
            order.nearFenceLocations(),
        )
    }
}
