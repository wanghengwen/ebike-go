package com.luopingtech.ebike.ops.data.task

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.OpsTaskType
import com.luopingtech.ebike.ops.domain.model.TaskAuditResult
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.put

@Serializable
data class TaskAuditResultDto(
    val remark: String? = null,
    val checkResult: Int? = null,
    val checkView: String? = null,
    val photo: List<String>? = null,
    val checkTime: String? = null,
    val checkManName: String? = null,
    val checkManPin: String? = null,
    val checkManPhone: String? = null,
    val state: Int? = null,
    val taskTicketId: Long? = null,
) {
    fun toDomain(): TaskAuditResult = TaskAuditResult(
        checkResult = checkResult,
        checkView = checkView.orEmpty(),
        remark = remark.orEmpty(),
        photo = photo.orEmpty().filter { it.isNotBlank() },
        checkTime = checkTime.orEmpty(),
        checkManName = checkManName.orEmpty(),
    )
}

@Serializable
data class MoveCarDetailAuditDto(
    val id: String = "",
    val checkResult: Int? = null,
    val checkInfo: TaskAuditResultDto? = null,
)

/**
 * Legacy TaskAuditViewModel:
 * - inspection: alarm/checkResultDetail
 * - repair: fix/check_result_detail
 * - move-car: move_car/task/detail → checkInfo (no dedicated check_result_detail on server)
 */
class TaskAuditApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    suspend fun load(taskId: String, type: OpsTaskType): OpsResult<TaskAuditResult> {
        return when (type) {
            OpsTaskType.Inspection -> checkResultDetail(
                source = "/business/ebike-operation/alarm/checkResultDetail",
                path = "business/ebike-operation/alarm/checkResultDetail",
                taskId = taskId,
            )
            OpsTaskType.Repair -> checkResultDetail(
                source = "/business/ebike-operation/fix/check_result_detail",
                path = "business/ebike-operation/fix/check_result_detail",
                taskId = taskId,
            )
            OpsTaskType.MoveCar -> moveCarCheckInfo(taskId)
            else -> OpsResult.Ok(TaskAuditResult())
        }
    }

    private suspend fun checkResultDetail(
        source: String,
        path: String,
        taskId: String,
    ): OpsResult<TaskAuditResult> {
        val body = CommonRequestBody.toJsonString(
            source = source,
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            val asLong = taskId.toLongOrNull()
            if (asLong != null) put("id", asLong) else put("id", taskId)
        }
        return when (
            val result = signedApi.post(
                path = path,
                bodyJson = body,
                deserializer = TaskAuditResultDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain())
            is OpsResult.Err -> result
        }
    }

    private suspend fun moveCarCheckInfo(taskId: String): OpsResult<TaskAuditResult> {
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
                deserializer = MoveCarDetailAuditDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> {
                val info = result.value.checkInfo?.toDomain()
                    ?: TaskAuditResult(checkResult = result.value.checkResult)
                OpsResult.Ok(info)
            }
            is OpsResult.Err -> result
        }
    }
}
