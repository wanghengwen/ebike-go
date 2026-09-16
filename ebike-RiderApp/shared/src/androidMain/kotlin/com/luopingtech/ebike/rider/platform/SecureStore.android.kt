package com.luopingtech.ebike.rider.platform

import android.content.Context
import android.content.SharedPreferences
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKeys

/**
 * Android encrypted prefs for tokens. Falls back to private prefs if crypto init fails.
 */
class AndroidSecureStore(context: Context) : SecureStore {
    private val prefs: SharedPreferences = runCatching {
        val appContext = context.applicationContext
        val masterKeyAlias = MasterKeys.getOrCreate(MasterKeys.AES256_GCM_SPEC)
        EncryptedSharedPreferences.create(
            PREFS_NAME,
            masterKeyAlias,
            appContext,
            EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
            EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM,
        )
    }.getOrElse {
        context.applicationContext.getSharedPreferences(PREFS_FALLBACK, Context.MODE_PRIVATE)
    }

    override fun getString(key: String): String? = prefs.getString(key, null)

    override fun putString(key: String, value: String) {
        // commit：运营区等关键键需在进程被杀前落盘，避免冷启动读不到记录又闪选区页。
        prefs.edit().putString(key, value).commit()
    }

    override fun remove(key: String) {
        prefs.edit().remove(key).commit()
    }

    override fun clear() {
        prefs.edit().clear().commit()
    }

    companion object {
        private const val PREFS_NAME = "ebike_rider_secure"
        private const val PREFS_FALLBACK = "ebike_rider_secure_fallback"
    }
}

actual fun createSecureStore(): SecureStore = InMemorySecureStore()
