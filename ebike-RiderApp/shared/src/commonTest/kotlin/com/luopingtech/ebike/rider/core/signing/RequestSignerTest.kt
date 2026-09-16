package com.luopingtech.ebike.rider.core.signing

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class RequestSignerTest {
    @Test
    fun getSign_isStableForSortedParams() {
        val secret = "test-secret"
        val timestamp = "1700000000000"
        val a = RequestSigner.signGet(
            mapOf("b" to "2", "a" to "1"),
            timestamp,
            secret,
        )
        val b = RequestSigner.signGet(
            mapOf("a" to "1", "b" to "2"),
            timestamp,
            secret,
        )
        assertEquals(a, b)
        assertEquals(64, a.length)
        assertTrue(a.all { it in '0'..'9' || it in 'a'..'f' })
    }

    @Test
    fun postSign_matchesKnownVector() {
        // body + "_t=" + timestamp + secret, then SHA-256 hex
        val body = """{"phone":"13800000000"}"""
        val timestamp = "1700000000000"
        val secret = "s3cret"
        val expectedPayload = body + "_t=" + timestamp + secret
        val sign = RequestSigner.signPostJson(body, timestamp, secret)
        assertEquals(
            com.luopingtech.ebike.rider.core.crypto.Sha256.hex(expectedPayload),
            sign,
        )
    }

    @Test
    fun formSign_matchesLegacyTenantQueryStyle() {
        val timestamp = "1700000000000"
        val secret = "biz-secret"
        val params = mapOf(
            "izFilterRootTenant" to "true",
            "traceId" to "t-1",
            "phone" to "+86-13800138000",
        )
        val expected = buildString {
            append("izFilterRootTenant=true&")
            append("phone=+86-13800138000&")
            append("traceId=t-1&")
            append("time=").append(timestamp).append("&secret=").append(secret)
        }
        assertEquals(
            com.luopingtech.ebike.rider.core.crypto.Sha256.hex(expected),
            RequestSigner.signFormParams(params, timestamp, secret),
        )
    }
}
