package com.luopingtech.ebike.ops.data.movecar

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.FreeMoveCar
import com.luopingtech.ebike.ops.domain.model.TeamWorker

interface FreeMoveCarRepository {
    suspend fun checkPermission(carId: String, serviceId: String): OpsResult<Unit>
    suspend fun start(carIds: List<String>, serviceId: String, izPushCar: Boolean = false): OpsResult<List<FreeMoveCar>>
    suspend fun list(serviceId: String): OpsResult<List<FreeMoveCar>>
    suspend fun finish(
        carIds: List<String>,
        phone: String,
        pictures: List<String> = emptyList(),
        remark: String? = null,
        teamWorkers: List<TeamWorker> = emptyList(),
    ): OpsResult<Unit>
    suspend fun remove(carId: String, serviceId: String): OpsResult<Unit>
}

class FreeMoveCarRepositoryImpl(
    private val demoMode: Boolean,
    private val api: FreeMoveCarApi? = null,
) : FreeMoveCarRepository {
    private val demoCars = linkedMapOf<String, FreeMoveCar>()

    override suspend fun checkPermission(carId: String, serviceId: String): OpsResult<Unit> {
        if (demoMode || api == null) {
            if (carId.isBlank()) {
                return OpsResult.Err(OpsError.business("MOVE_CAR", Strings.t(Str.EnterCarId)))
            }
            return OpsResult.Ok(Unit)
        }
        return api.checkPermission(carId, serviceId)
    }

    override suspend fun start(
        carIds: List<String>,
        serviceId: String,
        izPushCar: Boolean,
    ): OpsResult<List<FreeMoveCar>> {
        if (demoMode || api == null) {
            val added = carIds.map { id ->
                FreeMoveCar(carId = id, imei = "86${id.hashCode().toUInt().toString().padStart(13, '0').take(13)}", restBattery = 55, state = 1)
                    .also { demoCars[it.carId] = it }
            }
            return OpsResult.Ok(added)
        }
        return api.start(carIds, serviceId, izPushCar)
    }

    override suspend fun list(serviceId: String): OpsResult<List<FreeMoveCar>> {
        if (demoMode || api == null) {
            if (demoCars.isEmpty()) {
                // Seed one demo vehicle so UI is not empty after login.
                val seed = FreeMoveCar(carId = "D${serviceId}-002", imei = "860000000000002", restBattery = 48, state = 1)
                demoCars[seed.carId] = seed
            }
            return OpsResult.Ok(demoCars.values.toList())
        }
        return api.list(serviceId)
    }

    override suspend fun finish(
        carIds: List<String>,
        phone: String,
        pictures: List<String>,
        remark: String?,
        teamWorkers: List<TeamWorker>,
    ): OpsResult<Unit> {
        if (demoMode || api == null) {
            if (pictures.isEmpty()) {
                return OpsResult.Err(OpsError.business("23326", Strings.t(Str.NeedPhotoAudit)))
            }
            // Keep teamWorkers/phone for API parity; demo only requires photos.
            carIds.forEach { demoCars.remove(it) }
            return OpsResult.Ok(Unit)
        }
        return api.finish(carIds, phone, pictures, remark, teamWorkers)
    }

    override suspend fun remove(carId: String, serviceId: String): OpsResult<Unit> {
        if (demoMode || api == null) {
            demoCars.remove(carId)
            return OpsResult.Ok(Unit)
        }
        return api.remove(carId, serviceId)
    }
}
