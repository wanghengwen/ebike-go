package com.luopingtech.ebike.ops.ui.home

import androidx.compose.foundation.background
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
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.domain.vehicle.VehicleMapFilter
import com.luopingtech.ebike.ops.ui.theme.OpsFilterHandle
import com.luopingtech.ebike.ops.ui.theme.OpsStatBlue
import com.luopingtech.ebike.ops.ui.theme.OpsStatGreen
import com.luopingtech.ebike.ops.ui.theme.OpsStatItemBg
import com.luopingtech.ebike.ops.ui.theme.OpsStatOrange
import com.luopingtech.ebike.ops.ui.theme.OpsStatRed
import com.luopingtech.ebike.ops.ui.theme.OpsStatSelected
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import com.luopingtech.ebike.ops.ui.theme.OpsToolIcon

@Composable
fun HomeAreaTitleBar(
    areaName: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val colors = OpsTheme.colors
    Row(
        modifier = modifier
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            )
            .padding(horizontal = 12.dp, vertical = 8.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.Center,
    ) {
        Text(
            text = areaName,
            color = colors.textPrimary,
            fontSize = 16.sp,
            fontWeight = FontWeight.Medium,
        )
        Spacer(modifier = Modifier.width(4.dp))
        Text(text = "▾", color = colors.textPrimary, fontSize = 12.sp)
    }
}

@Composable
fun HomeMapToolsRail(
    refreshLabel: String,
    detailLabel: String,
    fenceLabel: String,
    switchLabel: String,
    detailSelected: Boolean,
    fenceSelected: Boolean,
    switchSelected: Boolean,
    onRefresh: () -> Unit,
    onDetail: () -> Unit,
    onFence: () -> Unit,
    onSwitch: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Column(
        modifier = modifier
            .width(40.dp)
            .shadow(4.dp, RoundedCornerShape(20.dp))
            .background(Color.White, RoundedCornerShape(20.dp))
            .padding(vertical = 16.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        MapToolItem(symbol = "↻", label = refreshLabel, selected = false, onClick = onRefresh)
        MapToolItem(symbol = "ⓘ", label = detailLabel, selected = detailSelected, onClick = onDetail)
        MapToolItem(symbol = "▦", label = fenceLabel, selected = fenceSelected, onClick = onFence)
        MapToolItem(symbol = "⇄", label = switchLabel, selected = switchSelected, onClick = onSwitch)
    }
}

@Composable
private fun MapToolItem(
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
            ),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text(text = symbol, color = color, fontSize = 16.sp)
        Text(text = label, color = color, fontSize = 11.sp)
    }
}

@Composable
fun HomeFilterHandle(
    label: String,
    open: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Box(
        modifier = modifier
            .width(28.dp)
            .height(88.dp)
            .background(
                color = OpsFilterHandle,
                shape = RoundedCornerShape(topStart = 8.dp, bottomStart = 8.dp),
            )
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            ),
        contentAlignment = Alignment.Center,
    ) {
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            Text(
                text = if (open) "›" else "‹",
                color = Color.White,
                fontSize = 12.sp,
            )
            label.forEach { ch ->
                Text(text = ch.toString(), color = Color.White, fontSize = 11.sp)
            }
        }
    }
}

data class HomeStatItem(
    val filter: VehicleMapFilter,
    val label: String,
    val count: Int,
    val valueColor: Color,
)

@Composable
fun HomeStatisticsPanel(
    items: List<HomeStatItem>,
    selected: VehicleMapFilter,
    onSelect: (VehicleMapFilter) -> Unit,
    modifier: Modifier = Modifier,
) {
    val colors = OpsTheme.colors
    Column(
        modifier = modifier
            .fillMaxWidth()
            .shadow(8.dp, RoundedCornerShape(topStart = 20.dp, topEnd = 20.dp))
            .background(
                colors.pageBackground,
                RoundedCornerShape(topStart = 20.dp, topEnd = 20.dp),
            )
            .padding(horizontal = 16.dp, vertical = 24.dp),
        verticalArrangement = Arrangement.spacedBy(9.dp),
    ) {
        items.chunked(4).forEach { row ->
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(9.dp),
            ) {
                row.forEach { item ->
                    val isSelected = selected == item.filter
                    Row(
                        modifier = Modifier
                            .weight(1f)
                            .height(33.dp)
                            .background(
                                color = if (isSelected) OpsStatSelected else OpsStatItemBg,
                                shape = RoundedCornerShape(8.dp),
                            )
                            .clickable(
                                interactionSource = remember { MutableInteractionSource() },
                                indication = null,
                                onClick = {
                                    onSelect(
                                        if (isSelected) VehicleMapFilter.All else item.filter,
                                    )
                                },
                            ),
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.Center,
                    ) {
                        Text(
                            text = item.label,
                            color = if (isSelected) colors.onPrimary else colors.textPrimary,
                            fontSize = 12.sp,
                            fontWeight = FontWeight.SemiBold,
                        )
                        Spacer(modifier = Modifier.width(4.dp))
                        Text(
                            text = item.count.toString(),
                            color = when {
                                isSelected -> colors.onPrimary
                                item.count > 0 -> item.valueColor
                                else -> colors.textPrimary
                            },
                            fontSize = 12.sp,
                            fontWeight = FontWeight.SemiBold,
                            textAlign = TextAlign.Center,
                        )
                    }
                }
            }
        }
    }
}

fun homeStatColor(filter: VehicleMapFilter): Color = when (filter) {
    VehicleMapFilter.Warehouse, VehicleMapFilter.Ready -> OpsStatGreen
    VehicleMapFilter.Booking, VehicleMapFilter.Riding, VehicleMapFilter.TempParking -> OpsStatBlue
    VehicleMapFilter.LowBattery, VehicleMapFilter.Repairing -> OpsStatRed
    VehicleMapFilter.Moving -> OpsStatOrange
    VehicleMapFilter.All -> Color(0xFF333333)
}
