package com.luopingtech.ebike.ops.platform

import kotlin.test.AfterTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNull

/**
 * 往返测试守的是一类静默故障：写进 Keychain 但读不回来。
 *
 * 那种情况下 [KeychainSecureStore.putString] 会把 defaults 里的回退值删掉，
 * 读又拿不到 Keychain 的值，登录态在冷启动后凭空消失 —— 界面上只表现为
 * 「又要重新登录」，不会有任何报错。
 */
class KeychainSecureStoreTest {
    private val store = KeychainSecureStore(service = "com.luopingtech.ebike.ops.test")

    @AfterTest
    fun tearDown() {
        store.remove(KEY)
        store.remove(SecureStore.KEY_ACCESS_TOKEN)
    }

    @Test
    fun put_then_get_returns_value() {
        store.putString(KEY, "token-42")
        assertEquals("token-42", store.getString(KEY))
    }

    @Test
    fun put_overwrites_previous_value() {
        store.putString(KEY, "first")
        store.putString(KEY, "second")
        assertEquals("second", store.getString(KEY))
    }

    @Test
    fun non_ascii_survives_round_trip() {
        store.putString(KEY, "管理员·东区")
        assertEquals("管理员·东区", store.getString(KEY))
    }

    @Test
    fun remove_clears_value() {
        store.putString(KEY, "token-42")
        store.remove(KEY)
        assertNull(store.getString(KEY))
    }

    @Test
    fun clear_drops_known_keys() {
        store.putString(SecureStore.KEY_ACCESS_TOKEN, "token-42")
        store.clear()
        assertNull(store.getString(SecureStore.KEY_ACCESS_TOKEN))
    }

    private companion object {
        const val KEY = "test_credential"
    }
}
