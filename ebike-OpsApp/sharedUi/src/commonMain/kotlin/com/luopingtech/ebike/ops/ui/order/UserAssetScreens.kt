package com.luopingtech.ebike.ops.ui.order

import androidx.compose.foundation.background
import androidx.compose.foundation.border
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
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Text
import androidx.compose.material3.TextField
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.domain.order.OrderDepositRecord
import com.luopingtech.ebike.ops.domain.order.OrderFormat
import com.luopingtech.ebike.ops.domain.order.OrderRideCard
import com.luopingtech.ebike.ops.domain.order.OrderRideCardRecord
import com.luopingtech.ebike.ops.domain.order.OrderUserAssets
import com.luopingtech.ebike.ops.domain.order.OrderUserDetail
import com.luopingtech.ebike.ops.domain.order.OrderWalletInfo
import com.luopingtech.ebike.ops.domain.order.OrderWalletRecord
import com.luopingtech.ebike.ops.feature.order.OrderQueryFeature
import com.luopingtech.ebike.ops.ui.icons.OpsBackChevron
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import kotlinx.coroutines.launch

private val TextPrimary = Color(0xFF242936)
private val ThemeBlue = Color(0xFF1180F9)
private val LabelGray = Color(0xFF7C87B1)
private val SubGray = Color(0xFF9FA7C7)
private val PaidGreen = Color(0xFF1DBA4F)
private val UnpaidRed = Color(0xFFFF2222)
private val TipOrangeBg = Color(0xFFFFAD00)
private val TipOrange = Color(0xFFFA6400)
private val SectionGap = Color(0xFFF6F6F6)
private val DividerLine = Color(0xFFE5E5E5)

private typealias AssetTr = (Str, Array<out Any?>) -> String

@Composable
internal fun UserWalletPage(
    t: AssetTr,
    feature: OrderQueryFeature,
    wallet: OrderWalletInfo?,
    records: List<OrderWalletRecord>,
    loading: Boolean,
    onBack: () -> Unit,
) {
    val colors = OpsTheme.colors
    val scope = rememberCoroutineScope()
    var dialog by remember { mutableStateOf<WalletDialog?>(null) }
    LaunchedEffect(Unit) { feature.loadWallet() }

    Column(Modifier.fillMaxSize().background(Color.White)) {
        AssetTopBar(t(Str.OrderQueryWalletBalance, emptyArray()), onBack)
        LazyColumn(Modifier.weight(1f)) {
            item {
                Column(Modifier.fillMaxWidth().padding(bottom = 8.dp)) {
                    Row(
                        Modifier.padding(start = 16.dp, top = 22.dp),
                        verticalAlignment = Alignment.Bottom,
                    ) {
                        Text(t(Str.OrderQueryTotalAmount, emptyArray()), color = TextPrimary, fontSize = 16.sp)
                        Text(
                            OrderFormat.yuan(wallet?.balanceFen),
                            color = ThemeBlue,
                            fontSize = 26.sp,
                            fontWeight = FontWeight.Bold,
                        )
                        Spacer(Modifier.width(4.dp))
                        Text(t(Str.OrderQueryCny, emptyArray()), color = ThemeBlue, fontSize = 15.sp)
                    }
                    Column(
                        Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 16.dp, vertical = 16.dp)
                            .background(Color(0x123366CD), RoundedCornerShape(8.dp))
                            .padding(16.dp),
                    ) {
                        Row(verticalAlignment = Alignment.Bottom) {
                            Text(OrderFormat.yuan(wallet?.rechargeFen), color = ThemeBlue, fontSize = 26.sp)
                            Spacer(Modifier.width(2.dp))
                            Text(t(Str.OrderQueryCny, emptyArray()), color = ThemeBlue, fontSize = 15.sp)
                        }
                        Text(
                            t(Str.OrderQueryRechargeAmount, emptyArray()),
                            color = SubGray,
                            fontSize = 16.sp,
                            modifier = Modifier.padding(top = 4.dp),
                        )
                    }
                    Row(
                        Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 16.dp)
                            .background(Color(0x123366CD), RoundedCornerShape(8.dp))
                            .padding(16.dp),
                    ) {
                        Column(Modifier.weight(1f)) {
                            Row(verticalAlignment = Alignment.Bottom) {
                                Text(OrderFormat.yuan(wallet?.presentFen), color = ThemeBlue, fontSize = 26.sp)
                                Spacer(Modifier.width(2.dp))
                                Text(t(Str.OrderQueryCny, emptyArray()), color = ThemeBlue, fontSize = 15.sp)
                            }
                            Text(
                                t(Str.OrderQueryPresentAmount, emptyArray()),
                                color = SubGray,
                                fontSize = 16.sp,
                                modifier = Modifier.padding(top = 4.dp),
                            )
                        }
                        Column(horizontalAlignment = Alignment.End) {
                            OutlineAction(t(Str.OrderQueryDebit, emptyArray())) {
                                dialog = WalletDialog.Debit
                            }
                            Spacer(Modifier.height(12.dp))
                            FillAction(t(Str.OrderQueryRecharge, emptyArray())) {
                                dialog = WalletDialog.Recharge
                            }
                        }
                    }
                }
            }
            items(records, key = { it.merchantTradeNo + it.paidAt }) { row ->
                WalletRecordBlock(t, row, onRefund = {
                    scope.launch {
                        val max = feature.refundableFen(row.merchantTradeNo, row.paidAt) ?: return@launch
                        dialog = WalletDialog.Refund(row, max)
                    }
                })
            }
            if (loading) {
                item {
                    Box(Modifier.fillMaxWidth().padding(16.dp), contentAlignment = Alignment.Center) {
                        CircularProgressIndicator(color = colors.primary)
                    }
                }
            }
        }
    }

    when (val d = dialog) {
        null -> Unit
        WalletDialog.Recharge -> AmountDialog(
            title = t(Str.OrderQueryPleaseEnterAmount, emptyArray()),
            hint = "",
            onDismiss = { dialog = null },
            onConfirm = { yuan ->
                scope.launch {
                    if (feature.editWalletPresentYuan(yuan)) dialog = null
                }
            },
        )
        WalletDialog.Debit -> AmountDialog(
            title = t(Str.OrderQueryPleaseEnterAmount, emptyArray()),
            hint = t(Str.OrderQueryDebitHint, arrayOf(OrderFormat.yuan(wallet?.presentFen))),
            onDismiss = { dialog = null },
            onConfirm = { yuan ->
                scope.launch {
                    if (feature.editWalletPresentYuan(-yuan)) dialog = null
                }
            },
        )
        is WalletDialog.Refund -> AmountDialog(
            title = t(Str.OrderQueryPleaseEnterAmount, emptyArray()),
            hint = t(Str.OrderQueryRefundHint, arrayOf(OrderFormat.yuan(d.maxFen.toLong()))) +
                t(
                    Str.OrderQueryWalletDetailsHint,
                    arrayOf(
                        OrderFormat.yuan(wallet?.balanceFen),
                        OrderFormat.yuan(wallet?.rechargeFen),
                        OrderFormat.yuan(wallet?.presentFen),
                    ),
                ),
            onDismiss = { dialog = null },
            onConfirm = { yuan ->
                scope.launch {
                    if (feature.refundWallet(d.record, yuan)) dialog = null
                }
            },
        )
    }
}

private sealed class WalletDialog {
    data object Recharge : WalletDialog()
    data object Debit : WalletDialog()
    data class Refund(val record: OrderWalletRecord, val maxFen: Int) : WalletDialog()
}

@Composable
internal fun UserDepositPage(
    t: AssetTr,
    feature: OrderQueryFeature,
    user: OrderUserDetail?,
    assets: OrderUserAssets?,
    records: List<OrderDepositRecord>,
    loading: Boolean,
    onBack: () -> Unit,
) {
    val colors = OpsTheme.colors
    val scope = rememberCoroutineScope()
    val paid = OrderFormat.paidDeposit(user)
    LaunchedEffect(Unit) { feature.loadDeposit() }

    Column(Modifier.fillMaxSize().background(Color.White)) {
        AssetTopBar(t(Str.OrderQueryDepositMembership, emptyArray()), onBack)
        LazyColumn(Modifier.weight(1f)) {
            item {
                Column(Modifier.fillMaxWidth()) {
                    Box(
                        Modifier
                            .fillMaxWidth()
                            .height(48.dp)
                            .background(TipOrangeBg)
                            .padding(start = 16.dp),
                        contentAlignment = Alignment.CenterStart,
                    ) {
                        Text(t(Str.OrderQueryConfirmRefundMerchant, emptyArray()), color = TipOrange, fontSize = 14.sp)
                    }
                    Row(
                        Modifier.fillMaxWidth().padding(horizontal = 16.dp).padding(top = 20.dp),
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Text(t(Str.OrderQueryPaidAmountLabel, emptyArray()), color = TextPrimary, fontSize = 14.sp)
                        Text(
                            OrderFormat.integrityAmount(user, assets),
                            color = TextPrimary,
                            fontSize = 14.sp,
                            fontWeight = FontWeight.Bold,
                        )
                        Spacer(Modifier.weight(1f))
                        StatusChip(
                            text = t(if (paid) Str.OrderQueryPaid else Str.OrderQueryUnpaid, emptyArray()),
                            paid = paid,
                        )
                        if (user?.izRidingType == 4) {
                            Spacer(Modifier.width(8.dp))
                            StatusChip(t(Str.OrderQueryWechatNoDeposit, emptyArray()), paid = true)
                        }
                    }
                    Row(Modifier.padding(horizontal = 16.dp).padding(top = 20.dp)) {
                        Text(t(Str.OrderQueryExpireTime, emptyArray()), color = TextPrimary, fontSize = 14.sp)
                        Text(
                            OrderFormat.depositExpireText(user, assets),
                            color = TextPrimary,
                            fontSize = 14.sp,
                            fontWeight = FontWeight.Bold,
                        )
                    }
                    Row(
                        Modifier.fillMaxWidth().padding(top = 20.dp),
                        horizontalArrangement = Arrangement.Center,
                    ) {
                        OutlineWide(t(Str.OrderQueryResetRideState, emptyArray())) {
                            scope.launch { feature.resetCarStatus() }
                        }
                        Spacer(Modifier.width(15.dp))
                        OutlineWide(t(Str.OrderQueryManualReturn, emptyArray())) {
                            scope.launch { feature.manualReturnDeposit() }
                        }
                    }
                    Box(Modifier.fillMaxWidth().height(8.dp).padding(top = 20.dp).background(SectionGap))
                    Text(
                        t(Str.OrderQueryPurchaseRecord, emptyArray()),
                        color = TextPrimary,
                        fontSize = 16.sp,
                        fontWeight = FontWeight.Bold,
                        modifier = Modifier.padding(start = 16.dp, top = 20.dp, bottom = 8.dp),
                    )
                }
            }
            items(records, key = { it.merchantTradeNo + it.paidAt }) { row ->
                DepositRecordBlock(t, row)
            }
            if (loading) {
                item {
                    Box(Modifier.fillMaxWidth().padding(16.dp), contentAlignment = Alignment.Center) {
                        CircularProgressIndicator(color = colors.primary)
                    }
                }
            }
        }
    }
}

@Composable
internal fun UserRideCardPage(
    t: AssetTr,
    feature: OrderQueryFeature,
    cards: List<OrderRideCard>,
    records: List<OrderRideCardRecord>,
    loading: Boolean,
    onBack: () -> Unit,
) {
    val colors = OpsTheme.colors
    val scope = rememberCoroutineScope()
    var tab by remember { mutableStateOf(0) }
    var refundTarget by remember { mutableStateOf<Pair<OrderRideCardRecord, Int>?>(null) }
    LaunchedEffect(tab) {
        if (tab == 0) feature.loadRideCards() else feature.loadRideRecords()
    }

    Column(Modifier.fillMaxSize().background(Color.White)) {
        AssetTopBar(t(Str.OrderQueryRideCard, emptyArray()), onBack)
        RideSegment(
            t = t,
            selected = tab,
            onSelect = { tab = it },
        )
        Box(Modifier.fillMaxWidth().height(8.dp).background(SectionGap))
        LazyColumn(Modifier.weight(1f)) {
            if (tab == 0) {
                items(cards, key = { it.name + it.cardExpiredDate }) { card ->
                    RideCardBlock(t, card)
                }
            } else {
                items(records, key = { it.merchantTradeNo + it.paidAt }) { row ->
                    RideRecordBlock(t, row, onRefund = {
                        scope.launch {
                            val max = feature.refundableFen(row.merchantTradeNo, row.paidAt) ?: return@launch
                            refundTarget = row to max
                        }
                    })
                }
            }
            if (loading) {
                item {
                    Box(Modifier.fillMaxWidth().padding(16.dp), contentAlignment = Alignment.Center) {
                        CircularProgressIndicator(color = colors.primary)
                    }
                }
            }
        }
    }

    refundTarget?.let { (row, max) ->
        AmountDialog(
            title = t(Str.OrderQueryPleaseEnterAmount, emptyArray()),
            hint = t(Str.OrderQueryRefundHint, arrayOf(OrderFormat.yuan(max.toLong()))),
            onDismiss = { refundTarget = null },
            onConfirm = { yuan ->
                scope.launch {
                    if (feature.refundRide(row, yuan)) refundTarget = null
                }
            },
        )
    }
}

@Composable
private fun RideSegment(t: AssetTr, selected: Int, onSelect: (Int) -> Unit) {
    Row(
        Modifier
            .fillMaxWidth()
            .padding(horizontal = 52.dp, vertical = 12.dp)
            .border(1.dp, ThemeBlue, RoundedCornerShape(4.dp)),
    ) {
        listOf(
            t(Str.OrderQueryRideAvailable, emptyArray()),
            t(Str.OrderQueryPurchaseRecord, emptyArray()),
        ).forEachIndexed { index, label ->
            Box(
                Modifier
                    .weight(1f)
                    .background(if (selected == index) ThemeBlue else Color.White)
                    .clickable { onSelect(index) }
                    .padding(vertical = 8.dp),
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    label,
                    color = if (selected == index) Color.White else ThemeBlue,
                    fontSize = 14.sp,
                )
            }
        }
    }
}

@Composable
private fun WalletRecordBlock(t: AssetTr, row: OrderWalletRecord, onRefund: () -> Unit) {
    Column(Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 16.dp)) {
        PairRow(t(Str.OrderQueryTradeNo, emptyArray()), OrderFormat.orDash(row.merchantTradeNo), t(Str.OrderQueryTradeAmount, emptyArray()), OrderFormat.yuanWithUnit(row.amountFen))
        PairRow(t(Str.OrderQueryTradeTime, emptyArray()), OrderFormat.orDash(row.paidAt), t(Str.OrderQueryRechargeAmount, emptyArray()) + ":", OrderFormat.yuanWithUnit(row.rechargeAmountFen))
        PairRow(t(Str.OrderQueryPayChannel, emptyArray()).replace("支付", "交易"), OrderFormat.orDash(row.channel), t(Str.OrderQueryPresentAmount, emptyArray()) + ":", OrderFormat.yuanWithUnit(row.presentAmountFen))
        PairRow(t(Str.OrderQueryTradeCategory, emptyArray()), OrderFormat.orDash(row.type), "", "")
        Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
            Text(t(Str.OrderQueryTradeType, emptyArray()), color = TextPrimary, fontSize = 14.sp)
            Text(OrderFormat.orDash(row.changeType), color = TextPrimary, fontSize = 14.sp)
            Spacer(Modifier.weight(1f))
            if (row.izRefund == 0) {
                OutlineAction(t(Str.OrderQueryRefund, emptyArray()), onRefund)
            }
        }
    }
    HorizontalDivider(color = Color(0xFFD9DCE6), thickness = 1.dp)
}

@Composable
private fun DepositRecordBlock(t: AssetTr, row: OrderDepositRecord) {
    val refunded = row.state.isNotBlank() && row.state != "已支付"
    Column(Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 16.dp)) {
        Text(OrderFormat.orDash(row.state), color = if (refunded) ThemeBlue else PaidGreen, fontSize = 14.sp, fontWeight = FontWeight.Bold)
        Spacer(Modifier.height(8.dp))
        PairRow(t(Str.OrderQueryTradeNo, emptyArray()), OrderFormat.orDash(row.merchantTradeNo), "", "")
        PairRow(t(Str.OrderQueryTradeTime, emptyArray()), OrderFormat.orDash(row.paidAt), t(Str.OrderQueryMemberType, emptyArray()), OrderFormat.orDash(row.depositType))
        PairRow(t(Str.OrderQueryPayChannel, emptyArray()), OrderFormat.orDash(row.channel), t(Str.OrderQueryTradeAmount, emptyArray()), OrderFormat.yuanWithUnit(row.amountFen))
        PairRow(t(Str.OrderQueryTradeCategory, emptyArray()), OrderFormat.orDash(row.type), t(Str.OrderQueryExpireTime, emptyArray()), OrderFormat.durationDaysOrDash(row.duration))
    }
    HorizontalDivider(color = DividerLine, thickness = 1.dp)
}

@Composable
private fun RideCardBlock(t: AssetTr, card: OrderRideCard) {
    Column(Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 12.dp)) {
        PairRow(
            t(Str.OrderQueryRideCardName, emptyArray()),
            OrderFormat.orDash(card.name),
            t(Str.OrderQueryRemainDuration, emptyArray()),
            OrderFormat.rideRemainingTime(card),
            labelColor = LabelGray,
        )
        PairRow(
            t(Str.OrderQueryExpireTime, emptyArray()),
            OrderFormat.orDash(card.cardExpiredDate),
            t(Str.OrderQueryRemainTimes, emptyArray()),
            OrderFormat.rideRemainTimes(card),
            labelColor = LabelGray,
        )
    }
    HorizontalDivider(color = DividerLine, thickness = 1.dp)
}

@Composable
private fun RideRecordBlock(t: AssetTr, row: OrderRideCardRecord, onRefund: () -> Unit) {
    Column(Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 12.dp)) {
        PairRow(t(Str.OrderQueryTradeNo, emptyArray()), OrderFormat.orDash(row.merchantTradeNo), "", "", labelColor = LabelGray)
        PairRow(
            t(Str.OrderQueryRideCardName, emptyArray()),
            OrderFormat.orDash(row.name),
            t(Str.OrderQueryRideCardAmount, emptyArray()),
            OrderFormat.yuanWithUnit(row.amountFen),
            labelColor = LabelGray,
        )
        PairRow(
            t(Str.OrderQueryTradeTime, emptyArray()),
            OrderFormat.orDash(row.paidAt),
            t(Str.OrderQueryDurationLabel, emptyArray()),
            OrderFormat.durationDaysOrDash(row.duration),
            labelColor = LabelGray,
        )
        PairRow(t(Str.OrderQueryPayChannel, emptyArray()), OrderFormat.orDash(row.channel), "", "", labelColor = LabelGray)
        Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
            Text(t(Str.OrderQueryTradeCategory, emptyArray()), color = LabelGray, fontSize = 14.sp)
            Text(OrderFormat.orDash(row.type), color = TextPrimary, fontSize = 14.sp)
            Spacer(Modifier.weight(1f))
            if (row.izRefund != 1) {
                OutlineAction(t(Str.OrderQueryRefund, emptyArray()), onRefund)
            }
        }
    }
    HorizontalDivider(color = DividerLine, thickness = 1.dp)
}

@Composable
private fun PairRow(
    leftLabel: String,
    leftValue: String,
    rightLabel: String,
    rightValue: String,
    labelColor: Color = TextPrimary,
) {
    Row(Modifier.fillMaxWidth().padding(vertical = 4.dp)) {
        Column(Modifier.weight(1f)) {
            Row {
                Text(leftLabel, color = labelColor, fontSize = 14.sp)
                Text(leftValue, color = TextPrimary, fontSize = 14.sp)
            }
        }
        if (rightLabel.isNotEmpty()) {
            Column(Modifier.weight(1f), horizontalAlignment = Alignment.End) {
                Row {
                    Text(rightLabel, color = labelColor, fontSize = 14.sp)
                    Text(rightValue, color = TextPrimary, fontSize = 14.sp)
                }
            }
        }
    }
}

@Composable
private fun StatusChip(text: String, paid: Boolean) {
    Text(
        text,
        color = if (paid) PaidGreen else UnpaidRed,
        fontSize = 12.sp,
        modifier = Modifier
            .background(
                if (paid) Color(0x1A1DBA4F) else Color(0x1AFF2222),
                RoundedCornerShape(4.dp),
            )
            .padding(horizontal = 16.dp, vertical = 3.dp),
    )
}

@Composable
private fun OutlineAction(text: String, onClick: () -> Unit) {
    Box(
        Modifier
            .width(50.dp)
            .height(24.dp)
            .border(1.dp, ThemeBlue, RoundedCornerShape(12.dp))
            .clickable(onClick = onClick),
        contentAlignment = Alignment.Center,
    ) {
        Text(text, color = ThemeBlue, fontSize = 14.sp)
    }
}

@Composable
private fun FillAction(text: String, onClick: () -> Unit) {
    Box(
        Modifier
            .width(50.dp)
            .height(24.dp)
            .background(ThemeBlue, RoundedCornerShape(12.dp))
            .clickable(onClick = onClick),
        contentAlignment = Alignment.Center,
    ) {
        Text(text, color = Color.White, fontSize = 14.sp)
    }
}

@Composable
private fun OutlineWide(text: String, onClick: () -> Unit) {
    Box(
        Modifier
            .width(164.dp)
            .height(48.dp)
            .border(1.dp, Color(0xFF999999), RoundedCornerShape(4.dp))
            .clickable(onClick = onClick),
        contentAlignment = Alignment.Center,
    ) {
        Text(text, color = TextPrimary, fontSize = 16.sp, fontWeight = FontWeight.Bold)
    }
}

@Composable
private fun AmountDialog(
    title: String,
    hint: String,
    onDismiss: () -> Unit,
    onConfirm: (Double) -> Unit,
) {
    var text by remember { mutableStateOf("") }
    Dialog(onDismissRequest = onDismiss) {
        Column(
            Modifier
                .fillMaxWidth()
                .background(Color.White, RoundedCornerShape(8.dp))
                .padding(16.dp),
        ) {
            Text(title, color = TextPrimary, fontSize = 16.sp, fontWeight = FontWeight.Bold)
            if (hint.isNotBlank()) {
                Text(hint, color = Color(0xFF666666), fontSize = 13.sp, modifier = Modifier.padding(top = 8.dp))
            }
            TextField(
                value = text,
                onValueChange = { text = it.filter { ch -> ch.isDigit() || ch == '.' } },
                singleLine = true,
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Decimal),
                modifier = Modifier.fillMaxWidth().padding(top = 8.dp),
            )
            Row(Modifier.fillMaxWidth().padding(top = 12.dp), horizontalArrangement = Arrangement.End) {
                Text("取消", color = Color(0xFF999999), modifier = Modifier.clickable(onClick = onDismiss).padding(8.dp))
                Spacer(Modifier.width(16.dp))
                Text(
                    "确定",
                    color = ThemeBlue,
                    modifier = Modifier.clickable {
                        text.toDoubleOrNull()?.let(onConfirm)
                    }.padding(8.dp),
                )
            }
        }
    }
}

@Composable
private fun AssetTopBar(title: String, onBack: () -> Unit) {
    val colors = OpsTheme.colors
    com.luopingtech.ebike.ops.ui.navigation.OpsBackHandler(onBack = onBack)
    Box(
        Modifier
            .fillMaxWidth()
            .background(colors.primary)
            .statusBarsPadding()
            .padding(horizontal = 8.dp, vertical = 10.dp),
    ) {
        OpsBackChevron(onClick = onBack, modifier = Modifier.align(Alignment.CenterStart))
        Text(
            title,
            color = colors.onPrimary,
            fontSize = 18.sp,
            fontWeight = FontWeight.Medium,
            modifier = Modifier.align(Alignment.Center),
        )
    }
}
