package com.luopingtech.ebike.rider.domain.scan

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
        val target = ScanCodeParser.parse("860000000000001")
        assertIs<ScanTarget.CarId>(target)
        assertEquals("860000000000001", target.value)
    }

    @Test
    fun queryCarId() {
        val target = ScanCodeParser.parse(
            "https://rider.example/q?carId=ABC123&x=1",
            allowedHosts = listOf("rider.example"),
        )
        assertIs<ScanTarget.CarId>(target)
        assertEquals("ABC123", target.value)
    }

    @Test
    fun queryImei() {
        val target = ScanCodeParser.parse(
            "https://rider.example/q?imei=860000000000001",
            allowedHosts = listOf("rider.example"),
        )
        assertIs<ScanTarget.Imei>(target)
        assertEquals("860000000000001", target.value)
    }

    @Test
    fun pathTail() {
        val target = ScanCodeParser.parse(
            "https://rider.example/bike/XY-99",
            allowedHosts = listOf("rider.example"),
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
                allowedHosts = listOf("rider.example"),
            ),
        )
    }

    @Test
    fun emptyAllowlistRejectsHttpUrl() {
        assertNull(ScanCodeParser.parse("https://rider.example/q?carId=ABC"))
    }
}
