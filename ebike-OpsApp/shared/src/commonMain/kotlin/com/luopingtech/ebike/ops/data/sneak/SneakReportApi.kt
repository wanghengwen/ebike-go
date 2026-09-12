package com.luopingtech.ebike.ops.data.sneak

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.SneakReportRecord
import com.luopingtech.ebike.ops.domain.model.SneakReportType
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.builtins.ListSerializer
import kotlinx.serialization.json.add
import kotlinx.serialization.json.put
import kotlinx.serialization.json.putJsonArray
import kotlinx.serialization.json.putJsonObject

/**
 * Legacy merchant sneak report APIs (1214) — snake_case paths under /ebike_operation/.
 * Distinct from FaultReport (1210 business/ebike-operation/repair).
 */
class SneakReportApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    suspend fun typeList(serviceId: String): OpsResult<List<SneakReportType>> {
        val body = CommonRequestBody.toJsonString(
            source = "/ebike_operation/tools/business/sneak_type_list/query",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("service_id", serviceId)
        }
        return when (
            val result = signedApi.post(
                path = "ebike_operation/tools/business/sneak_type_list/query",
                bodyJson = body,
                deserializer = ListSerializer(SneakTypeDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.map { it.toDomain() })
            is OpsResult.Err -> result
        }
    }

    suspend fun submit(
        carId: String,
        description: String,
        typeIds: List<String>,
        otherType: String,
        photoUrls: List<String>,
        itinId: String,
        reportedUserPin: String,
        reportManPin: String,
    ): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/ebike_operation/tools/business/sneak",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            putJsonObject("report_type") {
                putJsonArray("type") { typeIds.forEach { add(it) } }
                putJsonArray("other_type") { add(otherType) }
            }
            put("car_id", carId)
            put("description", description)
            putJsonArray("imgs") { photoUrls.forEach { add(it) } }
            put("itin_id", itinId)
            put("reported_user_pin", reportedUserPin)
            put("report_man_pin", reportManPin)
        }
        return signedApi.postUnit("ebike_operation/tools/business/sneak", body)
    }

    suspend fun myList(
        reportManPhone: String,
        page: Int = 1,
        size: Int = 50,
    ): OpsResult<List<SneakReportRecord>> {
        val body = CommonRequestBody.toJsonString(
            source = "/ebike_operation/platform/business/sneak_list",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("report_man_phone", reportManPhone)
            put("page", page)
            put("size", size)
        }
        return when (
            val result = signedApi.post(
                path = "ebike_operation/platform/business/sneak_list",
                bodyJson = body,
                deserializer = SneakHistoryPageDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.result.map { it.toDomain() })
            is OpsResult.Err -> result
        }
    }

    suspend fun cancel(sneakId: String, serviceId: String): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/ebike_operation/platform/business/merchant_cancel_sneak",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("sneak_id", sneakId)
            put("service_id", serviceId)
        }
        return signedApi.postUnit("ebike_operation/platform/business/merchant_cancel_sneak", body)
    }

    suspend fun checkServicePermission(carId: String, serviceId: String): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/carInfo/carPermissionCheck",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("carId", carId)
            put("serviceId", serviceId)
        }
        return signedApi.postUnit("business/ebike-management/carInfo/carPermissionCheck", body)
    }
}
