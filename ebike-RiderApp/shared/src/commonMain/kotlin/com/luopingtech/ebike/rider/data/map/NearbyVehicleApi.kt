package com.luopingtech.ebike.rider.data.map

import com.luopingtech.ebike.rider.core.json.LooseJson
import com.luopingtech.ebike.rider.core.network.CommonRequestBody
import com.luopingtech.ebike.rider.core.network.SignedApiClient
import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.domain.model.MapPin
import com.luopingtech.ebike.rider.platform.DeviceInfo
import kotlinx.serialization.json.JsonArray
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.put

/**
 * 附近可用车辆（对齐 UniApp `getNearBike` → `/client/paas/device/eBikeLocation`）。
 */
interface NearbyVehicleRemote {
    suspend fun nearby(lat: Double, lng: Double, serviceId: String = ""): RiderResult<List<MapPin>>
}

class NearbyVehicleApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) : NearbyVehicleRemote {
    override suspend fun nearby(lat: Double, lng: Double, serviceId: String): RiderResult<List<MapPin>> {
        // 无服务区时不强拉附近车，避免错误 path / 空参打出非 JSON 再 toast。
        if (serviceId.isBlank()) return RiderResult.Ok(emptyList())
        val body = CommonRequestBody.toJsonString(
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("lat", lat)
            put("lng", lng)
            put("serviceId", serviceId)
        }
        return when (
            val result = signedApi.post(
                path = PATH,
                bodyJson = body,
                deserializer = JsonElement.serializer(),
            )
        ) {
            is RiderResult.Ok -> parsePins(result.value)
            is RiderResult.Err -> result
        }
    }

    private fun parsePins(data: JsonElement): RiderResult<List<MapPin>> {
        return try {
            val items: JsonArray = when (data) {
                is JsonArray -> data
                is JsonObject -> {
                    (data["list"] as? JsonArray)
                        ?: (data["records"] as? JsonArray)
                        ?: return RiderResult.Ok(emptyList())
                }
                else -> return RiderResult.Ok(emptyList())
            }
            val pins = items.mapNotNull { el ->
                val obj = LooseJson.obj(el) ?: return@mapNotNull null
                val id = LooseJson.string(obj, "carId", "id", "imei").ifBlank { return@mapNotNull null }
                val pinLat = LooseJson.double(obj, "latitude", "lat") ?: return@mapNotNull null
                val pinLng = LooseJson.double(obj, "longitude", "lng") ?: return@mapNotNull null
                MapPin(
                    id = id,
                    lat = pinLat,
                    lng = pinLng,
                    title = id,
                    subtitle = LooseJson.string(obj, "address"),
                    restBattery = LooseJson.int(obj, "restBattery") ?: 0,
                    ridingState = LooseJson.int(obj, "ridingState"),
                )
            }
            RiderResult.Ok(pins)
        } catch (t: Throwable) {
            RiderResult.Err(RiderError.network("nearby parse failed: ${t.message}", t))
        }
    }

    companion object {
        const val PATH: String = "client/paas/device/eBikeLocation"
    }
}

/** Demo 模式固定几辆车，方便地图 / 扫码联调。 */
class DemoNearbyVehicleRemote : NearbyVehicleRemote {
    override suspend fun nearby(lat: Double, lng: Double, serviceId: String): RiderResult<List<MapPin>> {
        val baseLat = if (lat != 0.0) lat else 28.22
        val baseLng = if (lng != 0.0) lng else 112.94
        return RiderResult.Ok(
            listOf(
                MapPin("D1001-001", baseLat + 0.001, baseLng + 0.001, "D1001-001", restBattery = 85),
                MapPin("D1001-002", baseLat - 0.0008, baseLng + 0.0015, "D1001-002", restBattery = 42),
                MapPin("D1001-003", baseLat + 0.0005, baseLng - 0.0012, "D1001-003", restBattery = 18),
            ),
        )
    }
}
