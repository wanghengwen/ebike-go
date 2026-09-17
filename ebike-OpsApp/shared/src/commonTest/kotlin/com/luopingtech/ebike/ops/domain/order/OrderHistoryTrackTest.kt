package com.luopingtech.ebike.ops.domain.order

import com.luopingtech.ebike.ops.domain.model.MapPinIcon
import com.luopingtech.ebike.ops.domain.model.TrackPoint
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class OrderHistoryTrackTest {
    @Test
    fun interpolate_midpoint_by_timestamp() {
        val points = listOf(
            TrackPoint(lat = 30.0, lng = 120.0, timestamp = 1000),
            TrackPoint(lat = 31.0, lng = 121.0, timestamp = 2000),
        )
        val mid = interpolateTrackPoint(points, 0.5f)!!
        assertEquals(30.5, mid.lat, 1e-6)
        assertEquals(120.5, mid.lng, 1e-6)
        assertEquals(1500, mid.timestamp)
    }

    @Test
    fun pins_include_track_origin_end_and_playback() {
        val order = OrderRecord(
            id = "1",
            startLat = 30.1,
            startLng = 120.1,
            endLat = 30.2,
            endLng = 120.2,
        )
        val track = listOf(
            TrackPoint(lat = 30.0, lng = 120.0, timestamp = 1000),
            TrackPoint(lat = 31.0, lng = 121.0, timestamp = 2000),
        )
        val pins = buildOrderHistoryPins(order, track, 0f)
        assertTrue(pins.any { it.icon == MapPinIcon.TrackOrigin })
        assertTrue(pins.any { it.icon == MapPinIcon.TrackEnd })
        assertTrue(pins.any { it.icon == MapPinIcon.UserStart })
        assertTrue(pins.any { it.icon == MapPinIcon.UserEnd })
        assertTrue(pins.any { it.id == ORDER_PLAYBACK_PIN_ID && it.icon == MapPinIcon.VehicleRiding })
    }
}
