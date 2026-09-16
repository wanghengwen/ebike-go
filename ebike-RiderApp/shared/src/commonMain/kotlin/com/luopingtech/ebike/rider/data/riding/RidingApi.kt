package com.luopingtech.ebike.rider.data.riding

import com.luopingtech.ebike.rider.core.json.LooseJson
import com.luopingtech.ebike.rider.core.network.ApiOutcome
import com.luopingtech.ebike.rider.core.network.CommonRequestBody
import com.luopingtech.ebike.rider.core.network.SignedApiClient
import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.domain.riding.BackCarConfig
import com.luopingtech.ebike.rider.domain.riding.PartMatchResult
import com.luopingtech.ebike.rider.domain.riding.ReturnPermissionResult
import com.luopingtech.ebike.rider.domain.riding.RideInfo
import com.luopingtech.ebike.rider.domain.riding.SettlementSummary
import com.luopingtech.ebike.rider.domain.riding.VehicleDetail
import com.luopingtech.ebike.rider.platform.DeviceInfo
import com.luopingtech.ebike.rider.platform.GeoPoint
import kotlinx.serialization.json.JsonObjectBuilder
import kotlinx.serialization.json.put

/** 骑行请求里反复出现的三件套：用户 pin + 定位。 */
data class RideActor(
    val userPin: String,
    val at: GeoPoint?,
) {
    val hasLocation: Boolean get() = at != null && (at.latitude != 0.0 || at.longitude != 0.0)
}

/**
 * 骑行主链路。path 与 UniApp `api/riding.ts` / `api/preCycling.ts` 一一对应，
 * 一个都不新造 —— 网关按 path 做灰度和限流，改名等于线上 404。
 */
interface RidingRemote {
    suspend fun rideInfo(actor: RideActor): RiderResult<ApiOutcome>
    suspend fun carInfo(carId: String): RiderResult<VehicleDetail>

    /** 扫码埋点，失败不影响流程。 */
    suspend fun submitScan(carId: String): RiderResult<Unit>

    /** 远程开锁。成功返回 orderId。 */
    suspend fun networkRide(carId: String, actor: RideActor): RiderResult<ApiOutcome>

    suspend fun returnPermission(
        carId: String,
        orderId: String,
        actor: RideActor,
        /** `izSw=1` 表示走蓝牙链路，后端据此放宽车机在线校验。 */
        bleChannel: Boolean,
        /** 蓝牙侧读到的配件状态（beacon 等），只有蓝牙通道会带。 */
        accessories: List<BleAccessoryReport> = emptyList(),
    ): RiderResult<ApiOutcome>

    suspend fun returnByNet(
        carId: String,
        orderId: String,
        actor: RideActor,
        returnType: Int,
    ): RiderResult<ApiOutcome>

    suspend fun tempPark(carId: String, userPin: String): RiderResult<ApiOutcome>
    suspend fun endTempPark(carId: String, userPin: String): RiderResult<ApiOutcome>

    suspend fun partMatchByTempParking(
        carId: String,
        orderId: String,
        serviceId: String,
        userPin: String,
    ): RiderResult<PartMatchResult>

    suspend fun backCarConfig(serviceId: String): RiderResult<BackCarConfig>

    /** 出服务区断电后临时恢复动力。 */
    suspend fun tempUnlock(carId: String): RiderResult<Unit>

    /** 解冻上一单。旧版每次进骑行页都软调一次，缺定位就跳过。 */
    suspend fun unFrozenOrder(actor: RideActor): RiderResult<Unit>

    suspend fun unlockHelmet(carId: String): RiderResult<Unit>

    /** 响铃找车。 */
    suspend fun playVehicleVoice(carId: String, imei: String): RiderResult<Unit>

    /** 结费页订单详情（`client/order/detail` / `detailLast`）。 */
    suspend fun orderDetail(orderId: String): RiderResult<SettlementSummary>

    /** 钱包充值/赠送余额分桶。 */
    suspend fun walletBalances(): RiderResult<Pair<Int, Int>>
}

class RidingApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) : RidingRemote {

    private fun body(block: JsonObjectBuilder.() -> Unit = {}): String =
        CommonRequestBody.toJsonString(
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
            block = block,
        )

    private fun JsonObjectBuilder.putActor(actor: RideActor) {
        if (actor.userPin.isNotBlank()) put("userPin", actor.userPin)
        actor.at?.let {
            put("userLat", it.latitude)
            put("userLng", it.longitude)
        }
    }

    override suspend fun rideInfo(actor: RideActor): RiderResult<ApiOutcome> =
        signedApi.postOutcome(
            PATH_RIDE_INFO,
            body {
                putActor(actor)
                put("version", CLIENT_VERSION)
            },
        )

    override suspend fun carInfo(carId: String): RiderResult<VehicleDetail> {
        return when (val out = signedApi.postOutcome(PATH_CAR_INFO, body { put("carId", carId) })) {
            is RiderResult.Err -> out
            is RiderResult.Ok -> {
                val outcome = out.value
                if (!outcome.success) {
                    RiderResult.Err(
                        RiderError.business(
                            outcome.code.ifBlank { "CAR_INFO" },
                            outcome.message,
                        ),
                    )
                } else {
                    RiderResult.Ok(RidingParsers.vehicleDetail(carId, outcome.data))
                }
            }
        }
    }

    override suspend fun submitScan(carId: String): RiderResult<Unit> =
        signedApi.postUnit(PATH_SCAN, body { put("carId", carId) })

    override suspend fun networkRide(carId: String, actor: RideActor): RiderResult<ApiOutcome> =
        signedApi.postOutcome(
            PATH_NETWORK_RIDE,
            body {
                put("carId", carId)
                putActor(actor)
            },
        )

    override suspend fun returnPermission(
        carId: String,
        orderId: String,
        actor: RideActor,
        bleChannel: Boolean,
        accessories: List<BleAccessoryReport>,
    ): RiderResult<ApiOutcome> = signedApi.postOutcome(
        PATH_RETURN_PERMISSION,
        body {
            put("carId", carId)
            if (orderId.isNotBlank()) put("orderId", orderId)
            putActor(actor)
            // 旧版恒传 true：前端已支持满桩提示，后端据此才会回 6101 而不是直接拒绝。
            put("izFrontSuppotFullPile", true)
            if (bleChannel) put("izSw", 1)
            if (accessories.isNotEmpty()) putAccessories(accessories)
        },
    )

    override suspend fun returnByNet(
        carId: String,
        orderId: String,
        actor: RideActor,
        returnType: Int,
    ): RiderResult<ApiOutcome> = signedApi.postOutcome(
        PATH_RETURN_BY_NET,
        body {
            put("carId", carId)
            if (orderId.isNotBlank()) put("orderId", orderId)
            putActor(actor)
            put("returnType", returnType)
            put("izFrontSuppotFullPile", true)
        },
    )

    override suspend fun tempPark(carId: String, userPin: String): RiderResult<ApiOutcome> =
        signedApi.postOutcome(
            PATH_TEMP_PARKING,
            body {
                put("carId", carId)
                if (userPin.isNotBlank()) put("userPin", userPin)
            },
        )

    override suspend fun endTempPark(carId: String, userPin: String): RiderResult<ApiOutcome> =
        signedApi.postOutcome(
            PATH_END_PARKING,
            body {
                put("carId", carId)
                if (userPin.isNotBlank()) put("userPin", userPin)
            },
        )

    override suspend fun partMatchByTempParking(
        carId: String,
        orderId: String,
        serviceId: String,
        userPin: String,
    ): RiderResult<PartMatchResult> {
        val out = signedApi.postOutcome(
            PATH_PART_MATCH_TEMP_PARKING,
            body {
                put("carId", carId)
                if (orderId.isNotBlank()) put("orderId", orderId)
                if (serviceId.isNotBlank()) put("serviceId", serviceId)
                if (userPin.isNotBlank()) put("userPin", userPin)
            },
        )
        return when (out) {
            is RiderResult.Err -> out
            is RiderResult.Ok -> {
                if (!out.value.success) {
                    // 配件校验是软校验：接口自己失败时不拦还车。
                    RiderResult.Ok(PartMatchResult(matched = true))
                } else {
                    val o = LooseJson.obj(out.value.data)
                    RiderResult.Ok(
                        PartMatchResult(
                            matched = LooseJson.boolOrNull(o, "izMatch") ?: true,
                            partName = LooseJson.string(o, "partName"),
                        ),
                    )
                }
            }
        }
    }

    override suspend fun backCarConfig(serviceId: String): RiderResult<BackCarConfig> {
        val out = signedApi.postOutcome(
            PATH_BACK_CAR_CONFIG,
            body { if (serviceId.isNotBlank()) put("serviceId", serviceId) },
        )
        return when (out) {
            is RiderResult.Err -> out
            is RiderResult.Ok -> RiderResult.Ok(RidingParsers.backCarConfig(out.value.data))
        }
    }

    override suspend fun tempUnlock(carId: String): RiderResult<Unit> =
        signedApi.postUnit(PATH_TEMP_UNLOCK, body { put("carId", carId) })

    override suspend fun unFrozenOrder(actor: RideActor): RiderResult<Unit> {
        // 旧版同样在这里挡一道：后端要求 userLat/userLng，缺定位直接返回而不是发个残缺请求。
        if (!actor.hasLocation) {
            return RiderResult.Err(RiderError.business("NO_LOCATION", "unFrozenOrder needs location"))
        }
        return signedApi.postUnit(
            PATH_UN_FROZEN_ORDER,
            body {
                putActor(actor)
                put("version", CLIENT_VERSION)
            },
        )
    }

    override suspend fun unlockHelmet(carId: String): RiderResult<Unit> =
        signedApi.postUnit(PATH_HELMET_UNLOCK, body { put("carId", carId) })

    override suspend fun playVehicleVoice(carId: String, imei: String): RiderResult<Unit> =
        signedApi.postUnit(
            PATH_CAR_SEARCH_VOICE,
            body {
                if (carId.isNotBlank()) put("carId", carId)
                if (imei.isNotBlank()) put("imei", imei)
            },
        )

    override suspend fun orderDetail(orderId: String): RiderResult<SettlementSummary> {
        val path = if (orderId.isBlank()) PATH_ORDER_DETAIL_LAST else PATH_ORDER_DETAIL
        return when (
            val out = signedApi.postOutcome(
                path,
                body {
                    if (orderId.isNotBlank()) put("orderId", orderId)
                    put("izNewApp", true)
                },
            )
        ) {
            is RiderResult.Err -> out
            is RiderResult.Ok -> {
                val outcome = out.value
                if (!outcome.success) {
                    RiderResult.Err(
                        RiderError.business(
                            outcome.code.ifBlank { "ORDER_DETAIL" },
                            outcome.message,
                        ),
                    )
                } else {
                    RiderResult.Ok(RidingParsers.settlementSummary(orderId, outcome.data))
                }
            }
        }
    }

    override suspend fun walletBalances(): RiderResult<Pair<Int, Int>> {
        return when (val out = signedApi.postOutcome(PATH_WALLET_INFO, body {})) {
            is RiderResult.Err -> out
            is RiderResult.Ok -> {
                val o = LooseJson.obj(out.value.data)
                RiderResult.Ok(
                    (LooseJson.int(o, "recharge") ?: 0) to (LooseJson.int(o, "present") ?: 0),
                )
            }
        }
    }

    companion object {
        const val PATH_RIDE_INFO: String = "client/rent/getRideInfo"
        const val PATH_CAR_INFO: String = "client/rent/getCarInfo"
        const val PATH_SCAN: String = "client/rent/scan"
        const val PATH_NETWORK_RIDE: String = "client/rent/network/ride"
        const val PATH_RETURN_PERMISSION: String = "client/rent/returnPermission"
        const val PATH_RETURN_BY_NET: String = "client/rent/network/return"
        const val PATH_TEMP_PARKING: String = "client/rent/tempParking"
        const val PATH_END_PARKING: String = "client/rent/endParking"
        const val PATH_TEMP_UNLOCK: String = "client/rent/tempUnlock"
        const val PATH_UN_FROZEN_ORDER: String = "client/rent/unFrozenOrder"
        const val PATH_HELMET_UNLOCK: String = "client/helmet/unlock"
        const val PATH_BACK_CAR_CONFIG: String = "client/system/getbackCarConfig"
        const val PATH_PART_MATCH_TEMP_PARKING: String = "client/rent/part/partMatchByTempParking"
        const val PATH_CAR_SEARCH_VOICE: String = "client/paas/device/carSearchVoice"
        const val PATH_ORDER_DETAIL: String = "client/order/detail"
        const val PATH_ORDER_DETAIL_LAST: String = "client/order/detailLast"
        const val PATH_WALLET_INFO: String = "client/ebike-account/wallet/get_wallet_info"

        /** 旧版恒传 `1.0.0`，后端按它区分老小程序，不要跟 App 版本号绑一起。 */
        const val CLIENT_VERSION: String = "1.0.0"
    }
}

/**
 * `data` → 领域模型。抠在一个 object 里方便被 demo 实现和单测复用。
 */
object RidingParsers {

    fun vehicleDetail(fallbackCarId: String, data: kotlinx.serialization.json.JsonElement?): VehicleDetail {
        val o = LooseJson.obj(data)
        return VehicleDetail(
            carId = LooseJson.string(o, "carId", "id").ifBlank { fallbackCarId },
            imei = LooseJson.string(o, "imei"),
            restBattery = LooseJson.int(o, "restBattery", "soc", "battery"),
            restMileage = LooseJson.int(o, "restMileage"),
            disconnected = LooseJson.bool(o, "isDisconnect", "izDisconnect"),
            errType = LooseJson.int(o, "errType") ?: 0,
            ridingType = LooseJson.int(o, "ridingType"),
            outOfServiceArea = LooseJson.bool(o, "isOutofServAera", "izOutofServAera"),
            redEnvelope = LooseJson.bool(o, "izRedEnvelope", "isRedEnvelope", "izRedPaket"),
            startPriceFen = LooseJson.int(o, "startingPrice"),
            startPriceMinutes = millisToMinutes(LooseJson.long(o, "startingTime")),
            lat = LooseJson.double(o, "lat", "latitude", "carLat") ?: 0.0,
            lng = LooseJson.double(o, "lng", "longitude", "carLng") ?: 0.0,
        )
    }

    fun rideInfo(data: kotlinx.serialization.json.JsonElement?): RideInfo {
        val o = LooseJson.obj(data) ?: return RideInfo()
        val tempUnlockCo = LooseJson.nested(o, "tempUnLockCO", "tempUnlockCO")
        return RideInfo(
            orderId = LooseJson.string(o, "orderId"),
            carId = LooseJson.string(o, "carId"),
            imei = LooseJson.string(o, "imei"),
            ridingState = LooseJson.int(o, "ridingState"),
            costFeeFen = LooseJson.int(o, "costFee") ?: 0,
            rideTimeMillis = LooseJson.long(o, "rideTime") ?: 0L,
            rideDistanceMeters = LooseJson.int(o, "rideDistance") ?: 0,
            restBattery = LooseJson.int(o, "restBattery", "soc"),
            restMileage = LooseJson.int(o, "restMileage"),
            helmetState = LooseJson.int(o, "helmetState"),
            helmetPopup = LooseJson.int(o, "helmetPopup") ?: 0,
            fenceTypeCode = LooseJson.int(o, "type", "ridingType", "izRidingType") ?: 0,
            dispatchCostFen = LooseJson.int(o, "dispatchCost", "penalty") ?: 0,
            canReturn = LooseJson.bool(o, "izCanReturn"),
            lat = LooseJson.double(o, "lat", "latitude", "carLat") ?: 0.0,
            lng = LooseJson.double(o, "lng", "longitude", "carLng") ?: 0.0,
            overloadState = LooseJson.int(o, "overloadState") ?: 0,
            nearServiceEdge = LooseJson.bool(o, "izNearService"),
            inFreeTime = LooseJson.bool(o, "izFreeTime"),
            tempUnlockState = LooseJson.int(tempUnlockCo, "tempUnlockState") ?: 0,
            tempUnlocked = LooseJson.bool(tempUnlockCo, "izTemp"),
        )
    }

    fun returnPermission(data: kotlinx.serialization.json.JsonElement?): ReturnPermissionResult {
        val o = LooseJson.obj(data)
        return ReturnPermissionResult(
            canReturn = LooseJson.bool(o, "izCanReturn"),
            returnTypeCode = LooseJson.int(o, "returnType") ?: 0,
            penaltyFen = LooseJson.int(o, "penalty", "dispatchCost") ?: 0,
        )
    }

    fun backCarConfig(data: kotlinx.serialization.json.JsonElement?): BackCarConfig {
        val o = LooseJson.obj(data)
        return BackCarConfig(
            civilizationRemind = LooseJson.bool(o, "izCivilizationRemind"),
            showApplyEntry = LooseJson.bool(o, "dispatchFee"),
            autoLockMinutes = LooseJson.int(o, "autoLockTime", "outServiceAreaAutoLock") ?: 3,
        )
    }

    fun orderIdOf(data: kotlinx.serialization.json.JsonElement?): String =
        LooseJson.string(LooseJson.obj(data), "orderId")

    fun settlementSummary(
        fallbackOrderId: String,
        data: kotlinx.serialization.json.JsonElement?,
    ): SettlementSummary {
        val o = LooseJson.obj(data)
        val payCost = LooseJson.int(o, "payCost", "cost", "waitPayMoney") ?: 0
        val origin = LooseJson.int(o, "originCost") ?: payCost
        val izPaid = LooseJson.int(o, "izPaid") ?: 0
        val ridingTimeMs = LooseJson.long(o, "ridingTime", "rideTime") ?: 0L
        val mileMeters = LooseJson.int(o, "mile", "rideDistance") ?: 0
        val frozenRaw = LooseJson.long(o, "frozenTime")
        val frozenAt = when {
            frozenRaw == null || frozenRaw <= 0L -> 0L
            frozenRaw < 1_000_000_000_000L -> frozenRaw * 1000L // 秒 → 毫秒
            else -> frozenRaw
        }
        return SettlementSummary(
            orderId = LooseJson.string(o, "orderId").ifBlank { fallbackOrderId },
            costFeeFen = payCost,
            dispatchFeeFen = LooseJson.int(o, "dispatchCost", "penalty") ?: 0,
            rideTimeSeconds = if (ridingTimeMs > 0L) ridingTimeMs / 1000L else 0L,
            rideDistanceMeters = mileMeters,
            settled = izPaid == 3 || izPaid == 4 || payCost <= 0,
            originCostFen = origin,
            startPriceFen = LooseJson.int(o, "startPrice"),
            timeCostFen = LooseJson.int(o, "timeCost", "durationCost"),
            mileCostFen = LooseJson.int(o, "mileCost"),
            rechargeBalanceFen = LooseJson.int(o, "recharge") ?: 0,
            presentBalanceFen = LooseJson.int(o, "present") ?: 0,
            frozenAtMillis = frozenAt,
        )
    }

    private fun millisToMinutes(millis: Long?): Int {
        val v = millis ?: return 0
        if (v <= 0L) return 0
        return ((v + 30_000L) / 60_000L).toInt()
    }
}
