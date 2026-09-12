package com.luopingtech.ebike.ops.domain.vehicle

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class BatterySnNormalizerTest {
    @Test
    fun stripsBatteryPrefixAndAcceptsLengths() {
        assertEquals("12345678901", BatterySnNormalizer.normalize("battery_12345678901"))
        assertTrue(BatterySnNormalizer.isValid("12345678901"))
        assertTrue(BatterySnNormalizer.isValid("12345678901234"))
        assertTrue(BatterySnNormalizer.isValid("battery_12345678901234"))
        assertFalse(BatterySnNormalizer.isValid("12345"))
        assertFalse(BatterySnNormalizer.isValid(""))
    }
}
