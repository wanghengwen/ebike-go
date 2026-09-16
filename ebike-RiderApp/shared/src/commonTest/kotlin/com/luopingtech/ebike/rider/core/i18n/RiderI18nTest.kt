package com.luopingtech.ebike.rider.core.i18n

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class RiderI18nTest {
    @Test
    fun zhAndEnCatalogsCoverAllKeys() {
        val missingZh = Str.entries.filter { StringCatalogs.catalog(RiderLanguage.ZH_CN)[it] == null }
        val missingEn = Str.entries.filter { StringCatalogs.catalog(RiderLanguage.EN)[it] == null }
        assertTrue(missingZh.isEmpty(), "ZH missing: $missingZh")
        assertTrue(missingEn.isEmpty(), "EN missing: $missingEn")
    }

    @Test
    fun formatPlaceholders() {
        val zh = RiderI18n.fallback(RiderLanguage.ZH_CN)
        assertEquals("已识别车辆 · C1", zh.t(Str.ScannedVehicle, "C1"))
        val en = RiderI18n.fallback(RiderLanguage.EN)
        assertEquals("Vehicle · C1", en.t(Str.ScannedVehicle, "C1"))
    }

    @Test
    fun operationOkHasNoPlaceholders() {
        val zh = RiderI18n.fallback(RiderLanguage.ZH_CN)
        assertEquals("操作成功", zh.t(Str.OperationOk))
        val en = RiderI18n.fallback(RiderLanguage.EN)
        assertEquals("Done", en.t(Str.OperationOk))
        assertEquals("验证码已发送至 138****0000", zh.t(Str.SmsCodeSentTo, "138****0000"))
        assertEquals("Code sent to 138****0000", en.t(Str.SmsCodeSentTo, "138****0000"))
    }

    @Test
    fun acceptLanguageFollowsLocale() {
        val i18n = RiderI18n.fallback(RiderLanguage.EN)
        Strings.install(i18n)
        assertEquals("en-US", LocaleContext.acceptLanguage)
        i18n.setLanguage(RiderLanguage.ZH_CN)
        assertEquals("zh-CN", LocaleContext.acceptLanguage)
    }

    @Test
    fun fromSystemLanguage() {
        assertEquals(RiderLanguage.EN, RiderLanguage.fromSystemLanguage("en-US"))
        assertEquals(RiderLanguage.ZH_CN, RiderLanguage.fromSystemLanguage("zh-Hans-CN"))
        assertEquals(RiderLanguage.ZH_CN, RiderLanguage.fromSystemLanguage(null))
    }
}
