package com.luopingtech.ebike.ops.data.admin

import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.admin.BlacklistItem
import com.luopingtech.ebike.ops.domain.admin.CareerAuditDetail
import com.luopingtech.ebike.ops.domain.admin.CareerAuditItem
import com.luopingtech.ebike.ops.domain.admin.CertificationAuditSubmit
import com.luopingtech.ebike.ops.domain.admin.CertificationPageQuery
import com.luopingtech.ebike.ops.domain.admin.IdBindAuditDetail
import com.luopingtech.ebike.ops.domain.admin.IdBindAuditItem
import com.luopingtech.ebike.ops.domain.admin.ObjectionDealRequest
import com.luopingtech.ebike.ops.domain.admin.ObjectionOrderDetail
import com.luopingtech.ebike.ops.domain.admin.ObjectionOrderItem
import com.luopingtech.ebike.ops.domain.admin.ObjectionPageQuery
import com.luopingtech.ebike.ops.domain.admin.ObjectionSendMode
import com.luopingtech.ebike.ops.domain.admin.ObjectionStates
import com.luopingtech.ebike.ops.domain.admin.OperationLogEventNode
import com.luopingtech.ebike.ops.domain.admin.OperationLogItem
import com.luopingtech.ebike.ops.domain.admin.OperationLogListQuery
import com.luopingtech.ebike.ops.domain.admin.isPaid
import com.luopingtech.ebike.ops.domain.order.OrderRecord

interface AdminRepository {
    suspend fun careerPage(query: CertificationPageQuery): OpsResult<List<CareerAuditItem>>
    suspend fun careerDetail(id: String): OpsResult<CareerAuditDetail>
    suspend fun careerAudit(req: CertificationAuditSubmit): OpsResult<Unit>
    suspend fun careerSendMode(): OpsResult<ObjectionSendMode>
    suspend fun objectionPage(query: ObjectionPageQuery): OpsResult<List<ObjectionOrderItem>>
    suspend fun objectionDetail(id: String, orderId: String): OpsResult<ObjectionOrderDetail>
    suspend fun dealObjection(req: ObjectionDealRequest): OpsResult<Unit>
    suspend fun objectionSendMode(): OpsResult<ObjectionSendMode>
    suspend fun blacklistPage(
        serviceId: String,
        phone: String? = null,
        authName: String? = null,
        authNo: String? = null,
    ): OpsResult<List<BlacklistItem>>
    suspend fun blacklistCancel(id: String): OpsResult<Unit>
    suspend fun changeBindPage(query: CertificationPageQuery): OpsResult<List<IdBindAuditItem>>
    suspend fun changeBindDetail(id: String): OpsResult<IdBindAuditDetail>
    suspend fun changeBindAudit(req: CertificationAuditSubmit): OpsResult<Unit>
    suspend fun changeBindSendMode(): OpsResult<ObjectionSendMode>
    suspend fun operationLogEventTree(type: Int): OpsResult<List<OperationLogEventNode>>
    suspend fun operationLogList(query: OperationLogListQuery): OpsResult<List<OperationLogItem>>
}

class AdminRepositoryImpl(
    private val demoMode: Boolean,
    private val api: AdminApi? = null,
) : AdminRepository {
    private val demoCareers = mutableListOf(
        CareerAuditItem(
            id = "c1",
            name = "张三",
            phone = "13800001111",
            auditState = 0,
            createdAt = "2026-09-15 10:00:00",
            company = "某某大学",
            certNo = "2024001",
            frontCard = "",
            backCard = "",
        ),
        CareerAuditItem(
            id = "c2",
            name = "李四",
            phone = "13900002222",
            auditState = 1,
            createdAt = "2026-09-14 09:00:00",
            company = "某某公司",
            certNo = "EMP002",
        ),
    )
    private val demoObjections = mutableListOf(
        ObjectionOrderItem(
            id = "o1",
            orderId = "ord-100",
            userName = "王五",
            phone = "13700003333",
            carId = "100600021",
            state = ObjectionStates.Pending,
            userReason = "调度费异议",
            createdAt = "2026-09-15 10:00:00",
            izPaid = 4,
            initiator = 0,
            payCost = 100,
            dispatchCost = 0,
            helmetPenalty = 0,
            originCost = 100,
        ),
        ObjectionOrderItem(
            id = "o2",
            orderId = "ord-101",
            userName = "赵六",
            phone = "13600004444",
            carId = "100600022",
            state = ObjectionStates.Processed,
            userReason = "里程异常",
            createdAt = "2026-09-14 16:20:00",
            izPaid = 3,
            initiator = 0,
            opType = 1,
            opReason = "费用合理",
            opManName = "运维A",
            opManPhone = "13800001111",
            dealAt = "2026-09-14 17:00:00",
            payCost = 200,
            originCost = 200,
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
            createdAt = "2026-09-15 11:00:00",
            applyType = 1,
            authNo = "110101199001011234",
        ),
    )
    private val demoLogs = listOf(
        OperationLogItem(
            time = "2026-09-12 10:00:00",
            name = "运维A",
            phone = "13800001111",
            carId = "100600021",
            imei = "860000000000001",
            eventName = "开锁",
            result = "开锁成功",
        ),
        OperationLogItem(
            time = "2026-09-12 09:30:00",
            name = "运维B",
            phone = "13900002222",
            carId = "100600022",
            imei = "860000000000002",
            eventName = "换电",
            result = "换电完成",
        ),
    )

    private val demoEventTree = listOf(
        OperationLogEventNode(
            id = 1L,
            eventName = "平台",
            children = listOf(
                OperationLogEventNode(
                    id = 2L,
                    eventName = "车辆详情",
                    children = listOf(
                        OperationLogEventNode(id = 3L, eventName = "开锁"),
                        OperationLogEventNode(id = 4L, eventName = "换电"),
                    ),
                ),
            ),
        ),
        OperationLogEventNode(
            id = 10L,
            eventName = "商家端",
            children = listOf(
                OperationLogEventNode(id = 11L, eventName = "登录"),
                OperationLogEventNode(id = 12L, eventName = "调度"),
            ),
        ),
        OperationLogEventNode(
            id = 20L,
            eventName = "用户端",
            children = listOf(
                OperationLogEventNode(id = 21L, eventName = "开锁"),
            ),
        ),
        OperationLogEventNode(
            id = 30L,
            eventName = "自动触发",
            children = listOf(
                OperationLogEventNode(id = 31L, eventName = "定时任务"),
            ),
        ),
    )

    override suspend fun careerPage(query: CertificationPageQuery): OpsResult<List<CareerAuditItem>> {
        if (demoMode || api == null) {
            var list = demoCareers.toList()
            query.auditState?.let { s -> list = list.filter { it.auditState == s } }
            val kw = query.keyword.trim()
            if (kw.isNotEmpty()) {
                list = list.filter {
                    it.name.contains(kw) || it.phone.contains(kw.removePrefix("+86-").removePrefix("+86"))
                }
            }
            val from = ((query.pageNum - 1) * query.pageSize).coerceAtLeast(0)
            return OpsResult.Ok(list.drop(from).take(query.pageSize))
        }
        return api.careerPage(query)
    }

    override suspend fun careerDetail(id: String): OpsResult<CareerAuditDetail> {
        if (demoMode || api == null) {
            val item = demoCareers.firstOrNull { it.id == id }
                ?: return OpsResult.Err(OpsError.business("CAREER", "not found"))
            return OpsResult.Ok(
                CareerAuditDetail(
                    id = item.id,
                    name = item.name,
                    phone = item.phone,
                    company = item.company,
                    certNo = item.certNo,
                    frontCard = item.frontCard,
                    backCard = item.backCard,
                    auditState = item.auditState,
                    createdAt = item.createdAt,
                    auditName = if (item.auditState == 0) "" else "演示运维",
                    auditPhone = if (item.auditState == 0) "" else "13800001111",
                    auditAt = if (item.auditState == 0) "" else "2026-09-16 12:00:00",
                    reason = if (item.auditState == 2) "资料不符" else "",
                ),
            )
        }
        return api.careerDetail(id)
    }

    override suspend fun careerAudit(req: CertificationAuditSubmit): OpsResult<Unit> {
        if (demoMode || api == null) {
            demoCareers.indices.forEach { index ->
                if (demoCareers[index].id == req.id) {
                    demoCareers[index] = demoCareers[index].copy(auditState = if (req.pass) 1 else 2)
                }
            }
            return OpsResult.Ok(Unit)
        }
        return api.careerAudit(req)
    }

    override suspend fun careerSendMode(): OpsResult<ObjectionSendMode> {
        if (demoMode || api == null) {
            return OpsResult.Ok(ObjectionSendMode(izSys = true, izSms = true, izApp = true))
        }
        return api.careerSendMode()
    }

    override suspend fun objectionPage(query: ObjectionPageQuery): OpsResult<List<ObjectionOrderItem>> {
        if (demoMode || api == null) {
            var list = demoObjections.toList()
            query.state?.let { s -> list = list.filter { it.state == s } }
            val kw = query.keyword.trim()
            if (kw.isNotEmpty()) {
                list = list.filter {
                    it.userName.contains(kw) || it.phone.contains(kw.removePrefix("+86-"))
                }
            }
            val from = ((query.pageNum - 1) * query.pageSize).coerceAtLeast(0)
            return OpsResult.Ok(list.drop(from).take(query.pageSize))
        }
        return api.objectionPage(query)
    }

    override suspend fun objectionDetail(id: String, orderId: String): OpsResult<ObjectionOrderDetail> {
        if (demoMode || api == null) {
            val found = demoObjections.firstOrNull { it.id == id }
                ?: return OpsResult.Err(OpsError.business("OBJ", "not found"))
            return OpsResult.Ok(
                ObjectionOrderDetail(
                    ticket = found,
                    order = OrderRecord(
                        id = found.orderId,
                        carId = found.carId,
                        phone = found.phone,
                        userName = found.userName,
                        originCost = found.originCost,
                        payCost = found.payCost,
                        mile = 211,
                        ridingTimeRaw = "1200000",
                        izPaid = found.izPaid,
                        startTime = "2026-09-15 09:50:00",
                        endTime = "2026-09-15 10:00:00",
                        dispatchCost = found.dispatchCost,
                        helmetPenalty = found.helmetPenalty,
                    ),
                ),
            )
        }
        return api.objectionDetail(id, orderId)
    }

    override suspend fun dealObjection(req: ObjectionDealRequest): OpsResult<Unit> {
        if (demoMode || api == null) {
            demoObjections.indices.forEach { index ->
                if (demoObjections[index].id == req.id) {
                    demoObjections[index] = demoObjections[index].copy(
                        state = ObjectionStates.Processed,
                        opType = if (req.feeReasonable) 1 else if (demoObjections[index].isPaid()) 3 else 2,
                        opReason = req.opReason,
                        dealAt = "2026-09-16 12:00:00",
                        opManName = "演示运维",
                    )
                }
            }
            return OpsResult.Ok(Unit)
        }
        return api.dealObjection(req)
    }

    override suspend fun objectionSendMode(): OpsResult<ObjectionSendMode> {
        if (demoMode || api == null) {
            return OpsResult.Ok(ObjectionSendMode(izSys = true, izSms = true, izApp = true))
        }
        return api.objectionSendMode()
    }

    override suspend fun blacklistPage(
        serviceId: String,
        phone: String?,
        authName: String?,
        authNo: String?,
    ): OpsResult<List<BlacklistItem>> {
        if (demoMode || api == null) {
            var list = demoBlacklist.toList()
            phone?.takeIf { it.isNotBlank() }?.let { p ->
                val bare = p.removePrefix("+86-").removePrefix("+86")
                list = list.filter { it.phone.contains(bare) }
            }
            authName?.takeIf { it.isNotBlank() }?.let { n ->
                list = list.filter { it.authName.contains(n) }
            }
            authNo?.takeIf { it.isNotBlank() }?.let {
                // demo 无身份证字段，按空结果处理
                list = emptyList()
            }
            return OpsResult.Ok(list)
        }
        return api.blacklistPage(serviceId, phone, authName, authNo)
    }

    override suspend fun blacklistCancel(id: String): OpsResult<Unit> {
        if (demoMode || api == null) {
            demoBlacklist.removeAll { it.id == id }
            return OpsResult.Ok(Unit)
        }
        return api.blacklistCancel(id)
    }

    override suspend fun changeBindPage(query: CertificationPageQuery): OpsResult<List<IdBindAuditItem>> {
        if (demoMode || api == null) {
            var list = demoChangeBind.toList()
            query.auditState?.let { s -> list = list.filter { it.auditState == s } }
            val kw = query.keyword.trim()
            if (kw.isNotEmpty()) {
                list = list.filter {
                    it.authName.contains(kw) ||
                        it.applyPhone.contains(kw.removePrefix("+86-").removePrefix("+86"))
                }
            }
            val from = ((query.pageNum - 1) * query.pageSize).coerceAtLeast(0)
            return OpsResult.Ok(list.drop(from).take(query.pageSize))
        }
        return api.changeBindPage(query)
    }

    override suspend fun changeBindDetail(id: String): OpsResult<IdBindAuditDetail> {
        if (demoMode || api == null) {
            val item = demoChangeBind.firstOrNull { it.id == id }
                ?: return OpsResult.Err(OpsError.business("CHANGE_BIND", "not found"))
            return OpsResult.Ok(
                IdBindAuditDetail(
                    id = item.id,
                    authName = item.authName,
                    applyPhone = item.applyPhone,
                    originPhone = item.originPhone,
                    authNo = item.authNo,
                    applyType = item.applyType,
                    frontCard = item.frontCard,
                    backCard = item.backCard,
                    auditState = item.auditState,
                    createdAt = item.createdAt,
                    dealerName = if (item.auditState == 0) "" else "演示运维",
                    dealerPhone = if (item.auditState == 0) "" else "13800001111",
                    dealTime = if (item.auditState == 0) "" else "2026-09-16 12:00:00",
                    reason = if (item.auditState == 2) "信息不符" else "",
                ),
            )
        }
        return api.changeBindDetail(id)
    }

    override suspend fun changeBindAudit(req: CertificationAuditSubmit): OpsResult<Unit> {
        if (demoMode || api == null) {
            demoChangeBind.indices.forEach { index ->
                if (demoChangeBind[index].id == req.id) {
                    demoChangeBind[index] = demoChangeBind[index].copy(auditState = if (req.pass) 1 else 2)
                }
            }
            return OpsResult.Ok(Unit)
        }
        return api.changeBindAudit(req)
    }

    override suspend fun changeBindSendMode(): OpsResult<ObjectionSendMode> {
        if (demoMode || api == null) {
            return OpsResult.Ok(ObjectionSendMode(izSys = true, izSms = true, izApp = true))
        }
        return api.changeBindSendMode()
    }

    override suspend fun operationLogEventTree(type: Int): OpsResult<List<OperationLogEventNode>> {
        if (demoMode || api == null) return OpsResult.Ok(demoEventTree)
        return api.operationLogEventTree(type)
    }

    override suspend fun operationLogList(query: OperationLogListQuery): OpsResult<List<OperationLogItem>> {
        if (demoMode || api == null) {
            val from = ((query.pageNum - 1) * query.pageSize).coerceAtLeast(0)
            return OpsResult.Ok(demoLogs.drop(from).take(query.pageSize))
        }
        return api.operationLogList(query)
    }
}
