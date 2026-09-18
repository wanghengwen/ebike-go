package com.luopingtech.ebike.rider.data.pay

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertNotNull
import kotlin.test.assertTrue
import kotlinx.serialization.json.Json

class WeChatPayParamsTest {
    @Test
    fun parse_appPayFields() {
        val raw = Json.parseToJsonElement(
            """
            {
              "appid": "wx123",
              "partnerid": "190000",
              "prepayid": "wx2014",
              "package": "Sign=WXPay",
              "noncestr": "abc",
              "timestamp": "1412000000",
              "sign": "ABCDEF",
              "out_trade_no": "T-1"
            }
            """.trimIndent(),
        )
        val parsed = WeChatPayParams.parse(raw)
        assertNotNull(parsed)
        assertTrue(parsed.isAppPayReady)
        assertEquals("wx123", parsed.appId)
        assertEquals("190000", parsed.partnerId)
        assertEquals("wx2014", parsed.prepayId)
        assertEquals("Sign=WXPay", parsed.packageValue)
        assertEquals("abc", parsed.nonceStr)
        assertEquals("1412000000", parsed.timeStamp)
        assertEquals("ABCDEF", parsed.sign)
        assertEquals("T-1", parsed.outTradeNo)
    }

    @Test
    fun parse_nestedPayInfoString() {
        val raw = Json.parseToJsonElement(
            """
            {
              "pay_info": "{\"partnerid\":\"m1\",\"prepayid\":\"p1\",\"noncestr\":\"n\",\"timestamp\":\"1\",\"sign\":\"s\"}"
            }
            """.trimIndent(),
        )
        val parsed = WeChatPayParams.parse(raw, fallbackAppId = "wx-fallback")
        assertNotNull(parsed)
        assertTrue(parsed.isAppPayReady)
        assertEquals("wx-fallback", parsed.appId)
        assertEquals("m1", parsed.partnerId)
    }

    @Test
    fun parse_miniProgramFields_notAppReady() {
        val raw = Json.parseToJsonElement(
            """
            {
              "timeStamp": "1",
              "nonceStr": "n",
              "package": "prepay_id=wxmini",
              "paySign": "sig",
              "signType": "MD5"
            }
            """.trimIndent(),
        )
        val parsed = WeChatPayParams.parse(raw)
        assertNotNull(parsed)
        assertFalse(parsed.isAppPayReady)
        assertEquals("sig", parsed.sign)
        assertTrue(parsed.partnerId.isBlank())
        assertTrue(parsed.prepayId.isBlank())
    }
}
