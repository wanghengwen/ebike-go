package com.luopingtech.ebike.ops.domain.vehicle

import com.luopingtech.ebike.ops.domain.model.Vehicle
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class VehicleAlarmFilterLogicTest {
    @Test
    fun dropsSoldOutAndMatchesAnySelectedAlarm() {
        val soldOut = Vehicle(
            carId = "off",
            operationStates = listOf(VehicleOperationStates.OFF),
            alarmStates = listOf(VehicleAlarmStates.MOVE),
            isOnline = true,
        )
        val move = Vehicle(
            carId = "m",
            alarmStates = listOf(VehicleAlarmStates.MOVE),
            isOnline = true,
        )
        val offline = Vehicle(
            carId = "o",
            isOnline = false,
        )
        assertTrue(VehicleAlarmFilterLogic.isSoldOut(soldOut.operationStates))
        assertTrue(
            VehicleAlarmFilterLogic.matchesAlarms(
                move.alarmStates,
                move.isOnline,
                setOf(VehicleAlarmStates.MOVE, VehicleAlarmStates.LOST),
            ),
        )
        assertFalse(
            VehicleAlarmFilterLogic.matchesAlarms(
                move.alarmStates,
                move.isOnline,
                setOf(VehicleAlarmStates.LOST),
            ),
        )
        assertFalse(
            VehicleAlarmFilterLogic.matchesAlarms(
                offline.alarmStates,
                offline.isOnline,
                setOf(VehicleAlarmStates.OFFLINE),
            ),
        )
        assertTrue(
            VehicleAlarmFilterLogic.matchesAlarms(
                listOf(VehicleAlarmStates.OFFLINE),
                isOnline = true,
                selected = setOf(VehicleAlarmStates.OFFLINE),
            ),
        )
        assertTrue(VehicleAlarmFilterLogic.matchesAlarms(move.alarmStates, true, emptySet()))
    }

    @Test
    fun counts_ignoreSoldOut() {
        val list = listOf(
            Vehicle(
                carId = "1",
                alarmStates = listOf(VehicleAlarmStates.MOVE),
                isOnline = true,
            ),
            Vehicle(
                carId = "2",
                operationStates = listOf(VehicleOperationStates.OFF),
                alarmStates = listOf(VehicleAlarmStates.MOVE),
            ),
        )
        val counts = VehicleAlarmFilterLogic.counts(list)
        assertEquals(1, counts[VehicleAlarmFilter.Move])
    }
}
