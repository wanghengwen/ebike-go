package com.luopingtech.ebike.ops.data.relocation

import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.RelocationDevice
import com.luopingtech.ebike.ops.domain.scan.ScanTarget

interface RelocationRepository {
    suspend fun scanDevice(
        serviceId: String,
        target: ScanTarget,
    ): OpsResult<RelocationDevice>

    suspend fun reportLocation(
        imeiList: List<String>,
        latitude: Double,
        longitude: Double,
    ): OpsResult<Unit>
}

class RelocationRepositoryImpl(
    private val demoMode: Boolean,
    private val api: RelocationApi? = null,
) : RelocationRepository {
    override suspend fun scanDevice(
        serviceId: String,
        target: ScanTarget,
    ): OpsResult<RelocationDevice> {
        if (demoMode || api == null) {
            val carId = when (target) {
                is ScanTarget.CarId -> target.value
                is ScanTarget.Imei -> "D-${target.value.takeLast(4)}"
            }
            val imei = when (target) {
                is ScanTarget.CarId -> "86${carId.filter { it.isDigit() }.padStart(13, '0').takeLast(13)}"
                is ScanTarget.Imei -> target.value
            }
            return OpsResult.Ok(
                RelocationDevice(
                    carId = carId,
                    imei = imei,
                    lat = 28.22,
                    lng = 112.94,
                    restBattery = 40,
                    isOnline = 0,
                ),
            )
        }
        return when (target) {
            is ScanTarget.CarId -> api.deviceScan(serviceId = serviceId, carId = target.value)
            is ScanTarget.Imei -> api.deviceScan(serviceId = serviceId, imei = target.value)
        }
    }

    override suspend fun reportLocation(
        imeiList: List<String>,
        latitude: Double,
        longitude: Double,
    ): OpsResult<Unit> {
        if (imeiList.isEmpty()) return OpsResult.Ok(Unit)
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.locationReport(imeiList, latitude, longitude)
    }
}
