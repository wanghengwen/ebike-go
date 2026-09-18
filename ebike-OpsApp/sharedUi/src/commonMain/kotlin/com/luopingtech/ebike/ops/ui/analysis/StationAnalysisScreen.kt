package com.luopingtech.ebike.ops.ui.analysis

import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.domain.analysis.StationAnalyzeDetail
import com.luopingtech.ebike.ops.domain.analysis.StationAnalyzeItem
import com.luopingtech.ebike.ops.domain.analysis.StationOptStateFilter
import com.luopingtech.ebike.ops.domain.analysis.StationSortOrder
import com.luopingtech.ebike.ops.feature.analysis.StationAnalysisPage
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import kotlinx.coroutines.launch
import kotlin.math.max
import com.luopingtech.ebike.ops.ui.icons.OpsBackChevron

private val PageBg = Color(0xFFF6F7F9)
private val TextDark = Color(0xFF282828)
private val TextMuted = Color(0xFF646464)
private val StatusGreen = Color(0xFF1DBA4F)
private val StatusRed = Color(0xFFFF2222)
private val MetricBlue = Color(0xFF295FCC)
private val MetricBlueBg = Color(0x0D295FCC)
private val AxisGray = Color(0xFFC8C8C8)
private val LineBlue = Color(0xFF3868FF)

@Composable
fun StationAnalysisScreen(
    app: OpsApp,
    onClose: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val home by app.homeFeature.state.collectAsState()
    val state by app.stationAnalysisFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    val colors = OpsTheme.colors

    fun reload() {
        scope.launch {
            app.stationAnalysisFeature.loadTags()
            app.stationAnalysisFeature.refresh(home.currentArea)
        }
    }

    LaunchedEffect(home.currentArea?.id) {
        app.stationAnalysisFeature.loadTags()
        app.stationAnalysisFeature.refresh(home.currentArea)
    }

    LaunchedEffect(state.page, state.selected?.parkingId) {
        if (state.page == StationAnalysisPage.Detail && state.selected != null) {
            app.stationAnalysisFeature.loadDetail(home.currentArea)
        }
    }

    if (state.page == StationAnalysisPage.Detail) {
        StationDetailPane(
            app = app,
            onBack = { app.stationAnalysisFeature.closeDetail() },
        )
        return
    }

    var filterPanel by remember { mutableStateOf<StationFilterPanel?>(null) }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.White),
    ) {
        StationTopBar(
            title = t(Str.StationAnalysisTitle),
            primary = colors.primary,
            onBack = {
                app.stationAnalysisFeature.clear()
                onClose()
            },
        )
        StationFilterBar(
            tagLabel = if (state.selectedTagIds.isEmpty()) {
                t(Str.StationFilterAllTags)
            } else {
                "${t(Str.StationFilterAllTags)}(${state.selectedTagIds.size})"
            },
            statusLabel = when (state.optState) {
                StationOptStateFilter.All -> t(Str.StationFilterAllStatus)
                StationOptStateFilter.Operating -> t(Str.StationStatusOperating)
                StationOptStateFilter.Stopped -> t(Str.StationStatusStopped)
            },
            sortLabel = sortLabel(app, state.sort),
            onTag = { filterPanel = StationFilterPanel.Tags },
            onStatus = { filterPanel = StationFilterPanel.Status },
            onSort = { filterPanel = StationFilterPanel.Sort },
        )
        when {
            state.loading && state.items.isEmpty() -> {
                Column(
                    modifier = Modifier.fillMaxSize(),
                    verticalArrangement = Arrangement.Center,
                    horizontalAlignment = Alignment.CenterHorizontally,
                ) {
                    CircularProgressIndicator(color = colors.primary)
                }
            }
            state.errorMessage != null && state.items.isEmpty() -> {
                Text(
                    text = state.errorMessage.orEmpty(),
                    color = StatusRed,
                    modifier = Modifier.padding(24.dp),
                )
            }
            else -> {
                val listState = rememberLazyListState()
                val shouldLoadMore by remember {
                    derivedStateOf {
                        val info = listState.layoutInfo
                        val last = info.visibleItemsInfo.lastOrNull()?.index ?: 0
                        last >= info.totalItemsCount - 3
                    }
                }
                LaunchedEffect(shouldLoadMore, state.hasMore, state.loadingMore) {
                    if (shouldLoadMore && state.hasMore && !state.loadingMore) {
                        app.stationAnalysisFeature.loadMore(home.currentArea)
                    }
                }
                LazyColumn(
                    state = listState,
                    modifier = Modifier
                        .fillMaxSize()
                        .background(PageBg),
                ) {
                    if (state.items.isEmpty()) {
                        item {
                            Text(
                                t(Str.StationEmptyList),
                                color = TextMuted,
                                modifier = Modifier.padding(24.dp),
                            )
                        }
                    }
                    items(state.items, key = { it.parkingId }) { item ->
                        StationListCard(
                            app = app,
                            item = item,
                            onClick = {
                                app.stationAnalysisFeature.openDetail(item)
                            },
                        )
                    }
                    if (state.loadingMore) {
                        item {
                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(16.dp),
                                horizontalArrangement = Arrangement.Center,
                            ) {
                                CircularProgressIndicator(
                                    modifier = Modifier.size(24.dp),
                                    color = colors.primary,
                                    strokeWidth = 2.dp,
                                )
                            }
                        }
                    }
                }
            }
        }
    }

    filterPanel?.let { panel ->
        StationFilterSheet(
            app = app,
            panel = panel,
            onDismiss = { filterPanel = null },
            onApplied = {
                filterPanel = null
                reload()
            },
        )
    }
}

private enum class StationFilterPanel { Tags, Status, Sort }

@Composable
private fun StationTopBar(title: String, primary: Color, onBack: () -> Unit) {
    com.luopingtech.ebike.ops.ui.navigation.OpsBackHandler(onBack = onBack)
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(primary)
            .statusBarsPadding(),
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(48.dp)
                .padding(horizontal = 8.dp),
        ) {
            OpsBackChevron(
                onClick = onBack,
                modifier = Modifier.align(Alignment.CenterStart),
            )
            Text(
                text = title,
                color = Color.White,
                fontSize = 17.sp,
                fontWeight = FontWeight.SemiBold,
                modifier = Modifier.align(Alignment.Center),
            )
        }
    }
}

@Composable
private fun StationFilterBar(
    tagLabel: String,
    statusLabel: String,
    sortLabel: String,
    onTag: () -> Unit,
    onStatus: () -> Unit,
    onSort: () -> Unit,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .height(44.dp)
            .background(Color.White),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        FilterTrigger(label = tagLabel, modifier = Modifier.weight(1f), onClick = onTag)
        Spacer(modifier = Modifier.width(1.dp).height(16.dp).background(Color(0xFFE5E5E5)))
        FilterTrigger(label = statusLabel, modifier = Modifier.weight(1f), onClick = onStatus)
        Spacer(modifier = Modifier.width(1.dp).height(16.dp).background(Color(0xFFE5E5E5)))
        FilterTrigger(label = sortLabel, modifier = Modifier.weight(1f), onClick = onSort)
    }
    Spacer(modifier = Modifier.fillMaxWidth().height(1.dp).background(Color(0x14000000)))
}

@Composable
private fun FilterTrigger(label: String, modifier: Modifier, onClick: () -> Unit) {
    Row(
        modifier = modifier
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            ),
        horizontalArrangement = Arrangement.Center,
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(label, color = TextDark, fontSize = 13.sp, maxLines = 1, overflow = TextOverflow.Ellipsis)
        Text(" v", color = TextMuted, fontSize = 11.sp)
    }
}

@OptIn(ExperimentalLayoutApi::class)
@Composable
private fun StationFilterSheet(
    app: OpsApp,
    panel: StationFilterPanel,
    onDismiss: () -> Unit,
    onApplied: () -> Unit,
) {
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.stationAnalysisFeature.state.collectAsState()
    val colors = OpsTheme.colors
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color(0x66000000))
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onDismiss,
            ),
    ) {
        Column(
            modifier = Modifier
                .align(Alignment.TopCenter)
                .fillMaxWidth()
                .padding(top = 92.dp)
                .background(Color.White)
                .clickable(
                    interactionSource = remember { MutableInteractionSource() },
                    indication = null,
                    onClick = {},
                )
                .padding(16.dp),
        ) {
            when (panel) {
                StationFilterPanel.Tags -> {
                    FlowRow(
                        horizontalArrangement = Arrangement.spacedBy(8.dp),
                        verticalArrangement = Arrangement.spacedBy(8.dp),
                    ) {
                        state.tags.forEach { tag ->
                            val selected = tag.id in state.selectedTagIds
                            Text(
                                text = tag.name,
                                color = if (selected) colors.primary else TextDark,
                                fontSize = 13.sp,
                                modifier = Modifier
                                    .background(
                                        if (selected) colors.primary.copy(alpha = 0.12f) else PageBg,
                                        RoundedCornerShape(16.dp),
                                    )
                                    .clickable { app.stationAnalysisFeature.toggleTag(tag.id) }
                                    .padding(horizontal = 12.dp, vertical = 8.dp),
                            )
                        }
                    }
                    Spacer(modifier = Modifier.height(12.dp))
                    Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                        Text(
                            t(Str.StationFilterReset),
                            color = TextMuted,
                            modifier = Modifier
                                .clickable {
                                    app.stationAnalysisFeature.setSelectedTags(emptySet())
                                }
                                .padding(8.dp),
                        )
                        Spacer(modifier = Modifier.weight(1f))
                        Text(
                            t(Str.StationFilterDone),
                            color = colors.primary,
                            fontWeight = FontWeight.SemiBold,
                            modifier = Modifier
                                .clickable(onClick = onApplied)
                                .padding(8.dp),
                        )
                    }
                }
                StationFilterPanel.Status -> {
                    listOf(
                        StationOptStateFilter.All to t(Str.StationFilterAllStatus),
                        StationOptStateFilter.Operating to t(Str.StationStatusOperating),
                        StationOptStateFilter.Stopped to t(Str.StationStatusStopped),
                    ).forEach { (value, label) ->
                        Text(
                            text = label,
                            color = if (state.optState == value) colors.primary else TextDark,
                            fontSize = 14.sp,
                            modifier = Modifier
                                .fillMaxWidth()
                                .clickable {
                                    app.stationAnalysisFeature.setOptState(value)
                                    onApplied()
                                }
                                .padding(vertical = 12.dp),
                        )
                    }
                }
                StationFilterPanel.Sort -> {
                    StationSortOrder.entries.forEach { order ->
                        if (order == StationSortOrder.None) return@forEach
                        Text(
                            text = sortLabel(app, order),
                            color = if (state.sort == order) colors.primary else TextDark,
                            fontSize = 14.sp,
                            modifier = Modifier
                                .fillMaxWidth()
                                .clickable {
                                    app.stationAnalysisFeature.setSort(order)
                                    onApplied()
                                }
                                .padding(vertical = 12.dp),
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun StationListCard(
    app: OpsApp,
    item: StationAnalyzeItem,
    onClick: () -> Unit,
) {
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(Color.White)
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            )
            .padding(horizontal = 16.dp, vertical = 20.dp),
    ) {
        Row(modifier = Modifier.fillMaxWidth(), verticalAlignment = Alignment.Top) {
            Text(
                text = item.name,
                color = TextDark,
                fontSize = 14.sp,
                fontWeight = FontWeight.Bold,
                modifier = Modifier.weight(1f),
                maxLines = 2,
                overflow = TextOverflow.Ellipsis,
            )
            Text(
                text = if (item.operating) t(Str.StationStatusOperating) else t(Str.StationStatusStopped),
                color = if (item.operating) StatusGreen else StatusRed,
                fontSize = 14.sp,
                fontWeight = FontWeight.Bold,
            )
        }
        if (item.tags.isNotEmpty()) {
            Spacer(modifier = Modifier.height(8.dp))
            Row(horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                item.tags.take(4).forEach { tag ->
                    Text(
                        text = tag.name,
                        color = MetricBlue,
                        fontSize = 11.sp,
                        modifier = Modifier
                            .background(MetricBlueBg, RoundedCornerShape(4.dp))
                            .padding(horizontal = 6.dp, vertical = 2.dp),
                    )
                }
            }
        }
        Spacer(modifier = Modifier.height(12.dp))
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(6.dp),
        ) {
            MetricCell(
                value = item.canRent.toString(),
                label = t(Str.StationCanRent),
                valueColor = TextDark,
                bg = Color.Transparent,
                modifier = Modifier.weight(1f),
            )
            MetricCell(
                value = item.idle.toString(),
                label = t(Str.StationIdle24h),
                valueColor = MetricBlue,
                bg = MetricBlueBg,
                modifier = Modifier.weight(1f),
            )
            MetricCell(
                value = item.siteOut.toString(),
                label = t(Str.StationSiteOut),
                valueColor = MetricBlue,
                bg = MetricBlueBg,
                modifier = Modifier.weight(1f),
            )
            MetricCell(
                value = item.ddMissOrder.toString(),
                label = t(Str.StationLowBatteryLost),
                valueColor = MetricBlue,
                bg = MetricBlueBg,
                modifier = Modifier.weight(1f),
            )
        }
    }
    Spacer(modifier = Modifier.fillMaxWidth().height(1.dp).background(Color(0xFFE8E8E8)))
}

@Composable
private fun MetricCell(
    value: String,
    label: String,
    valueColor: Color,
    bg: Color,
    modifier: Modifier = Modifier,
) {
    Column(
        modifier = modifier
            .background(bg, RoundedCornerShape(6.dp))
            .padding(6.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text(value, color = valueColor, fontSize = 18.sp, fontWeight = FontWeight.SemiBold)
        Spacer(modifier = Modifier.height(4.dp))
        Text(
            label,
            color = valueColor,
            fontSize = 10.sp,
            maxLines = 2,
            overflow = TextOverflow.Ellipsis,
        )
    }
}

@Composable
private fun StationDetailPane(app: OpsApp, onBack: () -> Unit) {
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.stationAnalysisFeature.state.collectAsState()
    val colors = OpsTheme.colors
    var tab by remember { mutableStateOf(0) }
    val item = state.selected
    val detail = state.detail

    Column(modifier = Modifier.fillMaxSize().background(PageBg)) {
        StationTopBar(
            title = item?.name ?: t(Str.StationAnalysisTitle),
            primary = colors.primary,
            onBack = onBack,
        )
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .height(44.dp)
                .background(Color.White),
        ) {
            listOf(t(Str.StationDetailStats), t(Str.StationDetailAnalysis)).forEachIndexed { i, label ->
                Column(
                    modifier = Modifier
                        .weight(1f)
                        .fillMaxSize()
                        .clickable { tab = i },
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Center,
                ) {
                    Text(
                        label,
                        color = if (tab == i) colors.primary else TextDark,
                        fontWeight = FontWeight.SemiBold,
                        fontSize = 15.sp,
                    )
                    Spacer(modifier = Modifier.height(4.dp))
                    Box(
                        modifier = Modifier
                            .width(if (tab == i) 24.dp else 0.dp)
                            .height(3.dp)
                            .background(if (tab == i) colors.primary else Color.Transparent),
                    )
                }
            }
        }
        when {
            state.detailLoading && detail == null -> {
                Column(
                    modifier = Modifier.fillMaxSize(),
                    verticalArrangement = Arrangement.Center,
                    horizontalAlignment = Alignment.CenterHorizontally,
                ) {
                    CircularProgressIndicator(color = colors.primary)
                }
            }
            state.errorMessage != null && detail == null -> {
                Text(state.errorMessage.orEmpty(), color = StatusRed, modifier = Modifier.padding(24.dp))
            }
            detail != null -> {
                Column(
                    modifier = Modifier
                        .fillMaxSize()
                        .verticalScroll(rememberScrollState())
                        .padding(12.dp),
                    verticalArrangement = Arrangement.spacedBy(12.dp),
                ) {
                    if (tab == 0) {
                        DetailRealtimeRow(app, detail)
                        ChartCard(t(Str.StationChartCanUse), detail.canRentHours)
                        ChartCard(t(Str.StationChartOps), detail.operationHours)
                        ChartCard(t(Str.StationChartBooking), detail.bookingHours)
                        ChartCard(t(Str.StationChartAlarm), detail.alarmHours)
                        ChartCard(t(Str.StationChartFault), detail.faultHours)
                    } else {
                        IdleBarCard(app, detail.idleBuckets)
                        ChartCard(t(Str.StationChartSiteOut), detail.siteOutHours)
                        ChartCard(t(Str.StationChartDdMiss), detail.ddMissHours)
                    }
                }
            }
        }
    }
}

@Composable
private fun DetailRealtimeRow(app: OpsApp, detail: StationAnalyzeDetail) {
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .background(Color.White, RoundedCornerShape(12.dp))
            .padding(16.dp),
        horizontalArrangement = Arrangement.SpaceEvenly,
    ) {
        listOf(
            detail.canRent to t(Str.StationRealtimeCanUse),
            detail.booking to t(Str.StationRealtimeBooking),
            detail.operation to t(Str.StationRealtimeOps),
        ).forEach { (value, label) ->
            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                Text("$value", color = TextDark, fontSize = 22.sp, fontWeight = FontWeight.Bold)
                Text(label, color = TextMuted, fontSize = 12.sp)
            }
        }
    }
}

@Composable
private fun ChartCard(title: String, series: List<Int>) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(Color.White, RoundedCornerShape(12.dp))
            .padding(16.dp),
    ) {
        Text(title, color = TextDark, fontWeight = FontWeight.SemiBold, fontSize = 14.sp)
        Spacer(modifier = Modifier.height(8.dp))
        HourLineChart(series = series, modifier = Modifier.fillMaxWidth().height(140.dp))
    }
}

@Composable
private fun IdleBarCard(app: OpsApp, buckets: List<Int>) {
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val labels = listOf(
        t(Str.StationIdle1to3),
        t(Str.StationIdle3to6),
        t(Str.StationIdle6to12),
        t(Str.StationIdle12to24),
        t(Str.StationIdle24to48),
        t(Str.StationIdleOver48),
    )
    val values = labels.indices.map { buckets.getOrElse(it) { 0 } }
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(Color.White, RoundedCornerShape(12.dp))
            .padding(16.dp),
    ) {
        Text(t(Str.StationChartIdle), color = TextDark, fontWeight = FontWeight.SemiBold, fontSize = 14.sp)
        Spacer(modifier = Modifier.height(8.dp))
        BarChart(values = values, labels = labels, modifier = Modifier.fillMaxWidth().height(160.dp))
    }
}

@Composable
private fun HourLineChart(series: List<Int>, modifier: Modifier = Modifier) {
    val values = series.ifEmpty { List(24) { 0 } }.map { it.toFloat() }
    val maxY = max(1f, values.maxOrNull() ?: 1f)
    Canvas(modifier = modifier) {
        val left = 8f
        val right = size.width - 8f
        val top = 8f
        val bottom = size.height - 8f
        drawLine(AxisGray, Offset(left, top), Offset(left, bottom), strokeWidth = 2f)
        drawLine(AxisGray, Offset(left, bottom), Offset(right, bottom), strokeWidth = 2f)
        if (values.isEmpty()) return@Canvas
        val step = if (values.size == 1) 0f else (right - left) / (values.size - 1)
        val path = Path()
        values.forEachIndexed { i, v ->
            val x = left + step * i
            val y = bottom - (v / maxY) * (bottom - top)
            if (i == 0) path.moveTo(x, y) else path.lineTo(x, y)
        }
        drawPath(path, LineBlue, style = Stroke(width = 3f, cap = StrokeCap.Round))
    }
}

@Composable
private fun BarChart(values: List<Int>, labels: List<String>, modifier: Modifier = Modifier) {
    val maxY = max(1, values.maxOrNull() ?: 1).toFloat()
    Column(modifier = modifier) {
        Canvas(modifier = Modifier.fillMaxWidth().weight(1f)) {
            val n = values.size.coerceAtLeast(1)
            val gap = 8f
            val barW = (size.width - gap * (n + 1)) / n
            values.forEachIndexed { i, v ->
                val h = (v / maxY) * (size.height - 4f)
                val x = gap + i * (barW + gap)
                drawRect(
                    color = LineBlue,
                    topLeft = Offset(x, size.height - h),
                    size = Size(barW, h),
                )
            }
        }
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceEvenly,
        ) {
            labels.forEach {
                Text(it, color = AxisGray, fontSize = 9.sp, maxLines = 1)
            }
        }
    }
}

private fun sortLabel(app: OpsApp, order: StationSortOrder): String {
    fun t(key: Str) = app.i18n.t(key)
    return when (order) {
        StationSortOrder.None -> t(Str.StationFilterSort)
        StationSortOrder.CanRentDesc -> t(Str.StationSortCanRentDesc)
        StationSortOrder.CanRentAsc -> t(Str.StationSortCanRentAsc)
        StationSortOrder.IdleDesc -> t(Str.StationSortIdleDesc)
        StationSortOrder.IdleAsc -> t(Str.StationSortIdleAsc)
        StationSortOrder.SiteOutDesc -> t(Str.StationSortSiteOutDesc)
        StationSortOrder.SiteOutAsc -> t(Str.StationSortSiteOutAsc)
        StationSortOrder.DdMissDesc -> t(Str.StationSortDdMissDesc)
        StationSortOrder.DdMissAsc -> t(Str.StationSortDdMissAsc)
    }
}
