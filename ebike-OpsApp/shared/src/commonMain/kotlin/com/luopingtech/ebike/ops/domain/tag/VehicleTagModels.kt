package com.luopingtech.ebike.ops.domain.tag

data class VehicleTagType(
    val id: String,
    val name: String,
)

data class VehicleTagRecord(
    val id: String,
    val carId: String,
    val typeName: String,
    val createdAt: String = "",
    val operatorName: String = "",
)
