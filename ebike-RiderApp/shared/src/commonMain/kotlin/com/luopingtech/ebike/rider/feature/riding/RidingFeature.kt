package com.luopingtech.ebike.rider.feature.riding

import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.core.i18n.Strings
import com.luopingtech.ebike.rider.core.logging.NoOpLogger
import com.luopingtech.ebike.rider.core.logging.RiderLogger
import com.luopingtech.ebike.rider.core.network.ApiOutcome
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.core.time.nowEpochMillis
import com.luopingtech.ebike.rider.data.fence.FenceRemote
import com.luopingtech.ebike.rider.data.riding.ApplyReturnRemote
import com.luopingtech.ebike.rider.data.riding.BleAccessoryReport
import com.luopingtech.ebike.rider.data.riding.BleRideRemote
import com.luopingtech.ebike.rider.data.riding.RideActor
import com.luopingtech.ebike.rider.data.riding.RidingParsers
import com.luopingtech.ebike.rider.data.riding.RidingRemote
import com.luopingtech.ebike.rider.domain.ble.BleFrame
import com.luopingtech.ebike.rider.domain.ble.BleVoice
import com.luopingtech.ebike.rider.domain.model.FenceKind
import com.luopingtech.ebike.rider.domain.model.FencePolygon
import com.luopingtech.ebike.rider.domain.model.GeoLatLng
import com.luopingtech.ebike.rider.domain.riding.BackCarConfig
import com.luopingtech.ebike.rider.domain.riding.NearParking
import com.luopingtech.ebike.rider.domain.riding.ReturnByNetType
import com.luopingtech.ebike.rider.domain.riding.ReturnDecision
import com.luopingtech.ebike.rider.domain.riding.ReturnFlow
import com.luopingtech.ebike.rider.domain.riding.ReturnPermissionResult
import com.luopingtech.ebike.rider.domain.riding.RideEvent
import com.luopingtech.ebike.rider.domain.riding.RideInfo
import com.luopingtech.ebike.rider.domain.riding.RidePhase
import com.luopingtech.ebike.rider.domain.riding.RideSession
import com.luopingtech.ebike.rider.domain.riding.RideStateMachine
import com.luopingtech.ebike.rider.domain.riding.RidingFenceTip
import com.luopingtech.ebike.rider.domain.riding.RidingFenceTips
import com.luopingtech.ebike.rider.domain.riding.SettlementSummary
import com.luopingtech.ebike.rider.domain.riding.VehicleDetail
import com.luopingtech.ebike.rider.domain.tracking.TrackBufferEvent
import com.luopingtech.ebike.rider.domain.tracking.TrackPointBuffer
import com.luopingtech.ebike.rider.feature.ble.BleSessionFeature
import com.luopingtech.ebike.rider.platform.GeoPoint
import com.luopingtech.ebike.rider.platform.LocationTracker
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock

/** 实际走通的开锁通道，进 Debug 屏也进 [RidingUiState.unlockChannel]。 */
enum class UnlockChannel { Ble, Network }

/**
 * 通道策略。
 *
 * 与 UniApp **有意不同**：小程序没有稳定的常驻蓝牙，所以 `unlockWithBleFallback` 是
 * 网络优先、只在车机离线（17012）时才降级蓝牙。原生端反过来 —— 蓝牙握手通常比一次
 * 公网往返快，而且地库 / 电梯口没信号时只有蓝牙能用。所以默认 [BlePreferred]，
 * 保留 [NetworkPreferred] 给需要严格对齐小程序行为的租户。
 */
enum class UnlockPolicy { BlePreferred, NetworkPreferred, BleOnly }

/** 需要界面弹层 / 对话框的时刻。 */
sealed interface RidePrompt {
    /** 正常还车前的文明停车提醒。 */
    data object Civilization : RidePrompt

    /** 可还但要收调度费。 */
    data class Penalty(val returnTypeCode: Int, val penaltyFen: Int) : RidePrompt

    /** 不可还，解释原因。 */
    data class Blocked(val returnTypeCode: Int, val penaltyFen: Int) : RidePrompt

    /** 跳定制还车引导页。 */
    data class Guide(val pageType: Int, val returnTypeCode: Int) : RidePrompt

    /** 临停被头盔未归位拦下。 */
    data object HelmetNotReturned : RidePrompt

    /** 超载：[powerOff] 为真表示已断电。 */
    data class Overload(val powerOff: Boolean) : RidePrompt

    /** 平台已断电，问用户要不要临时恢复动力 [minutes] 分钟。 */
    data class RecoverPower(val minutes: Int) : RidePrompt

    /** 余额不足（15030），引导充值。 */
    data class InsufficientBalance(val message: String) : RidePrompt
}

/** 一次性提示。[nonce] 让相同文案连续出现两次也能被界面识别成两次。 */
data class RideMessage(
    val text: String,
    val isError: Boolean = false,
    val nonce: Long = 0L,
)

data class RidingUiState(
    val session: RideSession = RideSession.Idle,
    val vehicle: VehicleDetail? = null,
    val rideInfo: RideInfo? = null,
    /** null = 还没拉过。空配置和「未加载」必须分得开，否则每次进还车都会重拉。 */
    val backCarConfig: BackCarConfig? = null,
    val fenceTip: RidingFenceTip? = null,
    val fences: List<FencePolygon> = emptyList(),
    val nearParking: NearParking? = null,
    /** 本地攒出来的骑行折线，喂给骑行地图与行程回放。 */
    val track: List<GeoLatLng> = emptyList(),
    val myLocation: GeoPoint? = null,
    val busy: Boolean = false,
    val busyLabel: Str? = null,
    /** 本地秒表。服务端 `rideTime` 每 8 秒才回一次，中间靠它走。 */
    val elapsedSeconds: Long = 0L,
    val unlockChannel: UnlockChannel? = null,
    val prompt: RidePrompt? = null,
    val settlement: SettlementSummary? = null,
    val message: RideMessage? = null,
) {
    val phase: RidePhase get() = session.phase
    val carId: String get() = session.carId
    val config: BackCarConfig get() = backCarConfig ?: BackCarConfig()
}

/**
 * 骑行主循环编排。
 *
 * 职责边界：状态转移全部交给 [RideStateMachine]（纯函数、可单测），这里只负责
 * 「跑 IO、把结果翻译成事件、落盘、驱动前台定位」。任何一处想直接改 `_state.value.session.phase`
 * 都是 bug —— 走 [dispatch]。
 */
class RidingFeature(
    private val riding: RidingRemote,
    private val bleRide: BleRideRemote,
    private val applyReturn: ApplyReturnRemote,
    private val fences: FenceRemote,
    private val bleSession: BleSessionFeature,
    private val locationTracker: LocationTracker,
    private val store: RideSessionStore,
    private val logger: RiderLogger = NoOpLogger,
    private val unlockPolicy: UnlockPolicy = UnlockPolicy.BlePreferred,
    /** 是否有可用蓝牙硬件。false 时策略自动退化成纯网络。 */
    private val bleAvailable: () -> Boolean = { true },
) {
    private val _state = MutableStateFlow(RidingUiState())
    val state: StateFlow<RidingUiState> = _state.asStateFlow()

    /** 开锁 / 还车这类改状态机相位的动作串行化，避免双击造出两张单。 */
    private val commandMutex = Mutex()

    /**
     * 抽稀器只负责「这一批要不要收进折线」；[committedTrack] 才是完整轨迹。
     * [TrackPointBuffer] 是 OpsApp 的只读拷贝，不给它加方法。
     */
    private val trackBuffer = TrackPointBuffer()
    private val committedTrack = ArrayList<GeoLatLng>()
    private var started = false
    private var messageNonce = 0L

    // ── 生命周期 ────────────────────────────────────────────────────────────

    fun start(scope: CoroutineScope) {
        if (started) return
        started = true
        restore()
        scope.launch { pollLoop() }
        scope.launch { tickLoop() }
        scope.launch { locationLoop() }
        // 本地快照可能是 Idle，但服务端仍有进行中订单 —— 对齐 UniApp 首页 getPersonInfo。
        scope.launch { resumeActiveRideIfNeeded() }
    }

    /** 冷启动恢复。相位落在骑行域时立刻拉一次服务端真值。 */
    fun restore() {
        val restored = RideStateMachine.sanitizeRestored(store.load())
        _state.value = _state.value.copy(
            session = restored,
            elapsedSeconds = elapsedFrom(restored),
        )
        syncForegroundLocation(restored)
    }

    /**
     * 首页 / 冷启动：若本地不在骑行相位，向服务端确认是否仍有进行中订单。
     * 对齐 UniApp `home.onShow` → `getPersonInfo` ridingState∈{4,5,6} → `reLaunch` 骑行页。
     */
    suspend fun resumeActiveRideIfNeeded() {
        val phase = _state.value.phase
        if (phase != RidePhase.Idle && phase != RidePhase.Scanning) return
        refreshRideInfo()
    }

    private suspend fun pollLoop() {
        while (true) {
            val session = _state.value.session
            // Idle 也间歇探活：用户杀进程后仅留服务端订单时，首页要能抬回骑行。
            if (session.needsForegroundLocation ||
                session.phase == RidePhase.Settling ||
                session.phase == RidePhase.Idle
            ) {
                refreshRideInfo()
                // 申诉通过后回骑行页：自动再跑一次还车（旧版 `autoLock`）。
                // 只在轮询里做，不放进 refreshRideInfo —— 那个方法会在 commandMutex 内被调用。
                if (_state.value.session.isOnTrip && store.consumeAutoReturn()) {
                    beginReturn()
                }
            }
            delay(POLL_INTERVAL_MS)
        }
    }

    private suspend fun tickLoop() {
        while (true) {
            delay(1_000L)
            val current = _state.value
            // 临停不计时（对齐旧版：临停时长不进骑行时长）。
            if (current.phase == RidePhase.Riding) {
                _state.value = current.copy(elapsedSeconds = current.elapsedSeconds + 1)
            }
        }
    }

    /**
     * 骑行期间收集定位：喂 [TrackPointBuffer] 做抽稀，攒出本地折线。
     * 这里**不**上报轨迹 —— 服务端的骑行轨迹来自车机 GPS，前端再传一份只会打架。
     */
    private suspend fun locationLoop() {
        val flow = runCatching { locationTracker.track() }.getOrNull() ?: return
        runCatching {
            flow.collect { point ->
                _state.value = _state.value.copy(myLocation = point)
                if (!_state.value.session.isOnTrip) return@collect
                when (val event = trackBuffer.onTick(point)) {
                    is TrackBufferEvent.Flush -> {
                        // Flush = 这批点已确定进轨迹；转存后让抽稀器从头再攒。
                        committedTrack += event.points.map { GeoLatLng(it.latitude, it.longitude) }
                        trackBuffer.clearAfterSuccess()
                        _state.value = _state.value.copy(track = committedTrack.toList())
                    }
                    TrackBufferEvent.None -> _state.value = _state.value.copy(
                        track = committedTrack + trackBuffer.snapshot().map {
                            GeoLatLng(it.latitude, it.longitude)
                        },
                    )
                }
            }
        }.onFailure { logger.w(TAG, "location stream ended: ${it.message}") }
    }

    // ── 车辆选择 ────────────────────────────────────────────────────────────

    fun openScan() = dispatch(RideEvent.OpenScan)

    /** 放弃当前未开锁的车。骑行中调用无效（状态机会挡住）。 */
    fun cancelSelection() {
        dispatch(RideEvent.Reset)
        _state.value = _state.value.copy(vehicle = null, unlockChannel = null)
    }

    /**
     * 扫码 / 手输之后拉车辆详情，进确认页。
     *
     * `submitScan` 是埋点，旧版 `void` 掉不等结果 —— 这里也不让它拖慢界面。
     */
    suspend fun selectVehicle(carId: String, imei: String = "", manual: Boolean = false) {
        val trimmed = carId.trim()
        if (trimmed.isBlank()) {
            emitMessage(Strings.t(Str.RideCarIdMissing), isError = true)
            return
        }
        dispatch(if (manual) RideEvent.ManualId(trimmed) else RideEvent.Scan(trimmed, imei))
        withBusy(Str.Loading) {
            riding.submitScan(trimmed)
            when (val detail = riding.carInfo(trimmed)) {
                is RiderResult.Ok -> {
                    val vehicle = detail.value
                    _state.value = _state.value.copy(vehicle = vehicle)
                    dispatch(
                        RideEvent.VehicleLoaded(
                            carId = vehicle.carId,
                            imei = vehicle.imei.ifBlank { imei },
                        ),
                    )
                    if (!vehicle.available) {
                        emitMessage(Strings.t(Str.RideVehicleUnavailable), isError = true)
                    }
                }
                is RiderResult.Err -> {
                    _state.value = _state.value.copy(vehicle = null)
                    emitMessage(detail.error.message, isError = true)
                }
            }
        }
    }

    // ── 开锁 ────────────────────────────────────────────────────────────────

    /**
     * 确认开锁。
     *
     * 时序照搬 UniApp `unlockByBle`：**静音**下发 0x2c，成功后先向服务端建单，
     * 建单成功才播 `PLAY_VOICE(START)`。顺序反了会出现「车响了但没订单」——
     * 用户以为骑上了，实际是白骑，账也对不上。建单失败必须回滚上锁。
     */
    suspend fun confirmUnlock() = commandMutex.withLock {
        val session = _state.value.session
        if (session.phase != RidePhase.Confirming || !session.hasVehicle) {
            logger.w(TAG, "confirmUnlock ignored: phase=${session.phase} carId=${session.carId}")
            emitMessage(Strings.t(Str.RideUnlockFailed), isError = true)
            return@withLock
        }
        val vehicle = _state.value.vehicle
        if (vehicle == null) {
            emitMessage(Strings.t(Str.RideVehicleUnavailable), isError = true)
            return@withLock
        }
        if (!vehicle.available) {
            emitMessage(Strings.t(Str.RideVehicleUnavailable), isError = true)
            return@withLock
        }
        if (vehicle.outOfServiceArea) {
            emitMessage(Strings.t(Str.RideOutOfServiceCannotUnlock), isError = true)
            return@withLock
        }
        // 开锁前强制刷新定位，Confirming 相位不启前台追踪，不能只信缓存 myLocation。
        val actor = actorWithFreshLocation()
        if (actor == null) {
            emitMessage(Strings.t(Str.RideNeedLocation), isError = true)
            return@withLock
        }

        dispatch(RideEvent.Confirm)
        val imei = session.imei.ifBlank { vehicle.imei }
        val carId = session.carId

        withBusy(Str.RideUnlocking) {
            val order = runUnlock(carId = carId, imei = imei, actor = actor, vehicle = vehicle)
            when (order) {
                is UnlockOutcome.Success -> {
                    _state.value = _state.value.copy(unlockChannel = order.channel)
                    dispatch(
                        RideEvent.UnlockSuccess(
                            orderId = order.orderId,
                            at = nowEpochMillis(),
                            usedNetworkFallback = order.channel == UnlockChannel.Network,
                        ),
                    )
                    trackBuffer.reset()
                    committedTrack.clear()
                    _state.value = _state.value.copy(elapsedSeconds = 0L, track = emptyList())
                    emitMessage(Strings.t(Str.RideUnlockOk))
                    refreshRideInfo()
                    loadBackCarConfig()
                }
                is UnlockOutcome.InsufficientBalance -> {
                    dispatch(RideEvent.UnlockFail(order.message))
                    _state.value = _state.value.copy(
                        prompt = RidePrompt.InsufficientBalance(order.message),
                    )
                }
                is UnlockOutcome.Failure -> {
                    dispatch(RideEvent.UnlockFail(order.message))
                    emitMessage(order.message, isError = true)
                }
            }
        }
    }

    private sealed interface UnlockOutcome {
        data class Success(val orderId: String, val channel: UnlockChannel) : UnlockOutcome
        data class Failure(val message: String) : UnlockOutcome
        data class InsufficientBalance(val message: String) : UnlockOutcome
    }

    private suspend fun runUnlock(
        carId: String,
        imei: String,
        actor: RideActor,
        vehicle: VehicleDetail?,
    ): UnlockOutcome {
        // 生产空 imei 不要拼 demo IMEI：会先对假地址扫蓝牙，表现为「指令没下发」。
        val effectiveImei = imei.trim()
        val bleUsable = bleAvailable() && effectiveImei.isNotBlank()
        // 车机离线时网络指令必然超时，直接顶到蓝牙优先。
        val preferBle = when (unlockPolicy) {
            UnlockPolicy.BleOnly -> true
            UnlockPolicy.BlePreferred -> true
            UnlockPolicy.NetworkPreferred -> vehicle?.disconnected == true
        }

        if (preferBle) {
            if (!bleUsable) {
                if (unlockPolicy == UnlockPolicy.BleOnly) {
                    return UnlockOutcome.Failure(Strings.t(Str.BleUnavailable))
                }
                logger.w(TAG, "preferBle but no usable imei/ble, falling back to network")
                return unlockByNetwork(carId, actor, allowBleRetry = false, imei = effectiveImei)
            }
            when (val bleAttempt = unlockByBle(carId, effectiveImei, actor)) {
                is UnlockOutcome.Success -> return bleAttempt
                is UnlockOutcome.InsufficientBalance -> return bleAttempt
                is UnlockOutcome.Failure -> {
                    if (unlockPolicy == UnlockPolicy.BleOnly) return bleAttempt
                    logger.w(TAG, "ble unlock failed, falling back to network: ${bleAttempt.message}")
                }
            }
            return unlockByNetwork(carId, actor, allowBleRetry = false, imei = effectiveImei)
        }

        return unlockByNetwork(carId, actor, allowBleRetry = bleUsable, imei = effectiveImei)
    }

    private suspend fun unlockByBle(
        carId: String,
        imei: String,
        actor: RideActor,
    ): UnlockOutcome {
        val permission = when (val result = bleRide.ridePermission(carId, actor)) {
            is RiderResult.Err -> return UnlockOutcome.Failure(result.error.message)
            is RiderResult.Ok -> result.value
        }
        if (!permission.success) {
            if (permission.isInsufficientBalance) {
                return UnlockOutcome.InsufficientBalance(
                    permission.messageOr(Str.RideBalanceInsufficient),
                )
            }
            return UnlockOutcome.Failure(permission.messageOr(Str.RideUnlockFailed))
        }

        // 静音开锁：语音留到建单成功之后。
        when (val written = bleSession.unlock(imei, mute = true)) {
            is RiderResult.Err -> return UnlockOutcome.Failure(written.error.message)
            is RiderResult.Ok -> Unit
        }

        val reported = bleRide.rideReport(carId, actor)
        val reportedOk = (reported as? RiderResult.Ok)?.value?.takeIf { it.success }
        if (reportedOk == null) {
            // 回滚：车已经开了但服务端没建单，必须锁回去，否则用户白骑、账也对不上。
            logger.w(TAG, "ble ride report failed, locking back")
            bleSession.lock(imei, mute = true)
            return UnlockOutcome.Failure(reported.messageOr(Str.RideUnlockFailed))
        }

        // 建单成功之后才响：顺序反了会出现「车响了但没订单」。
        bleSession.playVoice(imei, BleVoice.START, BleVoice.DEFAULT_VOLUME)
        return UnlockOutcome.Success(
            orderId = RidingParsers.orderIdOf(reportedOk.data),
            channel = UnlockChannel.Ble,
        )
    }

    private suspend fun unlockByNetwork(
        carId: String,
        actor: RideActor,
        allowBleRetry: Boolean,
        imei: String,
    ): UnlockOutcome {
        val outcome = when (val result = riding.networkRide(carId, actor)) {
            is RiderResult.Err -> return UnlockOutcome.Failure(result.error.message)
            is RiderResult.Ok -> result.value
        }
        if (outcome.success) {
            return UnlockOutcome.Success(
                orderId = RidingParsers.orderIdOf(outcome.data),
                channel = UnlockChannel.Network,
            )
        }
        if (outcome.isInsufficientBalance) {
            return UnlockOutcome.InsufficientBalance(outcome.messageOr(Str.RideBalanceInsufficient))
        }
        if (allowBleRetry && outcome.isDeviceOffline) {
            logger.w(TAG, "network ride 17012, retry over ble")
            return unlockByBle(carId, imei, actor)
        }
        return UnlockOutcome.Failure(outcome.messageOr(Str.RideUnlockFailed))
    }

    // ── 临停 / 恢复 ─────────────────────────────────────────────────────────

    suspend fun toggleTempLock() = commandMutex.withLock {
        when (_state.value.session.phase) {
            RidePhase.Riding -> tempLock()
            RidePhase.TempLocked -> resumeRide()
            else -> Unit
        }
    }

    private suspend fun tempLock() {
        val session = _state.value.session
        val actor = actorOrPinOnly()
        withBusy(Str.RideTempLocking) {
            // 头盔没归位就别锁，否则用户回来发现头盔丢了还得赔。
            val match = riding.partMatchByTempParking(
                carId = session.carId,
                orderId = session.orderId,
                serviceId = store.serviceAreaId,
                userPin = actor.userPin,
            )
            if (match is RiderResult.Ok && match.value.helmetNotReturned) {
                _state.value = _state.value.copy(prompt = RidePrompt.HelmetNotReturned)
                return@withBusy
            }

            val net = riding.tempPark(session.carId, actor.userPin)
            if (net.succeeded) {
                dispatch(RideEvent.TempLock(nowEpochMillis()))
                emitMessage(Strings.t(Str.RideTempLockOk))
                refreshRideInfo()
                return@withBusy
            }
            if (!net.isDeviceOffline || session.imei.isBlank()) {
                emitMessage(net.messageOr(Str.RideTempLockFailed), isError = true)
                return@withBusy
            }
            // 车机离线：蓝牙临停锁 + 上报。
            when (bleSession.tempLock(session.imei)) {
                is RiderResult.Err -> emitMessage(Strings.t(Str.RideTempLockFailed), isError = true)
                is RiderResult.Ok -> {
                    val reported = bleRide.tempParkReport(session.carId, actor)
                    if (reported.succeeded) {
                        dispatch(RideEvent.TempLock(nowEpochMillis()))
                        emitMessage(Strings.t(Str.RideTempLockOk))
                        refreshRideInfo()
                    } else {
                        emitMessage(reported.messageOr(Str.RideTempLockFailed), isError = true)
                    }
                }
            }
        }
    }

    private suspend fun resumeRide() {
        val session = _state.value.session
        val actor = actorOrPinOnly()
        withBusy(Str.RideResuming) {
            val net = riding.endTempPark(session.carId, actor.userPin)
            if (net.succeeded) {
                dispatch(RideEvent.Resume)
                refreshRideInfo()
                return@withBusy
            }
            if (!net.isDeviceOffline || session.imei.isBlank()) {
                emitMessage(net.messageOr(Str.RideResumeFailed), isError = true)
                return@withBusy
            }
            when (bleSession.unlock(session.imei, mute = true)) {
                is RiderResult.Err -> emitMessage(Strings.t(Str.RideResumeFailed), isError = true)
                is RiderResult.Ok -> {
                    val reported = bleRide.endTempParkReport(session.carId, actor)
                    if (reported.succeeded) {
                        bleSession.playVoice(session.imei, BleVoice.START, BleVoice.DEFAULT_VOLUME)
                        dispatch(RideEvent.Resume)
                        refreshRideInfo()
                    } else {
                        emitMessage(reported.messageOr(Str.RideResumeFailed), isError = true)
                    }
                }
            }
        }
    }

    // ── 还车 ────────────────────────────────────────────────────────────────

    /**
     * 第一步：只跑 `returnPermission`，由 [ReturnDecision.flow] 决定界面走哪条路。
     * 旧版把判定和提交揉在一起，结果「文明提醒」弹层出现时车已经锁了。
     */
    suspend fun beginReturn() = commandMutex.withLock {
        val session = _state.value.session
        if (!session.isOnTrip) return@withLock
        if (session.carId.isBlank()) {
            emitMessage(Strings.t(Str.RideCarIdMissing), isError = true)
            return@withLock
        }
        val actor = actorWithLocation()
        if (actor == null) {
            emitMessage(Strings.t(Str.RideNeedLocation), isError = true)
            return@withLock
        }
        if (_state.value.backCarConfig == null) loadBackCarConfig()

        dispatch(RideEvent.StartReturn)
        withBusy(Str.RideReturning) {
            val preferBle = preferBleChannel()
            var result = riding.returnPermission(
                carId = session.carId,
                orderId = session.orderId,
                actor = actor,
                bleChannel = preferBle,
            )
            // permission 自己回 17012 时，旧版换蓝牙通道再问一次。
            if (!result.succeeded && result.isDeviceOffline && !preferBle) {
                result = riding.returnPermission(
                    carId = session.carId,
                    orderId = session.orderId,
                    actor = actor,
                    bleChannel = true,
                )
            }
            val outcome = (result as? RiderResult.Ok)?.value?.takeIf { it.success }
            if (outcome == null) {
                val message = result.messageOr(Str.RideReturnFailed)
                dispatch(RideEvent.Error(message))
                emitMessage(message, isError = true)
                return@withBusy
            }
            val permission = RidingParsers.returnPermission(outcome.data)
            dispatch(RideEvent.ReturnPermission(permission.returnTypeCode))
            applyReturnFlow(permission, actor)
        }
    }

    private fun preferBleChannel(): Boolean =
        _state.value.vehicle?.disconnected == true || unlockPolicy == UnlockPolicy.BleOnly

    private suspend fun applyReturnFlow(permission: ReturnPermissionResult, actor: RideActor) {
        val flow = ReturnDecision.flow(
            canReturn = permission.canReturn,
            returnTypeCode = permission.returnTypeCode,
            civilizationRemind = _state.value.config.civilizationRemind,
        )
        when (flow) {
            ReturnFlow.Normal -> commitReturn(forcePenalty = false, actor = actor)
            ReturnFlow.Civilization ->
                _state.value = _state.value.copy(prompt = RidePrompt.Civilization)
            ReturnFlow.PenaltySheet -> _state.value = _state.value.copy(
                prompt = RidePrompt.Penalty(permission.returnTypeCode, permission.penaltyFen),
            )
            ReturnFlow.BlockSheet -> _state.value = _state.value.copy(
                prompt = RidePrompt.Blocked(permission.returnTypeCode, permission.penaltyFen),
            )
            ReturnFlow.Guide -> {
                val pageType = ReturnDecision.guidePageType(permission.returnTypeCode)
                if (pageType == null) {
                    dispatch(RideEvent.Error(""))
                } else {
                    _state.value = _state.value.copy(
                        prompt = RidePrompt.Guide(pageType, permission.returnTypeCode),
                    )
                }
            }
            // 旧版在未知码上静默返回骑行页，保留这个行为但把相位收回去。
            ReturnFlow.Silent -> dispatch(RideEvent.Error(""))
        }
    }

    /**
     * 用户在弹层上点了确认（文明提醒 / 认罚）。
     * [forcePenalty] 对应 `returnByNet` 的 `returnType=1`。
     */
    suspend fun confirmReturn(forcePenalty: Boolean) = commandMutex.withLock {
        val actor = actorWithLocation() ?: actorOrPinOnly()
        _state.value = _state.value.copy(prompt = null)
        if (_state.value.session.phase != RidePhase.Returning) {
            dispatch(RideEvent.StartReturn)
        }
        withBusy(Str.RideReturning) { commitReturn(forcePenalty, actor) }
    }

    /**
     * 提交还车。失败矩阵逐条对齐 UniApp `commitReturn`：
     *
     * - 成功 → 结费
     * - `17012*` 车机离线 → 蓝牙锁车 + 上报
     * - `15002` 临时冻结单 → 也进结费（界面显示待支付倒计时）
     * - `15009` 订单已结束 → 当成功
     * - 业务码为空 / `0` → 围栏又拒了，带着 data 重开弹层矩阵
     * - 其它 → 提示原文，退回骑行
     */
    private suspend fun commitReturn(forcePenalty: Boolean, actor: RideActor) {
        val session = _state.value.session
        if (preferBleChannel() && session.imei.isNotBlank()) {
            returnByBle(actor)
            return
        }

        val outcome = when (
            val result = riding.returnByNet(
                carId = session.carId,
                orderId = session.orderId,
                actor = actor,
                returnType = if (forcePenalty) {
                    ReturnByNetType.ACCEPT_PENALTY
                } else {
                    ReturnByNetType.NORMAL
                },
            )
        ) {
            is RiderResult.Err -> {
                dispatch(RideEvent.Error(result.error.message))
                emitMessage(result.error.message, isError = true)
                return
            }
            is RiderResult.Ok -> result.value
        }
        if (outcome.success) {
            finishReturn(orderId = RidingParsers.orderIdOf(outcome.data))
            return
        }
        when {
            outcome.isDeviceOffline && session.imei.isNotBlank() -> returnByBle(actor)

            // 冻结单：钱没结清但车已还，界面进结费屏走待支付。
            outcome.code.trim() == CODE_TEMP_FROZEN -> finishReturn(orderId = session.orderId)

            outcome.code.trim() == CODE_ORDER_FINISHED -> finishReturn(orderId = session.orderId)

            outcome.code.isBlank() || outcome.code.trim() == "0" -> {
                // 围栏在提交这一刻又变了（用户挪了车 / 站点满了）—— 重跑一次判定矩阵。
                val permission = RidingParsers.returnPermission(outcome.data)
                dispatch(RideEvent.ReturnPermission(permission.returnTypeCode))
                applyReturnFlow(permission, actor)
            }

            else -> {
                dispatch(RideEvent.Error(outcome.messageOr(Str.RideReturnFailed)))
                emitMessage(outcome.messageOr(Str.RideReturnFailed), isError = true)
            }
        }
    }

    private suspend fun returnByBle(actor: RideActor) {
        val session = _state.value.session
        if (session.imei.isBlank()) {
            dispatch(RideEvent.Error(Strings.t(Str.RideReturnFailed)))
            emitMessage(Strings.t(Str.RideReturnFailed), isError = true)
            return
        }
        val accessories = collectBeaconAccessory(session.carId, session.imei)

        // 蓝牙还车前也要过一次 returnPermission（带 izSw=1 + 配件读数），
        // 否则后端不认这次锁车，会出现「车锁了但订单还在跑」。
        val permission = riding.returnPermission(
            carId = session.carId,
            orderId = session.orderId,
            actor = actor,
            bleChannel = true,
            accessories = accessories,
        )
        if (!permission.succeeded) {
            val message = permission.messageOr(Str.RideReturnFailed)
            dispatch(RideEvent.Error(message))
            emitMessage(message, isError = true)
            return
        }

        when (bleSession.lock(session.imei, mute = true)) {
            is RiderResult.Err -> {
                dispatch(RideEvent.Error(Strings.t(Str.RideReturnFailed)))
                emitMessage(Strings.t(Str.RideReturnFailed), isError = true)
                return
            }
            is RiderResult.Ok -> Unit
        }

        val reported = bleRide.returnReport(
            carId = session.carId,
            orderId = session.orderId,
            actor = actor,
            accessories = accessories,
        )
        if (reported.succeeded) {
            bleSession.playVoice(session.imei, BleVoice.LOCK, BleVoice.DEFAULT_VOLUME)
            finishReturn(orderId = session.orderId)
        } else {
            val message = reported.messageOr(Str.RideReturnFailed)
            dispatch(RideEvent.Error(message))
            emitMessage(message, isError = true)
        }
    }

    /** 蓝牙还车前的 beacon 读取。整段软失败：读不到就不带这个配件，别拦还车。 */
    private suspend fun collectBeaconAccessory(
        carId: String,
        imei: String,
    ): List<BleAccessoryReport> {
        val config = bleRide.returnConfig(carId).getOrNull() ?: return emptyList()
        if (!config.needBeacon) return emptyList()
        val beacon = bleSession.checkBeacon(imei)
        if (beacon !is RiderResult.Ok) {
            logger.w(TAG, "beacon check failed, submitting without it")
            return emptyList()
        }
        // 旧版判据是 `valueDecode.event` 有值；解析失败时 fields 是空 map。
        val fields = beacon.value.decoded.fields
        if (fields["event"].isNullOrBlank()) {
            logger.w(TAG, "beacon notify carried no event")
            return emptyList()
        }
        return listOf(BleAccessoryReport(name = "beacon", state = fields))
    }

    private fun finishReturn(orderId: String) {
        val info = _state.value.rideInfo
        val oid = orderId.ifBlank { _state.value.session.orderId }
        dispatch(RideEvent.ReturnSuccess(oid))
        _state.value = _state.value.copy(
            prompt = null,
            settlement = SettlementSummary(
                orderId = oid,
                costFeeFen = info?.costFeeFen ?: 0,
                dispatchFeeFen = info?.dispatchCostFen ?: 0,
                rideTimeSeconds = info?.rideTimeSeconds ?: _state.value.elapsedSeconds,
                rideDistanceMeters = info?.rideDistanceMeters ?: 0,
                originCostFen = info?.costFeeFen ?: 0,
                enteredAtMillis = nowEpochMillis(),
                frozenAtMillis = nowEpochMillis(),
            ),
        )
        emitMessage(Strings.t(Str.RideReturnOk))
    }

    /**
     * 结费页进入后拉订单详情 + 钱包分桶，把起步/时长/里程费补齐（对齐 UniApp `pay.vue` refresh）。
     */
    suspend fun hydrateSettlement() {
        val current = _state.value.settlement ?: return
        val oid = current.orderId
        var next = current
        when (val detail = riding.orderDetail(oid)) {
            is RiderResult.Ok -> {
                val d = detail.value
                next = next.copy(
                    orderId = d.orderId.ifBlank { oid },
                    costFeeFen = if (d.costFeeFen > 0) d.costFeeFen else next.costFeeFen,
                    dispatchFeeFen = if (d.dispatchFeeFen > 0) d.dispatchFeeFen else next.dispatchFeeFen,
                    rideTimeSeconds = if (d.rideTimeSeconds > 0L) d.rideTimeSeconds else next.rideTimeSeconds,
                    rideDistanceMeters = if (d.rideDistanceMeters > 0) {
                        d.rideDistanceMeters
                    } else {
                        next.rideDistanceMeters
                    },
                    settled = d.settled || next.settled,
                    originCostFen = if (d.originCostFen > 0) d.originCostFen else next.originCostFen,
                    startPriceFen = d.startPriceFen ?: next.startPriceFen,
                    timeCostFen = d.timeCostFen ?: next.timeCostFen,
                    mileCostFen = d.mileCostFen ?: next.mileCostFen,
                    rechargeBalanceFen = d.rechargeBalanceFen,
                    presentBalanceFen = d.presentBalanceFen,
                    frozenAtMillis = if (d.frozenAtMillis > 0L) d.frozenAtMillis else next.frozenAtMillis,
                    enteredAtMillis = if (next.enteredAtMillis > 0L) {
                        next.enteredAtMillis
                    } else {
                        nowEpochMillis()
                    },
                )
            }
            is RiderResult.Err -> logger.w(TAG, "orderDetail soft fail: ${detail.error.message}")
        }
        when (val wallet = riding.walletBalances()) {
            is RiderResult.Ok -> {
                val (recharge, present) = wallet.value
                if (next.rechargeBalanceFen == 0 && recharge > 0) {
                    next = next.copy(rechargeBalanceFen = recharge)
                }
                if (next.presentBalanceFen == 0 && present > 0) {
                    next = next.copy(presentBalanceFen = present)
                }
            }
            is RiderResult.Err -> Unit
        }
        if (_state.value.phase == RidePhase.Settling) {
            _state.value = _state.value.copy(settlement = next)
        }
    }

    /** 结费屏点「完成」。支付未接通时这就是全部收尾。 */
    fun finishSettlement() {
        dispatch(RideEvent.SettleDone)
        _state.value = _state.value.copy(
            vehicle = null,
            rideInfo = null,
            settlement = null,
            unlockChannel = null,
            elapsedSeconds = 0L,
            track = emptyList(),
            fenceTip = null,
        )
        trackBuffer.reset()
        committedTrack.clear()
    }

    // ── 申诉 / 辅助操作 ─────────────────────────────────────────────────────

    suspend fun canApplyReturn(applyType: Int?): Boolean =
        applyReturn.canApply(applyType).getOrNull() ?: false

    /**
     * 提交「无法还车」申诉。成功后置 autoReturn 标记，回骑行页自动再跑一次还车
     * （旧版 `storage.set('autoLock', true)`）。
     */
    suspend fun submitApplyReturn(
        photoUrls: List<String>,
        applyType: Int?,
        reason: String,
    ): Boolean {
        var ok = false
        withBusy(Str.Loading) {
            val result = applyReturn.createAudit(
                photoUrls = photoUrls,
                orderId = _state.value.session.orderId,
                applyType = applyType,
                reason = reason,
            )
            when (result) {
                is RiderResult.Ok -> {
                    store.markAutoReturn()
                    ok = true
                    emitMessage(Strings.t(Str.OperationOk))
                }
                is RiderResult.Err -> emitMessage(result.error.message, isError = true)
            }
        }
        return ok
    }

    suspend fun ringVehicle() {
        val session = _state.value.session
        if (session.carId.isBlank() && session.imei.isBlank()) return
        withBusy(Str.RideFindingVehicle) {
            when (riding.playVehicleVoice(session.carId, session.imei)) {
                is RiderResult.Ok -> emitMessage(Strings.t(Str.RideRingOk))
                is RiderResult.Err -> emitMessage(Strings.t(Str.RideRingFailed), isError = true)
            }
        }
    }

    suspend fun openHelmetLock() {
        val carId = _state.value.session.carId
        if (carId.isBlank()) return
        withBusy(Str.Loading) {
            when (val result = riding.unlockHelmet(carId)) {
                is RiderResult.Ok -> emitMessage(Strings.t(Str.RideHelmetUnlockOk))
                is RiderResult.Err -> emitMessage(result.error.message, isError = true)
            }
        }
    }

    /** 出服务区断电后临时恢复动力。 */
    suspend fun recoverPower() {
        val carId = _state.value.session.carId
        if (carId.isBlank()) return
        _state.value = _state.value.copy(prompt = null)
        withBusy(Str.Loading) {
            when (val result = riding.tempUnlock(carId)) {
                is RiderResult.Ok -> {
                    emitMessage(Strings.t(Str.RideTempUnlockOk))
                    refreshRideInfo()
                }
                is RiderResult.Err -> emitMessage(result.error.message, isError = true)
            }
        }
    }

    suspend fun loadFences() {
        val at = _state.value.myLocation ?: currentLocation() ?: return
        when (val result = fences.nearFences(at, store.serviceAreaId)) {
            is RiderResult.Ok -> _state.value = _state.value.copy(fences = result.value)
            is RiderResult.Err -> logger.w(TAG, "fence load failed: ${result.error.message}")
        }
    }

    suspend fun loadNearParking(at: GeoPoint? = null) {
        val target = at ?: _state.value.myLocation ?: currentLocation() ?: return
        withBusy(Str.Loading) {
            // 目的地可能落在别的服务区，先按坐标问一次而不是用缓存的 serviceId。
            val serviceId = fences.serviceAreaIdAt(target).getOrNull().orEmpty()
            when (val result = fences.nearParking(target, serviceId)) {
                is RiderResult.Ok -> _state.value = _state.value.copy(nearParking = result.value)
                is RiderResult.Err -> emitMessage(result.error.message, isError = true)
            }
        }
    }

    fun dismissPrompt() {
        val current = _state.value
        _state.value = current.copy(prompt = null)
        // 还车弹层被关掉 = 用户放弃这次还车，把相位收回骑行。
        if (current.session.phase == RidePhase.Returning) {
            dispatch(RideEvent.Error(""))
        }
    }

    fun consumeMessage() {
        _state.value = _state.value.copy(message = null)
    }

    // ── 轮询 ────────────────────────────────────────────────────────────────

    /**
     * 一轮 `getRideInfo`。服务端是订单状态的唯一权威 —— 车可能被运维远程还了，
     * 也可能上一次还车请求实际成功但响应丢在路上。
     */
    suspend fun refreshRideInfo() {
        val actor = actorWithLocation() ?: actorOrPinOnly()
        val outcome = when (val result = riding.rideInfo(actor)) {
            is RiderResult.Err -> {
                logger.w(TAG, "rideInfo failed: ${result.error.message}")
                return
            }
            is RiderResult.Ok -> result.value
        }
        if (!outcome.success) {
            when (outcome.code.trim()) {
                // 15042 = 无进行中订单，旧版软失败不弹错、不踢结费。
                CODE_NO_ORDER -> Unit
                // 15009 = 订单已结束 → 进结费（对齐 UniApp reLaunch pay）。
                CODE_ORDER_FINISHED -> {
                    dispatch(
                        RideEvent.ServerSync(ridingState = null, orderEnded = true),
                    )
                    ensureSettlementFromRideInfo()
                }
                else -> logger.w(TAG, "rideInfo business ${outcome.code}: ${outcome.message}")
            }
            return
        }

        val info = RidingParsers.rideInfo(outcome.data)
        _state.value = _state.value.copy(
            rideInfo = info,
            fenceTip = RidingFenceTips.of(
                fenceTypeCode = info.fenceTypeCode,
                dispatchCostFen = info.dispatchCostFen,
                canReturn = info.canReturn,
            ),
        )
        dispatch(
            RideEvent.ServerSync(
                ridingState = info.ridingState,
                carId = info.carId,
                imei = info.imei,
                orderId = info.orderId,
                orderEnded = false,
            ),
        )
        // 服务端时长是权威，本地秒表只在两次轮询之间补间。
        if (info.rideTimeSeconds > 0L) {
            _state.value = _state.value.copy(elapsedSeconds = info.rideTimeSeconds)
        }
        applyRideInfoPrompts(info)
        if (_state.value.phase == RidePhase.Settling) {
            ensureSettlementFromRideInfo()
        }
    }

    /** 被服务端踢进结费时，用最近一次 rideInfo 填结费摘要（还车成功路径会自己造 SettlementSummary）。 */
    private fun ensureSettlementFromRideInfo() {
        if (_state.value.settlement != null) return
        if (_state.value.phase != RidePhase.Settling) return
        val info = _state.value.rideInfo
        _state.value = _state.value.copy(
            settlement = SettlementSummary(
                orderId = info?.orderId.orEmpty().ifBlank { _state.value.session.orderId },
                costFeeFen = info?.costFeeFen ?: 0,
                dispatchFeeFen = info?.dispatchCostFen ?: 0,
                rideTimeSeconds = info?.rideTimeSeconds ?: _state.value.elapsedSeconds,
                rideDistanceMeters = info?.rideDistanceMeters ?: 0,
                originCostFen = info?.costFeeFen ?: 0,
                enteredAtMillis = nowEpochMillis(),
                frozenAtMillis = nowEpochMillis(),
            ),
        )
    }

    private fun applyRideInfoPrompts(info: RideInfo) {
        if (_state.value.prompt != null) return
        if (info.overloadState == 2 || info.overloadState == 3) {
            _state.value = _state.value.copy(
                prompt = RidePrompt.Overload(powerOff = info.overloadState == 2),
            )
            return
        }
        // 平台主动断电，等用户确认恢复动力。
        if (info.tempUnlockState == 1 && !info.tempUnlocked && _state.value.session.carId.isNotBlank()) {
            _state.value = _state.value.copy(
                prompt = RidePrompt.RecoverPower(_state.value.config.autoLockMinutes),
            )
        }
    }

    private suspend fun loadBackCarConfig() {
        when (val result = riding.backCarConfig(store.serviceAreaId)) {
            is RiderResult.Ok -> _state.value = _state.value.copy(backCarConfig = result.value)
            is RiderResult.Err -> logger.w(TAG, "backCarConfig failed: ${result.error.message}")
        }
    }

    /** 进骑行页时软调一次解冻。缺定位就跳过（后端要坐标）。 */
    suspend fun unfreezePreviousOrder() {
        val actor = actorWithLocation() ?: return
        riding.unFrozenOrder(actor)
    }

    // ── 内部 ────────────────────────────────────────────────────────────────

    private fun dispatch(event: RideEvent) {
        val before = _state.value.session
        val after = RideStateMachine.reduce(before, event)
        if (after == before) return
        _state.value = _state.value.copy(
            session = after,
            elapsedSeconds = if (after.phase == RidePhase.Idle) 0L else _state.value.elapsedSeconds,
        )
        store.save(after)
        if (before.needsForegroundLocation != after.needsForegroundLocation) {
            syncForegroundLocation(after)
        }
    }

    private fun syncForegroundLocation(session: RideSession) {
        runCatching {
            if (session.needsForegroundLocation) {
                locationTracker.startTracking()
            } else {
                locationTracker.stopTracking()
            }
        }.onFailure { logger.w(TAG, "foreground location toggle failed: ${it.message}") }
    }

    private suspend fun currentLocation(): GeoPoint? =
        when (val result = locationTracker.currentLocation()) {
            is RiderResult.Ok -> result.value.also {
                _state.value = _state.value.copy(myLocation = it)
            }
            is RiderResult.Err -> null
        }

    private suspend fun actorWithLocation(): RideActor? {
        val at = _state.value.myLocation ?: currentLocation() ?: return null
        if (at.latitude == 0.0 && at.longitude == 0.0) return null
        return RideActor(userPin = store.userPin, at = at)
    }

    /** 开锁专用：优先实时定位，避免 Confirming 阶段缓存为空时静默拦下。 */
    private suspend fun actorWithFreshLocation(): RideActor? {
        val at = currentLocation() ?: _state.value.myLocation ?: return null
        if (at.latitude == 0.0 && at.longitude == 0.0) return null
        return RideActor(userPin = store.userPin, at = at)
    }

    private fun actorOrPinOnly(): RideActor =
        RideActor(userPin = store.userPin, at = _state.value.myLocation)

    private suspend fun withBusy(label: Str, block: suspend () -> Unit) {
        _state.value = _state.value.copy(busy = true, busyLabel = label)
        try {
            block()
        } finally {
            _state.value = _state.value.copy(busy = false, busyLabel = null)
        }
    }

    private fun emitMessage(text: String, isError: Boolean = false) {
        if (text.isBlank()) return
        messageNonce += 1
        _state.value = _state.value.copy(
            message = RideMessage(text = text, isError = isError, nonce = messageNonce),
        )
    }

    private fun elapsedFrom(session: RideSession): Long {
        if (session.startedAtMillis <= 0L || !session.isOnTrip) return 0L
        return ((nowEpochMillis() - session.startedAtMillis) / 1000L).coerceAtLeast(0L)
    }

    companion object {
        private const val TAG = "RidingFeature"

        /** 旧版骑行页 8 秒一轮，跟后端的订单状态推进节奏一致。 */
        const val POLL_INTERVAL_MS: Long = 8_000L

        /** 余额不足。 */
        const val CODE_INSUFFICIENT_BALANCE: String = "15030"

        /** 临时冻结单，待支付。 */
        const val CODE_TEMP_FROZEN: String = "15002"

        /** 订单已结束。 */
        const val CODE_ORDER_FINISHED: String = "15009"

        /** 无进行中订单（软失败）。 */
        const val CODE_NO_ORDER: String = "15042"

        /** 车机离线 / 远程指令下发失败，整族 `17012xxx`。 */
        const val CODE_DEVICE_OFFLINE_PREFIX: String = "17012"
    }
}

// ── ApiOutcome 小工具 ───────────────────────────────────────────────────────

private val ApiOutcome.isDeviceOffline: Boolean
    get() = code.trim().startsWith(RidingFeature.CODE_DEVICE_OFFLINE_PREFIX)

private val ApiOutcome.isInsufficientBalance: Boolean
    get() = code.trim().contains(RidingFeature.CODE_INSUFFICIENT_BALANCE)

private val RiderResult<ApiOutcome>.succeeded: Boolean
    get() = this is RiderResult.Ok && value.success

private val RiderResult<ApiOutcome>.isDeviceOffline: Boolean
    get() = this is RiderResult.Ok && value.isDeviceOffline

private fun RiderResult<ApiOutcome>.messageOr(fallback: Str): String = when (this) {
    is RiderResult.Err -> error.message.ifBlank { Strings.t(fallback) }
    is RiderResult.Ok -> value.message.ifBlank { Strings.t(fallback) }
}

private fun ApiOutcome.messageOr(fallback: Str): String = message.ifBlank { Strings.t(fallback) }
