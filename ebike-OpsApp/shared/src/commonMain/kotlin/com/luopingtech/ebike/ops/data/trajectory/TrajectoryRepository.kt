package com.luopingtech.ebike.ops.data.trajectory

import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.time.nowEpochMillis
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.TrackPoint
import kotlin.math.cos
import kotlin.math.sin

interface TrajectoryRepository {
    suspend fun loadRealTime(
        imei: String,
        startTimeMs: Long,
        endTimeMs: Long,
    ): OpsResult<List<TrackPoint>>
}

class TrajectoryRepositoryImpl(
    private val demoMode: Boolean,
    private val api: TrajectoryApi? = null,
) : TrajectoryRepository {
    override suspend fun loadRealTime(
        imei: String,
        startTimeMs: Long,
        endTimeMs: Long,
    ): OpsResult<List<TrackPoint>> {
        if (imei.isBlank()) {
            return OpsResult.Err(OpsError.business("TRACK", "imei empty"))
        }
        if (demoMode || api == null) {
            return OpsResult.Ok(demoTrack(imei))
        }
        return api.realTime(imei.trim(), startTimeMs, endTimeMs)
    }

    companion object {
        fun demoTrack(imei: String): List<TrackPoint> {
            val baseLat = 28.221
            val baseLng = 112.941
            return (0 until 12).map { i ->
                val t = i / 12.0 * 6.28
                TrackPoint(
                    lat = baseLat + 0.001 * sin(t),
                    lng = baseLng + 0.001 * cos(t),
                    timestamp = nowEpochMillis() - (12 - i) * 60_000L,
                    speed = 8f,
                )
            }
        }
    }
}
