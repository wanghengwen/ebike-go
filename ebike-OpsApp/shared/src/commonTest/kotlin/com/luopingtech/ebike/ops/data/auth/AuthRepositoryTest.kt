package com.luopingtech.ebike.ops.data.auth

import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.platform.InMemorySecureStore
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertNull
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking

class AuthRepositoryTest {
    @Test
    fun demoPassword_hq_requiresBusinessSelect() = runBlocking {
        val repo = AuthRepositoryImpl(InMemorySecureStore(), demoMode = true)
        val result = repo.login("hq", "any")
        assertTrue(result is OpsResult.Err)
        assertEquals(AuthRepositoryImpl.CODE_SELECT_BUSINESS, (result as OpsResult.Err).error.code)
        assertEquals(2, repo.pendingBusinesses().size)
        assertNull(repo.currentSession())
    }

    @Test
    fun demoSms_thenSelectBusiness_signsIn() = runBlocking {
        val repo = AuthRepositoryImpl(InMemorySecureStore(), demoMode = true)
        assertTrue(repo.sendSmsCode("13800138000") is OpsResult.Ok)
        val login = repo.loginWithSms("13800138000", "1234")
        assertTrue(login is OpsResult.Err)
        val tenantId = repo.pendingBusinesses().first().tenantId
        val selected = repo.selectBusiness(tenantId)
        assertTrue(selected is OpsResult.Ok)
        assertEquals(tenantId, (selected as OpsResult.Ok).value.tenantId)
        assertNotNull(repo.currentSession())
        assertTrue(repo.pendingBusinesses().isEmpty())
    }

    @Test
    fun demoPassword_demo_signsInDirectly() = runBlocking {
        val repo = AuthRepositoryImpl(InMemorySecureStore(), demoMode = true)
        val result = repo.login("demo", "demo")
        assertTrue(result is OpsResult.Ok)
        assertEquals("demo", (result as OpsResult.Ok).value.displayName)
    }

    @Test
    fun demoNopwd_requiresSetPassword() = runBlocking {
        val repo = AuthRepositoryImpl(InMemorySecureStore(), demoMode = true)
        val result = repo.login("nopwd", "x")
        assertTrue(result is OpsResult.Ok)
        assertTrue(repo.needsSetPassword())
        val set = repo.setPassword("password1")
        assertTrue(set is OpsResult.Ok)
        assertTrue(!repo.needsSetPassword())
    }

    @Test
    fun demoForgetPassword_thenLogin() = runBlocking {
        val repo = AuthRepositoryImpl(InMemorySecureStore(), demoMode = true)
        assertTrue(repo.sendSmsCode("13800138000", scene = 4) is OpsResult.Ok)
        assertTrue(repo.forgetPassword("13800138000", "1234", "password1") is OpsResult.Ok)
    }
}
