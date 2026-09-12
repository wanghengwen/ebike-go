package com.luopingtech.ebike.ops.data.auth

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class PhoneNormalizerTest {
    @Test
    fun prependsDefaultAreaCode() {
        assertEquals("+86-13800138000", PhoneNormalizer.toLoginPhone("13800138000"))
    }

    @Test
    fun keepsPlusPrefix() {
        assertEquals("+86-13800138000", PhoneNormalizer.toLoginPhone("+86-13800138000"))
    }

    @Test
    fun usesSelectedAreaCode() {
        assertEquals("+48-501234567", PhoneNormalizer.toLoginPhone("501234567", "+48"))
        assertEquals("+48-501234567", PhoneNormalizer.toLoginPhone("501234567", "48"))
    }

    @Test
    fun normalizeAreaCodeStripsSpacesAndDoublePlus() {
        assertEquals("+86", PhoneNormalizer.normalizeAreaCode("++86"))
        assertEquals("+5999", PhoneNormalizer.normalizeAreaCode("+599 9"))
    }
}

class CallingCodeCatalogTest {
    @Test
    fun loadsLegacyCatalogWithCnPinned() {
        val all = CallingCodeCatalog.all()
        assertTrue(all.size > 100)
        assertEquals("CN", all.first().regionCode)
        assertEquals("+86", all.first().dialCode)
    }

    @Test
    fun filterByDialOrName() {
        assertTrue(CallingCodeCatalog.filter("48").any { it.regionCode == "PL" })
        assertTrue(CallingCodeCatalog.filter("Poland").any { it.regionCode == "PL" })
    }
}
