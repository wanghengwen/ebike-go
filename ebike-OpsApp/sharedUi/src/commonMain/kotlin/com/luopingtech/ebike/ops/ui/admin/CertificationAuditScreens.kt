package com.luopingtech.ebike.ops.ui.admin

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
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
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
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
import androidx.compose.ui.platform.LocalUriHandler
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.domain.admin.CareerAuditDetail
import com.luopingtech.ebike.ops.domain.admin.CareerAuditItem
import com.luopingtech.ebike.ops.domain.admin.CertificationAuditStates
import com.luopingtech.ebike.ops.domain.admin.IdBindAuditDetail
import com.luopingtech.ebike.ops.domain.admin.IdBindAuditItem
import com.luopingtech.ebike.ops.domain.admin.ObjectionSendMode
import com.luopingtech.ebike.ops.domain.admin.applyTypeLabel
import com.luopingtech.ebike.ops.domain.admin.showAuditActions
import com.luopingtech.ebike.ops.domain.admin.showAuditResult
import com.luopingtech.ebike.ops.domain.admin.showRejectReason
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsTimeRanges
import com.luopingtech.ebike.ops.feature.admin.CertificationNav
import com.luopingtech.ebike.ops.feature.admin.IdBindAuditUiState
import com.luopingtech.ebike.ops.feature.admin.ProfessionAuditUiState
import com.luopingtech.ebike.ops.ui.analysis.VcdTopBar
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.launch

private val PageBg = Color(0xFFF6F7F9)
private val TextPrimary = Color(0xFF242936)
private val TextMuted = Color(0xFF7C87B1)
private val PendingOrange = Color(0xFFFFAD00)
private val PassedGreen = Color(0xFF1DBA4F)
private val RejectedRed = Color(0xFFFF0808)
private val CancelledGray = Color(0xFF999999)
private val LinkBlue = Color(0xFF1180F9)
private val DividerColor = Color(0xFFEEF0F6)

private enum class CertDatePickTarget { Start, End }

@Composable
fun ProfessionAuditScreen(app: OpsApp, onClose: () -> Unit) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val home by app.homeFeature.state.collectAsState()
    val state by app.professionAuditFeature.state.collectAsState()
    val feature = app.professionAuditFeature
    val scope = rememberCoroutineScope()
    val primary = OpsTheme.colors.primary

    LaunchedEffect(home.currentArea?.id) {
        feature.load(home.currentArea)
    }

    when (state.nav) {
        CertificationNav.List -> ProfessionListPage(
            t = { key, args -> t(key, *args) },
            state = state,
            primary = primary,
            onClose = {
                feature.clear()
                onClose()
            },
            onKeyword = feature::setKeyword,
            onStateFilter = { filter ->
                feature.setAuditStateFilter(filter)
                scope.launch { feature.search(home.currentArea) }
            },
            onDatesChange = { start, end ->
                if (feature.setDates(start, end)) {
                    scope.launch { feature.search(home.currentArea) }
                }
            },
            onSearch = { scope.launch { feature.search(home.currentArea) } },
            onLoadMore = { scope.launch { feature.loadMore(home.currentArea) } },
            onOpen = { id -> scope.launch { feature.openDetail(id) } },
        )
        CertificationNav.Detail -> {
            val detail = state.detail
            if (detail == null) {
                feature.navigateBack()
            } else {
                ProfessionDetailPage(
                    t = { key, args -> t(key, *args) },
                    detail = detail,
                    loading = state.loading,
                    error = state.errorMessage,
                    primary = primary,
                    onBack = { feature.navigateBack() },
                    onPass = { scope.launch { feature.openSubmit(pass = true) } },
                    onReject = { scope.launch { feature.openSubmit(pass = false) } },
                )
            }
        }
        CertificationNav.Submit -> {
            if (state.detail == null) {
                feature.navigateBack()
            } else {
                CertificationSubmitPage(
                    t = { key, args -> t(key, *args) },
                    pass = state.submitPass,
                    rejectReason = state.rejectReason,
                    noticeSys = state.noticeSys,
                    noticeSms = state.noticeSms,
                    noticeApp = state.noticeApp,
                    sendMode = state.sendMode,
                    loading = state.loading,
                    error = state.errorMessage,
                    canSubmit = feature.canSubmit(),
                    primary = primary,
                    onBack = { feature.navigateBack() },
                    onRejectReason = feature::setRejectReason,
                    onToggleSys = feature::toggleNoticeSys,
                    onToggleSms = feature::toggleNoticeSms,
                    onToggleApp = feature::toggleNoticeApp,
                    onSubmit = { scope.launch { feature.submit(home.currentArea) } },
                )
            }
        }
    }
}

@Composable
fun IdBindAuditScreen(app: OpsApp, onClose: () -> Unit) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val home by app.homeFeature.state.collectAsState()
    val state by app.idBindAuditFeature.state.collectAsState()
    val feature = app.idBindAuditFeature
    val scope = rememberCoroutineScope()
    val primary = OpsTheme.colors.primary

    LaunchedEffect(home.currentArea?.id) {
        feature.load(home.currentArea)
    }

    when (state.nav) {
        CertificationNav.List -> IdBindListPage(
            t = { key, args -> t(key, *args) },
            state = state,
            primary = primary,
            onClose = {
                feature.clear()
                onClose()
            },
            onKeyword = feature::setKeyword,
            onStateFilter = { filter ->
                feature.setAuditStateFilter(filter)
                scope.launch { feature.search(home.currentArea) }
            },
            onDatesChange = { start, end ->
                if (feature.setDates(start, end)) {
                    scope.launch { feature.search(home.currentArea) }
                }
            },
            onSearch = { scope.launch { feature.search(home.currentArea) } },
            onLoadMore = { scope.launch { feature.loadMore(home.currentArea) } },
            onOpen = { id -> scope.launch { feature.openDetail(id) } },
        )
        CertificationNav.Detail -> {
            val detail = state.detail
            if (detail == null) {
                feature.navigateBack()
            } else {
                IdBindDetailPage(
                    t = { key, args -> t(key, *args) },
                    detail = detail,
                    loading = state.loading,
                    error = state.errorMessage,
                    primary = primary,
                    onBack = { feature.navigateBack() },
                    onPass = { scope.launch { feature.openSubmit(pass = true) } },
                    onReject = { scope.launch { feature.openSubmit(pass = false) } },
                )
            }
        }
        CertificationNav.Submit -> {
            if (state.detail == null) {
                feature.navigateBack()
            } else {
                CertificationSubmitPage(
                    t = { key, args -> t(key, *args) },
                    pass = state.submitPass,
                    rejectReason = state.rejectReason,
                    noticeSys = state.noticeSys,
                    noticeSms = state.noticeSms,
                    noticeApp = state.noticeApp,
                    sendMode = state.sendMode,
                    loading = state.loading,
                    error = state.errorMessage,
                    canSubmit = feature.canSubmit(),
                    primary = primary,
                    onBack = { feature.navigateBack() },
                    onRejectReason = feature::setRejectReason,
                    onToggleSys = feature::toggleNoticeSys,
                    onToggleSms = feature::toggleNoticeSms,
                    onToggleApp = feature::toggleNoticeApp,
                    onSubmit = { scope.launch { feature.submit(home.currentArea) } },
                )
            }
        }
    }
}

@Composable
private fun ProfessionListPage(
    t: (Str, Array<out Any?>) -> String,
    state: ProfessionAuditUiState,
    primary: Color,
    onClose: () -> Unit,
    onKeyword: (String) -> Unit,
    onStateFilter: (Int?) -> Unit,
    onDatesChange: (String, String) -> Unit,
    onSearch: () -> Unit,
    onLoadMore: () -> Unit,
    onOpen: (String) -> Unit,
) {
    val listState = rememberLazyListState()
    var datePickTarget by remember { mutableStateOf<CertDatePickTarget?>(null) }
    val startDate = state.startDate.ifBlank {
        OfflineOpsTimeRanges.formatDate(OfflineOpsTimeRanges.objectionDefaultRange().startMs)
    }
    val endDate = state.endDate.ifBlank {
        OfflineOpsTimeRanges.formatDate(OfflineOpsTimeRanges.objectionDefaultRange().endMs)
    }

    LaunchedEffect(listState, state.loadingMore, state.finished, state.loading) {
        snapshotFlow {
            val info = listState.layoutInfo
            val last = info.visibleItemsInfo.lastOrNull()?.index ?: 0
            last >= info.totalItemsCount - 2 && info.totalItemsCount > 0
        }.distinctUntilChanged().collect { nearEnd ->
            if (nearEnd && !state.loading && !state.loadingMore && !state.finished) onLoadMore()
        }
    }

    Column(Modifier.fillMaxSize().background(PageBg)) {
        VcdTopBar(title = t(Str.ProfessionAudit, emptyArray()), onBack = onClose)
        CertDateRangeBar(
            startDate = startDate,
            endDate = endDate,
            primary = primary,
            toLabel = t(Str.ObjectionDateTo, emptyArray()),
            onPickStart = { datePickTarget = CertDatePickTarget.Start },
            onPickEnd = { datePickTarget = CertDatePickTarget.End },
        )
        CertSearchAndFilters(
            keyword = state.keyword,
            hint = t(Str.CertSearchHint, emptyArray()),
            searchLabel = t(Str.Search, emptyArray()),
            primary = primary,
            selectedFilter = state.auditStateFilter,
            filters = professionStateFilters(t),
            onKeyword = onKeyword,
            onSearch = onSearch,
            onStateFilter = onStateFilter,
        )
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
                    ProfessionListRow(
                        item = item,
                        t = t,
                        primary = primary,
                        onOpen = { onOpen(item.id) },
                    )
                }
                item { CertListFooter(t, state.loadingMore, state.finished, primary) }
            }
        }
    }

    datePickTarget?.let { target ->
        val initial = if (target == CertDatePickTarget.Start) startDate else endDate
        CertDayPickerDialog(
            initialDate = initial,
            primary = primary,
            confirmLabel = t(Str.Confirm, emptyArray()),
            cancelLabel = t(Str.Cancel, emptyArray()),
            onDismiss = { datePickTarget = null },
            onConfirm = { picked ->
                datePickTarget = null
                if (target == CertDatePickTarget.Start) onDatesChange(picked, endDate)
                else onDatesChange(startDate, picked)
            },
        )
    }
}

@Composable
private fun IdBindListPage(
    t: (Str, Array<out Any?>) -> String,
    state: IdBindAuditUiState,
    primary: Color,
    onClose: () -> Unit,
    onKeyword: (String) -> Unit,
    onStateFilter: (Int?) -> Unit,
    onDatesChange: (String, String) -> Unit,
    onSearch: () -> Unit,
    onLoadMore: () -> Unit,
    onOpen: (String) -> Unit,
) {
    val listState = rememberLazyListState()
    var datePickTarget by remember { mutableStateOf<CertDatePickTarget?>(null) }
    val startDate = state.startDate.ifBlank {
        OfflineOpsTimeRanges.formatDate(OfflineOpsTimeRanges.objectionDefaultRange().startMs)
    }
    val endDate = state.endDate.ifBlank {
        OfflineOpsTimeRanges.formatDate(OfflineOpsTimeRanges.objectionDefaultRange().endMs)
    }

    LaunchedEffect(listState, state.loadingMore, state.finished, state.loading) {
        snapshotFlow {
            val info = listState.layoutInfo
            val last = info.visibleItemsInfo.lastOrNull()?.index ?: 0
            last >= info.totalItemsCount - 2 && info.totalItemsCount > 0
        }.distinctUntilChanged().collect { nearEnd ->
            if (nearEnd && !state.loading && !state.loadingMore && !state.finished) onLoadMore()
        }
    }

    Column(Modifier.fillMaxSize().background(PageBg)) {
        VcdTopBar(title = t(Str.IdBindAudit, emptyArray()), onBack = onClose)
        CertDateRangeBar(
            startDate = startDate,
            endDate = endDate,
            primary = primary,
            toLabel = t(Str.ObjectionDateTo, emptyArray()),
            onPickStart = { datePickTarget = CertDatePickTarget.Start },
            onPickEnd = { datePickTarget = CertDatePickTarget.End },
        )
        CertSearchAndFilters(
            keyword = state.keyword,
            hint = t(Str.CertSearchHint, emptyArray()),
            searchLabel = t(Str.Search, emptyArray()),
            primary = primary,
            selectedFilter = state.auditStateFilter,
            filters = idBindStateFilters(t),
            onKeyword = onKeyword,
            onSearch = onSearch,
            onStateFilter = onStateFilter,
        )
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
                    IdBindListRow(
                        item = item,
                        t = t,
                        primary = primary,
                        onOpen = { onOpen(item.id) },
                    )
                }
                item { CertListFooter(t, state.loadingMore, state.finished, primary) }
            }
        }
    }

    datePickTarget?.let { target ->
        val initial = if (target == CertDatePickTarget.Start) startDate else endDate
        CertDayPickerDialog(
            initialDate = initial,
            primary = primary,
            confirmLabel = t(Str.Confirm, emptyArray()),
            cancelLabel = t(Str.Cancel, emptyArray()),
            onDismiss = { datePickTarget = null },
            onConfirm = { picked ->
                datePickTarget = null
                if (target == CertDatePickTarget.Start) onDatesChange(picked, endDate)
                else onDatesChange(startDate, picked)
            },
        )
    }
}

@Composable
private fun CertDateRangeBar(
    startDate: String,
    endDate: String,
    primary: Color,
    toLabel: String,
    onPickStart: () -> Unit,
    onPickEnd: () -> Unit,
) {
    Column(
        Modifier
            .fillMaxWidth()
            .background(primary)
            .padding(horizontal = 16.dp),
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
                modifier = Modifier.clickable(onClick = onPickStart),
            )
            Text(
                toLabel,
                color = TextPrimary,
                fontSize = 16.sp,
                modifier = Modifier.padding(horizontal = 8.dp),
            )
            Text(
                endDate,
                color = LinkBlue,
                fontSize = 16.sp,
                modifier = Modifier.clickable(onClick = onPickEnd),
            )
        }
    }
}

@Composable
private fun CertSearchAndFilters(
    keyword: String,
    hint: String,
    searchLabel: String,
    primary: Color,
    selectedFilter: Int?,
    filters: List<Pair<Int?, String>>,
    onKeyword: (String) -> Unit,
    onSearch: () -> Unit,
    onStateFilter: (Int?) -> Unit,
) {
    Column(
        Modifier
            .fillMaxWidth()
            .background(Color.White)
            .padding(horizontal = 16.dp, vertical = 12.dp),
        verticalArrangement = Arrangement.spacedBy(10.dp),
    ) {
        Row(verticalAlignment = Alignment.CenterVertically) {
            BasicTextField(
                value = keyword,
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
                    if (keyword.isEmpty()) {
                        Text(hint, color = TextMuted, fontSize = 14.sp)
                    }
                    inner()
                },
            )
            Spacer(Modifier.width(8.dp))
            Button(onClick = onSearch) { Text(searchLabel) }
        }
        Row(
            Modifier
                .fillMaxWidth()
                .horizontalScroll(rememberScrollState()),
            horizontalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            filters.forEach { (value, label) ->
                CertStateChip(
                    label = label,
                    selected = selectedFilter == value,
                    primary = primary,
                    onClick = { onStateFilter(value) },
                )
            }
        }
    }
}

@Composable
private fun CertListFooter(
    t: (Str, Array<out Any?>) -> String,
    loadingMore: Boolean,
    finished: Boolean,
    primary: Color,
) {
    when {
        loadingMore -> Box(
            Modifier.fillMaxWidth().padding(16.dp),
            contentAlignment = Alignment.Center,
        ) { CircularProgressIndicator(color = primary) }
        finished -> Text(
            t(Str.VehicleListEnd, emptyArray()),
            modifier = Modifier.padding(16.dp).fillMaxWidth(),
            color = TextMuted,
            fontSize = 12.sp,
            textAlign = TextAlign.Center,
        )
    }
}

@Composable
private fun ProfessionListRow(
    item: CareerAuditItem,
    t: (Str, Array<out Any?>) -> String,
    primary: Color,
    onOpen: () -> Unit,
) {
    val pending = item.auditState == CertificationAuditStates.Pending
    Column(
        Modifier
            .fillMaxWidth()
            .background(Color.White)
            .clickable(onClick = onOpen)
            .padding(horizontal = 16.dp, vertical = 14.dp),
    ) {
        Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
            Text(t(Str.CertApplicant, emptyArray()), color = TextPrimary, fontSize = 14.sp)
            Text(
                item.name.ifBlank { "--" },
                color = TextPrimary,
                fontSize = 14.sp,
                fontWeight = FontWeight.Bold,
                modifier = Modifier.weight(1f).padding(start = 4.dp),
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            Text(
                certStateLabel(t, item.auditState),
                color = certStateColor(item.auditState),
                fontSize = 14.sp,
                fontWeight = FontWeight.Bold,
            )
        }
        Spacer(Modifier.height(8.dp))
        CertLabelValue(t(Str.CertPhone, emptyArray()), item.phone.ifBlank { "--" })
        Spacer(Modifier.height(8.dp))
        CertLabelValue(t(Str.CertApplyTime, emptyArray()), item.createdAt.ifBlank { "--" })
        Spacer(Modifier.height(8.dp))
        Row(verticalAlignment = Alignment.CenterVertically) {
            Text(t(Str.CertApplyNo, emptyArray()), color = TextPrimary, fontSize = 14.sp)
            Text(
                item.id.ifBlank { "--" },
                color = TextPrimary,
                fontSize = 14.sp,
                fontWeight = FontWeight.Medium,
                modifier = Modifier.weight(1f).padding(start = 4.dp),
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            CertActionButton(
                pending = pending,
                primary = primary,
                toAudit = t(Str.CertToAudit, emptyArray()),
                lookDetail = t(Str.CertLookDetail, emptyArray()),
                onClick = onOpen,
            )
        }
    }
    HorizontalDivider(color = DividerColor)
}

@Composable
private fun IdBindListRow(
    item: IdBindAuditItem,
    t: (Str, Array<out Any?>) -> String,
    primary: Color,
    onOpen: () -> Unit,
) {
    val pending = item.auditState == CertificationAuditStates.Pending
    val typeLabel = applyTypeLabel(
        item.applyType,
        t(Str.CertPhoneBind, emptyArray()),
        t(Str.CertIdBind, emptyArray()),
    )
    Column(
        Modifier
            .fillMaxWidth()
            .background(Color.White)
            .clickable(onClick = onOpen)
            .padding(horizontal = 16.dp, vertical = 14.dp),
    ) {
        Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
            Text(t(Str.CertApplicant, emptyArray()), color = TextPrimary, fontSize = 14.sp)
            Text(
                item.authName.ifBlank { "--" },
                color = TextPrimary,
                fontSize = 14.sp,
                fontWeight = FontWeight.Bold,
                modifier = Modifier.weight(1f).padding(start = 4.dp),
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            Text(
                certStateLabel(t, item.auditState),
                color = certStateColor(item.auditState),
                fontSize = 14.sp,
                fontWeight = FontWeight.Bold,
            )
        }
        Spacer(Modifier.height(8.dp))
        CertLabelValue(t(Str.CertPhone, emptyArray()), item.applyPhone.ifBlank { "--" })
        Spacer(Modifier.height(8.dp))
        CertLabelValue(t(Str.CertApplyType, emptyArray()), typeLabel)
        Spacer(Modifier.height(8.dp))
        Row(verticalAlignment = Alignment.CenterVertically) {
            Text(t(Str.CertApplyTime, emptyArray()), color = TextPrimary, fontSize = 14.sp)
            Text(
                item.createdAt.ifBlank { "--" },
                color = TextPrimary,
                fontSize = 14.sp,
                modifier = Modifier.weight(1f).padding(start = 4.dp),
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            CertActionButton(
                pending = pending,
                primary = primary,
                toAudit = t(Str.CertToAudit, emptyArray()),
                lookDetail = t(Str.CertLookDetail, emptyArray()),
                onClick = onOpen,
            )
        }
    }
    HorizontalDivider(color = DividerColor)
}

@Composable
private fun CertActionButton(
    pending: Boolean,
    primary: Color,
    toAudit: String,
    lookDetail: String,
    onClick: () -> Unit,
) {
    Box(
        Modifier
            .background(
                if (pending) primary else Color.White,
                RoundedCornerShape(5.dp),
            )
            .then(
                if (pending) Modifier
                else Modifier.border(1.dp, CancelledGray, RoundedCornerShape(5.dp)),
            )
            .clickable(onClick = onClick)
            .padding(horizontal = 12.dp, vertical = 6.dp),
    ) {
        Text(
            if (pending) toAudit else lookDetail,
            color = if (pending) Color.White else TextPrimary,
            fontSize = 14.sp,
        )
    }
}

@Composable
private fun CertLabelValue(label: String, value: String) {
    Row {
        Text(label, color = TextPrimary, fontSize = 14.sp)
        Text(
            value,
            color = TextPrimary,
            fontSize = 14.sp,
            fontWeight = FontWeight.Medium,
            modifier = Modifier.padding(start = 4.dp),
        )
    }
}

@Composable
private fun ProfessionDetailPage(
    t: (Str, Array<out Any?>) -> String,
    detail: CareerAuditDetail,
    loading: Boolean,
    error: String?,
    primary: Color,
    onBack: () -> Unit,
    onPass: () -> Unit,
    onReject: () -> Unit,
) {
    val uriHandler = LocalUriHandler.current
    Column(Modifier.fillMaxSize().background(PageBg)) {
        VcdTopBar(title = t(Str.ProfessionAudit, emptyArray()), onBack = onBack)
        if (loading && detail.id.isBlank()) {
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                CircularProgressIndicator(color = primary)
            }
            return
        }
        Column(
            Modifier
                .weight(1f)
                .verticalScroll(rememberScrollState())
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            error?.let { Text(it, color = Color(0xFFE02020)) }
            CertDetailCard {
                CertDetailStateRow(t, detail.auditState)
                CertDetailInfoRow(
                    t(Str.CertApplicant, emptyArray()),
                    detail.name.ifBlank { "--" },
                )
                CertDetailInfoRow(
                    t(Str.CertApplyTime, emptyArray()),
                    detail.createdAt.ifBlank { "--" },
                )
                CertDetailInfoRow(
                    t(Str.CertApplyNo, emptyArray()),
                    detail.id.ifBlank { "--" },
                )
                CertPhoneRow(
                    label = t(Str.CertPhone, emptyArray()),
                    phone = detail.phone,
                    onDial = { raw ->
                        if (raw.isNotBlank()) runCatching { uriHandler.openUri("tel:$raw") }
                    },
                )
                CertDetailInfoRow(
                    t(Str.CertCompany, emptyArray()),
                    detail.company.ifBlank { "--" },
                )
                CertDetailInfoRow(
                    t(Str.CertNo, emptyArray()),
                    detail.certNo.ifBlank { "--" },
                )
                CertPicsSection(
                    label = t(Str.CertPics, emptyArray()),
                    front = detail.frontCard,
                    back = detail.backCard,
                    onOpen = { url -> runCatching { uriHandler.openUri(url) } },
                )
            }
            if (detail.showAuditResult()) {
                CertDetailCard(title = t(Str.CertDealResult, emptyArray())) {
                    CertDetailInfoRow(
                        t(Str.CertHandler, emptyArray()),
                        buildString {
                            append(detail.auditName.ifBlank { "--" })
                            if (detail.auditAt.isNotBlank()) {
                                append("  ")
                                append(detail.auditAt)
                            }
                        },
                    )
                    CertPhoneRow(
                        label = t(Str.CertPhone, emptyArray()),
                        phone = detail.auditPhone,
                        onDial = { raw ->
                            if (raw.isNotBlank()) runCatching { uriHandler.openUri("tel:$raw") }
                        },
                    )
                    if (detail.showRejectReason()) {
                        CertDetailInfoRow(
                            t(Str.CertDealReason, emptyArray()),
                            detail.reason.ifBlank { "--" },
                        )
                    }
                }
            }
        }
        if (detail.showAuditActions()) {
            CertDetailActions(
                rejectLabel = t(Str.AuditReject, emptyArray()),
                passLabel = t(Str.AuditPass, emptyArray()),
                primary = primary,
                onReject = onReject,
                onPass = onPass,
            )
        }
    }
}

@Composable
private fun IdBindDetailPage(
    t: (Str, Array<out Any?>) -> String,
    detail: IdBindAuditDetail,
    loading: Boolean,
    error: String?,
    primary: Color,
    onBack: () -> Unit,
    onPass: () -> Unit,
    onReject: () -> Unit,
) {
    val uriHandler = LocalUriHandler.current
    val typeLabel = applyTypeLabel(
        detail.applyType,
        t(Str.CertPhoneBind, emptyArray()),
        t(Str.CertIdBind, emptyArray()),
    )
    Column(Modifier.fillMaxSize().background(PageBg)) {
        VcdTopBar(title = t(Str.IdBindAudit, emptyArray()), onBack = onBack)
        if (loading && detail.id.isBlank()) {
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                CircularProgressIndicator(color = primary)
            }
            return
        }
        Column(
            Modifier
                .weight(1f)
                .verticalScroll(rememberScrollState())
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            error?.let { Text(it, color = Color(0xFFE02020)) }
            CertDetailCard {
                CertDetailStateRow(t, detail.auditState)
                CertDetailInfoRow(
                    t(Str.CertApplicant, emptyArray()),
                    detail.authName.ifBlank { "--" },
                )
                CertDetailInfoRow(
                    t(Str.CertApplyTime, emptyArray()),
                    detail.createdAt.ifBlank { "--" },
                )
                CertDetailInfoRow(t(Str.CertApplyType, emptyArray()), typeLabel)
                CertPhoneRow(
                    label = t(Str.CertPhone, emptyArray()),
                    phone = detail.applyPhone,
                    onDial = { raw ->
                        if (raw.isNotBlank()) runCatching { uriHandler.openUri("tel:$raw") }
                    },
                )
                CertPhoneRow(
                    label = t(Str.CertOriginPhone, emptyArray()),
                    phone = detail.originPhone,
                    onDial = { raw ->
                        if (raw.isNotBlank()) runCatching { uriHandler.openUri("tel:$raw") }
                    },
                )
                CertDetailInfoRow(
                    t(Str.CertAuthNo, emptyArray()),
                    detail.authNo.ifBlank { "--" },
                )
                CertPicsSection(
                    label = t(Str.CertPics, emptyArray()),
                    front = detail.frontCard,
                    back = detail.backCard,
                    onOpen = { url -> runCatching { uriHandler.openUri(url) } },
                )
            }
            if (detail.showAuditResult()) {
                CertDetailCard(title = t(Str.CertDealResult, emptyArray())) {
                    CertDetailInfoRow(
                        t(Str.CertHandler, emptyArray()),
                        buildString {
                            append(detail.dealerName.ifBlank { "--" })
                            if (detail.dealTime.isNotBlank()) {
                                append("  ")
                                append(detail.dealTime)
                            }
                        },
                    )
                    CertPhoneRow(
                        label = t(Str.CertPhone, emptyArray()),
                        phone = detail.dealerPhone,
                        onDial = { raw ->
                            if (raw.isNotBlank()) runCatching { uriHandler.openUri("tel:$raw") }
                        },
                    )
                    CertDetailInfoRow(
                        t(Str.CertDealReason, emptyArray()),
                        detail.reason.ifBlank { "--" },
                    )
                }
            }
        }
        if (detail.showAuditActions()) {
            CertDetailActions(
                rejectLabel = t(Str.AuditReject, emptyArray()),
                passLabel = t(Str.AuditPass, emptyArray()),
                primary = primary,
                onReject = onReject,
                onPass = onPass,
            )
        }
    }
}

@Composable
private fun CertificationSubmitPage(
    t: (Str, Array<out Any?>) -> String,
    pass: Boolean,
    rejectReason: String,
    noticeSys: Boolean,
    noticeSms: Boolean,
    noticeApp: Boolean,
    sendMode: ObjectionSendMode,
    loading: Boolean,
    error: String?,
    canSubmit: Boolean,
    primary: Color,
    onBack: () -> Unit,
    onRejectReason: (String) -> Unit,
    onToggleSys: () -> Unit,
    onToggleSms: () -> Unit,
    onToggleApp: () -> Unit,
    onSubmit: () -> Unit,
) {
    val title = if (pass) t(Str.CertPassTitle, emptyArray()) else t(Str.CertRejectTitle, emptyArray())
    Column(Modifier.fillMaxSize().background(PageBg)) {
        Row(
            Modifier
                .fillMaxWidth()
                .background(Color.White)
                .padding(horizontal = 8.dp, vertical = 4.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            TextButton(onClick = onBack) {
                Text(t(Str.Cancel, emptyArray()), color = TextPrimary, fontSize = 16.sp)
            }
            Text(
                title,
                color = TextPrimary,
                fontSize = 17.sp,
                fontWeight = FontWeight.Bold,
                textAlign = TextAlign.Center,
                modifier = Modifier.weight(1f),
            )
            TextButton(
                onClick = onSubmit,
                enabled = canSubmit && !loading,
            ) {
                Text(
                    t(Str.Finish, emptyArray()),
                    color = if (canSubmit && !loading) primary else TextMuted,
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Bold,
                )
            }
        }
        HorizontalDivider(color = DividerColor)
        Column(
            Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState())
                .padding(horizontal = 16.dp, vertical = 12.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            error?.let { Text(it, color = Color(0xFFE02020), fontSize = 13.sp) }
            if (loading) {
                Box(Modifier.fillMaxWidth(), contentAlignment = Alignment.Center) {
                    CircularProgressIndicator(color = primary, modifier = Modifier.size(24.dp))
                }
            }

            CertNoticeSwitchRow(
                label = noticeLabel(t, Str.ObjectionNoticeSys, sendMode.izSys),
                checked = noticeSys,
                enabled = sendMode.izSys,
                onToggle = onToggleSys,
            )
            CertNoticeSwitchRow(
                label = noticeLabel(t, Str.ObjectionNoticeSms, sendMode.izSms),
                checked = noticeSms,
                enabled = sendMode.izSms,
                onToggle = onToggleSms,
            )
            CertNoticeSwitchRow(
                label = noticeLabel(t, Str.ObjectionNoticeApp, sendMode.izApp),
                checked = noticeApp,
                enabled = sendMode.izApp,
                onToggle = onToggleApp,
            )

            if (!pass) {
                Spacer(Modifier.height(8.dp))
                Text(
                    t(Str.CertRejectReason, emptyArray()),
                    color = TextPrimary,
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Bold,
                )
                BasicTextField(
                    value = rejectReason,
                    onValueChange = onRejectReason,
                    textStyle = TextStyle(color = TextPrimary, fontSize = 14.sp),
                    cursorBrush = SolidColor(primary),
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(88.dp)
                        .border(1.dp, Color(0xFFD7D7D7), RoundedCornerShape(5.dp))
                        .padding(12.dp),
                    decorationBox = { inner ->
                        Box(Modifier.fillMaxSize()) {
                            if (rejectReason.isEmpty()) {
                                Text(
                                    t(Str.CertRejectReasonHint, emptyArray()),
                                    color = TextMuted,
                                    fontSize = 14.sp,
                                )
                            }
                            inner()
                            Text(
                                "${rejectReason.length}/20",
                                color = TextMuted,
                                fontSize = 12.sp,
                                modifier = Modifier.align(Alignment.BottomEnd),
                            )
                        }
                    },
                )
            }
        }
    }
}

@Composable
private fun CertDetailCard(
    title: String? = null,
    content: @Composable () -> Unit,
) {
    Column(
        Modifier
            .fillMaxWidth()
            .background(Color.White, RoundedCornerShape(8.dp))
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(10.dp),
    ) {
        if (title != null) {
            Text(title, color = TextPrimary, fontSize = 16.sp, fontWeight = FontWeight.Bold)
            HorizontalDivider(color = DividerColor)
        }
        content()
    }
}

@Composable
private fun CertDetailStateRow(t: (Str, Array<out Any?>) -> String, auditState: Int) {
    val stateText = certStateLabel(t, auditState)
    Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
        Text(t(Str.AuditStateLabel, arrayOf("")).trimEnd(), color = TextMuted, fontSize = 14.sp)
        Text(
            stateText,
            color = certStateColor(auditState),
            fontSize = 14.sp,
            fontWeight = FontWeight.Bold,
        )
    }
}

@Composable
private fun CertDetailInfoRow(label: String, value: String) {
    Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
        Text(label, color = TextMuted, fontSize = 14.sp)
        Text(
            value,
            color = TextPrimary,
            fontSize = 14.sp,
            fontWeight = FontWeight.Medium,
            modifier = Modifier.padding(start = 12.dp),
            textAlign = TextAlign.End,
        )
    }
}

@Composable
private fun CertPhoneRow(label: String, phone: String, onDial: (String) -> Unit) {
    val display = phone.ifBlank { "--" }
    val dialable = phone.filter { it.isDigit() || it == '+' }
    Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
        Text(label, color = TextMuted, fontSize = 14.sp)
        Text(
            display,
            color = if (dialable.isNotBlank()) LinkBlue else TextPrimary,
            fontSize = 14.sp,
            fontWeight = FontWeight.Medium,
            modifier = Modifier
                .padding(start = 12.dp)
                .then(
                    if (dialable.isNotBlank()) Modifier.clickable { onDial(dialable) }
                    else Modifier,
                ),
        )
    }
}

@Composable
private fun CertPicsSection(
    label: String,
    front: String,
    back: String,
    onOpen: (String) -> Unit,
) {
    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Text(label, color = TextMuted, fontSize = 14.sp)
        Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            CertPicLink(url = front, onOpen = onOpen, modifier = Modifier.weight(1f))
            CertPicLink(url = back, onOpen = onOpen, modifier = Modifier.weight(1f))
        }
    }
}

@Composable
private fun CertPicLink(url: String, onOpen: (String) -> Unit, modifier: Modifier = Modifier) {
    val text = url.ifBlank { "-" }
    Text(
        text,
        color = if (url.isNotBlank()) LinkBlue else TextPrimary,
        fontSize = 13.sp,
        maxLines = 2,
        overflow = TextOverflow.Ellipsis,
        modifier = modifier.then(
            if (url.isNotBlank()) Modifier.clickable { onOpen(url) } else Modifier,
        ),
    )
}

@Composable
private fun CertDetailActions(
    rejectLabel: String,
    passLabel: String,
    primary: Color,
    onReject: () -> Unit,
    onPass: () -> Unit,
) {
    Row(
        Modifier
            .fillMaxWidth()
            .background(Color.White)
            .padding(horizontal = 16.dp, vertical = 12.dp),
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        OutlinedButton(
            onClick = onReject,
            modifier = Modifier.weight(1f).height(44.dp),
            shape = RoundedCornerShape(5.dp),
        ) {
            Text(rejectLabel, color = TextPrimary, fontSize = 16.sp)
        }
        Button(
            onClick = onPass,
            modifier = Modifier.weight(1f).height(44.dp),
            shape = RoundedCornerShape(5.dp),
            colors = ButtonDefaults.buttonColors(containerColor = primary),
        ) {
            Text(passLabel, color = Color.White, fontSize = 16.sp, fontWeight = FontWeight.Bold)
        }
    }
}

@Composable
private fun CertNoticeSwitchRow(
    label: String,
    checked: Boolean,
    enabled: Boolean,
    onToggle: () -> Unit,
) {
    Row(
        Modifier
            .fillMaxWidth()
            .height(56.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.SpaceBetween,
    ) {
        Text(label, color = TextPrimary, fontSize = 14.sp, fontWeight = FontWeight.Bold)
        Switch(
            checked = checked,
            onCheckedChange = { onToggle() },
            enabled = enabled,
        )
    }
}

@Composable
private fun CertStateChip(
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
private fun CertDayPickerDialog(
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

private fun professionStateFilters(t: (Str, Array<out Any?>) -> String): List<Pair<Int?, String>> = listOf(
    null to t(Str.AuditStateAll, emptyArray()),
    CertificationAuditStates.Pending to t(Str.AuditResultPending, emptyArray()),
    CertificationAuditStates.Passed to t(Str.AuditStatePassed, emptyArray()),
    CertificationAuditStates.Rejected to t(Str.AuditStateRejected, emptyArray()),
    CertificationAuditStates.Cancelled to t(Str.AuditStateCancelled, emptyArray()),
)

private fun idBindStateFilters(t: (Str, Array<out Any?>) -> String): List<Pair<Int?, String>> =
    professionStateFilters(t) + listOf(
        CertificationAuditStates.FacePassed to t(Str.AuditStateFacePassed, emptyArray()),
    )

private fun certStateLabel(t: (Str, Array<out Any?>) -> String, state: Int): String = when (state) {
    CertificationAuditStates.Pending -> t(Str.AuditResultPending, emptyArray())
    CertificationAuditStates.Passed -> t(Str.AuditStatePassed, emptyArray())
    CertificationAuditStates.Rejected -> t(Str.AuditStateRejected, emptyArray())
    CertificationAuditStates.Cancelled -> t(Str.AuditStateCancelled, emptyArray())
    CertificationAuditStates.FacePassed -> t(Str.AuditStateFacePassed, emptyArray())
    else -> state.toString()
}

private fun certStateColor(state: Int): Color = when (state) {
    CertificationAuditStates.Pending -> PendingOrange
    CertificationAuditStates.Passed, CertificationAuditStates.FacePassed -> PassedGreen
    CertificationAuditStates.Rejected -> RejectedRed
    CertificationAuditStates.Cancelled -> CancelledGray
    else -> TextMuted
}

private fun noticeLabel(t: (Str, Array<out Any?>) -> String, key: Str, configured: Boolean): String {
    val base = t(key, emptyArray())
    return if (configured) base else t(Str.CertNoticeUnconfigured, arrayOf(base))
}
