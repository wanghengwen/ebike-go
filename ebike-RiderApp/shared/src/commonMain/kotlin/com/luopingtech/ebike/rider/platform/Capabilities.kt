package com.luopingtech.ebike.rider.platform

import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.core.i18n.Strings
import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import kotlin.concurrent.Volatile
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flowOf

data class GeoPoint(val latitude: Double, val longitude: Double)

/**
 * Continuous location for rider track upload.
 * Hosts request runtime permission before [startTracking] on real devices.
 */
interface LocationTracker {
    suspend fun currentLocation(): RiderResult<GeoPoint>

    /** Stream of fixes while tracking is started. */
    fun track(): Flow<GeoPoint>

    fun startTracking() {}

    fun stopTracking() {}
}

class UnsupportedLocationTracker : LocationTracker {
    override suspend fun currentLocation(): RiderResult<GeoPoint> =
        RiderResult.Err(RiderError.unsupported("LocationTracker"))

    override fun track(): Flow<GeoPoint> = flowOf()
}

interface CodeScanner {
    suspend fun scanOnce(): RiderResult<String>
}

class UnsupportedCodeScanner : CodeScanner {
    override suspend fun scanOnce(): RiderResult<String> =
        RiderResult.Err(RiderError.unsupported("CodeScanner"))
}

/**
 * Host Activity binds a real scanner (CameraX / AVFoundation) after startup.
 * Shared features can depend on this port before UI is ready.
 */
class BindableCodeScanner : CodeScanner {
    @Volatile
    private var bound: CodeScanner? = null

    fun bind(scanner: CodeScanner) {
        bound = scanner
    }

    fun unbind(scanner: CodeScanner? = null) {
        if (scanner == null || bound === scanner) {
            bound = null
        }
    }

    val isBound: Boolean get() = bound != null

    override suspend fun scanOnce(): RiderResult<String> {
        val current = bound
        return current?.scanOnce()
            ?: RiderResult.Err(RiderError.unsupported("CodeScanner not bound to Activity"))
    }
}

/**
 * 拍照 / 从相册选图，返回平台本地引用（Android 是 `content://`，iOS 是文件路径），
 * 再交给 [MediaUploader] 上传——两者的字符串契约本来就是同一套。
 *
 * 做成端口而不是让界面直接调平台 API，是为了让上报类界面能放进 `commonMain`：
 * 运行时权限申请和 Activity result 都是 Android 专有的，留在 UI 里这些界面就搬不走。
 * 权限询问由实现方负责，调用侧只看结果。
 */
interface PhotoCapture {
    /** [prefix] 只用来拼缓存文件名，方便排查，不参与业务。 */
    suspend fun takePhoto(prefix: String = "photo"): RiderResult<String>

    suspend fun pickFromGallery(): RiderResult<String>
}

class UnsupportedPhotoCapture : PhotoCapture {
    override suspend fun takePhoto(prefix: String): RiderResult<String> =
        RiderResult.Err(RiderError.unsupported("PhotoCapture"))

    override suspend fun pickFromGallery(): RiderResult<String> =
        RiderResult.Err(RiderError.unsupported("PhotoCapture"))
}

/** 与 [BindableCodeScanner] 同构：宿主 Activity 起来之后再绑真实现。 */
class BindablePhotoCapture : PhotoCapture {
    @Volatile
    private var bound: PhotoCapture? = null

    fun bind(capture: PhotoCapture) {
        bound = capture
    }

    fun unbind(capture: PhotoCapture? = null) {
        if (capture == null || bound === capture) {
            bound = null
        }
    }

    val isBound: Boolean get() = bound != null

    override suspend fun takePhoto(prefix: String): RiderResult<String> =
        bound?.takePhoto(prefix)
            ?: RiderResult.Err(RiderError.unsupported("PhotoCapture not bound to Activity"))

    override suspend fun pickFromGallery(): RiderResult<String> =
        bound?.pickFromGallery()
            ?: RiderResult.Err(RiderError.unsupported("PhotoCapture not bound to Activity"))
}

enum class MapProviderKind { NONE, SIMULATOR, TENCENT, GOOGLE }

/**
 * Map rendering stays in the host UI. Shared only needs provider selection metadata.
 * [SIMULATOR] = open-source canvas projection (no vendor SDK / key).
 */
interface MapCapability {
    val kind: MapProviderKind
    val isReady: Boolean
}

class NoneMapCapability : MapCapability {
    override val kind: MapProviderKind = MapProviderKind.NONE
    override val isReady: Boolean = false
}

class SimulatorMapCapability : MapCapability {
    override val kind: MapProviderKind = MapProviderKind.SIMULATOR
    override val isReady: Boolean = true
}

/**
 * Tencent/Google adapters live in host modules. Until SDK is wired, [isReady] stays false
 * so UI can fall back to simulator rendering.
 */
class DeferredVendorMapCapability(
    override val kind: MapProviderKind,
    override val isReady: Boolean = false,
) : MapCapability

object MapCapabilityFactory {
    fun fromConfig(config: com.luopingtech.ebike.rider.core.config.TenantConfig): MapCapability {
        val raw = config.features.mapProvider.ifBlank { config.map.provider }.trim().lowercase()
        return when (raw) {
            "simulator", "sim", "canvas" -> SimulatorMapCapability()
            "tencent" -> {
                if (config.map.tencentKey.isNotBlank()) {
                    // Host Android module renders TextureMapView when isReady.
                    DeferredVendorMapCapability(MapProviderKind.TENCENT, isReady = true)
                } else {
                    // No key → simulator so rider UI remains usable.
                    SimulatorMapCapability()
                }
            }
            "google" -> {
                if (config.map.googleKey.isNotBlank()) {
                    DeferredVendorMapCapability(MapProviderKind.GOOGLE, isReady = false)
                } else {
                    SimulatorMapCapability()
                }
            }
            else -> NoneMapCapability()
        }
    }
}

/**
 * Upload local photo URIs (content:// / file:// / camera cache) to CDN/OSS via
 * backend `file/upload`. Already-remote http(s) / demo:// URLs pass through.
 */
interface MediaUploader {
    suspend fun upload(localUris: List<String>): RiderResult<List<String>>
}

/** Demo / offline: rewrite local URIs to stable https demo CDN keys. */
class DemoMediaUploader : MediaUploader {
    override suspend fun upload(localUris: List<String>): RiderResult<List<String>> {
        if (localUris.isEmpty()) {
            return RiderResult.Err(RiderError.business("MEDIA_EMPTY", Strings.t(Str.PhotoRequired)))
        }
        return RiderResult.Ok(
            localUris.map { uri ->
                val trimmed = uri.trim()
                when {
                    trimmed.startsWith("http://", ignoreCase = true) ||
                        trimmed.startsWith("https://", ignoreCase = true) -> trimmed
                    trimmed.startsWith("demo://") ->
                        "https://demo.cdn.rider/${trimmed.removePrefix("demo://")}"
                    else ->
                        "https://demo.cdn.rider/upload/${trimmed.hashCode().toUInt()}.jpg"
                }
            },
        )
    }
}

/**
 * Host binds a real uploader (multipart → `/business/ebike-management/file/upload`).
 */
class BindableMediaUploader : MediaUploader {
    @Volatile
    private var bound: MediaUploader? = null

    fun bind(uploader: MediaUploader) {
        bound = uploader
    }

    fun unbind(uploader: MediaUploader? = null) {
        if (uploader == null || bound === uploader) {
            bound = null
        }
    }

    val isBound: Boolean get() = bound != null

    override suspend fun upload(localUris: List<String>): RiderResult<List<String>> {
        val remoteReady = localUris.all { uri ->
            val t = uri.trim()
            t.startsWith("http://", ignoreCase = true) ||
                t.startsWith("https://", ignoreCase = true) ||
                t.startsWith("demo://")
        }
        if (remoteReady) {
            return DemoMediaUploader().upload(localUris)
        }
        val current = bound
        return current?.upload(localUris)
            ?: RiderResult.Err(RiderError.unsupported("MediaUploader not bound"))
    }
}

interface PushRegistrar {
    val isEnabled: Boolean
    suspend fun register(): RiderResult<String>
}

class DisabledPushRegistrar : PushRegistrar {
    override val isEnabled: Boolean = false
    override suspend fun register(): RiderResult<String> =
        RiderResult.Err(RiderError.unsupported("PushRegistrar"))
}
