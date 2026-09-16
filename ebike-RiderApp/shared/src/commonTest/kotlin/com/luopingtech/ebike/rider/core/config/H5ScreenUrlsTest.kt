package com.luopingtech.ebike.rider.core.config

import com.luopingtech.ebike.rider.core.i18n.RiderLanguage
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertNull
import kotlin.test.assertTrue

class H5ScreenUrlsTest {
    @Test
    fun catalogHasFortyNineLongTailScreens() {
        assertEquals(49, H5ScreenKind.entries.size)
    }

    @Test
    fun appendQuery_goesAfterHash() {
        val out = H5ScreenUrls.appendQuery(
            "https://cdn.example/h5/index.html#/pages-sub/pay/wallet/wallet",
            mapOf("lang" to "zh-CN", "tenantId" to "1001"),
        )
        assertEquals(
            "https://cdn.example/h5/index.html#/pages-sub/pay/wallet/wallet?lang=zh-CN&tenantId=1001",
            out,
        )
    }

    @Test
    fun appendQuery_dropsSecretKeys() {
        val out = H5ScreenUrls.appendQuery(
            "https://cdn.example/h5/index.html#/help",
            mapOf(
                "lang" to "en",
                "sign" to "s3cret",
                "tenantSecret" to "biz",
                "accessToken" to "tok",
            ),
        )
        assertEquals("https://cdn.example/h5/index.html#/help?lang=en", out)
        assertFalse(out.contains("s3cret"))
        assertFalse(out.contains("tok"))
    }

    @Test
    fun resolve_blankBase_isNull() {
        assertNull(H5ScreenUrls.resolve(TenantConfig(), H5ScreenKind.Wallet, RiderLanguage.ZH_CN))
        assertFalse(H5ScreenUrls.isConfigured(TenantConfig()))
    }

    @Test
    fun resolve_doesNotLeakSecrets() {
        val config = TenantConfig(
            tenantId = "1001",
            auth = AuthConfig(businessSecret = "biz-secret", signSecret = "sign-secret"),
            branding = BrandingConfig(primaryColor = "#3AA0E8"),
            h5 = H5ScreensConfig(baseUrl = "https://cdn.example/h5/index.html#/old"),
        )
        val url = H5ScreenUrls.resolve(config, H5ScreenKind.Wallet, RiderLanguage.EN)!!
        assertTrue(url.startsWith("https://cdn.example/h5/index.html#/pages-sub/pay/wallet/wallet"))
        assertTrue(url.contains("lang=en"))
        assertTrue(url.contains("tenantId=1001"))
        assertTrue(url.contains("themeColor="))
        assertFalse(url.contains("sign-secret"))
        assertFalse(url.contains("biz-secret"))
        assertFalse(url.contains("accessToken"))
        assertFalse(url.contains("sign="))
        assertFalse(url.contains("tenantSecret"))
    }

    @Test
    fun resolveHash_acceptsBarePath() {
        val config = TenantConfig(h5 = H5ScreensConfig(baseUrl = "https://host/app/"))
        val url = H5ScreenUrls.resolveHash(
            config,
            "/pages-sub/support/help/help",
            RiderLanguage.ZH_CN,
        )!!
        assertTrue(url.contains("#/pages-sub/support/help/help"))
        assertTrue(url.contains("lang=zh-CN"))
    }

    @Test
    fun rawUrl_stripsHashFromBase() {
        val config = TenantConfig(
            h5 = H5ScreensConfig(baseUrl = "https://host/h5/index.html#/somethingElse"),
        )
        assertEquals(
            "https://host/h5/index.html#/pages-sub/support/help/help",
            H5ScreenUrls.rawUrl(config, H5ScreenKind.Help),
        )
    }
}
