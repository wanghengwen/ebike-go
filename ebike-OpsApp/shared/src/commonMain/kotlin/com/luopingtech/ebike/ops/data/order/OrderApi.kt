package com.luopingtech.ebike.ops.data.order

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.LastOrder
import com.luopingtech.ebike.ops.domain.order.OrderDepositRecord
import com.luopingtech.ebike.ops.domain.order.OrderListQuery
import com.luopingtech.ebike.ops.domain.order.OrderRecord
import com.luopingtech.ebike.ops.domain.order.OrderRideCard
import com.luopingtech.ebike.ops.domain.order.OrderRideCardRecord
import com.luopingtech.ebike.ops.domain.order.OrderUserAssets
import com.luopingtech.ebike.ops.domain.order.OrderUserDetail
import com.luopingtech.ebike.ops.domain.order.OrderUserPageItem
import com.luopingtech.ebike.ops.domain.order.OrderWalletInfo
import com.luopingtech.ebike.ops.domain.order.OrderWalletRecord
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.builtins.ListSerializer
import kotlinx.serialization.builtins.serializer
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

    /** Legacy: /business/order/orderDetail — 含 deviceTrajectory（列表超 6 个月常无轨迹）。 */
    suspend fun orderDetail(orderId: String): OpsResult<OrderRecord> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/order/orderDetail",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("orderId", orderId)
        }
        return when (
            val result = signedApi.post(
                path = "business/order/orderDetail",
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
            // 对齐原版 OrderInformationRepositoryV2：近期列表传 endTime 区间
            query.endTimeMs?.let { (from, to) ->
                put("endTime", buildJsonArray {
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

    /** Legacy: /business/ebike-account/user_account — 钱包 / 骑行卡 / 押金卡。 */
    suspend fun getUserAssets(pin: String, serviceId: String): OpsResult<OrderUserAssets> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-account/user_account",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("pin", pin)
            put("serviceId", serviceId)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-account/user_account",
                bodyJson = body,
                deserializer = OrderUserAssetsDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain())
            is OpsResult.Err -> result
        }
    }

    suspend fun getWalletInfo(pin: String): OpsResult<OrderWalletInfo> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-account/wallet/detail",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) { put("pin", pin) }
        return when (
            val result = signedApi.post(
                path = "business/ebike-account/wallet/detail",
                bodyJson = body,
                deserializer = OrderWalletInfoDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain())
            is OpsResult.Err -> result
        }
    }

    suspend fun getWalletRecords(pin: String): OpsResult<List<OrderWalletRecord>> {
        val body = CommonRequestBody.toJsonString(
            source = "/ebike_visual/merchant/business/wallet/user_record",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) { put("pin", pin) }
        return when (
            val result = signedApi.post(
                path = "ebike_visual/merchant/business/wallet/user_record",
                bodyJson = body,
                deserializer = ListSerializer(OrderWalletRecordDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.map { it.toDomain() })
            is OpsResult.Err -> result
        }
    }

    suspend fun editWalletPresent(pin: String, changePresentFen: Int): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-account/wallet/edit",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("pin", pin)
            put("changePresent", changePresentFen)
        }
        return signedApi.postUnit(path = "business/ebike-account/wallet/edit", bodyJson = body)
    }

    suspend fun getDepositRecords(pin: String): OpsResult<List<OrderDepositRecord>> {
        val body = CommonRequestBody.toJsonString(
            source = "/ebike_visual/merchant/business/deposit/user_record",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) { put("pin", pin) }
        return when (
            val result = signedApi.post(
                path = "ebike_visual/merchant/business/deposit/user_record",
                bodyJson = body,
                deserializer = ListSerializer(OrderDepositRecordDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.map { it.toDomain() })
            is OpsResult.Err -> result
        }
    }

    suspend fun getRideCards(pin: String): OpsResult<List<OrderRideCard>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-account/riding_card/detail",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) { put("pin", pin) }
        return when (
            val result = signedApi.post(
                path = "business/ebike-account/riding_card/detail",
                bodyJson = body,
                deserializer = OrderRideCardListDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.used.orEmpty().map { it.toDomain() })
            is OpsResult.Err -> result
        }
    }

    suspend fun getRideCardRecords(pin: String): OpsResult<List<OrderRideCardRecord>> {
        val body = CommonRequestBody.toJsonString(
            source = "/ebike_visual/merchant/business/riding_card/user_record",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) { put("pin", pin) }
        return when (
            val result = signedApi.post(
                path = "ebike_visual/merchant/business/riding_card/user_record",
                bodyJson = body,
                deserializer = ListSerializer(OrderRideCardRecordDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.map { it.toDomain() })
            is OpsResult.Err -> result
        }
    }

    suspend fun refundableAmount(transactId: String, paidAt: String): OpsResult<Int> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/pay/refund/amount",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("transactId", transactId)
            put("paidAt", paidAt)
        }
        return signedApi.post(
            path = "business/pay/refund/amount",
            bodyJson = body,
            deserializer = Int.serializer(),
        )
    }

    suspend fun refund(
        pin: String,
        refundFeeFen: Int,
        saleType: String,
        tradeNo: String,
        tradeTime: String,
        outRefundNo: String,
    ): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/pay/refund",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("pin", pin)
            put("refund_fee", refundFeeFen)
            put("sale_type", saleType)
            put("trade_no", tradeNo)
            put("trade_time", tradeTime)
            put("out_refund_no", outRefundNo)
        }
        return signedApi.postUnit(path = "business/pay/refund", bodyJson = body)
    }

    suspend fun ridingRefund(pin: String, refundFeeFen: Int, tradeNo: String): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/pay/refund",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("sale_type", "RIDING_CARD")
            put("pin", pin)
            put("refund_fee", refundFeeFen)
            put("trade_no", tradeNo)
        }
        return signedApi.postUnit(path = "business/pay/refund", bodyJson = body)
    }

    suspend fun resetCarStatus(pin: String): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/user/user/recover",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) { put("pin", pin) }
        return signedApi.postUnit(path = "business/user/user/recover", bodyJson = body)
    }

    suspend fun manualReturnDeposit(pin: String, refundFeeFen: Int): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/pay/manualRefund",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("pin", pin)
            put("sale_type", "DEPOSIT")
            put("is_manual_refund", 1)
            put("refund_fee", refundFeeFen)
        }
        return signedApi.postUnit(path = "business/pay/manualRefund", bodyJson = body)
    }

    /**
     * Legacy: /business/rent/focusReturn — 强制结束行程。
     * 对齐 Merchant-Android UserOrderInfoRepository.getForceReturn。
     */
    suspend fun focusReturn(userPin: String, carId: String): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/rent/focusReturn",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("userPin", userPin)
            put("carId", carId)
        }
        return signedApi.postUnit(path = "business/rent/focusReturn", bodyJson = body)
    }

    /**
     * Legacy: /business/rent/focusReturnWithCost — 结束行程并按运维金额结算。
     * 金额单位：分。
     */
    suspend fun focusReturnWithCost(
        userPin: String,
        carId: String,
        modifyPayCostFen: Int,
        modifyDispatchCostFen: Int,
        modifyHelmetPenaltyFen: Int,
    ): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/rent/focusReturnWithCost",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("userPin", userPin)
            put("carId", carId)
            put("modifyPayCost", modifyPayCostFen)
            put("modifyDispatchCost", modifyDispatchCostFen)
            if (modifyHelmetPenaltyFen >= 0) {
                put("modifyHelmetPenalty", modifyHelmetPenaltyFen)
            }
        }
        return signedApi.postUnit(path = "business/rent/focusReturnWithCost", bodyJson = body)
    }

    /**
     * Legacy: /business/rent/tempUnlockConfig — 临时通电。
     * [unLockTimeSec] 秒。
     */
    suspend fun temporaryUnlock(carId: String, unLockTimeSec: Int): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/rent/tempUnlockConfig",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("carId", carId)
            put("unLockTime", unLockTimeSec)
        }
        return signedApi.postUnit(path = "business/rent/tempUnlockConfig", bodyJson = body)
    }

    /**
     * Legacy: /business/ebike-operation/tools/start — 启动（临停开锁）。
     * 对齐车辆控制 NetworkUnlock；用户详情页 openAcc → switchLockCar。
     */
    suspend fun startVehicle(carId: String): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/tools/start",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("carId", carId)
            put("izRiskControl", false)
        }
        return signedApi.postUnit(path = "business/ebike-operation/tools/start", bodyJson = body)
    }

    /**
     * Legacy: /business/createUpdateCostTicket — 未支付订单修改金额。
     * 金额单位：分。
     */
    suspend fun modifyOrderCost(
        orderId: String,
        modifyPayCostFen: Int,
        modifyDispatchCostFen: Int,
        modifyHelmetPenaltyFen: Int,
    ): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/createUpdateCostTicket",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("orderId", orderId)
            put("modifyPayCost", modifyPayCostFen)
            put("modifyDispatchCost", modifyDispatchCostFen)
            if (modifyHelmetPenaltyFen >= 0) {
                put("modifyHelmetPenalty", modifyHelmetPenaltyFen)
            }
        }
        return signedApi.postUnit(path = "business/createUpdateCostTicket", bodyJson = body)
    }
}
