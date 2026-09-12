package com.luopingtech.ebike.ops.core.network

import com.luopingtech.ebike.ops.core.config.AuthConfig
import com.luopingtech.ebike.ops.core.config.TenantConfig
import com.luopingtech.ebike.ops.core.crypto.Base64Text
import com.luopingtech.ebike.ops.core.signing.RequestSigner
import com.luopingtech.ebike.ops.platform.InMemorySecureStore
import com.luopingtech.ebike.ops.platform.SecureStore
import io.ktor.client.request.HttpRequestBuilder
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class RequestAuthTest {
    @Test
    fun basicClientAuthorization_matchesLegacyFormat() {
        val config = TenantConfig(
            tenantId = "tenant-1",
            auth = AuthConfig(businessSecret = "secret-1", signSecret = "sign"),
        )
        val auth = RequestAuth(config, InMemorySecureStore())
        val expected = "Basic " + Base64Text.encodeUtf8("tenant-1:secret-1")
        assertEquals(expected, auth.basicClientAuthorization())
        assertTrue(auth.basicClientAuthorization().startsWith("Basic "))
    }

    @Test
    fun bearerAuthorization_alwaysUsesCapitalBearerScheme() {
        val config = TenantConfig(
            tenantId = "1",
            auth = AuthConfig(businessSecret = "s", signSecret = "sign"),
        )
        val store = InMemorySecureStore()
        store.putString(SecureStore.KEY_ACCESS_TOKEN, "stale-token")
        val auth = RequestAuth(config, store)
        val builder = HttpRequestBuilder()
        auth.applyPostJson(
            builder,
            bodyJson = "{}",
            session = NetworkSession(accessToken = "fresh-jwt", tokenType = "bearer"),
            authMode = AuthHeaderMode.Bearer,
        )
        assertEquals("Bearer fresh-jwt", builder.headers[RequestSigner.HEADER_AUTHORIZATION])
    }

    @Test
    fun bearerAuthorization_fallsBackToStoreWhenSessionBlank() {
        val config = TenantConfig(
            tenantId = "1",
            auth = AuthConfig(businessSecret = "s", signSecret = "sign"),
        )
        val store = InMemorySecureStore()
        store.putString(SecureStore.KEY_ACCESS_TOKEN, "stored-jwt")
        val auth = RequestAuth(config, store)
        val builder = HttpRequestBuilder()
        auth.applyPostJson(
            builder,
            bodyJson = "{}",
            session = NetworkSession(accessToken = "", tokenType = "bearer"),
            authMode = AuthHeaderMode.Bearer,
        )
        assertEquals("Bearer stored-jwt", builder.headers[RequestSigner.HEADER_AUTHORIZATION])
    }
}
