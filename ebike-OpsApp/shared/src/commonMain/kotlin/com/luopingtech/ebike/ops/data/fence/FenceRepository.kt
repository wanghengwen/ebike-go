package com.luopingtech.ebike.ops.data.fence

import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.FenceBundle
import com.luopingtech.ebike.ops.domain.model.FenceKind
import com.luopingtech.ebike.ops.domain.model.FencePolygon
import com.luopingtech.ebike.ops.domain.model.GeoLatLng

interface FenceRepository {
    suspend fun loadByServiceId(serviceId: String): OpsResult<FenceBundle>

    suspend fun loadNearLocations(
        serviceId: String,
        locations: List<GeoLatLng>,
    ): OpsResult<FenceBundle>
}

class FenceRepositoryImpl(
    private val demoMode: Boolean,
    private val api: FenceApi? = null,
) : FenceRepository {
    override suspend fun loadByServiceId(serviceId: String): OpsResult<FenceBundle> {
        if (serviceId.isBlank()) {
            return OpsResult.Err(OpsError.business("FENCE", "serviceId empty"))
        }
        if (demoMode || api == null) {
            return OpsResult.Ok(demoFence(serviceId))
        }
        return api.getFenceByServiceId(serviceId.trim())
    }

    override suspend fun loadNearLocations(
        serviceId: String,
        locations: List<GeoLatLng>,
    ): OpsResult<FenceBundle> {
        if (serviceId.isBlank()) {
            return OpsResult.Err(OpsError.business("FENCE", "serviceId empty"))
        }
        if (locations.isEmpty()) {
            return loadByServiceId(serviceId)
        }
        if (demoMode || api == null) {
            return OpsResult.Ok(demoFence(serviceId, near = true))
        }
        return api.getNearFenceByLocations(serviceId.trim(), locations)
    }

    companion object {
        fun demoFence(serviceId: String, near: Boolean = false): FenceBundle {
            val cLat = 28.22
            val cLng = 112.94
            val delta = if (near) 0.004 else 0.008
            val ring = listOf(
                GeoLatLng(cLat + delta, cLng - delta),
                GeoLatLng(cLat + delta, cLng + delta),
                GeoLatLng(cLat - delta, cLng + delta),
                GeoLatLng(cLat - delta, cLng - delta),
            )
            val parkRing = listOf(
                GeoLatLng(cLat + 0.001, cLng - 0.001),
                GeoLatLng(cLat + 0.001, cLng + 0.001),
                GeoLatLng(cLat - 0.001, cLng + 0.001),
                GeoLatLng(cLat - 0.001, cLng - 0.001),
            )
            return FenceBundle(
                serviceAreas = listOf(
                    FencePolygon(
                        id = "svc-$serviceId",
                        name = "Demo service $serviceId",
                        points = ring,
                        kind = FenceKind.ServiceArea,
                        carCount = 128,
                        izEnable = true,
                    ),
                ),
                parkings = listOf(
                    FencePolygon(
                        id = "park-$serviceId",
                        name = "Demo parking",
                        points = parkRing,
                        kind = FenceKind.Parking,
                        carCount = 6,
                        currentParkingNumber = 6,
                        maxParkingNumber = 20,
                        izEnable = true,
                    ),
                ),
                noParkings = listOf(
                    FencePolygon(
                        id = "nopark-$serviceId",
                        name = "Demo no-parking",
                        points = parkRing.map { GeoLatLng(it.lat + 0.002, it.lng + 0.002) },
                        kind = FenceKind.NoParking,
                        izEnable = true,
                    ),
                ),
            )
        }
    }
}
