package com.luopingtech.ebike.ops.domain.order

import com.luopingtech.ebike.ops.domain.model.MapPin
import com.luopingtech.ebike.ops.domain.model.MapPinIcon
import com.luopingtech.ebike.ops.domain.model.TrackPoint

/** 轨迹回放车标；不参与地图 fit，避免拖进度条时镜头跟着跳。 */
const val ORDER_PLAYBACK_PIN_ID = "order-playback"

fun isValidMapLatLng(lat: Double?, lng: Double?): Boolean {
    if (lat == null || lng == null) return false
    return lat != 0.0 || lng != 0.0
}

/**
 * 对齐 OrderHistoryActivity.updatePlaybackUI：按轨迹时间戳插值；
 * 时间跨度为 0 时按点序插值，保证进度条仍能带动「骑」标。
 */
fun interpolateTrackPoint(points: List<TrackPoint>, progress: Float): TrackPoint? {
    if (points.isEmpty()) return null
    if (points.size == 1) return points.first()
    val p = progress.coerceIn(0f, 1f)
    val minT = points.first().timestamp
    val maxT = points.last().timestamp
    val span = maxT - minT
    if (span <= 0L) {
        val exact = p * points.lastIndex
        val i = exact.toInt().coerceIn(0, points.lastIndex - 1)
        val frac = (exact - i).toDouble()
        val a = points[i]
        val b = points[i + 1]
        return TrackPoint(
            lat = a.lat + (b.lat - a.lat) * frac,
            lng = a.lng + (b.lng - a.lng) * frac,
            timestamp = a.timestamp,
            speed = a.speed,
            course = a.course,
        )
    }
    val target = minT + (span * p.toDouble()).toLong()
    var p1 = points.first()
    var p2 = points.last()
    for (i in 0 until points.lastIndex) {
        if (target >= points[i].timestamp && target <= points[i + 1].timestamp) {
            p1 = points[i]
            p2 = points[i + 1]
            break
        }
    }
    if (target < points.first().timestamp) {
        p1 = points.first()
        p2 = points.first()
    } else if (target > points.last().timestamp) {
        p1 = points.last()
        p2 = points.last()
    }
    val t1 = p1.timestamp.toDouble()
    val t2 = p2.timestamp.toDouble()
    val fraction = if (t2 > t1) (target - t1) / (t2 - t1) else 0.0
    return TrackPoint(
        lat = p1.lat + (p2.lat - p1.lat) * fraction,
        lng = p1.lng + (p2.lng - p1.lng) * fraction,
        timestamp = target,
        speed = p1.speed,
        course = p1.course,
    )
}

fun formatOrderTrackClock(epoch: Long): String {
    if (epoch <= 0L) return "--:--:--"
    val ms = if (epoch < 1_000_000_000_000L) epoch * 1000 else epoch
    val totalSec = (ms / 1000) % 86_400
    val h = totalSec / 3600
    val m = (totalSec % 3600) / 60
    val s = totalSec % 60
    return "${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}"
}

/**
 * 对齐 OrderHistoryActivity + TrackDataHelp：
 * 轨迹首点 `btn_trajectory_origin`、末点 `btn_trajectory_end`、
 * 用车人起终点 `icon_user_start/end`、进度条位置 `icon_vehicle_riding`。
 */
fun buildOrderHistoryPins(
    order: OrderRecord?,
    track: List<TrackPoint>,
    playbackProgress: Float = 0f,
): List<MapPin> {
    if (order == null && track.isEmpty()) return emptyList()
    return buildList {
        val first = track.firstOrNull()
        val last = track.lastOrNull()
        if (first != null && isValidMapLatLng(first.lat, first.lng)) {
            add(
                MapPin(
                    id = "order-track-origin",
                    lat = first.lat,
                    lng = first.lng,
                    title = "起",
                    memberCount = 1,
                    memberIds = listOf("order-track-origin"),
                    icon = MapPinIcon.TrackOrigin,
                ),
            )
        }
        if (last != null && isValidMapLatLng(last.lat, last.lng) && track.size >= 2) {
            add(
                MapPin(
                    id = "order-track-end",
                    lat = last.lat,
                    lng = last.lng,
                    title = "终",
                    memberCount = 1,
                    memberIds = listOf("order-track-end"),
                    icon = MapPinIcon.TrackEnd,
                ),
            )
        }
        val startLat = order?.startLat
        val startLng = order?.startLng
        if (isValidMapLatLng(startLat, startLng)) {
            add(
                MapPin(
                    id = "order-user-start",
                    lat = startLat!!,
                    lng = startLng!!,
                    title = "起",
                    memberCount = 1,
                    memberIds = listOf("order-user-start"),
                    icon = MapPinIcon.UserStart,
                ),
            )
        }
        val endLat = order?.endLat
        val endLng = order?.endLng
        if (isValidMapLatLng(endLat, endLng)) {
            add(
                MapPin(
                    id = "order-user-end",
                    lat = endLat!!,
                    lng = endLng!!,
                    title = "终",
                    memberCount = 1,
                    memberIds = listOf("order-user-end"),
                    icon = MapPinIcon.UserEnd,
                ),
            )
        }
        val play = interpolateTrackPoint(track, playbackProgress)
        if (play != null && isValidMapLatLng(play.lat, play.lng) && track.size >= 2) {
            add(
                MapPin(
                    id = ORDER_PLAYBACK_PIN_ID,
                    lat = play.lat,
                    lng = play.lng,
                    title = "骑",
                    memberCount = 1,
                    memberIds = listOf(ORDER_PLAYBACK_PIN_ID),
                    icon = MapPinIcon.VehicleRiding,
                ),
            )
        }
    }
}
