package com.luopingtech.ebike.ops.ui.home

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.RowScope
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.domain.model.Vehicle
import com.luopingtech.ebike.ops.domain.vehicle.VehicleListStatusTone
import com.luopingtech.ebike.ops.domain.vehicle.listStatus
import com.luopingtech.ebike.ops.ui.theme.OpsStatGreen
import com.luopingtech.ebike.ops.ui.theme.OpsStatRed
import com.luopingtech.ebike.ops.ui.theme.OpsTheme

/**
 * Legacy [SimpleCarListActivity]: cluster tap at zoom ≥ 16 → 车辆列表 → 点行进详情.
 */
@Composable
fun SimpleCarListScreen(
    vehicles: List<Vehicle>,
    title: String,
    backLabel: String,
    colId: String,
    colStatus: String,
    colBattery: String,
    onBack: () -> Unit,
    onVehicleClick: (Vehicle) -> Unit,
) {
    val colors = OpsTheme.colors
    com.luopingtech.ebike.ops.ui.navigation.OpsBackHandler(onBack = onBack)
    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.White),
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .background(colors.primary)
                .statusBarsPadding()
                .height(50.dp)
                .padding(horizontal = 12.dp),
        ) {
            Text(
                text = backLabel,
                color = colors.onPrimary,
                fontSize = 16.sp,
                modifier = Modifier
                    .align(Alignment.CenterStart)
                    .clickable(
                        interactionSource = remember { MutableInteractionSource() },
                        indication = null,
                        onClick = onBack,
                    ),
            )
            Text(
                text = title,
                color = colors.onPrimary,
                fontSize = 18.sp,
                fontWeight = FontWeight.Medium,
                modifier = Modifier.align(Alignment.Center),
            )
        }
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .height(44.dp)
                .padding(horizontal = 8.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            HeaderCell(colId)
            HeaderCell(colStatus)
            HeaderCell(colBattery)
        }
        HorizontalDivider(color = Color(0xFFD8D8D8), thickness = 1.dp)
        LazyColumn(modifier = Modifier.fillMaxSize()) {
            items(vehicles, key = { it.carId }) { vehicle ->
                val status = vehicle.listStatus()
                val statusColor = when (status.tone) {
                    VehicleListStatusTone.Positive -> OpsStatGreen
                    VehicleListStatusTone.Alert -> OpsStatRed
                    VehicleListStatusTone.Neutral -> Color(0xFF333333)
                }
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(40.dp)
                        .clickable { onVehicleClick(vehicle) }
                        .padding(horizontal = 8.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text(
                        vehicle.carId,
                        modifier = Modifier.weight(1f),
                        textAlign = TextAlign.Center,
                        color = Color(0xFF333333),
                        fontSize = 14.sp,
                    )
                    Text(
                        status.label,
                        modifier = Modifier.weight(1f),
                        textAlign = TextAlign.Center,
                        color = statusColor,
                        fontSize = 14.sp,
                    )
                    Text(
                        vehicle.batteryLabel,
                        modifier = Modifier.weight(1f),
                        textAlign = TextAlign.Center,
                        color = Color(0xFF333333),
                        fontSize = 14.sp,
                    )
                }
                HorizontalDivider(color = Color(0xFFE6E6E6), thickness = 1.dp)
            }
        }
    }
}

@Composable
private fun RowScope.HeaderCell(text: String) {
    Text(
        text = text,
        modifier = Modifier.weight(1f),
        textAlign = TextAlign.Center,
        color = Color(0xFF4D4D4D),
        fontSize = 14.sp,
    )
}
