package com.luopingtech.ebike.rider.domain.tracking

import com.luopingtech.ebike.rider.domain.geo.GeoMath
import com.luopingtech.ebike.rider.platform.GeoPoint

/**
 * 骑行轨迹攒批。从 `ebike-OpsApp` 的 `domain/tracking/TrackPointBuffer.kt` 只读拷过来，
 * 语义一字不改（运维端那份对齐旧 `MainViewModel`）：
 *
 * - 每个定位 tick 让 [ticks] +1
 * - 缓冲为空、或位移超过 [TrackUploadPolicy.minDistanceMeters] 才入队
 * - [ticks] 到 [TrackUploadPolicy.flushEveryTicks] 就吐一次 [TrackBufferEvent.Flush]
 * - 超过 [TrackUploadPolicy.maxBuffer] 整桶丢弃（旧版行为：宁可丢轨迹也不让内存涨）
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
