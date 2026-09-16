package com.luopingtech.ebike.rider.domain.riding

import com.luopingtech.ebike.rider.core.i18n.Str

enum class FenceTipSeverity {
    /** 白底黑字，只是提示。 */
    Info,

    /** 红底白字，出站 / 禁停 / 出服务区。 */
    Warning,
}

/**
 * 骑行页顶部围栏横幅。
 *
 * UniApp `ridingFenceTips.ts` 把中文文案和 `#FF5936` 一起硬编码在里面；这里只产出
 * [Str] 键 + 参数 + 严重级，颜色交给 `RiderTheme`，文案交给词表 —— 否则英文包永远漏这条。
 */
data class RidingFenceTip(
    val text: Str,
    /** [text] 的格式化参数（目前只有调度费金额）。 */
    val textArg: String = "",
    val shortTip: Str? = null,
    val severity: FenceTipSeverity = FenceTipSeverity.Info,
)

object RidingFenceTips {

    /**
     * @param fenceTypeCode `getRideInfo().type`
     * @param dispatchCostFen 调度费（分）
     * @param canReturn `getRideInfo().izCanReturn`
     */
    fun of(fenceTypeCode: Int, dispatchCostFen: Int, canReturn: Boolean): RidingFenceTip? {
        if (fenceTypeCode == 0) return null
        return when (fenceTypeCode) {
            ReturnTypeCodes.NORMAL -> RidingFenceTip(
                text = Str.FenceTipRideInArea,
                severity = FenceTipSeverity.Info,
            )

            ReturnTypeCodes.OUT_OF_SPOT, ReturnTypeCodes.RFID -> {
                // 旧版：只有「可还 + 确实要收钱」才把金额写进横幅。
                val withFee = canReturn && dispatchCostFen != 0
                if (withFee) {
                    RidingFenceTip(
                        text = Str.FenceTipOutOfSpotWithFee,
                        textArg = RideFormat.yuan(dispatchCostFen),
                        shortTip = Str.FenceTipOutOfSpotShortWithFee,
                        severity = FenceTipSeverity.Warning,
                    )
                } else {
                    RidingFenceTip(
                        text = Str.FenceTipOutOfSpot,
                        shortTip = Str.FenceTipOutOfSpotShort,
                        severity = FenceTipSeverity.Warning,
                    )
                }
            }

            ReturnTypeCodes.NO_PARKING -> RidingFenceTip(
                text = Str.FenceTipNoParking,
                shortTip = Str.FenceTipNoParkingShort,
                severity = FenceTipSeverity.Warning,
            )

            ReturnTypeCodes.OUT_OF_SERVICE, ReturnTypeCodes.OUT_OF_SERVICE_HELMET -> RidingFenceTip(
                text = Str.FenceTipOutOfService,
                shortTip = Str.FenceTipOutOfServiceShort,
                severity = FenceTipSeverity.Warning,
            )

            else -> null
        }
    }
}
