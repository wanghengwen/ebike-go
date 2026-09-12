package com.luopingtech.ebike.ops.core.config

import com.luopingtech.ebike.ops.domain.model.UserSession
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNull
import kotlin.test.assertTrue

class H5ScreenUrlsTest {
    @Test
    fun appendQuery_afterHashRoute() {
        val out = H5ScreenUrls.appendQuery(
            "https://ebike.luopingtech.com/mop-saas/index.html#/newOperationScreen",
            mapOf("accessToken" to "tok", "platform" to "android"),
        )
        assertTrue(out.contains("#/newOperationScreen?"))
        assertTrue(out.contains("accessToken=tok"))
        assertTrue(out.contains("platform=android"))
    }

    @Test
    fun blankUrl_isNotConfigured() {
        val config = TenantConfig()
        assertEquals(false, H5ScreenUrls.isConfigured(config, H5ScreenKind.Operation))
        assertNull(H5ScreenUrls.resolve(config, H5ScreenKind.Operation, session = null, deviceId = "d"))
    }

    @Test
    fun resolve_appendsAuthWhenEnabled() {
        val config = TenantConfig(
            tenantId = "1003",
            api = ApiConfig(baseUrl = "https://bussiness.luopingtech.com"),
            auth = AuthConfig(businessSecret = "bs", signSecret = "ss"),
            h5 = H5ScreensConfig(
                operationUrl = H5ScreensConfig.DEFAULT_OPERATION_URL,
                appendAuthQuery = true,
            ),
        )
        val session = UserSession(
            accessToken = "a",
            refreshToken = "r",
            phone = "13800000000",
        )
        val url = H5ScreenUrls.resolve(config, H5ScreenKind.Operation, session, deviceId = "dev-1")!!
        assertTrue(url.contains("apiHost="))
        assertTrue(url.contains("accessToken=a"))
        assertTrue(url.contains("tenantId=1003"))
        assertTrue(url.contains("opMan=13800000000"))
    }

    @Test
    fun resolve_skipsAuthWhenDisabled() {
        val config = TenantConfig(
            h5 = H5ScreensConfig(
                revenueUrl = H5ScreensConfig.DEFAULT_REVENUE_URL,
                appendAuthQuery = false,
            ),
        )
        val url = H5ScreenUrls.resolve(
            config,
            H5ScreenKind.Revenue,
            session = UserSession(accessToken = "a"),
            deviceId = "d",
        )
        assertEquals(H5ScreensConfig.DEFAULT_REVENUE_URL, url)
    }

    @Test
    fun rawUrl_composesRouteOntoBaseUrl() {
        val config = TenantConfig(
            h5 = H5ScreensConfig(baseUrl = "https://host/mop-saas/index.html"),
        )
        assertEquals(
            "https://host/mop-saas/index.html#/newOperationScreen",
            H5ScreenUrls.rawUrl(config, H5ScreenKind.Operation),
        )
    }

    @Test
    fun rawUrl_stripsRouteFromMisconfiguredBaseUrl() {
        // baseUrl 本该只到 index.html，但很容易被照着大屏地址整串填进来。
        val config = TenantConfig(
            h5 = H5ScreensConfig(baseUrl = "https://host/mop-saas/index.html#/somethingElse"),
        )
        assertEquals(
            "https://host/mop-saas/index.html#/newRevenueHome",
            H5ScreenUrls.rawUrl(config, H5ScreenKind.Revenue),
        )
    }

    @Test
    fun rawUrl_derivesBaseFromRevenueUrlWhenOperationBlank() {
        // 老租户配置只填了大屏地址，没有 baseUrl，管理屏要能靠反推开出来。
        val config = TenantConfig(
            h5 = H5ScreensConfig(revenueUrl = H5ScreensConfig.DEFAULT_REVENUE_URL),
        )
        assertEquals(
            "https://ebike.luopingtech.com/mop-saas/index.html#/newOperationScreen",
            H5ScreenUrls.rawUrl(config, H5ScreenKind.Operation),
        )
    }

    @Test
    fun rawUrl_legacyUrlWinsOverBaseUrl() {
        val config = TenantConfig(
            h5 = H5ScreensConfig(
                baseUrl = "https://new-host/app/index.html",
                operationUrl = "https://legacy-host/mop-saas/index.html#/newOperationScreen",
            ),
        )
        assertEquals(
            "https://legacy-host/mop-saas/index.html#/newOperationScreen",
            H5ScreenUrls.rawUrl(config, H5ScreenKind.Operation),
        )
    }

    /** 路由必须和 `webH5/src/router/index.ts` 注册的 path 一字不差，写错只会 404。 */
    @Test
    fun rawUrl_orderScreenDerivesFromDashboardUrl() {
        val config = TenantConfig(
            h5 = H5ScreensConfig(operationUrl = H5ScreensConfig.DEFAULT_OPERATION_URL),
        )
        assertEquals(
            "https://ebike.luopingtech.com/mop-saas/index.html#/order/search",
            H5ScreenUrls.rawUrl(config, H5ScreenKind.Order),
        )
    }

    @Test
    fun rawUrl_blankWhenNothingConfigured() {
        assertEquals("", H5ScreenUrls.rawUrl(TenantConfig(), H5ScreenKind.Revenue))
    }
}
