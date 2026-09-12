package com.luopingtech.ebike.ops.data.geo

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.platform.DeviceInfo
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.put

class GeoApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    /** Legacy: /business/ebike-management/gaode/getAddress */
    suspend fun getAddress(lat: Double, lng: Double): OpsResult<String> {
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/gaode/getAddress",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("latitude", lat)
            put("longitude", lng)
        }
        return when (
            val result = signedApi.post(
                path = "business/ebike-management/gaode/getAddress",
                bodyJson = body,
                deserializer = AddressDto.serializer(),
            )
        ) {
            is OpsResult.Ok -> {
                val address = result.value.address?.trim().orEmpty()
                if (address.isBlank()) {
                    OpsResult.Err(OpsError.business("GEO", "empty address"))
                } else {
                    OpsResult.Ok(address)
                }
            }
            is OpsResult.Err -> result
        }
    }
}

@Serializable
internal data class AddressDto(
    val address: String? = null,
)
