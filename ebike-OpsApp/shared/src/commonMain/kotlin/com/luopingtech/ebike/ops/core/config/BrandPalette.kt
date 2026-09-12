package com.luopingtech.ebike.ops.core.config

/**
 * Resolved branding palette (ARGB). Built from [BrandingConfig] with legacy ManagerApp defaults.
 *
 * Field mapping from `renren.gradle` / flavor `resValue`:
 * - themeColor → [primary]
 * - themeColor1A → [primaryMuted]
 * - lightTxtColor → [onPrimarySurface]
 * - navigationBarTxtColor → [navigationBarText]
 * - navigationBackgroundColor → [navigationBarBackground]
 * - disabledColor → [disabled]
 * - textColorBlack3/6/9 → [textPrimary]/[textSecondary]/[textTertiary]
 */
data class BrandPalette(
    val primary: Long,
    val primaryMuted: Long,
    val onPrimary: Long,
    val onPrimarySurface: Long,
    val navigationBarText: Long,
    val navigationBarBackground: Long,
    val disabled: Long,
    val textPrimary: Long,
    val textSecondary: Long,
    val textTertiary: Long,
    val pageBackground: Long,
    val divider: Long,
    val chipBackground: Long,
    val tabUnselected: Long,
) {
    companion object {
        /** Legacy renren `themeColor`. */
        const val DEFAULT_PRIMARY: Long = 0xFF00A0E8L
        const val DEFAULT_PRIMARY_MUTED: Long = 0x1A00A0E8L
        const val DEFAULT_ON_PRIMARY: Long = 0xFFFFFFFFL
        const val DEFAULT_ON_PRIMARY_SURFACE: Long = 0xFF282828L
        const val DEFAULT_NAV_TEXT: Long = 0xFF282828L
        const val DEFAULT_NAV_BG: Long = 0xFF282828L
        const val DEFAULT_DISABLED: Long = 0xFFCCCCCCL
        const val DEFAULT_TEXT_PRIMARY: Long = 0xFF333333L
        const val DEFAULT_TEXT_SECONDARY: Long = 0xFF666666L
        const val DEFAULT_TEXT_TERTIARY: Long = 0xFF999999L
        const val DEFAULT_PAGE_BG: Long = 0xFFFFFFFFL
        const val DEFAULT_DIVIDER: Long = 0xFFD7D7D7L
        const val DEFAULT_CHIP_BG: Long = 0xFFF4F4F4L
        const val DEFAULT_TAB_UNSELECTED: Long = 0xFF242936L

        fun defaults(): BrandPalette = BrandPalette(
            primary = DEFAULT_PRIMARY,
            primaryMuted = DEFAULT_PRIMARY_MUTED,
            onPrimary = DEFAULT_ON_PRIMARY,
            onPrimarySurface = DEFAULT_ON_PRIMARY_SURFACE,
            navigationBarText = DEFAULT_NAV_TEXT,
            navigationBarBackground = DEFAULT_NAV_BG,
            disabled = DEFAULT_DISABLED,
            textPrimary = DEFAULT_TEXT_PRIMARY,
            textSecondary = DEFAULT_TEXT_SECONDARY,
            textTertiary = DEFAULT_TEXT_TERTIARY,
            pageBackground = DEFAULT_PAGE_BG,
            divider = DEFAULT_DIVIDER,
            chipBackground = DEFAULT_CHIP_BG,
            tabUnselected = DEFAULT_TAB_UNSELECTED,
        )

        fun from(branding: BrandingConfig): BrandPalette {
            val primary = ColorHex.parseArgbOr(branding.primaryColor, DEFAULT_PRIMARY)
            val mutedFallback = (primary and 0x00FFFFFFL) or 0x1A000000L
            return BrandPalette(
                primary = primary,
                primaryMuted = ColorHex.parseArgbOr(branding.primaryMutedColor, mutedFallback),
                onPrimary = ColorHex.parseArgbOr(branding.onPrimaryColor, DEFAULT_ON_PRIMARY),
                onPrimarySurface = ColorHex.parseArgbOr(
                    branding.lightTextColor,
                    DEFAULT_ON_PRIMARY_SURFACE,
                ),
                navigationBarText = ColorHex.parseArgbOr(
                    branding.navigationBarTextColor,
                    DEFAULT_NAV_TEXT,
                ),
                navigationBarBackground = ColorHex.parseArgbOr(
                    branding.navigationBarBackgroundColor,
                    DEFAULT_NAV_BG,
                ),
                disabled = ColorHex.parseArgbOr(branding.disabledColor, DEFAULT_DISABLED),
                textPrimary = ColorHex.parseArgbOr(
                    branding.textColorPrimary,
                    DEFAULT_TEXT_PRIMARY,
                ),
                textSecondary = ColorHex.parseArgbOr(
                    branding.textColorSecondary,
                    DEFAULT_TEXT_SECONDARY,
                ),
                textTertiary = ColorHex.parseArgbOr(
                    branding.textColorTertiary,
                    DEFAULT_TEXT_TERTIARY,
                ),
                pageBackground = ColorHex.parseArgbOr(
                    branding.pageBackgroundColor,
                    DEFAULT_PAGE_BG,
                ),
                divider = ColorHex.parseArgbOr(branding.dividerColor, DEFAULT_DIVIDER),
                chipBackground = ColorHex.parseArgbOr(
                    branding.chipBackgroundColor,
                    DEFAULT_CHIP_BG,
                ),
                tabUnselected = ColorHex.parseArgbOr(
                    branding.tabUnselectedColor,
                    DEFAULT_TAB_UNSELECTED,
                ),
            )
        }
        fun cssHexRgb(argb: Long): String {
            val rgb = argb and 0x00FFFFFFL
            return "#${rgb.toString(16).padStart(6, '0')}"
        }
    }
}
