package com.luopingtech.ebike.rider.domain.model

data class GeoLatLng(
    val lat: Double,
    val lng: Double,
)

data class FencePolygon(
    val id: String,
    val name: String,
    val points: List<GeoLatLng>,
    val kind: FenceKind = FenceKind.ServiceArea,
)

enum class FenceKind {
    ServiceArea,
    Parking,
    NoParking,
}

data class TrackPoint(
    val lat: Double,
    val lng: Double,
    val timestamp: Long = 0L,
)
