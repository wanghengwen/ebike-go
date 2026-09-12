package com.luopingtech.ebike.ops.domain.tracking

import com.luopingtech.ebike.ops.platform.GeoPoint
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertIs
import kotlin.test.assertTrue

class TrackPointBufferTest {
    private val origin = GeoPoint(30.0, 104.0)

    @Test
    fun enqueuesWhenMovedOver20m_andFlushesAt12Ticks() {
        val buffer = TrackPointBuffer(TrackUploadPolicy(flushEveryTicks = 12))
        var flush: TrackBufferEvent.Flush? = null
        repeat(12) { i ->
            // each step ~25m north
            val point = GeoPoint(30.0 + i * 0.00025, 104.0)
            val event = buffer.onTick(point)
            if (event is TrackBufferEvent.Flush) flush = event
        }
        assertIs<TrackBufferEvent.Flush>(flush)
        assertTrue(flush!!.points.size >= 2)
        assertEquals(12, buffer.ticks)
    }

    @Test
    fun ignoresSmallMoves_butStillCountsTicks() {
        val buffer = TrackPointBuffer(TrackUploadPolicy(flushEveryTicks = 3))
        buffer.onTick(origin)
        // ~2m
        val near = GeoPoint(30.00001, 104.0)
        assertIs<TrackBufferEvent.None>(buffer.onTick(near))
        assertEquals(1, buffer.size)
        val flush = buffer.onTick(near)
        assertIs<TrackBufferEvent.Flush>(flush)
        assertEquals(1, flush.points.size)
    }

    @Test
    fun clearAfterSuccess_resetsTicks() {
        val buffer = TrackPointBuffer(TrackUploadPolicy(flushEveryTicks = 2))
        buffer.onTick(origin)
        assertIs<TrackBufferEvent.Flush>(buffer.onTick(GeoPoint(30.001, 104.0)))
        buffer.clearAfterSuccess()
        assertEquals(0, buffer.size)
        assertEquals(0, buffer.ticks)
    }
}
