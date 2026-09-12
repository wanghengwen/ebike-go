package com.luopingtech.ebike.ops.data.workorder

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.model.WorkOrder
import com.luopingtech.ebike.ops.domain.model.WorkOrderKind

interface InspectionOrderRepository {
    suspend fun list(
        area: ServiceArea,
        alarmType: Int? = null,
        carId: String? = null,
    ): OpsResult<List<WorkOrder>>

    suspend fun accept(orderId: String): OpsResult<Unit>
    suspend fun finish(orderId: String): OpsResult<Unit>
}

class InspectionOrderRepositoryImpl(
    private val demoMode: Boolean,
    private val api: InspectionOrderApi? = null,
) : InspectionOrderRepository {
    private val demoOrders = linkedMapOf<String, WorkOrder>()

    override suspend fun list(
        area: ServiceArea,
        alarmType: Int?,
        carId: String?,
    ): OpsResult<List<WorkOrder>> {
        if (demoMode || api == null) {
            ensureDemo(area)
            val filtered = demoOrders.values
                .filter { it.serviceId == area.id && it.state in 0..1 }
                .filter { alarmType == null || it.alarmType == alarmType }
                .filter { carId.isNullOrBlank() || it.carId.equals(carId, ignoreCase = true) }
                .toList()
            return OpsResult.Ok(filtered)
        }
        return api.listPage(serviceId = area.id, alarmType = alarmType, carId = carId)
    }

    override suspend fun accept(orderId: String): OpsResult<Unit> {
        if (demoMode || api == null) {
            val order = demoOrders[orderId]
                ?: return OpsResult.Err(OpsError.business("WO", Strings.t(Str.WorkOrderNotFound)))
            if (order.state != 0) {
                return OpsResult.Err(OpsError.business("WO", Strings.t(Str.WorkOrderAlreadyHandled)))
            }
            demoOrders[orderId] = order.copy(
                state = 1,
                opManName = "Demo",
                opManPhone = "13900001111",
            )
            return OpsResult.Ok(Unit)
        }
        return api.accept(orderId)
    }

    override suspend fun finish(orderId: String): OpsResult<Unit> {
        if (demoMode || api == null) {
            val order = demoOrders[orderId]
                ?: return OpsResult.Err(OpsError.business("WO", Strings.t(Str.WorkOrderNotFound)))
            if (order.state != 1) {
                return OpsResult.Err(OpsError.business("WO", Strings.t(Str.WorkOrderAlreadyHandled)))
            }
            demoOrders[orderId] = order.copy(state = 2, fixedTime = "2026-09-11 12:00:00")
            return OpsResult.Ok(Unit)
        }
        return api.finish(orderId)
    }

    private fun ensureDemo(area: ServiceArea) {
        if (demoOrders.values.any { it.serviceId == area.id }) return
        demoCatalog(area).forEach { demoOrders[it.id] = it }
    }

    companion object {
        fun demoCatalog(area: ServiceArea): List<WorkOrder> = listOf(
            WorkOrder(
                id = "1208${area.id}01",
                kind = WorkOrderKind.Inspection,
                carId = "D${area.id}-W01",
                imei = "860000000000101",
                state = 0,
                alarmType = 3,
                createdAt = "2026-09-11 09:00:00",
                serviceId = area.id,
            ),
            WorkOrder(
                id = "1208${area.id}02",
                kind = WorkOrderKind.Inspection,
                carId = "D${area.id}-W02",
                imei = "860000000000102",
                state = 1,
                alarmType = 7,
                opManName = "Demo",
                opManPhone = "13900001111",
                createdAt = "2026-09-11 09:30:00",
                serviceId = area.id,
            ),
        )
    }
}

interface RepairOrderRepository {
    suspend fun list(
        area: ServiceArea,
        fixNames: List<String> = emptyList(),
        carId: String? = null,
    ): OpsResult<List<WorkOrder>>

    suspend fun accept(orderId: String): OpsResult<Unit>
    suspend fun finish(orderId: String, handlerType: Int = 2): OpsResult<Unit>
    fun filterNames(): List<String>
}

class RepairOrderRepositoryImpl(
    private val demoMode: Boolean,
    private val api: RepairOrderApi? = null,
) : RepairOrderRepository {
    private val demoOrders = linkedMapOf<String, WorkOrder>()

    override fun filterNames(): List<String> = listOf(
        Strings.t(Str.RepairTypeBrake),
        Strings.t(Str.RepairTypeBattery),
        Strings.t(Str.RepairTypeHelmet),
    )

    override suspend fun list(
        area: ServiceArea,
        fixNames: List<String>,
        carId: String?,
    ): OpsResult<List<WorkOrder>> {
        if (demoMode || api == null) {
            ensureDemo(area)
            val filtered = demoOrders.values
                .filter { it.serviceId == area.id && it.state in 0..1 }
                .filter { carId.isNullOrBlank() || it.carId.equals(carId, ignoreCase = true) }
                .filter { order ->
                    if (fixNames.isEmpty()) true
                    else {
                        fixNames.any { wanted ->
                            order.partNames.any { it.contains(wanted, ignoreCase = true) }
                        }
                    }
                }
                .toList()
            return OpsResult.Ok(filtered)
        }
        return api.listPage(serviceId = area.id, fixNames = fixNames, carId = carId)
    }

    override suspend fun accept(orderId: String): OpsResult<Unit> {
        if (demoMode || api == null) {
            val order = demoOrders[orderId]
                ?: return OpsResult.Err(OpsError.business("WO", Strings.t(Str.WorkOrderNotFound)))
            if (order.state != 0) {
                return OpsResult.Err(OpsError.business("WO", Strings.t(Str.WorkOrderAlreadyHandled)))
            }
            demoOrders[orderId] = order.copy(
                state = 1,
                opManName = "Demo",
                opManPhone = "13900001111",
            )
            return OpsResult.Ok(Unit)
        }
        return api.accept(orderId)
    }

    override suspend fun finish(orderId: String, handlerType: Int): OpsResult<Unit> {
        if (demoMode || api == null) {
            val order = demoOrders[orderId]
                ?: return OpsResult.Err(OpsError.business("WO", Strings.t(Str.WorkOrderNotFound)))
            if (order.state != 1) {
                return OpsResult.Err(OpsError.business("WO", Strings.t(Str.WorkOrderAlreadyHandled)))
            }
            val nextState = if (handlerType == 3) 3 else 2
            demoOrders[orderId] = order.copy(
                state = nextState,
                fixedTime = "2026-09-11 12:00:00",
                carStopped = handlerType == 3 || order.carStopped,
            )
            return OpsResult.Ok(Unit)
        }
        return api.finish(orderId, handlerType)
    }

    private fun ensureDemo(area: ServiceArea) {
        if (demoOrders.values.any { it.serviceId == area.id }) return
        demoCatalog(area).forEach { demoOrders[it.id] = it }
    }

    companion object {
        fun demoCatalog(area: ServiceArea): List<WorkOrder> = listOf(
            WorkOrder(
                id = "1209${area.id}01",
                kind = WorkOrderKind.Repair,
                carId = "D${area.id}-W11",
                imei = "860000000000111",
                state = 0,
                partNames = listOf(Strings.t(Str.RepairTypeBrake)),
                fixReason = "刹把松动",
                createdAt = "2026-09-11 10:00:00",
                serviceId = area.id,
            ),
            WorkOrder(
                id = "1209${area.id}02",
                kind = WorkOrderKind.Repair,
                carId = "D${area.id}-W12",
                imei = "860000000000112",
                state = 1,
                partNames = listOf(Strings.t(Str.RepairTypeBattery)),
                fixReason = "无法还车",
                opManName = "Demo",
                opManPhone = "13900001111",
                carStopped = true,
                createdAt = "2026-09-11 10:20:00",
                serviceId = area.id,
            ),
        )
    }
}
