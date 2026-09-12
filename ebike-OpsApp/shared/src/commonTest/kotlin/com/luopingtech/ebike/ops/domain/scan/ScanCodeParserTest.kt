package com.luopingtech.ebike.ops.domain.scan

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertIs
import kotlin.test.assertNull

class ScanCodeParserTest {
    @Test
    fun plainCarId() {
        val target = ScanCodeParser.parse("D1001-001")
        assertIs<ScanTarget.CarId>(target)
        assertEquals("D1001-001", target.value)
    }

    @Test
    fun fifteenDigitTreatedAsCarId() {
        // Legacy main scan treats bare tokens as carId, not IMEI.
        val target = ScanCodeParser.parse("860000000000001")
        assertIs<ScanTarget.CarId>(target)
        assertEquals("860000000000001", target.value)
    }

    @Test
    fun queryCarId() {
        val target = ScanCodeParser.parse(
            "https://ops.example/q?carId=ABC123&x=1",
            allowedHosts = listOf("ops.example"),
        )
        assertIs<ScanTarget.CarId>(target)
        assertEquals("ABC123", target.value)
    }

    @Test
    fun queryImei() {
        val target = ScanCodeParser.parse(
            "https://ops.example/q?imei=860000000000001",
            allowedHosts = listOf("ops.example"),
        )
        assertIs<ScanTarget.Imei>(target)
        assertEquals("860000000000001", target.value)
    }

    @Test
    fun pathTail() {
        val target = ScanCodeParser.parse(
            "https://ops.example/bike/XY-99",
            allowedHosts = listOf("ops.example"),
        )
        assertIs<ScanTarget.CarId>(target)
        assertEquals("XY-99", target.value)
    }

    @Test
    fun blankRejected() {
        assertNull(ScanCodeParser.parse("   "))
    }

    @Test
    fun imeiPrefix() {
        val target = ScanCodeParser.parse("IMEI:860000000000001")
        assertIs<ScanTarget.Imei>(target)
        assertEquals("860000000000001", target.value)
    }

    @Test
    fun hostAllowlistRejectsUnknown() {
        assertNull(
            ScanCodeParser.parse(
                "https://evil.example/q?carId=ABC",
                allowedHosts = listOf("ops.example"),
            ),
        )
    }

    @Test
    fun hostAllowlistAcceptsKnown() {
        val target = ScanCodeParser.parse(
            "https://ops.example/q?carId=ABC",
            allowedHosts = listOf("ops.example"),
        )
        assertIs<ScanTarget.CarId>(target)
        assertEquals("ABC", target.value)
    }

    @Test
    fun emptyAllowlistRejectsHttpUrl() {
        assertNull(ScanCodeParser.parse("https://ops.example/q?carId=ABC"))
        assertNull(
            ScanCodeParser.parse(
                "https://ops.example/q?carId=ABC",
                allowedHosts = emptyList(),
            ),
        )
    }
}
