package com.luopingtech.ebike.ops.feature.sneak

import com.luopingtech.ebike.ops.data.order.OrderRepositoryImpl
import com.luopingtech.ebike.ops.data.sneak.SneakReportRepositoryImpl
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking

class SneakReportFeatureTest {
    private fun feature(
        canPhone: String = "13900001111",
    ) = SneakReportFeature(
        repository = SneakReportRepositoryImpl(demoMode = true),
        orderRepository = OrderRepositoryImpl(demoMode = true),
        serviceAreaIdProvider = { "1001" },
        reportManPinProvider = { "demo-pin" },
        reportManPhoneProvider = { canPhone },
    )

    @Test
    fun submitRequiresTypeAndLastOrder() = runBlocking {
        val f = feature()
        f.openSubmit()
        f.ensureTypes()
        assertTrue(f.state.value.types.isNotEmpty())
        f.setCarId("D1001-001")
        f.lookupLastOrderIfReady()
        assertTrue(f.state.value.lastOrder != null)
        f.submit()
        assertTrue(f.state.value.errorMessage != null)

        f.toggleType(f.state.value.types.first().id)
        f.setDescription("blocked parking")
        f.submit()
        assertEquals(SneakPage.Hub, f.state.value.page)
        assertTrue(f.state.value.message?.isNotBlank() == true)
    }

    @Test
    fun historyAndCancelPending() = runBlocking {
        val f = feature()
        f.openHistory()
        assertTrue(f.state.value.records.size >= 3)
        val pending = f.state.value.records.first { it.checkResult == 0 }
        f.openDetail(pending)
        f.cancelSelected()
        f.openHistory()
        assertTrue(f.state.value.records.any { it.id == pending.id && it.checkResult == 5 })
    }

    @Test
    fun otherTypeRequiresText() = runBlocking {
        val f = feature()
        f.openSubmit()
        f.ensureTypes()
        f.setCarId("D1001-009")
        f.lookupLastOrderIfReady()
        val other = f.state.value.types.first { it.isOther }
        f.toggleType(other.id)
        f.submit()
        assertFalse(f.state.value.page == SneakPage.Hub)
        assertTrue(f.state.value.errorMessage != null)
        f.setOtherType("custom")
        f.submit()
        assertEquals(SneakPage.Hub, f.state.value.page)
    }
}
