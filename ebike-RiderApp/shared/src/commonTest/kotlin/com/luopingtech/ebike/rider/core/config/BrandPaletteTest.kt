package com.luopingtech.ebike.rider.core.config

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNull

class BrandPaletteTest {
    @Test
    fun parseHexVariants() {
        assertEquals(0xFF00A0E8L, ColorHex.parseArgb("#00A0E8"))
        assertEquals(0x1A00A0E8L, ColorHex.parseArgb("#1A00A0E8"))
        assertEquals(0xFF112233L, ColorHex.parseArgb("123"))
        assertNull(ColorHex.parseArgb(""))
        assertNull(ColorHex.parseArgb("#GG"))
    }

    @Test
    fun resolveFromBrandingUsesPrimaryAndDefaults() {
        val palette = BrandPalette.from(
            BrandingConfig(
                primaryColor = "#3AA0E8",
                textColorPrimary = "#111111",
            ),
        )
        assertEquals(0xFF3AA0E8L, palette.primary)
        assertEquals(0x1A3AA0E8L, palette.primaryMuted)
        assertEquals(0xFF111111L, palette.textPrimary)
        assertEquals(BrandPalette.DEFAULT_DISABLED, palette.disabled)
    }

    @Test
    fun cssHexRgb() {
        assertEquals("#00a0e8", BrandPalette.cssHexRgb(0xFF00A0E8L))
    }
}
