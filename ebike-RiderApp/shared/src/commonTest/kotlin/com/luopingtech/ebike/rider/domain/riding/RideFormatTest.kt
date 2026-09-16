package com.luopingtech.ebike.rider.domain.riding

import kotlin.test.Test
import kotlin.test.assertEquals

/**
 * 手写的格式化容易在边界上出错（Kotlin/Native 没有 `String.format`），
 * 而这些数字就是用户看到的钱和时间，所以逐个边界钉住。
 */
class RideFormatTest {

    @Test
    fun `fen to yuan always keeps two decimals`() {
        assertEquals("0.00", RideFormat.yuan(0))
        assertEquals("0.05", RideFormat.yuan(5))
        assertEquals("0.50", RideFormat.yuan(50))
        assertEquals("0.99", RideFormat.yuan(99))
        assertEquals("1.00", RideFormat.yuan(100))
        assertEquals("1.01", RideFormat.yuan(101))
        assertEquals("2.50", RideFormat.yuan(250))
        assertEquals("10.09", RideFormat.yuan(1009))
        assertEquals("123.45", RideFormat.yuan(12345))
    }

    @Test
    fun `negative amounts keep a single leading sign`() {
        assertEquals("-1.50", RideFormat.yuan(-150))
        assertEquals("-0.09", RideFormat.yuan(-9))
        assertEquals("-12.34", RideFormat.yuan(-1234))
    }

    @Test
    fun `duration is mm ss below an hour`() {
        assertEquals("00:00", RideFormat.duration(0))
        assertEquals("00:09", RideFormat.duration(9))
        assertEquals("01:00", RideFormat.duration(60))
        assertEquals("09:59", RideFormat.duration(599))
        assertEquals("59:59", RideFormat.duration(3_599))
    }

    @Test
    fun `duration grows an hours segment past an hour`() {
        // 骑行时长真的会破小时，只留两段会显示成 00:00。
        assertEquals("1:00:00", RideFormat.duration(3_600))
        assertEquals("1:00:01", RideFormat.duration(3_601))
        assertEquals("2:03:04", RideFormat.duration(7_384))
        assertEquals("10:00:00", RideFormat.duration(36_000))
    }

    @Test
    fun `a negative duration is clamped to zero`() {
        // 服务端时长偶尔会因为时钟漂移回负数。
        assertEquals("00:00", RideFormat.duration(-5))
    }

    @Test
    fun `distance stays in metres below a kilometre`() {
        assertEquals("0 m", RideFormat.distance(0))
        assertEquals("1 m", RideFormat.distance(1))
        assertEquals("999 m", RideFormat.distance(999))
    }

    @Test
    fun `distance switches to kilometres with one decimal`() {
        assertEquals("1.0 km", RideFormat.distance(1_000))
        assertEquals("1.0 km", RideFormat.distance(1_040))
        assertEquals("1.1 km", RideFormat.distance(1_050))
        assertEquals("2.5 km", RideFormat.distance(2_499))
        assertEquals("12.3 km", RideFormat.distance(12_345))
    }

    @Test
    fun `a negative distance is clamped to zero`() {
        assertEquals("0 m", RideFormat.distance(-10))
    }
}
