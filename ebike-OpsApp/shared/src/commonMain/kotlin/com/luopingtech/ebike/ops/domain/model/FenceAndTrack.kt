package com.luopingtech.ebike.ops.domain.model

data class GeoLatLng(
    val lat: Double,
    val lng: Double,
)

data class FencePolygon(
    val id: String,
    val name: String,
    /** Closed ring of [lng,lat] vertices (legacy pointList). */
    val points: List<GeoLatLng>,
    val kind: FenceKind = FenceKind.ServiceArea,
    /** Vehicles currently associated (legacy `carCount`, often used on parking markers). */
    val carCount: Int = 0,
    /** Occupancy / capacity for parking zones (legacy current/maxParkingNumber). */
    val currentParkingNumber: Int = 0,
    val maxParkingNumber: Int = 0,
    /** null = unknown; true = 运营中. */
    val izEnable: Boolean? = null,
    val address: String = "",
)

enum class FenceKind {
    ServiceArea,
    Parking,
    NoParking,
    Maintain,
    BanRiding,
}

data class FenceBundle(
    val serviceAreas: List<FencePolygon> = emptyList(),
    val parkings: List<FencePolygon> = emptyList(),
    val noParkings: List<FencePolygon> = emptyList(),
) {
    val all: List<FencePolygon>
        get() = serviceAreas + parkings + noParkings
}

data class TrackPoint(
    val lat: Double,
    val lng: Double,
    val timestamp: Long = 0L,
    val speed: Float = 0f,
    val course: Float = 0f,
)

/** Last order for vehicle detail map: embedded trajectory + start/end for near fences. */
data class LastOrder(
    val carId: String,
    val startLat: Double = 0.0,
    val startLng: Double = 0.0,
    val endLat: Double? = null,
    val endLng: Double? = null,
    val trajectory: List<TrackPoint> = emptyList(),
    /** Order / itinerary id — required for sneak report submit (`itin_id`). */
    val id: String = "",
    val userPin: String = "",
    val userPhone: String = "",
    val startTime: String = "",
    val endTime: String = "",
) {
    /** Locations for getNearFenceByLocations (legacy: start + end-or-last-track-point). */
    fun nearFenceLocations(): List<GeoLatLng> {
        val start = if (startLat != 0.0 || startLng != 0.0) {
            GeoLatLng(startLat, startLng)
        } else {
            null
        }
        val endFromOrder = if (endLat != null && endLng != null && (endLat != 0.0 || endLng != 0.0)) {
            GeoLatLng(endLat, endLng)
        } else {
            null
        }
        val end = endFromOrder
            ?: trajectory.lastOrNull()?.let { GeoLatLng(it.lat, it.lng) }
        return listOfNotNull(start, end).distinct()
    }
}
