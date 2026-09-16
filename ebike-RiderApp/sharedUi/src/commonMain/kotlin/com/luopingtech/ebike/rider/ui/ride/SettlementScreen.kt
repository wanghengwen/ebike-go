package com.luopingtech.ebike.rider.ui.ride

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableLongStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.rider.RiderApp
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.core.time.nowEpochMillis
import com.luopingtech.ebike.rider.domain.riding.RideFormat
import com.luopingtech.ebike.rider.feature.riding.RidingUiState
import com.luopingtech.ebike.rider.ui.theme.RiderTheme
import kotlinx.coroutines.delay

/**
 * 费用支付页 —— 对齐产品截图 / UniApp `pay.vue`：
 * 3 分钟倒计时 → 待支付金额 → 余额分桶 → 时长里程 → 可展开费用明细 → 报修 → 去支付。
 */
@Composable
fun SettlementScreen(
    app: RiderApp,
    state: RidingUiState,
    onFinish: () -> Unit,
    onPay: () -> Unit,
    onOpenBilling: () -> Unit,
    onOpenRepair: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val summary = state.settlement
    val rideSeconds = summary?.rideTimeSeconds ?: state.elapsedSeconds
    val waitPay = summary?.waitPayFen ?: 0
    val paid = summary?.settled == true || waitPay <= 0

    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)

    var showShortTrip by remember(summary?.orderId) { mutableStateOf(false) }
    var shortTripAsked by remember(summary?.orderId) { mutableStateOf(false) }
    var costOpen by remember { mutableStateOf(true) }
    var remainMs by remember(summary?.orderId) { mutableLongStateOf(PAY_WINDOW_MS) }

    LaunchedEffect(summary?.orderId) {
        app.ridingFeature.hydrateSettlement()
    }

    LaunchedEffect(summary?.orderId, rideSeconds) {
        if (shortTripAsked) return@LaunchedEffect
        if (rideSeconds in 1 until SHORT_TRIP_SECONDS) {
            shortTripAsked = true
            showShortTrip = true
        }
    }

    LaunchedEffect(summary?.orderId, summary?.frozenAtMillis, summary?.enteredAtMillis, paid) {
        if (paid) {
            remainMs = 0L
            return@LaunchedEffect
        }
        val start = when {
            (summary?.frozenAtMillis ?: 0L) > 0L -> summary!!.frozenAtMillis
            (summary?.enteredAtMillis ?: 0L) > 0L -> summary!!.enteredAtMillis
            else -> nowEpochMillis()
        }
        while (true) {
            val elapsed = (nowEpochMillis() - start).coerceAtLeast(0L)
            remainMs = (PAY_WINDOW_MS - elapsed).coerceAtLeast(0L)
            if (remainMs <= 0L) break
            delay(250L)
        }
    }

    val pageBg = Color(0xFFF5F5F5)
    val hintBg = Color(0xFFFFF4E8)
    val hintFg = Color(0xFFE67E22)
    val brand = RiderTheme.colors.primary

    Column(
        modifier = modifier
            .fillMaxSize()
            .background(pageBg),
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color.White)
                .padding(horizontal = 8.dp, vertical = 12.dp),
            contentAlignment = Alignment.Center,
        ) {
            Text(
                text = t(Str.SettleTitle),
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.SemiBold,
                color = RiderTheme.colors.textPrimary,
            )
        }

        if (!paid && remainMs > 0L) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(hintBg)
                    .padding(horizontal = 16.dp, vertical = 10.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Text(
                    text = "⏱",
                    color = hintFg,
                    modifier = Modifier.padding(end = 6.dp),
                )
                Text(
                    text = t(Str.SettlePayWindow),
                    style = MaterialTheme.typography.bodySmall,
                    color = hintFg,
                    modifier = Modifier.weight(1f),
                )
                Text(
                    text = RideFormat.countdownHms(remainMs),
                    style = MaterialTheme.typography.bodySmall,
                    fontWeight = FontWeight.SemiBold,
                    color = hintFg,
                )
            }
        }

        Column(
            modifier = Modifier
                .weight(1f)
                .fillMaxWidth()
                .verticalScroll(rememberScrollState())
                .padding(horizontal = 16.dp, vertical = 16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            Surface(
                shape = RoundedCornerShape(16.dp),
                color = Color.White,
                modifier = Modifier.fillMaxWidth(),
            ) {
                Column(
                    modifier = Modifier.padding(horizontal = 20.dp, vertical = 24.dp),
                    horizontalAlignment = Alignment.CenterHorizontally,
                ) {
                    Row(verticalAlignment = Alignment.Bottom) {
                        Text(
                            text = if (paid) t(Str.SettlePaidLabel) else t(Str.SettleUnpaidLabel),
                            style = MaterialTheme.typography.titleMedium,
                            color = RiderTheme.colors.textPrimary,
                            modifier = Modifier.padding(bottom = 4.dp, end = 6.dp),
                        )
                        Text(
                            text = RideFormat.yuanPlain(if (paid) summary?.costFeeFen ?: 0 else waitPay),
                            style = MaterialTheme.typography.displaySmall.copy(
                                fontWeight = FontWeight.Bold,
                                fontSize = 40.sp,
                            ),
                            color = RiderTheme.colors.textPrimary,
                        )
                        Text(
                            text = " ${t(Str.RideYuan)}",
                            style = MaterialTheme.typography.titleMedium,
                            color = RiderTheme.colors.textPrimary,
                            modifier = Modifier.padding(bottom = 6.dp, start = 2.dp),
                        )
                    }

                    if (!paid) {
                        Spacer(modifier = Modifier.height(10.dp))
                        Text(
                            text = "${t(Str.SettleRechargeBalance, RideFormat.yuanPlain(summary?.rechargeBalanceFen ?: 0))}  " +
                                t(Str.SettlePresentBalance, RideFormat.yuanPlain(summary?.presentBalanceFen ?: 0)),
                            style = MaterialTheme.typography.bodySmall,
                            color = RiderTheme.colors.textTertiary,
                            textAlign = TextAlign.Center,
                        )

                        Spacer(modifier = Modifier.height(14.dp))
                        Surface(
                            color = pageBg,
                            shape = RoundedCornerShape(8.dp),
                            modifier = Modifier.fillMaxWidth(),
                        ) {
                            Text(
                                text = t(
                                    Str.SettleRideDurationMile,
                                    RideFormat.durationChinese(rideSeconds),
                                    RideFormat.distanceKmChinese(summary?.rideDistanceMeters ?: 0),
                                ),
                                style = MaterialTheme.typography.bodyMedium,
                                color = RiderTheme.colors.textSecondary,
                                modifier = Modifier.padding(horizontal = 12.dp, vertical = 10.dp),
                                textAlign = TextAlign.Center,
                            )
                        }

                        if ((summary?.rideFeeParentFen ?: 0) > 0 || waitPay > 0) {
                            Spacer(modifier = Modifier.height(18.dp))
                            Row(
                                modifier = Modifier.fillMaxWidth(),
                                verticalAlignment = Alignment.CenterVertically,
                            ) {
                                Text(
                                    text = t(Str.SettleCostDetail),
                                    style = MaterialTheme.typography.titleSmall,
                                    fontWeight = FontWeight.SemiBold,
                                    color = RiderTheme.colors.textPrimary,
                                )
                                Text(
                                    text = " ?",
                                    color = RiderTheme.colors.textTertiary,
                                    modifier = Modifier
                                        .padding(start = 4.dp)
                                        .clickable(onClick = onOpenBilling),
                                )
                            }
                            Spacer(modifier = Modifier.height(8.dp))
                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .clickable { costOpen = !costOpen }
                                    .padding(vertical = 6.dp),
                                horizontalArrangement = Arrangement.SpaceBetween,
                            ) {
                                Text(
                                    text = t(Str.SettleRideFee),
                                    style = MaterialTheme.typography.bodyMedium,
                                    color = RiderTheme.colors.textPrimary,
                                )
                                Row(verticalAlignment = Alignment.CenterVertically) {
                                    Text(
                                        text = "${RideFormat.yuanPlain(summary?.rideFeeParentFen ?: waitPay)}${t(Str.RideYuan)}",
                                        style = MaterialTheme.typography.bodyMedium,
                                        color = RiderTheme.colors.textPrimary,
                                    )
                                    Text(
                                        text = if (costOpen) " ˄" else " ˅",
                                        color = RiderTheme.colors.textTertiary,
                                    )
                                }
                            }
                            if (costOpen) {
                                val yuan = t(Str.RideYuan)
                                CostChildRow(
                                    label = t(Str.SettleStartPrice),
                                    fen = summary?.startPriceFen,
                                    yuanUnit = yuan,
                                )
                                CostChildRow(
                                    label = t(Str.SettleDurationCost),
                                    fen = summary?.timeCostFen,
                                    yuanUnit = yuan,
                                )
                                CostChildRow(
                                    label = t(Str.SettleMileCost),
                                    fen = summary?.mileCostFen,
                                    yuanUnit = yuan,
                                )
                                if ((summary?.dispatchFeeFen ?: 0) > 0) {
                                    CostChildRow(
                                        label = t(Str.SettleDispatchFee),
                                        fen = summary?.dispatchFeeFen,
                                        yuanUnit = yuan,
                                    )
                                }
                            }
                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(top = 8.dp),
                                horizontalArrangement = Arrangement.End,
                            ) {
                                Text(
                                    text = t(
                                        Str.SettleSubtotal,
                                        RideFormat.yuanPlain(summary?.totalFen ?: waitPay),
                                    ),
                                    style = MaterialTheme.typography.bodyMedium,
                                    fontWeight = FontWeight.Medium,
                                    color = RiderTheme.colors.textPrimary,
                                )
                            }
                        }
                    }
                }
            }
        }

        Column(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color.White)
                .padding(horizontal = 16.dp, vertical = 12.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            TextButton(onClick = onOpenRepair) {
                Text(
                    text = "🔧 ${t(Str.SettleBikeRepair)}",
                    color = RiderTheme.colors.textSecondary,
                )
            }
            Button(
                onClick = {
                    if (paid) onFinish() else onPay()
                },
                modifier = Modifier
                    .fillMaxWidth()
                    .height(48.dp),
                shape = RoundedCornerShape(24.dp),
                colors = ButtonDefaults.buttonColors(
                    containerColor = brand,
                    contentColor = Color.White,
                ),
            ) {
                Text(
                    text = if (paid) {
                        t(Str.SettleFinish)
                    } else {
                        t(Str.SettleGoPay, RideFormat.yuanPlain(waitPay))
                    },
                    fontWeight = FontWeight.SemiBold,
                    fontSize = 16.sp,
                )
            }
        }
    }

    if (showShortTrip) {
        AlertDialog(
            onDismissRequest = { showShortTrip = false },
            title = { Text(t(Str.ShortTripFeedbackTitle)) },
            text = {
                Text(
                    text = t(Str.ShortTripFeedbackBody),
                    style = MaterialTheme.typography.bodyMedium,
                    color = RiderTheme.colors.textSecondary,
                )
            },
            confirmButton = {
                TextButton(
                    onClick = {
                        showShortTrip = false
                        onOpenRepair()
                    },
                ) {
                    Text(t(Str.ShortTripHasFault))
                }
            },
            dismissButton = {
                TextButton(onClick = { showShortTrip = false }) {
                    Text(t(Str.ShortTripNoFault))
                }
            },
        )
    }
}

@Composable
private fun CostChildRow(label: String, fen: Int?, yuanUnit: String = "元") {
    if (fen == null) return
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(start = 12.dp, top = 4.dp, bottom = 4.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
    ) {
        Text(
            text = label,
            style = MaterialTheme.typography.bodySmall,
            color = RiderTheme.colors.textTertiary,
        )
        Text(
            text = "${RideFormat.yuanPlain(fen)}$yuanUnit",
            style = MaterialTheme.typography.bodySmall,
            color = RiderTheme.colors.textTertiary,
        )
    }
}

private const val SHORT_TRIP_SECONDS = 3 * 60L
private const val PAY_WINDOW_MS = 3L * 60L * 1000L
