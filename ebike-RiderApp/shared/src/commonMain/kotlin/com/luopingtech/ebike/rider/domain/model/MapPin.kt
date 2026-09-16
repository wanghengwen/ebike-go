package com.luopingtech.ebike.rider.domain.model

/** Lightweight pin for host map rendering. Shared does not draw vendor maps. */
data class MapPin(
    val id: String,
    val lat: Double,
    val lng: Double,
    val title: String,
    val subtitle: String = "",
    val restBattery: Int = 0,
    val ridingState: Int? = null,
    /** 1 = single vehicle; >1 = cluster marker. */
    val memberCount: Int = 1,
    val memberIds: List<String> = emptyList(),
) {
    val isCluster: Boolean get() = memberCount > 1
}
