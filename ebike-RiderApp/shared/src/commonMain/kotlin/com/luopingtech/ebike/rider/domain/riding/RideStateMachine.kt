package com.luopingtech.ebike.rider.domain.riding

import kotlinx.serialization.Serializable

/**
 * 骑行主循环的相位。
 *
 * `Idle → Scanning → Confirming → Unlocking → Riding → TempLocked → Returning → Settling`，
 * [Settling] 结束回 [Idle]。UniApp 那边这套状态散在 `tempData.ride.status`、
 * `getRideInfo().ridingState` 和各页面的 `ref` 里，任何一处漏改就会出现「界面在骑行、
 * store 说已结束」。这里收成单一相位，转移只走 [RideStateMachine.reduce]。
 */
enum class RidePhase {
    Idle,
    Scanning,
    Confirming,
    Unlocking,
    Riding,
    TempLocked,
    Returning,
    Settling,
}

/**
 * 一次骑行的全部可持久化状态。字段刻意都是原始类型：这份快照要过 [com.luopingtech.ebike.rider.platform.SecureStore]，
 * 杀进程后原样读回。
 */
@Serializable
data class RideSession(
    val phase: RidePhase = RidePhase.Idle,
    val carId: String = "",
    val imei: String = "",
    val orderId: String = "",
    /** 解锁成功的时刻，计时基准。0 = 还没开始。 */
    val startedAtMillis: Long = 0L,
    /** 进临时锁车的时刻，用于「临停已 N 分钟」与自动还车倒计时。 */
    val tempLockedAtMillis: Long = 0L,
    /** `returnPermission.returnType`（1101/2101/…），不是 `returnByNet` 的 0|1。 */
    val returnTypeCode: Int = 0,
    /** 最近一次失败原因，进 [RidePhase.Confirming] / [RidePhase.Riding] 的提示条。 */
    val lastError: String = "",
    /** 同一辆车连续解锁失败次数，UI 用来决定要不要提示换车。 */
    val unlockAttempts: Int = 0,
    /** 走过网络远程开锁降级（诊断用，进 Debug 屏）。 */
    val usedNetworkFallback: Boolean = false,
) {
    val hasVehicle: Boolean get() = carId.isNotBlank()

    /** 计时 / 上报轨迹只在这两个相位里有意义。 */
    val isOnTrip: Boolean get() = phase == RidePhase.Riding || phase == RidePhase.TempLocked

    /** 前台定位服务的开关条件：从解锁开始一直到还车提交完。 */
    val needsForegroundLocation: Boolean
        get() = phase == RidePhase.Unlocking ||
            phase == RidePhase.Riding ||
            phase == RidePhase.TempLocked ||
            phase == RidePhase.Returning

    companion object {
        val Idle: RideSession = RideSession()
    }
}

/**
 * 主循环事件。命名对齐交付清单，[KillRestore] 是唯一一个「从外部灌入整份状态」的事件。
 */
sealed interface RideEvent {
    /** 点「扫码」 */
    data object OpenScan : RideEvent

    /** 扫码识别成功。UniApp 二维码里可能只有 carId，imei 要靠 `getCarInfo` 补。 */
    data class Scan(val carId: String, val imei: String = "") : RideEvent

    /** 手输车牌号（`pages-sub/ride/enter-id`）。 */
    data class ManualId(val carId: String) : RideEvent

    /** 车辆详情读到了，进确认页；也用于在 [RidePhase.Confirming] 内补齐 imei。 */
    data class VehicleLoaded(val carId: String, val imei: String) : RideEvent

    /** 确认页点「开锁」 */
    data object Confirm : RideEvent

    data class UnlockSuccess(
        val orderId: String,
        val at: Long,
        val usedNetworkFallback: Boolean = false,
    ) : RideEvent

    data class UnlockFail(val message: String) : RideEvent

    data class TempLock(val at: Long) : RideEvent

    data object Resume : RideEvent

    data object StartReturn : RideEvent

    /** `returnPermission` 回来了，带上围栏判定码。 */
    data class ReturnPermission(val returnTypeCode: Int) : RideEvent

    data class ReturnSuccess(val orderId: String = "") : RideEvent

    /** 结费屏走完（支付未接通时是「知道了」）。 */
    data object SettleDone : RideEvent

    /**
     * 可恢复的失败：退回上一个稳定相位并挂上原因，不清空车辆信息。
     */
    data class Error(val message: String) : RideEvent

    /** 用户主动放弃当前未开锁的车辆，或结费完毕后清场。 */
    data object Reset : RideEvent

    /**
     * 冷启动从 [com.luopingtech.ebike.rider.platform.SecureStore] 读回的快照。
     * 语义见 [RideStateMachine.sanitizeRestored]。
     */
    data class KillRestore(val restored: RideSession) : RideEvent

    /**
     * 服务端订单状态权威回灌（`getRideInfo`）。本地快照与它冲突时以它为准 ——
     * 车可能被运维远程还了，也可能上一次还车请求实际成功但响应丢了。
     *
     * [orderEnded]=true 表示业务码明确结束（如 15009），才允许从骑行推进到结费。
     * 仅 `ridingState == null` **不能**当作结束：开锁后首轮轮询常缺字段，UniApp 对 15042 是软忽略。
     */
    data class ServerSync(
        val ridingState: Int?,
        val carId: String = "",
        val imei: String = "",
        val orderId: String = "",
        val startedAtMillis: Long = 0L,
        val orderEnded: Boolean = false,
    ) : RideEvent
}

/**
 * 纯函数状态机。没有协程、没有 IO、没有时钟 —— 时间戳一律由事件带进来，
 * 这样每条转移都能在单测里钉死。
 *
 * 未定义的（相位, 事件）组合一律**原样返回**，不抛异常：BLE 回调、轮询和用户点击
 * 会并发到达，晚到的 ack 落在已经翻页的相位上属于常态。
 */
object RideStateMachine {

    /** 服务端 `getRideInfo().ridingState`：3 = 临停，4/5/6 = 骑行中。 */
    const val SERVER_STATE_TEMP_PARK: Int = 3
    val SERVER_STATES_RIDING: Set<Int> = setOf(4, 5, 6)

    fun reduce(state: RideSession, event: RideEvent): RideSession = when (event) {
        is RideEvent.OpenScan -> when (state.phase) {
            RidePhase.Idle, RidePhase.Scanning, RidePhase.Confirming ->
                state.copy(phase = RidePhase.Scanning, lastError = "")
            else -> state
        }

        is RideEvent.Scan -> when (state.phase) {
            RidePhase.Idle, RidePhase.Scanning, RidePhase.Confirming -> selectVehicle(
                state = state,
                carId = event.carId,
                imei = event.imei,
            )
            else -> state
        }

        is RideEvent.ManualId -> when (state.phase) {
            RidePhase.Idle, RidePhase.Scanning, RidePhase.Confirming -> selectVehicle(
                state = state,
                carId = event.carId,
                imei = "",
            )
            else -> state
        }

        is RideEvent.VehicleLoaded -> when (state.phase) {
            RidePhase.Idle, RidePhase.Scanning, RidePhase.Confirming -> selectVehicle(
                state = state,
                carId = event.carId,
                imei = event.imei,
            )
            // 骑行中补 imei（弱网还车要用），但不许改车。
            RidePhase.Riding, RidePhase.TempLocked ->
                if (event.carId.isNotBlank() && event.carId != state.carId) {
                    state
                } else {
                    state.copy(imei = event.imei.ifBlank { state.imei })
                }
            else -> state
        }

        is RideEvent.Confirm ->
            if (state.phase == RidePhase.Confirming && state.hasVehicle) {
                state.copy(phase = RidePhase.Unlocking, lastError = "")
            } else {
                state
            }

        is RideEvent.UnlockSuccess -> when (state.phase) {
            // 幂等：BLE 上报和网络降级都可能各回一次成功。
            RidePhase.Unlocking, RidePhase.Confirming -> state.copy(
                phase = RidePhase.Riding,
                orderId = event.orderId.ifBlank { state.orderId },
                startedAtMillis = if (event.at > 0L) event.at else state.startedAtMillis,
                lastError = "",
                unlockAttempts = 0,
                usedNetworkFallback = event.usedNetworkFallback,
            )
            else -> state
        }

        is RideEvent.UnlockFail ->
            if (state.phase == RidePhase.Unlocking) {
                state.copy(
                    phase = RidePhase.Confirming,
                    lastError = event.message,
                    unlockAttempts = state.unlockAttempts + 1,
                )
            } else {
                state
            }

        is RideEvent.TempLock ->
            if (state.phase == RidePhase.Riding) {
                state.copy(
                    phase = RidePhase.TempLocked,
                    tempLockedAtMillis = event.at,
                    lastError = "",
                )
            } else {
                state
            }

        is RideEvent.Resume ->
            if (state.phase == RidePhase.TempLocked) {
                state.copy(
                    phase = RidePhase.Riding,
                    tempLockedAtMillis = 0L,
                    lastError = "",
                )
            } else {
                state
            }

        is RideEvent.StartReturn -> when (state.phase) {
            RidePhase.Riding, RidePhase.TempLocked ->
                state.copy(phase = RidePhase.Returning, lastError = "")
            else -> state
        }

        is RideEvent.ReturnPermission ->
            if (state.phase == RidePhase.Returning) {
                state.copy(returnTypeCode = event.returnTypeCode)
            } else {
                state
            }

        is RideEvent.ReturnSuccess -> when (state.phase) {
            // 15009「订单已结束」也走这里，所以 Riding / TempLocked 直接进结费也放行。
            RidePhase.Returning, RidePhase.Riding, RidePhase.TempLocked -> state.copy(
                phase = RidePhase.Settling,
                orderId = event.orderId.ifBlank { state.orderId },
                lastError = "",
            )
            else -> state
        }

        is RideEvent.SettleDone ->
            if (state.phase == RidePhase.Settling) RideSession.Idle else state

        is RideEvent.Error -> when (state.phase) {
            // 还车失败要退回骑行，用户手上还有车。
            RidePhase.Returning -> state.copy(phase = RidePhase.Riding, lastError = event.message)
            RidePhase.Unlocking -> state.copy(
                phase = RidePhase.Confirming,
                lastError = event.message,
                unlockAttempts = state.unlockAttempts + 1,
            )
            RidePhase.Scanning -> state.copy(phase = RidePhase.Idle, lastError = event.message)
            RidePhase.Idle -> state.copy(lastError = event.message)
            else -> state.copy(lastError = event.message)
        }

        is RideEvent.Reset ->
            // 骑行中不允许被「返回」清掉，否则界面回首页而车还开着。
            if (state.isOnTrip) state else RideSession.Idle

        is RideEvent.KillRestore -> sanitizeRestored(event.restored)

        is RideEvent.ServerSync -> applyServerSync(state, event)
    }

    fun reduceAll(state: RideSession, events: List<RideEvent>): RideSession =
        events.fold(state, ::reduce)

    private fun selectVehicle(state: RideSession, carId: String, imei: String): RideSession {
        val id = carId.trim()
        if (id.isEmpty()) {
            return state.copy(phase = RidePhase.Scanning, lastError = "")
        }
        return state.copy(
            phase = RidePhase.Confirming,
            carId = id,
            imei = imei.ifBlank { state.imei },
            lastError = "",
            unlockAttempts = if (id != state.carId) 0 else state.unlockAttempts,
        )
    }

    /**
     * 杀进程恢复。原则：**只信稳定相位**。
     *
     * `Unlocking` / `Returning` 是「请求已发出、响应未确认」的窗口，进程死在那里时本地
     * 无法判断车到底开没开 —— 恢复成瞬时相位会让界面卡在转圈上。所以：
     *
     * - `Riding` / `TempLocked` / `Settling` → 原样保留（随后 `getRideInfo` 校正）
     * - `Unlocking` → `Confirming`：保留车辆，让用户重按一次；订单若已建，`ServerSync` 会把它抬进 `Riding`
     * - `Returning` → `Riding`：车还在手上，重走还车
     * - `Scanning` → `Idle`：相机会话不可能跨进程
     * - `Confirming` → 有车牌就留着，否则 `Idle`
     */
    fun sanitizeRestored(restored: RideSession): RideSession = when (restored.phase) {
        RidePhase.Riding, RidePhase.TempLocked, RidePhase.Settling -> restored
        RidePhase.Unlocking ->
            if (restored.hasVehicle) {
                restored.copy(phase = RidePhase.Confirming)
            } else {
                RideSession.Idle
            }
        RidePhase.Returning ->
            if (restored.hasVehicle) restored.copy(phase = RidePhase.Riding) else RideSession.Idle
        RidePhase.Confirming ->
            if (restored.hasVehicle) restored else RideSession.Idle
        RidePhase.Scanning, RidePhase.Idle -> RideSession.Idle
    }

    /**
     * 服务端权威状态回灌。
     *
     * 对齐 UniApp `riding.vue`：
     * - `ridingState` 3 → 临停；4/5/6 → 骑行中；其它非空值也按骑行中展示（不踢结费）
     * - 仅业务码明确结束（[RideEvent.ServerSync.orderEnded]）才进结费
     * - `ridingState == null` 且未结束：软保留当前相位（开锁后首轮常缺字段 / 15042）
     */
    private fun applyServerSync(state: RideSession, event: RideEvent.ServerSync): RideSession {
        val patched = state.copy(
            carId = event.carId.ifBlank { state.carId },
            imei = event.imei.ifBlank { state.imei },
            orderId = event.orderId.ifBlank { state.orderId },
            startedAtMillis = if (event.startedAtMillis > 0L) {
                event.startedAtMillis
            } else {
                state.startedAtMillis
            },
        )
        if (event.orderEnded) {
            return when {
                patched.isOnTrip || patched.phase == RidePhase.Returning ->
                    patched.copy(phase = RidePhase.Settling)
                else -> patched
            }
        }
        val serverState = event.ridingState
        return when {
            serverState == SERVER_STATE_TEMP_PARK -> when (patched.phase) {
                // 还车提交中不要被轮询打回临停。
                RidePhase.Returning, RidePhase.Settling -> patched
                else -> patched.copy(phase = RidePhase.TempLocked, lastError = "")
            }
            serverState in SERVER_STATES_RIDING -> when (patched.phase) {
                RidePhase.Returning, RidePhase.Settling -> patched
                RidePhase.Riding -> patched
                else -> patched.copy(phase = RidePhase.Riding, lastError = "")
            }
            // null / 0 / 未知：开锁后首轮常缺字段；Idle 上也不要因脏值误抬骑行。
            // 已在行程中则软保留当前相位（对齐 UniApp 15042 / 缺字段不踢页）。
            patched.phase == RidePhase.Riding ||
                patched.phase == RidePhase.TempLocked ||
                patched.phase == RidePhase.Returning -> patched
            else -> patched
        }
    }
}
