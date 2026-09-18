package com.luopingtech.ebike.rider.data.pay

import com.luopingtech.ebike.rider.core.json.LooseJson
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.contentOrNull

/**
 * 微信开放平台 APP 支付 [PayReq] 字段。
 *
 * 小程序 `requestPayment` 是另一套（timeStamp / nonceStr / package / paySign），
 * [isAppPayReady] 为 false 时不要硬调 SDK，回退 H5。
 */
data class WeChatPayParams(
    val appId: String,
    val partnerId: String,
    val prepayId: String,
    val packageValue: String,
    val nonceStr: String,
    val timeStamp: String,
    val sign: String,
    val outTradeNo: String = "",
) {
    val isAppPayReady: Boolean
        get() = partnerId.isNotBlank() &&
            prepayId.isNotBlank() &&
            nonceStr.isNotBlank() &&
            timeStamp.isNotBlank() &&
            sign.isNotBlank()

    companion object {
        private val json = Json { ignoreUnknownKeys = true; isLenient = true }

        fun parse(data: JsonElement?, fallbackAppId: String = ""): WeChatPayParams? {
            val payload = unwrap(data) ?: return null
            val outer = LooseJson.obj(data)
            val appId = LooseJson.string(payload, "appId", "appid", "app_id").ifBlank { fallbackAppId }
            val partnerId = LooseJson.string(
                payload,
                "partnerId",
                "partnerid",
                "mch_id",
                "mchId",
            )
            val prepayId = LooseJson.string(payload, "prepayId", "prepayid", "prepay_id")
            val packageValue = LooseJson.string(payload, "packageValue", "package", "packages")
                .ifBlank { "Sign=WXPay" }
            val nonceStr = LooseJson.string(payload, "nonceStr", "noncestr", "nonce_str")
            val timeStamp = LooseJson.string(payload, "timeStamp", "timestamp", "time_stamp")
            val sign = LooseJson.string(payload, "sign", "paySign", "pay_sign")
            val outTradeNo = LooseJson.string(
                payload,
                "out_trade_no",
                "outTradeNo",
                "tradeNo",
            ).ifBlank {
                LooseJson.string(outer, "out_trade_no", "outTradeNo", "tradeNo")
            }
            if (appId.isBlank() && partnerId.isBlank() && prepayId.isBlank() && sign.isBlank()) {
                return null
            }
            return WeChatPayParams(
                appId = appId,
                partnerId = partnerId,
                prepayId = prepayId,
                packageValue = packageValue,
                nonceStr = nonceStr,
                timeStamp = timeStamp,
                sign = sign,
                outTradeNo = outTradeNo,
            )
        }

        private fun unwrap(element: JsonElement?): JsonObject? {
            val decoded = decode(element) ?: return null
            for (key in NESTED_KEYS) {
                decoded[key]?.let { nested ->
                    unwrap(nested)?.let { return it }
                }
            }
            return decoded
        }

        private fun decode(element: JsonElement?): JsonObject? {
            if (element == null) return null
            LooseJson.obj(element)?.let { return it }
            val primitive = element as? JsonPrimitive ?: return null
            val text = primitive.contentOrNull?.trim().orEmpty()
            if (text.isEmpty() || text == "null") return null
            return runCatching { LooseJson.obj(json.parseToJsonElement(text)) }.getOrNull()
        }

        private val NESTED_KEYS = arrayOf(
            "miniPayRequest",
            "pay_info",
            "payInfo",
            "wc_pay_data",
            "wxPayData",
            "wx_pay",
        )
    }
}
