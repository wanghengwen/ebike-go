package com.luopingtech.ebike.ops.domain.permission

import kotlin.test.Test
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class OpsPermissionsTest {
    @Test
    fun demoFull_showsAllTabs() {
        val p = OpsPermissions.demoFull()
        assertTrue(p.showMap)
        assertTrue(p.showScan)
        assertTrue(p.showTaskCenter)
        assertTrue(p.showChangeBattery)
        assertTrue(p.canScanUnlock)
        assertTrue(p.showMaintainModule)
        assertTrue(p.showOperationModule)
        assertTrue(p.showWarehouse)
        assertTrue(p.showWarehouseIn)
        assertTrue(p.showProduction)
        assertTrue(p.showProductionDetect)
        assertTrue(p.showFaultReport)
        assertTrue(p.showSneakReport)
        assertTrue(p.canBindBatterySn)
        assertTrue(p.showUnlockedVehicles)
        assertTrue(p.canFilterUnlockedStaff)
        assertTrue(p.showRelocation)
        assertTrue(p.showInspectionOrder)
        assertTrue(p.canTakeInspectionOrder)
        assertTrue(p.showRepairOrder)
        assertTrue(p.canTakeRepairOrder)
    }

    @Test
    fun demoLimited_hidesTaskAndScan() {
        val p = OpsPermissions.forDemoAccount("limited")
        assertTrue(p.showMap)
        assertFalse(p.showScan)
        assertFalse(p.showTaskCenter)
        assertFalse(p.showChangeBattery)
        assertFalse(p.showWarehouse)
        assertFalse(p.showProduction)
        assertFalse(p.showFaultReport)
        assertFalse(p.showSneakReport)
        assertFalse(p.showUnlockedVehicles)
        assertFalse(p.showInspectionOrder)
        assertFalse(p.showRepairOrder)
    }

    @Test
    fun workOrder_entryVisibleWithoutTakeCode() {
        val entryOnly = OpsPermissions.fromCodes(
            listOf(OpsPermissionCodes.INSPECTION_ORDER, OpsPermissionCodes.REPAIR_ORDER),
        )
        assertTrue(entryOnly.showInspectionOrder)
        assertFalse(entryOnly.canTakeInspectionOrder)
        assertTrue(entryOnly.showRepairOrder)
        assertFalse(entryOnly.canTakeRepairOrder)
    }

    @Test
    fun remoteCodes_taskTileNeedsCenterAndChild() {
        val onlyChild = OpsPermissions.fromCodes(listOf(OpsPermissionCodes.TASK_CHANGE_BATTERY))
        assertFalse(onlyChild.showChangeBattery)

        val both = OpsPermissions.fromCodes(
            listOf(OpsPermissionCodes.TASK_CENTER, OpsPermissionCodes.TASK_CHANGE_BATTERY),
        )
        assertTrue(both.showChangeBattery)
        assertFalse(both.showMoveCar)
    }

    @Test
    fun warehouseChild_requiresOwnCode_notParentAlone() {
        val parentOnly = OpsPermissions.fromCodes(listOf(OpsPermissionCodes.WAREHOUSE_MODULE))
        assertTrue(parentOnly.showWarehouse)
        assertFalse(parentOnly.showWarehouseIn)
        assertFalse(parentOnly.showWarehouseOut)
        assertFalse(parentOnly.showWarehouseRecord)

        val withChild = OpsPermissions.fromCodes(
            listOf(OpsPermissionCodes.WAREHOUSE_MODULE, OpsPermissionCodes.WAREHOUSE_IN),
        )
        assertTrue(withChild.showWarehouseIn)
        assertFalse(withChild.showWarehouseOut)
    }

    @Test
    fun maintainSection_requiresParentModule() {
        val childOnly = OpsPermissions.fromCodes(listOf(OpsPermissionCodes.VEHICLE_LIST))
        assertFalse(childOnly.showMaintainModule)
        val withParent = OpsPermissions.fromCodes(
            listOf(OpsPermissionCodes.MAINTENANCE_MODULE, OpsPermissionCodes.VEHICLE_LIST),
        )
        assertTrue(withParent.showMaintainModule)
        assertTrue(withParent.showVehicleList)
    }
}
