package com.luopingtech.ebike.rider

import com.luopingtech.ebike.rider.core.config.H5ScreenKind
import com.luopingtech.ebike.rider.core.config.H5ScreenUrls
import com.luopingtech.ebike.rider.core.config.TenantConfig
import com.luopingtech.ebike.rider.core.i18n.RiderI18n
import com.luopingtech.ebike.rider.core.logging.RiderLogger
import com.luopingtech.ebike.rider.core.logging.StdoutLogger
import com.luopingtech.ebike.rider.core.network.AuthHeaderMode
import com.luopingtech.ebike.rider.core.network.CommonRequestBody
import com.luopingtech.ebike.rider.core.network.HttpClientFactory
import com.luopingtech.ebike.rider.core.network.NetworkSession
import com.luopingtech.ebike.rider.core.network.RequestAuth
import com.luopingtech.ebike.rider.core.network.SignedApiClient
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.core.signing.RequestSigner
import com.luopingtech.ebike.rider.core.time.nowEpochMillis
import com.luopingtech.ebike.rider.core.util.Ids
import com.luopingtech.ebike.rider.data.auth.AuthApi
import com.luopingtech.ebike.rider.data.auth.AuthRemote
import com.luopingtech.ebike.rider.data.auth.AuthRepository
import com.luopingtech.ebike.rider.data.auth.AuthRepositoryImpl
import com.luopingtech.ebike.rider.data.auth.PhoneNormalizer
import com.luopingtech.ebike.rider.data.auth.UserApi
import com.luopingtech.ebike.rider.data.auth.UserRemote
import com.luopingtech.ebike.rider.feature.auth.AuthFeature
import com.luopingtech.ebike.rider.data.ble.BleTokenApi
import com.luopingtech.ebike.rider.data.ble.BleTokenRemote
import com.luopingtech.ebike.rider.data.ble.DemoBleTokenApi
import com.luopingtech.ebike.rider.data.pay.DemoPayRemote
import com.luopingtech.ebike.rider.data.pay.PayApi
import com.luopingtech.ebike.rider.data.pay.PayRemote
import com.luopingtech.ebike.rider.data.map.DemoNearbyVehicleRemote
import com.luopingtech.ebike.rider.data.map.NearbyVehicleApi
import com.luopingtech.ebike.rider.data.map.NearbyVehicleRemote
import com.luopingtech.ebike.rider.data.home.DemoHomeNavRemote
import com.luopingtech.ebike.rider.data.home.HomeNavApi
import com.luopingtech.ebike.rider.data.home.HomeNavRemote
import com.luopingtech.ebike.rider.data.media.FileUploadApi
import com.luopingtech.ebike.rider.data.fence.FenceApi
import com.luopingtech.ebike.rider.data.fence.FenceRemote
import com.luopingtech.ebike.rider.data.riding.ApplyReturnApi
import com.luopingtech.ebike.rider.data.riding.ApplyReturnRemote
import com.luopingtech.ebike.rider.data.riding.BleRideApi
import com.luopingtech.ebike.rider.data.riding.BleRideRemote
import com.luopingtech.ebike.rider.data.riding.DemoApplyReturnRemote
import com.luopingtech.ebike.rider.data.riding.DemoBleRideRemote
import com.luopingtech.ebike.rider.data.riding.DemoFenceRemote
import com.luopingtech.ebike.rider.data.riding.DemoRideBook
import com.luopingtech.ebike.rider.data.riding.DemoRidingRemote
import com.luopingtech.ebike.rider.data.riding.RidingApi
import com.luopingtech.ebike.rider.data.riding.RidingRemote
import com.luopingtech.ebike.rider.feature.ble.BleSessionFeature
import com.luopingtech.ebike.rider.feature.riding.RideSessionStore
import com.luopingtech.ebike.rider.feature.riding.RidingFeature
import com.luopingtech.ebike.rider.feature.riding.UnlockPolicy
import com.luopingtech.ebike.rider.platform.BindableMediaUploader
import com.luopingtech.ebike.rider.platform.BindableWeChatPay
import com.luopingtech.ebike.rider.platform.BleTransport
import com.luopingtech.ebike.rider.platform.BleTransportFactory
import com.luopingtech.ebike.rider.platform.CodeScanner
import com.luopingtech.ebike.rider.platform.DemoMediaUploader
import com.luopingtech.ebike.rider.platform.DemoReverseGeocoder
import com.luopingtech.ebike.rider.platform.DeviceInfo
import com.luopingtech.ebike.rider.platform.LocationTracker
import com.luopingtech.ebike.rider.platform.MapCapability
import com.luopingtech.ebike.rider.platform.MapCapabilityFactory
import com.luopingtech.ebike.rider.platform.MediaUploader
import com.luopingtech.ebike.rider.platform.PhotoCapture
import com.luopingtech.ebike.rider.platform.QuickLoginProvider
import com.luopingtech.ebike.rider.platform.ReverseGeocoder
import com.luopingtech.ebike.rider.platform.SecureStore
import com.luopingtech.ebike.rider.platform.SimulatorLocationTracker
import com.luopingtech.ebike.rider.platform.UnsupportedCodeScanner
import com.luopingtech.ebike.rider.platform.UnsupportedLocationTracker
import com.luopingtech.ebike.rider.platform.UnsupportedPhotoCapture
import com.luopingtech.ebike.rider.platform.UnsupportedQuickLogin
import com.luopingtech.ebike.rider.platform.UnsupportedReverseGeocoder
import com.luopingtech.ebike.rider.platform.createDeviceInfo
import io.ktor.client.HttpClient
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.client.statement.bodyAsText
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.launch
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.contentOrNull
import kotlinx.serialization.json.jsonObject

/**
 * 组合根，对应 Android 的 `RiderApplication.onCreate` 与 iOS 的 `createIosRiderApp`。
 *
 * P1 认证与会话；P2 宿主能力端口；P3 BLE 帧 / 会话；P6 H5 长尾走 [signedApiClient] 代理。
 * [probe] 仍是签名通路的出口。
 */
class RiderApp private constructor(
    val config: TenantConfig,
    val logger: RiderLogger,
    val secureStore: SecureStore,
    val deviceInfo: DeviceInfo,
    /** 稳定设备标识，落在 [SecureStore.KEY_DEVICE_ID]，进公共请求体。 */
    val deviceId: String,
    val i18n: RiderI18n,
    val quickLogin: QuickLoginProvider,
    val codeScanner: CodeScanner,
    val photoCapture: PhotoCapture,
    val locationTracker: LocationTracker,
    val reverseGeocoder: ReverseGeocoder,
    val mapCapability: MapCapability,
    mediaUploader: MediaUploader?,
    bleTransportOverride: BleTransport?,
    nativeBle: BleTransport?,
) {
    private val httpFactory = HttpClientFactory(config, logger)
    private val json: Json = httpFactory.json
    private val httpClient: HttpClient = httpFactory.create()

    private val requestAuth = RequestAuth(config, secureStore)

    /** `api.baseUrl` 为空即 demo：不发请求，任意验证码可登录。 */
    val isDemoMode: Boolean get() = config.api.baseUrl.isBlank()

    private val tenantIdProvider: () -> String = {
        secureStore.getString(SecureStore.KEY_TENANT_ID)?.takeIf { it.isNotBlank() }
            ?: config.tenantId
    }

    private val authApi: AuthRemote? = if (isDemoMode) {
        null
    } else {
        AuthApi(
            client = httpClient,
            requestAuth = requestAuth,
            json = json,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = { deviceId },
            areaCodeProvider = {
                secureStore.getString(SecureStore.KEY_LOGIN_AREA_CODE)
                    ?.takeIf { it.isNotBlank() }
                    ?: PhoneNormalizer.DEFAULT_AREA_CODE
            },
        )
    }

    private var userRemote: UserRemote? = null

    val authRepository: AuthRepository = AuthRepositoryImpl(
        secureStore = secureStore,
        demoMode = isDemoMode,
        authApi = authApi,
        userApiProvider = { userRemote },
        defaultTenantId = config.tenantId,
    )

    /**
     * 业务接口走这里（00005 / 00013 单次刷新，00015 / 坏 refresh 清 session）。
     */
    val signedApiClient = SignedApiClient(
        client = httpClient,
        requestAuth = requestAuth,
        json = json,
        sessionProvider = {
            NetworkSession(
                accessToken = authRepository.currentSession()?.accessToken
                    ?: secureStore.getString(SecureStore.KEY_ACCESS_TOKEN).orEmpty(),
            )
        },
        refreshAccessToken = {
            when (val result = authRepository.refreshAccessToken()) {
                is RiderResult.Ok -> RiderResult.Ok(result.value.accessToken)
                is RiderResult.Err -> result
            }
        },
        onSessionInvalid = { _, message ->
            authRepository.markSessionInvalid(message)
        },
    )

    val authFeature: AuthFeature = AuthFeature(authRepository)

    val bleTransport: BleTransport = BleTransportFactory.fromConfig(
        config = config,
        override = bleTransportOverride,
        native = nativeBle,
    )

    val bleTokenApi: BleTokenApi = if (isDemoMode) {
        DemoBleTokenApi()
    } else {
        BleTokenRemote(
            signedApi = signedApiClient,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = { deviceId },
        )
    }

    val bleSession: BleSessionFeature = BleSessionFeature(
        transport = bleTransport,
        tokenApi = bleTokenApi,
        logger = logger,
    )

    // ── P4 骑行域 ───────────────────────────────────────────────────────────

    /** 两个 demo remote 必须共用一本假订单账本，否则蓝牙开锁后骑行页一直是 0 元。 */
    private val demoRideBook = DemoRideBook()

    val ridingRemote: RidingRemote = if (isDemoMode) {
        DemoRidingRemote(demoRideBook)
    } else {
        RidingApi(
            signedApi = signedApiClient,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = { deviceId },
        )
    }

    val bleRideRemote: BleRideRemote = if (isDemoMode) {
        DemoBleRideRemote(demoRideBook)
    } else {
        BleRideApi(
            signedApi = signedApiClient,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = { deviceId },
        )
    }

    val applyReturnRemote: ApplyReturnRemote = if (isDemoMode) {
        DemoApplyReturnRemote()
    } else {
        ApplyReturnApi(
            signedApi = signedApiClient,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = { deviceId },
        )
    }

    val fenceRemote: FenceRemote = if (isDemoMode) {
        DemoFenceRemote()
    } else {
        FenceApi(
            signedApi = signedApiClient,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = { deviceId },
        )
    }

    val rideSessionStore: RideSessionStore = RideSessionStore(secureStore)

    val wechatPay: BindableWeChatPay = BindableWeChatPay()

    val payRemote: PayRemote = if (isDemoMode) {
        DemoPayRemote()
    } else {
        PayApi(
            signedApi = signedApiClient,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = { deviceId },
        )
    }

    val ridingFeature: RidingFeature = RidingFeature(
        riding = ridingRemote,
        bleRide = bleRideRemote,
        applyReturn = applyReturnRemote,
        fences = fenceRemote,
        bleSession = bleSession,
        locationTracker = locationTracker,
        store = rideSessionStore,
        logger = logger,
        unlockPolicy = if (config.features.onlyBluetooth) {
            UnlockPolicy.BleOnly
        } else {
            // 对齐 UniApp：默认远程开锁，仅车机离线 / 17012 再降级蓝牙。
            // 蓝牙优先在地库场景有优势，但真机常因 imei/扫描失败导致「指令像没下发」。
            UnlockPolicy.NetworkPreferred
        },
        bleAvailable = { bleTransport.isAvailable },
        pay = payRemote,
        wechatPay = wechatPay,
        payChannelType = { config.pay.channelType },
        wechatAppId = { config.pay.wechatAppId },
        paySaleType = { config.pay.saleType },
    )

    val mediaUploader: MediaUploader = mediaUploader
        ?: if (isDemoMode) DemoMediaUploader() else BindableMediaUploader()

    val nearbyVehicleRemote: NearbyVehicleRemote = if (isDemoMode) {
        DemoNearbyVehicleRemote()
    } else {
        NearbyVehicleApi(
            signedApi = signedApiClient,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = { deviceId },
        )
    }

    val homeNavRemote: HomeNavRemote = if (isDemoMode) {
        DemoHomeNavRemote()
    } else {
        HomeNavApi(
            signedApi = signedApiClient,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = { deviceId },
        )
    }

    val fileUploadApi: FileUploadApi? = if (isDemoMode) {
        null
    } else {
        FileUploadApi(
            client = httpClient,
            requestAuth = requestAuth,
            json = json,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = { deviceId },
            sessionProvider = {
                NetworkSession(
                    accessToken = authRepository.currentSession()?.accessToken
                        ?: secureStore.getString(SecureStore.KEY_ACCESS_TOKEN).orEmpty(),
                )
            },
        )
    }

    init {
        if (!isDemoMode) {
            userRemote = UserApi(
                signedApi = signedApiClient,
                tenantIdProvider = tenantIdProvider,
                deviceInfo = deviceInfo,
                deviceIdProvider = { deviceId },
            )
        }
    }

    fun start(scope: CoroutineScope) {
        authFeature.start(scope)
        scope.launch { authFeature.restoreSession() }
        // pin 是骑行接口的必填参数，但它只在登录态里；镜像一份到 SecureStore，
        // 让冷启动恢复的骑行页不必等 restoreSession 回来才能调 getRideInfo。
        scope.launch {
            authFeature.state.collect { auth ->
                rideSessionStore.userPin = auth.session?.pin.orEmpty()
                if (auth.session == null) rideSessionStore.clear()
            }
        }
        ridingFeature.start(scope)
    }

    /**
     * 一次公开配置接口探针。选它是因为不需要登录态 —— 拿到业务码就等于证明
     * `_t` / `_s` 被网关认了，而不是先撞上 401。
     */
    suspend fun probe(path: String = DEFAULT_PROBE_PATH): ProbeResult {
        val body = CommonRequestBody.toJsonString(
            tenantId = config.tenantId,
            deviceInfo = deviceInfo,
            deviceId = deviceId,
        )
        if (isDemoMode) {
            val timestamp = nowEpochMillis().toString()
            return ProbeResult(
                path = path,
                requestBody = body,
                timestamp = timestamp,
                sign = RequestSigner.signPostJson(body, timestamp, config.auth.signSecret),
                httpStatus = null,
                businessCode = "",
                message = "demo mode — api.baseUrl 为空，未发出请求",
                isDemo = true,
            )
        }

        var timestamp = ""
        var sign = ""
        return try {
            val response = httpClient.post(path.trimStart('/')) {
                requestAuth.applyPostJson(this, body, NetworkSession(), AuthHeaderMode.None)
                timestamp = headers[RequestSigner.HEADER_TIMESTAMP].orEmpty()
                sign = headers[RequestSigner.HEADER_SIGN].orEmpty()
                setBody(body)
            }
            val raw = response.bodyAsText()
            ProbeResult(
                path = path,
                requestBody = body,
                timestamp = timestamp,
                sign = sign,
                httpStatus = response.status.value,
                businessCode = peek(raw, "code"),
                message = peek(raw, "msg").ifBlank { peek(raw, "message") },
                isDemo = false,
                responsePreview = raw.take(RESPONSE_PREVIEW_CHARS),
            )
        } catch (t: Throwable) {
            logger.e(TAG, "probe $path failed", t)
            ProbeResult(
                path = path,
                requestBody = body,
                timestamp = timestamp,
                sign = sign,
                httpStatus = null,
                businessCode = "",
                message = "",
                isDemo = false,
                error = t.message ?: t.toString(),
            )
        }
    }

    fun resolveH5Url(kind: H5ScreenKind): String? =
        H5ScreenUrls.resolve(config, kind, i18n.language)

    fun resolveH5Hash(route: String): String? =
        H5ScreenUrls.resolveHash(config, route, i18n.language)

    fun close() {
        httpClient.close()
    }

    private fun peek(raw: String, key: String): String {
        if (raw.isBlank()) return ""
        return runCatching {
            (json.parseToJsonElement(raw).jsonObject[key] as? JsonPrimitive)?.contentOrNull.orEmpty()
        }.getOrDefault("")
    }

    companion object {
        const val LIBRARY_VERSION: String = "0.1.0-p6.1"
        const val DEFAULT_PROBE_PATH: String = "client/systemConfig/getConfigBaseItem"
        private const val TAG = "RiderApp"
        private const val RESPONSE_PREVIEW_CHARS = 240

        /**
         * 宿主注入入口。参数一律显式传入，**不要**加默认值——Kotlin 的 `create$default`
         * 在 Apply Changes / 分层 dex 下极易与 androidApp 调用点签名错位（NoSuchMethodError）。
         */
        fun create(
            config: TenantConfig,
            secureStore: SecureStore,
            logger: RiderLogger,
            appVersion: String,
            deviceInfo: DeviceInfo,
            /** 首启跟随系统语言；用户存过的选择仍然优先（见 [RiderI18n.fromStore]）。 */
            systemLanguage: String?,
            quickLogin: QuickLoginProvider,
            codeScanner: CodeScanner,
            photoCapture: PhotoCapture,
            locationTracker: LocationTracker?,
            reverseGeocoder: ReverseGeocoder?,
            mediaUploader: MediaUploader?,
            bleTransport: BleTransport?,
            nativeBle: BleTransport?,
        ): RiderApp {
            if (config.tenantId.isNotBlank()) {
                secureStore.putString(SecureStore.KEY_TENANT_ID, config.tenantId)
            }
            return RiderApp(
                config = config,
                logger = logger,
                secureStore = secureStore,
                deviceInfo = deviceInfo,
                deviceId = resolveDeviceId(secureStore),
                i18n = RiderI18n.fromStore(secureStore, systemLanguage),
                quickLogin = quickLogin,
                codeScanner = codeScanner,
                photoCapture = photoCapture,
                locationTracker = locationTracker
                    ?: if (config.api.baseUrl.isBlank()) {
                        SimulatorLocationTracker()
                    } else {
                        UnsupportedLocationTracker()
                    },
                reverseGeocoder = reverseGeocoder
                    ?: if (config.api.baseUrl.isBlank()) {
                        DemoReverseGeocoder()
                    } else {
                        UnsupportedReverseGeocoder()
                    },
                mapCapability = MapCapabilityFactory.fromConfig(config),
                mediaUploader = mediaUploader,
                bleTransportOverride = bleTransport,
                nativeBle = nativeBle,
            )
        }

        /** 设备标识要跨启动稳定。新设备用 UniApp 规则；已存值原样复用。 */
        internal fun resolveDeviceId(secureStore: SecureStore): String {
            val stored = secureStore.getString(SecureStore.KEY_DEVICE_ID)
            if (!stored.isNullOrBlank()) return stored
            val fresh = Ids.uniAppDeviceId()
            secureStore.putString(SecureStore.KEY_DEVICE_ID, fresh)
            return fresh
        }
    }
}

/**
 * 探针结果。[timestamp] / [sign] 是真正发出去的 `_t` / `_s`，Debug 屏直接显示它们，
 * 好跟 UniApp / Go 网关的日志对照。
 */
data class ProbeResult(
    val path: String,
    val requestBody: String,
    val timestamp: String,
    val sign: String,
    /** null = 没发出去（demo 模式或传输失败）。 */
    val httpStatus: Int?,
    val businessCode: String,
    val message: String,
    val isDemo: Boolean,
    val responsePreview: String = "",
    val error: String? = null,
) {
    /** 签名前 8 位 + 后 8 位，够对日志又不整段抄出去。 */
    val signDigest: String
        get() = if (sign.length <= SIGN_DIGEST_EDGE * 2) {
            sign
        } else {
            sign.take(SIGN_DIGEST_EDGE) + "…" + sign.takeLast(SIGN_DIGEST_EDGE)
        }

    private companion object {
        const val SIGN_DIGEST_EDGE = 8
    }
}
