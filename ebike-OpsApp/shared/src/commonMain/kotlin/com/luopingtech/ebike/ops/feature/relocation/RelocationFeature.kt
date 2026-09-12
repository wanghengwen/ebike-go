package com.luopingtech.ebike.ops.feature.relocation

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.relocation.RelocationRepository
import com.luopingtech.ebike.ops.domain.model.RelocationDevice
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.scan.ScanCodeParser
import com.luopingtech.ebike.ops.platform.GeoPoint
import com.luopingtech.ebike.ops.platform.LocationTracker
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class RelocationUiState(
    val loading: Boolean = false,
    val reporting: Boolean = false,
    val carInput: String = "",
    val devices: List<RelocationDevice> = emptyList(),
    val pinLat: Double? = null,
    val pinLng: Double? = null,
    val message: String? = null,
    val errorMessage: String? = null,
) {
    val selected: List<RelocationDevice>
        get() = devices.filter { it.selected }

    val allSelected: Boolean
        get() = devices.isNotEmpty() && devices.all { it.selected }
}

/**
 * Legacy Flutter relocation: scan offline vehicles → report GPS for selected IMEIs.
 */
class RelocationFeature(
    private val repository: RelocationRepository,
    private val locationTracker: LocationTracker,
    private val qrHostsProvider: () -> List<String> = { emptyList() },
) {
    private val _state = MutableStateFlow(RelocationUiState())
    val state: StateFlow<RelocationUiState> = _state.asStateFlow()

    fun clear() {
        _state.value = RelocationUiState()
    }

    fun setCarInput(value: String) {
        _state.value = _state.value.copy(carInput = value, errorMessage = null)
    }

    fun setPin(lat: Double?, lng: Double?) {
        _state.value = _state.value.copy(pinLat = lat, pinLng = lng, errorMessage = null)
    }

    fun toggleSelected(carId: String) {
        val next = _state.value.devices.map {
            if (it.carId.equals(carId, ignoreCase = true)) it.copy(selected = !it.selected) else it
        }
        _state.value = _state.value.copy(devices = next)
    }

    fun setAllSelected(selected: Boolean) {
        _state.value = _state.value.copy(
            devices = _state.value.devices.map { it.copy(selected = selected) },
        )
    }

    fun remove(carId: String) {
        _state.value = _state.value.copy(
            devices = _state.value.devices.filterNot { it.carId.equals(carId, ignoreCase = true) },
        )
    }

    suspend fun refreshPinFromGps(): OpsResult<GeoPoint> {
        return when (val loc = locationTracker.currentLocation()) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    pinLat = loc.value.latitude,
                    pinLng = loc.value.longitude,
                    errorMessage = null,
                )
                loc
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(errorMessage = loc.error.message)
                loc
            }
        }
    }

    suspend fun addByRaw(raw: String, area: ServiceArea?): OpsResult<RelocationDevice> {
        val hosts = qrHostsProvider()
        val target = ScanCodeParser.parse(raw, hosts)
            ?: return fail(OpsError.business("SCAN", Strings.t(Str.CannotParseScan)))
        return addByTarget(target, area)
    }

    suspend fun addManual(area: ServiceArea?): OpsResult<RelocationDevice> {
        return addByRaw(_state.value.carInput, area)
    }

    private suspend fun addByTarget(
        target: com.luopingtech.ebike.ops.domain.scan.ScanTarget,
        area: ServiceArea?,
    ): OpsResult<RelocationDevice> {
        if (area == null || area.id.isBlank()) {
            return fail(OpsError.business("AREA", Strings.t(Str.SelectServiceAreaFirst)))
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        return when (val result = repository.scanDevice(area.id, target)) {
            is OpsResult.Err -> {
                _state.value = _state.value.copy(loading = false, errorMessage = result.error.message)
                result
            }
            is OpsResult.Ok -> {
                val device = result.value
                if (!device.isOffline) {
                    return fail(OpsError.business("ONLINE", Strings.t(Str.RelocationOfflineOnly)))
                }
                if (device.carId.isBlank() || device.imei.isBlank()) {
                    return fail(OpsError.business("DEVICE", Strings.t(Str.RelocationScanFailed)))
                }
                val exists = _state.value.devices.any {
                    it.carId.equals(device.carId, ignoreCase = true)
                }
                if (exists) {
                    return fail(OpsError.business("DUP", Strings.t(Str.RelocationAlreadyAdded)))
                }
                _state.value = _state.value.copy(
                    loading = false,
                    carInput = "",
                    devices = _state.value.devices + device.copy(selected = true),
                    message = Strings.t(Str.RelocationAdded, device.carId),
                )
                OpsResult.Ok(device)
            }
        }
    }

    suspend fun confirmLocation(): OpsResult<Unit> {
        val selected = _state.value.selected
        if (selected.isEmpty()) {
            return fail(OpsError.business("EMPTY", Strings.t(Str.RelocationNeedSelect)))
        }
        var lat = _state.value.pinLat
        var lng = _state.value.pinLng
        if (lat == null || lng == null) {
            when (val loc = locationTracker.currentLocation()) {
                is OpsResult.Ok -> {
                    lat = loc.value.latitude
                    lng = loc.value.longitude
                    _state.value = _state.value.copy(pinLat = lat, pinLng = lng)
                }
                is OpsResult.Err -> {
                    _state.value = _state.value.copy(errorMessage = loc.error.message)
                    return loc
                }
            }
        }
        val safeLat = lat ?: return fail(OpsError.business("GPS", Strings.t(Str.RelocationNeedGps)))
        val safeLng = lng ?: return fail(OpsError.business("GPS", Strings.t(Str.RelocationNeedGps)))
        _state.value = _state.value.copy(reporting = true, loading = true, errorMessage = null, message = null)
        val imeis = selected.map { it.imei }.filter { it.isNotBlank() }
        return when (val result = repository.reportLocation(imeis, safeLat, safeLng)) {
            is OpsResult.Ok -> {
                val remaining = _state.value.devices.filterNot { device ->
                    imeis.any { it.equals(device.imei, ignoreCase = true) }
                }
                _state.value = _state.value.copy(
                    loading = false,
                    reporting = false,
                    devices = remaining,
                    message = Strings.t(Str.RelocationOk, imeis.size),
                )
                result
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    reporting = false,
                    errorMessage = result.error.message,
                )
                result
            }
        }
    }

    private fun fail(error: OpsError): OpsResult.Err {
        _state.value = _state.value.copy(loading = false, reporting = false, errorMessage = error.message)
        return OpsResult.Err(error)
    }
}
