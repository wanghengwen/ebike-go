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
}
