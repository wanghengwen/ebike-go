package com.luopingtech.ebike.ops.data.staff

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.staff.AddEmployeeRequest
import com.luopingtech.ebike.ops.domain.staff.BatteryConfig
import com.luopingtech.ebike.ops.domain.staff.EmployeeTrackResult
import com.luopingtech.ebike.ops.domain.staff.MoveCarConfig
import com.luopingtech.ebike.ops.domain.staff.StaffEmployee
import com.luopingtech.ebike.ops.domain.staff.StaffEmployeePageQuery
import com.luopingtech.ebike.ops.domain.staff.StaffRealtimeLocation
import com.luopingtech.ebike.ops.domain.staff.StaffRole
import com.luopingtech.ebike.ops.domain.staff.StaffRoleOption
import com.luopingtech.ebike.ops.domain.staff.TrackOperation
import com.luopingtech.ebike.ops.domain.staff.TrackPoint
import com.luopingtech.ebike.ops.domain.staff.ViewConfig
import com.luopingtech.ebike.ops.domain.staff.VoltagePlanOption
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.Serializable
import kotlinx.serialization.builtins.ListSerializer
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.booleanOrNull
import kotlinx.serialization.json.buildJsonArray
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.contentOrNull
import kotlinx.serialization.json.doubleOrNull
import kotlinx.serialization.json.intOrNull
import kotlinx.serialization.json.longOrNull
import kotlinx.serialization.json.put

@Serializable
private data class PageListDto<T>(
    val list: List<T> = emptyList(),
)

@Serializable
private data class StaffUserDto(
    val id: JsonElement? = null,
    val name: String? = null,
    val phone: String? = null,
    val roleName: String? = null,
    val createdAt: String? = null,
    val status: JsonElement? = null,
    val pin: String? = null,
    val userPin: String? = null,
    val profession: String? = null,
    val izRoot: JsonElement? = null,
    val agentIds: List<JsonElement>? = null,
)

@Serializable
private data class StaffRoleDto(
    val id: JsonElement? = null,
    val name: String? = null,
    val status: JsonElement? = null,
    val remark: String? = null,
    val serviceName: String? = null,
    val tabCodeList: List<String>? = null,
)

@Serializable
private data class HomePermissionDto(
    val id: JsonElement? = null,
    val roleId: JsonElement? = null,
    val tabCodeList: List<String>? = null,
)

@Serializable
private data class MoveCarConfigDto(
    val carStatus: List<Long>? = null,
    val carIdStart: JsonElement? = null,
    val carIdEnd: JsonElement? = null,
)

@Serializable
private data class BatteryConfigDto(
    val voltagePlanIds: List<Long>? = null,
    val carIdStart: JsonElement? = null,
    val carIdEnd: JsonElement? = null,
)

@Serializable
private data class ViewConfigDto(
    val moveCarConfigCo: MoveCarConfigDto? = null,
    val changeBatteryConfigCo: BatteryConfigDto? = null,
)

@Serializable
private data class VoltagePlanDto(
    val id: JsonElement? = null,
    val name: String? = null,
)

@Serializable
private data class TrackSegmentDto(
    val createdAt: String? = null,
    val pointList: List<List<Double>>? = null,
)

@Serializable
private data class TrackOperationDto(
    val id: JsonElement? = null,
    val operationType: String? = null,
    val createdAt: String? = null,
    val carId: String? = null,
    val opMan: String? = null,
    val phone: String? = null,
)

@Serializable
private data class TrackAggDto(
    val operationList: List<TrackOperationDto>? = null,
)

@Serializable
private data class TrackListDto(
    val trackList: List<TrackSegmentDto>? = null,
    val operationAggregation: List<TrackAggDto>? = null,
)

@Serializable
private data class RealtimeUserDto(
    val id: JsonElement? = null,
    val name: String? = null,
    val phone: String? = null,
    val pin: String? = null,
    val userPin: String? = null,
    val pointList: List<List<Double>>? = null,
)

/**
 * Staff CRUD / role / view-config / track APIs.
 * Paths under business/ebike-management/...
 */
class StaffManagementApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    private fun body(source: String, block: kotlinx.serialization.json.JsonObjectBuilder.() -> Unit): String =
        CommonRequestBody.toJsonString(
            source = source,
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
            block = block,
        )

    suspend fun pageEmployees(query: StaffEmployeePageQuery): OpsResult<List<StaffEmployee>> {
        val source = "/business/ebike-management/user/pageList"
        val json = body(source) {
            put("pageNum", query.pageNum)
            put("pageSize", query.pageSize)
            put("status", query.status)
            query.nameOrPhone?.takeIf { it.isNotBlank() }?.let {
                put("nameOrPhone", it)
                if (it.all { ch -> ch.isDigit() }) put("phone", it) else put("name", it)
            }
            query.roleIds?.takeIf { it.isNotEmpty() }?.let { ids ->
                put("roleIds", buildJsonArray { ids.forEach { add(JsonPrimitive(it)) } })
            }
            query.agentIds?.takeIf { it.isNotEmpty() }?.let { ids ->
                put("agentIds", buildJsonArray { ids.forEach { add(JsonPrimitive(it)) } })
            }
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-management/user/pageList",
                bodyJson = json,
                deserializer = PageListDto.serializer(StaffUserDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.list.mapNotNull { it.toDomain() })
            is OpsResult.Err -> result
        }
    }

    suspend fun addEmployee(req: AddEmployeeRequest): OpsResult<Unit> {
        val source = "/business/ebike-management/user/add"
        val json = body(source) {
            put("name", req.name)
            put("phone", req.phone)
            put("profession", req.profession)
            put("izRoot", req.izRoot)
            put("roleIds", buildJsonArray { req.roleIds.forEach { add(JsonPrimitive(it)) } })
        }
        return signedApi.postUnit("business/ebike-management/user/add", json)
    }

    suspend fun setEmployeeStatus(id: String, enabled: Boolean): OpsResult<Unit> {
        val source = "/business/ebike-management/user/status"
        val json = body(source) {
            put("id", id)
            put("status", enabled)
        }
        return signedApi.postUnit("business/ebike-management/user/status", json)
    }

    suspend fun deleteEmployee(id: String): OpsResult<Unit> {
        val source = "/business/ebike-management/user/delete"
        val json = body(source) { put("id", id) }
        return signedApi.postUnit("business/ebike-management/user/delete", json)
    }

    suspend fun roleListForFilter(subTenantId: String): OpsResult<List<StaffRoleOption>> {
        val source = "/business/ebike-management/role/list"
        val json = body(source) { put("subTenantId", subTenantId) }
        return when (
            val result = signedApi.post(
                path = "business/ebike-management/role/list",
                bodyJson = json,
                deserializer = ListSerializer(StaffRoleDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.mapNotNull { it.toOption() })
            is OpsResult.Err -> result
        }
    }

    suspend fun roleAppList(
        userId: String,
        status: Int,
        subTenantId: String,
        name: String? = null,
    ): OpsResult<List<StaffRole>> {
        val source = "/business/ebike-management/role/appList"
        val json = body(source) {
            put("id", userId)
            put("status", status)
            put("subTenantId", subTenantId)
            name?.takeIf { it.isNotBlank() }?.let { put("name", it) }
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-management/role/appList",
                bodyJson = json,
                deserializer = ListSerializer(StaffRoleDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.mapNotNull { it.toDomain() })
            is OpsResult.Err -> result
        }
    }

    suspend fun roleDisableEnabled(id: String, name: String, status: Int): OpsResult<Unit> {
        val source = "/business/ebike-management/role/disableEnabled"
        val json = body(source) {
            put("id", id)
            put("name", name)
            put("status", status)
        }
        return signedApi.postUnit("business/ebike-management/role/disableEnabled", json)
    }

    suspend fun roleDelete(id: String, name: String, status: Int): OpsResult<Unit> {
        val source = "/business/ebike-management/role/delete"
        val json = body(source) {
            put("id", id)
            put("name", name)
            put("status", status)
        }
        return signedApi.postUnit("business/ebike-management/role/delete", json)
    }

    suspend fun getHomeTabsByRoleId(roleId: String): OpsResult<List<String>> {
        val source = "/business/ebike-management/home/page/role/getByRoleId"
        val json = body(source) { put("id", roleId) }
        return when (
            val result = signedApi.post(
                path = "business/ebike-management/home/page/role/getByRoleId",
                bodyJson = json,
                deserializer = HomePermissionDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.tabCodeList.orEmpty())
            is OpsResult.Err -> result
        }
    }

    suspend fun saveHomeTabs(roleId: String, tabCodeList: List<String>): OpsResult<Unit> {
        val source = "/business/ebike-management/home/page/role/addOrUpdate"
        val json = body(source) {
            put("roleId", roleId)
            put("tabCodeList", buildJsonArray { tabCodeList.forEach { add(JsonPrimitive(it)) } })
        }
        return signedApi.postUnit("business/ebike-management/home/page/role/addOrUpdate", json)
    }

    suspend fun getViewConfig(roleId: String): OpsResult<ViewConfig> {
        val source = "/business/ebike-management/app-data-role/get"
        val json = body(source) { put("roleId", roleId) }
        return when (
            val result = signedApi.post(
                path = "business/ebike-management/app-data-role/get",
                bodyJson = json,
                deserializer = ViewConfigDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain(roleId))
            is OpsResult.Err -> result
        }
    }

    suspend fun saveMoveCarConfig(roleId: String, config: MoveCarConfig): OpsResult<Unit> {
        val source = "/business/ebike-management/app-data-role/addOrUpdate"
        val json = body(source) {
            put("roleId", roleId)
            put(
                "carConfigCmd",
                buildJsonObject {
                    put("carIdStart", config.carIdStart.ifBlank { "0" })
                    put("carIdEnd", config.carIdEnd.ifBlank { "0" })
                    put("carStatus", buildJsonArray { config.carStatus.forEach { add(JsonPrimitive(it)) } })
                },
            )
        }
        return signedApi.postUnit("business/ebike-management/app-data-role/addOrUpdate", json)
    }

    suspend fun saveBatteryConfig(roleId: String, config: BatteryConfig): OpsResult<Unit> {
        val source = "/business/ebike-management/app-data-role/addOrUpdate"
        val json = body(source) {
            put("roleId", roleId)
            put(
                "batteryConfigCmd",
                buildJsonObject {
                    put("carIdStart", config.carIdStart.ifBlank { "0" })
                    put("carIdEnd", config.carIdEnd.ifBlank { "0" })
                    put(
                        "voltagePlanIds",
                        buildJsonArray { config.voltagePlanIds.forEach { add(JsonPrimitive(it)) } },
                    )
                },
            )
        }
        return signedApi.postUnit("business/ebike-management/app-data-role/addOrUpdate", json)
    }

    suspend fun voltagePlanList(): OpsResult<List<VoltagePlanOption>> {
        val source = "/business/ebike-management/voltagePlan/list"
        val json = body(source) {}
        return when (
            val result = signedApi.post(
                path = "business/ebike-management/voltagePlan/list",
                bodyJson = json,
                deserializer = ListSerializer(VoltagePlanDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.mapNotNull { it.toDomain() })
            is OpsResult.Err -> result
        }
    }

    suspend fun getTrack(
        userPin: String,
        startTime: String,
        endTime: String,
    ): OpsResult<EmployeeTrackResult> {
        val source = "/business/ebike-management/employeeTrack/getTrack"
        val json = body(source) {
            put("userPin", userPin)
            put("startTime", startTime)
            put("endTime", endTime)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-management/employeeTrack/getTrack",
                bodyJson = json,
                deserializer = TrackListDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain())
            is OpsResult.Err -> result
        }
    }

    suspend fun realTimeLocation(userPin: String? = null): OpsResult<List<StaffRealtimeLocation>> {
        val source = "/business/ebike-management/employeeTrack/realTimeLocation"
        val json = body(source) {
            userPin?.takeIf { it.isNotBlank() }?.let { put("userPin", it) }
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-management/employeeTrack/realTimeLocation",
                bodyJson = json,
                deserializer = ListSerializer(RealtimeUserDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.map { it.toDomain() })
            is OpsResult.Err -> result
        }
    }

    suspend fun getAllOperations(
        userPin: String,
        startTime: String? = null,
        endTime: String? = null,
    ): OpsResult<List<TrackOperation>> {
        val source = "/business/ebike-management/employeeTrack/getAllOperation"
        val json = body(source) {
            put("userPin", userPin)
            if (!startTime.isNullOrBlank() && !endTime.isNullOrBlank()) {
                put("startTime", startTime)
                put("endTime", endTime)
            }
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-management/employeeTrack/getAllOperation",
                bodyJson = json,
                deserializer = ListSerializer(TrackOperationDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.map { it.toDomain() })
            is OpsResult.Err -> result
        }
    }
}

private fun StaffUserDto.toDomain(): StaffEmployee? {
    val id = id.toFlexibleString()
    if (id.isBlank()) return null
    return StaffEmployee(
        id = id,
        name = name.orEmpty(),
        phone = phone.orEmpty(),
        roleName = roleName.orEmpty(),
        createdAt = createdAt.orEmpty(),
        enabled = status.toFlexibleBoolean(default = true),
        pin = userPin?.takeIf { it.isNotBlank() } ?: pin.orEmpty(),
        profession = profession.orEmpty(),
        izRoot = izRoot.toFlexibleBoolean(default = false),
        agentIds = agentIds.orEmpty().map { it.toFlexibleString() }.filter { it.isNotBlank() },
    )
}

private fun StaffRoleDto.toDomain(): StaffRole? {
    val id = id.toFlexibleString()
    if (id.isBlank()) return null
    return StaffRole(
        id = id,
        name = name.orEmpty(),
        status = status.toFlexibleInt(default = 1),
        remark = remark.orEmpty(),
        serviceName = serviceName.orEmpty(),
        tabCodeList = tabCodeList.orEmpty(),
    )
}

private fun StaffRoleDto.toOption(): StaffRoleOption? {
    val id = id.toFlexibleString()
    if (id.isBlank()) return null
    return StaffRoleOption(id = id, name = name.orEmpty())
}

private fun ViewConfigDto.toDomain(roleId: String): ViewConfig = ViewConfig(
    roleId = roleId,
    moveCar = MoveCarConfig(
        carStatus = moveCarConfigCo?.carStatus.orEmpty(),
        carIdStart = moveCarConfigCo?.carIdStart.toFlexibleString().ifBlank { "0" },
        carIdEnd = moveCarConfigCo?.carIdEnd.toFlexibleString().ifBlank { "0" },
    ),
    battery = BatteryConfig(
        voltagePlanIds = changeBatteryConfigCo?.voltagePlanIds.orEmpty(),
        carIdStart = changeBatteryConfigCo?.carIdStart.toFlexibleString().ifBlank { "0" },
        carIdEnd = changeBatteryConfigCo?.carIdEnd.toFlexibleString().ifBlank { "0" },
    ),
)

private fun VoltagePlanDto.toDomain(): VoltagePlanOption? {
    val id = id.toFlexibleLong() ?: return null
    return VoltagePlanOption(id = id, name = name.orEmpty().ifBlank { id.toString() })
}

private fun TrackListDto.toDomain(): EmployeeTrackResult {
    val points = trackList.orEmpty().flatMap { seg ->
        val at = seg.createdAt.orEmpty()
        seg.pointList.orEmpty().mapNotNull { pair ->
            if (pair.size < 2) return@mapNotNull null
            TrackPoint(createdAt = at, lng = pair[0], lat = pair[1])
        }
    }
    val operations = operationAggregation.orEmpty()
        .flatMap { it.operationList.orEmpty() }
        .map { it.toDomain() }
    return EmployeeTrackResult(points = points, operations = operations)
}

private fun RealtimeUserDto.toDomain(): StaffRealtimeLocation {
    val pts = pointList.orEmpty().mapNotNull { pair ->
        if (pair.size < 2) return@mapNotNull null
        TrackPoint(lng = pair[0], lat = pair[1])
    }
    return StaffRealtimeLocation(
        id = id.toFlexibleString(),
        pin = userPin?.takeIf { it.isNotBlank() } ?: pin.orEmpty(),
        name = name.orEmpty(),
        phone = phone.orEmpty(),
        points = pts,
    )
}

private fun TrackOperationDto.toDomain(): TrackOperation = TrackOperation(
    id = id.toFlexibleString(),
    operationType = operationType.orEmpty(),
    createdAt = createdAt.orEmpty(),
    carId = carId.orEmpty(),
    opMan = opMan.orEmpty(),
    phone = phone.orEmpty(),
)

private fun JsonElement?.toFlexibleString(): String = when (this) {
    null -> ""
    is JsonPrimitive -> contentOrNull ?: longOrNull?.toString().orEmpty()
        ?: intOrNull?.toString().orEmpty()
        ?: booleanOrNull?.toString().orEmpty()
    else -> toString().trim('"')
}

private fun JsonElement?.toFlexibleLong(): Long? = when (this) {
    null -> null
    is JsonPrimitive -> longOrNull
        ?: contentOrNull?.toLongOrNull()
        ?: doubleOrNull?.toLong()
    else -> toString().trim('"').toLongOrNull()
}

private fun JsonElement?.toFlexibleInt(default: Int): Int = when (this) {
    null -> default
    is JsonPrimitive -> intOrNull
        ?: longOrNull?.toInt()
        ?: booleanOrNull?.let { if (it) 1 else 0 }
        ?: contentOrNull?.toIntOrNull()
        ?: default
    else -> default
}

private fun JsonElement?.toFlexibleBoolean(default: Boolean): Boolean = when (this) {
    null -> default
    is JsonPrimitive -> booleanOrNull
        ?: intOrNull?.let { it != 0 }
        ?: longOrNull?.let { it != 0L }
        ?: contentOrNull?.let {
            when (it.lowercase()) {
                "true", "1" -> true
                "false", "0" -> false
                else -> default
            }
        }
        ?: default
    else -> default
}
