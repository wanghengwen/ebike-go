package com.luopingtech.ebike.rider

import com.luopingtech.ebike.rider.core.config.MapConfig
import com.luopingtech.ebike.rider.core.config.TenantConfig
import com.luopingtech.ebike.rider.core.config.TenantConfigLoader
import com.luopingtech.ebike.rider.platform.BindableMediaUploader
import com.luopingtech.ebike.rider.platform.CoreLocationTracker
import com.luopingtech.ebike.rider.platform.IosCodeScanner
import com.luopingtech.ebike.rider.platform.IosMediaUploader
import com.luopingtech.ebike.rider.platform.IosPhotoCapture
import com.luopingtech.ebike.rider.platform.IosReverseGeocoder
import com.luopingtech.ebike.rider.platform.NativeBleTransport
import com.luopingtech.ebike.rider.platform.SecureStore
import com.luopingtech.ebike.rider.platform.SimulatorLocationTracker
import com.luopingtech.ebike.rider.platform.createSecureStore
import kotlinx.cinterop.BetaInteropApi
import kotlinx.cinterop.ExperimentalForeignApi
import platform.Foundation.NSBundle
import platform.Foundation.NSLocale
import platform.Foundation.NSString
import platform.Foundation.NSUTF8StringEncoding
import platform.Foundation.preferredLanguages
import platform.Foundation.stringWithContentsOfFile

fun createIosRiderApp(config: TenantConfig): RiderApp {
    val merged = mergeMapKey(config, infoPlistTencentKey())
    val secureStore = createSecureStore()
    val demo = merged.api.baseUrl.isBlank()
    val app = RiderApp.create(
        config = merged,
        secureStore = secureStore,
        logger = com.luopingtech.ebike.rider.core.logging.StdoutLogger,
        appVersion = RiderApp.LIBRARY_VERSION,
        deviceInfo = com.luopingtech.ebike.rider.platform.createDeviceInfo(RiderApp.LIBRARY_VERSION),
        systemLanguage = systemLanguageTag(),
        quickLogin = com.luopingtech.ebike.rider.platform.UnsupportedQuickLogin,
        codeScanner = IosCodeScanner(cancelTitle = "取消"),
        photoCapture = IosPhotoCapture(),
        locationTracker = if (demo) SimulatorLocationTracker() else CoreLocationTracker(),
        reverseGeocoder = if (demo) null else IosReverseGeocoder(),
        mediaUploader = null,
        bleTransport = null,
        nativeBle = NativeBleTransport(),
    )
    (app.mediaUploader as? BindableMediaUploader)?.let { bridge ->
        app.fileUploadApi?.let { api -> bridge.bind(IosMediaUploader(api)) }
    }
    if (merged.tenantId.isNotBlank()) {
        secureStore.putString(SecureStore.KEY_TENANT_ID, merged.tenantId)
    }
    app.logger.i(
        "RiderApp",
        "started shared=${RiderApp.LIBRARY_VERSION} demo=${app.isDemoMode} " +
            "tenant=${merged.alias} map=${app.mapCapability.kind}/ready=${app.mapCapability.isReady}",
    )
    return app
}

fun createDemoIosRiderApp(): RiderApp = createIosRiderApp(TenantConfig.demo())

fun createBundleIosRiderApp(): RiderApp = createIosRiderApp(bundleTenantConfig())

@OptIn(BetaInteropApi::class, ExperimentalForeignApi::class)
private fun bundleTenantConfig(): TenantConfig {
    val path = NSBundle.mainBundle.pathForResource("tenant", ofType = "json")
        ?: return TenantConfig.demo()
    val raw = NSString.stringWithContentsOfFile(path, encoding = NSUTF8StringEncoding, error = null)
        ?: return TenantConfig.demo()
    return runCatching { TenantConfigLoader.fromJson(raw) }.getOrElse { TenantConfig.demo() }
}

private fun mergeMapKey(base: TenantConfig, plistKey: String): TenantConfig {
    val key = base.map.tencentKey.trim().ifBlank { plistKey.trim() }
    if (key.isBlank()) return base
    val providerRaw = base.features.mapProvider.ifBlank { base.map.provider }.trim()
    val useTencent = !providerRaw.equals("google", ignoreCase = true) && !base.features.overseas
    if (!useTencent) return base.copy(map = base.map.copy(tencentKey = key))
    return base.copy(
        map = MapConfig(provider = "tencent", tencentKey = key, googleKey = base.map.googleKey),
        features = base.features.copy(mapProvider = "tencent"),
    )
}

private fun infoPlistTencentKey(): String =
    (NSBundle.mainBundle.objectForInfoDictionaryKey("RiderTencentMapKey") as? String).orEmpty()

private fun bundleShortVersion(): String =
    (NSBundle.mainBundle.objectForInfoDictionaryKey("CFBundleShortVersionString") as? String)
        ?.takeIf { it.isNotBlank() }
        ?: RiderApp.LIBRARY_VERSION

private fun systemLanguageTag(): String? =
    (NSLocale.preferredLanguages.firstOrNull() as? String)?.takeIf { it.isNotBlank() }
