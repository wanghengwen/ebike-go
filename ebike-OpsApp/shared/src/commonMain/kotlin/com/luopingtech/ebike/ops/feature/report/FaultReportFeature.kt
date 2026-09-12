package com.luopingtech.ebike.ops.feature.report

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.report.FaultReportRepository
import com.luopingtech.ebike.ops.data.vehicle.VehicleRepository
import com.luopingtech.ebike.ops.domain.model.FaultReportRecord
import com.luopingtech.ebike.ops.domain.model.RepairType
import com.luopingtech.ebike.ops.domain.scan.ScanCodeParser
import com.luopingtech.ebike.ops.domain.scan.ScanTarget
import com.luopingtech.ebike.ops.platform.DemoMediaUploader
import com.luopingtech.ebike.ops.platform.MediaUploader
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

enum class ReportPage {
    Hub,
    Submit,
    History,
    Detail,
}

data class ReportUiState(
    val page: ReportPage = ReportPage.Hub,
    val loading: Boolean = false,
    val carId: String = "",
    val vehicleModel: String = "",
    val vehicleSummary: String? = null,
    val fixReason: String = "",
    val repairTypes: List<RepairType> = emptyList(),
    val selectedTypeIds: Set<String> = emptySet(),
    val photoUrls: List<String> = emptyList(),
    val izStop: Boolean? = null,
    val records: List<FaultReportRecord> = emptyList(),
    val selectedRecord: FaultReportRecord? = null,
    val message: String? = null,
    val errorMessage: String? = null,
)

class FaultReportFeature(
    private val repository: FaultReportRepository,
    private val serviceAreaIdProvider: () -> String,
    private val vehicleRepository: VehicleRepository,
    private val mediaUploader: MediaUploader = DemoMediaUploader(),
) {
    private val _state = MutableStateFlow(ReportUiState())
    val state: StateFlow<ReportUiState> = _state.asStateFlow()

    /** Last car id / imei that completed a lookup (success or clear). */
    private var lastLookupKey: String = ""

    fun openHub() {
        lastLookupKey = ""
        _state.value = ReportUiState()
    }

    fun openSubmit() {
        lastLookupKey = ""
        _state.value = ReportUiState(page = ReportPage.Submit)
    }

    fun setCarId(value: String) {
        // Legacy LengthFilter(10) on typed vehicle number; scan/IMEI path uses applyScanRaw.
        val clipped = value.take(MAX_CAR_ID_LEN)
        _state.value = _state.value.copy(carId = clipped, errorMessage = null)
    }

    fun applyScanRaw(raw: String, qrHosts: List<String> = emptyList()): Boolean {
        val target = ScanCodeParser.parse(raw, qrHosts)
            ?: run {
                _state.value = _state.value.copy(errorMessage = Strings.t(Str.CannotParseScan))
                return false
            }
        val carId = when (target) {
            is ScanTarget.CarId -> target.value
            is ScanTarget.Imei -> target.value
        }
        lastLookupKey = ""
        _state.value = _state.value.copy(
            carId = carId,
            message = Strings.t(Str.ScannedInVehicle, carId),
            errorMessage = null,
            vehicleModel = "",
            vehicleSummary = null,
        )
        return true
    }

    /**
     * Legacy VehicleRepairActivity: after carId length ≥ 7, getCarDetail → repairList(model).
     */
    suspend fun lookupVehicleIfReady() {
        val raw = _state.value.carId.trim()
        if (raw.length < MIN_CAR_ID_LEN) {
            if (lastLookupKey.isNotEmpty() || _state.value.vehicleModel.isNotEmpty()) {
                lastLookupKey = ""
                _state.value = _state.value.copy(
                    vehicleModel = "",
                    vehicleSummary = null,
                )
            }
            return
        }
        if (raw == lastLookupKey) return

        _state.value = _state.value.copy(loading = true, errorMessage = null)
        val target = resolveScanTarget(raw)
        when (val detail = vehicleRepository.findByScanTarget(target)) {
            is OpsResult.Ok -> {
                val vehicle = detail.value
                val model = vehicle.model.ifBlank { "1" }
                lastLookupKey = raw
                _state.value = _state.value.copy(
                    loading = false,
                    carId = vehicle.carId.ifBlank { raw },
                    vehicleModel = model,
                    vehicleSummary = Strings.t(
                        Str.FaultVehicleResolved,
                        vehicle.carId,
                        vehicle.restBattery,
                        model,
                    ),
                    message = Strings.t(Str.FaultModelTypesLoaded, model),
                    selectedTypeIds = emptySet(),
                    repairTypes = emptyList(),
                )
                reloadRepairTypes(model)
            }
            is OpsResult.Err -> {
                lastLookupKey = ""
                _state.value = _state.value.copy(
                    loading = false,
                    vehicleModel = "",
                    vehicleSummary = null,
                    errorMessage = Strings.t(Str.FaultCarLookupFailed, detail.error.message),
                )
            }
        }
    }

    fun setFixReason(value: String) {
        _state.value = _state.value.copy(fixReason = value.take(MAX_FIX_REASON_LEN), errorMessage = null)
    }

    fun setIzStop(value: Boolean?) {
        _state.value = _state.value.copy(izStop = value, errorMessage = null)
    }

    fun toggleType(typeId: String) {
        val selected = _state.value.selectedTypeIds.toMutableSet()
        if (!selected.add(typeId)) selected.remove(typeId)
        _state.value = _state.value.copy(selectedTypeIds = selected, errorMessage = null)
    }

    fun addDemoPhoto() {
        if (_state.value.photoUrls.size >= MAX_PHOTOS) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.PhotosRequired))
            return
        }
        val next = _state.value.photoUrls + "demo://photo/${_state.value.photoUrls.size + 1}"
        _state.value = _state.value.copy(photoUrls = next, errorMessage = null)
    }

    fun addPhotoUrl(url: String) {
        val trimmed = url.trim()
        if (trimmed.isBlank()) return
        if (_state.value.photoUrls.size >= MAX_PHOTOS) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.PhotosRequired))
            return
        }
        _state.value = _state.value.copy(
            photoUrls = _state.value.photoUrls + trimmed,
            errorMessage = null,
        )
    }

    fun removePhoto(url: String) {
        _state.value = _state.value.copy(
            photoUrls = _state.value.photoUrls.filterNot { it == url },
        )
    }

    suspend fun ensureRepairTypes() {
        if (_state.value.repairTypes.isNotEmpty()) return
        reloadRepairTypes(_state.value.vehicleModel)
    }

    private suspend fun reloadRepairTypes(model: String) {
        _state.value = _state.value.copy(loading = true, errorMessage = null)
        when (val result = repository.loadRepairTypes(model)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(loading = false, repairTypes = result.value)
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    errorMessage = result.error.message,
                )
            }
        }
    }

    suspend fun submit() {
        val current = _state.value
        val selected = current.repairTypes.filter { it.id in current.selectedTypeIds }
        _state.value = current.copy(loading = true, errorMessage = null, message = null)

        // Legacy VehicleRepairActivity has carPermissionCheck commented out; do not hard-block submit.

        val remotePhotos = when (val uploaded = mediaUploader.upload(current.photoUrls)) {
            is OpsResult.Ok -> uploaded.value
            is OpsResult.Err -> {
                _state.value = current.copy(
                    loading = false,
                    errorMessage = Strings.t(Str.PhotoUploadFailed, uploaded.error.message),
                )
                return
            }
        }

        when (
            val result = repository.submit(
                carId = current.carId,
                fixReason = current.fixReason,
                types = selected,
                photoUrls = remotePhotos,
                izStop = current.izStop,
            )
        ) {
            is OpsResult.Ok -> {
                lastLookupKey = ""
                _state.value = ReportUiState(
                    page = ReportPage.Hub,
                    message = Strings.t(Str.FaultReportSuccess, result.value.carId),
                )
            }
            is OpsResult.Err -> {
                _state.value = current.copy(
                    loading = false,
                    errorMessage = result.error.message,
                )
            }
        }
    }

    suspend fun openHistory() {
        _state.value = _state.value.copy(
            page = ReportPage.History,
            loading = true,
            errorMessage = null,
            selectedRecord = null,
        )
        when (val result = repository.myReports(serviceAreaIdProvider())) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(loading = false, records = result.value)
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    errorMessage = result.error.message,
                )
            }
        }
    }

    fun openDetail(record: FaultReportRecord) {
        _state.value = _state.value.copy(
            page = ReportPage.Detail,
            selectedRecord = record,
        )
    }

    fun clear() {
        lastLookupKey = ""
        _state.value = ReportUiState()
    }

    private fun resolveScanTarget(raw: String): ScanTarget {
        val digitsOnly = raw.all { it.isDigit() }
        return if (digitsOnly && raw.length >= 15) {
            ScanTarget.Imei(raw)
        } else {
            ScanTarget.CarId(raw)
        }
    }

    companion object {
        const val MIN_CAR_ID_LEN = 7
        /** Legacy isValidVehicleNum / LengthFilter(10) for typed car ids. */
        const val MAX_CAR_ID_LEN = 10
        const val MAX_FIX_REASON_LEN = 100
        const val MAX_PHOTOS = 3
    }
}
