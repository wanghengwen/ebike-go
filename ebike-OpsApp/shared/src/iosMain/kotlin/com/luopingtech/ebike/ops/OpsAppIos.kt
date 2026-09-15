package com.luopingtech.ebike.ops

import com.luopingtech.ebike.ops.core.config.MapConfig
import com.luopingtech.ebike.ops.core.config.TenantConfig
import com.luopingtech.ebike.ops.core.config.TenantConfigLoader
import com.luopingtech.ebike.ops.platform.BindableMediaUploader
import com.luopingtech.ebike.ops.platform.CoreLocationTracker
import com.luopingtech.ebike.ops.platform.IosCodeScanner
import com.luopingtech.ebike.ops.platform.IosMediaUploader
import com.luopingtech.ebike.ops.platform.IosPhotoCapture
import com.luopingtech.ebike.ops.platform.IosReverseGeocoder
import com.luopingtech.ebike.ops.platform.SecureStore
import com.luopingtech.ebike.ops.platform.SimulatorLocationTracker
import com.luopingtech.ebike.ops.platform.createSecureStore
import kotlinx.cinterop.BetaInteropApi
import kotlinx.cinterop.ExperimentalForeignApi
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import platform.Foundation.NSBundle
import platform.Foundation.NSLocale
import platform.Foundation.NSString
import platform.Foundation.NSUTF8StringEncoding
import platform.Foundation.preferredLanguages
import platform.Foundation.stringWithContentsOfFile

/**
 * iOS 宿主用的装配入口，对应 Android 的 `OpsApplication.onCreate`。
 *
 * Swift 那边优先调 [createBundleIosOpsApp]（读 bundle 里的 `tenant.json`，
 * 跟 Android 读 `assets/tenant.json` 是同一份配置）；没有资源文件时再退
 * [createDemoIosOpsApp]。地图 Key 的合并规则也跟 Android 的 `mergeMapKey` 对齐。
 */
fun createIosOpsApp(config: TenantConfig): OpsApp {
    val merged = mergeMapKey(config, infoPlistTencentKey())
    val secureStore = createSecureStore()
    val demo = merged.api.baseUrl.isBlank()
    val app = OpsApp.create(
        config = merged,
        secureStore = secureStore,
        codeScanner = IosCodeScanner(),
        photoCapture = IosPhotoCapture(),
        // demo（baseUrl 为空）用模拟轨迹：模拟器默认没有 GPS，
        // 真机联调时才切 CoreLocation，跟 OpsApp.create 里的降级规则一致。
        locationTracker = if (demo) SimulatorLocationTracker() else CoreLocationTracker(),
        reverseGeocoder = if (demo) null else IosReverseGeocoder(),
        // 首次启动跟随系统语言；用户存过的选择仍然优先（OpsI18n 里判断）。
        systemLanguage = systemLanguageTag(),
    )
    // 上报照片要走签名上传接口，非 demo 模式下共享层给的是待绑定的桥。
    (app.mediaUploader as? BindableMediaUploader)?.let { bridge ->
        app.fileUploadApi?.let { api -> bridge.bind(IosMediaUploader(api)) }
    }
    if (merged.tenantId.isNotBlank()) {
        secureStore.putString(SecureStore.KEY_TENANT_ID, merged.tenantId)
    }
    // Android 在 Application 里 bind(appScope)。少了这一步 feature 的 state 流不会启动，
    // 界面会永远停在「加载服务区…」——iOS 侧同样需要一个进程级 scope。
    app.bind(CoroutineScope(SupervisorJob() + Dispatchers.Main.immediate))
    app.logger.i(
        "OpsApp",
        "started shared=${OpsApp.LIBRARY_VERSION} demo=${app.isDemoMode} " +
            "tenant=${merged.alias} map=${app.mapCapability.kind}/ready=${app.mapCapability.isReady} " +
            "tencentKey=${if (merged.map.tencentKey.isNotBlank()) "set" else "empty"}",
    )
    return app
}

fun createDemoIosOpsApp(): OpsApp = createIosOpsApp(TenantConfig.demo())

/**
 * 从 app bundle 里的 `tenant.json` 装配，对应 Android 读 `assets/tenant.json`。
 * 文件缺失或解析失败都退回 demo 配置：宿主少放一个资源不该让 App 起不来。
 */
fun createBundleIosOpsApp(): OpsApp = createIosOpsApp(bundleTenantConfig())

/**
 * 对齐 Android `OpsApplication.mergeMapKey`：
 * 租户文件里的 Key 优先，否则用 Info.plist 的 `OpsTencentMapKey`
 * （由 `OpsTenant.xcconfig` / `sync_tenant.sh` 从 `local.properties` 注入）。
 * 有 Key 且不是海外/谷歌时，把 provider 切到 tencent。
 */
internal fun mergeMapKey(base: TenantConfig, buildKey: String): TenantConfig {
    val tenantKey = base.map.tencentKey.trim()
    val key = tenantKey.ifBlank { buildKey.trim() }
    if (key.isBlank()) return base

    val providerRaw = base.features.mapProvider.ifBlank { base.map.provider }.trim()
    val useTencent = !providerRaw.equals("google", ignoreCase = true) &&
        !base.features.overseas

    if (!useTencent) {
        return base.copy(map = base.map.copy(tencentKey = key))
    }

    return base.copy(
        map = MapConfig(
            provider = "tencent",
            tencentKey = key,
            googleKey = base.map.googleKey,
        ),
        features = base.features.copy(mapProvider = "tencent"),
    )
}

@OptIn(BetaInteropApi::class, ExperimentalForeignApi::class)
private fun bundleTenantConfig(): TenantConfig {
    val path = NSBundle.mainBundle.pathForResource("tenant", ofType = "json")
        ?: return TenantConfig.demo()
    val raw = NSString.stringWithContentsOfFile(path, encoding = NSUTF8StringEncoding, error = null)
        ?: return TenantConfig.demo()
    return runCatching { TenantConfigLoader.fromJson(raw) }.getOrElse { TenantConfig.demo() }
}

private fun infoPlistTencentKey(): String {
    val value = NSBundle.mainBundle.objectForInfoDictionaryKey("OpsTencentMapKey") as? String
    return value.orEmpty().trim()
}

private fun systemLanguageTag(): String? =
    (NSLocale.preferredLanguages.firstOrNull() as? String)?.takeIf { it.isNotBlank() }
