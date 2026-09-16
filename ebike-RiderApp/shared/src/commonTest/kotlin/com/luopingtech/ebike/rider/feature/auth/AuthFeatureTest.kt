package com.luopingtech.ebike.rider.feature.auth

import com.luopingtech.ebike.rider.core.config.ApiConfig
import com.luopingtech.ebike.rider.core.config.AuthConfig
import com.luopingtech.ebike.rider.core.config.TenantConfig
import com.luopingtech.ebike.rider.core.logging.NoOpLogger
import com.luopingtech.ebike.rider.data.auth.AuthRepositoryImpl
import com.luopingtech.ebike.rider.platform.DeviceInfo
import com.luopingtech.ebike.rider.platform.InMemorySecureStore
import com.luopingtech.ebike.rider.platform.SecureStore
import com.luopingtech.ebike.rider.RiderApp
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertNull
import kotlin.test.assertTrue
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.runCurrent
import kotlinx.coroutines.test.runTest

class AuthFeatureTest {
    @Test
    fun demoLogin_doesNotNeedHttp_andWritesSession() = runTest {
        val store = InMemorySecureStore()
        val app = riderApp(store)
        assertTrue(app.isDemoMode)
        app.authFeature.setAgreedProtocol(true)
        app.authFeature.loginWithSms("13800138000", "whatever")

        val state = app.authFeature.state.value
        assertFalse(state.needLogin)
        assertEquals(AuthRepositoryImpl.DEMO_ACCESS, state.session?.accessToken)
        assertEquals(AuthRepositoryImpl.DEMO_ACCESS, store.getString(SecureStore.KEY_ACCESS_TOKEN))
        app.close()
    }

    @OptIn(ExperimentalCoroutinesApi::class)
    @Test
    fun sessionInvalid_notifiesNeedLogin() = runTest {
        val repo = AuthRepositoryImpl(InMemorySecureStore(), demoMode = true)
        val feature = AuthFeature(repo)
        feature.setAgreedProtocol(true)
        feature.loginWithSms("13800138000", "1")
        assertTrue(feature.state.value.session != null)

        feature.start(backgroundScope)
        runCurrent()
        repo.markSessionInvalid("00015 kicked")
        runCurrent()

        assertTrue(feature.state.value.needLogin)
        assertNull(feature.state.value.session)
        assertEquals("00015 kicked", feature.state.value.error)
    }

    @Test
    fun restoreSession_readsStore() = runTest {
        val store = InMemorySecureStore()
        val repo = AuthRepositoryImpl(store, demoMode = true)
        repo.loginWithSms("13800138000", "1")
        val feature = AuthFeature(AuthRepositoryImpl(store, demoMode = true))
        assertTrue(feature.state.value.session != null)
        feature.restoreSession()
        assertFalse(feature.state.value.needLogin)
        assertTrue(feature.state.value.session?.accessToken == AuthRepositoryImpl.DEMO_ACCESS)
    }

    private fun riderApp(store: InMemorySecureStore) = RiderApp.create(
        config = TenantConfig(
            alias = "test",
            tenantId = "1001",
            api = ApiConfig(baseUrl = ""),
            auth = AuthConfig(businessSecret = "biz", signSecret = "s3cret"),
        ),
        secureStore = store,
        logger = NoOpLogger,
        appVersion = "0.0.0",
        deviceInfo = object : DeviceInfo {
            override val platform = "test"
            override val osVersion = "0"
            override val deviceModel = "unit test"
            override val appVersion = "0.0.0"
        },
        systemLanguage = null,
        quickLogin = com.luopingtech.ebike.rider.platform.UnsupportedQuickLogin,
        codeScanner = com.luopingtech.ebike.rider.platform.UnsupportedCodeScanner(),
        photoCapture = com.luopingtech.ebike.rider.platform.UnsupportedPhotoCapture(),
        locationTracker = null,
        reverseGeocoder = null,
        mediaUploader = null,
        bleTransport = null,
        nativeBle = null,
    )
}
