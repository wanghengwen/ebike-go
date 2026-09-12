package com.luopingtech.ebike.ops.data.order

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.LastOrder
import com.luopingtech.ebike.ops.domain.order.OrderListQuery
import com.luopingtech.ebike.ops.domain.order.OrderRecord
import com.luopingtech.ebike.ops.domain.order.OrderUserDetail
import com.luopingtech.ebike.ops.domain.order.OrderUserPageItem
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.json.add
import kotlinx.serialization.json.buildJsonArray
import kotlinx.serialization.json.put

class OrderApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    /** Legacy: /business/order/detailLast — last order + embedded deviceTrajectory. */
    suspend fun detailLast(carId: String): OpsResult<LastOrder> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/order/detailLast",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("carId", carId)
        }
        return when (
            val result = signedApi.post(
                path = "business/order/detailLast",
                bodyJson = body,
                deserializer = LastOrderDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain())
            is OpsResult.Err -> result
        }
    }

    suspend fun detailLastRecord(
        carId: String? = null,
        userPin: String? = null,
        imei: String? = null,
    ): OpsResult<OrderRecord> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/order/detailLast",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            carId?.takeIf { it.isNotBlank() }?.let { put("carId", it) }
            userPin?.takeIf { it.isNotBlank() }?.let { put("userPin", it) }
            imei?.takeIf { it.isNotBlank() }?.let { put("imei", it) }
        }
        return when (
            val result = signedApi.post(
                path = "business/order/detailLast",
                bodyJson = body,
                deserializer = LastOrderDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toOrderRecord())
            is OpsResult.Err -> result
        }
    }

    /** Legacy: /business/order/bList */
    suspend fun listOrders(query: OrderListQuery): OpsResult<List<OrderRecord>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/order/bList",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("pageNum", query.pageNum)
            put("pageSize", query.pageSize)
            query.userPin?.takeIf { it.isNotBlank() }?.let { put("userPin", it) }
            query.carId?.takeIf { it.isNotBlank() }?.let { put("carId", it) }
            query.imei?.takeIf { it.isNotBlank() }?.let { put("imei", it) }
            query.serviceId?.let { put("serviceId", it) }
            query.izPaid?.let { put("izPaid", it) }
            query.startTimeMs?.let { (from, to) ->
                put("startTime", buildJsonArray {
                    add(from)
                    add(to)
                })
            }
        }
        return when (
            val result = signedApi.post(
                path = "business/order/bList",
                bodyJson = body,
                deserializer = OrderPageDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.list.orEmpty().map { it.toOrderRecord() })
            is OpsResult.Err -> result
        }
    }

    /** Legacy: /business/user/user/detailByPhone — phone must include +86- prefix. */
    suspend fun getUserByPhone(phone: String, serviceIds: List<String>): OpsResult<OrderUserDetail> {
        val normalized = phone.trim().removePrefix("+86-").removePrefix("+86")
        val body = CommonRequestBody.toJsonString(
            source = "/business/user/user/detailByPhone",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("type", "1")
            put("phone", "+86-$normalized")
            put("serviceIds", buildJsonArray { serviceIds.forEach { add(it) } })
        }
        return when (
            val result = signedApi.post(
                path = "business/user/user/detailByPhone",
                bodyJson = body,
                deserializer = OrderUserDetailDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain())
            is OpsResult.Err -> result
        }
    }

    /** Legacy: /business/user/user/page — authName exact match; serviceId required. */
    suspend fun listUsersByName(
        authName: String,
        serviceIds: List<Long>,
        pageNum: Int,
        pageSize: Int,
    ): OpsResult<Pair<List<OrderUserPageItem>, Int>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/user/user/page",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("authName", authName)
            put("serviceId", buildJsonArray { serviceIds.forEach { add(it) } })
            put("pageNum", pageNum)
            put("pageSize", pageSize)
        }
        return when (
            val result = signedApi.post(
                path = "business/user/user/page",
                bodyJson = body,
                deserializer = OrderUserPageDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(
                result.value.list.orEmpty().map { it.toDomain() } to result.value.count,
            )
            is OpsResult.Err -> result
        }
    }

    suspend fun getUserByPin(pin: String): OpsResult<OrderUserDetail> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/user/user/detail",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("pin", pin)
        }
        return when (
            val result = signedApi.post(
                path = "business/user/user/detail",
                bodyJson = body,
                deserializer = OrderUserDetailDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain())
            is OpsResult.Err -> result
        }
    }
}
