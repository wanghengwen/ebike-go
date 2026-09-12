package com.luopingtech.ebike.ops.ui.task

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import com.luopingtech.ebike.ops.ui.theme.OpsToolIcon

@Composable
fun TaskFullscreenTopBar(
    title: String,
    primary: Color,
    onBack: () -> Unit,
    trailingLabel: String? = null,
    onTrailing: (() -> Unit)? = null,
) {
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
            Text(
                text = "<",
                color = Color.White,
                fontSize = 22.sp,
                modifier = Modifier
                    .align(Alignment.CenterStart)
                    .clickable(
                        interactionSource = remember { MutableInteractionSource() },
                        indication = null,
                        onClick = onBack,
                    )
                    .padding(8.dp),
            )
            Text(
                text = title,
                color = Color.White,
                fontSize = 17.sp,
                fontWeight = FontWeight.SemiBold,
                modifier = Modifier.align(Alignment.Center),
            )
            if (trailingLabel != null && onTrailing != null) {
                Text(
                    text = trailingLabel,
                    color = Color.White,
                    fontSize = 14.sp,
                    modifier = Modifier
                        .align(Alignment.CenterEnd)
                        .clickable(
                            interactionSource = remember { MutableInteractionSource() },
                            indication = null,
                            onClick = onTrailing,
                        )
                        .padding(8.dp),
                )
            }
        }
    }
}

/**
 * 对齐遗留 HomeMapControlView：缩放 +/-、刷新、定位、更多（停车区/卫星/车辆详情）。
 * 指南针由腾讯地图 SDK 自带。
 */
@Composable
fun TaskMapSideTools(
    refreshLabel: String,
    locateLabel: String,
    moreLabel: String,
    parkingLabel: String,
    satelliteLabel: String,
    detailLabel: String,
    onZoomIn: () -> Unit,
    onZoomOut: () -> Unit,
    onRefresh: () -> Unit,
    onLocate: () -> Unit,
    parkingSelected: Boolean,
    satelliteSelected: Boolean,
    detailSelected: Boolean,
    onToggleParking: () -> Unit,
    onToggleSatellite: () -> Unit,
    onToggleDetail: () -> Unit,
    modifier: Modifier = Modifier,
    showMoreMenu: Boolean = true,
) {
    var moreOpen by remember { mutableStateOf(false) }
    val colors = OpsTheme.colors

    Row(modifier = modifier, verticalAlignment = Alignment.Bottom) {
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            Column(
                modifier = Modifier
                    .width(40.dp)
                    .shadow(3.dp, RoundedCornerShape(8.dp))
                    .background(Color.White, RoundedCornerShape(8.dp)),
                horizontalAlignment = Alignment.CenterHorizontally,
            ) {
                Text(
                    text = "+",
                    fontSize = 20.sp,
                    fontWeight = FontWeight.Bold,
                    color = Color(0xFF48506C),
                    modifier = Modifier
                        .clickable(onClick = onZoomIn)
                        .padding(vertical = 10.dp),
                )
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(1.dp)
                        .background(Color(0xFFE8E8E8)),
                )
                Text(
                    text = "−",
                    fontSize = 20.sp,
                    fontWeight = FontWeight.Bold,
                    color = Color(0xFF48506C),
                    modifier = Modifier
                        .clickable(onClick = onZoomOut)
                        .padding(vertical = 10.dp),
                )
            }
            Spacer(modifier = Modifier.height(10.dp))
            Column(
                modifier = Modifier
                    .width(40.dp)
                    .shadow(3.dp, RoundedCornerShape(8.dp))
                    .background(Color.White, RoundedCornerShape(8.dp))
                    .padding(vertical = 6.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(4.dp),
            ) {
                SideToolItem(symbol = "↻", label = refreshLabel, selected = false, onClick = onRefresh)
                SideToolItem(symbol = "⌖", label = locateLabel, selected = false, onClick = onLocate)
                if (showMoreMenu) {
                    SideToolItem(
                        symbol = "▦",
                        label = moreLabel,
                        selected = moreOpen,
                        onClick = { moreOpen = !moreOpen },
                    )
                }
            }
        }
        if (showMoreMenu && moreOpen) {
            Spacer(modifier = Modifier.width(8.dp))
            Row(
                modifier = Modifier
                    .height(62.dp)
                    .shadow(3.dp, RoundedCornerShape(8.dp))
                    .background(Color.White, RoundedCornerShape(8.dp))
                    .padding(horizontal = 6.dp),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(4.dp),
            ) {
                MoreMenuItem(
                    symbol = "P",
                    label = parkingLabel,
                    selected = parkingSelected,
                    selectedColor = colors.primary,
                    onClick = onToggleParking,
                )
                MoreMenuItem(
                    symbol = "🛰",
                    label = satelliteLabel,
                    selected = satelliteSelected,
                    selectedColor = colors.primary,
                    onClick = onToggleSatellite,
                )
                MoreMenuItem(
                    symbol = "ⓘ",
                    label = detailLabel,
                    selected = detailSelected,
                    selectedColor = colors.primary,
                    onClick = onToggleDetail,
                )
            }
        }
    }
}

/** 巡检/维修半图：仅刷新 + 定位（对齐 Flutter TaskPage）。 */
@Composable
fun TaskMapSimpleSideTools(
    refreshLabel: String,
    locateLabel: String,
    onRefresh: () -> Unit,
    onLocate: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Column(
        modifier = modifier
            .width(40.dp)
            .shadow(3.dp, RoundedCornerShape(8.dp))
            .background(Color.White, RoundedCornerShape(8.dp))
            .padding(vertical = 6.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(4.dp),
    ) {
        SideToolItem(symbol = "↻", label = refreshLabel, selected = false, onClick = onRefresh)
        SideToolItem(symbol = "⌖", label = locateLabel, selected = false, onClick = onLocate)
    }
}

@Composable
fun TaskMapSwitchAreaButton(
    label: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Column(
        modifier = modifier
            .width(44.dp)
            .shadow(3.dp, RoundedCornerShape(8.dp))
            .background(Color.White, RoundedCornerShape(8.dp))
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            )
            .padding(vertical = 8.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text(text = "⧉", color = Color(0xFF48506C), fontSize = 16.sp)
        Spacer(modifier = Modifier.height(4.dp))
        Text(text = label, color = Color(0xFF48506C), fontSize = 10.sp)
    }
}

@Composable
private fun SideToolItem(
    symbol: String,
    label: String,
    selected: Boolean,
    onClick: () -> Unit,
) {
    val color = if (selected) OpsTheme.colors.primary else OpsToolIcon
    Column(
        modifier = Modifier
            .width(40.dp)
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            )
            .padding(vertical = 4.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text(text = symbol, color = color, fontSize = 15.sp)
        Text(text = label, color = color, fontSize = 10.sp)
    }
}

@Composable
private fun MoreMenuItem(
    symbol: String,
    label: String,
    selected: Boolean,
    selectedColor: Color,
    onClick: () -> Unit,
) {
    val color = if (selected) selectedColor else Color(0xFF48506C)
    Column(
        modifier = Modifier
            .width(52.dp)
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            )
            .padding(vertical = 4.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text(
            text = symbol,
            color = color,
            fontSize = 16.sp,
            fontWeight = if (selected) FontWeight.Bold else FontWeight.Normal,
        )
        Text(text = label, color = color, fontSize = 9.sp, maxLines = 1)
    }
}

@Composable
fun TaskMapFilterToggle(
    label: String,
    checked: Boolean,
    onToggle: () -> Unit,
) {
    val colors = OpsTheme.colors
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onToggle)
            .padding(vertical = 12.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Box(
            modifier = Modifier
                .size(18.dp)
                .border(
                    1.dp,
                    if (checked) colors.primary else Color(0xFFCCCCCC),
                    RoundedCornerShape(3.dp),
                )
                .background(
                    if (checked) colors.primary else Color.Transparent,
                    RoundedCornerShape(3.dp),
                ),
        )
        Spacer(modifier = Modifier.width(10.dp))
        Text(label, color = Color(0xFF333333), fontSize = 14.sp)
    }
}
