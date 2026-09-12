package com.luopingtech.ebike.ops.domain.vehicle

import com.luopingtech.ebike.ops.domain.model.Vehicle
import kotlin.test.Test
import kotlin.test.assertEquals

class VehicleListStatusTest {
    @Test
    fun operationIsPositive() {
        val status = Vehicle(
            carId = "1",
            ridingState = VehicleRidingStates.OPERATION,
            isOnline = true,
        ).listStatus()
        assertEquals(VehicleListStatusTone.Positive, status.tone)
    }

    @Test
    fun offlineIsAlert() {
        val status = Vehicle(
            carId = "1",
            ridingState = VehicleRidingStates.RIDEABLE,
            alarmStates = listOf(VehicleAlarmStates.OFFLINE),
            isOnline = false,
        ).listStatus()
        assertEquals(VehicleListStatusTone.Alert, status.tone)
    }
}
