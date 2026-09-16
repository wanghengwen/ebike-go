package com.luopingtech.ebike.rider.core.network

import com.luopingtech.ebike.rider.core.config.AuthConfig
import com.luopingtech.ebike.rider.core.config.TenantConfig
import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.platform.InMemorySecureStore
import com.luopingtech.ebike.rider.platform.SecureStore
import io.ktor.client.HttpClient
import io.ktor.client.engine.mock.MockEngine
import io.ktor.client.engine.mock.respond
import io.ktor.http.HttpHeaders
import io.ktor.http.HttpStatusCode
import io.ktor.http.headersOf
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import kotlinx.coroutines.async
import kotlinx.coroutines.delay
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.json.Json

class SignedApiClientAuthTest {
    @Test
    fun code00015_clearsSessionWithoutRefresh() = runTest {
        val events = mutableListOf<String>()
        var refreshCalls = 0
        val client = signedClient(
            handler = { envelope("00015", "other device") },
            sessionToken = "old-token",
            refresh = {
                refreshCalls++
                RiderResult.Ok("new-token")
            },
            onInvalid = { code, _ -> events += code },
        )

        val result = client.postUnit("client/ping", "{}")
        assertTrue(result is RiderResult.Err)
        assertEquals(listOf("00015"), events)
        assertEquals(0, refreshCalls)
    }

    @Test
    fun code00005_refreshesOnceThenRetries() = runTest {
        var calls = 0
        var refreshCalls = 0
        var stored = "old-token"
        val client = signedClient(
            handler = { request ->
                calls++
                val auth = request.headers[HttpHeaders.Authorization].orEmpty()
                if (auth == "Bearer new-token") {
                    """{"success":true,"code":"0","msg":"ok","data":{}}"""
                } else {
                    envelope("00005", "expired")
                }
            },
            sessionToken = { stored },
            refresh = {
                refreshCalls++
                stored = "new-token"
                RiderResult.Ok("new-token")
            },
        )

        val result = client.postUnit("client/ping", "{}")
        assertTrue(result is RiderResult.Ok)
        assertEquals(1, refreshCalls)
        assertEquals(2, calls)
    }

    @Test
    fun code00013_refreshesLike00005() = runTest {
        var stored = "old-token"
        var refreshCalls = 0
        val client = signedClient(
            handler = { request ->
                if (request.headers[HttpHeaders.Authorization] == "Bearer new-token") {
                    """{"success":true,"code":"0","msg":"ok","data":{}}"""
                } else {
                    envelope("00013", "invalid")
                }
            },
            sessionToken = { stored },
            refresh = {
                refreshCalls++
                stored = "new-token"
                RiderResult.Ok("new-token")
            },
        )
        assertTrue(client.postUnit("client/ping", "{}") is RiderResult.Ok)
        assertEquals(1, refreshCalls)
    }

    @Test
    fun concurrent00005_singleFlightRefresh() = runTest {
        var refreshCalls = 0
        var stored = "old-token"
        val client = signedClient(
            handler = { request ->
                if (request.headers[HttpHeaders.Authorization] == "Bearer new-token") {
                    """{"success":true,"code":"0","msg":"ok","data":{}}"""
                } else {
                    envelope("00005", "expired")
                }
            },
            sessionToken = { stored },
            refresh = {
                refreshCalls++
                delay(40)
                stored = "new-token"
                RiderResult.Ok("new-token")
            },
        )

        val a = async { client.postUnit("client/a", "{}") }
        val b = async { client.postUnit("client/b", "{}") }
        assertTrue(a.await() is RiderResult.Ok)
        assertTrue(b.await() is RiderResult.Ok)
        assertEquals(1, refreshCalls)
    }

    @Test
    fun code00006_guestDoesNotClearSession() = runTest {
        val events = mutableListOf<String>()
        val client = signedClient(
            handler = { envelope("00006", "no auth") },
            sessionToken = "",
            onInvalid = { code, _ -> events += code },
        )
        val result = client.postUnit("client/ping", "{}")
        assertTrue(result is RiderResult.Err)
        assertTrue(events.isEmpty())
    }

    @Test
    fun code00006_withTokenClearsSession() = runTest {
        val events = mutableListOf<String>()
        val client = signedClient(
            handler = { envelope("00006", "bad bearer") },
            sessionToken = "stale",
            onInvalid = { code, _ -> events += code },
        )
        val result = client.postUnit("client/ping", "{}")
        assertTrue(result is RiderResult.Err)
        assertEquals(listOf("00006"), events)
    }

    @Test
    fun badRefresh_notifiesSessionInvalid() = runTest {
        val events = mutableListOf<String>()
        val client = signedClient(
            handler = { envelope("00005", "expired") },
            sessionToken = "old-token",
            refresh = { RiderResult.Err(RiderError.unauthorized("refresh failed")) },
            onInvalid = { _, message -> events += message },
        )
        assertTrue(client.postUnit("client/ping", "{}") is RiderResult.Err)
        assertEquals(listOf("refresh failed"), events)
    }

    private fun envelope(code: String, msg: String): String =
        """{"success":false,"code":"$code","msg":"$msg","data":null}"""

    private fun signedClient(
        handler: (io.ktor.client.request.HttpRequestData) -> String,
        sessionToken: String,
        refresh: suspend () -> RiderResult<String> = { RiderResult.Ok("new-token") },
        onInvalid: (String, String) -> Unit = { _, _ -> },
    ): SignedApiClient = signedClient(
        handler = handler,
        sessionToken = { sessionToken },
        refresh = refresh,
        onInvalid = onInvalid,
    )

    private fun signedClient(
        handler: (io.ktor.client.request.HttpRequestData) -> String,
        sessionToken: () -> String,
        refresh: suspend () -> RiderResult<String>,
        onInvalid: (String, String) -> Unit = { _, _ -> },
    ): SignedApiClient {
        val engine = MockEngine { request ->
            respond(
                content = handler(request),
                status = HttpStatusCode.OK,
                headers = headersOf(HttpHeaders.ContentType, "application/json"),
            )
        }
        val http = HttpClient(engine)
        val store = InMemorySecureStore()
        store.putString(SecureStore.KEY_TENANT_ID, "1")
        val auth = RequestAuth(
            TenantConfig(tenantId = "1", auth = AuthConfig(businessSecret = "s", signSecret = "sign")),
            store,
        )
        return SignedApiClient(
            client = http,
            requestAuth = auth,
            json = Json { ignoreUnknownKeys = true; isLenient = true; coerceInputValues = true },
            sessionProvider = { NetworkSession(accessToken = sessionToken()) },
            refreshAccessToken = refresh,
            onSessionInvalid = onInvalid,
        )
    }
}
