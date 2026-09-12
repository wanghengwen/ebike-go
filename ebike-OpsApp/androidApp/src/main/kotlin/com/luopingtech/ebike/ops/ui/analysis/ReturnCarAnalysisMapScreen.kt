package com.luopingtech.ebike.ops.ui.analysis

import androidx.compose.foundation.background
import androidx.compose.foundation.border
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
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
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
import com.luopingtech.ebike.ops.domain.analysis.ReturnCarStatusFilter
import com.luopingtech.ebike.ops.ui.map.ReturnCarScatterMapView
import com.luopingtech.ebike.ops.ui.map.SimulatorMapView
import com.luopingtech.ebike.ops.domain.model.MapPin
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import kotlinx.coroutines.launch

/**
 * 还车分布：时间 Tab + 宿主散点地图 + 底部正常/异常/全部。
 */
@Composable
fun ReturnCarAnalysisMapScreen(
    app: OpsApp,
    onClose: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val home by app.homeFeature.state.collectAsState()
    val state by app.returnCarAnalysisFeature.state.collectAsState()
    val detailMap by app.vehicleDetailMapFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    val colors = OpsTheme.colors
    var showFence by remember { mutableStateOf(true) }
    var fitNonce by remember { mutableIntStateOf(0) }

    LaunchedEffect(home.currentArea?.id, state.period) {
        home.currentArea?.id?.takeIf { it.isNotBlank() }?.let {
            app.vehicleDetailMapFeature.loadFence(it)
        }
        app.returnCarAnalysisFeature.load(home.currentArea)
    }

    val isTencent = home.mapProviderKind.equals("tencent", ignoreCase = true) && home.mapReady
    val fencePolygons = if (showFence) detailMap.fence?.all.orEmpty() else emptyList()
    val showNormal = state.statusFilter != ReturnCarStatusFilter.Abnormal
    val showAbnormal = state.statusFilter != ReturnCarStatusFilter.Normal
    val result = state.result

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.White),
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .background(colors.primary)
                .statusBarsPadding(),
        ) {
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(48.dp)
                    .padding(horizontal = 8.dp),
            ) {
                Text(
                    text = "‹ " + t(Str.Back),
                    color = Color.White,
                    fontSize = 15.sp,
                    modifier = Modifier
                        .align(Alignment.CenterStart)
                        .clickable(
                            interactionSource = remember { MutableInteractionSource() },
                            indication = null,
                            onClick = {
                                app.returnCarAnalysisFeature.clear()
                                onClose()
                            },
                        )
                        .padding(8.dp),
                )
                Text(
                    text = t(Str.ReturnCarAnalysis),
                    color = Color.White,
                    fontSize = 17.sp,
                    fontWeight = FontWeight.SemiBold,
                    modifier = Modifier.align(Alignment.Center),
                )
            }
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(bottom = 4.dp),
                horizontalArrangement = Arrangement.SpaceEvenly,
            ) {
                periodTabs(app).forEach { (period, label) ->
                    val selected = state.period == period
                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally,
                        modifier = Modifier
                            .clickable {
                                app.returnCarAnalysisFeature.selectPeriod(period)
                            }
                            .padding(horizontal = 4.dp, vertical = 8.dp),
                    ) {
                        Text(
                            label,
                            color = Color.White.copy(alpha = if (selected) 1f else 0.75f),
                            fontSize = 13.sp,
                            fontWeight = if (selected) FontWeight.SemiBold else FontWeight.Normal,
                        )
                        Spacer(modifier = Modifier.height(4.dp))
                        Box(
                            modifier = Modifier
                                .width(if (selected) 28.dp else 0.dp)
                                .height(2.dp)
                                .background(Color.White),
                        )
                    }
                }
            }
        }

        Box(modifier = Modifier.weight(1f).fillMaxWidth()) {
            if (isTencent) {
                ReturnCarScatterMapView(
                    points = result.points,
                    showNormal = showNormal,
                    showAbnormal = showAbnormal,
                    fencePolygons = fencePolygons,
                    fitNonce = fitNonce,
                    modifier = Modifier.fillMaxSize(),
                )
            } else {
                val pins = state.visiblePoints.mapIndexed { i, p ->
                    MapPin(
                        id = "rc-$i",
                        lat = p.lat,
                        lng = p.lng,
                        title = if (p.abnormal) t(Str.ReturnCarAbnormal) else t(Str.ReturnCarNormal),
                        subtitle = "",
                        restBattery = if (p.abnormal) 10 else 80,
                        ridingState = 0,
                    )
                }
                SimulatorMapView(
                    pins = pins,
                    selectedCarId = null,
                    providerLabel = t(Str.ReturnCarAnalysis),
                    onSelectCarId = {},
                    clusterOverview = false,
                    modifier = Modifier.fillMaxSize(),
                )
            }

            if (state.loading) {
                CircularProgressIndicator(
                    color = colors.primary,
                    modifier = Modifier.align(Alignment.Center),
                )
            }
            state.errorMessage?.let { msg ->
                Text(
                    text = msg,
                    color = Color(0xFFE53935),
                    modifier = Modifier
                        .align(Alignment.TopCenter)
                        .padding(12.dp)
                        .background(Color.White.copy(alpha = 0.9f), RoundedCornerShape(8.dp))
                        .padding(8.dp),
                )
            }
            Text(
                text = t(Str.ReturnCarMapHint),
                color = Color(0xFF666666),
                fontSize = 11.sp,
                modifier = Modifier
                    .align(Alignment.BottomCenter)
                    .padding(bottom = 72.dp)
                    .background(Color.White.copy(alpha = 0.85f), RoundedCornerShape(6.dp))
                    .padding(horizontal = 10.dp, vertical = 4.dp),
            )
            Row(
                modifier = Modifier
                    .align(Alignment.CenterStart)
                    .padding(start = 12.dp),
                horizontalArrangement = Arrangement.spacedBy(0.dp),
            ) {
                Column(
                    modifier = Modifier
                        .background(Color.White, RoundedCornerShape(8.dp))
                        .padding(4.dp),
                ) {
                    ToolChip(t(Str.Refresh)) {
                        scope.launch {
                            app.returnCarAnalysisFeature.load(home.currentArea)
                            home.currentArea?.id?.let { app.vehicleDetailMapFeature.loadFence(it) }
                        }
                    }
                    ToolChip(t(Str.MapToolFence), selected = showFence) {
                        showFence = !showFence
                    }
                    ToolChip(t(Str.MapToolLocate)) { fitNonce += 1 }
                }
            }
        }

        Row(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color.White)
                .padding(horizontal = 16.dp, vertical = 12.dp),
            horizontalArrangement = Arrangement.spacedBy(10.dp),
        ) {
            StatusChip(
                label = "${t(Str.ReturnCarNormal)}(${result.normalCount})",
                selected = state.statusFilter == ReturnCarStatusFilter.Normal,
                selectedBg = Color(0xFF4CAF50),
                selectedBorder = Color(0xFF4CAF50),
                modifier = Modifier.weight(1f),
                onClick = { app.returnCarAnalysisFeature.selectStatus(ReturnCarStatusFilter.Normal) },
            )
            StatusChip(
                label = "${t(Str.ReturnCarAbnormal)}(${result.abnormalCount})",
                selected = state.statusFilter == ReturnCarStatusFilter.Abnormal,
                selectedBg = Color(0xFFE53935),
                selectedBorder = Color(0xFFE53935),
                modifier = Modifier.weight(1f),
                onClick = { app.returnCarAnalysisFeature.selectStatus(ReturnCarStatusFilter.Abnormal) },
            )
            StatusChip(
                label = "${t(Str.ReturnCarAll)}(${result.total})",
                selected = state.statusFilter == ReturnCarStatusFilter.All,
                selectedBg = Color(0xFF2196F3),
                selectedBorder = Color(0xFF2196F3),
                modifier = Modifier.weight(1f),
                onClick = { app.returnCarAnalysisFeature.selectStatus(ReturnCarStatusFilter.All) },
            )
        }
    }
}

@Composable
private fun ToolChip(label: String, selected: Boolean = false, onClick: () -> Unit) {
    Text(
        text = label,
        color = if (selected) Color(0xFF0087FF) else Color(0xFF333333),
        fontSize = 12.sp,
        modifier = Modifier
            .clickable(onClick = onClick)
            .padding(horizontal = 10.dp, vertical = 8.dp),
    )
}

@Composable
private fun StatusChip(
    label: String,
    selected: Boolean,
    selectedBg: Color,
    selectedBorder: Color,
    modifier: Modifier = Modifier,
    onClick: () -> Unit,
) {
    val bg = if (selected) selectedBg else Color.White
    val fg = if (selected) Color.White else selectedBorder
    Text(
        text = label,
        color = fg,
        fontSize = 13.sp,
        fontWeight = FontWeight.Medium,
        textAlign = androidx.compose.ui.text.style.TextAlign.Center,
        modifier = modifier
            .border(1.dp, selectedBorder, RoundedCornerShape(8.dp))
            .background(bg, RoundedCornerShape(8.dp))
            .clickable(onClick = onClick)
            .padding(vertical = 10.dp),
    )
}

private fun periodTabs(app: OpsApp): List<Pair<OfflineOpsPeriod, String>> {
    fun t(key: Str) = app.i18n.t(key)
    return listOf(
        OfflineOpsPeriod.Today to t(Str.OfflineOpsPeriodToday),
        OfflineOpsPeriod.Yesterday to t(Str.OfflineOpsPeriodYesterday),
        OfflineOpsPeriod.ThisWeek to t(Str.OfflineOpsPeriodThisWeek),
        OfflineOpsPeriod.LastWeek to t(Str.OfflineOpsPeriodLastWeek),
        OfflineOpsPeriod.ThisMonth to t(Str.OfflineOpsPeriodThisMonth),
        OfflineOpsPeriod.LastMonth to t(Str.OfflineOpsPeriodLastMonth),
    )
}
