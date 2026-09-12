package com.luopingtech.ebike.ops.data.task

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.OpsTask
import com.luopingtech.ebike.ops.domain.model.OpsTaskType
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.json.add
import kotlinx.serialization.json.put
import kotlinx.serialization.json.putJsonArray

/** Inspection (alarm) task APIs under alarm/task and alarm/accept. */
class InspectionTaskApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    suspend fun listPage(
        serviceId: String,
        state: Int,
        opPin: String? = null,
        pageSize: Int = 999,
    ): OpsResult<List<OpsTask>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/alarm/task/page",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("serviceId", serviceId)
            put("pageSize", pageSize)
            put("state", state)
            put("izShowDistance", true)
            if (!opPin.isNullOrBlank()) put("opPin", opPin)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-operation/alarm/task/page",
                bodyJson = body,
                deserializer = ChangeBatteryTaskListDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(
                result.value.list.map {
                    it.toDomain(serviceId).copy(type = OpsTaskType.Inspection)
                },
            )
            is OpsResult.Err -> result
        }
    }

    suspend fun claim(taskIds: List<String>, signPin: String): OpsResult<Unit> =
        assign(type = 1, taskIds = taskIds, signPin = signPin)

    /** type: 1 领取, 2 指派。 */
    suspend fun assign(type: Int, taskIds: List<String>, signPin: String): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/alarm/task/assign",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("type", type)
            putJsonArray("ids") { taskIds.forEach { add(it) } }
            put("signPin", signPin)
        }
        return signedApi.postUnit(
            path = "business/ebike-operation/alarm/task/assign",
            bodyJson = body,
        )
    }

    suspend fun pageStats(
        serviceId: String,
        state: Int,
        pageNum: Int,
        pageSize: Int,
        opPin: String,
        finishTimeStart: String,
        finishTimeEnd: String,
    ): OpsResult<ChangeBatteryTaskListDto> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/alarm/task/page",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("serviceId", serviceId)
            put("pageNum", pageNum)
            put("pageSize", pageSize)
            put("state", state)
            put("opPin", opPin)
            put("finishTimeStart", finishTimeStart)
            put("finishTimeEnd", finishTimeEnd)
        }
        return signedApi.post(
            path = "business/ebike-operation/alarm/task/page",
            bodyJson = body,
            deserializer = ChangeBatteryTaskListDto.serializer(),
        )
    }

    suspend fun start(taskId: String, pin: String): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/alarm/accept",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("id", taskId)
            put("pin", pin)
        }
        return signedApi.postUnit(
            path = "business/ebike-operation/alarm/accept",
            bodyJson = body,
        )
    }

    suspend fun finish(taskId: String, pictures: List<String> = emptyList(), remark: String? = null): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/alarm/task/finish",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("id", taskId)
            if (pictures.isNotEmpty()) {
                putJsonArray("photo") { pictures.forEach { add(it) } }
            }
            if (!remark.isNullOrBlank()) {
                put("remark", remark)
            }
        }
        return signedApi.postUnit(
            path = "business/ebike-operation/alarm/task/finish",
            bodyJson = body,
        )
    }
}
