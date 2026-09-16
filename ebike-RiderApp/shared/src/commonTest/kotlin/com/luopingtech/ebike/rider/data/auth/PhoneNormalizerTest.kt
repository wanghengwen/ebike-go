package com.luopingtech.ebike.rider.data.auth

import kotlin.test.Test
import kotlin.test.assertEquals

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

    @Test
    fun stripsLeading86LikeUniApp() {
        assertEquals("+86-13800138000", PhoneNormalizer.toLoginPhone("8613800138000"))
        assertEquals("+86-13800138000", PhoneNormalizer.toLoginPhone("86-13800138000"))
    }
}
