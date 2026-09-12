package com.luopingtech.ebike.ops.core.config

import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json

/**
 * Tenant runtime config. Mirrors `config/{tenant}_{mode}.json`.
 * Secrets stay empty in the public demo file.
 */
@Serializable
data class TenantConfig(
    val name: String = "",
    val alias: String = "",
    val description: String = "",
    val tenantId: String = "",
    val app: AppConfig = AppConfig(),
    val api: ApiConfig = ApiConfig(),
    val auth: AuthConfig = AuthConfig(),
    val map: MapConfig = MapConfig(),
    val push: PushConfig = PushConfig(),
    val ble: BleConfig = BleConfig(),
    val features: FeatureFlags = FeatureFlags(),
    val branding: BrandingConfig = BrandingConfig(),
    val documents: DocumentsConfig = DocumentsConfig(),
    /**
     * Optional remote H5 dashboards opened in-app WebView.
     * Blank URL ⇒ hide that workbench tile. Paths are tenant-configurable.
     */
    val h5: H5ScreensConfig = H5ScreensConfig(),
) {
    companion object {
        fun demo(): TenantConfig = TenantConfig(
            name = "Demo Ops",
            alias = "demo",
            description = "public demo tenant, no credentials included",
            app = AppConfig(
                androidApplicationId = "com.luopingtech.ebike.ops.demo",
                iosBundleId = "com.luopingtech.ebike.ops.demo",
                displayName = "Demo Ops",
            ),
            features = FeatureFlags(bleTransport = "simulator", mapProvider = "simulator"),
            branding = BrandingConfig(primaryColor = "#3AA0E8"),
            h5 = H5ScreensConfig(
                operationUrl = H5ScreensConfig.DEFAULT_OPERATION_URL,
                revenueUrl = H5ScreensConfig.DEFAULT_REVENUE_URL,
                appendAuthQuery = true,
            ),
        )
    }
}

@Serializable
data class AppConfig(
    val androidApplicationId: String = "",
    val iosBundleId: String = "",
    val displayName: String = "",
)

@Serializable
data class ApiConfig(
    val baseUrl: String = "",
    val logUploadUrl: String = "",
)

@Serializable
data class AuthConfig(
    val businessSecret: String = "",
    val signSecret: String = "",
)

@Serializable
data class MapConfig(
    val provider: String = "none",
    val tencentKey: String = "",
    val googleKey: String = "",
)

@Serializable
data class PushConfig(
    val provider: String = "none",
    val appKey: String = "",
)

@Serializable
data class BleConfig(
    val transport: String = "simulator",
    val defaultToken: List<String> = emptyList(),
)

@Serializable
data class FeatureFlags(
    val bleTransport: String = "simulator",
    val mapProvider: String = "none",
    val push: Boolean = false,
    val backgroundLocation: Boolean = false,
    val overseas: Boolean = false,
)

@Serializable
data class BrandingConfig(
    /** Legacy `themeColor`. */
    val primaryColor: String = "",
    /** Legacy `themeColor1A` (primary @ ~10% alpha). */
    val primaryMutedColor: String = "",
    /** Text / icons drawn on primary-filled surfaces (usually white). */
    val onPrimaryColor: String = "",
    /** Legacy `lightTxtColor` — text sitting on light/theme-tinted chrome. */
    val lightTextColor: String = "",
    /** Legacy `navigationBarTxtColor`. */
    val navigationBarTextColor: String = "",
    /** Legacy `navigationBackgroundColor`. */
    val navigationBarBackgroundColor: String = "",
    /** Legacy `disabledColor`. */
    val disabledColor: String = "",
    /** Legacy `textColorBlack3`. */
    val textColorPrimary: String = "",
    /** Legacy `textColorBlack6`. */
    val textColorSecondary: String = "",
    /** Legacy `textColorBlack9`. */
    val textColorTertiary: String = "",
    val pageBackgroundColor: String = "",
    val dividerColor: String = "",
    val chipBackgroundColor: String = "",
    val tabUnselectedColor: String = "",
    val logo: String = "",
)

@Serializable
data class DocumentsConfig(
    val privacyPolicy: String = "",
    val userAgreement: String = "",
)

/**
 * Legacy Merchant-H5 big screens (运营 / 营收).
 * Deployed on CDN; App only loads remote URLs (not bundled).
 */
@Serializable
data class H5ScreensConfig(
    /**
     * SPA 入口地址（不含 `#` 路由），例如 `https://host/mop-saas/index.html`。
     * 除两块大屏外的 H5 屏都由它拼 [H5ScreenKind.route] 得到。
     * 留空时从 [operationUrl] / [revenueUrl] 里取 `#` 之前的部分反推。
     */
    val baseUrl: String = "",
    /** Full URL, e.g. `https://host/mop-saas/index.html#/newOperationScreen`. Empty = hide. */
    val operationUrl: String = "",
    /** Full URL for revenue screen. Empty = hide. */
    val revenueUrl: String = "",
    /**
     * When true, append apiHost / tokens / tenant secrets like iOS [BigScreenViewController]
     * so the SPA can call APIs without a separate login.
     */
    val appendAuthQuery: Boolean = true,
) {
    companion object {
        const val DEFAULT_OPERATION_URL: String =
            "https://ebike.luopingtech.com/mop-saas/index.html#/newOperationScreen"
        const val DEFAULT_REVENUE_URL: String =
            "https://ebike.luopingtech.com/mop-saas/index.html#/newrevenueHome"
    }
}

object TenantConfigLoader {
    private val json = Json {
        ignoreUnknownKeys = true
        isLenient = true
        encodeDefaults = true
    }

    fun fromJson(raw: String): TenantConfig = json.decodeFromString(TenantConfig.serializer(), raw)

    fun demo(): TenantConfig = TenantConfig.demo()
}
