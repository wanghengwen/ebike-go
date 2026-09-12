package com.luopingtech.ebike.ops.feature.tools

import com.luopingtech.ebike.ops.core.i18n.OpsI18n
import com.luopingtech.ebike.ops.core.i18n.OpsLanguage
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.data.control.DemoNetworkVehicleControl
import com.luopingtech.ebike.ops.data.tools.UnlockedVehicleRepositoryImpl
import com.luopingtech.ebike.ops.domain.control.VehicleControlPolicy
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.platform.SimulatorBleTransport
import kotlin.test.BeforeTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking

class UnlockedVehicleFeatureTest {
    private val area = ServiceArea(id = "1001", name = "Demo")

    @BeforeTest
    fun installStrings() {
        Strings.install(OpsI18n.fallback(OpsLanguage.ZH_CN))
    }

    @Test
    fun withoutStaffPermission_forcesSelfPhoneQuery() = runBlocking {
        val feature = UnlockedVehicleFeature(
            repository = UnlockedVehicleRepositoryImpl(demoMode = true),
            control = VehicleControlPolicy(
                ble = SimulatorBleTransport(),
                network = DemoNetworkVehicleControl(),
            ),
            canFilterStaffProvider = { false },
            selfPhoneProvider = { "13900001111" },
        )
        feature.setQuery("should-ignore")
        feature.load(area)
        assertEquals(false, feature.state.value.canFilterStaff)
        assertEquals("13900001111", feature.state.value.query)
    }

    @Test
    fun loadAndNetworkLockRemovesRow() = runBlocking {
        val feature = UnlockedVehicleFeature(
            repository = UnlockedVehicleRepositoryImpl(demoMode = true),
            control = VehicleControlPolicy(
                ble = SimulatorBleTransport(),
                network = DemoNetworkVehicleControl(),
            ),
        )
        feature.load(area)
        assertTrue(feature.state.value.vehicles.isNotEmpty())
        val carId = feature.state.value.vehicles.first().carId
        assertTrue(feature.lock(carId).isOk)
        assertTrue(feature.state.value.vehicles.none { it.carId == carId })

        feature.load(area)
        assertTrue(feature.state.value.vehicles.none { it.carId == carId })
        assertEquals(true, feature.state.value.message?.isNotBlank())
    }
}
