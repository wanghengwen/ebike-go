package com.luopingtech.ebike.ops.ui.theme

import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.ReadOnlyComposable
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.graphics.Color
import com.luopingtech.ebike.ops.core.config.BrandPalette
import com.luopingtech.ebike.ops.core.config.BrandingConfig

/**
 * Tenant-driven UI colors. Prefer [OpsTheme.colors] inside Composables.
 * Values come from [BrandingConfig] / legacy ManagerApp `themeColor*` resValues.
 */
data class OpsColors(
    val primary: Color,
    val primaryMuted: Color,
    val onPrimary: Color,
    val onPrimarySurface: Color,
    val navigationBarText: Color,
    val navigationBarBackground: Color,
    val disabled: Color,
    val textPrimary: Color,
    val textSecondary: Color,
    val textTertiary: Color,
    val pageBackground: Color,
    val divider: Color,
    val chipBackground: Color,
    val tabUnselected: Color,
) {
    companion object {
        fun from(palette: BrandPalette): OpsColors = OpsColors(
            primary = Color(palette.primary),
            primaryMuted = Color(palette.primaryMuted),
            onPrimary = Color(palette.onPrimary),
            onPrimarySurface = Color(palette.onPrimarySurface),
            navigationBarText = Color(palette.navigationBarText),
            navigationBarBackground = Color(palette.navigationBarBackground),
            disabled = Color(palette.disabled),
            textPrimary = Color(palette.textPrimary),
            textSecondary = Color(palette.textSecondary),
            textTertiary = Color(palette.textTertiary),
            pageBackground = Color(palette.pageBackground),
            divider = Color(palette.divider),
            chipBackground = Color(palette.chipBackground),
            tabUnselected = Color(palette.tabUnselected),
        )

        fun from(branding: BrandingConfig): OpsColors = from(BrandPalette.from(branding))

        val Default: OpsColors = from(BrandPalette.defaults())
    }
}

val LocalOpsColors = staticCompositionLocalOf { OpsColors.Default }

object OpsTheme {
    val colors: OpsColors
        @Composable
        @ReadOnlyComposable
        get() = LocalOpsColors.current
}

/** Fixed semantic colors (home statistics / map chrome) — not tenant branding. */
val OpsToolIcon = Color(0xFF48506C)
val OpsStatGreen = Color(0xFF00C24A)
val OpsStatBlue = Color(0xFF14B6FF)
val OpsStatRed = Color(0xFFFF2900)
val OpsStatOrange = Color(0xFFFF8122)
val OpsStatItemBg = Color(0x1A828DB3)
val OpsStatSelected = Color(0xFF1887F8)
val OpsFilterHandle = Color(0xFF5A6278)

@Composable
fun OpsTheme(
    branding: BrandingConfig = BrandingConfig(),
    content: @Composable () -> Unit,
) {
    val colors = OpsColors.from(branding)
    CompositionLocalProvider(LocalOpsColors provides colors) {
        MaterialTheme(
            colorScheme = lightColorScheme(
                primary = colors.primary,
                onPrimary = colors.onPrimary,
                primaryContainer = colors.primaryMuted,
                background = colors.pageBackground,
                surface = colors.pageBackground,
                onBackground = colors.textPrimary,
                onSurface = colors.textPrimary,
                onSurfaceVariant = colors.textTertiary,
                outline = colors.divider,
                secondary = colors.textSecondary,
                onSecondary = colors.onPrimary,
                error = colors.statRedCompat(),
            ),
            content = content,
        )
    }
}

private fun OpsColors.statRedCompat(): Color = OpsStatRed
