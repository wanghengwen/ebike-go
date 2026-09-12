package com.luopingtech.ebike.ops.feature.sneak

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.order.OrderRepository
import com.luopingtech.ebike.ops.data.sneak.SneakReportRepository
import com.luopingtech.ebike.ops.domain.model.LastOrder
import com.luopingtech.ebike.ops.domain.model.SneakReportRecord
import com.luopingtech.ebike.ops.domain.model.SneakReportType
import com.luopingtech.ebike.ops.domain.scan.ScanCodeParser
import com.luopingtech.ebike.ops.domain.scan.ScanTarget
import com.luopingtech.ebike.ops.platform.DemoMediaUploader
import com.luopingtech.ebike.ops.platform.MediaUploader
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

enum class SneakPage {
    Hub,
    Submit,
    History,
    Detail,
}

data class SneakUiState(
    val page: SneakPage = SneakPage.Hub,
    val loading: Boolean = false,
    val carId: String = "",
    val lastOrder: LastOrder? = null,
    val description: String = "",
    val otherType: String = "",
    val types: List<SneakReportType> = emptyList(),
    val selectedTypeIds: Set<String> = emptySet(),
    val photoUrls: List<String> = emptyList(),
    val records: List<SneakReportRecord> = emptyList(),
    val selectedRecord: SneakReportRecord? = null,
    val message: String? = null,
    val errorMessage: String? = null,
)

/**
 * Field sneak report (legacy ReportActivity / 1214): submit + my list + cancel pending.
 */
class SneakReportFeature(
    private val repository: SneakReportRepository,
    private val orderRepository: OrderRepository,
    private val serviceAreaIdProvider: () -> String,
    private val reportManPinProvider: () -> String,
    private val reportManPhoneProvider: () -> String,
    private val mediaUploader: MediaUploader = DemoMediaUploader(),
) {
    private val _state = MutableStateFlow(SneakUiState())
    val state: StateFlow<SneakUiState> = _state.asStateFlow()

    private var lastLookupKey: String = ""

    fun openHub() {
        lastLookupKey = ""
        _state.value = SneakUiState()
    }

    fun openSubmit() {
        lastLookupKey = ""
        _state.value = SneakUiState(page = SneakPage.Submit)
    }

    fun setCarId(value: String) {
        _state.value = _state.value.copy(carId = value, errorMessage = null, lastOrder = null)
        lastLookupKey = ""
    }

    fun setDescription(value: String) {
        _state.value = _state.value.copy(description = value.take(100), errorMessage = null)
    }

    fun setOtherType(value: String) {
        _state.value = _state.value.copy(otherType = value.take(10), errorMessage = null)
    }

    fun toggleType(typeId: String) {
        val selected = _state.value.selectedTypeIds.toMutableSet()
        if (!selected.add(typeId)) selected.remove(typeId)
        _state.value = _state.value.copy(selectedTypeIds = selected, errorMessage = null)
    }

    fun addDemoPhoto() {
        if (_state.value.photoUrls.size >= MAX_PHOTOS) return
        val next = _state.value.photoUrls + "demo://sneak/${_state.value.photoUrls.size + 1}"
        _state.value = _state.value.copy(photoUrls = next, errorMessage = null)
    }

    fun addPhotoUrl(url: String) {
        val trimmed = url.trim()
        if (trimmed.isBlank() || _state.value.photoUrls.size >= MAX_PHOTOS) return
        _state.value = _state.value.copy(
            photoUrls = _state.value.photoUrls + trimmed,
            errorMessage = null,
        )
    }

    fun removePhoto(url: String) {
        _state.value = _state.value.copy(photoUrls = _state.value.photoUrls.filterNot { it == url })
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
            lastOrder = null,
            message = Strings.t(Str.ScannedInVehicle, carId),
            errorMessage = null,
        )
        return true
    }

    suspend fun ensureTypes() {
        if (_state.value.types.isNotEmpty()) return
        val serviceId = serviceAreaIdProvider()
        _state.value = _state.value.copy(loading = true, errorMessage = null)
        when (val result = repository.loadTypes(serviceId)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(loading = false, types = result.value)
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                errorMessage = result.error.message,
            )
        }
    }

    /** Load last order when car id is ready (legacy: must have order before submit). */
    suspend fun lookupLastOrderIfReady() {
        val raw = _state.value.carId.trim()
        if (raw.length < MIN_CAR_ID_LEN) {
            if (_state.value.lastOrder != null) {
                _state.value = _state.value.copy(lastOrder = null)
            }
            return
        }
        if (raw == lastLookupKey) return

        _state.value = _state.value.copy(loading = true, errorMessage = null)
        val serviceId = serviceAreaIdProvider()
        when (val permission = repository.checkServicePermission(raw, serviceId)) {
            is OpsResult.Err -> {
                lastLookupKey = ""
                _state.value = _state.value.copy(
                    loading = false,
                    lastOrder = null,
                    errorMessage = permission.error.message,
                )
                return
            }
            is OpsResult.Ok -> Unit
        }

        when (val order = orderRepository.detailLast(raw)) {
            is OpsResult.Ok -> {
                lastLookupKey = raw
                val value = order.value
                if (value.id.isBlank() || value.userPin.isBlank()) {
                    _state.value = _state.value.copy(
                        loading = false,
                        lastOrder = null,
                        errorMessage = Strings.t(Str.SneakNeedLastOrder),
                    )
                } else {
                    _state.value = _state.value.copy(
                        loading = false,
                        lastOrder = value,
                        message = Strings.t(Str.SneakLastOrderLoaded, value.id),
                    )
                }
            }
            is OpsResult.Err -> {
                lastLookupKey = ""
                _state.value = _state.value.copy(
                    loading = false,
                    lastOrder = null,
                    errorMessage = Strings.t(Str.SneakLastOrderFailed, order.error.message),
                )
            }
        }
    }

    suspend fun submit() {
        val current = _state.value
        val selected = current.types.filter { it.id in current.selectedTypeIds }
        val order = current.lastOrder
        _state.value = current.copy(loading = true, errorMessage = null, message = null)

        if (order == null || order.id.isBlank() || order.userPin.isBlank()) {
            _state.value = current.copy(
                loading = false,
                errorMessage = Strings.t(Str.SneakNeedLastOrder),
            )
            return
        }

        val remotePhotos = if (current.photoUrls.isEmpty()) {
            emptyList()
        } else {
            when (val uploaded = mediaUploader.upload(current.photoUrls)) {
                is OpsResult.Ok -> uploaded.value
                is OpsResult.Err -> {
                    _state.value = current.copy(
                        loading = false,
                        errorMessage = Strings.t(Str.PhotoUploadFailed, uploaded.error.message),
                    )
                    return
                }
            }
        }

        when (
            val result = repository.submit(
                carId = current.carId,
                description = current.description,
                types = selected,
                otherType = current.otherType,
                photoUrls = remotePhotos,
                itinId = order.id,
                reportedUserPin = order.userPin,
                reportedUserPhone = order.userPhone,
                reportManPin = reportManPinProvider(),
                serviceId = serviceAreaIdProvider(),
            )
        ) {
            is OpsResult.Ok -> {
                lastLookupKey = ""
                _state.value = SneakUiState(
                    page = SneakPage.Hub,
                    message = Strings.t(Str.SneakSubmitOk, result.value.carId),
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
            page = SneakPage.History,
            loading = true,
            errorMessage = null,
            selectedRecord = null,
        )
        when (val result = repository.myReports(reportManPhoneProvider())) {
            is OpsResult.Ok -> _state.value = _state.value.copy(loading = false, records = result.value)
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                errorMessage = result.error.message,
            )
        }
    }

    fun openDetail(record: SneakReportRecord) {
        _state.value = _state.value.copy(page = SneakPage.Detail, selectedRecord = record)
    }

    suspend fun cancelSelected() {
        val record = _state.value.selectedRecord ?: return
        if (!record.canCancel) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.SneakCannotCancel))
            return
        }
        val serviceId = record.serviceId.ifBlank { serviceAreaIdProvider() }
        _state.value = _state.value.copy(loading = true, errorMessage = null)
        when (val result = repository.cancel(record.id, serviceId)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    selectedRecord = record.copy(checkResult = 5),
                    message = Strings.t(Str.SneakCancelOk),
                )
                openHistory()
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    errorMessage = result.error.message,
                )
            }
        }
    }

    fun clear() {
        lastLookupKey = ""
        _state.value = SneakUiState()
    }

    companion object {
        const val MIN_CAR_ID_LEN = 7
        const val MAX_PHOTOS = 3
    }
}
