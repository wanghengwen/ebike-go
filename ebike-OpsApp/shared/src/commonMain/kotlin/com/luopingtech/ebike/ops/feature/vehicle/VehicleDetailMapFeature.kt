package com.luopingtech.ebike.ops.feature.vehicle

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.core.time.nowEpochMillis
import com.luopingtech.ebike.ops.data.fence.FenceRepository
import com.luopingtech.ebike.ops.data.geo.GeoApi
import com.luopingtech.ebike.ops.data.order.OrderRepository
import com.luopingtech.ebike.ops.data.trajectory.TrajectoryRepository
import com.luopingtech.ebike.ops.domain.model.FenceBundle
import com.luopingtech.ebike.ops.domain.model.LastOrder
import com.luopingtech.ebike.ops.domain.model.TrackPoint
import com.luopingtech.ebike.ops.domain.model.Vehicle
import com.luopingtech.ebike.ops.platform.ReverseGeocoder
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class VehicleDetailMapUiState(
    val loadingFence: Boolean = false,
    val loadingTrack: Boolean = false,
    val loadingAddress: Boolean = false,
    val fence: FenceBundle? = null,
    val lastOrder: LastOrder? = null,
    val track: List<TrackPoint> = emptyList(),
    val address: String? = null,
    val showFence: Boolean = true,
    val showTrack: Boolean = false,
    val errorMessage: String? = null,
)

/**
 * Detail extras aligned with legacy CarDetail:
 * - address via gaode/getAddress (fallback ReverseGeocoder)
 * - fence via getNearFenceByLocations when last-order points exist, else full service fence
 * - map track toggle = last-order deviceTrajectory; realtime = trajectory/realTime
 */
class VehicleDetailMapFeature(
    private val fenceRepository: FenceRepository,
    private val orderRepository: OrderRepository,
    private val trajectoryRepository: TrajectoryRepository,
    private val reverseGeocoder: ReverseGeocoder,
    private val geoApi: GeoApi? = null,
    private val demoMode: Boolean = false,
) {
    private val _state = MutableStateFlow(VehicleDetailMapUiState())
    val state: StateFlow<VehicleDetailMapUiState> = _state.asStateFlow()

    fun clear() {
        _state.value = VehicleDetailMapUiState()
    }

    fun setShowFence(show: Boolean) {
        _state.value = _state.value.copy(showFence = show)
    }

    fun setShowTrack(show: Boolean) {
        _state.value = _state.value.copy(showTrack = show)
    }

    suspend fun loadForVehicle(vehicle: Vehicle, serviceAreaId: String?) {
        val sid = serviceAreaId?.ifBlank { null }
            ?: vehicle.serviceId.ifBlank { null }
        reverseGeocode(vehicle.lat, vehicle.lng)
        val order = loadLastOrder(vehicle.carId)
        if (sid != null) {
            val near = order?.nearFenceLocations().orEmpty()
            if (near.isNotEmpty()) {
                loadNearFence(sid, near)
            } else {
                loadFence(sid)
            }
        }
    }

    suspend fun loadLastOrder(carId: String): LastOrder? {
        if (carId.isBlank()) return null
        return when (val result = orderRepository.detailLast(carId)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(lastOrder = result.value)
                result.value
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(lastOrder = null)
                null
            }
        }
    }

    suspend fun loadFence(serviceId: String) {
        _state.value = _state.value.copy(loadingFence = true, errorMessage = null)
        when (val result = fenceRepository.loadByServiceId(serviceId)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loadingFence = false,
                    fence = result.value,
                    showFence = true,
                )
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loadingFence = false,
                    errorMessage = result.error.message,
                )
            }
        }
    }

    suspend fun loadNearFence(
        serviceId: String,
        locations: List<com.luopingtech.ebike.ops.domain.model.GeoLatLng>,
    ) {
        _state.value = _state.value.copy(loadingFence = true, errorMessage = null)
        when (val result = fenceRepository.loadNearLocations(serviceId, locations)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loadingFence = false,
                    fence = result.value,
                    showFence = true,
                )
            }
            is OpsResult.Err -> {
                // Fall back to full service fence on near-API failure.
                loadFence(serviceId)
            }
        }
    }

    /** Legacy CarDetail track toggle: last-order deviceTrajectory. */
    suspend fun loadTrack(vehicle: Vehicle) {
        _state.value = _state.value.copy(loadingTrack = true, errorMessage = null)
        val cached = _state.value.lastOrder
        val order = cached?.takeIf { it.carId == vehicle.carId }
            ?: loadLastOrder(vehicle.carId)
        val track = order?.trajectory.orEmpty()
        if (track.isNotEmpty()) {
            _state.value = _state.value.copy(
                loadingTrack = false,
                track = track,
                showTrack = true,
                errorMessage = null,
            )
            refreshNearFenceFromOrder(vehicle, order)
            return
        }
        _state.value = _state.value.copy(
            loadingTrack = false,
            track = emptyList(),
            showTrack = false,
            errorMessage = Strings.t(Str.VehicleTrackEmpty),
        )
    }

    /** Legacy openTrajectory: IMEI time-window realtime path. */
    suspend fun loadRealtimeTrack(vehicle: Vehicle, hoursBack: Int = 6) {
        val imei = vehicle.imei.trim()
        if (imei.isBlank()) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.DetectOverloadNeedImei))
            return
        }
        val end = nowEpochMillis()
        val start = end - hoursBack * 3_600_000L
        _state.value = _state.value.copy(loadingTrack = true, errorMessage = null)
        when (val result = trajectoryRepository.loadRealTime(imei, start, end)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loadingTrack = false,
                    track = result.value,
                    showTrack = true,
                    errorMessage = if (result.value.isEmpty()) {
                        Strings.t(Str.VehicleTrackEmpty)
                    } else {
                        null
                    },
                )
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loadingTrack = false,
                    errorMessage = result.error.message,
                )
            }
        }
    }

    suspend fun reverseGeocode(lat: Double, lng: Double) {
        if (lat == 0.0 && lng == 0.0) return
        _state.value = _state.value.copy(loadingAddress = true)
        val server = when {
            demoMode -> null
            geoApi != null -> geoApi.getAddress(lat, lng)
            else -> null
        }
        if (server is OpsResult.Ok) {
            _state.value = _state.value.copy(loadingAddress = false, address = server.value)
            return
        }
        when (val result = reverseGeocoder.addressOf(lat, lng)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loadingAddress = false,
                    address = result.value,
                )
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loadingAddress = false,
                    address = null,
                )
            }
        }
    }

    private suspend fun refreshNearFenceFromOrder(vehicle: Vehicle, order: LastOrder?) {
        val sid = vehicle.serviceId.ifBlank { null } ?: return
        val near = order?.nearFenceLocations().orEmpty()
        if (near.isNotEmpty()) {
            loadNearFence(sid, near)
        }
    }
}
