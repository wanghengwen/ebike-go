package com.luopingtech.ebike.rider.core.h5

import com.luopingtech.ebike.rider.RiderApp
import com.luopingtech.ebike.rider.core.config.ApiConfig
import com.luopingtech.ebike.rider.core.config.AuthConfig
import com.luopingtech.ebike.rider.core.config.H5ScreensConfig
import com.luopingtech.ebike.rider.core.config.TenantConfig
import com.luopingtech.ebike.rider.core.i18n.RiderLanguage
import com.luopingtech.ebike.rider.core.logging.NoOpLogger
import com.luopingtech.ebike.rider.platform.DeviceInfo
import com.luopingtech.ebike.rider.platform.InMemorySecureStore
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.contentOrNull
import kotlinx.serialization.json.jsonPrimitive
import kotlinx.serialization.json.put
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class NativeHostBridgeTest {
    @Test
    fun requestGuard_onlyClientPaths() {
        assertTrue(H5RequestGuard.isAllowed("/client/user/user/personInfo"))
        assertTrue(H5RequestGuard.isAllowed("client/order/list"))
        assertFalse(H5RequestGuard.isAllowed("/oauth/token"))
        assertFalse(H5RequestGuard.isAllowed("https://api.example/oauth/token"))
        assertFalse(H5RequestGuard.isAllowed("/business/ebike-management/file/upload"))
        assertFalse(H5RequestGuard.isAllowed(""))
        assertEquals("/client/foo", H5RequestGuard.normalizePath("client/foo?x=1"))
    }

    @Test
    fun originPolicy_matchesTenantBase() {
        val base = "https://cdn.example/h5/index.html"
        assertTrue(H5OriginPolicy.isAllowed("https://cdn.example/h5/index.html#/wallet", base))
        assertFalse(H5OriginPolicy.isAllowed("https://evil.example/h5/index.html", base))
        assertFalse(H5OriginPolicy.isAllowed("about:blank", base))
        assertTrue(H5OriginPolicy.isAllowed("file:///android_asset/h5/index.html", ""))
        assertEquals("https://cdn.example", H5OriginPolicy.originOf(base))
    }

    @Test
    fun profile_hasDisplayFieldsOnly() = runTest {
        val app = riderApp()
        app.authRepository.loginWithSms("13800000000", "000000")
        val bridge = NativeHostBridge(app, NativeHostHooks())
        bridge.currentPageUrl = { "https://cdn.example/h5/index.html" }
        val profile = bridge.dispatch("getProfile", buildJsonObject {})
        assertFalse(H5ProfileView.containsSecrets(profile))
        assertEquals("demo-user", profile["userId"]?.jsonPrimitive?.contentOrNull)
        assertEquals(true, profile["loggedIn"]?.toString()?.contains("true"))
        assertFalse(profile.containsKey("accessToken"))
        assertFalse(profile.containsKey("refreshToken"))
        app.close()
    }

    @Test
    fun request_rejectsOauthAndPayIsPlaceholder() = runTest {
        val app = riderApp()
        val bridge = NativeHostBridge(app, NativeHostHooks())
        bridge.currentPageUrl = { "https://cdn.example/h5/index.html" }

        val forbidden = bridge.dispatch(
            "request",
            buildJsonObject { put("url", "/oauth/token") },
        )
        assertEquals("FORBIDDEN", forbidden["code"]?.jsonPrimitive?.contentOrNull)

        val pay = bridge.dispatch("pay", buildJsonObject {})
        assertEquals("UNSUPPORTED", pay["code"]?.jsonPrimitive?.contentOrNull)
        assertEquals(false, pay["success"]?.toString()?.contains("true") == true)

        val demo = bridge.dispatch(
            "request",
            buildJsonObject { put("url", "/client/user/user/personInfo") },
        )
        assertEquals("DEMO", demo["code"]?.jsonPrimitive?.contentOrNull)
        app.close()
    }

    @Test
    fun language_roundTripAndOriginReject() = runTest {
        val app = riderApp()
        val bridge = NativeHostBridge(app, NativeHostHooks())
        bridge.currentPageUrl = { "https://cdn.example/h5/index.html" }
        val set = bridge.dispatch(
            "setLanguage",
            buildJsonObject { put("language", "en-US") },
        )
        assertEquals(RiderLanguage.EN, app.i18n.language)
        assertEquals("en", set["language"]?.jsonPrimitive?.contentOrNull)

        bridge.currentPageUrl = { "https://evil.example/" }
        val blocked = bridge.dispatch("getProfile", buildJsonObject {})
        assertEquals("ORIGIN", blocked["code"]?.jsonPrimitive?.contentOrNull)
        app.close()
    }

    @Test
    fun handlePayload_repliesWithId() = runTest {
        val app = riderApp()
        val bridge = NativeHostBridge(app, NativeHostHooks())
        bridge.currentPageUrl = { "https://cdn.example/h5/index.html" }
        val (id, body) = bridge.handlePayload(
            """{"id":"7","method":"getLanguage","args":{}}""",
        )
        assertEquals("7", id)
        assertTrue(body.contains("zh-CN") || body.contains("en"))
        val js = bridge.replyJs(id, body)
        assertTrue(js.contains("__riderNativeOnResult"))
        assertTrue(js.contains("\"7\""))
        app.close()
    }

    private fun riderApp(): RiderApp = RiderApp.create(
        config = TenantConfig(
            alias = "test",
            tenantId = "1001",
            api = ApiConfig(baseUrl = ""),
            auth = AuthConfig(businessSecret = "biz", signSecret = "s3cret"),
            h5 = H5ScreensConfig(baseUrl = "https://cdn.example/h5/index.html"),
        ),
        secureStore = InMemorySecureStore(),
        logger = NoOpLogger,
        appVersion = "0.0.0",
        deviceInfo = FakeDeviceInfo,
        systemLanguage = null,
        quickLogin = com.luopingtech.ebike.rider.platform.UnsupportedQuickLogin,
        codeScanner = com.luopingtech.ebike.rider.platform.UnsupportedCodeScanner(),
        photoCapture = com.luopingtech.ebike.rider.platform.UnsupportedPhotoCapture(),
        locationTracker = null,
        reverseGeocoder = null,
        mediaUploader = null,
        bleTransport = null,
        nativeBle = null,
    )

    private object FakeDeviceInfo : DeviceInfo {
        override val platform = "test"
        override val osVersion = "0"
        override val deviceModel = "unit test"
        override val appVersion = "0.0.0"
    }
}
