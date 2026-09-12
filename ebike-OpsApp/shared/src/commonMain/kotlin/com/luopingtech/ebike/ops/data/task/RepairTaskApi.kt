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

/** Repair (fix) task APIs aligned with Flutter task_api.dart under fix/, not form/. */
class RepairTaskApi(
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
            source = "/business/ebike-operation/fix/task/page",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("serviceId", serviceId)
            put("pageSize", pageSize)
            put("state", state)
            if (!opPin.isNullOrBlank()) put("opPin", opPin)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-operation/fix/task/page",
                bodyJson = body,
                deserializer = ChangeBatteryTaskListDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(
                result.value.list.map {
                    it.toDomain(serviceId).copy(type = OpsTaskType.Repair)
                },
            )
            is OpsResult.Err -> result
        }
    }

    suspend fun claim(taskId: String, signPin: String): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/fix/task/assign",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("type", 1)
            put("id", taskId)
            put("signPin", signPin)
        }
        return signedApi.postUnit(
            path = "business/ebike-operation/fix/task/assign",
            bodyJson = body,
        )
    }

    suspend fun start(taskId: String, pin: String): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/fix/fix/accept",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("id", taskId)
            put("pin", pin)
        }
        return signedApi.postUnit(
            path = "business/ebike-operation/fix/fix/accept",
            bodyJson = body,
        )
    }

    suspend fun finish(taskId: String, pictures: List<String> = emptyList(), remark: String? = null): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/fix/task/end",
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
            path = "business/ebike-operation/fix/task/end",
            bodyJson = body,
        )
    }

    /**
     * Create drag-back move task.
     * Flutter path fix/task/create_drag (body id) + iOS fields (fixId / dragReason / dragAddress).
     */
    suspend fun createDrag(
        taskId: String,
        carId: String = "",
        areaId: String = "",
        dragReason: String,
        dragAddress: String,
    ): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/fix/task/create_drag",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("id", taskId)
            put("fixId", taskId)
            if (carId.isNotBlank()) put("carId", carId)
            if (areaId.isNotBlank()) put("areaId", areaId)
            put("dragReason", dragReason)
            put("dragAddress", dragAddress)
        }
        return signedApi.postUnit(
            path = "business/ebike-operation/fix/task/create_drag",
            bodyJson = body,
        )
    }
}
