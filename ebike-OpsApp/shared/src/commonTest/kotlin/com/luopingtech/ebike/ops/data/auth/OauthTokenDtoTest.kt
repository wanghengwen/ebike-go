package com.luopingtech.ebike.ops.data.auth

import com.luopingtech.ebike.ops.data.api.ApiEnvelope
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlinx.serialization.json.Json

class OauthTokenDtoTest {
    private val json = Json {
        ignoreUnknownKeys = true
        isLenient = true
        coerceInputValues = true
    }

    @Test
    fun parseCamelCaseAccessToken() {
        val raw = """
            {"success":true,"code":"0","msg":"ok","data":{"accessToken":"jwt-1","refreshToken":"r1","tokenType":"bearer","expiresIn":7200}}
        """.trimIndent()
        val envelope = json.decodeFromString(ApiEnvelope.serializer(OauthTokenDto.serializer()), raw)
        assertEquals("jwt-1", envelope.data?.accessToken)
        assertEquals("bearer", envelope.data?.tokenType)
    }

    @Test
    fun parseSnakeCaseAccessToken() {
        val raw = """
            {"success":true,"code":"0","msg":"ok","data":{"access_token":"jwt-2","refresh_token":"r2","token_type":"bearer","expires_in":3600}}
        """.trimIndent()
        val envelope = json.decodeFromString(ApiEnvelope.serializer(OauthTokenDto.serializer()), raw)
        assertEquals("jwt-2", envelope.data?.accessToken)
        assertEquals("r2", envelope.data?.refreshToken)
        assertEquals(3600L, envelope.data?.expiresIn)
    }
}
