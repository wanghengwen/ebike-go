package com.luopingtech.ebike.ops.data.task

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.OpsTask
import com.luopingtech.ebike.ops.domain.model.OpsTaskType
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.buildJsonArray
import kotlinx.serialization.json.put

/**
 * Move-car task APIs (Flutter task center):
 * page list, assign/claim, auto_task_start, auto_task_finish.
 */
class MoveCarTaskApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    suspend fun list(
        serviceId: String,
        pageNum: Int = 1,
        pageSize: Int = 50,
    ): OpsResult<List<OpsTask>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/move_car/task/page",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("serviceId", serviceId)
            put("izFilter", true)
            put("pageNum", pageNum)
            put("pageSize", pageSize)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-operation/move_car/task/page",
                bodyJson = body,
                deserializer = ChangeBatteryTaskListDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(
                result.value.list.map {
                    it.toDomain(serviceId).copy(type = OpsTaskType.MoveCar)
                },
            )
            is OpsResult.Err -> result
        }
    }

    suspend fun detail(taskId: String): OpsResult<OpsTask> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/move_car/task/detail",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("id", taskId)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-operation/move_car/task/detail",
                bodyJson = body,
                deserializer = ChangeBatteryTaskDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain().copy(type = OpsTaskType.MoveCar))
            is OpsResult.Err -> result
        }
    }

    /** type: 1 claim, 2 assign. */
    suspend fun claim(taskId: String, signPin: String): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/move_car/task/assign",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("type", 1)
            put("id", taskId)
            put("signPin", signPin)
        }
        return signedApi.postUnit(
            path = "business/ebike-operation/move_car/task/assign",
            bodyJson = body,
        )
    }

    suspend fun start(taskId: String): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/move_car/task/auto_task_start",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("taskId", taskId)
        }
        return signedApi.postUnit(
            path = "business/ebike-operation/move_car/task/auto_task_start",
            bodyJson = body,
        )
    }

    suspend fun finish(taskId: String, pictures: List<String> = emptyList(), remark: String? = null): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/move_car/task/auto_task_finish",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("taskId", taskId)
            if (pictures.isNotEmpty()) {
                put("photo", buildJsonArray { pictures.forEach { add(JsonPrimitive(it)) } })
            }
            if (!remark.isNullOrBlank()) {
                put("remark", remark)
            }
        }
        return signedApi.postUnit(
            path = "business/ebike-operation/move_car/task/auto_task_finish",
            bodyJson = body,
        )
    }
}
