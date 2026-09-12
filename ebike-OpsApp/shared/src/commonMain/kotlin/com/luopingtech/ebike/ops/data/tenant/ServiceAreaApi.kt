package com.luopingtech.ebike.ops.data.tenant

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.builtins.ListSerializer

class ServiceAreaApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    suspend fun listByToken(): OpsResult<List<ServiceArea>> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/fence/serviceArea/getByToken",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        )
        return when (
            val result = signedApi.post(
                path = "business/fence/serviceArea/getByToken",
                bodyJson = body,
                deserializer = ListSerializer(ServiceAreaDto.serializer()),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.map { it.toDomain() })
            is OpsResult.Err -> result
        }
    }
}
