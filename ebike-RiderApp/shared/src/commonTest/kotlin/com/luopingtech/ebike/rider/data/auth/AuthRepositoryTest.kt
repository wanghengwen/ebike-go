package com.luopingtech.ebike.rider.data.auth

import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.platform.InMemorySecureStore
import com.luopingtech.ebike.rider.platform.SecureStore
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertNull
import kotlin.test.assertTrue
import kotlinx.coroutines.test.runTest

class AuthRepositoryTest {
    @Test
    fun demoLogin_writesTokensWithoutCallingRemote() = runTest {
        val store = InMemorySecureStore()
        val remote = CountingAuthRemote()
        val repo = AuthRepositoryImpl(
            secureStore = store,
            demoMode = true,
            authApi = remote,
            defaultTenantId = "1001",
        )

        val result = repo.loginWithSms("13800138000", "0000")
        assertTrue(result is RiderResult.Ok)
        assertEquals(0, remote.calls)
        assertEquals(AuthRepositoryImpl.DEMO_ACCESS, store.getString(SecureStore.KEY_ACCESS_TOKEN))
        assertEquals(AuthRepositoryImpl.DEMO_REFRESH, store.getString(SecureStore.KEY_REFRESH_TOKEN))
        assertEquals("+86-13800138000", repo.currentSession()?.phone)
    }

    @Test
    fun realLogin_writesTokensFromOauth() = runTest {
        val store = InMemorySecureStore()
        val auth = FakeAuthRemote(
            token = OauthTokenDto(
                accessToken = "access-1",
                refreshToken = "refresh-1",
                phone = "+86-13800138000",
                pin = "pin-1",
                nickname = "Ada",
            ),
        )
        val user = FakeUserRemote(
            person = PersonInfoDto(pin = "pin-1", nickname = "Ada", phone = "+86-13800138000", balance = 42),
        )
        val repo = AuthRepositoryImpl(
            secureStore = store,
            demoMode = false,
            authApi = auth,
            userApiProvider = { user },
            defaultTenantId = "1001",
        )

        val result = repo.loginWithSms("13800138000", "123456")
        assertTrue(result is RiderResult.Ok)
        assertEquals(1, auth.loginCalls)
        assertEquals(0, auth.refreshCalls)
        assertEquals("access-1", store.getString(SecureStore.KEY_ACCESS_TOKEN))
        assertEquals("refresh-1", store.getString(SecureStore.KEY_REFRESH_TOKEN))
        assertEquals("Ada", repo.currentSession()?.displayName)
        assertEquals(42L, repo.currentSession()?.balance)
    }

    @Test
    fun refresh_updatesStoredTokens() = runTest {
        val store = InMemorySecureStore()
        val auth = FakeAuthRemote(
            token = OauthTokenDto(accessToken = "old", refreshToken = "r-old"),
            refresh = OauthTokenDto(accessToken = "new-access", refreshToken = "new-refresh"),
        )
        val repo = AuthRepositoryImpl(
            secureStore = store,
            demoMode = false,
            authApi = auth,
            defaultTenantId = "1001",
        )
        assertTrue(repo.loginWithSms("13800138000", "1") is RiderResult.Ok)

        val refreshed = repo.refreshAccessToken()
        assertTrue(refreshed is RiderResult.Ok)
        assertEquals(1, auth.refreshCalls)
        assertEquals("new-access", store.getString(SecureStore.KEY_ACCESS_TOKEN))
        assertEquals("new-refresh", store.getString(SecureStore.KEY_REFRESH_TOKEN))
        assertEquals("new-access", repo.currentSession()?.accessToken)
    }

    @Test
    fun badRefresh_clearsSession() = runTest {
        val store = InMemorySecureStore()
        val auth = FakeAuthRemote(
            token = OauthTokenDto(accessToken = "old", refreshToken = "r-old"),
            refreshError = RiderError.unauthorized("bad refresh"),
        )
        val repo = AuthRepositoryImpl(
            secureStore = store,
            demoMode = false,
            authApi = auth,
        )
        assertTrue(repo.loginWithSms("13800138000", "1") is RiderResult.Ok)

        val refreshed = repo.refreshAccessToken()
        assertTrue(refreshed is RiderResult.Err)
        assertNull(repo.currentSession())
        assertNull(store.getString(SecureStore.KEY_ACCESS_TOKEN))
        assertNull(store.getString(SecureStore.KEY_REFRESH_TOKEN))
        assertEquals("bad refresh", repo.sessionInvalidMessage.value)
    }

    @Test
    fun code00015_clearsSession() = runTest {
        val store = InMemorySecureStore()
        val repo = AuthRepositoryImpl(secureStore = store, demoMode = true)
        assertTrue(repo.loginWithSms("13800138000", "1") is RiderResult.Ok)
        assertNotNull(store.getString(SecureStore.KEY_ACCESS_TOKEN))

        repo.markSessionInvalid("00015 other device")
        assertNull(repo.currentSession())
        assertNull(store.getString(SecureStore.KEY_ACCESS_TOKEN))
        assertEquals("00015 other device", repo.sessionInvalidMessage.value)
    }

    @Test
    fun logout_clearsTokensButKeepsDeviceAndTenant() = runTest {
        val store = InMemorySecureStore()
        store.putString(SecureStore.KEY_DEVICE_ID, "device-keep")
        store.putString(SecureStore.KEY_TENANT_ID, "1001")
        val repo = AuthRepositoryImpl(secureStore = store, demoMode = true, defaultTenantId = "1001")
        assertTrue(repo.loginWithSms("13800138000", "1") is RiderResult.Ok)

        repo.logout()
        assertNull(repo.currentSession())
        assertNull(store.getString(SecureStore.KEY_ACCESS_TOKEN))
        assertEquals("device-keep", store.getString(SecureStore.KEY_DEVICE_ID))
        assertEquals("1001", store.getString(SecureStore.KEY_TENANT_ID))
    }
}

private class CountingAuthRemote : AuthRemote {
    var calls: Int = 0
        private set

    override suspend fun sendSmsCode(phone: String, scene: Int): RiderResult<Unit> {
        calls++
        error("demo must not call sendSms")
    }

    override suspend fun loginWithSms(phone: String, messageCode: String): RiderResult<OauthTokenDto> {
        calls++
        error("demo must not call login")
    }

    override suspend fun refreshToken(refreshToken: String): RiderResult<OauthTokenDto> {
        calls++
        error("demo must not call refresh")
    }
}

private class FakeAuthRemote(
    private val token: OauthTokenDto = OauthTokenDto(accessToken = "a", refreshToken = "r"),
    private val refresh: OauthTokenDto = token,
    private val refreshError: RiderError? = null,
) : AuthRemote {
    var loginCalls: Int = 0
    var refreshCalls: Int = 0

    override suspend fun sendSmsCode(phone: String, scene: Int): RiderResult<Unit> = RiderResult.Ok(Unit)

    override suspend fun loginWithSms(phone: String, messageCode: String): RiderResult<OauthTokenDto> {
        loginCalls++
        return RiderResult.Ok(token)
    }

    override suspend fun refreshToken(refreshToken: String): RiderResult<OauthTokenDto> {
        refreshCalls++
        refreshError?.let { return RiderResult.Err(it) }
        return RiderResult.Ok(refresh)
    }
}

private class FakeUserRemote(
    private val person: PersonInfoDto = PersonInfoDto(),
    private val account: UserAccountDto = UserAccountDto(),
) : UserRemote {
    override suspend fun fetchPersonInfo(): RiderResult<PersonInfoDto> = RiderResult.Ok(person)
    override suspend fun fetchUserAccount(): RiderResult<UserAccountDto> = RiderResult.Ok(account)
}
