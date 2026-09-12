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
            put("serviceId", serviceId)
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
