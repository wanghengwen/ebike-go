package com.luopingtech.ebike.ops.platform

import kotlinx.cinterop.BetaInteropApi
import kotlinx.cinterop.ExperimentalForeignApi
import kotlinx.cinterop.addressOf
import kotlinx.cinterop.alloc
import kotlinx.cinterop.memScoped
import kotlinx.cinterop.ptr
import kotlinx.cinterop.usePinned
import kotlinx.cinterop.value
import platform.CoreFoundation.CFDictionaryAddValue
import platform.CoreFoundation.CFDictionaryCreateMutable
import platform.CoreFoundation.CFMutableDictionaryRef
import platform.CoreFoundation.CFRelease
import platform.CoreFoundation.CFTypeRefVar
import platform.CoreFoundation.kCFAllocatorDefault
import platform.CoreFoundation.kCFBooleanTrue
import platform.Foundation.CFBridgingRelease
import platform.Foundation.CFBridgingRetain
import platform.Foundation.NSData
import platform.Foundation.NSUserDefaults
import platform.Foundation.create
import platform.Security.SecItemAdd
import platform.Security.SecItemCopyMatching
import platform.Security.SecItemDelete
import platform.Security.errSecSuccess
import platform.Security.kSecAttrAccessible
import platform.Security.kSecAttrAccessibleAfterFirstUnlock
import platform.Security.kSecAttrAccount
import platform.Security.kSecAttrService
import platform.Security.kSecClass
import platform.Security.kSecClassGenericPassword
import platform.Security.kSecMatchLimit
import platform.Security.kSecMatchLimitOne
import platform.Security.kSecReturnData
import platform.Security.kSecValueData
import platform.posix.memcpy

/**
 * Keychain 版凭证存储，对应 Android 的 `AndroidSecureStore`。
 *
 * 每个 key 一条 generic password item，service 固定、account 用业务 key。
 * `kSecAttrAccessibleAfterFirstUnlock` 是后台轨迹上报的硬要求：锁屏后进程被唤醒
 * 仍要读得到 token，用默认的 WhenUnlocked 会读空。
 *
 * 模拟器与部分未配 keychain entitlement 的宿主上 `SecItemAdd` 会返回
 * -34018，此时落到 [NSUserDefaults]，行为与 Android 的 prefs 回退一致：
 * 宁可降级也不要让登录态整个失效。
 */
@OptIn(ExperimentalForeignApi::class, BetaInteropApi::class)
class KeychainSecureStore(
    private val service: String = "com.luopingtech.ebike.ops",
) : SecureStore {

    private val defaults = NSUserDefaults.standardUserDefaults

    override fun getString(key: String): String? =
        keychainRead(key) ?: defaults.stringForKey(fallbackKey(key))

    override fun putString(key: String, value: String) {
        if (keychainWrite(key, value)) {
            defaults.removeObjectForKey(fallbackKey(key))
        } else {
            defaults.setObject(value, forKey = fallbackKey(key))
        }
    }

    override fun remove(key: String) {
        keychainDelete(key)
        defaults.removeObjectForKey(fallbackKey(key))
    }

    override fun clear() {
        KNOWN_KEYS.forEach { remove(it) }
    }

    private fun fallbackKey(key: String) = "$service.$key"

    private fun keychainRead(key: String): String? = memScoped {
        val out = alloc<CFTypeRefVar>()
        val retained = mutableListOf<Any?>()
        val query = baseQuery(key, retained) { dict ->
            CFDictionaryAddValue(dict, kSecReturnData, kCFBooleanTrue)
            CFDictionaryAddValue(dict, kSecMatchLimit, kSecMatchLimitOne)
        }
        val status = SecItemCopyMatching(query, out.ptr)
        CFRelease(query)
        retained.forEach { CFRelease(it as? kotlinx.cinterop.COpaquePointer) }
        if (status != errSecSuccess) return@memScoped null
        val data = CFBridgingRelease(out.value) as? NSData ?: return@memScoped null
        data.toByteArray()?.decodeToString()
    }

    private fun keychainWrite(key: String, value: String): Boolean {
        val payload = value.toNSData() ?: return false
        // Add 之前先删：SecItemUpdate 还要再拼一份 attributes，删+加更短且等效。
        keychainDelete(key)
        val retained = mutableListOf<Any?>()
        val attributes = baseQuery(key, retained) { dict ->
            val data = CFBridgingRetain(payload)
            retained += data
            CFDictionaryAddValue(dict, kSecValueData, data)
            CFDictionaryAddValue(dict, kSecAttrAccessible, kSecAttrAccessibleAfterFirstUnlock)
        }
        val status = SecItemAdd(attributes, null)
        CFRelease(attributes)
        retained.forEach { CFRelease(it as? kotlinx.cinterop.COpaquePointer) }
        return status == errSecSuccess
    }

    private fun keychainDelete(key: String) {
        val retained = mutableListOf<Any?>()
        val query = baseQuery(key, retained) {}
        SecItemDelete(query)
        CFRelease(query)
        retained.forEach { CFRelease(it as? kotlinx.cinterop.COpaquePointer) }
    }

    private inline fun baseQuery(
        key: String,
        retained: MutableList<Any?>,
        extra: (CFMutableDictionaryRef?) -> Unit,
    ): CFMutableDictionaryRef? {
        val dict = CFDictionaryCreateMutable(kCFAllocatorDefault, 0, null, null)
        CFDictionaryAddValue(dict, kSecClass, kSecClassGenericPassword)
        // CFBridgingRetain 收 Any?，Kotlin String 会自动桥接成 NSString；
        // 显式 `as NSString` 反而是编译器判定永不成立的强转。
        val serviceRef = CFBridgingRetain(service)
        retained += serviceRef
        CFDictionaryAddValue(dict, kSecAttrService, serviceRef)
        val accountRef = CFBridgingRetain(key)
        retained += accountRef
        CFDictionaryAddValue(dict, kSecAttrAccount, accountRef)
        extra(dict)
        return dict
    }

    private companion object {
        /**
         * Keychain 没有「按 service 清空」的单条调用，而共享层只会存这些键，
         * 逐个删比 SecItemCopyMatching 拉全表再删更省事。
         */
        val KNOWN_KEYS = listOf(
            SecureStore.KEY_ACCESS_TOKEN,
            SecureStore.KEY_REFRESH_TOKEN,
            SecureStore.KEY_TENANT_ID,
            SecureStore.KEY_DEVICE_ID,
            SecureStore.KEY_SERVICE_AREA_ID,
            SecureStore.KEY_SERVICE_AREA_NAME,
            SecureStore.KEY_LANGUAGE,
            SecureStore.KEY_LOGIN_AREA_CODE,
            SecureStore.KEY_LOGIN_AREA_REGION,
            SecureStore.KEY_TRACK_UPLOAD_ENABLED,
            SecureStore.KEY_COMMON_MODULE_IDS,
        )
    }
}

/**
 * Keychain 的载荷是 CFData，而 Kotlin String ↔ NSString 的强转在 K/N 上不成立
 * （编译器会判定 cast 永不成功，运行时读出来永远是 null，凭据会悄悄退到明文 defaults）。
 * 所以两个方向都走字节：编码/解码自己做，不碰 NSString 桥接。
 */
@OptIn(ExperimentalForeignApi::class, BetaInteropApi::class)
private fun String.toNSData(): NSData? {
    val bytes = encodeToByteArray()
    if (bytes.isEmpty()) return NSData()
    return bytes.usePinned { pinned ->
        NSData.create(bytes = pinned.addressOf(0), length = bytes.size.toULong())
    }
}

@OptIn(ExperimentalForeignApi::class)
private fun NSData.toByteArray(): ByteArray? {
    val size = length.toInt()
    if (size == 0) return ByteArray(0)
    val source = bytes ?: return null
    val out = ByteArray(size)
    out.usePinned { pinned -> memcpy(pinned.addressOf(0), source, length) }
    return out
}

actual fun createSecureStore(): SecureStore = KeychainSecureStore()
