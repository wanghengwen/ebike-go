package com.luopingtech.ebike.ops.platform

import com.luopingtech.ebike.ops.core.result.OpsResult
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking

class BindableCodeScannerTest {
    @Test
    fun unbound_returnsUnsupported() = runBlocking {
        val bridge = BindableCodeScanner()
        val result = bridge.scanOnce()
        assertTrue(result.isErr)
        assertEquals("UNSUPPORTED", (result as OpsResult.Err).error.code)
    }

    @Test
    fun bind_delegates() = runBlocking {
        val bridge = BindableCodeScanner()
        bridge.bind(
            object : CodeScanner {
                override suspend fun scanOnce(): OpsResult<String> = OpsResult.Ok("D1001-001")
            },
        )
        val result = bridge.scanOnce()
        assertTrue(result.isOk)
        assertEquals("D1001-001", result.getOrNull())
    }
}
