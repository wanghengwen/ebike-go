package com.luopingtech.ebike.ops.domain.warehouse

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class WarehouseComponentFilterTest {
    @Test
    fun apiAll_mapsToLocalizedDisplay() {
        val display = WarehouseComponentFilter.toDisplayList(
            listOf(WarehouseComponentFilter.API_ALL, "电池"),
        )
        assertEquals(Strings.t(Str.FilterAll), display.first())
        assertEquals("电池", display[1])
    }

    @Test
    fun localizedAll_omittedFromApiFilter() {
        assertTrue(WarehouseComponentFilter.isAll(Strings.t(Str.FilterAll)))
        assertTrue(WarehouseComponentFilter.isAll(WarehouseComponentFilter.API_ALL))
        assertEquals("", WarehouseComponentFilter.toApiFilter(Strings.t(Str.FilterAll)))
        assertEquals("电池", WarehouseComponentFilter.toApiFilter("电池"))
        assertFalse(WarehouseComponentFilter.isAll("电池"))
    }
}
