package com.luopingtech.ebike.rider.ui.theme

import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.ReadOnlyComposable
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.graphics.Color
import com.luopingtech.ebike.rider.core.config.BrandPalette
import com.luopingtech.ebike.rider.core.config.BrandingConfig

/**
 * Tenant-driven UI colors. Prefer [RiderTheme.colors] inside Composables.
 * Values come from [BrandingConfig] via [BrandPalette].
 */
data class RiderColors(
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
        fun from(palette: BrandPalette): RiderColors = RiderColors(
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

        fun from(branding: BrandingConfig): RiderColors = from(BrandPalette.from(branding))

        val Default: RiderColors = from(BrandPalette.defaults())
    }
}

val LocalRiderColors = staticCompositionLocalOf { RiderColors.Default }

object RiderTheme {
    val colors: RiderColors
        @Composable
        @ReadOnlyComposable
        get() = LocalRiderColors.current
}

@Composable
fun RiderTheme(
    branding: BrandingConfig = BrandingConfig(),
    content: @Composable () -> Unit,
) {
    val colors = RiderColors.from(branding)
    CompositionLocalProvider(LocalRiderColors provides colors) {
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
            ),
            content = content,
        )
    }
}
