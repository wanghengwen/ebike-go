package com.luopingtech.ebike.ops.ui.admin

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.horizontalScroll
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
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.domain.admin.OperationLogEventNode
import com.luopingtech.ebike.ops.domain.admin.OperationLogItem
import com.luopingtech.ebike.ops.domain.admin.OperationLogKind
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsTimeRanges
import com.luopingtech.ebike.ops.feature.admin.OperationLogNav
import com.luopingtech.ebike.ops.ui.analysis.VcdTopBar
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.launch

private val PageBg = Color(0xFFF6F7F9)
private val TextPrimary = Color(0xFF333333)
private val TextMuted = Color(0xFF999999)
private val LinkBlue = Color(0xFF1180F9)
private val DividerGray = Color(0xFFE5E5E5)

@Composable
fun OperationLogScreen(app: OpsApp, onClose: () -> Unit) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.operationLogFeature.state.collectAsState()
    val feature = app.operationLogFeature
    val scope = rememberCoroutineScope()
    val primary = OpsTheme.colors.primary

    LaunchedEffect(state.kind) {
        if (state.nav is OperationLogNav.Filter) feature.loadEventTree()
    }

    when (val nav = state.nav) {
        OperationLogNav.Filter -> OperationLogFilterPage(
            t = { key, args -> t(key, *args) },
            state = state,
            primary = primary,
            onClose = {
                feature.clear()
                onClose()
            },
            onKind = { kind ->
                feature.selectKind(kind)
                scope.launch { feature.loadEventTree() }
            },
            onTopInput = feature::setTopInput,
            onBottomPhone = feature::setBottomPhone,
            onStartDate = feature::setStartDate,
            onEndDate = feature::setEndDate,
            onEventSelected = feature::setEventSelection,
            onReset = {
                feature.resetFilter()
                scope.launch { feature.loadEventTree() }
            },
            onConfirm = {
                if (feature.confirmFilter()) {
                    scope.launch { feature.loadList(refresh = true) }
                }
            },
        )
        OperationLogNav.List -> OperationLogListPage(
            t = { key, args -> t(key, *args) },
            state = state,
            primary = primary,
            onBack = { feature.navigateBack() },
            onLoadMore = { scope.launch { feature.loadList(refresh = false) } },
            onOpen = feature::openDetail,
        )
        is OperationLogNav.Detail -> OperationLogDetailPage(
            t = { key, args -> t(key, *args) },
            result = nav.result,
            onBack = { feature.navigateBack() },
        )
    }
}

@Composable
private fun OperationLogFilterPage(
    t: (Str, Array<out Any?>) -> String,
    state: com.luopingtech.ebike.ops.feature.admin.OperationLogUiState,
    primary: Color,
    onClose: () -> Unit,
    onKind: (OperationLogKind) -> Unit,
    onTopInput: (String) -> Unit,
    onBottomPhone: (String) -> Unit,
    onStartDate: (String) -> Unit,
    onEndDate: (String) -> Unit,
    onEventSelected: (String?, String?) -> Unit,
    onReset: () -> Unit,
    onConfirm: () -> Unit,
) {
    var pickStart by remember { mutableStateOf(false) }
    var pickEnd by remember { mutableStateOf(false) }
    var showEventPicker by remember { mutableStateOf(false) }

    val topTitle = when (state.kind) {
        OperationLogKind.Car -> t(Str.OpLogCarIdLabel, emptyArray()).trimEnd('：', ':')
        OperationLogKind.Device -> t(Str.OpLogImeiLabel, emptyArray()).trimEnd('：', ':')
        OperationLogKind.Operator -> t(Str.OpLogPhoneLabel, emptyArray()).trimEnd('：', ':')
    }
    val topHint = when (state.kind) {
        OperationLogKind.Car -> t(Str.OpLogEnterCarId, emptyArray())
        OperationLogKind.Device -> t(Str.OpLogEnterImei, emptyArray())
        OperationLogKind.Operator -> t(Str.OpLogEnterPhone, emptyArray())
    }

    Column(Modifier.fillMaxSize().background(Color.White)) {
        VcdTopBar(title = t(Str.OpLogTitle, emptyArray()), onBack = onClose)
        Box(
            Modifier
                .fillMaxWidth()
                .background(primary)
                .padding(16.dp),
        ) {
            Row(
                Modifier
                    .fillMaxWidth()
                    .height(36.dp)
                    .border(1.dp, OpsTheme.colors.onPrimary, RoundedCornerShape(8.dp))
                    .background(primary, RoundedCornerShape(8.dp)),
            ) {
                KindTab(
                    label = t(Str.OpLogCar, emptyArray()),
                    selected = state.kind == OperationLogKind.Car,
                    primary = primary,
                    shape = RoundedCornerShape(topStart = 6.dp, bottomStart = 6.dp),
                    onClick = { onKind(OperationLogKind.Car) },
                    modifier = Modifier.weight(1f),
                )
                KindTab(
                    label = t(Str.OpLogDevice, emptyArray()),
                    selected = state.kind == OperationLogKind.Device,
                    primary = primary,
                    shape = RoundedCornerShape(0.dp),
                    onClick = { onKind(OperationLogKind.Device) },
                    modifier = Modifier.weight(1f),
                )
                KindTab(
                    label = t(Str.OpLogOperator, emptyArray()),
                    selected = state.kind == OperationLogKind.Operator,
                    primary = primary,
                    shape = RoundedCornerShape(topEnd = 6.dp, bottomEnd = 6.dp),
                    onClick = { onKind(OperationLogKind.Operator) },
                    modifier = Modifier.weight(1f),
                )
            }
        }

        Column(
            Modifier
                .weight(1f)
                .verticalScroll(rememberScrollState())
                .padding(horizontal = 16.dp, vertical = 12.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            state.errorMessage?.let { Text(it, color = Color(0xFFE02020), fontSize = 13.sp) }

            FilterField(
                title = topTitle,
                value = state.topInput,
                hint = topHint,
                keyboardType = if (state.kind == OperationLogKind.Device) KeyboardType.Text else KeyboardType.Number,
                onValueChange = onTopInput,
            )

            if (state.kind == OperationLogKind.Car) {
                FilterField(
                    title = t(Str.OpLogPhoneLabel, emptyArray()).trimEnd('：', ':'),
                    value = state.bottomPhone,
                    hint = t(Str.OpLogEnterPhone, emptyArray()),
                    keyboardType = KeyboardType.Number,
                    onValueChange = onBottomPhone,
                )
            }

            Row(
                Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween,
            ) {
                Text(
                    state.startDate,
                    color = LinkBlue,
                    fontSize = 16.sp,
                    modifier = Modifier.clickable { pickStart = true },
                )
                Text(t(Str.ObjectionDateTo, emptyArray()), color = TextPrimary, fontSize = 16.sp)
                Text(
                    state.endDate,
                    color = LinkBlue,
                    fontSize = 16.sp,
                    modifier = Modifier.clickable { pickEnd = true },
                )
            }

            if (state.kind != OperationLogKind.Device) {
                Row(
                    Modifier
                        .fillMaxWidth()
                        .clickable { showEventPicker = true }
                        .padding(vertical = 8.dp),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text(
                        state.eventLabel.ifBlank { t(Str.OpLogChooseEvent, emptyArray()) },
                        color = if (state.eventLabel.isBlank()) TextMuted else TextPrimary,
                        fontSize = 14.sp,
                        modifier = Modifier.weight(1f),
                    )
                    Text("›", color = TextMuted, fontSize = 18.sp)
                }
                HorizontalDivider(color = DividerGray)
            }
        }

        Row(
            Modifier
                .fillMaxWidth()
                .padding(16.dp),
            horizontalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            OutlinedButton(
                onClick = onReset,
                modifier = Modifier.weight(1f).height(48.dp),
            ) { Text(t(Str.OpLogReset, emptyArray())) }
            Button(
                onClick = onConfirm,
                modifier = Modifier.weight(1f).height(48.dp),
            ) { Text(t(Str.Confirm, emptyArray())) }
        }
    }

    if (pickStart) {
        OpLogDayPickerDialog(
            initialDate = state.startDate,
            primary = primary,
            confirmLabel = t(Str.Confirm, emptyArray()),
            cancelLabel = t(Str.Cancel, emptyArray()),
            onDismiss = { pickStart = false },
            onConfirm = {
                pickStart = false
                onStartDate(it)
            },
        )
    }
    if (pickEnd) {
        OpLogDayPickerDialog(
            initialDate = state.endDate,
            primary = primary,
            confirmLabel = t(Str.Confirm, emptyArray()),
            cancelLabel = t(Str.Cancel, emptyArray()),
            onDismiss = { pickEnd = false },
            onConfirm = {
                pickEnd = false
                onEndDate(it)
            },
        )
    }
    if (showEventPicker) {
        EventTreePickerDialog(
            roots = state.eventTree,
            pleaseSelect = t(Str.OpLogPleaseSelect, emptyArray()),
            confirmLabel = t(Str.Confirm, emptyArray()),
            cancelLabel = t(Str.Cancel, emptyArray()),
            primary = primary,
            onDismiss = { showEventPicker = false },
            onCancel = {
                showEventPicker = false
                onEventSelected(null, null)
            },
            onConfirm = { label, ids ->
                showEventPicker = false
                onEventSelected(label, ids)
            },
        )
    }
}

@Composable
private fun KindTab(
    label: String,
    selected: Boolean,
    primary: Color,
    shape: RoundedCornerShape,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val onPrimary = OpsTheme.colors.onPrimary
    Box(
        modifier
            .fillMaxHeight()
            .background(if (selected) Color.White else Color.Transparent, shape)
            .clickable(onClick = onClick),
        contentAlignment = Alignment.Center,
    ) {
        Text(
            label,
            // 选中：白底 + 主题色字；未选中：主题色底 + onPrimary 字（避免白底白字）
            color = if (selected) primary else onPrimary,
            fontSize = 14.sp,
            fontWeight = if (selected) FontWeight.Medium else FontWeight.Normal,
        )
    }
}

@Composable
private fun FilterField(
    title: String,
    value: String,
    hint: String,
    keyboardType: KeyboardType,
    onValueChange: (String) -> Unit,
) {
    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Text(title, color = TextPrimary, fontSize = 14.sp)
        BasicTextField(
            value = value,
            onValueChange = onValueChange,
            singleLine = true,
            keyboardOptions = KeyboardOptions(keyboardType = keyboardType),
            textStyle = TextStyle(color = TextPrimary, fontSize = 14.sp),
            modifier = Modifier
                .fillMaxWidth()
                .height(40.dp)
                .border(1.dp, Color(0xFFDCDCDC), RoundedCornerShape(4.dp))
                .padding(horizontal = 12.dp, vertical = 10.dp),
            decorationBox = { inner ->
                if (value.isEmpty()) Text(hint, color = TextMuted, fontSize = 14.sp)
                inner()
            },
        )
    }
}

@Composable
private fun OperationLogListPage(
    t: (Str, Array<out Any?>) -> String,
    state: com.luopingtech.ebike.ops.feature.admin.OperationLogUiState,
    primary: Color,
    onBack: () -> Unit,
    onLoadMore: () -> Unit,
    onOpen: (OperationLogItem) -> Unit,
) {
    val title = when (state.kind) {
        OperationLogKind.Car -> t(Str.OpLogCar, emptyArray())
        OperationLogKind.Device -> t(Str.OpLogDevice, emptyArray())
        OperationLogKind.Operator -> t(Str.OpLogOperator, emptyArray())
    }
    val listState = rememberLazyListState()
    val uriHandler = LocalUriHandler.current

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
        VcdTopBar(title = title, onBack = onBack)
        state.errorMessage?.let {
            Text(it, color = Color(0xFFE02020), modifier = Modifier.padding(16.dp))
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
                items(state.items, key = { "${it.time}-${it.phone}-${it.carId}-${it.eventName}-${it.result}" }) { item ->
                    when (state.kind) {
                        OperationLogKind.Car -> CarLogRow(t, item, onOpen) { phone ->
                            val raw = phone.filter { it.isDigit() || it == '+' }
                            if (raw.isNotBlank()) runCatching { uriHandler.openUri("tel:$raw") }
                        }
                        OperationLogKind.Device -> DeviceLogRow(t, item, onOpen)
                        OperationLogKind.Operator -> OperatorLogRow(t, item, onOpen) { phone ->
                            val raw = phone.filter { it.isDigit() || it == '+' }
                            if (raw.isNotBlank()) runCatching { uriHandler.openUri("tel:$raw") }
                        }
                    }
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
                            textAlign = TextAlign.Center,
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun CarLogRow(
    t: (Str, Array<out Any?>) -> String,
    item: OperationLogItem,
    onOpen: (OperationLogItem) -> Unit,
    onDial: (String) -> Unit,
) {
    Column(
        Modifier
            .fillMaxWidth()
            .background(Color.White)
            .clickable { onOpen(item) }
            .padding(horizontal = 16.dp, vertical = 24.dp),
    ) {
        Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
            Row {
                Text(t(Str.OpLogCarIdLabel, emptyArray()), color = TextPrimary, fontSize = 14.sp)
                Text(item.carId, color = TextPrimary, fontSize = 14.sp)
            }
            Row {
                Text(t(Str.OpLogEventLabel, emptyArray()), color = TextPrimary, fontSize = 14.sp)
                Text(item.eventName, color = LinkBlue, fontSize = 14.sp)
            }
        }
        Spacer(Modifier.height(8.dp))
        Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
            Row {
                Text(t(Str.OpLogNameLabel, emptyArray()), color = TextPrimary, fontSize = 14.sp)
                Text(item.name, color = TextPrimary, fontSize = 14.sp)
            }
            Row {
                Text(t(Str.OpLogPhoneLabel, emptyArray()), color = TextPrimary, fontSize = 14.sp)
                Text(
                    item.phone,
                    color = LinkBlue,
                    fontSize = 14.sp,
                    modifier = Modifier.clickable { onDial(item.phone) },
                )
            }
        }
        Spacer(Modifier.height(8.dp))
        Row {
            Text(t(Str.OpLogTimeLabel, emptyArray()), color = TextPrimary, fontSize = 14.sp)
            Text(item.time, color = TextPrimary, fontSize = 14.sp)
        }
        Spacer(Modifier.height(24.dp))
        HorizontalDivider(color = DividerGray)
    }
}

@Composable
private fun DeviceLogRow(
    t: (Str, Array<out Any?>) -> String,
    item: OperationLogItem,
    onOpen: (OperationLogItem) -> Unit,
) {
    Column(
        Modifier
            .fillMaxWidth()
            .background(Color.White)
            .clickable { onOpen(item) }
            .padding(horizontal = 16.dp, vertical = 24.dp),
    ) {
        Row {
            Text(t(Str.OpLogImeiLabel, emptyArray()), color = TextPrimary, fontSize = 14.sp)
            Text(item.imei, color = TextPrimary, fontSize = 14.sp)
        }
        Spacer(Modifier.height(8.dp))
        Row {
            Text(t(Str.OpLogEventLabel, emptyArray()), color = TextPrimary, fontSize = 14.sp)
            Text(item.eventName, color = LinkBlue, fontSize = 14.sp)
        }
        Spacer(Modifier.height(8.dp))
        Row {
            Text(t(Str.OpLogTimeLabel, emptyArray()), color = TextPrimary, fontSize = 14.sp)
            Text(item.time, color = TextPrimary, fontSize = 14.sp)
        }
        Spacer(Modifier.height(24.dp))
        HorizontalDivider(color = DividerGray)
    }
}

@Composable
private fun OperatorLogRow(
    t: (Str, Array<out Any?>) -> String,
    item: OperationLogItem,
    onOpen: (OperationLogItem) -> Unit,
    onDial: (String) -> Unit,
) {
    Column(
        Modifier
            .fillMaxWidth()
            .background(Color.White)
            .clickable { onOpen(item) }
            .padding(horizontal = 16.dp, vertical = 24.dp),
    ) {
        Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
            Row {
                Text(t(Str.OpLogPhoneLabel, emptyArray()), color = TextPrimary, fontSize = 14.sp)
                Text(
                    item.phone,
                    color = LinkBlue,
                    fontSize = 14.sp,
                    modifier = Modifier.clickable { onDial(item.phone) },
                )
            }
            Row {
                Text(t(Str.OpLogEventLabel, emptyArray()), color = TextPrimary, fontSize = 14.sp)
                Text(item.eventName, color = LinkBlue, fontSize = 14.sp)
            }
        }
        Spacer(Modifier.height(8.dp))
        Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
            Row {
                Text(t(Str.OpLogNameLabel, emptyArray()), color = TextPrimary, fontSize = 14.sp)
                Text(item.name, color = TextPrimary, fontSize = 14.sp)
            }
            Row {
                Text(t(Str.OpLogTimeLabel, emptyArray()), color = TextPrimary, fontSize = 14.sp)
                Text(item.time, color = TextPrimary, fontSize = 14.sp)
            }
        }
        Spacer(Modifier.height(24.dp))
        HorizontalDivider(color = DividerGray)
    }
}

@Composable
private fun OperationLogDetailPage(
    t: (Str, Array<out Any?>) -> String,
    result: String,
    onBack: () -> Unit,
) {
    Column(Modifier.fillMaxSize().background(Color.White)) {
        VcdTopBar(title = t(Str.OpLogDetail, emptyArray()), onBack = onBack)
        Text(
            result.ifBlank { "--" },
            color = TextPrimary,
            fontSize = 14.sp,
            modifier = Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState())
                .padding(16.dp),
        )
    }
}

@Composable
private fun EventTreePickerDialog(
    roots: List<OperationLogEventNode>,
    pleaseSelect: String,
    confirmLabel: String,
    cancelLabel: String,
    primary: Color,
    onDismiss: () -> Unit,
    onCancel: () -> Unit,
    onConfirm: (label: String, ids: String) -> Unit,
) {
    // 对齐原版 OperationLogFilterAdapter + TabLayout：
    // 每一级一个 Tab；点有子节点项时当前 Tab 改名并追加「请选择」Tab 展示子列表
    data class Level(
        val title: String,
        val options: List<OperationLogEventNode>,
        val selected: OperationLogEventNode? = null,
    )

    data class PickerState(
        val levels: List<Level>,
        val activeIndex: Int,
    )

    var picker by remember(roots) {
        mutableStateOf(
            PickerState(
                levels = listOf(Level(title = pleaseSelect, options = roots)),
                activeIndex = 0,
            ),
        )
    }
    val activeLevel = picker.levels.getOrNull(picker.activeIndex) ?: picker.levels.first()

    fun selectNode(node: OperationLogEventNode) {
        val idx = picker.activeIndex
        val truncated = picker.levels.take(idx + 1).toMutableList()
        truncated[idx] = truncated[idx].copy(title = node.eventName, selected = node)
        picker = if (node.children.isNotEmpty()) {
            // 有子级：追加「请选择」并切到该级（对齐 addTab + defaultSelected）
            truncated.add(Level(title = pleaseSelect, options = node.children))
            PickerState(levels = truncated, activeIndex = truncated.lastIndex)
        } else {
            // 叶子：只改当前 Tab 名，不追加空「请选择」
            PickerState(levels = truncated, activeIndex = idx)
        }
    }

    fun jumpToLevel(index: Int) {
        // 对齐 removeViewFromPosition：丢掉该级之后的 Tab
        picker = PickerState(
            levels = picker.levels.take(index + 1),
            activeIndex = index,
        )
    }

    Dialog(onDismissRequest = onDismiss) {
        Column(
            Modifier
                .fillMaxWidth()
                .background(Color.White, RoundedCornerShape(12.dp))
                .padding(16.dp),
        ) {
            Row(
                Modifier
                    .fillMaxWidth()
                    .horizontalScroll(rememberScrollState()),
                horizontalArrangement = Arrangement.spacedBy(8.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                picker.levels.forEachIndexed { index, level ->
                    if (index > 0) Text("›", color = TextMuted, fontSize = 13.sp)
                    Text(
                        level.title,
                        color = if (index == picker.activeIndex) primary else TextMuted,
                        fontSize = 13.sp,
                        modifier = Modifier.clickable { jumpToLevel(index) },
                    )
                }
            }
            Spacer(Modifier.height(12.dp))
            LazyColumn(
                Modifier
                    .fillMaxWidth()
                    .height(240.dp),
            ) {
                if (activeLevel.options.isEmpty()) {
                    item {
                        Text(
                            pleaseSelect,
                            color = TextMuted,
                            fontSize = 14.sp,
                            modifier = Modifier.padding(16.dp),
                        )
                    }
                } else {
                    items(
                        items = activeLevel.options,
                        key = { node -> "${node.id}-${node.eventName}" },
                    ) { node ->
                        val selected = activeLevel.selected?.let { sel ->
                            sel.id == node.id && sel.eventName == node.eventName
                        } == true
                        Text(
                            node.eventName,
                            color = if (selected) primary else TextPrimary,
                            fontSize = 15.sp,
                            fontWeight = if (selected) FontWeight.Medium else FontWeight.Normal,
                            modifier = Modifier
                                .fillMaxWidth()
                                .clickable { selectNode(node) }
                                .padding(vertical = 12.dp),
                        )
                        HorizontalDivider(color = DividerGray)
                    }
                }
            }
            Spacer(Modifier.height(12.dp))
            val canConfirm = picker.levels.any { it.selected != null }
            Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.End) {
                TextButton(onClick = onCancel) { Text(cancelLabel) }
                TextButton(
                    onClick = {
                        val selectedPath = picker.levels.mapNotNull { it.selected }
                        if (selectedPath.isEmpty()) {
                            onCancel()
                        } else {
                            onConfirm(
                                selectedPath.joinToString("-") { it.eventName },
                                selectedPath.joinToString(",") { it.id.toString() },
                            )
                        }
                    },
                    enabled = canConfirm,
                ) { Text(confirmLabel, color = primary) }
            }
        }
    }
}

@Composable
private fun OpLogDayPickerDialog(
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
                Text("‹", color = primary, fontSize = 24.sp, modifier = Modifier.clickable { shiftMonth(-1) }.padding(8.dp))
                Text(
                    "$year-${month.toString().padStart(2, '0')}",
                    color = TextPrimary,
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Medium,
                )
                Text("›", color = primary, fontSize = 24.sp, modifier = Modifier.clickable { shiftMonth(1) }.padding(8.dp))
            }
            Spacer(Modifier.height(8.dp))
            Row(Modifier.fillMaxWidth()) {
                listOf("一", "二", "三", "四", "五", "六", "日").forEach { label ->
                    Text(label, color = TextMuted, fontSize = 12.sp, textAlign = TextAlign.Center, modifier = Modifier.weight(1f))
                }
            }
            cells.chunked(7).forEach { week ->
                Row(Modifier.fillMaxWidth()) {
                    week.forEach { day ->
                        Box(Modifier.weight(1f).height(40.dp), contentAlignment = Alignment.Center) {
                            if (day != null) {
                                val selected = day == selectedDay
                                Box(
                                    Modifier
                                        .size(32.dp)
                                        .background(if (selected) primary else Color.Transparent, CircleShape)
                                        .clickable { selectedDay = day },
                                    contentAlignment = Alignment.Center,
                                ) {
                                    Text(
                                        day.toString(),
                                        color = if (selected) Color.White else TextPrimary,
                                        fontSize = 14.sp,
                                    )
                                }
                            }
                        }
                    }
                }
            }
            Spacer(Modifier.height(12.dp))
            Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.End) {
                TextButton(onClick = onDismiss) { Text(cancelLabel) }
                TextButton(
                    onClick = {
                        onConfirm(
                            buildString {
                                append(year.toString().padStart(4, '0'))
                                append('-')
                                append(month.toString().padStart(2, '0'))
                                append('-')
                                append(selectedDay.toString().padStart(2, '0'))
                            },
                        )
                    },
                ) { Text(confirmLabel, color = primary) }
            }
        }
    }
}
