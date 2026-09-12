package com.luopingtech.ebike.ops.data.control

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.control.NetworkVehicleControl
import com.luopingtech.ebike.ops.domain.control.VehicleAction
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.json.JsonObjectBuilder
import kotlinx.serialization.json.put

/**
 * Network channel:
 * - tools unlock/lock + change_battery ring
 * - paas device switches (acc / defend / battery / helmet / rear wheel)
 */
class RemoteNetworkVehicleControl(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) : NetworkVehicleControl {
    override suspend fun checkOpenBatteryBoxPermission(vehicleId: String): OpsResult<Unit> {
        if (vehicleId.isBlank()) {
            return OpsResult.Err(OpsError.business("CONTROL_INVALID", "vehicleId empty"))
        }
        return post(
            path = "business/ebike-operation/change_battery/permission",
            source = "/business/ebike-operation/change_battery/permission",
        ) {
            put("carId", vehicleId)
        }.let { result ->
            when (result) {
                is OpsResult.Ok -> result
                is OpsResult.Err -> OpsResult.Err(
                    OpsError.business(
                        result.error.code.ifBlank { "BAT_PERM" },
                        result.error.message.ifBlank { Strings.t(Str.OpenBatBoxNoPermission) },
                    ),
                )
            }
        }
    }

    override suspend fun execute(
        vehicleId: String,
        action: VehicleAction,
        imei: String,
    ): OpsResult<Unit> {
        if (vehicleId.isBlank()) {
            return OpsResult.Err(OpsError.business("CONTROL_INVALID", "vehicleId empty"))
        }
        return when (action) {
            VehicleAction.Unlock -> post(
                path = "business/ebike-operation/tools/start",
                source = "/business/ebike-operation/tools/start",
            ) {
                put("carId", vehicleId)
                put("izRiskControl", false)
            }
            VehicleAction.Lock -> post(
                path = "business/ebike-operation/tools/end",
                source = "/business/ebike-operation/tools/end",
            ) {
                put("carId", vehicleId)
                put("izRiskControl", false)
            }
            VehicleAction.Ring -> post(
                path = "business/ebike-operation/change_battery/car_searching",
                source = "/business/ebike-operation/change_battery/car_searching",
            ) {
                put("carId", vehicleId)
            }
            VehicleAction.AccOn -> postDevice(vehicleId, imei, "business/paas/device/acc") {
                put("acc", 1)
                put("izRiskControl", false)
            }
            VehicleAction.AccOff -> postDevice(vehicleId, imei, "business/paas/device/acc") {
                put("acc", 0)
                put("izRiskControl", false)
            }
            VehicleAction.DefendOn -> postDevice(vehicleId, imei, "business/paas/device/defend") {
                put("defend", 1)
            }
            VehicleAction.DefendOff -> postDevice(vehicleId, imei, "business/paas/device/defend") {
                put("defend", 0)
            }
            // Legacy: sw 0 = open, 1 = close
            VehicleAction.OpenBatteryBox -> postDevice(vehicleId, imei, "business/paas/device/batteryCompartment") {
                put("sw", 0)
            }
            VehicleAction.CloseBatteryBox -> postDevice(vehicleId, imei, "business/paas/device/batteryCompartment") {
                put("sw", 1)
            }
            VehicleAction.OpenHelmetLock -> postDevice(vehicleId, imei, "business/paas/device/helmetLock") {
                put("sw", 0)
            }
            VehicleAction.CloseHelmetLock -> postDevice(vehicleId, imei, "business/paas/device/helmetLock") {
                put("sw", 1)
            }
            VehicleAction.OpenBackWheelLock -> postDevice(vehicleId, imei, "business/paas/device/rearWheelLock") {
                put("sw", 0)
            }
            VehicleAction.CloseBackWheelLock -> postDevice(vehicleId, imei, "business/paas/device/rearWheelLock") {
                put("sw", 1)
            }
        }
    }

    private suspend fun postDevice(
        vehicleId: String,
        imei: String,
        path: String,
        block: JsonObjectBuilder.() -> Unit,
    ): OpsResult<Unit> = post(path = path, source = "/$path") {
        put("carId", vehicleId)
        put("imei", imei)
        put("izSync", true)
        block()
    }

    private suspend fun post(
        path: String,
        source: String,
        block: JsonObjectBuilder.() -> Unit,
    ): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = source,
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
            block = block,
        )
        return signedApi.postUnit(path = path, bodyJson = body)
    }
}

class DemoNetworkVehicleControl : NetworkVehicleControl {
    override suspend fun execute(
        vehicleId: String,
        action: VehicleAction,
        imei: String,
    ): OpsResult<Unit> {
        if (vehicleId.isBlank()) {
            return OpsResult.Err(OpsError.business("CONTROL_INVALID", "vehicleId empty"))
        }
        return OpsResult.Ok(Unit)
    }
}
