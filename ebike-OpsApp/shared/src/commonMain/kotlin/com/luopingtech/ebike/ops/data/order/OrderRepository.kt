package com.luopingtech.ebike.ops.data.order

import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.time.nowEpochMillis
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.LastOrder
import com.luopingtech.ebike.ops.domain.model.TrackPoint
import kotlin.math.cos
import kotlin.math.sin

interface OrderRepository {
    suspend fun detailLast(carId: String): OpsResult<LastOrder>
}

class OrderRepositoryImpl(
    private val demoMode: Boolean,
    private val api: OrderApi? = null,
) : OrderRepository {
    override suspend fun detailLast(carId: String): OpsResult<LastOrder> {
        if (carId.isBlank()) {
            return OpsResult.Err(OpsError.business("ORDER", "carId empty"))
        }
        if (demoMode || api == null) {
            return OpsResult.Ok(demoLastOrder(carId.trim()))
        }
        return api.detailLast(carId.trim())
    }

    companion object {
        fun demoLastOrder(carId: String): LastOrder {
            val baseLat = 28.221
            val baseLng = 112.941
            val track = (0 until 10).map { i ->
                val t = i / 10.0 * 6.28
                TrackPoint(
                    lat = baseLat + 0.0008 * sin(t),
                    lng = baseLng + 0.0008 * cos(t),
                    timestamp = nowEpochMillis() - (10 - i) * 90_000L,
                    speed = 10f,
                )
            }
            return LastOrder(
                carId = carId,
                startLat = track.first().lat,
                startLng = track.first().lng,
                endLat = track.last().lat,
                endLng = track.last().lng,
                trajectory = track,
                id = "demo-order-$carId",
                userPin = "demo-user-pin",
                userPhone = "13800138000",
                startTime = "2026-09-11 08:00:00",
                endTime = "2026-09-11 08:25:00",
            )
        }
    }
}
