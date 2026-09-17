package com.luopingtech.ebike.ops.domain.order

import com.luopingtech.ebike.ops.core.i18n.OpsI18n
import com.luopingtech.ebike.ops.core.i18n.OpsLanguage
import com.luopingtech.ebike.ops.core.i18n.Strings
import kotlin.test.BeforeTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class OrderSearchClassifierTest {
    @Test
    fun classify_priority() {
        assertEquals(OrderSearchKind.Phone, OrderSearchClassifier.classify("13800138000"))
        assertEquals(OrderSearchKind.Imei, OrderSearchClassifier.classify("860123456789012"))
        assertEquals(OrderSearchKind.CarId, OrderSearchClassifier.classify("123456789"))
        assertEquals(OrderSearchKind.CarId, OrderSearchClassifier.classify("0812345"))
        assertEquals(OrderSearchKind.Name, OrderSearchClassifier.classify("张三"))
    }

    @Test
    fun carId_08_prefix() {
        assertTrue(OrderSearchClassifier.isCarId("0812345"))
        assertTrue(OrderSearchClassifier.isCarId("08123456"))
        assertFalse(OrderSearchClassifier.isCarId("081234"))
    }
}

class OrderFormatTest {
    @BeforeTest
    fun installZh() {
        Strings.install(OpsI18n.fallback(OpsLanguage.ZH_CN))
    }

    @Test
    fun yuan_and_duration() {
        assertEquals("3.50", OrderFormat.yuan(350))
        assertEquals("0.00", OrderFormat.yuan(0))
        assertEquals("3.50 元", OrderFormat.yuanWithUnit(350))
        // 对齐 TimeStampUtils.timestampFormat：毫秒。1200000ms = 20 分
        assertEquals("20分", OrderFormat.duration("1200000"))
        assertEquals("25分", OrderFormat.duration("1500000"))
        assertEquals("1秒", OrderFormat.duration("1500"))
        assertEquals("2.30公里", OrderFormat.distance(2300))
        assertEquals("230米", OrderFormat.distance(230))
    }

    @Test
    fun labels_follow_language() {
        Strings.install(OpsI18n.fallback(OpsLanguage.EN))
        assertEquals("25m", OrderFormat.duration("1500000"))
        assertEquals("2.30 km", OrderFormat.distance(2300))
        assertEquals("Paid", OrderFormat.payStateLabel(OrderPayStates.Paid))
        assertEquals("3.50 CNY", OrderFormat.yuanWithUnit(350))
    }
}
