package com.luopingtech.ebike.ops.data.movecar

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.BatchMoveChild
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.ExperimentalSerializationApi
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonNames
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.buildJsonArray
import kotlinx.serialization.json.put

@Serializable
data class BatchMoveListDto(
    val sum: Int? = null,
    val list: List<BatchMoveChildDto> = emptyList(),
)

@OptIn(ExperimentalSerializationApi::class)
@Serializable
data class BatchMoveChildDto(
    val carId: String = "",
    val imei: String = "",
    val restBattery: Int? = null,
    val battery: Int? = null,
    val state: Int = 0,
    @JsonNames("task_id")
    val taskId: String = "",
    val izFinish: Boolean? = null,
    val checkResult: Int? = null,
) {
    fun toDomain(): BatchMoveChild = BatchMoveChild(
        taskId = taskId,
        carId = carId,
        imei = imei,
        restBattery = restBattery ?: battery ?: 0,
        state = state,
        izFinish = izFinish == true,
    )
}

/**
 * Man-made multi-vehicle move parent:
 * child list [man_made_move_list] with body field `id` (iOS TaskRequest.fetchMoveCarList),
 * batch start/finish with child task ids.
 */
class BatchMoveCarApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    suspend fun listChildren(parentTaskId: String): OpsResult<List<BatchMoveChild>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/move_car/task/man_made_move_list",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("id", parentTaskId)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-operation/move_car/task/man_made_move_list",
                bodyJson = body,
                deserializer = BatchMoveListDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.list.map { it.toDomain() })
            is OpsResult.Err -> result
        }
    }

    suspend fun startBatch(childTaskIds: List<String>, pin: String): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/move_car/task/start/batch",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("ids", buildJsonArray { childTaskIds.forEach { add(JsonPrimitive(it)) } })
            put("pin", pin)
        }
        return signedApi.postUnit(
            path = "business/ebike-operation/move_car/task/start/batch",
            bodyJson = body,
        )
    }

    suspend fun finishBatch(
        childTaskIds: List<String>,
        pin: String,
        pictures: List<String> = emptyList(),
        remark: String? = null,
    ): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/move_car/task/finish/batch",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("ids", buildJsonArray { childTaskIds.forEach { add(JsonPrimitive(it)) } })
            put("pin", pin)
            if (pictures.isNotEmpty()) {
                put("ignoreParkId", 1)
                put("photo", buildJsonArray { pictures.forEach { add(JsonPrimitive(it)) } })
            }
            if (!remark.isNullOrBlank()) {
                put("remark", remark)
            }
        }
        return signedApi.postUnit(
            path = "business/ebike-operation/move_car/task/finish/batch",
            bodyJson = body,
        )
    }
}
