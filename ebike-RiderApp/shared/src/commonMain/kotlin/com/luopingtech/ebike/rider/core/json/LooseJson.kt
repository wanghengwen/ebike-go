package com.luopingtech.ebike.rider.core.json

import kotlinx.serialization.json.JsonArray
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.contentOrNull
import kotlinx.serialization.json.jsonObject

/**
 * 宽松字段读取。
 *
 * 骑行域这些老接口同一个字段会在不同租户 / 不同版本里换类型：`izCanReturn` 可能是
 * `true` / `1` / `"1"`，`returnType` 可能是 `2101` 或 `"2101"`，`costFee` 可能带小数。
 * `@Serializable` DTO 撞上任何一种就整包解析失败，而这些接口正是还车主路径 ——
 * 宁可逐字段兜底，也不能因为一个类型漂移把用户卡在骑行页。
 *
 * 每个读取器都接受**多个候选键**，因为旧版同一语义有多套命名（`lat` / `latitude` / `carLat`）。
 */
object LooseJson {

    fun obj(element: JsonElement?): JsonObject? = when (element) {
        null -> null
        is JsonObject -> element
        else -> runCatching { element.jsonObject }.getOrNull()
    }

    fun array(element: JsonElement?, vararg keys: String): JsonArray? {
        if (element is JsonArray) return element
        val o = obj(element) ?: return null
        for (key in keys) {
            (o[key] as? JsonArray)?.let { return it }
        }
        return null
    }

    fun string(o: JsonObject?, vararg keys: String): String {
        val raw = primitive(o, *keys) ?: return ""
        val text = raw.contentOrNull.orEmpty().trim()
        // 后端偶发把缺失值序列化成字符串 "null"，直接当空。
        return if (text == "null" || text == "undefined") "" else text
    }

    fun int(o: JsonObject?, vararg keys: String): Int? = double(o, *keys)?.let {
        if (it.isNaN() || it.isInfinite()) null else it.toInt()
    }

    fun long(o: JsonObject?, vararg keys: String): Long? = double(o, *keys)?.let {
        if (it.isNaN() || it.isInfinite()) null else it.toLong()
    }

    fun double(o: JsonObject?, vararg keys: String): Double? {
        val raw = primitive(o, *keys) ?: return null
        raw.contentOrNull?.let { text ->
            val trimmed = text.trim()
            if (trimmed.isEmpty() || trimmed == "null") return null
            trimmed.toDoubleOrNull()?.let { return it }
            // "true"/"false" 当 1/0，旧版 Java 端偶尔这么发布尔位。
            if (trimmed.equals("true", ignoreCase = true)) return 1.0
            if (trimmed.equals("false", ignoreCase = true)) return 0.0
        }
        return null
    }

    /** `true` / `1` / `"1"` / `"true"` 都算真；缺失算假。 */
    fun bool(o: JsonObject?, vararg keys: String): Boolean {
        val raw = primitive(o, *keys) ?: return false
        val text = raw.contentOrNull?.trim().orEmpty()
        if (text.equals("true", ignoreCase = true)) return true
        if (text.equals("false", ignoreCase = true)) return false
        val n = text.toDoubleOrNull() ?: return false
        return n != 0.0
    }

    /** 区分「显式 false」与「字段不存在」的场合用这个。 */
    fun boolOrNull(o: JsonObject?, vararg keys: String): Boolean? {
        primitive(o, *keys) ?: return null
        return bool(o, *keys)
    }

    fun nested(o: JsonObject?, vararg keys: String): JsonObject? {
        val root = o ?: return null
        for (key in keys) {
            (root[key] as? JsonObject)?.let { return it }
        }
        return null
    }

    private fun primitive(o: JsonObject?, vararg keys: String): JsonPrimitive? {
        val root = o ?: return null
        for (key in keys) {
            val element = root[key] ?: continue
            val p = element as? JsonPrimitive ?: continue
            if (p.contentOrNull == null) continue
            return p
        }
        return null
    }
}
