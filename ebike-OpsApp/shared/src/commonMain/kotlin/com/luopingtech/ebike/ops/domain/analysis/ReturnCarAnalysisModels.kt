package com.luopingtech.ebike.ops.domain.analysis

import com.luopingtech.ebike.ops.domain.model.GeoLatLng

enum class ReturnCarStatusFilter {
    /** 正常还车（进站） */
    Normal,
    /** 异常还车（未进站） */
    Abnormal,
    All,
}

data class ReturnCarPoint(
    val lat: Double,
    val lng: Double,
    /** true = 异常（p） */
    val abnormal: Boolean,
) {
    fun toLatLng(): GeoLatLng = GeoLatLng(lat = lat, lng = lng)
}

data class ReturnCarAnalyzeResult(
    val total: Int = 0,
    val normalCount: Int = 0,
    val abnormalCount: Int = 0,
    val points: List<ReturnCarPoint> = emptyList(),
) {
    fun visible(filter: ReturnCarStatusFilter): List<ReturnCarPoint> = when (filter) {
        ReturnCarStatusFilter.Normal -> points.filter { !it.abnormal }
        ReturnCarStatusFilter.Abnormal -> points.filter { it.abnormal }
        ReturnCarStatusFilter.All -> points
    }
}
