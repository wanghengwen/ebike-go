package com.luopingtech.ebike.ops.data.movecar

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.FreeMoveCar
import com.luopingtech.ebike.ops.domain.model.TeamWorker
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.Serializable
import kotlinx.serialization.builtins.ListSerializer
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.add
import kotlinx.serialization.json.buildJsonArray
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.put

@Serializable
data class FreeMoveCarDto(
    val carId: String = "",
    val imei: String = "",
    val restBattery: Int? = 0,
    val battery: Int? = null,
    val state: Int = 1,
    val izFinish: Boolean = false,
) {
    fun toDomain(): FreeMoveCar = FreeMoveCar(
        carId = carId,
        imei = imei,
        restBattery = restBattery ?: battery ?: 0,
        state = state,
        izFinish = izFinish,
    )
}

/**
 * Free move-car (????????�?:
 * start_permission / start / batch_list / end / remove.
 * Finish photos use field [pictures] (task finish uses [photo]).
 */
class FreeMoveCarApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    suspend fun checkPermission(carId: String, serviceId: String): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/move_car/start_permission",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("carId", carId)
            put("serviceId", serviceId)
        }
        return signedApi.postUnit(
            path = "business/ebike-operation/move_car/start_permission",
            bodyJson = body,
        )
    }

    suspend fun start(
        carIds: List<String>,
        serviceId: String,
        izPushCar: Boolean = false,
    ): OpsResult<List<FreeMoveCar>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/move_car/start",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("serviceId", serviceId)
            put("izPushCar", izPushCar)
            put("carIds", buildJsonArray { carIds.forEach { add(JsonPrimitive(it)) } })
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-operation/move_car/start",
                bodyJson = body,
                deserializer = ListSerializer(FreeMoveCarDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.map { it.toDomain() })
            is OpsResult.Err -> result
        }
    }

    suspend fun list(serviceId: String): OpsResult<List<FreeMoveCar>> {
        // Legacy MoveVehicleRepositoryV2.getAllVehicle: common params only (no serviceId).
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/move_car/batch_list",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            // Keep serviceId when backend variants require it; legacy omits it.
            if (serviceId.isNotBlank()) {
                put("serviceId", serviceId)
            }
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-operation/move_car/batch_list",
                bodyJson = body,
                deserializer = ListSerializer(FreeMoveCarDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.map { it.toDomain() })
            is OpsResult.Err -> result
        }
    }

    suspend fun finish(
        carIds: List<String>,
        serviceId: String,
        phone: String,
        pictures: List<String> = emptyList(),
        remark: String? = null,
        teamWorkers: List<TeamWorker> = emptyList(),
    ): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/move_car/end",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            // Legacy DevicesRepositoryV2.moveBikesBatchEnd (MoveCarActivity 关锁/完成):
            // serviceId + carIds (+ pictures/remark). phone/teamWorker 来自拍照审核路径。
            put("serviceId", serviceId)
            put("carIds", buildJsonArray { carIds.forEach { add(JsonPrimitive(it)) } })
            if (phone.isNotBlank()) {
                put("phone", phone)
            }
            if (pictures.isNotEmpty()) {
                put("pictures", buildJsonArray { pictures.forEach { add(JsonPrimitive(it)) } })
            }
            if (!remark.isNullOrBlank()) {
                put("remark", remark)
            }
            // Legacy MoveVehicleRepositoryV2: teamWorker = list of JSON *strings* {"name","phone"}.
            if (teamWorkers.isNotEmpty()) {
                put(
                    "teamWorker",
                    buildJsonArray {
                        teamWorkers.forEach { worker ->
                            val payload = buildJsonObject {
                                put("name", worker.name)
                                put("phone", worker.phone)
                            }.toString()
                            add(JsonPrimitive(payload))
                        }
                    },
                )
            }
        }
        return signedApi.postUnit(
            path = "business/ebike-operation/move_car/end",
            bodyJson = body,
        )
    }

    suspend fun remove(carId: String, serviceId: String): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/move_car/remove",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("carId", carId)
            put("serviceId", serviceId)
        }
        return signedApi.postUnit(
            path = "business/ebike-operation/move_car/remove",
            bodyJson = body,
        )
    }
}
