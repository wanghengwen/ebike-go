package com.luopingtech.ebike.ops.ui.admin

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
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.domain.admin.ObjectionOrderDetail
import com.luopingtech.ebike.ops.domain.admin.ObjectionOrderItem
import com.luopingtech.ebike.ops.domain.admin.ObjectionStates
import com.luopingtech.ebike.ops.domain.admin.favorableTypeLabel
import com.luopingtech.ebike.ops.domain.admin.handleWayLabel
import com.luopingtech.ebike.ops.domain.admin.isPaid
import com.luopingtech.ebike.ops.domain.admin.isProcessed
import com.luopingtech.ebike.ops.domain.admin.opTypeLabel
import com.luopingtech.ebike.ops.domain.admin.processAmounts
import com.luopingtech.ebike.ops.domain.admin.stateLabel
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsTimeRanges
import com.luopingtech.ebike.ops.domain.order.OrderFormat
import com.luopingtech.ebike.ops.feature.admin.ObjectionNav
import com.luopingtech.ebike.ops.ui.analysis.VcdTopBar
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.launch

private val PageBg = Color(0xFFF6F7F9)
private val TextPrimary = Color(0xFF242936)
private val TextMuted = Color(0xFF7C87B1)
private val PendingOrange = Color(0xFFFFAD00)
private val ProcessedGreen = Color(0xFF1DBA4F)
private val UnpaidRed = Color(0xFFFF0808)
private val TipOrange = Color(0xFFFA6400)
private val LinkBlue = Color(0xFF1180F9)

private enum class ObjectionDatePickTarget { Start, End }

@Composable
fun ObjectionOrderScreen(app: OpsApp, onClose: () -> Unit) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val home by app.homeFeature.state.collectAsState()
    val state by app.objectionOrderFeature.state.collectAsState()
    val feature = app.objectionOrderFeature
    val scope = rememberCoroutineScope()
    val colors = OpsTheme.colors

    LaunchedEffect(home.currentArea?.id) {
        feature.load(home.currentArea)
    }

    when (val nav = state.nav) {
        ObjectionNav.List -> ObjectionListPage(
            t = { key, args -> t(key, *args) },
            state = state,
            primary = colors.primary,
            onClose = {
                feature.clear()
                onClose()
            },
            onKeyword = feature::setKeyword,
            onStateFilter = { filter ->
                feature.setStateFilter(filter)
                scope.launch { feature.search(home.currentArea) }
            },
            onDatesChange = { start, end ->
                if (feature.setCreatedDates(start, end)) {
                    scope.launch { feature.search(home.currentArea) }
                }
            },
            onSearch = { scope.launch { feature.search(home.currentArea) } },
            onLoadMore = { scope.launch { feature.loadMore(home.currentArea) } },
            onOpen = { item -> scope.launch { feature.openDetail(home.currentArea, item) } },
        )
        ObjectionNav.Detail -> {
            val detail = state.detail
            if (detail == null) {
                feature.navigateBack()
            } else {
                ObjectionDetailPage(
                    t = { key, args -> t(key, *args) },
                    detail = detail,
                    loading = state.loading,
                    error = state.errorMessage,
                    primary = colors.primary,
                    onBack = { feature.navigateBack() },
                    onAdopt = { feature.openProcess(feeReasonable = false) },
                    onReject = { feature.openProcess(feeReasonable = true) },
                )
            }
        }
        is ObjectionNav.Process -> {
            val detail = state.detail
            if (detail == null) {
                feature.navigateBack()
            } else {
                LaunchedEffect(detail.ticket.id, nav.feeReasonable) {
                    feature.loadSendMode()
                }
                ObjectionProcessPage(
                    t = { key, args -> t(key, *args) },
                    detail = detail,
                    feeReasonable = nav.feeReasonable,
                    state = state,
                    primary = colors.primary,
                    canSubmit = feature.canSubmitProcess(),
                    onBack = { feature.navigateBack() },
                    onReason = feature::setProcessReason,
                    onModifyPay = feature::setModifyPayYuan,
                    onModifyDispatch = feature::setModifyDispatchYuan,
                    onModifyHelmet = feature::setModifyHelmetYuan,
                    onRefundPay = feature::setRefundPayYuan,
                    onRefundDispatch = feature::setRefundDispatchYuan,
                    onRefundHelmet = feature::setRefundHelmetYuan,
                    onRefundCard = feature::setRefundCardTimes,
                    onToggleSys = feature::toggleNoticeSys,
                    onToggleSms = feature::toggleNoticeSms,
                    onToggleApp = feature::toggleNoticeApp,
                    onSubmit = { scope.launch { feature.submitDeal(home.currentArea) } },
                )
            }
        }
    }
}

@Composable
private fun ObjectionListPage(
    t: (Str, Array<out Any?>) -> String,
    state: com.luopingtech.ebike.ops.feature.admin.ObjectionOrderUiState,
    primary: Color,
    onClose: () -> Unit,
    onKeyword: (String) -> Unit,
    onStateFilter: (Int?) -> Unit,
    onDatesChange: (startDate: String, endDate: String) -> Unit,
    onSearch: () -> Unit,
    onLoadMore: () -> Unit,
    onOpen: (ObjectionOrderItem) -> Unit,
) {
    val listState = rememberLazyListState()
    var datePickTarget by remember { mutableStateOf<ObjectionDatePickTarget?>(null) }
    val startDate = state.createdTimeStart.take(10).ifBlank { OfflineOpsTimeRanges.formatDate(OfflineOpsTimeRanges.objectionDefaultRange().startMs) }
    val endDate = state.createdTimeEnd.take(10).ifBlank { OfflineOpsTimeRanges.formatDate(OfflineOpsTimeRanges.objectionDefaultRange().endMs) }

    LaunchedEffect(listState, state.loadingMore, state.finished, state.loading) {
        snapshotFlow {
            val info = listState.layoutInfo
            val last = info.visibleItemsInfo.lastOrNull()?.index ?: 0
            last >= info.totalItemsCount - 2 && info.totalItemsCount > 0
        }.distinctUntilChanged().collect { nearEnd ->
            if (nearEnd && !state.loading && !state.loadingMore && !state.finished) onLoadMore()
        }
    }

    Column(
        Modifier
            .fillMaxSize()
            .background(PageBg),
    ) {
        VcdTopBar(title = t(Str.ObjectionOrder, emptyArray()), onBack = onClose)
        Column(
            Modifier
                .fillMaxWidth()
                .background(primary)
                .padding(horizontal = 16.dp, vertical = 0.dp),
        ) {
            Row(
                Modifier
                    .fillMaxWidth()
                    .padding(bottom = 12.dp)
                    .background(Color.White, RoundedCornerShape(4.dp))
                    .padding(horizontal = 12.dp, vertical = 10.dp),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.Center,
            ) {
                Text(
                    startDate,
                    color = LinkBlue,
                    fontSize = 16.sp,
                    modifier = Modifier.clickable { datePickTarget = ObjectionDatePickTarget.Start },
                )
                Text(
                    t(Str.ObjectionDateTo, emptyArray()),
                    color = TextPrimary,
                    fontSize = 16.sp,
                    modifier = Modifier.padding(horizontal = 8.dp),
                )
                Text(
                    endDate,
                    color = LinkBlue,
                    fontSize = 16.sp,
                    modifier = Modifier.clickable { datePickTarget = ObjectionDatePickTarget.End },
                )
            }
        }
        Column(
            Modifier
                .fillMaxWidth()
                .background(Color.White)
                .padding(horizontal = 16.dp, vertical = 12.dp),
            verticalArrangement = Arrangement.spacedBy(10.dp),
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                BasicTextField(
                    value = state.keyword,
                    onValueChange = onKeyword,
                    singleLine = true,
                    textStyle = TextStyle(color = TextPrimary, fontSize = 14.sp),
                    cursorBrush = SolidColor(primary),
                    modifier = Modifier
                        .weight(1f)
                        .height(40.dp)
                        .border(1.dp, Color(0xFFD9DCE6), RoundedCornerShape(6.dp))
                        .padding(horizontal = 12.dp, vertical = 10.dp),
                    decorationBox = { inner ->
                        if (state.keyword.isEmpty()) {
                            Text(t(Str.OrderQueryInputHint, emptyArray()), color = TextMuted, fontSize = 14.sp)
                        }
                        inner()
                    },
                )
                Spacer(Modifier.width(8.dp))
                Button(onClick = onSearch) { Text(t(Str.OrderQuerySearchAction, emptyArray())) }
            }
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                StateChip(
                    label = t(Str.ObjectionStateAll, emptyArray()),
                    selected = state.stateFilter == null,
                    primary = primary,
                    onClick = { onStateFilter(null) },
                )
                StateChip(
                    label = t(Str.ObjectionPending, emptyArray()),
                    selected = state.stateFilter == ObjectionStates.Pending,
                    primary = primary,
                    onClick = { onStateFilter(ObjectionStates.Pending) },
                )
                StateChip(
                    label = t(Str.ObjectionProcessed, emptyArray()),
                    selected = state.stateFilter == ObjectionStates.Processed,
                    primary = primary,
                    onClick = { onStateFilter(ObjectionStates.Processed) },
                )
            }
        }
        state.errorMessage?.let {
            Text(it, color = Color(0xFFE02020), modifier = Modifier.padding(16.dp))
        }
        state.message?.let {
            Text(it, color = primary, modifier = Modifier.padding(16.dp))
        }
        when {
            state.loading && state.items.isEmpty() -> Box(
                Modifier.fillMaxSize(),
                contentAlignment = Alignment.Center,
            ) { CircularProgressIndicator(color = primary) }
            state.items.isEmpty() -> Text(
                t(Str.AdminEmptyList, emptyArray()),
                modifier = Modifier.padding(40.dp),
                color = TextMuted,
            )
            else -> LazyColumn(state = listState, modifier = Modifier.fillMaxSize()) {
                items(state.items, key = { it.id }) { item ->
                    ObjectionListItem(
                        item = item,
                        t = t,
                        primary = primary,
                        onClick = { onOpen(item) },
                    )
                }
                item {
                    when {
                        state.loadingMore -> Box(
                            Modifier.fillMaxWidth().padding(16.dp),
                            contentAlignment = Alignment.Center,
                        ) { CircularProgressIndicator(color = primary) }
                        state.finished -> Text(
                            t(Str.VehicleListEnd, emptyArray()),
                            modifier = Modifier.padding(16.dp).fillMaxWidth(),
                            color = TextMuted,
                            fontSize = 12.sp,
                        )
                    }
                }
            }
        }
    }

    datePickTarget?.let { target ->
        val initial = if (target == ObjectionDatePickTarget.Start) startDate else endDate
        ObjectionDayPickerDialog(
            initialDate = initial,
            primary = primary,
            confirmLabel = t(Str.Confirm, emptyArray()),
            cancelLabel = t(Str.Cancel, emptyArray()),
            onDismiss = { datePickTarget = null },
            onConfirm = { picked ->
                datePickTarget = null
                if (target == ObjectionDatePickTarget.Start) {
                    onDatesChange(picked, endDate)
                } else {
                    onDatesChange(startDate, picked)
                }
            },
        )
    }
}

@Composable
private fun ObjectionDayPickerDialog(
    initialDate: String,
    primary: Color,
    confirmLabel: String,
    cancelLabel: String,
    onDismiss: () -> Unit,
    onConfirm: (String) -> Unit,
) {
    val initialParts = remember(initialDate) {
        OfflineOpsTimeRanges.parseDateStart(initialDate)?.let { OfflineOpsTimeRanges.ymdParts(it) }
            ?: OfflineOpsTimeRanges.ymdParts(OfflineOpsTimeRanges.objectionDefaultRange().endMs)
    }
    var year by remember { mutableStateOf(initialParts[0]) }
    var month by remember { mutableStateOf(initialParts[1]) }
    var selectedDay by remember { mutableStateOf(initialParts[2]) }

    fun shiftMonth(delta: Int) {
        var y = year
        var m = month + delta
        while (m < 1) {
            m += 12
            y -= 1
        }
        while (m > 12) {
            m -= 12
            y += 1
        }
        year = y
        month = m
        val maxDay = OfflineOpsTimeRanges.daysInMonth(y, m)
        if (selectedDay > maxDay) selectedDay = maxDay
    }

    val firstDow = OfflineOpsTimeRanges.weekdayMondayIndex(
        OfflineOpsTimeRanges.dateStartOf(year, month, 1),
    )
    val daysInMonth = OfflineOpsTimeRanges.daysInMonth(year, month)
    val cells = remember(year, month) {
        buildList {
            repeat(firstDow) { add(null) }
            for (d in 1..daysInMonth) add(d)
            while (size % 7 != 0) add(null)
        }
    }

    Dialog(onDismissRequest = onDismiss) {
        Column(
            Modifier
                .fillMaxWidth()
                .background(Color.White, RoundedCornerShape(12.dp))
                .padding(16.dp),
        ) {
            Row(
                Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween,
            ) {
                Text(
                    "‹",
                    color = primary,
                    fontSize = 24.sp,
                    modifier = Modifier
                        .clickable { shiftMonth(-1) }
                        .padding(8.dp),
                )
                Text(
                    "$year-${month.toString().padStart(2, '0')}",
                    color = TextPrimary,
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Medium,
                )
                Text(
                    "›",
                    color = primary,
                    fontSize = 24.sp,
                    modifier = Modifier
                        .clickable { shiftMonth(1) }
                        .padding(8.dp),
                )
            }
            Spacer(Modifier.height(8.dp))
            Row(Modifier.fillMaxWidth()) {
                listOf("一", "二", "三", "四", "五", "六", "日").forEach { label ->
                    Text(
                        label,
                        color = TextMuted,
                        fontSize = 12.sp,
                        textAlign = TextAlign.Center,
                        modifier = Modifier.weight(1f),
                    )
                }
            }
            Spacer(Modifier.height(4.dp))
            cells.chunked(7).forEach { week ->
                Row(Modifier.fillMaxWidth()) {
                    week.forEach { day ->
                        Box(
                            Modifier
                                .weight(1f)
                                .height(40.dp),
                            contentAlignment = Alignment.Center,
                        ) {
                            if (day != null) {
                                val selected = day == selectedDay
                                Box(
                                    Modifier
                                        .size(32.dp)
                                        .background(
                                            if (selected) primary else Color.Transparent,
                                            CircleShape,
                                        )
                                        .clickable { selectedDay = day },
                                    contentAlignment = Alignment.Center,
                                ) {
                                    Text(
                                        day.toString(),
                                        color = if (selected) Color.White else TextPrimary,
                                        fontSize = 14.sp,
                                        textAlign = TextAlign.Center,
                                    )
                                }
                            }
                        }
                    }
                }
            }
            Spacer(Modifier.height(12.dp))
            Row(
                Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.End,
            ) {
                TextButton(onClick = onDismiss) { Text(cancelLabel) }
                TextButton(
                    onClick = {
                        val date = buildString {
                            append(year.toString().padStart(4, '0'))
                            append('-')
                            append(month.toString().padStart(2, '0'))
                            append('-')
                            append(selectedDay.toString().padStart(2, '0'))
                        }
                        onConfirm(date)
                    },
                ) { Text(confirmLabel, color = primary) }
            }
        }
    }
}

@Composable
private fun StateChip(
    label: String,
    selected: Boolean,
    primary: Color,
    onClick: () -> Unit,
) {
    val bg = if (selected) primary.copy(alpha = 0.12f) else Color(0xFFF4F6FF)
    val fg = if (selected) primary else TextPrimary
    Text(
        label,
        color = fg,
        fontSize = 13.sp,
        modifier = Modifier
            .background(bg, RoundedCornerShape(14.dp))
            .clickable(onClick = onClick)
            .padding(horizontal = 12.dp, vertical = 6.dp),
    )
}

@Composable
private fun ObjectionListItem(
    item: ObjectionOrderItem,
    t: (Str, Array<out Any?>) -> String,
    primary: Color,
    onClick: () -> Unit,
) {
    val processed = item.isProcessed()
    Column(
        Modifier
            .fillMaxWidth()
            .padding(horizontal = 12.dp, vertical = 6.dp)
            .background(Color.White, RoundedCornerShape(8.dp))
            .clickable(onClick = onClick)
            .padding(16.dp),
    ) {
        Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
            Text(t(Str.OrderQueryNameLabel, emptyArray()), color = TextPrimary, fontSize = 14.sp)
            Text(
                item.userName.ifBlank { "--" },
                color = TextPrimary,
                fontSize = 14.sp,
                fontWeight = FontWeight.Bold,
                modifier = Modifier.weight(1f).padding(start = 4.dp),
            )
            Text(
                item.stateLabel(),
                color = if (processed) ProcessedGreen else PendingOrange,
                fontSize = 16.sp,
                fontWeight = FontWeight.Bold,
            )
        }
        Spacer(Modifier.height(8.dp))
        Row {
            Text(t(Str.PhoneColonLabel, emptyArray()), color = TextPrimary, fontSize = 14.sp)
            Text(item.phone.ifBlank { "--" }, color = TextPrimary, fontSize = 14.sp, fontWeight = FontWeight.Bold)
        }
        Spacer(Modifier.height(8.dp))
        Row(verticalAlignment = Alignment.CenterVertically) {
            Text(t(Str.ObjectionCreatedAt, emptyArray()), color = TextPrimary, fontSize = 14.sp)
            Text(
                item.createdAt.ifBlank { "--" },
                color = TextPrimary,
                fontSize = 14.sp,
                modifier = Modifier.weight(1f),
            )
            Box(
                Modifier
                    .background(
                        if (processed) Color.White else primary,
                        RoundedCornerShape(5.dp),
                    )
                    .then(
                        if (processed) {
                            Modifier.border(1.dp, Color(0xFF999999), RoundedCornerShape(5.dp))
                        } else {
                            Modifier
                        },
                    )
                    .padding(horizontal = 12.dp, vertical = 6.dp),
            ) {
                Text(
                    if (processed) t(Str.ObjectionLookDetail, emptyArray())
                    else t(Str.ObjectionToProcess, emptyArray()),
                    color = if (processed) TextPrimary else Color.White,
                    fontSize = 14.sp,
                )
            }
        }
    }
}

@Composable
private fun ObjectionDetailPage(
    t: (Str, Array<out Any?>) -> String,
    detail: com.luopingtech.ebike.ops.domain.admin.ObjectionOrderDetail,
    loading: Boolean,
    error: String?,
    primary: Color,
    onBack: () -> Unit,
    onAdopt: () -> Unit,
    onReject: () -> Unit,
) {
    val ticket = detail.ticket
    val order = detail.order
    Column(Modifier.fillMaxSize().background(PageBg)) {
        VcdTopBar(title = t(Str.ObjectionOrderDetail, emptyArray()), onBack = onBack)
        if (loading) {
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                CircularProgressIndicator(color = primary)
            }
            return
        }
        Column(
            Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState())
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            error?.let { Text(it, color = Color(0xFFE02020)) }
            SectionCard(title = t(Str.ObjectionTicketSection, emptyArray())) {
                Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                    Text(t(Str.AuditStateLabel, emptyArray()), color = TextMuted, fontSize = 14.sp)
                    Text(
                        ticket.stateLabel(),
                        color = if (ticket.isProcessed()) ProcessedGreen else PendingOrange,
                        fontWeight = FontWeight.Bold,
                    )
                }
                InfoRow(t(Str.ObjectionCreatedAt, emptyArray()), ticket.createdAt.ifBlank { "--" })
                InfoRow(t(Str.ObjectionTicketId, emptyArray()), ticket.id)
                InfoRow(t(Str.ObjectionUserReason, emptyArray()), ticket.userReason.ifBlank { "--" })
                if (ticket.isProcessed()) {
                    InfoRow(
                        t(Str.ObjectionOpReason, emptyArray()),
                        ticket.opReason.ifBlank { ticket.opTypeLabel() },
                    )
                    ticket.refundCost?.let {
                        InfoRow(t(Str.ObjectionRefundRideFee, emptyArray()), OrderFormat.yuanWithUnit(it))
                    }
                    ticket.updatePayCost?.let {
                        InfoRow(t(Str.ObjectionModifyRideFee, emptyArray()), OrderFormat.yuanWithUnit(it))
                    }
                } else {
                    Spacer(Modifier.height(8.dp))
                    Row(
                        Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.Center,
                    ) {
                        // 对齐原版：通过 → feeReasonable=false；驳回 → feeReasonable=true
                        OutlinedButton(onClick = onAdopt, modifier = Modifier.padding(end = 24.dp)) {
                            Text(t(Str.ObjectionAdopt, emptyArray()), color = LinkBlue)
                        }
                        OutlinedButton(onClick = onReject) {
                            Text(t(Str.ObjectionReject, emptyArray()), color = LinkBlue)
                        }
                    }
                }
            }
            SectionCard(title = t(Str.ObjectionOrderSection, emptyArray())) {
                val name = order?.userName?.ifBlank { null } ?: ticket.userName
                val phone = order?.phone?.ifBlank { null } ?: ticket.phone
                InfoRow(t(Str.OrderQueryNameLabel, emptyArray()), name.ifBlank { "--" })
                InfoRow(
                    t(Str.PhoneColonLabel, emptyArray()),
                    phone.ifBlank { "--" },
                    valueColor = LinkBlue,
                )
                InfoRow(
                    t(Str.VehicleCarNumberColon, emptyArray()),
                    (order?.carId ?: ticket.carId).ifBlank { "--" },
                )
                InfoRow(t(Str.OrderIdColon, emptyArray()), (order?.id ?: ticket.orderId).ifBlank { "--" })
                val pay = order?.payCost ?: ticket.payCost
                val paidLabel = when {
                    ticket.isPaid() || order?.izPaid == 4 -> t(Str.ObjectionPaid, emptyArray())
                    ticket.izPaid == 3 || order?.izPaid == 3 -> t(Str.ObjectionUnpaid, emptyArray())
                    else -> "--"
                }
                InfoRow(
                    "支付金额",
                    "${OrderFormat.yuanWithUnit(pay)} $paidLabel",
                )
                InfoRow(
                    t(Str.OrderDurationColon, emptyArray()),
                    OrderFormat.duration(order?.ridingTimeRaw ?: ticket.ridingTimeRaw),
                )
                InfoRow(
                    t(Str.MileageColon, emptyArray()),
                    OrderFormat.distance(order?.mile ?: ticket.mile),
                )
                InfoRow("开始时间:", order?.startTime?.ifBlank { "--" } ?: "--")
                InfoRow("结束时间:", order?.endTime?.ifBlank { "--" } ?: "--")
            }
            if (ticket.isProcessed()) {
                SectionCard(title = t(Str.ObjectionHandlerSection, emptyArray())) {
                    InfoRow(t(Str.OrderQueryNameLabel, emptyArray()), ticket.opManName.ifBlank { "--" })
                    InfoRow(t(Str.PhoneColonLabel, emptyArray()), ticket.opManPhone.ifBlank { "--" })
                    InfoRow("处理时间:", ticket.dealAt.ifBlank { "--" })
                }
            }
        }
    }
}

@Composable
private fun ObjectionProcessPage(
    t: (Str, Array<out Any?>) -> String,
    detail: ObjectionOrderDetail,
    feeReasonable: Boolean,
    state: com.luopingtech.ebike.ops.feature.admin.ObjectionOrderUiState,
    primary: Color,
    canSubmit: Boolean,
    onBack: () -> Unit,
    onReason: (String) -> Unit,
    onModifyPay: (String) -> Unit,
    onModifyDispatch: (String) -> Unit,
    onModifyHelmet: (String) -> Unit,
    onRefundPay: (String) -> Unit,
    onRefundDispatch: (String) -> Unit,
    onRefundHelmet: (String) -> Unit,
    onRefundCard: (String) -> Unit,
    onToggleSys: () -> Unit,
    onToggleSms: () -> Unit,
    onToggleApp: () -> Unit,
    onSubmit: () -> Unit,
) {
    val ticket = detail.ticket
    val amounts = detail.processAmounts()
    val paid = amounts.isPaid()
    val title = if (feeReasonable) {
        t(Str.ObjectionPassHandle, emptyArray())
    } else {
        t(Str.ObjectionBackHandle, emptyArray())
    }
    Column(Modifier.fillMaxSize().background(PageBg)) {
        VcdTopBar(title = t(Str.ObjectionProcessTitle, emptyArray()), onBack = onBack)
        Column(
            Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState())
                .padding(horizontal = 16.dp, vertical = 12.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            state.errorMessage?.let { Text(it, color = Color(0xFFE02020), fontSize = 13.sp) }
            state.noticeMessage?.let { Text(it, color = TipOrange, fontSize = 13.sp) }

            Text(title, color = TextPrimary, fontSize = 16.sp, fontWeight = FontWeight.Bold)

            InfoRow(
                t(Str.ObjectionPayStatus, emptyArray()),
                if (paid) t(Str.ObjectionPaid, emptyArray()) else t(Str.ObjectionUnpaid, emptyArray()),
                valueColor = if (paid) ProcessedGreen else UnpaidRed,
            )
            InfoRow(t(Str.ObjectionFavorable, emptyArray()), ticket.favorableTypeLabel())
            InfoRow(t(Str.ObjectionTravelFee, emptyArray()), OrderFormat.yuanWithUnit(amounts.payCostFen))
            if (!feeReasonable) {
                InfoRow(t(Str.ObjectionRidingFeeLabel, emptyArray()), OrderFormat.yuanWithUnit(amounts.originCostFen))
                InfoRow(t(Str.ObjectionDispatchFeeLabel, emptyArray()), OrderFormat.yuanWithUnit(amounts.dispatchCostFen))
            }
            InfoRow(t(Str.ObjectionHelmetLabel, emptyArray()), OrderFormat.yuanWithUnit(amounts.helmetPenaltyFen))
            if (!feeReasonable) {
                InfoRow(t(Str.ObjectionHandleWay, emptyArray()), ticket.handleWayLabel(false, amounts))
            }

            // 卡券：有优惠时展示（驳回也可填，对齐原版 XML）
            if (amounts.showCardTimes()) {
                InlineAmountField(
                    label = t(Str.ObjectionRefundCardTimes, emptyArray()),
                    value = state.refundCardTimes,
                    hint = t(Str.ObjectionTimesHint, emptyArray()),
                    keyboardType = KeyboardType.Number,
                    onValueChange = onRefundCard,
                )
            }

            if (!feeReasonable) {
                if (paid) {
                    if (amounts.showRefundOrigin()) {
                        InlineAmountField(
                            label = t(Str.ObjectionRefundRideLabel, emptyArray()),
                            value = state.refundPayYuan,
                            hint = t(Str.ObjectionAmountHint, emptyArray()),
                            onValueChange = onRefundPay,
                        )
                    }
                    if (amounts.showRefundDispatch()) {
                        InlineAmountField(
                            label = t(Str.ObjectionRefundDispatchLabel, emptyArray()),
                            value = state.refundDispatchYuan,
                            hint = t(Str.ObjectionAmountHint, emptyArray()),
                            onValueChange = onRefundDispatch,
                        )
                    }
                } else {
                    InlineAmountField(
                        label = t(Str.ObjectionModifyRideLabel, emptyArray()),
                        value = state.modifyPayYuan,
                        hint = t(Str.ObjectionAmountHint, emptyArray()),
                        onValueChange = onModifyPay,
                    )
                    InlineAmountField(
                        label = t(Str.ObjectionModifyDispatchLabel, emptyArray()),
                        value = state.modifyDispatchYuan,
                        hint = t(Str.ObjectionAmountHint, emptyArray()),
                        onValueChange = onModifyDispatch,
                    )
                }
                if (amounts.showHelmetInput()) {
                    InlineAmountField(
                        label = if (paid) {
                            t(Str.ObjectionRefundHelmetLabel, emptyArray())
                        } else {
                            t(Str.ObjectionModifyHelmetLabel, emptyArray())
                        },
                        value = if (paid) state.refundHelmetYuan else state.modifyHelmetYuan,
                        hint = t(Str.ObjectionAmountHint, emptyArray()),
                        onValueChange = if (paid) onRefundHelmet else onModifyHelmet,
                    )
                }
                Text(
                    if (paid) t(Str.ObjectionRefundTip, emptyArray())
                    else t(Str.ObjectionModifyTip, emptyArray()),
                    color = TipOrange,
                    fontSize = 14.sp,
                )
            }

            Spacer(Modifier.height(16.dp))
            Text(t(Str.ObjectionRemark, emptyArray()), color = TextPrimary, fontSize = 16.sp, fontWeight = FontWeight.Bold)
            BasicTextField(
                value = state.processReason,
                onValueChange = onReason,
                textStyle = TextStyle(color = TextPrimary, fontSize = 14.sp),
                cursorBrush = SolidColor(primary),
                modifier = Modifier
                    .fillMaxWidth()
                    .height(64.dp)
                    .border(1.dp, Color(0xFFD7D7D7), RoundedCornerShape(5.dp))
                    .padding(12.dp),
                decorationBox = { inner ->
                    Box(Modifier.fillMaxSize()) {
                        if (state.processReason.isEmpty()) {
                            Text(t(Str.ObjectionReasonInputHint, emptyArray()), color = TextMuted, fontSize = 14.sp)
                        }
                        inner()
                        Text(
                            "${state.processReason.length}/25",
                            color = TextMuted,
                            fontSize = 12.sp,
                            modifier = Modifier.align(Alignment.BottomEnd),
                        )
                    }
                },
            )

            Spacer(Modifier.height(8.dp))
            Box(Modifier.fillMaxWidth().height(8.dp).background(Color(0xFFF6F6F6)))
            Spacer(Modifier.height(8.dp))

            Text(t(Str.ObjectionNoticeTitle, emptyArray()), color = TextPrimary, fontSize = 16.sp, fontWeight = FontWeight.Bold)
            Text(t(Str.ObjectionNoticeTip, emptyArray()), color = TipOrange, fontSize = 13.sp)
            NoticeSwitchRow(t(Str.ObjectionNoticeSys, emptyArray()), state.noticeSys, onToggleSys)
            NoticeSwitchRow(t(Str.ObjectionNoticeSms, emptyArray()), state.noticeSms, onToggleSms)
            NoticeSwitchRow(t(Str.ObjectionNoticeApp, emptyArray()), state.noticeApp, onToggleApp)

            Spacer(Modifier.height(12.dp))
            Button(
                onClick = onSubmit,
                enabled = canSubmit && !state.submitting,
                modifier = Modifier.fillMaxWidth().height(48.dp),
            ) {
                if (state.submitting) {
                    CircularProgressIndicator(color = Color.White, modifier = Modifier.height(18.dp).width(18.dp))
                } else {
                    Text(t(Str.Confirm, emptyArray()), fontSize = 16.sp, fontWeight = FontWeight.Bold)
                }
            }
            Spacer(Modifier.height(24.dp))
        }
    }
}

@Composable
private fun NoticeSwitchRow(label: String, checked: Boolean, onToggle: () -> Unit) {
    Row(
        Modifier
            .fillMaxWidth()
            .height(56.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.SpaceBetween,
    ) {
        Text(label, color = TextPrimary, fontSize = 14.sp, fontWeight = FontWeight.Bold)
        Switch(checked = checked, onCheckedChange = { onToggle() })
    }
}

@Composable
private fun InlineAmountField(
    label: String,
    value: String,
    hint: String,
    keyboardType: KeyboardType = KeyboardType.Decimal,
    onValueChange: (String) -> Unit,
) {
    Row(
        Modifier.fillMaxWidth().padding(vertical = 4.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(label, color = TextPrimary, fontSize = 14.sp, fontWeight = FontWeight.Medium)
        Spacer(Modifier.width(10.dp))
        BasicTextField(
            value = value,
            onValueChange = onValueChange,
            singleLine = true,
            keyboardOptions = KeyboardOptions(keyboardType = keyboardType),
            textStyle = TextStyle(color = TextPrimary, fontSize = 12.sp),
            modifier = Modifier
                .width(100.dp)
                .height(34.dp)
                .border(1.dp, Color(0xFFD7D7D7), RoundedCornerShape(5.dp))
                .padding(horizontal = 10.dp, vertical = 8.dp),
            decorationBox = { inner ->
                if (value.isEmpty()) {
                    Text(hint, color = TextMuted, fontSize = 12.sp)
                }
                inner()
            },
        )
    }
}

@Composable
private fun SectionCard(title: String, content: @Composable () -> Unit) {
    Column(
        Modifier
            .fillMaxWidth()
            .background(Color.White, RoundedCornerShape(8.dp))
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(10.dp),
    ) {
        Text(title, color = TextPrimary, fontSize = 16.sp, fontWeight = FontWeight.Bold)
        HorizontalDivider(color = Color(0xFFEEF0F6))
        content()
    }
}

@Composable
private fun InfoRow(label: String, value: String, valueColor: Color = TextPrimary) {
    if (label.isEmpty() && value.isEmpty()) return
    Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
        Text(label, color = TextMuted, fontSize = 14.sp)
        Text(value, color = valueColor, fontSize = 14.sp, fontWeight = FontWeight.Medium)
    }
}
