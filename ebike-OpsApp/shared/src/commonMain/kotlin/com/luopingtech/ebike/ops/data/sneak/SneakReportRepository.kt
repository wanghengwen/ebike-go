package com.luopingtech.ebike.ops.data.sneak

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.SneakReportRecord
import com.luopingtech.ebike.ops.domain.model.SneakReportType

interface SneakReportRepository {
    suspend fun loadTypes(serviceId: String): OpsResult<List<SneakReportType>>
    suspend fun checkServicePermission(carId: String, serviceId: String): OpsResult<Unit>
    suspend fun submit(
        carId: String,
        description: String,
        types: List<SneakReportType>,
        otherType: String,
        photoUrls: List<String>,
        itinId: String,
        reportedUserPin: String,
        reportedUserPhone: String,
        reportManPin: String,
        serviceId: String,
    ): OpsResult<SneakReportRecord>

    suspend fun myReports(reportManPhone: String): OpsResult<List<SneakReportRecord>>
    suspend fun cancel(sneakId: String, serviceId: String): OpsResult<Unit>
}

class SneakReportRepositoryImpl(
    private val demoMode: Boolean,
    private val api: SneakReportApi? = null,
) : SneakReportRepository {
    private val demoHistory = mutableListOf(
        SneakReportRecord(
            id = "SK-1001",
            carId = "D1001-001",
            reportedUserName = "Demo Rider",
            reportedUserPhone = "13800138000",
            reportedUserPin = "demo-user-pin",
            itinId = "demo-order-D1001-001",
            typeLabels = listOf(Strings.t(Str.SneakTypePrivateLock)),
            description = Strings.t(Str.DemoSneakReason),
            createdAt = "2026-09-10 16:00",
            checkResult = 0,
            serviceId = "1001",
        ),
        SneakReportRecord(
            id = "SK-1002",
            carId = "D1001-002",
            reportedUserName = "Demo Rider 2",
            reportedUserPhone = "13900139000",
            reportedUserPin = "demo-user-2",
            itinId = "demo-order-2",
            typeLabels = listOf(Strings.t(Str.SneakTypeDamage)),
            createdAt = "2026-09-09 11:20",
            checkResult = 3,
            remark = "ok",
            serviceId = "1001",
        ),
        SneakReportRecord(
            id = "SK-1003",
            carId = "D1001-003",
            reportedUserName = "Demo Rider 3",
            reportedUserPhone = "13700137000",
            typeLabels = listOf(Strings.t(Str.SneakTypeOther)),
            otherType = "custom",
            createdAt = "2026-09-08 09:00",
            checkResult = 5,
            serviceId = "1001",
        ),
    )

    override suspend fun loadTypes(serviceId: String): OpsResult<List<SneakReportType>> {
        if (demoMode || api == null) {
            return OpsResult.Ok(
                listOf(
                    SneakReportType(id = "1", name = Strings.t(Str.SneakTypePrivateLock), type = 1),
                    SneakReportType(id = "2", name = Strings.t(Str.SneakTypeDamage), type = 2),
                    SneakReportType(id = "3", name = Strings.t(Str.SneakTypeOccupy), type = 3),
                    SneakReportType(id = "12", name = Strings.t(Str.SneakTypeOther), type = 12),
                ),
            )
        }
        return api.typeList(serviceId)
    }

    override suspend fun checkServicePermission(carId: String, serviceId: String): OpsResult<Unit> {
        if (carId.isBlank()) {
            return OpsResult.Err(OpsError.business("SNEAK_EMPTY", Strings.t(Str.CarIdRequired)))
        }
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.checkServicePermission(carId.trim(), serviceId)
    }

    override suspend fun submit(
        carId: String,
        description: String,
        types: List<SneakReportType>,
        otherType: String,
        photoUrls: List<String>,
        itinId: String,
        reportedUserPin: String,
        reportedUserPhone: String,
        reportManPin: String,
        serviceId: String,
    ): OpsResult<SneakReportRecord> {
        if (carId.isBlank()) {
            return OpsResult.Err(OpsError.business("SNEAK_EMPTY", Strings.t(Str.CarIdRequired)))
        }
        if (types.isEmpty()) {
            return OpsResult.Err(OpsError.business("SNEAK_TYPE", Strings.t(Str.SneakSelectType)))
        }
        if (itinId.isBlank() || reportedUserPin.isBlank()) {
            return OpsResult.Err(OpsError.business("SNEAK_ORDER", Strings.t(Str.SneakNeedLastOrder)))
        }
        val hasOther = types.any { it.isOther }
        if (hasOther && otherType.trim().isBlank()) {
            return OpsResult.Err(OpsError.business("SNEAK_OTHER", Strings.t(Str.SneakOtherRequired)))
        }

        if (demoMode || api == null) {
            val record = SneakReportRecord(
                id = "SK-D${demoHistory.size + 1}",
                carId = carId.trim(),
                reportedUserName = "",
                reportedUserPhone = reportedUserPhone,
                reportedUserPin = reportedUserPin,
                itinId = itinId,
                typeLabels = types.map { it.name },
                otherType = otherType.trim(),
                description = description.trim(),
                photoUrls = photoUrls,
                createdAt = "demo-now",
                checkResult = 0,
                serviceId = serviceId,
            )
            demoHistory.add(0, record)
            return OpsResult.Ok(record)
        }

        return when (
            val result = api.submit(
                carId = carId.trim(),
                description = description.trim().take(100),
                typeIds = types.map { it.id },
                otherType = otherType.trim().take(10),
                photoUrls = photoUrls,
                itinId = itinId,
                reportedUserPin = reportedUserPin,
                reportManPin = reportManPin,
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(
                SneakReportRecord(
                    id = "remote-${carId.trim()}",
                    carId = carId.trim(),
                    reportedUserPhone = reportedUserPhone,
                    reportedUserPin = reportedUserPin,
                    itinId = itinId,
                    typeLabels = types.map { it.name },
                    otherType = otherType.trim(),
                    description = description.trim(),
                    photoUrls = photoUrls,
                    checkResult = 0,
                    serviceId = serviceId,
                ),
            )
            is OpsResult.Err -> result
        }
    }

    override suspend fun myReports(reportManPhone: String): OpsResult<List<SneakReportRecord>> {
        if (demoMode || api == null) {
            return OpsResult.Ok(demoHistory.toList())
        }
        return api.myList(reportManPhone = reportManPhone)
    }

    override suspend fun cancel(sneakId: String, serviceId: String): OpsResult<Unit> {
        if (demoMode || api == null) {
            val idx = demoHistory.indexOfFirst { it.id == sneakId }
            if (idx < 0) {
                return OpsResult.Err(OpsError.business("SNEAK", Strings.t(Str.SneakNotFound)))
            }
            val cur = demoHistory[idx]
            if (!cur.canCancel) {
                return OpsResult.Err(OpsError.business("SNEAK", Strings.t(Str.SneakCannotCancel)))
            }
            demoHistory[idx] = cur.copy(checkResult = 5)
            return OpsResult.Ok(Unit)
        }
        return api.cancel(sneakId = sneakId, serviceId = serviceId)
    }
}
