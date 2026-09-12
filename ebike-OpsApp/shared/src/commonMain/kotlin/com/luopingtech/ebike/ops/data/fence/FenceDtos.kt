package com.luopingtech.ebike.ops.data.fence

import com.luopingtech.ebike.ops.data.analysis.FlexibleIntSerializer
import com.luopingtech.ebike.ops.data.analysis.FlexibleStringSerializer
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
        serviceAreas = serviceAreas.orEmpty().map { it.toDomain(FenceKind.ServiceArea) },
        parkings = parkings.orEmpty().map { it.toDomain(FenceKind.Parking) },
        noParkings = noParkings.orEmpty().map { it.toDomain(FenceKind.NoParking) },
    )
}

@Serializable
data class FenceInfoDto(
    @Serializable(with = FlexibleStringSerializer::class)
    val id: String = "",
    val name: String = "",
    val pointList: List<List<Double>>? = null,
    @Serializable(with = FlexibleIntSerializer::class)
    val carCount: Int = 0,
    @Serializable(with = FlexibleIntSerializer::class)
    val currentParkingNumber: Int = 0,
    @Serializable(with = FlexibleIntSerializer::class)
    val maxParkingNumber: Int = 0,
    val izEnable: Boolean? = null,
    val address: String? = null,
) {
    fun toDomain(kind: FenceKind): FencePolygon {
        val pts = pointList.orEmpty().mapNotNull { pair ->
            if (pair.size < 2) return@mapNotNull null
            // Legacy geoJson style: [lng, lat]
            GeoLatLng(lat = pair[1], lng = pair[0])
        }
        return FencePolygon(
            id = id.ifBlank { name.ifBlank { kind.name } },
            name = name.ifBlank { id.ifBlank { kind.name } },
            points = pts,
            kind = kind,
            carCount = carCount,
            currentParkingNumber = currentParkingNumber,
            maxParkingNumber = maxParkingNumber,
            izEnable = izEnable,
            address = address.orEmpty(),
        )
    }
}
