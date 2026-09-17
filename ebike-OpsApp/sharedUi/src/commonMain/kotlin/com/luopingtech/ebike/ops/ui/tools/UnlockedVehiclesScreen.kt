package com.luopingtech.ebike.ops.ui.tools

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.model.Vehicle
import com.luopingtech.ebike.ops.ui.analysis.VcdTopBar
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import com.luopingtech.ebike.ops.ui.vehicle.LocalOpenVehicleDetail
import kotlinx.coroutines.launch

/**
 * 对齐原版 UnLockedVehiclesActivity：搜索 + 94dp 行（车辆详情 / 关锁）。
 */
@Composable
fun UnlockedVehiclesScreen(
    app: OpsApp,
    currentArea: ServiceArea?,
    onClose: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val state by app.unlockedVehicleFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    val openVehicleDetail = LocalOpenVehicleDetail.current
    val colors = OpsTheme.colors

    LaunchedEffect(currentArea?.id) {
        app.unlockedVehicleFeature.load(currentArea)
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.White),
    ) {
        VcdTopBar(
            title = t(Str.UnlockedVehicles),
            onBack = {
                app.unlockedVehicleFeature.clear()
                onClose()
            },
        )
        if (state.canFilterStaff) {
            OutlinedTextField(
                value = state.query,
                onValueChange = { app.unlockedVehicleFeature.setQuery(it) },
                placeholder = { Text(t(Str.UnlockedVehiclesSearch)) },
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp, vertical = 12.dp),
                singleLine = true,
            )
        } else {
            Text(
                text = t(Str.UnlockedSelfOnlyHint),
                color = Color(0xFF999999),
                fontSize = 13.sp,
                modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp),
            )
        }
        if (state.loading && state.vehicles.isEmpty()) {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                CircularProgressIndicator(color = colors.primary)
            }
        } else {
            LazyColumn(modifier = Modifier.fillMaxSize()) {
                items(state.vehicles, key = { it.carId }) { row ->
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(94.dp)
                            .padding(horizontal = 16.dp),
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Text(
                            text = row.carId,
                            color = Color.Black,
                            fontSize = 18.sp,
                            modifier = Modifier.weight(1f),
                        )
                        OutlinedButton(
                            onClick = {
                                scope.launch {
                                    val fresh = when (val r = app.vehicleFeature.refreshDetail(row.carId)) {
                                        is OpsResult.Ok -> r.value
                                        is OpsResult.Err -> Vehicle(carId = row.carId, imei = row.imei)
                                    }
                                    openVehicleDetail.open(fresh, currentArea?.id)
                                }
                            },
                            modifier = Modifier.height(32.dp),
                            shape = RoundedCornerShape(4.dp),
                            border = BorderStroke(1.dp, Color(0xFF999999)),
                            contentPadding = PaddingValues(horizontal = 12.dp),
                        ) {
                            Text(t(Str.OrderQueryOpenVehicle), fontWeight = FontWeight.Bold, fontSize = 14.sp, color = Color(0xFF333333))
                        }
                        Spacer(modifier = Modifier.width(12.dp))
                        Button(
                            onClick = { scope.launch { app.unlockedVehicleFeature.lock(row.carId) } },
                            enabled = !state.loading && state.lockingCarId != row.carId,
                            modifier = Modifier.height(32.dp),
                            shape = RoundedCornerShape(4.dp),
                            colors = ButtonDefaults.buttonColors(
                                containerColor = colors.primary,
                                contentColor = colors.onPrimary,
                            ),
                            contentPadding = PaddingValues(horizontal = 26.dp),
                        ) {
                            Text(
                                if (state.lockingCarId == row.carId) t(Str.LoadingEllipsis)
                                else t(Str.UnlockedLock),
                                fontSize = 14.sp,
                            )
                        }
                    }
                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(0.5.dp)
                            .background(Color(0xFFCCCCCC)),
                    )
                }
                if (!state.loading && state.vehicles.isEmpty()) {
                    item {
                        Text(
                            text = t(Str.UnlockedVehiclesCount, 0),
                            color = Color(0xFF999999),
                            modifier = Modifier.padding(24.dp),
                        )
                    }
                }
            }
        }
        state.message?.let {
            Text(it, color = colors.primary, modifier = Modifier.padding(16.dp))
        }
        state.errorMessage?.let {
            Text(it, color = Color(0xFFE02020), modifier = Modifier.padding(16.dp))
        }
    }
}
