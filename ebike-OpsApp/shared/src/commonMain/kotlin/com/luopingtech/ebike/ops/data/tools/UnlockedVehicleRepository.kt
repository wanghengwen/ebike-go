package com.luopingtech.ebike.ops.data.tools

import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.UnlockedVehicle

interface UnlockedVehicleRepository {
    suspend fun list(serviceId: String, query: String = ""): OpsResult<List<UnlockedVehicle>>

    /** Demo/local: drop a row after successful lock. Remote list refreshes from server. */
    suspend fun noteLocked(carId: String)
}

class UnlockedVehicleRepositoryImpl(
    private val demoMode: Boolean,
    private val api: UnlockedVehicleApi? = null,
) : UnlockedVehicleRepository {
    private val demoSeed = linkedMapOf(
        "D1001-002" to UnlockedVehicle(carId = "D1001-002", imei = "860000000000002"),
        "D1001-006" to UnlockedVehicle(carId = "D1001-006", imei = "860000000000006"),
        "D1002-001" to UnlockedVehicle(carId = "D1002-001", imei = "860000000000011"),
    )

    override suspend fun list(serviceId: String, query: String): OpsResult<List<UnlockedVehicle>> {
        if (demoMode || api == null) {
            val q = query.trim()
            val rows = demoSeed.values.filter { row ->
                (serviceId.isBlank() || row.carId.contains(serviceId)) &&
                    (
                        q.isEmpty() ||
                            row.carId.contains(q, ignoreCase = true) ||
                            row.imei.contains(q)
                        )
            }
            return OpsResult.Ok(rows.toList())
        }
        return api.list(serviceId, query)
    }

    override suspend fun noteLocked(carId: String) {
        if (demoMode || api == null) {
            demoSeed.remove(carId)
        }
    }
}
