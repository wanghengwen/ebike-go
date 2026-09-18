package com.luopingtech.ebike.ops.feature.vehicle

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.vehicle.VehicleChangeBatteryHistoryDto
import com.luopingtech.ebike.ops.data.vehicle.VehicleHistoryApi
import com.luopingtech.ebike.ops.data.vehicle.VehicleMoveHistoryDto
import com.luopingtech.ebike.ops.data.vehicle.VehicleScanLogDto
import com.luopingtech.ebike.ops.data.vehicle.VehicleSwitchLockDto
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsTimeRanges
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class VehicleHistoryUiState(
    val loading: Boolean = false,
    val errorMessage: String? = null,
    val scanLogs: List<VehicleScanLogDto> = emptyList(),
    val switchLocks: List<VehicleSwitchLockDto> = emptyList(),
    val changeBattery: List<VehicleChangeBatteryHistoryDto> = emptyList(),
    val moveCar: List<VehicleMoveHistoryDto> = emptyList(),
)

/**
 * Loads secondary history lists opened from CarDetail menus.
 */
class VehicleHistoryFeature(
    private val api: VehicleHistoryApi?,
    private val demoMode: Boolean = false,
) {
    private val _state = MutableStateFlow(VehicleHistoryUiState())
    val state: StateFlow<VehicleHistoryUiState> = _state.asStateFlow()

    fun clear() {
        _state.value = VehicleHistoryUiState()
    }

    suspend fun loadScanLogs(carId: String) {
        val id = carId.trim()
        if (id.isBlank()) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.EnterCarId))
            return
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null, scanLogs = emptyList())
        val client = api
        if (client == null) {
            _state.value = _state.value.copy(
                loading = false,
                scanLogs = if (demoMode) demoScan(id) else emptyList(),
                errorMessage = if (demoMode) null else Strings.t(Str.NoRecords),
            )
            return
        }
        when (val result = client.scanLogs(id)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(
                loading = false,
                scanLogs = result.value.distinctBy { it.time },
            )
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                errorMessage = result.error.message,
            )
        }
    }

    suspend fun loadSwitchLocks(carId: String) {
        val id = carId.trim()
        if (id.isBlank()) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.EnterCarId))
            return
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null, switchLocks = emptyList())
        val client = api
        if (client == null) {
            _state.value = _state.value.copy(
                loading = false,
                switchLocks = if (demoMode) demoSwitch(id) else emptyList(),
            )
            return
        }
        when (val result = client.switchLockLogs(id)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(loading = false, switchLocks = result.value)
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                errorMessage = result.error.message,
            )
        }
    }

    suspend fun loadChangeBattery(carId: String, serviceId: String?) {
        val id = carId.trim()
        val sid = serviceId?.trim().orEmpty()
        if (id.isBlank() || sid.isBlank()) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.SelectServiceAreaFirst))
            return
        }
        val range = OfflineOpsTimeRanges.lastDays(7)
        _state.value = _state.value.copy(loading = true, errorMessage = null, changeBattery = emptyList())
        val client = api
        if (client == null) {
            _state.value = _state.value.copy(
                loading = false,
                changeBattery = if (demoMode) demoBattery(id) else emptyList(),
            )
            return
        }
        when (
            val result = client.changeBatteryHistory(
                carId = id,
                serviceId = sid,
                startTime = range.startText,
                endTime = range.endText,
            )
        ) {
            is OpsResult.Ok -> _state.value = _state.value.copy(loading = false, changeBattery = result.value)
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                errorMessage = result.error.message,
            )
        }
    }

    suspend fun loadMoveCar(carId: String, serviceId: String?) {
        val id = carId.trim()
        val sid = serviceId?.trim().orEmpty()
        if (id.isBlank() || sid.isBlank()) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.SelectServiceAreaFirst))
            return
        }
        val range = OfflineOpsTimeRanges.lastDays(7)
        _state.value = _state.value.copy(loading = true, errorMessage = null, moveCar = emptyList())
        val client = api
        if (client == null) {
            _state.value = _state.value.copy(
                loading = false,
                moveCar = if (demoMode) demoMove(id) else emptyList(),
            )
            return
        }
        when (
            val result = client.moveCarHistory(
                carId = id,
                serviceId = sid,
                start = range.startText,
                end = range.endText,
            )
        ) {
            is OpsResult.Ok -> _state.value = _state.value.copy(loading = false, moveCar = result.value)
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                errorMessage = result.error.message,
            )
        }
    }

    private fun demoScan(carId: String) = listOf(
        VehicleScanLogDto(
            carId = carId,
            phone = "13800000000",
            time = "2026-01-01 12:00:00",
            address = "Demo scan point",
            latitude = 31.23,
            longitude = 121.47,
        ),
    )

    private fun demoSwitch(carId: String) = listOf(
        VehicleSwitchLockDto(
            carId = carId,
            name = "Demo",
            phone = "13800000000",
            time = "2026-01-01 12:00:00",
            eventType = "17514",
        ),
        VehicleSwitchLockDto(
            carId = carId,
            name = "Demo",
            phone = "13800000000",
            time = "2026-01-01 13:00:00",
            eventType = "17516",
        ),
    )

    private fun demoBattery(carId: String) = listOf(
        VehicleChangeBatteryHistoryDto(
            carId = carId,
            opMan = "Demo",
            phone = "13800000000",
            openBatBoxTime = "2026-01-01 12:00:00",
            closeBatBoxTime = "2026-01-01 12:05:00",
            restBatteryBefore = 20,
            restBatteryAfter = 95,
            state = 1,
            checkResult = 2,
        ),
    )

    private fun demoMove(carId: String) = listOf(
        VehicleMoveHistoryDto(
            carId = carId,
            operator = "Demo",
            phone = "13800000000",
            startTime = "2026-01-01 12:00:00",
            endTime = "2026-01-01 12:10:00",
            lastTime = "600",
            checkResult = 2,
        ),
    )
}
