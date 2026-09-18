package com.luopingtech.ebike.ops.data.vehicle

import kotlinx.serialization.KSerializer
import kotlinx.serialization.descriptors.PrimitiveKind
import kotlinx.serialization.descriptors.PrimitiveSerialDescriptor
import kotlinx.serialization.descriptors.SerialDescriptor
import kotlinx.serialization.encoding.Decoder
import kotlinx.serialization.encoding.Encoder
import kotlinx.serialization.json.JsonDecoder
import kotlinx.serialization.json.JsonNull
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.longOrNull

/**
 * device/detail 的 timestamp 可能是 number 或数字字符串（秒/毫秒）。
 */
object FlexibleNullableEpochMsSerializer : KSerializer<Long?> {
    override val descriptor: SerialDescriptor =
        PrimitiveSerialDescriptor("FlexibleNullableEpochMs", PrimitiveKind.LONG)

    override fun deserialize(decoder: Decoder): Long? {
        val jsonDecoder = decoder as? JsonDecoder ?: return decoder.decodeLong()
        val element = jsonDecoder.decodeJsonElement()
        if (element is JsonNull) return null
        val primitive = element as? JsonPrimitive ?: return null
        val raw = primitive.longOrNull ?: primitive.content.toLongOrNull() ?: return null
        return normalizeEpochMs(raw)
    }

    override fun serialize(encoder: Encoder, value: Long?) {
        if (value == null) encoder.encodeNull() else encoder.encodeLong(value)
    }
}

/** Legacy TimeStampUtils: 10-digit seconds → ms. */
fun normalizeEpochMs(raw: Long): Long =
    if (raw in 1 until 10_000_000_000L) raw * 1000L else raw

fun normalizeEpochMsOrNull(raw: Long?): Long? {
    if (raw == null || raw <= 0L) return null
    return normalizeEpochMs(raw)
}
