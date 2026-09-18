package com.luopingtech.ebike.rider.ui.ride

import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import com.luopingtech.ebike.rider.RiderApp
import com.luopingtech.ebike.rider.core.config.H5ScreenKind
import com.luopingtech.ebike.rider.core.config.H5ScreenUrls
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.domain.riding.ReturnDecision
import com.luopingtech.ebike.rider.domain.riding.RidePhase
import com.luopingtech.ebike.rider.feature.pay.NativePayOutcome
import com.luopingtech.ebike.rider.ui.feedback.LocalRiderToast
import com.luopingtech.ebike.rider.ui.home.HomeMapScreen
import com.luopingtech.ebike.rider.ui.permission.LocalRiderBlePermissionGate
import com.luopingtech.ebike.rider.ui.permission.LocalRiderLocationPermissionGate
import kotlinx.coroutines.launch

/**
 * 骑行域的「叠加层」路由。
 *
 * 为什么由相位驱动而不是导航栈：骑行相位可以被服务端轮询和杀进程恢复改写，
 * 用栈的话会出现「相位已回 Idle 但界面还压着骑行页」。这里让 [RidePhase] 当唯一真源，
 * 只有相位内部的次级页面（找 P / 引导 / 申诉 / 轨迹）用局部状态。
 */
@Composable
fun RideFlowHost(
    app: RiderApp,
    onOpenH5: (H5ScreenKind?, String?) -> Unit,
    modifier: Modifier = Modifier,
) {
    val feature = app.ridingFeature
    val state by feature.state.collectAsState()
    val scope = rememberCoroutineScope()
    val toast = LocalRiderToast.current
    val locationGate = LocalRiderLocationPermissionGate.current
    val blePermissionGate = LocalRiderBlePermissionGate.current

    var overlay by remember { mutableStateOf<RideOverlay>(RideOverlay.None) }

    // feature 的一次性提示统一在这里落成 toast，各屏不用各自接。
    LaunchedEffect(state.message?.nonce) {
        state.message?.let {
            toast(it.text)
            feature.consumeMessage()
        }
    }

    // 相位跌回骑行 / 空闲时，次级页面要跟着收起来，否则会挂在过期的上下文上。
    LaunchedEffect(state.phase) {
        if (state.phase == RidePhase.Idle || state.phase == RidePhase.Settling) {
            overlay = RideOverlay.None
        }
    }

    LaunchedEffect(state.session.isOnTrip) {
        if (state.session.isOnTrip) {
            feature.unfreezePreviousOrder()
            feature.loadFences()
        }
    }

    fun withLocation(block: suspend () -> Unit) {
        locationGate.ensure { denied ->
            if (denied != null) {
                toast(denied)
                return@ensure
            }
            scope.launch { block() }
        }
    }

    /** 开锁要同时拿到定位与蓝牙权限：BLE 优先策略下没有蓝牙就只能走网络降级。 */
    fun unlock() {
        locationGate.ensure { deniedLocation ->
            if (deniedLocation != null) {
                toast(deniedLocation)
                return@ensure
            }
            blePermissionGate.ensure { deniedBle ->
                // 蓝牙被拒不拦：网络远程开锁仍然可用，只是慢一点。
                if (deniedBle != null) toast(deniedBle)
                scope.launch { feature.confirmUnlock() }
            }
        }
    }

    when (val active = overlay) {
        is RideOverlay.ManualId -> {
            ManualIdScreen(
                app = app,
                onSubmit = { carId ->
                    overlay = RideOverlay.None
                    scope.launch { feature.selectVehicle(carId = carId, manual = true) }
                },
                onBack = { overlay = RideOverlay.None },
                modifier = modifier,
            )
            return
        }

        is RideOverlay.ParkSearch -> {
            ParkSearchScreen(
                app = app,
                state = state,
                onRefresh = { withLocation { feature.loadNearParking() } },
                onBack = { overlay = RideOverlay.None },
                modifier = modifier,
            )
            return
        }

        is RideOverlay.Guide -> {
            ReturnGuideScreen(
                app = app,
                pageType = active.pageType,
                showApplyEntry = state.config.showApplyEntry,
                onRetryReturn = {
                    overlay = RideOverlay.None
                    withLocation { feature.beginReturn() }
                },
                onApply = {
                    overlay = RideOverlay.Apply(
                        applyType = ReturnDecision.guideApplyType(active.pageType),
                    )
                },
                onBack = { overlay = RideOverlay.None },
                modifier = modifier,
            )
            return
        }

        is RideOverlay.Apply -> {
            ApplyReturnScreen(
                app = app,
                applyType = active.applyType,
                // 申诉通过后 feature 会置 autoReturn，回骑行页由轮询自动再走一次还车。
                onSubmitted = { overlay = RideOverlay.None },
                onBack = { overlay = RideOverlay.None },
                modifier = modifier,
            )
            return
        }

        is RideOverlay.TripMap -> {
            TripMapScreen(
                app = app,
                state = state,
                onBack = { overlay = RideOverlay.None },
                modifier = modifier,
            )
            return
        }

        RideOverlay.None -> Unit
    }

    when (state.phase) {
        RidePhase.Idle -> HomeMapScreen(
            app = app,
            onOpenH5 = onOpenH5,
            onStartRide = { carId, imei ->
                scope.launch { feature.selectVehicle(carId = carId, imei = imei) }
            },
            onOpenScan = { feature.openScan() },
            modifier = modifier,
        )

        RidePhase.Scanning -> ScanScreen(
            app = app,
            onScanned = { carId, imei ->
                scope.launch { feature.selectVehicle(carId = carId, imei = imei) }
            },
            onBack = { feature.cancelSelection() },
            modifier = modifier,
        )

        RidePhase.Confirming, RidePhase.Unlocking -> PreCyclingScreen(
            app = app,
            state = state,
            onUnlock = { unlock() },
            onChangeVehicle = { feature.openScan() },
            onBack = { feature.cancelSelection() },
            onOpenBilling = { onOpenH5(H5ScreenKind.BillingRules, null) },
            modifier = modifier,
        )

        RidePhase.Riding, RidePhase.TempLocked, RidePhase.Returning -> RidingScreen(
            app = app,
            state = state,
            onTempLockToggle = { scope.launch { feature.toggleTempLock() } },
            onReturn = { withLocation { feature.beginReturn() } },
            onRing = { scope.launch { feature.ringVehicle() } },
            onHelmet = { scope.launch { feature.openHelmetLock() } },
            onFindParking = {
                overlay = RideOverlay.ParkSearch
                withLocation { feature.loadNearParking() }
            },
            onOpenTripMap = { overlay = RideOverlay.TripMap },
            onApplyReturn = {
                overlay = RideOverlay.Apply(
                    applyType = ReturnDecision.applyType(state.session.returnTypeCode),
                )
            },
            onDismissPrompt = { feature.dismissPrompt() },
            onConfirmReturn = { forcePenalty ->
                scope.launch { feature.confirmReturn(forcePenalty) }
            },
            onOpenGuide = { pageType, _ ->
                feature.dismissPrompt()
                overlay = RideOverlay.Guide(pageType)
            },
            onRecoverPower = { scope.launch { feature.recoverPower() } },
            onRecharge = {
                feature.dismissPrompt()
                onOpenH5(H5ScreenKind.Wallet, null)
            },
            modifier = modifier,
        )

        RidePhase.Settling -> SettlementScreen(
            app = app,
            state = state,
            onFinish = { feature.finishSettlement() },
            onPay = {
                scope.launch {
                    when (val outcome = feature.payPendingOrder()) {
                        NativePayOutcome.Paid -> Unit
                        NativePayOutcome.Cancelled -> toast(app.i18n.t(Str.PayCancelled))
                        is NativePayOutcome.Failed -> toast(
                            outcome.message.ifBlank { app.i18n.t(Str.PayFailed) },
                        )
                        NativePayOutcome.FallbackH5 -> {
                            val orderId = state.settlement?.orderId.orEmpty()
                                .ifBlank { state.session.orderId }
                            val hash = H5ScreenUrls.appendQuery(
                                H5ScreenKind.Pay.route,
                                if (orderId.isBlank()) emptyMap() else mapOf("orderId" to orderId),
                            )
                            onOpenH5(null, hash)
                        }
                    }
                }
            },
            onOpenBilling = { onOpenH5(H5ScreenKind.BillingRules, null) },
            onOpenRepair = {
                val carId = state.carId
                val orderId = state.settlement?.orderId.orEmpty().ifBlank { state.session.orderId }
                val params = buildMap {
                    if (carId.isNotBlank()) put("carId", carId)
                    if (orderId.isNotBlank()) put("orderId", orderId)
                }
                onOpenH5(null, H5ScreenUrls.appendQuery(H5ScreenKind.Repair.route, params))
            },
            modifier = modifier,
        )
    }
}

/** 相位内部的次级页面。与 [RidePhase] 正交，所以单独用局部状态管。 */
private sealed interface RideOverlay {
    data object None : RideOverlay
    data object ManualId : RideOverlay
    data object ParkSearch : RideOverlay
    data object TripMap : RideOverlay
    data class Guide(val pageType: Int) : RideOverlay
    data class Apply(val applyType: Int) : RideOverlay
}
