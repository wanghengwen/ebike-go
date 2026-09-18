package com.luopingtech.ebike.ops

import android.app.Application
import com.luopingtech.ebike.ops.core.config.MapConfig
import com.luopingtech.ebike.ops.core.config.TenantConfig
import com.luopingtech.ebike.ops.core.config.TenantConfigLoader
import com.luopingtech.ebike.ops.core.logging.LogLevel
import com.luopingtech.ebike.ops.core.logging.OpsLogger
import com.luopingtech.ebike.ops.platform.AndroidLocationTracker
import com.luopingtech.ebike.ops.platform.AndroidMediaUploader
import com.luopingtech.ebike.ops.platform.AndroidReverseGeocoder
import com.luopingtech.ebike.ops.platform.AndroidSecureStore
import com.luopingtech.ebike.ops.platform.BindableCodeScanner
import com.luopingtech.ebike.ops.platform.BindableMediaUploader
import com.luopingtech.ebike.ops.platform.BindablePhotoCapture
import com.luopingtech.ebike.ops.platform.SecureStore
import com.luopingtech.ebike.ops.platform.SimulatorLocationTracker
import com.luopingtech.ebike.ops.platform.DemoReverseGeocoder
import com.luopingtech.ebike.ops.platform.bindOpsDiskCacheContext
import com.luopingtech.ebike.ops.platform.createDeviceInfo
import com.tencent.tencentmap.mapsdk.maps.TencentMapInitializer
import java.util.Locale
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob

class OpsApplication : Application() {
    private val appJob = SupervisorJob()
    val appScope: CoroutineScope = CoroutineScope(appJob + Dispatchers.Main.immediate)

    val codeScannerBridge: BindableCodeScanner = BindableCodeScanner()
    val photoCaptureBridge: BindablePhotoCapture = BindablePhotoCapture()

    lateinit var opsApp: OpsApp
        private set

    val isOpsAppReady: Boolean
        get() = ::opsApp.isInitialized

    override fun onCreate() {
        super.onCreate()
        bindOpsDiskCacheContext(this)
        // Required by Tencent Map SDK ≥4.5.6 before any MapView is created.
        TencentMapInitializer.setAgreePrivacy(true)

        val logger = object : OpsLogger {
            override fun log(
                level: LogLevel,
                tag: String,
                message: String,
                throwable: Throwable?,
            ) {
                when (level) {
                    LogLevel.DEBUG -> android.util.Log.d(tag, message, throwable)
                    LogLevel.INFO -> android.util.Log.i(tag, message, throwable)
                    LogLevel.WARN -> android.util.Log.w(tag, message, throwable)
                    LogLevel.ERROR -> android.util.Log.e(tag, message, throwable)
                }
            }
        }
        val secureStore = AndroidSecureStore(this)
        val config = mergeMapKey(loadTenantConfig())
        val versionName = runCatching {
            @Suppress("DEPRECATION")
            packageManager.getPackageInfo(packageName, 0).versionName
        }.getOrNull().orEmpty().ifBlank { "1.0.0" }
        val locationTracker = if (config.api.baseUrl.isBlank()) {
            SimulatorLocationTracker(intervalMs = 5_000L)
        } else {
            AndroidLocationTracker(this)
        }
        val reverseGeocoder = if (config.api.baseUrl.isBlank()) {
            DemoReverseGeocoder()
        } else {
            AndroidReverseGeocoder(this)
        }
        opsApp = OpsApp.create(
            config = config,
            logger = logger,
            secureStore = secureStore,
            deviceInfo = createDeviceInfo(appVersion = versionName),
            codeScanner = codeScannerBridge,
            photoCapture = photoCaptureBridge,
            locationTracker = locationTracker,
            reverseGeocoder = reverseGeocoder,
            // Follow device locale on first launch; a stored user choice still wins.
            systemLanguage = Locale.getDefault().toLanguageTag(),
        )
        bindMediaUploader()
        if (config.tenantId.isNotBlank()) {
            secureStore.putString(SecureStore.KEY_TENANT_ID, config.tenantId)
        }
        opsApp.bind(appScope)
        logger.i(
            "OpsApp",
            "started shared=${OpsApp.LIBRARY_VERSION} demo=${opsApp.isDemoMode} " +
                "tenant=${config.alias} map=${opsApp.mapCapability.kind}/ready=${opsApp.mapCapability.isReady}",
        )
    }

    private fun bindMediaUploader() {
        val api = opsApp.fileUploadApi ?: return
        val bridge = opsApp.mediaUploader as? BindableMediaUploader ?: return
        bridge.bind(AndroidMediaUploader(this, api))
    }

    private fun loadTenantConfig(): TenantConfig {
        val fromAssets = runCatching {
            assets.open(ASSET_TENANT).bufferedReader().use { it.readText() }
        }.getOrNull()
        if (!fromAssets.isNullOrBlank()) {
            return runCatching { TenantConfigLoader.fromJson(fromAssets) }
                .getOrElse { TenantConfig.demo() }
        }
        return TenantConfig.demo()
    }

    /**
     * Prefer tenant.json map key; else inject BuildConfig key from local.properties.
     * When a Tencent key exists and provider is unset/none, switch to tencent.
     */
    private fun mergeMapKey(base: TenantConfig): TenantConfig {
        val buildKey = BuildConfig.TENCENT_MAP_KEY.trim()
        val tenantKey = base.map.tencentKey.trim()
        val key = tenantKey.ifBlank { buildKey }
        if (key.isBlank()) return base

        val providerRaw = base.features.mapProvider.ifBlank { base.map.provider }.trim()
        // Key present ⇒ prefer Tencent unless explicitly google/overseas.
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

    override fun onTerminate() {
        appJob.cancel()
        if (::opsApp.isInitialized) {
            opsApp.close()
        }
        super.onTerminate()
    }

    companion object {
        const val ASSET_TENANT = "tenant.json"
    }
}
