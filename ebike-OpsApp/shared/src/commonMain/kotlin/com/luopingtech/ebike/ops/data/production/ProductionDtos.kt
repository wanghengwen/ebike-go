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
    val carId: String? = null,
    val carNo: String? = null,
    val brand: String? = null,
    val serviceId: String? = null,
    val serviceName: String? = null,
) {
    fun toDomain(fallbackCarId: String = ""): ShelfCheckResult = ShelfCheckResult(
        carId = carId?.takeIf { it.isNotBlank() } ?: fallbackCarId,
        carNo = carNo.orEmpty(),
        brand = brand.orEmpty(),
        serviceId = serviceId.orEmpty(),
        serviceName = serviceName.orEmpty(),
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
