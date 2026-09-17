package com.luopingtech.ebike.ops.data.vehicle

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsPeriod
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsTimeRanges
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.builtins.ListSerializer
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.put

/**
 * Legacy CarDetail secondary pages:
 * - scan locate: /business/ebike-management/log/list type=15500
 * - switch lock: /business/ebike-management/log/carSwitchLockLog
 * - change battery: /business/ebike-operation/change_battery/log
 * - move car: /business/ebike-operation/move_car/page
 */
class VehicleHistoryApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    private val contentJson = Json { ignoreUnknownKeys = true; isLenient = true }

    suspend fun scanLogs(
        carId: String,
        pageNum: Int = 1,
        pageSize: Int = 100,
    ): OpsResult<List<VehicleScanLogDto>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/log/list",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("pageNum", pageNum)
            put("pageSize", pageSize)
            put("type", 15500)
            put("carId", carId)
            put("reverse", true)
            // 对齐遗留 VehicleScanLogRepositoryV2：近 7 天（秒），字段名 entTime 为后端拼写。
            val endSec = OfflineOpsTimeRanges.period(OfflineOpsPeriod.Today).endMs / 1000L
            put("startTime", endSec - 604_799L)
            put("entTime", endSec)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-management/log/list",
                bodyJson = body,
                deserializer = ListSerializer(VehicleScanLogDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.map { enrichScanLog(it) })
            is OpsResult.Err -> result
        }
    }

    suspend fun switchLockLogs(carId: String): OpsResult<List<VehicleSwitchLockDto>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/log/carSwitchLockLog",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("carId", carId)
        }
        return signedApi.post(
            path = "business/ebike-management/log/carSwitchLockLog",
            bodyJson = body,
            deserializer = ListSerializer(VehicleSwitchLockDto.serializer()),
        )
    }

    suspend fun changeBatteryHistory(
        carId: String,
        serviceId: String,
        startTime: String,
        endTime: String,
        pageNum: Int = 1,
        pageSize: Int = 50,
    ): OpsResult<List<VehicleChangeBatteryHistoryDto>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/change_battery/log",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("carId", carId)
            put("latest", false)
            put("serviceId", serviceId)
            put("pageNum", pageNum)
            put("pageSize", pageSize)
            put("startTime", startTime)
            put("endTime", endTime)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-operation/change_battery/log",
                bodyJson = body,
                deserializer = VehicleChangeBatteryHistoryPageDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.list)
            is OpsResult.Err -> result
        }
    }

    suspend fun moveCarHistory(
        carId: String,
        serviceId: String,
        start: String,
        end: String,
        pageNum: Int = 1,
        pageSize: Int = 50,
    ): OpsResult<List<VehicleMoveHistoryDto>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-operation/move_car/page",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("carId", carId)
            put("serviceId", serviceId)
            put("pageNum", pageNum)
            put("pageSize", pageSize)
            put("start", start)
            put("end", end)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-operation/move_car/page",
                bodyJson = body,
                deserializer = VehicleMoveHistoryPageDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.list)
            is OpsResult.Err -> result
        }
    }

    private fun enrichScanLog(item: VehicleScanLogDto): VehicleScanLogDto {
        if (item.latitude != null && item.longitude != null) return item
        val raw = item.content.trim()
        if (raw.isBlank()) return item
        val candidates = listOf(
            raw,
            unescapeJavaLite(raw),
            unescapeJavaLite(raw).removeSurrounding("\""),
            raw.replace("\\\"", "\"").removeSurrounding("\""),
        ).distinct()
        for (text in candidates) {
            runCatching {
                contentJson.decodeFromString(VehicleScanLogContentDto.serializer(), text)
            }.getOrNull()?.let { content ->
                val lat = content.userLat
                val lng = content.userLng
                if (lat != null && lng != null) {
                    return item.copy(latitude = lat, longitude = lng)
                }
            }
        }
        return item
    }

    /** 对齐遗留 StringEscapeUtils.unescapeJava 的常见转义。 */
    private fun unescapeJavaLite(input: String): String {
        if (!input.contains('\\')) return input
        val out = StringBuilder(input.length)
        var i = 0
        while (i < input.length) {
            val c = input[i]
            if (c != '\\' || i + 1 >= input.length) {
                out.append(c)
                i++
                continue
            }
            when (val n = input[i + 1]) {
                'n' -> { out.append('\n'); i += 2 }
                'r' -> { out.append('\r'); i += 2 }
                't' -> { out.append('\t'); i += 2 }
                '"' -> { out.append('"'); i += 2 }
                '\'' -> { out.append('\''); i += 2 }
                '\\' -> { out.append('\\'); i += 2 }
                'u' -> {
                    if (i + 5 < input.length) {
                        val hex = input.substring(i + 2, i + 6)
                        val code = hex.toIntOrNull(16)
                        if (code != null) {
                            out.append(code.toChar())
                            i += 6
                        } else {
                            out.append(c)
                            i++
                        }
                    } else {
                        out.append(c)
                        i++
                    }
                }
                else -> {
                    out.append(c)
                    i++
                }
            }
        }
        return out.toString()
    }
}
