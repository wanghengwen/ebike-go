package com.luopingtech.ebike.ops.data.tools

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.tools.OpsSettingConfig
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.put

@Serializable
data class OpsSettingDto(
    val swapBatteryThreshold: Int = 20,
    val izAutoSwapBattery: Boolean = false,
) {
    fun toDomain(): OpsSettingConfig = OpsSettingConfig(
        swapBatteryThreshold = swapBatteryThreshold,
        izAutoSwapBattery = izAutoSwapBattery,
    )
}

class OpsSettingApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    suspend fun load(serviceId: String): OpsResult<OpsSettingConfig> {
        val body = bodyJson(
            source = "/business/systemConfig/getConfigBaseItem",
            serviceId = serviceId,
        )
        return when (
            val result = signedApi.post(
                path = "business/systemConfig/getConfigBaseItem",
                bodyJson = body,
                deserializer = OpsSettingDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> OpsResult.Ok(result.value.toDomain())
            is OpsResult.Err -> result
        }
    }

    suspend fun save(
        serviceId: String,
        threshold: Int,
        izAutoSwapBattery: Boolean,
    ): OpsResult<Unit> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/systemConfig/updateSwapBatteryThreshold",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            serviceId.toLongOrNull()?.let { put("serviceId", it) } ?: put("serviceId", serviceId)
            put("swapBatteryThreshold", threshold)
            put("izAutoSwapBattery", izAutoSwapBattery)
        }
        return signedApi.postUnit(
            path = "business/systemConfig/updateSwapBatteryThreshold",
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
            serviceId.toLongOrNull()?.let { put("serviceId", it) } ?: put("serviceId", serviceId)
        }
}
