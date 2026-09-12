package com.luopingtech.ebike.ops.domain.analysis

import com.luopingtech.ebike.ops.core.i18n.OpsI18n
import com.luopingtech.ebike.ops.core.i18n.OpsLanguage
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.domain.model.Vehicle
import com.luopingtech.ebike.ops.domain.vehicle.VehicleOperationStates
import kotlin.test.BeforeTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNull
import kotlin.test.assertTrue

class VehicleConditionBucketsTest {
    @BeforeTest
    fun installStrings() {
        Strings.install(OpsI18n.fallback(OpsLanguage.ZH_CN))
    }

    @Test
    fun batteryUsesHalfOpenIntervalsAndDropsSoldOut() {
        val vehicles = listOf(
            vehicle("a", battery = 91),
            vehicle("b", battery = 90),
            vehicle("c", battery = 10),
            vehicle("d", battery = 0),
            vehicle("off", battery = 99, soldOut = true),
        )
        val result = VehicleConditionBuckets.battery(vehicles)
        assertEquals(4, result.total)
        assertEquals(1, result.buckets[0].count) // >90 → a
        assertEquals(1, result.buckets[1].count) // 90 → 80–90
        assertEquals(2, result.buckets[9].count) // ≤10 → c,d
        assertEquals(25, result.buckets[0].percent) // 1*100/4
    }

    @Test
    fun idleRequiresLockAfterUnlockAndStrictlyOverOneHour() {
        val now = 1_000_000_000_000L
        val hourMs = 3_600_000L
        val vehicles = listOf(
            vehicle("riding", lock = now - hourMs, unlock = now - hourMs / 2),
            vehicle("short", lock = now - hourMs, unlock = now - 2 * hourMs), // idle == 1h → 不进档
            vehicle("1to3", lock = now - 2 * hourMs, unlock = now - 3 * hourMs),
            vehicle("exact3", lock = now - 3 * hourMs, unlock = now - 4 * hourMs), // ==3h → 1–3
            vehicle("over3d", lock = now - 80 * hourMs, unlock = now - 90 * hourMs),
            vehicle("sold", lock = now - 10 * hourMs, unlock = now - 20 * hourMs, soldOut = true),
        )
        val result = VehicleConditionBuckets.idle(vehicles, now)
        assertEquals(5, result.total) // 排除 sold
        assertEquals(2, result.buckets[0].count) // 1to3 + exact3
        assertEquals(1, result.buckets[6].count)
        assertNull(VehicleConditionBuckets.idleIndex(now - hourMs, now - 2 * hourMs, now))
        assertEquals(0, VehicleConditionBuckets.idleIndex(now - 3 * hourMs, now - 4 * hourMs, now))
        assertTrue(result.buckets.sumOf { it.count } < result.total)
    }

    private fun vehicle(
        id: String,
        battery: Int = 50,
        lock: Long = 0L,
        unlock: Long = 0L,
        soldOut: Boolean = false,
    ): Vehicle = Vehicle(
        carId = id,
        restBattery = battery,
        operationStates = if (soldOut) listOf(VehicleOperationStates.OFF) else emptyList(),
        lockTimeMs = lock,
        unlockTimeMs = unlock,
    )
}
