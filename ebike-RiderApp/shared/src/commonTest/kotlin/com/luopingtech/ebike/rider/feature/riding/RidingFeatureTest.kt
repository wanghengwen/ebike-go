package com.luopingtech.ebike.rider.feature.riding

import com.luopingtech.ebike.rider.core.network.ApiOutcome
import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.data.ble.DemoBleTokenApi
import com.luopingtech.ebike.rider.data.fence.FenceRemote
import com.luopingtech.ebike.rider.data.riding.ApplyReturnRemote
import com.luopingtech.ebike.rider.data.riding.BleAccessoryReport
import com.luopingtech.ebike.rider.data.riding.BleRideRemote
import com.luopingtech.ebike.rider.data.riding.RideActor
import com.luopingtech.ebike.rider.data.riding.RidingRemote
import com.luopingtech.ebike.rider.domain.ble.BleCommands
import com.luopingtech.ebike.rider.domain.ble.BleConvert
import com.luopingtech.ebike.rider.domain.ble.BleFrame
import com.luopingtech.ebike.rider.domain.model.FencePolygon
import com.luopingtech.ebike.rider.domain.riding.BackCarConfig
import com.luopingtech.ebike.rider.domain.riding.BleReturnConfig
import com.luopingtech.ebike.rider.domain.riding.NearParking
import com.luopingtech.ebike.rider.domain.riding.PartMatchResult
import com.luopingtech.ebike.rider.domain.riding.RidePhase
import com.luopingtech.ebike.rider.domain.riding.ReturnTypeCodes
import com.luopingtech.ebike.rider.domain.riding.SettlementSummary
import com.luopingtech.ebike.rider.domain.riding.VehicleDetail
import com.luopingtech.ebike.rider.feature.ble.BleSessionFeature
import com.luopingtech.ebike.rider.platform.BleTransport
import com.luopingtech.ebike.rider.platform.GeoPoint
import com.luopingtech.ebike.rider.platform.InMemorySecureStore
import com.luopingtech.ebike.rider.platform.LocationTracker
import com.luopingtech.ebike.rider.platform.SimulatorBleTransport
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.put
import kotlin.test.Test
import kotlin.test.assertContains
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertNotNull
import kotlin.test.assertTrue

/**
 * 编排层单测。重点不是"能跑通"，而是那些顺序 / 降级错了会真赔钱的地方：
 * 静音开锁与建单的先后、建单失败要回滚上锁、蓝牙失败降级网络、余额不足要弹充值。
 */
class RidingFeatureTest {

    // ── 开锁通道 ────────────────────────────────────────────────────────────

    @Test
    fun blePreferredUnlockGoesThroughBluetoothAndOpensAnOrder() = runTest {
        val env = env(policy = UnlockPolicy.BlePreferred)
        env.feature.selectVehicle(carId = "C1")
        env.feature.confirmUnlock()

        val state = env.feature.state.value
        assertEquals(RidePhase.Riding, state.phase)
        assertEquals(UnlockChannel.Ble, state.unlockChannel)
        assertEquals("ORDER-BLE", state.session.orderId)
        // 蓝牙成功时不该再花一轮公网往返。
        assertEquals(0, env.riding.networkRideCalls)
    }

    @Test
    fun bleUnlockIsMutedAndVoicePlaysOnlyAfterTheOrderExists() = runTest {
        val env = env(policy = UnlockPolicy.BlePreferred)
        env.feature.selectVehicle(carId = "C1")
        env.feature.confirmUnlock()

        // 顺序反了会出现「车响了但没订单」——用户以为骑上了，实际白骑。
        val unlockIndex = env.ble.writes.indexOfFirst { it.startsWith(unlockPrefix()) }
        val voiceIndex = env.ble.writes.indexOfFirst { it.startsWith(voicePrefix()) }
        assertTrue(unlockIndex >= 0, "no unlock frame written: ${env.ble.writes}")
        assertTrue(voiceIndex > unlockIndex, "voice must follow unlock: ${env.ble.writes}")
        assertTrue(
            env.bleRide.rideReportAt in (unlockIndex + 1)..voiceIndex,
            "order must be created between unlock and voice",
        )
        // 开锁帧走静音载荷，语音留给建单成功之后。
        assertEquals(BleFrame.build(BleCommands.UNLOCK, BleFrame.MUTE_PAYLOAD), env.ble.writes[unlockIndex])
    }

    @Test
    fun aFailedOrderReportLocksTheBikeBackAndKeepsTheUserOnConfirm() = runTest {
        val env = env(policy = UnlockPolicy.BlePreferred, bleReportFails = true, networkRideFails = true)
        env.feature.selectVehicle(carId = "C1")
        env.feature.confirmUnlock()

        // 车已经开了但服务端没建单 —— 必须锁回去，否则用户白骑、账也对不上。
        assertContains(env.ble.writes.map { it.take(2) }, lockPrefix())
        assertEquals(RidePhase.Confirming, env.feature.state.value.phase)
        assertEquals(1, env.feature.state.value.session.unlockAttempts)
    }

    @Test
    fun bleFailureFallsBackToTheNetworkUnlock() = runTest {
        val env = env(policy = UnlockPolicy.BlePreferred, blePermissionFails = true)
        env.feature.selectVehicle(carId = "C1")
        env.feature.confirmUnlock()

        val state = env.feature.state.value
        assertEquals(RidePhase.Riding, state.phase)
        assertEquals(UnlockChannel.Network, state.unlockChannel)
        assertEquals("ORDER-NET", state.session.orderId)
        assertTrue(state.session.usedNetworkFallback)
        assertEquals(1, env.riding.networkRideCalls)
    }

    @Test
    fun bleOnlyPolicyNeverTouchesTheNetworkUnlock() = runTest {
        // 车队没有联网车机时，一轮必然超时的远程指令是纯浪费。
        val env = env(policy = UnlockPolicy.BleOnly, blePermissionFails = true)
        env.feature.selectVehicle(carId = "C1")
        env.feature.confirmUnlock()

        assertEquals(0, env.riding.networkRideCalls)
        assertEquals(RidePhase.Confirming, env.feature.state.value.phase)
    }

    @Test
    fun withoutBluetoothHardwareUnlockGoesStraightToTheNetwork() = runTest {
        val env = env(policy = UnlockPolicy.BlePreferred, bleAvailable = false)
        env.feature.selectVehicle(carId = "C1")
        env.feature.confirmUnlock()

        assertEquals(UnlockChannel.Network, env.feature.state.value.unlockChannel)
        assertTrue(env.ble.writes.isEmpty())
    }

    @Test
    fun networkPreferredPolicyOnlyUsesBleWhenTheBikeIsOffline() = runTest {
        val online = env(policy = UnlockPolicy.NetworkPreferred)
        online.feature.selectVehicle(carId = "C1")
        online.feature.confirmUnlock()
        assertEquals(UnlockChannel.Network, online.feature.state.value.unlockChannel)

        val offline = env(policy = UnlockPolicy.NetworkPreferred, vehicleDisconnected = true)
        offline.feature.selectVehicle(carId = "C1")
        offline.feature.confirmUnlock()
        assertEquals(UnlockChannel.Ble, offline.feature.state.value.unlockChannel)
    }

    @Test
    fun deviceOffline17012RetriesOverBluetooth() = runTest {
        val env = env(
            policy = UnlockPolicy.NetworkPreferred,
            networkRideCode = "17012001",
        )
        env.feature.selectVehicle(carId = "C1")
        env.feature.confirmUnlock()

        assertEquals(UnlockChannel.Ble, env.feature.state.value.unlockChannel)
        assertEquals(RidePhase.Riding, env.feature.state.value.phase)
    }

    // ── 开锁前置校验 ────────────────────────────────────────────────────────

    @Test
    fun unlockIsRefusedWhenTheBikeSitsOutsideTheServiceArea() = runTest {
        val env = env(vehicleOutOfService = true)
        env.feature.selectVehicle(carId = "C1")
        env.feature.confirmUnlock()

        assertEquals(RidePhase.Confirming, env.feature.state.value.phase)
        assertTrue(env.ble.writes.isEmpty())
        assertEquals(0, env.riding.networkRideCalls)
    }

    @Test
    fun unlockIsRefusedWithoutAUsableFix() = runTest {
        // 后端要坐标，没定位发过去必然被拒；本地先挡住并给出可操作的提示。
        val env = env(location = null)
        env.feature.selectVehicle(carId = "C1")
        env.feature.confirmUnlock()

        assertEquals(RidePhase.Confirming, env.feature.state.value.phase)
        assertEquals(0, env.riding.networkRideCalls)
        assertNotNull(env.feature.state.value.message)
    }

    @Test
    fun aZeroZeroFixIsTreatedAsNoFix() = runTest {
        val env = env(location = GeoPoint(0.0, 0.0))
        env.feature.selectVehicle(carId = "C1")
        env.feature.confirmUnlock()
        assertEquals(0, env.riding.networkRideCalls)
    }

    @Test
    fun insufficientBalanceRaisesTheTopUpPrompt() = runTest {
        val env = env(policy = UnlockPolicy.BlePreferred, blePermissionCode = "15030")
        env.feature.selectVehicle(carId = "C1")
        env.feature.confirmUnlock()

        val prompt = env.feature.state.value.prompt
        assertTrue(prompt is RidePrompt.InsufficientBalance, "prompt=$prompt")
        assertEquals(RidePhase.Confirming, env.feature.state.value.phase)
        // 余额不足是确定性失败，不要再去网络重试一遍。
        assertEquals(0, env.riding.networkRideCalls)
    }

    // ── 临停 ────────────────────────────────────────────────────────────────

    @Test
    fun tempLockAndResumeWalkThePhaseBackAndForth() = runTest {
        val env = ridingEnv()
        env.feature.toggleTempLock()
        assertEquals(RidePhase.TempLocked, env.feature.state.value.phase)

        env.feature.toggleTempLock()
        assertEquals(RidePhase.Riding, env.feature.state.value.phase)
    }

    @Test
    fun anUnstowedHelmetBlocksTheTempLock() = runTest {
        // 锁了车用户回来发现头盔丢了还得赔，这一步不能省。
        val env = ridingEnv(helmetStowed = false)
        env.feature.toggleTempLock()

        assertEquals(RidePhase.Riding, env.feature.state.value.phase)
        assertEquals(RidePrompt.HelmetNotReturned, env.feature.state.value.prompt)
        assertEquals(0, env.riding.tempParkCalls)
    }

    @Test
    fun anOfflineBikeIsTempLockedOverBluetooth() = runTest {
        val env = ridingEnv(tempParkCode = "17012001")
        env.feature.toggleTempLock()

        assertEquals(RidePhase.TempLocked, env.feature.state.value.phase)
        assertTrue(env.bleRide.tempParkReported)
    }

    // ── 还车 ────────────────────────────────────────────────────────────────

    @Test
    fun aNormalReturnSubmitsImmediatelyAndSettles() = runTest {
        val env = ridingEnv()
        env.feature.beginReturn()

        val state = env.feature.state.value
        assertEquals(RidePhase.Settling, state.phase)
        assertNotNull(state.settlement)
    }

    @Test
    fun theCivilizationReminderIsShownBeforeCommitting() = runTest {
        val env = ridingEnv(civilizationRemind = true)
        env.feature.beginReturn()

        // 弹层出现时车还不能锁 —— 旧版这里已经把车锁了。
        assertEquals(RidePrompt.Civilization, env.feature.state.value.prompt)
        assertEquals(RidePhase.Returning, env.feature.state.value.phase)
        assertEquals(0, env.riding.returnByNetCalls)

        env.feature.confirmReturn(forcePenalty = false)
        assertEquals(RidePhase.Settling, env.feature.state.value.phase)
        assertEquals(1, env.riding.returnByNetCalls)
    }

    @Test
    fun anOutOfSpotReturnAsksTheUserToAcceptTheFee() = runTest {
        val env = ridingEnv(returnTypeCode = ReturnTypeCodes.OUT_OF_SPOT, penaltyFen = 200)
        env.feature.beginReturn()

        val prompt = env.feature.state.value.prompt
        assertTrue(prompt is RidePrompt.Penalty, "prompt=$prompt")
        assertEquals(200, prompt.penaltyFen)

        env.feature.confirmReturn(forcePenalty = true)
        // returnType=1 才是「认罚」，传 0 后端会再拒一次。
        assertEquals(1, env.riding.lastReturnType)
        assertEquals(RidePhase.Settling, env.feature.state.value.phase)
    }

    @Test
    fun aBlockedReturnExplainsItselfAndKeepsTheBike() = runTest {
        val env = ridingEnv(canReturn = false, returnTypeCode = ReturnTypeCodes.NO_PARKING)
        env.feature.beginReturn()

        assertTrue(env.feature.state.value.prompt is RidePrompt.Blocked)
        assertEquals(0, env.riding.returnByNetCalls)

        env.feature.dismissPrompt()
        // 放弃还车要回到骑行，车还在用户手上。
        assertEquals(RidePhase.Riding, env.feature.state.value.phase)
    }

    @Test
    fun aGuidedReturnCodeOpensTheGuideInsteadOfSubmitting() = runTest {
        val env = ridingEnv(canReturn = false, returnTypeCode = ReturnTypeCodes.HELMET_MISS_A)
        env.feature.beginReturn()

        val prompt = env.feature.state.value.prompt
        assertTrue(prompt is RidePrompt.Guide, "prompt=$prompt")
        assertEquals(6, prompt.pageType)
        assertEquals(0, env.riding.returnByNetCalls)
    }

    @Test
    fun anAlreadyFinishedOrderIsTreatedAsASuccessfulReturn() = runTest {
        // 15009：上一次还车其实成功了，只是响应丢在路上。
        val env = ridingEnv(returnByNetCode = "15009")
        env.feature.beginReturn()
        assertEquals(RidePhase.Settling, env.feature.state.value.phase)
    }

    @Test
    fun aFrozenOrderStillReachesSettlementForThePendingPayment() = runTest {
        val env = ridingEnv(returnByNetCode = "15002")
        env.feature.beginReturn()
        assertEquals(RidePhase.Settling, env.feature.state.value.phase)
        assertNotNull(env.feature.state.value.settlement)
    }

    @Test
    fun aFenceRejectionAtSubmitTimeReopensTheDecisionMatrix() = runTest {
        // 用户在弹层上犹豫的几秒里把车挪了 / 站点满了。
        val env = ridingEnv(
            returnByNetCode = "0",
            returnByNetRetypeTo = ReturnTypeCodes.OUT_OF_SPOT,
            returnByNetRepenaltyFen = 300,
        )
        env.feature.beginReturn()

        val prompt = env.feature.state.value.prompt
        assertTrue(prompt is RidePrompt.Penalty, "prompt=$prompt")
        assertEquals(300, prompt.penaltyFen)
        assertEquals(RidePhase.Returning, env.feature.state.value.phase)
    }

    @Test
    fun anOfflineBikeIsReturnedOverBluetooth() = runTest {
        val env = ridingEnv(returnByNetCode = "17012001")
        env.feature.beginReturn()

        assertTrue(env.bleRide.returnReported)
        assertEquals(RidePhase.Settling, env.feature.state.value.phase)
    }

    @Test
    fun aReturnFailureKeepsTheBikeAndAllowsARetry() = runTest {
        val env = ridingEnv(returnByNetCode = "19999")
        env.feature.beginReturn()

        assertEquals(RidePhase.Riding, env.feature.state.value.phase)
        assertEquals("C1", env.feature.state.value.carId)
    }

    @Test
    fun finishingSettlementClearsTheSession() = runTest {
        val env = ridingEnv()
        env.feature.beginReturn()
        env.feature.finishSettlement()

        val state = env.feature.state.value
        assertEquals(RidePhase.Idle, state.phase)
        assertEquals(null, state.vehicle)
        assertEquals(null, state.settlement)
        assertEquals(0L, state.elapsedSeconds)
        // 结完账磁盘上不该再留骑行快照。
        assertEquals(RidePhase.Idle, env.store.load().phase)
    }

    // ── 前台定位 ────────────────────────────────────────────────────────────

    @Test
    fun foregroundLocationFollowsTheRidePhase() = runTest {
        val env = env()
        assertFalse(env.location.tracking)

        env.feature.selectVehicle(carId = "C1")
        assertFalse(env.location.tracking)

        env.feature.confirmUnlock()
        assertTrue(env.location.tracking)

        env.feature.beginReturn()
        env.feature.finishSettlement()
        // 还完车要放掉定位，否则电量和隐私提示都下不去。
        assertFalse(env.location.tracking)
    }

    // ── 持久化 ──────────────────────────────────────────────────────────────

    @Test
    fun theRidingPhaseIsPersistedAndRestored() = runTest {
        val env = env()
        env.feature.selectVehicle(carId = "C1")
        env.feature.confirmUnlock()
        assertEquals(RidePhase.Riding, env.store.load().phase)

        // 换一个 feature 实例模拟冷启动。
        val revived = env.copyFeature()
        revived.restore()
        assertEquals(RidePhase.Riding, revived.state.value.phase)
        assertEquals("C1", revived.state.value.carId)
    }

    @Test
    fun restoringARidingPhaseTurnsForegroundLocationBackOn() = runTest {
        val env = env()
        env.feature.selectVehicle(carId = "C1")
        env.feature.confirmUnlock()

        val fresh = FakeLocationTracker(GeoPoint(28.2, 112.9))
        val revived = env.copyFeature(location = fresh)
        revived.restore()
        assertTrue(fresh.tracking)
    }

    // ── 申诉 ────────────────────────────────────────────────────────────────

    @Test
    fun aSuccessfulAppealArmsTheAutomaticReturnRetry() = runTest {
        val env = ridingEnv()
        assertTrue(env.feature.submitApplyReturn(listOf("https://x/1.jpg"), applyType = 1, reason = "挡道"))
        // 回骑行页由轮询消费这个标记，自动再跑一次还车（旧版 autoLock）。
        assertTrue(env.store.consumeAutoReturn())
    }

    @Test
    fun aFailedAppealDoesNotArmTheRetry() = runTest {
        val env = ridingEnv(applyFails = true)
        assertFalse(env.feature.submitApplyReturn(listOf("https://x/1.jpg"), applyType = 1, reason = "挡道"))
        assertFalse(env.store.consumeAutoReturn())
    }

    // ── 测试脚手架 ──────────────────────────────────────────────────────────

    /** 帧的头两个 hex 字符就是指令字节。 */
    private fun unlockPrefix() = BleConvert.toPad2HexStr(BleCommands.UNLOCK)

    private fun voicePrefix() = BleConvert.toPad2HexStr(BleCommands.PLAY_VOICE)

    private fun lockPrefix() = BleConvert.toPad2HexStr(BleCommands.LOCK)

    private class Env(
        val feature: RidingFeature,
        val riding: FakeRidingRemote,
        val bleRide: FakeBleRideRemote,
        val ble: RecordingBleTransport,
        val location: FakeLocationTracker,
        val store: RideSessionStore,
        val applyReturn: FakeApplyReturnRemote,
        private val policy: UnlockPolicy,
        private val bleAvailable: Boolean,
    ) {
        /** 用同一份磁盘状态另起一个 feature，模拟杀进程后冷启动。 */
        fun copyFeature(location: FakeLocationTracker? = null): RidingFeature = RidingFeature(
            riding = riding,
            bleRide = bleRide,
            applyReturn = applyReturn,
            fences = FakeFenceRemote(),
            bleSession = BleSessionFeature(ble, DemoBleTokenApi()),
            locationTracker = location ?: this.location,
            store = store,
            unlockPolicy = policy,
            bleAvailable = { bleAvailable },
        )
    }

    private fun env(
        policy: UnlockPolicy = UnlockPolicy.BlePreferred,
        bleAvailable: Boolean = true,
        location: GeoPoint? = GeoPoint(28.22, 112.94),
        vehicleDisconnected: Boolean = false,
        vehicleOutOfService: Boolean = false,
        blePermissionFails: Boolean = false,
        blePermissionCode: String = "",
        bleReportFails: Boolean = false,
        networkRideFails: Boolean = false,
        networkRideCode: String = "",
        civilizationRemind: Boolean = false,
        canReturn: Boolean = true,
        returnTypeCode: Int = ReturnTypeCodes.NORMAL,
        penaltyFen: Int = 0,
        returnByNetCode: String = "",
        returnByNetRetypeTo: Int = ReturnTypeCodes.NORMAL,
        returnByNetRepenaltyFen: Int = 0,
        tempParkCode: String = "",
        helmetStowed: Boolean = true,
        applyFails: Boolean = false,
    ): Env {
        val riding = FakeRidingRemote(
            vehicleDisconnected = vehicleDisconnected,
            vehicleOutOfService = vehicleOutOfService,
            networkRideFails = networkRideFails,
            networkRideCode = networkRideCode,
            civilizationRemind = civilizationRemind,
            canReturn = canReturn,
            returnTypeCode = returnTypeCode,
            penaltyFen = penaltyFen,
            returnByNetCode = returnByNetCode,
            returnByNetRetypeTo = returnByNetRetypeTo,
            returnByNetRepenaltyFen = returnByNetRepenaltyFen,
            tempParkCode = tempParkCode,
            helmetStowed = helmetStowed,
        )
        val ble = RecordingBleTransport()
        val bleRide = FakeBleRideRemote(
            permissionFails = blePermissionFails,
            permissionCode = blePermissionCode,
            reportFails = bleReportFails,
            writeLog = ble.writes,
            onTempParkChanged = { riding.tempParked = it },
        )
        val tracker = FakeLocationTracker(location)
        val store = RideSessionStore(InMemorySecureStore()).apply { userPin = "PIN-1" }
        val apply = FakeApplyReturnRemote(fails = applyFails)
        val feature = RidingFeature(
            riding = riding,
            bleRide = bleRide,
            applyReturn = apply,
            fences = FakeFenceRemote(),
            bleSession = BleSessionFeature(ble, DemoBleTokenApi()),
            locationTracker = tracker,
            store = store,
            unlockPolicy = policy,
            bleAvailable = { bleAvailable },
        )
        return Env(
            feature = feature,
            riding = riding,
            bleRide = bleRide,
            ble = ble,
            location = tracker,
            store = store,
            applyReturn = apply,
            policy = policy,
            bleAvailable = bleAvailable,
        )
    }

    /** 已经骑上车的环境，还车 / 临停用例的起点。走网络开锁，避免蓝牙写日志混进断言。 */
    private suspend fun ridingEnv(
        civilizationRemind: Boolean = false,
        canReturn: Boolean = true,
        returnTypeCode: Int = ReturnTypeCodes.NORMAL,
        penaltyFen: Int = 0,
        returnByNetCode: String = "",
        returnByNetRetypeTo: Int = ReturnTypeCodes.NORMAL,
        returnByNetRepenaltyFen: Int = 0,
        tempParkCode: String = "",
        helmetStowed: Boolean = true,
        applyFails: Boolean = false,
    ): Env {
        val env = env(
            policy = UnlockPolicy.NetworkPreferred,
            civilizationRemind = civilizationRemind,
            canReturn = canReturn,
            returnTypeCode = returnTypeCode,
            penaltyFen = penaltyFen,
            returnByNetCode = returnByNetCode,
            returnByNetRetypeTo = returnByNetRetypeTo,
            returnByNetRepenaltyFen = returnByNetRepenaltyFen,
            tempParkCode = tempParkCode,
            helmetStowed = helmetStowed,
            applyFails = applyFails,
        )
        env.feature.selectVehicle(carId = "C1")
        env.feature.confirmUnlock()
        check(env.feature.state.value.phase == RidePhase.Riding) {
            "fixture failed to reach Riding: ${env.feature.state.value.phase}"
        }
        env.riding.resetCounters()
        return env
    }
}

// ── 假实现 ──────────────────────────────────────────────────────────────────

private fun ok(data: JsonElement? = null) =
    ApiOutcome(success = true, code = "0", message = "", data = data)

private fun fail(code: String, message: String = "fake failure", data: JsonElement? = null) =
    ApiOutcome(success = false, code = code, message = message, data = data)

private class FakeRidingRemote(
    private val vehicleDisconnected: Boolean,
    private val vehicleOutOfService: Boolean,
    private val networkRideFails: Boolean,
    private val networkRideCode: String,
    private val civilizationRemind: Boolean,
    private val canReturn: Boolean,
    private val returnTypeCode: Int,
    private val penaltyFen: Int,
    private val returnByNetCode: String,
    private val returnByNetRetypeTo: Int,
    private val returnByNetRepenaltyFen: Int,
    private val tempParkCode: String,
    private val helmetStowed: Boolean,
) : RidingRemote {

    var networkRideCalls = 0
    var returnByNetCalls = 0
    var tempParkCalls = 0
    var lastReturnType = -1

    /** 真后端临停后 `ridingState` 会变 3；假实现也得跟上，否则轮询会把相位拽回骑行。 */
    var tempParked = false
    private var returnByNetSeen = 0
    private var orderOpen = false

    fun resetCounters() {
        networkRideCalls = 0
        returnByNetCalls = 0
        tempParkCalls = 0
    }

    override suspend fun rideInfo(actor: RideActor): RiderResult<ApiOutcome> =
        if (!orderOpen) {
            RiderResult.Ok(fail("15042", "no order"))
        } else {
            RiderResult.Ok(
                ok(
                    buildJsonObject {
                        put("orderId", "ORDER-NET")
                        put("carId", "C1")
                        put("ridingState", if (tempParked) 3 else 4)
                        put("costFee", 150)
                        put("rideTime", 0)
                        put("rideDistance", 0)
                        put("type", returnTypeCode)
                        put("izCanReturn", canReturn)
                        put("dispatchCost", penaltyFen)
                    },
                ),
            )
        }

    override suspend fun carInfo(carId: String): RiderResult<VehicleDetail> = RiderResult.Ok(
        VehicleDetail(
            carId = carId,
            imei = BleFrame.demoImeiForCarId(carId),
            disconnected = vehicleDisconnected,
            outOfServiceArea = vehicleOutOfService,
            restBattery = 80,
        ),
    )

    override suspend fun submitScan(carId: String): RiderResult<Unit> = RiderResult.Ok(Unit)

    override suspend fun networkRide(carId: String, actor: RideActor): RiderResult<ApiOutcome> {
        networkRideCalls += 1
        if (networkRideCode.isNotBlank()) return RiderResult.Ok(fail(networkRideCode))
        if (networkRideFails) return RiderResult.Ok(fail("19999"))
        orderOpen = true
        return RiderResult.Ok(ok(buildJsonObject { put("orderId", "ORDER-NET") }))
    }

    override suspend fun returnPermission(
        carId: String,
        orderId: String,
        actor: RideActor,
        bleChannel: Boolean,
        accessories: List<BleAccessoryReport>,
    ): RiderResult<ApiOutcome> = RiderResult.Ok(
        ok(
            buildJsonObject {
                put("izCanReturn", canReturn)
                put("returnType", returnTypeCode)
                put("penalty", penaltyFen)
            },
        ),
    )

    override suspend fun returnByNet(
        carId: String,
        orderId: String,
        actor: RideActor,
        returnType: Int,
    ): RiderResult<ApiOutcome> {
        returnByNetCalls += 1
        returnByNetSeen += 1
        lastReturnType = returnType
        // 围栏重判只演一次，否则用例会陷在弹层循环里。
        if (returnByNetCode == "0" && returnByNetSeen == 1) {
            return RiderResult.Ok(
                fail(
                    code = "0",
                    data = buildJsonObject {
                        put("izCanReturn", true)
                        put("returnType", returnByNetRetypeTo)
                        put("penalty", returnByNetRepenaltyFen)
                    },
                ),
            )
        }
        if (returnByNetCode.isNotBlank() && returnByNetCode != "0") {
            return RiderResult.Ok(fail(returnByNetCode))
        }
        orderOpen = false
        return RiderResult.Ok(ok(buildJsonObject { put("orderId", "ORDER-NET") }))
    }

    override suspend fun tempPark(carId: String, userPin: String): RiderResult<ApiOutcome> {
        tempParkCalls += 1
        if (tempParkCode.isNotBlank()) return RiderResult.Ok(fail(tempParkCode))
        tempParked = true
        return RiderResult.Ok(ok())
    }

    override suspend fun endTempPark(carId: String, userPin: String): RiderResult<ApiOutcome> {
        tempParked = false
        return RiderResult.Ok(ok())
    }

    override suspend fun partMatchByTempParking(
        carId: String,
        orderId: String,
        serviceId: String,
        userPin: String,
    ): RiderResult<PartMatchResult> = RiderResult.Ok(
        if (helmetStowed) {
            PartMatchResult(matched = true)
        } else {
            PartMatchResult(matched = false, partName = "helmet")
        },
    )

    override suspend fun backCarConfig(serviceId: String): RiderResult<BackCarConfig> =
        RiderResult.Ok(BackCarConfig(civilizationRemind = civilizationRemind, showApplyEntry = true))

    override suspend fun tempUnlock(carId: String): RiderResult<Unit> = RiderResult.Ok(Unit)

    override suspend fun unFrozenOrder(actor: RideActor): RiderResult<Unit> = RiderResult.Ok(Unit)

    override suspend fun unlockHelmet(carId: String): RiderResult<Unit> = RiderResult.Ok(Unit)

    override suspend fun playVehicleVoice(carId: String, imei: String): RiderResult<Unit> =
        RiderResult.Ok(Unit)

    override suspend fun orderDetail(orderId: String): RiderResult<SettlementSummary> =
        RiderResult.Ok(
            SettlementSummary(
                orderId = orderId.ifBlank { "TEST-ORDER" },
                costFeeFen = 100,
                originCostFen = 100,
                startPriceFen = 100,
                timeCostFen = 0,
                mileCostFen = 0,
                rideTimeSeconds = 148,
                rideDistanceMeters = 246,
            ),
        )

    override suspend fun walletBalances(): RiderResult<Pair<Int, Int>> =
        RiderResult.Ok(0 to 0)
}

private class FakeBleRideRemote(
    private val permissionFails: Boolean,
    private val permissionCode: String,
    private val reportFails: Boolean,
    private val writeLog: List<String>,
    /** 蓝牙上报也要让「后端」知道临停了，否则下一轮轮询把相位拽回骑行。 */
    private val onTempParkChanged: (Boolean) -> Unit = {},
) : BleRideRemote {

    /** 建单发生在第几次蓝牙写之后 —— 用来钉住「先建单再播语音」的顺序。 */
    var rideReportAt: Int = -1
    var returnReported = false
    var tempParkReported = false

    override suspend fun ridePermission(
        carId: String,
        actor: RideActor,
        accessories: List<BleAccessoryReport>,
    ): RiderResult<ApiOutcome> = when {
        permissionCode.isNotBlank() -> RiderResult.Ok(fail(permissionCode))
        permissionFails -> RiderResult.Ok(fail("19999"))
        else -> RiderResult.Ok(ok())
    }

    override suspend fun rideReport(carId: String, actor: RideActor): RiderResult<ApiOutcome> {
        rideReportAt = writeLog.size
        if (reportFails) return RiderResult.Ok(fail("19999"))
        return RiderResult.Ok(ok(buildJsonObject { put("orderId", "ORDER-BLE") }))
    }

    override suspend fun returnReport(
        carId: String,
        orderId: String,
        actor: RideActor,
        accessories: List<BleAccessoryReport>,
    ): RiderResult<ApiOutcome> {
        returnReported = true
        return RiderResult.Ok(ok())
    }

    override suspend fun tempParkReport(carId: String, actor: RideActor): RiderResult<ApiOutcome> {
        tempParkReported = true
        onTempParkChanged(true)
        return RiderResult.Ok(ok())
    }

    override suspend fun endTempParkReport(
        carId: String,
        actor: RideActor,
    ): RiderResult<ApiOutcome> {
        onTempParkChanged(false)
        return RiderResult.Ok(ok())
    }

    override suspend fun returnConfig(carId: String): RiderResult<BleReturnConfig> =
        RiderResult.Ok(BleReturnConfig(needBeacon = false))
}

private class FakeApplyReturnRemote(private val fails: Boolean) : ApplyReturnRemote {
    override suspend fun canApply(applyType: Int?): RiderResult<Boolean> = RiderResult.Ok(!fails)

    override suspend fun createAudit(
        photoUrls: List<String>,
        orderId: String,
        applyType: Int?,
        reason: String,
    ): RiderResult<Unit> = if (fails) {
        RiderResult.Err(RiderError.business("19999", "rejected"))
    } else {
        RiderResult.Ok(Unit)
    }
}

private class FakeFenceRemote : FenceRemote {
    override suspend fun serviceAreaIdAt(at: GeoPoint): RiderResult<String> = RiderResult.Ok("SA-1")

    override suspend fun nearFences(
        at: GeoPoint,
        serviceId: String,
    ): RiderResult<List<FencePolygon>> = RiderResult.Ok(emptyList())

    override suspend fun nearParking(at: GeoPoint, serviceId: String): RiderResult<NearParking> =
        RiderResult.Ok(NearParking(count = 1))
}

private class FakeLocationTracker(private val fix: GeoPoint?) : LocationTracker {
    var tracking = false

    override suspend fun currentLocation(): RiderResult<GeoPoint> = fix
        ?.let { RiderResult.Ok(it) }
        ?: RiderResult.Err(RiderError.business("LOC", "no fix"))

    override fun track(): Flow<GeoPoint> = fix?.let { flowOf(it) } ?: flowOf()

    override fun startTracking() {
        tracking = true
    }

    override fun stopTracking() {
        tracking = false
    }
}

/** 记录写出去的每一帧，顺序断言全靠它。 */
private class RecordingBleTransport : BleTransport {
    private val delegate = SimulatorBleTransport()
    val writes = mutableListOf<String>()

    override val isAvailable: Boolean get() = delegate.isAvailable

    override fun connectionState() = delegate.connectionState()

    override suspend fun openAdapter() = delegate.openAdapter()

    override suspend fun closeAdapter() = delegate.closeAdapter()

    override suspend fun findByImei(imei: String, timeoutMs: Long) =
        delegate.findByImei(imei, timeoutMs)

    override suspend fun connect(deviceId: String, timeoutMs: Long) =
        delegate.connect(deviceId, timeoutMs)

    override suspend fun disconnect() = delegate.disconnect()

    override suspend fun writeAndAwait(hex: String, timeoutMs: Long) =
        delegate.writeAndAwait(hex, timeoutMs).also { writes += hex }
}
