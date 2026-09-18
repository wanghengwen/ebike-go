package com.luopingtech.ebike.rider.core.h5

import com.luopingtech.ebike.rider.RiderApp
import com.luopingtech.ebike.rider.core.i18n.RiderLanguage
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.core.util.Ids
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.contentOrNull
import kotlinx.serialization.json.doubleOrNull
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import kotlinx.serialization.json.put

/**
 * UniApp `nativeHost.ts` 的原生实现。Android 挂 `window.__riderNative`，
 * iOS 挂 `webkit.messageHandlers.riderNative`；两端都把 `{id,method,args}` JSON 送到这里。
 */
class NativeHostHooks(
    var close: () -> Unit = {},
    var setTitle: (String) -> Unit = {},
    var toast: (String) -> Unit = {},
    var navigate: (type: String, url: String?) -> Unit = { _, _ -> },
    var onLoginRequired: () -> Unit = {},
    var openNavigation: (lat: Double, lng: Double, name: String) -> Unit = { _, _, _ -> },
)

class NativeHostBridge(
    private val app: RiderApp,
    val hooks: NativeHostHooks,
    private val json: Json = Json { ignoreUnknownKeys = true; isLenient = true },
) {
    var evaluateJs: (String) -> Unit = {}
    var currentPageUrl: () -> String? = { null }

    fun platform(): String = app.deviceInfo.platform

    suspend fun handlePayload(payload: String): Pair<String, String> {
        val root = runCatching { json.parseToJsonElement(payload).jsonObject }.getOrNull()
        val id = root?.string("id").orEmpty()
        if (root == null) {
            return id to errorObj("BAD_PAYLOAD", "invalid bridge payload").toString()
        }
        val method = root.string("method").orEmpty()
        val args = root["args"] as? JsonObject ?: buildJsonObject {}
        val result = runCatching { dispatch(method, args) }
            .getOrElse { errorObj("BRIDGE", it.message ?: it.toString()) }
        return id to result.toString()
    }

    suspend fun dispatch(method: String, args: JsonObject): JsonObject {
        if (!originAllowed() && method !in ORIGINLESS) {
            return errorObj("ORIGIN", "origin not allowed")
        }
        return when (method) {
            "platform" -> buildJsonObject { put("platform", platform()) }
            "getLanguage" -> buildJsonObject { put("language", app.i18n.language.tag) }
            "setLanguage" -> {
                val tag = args.string("language") ?: args.string("lang").orEmpty()
                app.i18n.setLanguage(RiderLanguage.fromTag(tag))
                buildJsonObject {
                    put("ok", true)
                    put("language", app.i18n.language.tag)
                }
            }
            "getProfile" -> resolveProfile()
            "request" -> proxyRequest(args)
            "pay" -> buildJsonObject {
                put("success", false)
                put("code", "UNSUPPORTED")
                put("msg", app.i18n.t(Str.H5PayUnsupported))
            }
            "scanCode" -> when (val result = app.codeScanner.scanOnce()) {
                is RiderResult.Ok -> buildJsonObject { put("code", result.value) }
                is RiderResult.Err -> errorObj(result.error.code, result.error.message)
            }
            "capturePhoto" -> captureAndMaybeUpload(args)
            "currentLocation" -> when (val result = app.locationTracker.currentLocation()) {
                is RiderResult.Ok -> buildJsonObject {
                    put("latitude", result.value.latitude)
                    put("longitude", result.value.longitude)
                }
                is RiderResult.Err -> errorObj(result.error.code, result.error.message)
            }
            "openNavigation" -> {
                val lat = args.double("latitude") ?: args.double("lat")
                val lng = args.double("longitude") ?: args.double("lng")
                if (lat == null || lng == null) {
                    errorObj("BAD_ARGS", "latitude/longitude required")
                } else {
                    hooks.openNavigation(lat, lng, args.string("name").orEmpty())
                    ok()
                }
            }
            "navigate" -> {
                val type = args.string("type") ?: "to"
                val url = args.string("url")
                if (isLoginNavigation(url)) {
                    hooks.onLoginRequired()
                    hooks.close()
                } else {
                    hooks.navigate(type, url)
                }
                ok()
            }
            "close" -> {
                hooks.close()
                ok()
            }
            "setTitle" -> {
                hooks.setTitle(args.string("title").orEmpty())
                ok()
            }
            "toast" -> {
                val text = args.string("title") ?: args.string("message") ?: args.string("text").orEmpty()
                hooks.toast(text)
                ok()
            }
            else -> errorObj("UNKNOWN", "unknown method: $method")
        }
    }

    fun replyJs(id: String, resultJson: String): String {
        val idJson = JsonPrimitive(id).toString()
        return "window.__riderNativeOnResult&&window.__riderNativeOnResult($idJson,$resultJson);"
    }

    private fun originAllowed(): Boolean =
        H5OriginPolicy.isAllowed(currentPageUrl(), app.config.h5.baseUrl)

    private suspend fun resolveProfile(): JsonObject {
        var sid = app.rideSessionStore.serviceAreaId
        if (sid.isBlank() && !app.isDemoMode) {
            when (val loc = app.locationTracker.currentLocation()) {
                is RiderResult.Ok -> when (val area = app.fenceRemote.serviceAreaIdAt(loc.value)) {
                    is RiderResult.Ok -> {
                        if (area.value.isNotBlank()) {
                            app.rideSessionStore.serviceAreaId = area.value
                            sid = area.value
                        }
                    }
                    is RiderResult.Err -> Unit
                }
                is RiderResult.Err -> Unit
            }
        }
        return H5ProfileView.from(
            session = app.authRepository.currentSession(),
            tenantId = app.config.tenantId,
            serviceAreaId = sid,
        )
    }

    private suspend fun proxyRequest(args: JsonObject): JsonObject {
        val rawUrl = args.string("url") ?: args.string("path").orEmpty()
        if (!H5RequestGuard.isAllowed(rawUrl)) {
            return envelope(success = false, code = "FORBIDDEN", msg = "path not allowed")
        }
        if (app.isDemoMode) {
            return envelope(success = false, code = "DEMO", msg = "demo mode — api.baseUrl empty")
        }
        val extra = args["data"] as? JsonObject
        val body = mergeClientBody(extra)
        val path = H5RequestGuard.normalizePath(rawUrl).trimStart('/')
        return try {
            when (val result = app.signedApiClient.postRaw(path, body)) {
                is RiderResult.Ok -> parseEnvelope(result.value)
                is RiderResult.Err -> {
                    if (result.error.code == "UNAUTHORIZED") {
                        hooks.onLoginRequired()
                    }
                    envelope(false, result.error.code, result.error.message)
                }
            }
        } catch (t: Throwable) {
            envelope(false, "NETWORK", t.message ?: app.i18n.t(Str.NetworkError))
        }
    }

    private suspend fun captureAndMaybeUpload(args: JsonObject): JsonObject {
        val source = args.string("source")
            ?.trim()
            ?.lowercase()
            .orEmpty()
        hooks.toast(app.i18n.t(Str.ApplyReturnAddPhoto))
        val shot = when (source) {
            "album", "gallery" -> app.photoCapture.pickFromGallery()
            else -> app.photoCapture.takePhoto("h5")
        }
        return when (shot) {
            is RiderResult.Err -> {
                hooks.toast(shot.error.message.ifBlank { shot.error.code })
                errorObj(shot.error.code, shot.error.message)
            }
            is RiderResult.Ok -> {
                when (val uploaded = app.mediaUploader.upload(listOf(shot.value))) {
                    is RiderResult.Ok -> buildJsonObject {
                        put("uri", uploaded.value.firstOrNull().orEmpty())
                    }
                    is RiderResult.Err -> {
                        hooks.toast(uploaded.error.message.ifBlank { uploaded.error.code })
                        errorObj(uploaded.error.code, uploaded.error.message)
                    }
                }
            }
        }
    }

    private fun mergeClientBody(extra: JsonObject?): String = buildJsonObject {
        extra?.forEach { (key, value) ->
            if (key !in STAMPED_KEYS) put(key, value)
        }
        put("traceId", Ids.uuidV4())
        put("platform", app.deviceInfo.platform)
        put("deviceId", app.deviceId)
        put(
            "tenantId",
            app.authRepository.currentSession()?.tenantId?.ifBlank { app.config.tenantId }
                ?: app.config.tenantId,
        )
    }.toString()

    private fun parseEnvelope(raw: String): JsonObject {
        if (raw.isBlank()) return envelope(false, "EMPTY", "empty response")
        return runCatching { json.parseToJsonElement(raw).jsonObject }
            .getOrElse { envelope(false, "PARSE", it.message ?: "parse failed") }
    }

    companion object {
        private val ORIGINLESS = setOf("getLanguage", "platform")
        private val STAMPED_KEYS = setOf("traceId", "platform", "deviceId", "tenantId")

        fun isLoginNavigation(url: String?): Boolean {
            if (url.isNullOrBlank()) return false
            val u = url.lowercase()
            return u == "login" ||
                u.startsWith("native://login") ||
                u.contains("/pages/auth/phone-login") ||
                u.contains("/pages/auth/quick-login")
        }

        fun ok(): JsonObject = buildJsonObject { put("ok", true) }

        fun errorObj(code: String, msg: String): JsonObject = buildJsonObject {
            put("__error", true)
            put("success", false)
            put("code", code)
            put("msg", msg)
        }

        fun envelope(success: Boolean, code: String, msg: String, data: JsonElement? = null): JsonObject =
            buildJsonObject {
                put("success", success)
                put("code", code)
                put("msg", msg)
                if (data != null) put("data", data)
            }
    }
}

private fun JsonObject.string(key: String): String? =
    (this[key] as? JsonPrimitive)?.contentOrNull

private fun JsonObject.double(key: String): Double? {
    val primitive = this[key] as? JsonPrimitive ?: return null
    primitive.doubleOrNull?.let { return it }
    return primitive.contentOrNull?.toDoubleOrNull()
}
