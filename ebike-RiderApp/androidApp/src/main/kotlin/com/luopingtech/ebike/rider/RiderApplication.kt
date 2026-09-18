package com.luopingtech.ebike.rider

import android.app.Application
import com.luopingtech.ebike.rider.core.config.MapConfig
import com.luopingtech.ebike.rider.core.config.TenantConfig
import com.luopingtech.ebike.rider.core.config.TenantConfigLoader
import com.luopingtech.ebike.rider.core.logging.LogLevel
import com.luopingtech.ebike.rider.core.logging.RiderLogger
import com.luopingtech.ebike.rider.platform.AndroidLocationTracker
import com.luopingtech.ebike.rider.platform.NativeBleTransport
import com.luopingtech.ebike.rider.platform.AndroidMediaUploader
import com.luopingtech.ebike.rider.platform.AndroidReverseGeocoder
import com.luopingtech.ebike.rider.platform.AndroidSecureStore
import com.luopingtech.ebike.rider.platform.BindableCodeScanner
import com.luopingtech.ebike.rider.platform.BindableMediaUploader
import com.luopingtech.ebike.rider.platform.BindablePhotoCapture
import com.luopingtech.ebike.rider.pay.AndroidWeChatPay
import com.luopingtech.ebike.rider.platform.DemoReverseGeocoder
import com.luopingtech.ebike.rider.platform.SecureStore
import com.luopingtech.ebike.rider.platform.SimulatorLocationTracker
import com.luopingtech.ebike.rider.platform.UnsupportedQuickLogin
import com.luopingtech.ebike.rider.platform.createDeviceInfo
import com.tencent.tencentmap.mapsdk.maps.TencentMapInitializer
import java.util.Locale
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob

class RiderApplication : Application() {
    private val appJob = SupervisorJob()
    val appScope: CoroutineScope = CoroutineScope(appJob + Dispatchers.Main.immediate)

    val codeScannerBridge: BindableCodeScanner = BindableCodeScanner()
    val photoCaptureBridge: BindablePhotoCapture = BindablePhotoCapture()

    lateinit var riderApp: RiderApp
        private set

    override fun onCreate() {
        super.onCreate()
        TencentMapInitializer.setAgreePrivacy(true)

        val logger = object : RiderLogger {
            override fun log(level: LogLevel, tag: String, message: String, throwable: Throwable?) {
                when (level) {
                    LogLevel.DEBUG -> android.util.Log.d(tag, message, throwable)
                    LogLevel.INFO -> android.util.Log.i(tag, message, throwable)
                    LogLevel.WARN -> android.util.Log.w(tag, message, throwable)
                    LogLevel.ERROR -> android.util.Log.e(tag, message, throwable)
                }
            }
        }
        val secureStore = AndroidSecureStore(this)
        val config = mergePay(mergeMapKey(loadTenantConfig()))
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
        riderApp = RiderApp.create(
            config = config,
            secureStore = secureStore,
            logger = logger,
            appVersion = versionName(),
            deviceInfo = createDeviceInfo(versionName()),
            systemLanguage = Locale.getDefault().toLanguageTag(),
            quickLogin = UnsupportedQuickLogin,
            codeScanner = codeScannerBridge,
            photoCapture = photoCaptureBridge,
            locationTracker = locationTracker,
            reverseGeocoder = reverseGeocoder,
            mediaUploader = null,
            bleTransport = null,
            nativeBle = NativeBleTransport(this),
        )
        bindMediaUploader()
        bindWeChatPay()
        if (config.tenantId.isNotBlank()) {
            secureStore.putString(SecureStore.KEY_TENANT_ID, config.tenantId)
        }
        logger.i(
            TAG,
            "started shared=${RiderApp.LIBRARY_VERSION} demo=${riderApp.isDemoMode} " +
                "tenant=${config.alias} map=${riderApp.mapCapability.kind}/ready=${riderApp.mapCapability.isReady}",
        )
    }

    private fun bindMediaUploader() {
        val api = riderApp.fileUploadApi ?: return
        val bridge = riderApp.mediaUploader as? BindableMediaUploader ?: return
        bridge.bind(AndroidMediaUploader(this, api))
    }

    private fun bindWeChatPay() {
        riderApp.wechatPay.bind(
            AndroidWeChatPay(this) {
                riderApp.config.pay.wechatAppId.ifBlank { BuildConfig.WECHAT_APP_ID }
            },
        )
    }

    private fun mergePay(base: TenantConfig): TenantConfig {
        val id = base.pay.wechatAppId.trim().ifBlank { BuildConfig.WECHAT_APP_ID.trim() }
        return if (id.isBlank()) base else base.copy(pay = base.pay.copy(wechatAppId = id))
    }

    private fun loadTenantConfig(): TenantConfig {
        val raw = runCatching {
            assets.open(ASSET_TENANT).bufferedReader().use { it.readText() }
        }.getOrNull()
        if (raw.isNullOrBlank()) return TenantConfig.demo()
        return runCatching { TenantConfigLoader.fromJson(raw) }.getOrElse { TenantConfig.demo() }
    }

    private fun mergeMapKey(base: TenantConfig): TenantConfig {
        val buildKey = BuildConfig.TENCENT_MAP_KEY.trim()
        val tenantKey = base.map.tencentKey.trim()
        val key = tenantKey.ifBlank { buildKey }
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

    private fun versionName(): String = runCatching {
        packageManager.getPackageInfo(packageName, 0).versionName
    }.getOrNull().orEmpty().ifBlank { RiderApp.LIBRARY_VERSION }

    override fun onTerminate() {
        appJob.cancel()
        if (::riderApp.isInitialized) {
            riderApp.close()
        }
        super.onTerminate()
    }

    companion object {
        const val ASSET_TENANT = "tenant.json"
        private const val TAG = "RiderApp"
    }
}
