package com.luopingtech.ebike.ops.ui.analysis

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.domain.analysis.VehicleConditionBuckets
import com.luopingtech.ebike.ops.feature.analysis.VcdPage
import com.luopingtech.ebike.ops.feature.analysis.VcdTab
import com.luopingtech.ebike.ops.ui.theme.OpsTheme

/**
 * 车况分布列表 —— 对齐遗留 Flutter 页：品牌顶栏、分段下划线、灰底色条 + 旁注数量。
 */
@Composable
fun VehicleConditionDistributionScreen(
    app: OpsApp,
    onClose: () -> Unit,
    onOpenMap: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val home by app.homeFeature.state.collectAsState()
    val state by app.vehicleConditionDistributionFeature.state.collectAsState()

    LaunchedEffect(home.currentArea?.id) {
        app.vehicleConditionDistributionFeature.load(home.currentArea)
    }

    LaunchedEffect(state.page) {
        if (state.page == VcdPage.Map) {
            onOpenMap()
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.White),
    ) {
        VcdTopBar(
            title = t(Str.VehicleDistribution),
            onBack = {
                app.vehicleConditionDistributionFeature.clear()
                onClose()
            },
        )
        VcdSegmentTabs(
            batteryLabel = t(Str.BatteryDistribution),
            idleLabel = t(Str.IdleDurationDistribution),
            selected = state.tab,
            onSelect = { app.vehicleConditionDistributionFeature.selectTab(it) },
        )
        if (state.loading) {
            VcdLoadingPane()
        } else if (state.errorMessage != null) {
            Text(
                text = state.errorMessage.orEmpty(),
                color = Color(0xFFE53935),
                modifier = Modifier.padding(24.dp),
            )
        } else {
            LazyColumn(
                modifier = Modifier
                    .fillMaxSize()
                    .background(Color.White),
            ) {
                itemsIndexed(
                    items = state.listBuckets,
                    key = { _, bucket -> "${state.tab}-${bucket.index}" },
                ) { index, bucket ->
                    VcdBucketRow(
                        label = t(bucket.labelKey),
                        countLabel = t(Str.VcdCountPercent, bucket.count, bucket.percent),
                        percent = bucket.percent,
                        colorArgb = bucket.colorArgb,
                        useIdleGradient = state.tab == VcdTab.Idle,
                        topPad = if (index == 0) 20.dp else 10.dp,
                        onClick = {
                            app.vehicleConditionDistributionFeature.openMap(bucket.index)
                        },
                    )
                }
                item { Spacer(modifier = Modifier.height(24.dp)) }
            }
        }
    }
}

@Composable
private fun VcdLoadingPane() {
    Column(
        modifier = Modifier.fillMaxSize(),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
    ) {
        CircularProgressIndicator(color = OpsTheme.colors.primary)
    }
}

@Composable
fun VcdTopBar(
    title: String,
    onBack: () -> Unit,
) {
    val colors = OpsTheme.colors
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .background(colors.primary)
            .statusBarsPadding()
            .height(48.dp),
    ) {
        Text(
            text = "‹",
            color = colors.onPrimary,
            fontSize = 28.sp,
            fontWeight = FontWeight.Light,
            modifier = Modifier
                .align(Alignment.CenterStart)
                .clickable(
                    interactionSource = remember { MutableInteractionSource() },
                    indication = null,
                    onClick = onBack,
                )
                .padding(horizontal = 16.dp),
        )
        Text(
            text = title,
            color = colors.onPrimary,
            fontSize = 18.sp,
            fontWeight = FontWeight.Medium,
            modifier = Modifier.align(Alignment.Center),
        )
    }
}

@Composable
private fun VcdSegmentTabs(
    batteryLabel: String,
    idleLabel: String,
    selected: VcdTab,
    onSelect: (VcdTab) -> Unit,
) {
    Column(modifier = Modifier.fillMaxWidth()) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .height(54.dp)
                .background(Color.White),
        ) {
            VcdSegmentTab(
                label = batteryLabel,
                selected = selected == VcdTab.Battery,
                modifier = Modifier.weight(1f),
                onClick = { onSelect(VcdTab.Battery) },
            )
            VcdSegmentTab(
                label = idleLabel,
                selected = selected == VcdTab.Idle,
                modifier = Modifier.weight(1f),
                onClick = { onSelect(VcdTab.Idle) },
            )
        }
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(1.dp)
                .background(Color(0xFF000000).copy(alpha = 0.12f)),
        )
    }
}

@Composable
private fun VcdSegmentTab(
    label: String,
    selected: Boolean,
    modifier: Modifier = Modifier,
    onClick: () -> Unit,
) {
    val colors = OpsTheme.colors
    Column(
        modifier = modifier
            .fillMaxHeight()
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            ),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
    ) {
        Text(
            text = label,
            color = if (selected) colors.primary else Color(0xFF282828),
            fontSize = 15.sp,
            fontWeight = FontWeight.SemiBold,
        )
        Spacer(modifier = Modifier.height(2.dp))
        Box(
            modifier = Modifier
                .width(if (selected) 22.dp else 0.dp)
                .height(4.dp)
                .clip(RoundedCornerShape(2.dp))
                .background(if (selected) colors.primary else Color.Transparent),
        )
    }
}

@Composable
private fun VcdBucketRow(
    label: String,
    countLabel: String,
    percent: Int,
    colorArgb: Long,
    useIdleGradient: Boolean,
    topPad: Dp,
    onClick: () -> Unit,
) {
    val barColor = Color(colorArgb)
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            )
            .padding(start = 16.dp, end = 16.dp, top = topPad),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            text = label,
            fontSize = 11.sp,
            fontWeight = FontWeight.Medium,
            color = Color(0xFF333333),
            textAlign = TextAlign.Start,
            modifier = Modifier.width(64.dp),
        )
        Spacer(modifier = Modifier.width(6.dp))
        BoxWithConstraints(
            modifier = Modifier
                .weight(1f)
                .height(24.dp)
                .clip(RoundedCornerShape(2.dp))
                .background(Color(0xFFEBEBEB)),
        ) {
            val trackWidth = maxWidth
            val fillWidth = trackWidth * (percent.coerceIn(0, 100) / 100f)
            val remain = trackWidth - fillWidth
            val textEnough = remain > 72.dp
            val fillBrush = if (useIdleGradient) {
                Brush.horizontalGradient(
                    listOf(Color(0xFF188DF0), Color(VehicleConditionBuckets.idleGradientEndArgb)),
                )
            } else {
                null
            }

            if (textEnough) {
                Row(
                    modifier = Modifier.fillMaxSize(),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    if (percent > 0) {
                        val fillMod = Modifier
                            .width(fillWidth.coerceAtLeast(2.dp))
                            .fillMaxHeight()
                        Box(
                            modifier = if (fillBrush != null) {
                                fillMod.background(fillBrush, RoundedCornerShape(2.dp))
                            } else {
                                fillMod.background(barColor, RoundedCornerShape(2.dp))
                            },
                        )
                        Spacer(modifier = Modifier.width(8.dp))
                    }
                    Text(
                        text = countLabel,
                        fontSize = 11.sp,
                        fontWeight = FontWeight.Medium,
                        color = Color(0xFF333333),
                    )
                }
            } else {
                Box(modifier = Modifier.fillMaxSize()) {
                    if (percent > 0) {
                        val fullMod = Modifier.fillMaxSize()
                        Box(
                            modifier = if (fillBrush != null) {
                                fullMod.background(fillBrush)
                            } else {
                                fullMod.background(barColor)
                            },
                        )
                    }
                    Text(
                        text = countLabel,
                        fontSize = 11.sp,
                        fontWeight = FontWeight.Medium,
                        color = Color(0xFF333333),
                        modifier = Modifier.align(Alignment.Center),
                    )
                }
            }
        }
    }
}
