package com.luopingtech.ebike.ops.feature.home

import com.luopingtech.ebike.ops.data.auth.AuthRepositoryImpl
import com.luopingtech.ebike.ops.data.tenant.ServiceAreaRepositoryImpl
import com.luopingtech.ebike.ops.data.vehicle.VehicleRepositoryImpl
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.model.UserSession
import com.luopingtech.ebike.ops.feature.auth.AuthFeature
import com.luopingtech.ebike.ops.feature.tenant.ServiceAreaFeature
import com.luopingtech.ebike.ops.feature.vehicle.VehicleFeature
import com.luopingtech.ebike.ops.platform.SecureStore
import com.luopingtech.ebike.ops.platform.InMemorySecureStore
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNull

class HomeFeatureTest {
    @Test
    fun resolveSignedInHomeGate_skipsPickerWhenAreaPersisted() {
        val saved = ServiceArea(id = "1001", name = "东区")
        assertEquals(
            SignedInHomeGate.Home,
            resolveSignedInHomeGate(
                pickingArea = false,
                currentArea = saved,
                persistedAreaId = null,
                areasLoaded = false,
            ),
        )
        assertEquals(
            SignedInHomeGate.Home,
            resolveSignedInHomeGate(
                pickingArea = false,
                currentArea = null,
                persistedAreaId = "1001",
                areasLoaded = false,
            ),
        )
        assertEquals(
            SignedInHomeGate.AreaPicker,
            resolveSignedInHomeGate(
                pickingArea = true,
                currentArea = saved,
                persistedAreaId = "1001",
                areasLoaded = true,
            ),
        )
        assertEquals(
            SignedInHomeGate.LoadingAreas,
            resolveSignedInHomeGate(
                pickingArea = false,
                currentArea = null,
                persistedAreaId = null,
                areasLoaded = false,
            ),
        )
        assertEquals(
            SignedInHomeGate.AreaPicker,
            resolveSignedInHomeGate(
                pickingArea = false,
                currentArea = null,
                persistedAreaId = null,
                areasLoaded = true,
            ),
        )
    }

    @Test
    fun resolveCurrentArea_prefersRepositoryThenSession() {
        val selected = ServiceArea(id = "1002", name = "西区")
        val session = UserSession(serviceAreaId = "1001", serviceAreaName = "东区")
        assertEquals("1002", resolveCurrentArea(selected, session)?.id)
        assertEquals("1001", resolveCurrentArea(null, session)?.id)
        assertEquals("东区", resolveCurrentArea(null, session)?.name)
        assertNull(resolveCurrentArea(null, UserSession()))
    }

    @Test
    fun initialState_usesRestoredAreaWithoutWaitingForStart() {
        val store = InMemorySecureStore()
        store.putString(SecureStore.KEY_ACCESS_TOKEN, "tok")
        store.putString(SecureStore.KEY_SERVICE_AREA_ID, "1002")
        store.putString(SecureStore.KEY_SERVICE_AREA_NAME, "Demo 城西服务区")
        val home = HomeFeature(
            authFeature = AuthFeature(AuthRepositoryImpl(store, demoMode = true)),
            serviceAreaFeature = ServiceAreaFeature(
                ServiceAreaRepositoryImpl(secureStore = store, demoMode = true),
            ),
            vehicleFeature = VehicleFeature(VehicleRepositoryImpl(demoMode = true)),
            mapReady = false,
        )
        assertEquals("1002", home.state.value.currentArea?.id)
        assertEquals(
            SignedInHomeGate.Home,
            resolveSignedInHomeGate(
                pickingArea = false,
                currentArea = home.state.value.currentArea,
                persistedAreaId = home.state.value.session?.serviceAreaId,
                areasLoaded = home.state.value.areasLoaded,
            ),
        )
    }
}
