package com.luopingtech.ebike.ops.feature.production

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.production.ProductionRepository
import com.luopingtech.ebike.ops.data.vehicle.VehicleRepository
import com.luopingtech.ebike.ops.domain.control.ControlChannel
import com.luopingtech.ebike.ops.domain.control.VehicleAction
import com.luopingtech.ebike.ops.domain.control.VehicleControlPolicy
import com.luopingtech.ebike.ops.domain.model.BindVehicleInfo
import com.luopingtech.ebike.ops.domain.model.DetectStepKind
import com.luopingtech.ebike.ops.domain.model.DetectStepState
import com.luopingtech.ebike.ops.domain.model.DetectStepStatus
import com.luopingtech.ebike.ops.domain.model.SaddleOverloadContact
import com.luopingtech.ebike.ops.domain.model.ShelfCheckResult
import com.luopingtech.ebike.ops.domain.model.Vehicle
import com.luopingtech.ebike.ops.domain.scan.ScanCodeParser
import com.luopingtech.ebike.ops.domain.scan.ScanTarget
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock

enum class ProductionPage {
    Hub,
    Detect,
    Bind,
    Shelves,
    Overload,
}

enum class ShelfMode {
    PutOn,
    PullOff,
}

/** Legacy VehicleDetectActivity pager: 网络检测 / 蓝牙检测. */
enum class DetectChannelTab {
    Network,
    Bluetooth,
}

/** Switch options on detect tabs (legacy DetectOptionEnum switches; location is a separate action). */
enum class DetectSwitchKind {
    Acc,
    Defend,
    Battery,
    Helmet,
    Wheel,
}

data class DetectSwitchState(
    val kind: DetectSwitchKind,
    val label: String,
    /** Optimistic on/open state after a successful toggle. */
    val on: Boolean = false,
    val busy: Boolean = false,
)

data class ProductionUiState(
    val page: ProductionPage = ProductionPage.Hub,
    val loading: Boolean = false,
    val carId: String = "",
    val imei: String = "",
    val helmet: String = "",
    val bindInfo: BindVehicleInfo? = null,
    /** Single search field on detect (legacy VehicleSearchView). */
    val searchInput: String = "",
    /** Last queried detect vehicle; drives info rows + switch polarity. */
    val detectVehicle: Vehicle? = null,
    val detectChannel: DetectChannelTab = DetectChannelTab.Network,
    val detectSteps: List<DetectStepState> = defaultDetectSteps(),
    val detectSwitches: List<DetectSwitchState> = defaultDetectSwitches(),
    val bleAvailable: Boolean = true,
    val shelfMode: ShelfMode = ShelfMode.PutOn,
    val shelfServiceId: String = "",
    /** Legacy originPutList / originPullList — kept separate when switching tabs. */
    val putShelfQueue: List<ShelfCheckResult> = emptyList(),
    val pullShelfQueue: List<ShelfCheckResult> = emptyList(),
    val overloadContacts: SaddleOverloadContact = SaddleOverloadContact.Idle,
    val overloadCountdown: Int = 0,
    val overloadRunning: Boolean = false,
    val hasOverloadDevice: Boolean? = null,
    val message: String? = null,
    val errorMessage: String? = null,
) {
    val shelfQueue: List<ShelfCheckResult>
        get() = if (shelfMode == ShelfMode.PutOn) putShelfQueue else pullShelfQueue

    val controlChannel: ControlChannel
        get() = when (detectChannel) {
            DetectChannelTab.Network -> ControlChannel.NetworkOnly
            DetectChannelTab.Bluetooth -> ControlChannel.BleOnly
        }
}

private fun defaultDetectSteps(): List<DetectStepState> = listOf(
    DetectStepState(DetectStepKind.Ring, Strings.t(Str.DetectFindSound)),
    DetectStepState(DetectStepKind.Refresh, Strings.t(Str.Refresh)),
    DetectStepState(DetectStepKind.Reboot, Strings.t(Str.DetectReboot)),
)

private fun defaultDetectSwitches(): List<DetectSwitchState> = listOf(
    DetectSwitchState(DetectSwitchKind.Acc, Strings.t(Str.DetectSwitchAcc)),
    DetectSwitchState(DetectSwitchKind.Defend, Strings.t(Str.DetectSwitchDefend)),
    DetectSwitchState(DetectSwitchKind.Battery, Strings.t(Str.DetectSwitchBattery)),
    DetectSwitchState(DetectSwitchKind.Helmet, Strings.t(Str.DetectSwitchHelmet)),
    DetectSwitchState(DetectSwitchKind.Wheel, Strings.t(Str.DetectSwitchWheel)),
)

class ProductionFeature(
    private val repository: ProductionRepository,
    private val control: VehicleControlPolicy,
    private val serviceAreaIdProvider: () -> String,
    private val vehicleRepository: VehicleRepository,
    private val bleAvailableProvider: () -> Boolean = { true },
    /** Test hook: shorten overload poll window (legacy UI uses 60s). */
    private val overloadTimeoutSec: Int = 60,
    private val overloadPollDelayMs: Long = 1_000L,
    /**
     * Legacy handlerCallback waits 1500ms then re-queries device.
     * Pass a negative value in tests to skip the refresh.
     */
    private val refreshAfterControlDelayMs: Long = 1_500L,
) {
    private val _state = MutableStateFlow(ProductionUiState(bleAvailable = bleAvailableProvider()))
    val state: StateFlow<ProductionUiState> = _state.asStateFlow()
    private val overloadMutex = Mutex()
    private var overloadGeneration: Int = 0
    /** Skip one auto-query after programmatic carId fill (IMEI reverse lookup / scan). */
    private var suppressCarIdAutoQuery = false
    /** Skip one auto-query after programmatic imei fill (bind info / scan). */
    private var suppressImeiAutoQuery = false

    fun openHub() {
        stopOverloadCheck()
        _state.value = ProductionUiState(
            shelfServiceId = serviceAreaIdProvider(),
            bleAvailable = bleAvailableProvider(),
        )
    }

    fun openDetect() {
        stopOverloadCheck()
        _state.value = ProductionUiState(
            page = ProductionPage.Detect,
            shelfServiceId = serviceAreaIdProvider(),
            bleAvailable = bleAvailableProvider(),
            detectChannel = DetectChannelTab.Network,
        )
    }

    fun openBind() {
        stopOverloadCheck()
        suppressCarIdAutoQuery = false
        suppressImeiAutoQuery = false
        _state.value = ProductionUiState(
            page = ProductionPage.Bind,
            shelfServiceId = serviceAreaIdProvider(),
            bleAvailable = bleAvailableProvider(),
        )
    }

    fun openShelves(mode: ShelfMode) {
        stopOverloadCheck()
        _state.value = ProductionUiState(
            page = ProductionPage.Shelves,
            shelfMode = mode,
            shelfServiceId = serviceAreaIdProvider(),
            bleAvailable = bleAvailableProvider(),
        )
    }

    /** Switch 上架/下架 tab without wiping the other list (legacy put/pull). */
    fun setShelfMode(mode: ShelfMode) {
        if (_state.value.page != ProductionPage.Shelves) {
            openShelves(mode)
            return
        }
        if (_state.value.shelfMode == mode) return
        _state.value = _state.value.copy(
            shelfMode = mode,
            carId = "",
            message = null,
            errorMessage = null,
        )
    }

    /** Leave overload pane without wiping carId / imei. */
    fun backToDetect() {
        stopOverloadCheck()
        _state.value = _state.value.copy(
            page = ProductionPage.Detect,
            overloadRunning = false,
            overloadCountdown = 0,
            overloadContacts = SaddleOverloadContact.Idle,
            message = null,
            errorMessage = null,
        )
    }

    fun setDetectChannel(tab: DetectChannelTab) {
        if (_state.value.detectChannel == tab) return
        _state.value = _state.value.copy(
            detectChannel = tab,
            bleAvailable = bleAvailableProvider(),
            errorMessage = null,
        )
    }

    fun setSearchInput(value: String) {
        _state.value = _state.value.copy(searchInput = value, errorMessage = null)
    }

    fun setCarId(value: String) {
        _state.value = _state.value.copy(carId = value, errorMessage = null)
    }

    fun setImei(value: String) {
        _state.value = _state.value.copy(imei = value, errorMessage = null)
    }

    /**
     * Field scan → fill [carId] (车号 QR) or [imei] (15 位中控).
     * @return true when parse succeeded.
     */
    fun applyScanRaw(raw: String, qrHosts: List<String> = emptyList()): Boolean {
        val trimmed = raw.trim()
        // Production floor: bare 15-digit barcodes are controller IMEIs (not carId).
        val target = if (trimmed.length == 15 && trimmed.all { it.isDigit() }) {
            ScanTarget.Imei(trimmed)
        } else {
            ScanCodeParser.parse(trimmed, qrHosts)
        } ?: run {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.CannotParseScan))
            return false
        }
        return when (target) {
            is ScanTarget.CarId -> {
                _state.value = _state.value.copy(
                    carId = target.value,
                    message = Strings.t(Str.ScannedInVehicle, target.value),
                    errorMessage = null,
                )
                true
            }
            is ScanTarget.Imei -> {
                _state.value = _state.value.copy(
                    imei = target.value,
                    // Detect / shelves 常用 IMEI 条码当车辆键；绑定页同时填 imei。
                    carId = _state.value.carId.ifBlank { target.value },
                    message = Strings.t(Str.ScannedImei, target.value),
                    errorMessage = null,
                )
                true
            }
        }
    }

    fun setHelmet(value: String) {
        _state.value = _state.value.copy(helmet = value, errorMessage = null)
    }

    fun setTransientMessage(message: String) {
        _state.value = _state.value.copy(message = message, errorMessage = null)
    }

    /** Legacy checkClickEnable: 车号 1..10 且（设备号恰好 15 位 或 头盔非空）. */
    fun canBindOrUnbind(): Boolean {
        val carId = _state.value.carId.trim()
        val imei = _state.value.imei.trim()
        val helmet = _state.value.helmet.trim()
        val carOk = carId.length in MIN_CAR_ID_LEN..MAX_CAR_ID_LEN
        return carOk && (imei.length == BIND_IMEI_EXACT_LEN || helmet.isNotEmpty())
    }

    /**
     * Legacy CenterControlBindActivity barcodeResult:
     * IMEI:… → 设备号；含 carId → 车辆号；含 MAC → 头盔号。
     */
    suspend fun applyBindScan(raw: String, qrHosts: List<String> = emptyList()): Boolean {
        val text = raw.trim()
        if (text.isEmpty()) return false

        parseHelmetMac(text)?.let { mac ->
            _state.value = _state.value.copy(
                helmet = mac.take(BIND_HELMET_MAX_LEN),
                message = Strings.t(Str.ScannedHelmetMac, mac),
                errorMessage = null,
            )
            return true
        }

        if (text.startsWith("IMEI", ignoreCase = true)) {
            val deviceId = extractImeiFromScan(text)
            if (deviceId != null) {
                suppressImeiAutoQuery = true
                _state.value = _state.value.copy(
                    imei = deviceId.take(BIND_IMEI_EXACT_LEN),
                    message = Strings.t(Str.ScannedImei, deviceId),
                    errorMessage = null,
                )
                if (_state.value.carId.isBlank()) {
                    fillCarIdFromImei(deviceId)
                }
                return true
            }
        }

        val target = when {
            text.length == BIND_IMEI_EXACT_LEN && text.all { it.isDigit() } -> ScanTarget.Imei(text)
            else -> ScanCodeParser.parse(text, qrHosts)
        }
        return when (target) {
            is ScanTarget.CarId -> {
                val carId = target.value.trim().take(MAX_CAR_ID_LEN)
                if (carId.isEmpty()) {
                    _state.value = _state.value.copy(errorMessage = Strings.t(Str.CannotParseScan))
                    return false
                }
                suppressCarIdAutoQuery = true
                _state.value = _state.value.copy(
                    carId = carId,
                    message = Strings.t(Str.ScannedInVehicle, carId),
                    errorMessage = null,
                )
                loadBindInfo(autoFill = true)
                true
            }
            is ScanTarget.Imei -> {
                suppressImeiAutoQuery = true
                _state.value = _state.value.copy(
                    imei = target.value.filter { it.isDigit() }.take(BIND_IMEI_EXACT_LEN),
                    message = Strings.t(Str.ScannedImei, target.value),
                    errorMessage = null,
                )
                if (_state.value.carId.isBlank()) {
                    fillCarIdFromImei(target.value)
                }
                true
            }
            null -> {
                _state.value = _state.value.copy(errorMessage = Strings.t(Str.CannotParseScan))
                false
            }
        }
    }

    /** Legacy carIdWatcher: 有效车号后自动 getBindVehicleInfo. */
    suspend fun onBindCarIdChanged() {
        if (suppressCarIdAutoQuery) {
            suppressCarIdAutoQuery = false
            return
        }
        val carId = _state.value.carId.trim()
        if (carId.length in MIN_CAR_ID_LEN..MAX_CAR_ID_LEN) {
            loadBindInfo(autoFill = true)
        }
    }

    /** Legacy deviceIdWatcher: carId 空且 imei 长度>6 时按 IMEI 反查车号. */
    suspend fun onBindImeiChanged() {
        if (suppressImeiAutoQuery) {
            suppressImeiAutoQuery = false
            return
        }
        val carId = _state.value.carId.trim()
        val imei = _state.value.imei.trim()
        if (carId.isEmpty() && imei.length > 6) {
            fillCarIdFromImei(imei)
        }
    }

    private suspend fun fillCarIdFromImei(imei: String) {
        when (val result = vehicleRepository.findByScanTarget(ScanTarget.Imei(imei.trim()))) {
            is OpsResult.Ok -> {
                val carId = result.value.carId.trim()
                if (carId.isNotEmpty()) {
                    suppressCarIdAutoQuery = true
                    _state.value = _state.value.copy(
                        carId = carId.take(MAX_CAR_ID_LEN),
                        errorMessage = null,
                    )
                }
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(errorMessage = result.error.message)
            }
        }
    }

    fun setShelfServiceId(value: String) {
        _state.value = _state.value.copy(shelfServiceId = value, errorMessage = null)
    }

    fun removeShelfItem(carId: String) {
        val s = _state.value
        _state.value = if (s.shelfMode == ShelfMode.PutOn) {
            s.copy(putShelfQueue = s.putShelfQueue.filterNot { it.carId.equals(carId, ignoreCase = true) })
        } else {
            s.copy(pullShelfQueue = s.pullShelfQueue.filterNot { it.carId.equals(carId, ignoreCase = true) })
        }
    }

    fun setShelfItemSelected(carId: String, selected: Boolean) {
        fun mapList(list: List<ShelfCheckResult>) =
            list.map { if (it.carId.equals(carId, ignoreCase = true)) it.copy(selected = selected) else it }
        val s = _state.value
        _state.value = if (s.shelfMode == ShelfMode.PutOn) {
            s.copy(putShelfQueue = mapList(s.putShelfQueue))
        } else {
            s.copy(pullShelfQueue = mapList(s.pullShelfQueue))
        }
    }

    fun setShelfSelectAll(selected: Boolean) {
        fun mapList(list: List<ShelfCheckResult>) = list.map { it.copy(selected = selected) }
        val s = _state.value
        _state.value = if (s.shelfMode == ShelfMode.PutOn) {
            s.copy(putShelfQueue = mapList(s.putShelfQueue))
        } else {
            s.copy(pullShelfQueue = mapList(s.pullShelfQueue))
        }
    }

    fun canAddShelfCar(): Boolean {
        val carId = _state.value.carId.trim()
        return carId.length in MIN_CAR_ID_LEN..MAX_CAR_ID_LEN && !_state.value.loading
    }

    fun hasShelfSelection(): Boolean = _state.value.shelfQueue.any { it.selected }

    fun isShelfAllSelected(): Boolean {
        val q = _state.value.shelfQueue
        return q.isNotEmpty() && q.all { it.selected }
    }

    /**
     * Legacy shelves scan: decode carId then auto-add.
     * Also accepts device QR (`IMEI:…` / 15-digit barcode) → resolve carId then add.
     */
    suspend fun applyShelvesScan(raw: String, qrHosts: List<String> = emptyList()): Boolean {
        val text = raw.trim()
        if (text.isEmpty()) return false

        val imeiFromDevice = when {
            text.startsWith("IMEI", ignoreCase = true) -> extractImeiFromScan(text)
            text.length in MIN_IMEI_LEN..MAX_IMEI_LEN && text.all { it.isDigit() } -> text
            else -> null
        }
        if (imeiFromDevice != null) {
            when (val found = vehicleRepository.findByScanTarget(ScanTarget.Imei(imeiFromDevice))) {
                is OpsResult.Ok -> {
                    val carId = found.value.carId.trim().take(MAX_CAR_ID_LEN)
                    if (carId.isEmpty()) {
                        _state.value = _state.value.copy(
                            errorMessage = Strings.t(Str.CannotParseScan),
                            message = null,
                        )
                        return false
                    }
                    _state.value = _state.value.copy(carId = carId, errorMessage = null, message = null)
                    addShelfCar()
                    return _state.value.errorMessage == null
                }
                is OpsResult.Err -> {
                    _state.value = _state.value.copy(
                        errorMessage = found.error.message,
                        message = null,
                    )
                    return false
                }
            }
        }

        val target = ScanCodeParser.parse(text, qrHosts)
        val carId = when (target) {
            is ScanTarget.CarId -> target.value.trim().take(MAX_CAR_ID_LEN)
            is ScanTarget.Imei -> {
                // Parser may classify bare digits / imei= as Imei.
                when (val found = vehicleRepository.findByScanTarget(target)) {
                    is OpsResult.Ok -> found.value.carId.trim().take(MAX_CAR_ID_LEN)
                    is OpsResult.Err -> {
                        _state.value = _state.value.copy(
                            errorMessage = found.error.message,
                            message = null,
                        )
                        return false
                    }
                }
            }
            null -> {
                _state.value = _state.value.copy(
                    errorMessage = Strings.t(Str.CannotParseScan),
                    message = null,
                )
                return false
            }
        }
        if (carId.isEmpty()) {
            _state.value = _state.value.copy(
                errorMessage = Strings.t(Str.CannotParseScan),
                message = null,
            )
            return false
        }
        _state.value = _state.value.copy(carId = carId, errorMessage = null, message = null)
        addShelfCar()
        return _state.value.errorMessage == null
    }

    /**
     * Legacy onQueryAction / scan: 1–10 位车号或 14–18 位中控码，成功后回填开关状态。
     */
    suspend fun queryDetectVehicle(
        raw: String? = null,
        qrHosts: List<String> = emptyList(),
        showLoading: Boolean = true,
    ) {
        val input = (raw ?: _state.value.searchInput).trim()
        val target = if (input.isNotBlank()) {
            resolveDetectInput(input, qrHosts)
        } else {
            val carId = _state.value.carId.trim()
            val imei = _state.value.imei.trim()
            when {
                carId.isNotBlank() -> ScanTarget.CarId(carId)
                imei.isNotBlank() -> ScanTarget.Imei(imei)
                else -> null
            }
        }
        if (target == null) {
            _state.value = _state.value.copy(
                errorMessage = Strings.t(Str.DetectInvalidCarOrImei),
                message = null,
            )
            return
        }
        if (raw != null) {
            _state.value = _state.value.copy(searchInput = raw.trim())
        }
        if (showLoading) {
            _state.value = _state.value.copy(
                loading = true,
                errorMessage = null,
                message = Strings.t(Str.Querying),
            )
        }
        when (val result = vehicleRepository.findByScanTarget(target)) {
            is OpsResult.Ok -> applyVehicleToDetect(result.value)
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    errorMessage = result.error.message,
                    message = null,
                )
            }
        }
    }

    /** Legacy footer: 寻车音 / 刷新 / 重启. Ring+Reboot use NETWORK_FIRST. */
    suspend fun runDetectStep(kind: DetectStepKind) {
        val vehicle = _state.value.detectVehicle
        val carId = vehicle?.carId?.ifBlank { _state.value.carId.trim() }.orEmpty()
        if (vehicle == null && carId.isBlank()) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.DetectNoVehicleInfo))
            return
        }
        when (kind) {
            DetectStepKind.Refresh -> {
                updateStep(kind, DetectStepStatus.Running)
                queryDetectVehicle(showLoading = true)
                val ok = _state.value.errorMessage == null && _state.value.detectVehicle != null
                updateStep(
                    kind,
                    if (ok) DetectStepStatus.Ok else DetectStepStatus.Failed,
                    _state.value.errorMessage.orEmpty(),
                )
            }
            DetectStepKind.Ring, DetectStepKind.Reboot -> {
                updateStep(kind, DetectStepStatus.Running)
                val action = if (kind == DetectStepKind.Ring) {
                    VehicleAction.Ring
                } else {
                    VehicleAction.Restart
                }
                if (kind == DetectStepKind.Ring) {
                    _state.value = _state.value.copy(
                        message = Strings.t(Str.DetectPlayFindSound),
                        errorMessage = null,
                    )
                }
                when (
                    val result = control.execute(
                        vehicleId = carId.ifBlank { vehicle?.carId.orEmpty() },
                        action = action,
                        channel = ControlChannel.NetworkPreferred,
                        imei = _state.value.imei.trim().ifBlank { vehicle?.imei.orEmpty() },
                    )
                ) {
                    is OpsResult.Ok -> {
                        val okMsg = if (kind == DetectStepKind.Reboot) {
                            Strings.t(Str.DetectRebootOk)
                        } else {
                            Strings.t(Str.DetectPlayFindSound)
                        }
                        updateStep(kind, DetectStepStatus.Ok, okMsg)
                        _state.value = _state.value.copy(message = okMsg, errorMessage = null)
                    }
                    is OpsResult.Err -> {
                        val failMsg = if (kind == DetectStepKind.Reboot) {
                            result.error.message.ifBlank { Strings.t(Str.DetectRebootFail) }
                        } else {
                            result.error.message
                        }
                        updateStep(kind, DetectStepStatus.Failed, failMsg)
                    }
                }
                refreshDetectAfterControl()
            }
        }
    }

    /**
     * Legacy OPTION_LOCATION: resolve detail by carId (preferred) or IMEI, then host UI
     * closes detect and focuses the home map on the vehicle (no separate Location Activity).
     */
    suspend fun locateVehicleOnMap(): OpsResult<Vehicle> {
        val carId = _state.value.carId.trim()
        val imei = _state.value.imei.trim()
        if (carId.isBlank() && imei.isBlank()) {
            val msg = Strings.t(Str.EnterCarIdOrScan)
            _state.value = _state.value.copy(errorMessage = msg)
            return OpsResult.Err(OpsError.business("CAR", msg))
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        val target = when {
            carId.isNotBlank() -> ScanTarget.CarId(carId)
            else -> ScanTarget.Imei(imei)
        }
        return when (val result = vehicleRepository.findByScanTarget(target)) {
            is OpsResult.Ok -> {
                val vehicle = result.value
                _state.value = _state.value.copy(
                    loading = false,
                    carId = vehicle.carId.ifBlank { carId },
                    imei = vehicle.imei.ifBlank { imei },
                    message = Strings.t(Str.DetectLocationOk, vehicle.carId),
                    errorMessage = null,
                )
                result
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    errorMessage = result.error.message,
                )
                result
            }
        }
    }

    /**
     * Legacy detect switch tiles: acc / defend / battery / helmet / wheel.
     * [on] = true means open/acc-on/defend-on (same polarity as Photograph UI switches).
     */
    suspend fun toggleDetectSwitch(kind: DetectSwitchKind, on: Boolean) {
        val vehicle = _state.value.detectVehicle
        val carId = vehicle?.carId?.ifBlank { _state.value.carId.trim() }.orEmpty()
        if (vehicle == null && carId.isBlank()) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.DetectNoVehicleInfo))
            return
        }
        val channel = _state.value.controlChannel
        if (channel == ControlChannel.BleOnly && !bleAvailableProvider()) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.DetectBleUnavailableHint))
            return
        }
        updateSwitch(kind) { it.copy(busy = true) }
        val action = actionForSwitch(kind, on)
        val channelLabel = when (_state.value.detectChannel) {
            DetectChannelTab.Network -> Strings.t(Str.DetectChannelNetwork)
            DetectChannelTab.Bluetooth -> Strings.t(Str.DetectChannelBle)
        }
        when (
            val result = control.execute(
                vehicleId = carId.ifBlank { vehicle?.carId.orEmpty() },
                action = action,
                channel = channel,
                imei = _state.value.imei.trim().ifBlank { vehicle?.imei.orEmpty() },
            )
        ) {
            is OpsResult.Ok -> {
                updateSwitch(kind) { it.copy(on = on, busy = false) }
                _state.value = _state.value.copy(
                    message = Strings.t(Str.DetectSwitchOk, kindLabel(kind), channelLabel),
                    errorMessage = null,
                )
            }
            is OpsResult.Err -> {
                updateSwitch(kind) { it.copy(busy = false) }
                _state.value = _state.value.copy(errorMessage = result.error.message)
            }
        }
        refreshDetectAfterControl()
    }

    private fun actionForSwitch(kind: DetectSwitchKind, on: Boolean): VehicleAction = when (kind) {
        DetectSwitchKind.Acc -> if (on) VehicleAction.AccOn else VehicleAction.AccOff
        DetectSwitchKind.Defend -> if (on) VehicleAction.DefendOn else VehicleAction.DefendOff
        DetectSwitchKind.Battery -> if (on) VehicleAction.OpenBatteryBox else VehicleAction.CloseBatteryBox
        DetectSwitchKind.Helmet -> if (on) VehicleAction.OpenHelmetLock else VehicleAction.CloseHelmetLock
        DetectSwitchKind.Wheel -> if (on) VehicleAction.OpenBackWheelLock else VehicleAction.CloseBackWheelLock
    }

    private fun kindLabel(kind: DetectSwitchKind): String =
        _state.value.detectSwitches.firstOrNull { it.kind == kind }?.label
            ?: kind.name

    private fun updateSwitch(kind: DetectSwitchKind, transform: (DetectSwitchState) -> DetectSwitchState) {
        _state.value = _state.value.copy(
            detectSwitches = _state.value.detectSwitches.map { switch ->
                if (switch.kind == kind) transform(switch) else switch
            },
        )
    }

    suspend fun loadBindInfo(autoFill: Boolean = false) {
        val carId = _state.value.carId.trim()
        if (carId.isBlank()) {
            if (!autoFill) {
                _state.value = _state.value.copy(errorMessage = Strings.t(Str.EnterCarId))
            }
            return
        }
        if (!autoFill) {
            _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        }
        when (val result = repository.getBind(carId)) {
            is OpsResult.Ok -> {
                val info = result.value
                if (info.imei.isNotBlank()) {
                    suppressImeiAutoQuery = true
                }
                _state.value = _state.value.copy(
                    loading = false,
                    bindInfo = info,
                    imei = info.imei.ifBlank { _state.value.imei },
                    helmet = info.helmet.ifBlank { _state.value.helmet },
                    message = when {
                        autoFill && info.imei.isBlank() -> _state.value.message
                        info.imei.isBlank() -> Strings.t(Str.NotBoundImei)
                        else -> Strings.t(Str.BoundImei, info.imei)
                    },
                    errorMessage = null,
                )
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    errorMessage = if (autoFill) null else result.error.message,
                )
            }
        }
    }

    suspend fun bindCenter() {
        val carId = _state.value.carId.trim()
        val imei = _state.value.imei.trim()
        val helmet = _state.value.helmet.trim()
        if (carId.isBlank()) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.EnterCarId))
            return
        }
        if (imei.isBlank() && helmet.isBlank()) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.BindNeedDeviceOrHelmet))
            return
        }
        if (!canBindOrUnbind()) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.BindNeedDeviceOrHelmet))
            return
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        when (
            val result = repository.bind(carId, imei.ifBlank { null }, helmet.ifBlank { null })
        ) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(loading = false, message = Strings.t(Str.BindOk))
                loadBindInfo()
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    errorMessage = result.error.message,
                )
            }
        }
    }

    suspend fun unbindCenter() {
        val carId = _state.value.carId.trim()
        val imei = _state.value.imei.trim()
        val helmet = _state.value.helmet.trim()
        if (carId.isBlank()) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.EnterCarId))
            return
        }
        if (imei.isBlank() && helmet.isBlank()) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.BindNeedDeviceOrHelmet))
            return
        }
        if (!canBindOrUnbind()) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.BindNeedDeviceOrHelmet))
            return
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        when (
            val result = repository.unbind(
                carId,
                imei.ifBlank { null },
                helmet.ifBlank { null },
            )
        ) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    message = Strings.t(Str.UnbindOk),
                    bindInfo = null,
                    imei = "",
                    helmet = "",
                )
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    errorMessage = result.error.message,
                )
            }
        }
    }

    suspend fun addShelfCar() {
        val carId = _state.value.carId.trim()
        if (carId.isBlank() || carId.length !in MIN_CAR_ID_LEN..MAX_CAR_ID_LEN) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.EnterCarId))
            return
        }
        val queue = _state.value.shelfQueue
        if (queue.any { it.carId.equals(carId, ignoreCase = true) }) {
            _state.value = _state.value.copy(
                errorMessage = Strings.t(Str.AlreadyInList),
                carId = "",
            )
            return
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        val result = when (_state.value.shelfMode) {
            ShelfMode.PutOn -> repository.onlineCheck(carId)
            ShelfMode.PullOff -> repository.offlineCheck(carId)
        }
        when (result) {
            is OpsResult.Ok -> {
                // Legacy list rows use typed carNumberCode; API may omit carId / serviceName.
                val item = result.value.copy(
                    carId = carId,
                    selected = false,
                )
                val s = _state.value
                _state.value = if (s.shelfMode == ShelfMode.PutOn) {
                    s.copy(
                        loading = false,
                        putShelfQueue = listOf(item) + s.putShelfQueue,
                        carId = "",
                        message = Strings.t(Str.JoinedItem, item.carId),
                        errorMessage = null,
                    )
                } else {
                    s.copy(
                        loading = false,
                        pullShelfQueue = listOf(item) + s.pullShelfQueue,
                        carId = "",
                        message = Strings.t(Str.JoinedItem, item.carId),
                        errorMessage = null,
                    )
                }
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    carId = "",
                    errorMessage = result.error.message,
                )
            }
        }
    }

    /**
     * Submit checked items only.
     * Put-on requires [serviceId] (legacy 选择上架服务区).
     */
    suspend fun submitShelfQueue(serviceId: String? = null) {
        val queue = _state.value.shelfQueue.filter { it.selected }
        if (queue.isEmpty()) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.ShelfSelectItemsFirst))
            return
        }
        val carIds = queue.map { it.carId }
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        val result = when (_state.value.shelfMode) {
            ShelfMode.PutOn -> {
                val sid = serviceId?.trim().orEmpty()
                    .ifBlank { _state.value.shelfServiceId }
                    .ifBlank { serviceAreaIdProvider() }
                if (sid.isBlank()) {
                    _state.value = _state.value.copy(
                        loading = false,
                        errorMessage = Strings.t(Str.SelectServiceAreaFirst),
                    )
                    return
                }
                _state.value = _state.value.copy(shelfServiceId = sid)
                repository.onlineByCarList(sid, carIds)
            }
            ShelfMode.PullOff -> repository.offlineByCarList(carIds)
        }
        when (result) {
            is OpsResult.Ok -> {
                val label = if (_state.value.shelfMode == ShelfMode.PutOn) {
                    Strings.t(Str.ShelfPutOn)
                } else {
                    Strings.t(Str.ShelfPullOff)
                }
                val s = _state.value
                val selectedIds = carIds.map { it.lowercase() }.toSet()
                fun remain(list: List<ShelfCheckResult>) =
                    list.filterNot { it.carId.lowercase() in selectedIds }
                _state.value = if (s.shelfMode == ShelfMode.PutOn) {
                    s.copy(
                        loading = false,
                        putShelfQueue = remain(s.putShelfQueue),
                        message = Strings.t(Str.ShelfSubmitOk, label, carIds.size),
                    )
                } else {
                    s.copy(
                        loading = false,
                        pullShelfQueue = remain(s.pullShelfQueue),
                        message = Strings.t(Str.ShelfSubmitOk, label, carIds.size),
                    )
                }
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
        stopOverloadCheck()
        suppressCarIdAutoQuery = false
        suppressImeiAutoQuery = false
        _state.value = ProductionUiState()
    }

    /**
     * Legacy OPTION_OVERLOAD_CHECK: resolve detail, gate on [Vehicle.izHaveOverload], open pane.
     */
    suspend fun openOverloadCheck(): OpsResult<Unit> {
        val carId = _state.value.carId.trim()
        val imei = _state.value.imei.trim()
        if (carId.isBlank() && imei.isBlank()) {
            val msg = Strings.t(Str.EnterCarIdOrScan)
            _state.value = _state.value.copy(errorMessage = msg)
            return OpsResult.Err(OpsError.business("CAR", msg))
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        val target = when {
            carId.isNotBlank() -> ScanTarget.CarId(carId)
            else -> ScanTarget.Imei(imei)
        }
        return when (val detail = vehicleRepository.findByScanTarget(target)) {
            is OpsResult.Ok -> {
                val vehicle = detail.value
                if (!vehicle.izHaveOverload) {
                    val msg = Strings.t(Str.DetectOverloadUnavailable)
                    _state.value = _state.value.copy(
                        loading = false,
                        hasOverloadDevice = false,
                        carId = vehicle.carId.ifBlank { carId },
                        imei = vehicle.imei.ifBlank { imei },
                        errorMessage = msg,
                    )
                    return OpsResult.Err(OpsError.business("OVERLOAD", msg))
                }
                stopOverloadCheck()
                _state.value = _state.value.copy(
                    loading = false,
                    page = ProductionPage.Overload,
                    carId = vehicle.carId.ifBlank { carId },
                    imei = vehicle.imei.ifBlank { imei },
                    hasOverloadDevice = true,
                    overloadContacts = SaddleOverloadContact.Idle,
                    overloadCountdown = 0,
                    overloadRunning = false,
                    message = null,
                    errorMessage = null,
                )
                OpsResult.Ok(Unit)
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    errorMessage = detail.error.message,
                )
                detail
            }
        }
    }

    fun stopOverloadCheck() {
        overloadGeneration += 1
        _state.value = _state.value.copy(
            overloadRunning = false,
            overloadCountdown = 0,
        )
    }

    /**
     * Start / restart overload window: triggerTempState(stateCmd=1) then poll contacts for [overloadTimeoutSec].
     */
    suspend fun startOverloadCheck() {
        stopOverloadCheck()
        overloadMutex.withLock {
        val carId = _state.value.carId.trim()
        var imei = _state.value.imei.trim()
        if (carId.isBlank()) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.EnterCarIdFirst))
            return
        }
        if (imei.isBlank()) {
            when (val detail = vehicleRepository.getDetail(carId)) {
                is OpsResult.Ok -> {
                    imei = detail.value.imei.trim()
                    _state.value = _state.value.copy(
                        imei = imei,
                        hasOverloadDevice = detail.value.izHaveOverload,
                    )
                }
                is OpsResult.Err -> {
                    _state.value = _state.value.copy(errorMessage = detail.error.message)
                    return
                }
            }
        }
        if (imei.isBlank()) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.DetectOverloadNeedImei))
            return
        }
        val gen = ++overloadGeneration
        _state.value = _state.value.copy(
            loading = true,
            overloadRunning = true,
            overloadCountdown = overloadTimeoutSec,
            overloadContacts = SaddleOverloadContact.Idle,
            message = null,
            errorMessage = null,
        )
        when (val trigger = repository.triggerOverloadCheck(carId, imei, overloadTimeoutSec)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    message = Strings.t(Str.DetectOverloadStarted),
                )
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    overloadRunning = false,
                    overloadCountdown = 0,
                    errorMessage = trigger.error.message.ifBlank {
                        Strings.t(Str.DetectOverloadStartFail)
                    },
                )
                return
            }
        }
        while (
            gen == overloadGeneration &&
            _state.value.overloadRunning &&
            _state.value.overloadCountdown > 0
        ) {
            delay(overloadPollDelayMs)
            if (gen != overloadGeneration || !_state.value.overloadRunning) break
            when (val poll = repository.queryOverloadContact(imei)) {
                is OpsResult.Ok -> {
                    _state.value = _state.value.copy(
                        overloadContacts = poll.value,
                        overloadCountdown = (_state.value.overloadCountdown - 1).coerceAtLeast(0),
                    )
                }
                is OpsResult.Err -> {
                    _state.value = _state.value.copy(
                        overloadCountdown = (_state.value.overloadCountdown - 1).coerceAtLeast(0),
                    )
                }
            }
        }
        if (gen == overloadGeneration) {
            _state.value = _state.value.copy(
                overloadRunning = false,
                overloadCountdown = 0,
                message = _state.value.message ?: Strings.t(Str.DetectOverloadDone),
            )
        }
        }
    }

    private fun updateStep(kind: DetectStepKind, status: DetectStepStatus, detail: String = "") {
        _state.value = _state.value.copy(
            detectSteps = _state.value.detectSteps.map { step ->
                if (step.kind == kind) step.copy(status = status, detail = detail) else step
            },
            errorMessage = when (status) {
                DetectStepStatus.Failed -> detail
                DetectStepStatus.Ok -> null
                else -> _state.value.errorMessage
            },
            message = if (status == DetectStepStatus.Ok && detail.isNotBlank()) {
                detail
            } else {
                _state.value.message
            },
        )
    }

    private suspend fun refreshDetectAfterControl() {
        if (refreshAfterControlDelayMs < 0) return
        delay(refreshAfterControlDelayMs)
        queryDetectVehicle(showLoading = false)
    }

    private fun applyVehicleToDetect(vehicle: Vehicle) {
        _state.value = _state.value.copy(
            loading = false,
            carId = vehicle.carId,
            imei = vehicle.imei,
            detectVehicle = vehicle,
            hasOverloadDevice = vehicle.izHaveOverload,
            detectSwitches = switchesFromVehicle(vehicle),
            message = null,
            errorMessage = null,
        )
    }

    private fun switchesFromVehicle(vehicle: Vehicle): List<DetectSwitchState> = listOf(
        DetectSwitchState(DetectSwitchKind.Acc, Strings.t(Str.DetectSwitchAcc), on = vehicle.acc == 1),
        DetectSwitchState(DetectSwitchKind.Defend, Strings.t(Str.DetectSwitchDefend), on = vehicle.defend == 1),
        DetectSwitchState(DetectSwitchKind.Battery, Strings.t(Str.DetectSwitchBattery), on = vehicle.batteryLock == 0),
        DetectSwitchState(DetectSwitchKind.Helmet, Strings.t(Str.DetectSwitchHelmet), on = vehicle.helmetLock == 0),
        DetectSwitchState(DetectSwitchKind.Wheel, Strings.t(Str.DetectSwitchWheel), on = vehicle.backWheelLock == 0),
    )

    /**
     * Typed: carId length 1..10 or IMEI 14..18.
     * Scan: 15-digit barcode is IMEI; otherwise [ScanCodeParser].
     */
    fun resolveDetectInput(raw: String, qrHosts: List<String> = emptyList()): ScanTarget? {
        val trimmed = raw.trim()
        if (trimmed.isEmpty()) return null
        val isUrl = trimmed.startsWith("http://", ignoreCase = true) ||
            trimmed.startsWith("https://", ignoreCase = true)
        if (isUrl || trimmed.contains("carId=", ignoreCase = true) ||
            trimmed.startsWith("IMEI", ignoreCase = true)
        ) {
            return ScanCodeParser.parse(trimmed, qrHosts)
        }
        if (trimmed.length == 15 && trimmed.all { it.isDigit() }) {
            return ScanTarget.Imei(trimmed)
        }
        if (trimmed.length in MIN_CAR_ID_LEN..MAX_CAR_ID_LEN) {
            return ScanTarget.CarId(trimmed)
        }
        if (trimmed.length in MIN_IMEI_LEN..MAX_IMEI_LEN) {
            return ScanTarget.Imei(trimmed)
        }
        return null
    }

    companion object {
        /** Legacy AppConfig.isValidVehicleNum. */
        const val MIN_CAR_ID_LEN = 1
        const val MAX_CAR_ID_LEN = 10
        /** Legacy AppConfig.isValidIMEINum. */
        const val MIN_IMEI_LEN = 14
        const val MAX_IMEI_LEN = 18
        /** Legacy checkClickEnable uses exact 15 for deviceId length. */
        const val BIND_IMEI_EXACT_LEN = 15
        const val BIND_HELMET_MAX_LEN = 12

        private val imeiScanRegex = Regex("""^IMEI[:：]\s*([0-9]{14,18})""", RegexOption.IGNORE_CASE)

        fun extractImeiFromScan(raw: String): String? =
            imeiScanRegex.find(raw.trim())?.groupValues?.getOrNull(1)?.trim()?.takeIf { it.isNotEmpty() }

        /** Legacy: `MAC:xx:xx SN:...` → strip colons between MAC and SN. */
        fun parseHelmetMac(raw: String): String? {
            val text = raw.trim()
            val macIdx = text.indexOf("MAC", ignoreCase = true)
            if (macIdx < 0) return null
            val afterMac = text.substring(macIdx + 3).trimStart(':', ' ', '=')
            val snIdx = afterMac.indexOf(" SN", ignoreCase = true)
            val macPart = if (snIdx >= 0) afterMac.substring(0, snIdx) else afterMac.substringBefore(' ')
            val cleaned = macPart.replace(":", "").replace("-", "").trim()
            return cleaned.takeIf { it.isNotEmpty() }
        }
    }
}
