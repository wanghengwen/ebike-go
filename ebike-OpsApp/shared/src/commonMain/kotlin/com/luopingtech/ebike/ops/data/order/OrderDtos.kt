package com.luopingtech.ebike.ops.data.order

import com.luopingtech.ebike.ops.data.trajectory.TrackPointDto
import com.luopingtech.ebike.ops.domain.model.LastOrder
import com.luopingtech.ebike.ops.domain.model.TrackPoint
import kotlinx.serialization.Serializable

@Serializable
data class LastOrderDto(
    val carId: String? = null,
    val startLat: Double? = null,
    val startLng: Double? = null,
    val endLat: Double? = null,
    val endLng: Double? = null,
    val deviceTrajectory: List<TrackPointDto>? = null,
    val id: String? = null,
    val userPin: String? = null,
    val phone: String? = null,
    val userPhone: String? = null,
    val startTime: String? = null,
    val endTime: String? = null,
) {
    fun toDomain(): LastOrder {
        val track = deviceTrajectory.orEmpty().mapNotNull { it.toDomain() }
        return LastOrder(
            carId = carId.orEmpty(),
            startLat = startLat ?: 0.0,
            startLng = startLng ?: 0.0,
            endLat = endLat,
            endLng = endLng,
            trajectory = track,
            id = id.orEmpty(),
            userPin = userPin.orEmpty(),
            userPhone = phone?.takeIf { it.isNotBlank() } ?: userPhone.orEmpty(),
            startTime = startTime.orEmpty(),
            endTime = endTime.orEmpty(),
        )
    }
}
