package com.luopingtech.ebike.ops.data.tenant

import com.luopingtech.ebike.ops.platform.InMemorySecureStore
import com.luopingtech.ebike.ops.platform.SecureStore
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNull
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking

class ServiceAreaRepositoryTest {
    @Test
    fun demoMode_loadsWithoutAutoSelect() = runBlocking {
        val store = InMemorySecureStore()
        val repo = ServiceAreaRepositoryImpl(secureStore = store, demoMode = true)
        val result = repo.loadAreas()
        assertTrue(result.isOk)
        assertEquals(3, result.getOrNull()?.size)
        assertNull(repo.currentArea())

        repo.selectArea(ServiceAreaRepositoryImpl.DEMO_AREAS[1])
        assertEquals("1002", repo.currentArea()?.id)
        assertEquals("1002", store.getString(SecureStore.KEY_SERVICE_AREA_ID))
    }

    @Test
    fun restore_fallsBackToIdAndName_whenJsonMissing() {
        val store = InMemorySecureStore()
        store.putString(SecureStore.KEY_SERVICE_AREA_ID, "1002")
        store.putString(SecureStore.KEY_SERVICE_AREA_NAME, "Demo 城西服务区")
        val repo = ServiceAreaRepositoryImpl(secureStore = store, demoMode = true)
        assertEquals("1002", repo.currentArea()?.id)
        assertEquals("Demo 城西服务区", repo.currentArea()?.name)
    }

    @Test
    fun restore_fromJson_onNewInstance() {
        val store = InMemorySecureStore()
        val first = ServiceAreaRepositoryImpl(secureStore = store, demoMode = true)
        first.selectArea(ServiceAreaRepositoryImpl.DEMO_AREAS[0])
        val restored = ServiceAreaRepositoryImpl(secureStore = store, demoMode = true)
        assertEquals("1001", restored.currentArea()?.id)
        assertEquals("Demo 城东服务区", restored.currentArea()?.name)
    }

    @Test
    fun demoMode_loadAreas_reselectsPersistedId() = runBlocking {
        val store = InMemorySecureStore()
        store.putString(SecureStore.KEY_SERVICE_AREA_ID, "1003")
        val repo = ServiceAreaRepositoryImpl(secureStore = store, demoMode = true)
        assertEquals("1003", repo.currentArea()?.id)
        val result = repo.loadAreas()
        assertTrue(result.isOk)
        assertEquals("1003", repo.currentArea()?.id)
        assertEquals("Demo 高新区", repo.currentArea()?.name)
    }
}
