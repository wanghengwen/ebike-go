package com.luopingtech.ebike.ops.platform

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
        prefs.edit().putString(key, value).apply()
    }

    override fun remove(key: String) {
        prefs.edit().remove(key).apply()
    }

    override fun clear() {
        prefs.edit().clear().apply()
    }

    companion object {
        private const val PREFS_NAME = "ebike_ops_secure"
        private const val PREFS_FALLBACK = "ebike_ops_secure_fallback"
    }
}

actual fun createSecureStore(): SecureStore = InMemorySecureStore()
