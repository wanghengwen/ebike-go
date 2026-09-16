package com.luopingtech.ebike.rider.core.util

import com.luopingtech.ebike.rider.RiderApp
import com.luopingtech.ebike.rider.platform.InMemorySecureStore
import com.luopingtech.ebike.rider.platform.SecureStore
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class DeviceIdTest {
    @Test
    fun uniAppFormat_isTimestampPlusRoundedRandom() {
        val id = Ids.uniAppDeviceId(nowMillis = 1_726_380_000_000L, randomFraction = 0.5)
        assertEquals("172638000000050000000000", id)
        assertTrue(Ids.looksLikeUniAppDeviceId(id))
        assertTrue(id.all { it.isDigit() })
    }

    @Test
    fun resolveDeviceId_keepsStoredValue() {
        val store = InMemorySecureStore()
        store.putString(SecureStore.KEY_DEVICE_ID, "legacy-uuid-keep")
        assertEquals("legacy-uuid-keep", RiderApp.resolveDeviceId(store))
        assertFalse(Ids.looksLikeUniAppDeviceId("legacy-uuid-keep"))
    }

    @Test
    fun resolveDeviceId_newDeviceMatchesUniApp() {
        val store = InMemorySecureStore()
        val id = RiderApp.resolveDeviceId(store)
        assertTrue(Ids.looksLikeUniAppDeviceId(id), id)
        assertEquals(id, store.getString(SecureStore.KEY_DEVICE_ID))
        assertEquals(id, RiderApp.resolveDeviceId(store))
    }
}
