package com.luopingtech.ebike.ops.data.trajectory

import com.luopingtech.ebike.ops.domain.model.TrackPoint
import kotlinx.serialization.Serializable

@Serializable
data class TrackPointDto(
    val lat: Double? = null,
    val lng: Double? = null,
    val timestamp: Long? = null,
    val speed: Float? = null,
    val course: Float? = null,
) {
    fun toDomain(): TrackPoint? {
        val la = lat ?: return null
        val ln = lng ?: return null
        return TrackPoint(
            lat = la,
            lng = ln,
            timestamp = timestamp ?: 0L,
            speed = speed ?: 0f,
            course = course ?: 0f,
        )
    }
}
