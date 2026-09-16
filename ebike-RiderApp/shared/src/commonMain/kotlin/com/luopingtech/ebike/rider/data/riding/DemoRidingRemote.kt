package com.luopingtech.ebike.rider.data.riding

import com.luopingtech.ebike.rider.core.network.ApiOutcome
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.core.time.nowEpochMillis
import com.luopingtech.ebike.rider.data.fence.FenceRemote
import com.luopingtech.ebike.rider.domain.ble.BleFrame
import com.luopingtech.ebike.rider.domain.model.FenceKind
import com.luopingtech.ebike.rider.domain.model.FencePolygon
import com.luopingtech.ebike.rider.domain.model.GeoLatLng
import com.luopingtech.ebike.rider.domain.riding.BackCarConfig
import com.luopingtech.ebike.rider.domain.riding.BleReturnConfig
import com.luopingtech.ebike.rider.domain.riding.NearParking
import com.luopingtech.ebike.rider.domain.riding.PartMatchResult
import com.luopingtech.ebike.rider.domain.riding.ReturnTypeCodes
import com.luopingtech.ebike.rider.domain.riding.SettlementSummary
import com.luopingtech.ebike.rider.domain.riding.VehicleDetail
import com.luopingtech.ebike.rider.platform.GeoPoint
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.put

/**
 * Demo 的假订单账本。
 *
 * 必须被 [DemoRidingRemote] 与 [DemoBleRideRemote] 共用：默认策略是蓝牙优先，
 * 订单由 `bleRide.rideReport` 建出来，而 `getRideInfo` 是 riding 侧回的。
 * 各存一份的话蓝牙开锁后骑行页会一直显示 0 元 —— 演示直接失去意义。
 */
class DemoRideBook {
    internal var orderSeq: Int = 0
    internal var currentOrderId: String = ""
    internal var currentCarId: String = ""
    internal var startedAt: Long = 0L
    internal var tempParked: Boolean = false
    internal var finished: Boolean = false

    /** 开出一张假单，返回订单号。 */
    internal fun open(carId: String): String {
        orderSeq += 1
        currentOrderId = "DEMO-ORDER-$orderSeq"
        currentCarId = carId
        startedAt = nowEpochMillis()
        tempParked = false
        finished = false
        return currentOrderId
    }

    internal fun close() {
        finished = true
        tempParked = false
    }

    internal val hasOpenOrder: Boolean get() = currentOrderId.isNotBlank() && !finished
}

/**
 * 无后端 demo。`api.baseUrl` 为空时装配这一套，整条主循环（假解锁 → 骑行 → 假还车 → 结费）
 * 可以在模拟器上跑完。
 *
 * 计费故意做得"能看出在动"：按秒累加，两分钟起步价之后每分钟 1 毛。
 */
class DemoRidingRemote(
    private val book: DemoRideBook = DemoRideBook(),
) : RidingRemote {

    override suspend fun rideInfo(actor: RideActor): RiderResult<ApiOutcome> {
        if (!book.hasOpenOrder) {
            // 没有进行中订单：15042 是旧版「无订单」的软失败码，界面不弹错。
            return RiderResult.Ok(
                ApiOutcome(success = false, code = "15042", message = "no riding order", data = null),
            )
        }
        val elapsedMs = (nowEpochMillis() - book.startedAt).coerceAtLeast(0L)
        val elapsedSec = elapsedMs / 1000L
        val data = buildJsonObject {
            put("orderId", book.currentOrderId)
            put("carId", book.currentCarId)
            put("imei", BleFrame.demoImeiForCarId(book.currentCarId))
            put("ridingState", if (book.tempParked) 3 else 4)
            put("costFee", demoCostFen(elapsedSec))
            put("rideTime", elapsedMs)
            put("rideDistance", (elapsedSec * DEMO_METERS_PER_SECOND).toInt())
            put("restBattery", (88 - elapsedSec / 60).coerceAtLeast(5))
            put("restMileage", (42 - elapsedSec / 120).coerceAtLeast(1))
            put("helmetState", 1)
            put("type", ReturnTypeCodes.NORMAL)
            put("izCanReturn", true)
            put("dispatchCost", 0)
            put("lat", DEMO_LAT)
            put("lng", DEMO_LNG)
        }
        return RiderResult.Ok(ApiOutcome(success = true, code = "0", message = "", data = data))
    }

    override suspend fun carInfo(carId: String): RiderResult<VehicleDetail> = RiderResult.Ok(
        VehicleDetail(
            carId = carId,
            imei = BleFrame.demoImeiForCarId(carId),
            restBattery = 76,
            restMileage = 38,
            startPriceFen = DEMO_START_PRICE_FEN,
            startPriceMinutes = DEMO_START_MINUTES,
            lat = DEMO_LAT,
            lng = DEMO_LNG,
        ),
    )

    override suspend fun submitScan(carId: String): RiderResult<Unit> = RiderResult.Ok(Unit)

    override suspend fun networkRide(carId: String, actor: RideActor): RiderResult<ApiOutcome> =
        RiderResult.Ok(ok(buildJsonObject { put("orderId", book.open(carId)) }))

    override suspend fun returnPermission(
        carId: String,
        orderId: String,
        actor: RideActor,
        bleChannel: Boolean,
        accessories: List<BleAccessoryReport>,
    ): RiderResult<ApiOutcome> = RiderResult.Ok(
        ok(
            buildJsonObject {
                put("izCanReturn", true)
                put("returnType", ReturnTypeCodes.NORMAL)
                put("penalty", 0)
            },
        ),
    )

    override suspend fun returnByNet(
        carId: String,
        orderId: String,
        actor: RideActor,
        returnType: Int,
    ): RiderResult<ApiOutcome> {
        val closed = book.currentOrderId
        book.close()
        return RiderResult.Ok(ok(buildJsonObject { put("orderId", closed) }))
    }

    override suspend fun tempPark(carId: String, userPin: String): RiderResult<ApiOutcome> {
        book.tempParked = true
        return RiderResult.Ok(ok(null))
    }

    override suspend fun endTempPark(carId: String, userPin: String): RiderResult<ApiOutcome> {
        book.tempParked = false
        return RiderResult.Ok(ok(null))
    }

    override suspend fun partMatchByTempParking(
        carId: String,
        orderId: String,
        serviceId: String,
        userPin: String,
    ): RiderResult<PartMatchResult> = RiderResult.Ok(PartMatchResult(matched = true))

    override suspend fun backCarConfig(serviceId: String): RiderResult<BackCarConfig> =
        RiderResult.Ok(
            BackCarConfig(
                civilizationRemind = false,
                showApplyEntry = true,
                autoLockMinutes = 3,
            ),
        )

    override suspend fun tempUnlock(carId: String): RiderResult<Unit> = RiderResult.Ok(Unit)

    override suspend fun unFrozenOrder(actor: RideActor): RiderResult<Unit> = RiderResult.Ok(Unit)

    override suspend fun unlockHelmet(carId: String): RiderResult<Unit> = RiderResult.Ok(Unit)

    override suspend fun playVehicleVoice(carId: String, imei: String): RiderResult<Unit> =
        RiderResult.Ok(Unit)

    override suspend fun orderDetail(orderId: String): RiderResult<SettlementSummary> {
        val oid = orderId.ifBlank { book.currentOrderId }
        val seconds = 148L
        return RiderResult.Ok(
            SettlementSummary(
                orderId = oid,
                costFeeFen = DEMO_START_PRICE_FEN,
                originCostFen = DEMO_START_PRICE_FEN,
                startPriceFen = DEMO_START_PRICE_FEN,
                timeCostFen = 0,
                mileCostFen = 0,
                rideTimeSeconds = seconds,
                rideDistanceMeters = 246,
                rechargeBalanceFen = 0,
                presentBalanceFen = 0,
                frozenAtMillis = nowEpochMillis(),
                settled = false,
            ),
        )
    }

    override suspend fun walletBalances(): RiderResult<Pair<Int, Int>> =
        RiderResult.Ok(0 to 0)

    private fun ok(data: JsonElement?) =
        ApiOutcome(success = true, code = "0", message = "", data = data)

    companion object {
        const val DEMO_START_PRICE_FEN: Int = 150
        const val DEMO_START_MINUTES: Int = 2
        private const val DEMO_PER_MINUTE_FEN: Int = 10
        private const val DEMO_METERS_PER_SECOND: Int = 4
        private const val DEMO_LAT: Double = 28.221
        private const val DEMO_LNG: Double = 112.941

        internal fun demoCostFen(elapsedSeconds: Long): Int {
            val freeSeconds = DEMO_START_MINUTES * 60L
            if (elapsedSeconds <= freeSeconds) return DEMO_START_PRICE_FEN
            val extraMinutes = ((elapsedSeconds - freeSeconds) + 59L) / 60L
            return DEMO_START_PRICE_FEN + (extraMinutes * DEMO_PER_MINUTE_FEN).toInt()
        }
    }
}

/** Demo BLE 上报：一律成功，让蓝牙优先分支真的走通。 */
class DemoBleRideRemote(
    private val book: DemoRideBook = DemoRideBook(),
) : BleRideRemote {

    override suspend fun ridePermission(
        carId: String,
        actor: RideActor,
        accessories: List<BleAccessoryReport>,
    ): RiderResult<ApiOutcome> = RiderResult.Ok(ok(null))

    override suspend fun rideReport(carId: String, actor: RideActor): RiderResult<ApiOutcome> =
        RiderResult.Ok(ok(buildJsonObject { put("orderId", book.open(carId)) }))

    override suspend fun returnReport(
        carId: String,
        orderId: String,
        actor: RideActor,
        accessories: List<BleAccessoryReport>,
    ): RiderResult<ApiOutcome> {
        book.close()
        return RiderResult.Ok(ok(null))
    }

    override suspend fun tempParkReport(carId: String, actor: RideActor): RiderResult<ApiOutcome> {
        book.tempParked = true
        return RiderResult.Ok(ok(null))
    }

    override suspend fun endTempParkReport(carId: String, actor: RideActor): RiderResult<ApiOutcome> {
        book.tempParked = false
        return RiderResult.Ok(ok(null))
    }

    override suspend fun returnConfig(carId: String): RiderResult<BleReturnConfig> =
        RiderResult.Ok(BleReturnConfig(needBeacon = false))

    private fun ok(data: JsonElement?) =
        ApiOutcome(success = true, code = "0", message = "", data = data)
}

class DemoApplyReturnRemote : ApplyReturnRemote {
    override suspend fun canApply(applyType: Int?): RiderResult<Boolean> = RiderResult.Ok(true)

    override suspend fun createAudit(
        photoUrls: List<String>,
        orderId: String,
        applyType: Int?,
        reason: String,
    ): RiderResult<Unit> = RiderResult.Ok(Unit)
}

/**
 * Demo 围栏：车周围一个方形服务区 + 一个停车点，够验证「区内 / 区外」两种界面。
 */
class DemoFenceRemote : FenceRemote {
    override suspend fun serviceAreaIdAt(at: GeoPoint): RiderResult<String> =
        RiderResult.Ok("DEMO-SERVICE-1")

    override suspend fun nearFences(at: GeoPoint, serviceId: String): RiderResult<List<FencePolygon>> {
        val lat = if (at.latitude != 0.0) at.latitude else 28.22
        val lng = if (at.longitude != 0.0) at.longitude else 112.94
        return RiderResult.Ok(
            listOf(
                FencePolygon(
                    id = "demo-service",
                    name = "Demo 服务区",
                    kind = FenceKind.ServiceArea,
                    points = square(lat, lng, 0.008),
                ),
                FencePolygon(
                    id = "demo-parking",
                    name = "Demo 停车点 P",
                    kind = FenceKind.Parking,
                    points = square(lat + 0.0015, lng + 0.0015, 0.0006),
                ),
            ),
        )
    }

    override suspend fun nearParking(at: GeoPoint, serviceId: String): RiderResult<NearParking> {
        val lat = if (at.latitude != 0.0) at.latitude else 28.22
        val lng = if (at.longitude != 0.0) at.longitude else 112.94
        return RiderResult.Ok(
            NearParking(
                count = 3,
                distanceMeters = 180.0,
                lat = lat + 0.0015,
                lng = lng + 0.0015,
                name = "Demo 停车点 P",
            ),
        )
    }

    private fun square(lat: Double, lng: Double, half: Double): List<GeoLatLng> = listOf(
        GeoLatLng(lat - half, lng - half),
        GeoLatLng(lat - half, lng + half),
        GeoLatLng(lat + half, lng + half),
        GeoLatLng(lat + half, lng - half),
    )
}
