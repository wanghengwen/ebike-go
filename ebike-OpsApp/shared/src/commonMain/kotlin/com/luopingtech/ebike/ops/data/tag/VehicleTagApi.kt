package com.luopingtech.ebike.ops.data.tag

import com.luopingtech.ebike.ops.core.i18n.LocaleContext
import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.core.util.Ids
import com.luopingtech.ebike.ops.data.analysis.FlexibleStringSerializer
import com.luopingtech.ebike.ops.domain.tag.VehicleTagRecord
import com.luopingtech.ebike.ops.domain.tag.VehicleTagType
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.Serializable
import kotlinx.serialization.builtins.ListSerializer
import kotlinx.serialization.json.add
import kotlinx.serialization.json.buildJsonArray
import kotlinx.serialization.json.put

@Serializable
data class VehicleTagTypeDto(
    val id: Long? = null,
    val typeId: Long? = null,
    val name: String? = null,
) {
    fun toDomain(): VehicleTagType? {
        val resolvedId = (typeId ?: id)?.toString().orEmpty()
        val label = name?.trim().orEmpty()
        if (resolvedId.isBlank() || label.isBlank()) return null
        return VehicleTagType(id = resolvedId, name = label)
    }
}

@Serializable
data class VehicleTagRecordDto(
    val id: Long? = null,
    val carId: String? = null,
    val typeName: String? = null,
    @Serializable(with = FlexibleStringSerializer::class)
    val createdAt: String = "",
    val opName: String? = null,
)

@Serializable
data class VehicleTagPageDto(
    val list: List<VehicleTagRecordDto> = emptyList(),
)

class VehicleTagApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
    private val pinProvider: () -> String = { "" },
) {
    suspend fun listTypes(serviceId: String): OpsResult<List<VehicleTagType>> {
        // Align with ManagerApp ApiVehicleTagService.getTagTypeList
        val path = "/business/ebike-management/carTag/type/list"
        val body = bodyJson(path, serviceId)
        return when (
            val result = signedApi.post(
                path = path,
                bodyJson = body,
                deserializer = ListSerializer(VehicleTagTypeDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.mapNotNull { it.toDomain() })
            is OpsResult.Err -> result
        }
    }

    suspend fun listRecords(serviceId: String, carId: String = ""): OpsResult<List<VehicleTagRecord>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/carTag/record/page",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            serviceId.toLongOrNull()?.let { put("serviceId", it) } ?: put("serviceId", serviceId)
            put("pageNum", 1)
            put("pageSize", 50)
            if (carId.isNotBlank()) put("carId", carId)
        }
        return when (
            val result = signedApi.post(
                path = "/business/ebike-management/carTag/record/page",
                bodyJson = body,
                deserializer = VehicleTagPageDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(
                result.value.list.mapNotNull { dto ->
                    val id = dto.id?.toString().orEmpty()
                    val cid = dto.carId.orEmpty()
                    if (id.isBlank() || cid.isBlank()) return@mapNotNull null
                    VehicleTagRecord(
                        id = id,
                        carId = cid,
                        typeName = dto.typeName.orEmpty(),
                        createdAt = dto.createdAt,
                        operatorName = dto.opName.orEmpty(),
                    )
                },
            )
            is OpsResult.Err -> result
        }
    }

    /**
     * Legacy forces multipart: type=1, typeId, serviceId, carIds (comma-joined).
     * BFF (`cartag_record.go`) only reads PostForm — JSON body yields「参数不合法」.
     */
    suspend fun addRecords(
        serviceId: String,
        carIds: List<String>,
        typeId: String,
    ): OpsResult<Unit> {
        val joined = carIds.map { it.trim() }.filter { it.isNotEmpty() }.joinToString(",")
        if (joined.isBlank()) {
            return OpsResult.Err(
                com.luopingtech.ebike.ops.core.result.OpsError.business("TAG", "carIds empty"),
            )
        }
        val pin = pinProvider()
        val fields = linkedMapOf(
            "traceId" to Ids.uuidV4(),
            "platform" to deviceInfo.platform,
            "version" to deviceInfo.appVersion,
            "source" to "/business/ebike-management/carTag/record/add",
            "stressTesting" to "false",
            "tenantId" to tenantIdProvider(),
            "Accept-Language" to LocaleContext.acceptLanguage,
            "deviceId" to deviceIdProvider(),
            "type" to "1",
            "typeId" to typeId,
            "serviceId" to serviceId,
            "carIds" to joined,
        )
        if (pin.isNotBlank()) {
            fields["pin"] = pin
        }
        return signedApi.postMultipartUnit(
            path = "/business/ebike-management/carTag/record/add",
            fields = fields,
        )
    }

    suspend fun deleteRecord(serviceId: String, carId: String, recordId: String): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/carTag/record/del",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            serviceId.toLongOrNull()?.let { put("serviceId", it) } ?: put("serviceId", serviceId)
            recordId.toLongOrNull()?.let { id ->
                put(
                    "ids",
                    buildJsonArray {
                        add(id)
                    },
                )
            }
        }
        return signedApi.postUnit(
            path = "/business/ebike-management/carTag/record/del",
            bodyJson = body,
        )
    }

    private fun bodyJson(source: String, serviceId: String): String =
        CommonRequestBody.toJsonString(
            source = source,
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            // ManagerApp puts AppConfig.SERVICE_ID as String — keep the same type.
            put("serviceId", serviceId)
        }
}
