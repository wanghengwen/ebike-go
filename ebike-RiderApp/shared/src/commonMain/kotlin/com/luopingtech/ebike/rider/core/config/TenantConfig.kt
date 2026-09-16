package com.luopingtech.ebike.rider.core.config

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
    /** UniApp H5 长尾页基址。空则原生只展示占位，不加载 WebView。 */
    val h5: H5ScreensConfig = H5ScreensConfig(),
) {
    companion object {
        fun demo(): TenantConfig = TenantConfig(
            name = "Demo Rider",
            alias = "demo",
            description = "public demo tenant, no credentials included",
            app = AppConfig(
                androidApplicationId = "com.luopingtech.ebike.rider.demo",
                iosBundleId = "com.luopingtech.ebike.rider.demo",
                displayName = "Demo Rider",
            ),
            features = FeatureFlags(bleTransport = "simulator", mapProvider = "simulator"),
            branding = BrandingConfig(primaryColor = "#3AA0E8"),
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
    /**
     * UniApp 租户配置里的 `onlyBluetooth`：车队没有联网车机，开锁 / 还车只能走蓝牙。
     * 置位后不再尝试远程指令，省掉必然超时的那一轮。
     */
    val onlyBluetooth: Boolean = false,
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
    /** UniApp `iconCfg.default_avatar` — 首页左上角个人中心。 */
    val defaultAvatar: String = "",
    /** UniApp `mapCfg.iconScan` — 「立即用车」按钮图标。 */
    val iconScan: String = "",
    /** 首页快捷入口兜底图标（接口未下发时）。 */
    val homeIconTrips: String = "",
    val homeIconWallet: String = "",
    val homeIconService: String = "",
    val homeIconBilling: String = "",
)

@Serializable
data class DocumentsConfig(
    val privacyPolicy: String = "",
    val userAgreement: String = "",
)

/**
 * C 端 H5 容器配置。只允许填公开基址，**不要**把 signSecret / businessSecret 放进这里。
 *
 * [baseUrl] 例：`https://cdn.example.com/h5/index.html` 或本地调试 `http://192.168.1.8:5173/`。
 * 桥会用它的 origin 做白名单；文档基址（`#` 前）变化才 reload WebView。
 */
@Serializable
data class H5ScreensConfig(
    val baseUrl: String = "",
)

object TenantConfigLoader {
    private val json = Json {
        ignoreUnknownKeys = true
        isLenient = true
        encodeDefaults = true
    }

    fun fromJson(raw: String): TenantConfig = json.decodeFromString(TenantConfig.serializer(), raw)

    fun demo(): TenantConfig = TenantConfig.demo()
}
