package com.luopingtech.ebike.ops.data.fence

import com.luopingtech.ebike.ops.domain.model.FenceBundle
import com.luopingtech.ebike.ops.domain.model.FenceKind
import com.luopingtech.ebike.ops.domain.model.FencePolygon
import com.luopingtech.ebike.ops.domain.model.GeoLatLng
import kotlinx.serialization.Serializable

@Serializable
data class FenceBundleDto(
    val serviceAreas: List<FenceInfoDto>? = null,
    val parkings: List<FenceInfoDto>? = null,
    val noParkings: List<FenceInfoDto>? = null,
    val maintainAreas: List<FenceInfoDto>? = null,
    val banRidings: List<FenceInfoDto>? = null,
) {
    fun toDomain(): FenceBundle = FenceBundle(
        serviceAreas = serviceAreas.orEmpty().mapNotNull { it.toDomain(FenceKind.ServiceArea) },
        parkings = parkings.orEmpty().mapNotNull { it.toDomain(FenceKind.Parking) },
        noParkings = noParkings.orEmpty().mapNotNull { it.toDomain(FenceKind.NoParking) },
    )
}

@Serializable
data class FenceInfoDto(
    val id: String = "",
    val name: String = "",
    val pointList: List<List<Double>>? = null,
) {
    fun toDomain(kind: FenceKind): FencePolygon? {
        val pts = pointList.orEmpty().mapNotNull { pair ->
            if (pair.size < 2) return@mapNotNull null
            // Legacy geoJson style: [lng, lat]
            GeoLatLng(lat = pair[1], lng = pair[0])
        }
        if (pts.size < 3) return null
        return FencePolygon(
            id = id.ifBlank { name.ifBlank { kind.name } },
            name = name,
            points = pts,
            kind = kind,
        )
    }
}
