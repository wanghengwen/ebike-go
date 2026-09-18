package com.luopingtech.ebike.ops.data.fence

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.FenceBundle
import com.luopingtech.ebike.ops.domain.model.FenceNoParkingCreate
import com.luopingtech.ebike.ops.domain.model.FenceNoParkingUpdate
import com.luopingtech.ebike.ops.domain.model.FenceParkingCreate
import com.luopingtech.ebike.ops.domain.model.FenceParkingUpdate
import com.luopingtech.ebike.ops.domain.model.GeoLatLng
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.json.JsonObjectBuilder
import kotlinx.serialization.json.add
import kotlinx.serialization.json.buildJsonArray
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.put

class FenceApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    /** Legacy: /business/fence/serviceArea/getFenceByServiceId */
    suspend fun getFenceByServiceId(serviceId: String): OpsResult<FenceBundle> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/fence/serviceArea/getFenceByServiceId",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            val asLong = serviceId.toLongOrNull()
            if (asLong != null) put("id", asLong) else put("id", serviceId)
        }
        return when (
            val result = signedApi.post(
                path = "business/fence/serviceArea/getFenceByServiceId",
                bodyJson = body,
                deserializer = FenceBundleDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain())
            is OpsResult.Err -> result
        }
    }

    /** Legacy: /business/fence/serviceArea/getNearFenceByLocations */
    suspend fun getNearFenceByLocations(
        serviceId: String,
        locations: List<GeoLatLng>,
    ): OpsResult<FenceBundle> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/fence/serviceArea/getNearFenceByLocations",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("serviceId", serviceId)
            put(
                "locations",
                buildJsonArray {
                    locations.forEach { loc ->
                        add(
                            buildJsonObject {
                                put("lat", loc.lat)
                                put("lng", loc.lng)
                            },
                        )
                    }
                },
            )
        }
        return when (
            val result = signedApi.post(
                path = "business/fence/serviceArea/getNearFenceByLocations",
                bodyJson = body,
                deserializer = FenceBundleDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain())
            is OpsResult.Err -> result
        }
    }

    /** Legacy: POST /business/fence/parking/createParking */
    suspend fun createParking(req: FenceParkingCreate): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/fence/parking/createParking",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            putParkingFields(req.serviceId, req.name, req.centerLat, req.centerLng, req.points,
                req.maxParkingNumber, req.izEnable, req.bufferDistance, req.direction,
                req.tbeacon, req.rfid, req.directional, req.kickstand, req.camera,
                req.izFullPileNoStop, req.openingHoursBegin, req.openingHoursEnd)
        }
        return signedApi.postUnit(path = "business/fence/parking/createParking", bodyJson = body)
    }

    /** Legacy: POST /business/fence/noParking/createNoParking */
    suspend fun createNoParking(req: FenceNoParkingCreate): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/fence/noParking/createNoParking",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            putServiceId(req.serviceId)
            put("name", req.name)
            put("shapeType", "polygon")
            put("centerLat", req.centerLat)
            put("centerLng", req.centerLng)
            put("pointList", pointListJson(req.points))
        }
        return signedApi.postUnit(path = "business/fence/noParking/createNoParking", bodyJson = body)
    }

    /** Legacy: POST /business/fence/parking/updateParking */
    suspend fun updateParking(req: FenceParkingUpdate): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/fence/parking/updateParking",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            putId(req.id)
            putParkingFields(req.serviceId, req.name, req.centerLat, req.centerLng, req.points,
                req.maxParkingNumber, req.izEnable, req.bufferDistance, req.direction,
                req.tbeacon, req.rfid, req.directional, req.kickstand, req.camera,
                req.izFullPileNoStop, req.openingHoursBegin, req.openingHoursEnd)
        }
        return signedApi.postUnit(path = "business/fence/parking/updateParking", bodyJson = body)
    }

    /** Legacy: POST /business/fence/noParking/updateNoParking */
    suspend fun updateNoParking(req: FenceNoParkingUpdate): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/fence/noParking/updateNoParking",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            putId(req.id)
            putServiceId(req.serviceId)
            put("name", req.name)
            put("shapeType", "polygon")
            put("centerLat", req.centerLat)
            put("centerLng", req.centerLng)
            put("pointList", pointListJson(req.points))
        }
        return signedApi.postUnit(path = "business/fence/noParking/updateNoParking", bodyJson = body)
    }

    suspend fun deleteParking(id: String): OpsResult<Unit> =
        postId("business/fence/parking/deleteParking", "/business/fence/parking/deleteParking", id)

    suspend fun deleteNoParking(id: String): OpsResult<Unit> =
        postId("business/fence/noParking/deleteNoParking", "/business/fence/noParking/deleteNoParking", id)

    suspend fun enableParkingBatch(ids: List<String>): OpsResult<Unit> =
        postIds("business/fence/parking/enableBatch", "/business/fence/parking/enableBatch", ids)

    suspend fun disableParkingBatch(ids: List<String>): OpsResult<Unit> =
        postIds("business/fence/parking/disableBatch", "/business/fence/parking/disableBatch", ids)

    suspend fun deleteParkingBatch(ids: List<String>): OpsResult<Unit> =
        postIds("business/fence/parking/deleteParkingBatch", "/business/fence/parking/deleteParkingBatch", ids)

    suspend fun deleteNoParkingBatch(ids: List<String>): OpsResult<Unit> =
        postIds(
            "business/fence/noParking/deleteNoParkingBatch",
            "/business/fence/noParking/deleteNoParkingBatch",
            ids,
        )

    private suspend fun postId(path: String, source: String, id: String): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = source,
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            val asLong = id.toLongOrNull()
            if (asLong != null) put("id", asLong) else put("id", id)
        }
        return signedApi.postUnit(path = path, bodyJson = body)
    }

    private suspend fun postIds(path: String, source: String, ids: List<String>): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = source,
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put(
                "ids",
                buildJsonArray {
                    ids.forEach { id ->
                        val asLong = id.toLongOrNull()
                        if (asLong != null) add(asLong) else add(id)
                    }
                },
            )
        }
        return signedApi.postUnit(path = path, bodyJson = body)
    }

    private fun JsonObjectBuilder.putParkingFields(
        serviceId: String,
        name: String,
        centerLat: Double,
        centerLng: Double,
        points: List<GeoLatLng>,
        maxParkingNumber: Int,
        izEnable: Boolean,
        bufferDistance: Int,
        direction: Int,
        tbeacon: Boolean,
        rfid: Boolean,
        directional: Boolean,
        kickstand: Boolean,
        camera: Boolean,
        izFullPileNoStop: Boolean,
        openingHoursBegin: String,
        openingHoursEnd: String,
    ) {
        putServiceId(serviceId)
        put("name", name)
        put("shapeType", "polygon")
        put("centerLat", centerLat)
        put("centerLng", centerLng)
        put("pointList", pointListJson(points))
        put("maxParkingNumber", maxParkingNumber)
        put("izEnable", izEnable)
        // 服务端 ParkingDTO.bufferDistance / direction 为 Double
        put("bufferDistance", bufferDistance.toDouble())
        put("direction", direction.toDouble())
        put("tbeacon", tbeacon)
        put("rfid", rfid)
        put("directional", directional)
        put("izCameraDirectionalBackcar", false)
        put("izCameraPointBackcar", false)
        put("coefficientOfDifficult", 0.1)
        put("kickstand", kickstand)
        put("camera", camera)
        put("izFullPileNoStop", if (izFullPileNoStop) 1 else 0)
        // Jackson Date：必须 yyyy-MM-dd HH:mm:ss，不能只传 HH:mm:ss
        val begin = openingDateTime(openingHoursBegin.ifBlank { "00:00:00" })
        val end = openingDateTime(openingHoursEnd.ifBlank { "23:59:00" })
        put("openingHoursBegin", begin)
        put("openingHoursEnd", end)
        put("izOpenAllDay", openingHoursBegin.ifBlank { "00:00:00" } == "00:00:00" &&
            openingHoursEnd.ifBlank { "23:59:00" } == "23:59:00")
    }

    /** 对齐 ManagerApp FenceParamsModel.getOpeningBeginTime：当天日期 + 时分秒。 */
    private fun openingDateTime(hms: String): String {
        val now = com.luopingtech.ebike.ops.domain.analysis.OfflineOpsTimeRanges.formatDateTime(
            com.luopingtech.ebike.ops.core.time.nowEpochMillis(),
        )
        val date = now.substringBefore(' ')
        val time = if (hms.length >= 8) hms.take(8) else hms
        return "$date $time"
    }

    private fun JsonObjectBuilder.putServiceId(serviceId: String) {
        val asLong = serviceId.toLongOrNull()
        if (asLong != null) put("serviceId", asLong) else put("serviceId", serviceId)
    }

    private fun JsonObjectBuilder.putId(id: String) {
        val asLong = id.toLongOrNull()
        if (asLong != null) put("id", asLong) else put("id", id)
    }

    private fun pointListJson(points: List<GeoLatLng>) = buildJsonArray {
        points.forEach { p ->
            add(buildJsonArray {
                add(p.lng)
                add(p.lat)
            })
        }
    }
}
