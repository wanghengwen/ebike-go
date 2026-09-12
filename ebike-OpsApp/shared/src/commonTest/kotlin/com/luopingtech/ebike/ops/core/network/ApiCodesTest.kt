package com.luopingtech.ebike.ops.core.network

import kotlin.test.Test
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class ApiCodesTest {
    @Test
    fun refreshAndReloginSets() {
        assertTrue(ApiCodes.shouldRefresh("00005"))
        assertTrue(ApiCodes.shouldRefresh("00006"))
        assertFalse(ApiCodes.shouldRefresh("00013"))
        assertTrue(ApiCodes.shouldForceRelogin("00013"))
        assertTrue(ApiCodes.shouldForceRelogin("00015"))
        assertTrue(ApiCodes.shouldForceRelogin("00012"))
        assertFalse(ApiCodes.shouldForceRelogin("00005"))
    }
}
