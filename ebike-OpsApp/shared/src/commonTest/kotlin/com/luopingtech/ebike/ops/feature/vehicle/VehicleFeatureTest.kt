package com.luopingtech.ebike.ops.feature.vehicle

import com.luopingtech.ebike.ops.core.i18n.OpsI18n
import com.luopingtech.ebike.ops.core.i18n.OpsLanguage
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.data.vehicle.VehicleRepositoryImpl
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import kotlin.test.BeforeTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking

class VehicleFeatureTest {
    @BeforeTest
    fun installStrings() {
        Strings.install(OpsI18n.fallback(OpsLanguage.ZH_CN))
    }

    @Test
    fun refreshDetail_mergesIntoList() = runBlocking {
        val feature = VehicleFeature(VehicleRepositoryImpl(demoMode = true))
        val area = ServiceArea(id = "1001", name = "Demo", centerLat = 28.22, centerLng = 112.94)
        feature.loadForArea(area)
        assertEquals(8, feature.state.value.vehicles.size)
        feature.selectVehicle("D1001-003")
        assertTrue(feature.refreshDetail("D1001-003").isOk)
        val selected = feature.state.value.vehicles.first { it.carId == "D1001-003" }
        assertEquals(true, selected.isOutOfServiceArea)
        assertEquals(42_800, selected.voltageMv)
        assertEquals("D1001-003", feature.state.value.selectedCarId)
    }

    @Test
    fun upsertAndSelect_addsMissingVehicle() {
        val feature = VehicleFeature(VehicleRepositoryImpl(demoMode = true))
        val orphan = runBlocking {
            VehicleRepositoryImpl(demoMode = true).getDetail("SCAN-ONLY").getOrNull()!!
        }
        feature.upsertAndSelect(orphan)
        assertTrue(feature.state.value.vehicles.any { it.carId == orphan.carId })
        assertEquals(orphan.carId, feature.state.value.selectedCarId)
    }
}
