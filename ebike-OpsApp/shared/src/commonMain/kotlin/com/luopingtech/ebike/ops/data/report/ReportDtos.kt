package com.luopingtech.ebike.ops.data.report

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.domain.model.FaultReportRecord
import com.luopingtech.ebike.ops.domain.model.RepairType
import kotlinx.serialization.Serializable

@Serializable
data class RepairTypeDto(
    val id: String = "",
    val content: String = "",
    val type: Int = 0,
) {
    fun toDomain(): RepairType = RepairType(
        id = id,
        name = content,
        type = type,
    )
}

/** Page from fix/task/page — repair/add writes TaskFixTicket rows. */
@Serializable
data class RepairHistoryPageDto(
    val list: List<RepairHistoryDto> = emptyList(),
    val count: Int? = null,
)

@Serializable
data class RepairHistoryDto(
    val id: Long? = null,
    val carId: String = "",
    val fixReason: String = "",
    val nameExtraInfo: List<String> = emptyList(),
    val photo: List<String> = emptyList(),
    val izStop: Boolean = false,
    val createdAt: String = "",
    val state: Int = 0,
    val checkResult: Int? = null,
) {
    fun toDomain(): FaultReportRecord {
        val base = when (state) {
            0 -> Strings.t(Str.TaskStatePending)
            1 -> Strings.t(Str.TaskStateInProgress)
            2 -> Strings.t(Str.TaskStateDone)
            3 -> Strings.t(Str.TaskStateClosed)
            else -> Strings.t(Str.TaskStatePending)
        }
        val statusLabel = if (izStop) "$base · ${Strings.t(Str.WorkOrderStopped)}" else base
        return FaultReportRecord(
            id = id?.toString().orEmpty(),
            carId = carId,
            fixReason = fixReason,
            typeNames = nameExtraInfo.filter { it.isNotBlank() },
            photoUrls = photo.filter { it.isNotBlank() },
            izStop = izStop,
            createdAt = createdAt,
            statusLabel = statusLabel,
        )
    }
}
