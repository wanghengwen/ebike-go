package com.luopingtech.ebike.ops.data.vehicle

import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.core.time.nowEpochMillis
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.model.Vehicle
import com.luopingtech.ebike.ops.domain.scan.ScanTarget
import com.luopingtech.ebike.ops.domain.vehicle.VehicleAlarmStates
import com.luopingtech.ebike.ops.domain.vehicle.VehicleOperationStates
import com.luopingtech.ebike.ops.domain.vehicle.VehicleRidingStates

interface VehicleRepository {
    suspend fun loadByServiceArea(area: ServiceArea): OpsResult<List<Vehicle>>
    suspend fun findByScanTarget(target: ScanTarget): OpsResult<Vehicle>
    suspend fun getDetail(carId: String): OpsResult<Vehicle>
    suspend fun checkServicePermission(carId: String, serviceId: String): OpsResult<Unit>
    suspend fun bindBatterySn(carId: String, batterySn: String): OpsResult<Unit>
}

class VehicleRepositoryImpl(
    private val demoMode: Boolean,
    private val api: VehicleApi? = null,
) : VehicleRepository {
    override suspend fun loadByServiceArea(area: ServiceArea): OpsResult<List<Vehicle>> {
        if (demoMode || api == null) {
            return OpsResult.Ok(demoVehicles(area))
        }
        return api.listByServiceIds(listOf(area.id))
    }

    override suspend fun findByScanTarget(target: ScanTarget): OpsResult<Vehicle> {
        if (demoMode || api == null) {
            return OpsResult.Ok(demoVehicleFor(target))
        }
        return when (target) {
            is ScanTarget.CarId -> api.getDetail(carId = target.value)
            is ScanTarget.Imei -> api.getDetail(imei = target.value)
        }
    }

    override suspend fun getDetail(carId: String): OpsResult<Vehicle> {
        if (demoMode || api == null) {
            return OpsResult.Ok(demoVehicleFor(ScanTarget.CarId(carId)))
        }
        return api.getDetail(carId = carId)
    }

    override suspend fun checkServicePermission(carId: String, serviceId: String): OpsResult<Unit> {
        if (carId.isBlank() || serviceId.isBlank()) {
            return OpsResult.Ok(Unit)
        }
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.checkServicePermission(carId.trim(), serviceId.trim())
    }

    override suspend fun bindBatterySn(carId: String, batterySn: String): OpsResult<Unit> {
        val id = carId.trim()
        val sn = batterySn.trim()
        if (id.isBlank() || sn.isBlank()) {
            return OpsResult.Ok(Unit)
        }
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.bindBatterySn(id, sn)
    }

    companion object {
        fun demoVehicles(area: ServiceArea): List<Vehicle> {
            val baseLat = if (area.centerLat != 0.0) area.centerLat else 28.22
            val baseLng = if (area.centerLng != 0.0) area.centerLng else 112.94
            val now = nowEpochMillis()
            val hour = 3_600_000L
            return listOf(
                Vehicle(
                    carId = "D${area.id}-001",
                    imei = "860000000000001",
                    lat = baseLat + 0.001,
                    lng = baseLng + 0.001,
                    restBattery = 86,
                    ridingState = VehicleRidingStates.RIDEABLE,
                    isOnline = true,
                    serviceId = area.id,
                    voltageMv = 54_200,
                    batterySn = "BSN-${area.id}-001",
                    batteryLock = 1,
                    helmetLock = 1,
                    helmetMac = "AA:BB:CC:01",
                    forParkName = "${area.name} · A站",
                    totalMiles = 1280.0,
                    serviceName = area.name,
                    model = "1",
                    izHaveOverload = true,
                    acc = 0,
                    defend = 1,
                    gsmSignal = -71,
                    lockTimeMs = now - 2 * hour,
                    unlockTimeMs = now - 5 * hour,
                ),
                Vehicle(
                    carId = "D${area.id}-002",
                    imei = "860000000000002",
                    lat = baseLat - 0.0012,
                    lng = baseLng + 0.0008,
                    restBattery = 42,
                    ridingState = VehicleRidingStates.RIDING,
                    alarmStates = listOf(VehicleAlarmStates.ORDER_WITHOUT_GPS),
                    isOnline = true,
                    serviceId = area.id,
                    voltageMv = 48_100,
                    batterySn = "BSN-${area.id}-002",
                    batteryLock = 0,
                    forParkName = "${area.name} · B站",
                    totalMiles = 640.5,
                    serviceName = area.name,
                    model = "1",
                    lockTimeMs = now - hour,
                    unlockTimeMs = now - hour / 2,
                ),
                Vehicle(
                    carId = "D${area.id}-003",
                    imei = "860000000000003",
                    lat = baseLat + 0.0005,
                    lng = baseLng - 0.0015,
                    restBattery = 17,
                    ridingState = VehicleRidingStates.TEMP_PARKING,
                    operationStates = listOf(VehicleOperationStates.LOW_BATTERY),
                    alarmStates = listOf(VehicleAlarmStates.OFFLINE, VehicleAlarmStates.OUT_GFENCE),
                    isOnline = false,
                    serviceId = area.id,
                    voltageMv = 42_800,
                    batterySn = "BSN-${area.id}-003",
                    batteryLock = 1,
                    helmetLock = 0,
                    forParkName = "${area.name} · 边缘站",
                    noParkName = "禁停样例",
                    isOutOfServiceArea = true,
                    totalMiles = 2100.0,
                    serviceName = area.name,
                    model = "2",
                    izHaveOverload = true,
                    lockTimeMs = now - 30 * hour,
                    unlockTimeMs = now - 40 * hour,
                ),
                Vehicle(
                    carId = "D${area.id}-004",
                    imei = "860000000000004",
                    lat = baseLat - 0.0007,
                    lng = baseLng - 0.0004,
                    restBattery = 63,
                    ridingState = VehicleRidingStates.RESERVE,
                    alarmStates = listOf(VehicleAlarmStates.MOVE),
                    isOnline = true,
                    serviceId = area.id,
                    voltageMv = 51_000,
                    batterySn = "BSN-${area.id}-004",
                    batteryLock = 1,
                    forParkName = "${area.name} · C站",
                    totalMiles = 88.0,
                    serviceName = area.name,
                    model = "1",
                    lockTimeMs = now - 8 * hour,
                    unlockTimeMs = now - 10 * hour,
                ),
                Vehicle(
                    carId = "D${area.id}-005",
                    imei = "860000000000005",
                    lat = baseLat + 0.00105,
                    lng = baseLng + 0.00102,
                    restBattery = 22,
                    ridingState = VehicleRidingStates.OPERATION,
                    operationStates = listOf(VehicleOperationStates.REPAIRING),
                    alarmStates = listOf(VehicleAlarmStates.POWER_CUT),
                    isOnline = true,
                    serviceId = area.id,
                    voltageMv = 0,
                    batterySn = "",
                    batteryLock = 0,
                    forParkName = "${area.name} · 维修点",
                    totalMiles = 450.0,
                    serviceName = area.name,
                    model = "1",
                    lockTimeMs = now - 50 * hour,
                    unlockTimeMs = now - 55 * hour,
                ),
                Vehicle(
                    carId = "D${area.id}-006",
                    imei = "860000000000006",
                    lat = baseLat + 0.00098,
                    lng = baseLng + 0.00095,
                    restBattery = 31,
                    ridingState = VehicleRidingStates.RIDEABLE,
                    operationStates = listOf(VehicleOperationStates.MOVING_CAR),
                    alarmStates = listOf(VehicleAlarmStates.NO_PARKING_ZONE),
                    isOnline = true,
                    serviceId = area.id,
                    voltageMv = 46_500,
                    batterySn = "BSN-${area.id}-006",
                    batteryLock = 0,
                    noParkName = "商场门口禁停",
                    totalMiles = 320.0,
                    serviceName = area.name,
                    model = "1",
                    lockTimeMs = now - 80 * hour,
                    unlockTimeMs = now - 90 * hour,
                ),
                Vehicle(
                    carId = "D${area.id}-007",
                    imei = "860000000000007",
                    lat = baseLat + 0.00112,
                    lng = baseLng + 0.00088,
                    restBattery = 9,
                    ridingState = VehicleRidingStates.TEMP_PARKING,
                    operationStates = listOf(VehicleOperationStates.LOW_BATTERY),
                    alarmStates = listOf(VehicleAlarmStates.HELMET_LOST),
                    isOnline = true,
                    serviceId = area.id,
                    voltageMv = 40_200,
                    batterySn = "BSN-${area.id}-007",
                    batteryLock = 1,
                    helmetLock = 0,
                    helmetMac = "",
                    forParkName = "${area.name} · D站",
                    totalMiles = 990.0,
                    serviceName = area.name,
                    model = "1",
                    lockTimeMs = now - 4 * hour,
                    unlockTimeMs = now - 6 * hour,
                ),
                // Sold-out: should never appear on home map.
                Vehicle(
                    carId = "D${area.id}-OFF",
                    imei = "860000000000099",
                    lat = baseLat,
                    lng = baseLng,
                    restBattery = 50,
                    ridingState = VehicleRidingStates.OPERATION,
                    operationStates = listOf(VehicleOperationStates.OFF),
                    isOnline = false,
                    serviceId = area.id,
                    model = "1",
                    lockTimeMs = now - 100 * hour,
                    unlockTimeMs = now - 110 * hour,
                ),
            )
        }

        fun demoVehicleFor(target: ScanTarget): Vehicle {
            val catalog = listOf("1001", "1002", "1003").flatMap { id ->
                demoVehicles(ServiceArea(id = id, name = "Demo-$id"))
            }
            return when (target) {
                is ScanTarget.CarId -> catalog.firstOrNull { it.carId.equals(target.value, ignoreCase = true) }
                    ?: Vehicle(
                        carId = target.value,
                        imei = "860000000000099",
                        restBattery = 55,
                        ridingState = VehicleRidingStates.RIDEABLE,
                        isOnline = true,
                        serviceId = "demo",
                        voltageMv = 50_000,
                        batterySn = "BSN-SCAN",
                        batteryLock = 1,
                        forParkName = "Demo site",
                        totalMiles = 10.0,
                        serviceName = "Demo",
                        model = "1",
                        izHaveOverload = true,
                    )
                is ScanTarget.Imei -> catalog.firstOrNull { it.imei == target.value }
                    ?: Vehicle(
                        carId = "SCAN-${target.value.takeLast(4)}",
                        imei = target.value,
                        restBattery = 55,
                        ridingState = VehicleRidingStates.RIDEABLE,
                        isOnline = true,
                        serviceId = "demo",
                        voltageMv = 50_000,
                        batterySn = "BSN-SCAN",
                        batteryLock = 1,
                        forParkName = "Demo site",
                        totalMiles = 10.0,
                        serviceName = "Demo",
                        model = "1",
                        izHaveOverload = true,
                    )
            }
        }
    }
}
