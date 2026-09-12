package com.luopingtech.ebike.ops.feature.analysis

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.core.time.nowEpochMillis
import com.luopingtech.ebike.ops.data.vehicle.VehicleRepository
import com.luopingtech.ebike.ops.data.vehicle.toMapPin
import com.luopingtech.ebike.ops.domain.analysis.VehicleConditionBuckets
import com.luopingtech.ebike.ops.domain.model.MapPin
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.model.Vehicle
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

enum class VcdTab {
    Battery,
    Idle,
}

enum class VcdPage {
    List,
    Map,
}

data class VehicleConditionDistributionUiState(
    val page: VcdPage = VcdPage.List,
    val loading: Boolean = false,
    val tab: VcdTab = VcdTab.Battery,
    val totalCount: Int = 0,
    val batteryBuckets: List<VehicleConditionBuckets.Bucket> = emptyList(),
    val idleBuckets: List<VehicleConditionBuckets.Bucket> = emptyList(),
    val mapTab: VcdTab = VcdTab.Battery,
    val mapBucketIndex: Int = 0,
    val selectedCarId: String? = null,
    val errorMessage: String? = null,
) {
    val listBuckets: List<VehicleConditionBuckets.Bucket>
        get() = when (tab) {
            VcdTab.Battery -> batteryBuckets
            VcdTab.Idle -> idleBuckets
        }

    val mapBuckets: List<VehicleConditionBuckets.Bucket>
        get() = when (mapTab) {
            VcdTab.Battery -> batteryBuckets
            VcdTab.Idle -> idleBuckets
        }

    val mapVehicles: List<Vehicle>
        get() = mapBuckets.getOrNull(mapBucketIndex)?.vehicles.orEmpty()

    val mapPins: List<MapPin>
        get() = mapVehicles
            .filter { it.lat != 0.0 || it.lng != 0.0 }
            .map { it.toMapPin() }
}

class VehicleConditionDistributionFeature(
    private val repository: VehicleRepository,
) {
    private val _state = MutableStateFlow(VehicleConditionDistributionUiState())
    val state: StateFlow<VehicleConditionDistributionUiState> = _state.asStateFlow()

    fun open() {
        _state.value = VehicleConditionDistributionUiState()
    }

    fun clear() {
        _state.value = VehicleConditionDistributionUiState()
    }

    fun selectTab(tab: VcdTab) {
        _state.value = _state.value.copy(tab = tab, errorMessage = null)
    }

    fun openMap(bucketIndex: Int) {
        val s = _state.value
        val max = s.listBuckets.lastIndex.coerceAtLeast(0)
        _state.value = s.copy(
            page = VcdPage.Map,
            mapTab = s.tab,
            mapBucketIndex = bucketIndex.coerceIn(0, max),
            selectedCarId = null,
            errorMessage = null,
        )
    }

    fun selectMapBucket(index: Int) {
        val s = _state.value
        val max = s.mapBuckets.lastIndex.coerceAtLeast(0)
        _state.value = s.copy(
            mapBucketIndex = index.coerceIn(0, max),
            selectedCarId = null,
        )
    }

    fun closeMap() {
        _state.value = _state.value.copy(page = VcdPage.List, selectedCarId = null)
    }

    fun selectVehicle(carId: String?) {
        _state.value = _state.value.copy(selectedCarId = carId)
    }

    fun findMapVehicle(carId: String): Vehicle? =
        _state.value.mapVehicles.firstOrNull { it.carId == carId }

    suspend fun load(area: ServiceArea?) {
        if (area == null) {
            _state.value = _state.value.copy(
                loading = false,
                errorMessage = Strings.t(Str.SelectServiceAreaFirst),
                batteryBuckets = emptyList(),
                idleBuckets = emptyList(),
                totalCount = 0,
            )
            return
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null)
        when (val result = repository.loadByServiceArea(area)) {
            is OpsResult.Ok -> applyVehicles(result.value)
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    errorMessage = result.error.message,
                )
            }
        }
    }

    private fun applyVehicles(vehicles: List<Vehicle>) {
        val now = nowEpochMillis()
        val battery = VehicleConditionBuckets.battery(vehicles)
        val idle = VehicleConditionBuckets.idle(vehicles, now)
        val s = _state.value
        val mapMax = when (s.mapTab) {
            VcdTab.Battery -> battery.buckets.lastIndex
            VcdTab.Idle -> idle.buckets.lastIndex
        }.coerceAtLeast(0)
        _state.value = s.copy(
            loading = false,
            totalCount = battery.total,
            batteryBuckets = battery.buckets,
            idleBuckets = idle.buckets,
            mapBucketIndex = s.mapBucketIndex.coerceIn(0, mapMax),
            selectedCarId = s.selectedCarId
                ?.takeIf { id ->
                    val bucket = when (s.mapTab) {
                        VcdTab.Battery -> battery.buckets
                        VcdTab.Idle -> idle.buckets
                    }.getOrNull(s.mapBucketIndex.coerceIn(0, mapMax))
                    bucket?.vehicles?.any { it.carId == id } == true
                },
            errorMessage = null,
        )
    }
}
