package com.luopingtech.ebike.ops.data.order

import com.luopingtech.ebike.ops.data.trajectory.TrackPointDto
import com.luopingtech.ebike.ops.domain.model.LastOrder
import com.luopingtech.ebike.ops.domain.order.OrderRecord
import com.luopingtech.ebike.ops.domain.order.OrderUserDetail
import com.luopingtech.ebike.ops.domain.order.OrderUserPageItem
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonPrimitive

@Serializable
data class LastOrderDto(
    val carId: String? = null,
    val startLat: Double? = null,
    val startLng: Double? = null,
    val endLat: Double? = null,
    val endLng: Double? = null,
    val deviceTrajectory: List<TrackPointDto>? = null,
    val id: String? = null,
    val userPin: String? = null,
    val phone: String? = null,
    val userPhone: String? = null,
    val startTime: String? = null,
    val endTime: String? = null,
    val userName: String? = null,
    val originCost: Long? = null,
    val payCost: Long? = null,
    val mile: Long? = null,
    val ridingTime: JsonElement? = null,
    val izPaid: Int? = null,
    val payTime: String? = null,
    val hasPaid: Long? = null,
    val dispatchCost: Long? = null,
    val helmetPenalty: Long? = null,
    val carState: Int? = null,
) {
    fun toDomain(): LastOrder {
        val track = deviceTrajectory.orEmpty().mapNotNull { it.toDomain() }
        return LastOrder(
            carId = carId.orEmpty(),
            startLat = startLat ?: 0.0,
            startLng = startLng ?: 0.0,
            endLat = endLat,
            endLng = endLng,
            trajectory = track,
            id = id.orEmpty(),
            userPin = userPin.orEmpty(),
            userPhone = phone?.takeIf { it.isNotBlank() } ?: userPhone.orEmpty(),
            startTime = startTime.orEmpty(),
            endTime = endTime.orEmpty(),
        )
    }

    fun toOrderRecord(): OrderRecord = OrderRecord(
        id = id.orEmpty(),
        carId = carId.orEmpty(),
        phone = phone?.takeIf { it.isNotBlank() } ?: userPhone.orEmpty(),
        userPin = userPin.orEmpty(),
        userName = userName.orEmpty(),
        originCost = originCost,
        payCost = payCost,
        mile = mile,
        ridingTimeRaw = ridingTime.toFlexibleString(),
        izPaid = izPaid,
        startLat = startLat,
        startLng = startLng,
        endLat = endLat,
        endLng = endLng,
        startTime = startTime.orEmpty(),
        endTime = endTime.orEmpty(),
        payTime = payTime.orEmpty(),
        hasPaid = hasPaid,
        dispatchCost = dispatchCost,
        helmetPenalty = helmetPenalty,
        carState = carState,
    )
}

@Serializable
data class OrderPageDto(
    val list: List<LastOrderDto>? = null,
    val count: String? = null,
    val pageNum: Int = 1,
    val pageSize: Int = 10,
)

@Serializable
data class OrderUserDetailDto(
    val pin: String? = null,
    val authName: String? = null,
    val phone: String? = null,
    val createdAt: String? = null,
    val izAuth: Boolean? = null,
    val ridingState: Int? = null,
    val balance: Long? = null,
    val serviceName: String? = null,
) {
    fun toDomain(): OrderUserDetail = OrderUserDetail(
        pin = pin.orEmpty(),
        authName = authName.orEmpty(),
        phone = phone.orEmpty(),
        createdAt = createdAt.orEmpty(),
        izAuth = izAuth,
        ridingState = ridingState,
        balance = balance,
        serviceName = serviceName.orEmpty(),
    )
}

@Serializable
data class OrderUserPageItemDto(
    val pin: String? = null,
    val authName: String? = null,
    val phone: String? = null,
    val izAuth: Boolean? = null,
    val ridingState: Int? = null,
    val createdAt: String? = null,
    val balance: Long? = null,
) {
    fun toDomain(): OrderUserPageItem = OrderUserPageItem(
        pin = pin.orEmpty(),
        authName = authName.orEmpty(),
        phone = phone.orEmpty(),
        izAuth = izAuth,
        ridingState = ridingState,
        createdAt = createdAt.orEmpty(),
        balance = balance,
    )
}

@Serializable
data class OrderUserPageDto(
    val list: List<OrderUserPageItemDto>? = null,
    val count: Int = 0,
    val pageNum: Int = 1,
    val pageSize: Int = 15,
)

private fun JsonElement?.toFlexibleString(): String = when (this) {
    null -> ""
    is JsonPrimitive -> content
    else -> toString()
}
