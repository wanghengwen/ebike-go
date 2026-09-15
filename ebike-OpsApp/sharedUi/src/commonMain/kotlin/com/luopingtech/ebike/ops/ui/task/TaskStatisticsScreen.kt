package com.luopingtech.ebike.ops.ui.task

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
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsPeriod
import com.luopingtech.ebike.ops.domain.analysis.TaskStatisticsItem
import com.luopingtech.ebike.ops.domain.analysis.TaskStatisticsKind
import com.luopingtech.ebike.ops.ui.theme.OpsTheme

@Composable
fun TaskStatisticsScreen(
    app: OpsApp,
    kind: TaskStatisticsKind,
    onClose: () -> Unit,
) {
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    @Suppress("UNUSED_VARIABLE")
    val language by app.i18n.languageFlow.collectAsState()
    val home by app.homeFeature.state.collectAsState()
    val state by app.taskStatisticsFeature.state.collectAsState()
    val colors = OpsTheme.colors
    var periodOpen by remember { mutableStateOf(false) }

    LaunchedEffect(kind) {
        app.taskStatisticsFeature.open(kind)
        app.taskStatisticsFeature.refresh(home.currentArea)
    }

    LaunchedEffect(state.period, state.validTab, home.currentArea?.id) {
        if (state.kind == kind) {
            app.taskStatisticsFeature.refresh(home.currentArea)
        }
    }

    val title = when (kind) {
        TaskStatisticsKind.ChangeBattery -> t(Str.ChangeBatteryStats)
        TaskStatisticsKind.MoveCar -> t(Str.MoveCarStats)
        TaskStatisticsKind.Inspection -> t(Str.InspectionStats)
        TaskStatisticsKind.Repair -> t(Str.RepairStats)
    }
    val typeLabel = when (kind) {
        TaskStatisticsKind.ChangeBattery -> t(Str.ChangeBatteryShort)
        TaskStatisticsKind.MoveCar -> t(Str.MoveCarShort)
        TaskStatisticsKind.Inspection -> t(Str.InspectionShort)
        TaskStatisticsKind.Repair -> t(Str.RepairShort)
    }

    Column(modifier = Modifier.fillMaxSize().background(Color.White)) {
        TaskFullscreenTopBar(
            title = title,
            primary = colors.primary,
            onBack = {
                app.taskStatisticsFeature.clear()
                onClose()
            },
            trailingLabel = periodLabel(app, state.period),
            onTrailing = { periodOpen = !periodOpen },
        )
        if (periodOpen) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(Color(0xFFF7F7F7))
                    .padding(vertical = 4.dp),
            ) {
                OfflineOpsPeriod.entries.forEach { p ->
                    Text(
                        text = periodLabel(app, p),
                        color = if (p == state.period) colors.primary else Color(0xFF333333),
                        modifier = Modifier
                            .fillMaxWidth()
                            .clickable {
                                app.taskStatisticsFeature.selectPeriod(p)
                                periodOpen = false
                            }
                            .padding(horizontal = 16.dp, vertical = 12.dp),
                    )
                }
            }
        }
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .height(44.dp)
                .background(Color.White),
        ) {
            listOf(true to t(Str.TaskStatsValid, typeLabel), false to t(Str.TaskStatsInvalid, typeLabel)).forEach { (valid, label) ->
                val selected = state.validTab == valid
                Column(
                    modifier = Modifier
                        .weight(1f)
                        .fillMaxSize()
                        .clickable { app.taskStatisticsFeature.selectValidTab(valid) },
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Center,
                ) {
                    Text(
                        "$label(${if (selected) state.total else "·"})",
                        color = if (selected) colors.primary else Color(0xFF333333),
                        fontWeight = FontWeight.SemiBold,
                        fontSize = 14.sp,
                    )
                    Spacer(modifier = Modifier.height(4.dp))
                    Box(
                        modifier = Modifier
                            .width(if (selected) 36.dp else 0.dp)
                            .height(3.dp)
                            .background(if (selected) colors.primary else Color.Transparent),
                    )
                }
            }
        }
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
                Text(state.errorMessage.orEmpty(), color = Color(0xFFE53935), modifier = Modifier.padding(24.dp))
            }
            else -> {
                val listState = rememberLazyListState()
                val shouldLoadMore by remember {
                    derivedStateOf {
                        val last = listState.layoutInfo.visibleItemsInfo.lastOrNull()?.index ?: 0
                        last >= listState.layoutInfo.totalItemsCount - 2
                    }
                }
                LaunchedEffect(shouldLoadMore, state.hasMore) {
                    if (shouldLoadMore && state.hasMore) {
                        app.taskStatisticsFeature.loadMore(home.currentArea)
                    }
                }
                LazyColumn(state = listState, modifier = Modifier.fillMaxSize().background(Color(0xFFF6F7F9))) {
                    items(state.items, key = { it.id }) { item ->
                        StatsRow(app = app, item = item)
                    }
                    if (state.loadingMore) {
                        item {
                            Row(
                                modifier = Modifier.fillMaxWidth().padding(16.dp),
                                horizontalArrangement = Arrangement.Center,
                            ) {
                                CircularProgressIndicator(
                                    modifier = Modifier.size(24.dp),
                                    strokeWidth = 2.dp,
                                    color = colors.primary,
                                )
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun StatsRow(app: OpsApp, item: TaskStatisticsItem) {
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 12.dp, vertical = 6.dp)
            .background(Color.White, RoundedCornerShape(10.dp))
            .padding(14.dp),
    ) {
        Text("${t(Str.VehicleId)}: ${item.carId}", color = Color(0xFF333333), fontWeight = FontWeight.SemiBold)
        Spacer(modifier = Modifier.height(6.dp))
        when (item.kind) {
            TaskStatisticsKind.ChangeBattery -> {
                Text("换前 ${item.restBatteryBefore ?: 0}% → 换后 ${item.restBatteryAfter ?: 0}%", color = Color(0xFF666666), fontSize = 12.sp)
                Text("开仓 ${item.openBatBoxTime.ifBlank { "-" }}", color = Color(0xFF666666), fontSize = 12.sp)
                Text("关仓 ${item.closeBatBoxTime.ifBlank { "-" }}", color = Color(0xFF666666), fontSize = 12.sp)
            }
            TaskStatisticsKind.MoveCar -> {
                if (item.sourceLabel.isNotBlank()) {
                    Text("${t(Str.TaskSource)}: ${item.sourceLabel}", color = Color(0xFF666666), fontSize = 12.sp)
                }
                if (item.distance.isNotBlank()) {
                    Text("距离: ${item.distance}", color = Color(0xFF666666), fontSize = 12.sp)
                }
                Text("${item.startTime.ifBlank { "-" }} → ${item.finishTime.ifBlank { "-" }}", color = Color(0xFF666666), fontSize = 12.sp)
            }
            else -> {
                Text("${item.startTime.ifBlank { "-" }} → ${item.finishTime.ifBlank { "-" }}", color = Color(0xFF666666), fontSize = 12.sp)
            }
        }
        Spacer(modifier = Modifier.height(4.dp))
        Text("耗时: ${item.durationText.ifBlank { "-" }}", color = Color(0xFF999999), fontSize = 12.sp)
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
