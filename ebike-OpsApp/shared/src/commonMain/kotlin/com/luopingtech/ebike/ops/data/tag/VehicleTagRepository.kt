package com.luopingtech.ebike.ops.data.tag

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.tag.VehicleTagRecord
import com.luopingtech.ebike.ops.domain.tag.VehicleTagType

interface VehicleTagRepository {
    suspend fun listTypes(serviceId: String): OpsResult<List<VehicleTagType>>
    suspend fun listRecords(serviceId: String, carId: String = ""): OpsResult<List<VehicleTagRecord>>
    suspend fun addRecords(serviceId: String, carIds: List<String>, typeId: String): OpsResult<Unit>
    suspend fun deleteRecord(serviceId: String, carId: String, recordId: String): OpsResult<Unit>
}

class VehicleTagRepositoryImpl(
    private val demoMode: Boolean,
    private val api: VehicleTagApi? = null,
) : VehicleTagRepository {
    private val demoTypes = listOf(
        VehicleTagType(id = "1", name = Strings.t(Str.DemoTagTypeA)),
        VehicleTagType(id = "2", name = Strings.t(Str.DemoTagTypeB)),
    )
    private val demoRecords = mutableListOf(
        VehicleTagRecord(id = "101", carId = "D1001", typeName = Strings.t(Str.DemoTagTypeA)),
    )

    override suspend fun listTypes(serviceId: String): OpsResult<List<VehicleTagType>> {
        if (demoMode || api == null) return OpsResult.Ok(demoTypes)
        return api.listTypes(serviceId)
    }

    override suspend fun listRecords(serviceId: String, carId: String): OpsResult<List<VehicleTagRecord>> {
        if (demoMode || api == null) {
            val filtered = if (carId.isBlank()) demoRecords else demoRecords.filter { it.carId == carId }
            return OpsResult.Ok(filtered.toList())
        }
        return api.listRecords(serviceId, carId)
    }

    override suspend fun addRecords(serviceId: String, carIds: List<String>, typeId: String): OpsResult<Unit> {
        if (demoMode || api == null) {
            val typeName = demoTypes.firstOrNull { it.id == typeId }?.name.orEmpty()
            carIds.forEach { carId ->
                demoRecords.add(
                    0,
                    VehicleTagRecord(
                        id = "demo-${demoRecords.size + 1}",
                        carId = carId,
                        typeName = typeName,
                    ),
                )
            }
            return OpsResult.Ok(Unit)
        }
        return api.addRecords(serviceId, carIds, typeId)
    }

    override suspend fun deleteRecord(serviceId: String, carId: String, recordId: String): OpsResult<Unit> {
        if (demoMode || api == null) {
            demoRecords.removeAll { it.id == recordId }
            return OpsResult.Ok(Unit)
        }
        return api.deleteRecord(serviceId, carId, recordId)
    }
}
