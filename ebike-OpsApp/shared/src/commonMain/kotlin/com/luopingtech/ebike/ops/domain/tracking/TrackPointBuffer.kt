package com.luopingtech.ebike.ops.domain.tracking

import com.luopingtech.ebike.ops.domain.geo.GeoMath
import com.luopingtech.ebike.ops.platform.GeoPoint

/**
 * Legacy MainViewModel semantics:
 * - every location tick increments [tickCount]
 * - enqueue when buffer empty or move > [minDistanceMeters]
 * - flush when [tickCount] >= [flushEveryTicks]
 * - drop whole buffer if size exceeds [maxBuffer]
 */
data class TrackUploadPolicy(
    val minDistanceMeters: Double = 20.0,
    val flushEveryTicks: Int = 12,
    val maxBuffer: Int = 200,
)

sealed class TrackBufferEvent {
    data object None : TrackBufferEvent()
    data class Flush(val points: List<GeoPoint>) : TrackBufferEvent()
}

class TrackPointBuffer(
    private val policy: TrackUploadPolicy = TrackUploadPolicy(),
) {
    private val points = ArrayList<GeoPoint>()
    private var tickCount: Int = 0

    val size: Int get() = points.size
    val ticks: Int get() = tickCount
    fun snapshot(): List<GeoPoint> = points.toList()

    fun onTick(point: GeoPoint): TrackBufferEvent {
        tickCount++
        if (points.size > policy.maxBuffer) {
            points.clear()
        }
        if (points.isEmpty()) {
            points.add(point)
        } else {
            val last = points.last()
            if (GeoMath.distanceMeters(last, point) > policy.minDistanceMeters) {
                points.add(point)
            }
        }
        if (tickCount >= policy.flushEveryTicks) {
            return TrackBufferEvent.Flush(points.toList())
        }
        return TrackBufferEvent.None
    }

    fun clearAfterSuccess() {
        points.clear()
        tickCount = 0
    }

    fun reset() {
        points.clear()
        tickCount = 0
    }
}
