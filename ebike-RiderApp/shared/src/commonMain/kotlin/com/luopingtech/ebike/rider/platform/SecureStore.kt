package com.luopingtech.ebike.rider.platform

/**
 * Secure credential storage. Host apps provide platform backends;
 * [InMemorySecureStore] is for tests and the demo host.
 */
interface SecureStore {
    fun getString(key: String): String?
    fun putString(key: String, value: String)
    fun remove(key: String)
    fun clear()

    companion object {
        const val KEY_ACCESS_TOKEN = "access_token"
        const val KEY_REFRESH_TOKEN = "refresh_token"
        const val KEY_TENANT_ID = "tenant_id"
        const val KEY_DEVICE_ID = "device_id"
        const val KEY_SERVICE_AREA_ID = "service_area_id"
        const val KEY_SERVICE_AREA_NAME = "service_area_name"
        /** Persisted [com.luopingtech.ebike.rider.core.i18n.RiderLanguage.tag]. */
        const val KEY_LANGUAGE = "rider_language"
        /** Login dial code, e.g. `+86` (legacy UserAreaCodeActivity). */
        const val KEY_LOGIN_AREA_CODE = "login_area_code"
        /** ISO region for [KEY_LOGIN_AREA_CODE], e.g. `CN`. */
        const val KEY_LOGIN_AREA_REGION = "login_area_region"
        /** "1" when employee track upload should resume after process death. */
        const val KEY_TRACK_UPLOAD_ENABLED = "track_upload_enabled"
        /** Pipe-separated workbench favorite module ids (legacy toolNavCodes). */
        const val KEY_COMMON_MODULE_IDS = "common_module_ids"

        /**
         * JSON [com.luopingtech.ebike.rider.domain.riding.RideSession]. 杀进程后靠它恢复骑行相位。
         * 单靠 `getRideInfo` 不够：冷启动时那次请求可能还没回来，界面就已经要决定画首页还是画骑行页。
         */
        const val KEY_RIDE_SESSION = "ride_session"

        /** 用户 pin，骑行接口的 `userPin` 字段。 */
        const val KEY_USER_PIN = "user_pin"

        /** "1" = 申诉提交成功，回骑行页自动再跑一次还车（legacy `autoLock`）。 */
        const val KEY_RIDE_AUTO_RETURN = "ride_auto_return"
    }
}

class InMemorySecureStore : SecureStore {
    private val map = mutableMapOf<String, String>()

    override fun getString(key: String): String? = map[key]

    override fun putString(key: String, value: String) {
        map[key] = value
    }

    override fun remove(key: String) {
        map.remove(key)
    }

    override fun clear() {
        map.clear()
    }
}

expect fun createSecureStore(): SecureStore
