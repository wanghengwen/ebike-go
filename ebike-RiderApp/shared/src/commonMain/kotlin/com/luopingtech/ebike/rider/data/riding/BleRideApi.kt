package com.luopingtech.ebike.rider.data.riding

import com.luopingtech.ebike.rider.core.json.LooseJson
import com.luopingtech.ebike.rider.core.network.ApiOutcome
import com.luopingtech.ebike.rider.core.network.CommonRequestBody
import com.luopingtech.ebike.rider.core.network.SignedApiClient
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.domain.riding.BleReturnConfig
import com.luopingtech.ebike.rider.domain.riding.ReturnByNetType
import com.luopingtech.ebike.rider.platform.DeviceInfo
import kotlinx.serialization.json.JsonObjectBuilder
import kotlinx.serialization.json.buildJsonArray
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.put

/**
 * 蓝牙链路的服务端配套接口（UniApp `api/ble.ts`）。
 *
 * 蓝牙开锁**不是**纯本地操作：先 `ridePermission` 拿放行、再写 BLE 帧、成功后 `ride` 建单。
 * 少了任何一步都会出现「车开了但没订单」或「有订单但车没开」。
 */
interface BleRideRemote {
    /** 开锁前放行校验。 */
    suspend fun ridePermission(
        carId: String,
        actor: RideActor,
        accessories: List<BleAccessoryReport> = emptyList(),
    ): RiderResult<ApiOutcome>

    /** 开锁成功后建单。 */
    suspend fun rideReport(carId: String, actor: RideActor): RiderResult<ApiOutcome>

    suspend fun returnReport(
        carId: String,
        orderId: String,
        actor: RideActor,
        accessories: List<BleAccessoryReport> = emptyList(),
    ): RiderResult<ApiOutcome>

    suspend fun tempParkReport(carId: String, actor: RideActor): RiderResult<ApiOutcome>
    suspend fun endTempParkReport(carId: String, actor: RideActor): RiderResult<ApiOutcome>

    /** 还车前是否要读 beacon。 */
    suspend fun returnConfig(carId: String): RiderResult<BleReturnConfig>
}

/** BLE 侧读到的配件状态，作为 `result` 数组回传后端。 */
data class BleAccessoryReport(
    val name: String,
    val exists: Boolean = true,
    val usable: Boolean = true,
    val passed: Boolean = true,
    /** notify 解出来的原始字段，原样回传给后端存证。 */
    val state: Map<String, String> = emptyMap(),
)

/**
 * 配件数组写成旧版那套键名（`izExist` / `canUse` / `result`）。
 * `ridePermission`、`returnPermission`、`blue/return` 三处共用同一份形状。
 */
internal fun JsonObjectBuilder.putAccessories(accessories: List<BleAccessoryReport>) {
    put(
        "result",
        buildJsonArray {
            accessories.forEach { item ->
                add(
                    buildJsonObject {
                        put("name", item.name)
                        put("izExist", item.exists)
                        put("canUse", item.usable)
                        put("result", item.passed)
                        if (item.state.isNotEmpty()) {
                            put(
                                "state",
                                buildJsonObject {
                                    item.state.forEach { (key, value) -> put(key, value) }
                                },
                            )
                        }
                    },
                )
            }
        },
    )
}

class BleRideApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) : BleRideRemote {

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

    override suspend fun ridePermission(
        carId: String,
        actor: RideActor,
        accessories: List<BleAccessoryReport>,
    ): RiderResult<ApiOutcome> = signedApi.postOutcome(
        PATH_RIDE_PERMISSION,
        body {
            put("carId", carId)
            put("izSw", 1)
            putActor(actor)
            putAccessories(accessories)
        },
    )

    override suspend fun rideReport(carId: String, actor: RideActor): RiderResult<ApiOutcome> =
        signedApi.postOutcome(
            PATH_RIDE_REPORT,
            body {
                put("carId", carId)
                putActor(actor)
            },
        )

    override suspend fun returnReport(
        carId: String,
        orderId: String,
        actor: RideActor,
        accessories: List<BleAccessoryReport>,
    ): RiderResult<ApiOutcome> = signedApi.postOutcome(
        PATH_RETURN_REPORT,
        body {
            put("carId", carId)
            if (orderId.isNotBlank()) put("orderId", orderId)
            put("izSw", 1)
            put("returnType", ReturnByNetType.BLE_REPORT)
            putActor(actor)
            putAccessories(accessories)
            put("izFrontSuppotFullPile", true)
        },
    )

    override suspend fun tempParkReport(carId: String, actor: RideActor): RiderResult<ApiOutcome> =
        signedApi.postOutcome(
            PATH_TEMP_PARK_REPORT,
            body {
                put("carId", carId)
                putActor(actor)
            },
        )

    override suspend fun endTempParkReport(carId: String, actor: RideActor): RiderResult<ApiOutcome> =
        signedApi.postOutcome(
            PATH_END_PARK_REPORT,
            body {
                put("carId", carId)
                putActor(actor)
            },
        )

    override suspend fun returnConfig(carId: String): RiderResult<BleReturnConfig> {
        val out = signedApi.postOutcome(PATH_RETURN_CONFIG, body { put("id", carId) })
        return when (out) {
            is RiderResult.Err -> out
            is RiderResult.Ok -> RiderResult.Ok(
                BleReturnConfig(
                    needBeacon = LooseJson.bool(LooseJson.obj(out.value.data), "izBeacon"),
                ),
            )
        }
    }

    companion object {
        const val PATH_RIDE_PERMISSION: String = "client/rent/blue/ridePermission"
        const val PATH_RIDE_REPORT: String = "client/rent/blue/ride"
        const val PATH_RETURN_REPORT: String = "client/rent/blue/return"
        const val PATH_TEMP_PARK_REPORT: String = "client/rent/blue/tempParking"
        const val PATH_END_PARK_REPORT: String = "client/rent/blue/endParking"
        const val PATH_RETURN_CONFIG: String = "client/system/getbackCarConfigByCarId"
    }
}
