package com.luopingtech.ebike.ops.data.admin

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
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
import com.luopingtech.ebike.ops.domain.admin.OperationLogEventNode
import com.luopingtech.ebike.ops.domain.admin.OperationLogItem
import com.luopingtech.ebike.ops.domain.admin.OperationLogListQuery
import com.luopingtech.ebike.ops.domain.admin.ObjectionSendMode
import com.luopingtech.ebike.ops.domain.order.OrderRecord
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.builtins.ListSerializer
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.buildJsonArray
import kotlinx.serialization.json.contentOrNull
import kotlinx.serialization.json.doubleOrNull
import kotlinx.serialization.json.longOrNull
import kotlinx.serialization.json.put


@Serializable
private data class PageListDto<T>(
    val list: List<T> = emptyList(),
)

@Serializable
private data class CareerDto(
    val careerId: String? = null,
    val id: String? = null,
    val name: String? = null,
    val phone: String? = null,
    val company: String? = null,
    val certNo: String? = null,
    val frontCard: String? = null,
    val backCard: String? = null,
    val auditState: Int? = null,
    val createdAt: String? = null,
    val auditName: String? = null,
    val auditPhone: String? = null,
    val auditAt: String? = null,
    val reason: String? = null,
)

@Serializable
private data class ChangeBindDto(
    val changeBindId: String? = null,
    val id: String? = null,
    val authName: String? = null,
    val applyPhone: String? = null,
    val originPhone: String? = null,
    val authNo: String? = null,
    val applyType: Int? = null,
    val frontCard: String? = null,
    val backCard: String? = null,
    val auditState: Int? = null,
    val createdAt: String? = null,
    val dealerName: String? = null,
    val dealerPhone: String? = null,
    val dealTime: String? = null,
    val reason: String? = null,
)

@Serializable
private data class BlacklistDto(
    val blacklistId: String? = null,
    val authName: String? = null,
    val phone: String? = null,
    val reason: String? = null,
    val state: Int? = null,
    val createdAt: String? = null,
)

@Serializable
private data class ObjectionDto(
    val id: JsonElement? = null,
    val orderId: JsonElement? = null,
    val userName: String? = null,
    val phone: String? = null,
    val carId: String? = null,
    val state: Int? = null,
    val userReason: String? = null,
    val createdAt: String? = null,
    val izPaid: Int? = null,
    val initiator: Int? = null,
    val opType: Int? = null,
    val payCost: Long? = null,
    val dispatchCost: Long? = null,
    val helmetPenalty: Long? = null,
    val refundCost: Long? = null,
    val refundDispatchCost: Long? = null,
    val refundHelmetPenalty: Long? = null,
    val refundCardTimes: Int? = null,
    val updatePayCost: Long? = null,
    val updateDispatchCost: Long? = null,
    val updateHelmetPenalty: Long? = null,
    val opReason: String? = null,
    val photoUrl: String? = null,
    val opManName: String? = null,
    val opManPhone: String? = null,
    val dealAt: String? = null,
    val updatedAt: String? = null,
    val favorableTypes: List<Int>? = null,
    val ridingTime: JsonElement? = null,
    val mile: Long? = null,
    val originCost: Long? = null,
)

@Serializable
private data class ObjectionOrderDto(
    val id: JsonElement? = null,
    val carId: String? = null,
    val phone: String? = null,
    val userPhone: String? = null,
    val userName: String? = null,
    val originCost: Long? = null,
    val payCost: Long? = null,
    val hasPaid: Long? = null,
    val mile: Long? = null,
    val ridingTime: JsonElement? = null,
    val izPaid: Int? = null,
    val startLat: Double? = null,
    val startLng: Double? = null,
    val endLat: Double? = null,
    val endLng: Double? = null,
    val startTime: String? = null,
    val endTime: String? = null,
    val dispatchCost: Long? = null,
    val helmetPenalty: Long? = null,
)

@Serializable
private data class ObjectionDetailDto(
    val userTicketDetailCO: ObjectionDto? = null,
    @SerialName("bOrderDetailCO")
    val bOrderDetailCO: ObjectionOrderDto? = null,
    @SerialName("borderDetailCO")
    val borderDetailCO: ObjectionOrderDto? = null,
)

@Serializable
private data class OperationLogDto(
    val time: String? = null,
    val name: String? = null,
    val phone: String? = null,
    val carId: String? = null,
    val imei: String? = null,
    val eventName: String? = null,
    val content: String? = null,
    /** 后端为 Object，可能是字符串或结构化 JSON */
    val result: JsonElement? = null,
)

@Serializable
private data class OperationLogEventNodeDto(
    val id: JsonElement? = null,
    val eventName: String? = null,
    val children: List<OperationLogEventNodeDto>? = null,
)

@Serializable
private data class SendModeDto(
    val izApp: Boolean? = null,
    val izSms: Boolean? = null,
    val izSys: Boolean? = null,
)

class AdminApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    suspend fun careerPage(query: CertificationPageQuery): OpsResult<List<CareerAuditItem>> {
        val body = certificationPageBody("/business/user/career/page", query, career = true)
        return when (
            val result = signedApi.post(
                path = "business/user/career/page",
                bodyJson = body,
                deserializer = PageListDto.serializer(CareerDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.list.mapNotNull { it.toCareerItem() })
            is OpsResult.Err -> result
        }
    }

    suspend fun careerDetail(id: String): OpsResult<CareerAuditDetail> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/user/career/detail",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("id", id)
        }
        return when (
            val result = signedApi.post(
                path = "business/user/career/detail",
                bodyJson = body,
                deserializer = CareerDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> {
                val detail = result.value.toCareerDetail()
                    ?: return OpsResult.Err(OpsError.business("CAREER", "empty detail"))
                OpsResult.Ok(detail)
            }
            is OpsResult.Err -> result
        }
    }

    suspend fun careerAudit(req: CertificationAuditSubmit): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/user/career/audit",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("careerId", req.id)
            put("auditState", if (req.pass) 1 else 2)
            if (!req.pass && req.reason.isNotBlank()) put("reason", req.reason)
            if (req.remindTypes.isNotEmpty()) {
                put("remindTypes", buildJsonArray { req.remindTypes.forEach { add(JsonPrimitive(it)) } })
            }
        }
        return signedApi.postUnit(path = "business/user/career/audit", bodyJson = body)
    }

    /** type=5 职业认证通知模板 */
    suspend fun careerSendMode(): OpsResult<ObjectionSendMode> = getSendMode(type = 5)

    suspend fun objectionPage(query: ObjectionPageQuery): OpsResult<List<ObjectionOrderItem>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/pageUserTickets",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            query.serviceId.toLongOrNull()?.let { put("serviceId", it) }
                ?: put("serviceId", query.serviceId)
            put("pageNum", query.pageNum)
            put("pageSize", query.pageSize)
            if (query.createdTimeStart.isNotBlank()) put("createdTimeStart", query.createdTimeStart)
            if (query.createdTimeEnd.isNotBlank()) put("createdTimeEnd", query.createdTimeEnd)
            query.state?.let { put("state", it) }
            val kw = query.keyword.trim()
            if (kw.isNotEmpty()) {
                if (kw.all { it.isDigit() } || kw.contains("+86")) {
                    val phone = if (kw.contains("+86")) kw else "+86-$kw"
                    put("phone", phone)
                } else {
                    put("name", kw)
                }
            }
        }
        return when (
            val result = signedApi.post(
                path = "business/pageUserTickets",
                bodyJson = body,
                deserializer = PageListDto.serializer(ObjectionDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.list.mapNotNull { it.toObjection() })
            is OpsResult.Err -> result
        }
    }

    suspend fun objectionDetail(id: String, orderId: String): OpsResult<ObjectionOrderDetail> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/userTicketDetail",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("id", id)
            put("orderId", orderId)
        }
        return when (
            val result = signedApi.post(
                path = "business/userTicketDetail",
                bodyJson = body,
                deserializer = ObjectionDetailDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> {
                val ticket = result.value.userTicketDetailCO?.toObjection()
                    ?: return OpsResult.Err(com.luopingtech.ebike.ops.core.result.OpsError.business("OBJ", "empty"))
                val orderDto = result.value.bOrderDetailCO ?: result.value.borderDetailCO
                OpsResult.Ok(
                    ObjectionOrderDetail(
                        ticket = ticket,
                        order = orderDto?.toOrderRecord(),
                    ),
                )
            }
            is OpsResult.Err -> result
        }
    }

    suspend fun dealObjection(req: ObjectionDealRequest): OpsResult<Unit> {
        val isRefund = !req.feeReasonable && req.isPaidRefund
        val body = CommonRequestBody.toJsonString(
            source = "/business/dealUserTicket",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("id", req.id)
            put("orderId", req.orderId)
            // 对齐 UserRepository.dealObjectionOrder：发 !feeReasonable
            put("izAccepted", !req.feeReasonable)
            put("initiator", req.initiator)
            put("opReason", req.opReason)
            if (req.remindWay.isNotEmpty()) {
                put("remindWay", buildJsonArray { req.remindWay.forEach { add(JsonPrimitive(it)) } })
            }
            if (isRefund) {
                req.refundCostFen?.let { put("refundCost", it) }
                req.refundDispatchCostFen?.let { put("refundDispatchCost", it) }
                req.refundHelmetPenaltyFen?.let { put("refundHelmetPenalty", it) }
            } else if (!req.feeReasonable) {
                req.modifyPayCostFen?.let { put("modifyPayCost", it) }
                req.modifyDispatchCostFen?.let { put("modifyDispatchCost", it) }
                req.modifyHelmetPenaltyFen?.let { put("modifyHelmetPenalty", it) }
            }
            // 卡券次数不受 isRefund 限制（驳回填了也会发）
            req.refundCardTimes?.let { put("refundCardTimes", it) }
        }
        return signedApi.postUnit(path = "business/dealUserTicket", bodyJson = body)
    }

    /** type=12 对齐 MsgModeTypeEnum.MODE_OBJECTION_ORDER */
    suspend fun objectionSendMode(): OpsResult<ObjectionSendMode> = getSendMode(type = 12)

    private suspend fun getSendMode(type: Int): OpsResult<ObjectionSendMode> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/msgTemplate/getSendMode",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("type", type)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-management/msgTemplate/getSendMode",
                bodyJson = body,
                deserializer = SendModeDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(
                ObjectionSendMode(
                    izSys = result.value.izSys == true,
                    izSms = result.value.izSms == true,
                    izApp = result.value.izApp == true,
                ),
            )
            is OpsResult.Err -> result
        }
    }

    suspend fun blacklistPage(
        serviceId: String,
        phone: String? = null,
        authName: String? = null,
        authNo: String? = null,
    ): OpsResult<List<BlacklistItem>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/blacklist/page",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            // 对齐原版 BlackListRepository：serviceId 为 Long[]（listOf(SERVICE_ID)）
            val sid = serviceId.toLongOrNull()
            put(
                "serviceId",
                buildJsonArray {
                    if (sid != null) add(JsonPrimitive(sid)) else add(JsonPrimitive(serviceId))
                },
            )
            put("pageNum", 1)
            put("pageSize", 50)
            phone?.takeIf { it.isNotBlank() }?.let {
                val normalized = if (it.contains("+86")) it else "+86-$it"
                put("phone", normalized)
            }
            authName?.takeIf { it.isNotBlank() }?.let { put("authName", it) }
            authNo?.takeIf { it.isNotBlank() }?.let { put("authNo", it) }
        }
        return when (
            val result = signedApi.post(
                path = "business/blacklist/page",
                bodyJson = body,
                deserializer = PageListDto.serializer(BlacklistDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.list.mapNotNull { it.toBlacklist() })
            is OpsResult.Err -> result
        }
    }

    suspend fun blacklistCancel(id: String): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/blacklist/cancel",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("blacklistId", id)
        }
        return signedApi.postUnit(path = "business/blacklist/cancel", bodyJson = body)
    }

    suspend fun changeBindPage(query: CertificationPageQuery): OpsResult<List<IdBindAuditItem>> {
        val body = certificationPageBody("/business/user/changeBind/page", query, career = false)
        return when (
            val result = signedApi.post(
                path = "business/user/changeBind/page",
                bodyJson = body,
                deserializer = PageListDto.serializer(ChangeBindDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.list.mapNotNull { it.toChangeBindItem() })
            is OpsResult.Err -> result
        }
    }

    suspend fun changeBindDetail(id: String): OpsResult<IdBindAuditDetail> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/user/changeBind/detail",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("id", id)
        }
        return when (
            val result = signedApi.post(
                path = "business/user/changeBind/detail",
                bodyJson = body,
                deserializer = ChangeBindDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> {
                val detail = result.value.toChangeBindDetail()
                    ?: return OpsResult.Err(OpsError.business("CHANGE_BIND", "empty detail"))
                OpsResult.Ok(detail)
            }
            is OpsResult.Err -> result
        }
    }

    suspend fun changeBindAudit(req: CertificationAuditSubmit): OpsResult<Unit> {
        // Legacy ChangeBindCertificationViewModel: remindTypes 0/1/2 → 1/2/3
        val remind = req.remindTypes.map { it + 1 }
        val body = CommonRequestBody.toJsonString(
            source = "/business/user/changeBind/audit",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("changeBindId", req.id)
            put("auditState", if (req.pass) 1 else 2)
            if (!req.pass && req.reason.isNotBlank()) put("reason", req.reason)
            if (remind.isNotEmpty()) {
                put("remindTypes", buildJsonArray { remind.forEach { add(JsonPrimitive(it)) } })
            }
        }
        return signedApi.postUnit(path = "business/user/changeBind/audit", bodyJson = body)
    }

    /** type=7 实名换绑通知模板 */
    suspend fun changeBindSendMode(): OpsResult<ObjectionSendMode> = getSendMode(type = 7)

    private fun certificationPageBody(
        source: String,
        query: CertificationPageQuery,
        career: Boolean,
    ): String =
        CommonRequestBody.toJsonString(
            source = source,
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            query.serviceId.toLongOrNull()?.let { put("serviceId", it) }
                ?: put("serviceId", query.serviceId)
            put("pageNum", query.pageNum)
            put("pageSize", query.pageSize)
            if (query.startTime.isNotBlank()) put("startTime", query.startTime)
            if (query.endTime.isNotBlank()) put("endTime", query.endTime)
            query.auditState?.let { put("auditState", it) }
            val kw = query.keyword.trim()
            if (kw.isNotEmpty()) {
                val isPhone = kw.all { it.isDigit() } || kw.contains("+86")
                if (isPhone) {
                    val phone = if (kw.contains("+86")) kw else "+86-$kw"
                    if (career) {
                        put("phone", phone)
                        put("applyPhone", phone)
                    } else {
                        put("applyPhone", phone)
                    }
                } else if (career) {
                    put("name", kw)
                } else {
                    put("authName", kw)
                }
            }
        }

    private fun pagedBody(source: String, serviceId: String): String =
        CommonRequestBody.toJsonString(
            source = source,
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            serviceId.toLongOrNull()?.let { put("serviceId", it) } ?: put("serviceId", serviceId)
            put("pageNum", 1)
            put("pageSize", 50)
        }

    suspend fun operationLogEventTree(type: Int): OpsResult<List<OperationLogEventNode>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/log/eventTree",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("type", type)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-management/log/eventTree",
                bodyJson = body,
                deserializer = ListSerializer(OperationLogEventNodeDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.map { it.toNode() })
            is OpsResult.Err -> result
        }
    }

    suspend fun operationLogList(query: OperationLogListQuery): OpsResult<List<OperationLogItem>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/log/list",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            // 对齐原版：不传 serviceId；pageSize=10
            put("pageNum", query.pageNum)
            put("pageSize", query.pageSize)
            query.carId?.let { put("carId", it) }
            query.imei?.let { put("imei", it) }
            query.phone?.takeIf { it.isNotBlank() }?.let { phone ->
                val normalized = if (phone.contains("+86")) phone else "+86-$phone"
                put("phone", normalized)
            }
            query.startTimeSec?.let { put("startTime", it) }
            // 后端字段拼写就是 entTime
            query.endTimeSec?.let { put("entTime", it) }
            query.eventTypeIds?.takeIf { it.isNotBlank() }?.let { put("type", it) }
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-management/log/list",
                bodyJson = body,
                deserializer = ListSerializer(OperationLogDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.map { it.toItem() })
            is OpsResult.Err -> result
        }
    }
}

private fun CareerDto.toCareerItem(): CareerAuditItem? {
    val resolvedId = careerId ?: id ?: return null
    return CareerAuditItem(
        id = resolvedId,
        name = name.orEmpty(),
        phone = phone.orEmpty(),
        auditState = auditState ?: 0,
        createdAt = createdAt.orEmpty(),
        company = company.orEmpty(),
        certNo = certNo.orEmpty(),
        frontCard = frontCard.orEmpty(),
        backCard = backCard.orEmpty(),
    )
}

private fun CareerDto.toCareerDetail(): CareerAuditDetail? {
    val resolvedId = careerId ?: id ?: return null
    return CareerAuditDetail(
        id = resolvedId,
        name = name.orEmpty(),
        phone = phone.orEmpty(),
        company = company.orEmpty(),
        certNo = certNo.orEmpty(),
        frontCard = frontCard.orEmpty(),
        backCard = backCard.orEmpty(),
        auditState = auditState ?: 0,
        createdAt = createdAt.orEmpty(),
        auditName = auditName.orEmpty(),
        auditPhone = auditPhone.orEmpty(),
        auditAt = auditAt.orEmpty(),
        reason = reason.orEmpty(),
    )
}

private fun ChangeBindDto.toChangeBindItem(): IdBindAuditItem? {
    val resolvedId = changeBindId ?: id ?: return null
    return IdBindAuditItem(
        id = resolvedId,
        authName = authName.orEmpty(),
        applyPhone = applyPhone.orEmpty(),
        originPhone = originPhone.orEmpty(),
        auditState = auditState ?: 0,
        createdAt = createdAt.orEmpty(),
        applyType = applyType ?: 0,
        authNo = authNo.orEmpty(),
        frontCard = frontCard.orEmpty(),
        backCard = backCard.orEmpty(),
    )
}

private fun ChangeBindDto.toChangeBindDetail(): IdBindAuditDetail? {
    val resolvedId = changeBindId ?: id ?: return null
    return IdBindAuditDetail(
        id = resolvedId,
        authName = authName.orEmpty(),
        applyPhone = applyPhone.orEmpty(),
        originPhone = originPhone.orEmpty(),
        authNo = authNo.orEmpty(),
        applyType = applyType ?: 0,
        frontCard = frontCard.orEmpty(),
        backCard = backCard.orEmpty(),
        auditState = auditState ?: 0,
        createdAt = createdAt.orEmpty(),
        dealerName = dealerName.orEmpty(),
        dealerPhone = dealerPhone.orEmpty(),
        dealTime = dealTime.orEmpty(),
        reason = reason.orEmpty(),
    )
}

private fun BlacklistDto.toBlacklist(): BlacklistItem? {
    val resolvedId = blacklistId ?: return null
    return BlacklistItem(
        id = resolvedId,
        authName = authName.orEmpty(),
        phone = phone.orEmpty(),
        reason = reason.orEmpty(),
        state = state ?: 0,
        createdAt = createdAt.orEmpty(),
    )
}

private fun ObjectionDto.toObjection(): ObjectionOrderItem? {
    val resolvedId = id.toFlexibleString()
    if (resolvedId.isBlank()) return null
    return ObjectionOrderItem(
        id = resolvedId,
        orderId = orderId.toFlexibleString(),
        userName = userName.orEmpty(),
        phone = phone.orEmpty(),
        carId = carId.orEmpty(),
        state = state ?: 0,
        userReason = userReason.orEmpty(),
        createdAt = createdAt.orEmpty(),
        izPaid = izPaid,
        initiator = initiator ?: 0,
        opType = opType,
        payCost = payCost,
        dispatchCost = dispatchCost,
        helmetPenalty = helmetPenalty,
        refundCost = refundCost,
        refundDispatchCost = refundDispatchCost,
        refundHelmetPenalty = refundHelmetPenalty,
        refundCardTimes = refundCardTimes,
        updatePayCost = updatePayCost,
        updateDispatchCost = updateDispatchCost,
        updateHelmetPenalty = updateHelmetPenalty,
        opReason = opReason.orEmpty(),
        photoUrl = photoUrl.orEmpty(),
        opManName = opManName.orEmpty(),
        opManPhone = opManPhone.orEmpty(),
        dealAt = dealAt?.takeIf { it.isNotBlank() } ?: updatedAt.orEmpty(),
        favorableTypes = favorableTypes.orEmpty(),
        ridingTimeRaw = ridingTime.toFlexibleString(),
        mile = mile,
        originCost = originCost,
    )
}

private fun ObjectionOrderDto.toOrderRecord(): OrderRecord = OrderRecord(
    id = id.toFlexibleString(),
    carId = carId.orEmpty(),
    phone = phone?.takeIf { it.isNotBlank() } ?: userPhone.orEmpty(),
    userName = userName.orEmpty(),
    originCost = originCost,
    payCost = payCost,
    mile = mile,
    ridingTimeRaw = ridingTime.toFlexibleString(),
    izPaid = izPaid,
    startLat = startLat,
    startLng = startLng,
    endLat = endLat,
    endLng = endLng,
    startTime = startTime.orEmpty(),
    endTime = endTime.orEmpty(),
    hasPaid = hasPaid,
    dispatchCost = dispatchCost,
    helmetPenalty = helmetPenalty,
)

private fun JsonElement?.toFlexibleString(): String = when (this) {
    null -> ""
    is JsonPrimitive -> contentOrNull ?: longOrNull?.toString().orEmpty()
    else -> toString().trim('"')
}

private fun OperationLogDto.toItem(): OperationLogItem = OperationLogItem(
    time = time?.ifBlank { "--" } ?: "--",
    name = name?.ifBlank { "--" } ?: "--",
    phone = phone?.ifBlank { "--" } ?: "--",
    carId = carId?.ifBlank { "--" } ?: "--",
    imei = imei?.ifBlank { "--" } ?: "--",
    eventName = eventName?.ifBlank { "--" } ?: "--",
    result = result.toFlexibleString(),
    content = content.orEmpty(),
)

private fun OperationLogEventNodeDto.toNode(): OperationLogEventNode {
    // 对齐原版 EventTreeModel：id 可空也不丢节点，否则子级会被 mapNotNull 整支裁掉
    return OperationLogEventNode(
        id = id.toFlexibleLong() ?: 0L,
        eventName = eventName.orEmpty(),
        children = children.orEmpty().map { it.toNode() },
    )
}

private fun JsonElement?.toFlexibleLong(): Long? = when (this) {
    null -> null
    is JsonPrimitive -> longOrNull
        ?: contentOrNull?.toLongOrNull()
        ?: doubleOrNull?.toLong()
        ?: contentOrNull?.toDoubleOrNull()?.toLong()
    else -> toString().trim('"').toLongOrNull()
}
