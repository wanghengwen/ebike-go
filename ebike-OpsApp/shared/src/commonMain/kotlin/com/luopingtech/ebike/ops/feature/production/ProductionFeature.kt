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
    val detectChannel: DetectChannelTab = DetectChannelTab.Network,
    val detectSteps: List<DetectStepState> = defaultDetectSteps(),
    val detectSwitches: List<DetectSwitchState> = defaultDetectSwitches(),
    val bleAvailable: Boolean = true,
    val shelfMode: ShelfMode = ShelfMode.PutOn,
    val shelfServiceId: String = "",
    val shelfQueue: List<ShelfCheckResult> = emptyList(),
    val overloadContacts: SaddleOverloadContact = SaddleOverloadContact.Idle,
    val overloadCountdown: Int = 0,
    val overloadRunning: Boolean = false,
    val hasOverloadDevice: Boolean? = null,
    val message: String? = null,
    val errorMessage: String? = null,
) {
    val controlChannel: ControlChannel
        get() = when (detectChannel) {
            DetectChannelTab.Network -> ControlChannel.NetworkOnly
            DetectChannelTab.Bluetooth -> ControlChannel.BleOnly
        }
}

private fun defaultDetectSteps(): List<DetectStepState> = listOf(
    DetectStepState(DetectStepKind.Ring, Strings.t(Str.DetectFindSound)),
    DetectStepState(DetectStepKind.Unlock, Strings.t(Str.ScanUnlock)),
    DetectStepState(DetectStepKind.Lock, Strings.t(Str.ScanLock)),
    DetectStepState(DetectStepKind.OpenBox, Strings.t(Str.OpenBatteryBox)),
    DetectStepState(DetectStepKind.CloseBox, Strings.t(Str.DetectCloseBox)),
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
) {
    private val _state = MutableStateFlow(ProductionUiState(bleAvailable = bleAvailableProvider()))
    val state: StateFlow<ProductionUiState> = _state.asStateFlow()
    private val overloadMutex = Mutex()
    private var overloadGeneration: Int = 0

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
            detectSteps = defaultDetectSteps(),
            detectSwitches = defaultDetectSwitches(),
            bleAvailable = bleAvailableProvider(),
            message = null,
            errorMessage = null,
        )
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

    fun setShelfServiceId(value: String) {
        _state.value = _state.value.copy(shelfServiceId = value, errorMessage = null)
    }

    fun removeShelfItem(carId: String) {
        _state.value = _state.value.copy(
            shelfQueue = _state.value.shelfQueue.filterNot { it.carId == carId },
        )
    }

    suspend fun runDetectStep(kind: DetectStepKind) {
        val carId = _state.value.carId.trim()
        if (carId.isBlank()) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.EnterCarIdFirst))
            return
        }
        val channel = _state.value.controlChannel
        if (channel == ControlChannel.BleOnly && !bleAvailableProvider()) {
            updateStep(kind, DetectStepStatus.Failed, Strings.t(Str.DetectBleUnavailableHint))
            return
        }
        updateStep(kind, DetectStepStatus.Running)
        val action = when (kind) {
            DetectStepKind.Unlock -> VehicleAction.Unlock
            DetectStepKind.Lock -> VehicleAction.Lock
            DetectStepKind.Ring -> VehicleAction.Ring
            DetectStepKind.OpenBox -> VehicleAction.OpenBatteryBox
            DetectStepKind.CloseBox -> VehicleAction.CloseBatteryBox
        }
        val channelLabel = when (_state.value.detectChannel) {
            DetectChannelTab.Network -> Strings.t(Str.DetectChannelNetwork)
            DetectChannelTab.Bluetooth -> Strings.t(Str.DetectChannelBle)
        }
        when (val result = control.execute(carId, action, channel, imei = _state.value.imei.trim())) {
            is OpsResult.Ok -> updateStep(kind, DetectStepStatus.Ok, "OK · $channelLabel")
            is OpsResult.Err -> updateStep(kind, DetectStepStatus.Failed, result.error.message)
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
        val carId = _state.value.carId.trim()
        if (carId.isBlank()) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.EnterCarIdFirst))
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
                vehicleId = carId,
                action = action,
                channel = channel,
                imei = _state.value.imei.trim(),
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

    suspend fun loadBindInfo() {
        val carId = _state.value.carId.trim()
        if (carId.isBlank()) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.EnterCarId))
            return
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        when (val result = repository.getBind(carId)) {
            is OpsResult.Ok -> {
                val info = result.value
                _state.value = _state.value.copy(
                    loading = false,
                    bindInfo = info,
                    imei = info.imei.ifBlank { _state.value.imei },
                    helmet = info.helmet.ifBlank { _state.value.helmet },
                    message = if (info.imei.isBlank()) {
                        Strings.t(Str.NotBoundImei)
                    } else {
                        Strings.t(Str.BoundImei, info.imei)
                    },
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

    suspend fun bindCenter() {
        val carId = _state.value.carId.trim()
        val imei = _state.value.imei.trim()
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        when (
            val result = repository.bind(carId, imei, _state.value.helmet.trim().ifBlank { null })
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
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        when (
            val result = repository.unbind(
                carId,
                _state.value.imei.trim().ifBlank { null },
                _state.value.helmet.trim().ifBlank { null },
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
        if (carId.isBlank()) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.EnterCarId))
            return
        }
        if (_state.value.shelfQueue.any { it.carId.equals(carId, ignoreCase = true) }) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.AlreadyInList), carId = "")
            return
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        val result = when (_state.value.shelfMode) {
            ShelfMode.PutOn -> repository.onlineCheck(carId)
            ShelfMode.PullOff -> repository.offlineCheck(carId)
        }
        when (result) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    shelfQueue = _state.value.shelfQueue + result.value,
                    carId = "",
                    message = Strings.t(Str.JoinedItem, result.value.carId),
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

    suspend fun submitShelfQueue() {
        val queue = _state.value.shelfQueue
        val carIds = queue.map { it.carId }
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        val result = when (_state.value.shelfMode) {
            ShelfMode.PutOn -> {
                val serviceId = _state.value.shelfServiceId.ifBlank { serviceAreaIdProvider() }
                repository.onlineByCarList(serviceId, carIds)
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
                _state.value = _state.value.copy(
                    loading = false,
                    page = ProductionPage.Shelves,
                    shelfQueue = emptyList(),
                    message = Strings.t(Str.ShelfSubmitOk, label, carIds.size),
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

    fun clear() {
        stopOverloadCheck()
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
            errorMessage = if (status == DetectStepStatus.Failed) detail else null,
            message = if (status == DetectStepStatus.Ok) "${kind.name} OK" else _state.value.message,
        )
    }
}
