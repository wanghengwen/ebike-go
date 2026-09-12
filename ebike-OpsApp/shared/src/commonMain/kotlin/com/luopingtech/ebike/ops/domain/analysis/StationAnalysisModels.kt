package com.luopingtech.ebike.ops.domain.analysis

enum class StationOptStateFilter {
    All,
    Operating,
    Stopped,
}

enum class StationSortOrder(val apiValue: String?) {
    None(null),
    CanRentDesc("canRent_desc"),
    CanRentAsc("canRent_asc"),
    IdleDesc("idle_desc"),
    IdleAsc("idle_asc"),
    SiteOutDesc("siteOut_desc"),
    SiteOutAsc("siteOut_asc"),
    DdMissDesc("ddMissOrder_desc"),
    DdMissAsc("ddMissOrder_asc"),
}

data class StationTag(
    val id: String,
    val name: String,
)

data class StationAnalyzeItem(
    val parkingId: String,
    val serviceId: String,
    val name: String,
    val operating: Boolean,
    val tags: List<StationTag> = emptyList(),
    val canRent: Int = 0,
    val idle: Int = 0,
    val siteOut: Int = 0,
    val ddMissOrder: Int = 0,
    val booking: Int = 0,
    val operation: Int = 0,
)

data class StationAnalyzePage(
    val items: List<StationAnalyzeItem>,
    val total: Int,
    val pageNum: Int,
    val pageSize: Int,
) {
    val hasMore: Boolean get() = items.size < total && items.isNotEmpty()
}

/** 详情：实时三卡 + 24h 序列 + 闲置分档柱。 */
data class StationAnalyzeDetail(
    val parkingId: String,
    val serviceId: String,
    val canRent: Int = 0,
    val booking: Int = 0,
    val operation: Int = 0,
    val canRentHours: List<Int> = emptyList(),
    val bookingHours: List<Int> = emptyList(),
    val operationHours: List<Int> = emptyList(),
    val alarmHours: List<Int> = emptyList(),
    val faultHours: List<Int> = emptyList(),
    val siteOutHours: List<Int> = emptyList(),
    val ddMissHours: List<Int> = emptyList(),
    /** 闲置分档：1~3h … 48h以上，共 6 档。 */
    val idleBuckets: List<Int> = emptyList(),
)
