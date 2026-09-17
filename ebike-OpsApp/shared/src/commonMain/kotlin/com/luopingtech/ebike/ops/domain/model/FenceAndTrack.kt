package com.luopingtech.ebike.ops.domain.model

import com.luopingtech.ebike.ops.domain.analysis.StationTag
import kotlin.math.PI
import kotlin.math.atan2
import kotlin.math.cos
import kotlin.math.sin
import kotlin.math.sqrt

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
    /** Legacy centerLat / centerLng；缺省时用顶点重心。 */
    val centerLat: Double? = null,
    val centerLng: Double? = null,
    /** 面积（㎡），禁停底卡用。 */
    val area: Double? = null,
    /** 停车区标签（legacy refTags）。 */
    val tags: List<StationTag> = emptyList(),
) {
    fun centerOrCentroid(): GeoLatLng? {
        val lat = centerLat
        val lng = centerLng
        if (lat != null && lng != null && (lat != 0.0 || lng != 0.0)) {
            return GeoLatLng(lat, lng)
        }
        if (points.isEmpty()) return null
        return GeoLatLng(
            lat = points.map { it.lat }.average(),
            lng = points.map { it.lng }.average(),
        )
    }
}

enum class FenceKind {
    ServiceArea,
    Parking,
    NoParking,
    Maintain,
    BanRiding,
}

/**
 * 对齐遗留 FenceInfoModel.optionIcon：禁停 / 停用停车 / 普通停车带杆图钉。
 * （正规站点 icon_parking_function_unselect 需 tbeacon/rfid 等字段，后续再补。）
 */
fun FencePolygon.fenceMapPinIcon(): MapPinIcon = when (kind) {
    FenceKind.NoParking -> MapPinIcon.NoParking
    FenceKind.Parking -> if (izEnable == false) {
        MapPinIcon.ParkingHidden
    } else {
        MapPinIcon.Parking
    }
    else -> MapPinIcon.Default
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

/** Haversine distance in meters. */
fun metersBetween(a: GeoLatLng, b: GeoLatLng): Double {
    val earth = 6_371_000.0
    fun rad(deg: Double) = deg * PI / 180.0
    val dLat = rad(b.lat - a.lat)
    val dLng = rad(b.lng - a.lng)
    val lat1 = rad(a.lat)
    val lat2 = rad(b.lat)
    val h = sin(dLat / 2) * sin(dLat / 2) +
        cos(lat1) * cos(lat2) * sin(dLng / 2) * sin(dLng / 2)
    return 2 * earth * atan2(sqrt(h), sqrt(1 - h))
}
