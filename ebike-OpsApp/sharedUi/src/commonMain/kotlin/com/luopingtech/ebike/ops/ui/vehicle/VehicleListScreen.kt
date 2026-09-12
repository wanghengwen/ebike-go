package com.luopingtech.ebike.ops.ui.vehicle

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.vehicle.VehicleListStatusTone
import com.luopingtech.ebike.ops.domain.vehicle.listStatus
import com.luopingtech.ebike.ops.ui.theme.OpsStatGreen
import com.luopingtech.ebike.ops.ui.theme.OpsStatRed
import com.luopingtech.ebike.ops.ui.theme.OpsTheme

@Composable
fun VehicleListScreen(
    app: OpsApp,
    currentArea: ServiceArea?,
    onClose: () -> Unit,
    onChangeArea: () -> Unit = {},
    onVehicleClick: (String) -> Unit = {},
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val vehicleState by app.vehicleFeature.state.collectAsState()
    var query by remember { mutableStateOf("") }
    var appliedQuery by remember { mutableStateOf("") }
    val colors = OpsTheme.colors

    LaunchedEffect(currentArea?.id) {
        app.vehicleFeature.loadForArea(currentArea)
    }

    val filtered = remember(vehicleState.vehicles, appliedQuery) {
        val q = appliedQuery.trim()
        if (q.isEmpty()) {
            vehicleState.vehicles
        } else {
            vehicleState.vehicles.filter {
                it.carId.contains(q, ignoreCase = true) ||
                    it.imei.contains(q, ignoreCase = true)
            }
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(colors.pageBackground),
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .background(colors.primary)
                .statusBarsPadding()
                .padding(horizontal = 12.dp, vertical = 10.dp),
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Text(
                    text = "<",
                    color = colors.onPrimary,
                    fontSize = 28.sp,
                    modifier = Modifier
                        .clickable(onClick = onClose)
                        .padding(4.dp),
                )
                Row(
                    modifier = Modifier
                        .weight(1f)
                        .clickable(onClick = onChangeArea),
                    horizontalArrangement = Arrangement.Center,
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text(
                        text = currentArea?.name?.takeIf { it.isNotBlank() }
                            ?: t(Str.SelectServiceArea),
                        color = colors.onPrimary,
                        fontSize = 17.sp,
                        fontWeight = FontWeight.Medium,
                    )
                    Text(" v", color = colors.onPrimary, fontSize = 12.sp)
                }
                Spacer(modifier = Modifier.width(36.dp))
            }
            Spacer(modifier = Modifier.height(10.dp))
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(Color.White, RoundedCornerShape(8.dp))
                    .padding(horizontal = 12.dp, vertical = 8.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                BasicTextField(
                    value = query,
                    onValueChange = { query = it },
                    singleLine = true,
                    textStyle = TextStyle(color = colors.textPrimary, fontSize = 15.sp),
                    cursorBrush = SolidColor(colors.primary),
                    keyboardOptions = KeyboardOptions(imeAction = ImeAction.Search),
                    keyboardActions = KeyboardActions(
                        onSearch = { appliedQuery = query },
                    ),
                    modifier = Modifier.weight(1f),
                    decorationBox = { inner ->
                        if (query.isEmpty()) {
                            Text(
                                t(Str.VehicleListSearchHint),
                                color = colors.textTertiary,
                                fontSize = 15.sp,
                            )
                        }
                        inner()
                    },
                )
                Text(
                    text = t(Str.Search),
                    color = colors.primary,
                    fontWeight = FontWeight.Medium,
                    modifier = Modifier
                        .clickable { appliedQuery = query }
                        .padding(start = 8.dp),
                )
            }
        }

        Column(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color(0xFFF2F3F5))
                .padding(vertical = 8.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Text(
                text = t(Str.VehicleListFilteredCount, filtered.size),
                color = colors.textSecondary,
                fontSize = 13.sp,
            )
        }

        Row(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color.White)
                .padding(horizontal = 16.dp, vertical = 10.dp),
        ) {
            Text(
                t(Str.VehicleListColId),
                modifier = Modifier.weight(1.2f),
                fontWeight = FontWeight.Medium,
                fontSize = 14.sp,
            )
            Text(
                t(Str.VehicleListColStatus),
                modifier = Modifier.weight(1f),
                textAlign = TextAlign.Center,
                fontWeight = FontWeight.Medium,
                fontSize = 14.sp,
            )
            Text(
                t(Str.VehicleListColBattery),
                modifier = Modifier.weight(0.8f),
                textAlign = TextAlign.End,
                fontWeight = FontWeight.Medium,
                fontSize = 14.sp,
            )
        }
        HorizontalDivider(color = colors.divider)

        if (vehicleState.loading && filtered.isEmpty()) {
            Column(
                modifier = Modifier.fillMaxSize(),
                verticalArrangement = Arrangement.Center,
                horizontalAlignment = Alignment.CenterHorizontally,
            ) {
                Text(t(Str.LoadingEllipsis), color = colors.textSecondary)
            }
        } else {
            LazyColumn(modifier = Modifier.fillMaxSize()) {
                items(filtered, key = { it.carId }) { vehicle ->
                    val status = vehicle.listStatus()
                    val statusColor = when (status.tone) {
                        VehicleListStatusTone.Positive -> OpsStatGreen
                        VehicleListStatusTone.Alert -> OpsStatRed
                        VehicleListStatusTone.Neutral -> colors.textPrimary
                    }
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .background(Color.White)
                            .clickable { onVehicleClick(vehicle.carId) }
                            .padding(horizontal = 16.dp, vertical = 14.dp),
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Text(
                            vehicle.carId,
                            modifier = Modifier.weight(1.2f),
                            color = colors.textPrimary,
                            fontSize = 15.sp,
                        )
                        Text(
                            status.label,
                            modifier = Modifier.weight(1f),
                            textAlign = TextAlign.Center,
                            color = statusColor,
                            fontSize = 15.sp,
                        )
                        Text(
                            vehicle.batteryLabel,
                            modifier = Modifier.weight(0.8f),
                            textAlign = TextAlign.End,
                            color = colors.textPrimary,
                            fontSize = 15.sp,
                        )
                    }
                    HorizontalDivider(color = colors.divider)
                }
                item {
                    Text(
                        text = when {
                            vehicleState.errorMessage != null -> vehicleState.errorMessage!!
                            filtered.isEmpty() -> t(Str.FilterNoVehicles)
                            else -> t(Str.VehicleListEnd)
                        },
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(24.dp),
                        textAlign = TextAlign.Center,
                        color = colors.textTertiary,
                        fontSize = 13.sp,
                    )
                }
            }
        }
    }
}
