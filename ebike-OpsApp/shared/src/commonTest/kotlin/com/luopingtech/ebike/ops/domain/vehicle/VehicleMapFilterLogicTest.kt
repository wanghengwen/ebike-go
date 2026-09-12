package com.luopingtech.ebike.ops.domain.vehicle

import com.luopingtech.ebike.ops.domain.model.Vehicle
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class VehicleMapFilterLogicTest {
    @Test
    fun ridingAndOperationFilters_matchLegacyCodes() {
        val ready = Vehicle(carId = "a", ridingState = VehicleRidingStates.RIDEABLE)
        val riding = Vehicle(carId = "b", ridingState = VehicleRidingStates.RIDING)
        val low = Vehicle(
            carId = "c",
            restBattery = 12,
            ridingState = VehicleRidingStates.TEMP_PARKING,
            operationStates = listOf(VehicleOperationStates.LOW_BATTERY),
        )
        val moving = Vehicle(
            carId = "d",
            ridingState = VehicleRidingStates.RIDEABLE,
            operationStates = listOf(VehicleOperationStates.MOVING_CAR),
        )

        assertTrue(
            VehicleMapFilterLogic.matches(
                ready.ridingState,
                ready.operationStates,
                ready.restBattery,
                VehicleMapFilter.Ready,
            ),
        )
        assertTrue(
            VehicleMapFilterLogic.matches(
                riding.ridingState,
                riding.operationStates,
                riding.restBattery,
                VehicleMapFilter.Riding,
            ),
        )
        assertFalse(
            VehicleMapFilterLogic.matches(
                riding.ridingState,
                riding.operationStates,
                riding.restBattery,
                VehicleMapFilter.Ready,
            ),
        )
        assertTrue(
            VehicleMapFilterLogic.matches(
                low.ridingState,
                low.operationStates,
                low.restBattery,
                VehicleMapFilter.LowBattery,
            ),
        )
        // Legacy: low battery is operationState 4 only — do not invent 1..30 threshold.
        assertFalse(
            VehicleMapFilterLogic.matches(
                ridingState = VehicleRidingStates.RIDEABLE,
                operationStates = emptyList(),
                restBattery = 12,
                filter = VehicleMapFilter.LowBattery,
            ),
        )
        assertTrue(
            VehicleMapFilterLogic.matches(
                moving.ridingState,
                moving.operationStates,
                moving.restBattery,
                VehicleMapFilter.Moving,
            ),
        )
        // Legacy warehouse stub always empty.
        assertFalse(
            VehicleMapFilterLogic.matches(
                ready.ridingState,
                ready.operationStates,
                ready.restBattery,
                VehicleMapFilter.Warehouse,
            ),
        )
    }

    @Test
    fun counts_includeAllBucket() {
        val list = listOf(
            Vehicle(carId = "1", ridingState = VehicleRidingStates.RIDEABLE),
            Vehicle(carId = "2", ridingState = VehicleRidingStates.RIDING),
        )
        val counts = VehicleMapFilterLogic.counts(list)
        assertEquals(2, counts[VehicleMapFilter.All])
        assertEquals(1, counts[VehicleMapFilter.Ready])
        assertEquals(1, counts[VehicleMapFilter.Riding])
        assertEquals(0, counts[VehicleMapFilter.Warehouse])
    }
}
