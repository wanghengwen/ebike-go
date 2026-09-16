package com.luopingtech.ebike.rider.core.signing

import com.luopingtech.ebike.rider.core.config.AuthConfig
import com.luopingtech.ebike.rider.core.config.TenantConfig
import com.luopingtech.ebike.rider.core.crypto.Base64Text
import com.luopingtech.ebike.rider.core.crypto.Sha256
import com.luopingtech.ebike.rider.core.network.AuthHeaderMode
import com.luopingtech.ebike.rider.core.network.CommonRequestBody
import com.luopingtech.ebike.rider.core.network.NetworkSession
import com.luopingtech.ebike.rider.core.network.RequestAuth
import com.luopingtech.ebike.rider.platform.DeviceInfo
import com.luopingtech.ebike.rider.platform.InMemorySecureStore
import io.ktor.client.request.HttpRequestBuilder
import kotlin.test.Test
import kotlin.test.assertEquals

/**
 * 签名黄金向量。期望 hex **不是**用本仓库的 [Sha256] 算出来再回填的，而是外部
 * SHA-256（.NET `System.Security.Cryptography`）对同一串明文的结果 —— 自己算自己对
 * 只能证明代码没变，证明不了算法对。UniApp(Node) 与 Go 网关照 docs/signing.md
 * 的 payload 列直接跑，三栈应当得到同一批 hex。
 *
 * 时间戳固定，所以 [RequestAuth] 必须能注入 clock；线上那份用 `nowEpochMillis()`。
 */
class RiderSignVectorsTest {
    // ---- POST JSON: `{body}_t={timestamp}{signSecret}` ----

    @Test
    fun v1_postJsonBaseline() {
        val body = """{"phone":"13800000000"}"""
        assertEquals(
            "a57fc306490f67409b0bdf770a672b97f066e71fb62c159d2c30062664cd4c8f",
            RequestSigner.signPostJson(body, TIMESTAMP, SIGN_SECRET),
        )
        // 明文拼接顺序也钉住：换个顺序 hex 一样会变，但错在哪就看不出来了。
        assertEquals(body + "_t=" + TIMESTAMP + SIGN_SECRET, payloadOf(body))
    }

    /**
     * C 端公共体。键序就是签名的一部分：[CommonRequestBody] 用 `buildJsonObject`
     * 按写入顺序输出，改动字段顺序等于换掉所有签名。
     */
    @Test
    fun v2_commonRequestBodyKeyOrderAndSign() {
        val real = CommonRequestBody.toJsonString(
            tenantId = TENANT_ID,
            deviceInfo = FixedDeviceInfo,
            deviceId = DEVICE_ID,
        )
        // traceId 每次请求都是新的 uuid，钉成固定值才好跟别的栈比字节。
        val pinned = real.replaceFirst(
            Regex("\"traceId\":\"[^\"]+\""),
            "\"traceId\":\"$TRACE_ID\"",
        )
        assertEquals(COMMON_BODY, pinned)
        assertEquals(
            "7c7ef6681ea2e1a07df307911e51242b49f69572b880c103d5fd869a772bb298",
            RequestSigner.signPostJson(pinned, TIMESTAMP, SIGN_SECRET),
        )
    }

    @Test
    fun v3_emptyObjectBody() {
        assertEquals(
            "1039e50112b6794e94322ccd929949d58ed8e4e27df0bf71a2b26d171c2accc2",
            RequestSigner.signPostJson("{}", TIMESTAMP, SIGN_SECRET),
        )
    }

    /** UTF-8：签名必须按字节而不是按字符长度算，中文是最容易掉进 Latin-1 的一格。 */
    @Test
    fun v4_utf8ChineseBody() {
        val body = """{"name":"张三"}"""
        assertEquals(
            "162c83b932820c62e572b162addb84727b11567eb33bf5f55b328e61a21bed7b",
            RequestSigner.signPostJson(body, TIMESTAMP, SIGN_SECRET),
        )
    }

    /** `+` / `&` / `=` 不做任何转义 —— JSON 体是原样进签名串的。 */
    @Test
    fun v5_specialCharsBody() {
        val body = """{"q":"a+b&c=d","note":"1+1=2&ok"}"""
        assertEquals(
            "64d826d4ff07483ac3343fb1090919871e1a86c7f2983d4da50c9c507fa0d5c7",
            RequestSigner.signPostJson(body, TIMESTAMP, SIGN_SECRET),
        )
    }

    // ---- GET / form ----

    /** 排序 `k=v&` 拼接 + `_t={ts}{secret}`，入参顺序不影响结果。 */
    @Test
    fun v6_getSortedParams() {
        val expected = "105dbecf3895d681684f70ef6d2d1d213fd1d9e43ad4762f4fb790f994f2df6e"
        assertEquals(
            expected,
            RequestSigner.signGet(mapOf("a" to "1", "b" to "2"), TIMESTAMP, SIGN_SECRET),
        )
        assertEquals(
            expected,
            RequestSigner.signGet(mapOf("b" to "2", "a" to "1"), TIMESTAMP, SIGN_SECRET),
        )
    }

    /** 表单式：排序 `k=v&` + `time={ts}&secret={secret}`（旧 tenant/queryList 那条路）。 */
    @Test
    fun v7_formParams() {
        assertEquals(
            "96aa9481fd451e888b15f23c201b6801e3d6628780d11ec24b92723112e914c2",
            RequestSigner.signFormParams(
                mapOf(
                    "phone" to "+86-13800138000",
                    "izFilterRootTenant" to "true",
                ),
                TIMESTAMP,
                BUSINESS_SECRET,
            ),
        )
    }

    // ---- Authorization ----

    @Test
    fun v8_basicClientAuthorization() {
        val expected = "MTAwMTpiaXotc2VjcmV0"
        assertEquals(expected, Base64Text.encodeUtf8("$TENANT_ID:$BUSINESS_SECRET"))

        val auth = RequestAuth(tenantConfig(), InMemorySecureStore(), clock = { FIXED_CLOCK })
        assertEquals("Basic $expected", auth.basicClientAuthorization())
    }

    /**
     * 端到端：固定 clock 下 [RequestAuth] 写出的 `_t` / `_s` 就是 V1 那对值。
     * 这一条挂了说明 header 组装与 [RequestSigner] 走岔了，光测 signer 拦不住。
     */
    @Test
    fun requestAuth_withFixedClockEmitsGoldenHeaders() {
        val body = """{"phone":"13800000000"}"""
        val auth = RequestAuth(tenantConfig(), InMemorySecureStore(), clock = { FIXED_CLOCK })
        val builder = HttpRequestBuilder()
        auth.applyPostJson(
            builder,
            bodyJson = body,
            session = NetworkSession(),
            authMode = AuthHeaderMode.None,
        )
        assertEquals(TIMESTAMP, builder.headers[RequestSigner.HEADER_TIMESTAMP])
        assertEquals(
            "a57fc306490f67409b0bdf770a672b97f066e71fb62c159d2c30062664cd4c8f",
            builder.headers[RequestSigner.HEADER_SIGN],
        )
    }

    /** [Sha256] 自身对空串的 NIST 向量，SHA-256 实现整体跑偏时先在这里炸。 */
    @Test
    fun sha256_matchesNistEmptyStringVector() {
        assertEquals(
            "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
            Sha256.hex(""),
        )
    }

    private fun payloadOf(body: String) = body + "_t=" + TIMESTAMP + SIGN_SECRET

    private fun tenantConfig() = TenantConfig(
        tenantId = TENANT_ID,
        auth = AuthConfig(businessSecret = BUSINESS_SECRET, signSecret = SIGN_SECRET),
    )

    private object FixedDeviceInfo : DeviceInfo {
        override val platform = "android"
        override val osVersion = "14"
        override val deviceModel = "golden vector"
        override val appVersion = "1.0.0"
    }

    private companion object {
        const val TIMESTAMP = "1700000000000"
        const val FIXED_CLOCK = 1700000000000L
        const val SIGN_SECRET = "s3cret"
        const val BUSINESS_SECRET = "biz-secret"
        const val TENANT_ID = "1001"
        const val DEVICE_ID = "dev-abc"
        const val TRACE_ID = "11111111-1111-4111-8111-111111111111"
        const val COMMON_BODY =
            """{"traceId":"$TRACE_ID","platform":"android","deviceId":"$DEVICE_ID","tenantId":"$TENANT_ID"}"""
    }
}
