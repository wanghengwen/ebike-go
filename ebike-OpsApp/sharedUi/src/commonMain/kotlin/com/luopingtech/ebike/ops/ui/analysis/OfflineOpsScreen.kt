package com.luopingtech.ebike.ops.ui.analysis

import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
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
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
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
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsPeriod
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsTab
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsTimeRanges
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsTrendGrain
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsTrendPoint
import com.luopingtech.ebike.ops.domain.analysis.PieSlice
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import kotlinx.coroutines.launch
import kotlin.math.max

private val PageBg = Color(0xFFF6F7F9)
private val CardWhite = Color.White
private val CellGray = Color(0xFFF6F6F6)
private val TextDark = Color(0xFF333333)
private val TextMuted = Color(0xFF646464)
private val ValidGreen = Color(0xFF63D144)
private val InvalidRed = Color(0xFFFF5936)
private val LinkBlue = Color(0xFF0087FF)
private val LineBlue = Color(0xFF3868FF)
private val AxisGray = Color(0xFFC8C8C8)

@Composable
fun OfflineOpsScreen(
    app: OpsApp,
    onClose: () -> Unit,
    mineOnly: Boolean = false,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val home by app.homeFeature.state.collectAsState()
    val state by app.offlineOpsFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    val colors = OpsTheme.colors

    val opPin = if (mineOnly) app.authFeature.state.value.session?.userId.orEmpty() else null

    fun reload() {
        scope.launch { app.offlineOpsFeature.load(home.currentArea, mineOnly = mineOnly, opPin = opPin) }
    }

    LaunchedEffect(home.currentArea?.id, state.tab, state.period, state.trendGrain, mineOnly, opPin) {
        app.offlineOpsFeature.load(home.currentArea, mineOnly = mineOnly, opPin = opPin)
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(PageBg),
    ) {
        VcdTopBar(
            title = if (mineOnly) t(Str.MyTaskTool) else t(Str.OfflineOperation),
            onBack = {
                app.offlineOpsFeature.clear()
                onClose()
            },
        )
        OfflineOpsTabs(
            selected = state.tab,
            labels = listOf(
                OfflineOpsTab.ChangeBattery to t(Str.ChangeBatteryShort),
                OfflineOpsTab.MoveCar to t(Str.MoveCarShort),
                OfflineOpsTab.Inspection to t(Str.InspectionShort),
                OfflineOpsTab.Repair to t(Str.RepairShort),
            ),
            onSelect = {
                app.offlineOpsFeature.selectTab(it)
            },
        )

        if (state.loading && state.dashboard.updatedAtText.isBlank()) {
            Column(
                modifier = Modifier.fillMaxSize(),
                verticalArrangement = Arrangement.Center,
                horizontalAlignment = Alignment.CenterHorizontally,
            ) {
                CircularProgressIndicator(color = colors.primary)
            }
        } else if (state.errorMessage != null && state.dashboard.updatedAtText.isBlank()) {
            Text(
                text = state.errorMessage.orEmpty(),
                color = InvalidRed,
                modifier = Modifier.padding(24.dp),
            )
        } else {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .verticalScroll(rememberScrollState())
                    .padding(bottom = 24.dp),
            ) {
                val tabLabel = when (state.tab) {
                    OfflineOpsTab.ChangeBattery -> t(Str.ChangeBatteryShort)
                    OfflineOpsTab.MoveCar -> t(Str.MoveCarShort)
                    OfflineOpsTab.Inspection -> t(Str.InspectionShort)
                    OfflineOpsTab.Repair -> t(Str.RepairShort)
                }

                StatsCard(
                    app = app,
                    tab = state.tab,
                    period = state.period,
                    onPeriod = {
                        app.offlineOpsFeature.selectPeriod(it)
                    },
                )
                TrendCard(
                    app = app,
                    title = t(Str.OfflineOpsTrend, tabLabel),
                    grain = state.trendGrain,
                    points = state.dashboard.trendPoints,
                    updatedAt = state.dashboard.updatedAtText,
                    onGrain = { app.offlineOpsFeature.selectTrendGrain(it) },
                )
                if (!mineOnly && state.tab != OfflineOpsTab.Repair) {
                    RankCard(
                        app = app,
                        title = t(Str.OfflineOpsRank, tabLabel),
                        period = state.period,
                        updatedAt = state.dashboard.updatedAtText,
                        ranks = state.dashboard.ranks,
                        onPeriod = { app.offlineOpsFeature.selectPeriod(it) },
                    )
                }
            }
        }
    }
}

@Composable
private fun OfflineOpsTabs(
    selected: OfflineOpsTab,
    labels: List<Pair<OfflineOpsTab, String>>,
    onSelect: (OfflineOpsTab) -> Unit,
) {
    val colors = OpsTheme.colors
    Column(modifier = Modifier.fillMaxWidth().background(CardWhite)) {
        Row(modifier = Modifier.fillMaxWidth().height(54.dp)) {
            labels.forEach { (tab, label) ->
                val on = selected == tab
                Column(
                    modifier = Modifier
                        .weight(1f)
                        .fillMaxSize()
                        .clickable(
                            interactionSource = remember { MutableInteractionSource() },
                            indication = null,
                            onClick = { onSelect(tab) },
                        ),
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Center,
                ) {
                    Text(
                        text = label,
                        color = if (on) colors.primary else Color(0xFF282828),
                        fontSize = 15.sp,
                        fontWeight = FontWeight.SemiBold,
                    )
                    Spacer(modifier = Modifier.height(2.dp))
                    Box(
                        modifier = Modifier
                            .width(if (on) 24.dp else 0.dp)
                            .height(4.dp)
                            .background(
                                if (on) colors.primary else Color.Transparent,
                                RoundedCornerShape(2.dp),
                            ),
                    )
                }
            }
        }
        Box(modifier = Modifier.fillMaxWidth().height(1.dp).background(Color(0x4D9E9E9E)))
    }
}

@Composable
private fun StatsCard(
    app: OpsApp,
    tab: OfflineOpsTab,
    period: OfflineOpsPeriod,
    onPeriod: (OfflineOpsPeriod) -> Unit,
) {
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.offlineOpsFeature.state.collectAsState()
    val d = state.dashboard
    val analyze = d.analyze
    WhiteCard {
        CardHeader(
            title = t(Str.OfflineOpsStats),
            updatedAt = t(Str.OfflineOpsUpdatedAt, d.updatedAtText),
            menuLabel = periodLabel(app, period),
            menuItems = periodMenu(app),
            onSelect = { onPeriod(it) },
        )
        Spacer(modifier = Modifier.height(10.dp))
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .height(41.dp)
                .background(CellGray, RoundedCornerShape(10.dp))
                .padding(horizontal = 12.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(t(Str.OfflineOpsTotal), color = TextDark, fontSize = 13.sp, fontWeight = FontWeight.Medium)
            Spacer(modifier = Modifier.width(12.dp))
            Text(t(Str.OfflineOpsValid), color = TextMuted, fontSize = 12.sp)
            Text(" ${d.totalEffective}", color = ValidGreen, fontSize = 16.sp, fontWeight = FontWeight.Bold)
            Spacer(modifier = Modifier.width(12.dp))
            Text(t(Str.OfflineOpsInvalid), color = TextMuted, fontSize = 12.sp)
            Text(" ${d.totalVoid}", color = InvalidRed, fontSize = 16.sp, fontWeight = FontWeight.Bold)
        }
        Spacer(modifier = Modifier.height(10.dp))
        MetricGrid(metrics = metricsFor(app, tab))
        when (tab) {
            OfflineOpsTab.MoveCar -> if (analyze.movePie.isNotEmpty()) {
                Spacer(modifier = Modifier.height(12.dp))
                Text(t(Str.OfflineOpsMoveComposition), color = TextDark, fontWeight = FontWeight.SemiBold, fontSize = 14.sp)
                Spacer(modifier = Modifier.height(8.dp))
                SimplePieChart(slices = analyze.movePie, modifier = Modifier.fillMaxWidth().height(180.dp))
            }
            OfflineOpsTab.Repair -> {
                if (analyze.repairPartPie.isNotEmpty()) {
                    Spacer(modifier = Modifier.height(12.dp))
                    Text(t(Str.OfflineOpsRepairParts), color = TextDark, fontWeight = FontWeight.SemiBold, fontSize = 14.sp)
                    Spacer(modifier = Modifier.height(8.dp))
                    SimplePieChart(slices = analyze.repairPartPie, modifier = Modifier.fillMaxWidth().height(160.dp))
                }
                if (analyze.repairSourcePie.isNotEmpty()) {
                    Spacer(modifier = Modifier.height(12.dp))
                    Text(t(Str.OfflineOpsRepairSource), color = TextDark, fontWeight = FontWeight.SemiBold, fontSize = 14.sp)
                    Spacer(modifier = Modifier.height(8.dp))
                    SimplePieChart(slices = analyze.repairSourcePie, modifier = Modifier.fillMaxWidth().height(160.dp))
                }
            }
            else -> Unit
        }
    }
}

@Composable
private fun TrendCard(
    app: OpsApp,
    title: String,
    grain: OfflineOpsTrendGrain,
    points: List<OfflineOpsTrendPoint>,
    updatedAt: String,
    onGrain: (OfflineOpsTrendGrain) -> Unit,
) {
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    WhiteCard {
        CardHeader(
            title = title,
            updatedAt = t(Str.OfflineOpsUpdatedAt, updatedAt),
            menuLabel = grainLabel(app, grain),
            menuItems = grainMenu(app),
            onSelect = { onGrain(it) },
        )
        Spacer(modifier = Modifier.height(8.dp))
        Text(t(Str.OfflineOpsTrendUnit), color = AxisGray, fontSize = 11.sp)
        Spacer(modifier = Modifier.height(4.dp))
        SimpleLineChart(
            points = points,
            modifier = Modifier
                .fillMaxWidth()
                .height(200.dp),
        )
    }
}

@Composable
private fun RankCard(
    app: OpsApp,
    title: String,
    period: OfflineOpsPeriod,
    updatedAt: String,
    ranks: List<com.luopingtech.ebike.ops.domain.analysis.OfflineOpsRankRow>,
    onPeriod: (OfflineOpsPeriod) -> Unit,
) {
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    WhiteCard {
        CardHeader(
            title = title,
            updatedAt = t(Str.OfflineOpsUpdatedAt, updatedAt),
            menuLabel = periodLabel(app, period),
            menuItems = periodMenu(app),
            onSelect = { onPeriod(it) },
        )
        Spacer(modifier = Modifier.height(8.dp))
        if (ranks.isEmpty()) {
            Text(
                t(Str.OfflineOpsEmptyRank),
                color = TextMuted,
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(vertical = 28.dp),
                fontSize = 13.sp,
            )
        } else {
            ranks.forEachIndexed { index, row ->
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(vertical = 10.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Box(
                        modifier = Modifier
                            .size(22.dp)
                            .background(CellGray, RoundedCornerShape(4.dp)),
                        contentAlignment = Alignment.Center,
                    ) {
                        Text("${index + 1}", fontSize = 11.sp, color = TextMuted)
                    }
                    Spacer(modifier = Modifier.width(10.dp))
                    Text(row.opName, color = LinkBlue, fontSize = 15.sp, fontWeight = FontWeight.SemiBold, modifier = Modifier.weight(1f))
                    Text("${t(Str.OfflineOpsValid)}${row.effectiveNum}", color = TextMuted, fontSize = 12.sp)
                    Spacer(modifier = Modifier.width(8.dp))
                    Text("${t(Str.OfflineOpsInvalid)}${row.voidNum}", color = TextMuted, fontSize = 12.sp)
                    Spacer(modifier = Modifier.width(10.dp))
                    Text("${row.countNum}", color = TextDark, fontSize = 16.sp, fontWeight = FontWeight.Bold)
                }
                if (index != ranks.lastIndex) {
                    Box(modifier = Modifier.fillMaxWidth().height(1.dp).background(Color(0xFFE5E5E5)))
                }
            }
        }
    }
}

@Composable
private fun <T> CardHeader(
    title: String,
    updatedAt: String,
    menuLabel: String,
    menuItems: List<Pair<T, String>>,
    onSelect: (T) -> Unit,
) {
    var open by remember { mutableStateOf(false) }
    Column(modifier = Modifier.fillMaxWidth()) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(title, color = TextDark, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
            Spacer(modifier = Modifier.width(6.dp))
            Text(
                updatedAt,
                color = TextMuted,
                fontSize = 11.sp,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
                modifier = Modifier.weight(1f),
            )
            Text(
                text = "$menuLabel ▾",
                color = LinkBlue,
                fontSize = 13.sp,
                modifier = Modifier.clickable(
                    interactionSource = remember { MutableInteractionSource() },
                    indication = null,
                    onClick = { open = !open },
                ),
            )
        }
        if (open) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 6.dp)
                    .background(CardWhite, RoundedCornerShape(8.dp))
                    .padding(vertical = 4.dp),
            ) {
                menuItems.forEach { (value, label) ->
                    Text(
                        text = label,
                        color = if (label == menuLabel) LinkBlue else TextDark,
                        fontSize = 13.sp,
                        modifier = Modifier
                            .fillMaxWidth()
                            .clickable {
                                onSelect(value)
                                open = false
                            }
                            .padding(horizontal = 12.dp, vertical = 10.dp),
                    )
                }
            }
        }
    }
}

@Composable
private fun WhiteCard(content: @Composable () -> Unit) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(start = 12.dp, end = 12.dp, top = 8.dp)
            .background(CardWhite, RoundedCornerShape(16.dp))
            .padding(16.dp),
    ) {
        content()
    }
}

@Composable
private fun MetricGrid(metrics: List<Pair<String, String>>) {
    Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
        metrics.chunked(2).forEach { row ->
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(10.dp),
            ) {
                row.forEach { (label, value) ->
                    Column(
                        modifier = Modifier
                            .weight(1f)
                            .height(41.dp)
                            .background(CellGray, RoundedCornerShape(10.dp))
                            .padding(horizontal = 10.dp),
                        verticalArrangement = Arrangement.Center,
                    ) {
                        Row(verticalAlignment = Alignment.Bottom) {
                            Text(label, color = TextMuted, fontSize = 11.sp, maxLines = 1, overflow = TextOverflow.Ellipsis, modifier = Modifier.weight(1f, fill = false))
                            Spacer(modifier = Modifier.width(4.dp))
                            Text(value, color = TextDark, fontSize = 13.sp, fontWeight = FontWeight.SemiBold, maxLines = 1)
                        }
                    }
                }
                if (row.size == 1) Spacer(modifier = Modifier.weight(1f))
            }
        }
    }
}

@Composable
private fun metricsFor(app: OpsApp, tab: OfflineOpsTab): List<Pair<String, String>> {
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val a = app.offlineOpsFeature.state.value.dashboard.analyze
    val validRate = run {
        val sum = a.validCount + a.invalidCount
        if (sum <= 0) "0%" else "${((a.validCount * 100.0) / sum).let { (it * 100).toInt() / 100.0 }}%"
    }
    val duration = OfflineOpsTimeRanges.formatDurationSeconds(a.taskTime.now)
    return when (tab) {
        OfflineOpsTab.ChangeBattery -> listOf(
            t(Str.OfflineOpsAvgSwapDuration) to duration,
            t(Str.OfflineOpsSwapValidRate) to validRate,
            t(Str.OfflineOpsLowBatteryLost) to t(Str.UnitOrderCount, a.order.now.toInt()),
            t(Str.OfflineOpsAvgRestBattery) to "${a.battery.now.toInt()}%",
        )
        OfflineOpsTab.MoveCar -> listOf(
            t(Str.OfflineOpsAvgMoveDuration) to duration,
            t(Str.OfflineOpsMoveValidRate) to validRate,
            t(Str.OfflineOpsAvgMoveDistance) to t(Str.UnitKilometers, a.distance.now),
            t(Str.OfflineOpsMoveOrderEffect) to OfflineOpsTimeRanges.formatDurationSeconds(a.order.now),
            t(Str.OfflineOpsMoveOrder12h) to t(Str.UnitOrderCount, a.orderCountOf12.now.toInt()),
        )
        OfflineOpsTab.Inspection -> listOf(
            t(Str.OfflineOpsAvgInspectDuration) to duration,
            t(Str.OfflineOpsAvgInspectDistance) to t(Str.UnitKilometers, a.distance.now),
            t(Str.OfflineOpsInspectOrderEffect) to OfflineOpsTimeRanges.formatDurationSeconds(a.order.now),
            t(Str.OfflineOpsInspectValidRate) to validRate,
        )
        OfflineOpsTab.Repair -> listOf(
            t(Str.OfflineOpsAvgRepairDuration) to duration,
        )
    }
}

private fun periodLabel(app: OpsApp, period: OfflineOpsPeriod): String {
    fun t(key: Str) = app.i18n.t(key)
    return when (period) {
        OfflineOpsPeriod.Today -> t(Str.OfflineOpsPeriodToday)
        OfflineOpsPeriod.Yesterday -> t(Str.OfflineOpsPeriodYesterday)
        OfflineOpsPeriod.ThisWeek -> t(Str.OfflineOpsPeriodThisWeek)
        OfflineOpsPeriod.LastWeek -> t(Str.OfflineOpsPeriodLastWeek)
        OfflineOpsPeriod.ThisMonth -> t(Str.OfflineOpsPeriodThisMonth)
        OfflineOpsPeriod.LastMonth -> t(Str.OfflineOpsPeriodLastMonth)
    }
}

private fun grainLabel(app: OpsApp, grain: OfflineOpsTrendGrain): String {
    fun t(key: Str) = app.i18n.t(key)
    return when (grain) {
        OfflineOpsTrendGrain.Daily -> t(Str.OfflineOpsGrainDaily)
        OfflineOpsTrendGrain.Weekly -> t(Str.OfflineOpsGrainWeekly)
        OfflineOpsTrendGrain.Monthly -> t(Str.OfflineOpsGrainMonthly)
    }
}

private fun periodMenu(app: OpsApp): List<Pair<OfflineOpsPeriod, String>> =
    OfflineOpsPeriod.entries.map { it to periodLabel(app, it) }

private fun grainMenu(app: OpsApp): List<Pair<OfflineOpsTrendGrain, String>> =
    OfflineOpsTrendGrain.entries.map { it to grainLabel(app, it) }

@Composable
private fun SimpleLineChart(
    points: List<OfflineOpsTrendPoint>,
    modifier: Modifier = Modifier,
) {
    val values = points.map { it.count.toFloat() }
    val maxY = max(1f, values.maxOrNull() ?: 1f)
    Canvas(modifier = modifier.padding(start = 8.dp, end = 8.dp, bottom = 18.dp, top = 8.dp)) {
        val w = size.width
        val h = size.height
        val left = 28f
        val bottom = h - 4f
        val top = 8f
        val right = w - 8f
        // axes
        drawLine(AxisGray, Offset(left, top), Offset(left, bottom), strokeWidth = 2f)
        drawLine(AxisGray, Offset(left, bottom), Offset(right, bottom), strokeWidth = 2f)
        if (points.isEmpty()) return@Canvas
        val stepX = if (points.size == 1) 0f else (right - left) / (points.size - 1)
        val path = Path()
        points.forEachIndexed { i, p ->
            val x = left + stepX * i
            val y = bottom - (p.count / maxY) * (bottom - top)
            if (i == 0) path.moveTo(x, y) else path.lineTo(x, y)
        }
        drawPath(path, LineBlue, style = Stroke(width = 3f, cap = StrokeCap.Round))
        points.forEachIndexed { i, p ->
            val x = left + stepX * i
            val y = bottom - (p.count / maxY) * (bottom - top)
            drawCircle(LineBlue, radius = 6f, center = Offset(x, y))
            drawCircle(Color.White, radius = 3f, center = Offset(x, y))
        }
    }
    // x labels under chart
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 28.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
    ) {
        points.forEachIndexed { i, p ->
            if (points.size <= 8 || i == 0 || i == points.lastIndex || i == points.size / 2) {
                Text(p.label, color = AxisGray, fontSize = 9.sp, maxLines = 1)
            } else {
                Spacer(modifier = Modifier.width(1.dp))
            }
        }
    }
}

@Composable
private fun SimplePieChart(
    slices: List<PieSlice>,
    modifier: Modifier = Modifier,
) {
    val total = slices.sumOf { it.value }.let { if (it <= 0) 1.0 else it }
    val palette = listOf(
        Color(0xFF5470C6), Color(0xFF91CC75), Color(0xFFFAC858),
        Color(0xFFEE6666), Color(0xFF73C0DE), Color(0xFF3BA272),
    )
    Row(modifier = modifier, verticalAlignment = Alignment.CenterVertically) {
        Canvas(modifier = Modifier.size(140.dp)) {
            var start = -90f
            slices.forEachIndexed { i, slice ->
                val sweep = ((slice.value / total) * 360.0).toFloat()
                drawArc(
                    color = palette[i % palette.size],
                    startAngle = start,
                    sweepAngle = max(0.5f, sweep),
                    useCenter = true,
                )
                start += sweep
            }
        }
        Spacer(modifier = Modifier.width(12.dp))
        Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
            slices.forEachIndexed { i, slice ->
                val pct = ((slice.value * 100.0) / total).toInt()
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Box(
                        modifier = Modifier
                            .size(8.dp)
                            .background(palette[i % palette.size], RoundedCornerShape(2.dp)),
                    )
                    Spacer(modifier = Modifier.width(6.dp))
                    Text(
                        "${slice.name} ${slice.value.toInt()} ($pct%)",
                        color = TextMuted,
                        fontSize = 11.sp,
                    )
                }
            }
        }
    }
}
