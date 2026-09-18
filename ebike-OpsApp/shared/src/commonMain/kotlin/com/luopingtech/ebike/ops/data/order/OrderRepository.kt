package com.luopingtech.ebike.ops.data.order

import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.core.time.nowEpochMillis
import com.luopingtech.ebike.ops.domain.model.LastOrder
import com.luopingtech.ebike.ops.domain.model.TrackPoint
import com.luopingtech.ebike.ops.domain.order.OrderDepositRecord
import com.luopingtech.ebike.ops.domain.order.OrderListQuery
import com.luopingtech.ebike.ops.domain.order.OrderPayStates
import com.luopingtech.ebike.ops.domain.order.OrderRecord
import com.luopingtech.ebike.ops.domain.order.OrderRideCard
import com.luopingtech.ebike.ops.domain.order.OrderRideCardRecord
import com.luopingtech.ebike.ops.domain.order.OrderUserAssets
import com.luopingtech.ebike.ops.domain.order.OrderUserDetail
import com.luopingtech.ebike.ops.domain.order.OrderUserPageItem
import com.luopingtech.ebike.ops.domain.order.OrderWalletInfo
import com.luopingtech.ebike.ops.domain.order.OrderWalletRecord
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

    /** Legacy: /business/ebike-account/user_account */
    suspend fun getUserAssets(pin: String, serviceId: String): OpsResult<OrderUserAssets>

    suspend fun getWalletInfo(pin: String): OpsResult<OrderWalletInfo>
    suspend fun getWalletRecords(pin: String): OpsResult<List<OrderWalletRecord>>
    suspend fun editWalletPresent(pin: String, changePresentFen: Int): OpsResult<Unit>
    suspend fun getDepositRecords(pin: String): OpsResult<List<OrderDepositRecord>>
    suspend fun getRideCards(pin: String): OpsResult<List<OrderRideCard>>
    suspend fun getRideCardRecords(pin: String): OpsResult<List<OrderRideCardRecord>>
    suspend fun refundableAmount(transactId: String, paidAt: String): OpsResult<Int>
    suspend fun refundWallet(
        pin: String,
        refundFeeFen: Int,
        tradeNo: String,
        tradeTime: String,
    ): OpsResult<Unit>
    suspend fun ridingRefund(pin: String, refundFeeFen: Int, tradeNo: String): OpsResult<Unit>
    suspend fun resetCarStatus(pin: String): OpsResult<Unit>
    suspend fun manualReturnDeposit(pin: String, refundFeeFen: Int): OpsResult<Unit>

    /** Legacy orderDetail：补全轨迹。 */
    suspend fun orderDetail(orderId: String): OpsResult<OrderRecord>

    /** Legacy: /business/rent/focusReturn */
    suspend fun focusReturn(userPin: String, carId: String): OpsResult<Unit>

    /** Legacy: /business/rent/focusReturnWithCost（金额分） */
    suspend fun focusReturnWithCost(
        userPin: String,
        carId: String,
        modifyPayCostFen: Int,
        modifyDispatchCostFen: Int,
        modifyHelmetPenaltyFen: Int,
    ): OpsResult<Unit>

    /** Legacy: /business/rent/tempUnlockConfig（秒） */
    suspend fun temporaryUnlock(carId: String, unLockTimeSec: Int): OpsResult<Unit>

    /** Legacy: /business/ebike-operation/tools/start */
    suspend fun startVehicle(carId: String): OpsResult<Unit>

    /** Legacy: /business/createUpdateCostTicket（金额分） */
    suspend fun modifyOrderCost(
        orderId: String,
        modifyPayCostFen: Int,
        modifyDispatchCostFen: Int,
        modifyHelmetPenaltyFen: Int,
    ): OpsResult<Unit>
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
                    authNo = "430***********1234",
                    izDeposited = true,
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
                    authNo = "430***********1234",
                    izDeposited = true,
                ),
            )
        }
        return api.getUserByPin(pin)
    }

    override suspend fun getUserAssets(pin: String, serviceId: String): OpsResult<OrderUserAssets> {
        if (pin.isBlank()) {
            return OpsResult.Err(OpsError.business("ORDER", "pin empty"))
        }
        if (demoMode || api == null) {
            return OpsResult.Ok(
                OrderUserAssets(
                    walletBalanceFen = 12800,
                    depositedMountFen = 9900,
                    depositedStats = 1,
                    ridingCardName = "演示骑行卡",
                    depositCardDiscountFen = 0,
                ),
            )
        }
        return api.getUserAssets(pin, serviceId)
    }

    override suspend fun getWalletInfo(pin: String): OpsResult<OrderWalletInfo> {
        if (demoMode || api == null) {
            return OpsResult.Ok(OrderWalletInfo(balanceFen = 12800, rechargeFen = 8000, presentFen = 4800))
        }
        return api.getWalletInfo(pin)
    }

    override suspend fun getWalletRecords(pin: String): OpsResult<List<OrderWalletRecord>> {
        if (demoMode || api == null) {
            return OpsResult.Ok(
                listOf(
                    OrderWalletRecord(
                        amountFen = 2000,
                        changeType = "充值",
                        channel = "微信",
                        izRefund = 0,
                        merchantTradeNo = "demo-wallet-1",
                        paidAt = "2026-09-01 10:00:00",
                        presentAmountFen = 0,
                        rechargeAmountFen = 2000,
                        type = "钱包",
                    ),
                ),
            )
        }
        return api.getWalletRecords(pin)
    }

    override suspend fun editWalletPresent(pin: String, changePresentFen: Int): OpsResult<Unit> {
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.editWalletPresent(pin, changePresentFen)
    }

    override suspend fun getDepositRecords(pin: String): OpsResult<List<OrderDepositRecord>> {
        if (demoMode || api == null) {
            return OpsResult.Ok(
                listOf(
                    OrderDepositRecord(
                        paidAt = "2026-01-01 10:00:00",
                        depositType = "诚信金",
                        amountFen = 9900,
                        duration = "--",
                        state = "已支付",
                        type = "缴纳",
                        channel = "微信",
                        merchantTradeNo = "demo-dep-1",
                    ),
                ),
            )
        }
        return api.getDepositRecords(pin)
    }

    override suspend fun getRideCards(pin: String): OpsResult<List<OrderRideCard>> {
        if (demoMode || api == null) {
            return OpsResult.Ok(listOf(OrderRideCard(name = "演示骑行卡", cardExpiredDate = "2026-12-31", remainTimes = "3")))
        }
        return api.getRideCards(pin)
    }

    override suspend fun getRideCardRecords(pin: String): OpsResult<List<OrderRideCardRecord>> {
        if (demoMode || api == null) {
            return OpsResult.Ok(
                listOf(
                    OrderRideCardRecord(
                        paidAt = "2026-08-01 12:00:00",
                        name = "演示骑行卡",
                        amountFen = 19900,
                        duration = "30",
                        type = "购买",
                        channel = "微信",
                        merchantTradeNo = "demo-ride-1",
                    ),
                ),
            )
        }
        return api.getRideCardRecords(pin)
    }

    override suspend fun refundableAmount(transactId: String, paidAt: String): OpsResult<Int> {
        if (demoMode || api == null) return OpsResult.Ok(100)
        return api.refundableAmount(transactId, paidAt)
    }

    override suspend fun refundWallet(
        pin: String,
        refundFeeFen: Int,
        tradeNo: String,
        tradeTime: String,
    ): OpsResult<Unit> {
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.refund(pin, refundFeeFen, "WALLET", tradeNo, tradeTime, tradeNo)
    }

    override suspend fun ridingRefund(pin: String, refundFeeFen: Int, tradeNo: String): OpsResult<Unit> {
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.ridingRefund(pin, refundFeeFen, tradeNo)
    }

    override suspend fun resetCarStatus(pin: String): OpsResult<Unit> {
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.resetCarStatus(pin)
    }

    override suspend fun manualReturnDeposit(pin: String, refundFeeFen: Int): OpsResult<Unit> {
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.manualReturnDeposit(pin, refundFeeFen)
    }

    override suspend fun orderDetail(orderId: String): OpsResult<OrderRecord> {
        val id = orderId.trim()
        if (id.isEmpty() || id == "0") {
            return OpsResult.Err(OpsError.business("ORDER", "orderId empty"))
        }
        if (demoMode || api == null) {
            val demo = demoLastOrder("demo-car")
            return OpsResult.Ok(
                demoOrderRecord("demo-car", "demo-user-pin", "13800138000").copy(
                    id = id,
                    trajectory = demo.trajectory,
                    startLat = demo.startLat,
                    startLng = demo.startLng,
                    endLat = demo.endLat,
                    endLng = demo.endLng,
                ),
            )
        }
        return api.orderDetail(id)
    }

    override suspend fun focusReturn(userPin: String, carId: String): OpsResult<Unit> {
        if (userPin.isBlank() || carId.isBlank()) {
            return OpsResult.Err(OpsError.business("ORDER", "userPin/carId empty"))
        }
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.focusReturn(userPin.trim(), carId.trim())
    }

    override suspend fun focusReturnWithCost(
        userPin: String,
        carId: String,
        modifyPayCostFen: Int,
        modifyDispatchCostFen: Int,
        modifyHelmetPenaltyFen: Int,
    ): OpsResult<Unit> {
        if (userPin.isBlank() || carId.isBlank()) {
            return OpsResult.Err(OpsError.business("ORDER", "userPin/carId empty"))
        }
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.focusReturnWithCost(
            userPin.trim(),
            carId.trim(),
            modifyPayCostFen,
            modifyDispatchCostFen,
            modifyHelmetPenaltyFen,
        )
    }

    override suspend fun temporaryUnlock(carId: String, unLockTimeSec: Int): OpsResult<Unit> {
        if (carId.isBlank()) {
            return OpsResult.Err(OpsError.business("ORDER", "carId empty"))
        }
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.temporaryUnlock(carId.trim(), unLockTimeSec)
    }

    override suspend fun startVehicle(carId: String): OpsResult<Unit> {
        if (carId.isBlank()) {
            return OpsResult.Err(OpsError.business("ORDER", "carId empty"))
        }
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.startVehicle(carId.trim())
    }

    override suspend fun modifyOrderCost(
        orderId: String,
        modifyPayCostFen: Int,
        modifyDispatchCostFen: Int,
        modifyHelmetPenaltyFen: Int,
    ): OpsResult<Unit> {
        if (orderId.isBlank() || orderId == "0") {
            return OpsResult.Err(OpsError.business("ORDER", "orderId empty"))
        }
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.modifyOrderCost(
            orderId.trim(),
            modifyPayCostFen,
            modifyDispatchCostFen,
            modifyHelmetPenaltyFen,
        )
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
            izPaid: Int = OrderPayStates.Riding,
            carState: Int? = 2,
        ): OrderRecord = OrderRecord(
            id = "demo-order-$carId",
            carId = carId,
            phone = phone,
            userPin = userPin,
            userName = "演示用户",
            originCost = 350,
            payCost = 350,
            mile = 2300,
            ridingTimeRaw = "1500000",
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
            carState = carState,
            trajectory = demoLastOrder(carId).trajectory,
        )
    }
}
