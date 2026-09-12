package com.luopingtech.ebike.ops.feature.workorder

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.workorder.InspectionOrderRepository
import com.luopingtech.ebike.ops.data.workorder.RepairOrderRepository
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.model.WorkOrder
import com.luopingtech.ebike.ops.domain.model.WorkOrderKind
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class WorkOrderUiState(
    val kind: WorkOrderKind,
    val loading: Boolean = false,
    val actingId: String? = null,
    val canTake: Boolean = false,
    val orders: List<WorkOrder> = emptyList(),
    /** Inspection: AlarmState code; null = all. */
    val selectedAlarmType: Int? = null,
    /** Repair: part name; null/blank = all. */
    val selectedFixName: String? = null,
    val fixNameChips: List<String> = emptyList(),
    val carIdQuery: String = "",
    val serviceAreaId: String? = null,
    val message: String? = null,
    val errorMessage: String? = null,
)

/**
 * Legacy work-order ledger (1208 inspection / 1209 repair): list + accept + finish (no photos).
 */
class WorkOrderFeature(
    private val kind: WorkOrderKind,
    private val inspectionRepository: InspectionOrderRepository? = null,
    private val repairRepository: RepairOrderRepository? = null,
    private val canTakeProvider: () -> Boolean = { true },
) {
    private val _state = MutableStateFlow(
        WorkOrderUiState(kind = kind, canTake = canTakeProvider()),
    )
    val state: StateFlow<WorkOrderUiState> = _state.asStateFlow()

    fun setAlarmType(type: Int?) {
        if (kind != WorkOrderKind.Inspection) return
        _state.value = _state.value.copy(selectedAlarmType = type, errorMessage = null)
    }

    fun setFixName(name: String?) {
        if (kind != WorkOrderKind.Repair) return
        _state.value = _state.value.copy(
            selectedFixName = name?.takeIf { it.isNotBlank() },
            errorMessage = null,
        )
    }

    fun setCarIdQuery(value: String) {
        _state.value = _state.value.copy(carIdQuery = value, errorMessage = null)
    }

    fun clear() {
        _state.value = WorkOrderUiState(kind = kind, canTake = canTakeProvider())
    }

    suspend fun load(area: ServiceArea?) {
        if (area == null) {
            _state.value = WorkOrderUiState(
                kind = kind,
                canTake = canTakeProvider(),
                errorMessage = Strings.t(Str.SelectServiceAreaFirst),
            )
            return
        }
        val canTake = canTakeProvider()
        val fixChips = if (kind == WorkOrderKind.Repair) {
            repairRepository?.filterNames().orEmpty()
        } else {
            emptyList()
        }
        _state.value = _state.value.copy(
            loading = true,
            errorMessage = null,
            message = null,
            serviceAreaId = area.id,
            canTake = canTake,
            fixNameChips = fixChips,
        )
        val carId = _state.value.carIdQuery.trim().takeIf { it.isNotEmpty() }
        val result = when (kind) {
            WorkOrderKind.Inspection -> {
                val repo = inspectionRepository
                if (repo == null) {
                    failMissingRepo()
                    return
                }
                repo.list(
                    area = area,
                    alarmType = _state.value.selectedAlarmType,
                    carId = carId,
                )
            }
            WorkOrderKind.Repair -> {
                val repo = repairRepository
                if (repo == null) {
                    failMissingRepo()
                    return
                }
                val name = _state.value.selectedFixName
                repo.list(
                    area = area,
                    fixNames = if (name.isNullOrBlank()) emptyList() else listOf(name),
                    carId = carId,
                )
            }
        }
        when (result) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    orders = result.value,
                    message = Strings.t(Str.WorkOrderCount, result.value.size),
                )
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    errorMessage = result.error.message,
                )
            }
        }
    }

    suspend fun accept(orderId: String): OpsResult<Unit> {
        if (!canTakeProvider()) {
            return denyTake()
        }
        _state.value = _state.value.copy(actingId = orderId, loading = true, errorMessage = null, message = null)
        val result = when (kind) {
            WorkOrderKind.Inspection -> inspectionRepository!!.accept(orderId)
            WorkOrderKind.Repair -> repairRepository!!.accept(orderId)
        }
        return afterMutate(result, Strings.t(Str.WorkOrderAcceptOk))
    }

    suspend fun finish(orderId: String): OpsResult<Unit> {
        if (!canTakeProvider()) {
            return denyTake()
        }
        _state.value = _state.value.copy(actingId = orderId, loading = true, errorMessage = null, message = null)
        val result = when (kind) {
            WorkOrderKind.Inspection -> inspectionRepository!!.finish(orderId)
            WorkOrderKind.Repair -> repairRepository!!.finish(orderId, handlerType = 2)
        }
        return afterMutate(result, Strings.t(Str.WorkOrderFinishOk))
    }

    private suspend fun afterMutate(result: OpsResult<Unit>, okMessage: String): OpsResult<Unit> {
        when (result) {
            is OpsResult.Ok -> {
                val areaId = _state.value.serviceAreaId
                if (areaId != null) {
                    load(ServiceArea(id = areaId, name = areaId))
                }
                _state.value = _state.value.copy(
                    actingId = null,
                    message = okMessage,
                )
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    actingId = null,
                    errorMessage = result.error.message,
                )
            }
        }
        return result
    }

    private fun denyTake(): OpsResult<Unit> {
        val err = OpsResult.Err(OpsError.business("PERM", Strings.t(Str.WorkOrderNoTakePerm)))
        _state.value = _state.value.copy(errorMessage = err.error.message)
        return err
    }

    private fun failMissingRepo() {
        val err = OpsResult.Err(OpsError.business("WO", Strings.t(Str.WorkOrderNotFound)))
        _state.value = _state.value.copy(loading = false, errorMessage = err.error.message)
    }
}
