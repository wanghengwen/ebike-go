package com.luopingtech.ebike.ops.data.admin

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.admin.BlacklistItem
import com.luopingtech.ebike.ops.domain.admin.CareerAuditItem
import com.luopingtech.ebike.ops.domain.admin.IdBindAuditItem
import com.luopingtech.ebike.ops.domain.admin.ObjectionOrderItem
import com.luopingtech.ebike.ops.domain.admin.OperationLogItem
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.Serializable
import kotlinx.serialization.builtins.ListSerializer
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
    val auditState: Int? = null,
    val createdAt: String? = null,
)

@Serializable
private data class ChangeBindDto(
    val changeBindId: String? = null,
    val id: String? = null,
    val authName: String? = null,
    val applyPhone: String? = null,
    val originPhone: String? = null,
    val auditState: Int? = null,
    val createdAt: String? = null,
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
    val id: String? = null,
    val orderId: String? = null,
    val userName: String? = null,
    val phone: String? = null,
    val carId: String? = null,
    val state: Int? = null,
    val userReason: String? = null,
    val createdAt: String? = null,
)

@Serializable
private data class ObjectionDetailDto(
    val userTicketDetailCO: ObjectionDto? = null,
)

@Serializable
private data class OperationLogDto(
    val time: String? = null,
    val name: String? = null,
    val content: String? = null,
    val carId: String? = null,
)

class AdminApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    suspend fun careerPage(serviceId: String): OpsResult<List<CareerAuditItem>> {
        val body = pagedBody("/business/user/career/page", serviceId)
        return when (
            val result = signedApi.post(
                path = "business/user/career/page",
                bodyJson = body,
                deserializer = PageListDto.serializer(CareerDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.list.mapNotNull { it.toCareer() })
            is OpsResult.Err -> result
        }
    }

    suspend fun careerAudit(id: String, pass: Boolean): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/user/career/audit",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("careerId", id)
            put("auditState", if (pass) 1 else 2)
        }
        return signedApi.postUnit(path = "business/user/career/audit", bodyJson = body)
    }

    suspend fun objectionPage(serviceId: String): OpsResult<List<ObjectionOrderItem>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/pageUserTickets",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            serviceId.toLongOrNull()?.let { put("serviceId", it) } ?: put("serviceId", serviceId)
            put("pageNum", 1)
            put("pageSize", 50)
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

    suspend fun objectionDetail(id: String, orderId: String): OpsResult<ObjectionOrderItem> {
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
                val item = result.value.userTicketDetailCO?.toObjection()
                    ?: return OpsResult.Err(com.luopingtech.ebike.ops.core.result.OpsError.business("OBJ", "empty"))
                OpsResult.Ok(item)
            }
            is OpsResult.Err -> result
        }
    }

    suspend fun dealObjection(id: String, orderId: String): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/dealUserTicket",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("id", id)
            put("orderId", orderId)
            put("izAccepted", true)
            put("initiator", 1)
        }
        return signedApi.postUnit(path = "business/dealUserTicket", bodyJson = body)
    }

    suspend fun blacklistPage(serviceId: String): OpsResult<List<BlacklistItem>> {
        val body = pagedBody("/business/blacklist/page", serviceId)
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

    suspend fun changeBindPage(serviceId: String): OpsResult<List<IdBindAuditItem>> {
        val body = pagedBody("/business/user/changeBind/page", serviceId)
        return when (
            val result = signedApi.post(
                path = "business/user/changeBind/page",
                bodyJson = body,
                deserializer = PageListDto.serializer(ChangeBindDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.list.mapNotNull { it.toChangeBind() })
            is OpsResult.Err -> result
        }
    }

    suspend fun changeBindAudit(id: String, pass: Boolean): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/user/changeBind/audit",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("changeBindId", id)
            put("auditState", if (pass) 1 else 2)
        }
        return signedApi.postUnit(path = "business/user/changeBind/audit", bodyJson = body)
    }

    suspend fun operationLogList(serviceId: String): OpsResult<List<OperationLogItem>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/log/list",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            serviceId.toLongOrNull()?.let { put("serviceId", it) } ?: put("serviceId", serviceId)
            put("pageNum", 1)
            put("pageSize", 50)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-management/log/list",
                bodyJson = body,
                deserializer = ListSerializer(OperationLogDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(
                result.value.map {
                    OperationLogItem(
                        time = it.time.orEmpty(),
                        operatorName = it.name.orEmpty(),
                        content = it.content.orEmpty(),
                        carId = it.carId.orEmpty(),
                    )
                },
            )
            is OpsResult.Err -> result
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
}

private fun CareerDto.toCareer(): CareerAuditItem? {
    val resolvedId = careerId ?: id ?: return null
    return CareerAuditItem(
        id = resolvedId,
        name = name.orEmpty(),
        phone = phone.orEmpty(),
        auditState = auditState ?: 0,
        createdAt = createdAt.orEmpty(),
    )
}

private fun ChangeBindDto.toChangeBind(): IdBindAuditItem? {
    val resolvedId = changeBindId ?: id ?: return null
    return IdBindAuditItem(
        id = resolvedId,
        authName = authName.orEmpty(),
        applyPhone = applyPhone.orEmpty(),
        originPhone = originPhone.orEmpty(),
        auditState = auditState ?: 0,
        createdAt = createdAt.orEmpty(),
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
    val resolvedId = id ?: return null
    return ObjectionOrderItem(
        id = resolvedId,
        orderId = orderId.orEmpty(),
        userName = userName.orEmpty(),
        phone = phone.orEmpty(),
        carId = carId.orEmpty(),
        state = state ?: 0,
        userReason = userReason.orEmpty(),
        createdAt = createdAt.orEmpty(),
    )
}
