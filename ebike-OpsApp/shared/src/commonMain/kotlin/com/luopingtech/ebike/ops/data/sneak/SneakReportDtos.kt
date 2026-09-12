package com.luopingtech.ebike.ops.data.sneak

import com.luopingtech.ebike.ops.domain.model.SneakReportRecord
import com.luopingtech.ebike.ops.domain.model.SneakReportType
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class SneakTypeDto(
    val id: String? = null,
    val content: String? = null,
    val type: Int? = null,
) {
    fun toDomain(): SneakReportType = SneakReportType(
        id = id.orEmpty(),
        name = content.orEmpty(),
        type = type ?: 0,
    )
}

@Serializable
data class SneakReportTypePayloadDto(
    val type: List<String> = emptyList(),
    @SerialName("other_type")
    val otherType: List<String> = emptyList(),
)

@Serializable
data class SneakHistoryPageDto(
    val count: Int? = null,
    @SerialName("total_page")
    val totalPage: Int? = null,
    val result: List<SneakRecordDto> = emptyList(),
)

@Serializable
data class SneakDetailWrapDto(
    val list: SneakRecordDto? = null,
)

@Serializable
data class SneakRecordDto(
    val id: Long? = null,
    @SerialName("service_id")
    val serviceId: Long? = null,
    @SerialName("reported_user_pin")
    val reportedUserPin: String? = null,
    @SerialName("reported_user_name")
    val reportedUserName: String? = null,
    @SerialName("reported_user_phone")
    val reportedUserPhone: String? = null,
    val imei: String? = null,
    @SerialName("itin_id")
    val itinId: String? = null,
    @SerialName("car_id")
    val carId: String? = null,
    @SerialName("report_type")
    val reportType: SneakReportTypePayloadDto? = null,
    val description: String? = null,
    val imgs: List<String>? = null,
    @SerialName("report_man_phone")
    val reportManPhone: String? = null,
    @SerialName("report_man_name")
    val reportManName: String? = null,
    @SerialName("created_at")
    val createdAt: String? = null,
    val remark: String? = null,
    @SerialName("handle_type")
    val handleType: Int? = null,
    @SerialName("check_at")
    val checkAt: String? = null,
    @SerialName("check_result")
    val checkResult: Int? = null,
) {
    fun toDomain(typeNameLookup: (String) -> String = { it }): SneakReportRecord {
        val typeIds = reportType?.type.orEmpty()
        val other = reportType?.otherType.orEmpty().firstOrNull().orEmpty()
        return SneakReportRecord(
            id = id?.toString().orEmpty(),
            carId = carId.orEmpty(),
            reportedUserName = reportedUserName.orEmpty(),
            reportedUserPhone = reportedUserPhone.orEmpty(),
            reportedUserPin = reportedUserPin.orEmpty(),
            itinId = itinId.orEmpty(),
            typeLabels = typeIds.map(typeNameLookup),
            otherType = other,
            description = description.orEmpty(),
            photoUrls = imgs.orEmpty().filter { it.isNotBlank() },
            createdAt = createdAt.orEmpty(),
            checkResult = checkResult ?: 0,
            remark = remark.orEmpty(),
            handleType = handleType ?: 0,
            serviceId = serviceId?.toString().orEmpty(),
        )
    }
}
