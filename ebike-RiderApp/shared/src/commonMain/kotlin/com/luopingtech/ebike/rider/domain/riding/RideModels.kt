package com.luopingtech.ebike.rider.domain.riding

/**
 * `client/rent/getCarInfo` 的车辆详情（确认页 / 弱网还车判定都用它）。
 */
data class VehicleDetail(
    val carId: String,
    val imei: String = "",
    /** 剩余电量百分比，缺失为 null（界面显示 `--` 而不是 0）。 */
    val restBattery: Int? = null,
    /** 剩余可骑里程（km）。 */
    val restMileage: Int? = null,
    /**
     * 旧版 `isDisconnect`：车机离线。为真时还车 / 开锁直接走 BLE，
     * 别再花一轮超时去等远程指令。
     */
    val disconnected: Boolean = false,
    /** 旧版 `errType`：非 0 即车辆不可用，界面换「换车」布局。 */
    val errType: Int = 0,
    /**
     * 旧版 `ridingType`：有值且 ≠ 1 时与 UniApp 一样视为不可用
     *（即使 `errType` 仍是 0）。
     */
    val ridingType: Int? = null,
    /** 旧版 `isOutofServAera`：在服务区外，禁止开锁。 */
    val outOfServiceArea: Boolean = false,
    /** 红包车（免骑行费）。 */
    val redEnvelope: Boolean = false,
    val startPriceFen: Int? = null,
    val startPriceMinutes: Int = 0,
    val lat: Double = 0.0,
    val lng: Double = 0.0,
) {
    val available: Boolean
        get() = errType == 0 && (ridingType == null || ridingType == 1)

    val hasLocation: Boolean get() = lat != 0.0 || lng != 0.0
}

/**
 * `client/rent/getRideInfo` 的一轮轮询结果。
 *
 * 单位在这里就统一好：后端 `rideTime` 是毫秒、`rideDistance` 是米、费用是分。
 * 旧版把换算散在各个 `computed` 里，改一处漏一处。
 */
data class RideInfo(
    val orderId: String = "",
    val carId: String = "",
    val imei: String = "",
    /** 3 = 临停，4/5/6 = 骑行中，null = 无进行中订单。 */
    val ridingState: Int? = null,
    val costFeeFen: Int = 0,
    val rideTimeMillis: Long = 0L,
    val rideDistanceMeters: Int = 0,
    val restBattery: Int? = null,
    val restMileage: Int? = null,
    /** 2 = 无头盔锁；其余值显示头盔按钮。 */
    val helmetState: Int? = null,
    /** 1 = 开锁后首次提醒，2 = 佩戴提醒。 */
    val helmetPopup: Int = 0,
    /** 围栏码，喂给 [RidingFenceTips]。 */
    val fenceTypeCode: Int = 0,
    val dispatchCostFen: Int = 0,
    val canReturn: Boolean = false,
    val lat: Double = 0.0,
    val lng: Double = 0.0,
    /** 2 / 3 = 超载（断电），弹阻塞提示。 */
    val overloadState: Int = 0,
    /** 在运营区边缘，骑出会自动断电。 */
    val nearServiceEdge: Boolean = false,
    /** 处在免费还车时段内。 */
    val inFreeTime: Boolean = false,
    /** 平台主动临时通电：1 = 待用户确认恢复动力。 */
    val tempUnlockState: Int = 0,
    /** 已处于临时通电状态，不要重复弹窗。 */
    val tempUnlocked: Boolean = false,
) {
    val hasVehicleLocation: Boolean get() = lat != 0.0 && lng != 0.0
    val rideTimeSeconds: Long get() = rideTimeMillis / 1000L
}

/** `client/rent/returnPermission` 的判定结果。 */
data class ReturnPermissionResult(
    val canReturn: Boolean,
    val returnTypeCode: Int,
    val penaltyFen: Int = 0,
)

/** `client/system/getbackCarConfig` —— 还车相关的租户开关。 */
data class BackCarConfig(
    /** 正常还车前先弹文明停车提醒。 */
    val civilizationRemind: Boolean = false,
    /**
     * 旧版 `dispatchFee` 在这个接口里是布尔（是否展示「无法还车？」申诉入口），
     * 不是金额 —— 名字骗过不止一个人。
     */
    val showApplyEntry: Boolean = false,
    /** 出服务区后自动锁车的分钟数。 */
    val autoLockMinutes: Int = 3,
)

/** BLE 还车前要不要读 beacon（`getbackCarConfigByCarId`）。 */
data class BleReturnConfig(
    val needBeacon: Boolean = false,
)

/** 结费 / 费用支付屏数据（对齐 UniApp `pay.vue` 订单详情字段）。 */
data class SettlementSummary(
    val orderId: String,
    /** 待支付金额（分），优先 payCost / waitPay。 */
    val costFeeFen: Int,
    val dispatchFeeFen: Int = 0,
    val rideTimeSeconds: Long = 0L,
    val rideDistanceMeters: Int = 0,
    /** true = 已付清，无需再付。 */
    val settled: Boolean = false,
    /** 骑行费用合计 originCost（分），费用明细父项。 */
    val originCostFen: Int = 0,
    val startPriceFen: Int? = null,
    val timeCostFen: Int? = null,
    val mileCostFen: Int? = null,
    /** 充值余额（分）。 */
    val rechargeBalanceFen: Int = 0,
    /** 赠送余额（分）。 */
    val presentBalanceFen: Int = 0,
    /** 冻结起点 epoch millis；>0 时显示 3 分钟支付倒计时。 */
    val frozenAtMillis: Long = 0L,
    /** 进入结费页的本地时间，frozenAt 缺失时用它做 3 分钟窗。 */
    val enteredAtMillis: Long = 0L,
) {
    val totalFen: Int get() = costFeeFen + dispatchFeeFen
    val waitPayFen: Int get() = if (settled) 0 else costFeeFen.coerceAtLeast(0)
    val rideFeeParentFen: Int get() = if (originCostFen > 0) originCostFen else costFeeFen
}

/** `client/fence/parking/nearParkingNum` —— 找 P 点。 */
data class NearParking(
    val count: Int = 0,
    val distanceMeters: Double? = null,
    val lat: Double? = null,
    val lng: Double? = null,
    val name: String = "",
    val outOfService: Boolean = false,
) {
    val hasTarget: Boolean get() = (lat ?: 0.0) != 0.0 && (lng ?: 0.0) != 0.0
}

/** 临停前的配件校验（`partMatchByTempParking`）：头盔没归位就拦。 */
data class PartMatchResult(
    val matched: Boolean = true,
    val partName: String = "",
) {
    val helmetNotReturned: Boolean get() = !matched && partName == "helmet"
}
