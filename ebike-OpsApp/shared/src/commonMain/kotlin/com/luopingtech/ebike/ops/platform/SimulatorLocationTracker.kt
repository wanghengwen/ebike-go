package com.luopingtech.ebike.ops.platform

import com.luopingtech.ebike.ops.core.result.OpsError
import kotlin.concurrent.Volatile
import com.luopingtech.ebike.ops.core.result.OpsResult
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.isActive
import kotlin.coroutines.coroutineContext

/**
 * Walks a small path around [origin] so demo can flush without device GPS.
 * Default interval matches legacy Android GPS minTime (~5s).
 */
class SimulatorLocationTracker(
    // Align with demo ServiceArea / task catalog (Changsha) so 到点校验可用。
    private val origin: GeoPoint = GeoPoint(latitude = 28.22, longitude = 112.94),
    private val intervalMs: Long = 5_000L,
    private val stepMetersApprox: Double = 25.0,
) : LocationTracker {
    @Volatile
    private var running = false
    private var step = 0

    override suspend fun currentLocation(): OpsResult<GeoPoint> = OpsResult.Ok(peekNext())

    override fun track(): Flow<GeoPoint> = flow {
        while (coroutineContext.isActive) {
            if (running) {
                emit(nextPoint())
            }
            delay(intervalMs)
        }
    }

    override fun startTracking() {
        running = true
    }

    override fun stopTracking() {
        running = false
    }

    /** Deterministic nudge (degrees ≈ meters/111_320). */
    fun nextPoint(): GeoPoint {
        val point = peekNext()
        step++
        return point
    }

    private fun peekNext(): GeoPoint {
        val deltaDeg = stepMetersApprox / 111_320.0
        return GeoPoint(
            latitude = origin.latitude + step * deltaDeg,
            longitude = origin.longitude + (step % 3) * deltaDeg,
        )
    }
}
