package com.luopingtech.ebike.rider.core.network

import com.luopingtech.ebike.rider.core.util.Ids
import com.luopingtech.ebike.rider.platform.DeviceInfo
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.JsonObjectBuilder
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.put

/**
 * C 端公共请求体字段。
 *
 * 只有四个字段：[traceId] 排查链路、platform 与 deviceId 定位设备、tenantId 分租户。
 * Accept-Language 不进 body —— 它是 HTTP 语义，由 [RequestAuth] 写在 header 上，
 * 放进 body 还会一起进签名串，换个语言签名就变了。
 */
object CommonRequestBody {
    fun build(
        tenantId: String,
        deviceInfo: DeviceInfo,
        deviceId: String,
        block: JsonObjectBuilder.() -> Unit = {},
    ): JsonObject = buildJsonObject {
        put("traceId", Ids.uuidV4())
        put("platform", deviceInfo.platform)
        put("deviceId", deviceId)
        put("tenantId", tenantId)
        block()
    }

    fun toJsonString(
        tenantId: String,
        deviceInfo: DeviceInfo,
        deviceId: String,
        block: JsonObjectBuilder.() -> Unit = {},
    ): String = build(tenantId, deviceInfo, deviceId, block).toString()
}
