package com.luopingtech.ebike.rider.data.ble

import com.luopingtech.ebike.rider.core.network.CommonRequestBody
import com.luopingtech.ebike.rider.core.network.SignedApiClient
import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.data.auth.FlexibleStringSerializer
import com.luopingtech.ebike.rider.domain.ble.BleFrame
import com.luopingtech.ebike.rider.platform.DeviceInfo
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.put

/**
 * UniApp `getBlueToothToken` → `POST /client/paas/device/getBlueToothToken`.
 */
interface BleTokenApi {
    suspend fun getToken(imei: String): RiderResult<String>
}

class DemoBleTokenApi : BleTokenApi {
    override suspend fun getToken(imei: String): RiderResult<String> =
        RiderResult.Ok(BleFrame.DEFAULT_TOKEN_DECIMAL)
}

class BleTokenRemote(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) : BleTokenApi {
    override suspend fun getToken(imei: String): RiderResult<String> {
        val body = CommonRequestBody.toJsonString(
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("imei", imei)
        }
        return when (
            val result = signedApi.post(
                path = PATH,
                bodyJson = body,
                deserializer = BleTokenDto.serializer(),
            )
        ) {
            is RiderResult.Ok -> {
                val token = result.value.token.trim()
                if (token.isEmpty()) {
                    RiderResult.Err(RiderError.business("BLE_TOKEN_EMPTY", "empty bluetooth token"))
                } else {
                    RiderResult.Ok(token)
                }
            }
            is RiderResult.Err -> result
        }
    }

    companion object {
        const val PATH: String = "client/paas/device/getBlueToothToken"
    }
}

@Serializable
data class BleTokenDto(
    @Serializable(with = FlexibleStringSerializer::class)
    val token: String = "",
)
