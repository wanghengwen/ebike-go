package com.luopingtech.ebike.rider.data.fence

import com.luopingtech.ebike.rider.core.json.LooseJson
import com.luopingtech.ebike.rider.core.network.CommonRequestBody
import com.luopingtech.ebike.rider.core.network.SignedApiClient
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.domain.model.FenceKind
import com.luopingtech.ebike.rider.domain.model.FencePolygon
import com.luopingtech.ebike.rider.domain.model.GeoLatLng
import com.luopingtech.ebike.rider.domain.riding.NearParking
import com.luopingtech.ebike.rider.platform.DeviceInfo
import com.luopingtech.ebike.rider.platform.GeoPoint
import kotlinx.serialization.json.JsonArray
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.put

/**
 * 围栏 / 停车点（UniApp `api/map.ts` 的围栏子集）。
 *
 * 骑行域只需要三件事：我在哪个服务区、附近的围栏多边形、最近的 P 点。
 * 首页运营位、活动入口那些留给别的阶段。
 */
interface FenceRemote {
    /** 当前坐标属于哪个服务区。空串 = 不在任何服务区内。 */
    suspend fun serviceAreaIdAt(at: GeoPoint): RiderResult<String>

    /** 骑行页 / 找 P 页要画的围栏。 */
    suspend fun nearFences(at: GeoPoint, serviceId: String): RiderResult<List<FencePolygon>>

    suspend fun nearParking(at: GeoPoint, serviceId: String): RiderResult<NearParking>
}

class FenceApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) : FenceRemote {

    override suspend fun serviceAreaIdAt(at: GeoPoint): RiderResult<String> {
        val body = base {
            put("lat", at.latitude)
            put("lng", at.longitude)
        }
        return when (val out = signedApi.postOutcome(PATH_SERVICE_BY_LOCATION, body)) {
            is RiderResult.Err -> out
            is RiderResult.Ok ->
                if (!out.value.success) {
                    RiderResult.Ok("")
                } else {
                    RiderResult.Ok(LooseJson.string(LooseJson.obj(out.value.data), "id", "serviceId"))
                }
        }
    }

    override suspend fun nearFences(at: GeoPoint, serviceId: String): RiderResult<List<FencePolygon>> {
        val body = base {
            put("lat", at.latitude)
            put("lng", at.longitude)
            if (serviceId.isNotBlank()) put("serviceId", serviceId)
        }
        return when (val out = signedApi.postOutcome(PATH_NEAR_FENCE, body)) {
            is RiderResult.Err -> out
            is RiderResult.Ok -> RiderResult.Ok(FenceParsers.polygons(out.value.data))
        }
    }

    override suspend fun nearParking(at: GeoPoint, serviceId: String): RiderResult<NearParking> {
        // 这个接口的坐标裹在 locationDTO 里，而且键是 lng/lat —— 不是别处的 userLng/userLat。
        val body = base {
            put(
                "locationDTO",
                buildJsonObject {
                    put("lng", at.longitude)
                    put("lat", at.latitude)
                },
            )
            if (serviceId.isNotBlank()) put("serviceId", serviceId)
        }
        return when (val out = signedApi.postOutcome(PATH_NEAR_PARKING_NUM, body)) {
            is RiderResult.Err -> out
            is RiderResult.Ok ->
                if (!out.value.success) {
                    RiderResult.Ok(NearParking())
                } else {
                    RiderResult.Ok(FenceParsers.nearParking(out.value.data))
                }
        }
    }

    private fun base(block: kotlinx.serialization.json.JsonObjectBuilder.() -> Unit): String =
        CommonRequestBody.toJsonString(
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
            block = block,
        )

    companion object {
        const val PATH_SERVICE_BY_LOCATION: String = "client/fence/serviceArea/getByLocation"
        const val PATH_NEAR_FENCE: String = "client/fence/serviceArea/getNearFence"
        const val PATH_FENCE_BY_SERVICE: String = "client/fence/serviceArea/getFenceByServiceId"
        const val PATH_NEAR_PARKING_NUM: String = "client/fence/parking/nearParkingNum"
    }
}

object FenceParsers {

    /**
     * 围栏 `type`：1 禁停、2 停车点，其余按服务区。旧版把这个映射抄在三个页面里，
     * 每处都少认一两个值。
     */
    fun kindOf(type: Int?): FenceKind = when (type) {
        1 -> FenceKind.NoParking
        2 -> FenceKind.Parking
        else -> FenceKind.ServiceArea
    }

    fun polygons(data: JsonElement?): List<FencePolygon> {
        val items = LooseJson.array(data, "list", "records", "fenceList") ?: return emptyList()
        return items.mapIndexedNotNull { index, element ->
            val o = element as? JsonObject ?: return@mapIndexedNotNull null
            val points = points(o)
            if (points.size < 3) return@mapIndexedNotNull null
            FencePolygon(
                id = LooseJson.string(o, "id", "fenceId").ifBlank { "fence-$index" },
                name = LooseJson.string(o, "name", "fenceName"),
                points = points,
                kind = kindOf(LooseJson.int(o, "type", "fenceType")),
            )
        }
    }

    /**
     * 顶点数组的形状在租户之间不统一：可能是 `[{lat,lng}]`、`[{latitude,longitude}]`，
     * 也可能是 `"lng,lat;lng,lat"` 这种旧 GIS 串。三种都认。
     */
    private fun points(o: JsonObject): List<GeoLatLng> {
        val array = LooseJson.array(o["points"], "points")
            ?: LooseJson.array(o["fencePoints"], "fencePoints")
            ?: LooseJson.array(o["coordinates"], "coordinates")
        if (array is JsonArray) {
            return array.mapNotNull { element ->
                val po = element as? JsonObject ?: return@mapNotNull null
                val lat = LooseJson.double(po, "lat", "latitude") ?: return@mapNotNull null
                val lng = LooseJson.double(po, "lng", "longitude") ?: return@mapNotNull null
                GeoLatLng(lat, lng)
            }
        }
        val raw = LooseJson.string(o, "fenceStr", "polygon", "points")
        if (raw.isBlank()) return emptyList()
        return raw.split(';', '|').mapNotNull { pair ->
            val parts = pair.split(',')
            if (parts.size < 2) return@mapNotNull null
            // GIS 串是 lng,lat 顺序，反了会把车画到南极。
            val lng = parts[0].trim().toDoubleOrNull() ?: return@mapNotNull null
            val lat = parts[1].trim().toDoubleOrNull() ?: return@mapNotNull null
            GeoLatLng(lat, lng)
        }
    }

    fun nearParking(data: JsonElement?): NearParking {
        // 有的租户直接回一个数字（只有数量）。
        (data as? kotlinx.serialization.json.JsonPrimitive)?.let { primitive ->
            val count = primitive.content.trim().toIntOrNull() ?: 0
            return NearParking(count = count)
        }
        val o = LooseJson.obj(data) ?: return NearParking()
        val nearest = LooseJson.nested(o, "nearest", "parking") ?: o
        return NearParking(
            count = LooseJson.int(o, "parkingNum", "num", "count") ?: 0,
            distanceMeters = LooseJson.double(o, "distance"),
            lat = LooseJson.double(nearest, "lat", "latitude", "centerLat"),
            lng = LooseJson.double(nearest, "lng", "longitude", "centerLng"),
            name = LooseJson.string(nearest, "name", "title"),
            outOfService = LooseJson.bool(o, "izOutService", "outOfService", "isOutService"),
        )
    }
}
