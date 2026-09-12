package com.luopingtech.ebike.ops.data.workorder

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.WorkOrder
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.json.JsonObjectBuilder
import kotlinx.serialization.json.add
import kotlinx.serialization.json.put
import kotlinx.serialization.json.putJsonArray

/**
 * Legacy inspection work-order ledger (1208):
 * - alarm/page (izAlarmProcessing=true)
 * - alarm/accept
 * - alarm/finish
 */
class InspectionOrderApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    suspend fun listPage(
        serviceId: String,
        alarmType: Int? = null,
        carId: String? = null,
        pageNum: Int = 1,
        pageSize: Int = 50,
    ): OpsResult<List<WorkOrder>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/alarm/page",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("serviceId", serviceId)
            put("pageNum", pageNum)
            put("pageSize", pageSize)
            put("izAlarmProcessing", true)
            if (alarmType != null) put("type", alarmType)
            if (!carId.isNullOrBlank()) put("carId", carId)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-operation/alarm/page",
                bodyJson = body,
                deserializer = AlarmTicketPageDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.list.map { it.toDomain(serviceId) })
            is OpsResult.Err -> result
        }
    }

    suspend fun accept(orderId: String): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/alarm/accept",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            putOrderId(orderId)
        }
        return signedApi.postUnit("business/ebike-operation/alarm/accept", body)
    }

    suspend fun finish(orderId: String): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/alarm/finish",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            putOrderId(orderId)
        }
        return signedApi.postUnit("business/ebike-operation/alarm/finish", body)
    }
}

/**
 * Legacy repair work-order ledger (1209):
 * - fix/page
 * - fix/fix/accept
 * - fix/finish (handlerType 2=done)
 */
class RepairOrderApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    suspend fun listPage(
        serviceId: String,
        fixNames: List<String> = emptyList(),
        carId: String? = null,
        pageNum: Int = 1,
        pageSize: Int = 50,
    ): OpsResult<List<WorkOrder>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/fix/page",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("serviceId", serviceId)
            put("pageNum", pageNum)
            put("pageSize", pageSize)
            if (fixNames.isNotEmpty()) {
                putJsonArray("fixName") {
                    fixNames.forEach { add(it) }
                }
            }
            if (!carId.isNullOrBlank()) put("carId", carId)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-operation/fix/page",
                bodyJson = body,
                deserializer = FixTicketPageDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.list.map { it.toDomain(serviceId) })
            is OpsResult.Err -> result
        }
    }

    suspend fun accept(orderId: String): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/fix/fix/accept",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            putOrderId(orderId)
        }
        return signedApi.postUnit("business/ebike-operation/fix/fix/accept", body)
    }

    suspend fun finish(orderId: String, handlerType: Int = 2): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/fix/finish",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            putOrderId(orderId)
            put("handlerType", handlerType)
        }
        return signedApi.postUnit("business/ebike-operation/fix/finish", body)
    }
}

private fun JsonObjectBuilder.putOrderId(orderId: String) {
    val asLong = orderId.toLongOrNull()
    if (asLong != null) put("id", asLong) else put("id", orderId)
}
