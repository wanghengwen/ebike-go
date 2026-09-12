package com.luopingtech.ebike.ops.data.vehicle

import com.luopingtech.ebike.ops.domain.model.ServiceArea
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking

class VehicleRepositoryTest {
    @Test
    fun demoMode_returnsPinsAroundArea() = runBlocking {
        val repo = VehicleRepositoryImpl(demoMode = true)
        val area = ServiceArea(id = "1001", name = "Demo", centerLat = 28.22, centerLng = 112.94)
        val result = repo.loadByServiceArea(area)
        assertTrue(result.isOk)
        val vehicles = result.getOrNull().orEmpty()
        assertEquals(8, vehicles.size)
        assertTrue(vehicles.all { it.serviceId == "1001" })
        assertTrue(vehicles.all { it.lat != 0.0 && it.lng != 0.0 })
        val rich = vehicles.first { it.carId == "D1001-001" }
        assertEquals(54_200, rich.voltageMv)
        assertEquals("BSN-1001-001", rich.batterySn)
    }
}
