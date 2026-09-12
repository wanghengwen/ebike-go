package com.luopingtech.ebike.ops.data.analysis

import com.luopingtech.ebike.ops.domain.analysis.StationAnalyzeDetail
import com.luopingtech.ebike.ops.domain.analysis.StationAnalyzeItem
import com.luopingtech.ebike.ops.domain.analysis.StationAnalyzePage
import com.luopingtech.ebike.ops.domain.analysis.StationTag
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class StationTagDto(
    @Serializable(with = FlexibleStringSerializer::class)
    val tagId: String = "",
    val tagName: String = "",
) {
    fun toDomain(): StationTag = StationTag(id = tagId, name = tagName)
}

@Serializable
data class StationParkingInfoDto(
    @Serializable(with = FlexibleStringSerializer::class)
    val id: String = "",
    val name: String? = null,
    val izEnable: Boolean? = null,
    val refTags: List<StationTagDto> = emptyList(),
)

@Serializable
data class StationAnalyzeItemDto(
    val parkingInfo: StationParkingInfoDto? = null,
    @Serializable(with = FlexibleStringSerializer::class)
    val serviceId: String = "",
    @Serializable(with = FlexibleStringSerializer::class)
    val parkingId: String = "",
    @Serializable(with = FlexibleIntSerializer::class)
    val canRent: Int = 0,
    @Serializable(with = FlexibleIntSerializer::class)
    val idle: Int = 0,
    @Serializable(with = FlexibleIntSerializer::class)
    val siteOut: Int = 0,
    @Serializable(with = FlexibleIntSerializer::class)
    val ddMissOrder: Int = 0,
    @Serializable(with = FlexibleIntSerializer::class)
    val booking: Int = 0,
    @Serializable(with = FlexibleIntSerializer::class)
    val operation: Int = 0,
) {
    fun toDomain(): StationAnalyzeItem {
        val pid = parkingId.ifBlank { parkingInfo?.id.orEmpty() }
        return StationAnalyzeItem(
            parkingId = pid,
            serviceId = serviceId,
            name = parkingInfo?.name?.takeIf { it.isNotBlank() } ?: pid,
            operating = parkingInfo?.izEnable == true,
            tags = parkingInfo?.refTags.orEmpty().map { it.toDomain() }.filter { it.id.isNotBlank() },
            canRent = canRent,
            idle = idle,
            siteOut = siteOut,
            ddMissOrder = ddMissOrder,
            booking = booking,
            operation = operation,
        )
    }
}

@Serializable
data class StationAnalyzePageDto(
    val list: List<StationAnalyzeItemDto> = emptyList(),
    @Serializable(with = FlexibleIntSerializer::class)
    val count: Int = 0,
    @Serializable(with = FlexibleIntSerializer::class)
    val pageNum: Int = 1,
    @Serializable(with = FlexibleIntSerializer::class)
    val pageSize: Int = 10,
) {
    fun toDomain(): StationAnalyzePage = StationAnalyzePage(
        items = list.map { it.toDomain() },
        total = count,
        pageNum = pageNum,
        pageSize = pageSize,
    )
}

@Serializable
data class StationCarMapDto(
    val canRent: List<Int> = emptyList(),
    val booking: List<Int> = emptyList(),
    val operation: List<Int> = emptyList(),
    val alarm: List<Int> = emptyList(),
    val fault: List<Int> = emptyList(),
    val idle: List<Int> = emptyList(),
    val siteOut: List<Int> = emptyList(),
    val ddMissOrder: List<Int> = emptyList(),
)

@Serializable
data class StationAnalyzeOneDto(
    @Serializable(with = FlexibleStringSerializer::class)
    val parkingId: String = "",
    @Serializable(with = FlexibleStringSerializer::class)
    val serviceId: String = "",
    @Serializable(with = FlexibleIntSerializer::class)
    val canRent: Int = 0,
    @Serializable(with = FlexibleIntSerializer::class)
    val booking: Int = 0,
    @Serializable(with = FlexibleIntSerializer::class)
    val operation: Int = 0,
    @SerialName("carMap")
    val carMap: StationCarMapDto? = null,
) {
    fun toDomain(): StationAnalyzeDetail {
        val map = carMap
        return StationAnalyzeDetail(
            parkingId = parkingId,
            serviceId = serviceId,
            canRent = canRent,
            booking = booking,
            operation = operation,
            canRentHours = map?.canRent.orEmpty(),
            bookingHours = map?.booking.orEmpty(),
            operationHours = map?.operation.orEmpty(),
            alarmHours = map?.alarm.orEmpty(),
            faultHours = map?.fault.orEmpty(),
            siteOutHours = map?.siteOut.orEmpty(),
            ddMissHours = map?.ddMissOrder.orEmpty(),
            idleBuckets = map?.idle.orEmpty(),
        )
    }
}
