package com.luopingtech.ebike.rider.domain.riding

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertNotEquals
import kotlin.test.assertSame
import kotlin.test.assertTrue

/**
 * 状态机全量单测。
 *
 * 除了主链路，这里刻意压了三类容易在真机上翻车的情形：
 * 1. 迟到事件落在已经翻页的相位上（BLE ack / 轮询与用户点击并发）
 * 2. 杀进程恢复的每一个相位
 * 3. `ServerSync` 与本地相位冲突时谁说了算
 */
class RideStateMachineTest {

    private val idle = RideSession.Idle

    private fun reduce(state: RideSession, vararg events: RideEvent): RideSession =
        RideStateMachine.reduceAll(state, events.toList())

    /** 走到某相位的最短事件序列，供各用例复用。 */
    private fun riding(): RideSession = reduce(
        idle,
        RideEvent.Scan("C1", "I1"),
        RideEvent.Confirm,
        RideEvent.UnlockSuccess(orderId = "O1", at = 1_000L),
    )

    // ── 主链路 ──────────────────────────────────────────────────────────────

    @Test
    fun `happy path walks idle to settling and back to idle`() {
        var s = idle
        assertEquals(RidePhase.Idle, s.phase)

        s = reduce(s, RideEvent.OpenScan)
        assertEquals(RidePhase.Scanning, s.phase)

        s = reduce(s, RideEvent.Scan("C1", "I1"))
        assertEquals(RidePhase.Confirming, s.phase)
        assertEquals("C1", s.carId)
        assertEquals("I1", s.imei)

        s = reduce(s, RideEvent.Confirm)
        assertEquals(RidePhase.Unlocking, s.phase)

        s = reduce(s, RideEvent.UnlockSuccess(orderId = "O1", at = 1_000L))
        assertEquals(RidePhase.Riding, s.phase)
        assertEquals("O1", s.orderId)
        assertEquals(1_000L, s.startedAtMillis)

        s = reduce(s, RideEvent.TempLock(at = 2_000L))
        assertEquals(RidePhase.TempLocked, s.phase)
        assertEquals(2_000L, s.tempLockedAtMillis)

        s = reduce(s, RideEvent.Resume)
        assertEquals(RidePhase.Riding, s.phase)
        assertEquals(0L, s.tempLockedAtMillis)

        s = reduce(s, RideEvent.StartReturn)
        assertEquals(RidePhase.Returning, s.phase)

        s = reduce(s, RideEvent.ReturnPermission(ReturnTypeCodes.NORMAL))
        assertEquals(ReturnTypeCodes.NORMAL, s.returnTypeCode)

        s = reduce(s, RideEvent.ReturnSuccess("O1"))
        assertEquals(RidePhase.Settling, s.phase)

        s = reduce(s, RideEvent.SettleDone)
        assertEquals(idle, s)
    }

    @Test
    fun `manual id reaches the same confirming state as a scan`() {
        val scanned = reduce(idle, RideEvent.Scan("C9"))
        val typed = reduce(idle, RideEvent.ManualId("C9"))
        assertEquals(scanned, typed)
        assertEquals(RidePhase.Confirming, typed.phase)
    }

    @Test
    fun `manual id trims whitespace from the typed number`() {
        val s = reduce(idle, RideEvent.ManualId("  C9  "))
        assertEquals("C9", s.carId)
    }

    @Test
    fun `blank scan payload keeps the user on the scanner`() {
        val s = reduce(idle, RideEvent.OpenScan, RideEvent.Scan("   "))
        assertEquals(RidePhase.Scanning, s.phase)
        assertFalse(s.hasVehicle)
    }

    // ── 开锁失败 ────────────────────────────────────────────────────────────

    @Test
    fun `unlock failure returns to confirming and counts the attempt`() {
        val s = reduce(
            idle,
            RideEvent.Scan("C1", "I1"),
            RideEvent.Confirm,
            RideEvent.UnlockFail("ble timeout"),
        )
        assertEquals(RidePhase.Confirming, s.phase)
        assertEquals("ble timeout", s.lastError)
        assertEquals(1, s.unlockAttempts)
        // 车牌要留着，用户重按一次就能再试。
        assertEquals("C1", s.carId)
    }

    @Test
    fun `repeated failures on the same vehicle accumulate attempts`() {
        val s = reduce(
            idle,
            RideEvent.Scan("C1"),
            RideEvent.Confirm,
            RideEvent.UnlockFail("e1"),
            RideEvent.Confirm,
            RideEvent.UnlockFail("e2"),
            RideEvent.Confirm,
            RideEvent.UnlockFail("e3"),
        )
        assertEquals(3, s.unlockAttempts)
        assertEquals("e3", s.lastError)
    }

    @Test
    fun `switching to another vehicle resets the attempt counter`() {
        val s = reduce(
            idle,
            RideEvent.Scan("C1"),
            RideEvent.Confirm,
            RideEvent.UnlockFail("e1"),
            RideEvent.Scan("C2"),
        )
        assertEquals("C2", s.carId)
        assertEquals(0, s.unlockAttempts)
    }

    @Test
    fun `rescanning the same vehicle keeps the attempt counter`() {
        val s = reduce(
            idle,
            RideEvent.Scan("C1"),
            RideEvent.Confirm,
            RideEvent.UnlockFail("e1"),
            RideEvent.Scan("C1"),
        )
        assertEquals(1, s.unlockAttempts)
    }

    @Test
    fun `unlock success clears the previous error and attempt count`() {
        val s = reduce(
            idle,
            RideEvent.Scan("C1"),
            RideEvent.Confirm,
            RideEvent.UnlockFail("e1"),
            RideEvent.Confirm,
            RideEvent.UnlockSuccess(orderId = "O1", at = 5L),
        )
        assertEquals(RidePhase.Riding, s.phase)
        assertEquals("", s.lastError)
        assertEquals(0, s.unlockAttempts)
    }

    @Test
    fun `network fallback flag rides along into the session`() {
        val s = reduce(
            idle,
            RideEvent.Scan("C1"),
            RideEvent.Confirm,
            RideEvent.UnlockSuccess(orderId = "O1", at = 5L, usedNetworkFallback = true),
        )
        assertTrue(s.usedNetworkFallback)
    }

    // ── 迟到 / 越序事件 ─────────────────────────────────────────────────────

    @Test
    fun `duplicate unlock success while riding is idempotent`() {
        val first = riding()
        val second = reduce(first, RideEvent.UnlockSuccess(orderId = "O-late", at = 9_999L))
        // 已经在骑了，第二份 ack 不许改订单号或计时基准。
        assertEquals(first, second)
    }

    @Test
    fun `unlock success straight from confirming is accepted`() {
        // BLE 上报比 Confirm 的相位写入还快时会走到这里。
        val s = reduce(idle, RideEvent.Scan("C1"), RideEvent.UnlockSuccess("O1", at = 1L))
        assertEquals(RidePhase.Riding, s.phase)
    }

    @Test
    fun `late unlock failure while riding is ignored`() {
        val s = riding()
        assertEquals(s, reduce(s, RideEvent.UnlockFail("late")))
    }

    @Test
    fun `temp lock is ignored unless the user is actually riding`() {
        val confirming = reduce(idle, RideEvent.Scan("C1"))
        assertEquals(confirming, reduce(confirming, RideEvent.TempLock(at = 1L)))

        val tempLocked = reduce(riding(), RideEvent.TempLock(at = 1L))
        // 已临停，重复的临停 ack 不刷新时间戳。
        assertEquals(tempLocked, reduce(tempLocked, RideEvent.TempLock(at = 7L)))
        assertEquals(1L, tempLocked.tempLockedAtMillis)
    }

    @Test
    fun `resume is ignored unless the ride is temp locked`() {
        val s = riding()
        assertEquals(s, reduce(s, RideEvent.Resume))
    }

    @Test
    fun `return permission outside the returning phase is dropped`() {
        val s = riding()
        assertEquals(s, reduce(s, RideEvent.ReturnPermission(ReturnTypeCodes.OUT_OF_SPOT)))
    }

    @Test
    fun `settle done only fires from settling`() {
        val s = riding()
        assertEquals(s, reduce(s, RideEvent.SettleDone))
    }

    @Test
    fun `open scan is refused while a ride is in progress`() {
        val s = riding()
        assertEquals(s, reduce(s, RideEvent.OpenScan))
        // 更要紧的是不能换车。
        assertEquals(s, reduce(s, RideEvent.Scan("C-OTHER")))
    }

    @Test
    fun `reset cannot drop a ride in progress`() {
        val ridingSession = riding()
        assertEquals(ridingSession, reduce(ridingSession, RideEvent.Reset))

        val tempLocked = reduce(ridingSession, RideEvent.TempLock(at = 1L))
        assertEquals(tempLocked, reduce(tempLocked, RideEvent.Reset))
    }

    @Test
    fun `reset clears an unconfirmed vehicle`() {
        val s = reduce(idle, RideEvent.Scan("C1"), RideEvent.Reset)
        assertEquals(idle, s)
    }

    @Test
    fun `reset from settling clears the session`() {
        val s = reduce(riding(), RideEvent.ReturnSuccess("O1"), RideEvent.Reset)
        assertEquals(idle, s)
    }

    // ── 还车 ────────────────────────────────────────────────────────────────

    @Test
    fun `return success is accepted straight from riding for code 15009`() {
        // 服务端说「订单已结束」时没有经过 Returning。
        val s = reduce(riding(), RideEvent.ReturnSuccess("O1"))
        assertEquals(RidePhase.Settling, s.phase)
    }

    @Test
    fun `return success from temp locked also settles`() {
        val s = reduce(riding(), RideEvent.TempLock(at = 1L), RideEvent.ReturnSuccess("O1"))
        assertEquals(RidePhase.Settling, s.phase)
    }

    @Test
    fun `return success keeps the existing order id when the event carries none`() {
        val s = reduce(riding(), RideEvent.StartReturn, RideEvent.ReturnSuccess())
        assertEquals("O1", s.orderId)
    }

    @Test
    fun `return failure falls back to riding with the vehicle intact`() {
        val s = reduce(riding(), RideEvent.StartReturn, RideEvent.Error("fence rejected"))
        assertEquals(RidePhase.Riding, s.phase)
        assertEquals("fence rejected", s.lastError)
        assertEquals("C1", s.carId)
        assertEquals("O1", s.orderId)
    }

    @Test
    fun `return can be retried after a failure`() {
        val s = reduce(
            riding(),
            RideEvent.StartReturn,
            RideEvent.Error(""),
            RideEvent.StartReturn,
            RideEvent.ReturnSuccess("O1"),
        )
        assertEquals(RidePhase.Settling, s.phase)
    }

    @Test
    fun `start return is refused before the ride begins`() {
        val confirming = reduce(idle, RideEvent.Scan("C1"))
        assertEquals(confirming, reduce(confirming, RideEvent.StartReturn))
    }

    // ── Error ───────────────────────────────────────────────────────────────

    @Test
    fun `error during unlocking goes back to confirming`() {
        val s = reduce(idle, RideEvent.Scan("C1"), RideEvent.Confirm, RideEvent.Error("boom"))
        assertEquals(RidePhase.Confirming, s.phase)
        assertEquals(1, s.unlockAttempts)
    }

    @Test
    fun `error while scanning drops back to idle`() {
        val s = reduce(idle, RideEvent.OpenScan, RideEvent.Error("camera denied"))
        assertEquals(RidePhase.Idle, s.phase)
        assertEquals("camera denied", s.lastError)
    }

    @Test
    fun `error while riding only records the reason`() {
        val s = reduce(riding(), RideEvent.Error("gps lost"))
        assertEquals(RidePhase.Riding, s.phase)
        assertEquals("gps lost", s.lastError)
    }

    // ── 杀进程恢复 ──────────────────────────────────────────────────────────

    @Test
    fun `restore keeps riding temp locked and settling untouched`() {
        for (phase in listOf(RidePhase.Riding, RidePhase.TempLocked, RidePhase.Settling)) {
            val snapshot = RideSession(phase = phase, carId = "C1", orderId = "O1")
            assertEquals(snapshot, RideStateMachine.sanitizeRestored(snapshot))
            assertEquals(snapshot, reduce(idle, RideEvent.KillRestore(snapshot)))
        }
    }

    @Test
    fun `restore rewinds unlocking to confirming so the user can retry`() {
        val snapshot = RideSession(phase = RidePhase.Unlocking, carId = "C1", imei = "I1")
        val s = RideStateMachine.sanitizeRestored(snapshot)
        assertEquals(RidePhase.Confirming, s.phase)
        assertEquals("C1", s.carId)
        assertEquals("I1", s.imei)
    }

    @Test
    fun `restore rewinds returning to riding because the bike is still held`() {
        val snapshot = RideSession(phase = RidePhase.Returning, carId = "C1", orderId = "O1")
        val s = RideStateMachine.sanitizeRestored(snapshot)
        assertEquals(RidePhase.Riding, s.phase)
        assertEquals("O1", s.orderId)
    }

    @Test
    fun `restore drops scanning because a camera session cannot survive`() {
        val snapshot = RideSession(phase = RidePhase.Scanning, carId = "C1")
        assertEquals(idle, RideStateMachine.sanitizeRestored(snapshot))
    }

    @Test
    fun `restore keeps confirming only when a vehicle is known`() {
        val withVehicle = RideSession(phase = RidePhase.Confirming, carId = "C1")
        assertEquals(withVehicle, RideStateMachine.sanitizeRestored(withVehicle))

        val without = RideSession(phase = RidePhase.Confirming)
        assertEquals(idle, RideStateMachine.sanitizeRestored(without))
    }

    @Test
    fun `restore drops half states that lost their vehicle`() {
        for (phase in listOf(RidePhase.Unlocking, RidePhase.Returning)) {
            assertEquals(idle, RideStateMachine.sanitizeRestored(RideSession(phase = phase)))
        }
    }

    @Test
    fun `restore is idempotent`() {
        val snapshots = RidePhase.entries.map { RideSession(phase = it, carId = "C1", orderId = "O1") }
        for (snapshot in snapshots) {
            val once = RideStateMachine.sanitizeRestored(snapshot)
            assertEquals(once, RideStateMachine.sanitizeRestored(once), "phase=${snapshot.phase}")
        }
    }

    @Test
    fun `restore never lands on a transient phase`() {
        val transient = setOf(RidePhase.Scanning, RidePhase.Unlocking, RidePhase.Returning)
        for (phase in RidePhase.entries) {
            val restored = RideStateMachine.sanitizeRestored(
                RideSession(phase = phase, carId = "C1", orderId = "O1"),
            )
            assertFalse(restored.phase in transient, "phase=$phase restored to ${restored.phase}")
        }
    }

    @Test
    fun `kill restore overwrites whatever the current state was`() {
        val snapshot = RideSession(phase = RidePhase.Riding, carId = "C-DISK", orderId = "O-DISK")
        val s = reduce(reduce(idle, RideEvent.Scan("C-MEM")), RideEvent.KillRestore(snapshot))
        assertEquals("C-DISK", s.carId)
        assertEquals(RidePhase.Riding, s.phase)
    }

    // ── ServerSync ──────────────────────────────────────────────────────────

    @Test
    fun `server sync lifts a stale idle state into riding`() {
        // 冷启动读盘失败（比如换了词表清了快照），只靠服务端也要能回到骑行页。
        val s = reduce(
            idle,
            RideEvent.ServerSync(ridingState = 4, carId = "C1", imei = "I1", orderId = "O1"),
        )
        assertEquals(RidePhase.Riding, s.phase)
        assertEquals("C1", s.carId)
        assertEquals("I1", s.imei)
        assertEquals("O1", s.orderId)
    }

    @Test
    fun `all documented riding states map to the riding phase`() {
        for (code in RideStateMachine.SERVER_STATES_RIDING) {
            val s = reduce(idle, RideEvent.ServerSync(ridingState = code, carId = "C1"))
            assertEquals(RidePhase.Riding, s.phase, "ridingState=$code")
        }
    }

    @Test
    fun `server temp park state moves the local phase into temp locked`() {
        val s = reduce(
            riding(),
            RideEvent.ServerSync(ridingState = RideStateMachine.SERVER_STATE_TEMP_PARK),
        )
        assertEquals(RidePhase.TempLocked, s.phase)
    }

    @Test
    fun `server riding state pulls the ride back out of temp lock`() {
        // 用户在车上按了物理解锁键，服务端先知道。
        val s = reduce(
            riding(),
            RideEvent.TempLock(at = 1L),
            RideEvent.ServerSync(ridingState = 5),
        )
        assertEquals(RidePhase.Riding, s.phase)
    }

    @Test
    fun `server sync does not disturb a return in flight`() {
        val returning = reduce(riding(), RideEvent.StartReturn)
        for (code in listOf(RideStateMachine.SERVER_STATE_TEMP_PARK, 4, 5, 6)) {
            val s = reduce(returning, RideEvent.ServerSync(ridingState = code))
            assertEquals(RidePhase.Returning, s.phase, "ridingState=$code")
        }
    }

    @Test
    fun `server sync does not disturb settling`() {
        val settling = reduce(riding(), RideEvent.ReturnSuccess("O1"))
        for (code in listOf(RideStateMachine.SERVER_STATE_TEMP_PARK, 4, 5, 6)) {
            assertEquals(RidePhase.Settling, reduce(settling, RideEvent.ServerSync(code)).phase)
        }
    }

    @Test
    fun `no active order while riding settles only when orderEnded`() {
        // 开锁后首轮 ridingState 缺失：必须留在骑行，不能误进结费。
        val soft = reduce(riding(), RideEvent.ServerSync(ridingState = null))
        assertEquals(RidePhase.Riding, soft.phase)

        // 运维远程还车 / 15009：显式 orderEnded 才结费。
        val s = reduce(riding(), RideEvent.ServerSync(ridingState = null, orderEnded = true))
        assertEquals(RidePhase.Settling, s.phase)
    }

    @Test
    fun `no active order while returning settles only when orderEnded`() {
        val soft = reduce(riding(), RideEvent.StartReturn, RideEvent.ServerSync(ridingState = null))
        assertEquals(RidePhase.Returning, soft.phase)

        val s = reduce(
            riding(),
            RideEvent.StartReturn,
            RideEvent.ServerSync(ridingState = null, orderEnded = true),
        )
        assertEquals(RidePhase.Settling, s.phase)
    }

    @Test
    fun `no active order leaves pre-ride phases alone`() {
        for (phase in listOf(RidePhase.Idle, RidePhase.Scanning, RidePhase.Confirming)) {
            val before = RideSession(phase = phase, carId = "C1")
            val after = RideStateMachine.reduce(before, RideEvent.ServerSync(ridingState = null))
            assertEquals(phase, after.phase)
        }
    }

    @Test
    fun `unknown server riding state keeps the trip like UniApp`() {
        // UniApp: Number(ridingState)!==3 → 仍按骑行展示，不踢结费。
        val s = reduce(riding(), RideEvent.ServerSync(ridingState = 99))
        assertEquals(RidePhase.Riding, s.phase)
    }

    @Test
    fun `server riding state from idle resumes the trip`() {
        val s = reduce(
            RideSession.Idle,
            RideEvent.ServerSync(ridingState = 4, carId = "C9", orderId = "O9"),
        )
        assertEquals(RidePhase.Riding, s.phase)
        assertEquals("C9", s.carId)
        assertEquals("O9", s.orderId)
    }

    @Test
    fun `server sync never blanks fields it does not carry`() {
        val s = reduce(riding(), RideEvent.ServerSync(ridingState = 4))
        assertEquals("C1", s.carId)
        assertEquals("I1", s.imei)
        assertEquals("O1", s.orderId)
        assertEquals(1_000L, s.startedAtMillis)
    }

    @Test
    fun `server sync can supply a start timestamp the local snapshot lacks`() {
        val s = reduce(
            idle,
            RideEvent.ServerSync(ridingState = 4, carId = "C1", startedAtMillis = 42L),
        )
        assertEquals(42L, s.startedAtMillis)
    }

    @Test
    fun `server sync repeated while riding is stable`() {
        val once = reduce(riding(), RideEvent.ServerSync(ridingState = 4, carId = "C1"))
        val twice = reduce(once, RideEvent.ServerSync(ridingState = 4, carId = "C1"))
        assertEquals(once, twice)
    }

    // ── imei 补齐 ───────────────────────────────────────────────────────────

    @Test
    fun `vehicle loaded fills in the imei the qr code lacked`() {
        val s = reduce(idle, RideEvent.Scan("C1"), RideEvent.VehicleLoaded("C1", "I-FROM-API"))
        assertEquals("I-FROM-API", s.imei)
        assertEquals(RidePhase.Confirming, s.phase)
    }

    @Test
    fun `vehicle loaded may fill the imei mid-ride for a weak-network return`() {
        val s = reduce(
            idle,
            RideEvent.Scan("C1"),
            RideEvent.Confirm,
            RideEvent.UnlockSuccess("O1", at = 1L),
            RideEvent.VehicleLoaded("C1", "I-LATE"),
        )
        assertEquals("I-LATE", s.imei)
        assertEquals(RidePhase.Riding, s.phase)
    }

    @Test
    fun `vehicle loaded for another car is refused mid-ride`() {
        val s = riding()
        assertEquals(s, reduce(s, RideEvent.VehicleLoaded("C-OTHER", "I-OTHER")))
    }

    @Test
    fun `blank imei never overwrites a known one`() {
        val s = reduce(idle, RideEvent.Scan("C1", "I1"), RideEvent.VehicleLoaded("C1", ""))
        assertEquals("I1", s.imei)
    }

    // ── 派生属性 ────────────────────────────────────────────────────────────

    @Test
    fun `foreground location covers unlock through return`() {
        val needing = listOf(
            RidePhase.Unlocking,
            RidePhase.Riding,
            RidePhase.TempLocked,
            RidePhase.Returning,
        )
        for (phase in RidePhase.entries) {
            assertEquals(
                phase in needing,
                RideSession(phase = phase).needsForegroundLocation,
                "phase=$phase",
            )
        }
    }

    @Test
    fun `is on trip covers exactly riding and temp locked`() {
        for (phase in RidePhase.entries) {
            assertEquals(
                phase == RidePhase.Riding || phase == RidePhase.TempLocked,
                RideSession(phase = phase).isOnTrip,
                "phase=$phase",
            )
        }
    }

    // ── 不变式 ──────────────────────────────────────────────────────────────

    @Test
    fun `unhandled combinations return the identical instance`() {
        // 相同实例（而不只是相等）让 RidingFeature.dispatch 的短路判断成立。
        val settling = reduce(riding(), RideEvent.ReturnSuccess("O1"))
        val ignored = listOf(
            RideEvent.OpenScan,
            RideEvent.Scan("C2"),
            RideEvent.ManualId("C2"),
            RideEvent.Confirm,
            RideEvent.UnlockFail("x"),
            RideEvent.TempLock(1L),
            RideEvent.Resume,
            RideEvent.StartReturn,
            RideEvent.ReturnPermission(ReturnTypeCodes.NORMAL),
        )
        for (event in ignored) {
            assertSame(settling, RideStateMachine.reduce(settling, event), "event=$event")
        }
    }

    @Test
    fun `no event sequence can reach riding without a vehicle`() {
        val s = reduce(idle, RideEvent.Confirm, RideEvent.UnlockSuccess("O1", at = 1L))
        assertEquals(RidePhase.Idle, s.phase)
        assertNotEquals(RidePhase.Riding, s.phase)
    }

    @Test
    fun `reduce all folds in order`() {
        val events = listOf(
            RideEvent.Scan("C1", "I1"),
            RideEvent.Confirm,
            RideEvent.UnlockSuccess("O1", at = 1L),
            RideEvent.TempLock(2L),
        )
        assertEquals(RidePhase.TempLocked, RideStateMachine.reduceAll(idle, events).phase)
        // 顺序反了就到不了。
        assertEquals(RidePhase.Confirming, RideStateMachine.reduceAll(idle, events.reversed()).phase)
    }
}
