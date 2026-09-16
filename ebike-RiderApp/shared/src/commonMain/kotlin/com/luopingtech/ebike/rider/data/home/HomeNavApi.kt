package com.luopingtech.ebike.rider.data.home

import com.luopingtech.ebike.rider.core.json.LooseJson
import com.luopingtech.ebike.rider.core.network.CommonRequestBody
import com.luopingtech.ebike.rider.core.network.SignedApiClient
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.domain.home.HomeNavItem
import com.luopingtech.ebike.rider.domain.home.HomeNavJump
import com.luopingtech.ebike.rider.platform.DeviceInfo
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonArray
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.contentOrNull
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.put

/**
 * 首页底栏动态导航（UniApp `getHomeNav` → `/client/helpConfig/getHomeNavByServiceId`）。
 *
 * 注意：无 `serviceId` 时后端会报错（甚至误 INSERT）；人人骑行多个服务区现网常返回 `data:[]`。
 */
interface HomeNavRemote {
    suspend fun load(serviceId: String): RiderResult<List<HomeNavItem>>
}

class DemoHomeNavRemote : HomeNavRemote {
    override suspend fun load(serviceId: String): RiderResult<List<HomeNavItem>> =
        RiderResult.Ok(HomeNavJump.defaultItems())
}

class HomeNavApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) : HomeNavRemote {

    private val looseJson = Json { ignoreUnknownKeys = true; isLenient = true }

    override suspend fun load(serviceId: String): RiderResult<List<HomeNavItem>> {
        val sid = serviceId.trim()
        if (sid.isBlank()) {
            // 缺 serviceId 时后端 DataIntegrityViolation，直接空列表交给 UI 兜底
            return RiderResult.Ok(emptyList())
        }
        val body = CommonRequestBody.toJsonString(
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("carType", 0)
            put("serviceId", sid)
        }
        return when (val out = signedApi.postOutcome(PATH, body)) {
            is RiderResult.Err -> out
            is RiderResult.Ok -> {
                if (!out.value.success) {
                    RiderResult.Ok(emptyList())
                } else {
                    RiderResult.Ok(parseList(out.value.data))
                }
            }
        }
    }

    private fun parseList(data: kotlinx.serialization.json.JsonElement?): List<HomeNavItem> {
        val arr = when (data) {
            is JsonArray -> data
            is JsonObject -> data["list"] as? JsonArray
                ?: data["records"] as? JsonArray
                ?: data["rows"] as? JsonArray
            else -> null
        } ?: return emptyList()
        return arr.mapNotNull { el ->
            val o = LooseJson.obj(el) ?: return@mapNotNull null
            val name = LooseJson.string(o, "name", "linkTitle", "title").trim()
            if (name.isBlank()) return@mapNotNull null
            val jump = parseJumpPage(o["jumpPage"])
            HomeNavItem(
                id = LooseJson.string(o, "id"),
                name = name,
                iconUrl = LooseJson.string(o, "icon", "iconUrl", "img", "image").ifBlank {
                    LooseJson.string(jump, "icon", "iconUrl")
                },
                chainType = LooseJson.int(jump, "chainType") ?: 0,
                linkUrl = LooseJson.string(jump, "linkUrl").ifBlank {
                    LooseJson.string(o, "linkUrl")
                },
                linkTitle = LooseJson.string(jump, "linkTitle"),
                appId = LooseJson.string(jump, "appId", "appid"),
            )
        }
    }

    /** 后端偶发把 jumpPage 序列化成 JSON 字符串。 */
    private fun parseJumpPage(raw: kotlinx.serialization.json.JsonElement?): JsonObject? = when (raw) {
        is JsonObject -> raw
        is JsonPrimitive -> {
            val text = raw.contentOrNull?.trim().orEmpty()
            if (text.isBlank() || text == "null") {
                null
            } else {
                runCatching { looseJson.parseToJsonElement(text).jsonObject }.getOrNull()
            }
        }
        else -> null
    }

    companion object {
        const val PATH: String = "client/helpConfig/getHomeNavByServiceId"
    }
}
