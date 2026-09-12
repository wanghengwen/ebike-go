package com.luopingtech.ebike.ops.feature.home

import com.luopingtech.ebike.ops.domain.model.MapPin
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.model.UserSession
import com.luopingtech.ebike.ops.domain.model.Vehicle
import com.luopingtech.ebike.ops.feature.auth.AuthFeature
import com.luopingtech.ebike.ops.feature.tenant.ServiceAreaFeature
import com.luopingtech.ebike.ops.feature.vehicle.VehicleFeature
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
    val vehicles: List<Vehicle> = emptyList(),
    val mapPins: List<MapPin> = emptyList(),
    val selectedCarId: String? = null,
    val loadingVehicles: Boolean = false,
    val errorMessage: String? = null,
    val mapReady: Boolean = false,
    val mapProviderKind: String = "none",
)

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
        HomeUiState(
            mapReady = mapReady,
            mapProviderKind = mapProviderKind.name.lowercase(),
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
                HomeUiState(
                    session = auth.session,
                    serviceAreas = areas.areas,
                    currentArea = areas.selected,
                    loadingAreas = areas.loading,
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
        val area = serviceAreaFeature.state.value.selected ?: return
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
}
