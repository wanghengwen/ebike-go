package com.luopingtech.ebike.rider.ui.ride

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.safeDrawingPadding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.rider.RiderApp
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.domain.model.MapPin
import com.luopingtech.ebike.rider.domain.model.TrackPoint
import com.luopingtech.ebike.rider.domain.riding.FenceTipSeverity
import com.luopingtech.ebike.rider.domain.riding.RideFormat
import com.luopingtech.ebike.rider.domain.riding.RidePhase
import com.luopingtech.ebike.rider.feature.riding.RidingUiState
import com.luopingtech.ebike.rider.feature.riding.UnlockChannel
import com.luopingtech.ebike.rider.platform.MapProviderKind
import com.luopingtech.ebike.rider.ui.map.RiderMapSpec
import com.luopingtech.ebike.rider.ui.map.RiderMapView
import com.luopingtech.ebike.rider.ui.theme.RiderTheme

/**
 * 骑行主界面（对应 UniApp `pages/riding/riding.vue`）。
 *
 * 费用 / 操作区落在地图上方的白底圆角面板里，避免按键直接叠在地图上不清晰。
 */
@Composable
fun RidingScreen(
    app: RiderApp,
    state: RidingUiState,
    onTempLockToggle: () -> Unit,
    onReturn: () -> Unit,
    onRing: () -> Unit,
    onHelmet: () -> Unit,
    onFindParking: () -> Unit,
    onOpenTripMap: () -> Unit,
    onApplyReturn: () -> Unit,
    onDismissPrompt: () -> Unit,
    onConfirmReturn: (forcePenalty: Boolean) -> Unit,
    onOpenGuide: (pageType: Int, returnTypeCode: Int) -> Unit,
    onRecoverPower: () -> Unit,
    onRecharge: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val info = state.rideInfo
    val tempLocked = state.phase == RidePhase.TempLocked
    val brand = RiderTheme.colors.primary
    val panelShape = RoundedCornerShape(topStart = 16.dp, topEnd = 16.dp)
    val pillShape = RoundedCornerShape(24.dp)
    val toolShape = RoundedCornerShape(16.dp)

    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)

    val providerLabel = when (app.mapCapability.kind) {
        MapProviderKind.TENCENT -> t(Str.TencentMap)
        MapProviderKind.SIMULATOR -> t(Str.SimulatorMap)
        else -> app.mapCapability.kind.name
    }

    Box(modifier = modifier.fillMaxSize()) {
        RiderMapView(
            spec = RiderMapSpec(
                pins = vehiclePin(state),
                providerLabel = providerLabel,
                clusterOverview = false,
                fencePolygons = state.fences,
                trackPoints = state.track.map { TrackPoint(lat = it.lat, lng = it.lng) },
                showStatusOverlay = false,
            ),
            modifier = Modifier.fillMaxSize(),
        )

        Column(
            modifier = Modifier
                .align(Alignment.TopCenter)
                .fillMaxWidth()
                .safeDrawingPadding()
                .padding(12.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            state.fenceTip?.let { tip ->
                val warning = tip.severity == FenceTipSeverity.Warning
                Surface(
                    color = if (warning) {
                        MaterialTheme.colorScheme.error
                    } else {
                        Color.White
                    },
                    shape = RoundedCornerShape(10.dp),
                    shadowElevation = 2.dp,
                    modifier = Modifier.fillMaxWidth(),
                ) {
                    Text(
                        text = t(tip.text, tip.textArg),
                        style = MaterialTheme.typography.bodySmall,
                        color = if (warning) {
                            MaterialTheme.colorScheme.onError
                        } else {
                            RiderTheme.colors.textPrimary
                        },
                        textAlign = TextAlign.Center,
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 12.dp, vertical = 8.dp),
                    )
                }
            }

            if (info?.nearServiceEdge == true) {
                Surface(
                    color = Color.White,
                    shape = RoundedCornerShape(10.dp),
                    shadowElevation = 2.dp,
                ) {
                    Text(
                        text = t(Str.RideNearServiceEdge),
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.error,
                        modifier = Modifier.padding(horizontal = 12.dp, vertical = 8.dp),
                    )
                }
            }
        }

        Surface(
            color = Color.White,
            shape = panelShape,
            shadowElevation = 8.dp,
            modifier = Modifier
                .align(Alignment.BottomCenter)
                .fillMaxWidth(),
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .safeDrawingPadding()
                    .padding(horizontal = 16.dp, vertical = 14.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text(
                        text = t(Str.RideVehicleNo, state.carId),
                        style = MaterialTheme.typography.titleSmall,
                        color = RiderTheme.colors.textPrimary,
                    )
                    state.unlockChannel?.let { channel ->
                        Text(
                            text = when (channel) {
                                UnlockChannel.Ble -> t(Str.RideUnlockChannelBle)
                                UnlockChannel.Network -> t(Str.RideUnlockChannelNetwork)
                            },
                            style = MaterialTheme.typography.labelSmall,
                            color = RiderTheme.colors.textTertiary,
                        )
                    }
                }

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                ) {
                    RideMetric(
                        label = t(Str.RideCost),
                        value = RideFormat.yuan(info?.costFeeFen ?: 0),
                        suffix = t(Str.RideYuan),
                        modifier = Modifier.weight(1f),
                    )
                    RideMetric(
                        label = t(Str.RideDuration),
                        value = RideFormat.duration(state.elapsedSeconds),
                        modifier = Modifier.weight(1f),
                    )
                    RideMetric(
                        label = t(Str.RideDistance),
                        value = RideFormat.distance(info?.rideDistanceMeters ?: 0),
                        modifier = Modifier.weight(1f),
                    )
                }

                info?.restMileage?.let { km ->
                    Text(
                        text = t(Str.RideRestMileage, km),
                        style = MaterialTheme.typography.bodySmall,
                        color = RiderTheme.colors.textTertiary,
                    )
                }
                if (tempLocked) {
                    Text(
                        text = t(Str.RideTempLocked),
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.error,
                    )
                }

                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    RidingToolButton(
                        text = t(Str.RideFindVehicle),
                        enabled = !state.busy,
                        onClick = onRing,
                        shape = toolShape,
                        modifier = Modifier.weight(1f),
                    )
                    if (info?.helmetState != null && info.helmetState != HELMET_STATE_ABSENT) {
                        RidingToolButton(
                            text = t(Str.RideHelmet),
                            enabled = !state.busy,
                            onClick = onHelmet,
                            shape = toolShape,
                            modifier = Modifier.weight(1f),
                        )
                    }
                    RidingToolButton(
                        text = t(Str.ParkSearchTitle),
                        enabled = !state.busy,
                        onClick = onFindParking,
                        shape = toolShape,
                        modifier = Modifier.weight(1f),
                    )
                }

                Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                    OutlinedButton(
                        onClick = onTempLockToggle,
                        enabled = !state.busy,
                        modifier = Modifier
                            .weight(1f)
                            .height(48.dp),
                        shape = pillShape,
                        border = BorderStroke(1.dp, Color(0xFFCCCCCC)),
                        colors = ButtonDefaults.outlinedButtonColors(
                            containerColor = Color.White,
                            contentColor = RiderTheme.colors.textPrimary,
                        ),
                    ) {
                        Text(
                            text = if (tempLocked) t(Str.RideEndTempLock) else t(Str.RideTempLock),
                            fontWeight = FontWeight.SemiBold,
                            fontSize = 15.sp,
                        )
                    }
                    Button(
                        onClick = onReturn,
                        enabled = !state.busy,
                        modifier = Modifier
                            .weight(1.2f)
                            .height(48.dp),
                        shape = pillShape,
                        colors = ButtonDefaults.buttonColors(
                            containerColor = brand,
                            contentColor = Color.White,
                        ),
                    ) {
                        Text(
                            text = t(Str.RideReturn),
                            fontWeight = FontWeight.SemiBold,
                            fontSize = 15.sp,
                        )
                    }
                }

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                ) {
                    TextButton(onClick = onOpenTripMap) {
                        Text(t(Str.TripMapTitle), color = RiderTheme.colors.textSecondary)
                    }
                    if (state.config.showApplyEntry) {
                        TextButton(onClick = onApplyReturn) {
                            Text(t(Str.RideCannotReturn), color = RiderTheme.colors.textSecondary)
                        }
                    } else {
                        Spacer(modifier = Modifier)
                    }
                }
            }
        }

        if (state.busy) {
            Surface(
                color = RiderTheme.colors.pageBackground.copy(alpha = 0.6f),
                modifier = Modifier
                    .align(Alignment.Center)
                    .padding(24.dp),
                shape = RoundedCornerShape(12.dp),
            ) {
                Column(
                    modifier = Modifier.padding(20.dp),
                    horizontalAlignment = Alignment.CenterHorizontally,
                ) {
                    CircularProgressIndicator()
                    state.busyLabel?.let {
                        Text(
                            text = t(it),
                            style = MaterialTheme.typography.bodySmall,
                            color = RiderTheme.colors.textSecondary,
                            modifier = Modifier.padding(top = 8.dp),
                        )
                    }
                }
            }
        }

        state.prompt?.let { prompt ->
            RidePromptHost(
                i18n = app.i18n,
                prompt = prompt,
                onDismiss = onDismissPrompt,
                onConfirmReturn = onConfirmReturn,
                onOpenGuide = onOpenGuide,
                onFindParking = onFindParking,
                onRecoverPower = onRecoverPower,
                onRecharge = onRecharge,
                onRefreshReturn = onReturn,
            )
        }
    }
}

@Composable
private fun RidingToolButton(
    text: String,
    enabled: Boolean,
    onClick: () -> Unit,
    shape: RoundedCornerShape,
    modifier: Modifier = Modifier,
) {
    Button(
        onClick = onClick,
        enabled = enabled,
        modifier = modifier.height(44.dp),
        shape = shape,
        colors = ButtonDefaults.buttonColors(
            containerColor = Color(0xFFF0F1F3),
            contentColor = RiderTheme.colors.textPrimary,
            disabledContainerColor = Color(0xFFF0F1F3).copy(alpha = 0.6f),
            disabledContentColor = RiderTheme.colors.textTertiary,
        ),
        elevation = ButtonDefaults.buttonElevation(defaultElevation = 0.dp),
    ) {
        Text(
            text = text,
            style = MaterialTheme.typography.labelLarge,
            maxLines = 1,
        )
    }
}

@Composable
private fun RideMetric(
    label: String,
    value: String,
    suffix: String = "",
    modifier: Modifier = Modifier,
) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        modifier = modifier,
    ) {
        Text(
            text = if (suffix.isBlank()) value else "$value $suffix",
            style = MaterialTheme.typography.headlineSmall.copy(
                fontWeight = FontWeight.SemiBold,
                fontSize = 26.sp,
            ),
            color = RiderTheme.colors.textPrimary,
        )
        Text(
            text = label,
            style = MaterialTheme.typography.labelSmall,
            color = RiderTheme.colors.textTertiary,
            modifier = Modifier.padding(top = 4.dp),
        )
    }
}

/** 车机上报的坐标；没有就不画 pin（画 0,0 会把地图甩到几内亚湾）。 */
private fun vehiclePin(state: RidingUiState): List<MapPin> {
    val info = state.rideInfo ?: return emptyList()
    if (!info.hasVehicleLocation) return emptyList()
    return listOf(
        MapPin(
            id = state.carId,
            lat = info.lat,
            lng = info.lng,
            title = state.carId,
            restBattery = info.restBattery ?: 0,
            ridingState = info.ridingState,
        ),
    )
}

private const val HELMET_STATE_ABSENT = 2
