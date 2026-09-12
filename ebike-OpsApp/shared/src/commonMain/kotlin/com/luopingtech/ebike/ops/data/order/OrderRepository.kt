package com.luopingtech.ebike.ops.data.order

import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.core.time.nowEpochMillis
import com.luopingtech.ebike.ops.domain.model.LastOrder
import com.luopingtech.ebike.ops.domain.model.TrackPoint
import com.luopingtech.ebike.ops.domain.order.OrderListQuery
import com.luopingtech.ebike.ops.domain.order.OrderPayStates
import com.luopingtech.ebike.ops.domain.order.OrderRecord
import com.luopingtech.ebike.ops.domain.order.OrderUserDetail
import com.luopingtech.ebike.ops.domain.order.OrderUserPageItem
import kotlin.math.cos
import kotlin.math.sin

interface OrderRepository {
    suspend fun detailLast(carId: String): OpsResult<LastOrder>

    suspend fun detailLastRecord(
        carId: String? = null,
        userPin: String? = null,
        imei: String? = null,
    ): OpsResult<OrderRecord>

    suspend fun listOrders(query: OrderListQuery): OpsResult<List<OrderRecord>>

    suspend fun getUserByPhone(phone: String, serviceIds: List<String>): OpsResult<OrderUserDetail>

    suspend fun listUsersByName(
        authName: String,
        serviceIds: List<Long>,
        pageNum: Int,
        pageSize: Int,
    ): OpsResult<Pair<List<OrderUserPageItem>, Int>>

    suspend fun getUserByPin(pin: String): OpsResult<OrderUserDetail>
}

class OrderRepositoryImpl(
    private val demoMode: Boolean,
    private val api: OrderApi? = null,
) : OrderRepository {
    override suspend fun detailLast(carId: String): OpsResult<LastOrder> {
        if (carId.isBlank()) {
            return OpsResult.Err(OpsError.business("ORDER", "carId empty"))
        }
        if (demoMode || api == null) {
            return OpsResult.Ok(demoLastOrder(carId.trim()))
        }
        return api.detailLast(carId.trim())
    }

    override suspend fun detailLastRecord(
        carId: String?,
        userPin: String?,
        imei: String?,
    ): OpsResult<OrderRecord> {
        if (carId.isNullOrBlank() && userPin.isNullOrBlank() && imei.isNullOrBlank()) {
            return OpsResult.Err(OpsError.business("ORDER", "scope empty"))
        }
        if (demoMode || api == null) {
            return OpsResult.Ok(
                demoOrderRecord(
                    carId = carId.orEmpty().ifBlank { "demo-car" },
                    userPin = userPin.orEmpty().ifBlank { "demo-user-pin" },
                    phone = "13800138000",
                ),
            )
        }
        return api.detailLastRecord(carId = carId, userPin = userPin, imei = imei)
    }

    override suspend fun listOrders(query: OrderListQuery): OpsResult<List<OrderRecord>> {
        if (demoMode || api == null) {
            if (query.pageNum > 1) return OpsResult.Ok(emptyList())
            return OpsResult.Ok(
                listOf(
                    demoOrderRecord("100860001", "demo-user-pin", "13800138000"),
                    demoOrderRecord("100860002", "demo-user-2", "13900139000", izPaid = OrderPayStates.Paid),
                ),
            )
        }
        return api.listOrders(query)
    }

    override suspend fun getUserByPhone(
        phone: String,
        serviceIds: List<String>,
    ): OpsResult<OrderUserDetail> {
        if (phone.isBlank()) {
            return OpsResult.Err(OpsError.business("ORDER", "phone empty"))
        }
        if (demoMode || api == null) {
            return OpsResult.Ok(
                OrderUserDetail(
                    pin = "demo-user-pin",
                    authName = "演示用户",
                    phone = phone,
                    createdAt = "2026-01-01T10:00:00",
                    izAuth = true,
                    ridingState = 3,
                    balance = 12800,
                    serviceName = "演示服务区",
                ),
            )
        }
        return api.getUserByPhone(phone, serviceIds)
    }

    override suspend fun listUsersByName(
        authName: String,
        serviceIds: List<Long>,
        pageNum: Int,
        pageSize: Int,
    ): OpsResult<Pair<List<OrderUserPageItem>, Int>> {
        if (authName.isBlank()) {
            return OpsResult.Err(OpsError.business("ORDER", "authName empty"))
        }
        if (serviceIds.isEmpty()) {
            return OpsResult.Err(OpsError.business("ORDER", "serviceId empty"))
        }
        if (demoMode || api == null) {
            if (pageNum > 1) return OpsResult.Ok(emptyList<OrderUserPageItem>() to 1)
            return OpsResult.Ok(
                listOf(
                    OrderUserPageItem(
                        pin = "demo-user-pin",
                        authName = authName,
                        phone = "13800138000",
                        izAuth = true,
                        ridingState = 3,
                        createdAt = "2026-01-01T10:00:00",
                        balance = 12800,
                    ),
                ) to 1,
            )
        }
        return api.listUsersByName(authName, serviceIds, pageNum, pageSize)
    }

    override suspend fun getUserByPin(pin: String): OpsResult<OrderUserDetail> {
        if (pin.isBlank()) {
            return OpsResult.Err(OpsError.business("ORDER", "pin empty"))
        }
        if (demoMode || api == null) {
            return OpsResult.Ok(
                OrderUserDetail(
                    pin = pin,
                    authName = "演示用户",
                    phone = "13800138000",
                    createdAt = "2026-01-01T10:00:00",
                    izAuth = true,
                    ridingState = 6,
                    balance = 12800,
                    serviceName = "演示服务区",
                ),
            )
        }
        return api.getUserByPin(pin)
    }

    companion object {
        fun demoLastOrder(carId: String): LastOrder {
            val baseLat = 28.221
            val baseLng = 112.941
            val track = (0 until 10).map { i ->
                val t = i / 10.0 * 6.28
                TrackPoint(
                    lat = baseLat + 0.0008 * sin(t),
                    lng = baseLng + 0.0008 * cos(t),
                    timestamp = nowEpochMillis() - (10 - i) * 90_000L,
                    speed = 10f,
                )
            }
            return LastOrder(
                carId = carId,
                startLat = track.first().lat,
                startLng = track.first().lng,
                endLat = track.last().lat,
                endLng = track.last().lng,
                trajectory = track,
                id = "demo-order-$carId",
                userPin = "demo-user-pin",
                userPhone = "13800138000",
                startTime = "2026-09-11 08:00:00",
                endTime = "2026-09-11 08:25:00",
            )
        }

        fun demoOrderRecord(
            carId: String,
            userPin: String,
            phone: String,
            izPaid: Int = OrderPayStates.ToPay,
        ): OrderRecord = OrderRecord(
            id = "demo-order-$carId",
            carId = carId,
            phone = phone,
            userPin = userPin,
            userName = "演示用户",
            originCost = 350,
            payCost = 350,
            mile = 2300,
            ridingTimeRaw = "1500",
            izPaid = izPaid,
            startLat = 28.221,
            startLng = 112.941,
            endLat = 28.225,
            endLng = 112.945,
            startTime = "2026-09-11 08:00:00",
            endTime = "2026-09-11 08:25:00",
            hasPaid = 0,
            dispatchCost = 0,
            helmetPenalty = 0,
        )
    }
}
