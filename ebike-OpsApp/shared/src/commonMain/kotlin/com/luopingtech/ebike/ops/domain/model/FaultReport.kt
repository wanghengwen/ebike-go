package com.luopingtech.ebike.ops.domain.model

data class RepairType(
    val id: String,
    val name: String,
    val type: Int = 0,
)

data class FaultReportDraft(
    val carId: String = "",
    val fixReason: String = "",
    val typeIds: List<String> = emptyList(),
    val typeNames: List<String> = emptyList(),
    val photoUrls: List<String> = emptyList(),
    /** null = unset, true = stop ops, false = keep running. */
    val izStop: Boolean? = null,
)

data class FaultReportRecord(
    val id: String,
    val carId: String,
    val fixReason: String,
    val typeNames: List<String> = emptyList(),
    val photoUrls: List<String> = emptyList(),
    val izStop: Boolean = false,
    val createdAt: String = "",
    val statusLabel: String = "",
)
