package com.luopingtech.ebike.ops.data.production

import com.luopingtech.ebike.ops.domain.model.BindVehicleInfo
import com.luopingtech.ebike.ops.domain.model.SaddleOverloadContact
import com.luopingtech.ebike.ops.domain.model.ShelfCheckResult
import kotlinx.serialization.Serializable

@Serializable
data class BindVehicleInfoDto(
    val id: String = "",
    val carId: String = "",
    val carNo: String = "",
    val brand: String = "",
    val model: String = "",
    val imei: String = "",
    val helmet: String = "",
    val serviceId: String = "",
    val bindTime: String = "",
) {
    fun toDomain(): BindVehicleInfo = BindVehicleInfo(
        id = id,
        carId = carId,
        carNo = carNo,
        brand = brand,
        model = model,
        imei = imei,
        helmet = helmet,
        serviceId = serviceId,
        bindTime = bindTime,
    )
}

@Serializable
data class ShelfCheckDto(
    val id: String = "",
    val carId: String = "",
    val carNo: String = "",
    val brand: String = "",
    val serviceId: String = "",
    val serviceName: String = "",
) {
    fun toDomain(): ShelfCheckResult = ShelfCheckResult(
        carId = carId,
        carNo = carNo,
        brand = brand,
        serviceId = serviceId,
        serviceName = serviceName,
    )
}

@Serializable
data class SaddleOverloadContactDto(
    val frontSaddleContact: Int = 0,
    val centSaddleContact: Int = 0,
    val backSaddleContact: Int = 0,
) {
    fun toDomain(): SaddleOverloadContact = SaddleOverloadContact(
        frontSaddleContact = frontSaddleContact,
        centSaddleContact = centSaddleContact,
        backSaddleContact = backSaddleContact,
    )
}
