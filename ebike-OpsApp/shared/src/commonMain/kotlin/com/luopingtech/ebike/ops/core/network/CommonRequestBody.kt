package com.luopingtech.ebike.ops.core.network

import com.luopingtech.ebike.ops.core.util.Ids
import com.luopingtech.ebike.ops.platform.DeviceInfo
import com.luopingtech.ebike.ops.platform.GeoPoint
import kotlin.concurrent.Volatile
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.JsonObjectBuilder
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.put

/**
 * Legacy ParamsUtil.getCommonParams: inject latitude/longitude when available.
 * Hosts bind [locationSnapshot] from [LocationTracker.lastKnownOrNull].
 */
object CommonRequestBody {
    @Volatile
    var locationSnapshot: () -> GeoPoint? = { null }

    fun build(
        source: String,
        tenantId: String,
        deviceInfo: DeviceInfo,
        deviceId: String,
        block: JsonObjectBuilder.() -> Unit = {},
    ): JsonObject = buildJsonObject {
        put("traceId", Ids.uuidV4())
        put("platform", deviceInfo.platform)
        put("version", deviceInfo.appVersion)
        put("source", source)
        put("stressTesting", false)
        put("tenantId", tenantId)
        put("Accept-Language", com.luopingtech.ebike.ops.core.i18n.LocaleContext.acceptLanguage)
        put("deviceId", deviceId)
        locationSnapshot()?.let { point ->
            put("latitude", "${point.latitude}")
            put("longitude", "${point.longitude}")
        }
        block()
    }

    fun toJsonString(
        source: String,
        tenantId: String,
        deviceInfo: DeviceInfo,
        deviceId: String,
        block: JsonObjectBuilder.() -> Unit = {},
    ): String = build(source, tenantId, deviceInfo, deviceId, block).toString()
}
