package com.luopingtech.ebike.ops.data.fence

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.FenceBundle
import com.luopingtech.ebike.ops.domain.model.GeoLatLng
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.json.buildJsonArray
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.put

class FenceApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    /** Legacy: /business/fence/serviceArea/getFenceByServiceId */
    suspend fun getFenceByServiceId(serviceId: String): OpsResult<FenceBundle> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/fence/serviceArea/getFenceByServiceId",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            // 后端 FenceIdDTO / FenceTypeDTO 校验字段是 id（@NotNull Long），
            // 不是 serviceId；写错会回「id不能为空」。与遗留 App / StationAnalysisApi 对齐。
            val asLong = serviceId.toLongOrNull()
            if (asLong != null) put("id", asLong) else put("id", serviceId)
        }
        return when (
            val result = signedApi.post(
                path = "business/fence/serviceArea/getFenceByServiceId",
                bodyJson = body,
                deserializer = FenceBundleDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain())
            is OpsResult.Err -> result
        }
    }

    /** Legacy: /business/fence/serviceArea/getNearFenceByLocations */
    suspend fun getNearFenceByLocations(
        serviceId: String,
        locations: List<GeoLatLng>,
    ): OpsResult<FenceBundle> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/fence/serviceArea/getNearFenceByLocations",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("serviceId", serviceId)
            put(
                "locations",
                buildJsonArray {
                    locations.forEach { loc ->
                        add(
                            buildJsonObject {
                                put("lat", loc.lat)
                                put("lng", loc.lng)
                            },
                        )
                    }
                },
            )
        }
        return when (
            val result = signedApi.post(
                path = "business/fence/serviceArea/getNearFenceByLocations",
                bodyJson = body,
                deserializer = FenceBundleDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain())
            is OpsResult.Err -> result
        }
    }
}
