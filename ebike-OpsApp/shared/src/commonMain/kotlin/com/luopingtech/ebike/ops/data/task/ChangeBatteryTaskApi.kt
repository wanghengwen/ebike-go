package com.luopingtech.ebike.ops.data.task

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.BatteryRange
import com.luopingtech.ebike.ops.domain.model.OpsTask
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.put

@Serializable
data class BatteryRangeDto(
    val min: Int = 0,
    val max: Int = 30,
) {
    fun toDomain(): BatteryRange = BatteryRange(
        min = min.coerceIn(0, 100),
        max = max.coerceIn(0, 100).coerceAtLeast(min),
    )
}

/** Flutter-aligned change-battery task APIs (list / detail / start / finish / range). */
class ChangeBatteryTaskApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    suspend fun batteryRange(serviceId: String): OpsResult<BatteryRange> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/task_rules/battery_range",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("serviceId", serviceId)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-operation/task_rules/battery_range",
                bodyJson = body,
                deserializer = BatteryRangeDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain())
            is OpsResult.Err -> result
        }
    }

    suspend fun list(
        serviceId: String,
        maxBattery: Int = 30,
        pageSize: Int = 999,
    ): OpsResult<List<OpsTask>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/change_battery/task/list",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("serviceId", serviceId)
            put("maxBattery", maxBattery)
            put("pageSize", pageSize)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-operation/change_battery/task/list",
                bodyJson = body,
                deserializer = ChangeBatteryTaskListDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.list.map { it.toDomain(serviceId) })
            is OpsResult.Err -> result
        }
    }

    suspend fun detail(taskId: String): OpsResult<OpsTask> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/change_battery/task/detail",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("id", taskId)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-operation/change_battery/task/detail",
                bodyJson = body,
                deserializer = ChangeBatteryTaskDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain())
            is OpsResult.Err -> result
        }
    }

    suspend fun start(
        taskId: String,
        serviceId: String,
        pin: String,
        izBlue: Boolean = false,
    ): OpsResult<Unit> = postAction(
        path = "business/ebike-operation/change_battery/task/start",
        source = "/business/ebike-operation/change_battery/task/start",
        taskId = taskId,
        serviceId = serviceId,
        pin = pin,
        izBlue = izBlue,
    )

    suspend fun finish(
        taskId: String,
        serviceId: String,
        pin: String,
        izBlue: Boolean = false,
    ): OpsResult<Unit> = postAction(
        path = "business/ebike-operation/change_battery/task/finish",
        source = "/business/ebike-operation/change_battery/task/finish",
        taskId = taskId,
        serviceId = serviceId,
        pin = pin,
        izBlue = izBlue,
    )

    private suspend fun postAction(
        path: String,
        source: String,
        taskId: String,
        serviceId: String,
        pin: String,
        izBlue: Boolean,
    ): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = source,
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("taskId", taskId)
            put("serviceId", serviceId)
            put("pin", pin)
            put("izBlue", izBlue)
        }
        return signedApi.postUnit(path = path, bodyJson = body)
    }
}
