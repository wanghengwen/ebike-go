package com.luopingtech.ebike.ops.core.i18n

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class OpsI18nTest {
    @Test
    fun zhAndEnCatalogsCoverAllKeys() {
        val missingZh = Str.entries.filter { StringCatalogs.catalog(OpsLanguage.ZH_CN)[it] == null }
        val missingEn = Str.entries.filter { StringCatalogs.catalog(OpsLanguage.EN)[it] == null }
        assertTrue(missingZh.isEmpty(), "ZH missing: $missingZh")
        assertTrue(missingEn.isEmpty(), "EN missing: $missingEn")
    }

    @Test
    fun formatPlaceholders() {
        val i18n = OpsI18n.fallback(OpsLanguage.ZH_CN)
        assertEquals("已识别车辆 · C1", i18n.t(Str.ResolvedVehicle, "C1"))
        val en = OpsI18n.fallback(OpsLanguage.EN)
        assertEquals("Vehicle · C1", en.t(Str.ResolvedVehicle, "C1"))
    }

    @Test
    fun operationOkHasNoPlaceholders() {
        val zh = OpsI18n.fallback(OpsLanguage.ZH_CN)
        assertEquals("操作成功", zh.t(Str.OperationOk))
        val en = OpsI18n.fallback(OpsLanguage.EN)
        assertEquals("Done", en.t(Str.OperationOk))
        assertEquals("已领取 3 项", zh.t(Str.ClaimOk, 3))
        assertEquals("Claimed 3", en.t(Str.ClaimOk, 3))
    }

    @Test
    fun acceptLanguageFollowsLocale() {
        val i18n = OpsI18n.fallback(OpsLanguage.EN)
        Strings.install(i18n)
        assertEquals("en-US", LocaleContext.acceptLanguage)
        i18n.setLanguage(OpsLanguage.ZH_CN)
        assertEquals("zh-CN", LocaleContext.acceptLanguage)
    }

    @Test
    fun fromSystemLanguage() {
        assertEquals(OpsLanguage.EN, OpsLanguage.fromSystemLanguage("en-US"))
        assertEquals(OpsLanguage.ZH_CN, OpsLanguage.fromSystemLanguage("zh-Hans-CN"))
        assertEquals(OpsLanguage.ZH_CN, OpsLanguage.fromSystemLanguage(null))
    }
}
