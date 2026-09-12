package com.luopingtech.ebike.ops.feature.home

import com.luopingtech.ebike.ops.domain.model.MapPin
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.model.UserSession
import com.luopingtech.ebike.ops.domain.model.Vehicle
import com.luopingtech.ebike.ops.feature.auth.AuthFeature
import com.luopingtech.ebike.ops.feature.auth.AuthUiState
import com.luopingtech.ebike.ops.feature.tenant.ServiceAreaFeature
import com.luopingtech.ebike.ops.feature.tenant.ServiceAreaUiState
import com.luopingtech.ebike.ops.feature.vehicle.VehicleFeature
import com.luopingtech.ebike.ops.feature.vehicle.VehicleUiState
import com.luopingtech.ebike.ops.platform.MapProviderKind
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.launch

data class HomeUiState(
    val session: UserSession? = null,
    val serviceAreas: List<ServiceArea> = emptyList(),
    val currentArea: ServiceArea? = null,
    val loadingAreas: Boolean = false,
    val areasLoaded: Boolean = false,
    val vehicles: List<Vehicle> = emptyList(),
    val mapPins: List<MapPin> = emptyList(),
    val selectedCarId: String? = null,
    val loadingVehicles: Boolean = false,
    val errorMessage: String? = null,
    val mapReady: Boolean = false,
    val mapProviderKind: String = "none",
)

enum class SignedInHomeGate {
    LoadingAreas,
    AreaPicker,
    Home,
}

/**
 * Cold-start routing after login: never flash the area picker when a persisted
 * service area already exists (repository selection or session.serviceAreaId).
 */
fun resolveSignedInHomeGate(
    pickingArea: Boolean,
    currentArea: ServiceArea?,
    persistedAreaId: String?,
    areasLoaded: Boolean,
): SignedInHomeGate {
    if (pickingArea) return SignedInHomeGate.AreaPicker
    if (currentArea != null || !persistedAreaId.isNullOrBlank()) return SignedInHomeGate.Home
    if (!areasLoaded) return SignedInHomeGate.LoadingAreas
    return SignedInHomeGate.AreaPicker
}

internal fun resolveCurrentArea(
    selected: ServiceArea?,
    session: UserSession?,
): ServiceArea? {
    selected?.takeIf { it.id.isNotBlank() }?.let { return it }
    val id = session?.serviceAreaId?.takeIf { it.isNotBlank() } ?: return null
    return ServiceArea(id = id, name = session.serviceAreaName)
}

/**
 * Home shell: session + service area + vehicle pins.
 * Real map SDK rendering stays in the host UI; [mapPins] is the shared contract.
 */
class HomeFeature(
    private val authFeature: AuthFeature,
    private val serviceAreaFeature: ServiceAreaFeature,
    private val vehicleFeature: VehicleFeature,
    private val mapReady: Boolean,
    private val mapProviderKind: MapProviderKind = MapProviderKind.NONE,
) {
    private val _state = MutableStateFlow(
        composeState(
            authFeature.state.value,
            serviceAreaFeature.state.value,
            vehicleFeature.state.value,
        ),
    )
    val state: StateFlow<HomeUiState> = _state.asStateFlow()

    fun start(scope: CoroutineScope) {
        scope.launch {
            combine(
                authFeature.state,
                serviceAreaFeature.state,
                vehicleFeature.state,
            ) { auth, areas, vehicles ->
                composeState(auth, areas, vehicles)
            }.collect { _state.value = it }
        }
    }

    suspend fun ensureAreasLoaded() {
        if (authFeature.state.value.session == null) return
        if (serviceAreaFeature.state.value.areas.isNotEmpty()) {
            ensureVehiclesLoaded()
            return
        }
        serviceAreaFeature.load()
        ensureVehiclesLoaded()
    }

    suspend fun ensureVehiclesLoaded() {
        val area = serviceAreaFeature.state.value.selected
            ?: resolveCurrentArea(null, authFeature.state.value.session)
            ?: return
        val vehicles = vehicleFeature.state.value
        if (vehicles.serviceAreaId == area.id && vehicles.vehicles.isNotEmpty()) return
        vehicleFeature.loadForArea(area)
    }

    suspend fun selectArea(area: ServiceArea) {
        serviceAreaFeature.select(area)
        vehicleFeature.loadForArea(area)
    }

    fun selectVehicle(carId: String?) {
        vehicleFeature.selectVehicle(carId)
    }

    suspend fun refreshSelectedDetail() {
        val carId = vehicleFeature.state.value.selectedCarId ?: return
        vehicleFeature.refreshDetail(carId)
    }

    suspend fun reloadVehicles() {
        vehicleFeature.loadForArea(serviceAreaFeature.state.value.selected)
    }

    private fun composeState(
        auth: AuthUiState,
        areas: ServiceAreaUiState,
        vehicles: VehicleUiState,
    ): HomeUiState = HomeUiState(
        session = auth.session,
        serviceAreas = areas.areas,
        currentArea = resolveCurrentArea(areas.selected, auth.session),
        loadingAreas = areas.loading,
        areasLoaded = areas.listLoaded,
        vehicles = vehicles.vehicles,
        mapPins = vehicles.pins,
        selectedCarId = vehicles.selectedCarId,
        loadingVehicles = vehicles.loading,
        errorMessage = vehicles.errorMessage
            ?: areas.errorMessage
            ?: auth.errorMessage,
        mapReady = mapReady,
        mapProviderKind = mapProviderKind.name.lowercase(),
    )
}
