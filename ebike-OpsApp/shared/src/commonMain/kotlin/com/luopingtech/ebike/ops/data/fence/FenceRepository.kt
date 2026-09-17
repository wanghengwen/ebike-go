package com.luopingtech.ebike.ops.data.fence

import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.analysis.StationTag
import com.luopingtech.ebike.ops.domain.model.FenceBundle
import com.luopingtech.ebike.ops.domain.model.FenceKind
import com.luopingtech.ebike.ops.domain.model.FenceNoParkingCreate
import com.luopingtech.ebike.ops.domain.model.FenceNoParkingUpdate
import com.luopingtech.ebike.ops.domain.model.FenceParkingCreate
import com.luopingtech.ebike.ops.domain.model.FenceParkingUpdate
import com.luopingtech.ebike.ops.domain.model.FencePolygon
import com.luopingtech.ebike.ops.domain.model.GeoLatLng

interface FenceRepository {
    suspend fun loadByServiceId(serviceId: String): OpsResult<FenceBundle>

    suspend fun loadNearLocations(
        serviceId: String,
        locations: List<GeoLatLng>,
    ): OpsResult<FenceBundle>

    suspend fun createParking(req: FenceParkingCreate): OpsResult<Unit>

    suspend fun createNoParking(req: FenceNoParkingCreate): OpsResult<Unit>

    suspend fun updateParking(req: FenceParkingUpdate): OpsResult<Unit>

    suspend fun updateNoParking(req: FenceNoParkingUpdate): OpsResult<Unit>

    suspend fun deleteParking(id: String): OpsResult<Unit>

    suspend fun deleteNoParking(id: String): OpsResult<Unit>

    suspend fun enableParkingBatch(ids: List<String>): OpsResult<Unit>

    suspend fun disableParkingBatch(ids: List<String>): OpsResult<Unit>

    suspend fun deleteParkingBatch(ids: List<String>): OpsResult<Unit>

    suspend fun deleteNoParkingBatch(ids: List<String>): OpsResult<Unit>
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

    override suspend fun createParking(req: FenceParkingCreate): OpsResult<Unit> {
        if (req.points.size < 3) {
            return OpsResult.Err(OpsError.business("FENCE", "need >= 3 points"))
        }
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.createParking(req)
    }

    override suspend fun createNoParking(req: FenceNoParkingCreate): OpsResult<Unit> {
        if (req.points.size < 3) {
            return OpsResult.Err(OpsError.business("FENCE", "need >= 3 points"))
        }
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.createNoParking(req)
    }

    override suspend fun updateParking(req: FenceParkingUpdate): OpsResult<Unit> {
        if (req.id.isBlank()) return OpsResult.Err(OpsError.business("FENCE", "id empty"))
        if (req.points.size < 3) {
            return OpsResult.Err(OpsError.business("FENCE", "need >= 3 points"))
        }
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.updateParking(req)
    }

    override suspend fun updateNoParking(req: FenceNoParkingUpdate): OpsResult<Unit> {
        if (req.id.isBlank()) return OpsResult.Err(OpsError.business("FENCE", "id empty"))
        if (req.points.size < 3) {
            return OpsResult.Err(OpsError.business("FENCE", "need >= 3 points"))
        }
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.updateNoParking(req)
    }

    override suspend fun deleteParking(id: String): OpsResult<Unit> {
        if (id.isBlank()) return OpsResult.Err(OpsError.business("FENCE", "id empty"))
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.deleteParking(id.trim())
    }

    override suspend fun deleteNoParking(id: String): OpsResult<Unit> {
        if (id.isBlank()) return OpsResult.Err(OpsError.business("FENCE", "id empty"))
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.deleteNoParking(id.trim())
    }

    override suspend fun enableParkingBatch(ids: List<String>): OpsResult<Unit> {
        if (ids.isEmpty()) return OpsResult.Err(OpsError.business("FENCE", "ids empty"))
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.enableParkingBatch(ids)
    }

    override suspend fun disableParkingBatch(ids: List<String>): OpsResult<Unit> {
        if (ids.isEmpty()) return OpsResult.Err(OpsError.business("FENCE", "ids empty"))
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.disableParkingBatch(ids)
    }

    override suspend fun deleteParkingBatch(ids: List<String>): OpsResult<Unit> {
        if (ids.isEmpty()) return OpsResult.Err(OpsError.business("FENCE", "ids empty"))
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.deleteParkingBatch(ids)
    }

    override suspend fun deleteNoParkingBatch(ids: List<String>): OpsResult<Unit> {
        if (ids.isEmpty()) return OpsResult.Err(OpsError.business("FENCE", "ids empty"))
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.deleteNoParkingBatch(ids)
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
                        centerLat = cLat,
                        centerLng = cLng,
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
                        centerLat = cLat,
                        centerLng = cLng,
                        address = "Demo street 1",
                        tags = listOf(StationTag("1", "学校")),
                    ),
                ),
                noParkings = listOf(
                    FencePolygon(
                        id = "nopark-$serviceId",
                        name = "Demo no-parking",
                        points = parkRing.map { GeoLatLng(it.lat + 0.002, it.lng + 0.002) },
                        kind = FenceKind.NoParking,
                        izEnable = true,
                        centerLat = cLat + 0.002,
                        centerLng = cLng + 0.002,
                        area = 120.5,
                        address = "Demo street 2",
                    ),
                ),
            )
        }
    }
}
