package com.luopingtech.ebike.rider.domain.riding

import com.luopingtech.ebike.rider.core.i18n.Str

/**
 * `returnPermission` 的围栏判定码（**不是** `returnByNet` 用的 0|1 returnType）。
 * 逐条对齐 UniApp `features/bike/returnTypes.ts` 的 `RETURN_TYPE_META`。
 */
object ReturnTypeCodes {
    const val NORMAL: Int = 1101
    const val OUT_OF_PARK: Int = 1102
    const val OUT_OF_SPOT: Int = 2101
    const val DIRECTIONAL: Int = 2201
    const val RFID: Int = 2211
    const val KICKSTAND: Int = 2231
    const val CAMERA_DIRECTION: Int = 2241
    const val CAMERA_POINT: Int = 2242
    const val CAMERA_MISS: Int = 2243
    const val HELMET_MISS_A: Int = 2251
    const val HELMET_MISS_B: Int = 2252
    const val NO_PARKING: Int = 3101
    const val NO_PARKING_HELMET: Int = 3102
    const val OUT_OF_SERVICE: Int = 4101
    const val OUT_OF_SERVICE_HELMET: Int = 4102
    const val FORBID_ZONE: Int = 5101
    const val FULL_PILE: Int = 6101

    /** 骑行中 `getRideInfo` 主动弹的「已出服务区」，不是 permission 返回值。 */
    const val RIDE_INFO_OUT_OF_SERVICE: Int = 10001
}

/** `returnByNet` 的 returnType：0 普通、1 认罚（付调度费）。 */
object ReturnByNetType {
    const val NORMAL: Int = 0
    const val ACCEPT_PENALTY: Int = 1

    /** BLE 还车上报固定 3。 */
    const val BLE_REPORT: Int = 3
}

/** 只有 `11xx` 段算正常还车，其余都要走认罚 / 引导。 */
enum class ReturnKind { Normal, Unnormal }

/** `returnPermission` 之后界面该走哪条路。对齐 `decideReturnFlow`。 */
enum class ReturnFlow {
    /** 文明停车提醒（可还 + 正常 + 租户开了提醒）。 */
    Civilization,

    /** 认罚弹层：可还但不在站点，收调度费。 */
    PenaltySheet,

    /** 跳定制还车引导页（头盔 / RFID / 边撑 / 摄像头）。 */
    Guide,

    /** 不可还且是已知围栏码：弹阻塞层解释原因。 */
    BlockSheet,

    /** 可以直接提交。 */
    Normal,

    /** 不可还且是未知码 —— 旧版在这里什么都不做，保留。 */
    Silent,
}

/** 定制还车引导页的 pageType（legacy `customizedReturn`）。 */
object ReturnGuidePageType {
    const val DIRECTIONAL: Int = 4
    const val RFID: Int = 5
    const val HELMET: Int = 6
    const val KICKSTAND: Int = 8
    const val CAMERA: Int = 11
}

/** 还车原因文案 + 是否附带「出服务区」补充说明。 */
data class ReturnTypeMeta(
    val reason: Str,
    val tips: Str? = null,
    val title: Str? = null,
    /** true = 旧版用居中对话框而不是底部弹层（目前只有满桩）。 */
    val dialog: Boolean = false,
)

object ReturnDecision {

    private val META: Map<Int, ReturnTypeMeta> = mapOf(
        ReturnTypeCodes.NORMAL to ReturnTypeMeta(Str.ReturnReasonNormal),
        ReturnTypeCodes.OUT_OF_PARK to ReturnTypeMeta(Str.ReturnReasonOutOfPark),
        ReturnTypeCodes.OUT_OF_SPOT to ReturnTypeMeta(Str.ReturnReasonOutOfPark),
        ReturnTypeCodes.DIRECTIONAL to ReturnTypeMeta(Str.ReturnReasonOutOfPark),
        ReturnTypeCodes.RFID to ReturnTypeMeta(Str.ReturnReasonOutOfPark),
        ReturnTypeCodes.KICKSTAND to ReturnTypeMeta(Str.ReturnReasonOutOfPark),
        ReturnTypeCodes.CAMERA_DIRECTION to ReturnTypeMeta(Str.ReturnReasonCameraDirection),
        ReturnTypeCodes.CAMERA_POINT to ReturnTypeMeta(Str.ReturnReasonCameraPoint),
        ReturnTypeCodes.CAMERA_MISS to ReturnTypeMeta(Str.ReturnReasonCameraMiss),
        ReturnTypeCodes.HELMET_MISS_A to ReturnTypeMeta(Str.ReturnReasonHelmetMiss),
        ReturnTypeCodes.HELMET_MISS_B to ReturnTypeMeta(Str.ReturnReasonHelmetMiss),
        ReturnTypeCodes.NO_PARKING to ReturnTypeMeta(Str.ReturnReasonNoParking),
        ReturnTypeCodes.NO_PARKING_HELMET to ReturnTypeMeta(Str.ReturnReasonHelmetMiss),
        ReturnTypeCodes.OUT_OF_SERVICE to ReturnTypeMeta(
            reason = Str.ReturnReasonOutOfService,
            tips = Str.ReturnReasonOutOfServiceTips,
        ),
        ReturnTypeCodes.OUT_OF_SERVICE_HELMET to ReturnTypeMeta(Str.ReturnReasonHelmetMiss),
        ReturnTypeCodes.FORBID_ZONE to ReturnTypeMeta(Str.ReturnReasonForbidZone),
        ReturnTypeCodes.FULL_PILE to ReturnTypeMeta(
            reason = Str.ReturnReasonFullPile,
            title = Str.ReturnReasonFullPileTitle,
            dialog = true,
        ),
        ReturnTypeCodes.RIDE_INFO_OUT_OF_SERVICE to ReturnTypeMeta(
            reason = Str.ReturnReasonOutOfService,
            tips = Str.ReturnReasonOutOfServiceTips,
        ),
    )

    fun meta(returnTypeCode: Int): ReturnTypeMeta =
        META[returnTypeCode] ?: ReturnTypeMeta(Str.ReturnReasonOutOfPark)

    /** `String(type).startsWith('11')` 的等价实现，避免负数与前导零上的字符串技巧。 */
    fun kind(returnTypeCode: Int): ReturnKind =
        if (returnTypeCode in 1100..1199) ReturnKind.Normal else ReturnKind.Unnormal

    fun flow(
        canReturn: Boolean,
        returnTypeCode: Int,
        civilizationRemind: Boolean,
    ): ReturnFlow {
        val kind = kind(returnTypeCode)
        if (canReturn && kind == ReturnKind.Normal) {
            return if (civilizationRemind) ReturnFlow.Civilization else ReturnFlow.Normal
        }
        if (canReturn) return ReturnFlow.PenaltySheet
        if (guidePageType(returnTypeCode) != null) return ReturnFlow.Guide
        if (showSheetWhenCannotReturn(returnTypeCode)) return ReturnFlow.BlockSheet
        return ReturnFlow.Silent
    }

    /** `applyReturnEbike.applyType`：1 站点外 / 2 禁停区 / 3 其它。 */
    fun applyType(returnTypeCode: Int): Int = when (returnTypeCode) {
        in 2000..2999 -> 1
        in 3000..3999 -> 2
        else -> 3
    }

    fun guidePageType(returnTypeCode: Int): Int? = when (returnTypeCode) {
        ReturnTypeCodes.DIRECTIONAL -> ReturnGuidePageType.DIRECTIONAL
        ReturnTypeCodes.RFID -> ReturnGuidePageType.RFID
        ReturnTypeCodes.HELMET_MISS_A,
        ReturnTypeCodes.HELMET_MISS_B,
        ReturnTypeCodes.OUT_OF_SERVICE_HELMET,
        ReturnTypeCodes.NO_PARKING_HELMET,
        -> ReturnGuidePageType.HELMET
        ReturnTypeCodes.KICKSTAND -> ReturnGuidePageType.KICKSTAND
        ReturnTypeCodes.CAMERA_MISS -> ReturnGuidePageType.CAMERA
        else -> null
    }

    /**
     * 不可还但仍弹层解释（而不是跳引导页）的码。旧版 `goReturnBikeGuide` 只列了这四个；
     * 10001 是骑行页自己弹的，不在 permission 分支里。
     */
    fun showSheetWhenCannotReturn(returnTypeCode: Int): Boolean = returnTypeCode in setOf(
        ReturnTypeCodes.OUT_OF_SPOT,
        ReturnTypeCodes.NO_PARKING,
        ReturnTypeCodes.OUT_OF_SERVICE,
        ReturnTypeCodes.FULL_PILE,
    )

    /** 引导页拍照申诉的 applyType（legacy `takePhotoReturnCarType`）：11 → 7，其余同值。 */
    fun guideApplyType(pageType: Int): Int =
        if (pageType == ReturnGuidePageType.CAMERA) 7 else pageType
}
