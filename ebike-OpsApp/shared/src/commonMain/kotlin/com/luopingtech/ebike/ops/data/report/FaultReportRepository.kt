package com.luopingtech.ebike.ops.data.report

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.FaultReportRecord
import com.luopingtech.ebike.ops.domain.model.RepairType

interface FaultReportRepository {
    suspend fun loadRepairTypes(model: String = ""): OpsResult<List<RepairType>>
    suspend fun checkServicePermission(carId: String, serviceId: String): OpsResult<Unit>
    suspend fun submit(
        carId: String,
        fixReason: String,
        types: List<RepairType>,
        photoUrls: List<String>,
        izStop: Boolean?,
    ): OpsResult<FaultReportRecord>

    suspend fun myReports(serviceId: String = ""): OpsResult<List<FaultReportRecord>>
}

class FaultReportRepositoryImpl(
    private val demoMode: Boolean,
    private val api: FaultReportApi? = null,
) : FaultReportRepository {
    private val demoHistory = mutableListOf(
        FaultReportRecord(
            id = "FR-1001",
            carId = "D1001-003",
            fixReason = Strings.t(Str.DemoFaultReason),
            typeNames = listOf(Strings.t(Str.RepairTypeWheel)),
            photoUrls = listOf("demo://photo/sample"),
            izStop = false,
            createdAt = "2026-09-09 18:20",
            statusLabel = Strings.t(Str.FaultStatusSubmitted),
        ),
    )

    override suspend fun loadRepairTypes(model: String): OpsResult<List<RepairType>> {
        if (demoMode || api == null) {
            val base = mutableListOf(
                RepairType(id = "1", name = Strings.t(Str.RepairTypeBattery), type = 1),
                RepairType(id = "2", name = Strings.t(Str.RepairTypeEcu), type = 2),
                RepairType(id = "3", name = Strings.t(Str.RepairTypeWheel), type = 3),
                RepairType(id = "4", name = Strings.t(Str.RepairTypeBrake), type = 4),
                RepairType(id = "5", name = Strings.t(Str.RepairTypeOther), type = -1),
            )
            // Demo: model "2" exposes an extra helmet type so UI/tests prove model-scoped reload.
            if (model.trim() == "2") {
                base += RepairType(id = "6", name = Strings.t(Str.RepairTypeHelmet), type = 6)
            }
            return OpsResult.Ok(base)
        }
        return api.repairTypeList(model)
    }

    override suspend fun checkServicePermission(carId: String, serviceId: String): OpsResult<Unit> {
        if (carId.isBlank()) {
            return OpsResult.Err(OpsError.business("RPT_EMPTY", Strings.t(Str.CarIdRequired)))
        }
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.checkServicePermission(carId.trim(), serviceId)
    }

    override suspend fun submit(
        carId: String,
        fixReason: String,
        types: List<RepairType>,
        photoUrls: List<String>,
        izStop: Boolean?,
    ): OpsResult<FaultReportRecord> {
        if (carId.isBlank()) {
            return OpsResult.Err(OpsError.business("RPT_EMPTY", Strings.t(Str.CarIdRequired)))
        }
        if (types.isEmpty()) {
            return OpsResult.Err(OpsError.business("RPT_TYPE", Strings.t(Str.SelectRepairType)))
        }
        if (fixReason.isBlank()) {
            return OpsResult.Err(OpsError.business("RPT_REASON", Strings.t(Str.DescriptionRequired)))
        }
        if (photoUrls.isEmpty()) {
            return OpsResult.Err(OpsError.business("RPT_PHOTO", Strings.t(Str.PhotoRequired)))
        }
        if (izStop == null) {
            return OpsResult.Err(OpsError.business("RPT_STOP", Strings.t(Str.SelectIzStop)))
        }

        if (demoMode || api == null) {
            val record = FaultReportRecord(
                id = "FR-D${demoHistory.size + 1}",
                carId = carId.trim(),
                fixReason = fixReason.trim(),
                typeNames = types.map { it.name },
                photoUrls = photoUrls,
                izStop = izStop,
                createdAt = "demo-now",
                statusLabel = if (izStop) {
                    Strings.t(Str.FaultStatusSubmittedStopped)
                } else {
                    Strings.t(Str.FaultStatusSubmitted)
                },
            )
            demoHistory.add(0, record)
            return OpsResult.Ok(record)
        }

        val typeIds = types.mapNotNull { it.id.toLongOrNull() }
        if (typeIds.size != types.size) {
            return OpsResult.Err(OpsError.business("RPT_TYPE", Strings.t(Str.SelectRepairType)))
        }
        return when (
            val result = api.submitRepair(
                carId = carId.trim(),
                fixReason = fixReason.trim(),
                typeIds = typeIds,
                typeNames = types.map { it.name },
                photoUrls = photoUrls,
                izStop = izStop,
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(
                FaultReportRecord(
                    id = "remote-${carId.trim()}",
                    carId = carId.trim(),
                    fixReason = fixReason.trim(),
                    typeNames = types.map { it.name },
                    photoUrls = photoUrls,
                    izStop = izStop,
                    createdAt = "",
                    statusLabel = Strings.t(Str.FaultStatusSubmitted),
                ),
            )
            is OpsResult.Err -> result
        }
    }

    override suspend fun myReports(serviceId: String): OpsResult<List<FaultReportRecord>> {
        if (demoMode || api == null) {
            return OpsResult.Ok(demoHistory.toList())
        }
        if (serviceId.isBlank()) {
            return OpsResult.Err(OpsError.business("RPT_AREA", Strings.t(Str.SelectServiceAreaFirst)))
        }
        return api.myReports(serviceId.trim())
    }
}
