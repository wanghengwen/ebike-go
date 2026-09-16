package com.luopingtech.ebike.rider.data.riding

import com.luopingtech.ebike.rider.core.json.LooseJson
import com.luopingtech.ebike.rider.core.network.CommonRequestBody
import com.luopingtech.ebike.rider.core.network.SignedApiClient
import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.platform.DeviceInfo
import kotlinx.serialization.json.put

/**
 * 「无法还车」人工申诉（UniApp `api/applyReturn.ts`）。
 *
 * 提交成功后旧版会往 storage 写 `autoLock=true`，回骑行页时自动再跑一次还车 ——
 * 这里由 [com.luopingtech.ebike.rider.feature.riding.RidingFeature.consumeAutoReturn] 承接。
 */
interface ApplyReturnRemote {
    /** 这个 applyType 现在还能不能申诉（`izCapable`）。 */
    suspend fun canApply(applyType: Int?): RiderResult<Boolean>

    suspend fun createAudit(
        photoUrls: List<String>,
        orderId: String,
        applyType: Int?,
        reason: String,
    ): RiderResult<Unit>
}

class ApplyReturnApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) : ApplyReturnRemote {

    override suspend fun canApply(applyType: Int?): RiderResult<Boolean> {
        val body = CommonRequestBody.toJsonString(
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            if (applyType != null) put("applyType", applyType)
        }
        return when (val out = signedApi.postOutcome(PATH_CAPABLE, body)) {
            is RiderResult.Err -> out
            is RiderResult.Ok ->
                if (!out.value.success) {
                    RiderResult.Ok(false)
                } else {
                    RiderResult.Ok(parseCapable(out.value.data))
                }
        }
    }

    override suspend fun createAudit(
        photoUrls: List<String>,
        orderId: String,
        applyType: Int?,
        reason: String,
    ): RiderResult<Unit> {
        if (photoUrls.isEmpty()) {
            return RiderResult.Err(RiderError.business("PHOTO_REQUIRED", "at least one photo"))
        }
        val body = CommonRequestBody.toJsonString(
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            // 旧版就是逗号拼接一个字符串字段，不是数组。
            put("photoUrl", photoUrls.joinToString(","))
            if (orderId.isNotBlank()) put("orderId", orderId)
            if (applyType != null) put("applyType", applyType)
            put("userReason", reason)
        }
        return signedApi.postUnit(PATH_CREATE, body)
    }

    companion object {
        const val PATH_CAPABLE: String = "client/returnBikeAudit/izCapable"
        const val PATH_CREATE: String = "client/returnBikeAudit/createReturnBikeAudit"

        /**
         * `izCapable` 的 `data` 有两种形态：裸布尔 `true`，或 `{ "izCapable": true }`。
         * 都不认识时按「不能申诉」处理 —— 误开申诉入口比误关更难解释。
         */
        internal fun parseCapable(data: kotlinx.serialization.json.JsonElement?): Boolean {
            if (data == null) return false
            (data as? kotlinx.serialization.json.JsonPrimitive)?.let { primitive ->
                val text = primitive.content.trim()
                if (text.equals("true", ignoreCase = true)) return true
                if (text.equals("false", ignoreCase = true)) return false
                return (text.toDoubleOrNull() ?: 0.0) != 0.0
            }
            val o = LooseJson.obj(data)
            return LooseJson.boolOrNull(o, "izCapable", "capable", "result") ?: false
        }
    }
}
