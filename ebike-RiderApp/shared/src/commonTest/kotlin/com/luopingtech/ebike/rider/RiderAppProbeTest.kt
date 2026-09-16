package com.luopingtech.ebike.rider

import com.luopingtech.ebike.rider.core.config.ApiConfig
import com.luopingtech.ebike.rider.core.config.AuthConfig
import com.luopingtech.ebike.rider.core.config.TenantConfig
import com.luopingtech.ebike.rider.core.logging.NoOpLogger
import com.luopingtech.ebike.rider.core.signing.RequestSigner
import com.luopingtech.ebike.rider.platform.DeviceInfo
import com.luopingtech.ebike.rider.platform.InMemorySecureStore
import com.luopingtech.ebike.rider.platform.SecureStore
import kotlinx.coroutines.test.runTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNull
import kotlin.test.assertTrue

/**
 * 组合根的装配契约。P0 没有业务 feature，这几条就是「空壳装对了」的全部定义。
 */
class RiderAppProbeTest {
    @Test
    fun demoProbe_signsBodyWithoutSendingRequest() = runTest {
        val app = riderApp(baseUrl = "")
        val probe = app.probe()

        assertTrue(probe.isDemo)
        // 没有 baseUrl 就不该有请求发出去 —— 空壳最容易犯的错是照着相对路径撞出一个异常。
        assertNull(probe.httpStatus)
        assertEquals(
            RequestSigner.signPostJson(probe.requestBody, probe.timestamp, SIGN_SECRET),
            probe.sign,
        )
        app.close()
    }

    /** 公共请求体的四个字段必须都填上，缺 tenantId / deviceId 网关会直接拒。 */
    @Test
    fun demoProbe_bodyCarriesCommonFields() = runTest {
        val app = riderApp(baseUrl = "")
        val body = app.probe().requestBody

        assertTrue(body.contains("\"traceId\":\""), body)
        assertTrue(body.contains("\"tenantId\":\"$TENANT_ID\""), body)
        assertTrue(body.contains("\"deviceId\":\"${app.deviceId}\""), body)
        assertTrue(body.contains("\"platform\":\"test\""), body)
        app.close()
    }

    /** 设备标识跨实例稳定：每次冷启动换一个，后端看到的就是一台新设备。 */
    @Test
    fun deviceId_isPersistedAndReusedAcrossInstances() {
        val store = InMemorySecureStore()
        val first = riderApp(baseUrl = "", store = store).deviceId
        val second = riderApp(baseUrl = "", store = store).deviceId

        assertTrue(first.isNotBlank())
        assertTrue(first.all { it.isDigit() }, first)
        assertEquals(first, second)
        assertEquals(first, store.getString(SecureStore.KEY_DEVICE_ID))
    }

    @Test
    fun create_persistsTenantIdForLaterBasicAuth() {
        val store = InMemorySecureStore()
        riderApp(baseUrl = "", store = store)
        assertEquals(TENANT_ID, store.getString(SecureStore.KEY_TENANT_ID))
    }

    @Test
    fun isDemoMode_followsBaseUrl() {
        assertTrue(riderApp(baseUrl = "").isDemoMode)
        assertTrue(!riderApp(baseUrl = "https://example.invalid/api").isDemoMode)
    }

    private fun riderApp(
        baseUrl: String,
        store: InMemorySecureStore = InMemorySecureStore(),
    ) = RiderApp.create(
        config = TenantConfig(
            alias = "test",
            tenantId = TENANT_ID,
            api = ApiConfig(baseUrl = baseUrl),
            auth = AuthConfig(businessSecret = "biz", signSecret = SIGN_SECRET),
        ),
        secureStore = store,
        logger = NoOpLogger,
        appVersion = "0.0.0",
        // 不用 createDeviceInfo()：Android 那份读 android.os.Build，
        // JVM 单测里那些字段全是 null。设备信息本来就该由宿主注入。
        deviceInfo = FakeDeviceInfo,
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

    private object FakeDeviceInfo : DeviceInfo {
        override val platform = "test"
        override val osVersion = "0"
        override val deviceModel = "unit test"
        override val appVersion = "0.0.0"
    }

    private companion object {
        const val TENANT_ID = "1001"
        const val SIGN_SECRET = "s3cret"
    }
}
