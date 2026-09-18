package com.luopingtech.ebike.rider.data.pay

import com.luopingtech.ebike.rider.core.network.CommonRequestBody
import com.luopingtech.ebike.rider.core.network.SignedApiClient
import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.platform.DeviceInfo
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonObjectBuilder
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.put

interface PayRemote {
    suspend fun create(
        amountFen: Int,
        orderId: String,
        pin: String,
        serviceAreaId: String,
        channelType: String,
        saleType: String,
    ): RiderResult<JsonElement?>

    suspend fun cancel(pin: String, outTradeNo: String): RiderResult<Unit>
}

class PayApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) : PayRemote {

    private fun body(block: JsonObjectBuilder.() -> Unit = {}): String =
        CommonRequestBody.toJsonString(
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
            block = block,
        )

    override suspend fun create(
        amountFen: Int,
        orderId: String,
        pin: String,
        serviceAreaId: String,
        channelType: String,
        saleType: String,
    ): RiderResult<JsonElement?> {
        val json = body {
            put("channel_type", channelType)
            put("channel_info", buildJsonObject { })
            put("sale_type", saleType)
            put("sale_info", buildJsonObject { put("total_fee", amountFen) })
            if (pin.isNotBlank()) put("pin", pin)
            if (serviceAreaId.isNotBlank()) put("service_id", serviceAreaId)
            if (orderId.isNotBlank()) {
                put("orderId", orderId)
                put("id", orderId)
            }
        }
        return when (val out = signedApi.postOutcome(PATH_CREATE, json)) {
            is RiderResult.Err -> out
            is RiderResult.Ok -> {
                if (out.value.success) {
                    RiderResult.Ok(out.value.data)
                } else {
                    RiderResult.Err(RiderError.business(out.value.code, out.value.message))
                }
            }
        }
    }

    override suspend fun cancel(pin: String, outTradeNo: String): RiderResult<Unit> {
        val json = body {
            if (pin.isNotBlank()) put("pin", pin)
            if (outTradeNo.isNotBlank()) put("out_trade_no", outTradeNo)
        }
        return when (val out = signedApi.postOutcome(PATH_CANCEL, json)) {
            is RiderResult.Err -> out
            is RiderResult.Ok -> {
                if (out.value.success) RiderResult.Ok(Unit)
                else RiderResult.Err(RiderError.business(out.value.code, out.value.message))
            }
        }
    }

    companion object {
        const val PATH_CREATE: String = "client/ebike-pay/pay/create"
        const val PATH_CANCEL: String = "client/ebike-pay/pay/cancel"
    }
}

/** 凭证未就绪时走 H5，demo 包没有商户号。 */
class DemoPayRemote : PayRemote {
    override suspend fun create(
        amountFen: Int,
        orderId: String,
        pin: String,
        serviceAreaId: String,
        channelType: String,
        saleType: String,
    ): RiderResult<JsonElement?> = RiderResult.Err(RiderError.business(CODE_FALLBACK_H5, "demo pay"))

    override suspend fun cancel(pin: String, outTradeNo: String): RiderResult<Unit> =
        RiderResult.Ok(Unit)

    companion object {
        const val CODE_FALLBACK_H5: String = "FALLBACK_H5"
    }
}
