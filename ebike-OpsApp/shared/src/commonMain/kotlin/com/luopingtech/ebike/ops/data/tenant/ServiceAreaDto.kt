package com.luopingtech.ebike.ops.data.tenant

import com.luopingtech.ebike.ops.domain.model.ServiceArea
import kotlinx.serialization.Serializable

@Serializable
data class ServiceAreaDto(
    val id: Long = 0,
    val name: String = "",
    val centerLat: Double = 0.0,
    val centerLng: Double = 0.0,
    val agentId: String = "",
    val minDistance: Double = 0.0,
    val type: Int = 0,
    val shapeType: String = "",
) {
    fun toDomain(): ServiceArea = ServiceArea(
        id = id.toString(),
        name = name,
        centerLat = centerLat,
        centerLng = centerLng,
        agentId = agentId,
        minDistance = minDistance,
    )
}
