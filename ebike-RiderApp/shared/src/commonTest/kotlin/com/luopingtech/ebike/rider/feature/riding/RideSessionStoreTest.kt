package com.luopingtech.ebike.rider.feature.riding

import com.luopingtech.ebike.rider.domain.riding.RidePhase
import com.luopingtech.ebike.rider.domain.riding.RideSession
import com.luopingtech.ebike.rider.platform.InMemorySecureStore
import com.luopingtech.ebike.rider.platform.SecureStore
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertNull
import kotlin.test.assertTrue

/** 杀进程恢复的落盘 / 读回。 */
class RideSessionStoreTest {

    private fun store(): Pair<RideSessionStore, InMemorySecureStore> {
        val secure = InMemorySecureStore()
        return RideSessionStore(secure) to secure
    }

    @Test
    fun `an empty store restores to idle`() {
        val (rides, _) = store()
        assertEquals(RideSession.Idle, rides.load())
    }

    @Test
    fun `a riding session survives a round trip intact`() {
        val (rides, _) = store()
        val session = RideSession(
            phase = RidePhase.Riding,
            carId = "C1",
            imei = "I1",
            orderId = "O1",
            startedAtMillis = 1_700_000_000_000L,
            returnTypeCode = 1101,
            usedNetworkFallback = true,
        )
        rides.save(session)
        assertEquals(session, rides.load())
    }

    @Test
    fun `a temp locked session keeps its pause timestamp`() {
        val (rides, _) = store()
        val session = RideSession(
            phase = RidePhase.TempLocked,
            carId = "C1",
            orderId = "O1",
            startedAtMillis = 1_000L,
            tempLockedAtMillis = 2_000L,
        )
        rides.save(session)
        assertEquals(2_000L, rides.load().tempLockedAtMillis)
    }

    @Test
    fun `loading runs the restore sanitizer`() {
        val (rides, _) = store()
        rides.save(RideSession(phase = RidePhase.Unlocking, carId = "C1", imei = "I1"))
        // 磁盘上是 Unlocking，读回必须是可操作的 Confirming。
        val restored = rides.load()
        assertEquals(RidePhase.Confirming, restored.phase)
        assertEquals("C1", restored.carId)
    }

    @Test
    fun `a returning snapshot is restored as riding`() {
        val (rides, _) = store()
        rides.save(RideSession(phase = RidePhase.Returning, carId = "C1", orderId = "O1"))
        assertEquals(RidePhase.Riding, rides.load().phase)
    }

    @Test
    fun `scanning is downgraded before it ever hits the disk`() {
        val (rides, secure) = store()
        rides.save(RideSession(phase = RidePhase.Scanning, carId = "C1"))
        val raw = secure.getString(SecureStore.KEY_RIDE_SESSION)
        assertTrue(raw != null && !raw.contains("Scanning"), "raw=$raw")
        assertEquals(RideSession.Idle, rides.load())
    }

    @Test
    fun `saving an empty idle session clears the slot`() {
        val (rides, secure) = store()
        rides.save(RideSession(phase = RidePhase.Riding, carId = "C1", orderId = "O1"))
        assertTrue(secure.getString(SecureStore.KEY_RIDE_SESSION) != null)

        rides.save(RideSession.Idle)
        assertNull(secure.getString(SecureStore.KEY_RIDE_SESSION))
    }

    @Test
    fun `clear wipes the snapshot`() {
        val (rides, _) = store()
        rides.save(RideSession(phase = RidePhase.Riding, carId = "C1"))
        rides.clear()
        assertEquals(RideSession.Idle, rides.load())
    }

    @Test
    fun `a corrupt snapshot is discarded instead of trapping the user`() {
        val (rides, secure) = store()
        secure.putString(SecureStore.KEY_RIDE_SESSION, "{not json at all")
        assertEquals(RideSession.Idle, rides.load())
        // 坏快照要被清掉，否则每次冷启动都重复解析失败。
        assertNull(secure.getString(SecureStore.KEY_RIDE_SESSION))
    }

    @Test
    fun `an unknown field from a newer build is tolerated`() {
        val (rides, secure) = store()
        secure.putString(
            SecureStore.KEY_RIDE_SESSION,
            """{"phase":"Riding","carId":"C1","orderId":"O1","futureField":42}""",
        )
        val restored = rides.load()
        assertEquals(RidePhase.Riding, restored.phase)
        assertEquals("C1", restored.carId)
    }

    @Test
    fun `user pin round trips and blanks clear the slot`() {
        val (rides, secure) = store()
        rides.userPin = "PIN-1"
        assertEquals("PIN-1", rides.userPin)

        rides.userPin = ""
        assertEquals("", rides.userPin)
        assertNull(secure.getString(SecureStore.KEY_USER_PIN))
    }

    @Test
    fun `service area id round trips`() {
        val (rides, _) = store()
        assertEquals("", rides.serviceAreaId)
        rides.serviceAreaId = "SA-9"
        assertEquals("SA-9", rides.serviceAreaId)
    }

    @Test
    fun `auto return is a one-shot flag`() {
        val (rides, _) = store()
        assertFalse(rides.consumeAutoReturn())

        rides.markAutoReturn()
        assertTrue(rides.consumeAutoReturn())
        // 读一次就该没了，否则申诉后会反复自动还车。
        assertFalse(rides.consumeAutoReturn())
    }

    @Test
    fun `the ride snapshot and the pin live in separate slots`() {
        val (rides, _) = store()
        rides.userPin = "PIN-1"
        rides.save(RideSession(phase = RidePhase.Riding, carId = "C1"))
        rides.clear()
        // 清骑行快照不该把登录态里的 pin 一起带走。
        assertEquals("PIN-1", rides.userPin)
    }
}
