package com.luopingtech.ebike.ops.domain.model

import com.luopingtech.ebike.ops.domain.vehicle.VehicleOperationStates
import com.luopingtech.ebike.ops.domain.vehicle.VehicleRidingStates
import kotlin.test.Test
import kotlin.test.assertEquals

class HomeMapPinIconTest {
    @Test
    fun ridingStateIconsMatchLegacyGetVehicleIcon() {
        assertEquals(MapPinIcon.VehicleReady, vehicle(VehicleRidingStates.RIDEABLE).homeMapPinIcon())
        assertEquals(MapPinIcon.VehicleRiding, vehicle(VehicleRidingStates.RIDING).homeMapPinIcon())
        assertEquals(MapPinIcon.VehicleTempParking, vehicle(VehicleRidingStates.TEMP_PARKING).homeMapPinIcon())
        assertEquals(MapPinIcon.VehicleBooking, vehicle(VehicleRidingStates.RESERVE).homeMapPinIcon())
        assertEquals(MapPinIcon.VehicleHome, vehicle(VehicleRidingStates.OPERATION).homeMapPinIcon())
    }

    @Test
    fun operationStateOverridesInLegacyOrder() {
        val both = vehicle(
            riding = VehicleRidingStates.RIDING,
            ops = listOf(
                VehicleOperationStates.LOW_BATTERY,
                VehicleOperationStates.MOVING_CAR,
                VehicleOperationStates.REPAIRING,
            ),
        )
        assertEquals(MapPinIcon.VehicleRepairing, both.homeMapPinIcon())
        assertEquals(
            MapPinIcon.VehicleMoving,
            vehicle(
                VehicleRidingStates.RIDEABLE,
                listOf(VehicleOperationStates.LOW_BATTERY, VehicleOperationStates.MOVING_CAR),
            ).homeMapPinIcon(),
        )
        assertEquals(
            MapPinIcon.VehicleLowBattery,
            vehicle(VehicleRidingStates.RIDEABLE, listOf(VehicleOperationStates.LOW_BATTERY)).homeMapPinIcon(),
        )
    }

    @Test
    fun drawableNamesMatchLegacyResources() {
        assertEquals("icon_vehicle_ready", MapPinIcon.VehicleReady.legacyDrawableName())
        assertEquals("icon_vehicle_repairing", MapPinIcon.VehicleRepairing.legacyDrawableName())
        assertEquals("icon_vehicle", MapPinIcon.VehicleHome.legacyDrawableName())
        assertEquals("ico_vehicle_normal", MapPinIcon.VehicleNormal.legacyDrawableName())
        assertEquals("icon_parking_normal_unselect", MapPinIcon.Parking.legacyDrawableName())
        assertEquals("icon_parking_hiden_unselect", MapPinIcon.ParkingHidden.legacyDrawableName())
        assertEquals("icon_no_parking_unselect", MapPinIcon.NoParking.legacyDrawableName())
    }

    @Test
    fun badgeDrawableMatchesLegacyPriority() {
        assertEquals(
            "tips_parking",
            vehicle(VehicleRidingStates.TEMP_PARKING).homeMapBadgeDrawableName(),
        )
        assertEquals(
            "tips_riding",
            vehicle(VehicleRidingStates.RIDING).homeMapBadgeDrawableName(),
        )
        assertEquals(
            "tips_offline",
            vehicle(VehicleRidingStates.RIDEABLE, online = false).homeMapBadgeDrawableName(),
        )
        assertEquals(
            "tips_repair",
            vehicle(
                VehicleRidingStates.RIDEABLE,
                ops = listOf(VehicleOperationStates.REPAIRING),
            ).homeMapBadgeDrawableName(),
        )
    }

    private fun vehicle(
        riding: Int,
        ops: List<Int> = emptyList(),
        online: Boolean = true,
    ) = Vehicle(
        carId = "1",
        ridingState = riding,
        operationStates = ops,
        isOnline = online,
    )
}
