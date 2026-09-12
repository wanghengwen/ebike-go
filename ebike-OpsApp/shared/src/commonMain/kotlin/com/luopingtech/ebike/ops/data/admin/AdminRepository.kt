package com.luopingtech.ebike.ops.data.admin

import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.admin.BlacklistItem
import com.luopingtech.ebike.ops.domain.admin.CareerAuditItem
import com.luopingtech.ebike.ops.domain.admin.IdBindAuditItem
import com.luopingtech.ebike.ops.domain.admin.ObjectionOrderItem
import com.luopingtech.ebike.ops.domain.admin.OperationLogItem

interface AdminRepository {
    suspend fun careerPage(serviceId: String): OpsResult<List<CareerAuditItem>>
    suspend fun careerAudit(id: String, pass: Boolean): OpsResult<Unit>
    suspend fun objectionPage(serviceId: String): OpsResult<List<ObjectionOrderItem>>
    suspend fun objectionDetail(id: String, orderId: String): OpsResult<ObjectionOrderItem>
    suspend fun dealObjection(id: String, orderId: String): OpsResult<Unit>
    suspend fun blacklistPage(serviceId: String): OpsResult<List<BlacklistItem>>
    suspend fun blacklistCancel(id: String): OpsResult<Unit>
    suspend fun changeBindPage(serviceId: String): OpsResult<List<IdBindAuditItem>>
    suspend fun changeBindAudit(id: String, pass: Boolean): OpsResult<Unit>
    suspend fun operationLogList(serviceId: String): OpsResult<List<OperationLogItem>>
}

class AdminRepositoryImpl(
    private val demoMode: Boolean,
    private val api: AdminApi? = null,
) : AdminRepository {
    private val demoCareers = mutableListOf(
        CareerAuditItem(id = "c1", name = "张三", phone = "13800001111", auditState = 0),
        CareerAuditItem(id = "c2", name = "李四", phone = "13900002222", auditState = 1),
    )
    private val demoObjections = mutableListOf(
        ObjectionOrderItem(
            id = "o1",
            orderId = "ord-100",
            userName = "王五",
            phone = "13700003333",
            carId = "D1002",
            state = 0,
            userReason = "调度费异议",
        ),
    )
    private val demoBlacklist = mutableListOf(
        BlacklistItem(id = "b1", authName = "赵六", phone = "13600004444", reason = "恶意破坏", state = 1),
    )
    private val demoChangeBind = mutableListOf(
        IdBindAuditItem(
            id = "cb1",
            authName = "钱七",
            applyPhone = "13500005555",
            originPhone = "13400006666",
            auditState = 0,
        ),
    )
    private val demoLogs = listOf(
        OperationLogItem(time = "2026-09-12 10:00", operatorName = "运维A", content = "开锁", carId = "D1001"),
        OperationLogItem(time = "2026-09-12 09:30", operatorName = "运维B", content = "换电", carId = "D1003"),
    )

    override suspend fun careerPage(serviceId: String): OpsResult<List<CareerAuditItem>> {
        if (demoMode || api == null) return OpsResult.Ok(demoCareers.toList())
        return api.careerPage(serviceId)
    }

    override suspend fun careerAudit(id: String, pass: Boolean): OpsResult<Unit> {
        if (demoMode || api == null) {
            demoCareers.indices.forEach { index ->
                if (demoCareers[index].id == id) {
                    demoCareers[index] = demoCareers[index].copy(auditState = if (pass) 1 else 2)
                }
            }
            return OpsResult.Ok(Unit)
        }
        return api.careerAudit(id, pass)
    }

    override suspend fun objectionPage(serviceId: String): OpsResult<List<ObjectionOrderItem>> {
        if (demoMode || api == null) return OpsResult.Ok(demoObjections.toList())
        return api.objectionPage(serviceId)
    }

    override suspend fun objectionDetail(id: String, orderId: String): OpsResult<ObjectionOrderItem> {
        if (demoMode || api == null) {
            val found = demoObjections.firstOrNull { it.id == id }
                ?: return OpsResult.Err(com.luopingtech.ebike.ops.core.result.OpsError.business("OBJ", "not found"))
            return OpsResult.Ok(found)
        }
        return api.objectionDetail(id, orderId)
    }

    override suspend fun dealObjection(id: String, orderId: String): OpsResult<Unit> {
        if (demoMode || api == null) {
            demoObjections.indices.forEach { index ->
                if (demoObjections[index].id == id) {
                    demoObjections[index] = demoObjections[index].copy(state = 1)
                }
            }
            return OpsResult.Ok(Unit)
        }
        return api.dealObjection(id, orderId)
    }

    override suspend fun blacklistPage(serviceId: String): OpsResult<List<BlacklistItem>> {
        if (demoMode || api == null) return OpsResult.Ok(demoBlacklist.toList())
        return api.blacklistPage(serviceId)
    }

    override suspend fun blacklistCancel(id: String): OpsResult<Unit> {
        if (demoMode || api == null) {
            demoBlacklist.removeAll { it.id == id }
            return OpsResult.Ok(Unit)
        }
        return api.blacklistCancel(id)
    }

    override suspend fun changeBindPage(serviceId: String): OpsResult<List<IdBindAuditItem>> {
        if (demoMode || api == null) return OpsResult.Ok(demoChangeBind.toList())
        return api.changeBindPage(serviceId)
    }

    override suspend fun changeBindAudit(id: String, pass: Boolean): OpsResult<Unit> {
        if (demoMode || api == null) {
            demoChangeBind.indices.forEach { index ->
                if (demoChangeBind[index].id == id) {
                    demoChangeBind[index] = demoChangeBind[index].copy(auditState = if (pass) 1 else 2)
                }
            }
            return OpsResult.Ok(Unit)
        }
        return api.changeBindAudit(id, pass)
    }

    override suspend fun operationLogList(serviceId: String): OpsResult<List<OperationLogItem>> {
        if (demoMode || api == null) return OpsResult.Ok(demoLogs)
        return api.operationLogList(serviceId)
    }
}
