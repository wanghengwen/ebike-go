package com.luopingtech.ebike.rider.ui.ride

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.safeDrawingPadding
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.rider.RiderApp
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.domain.model.MapPin
import com.luopingtech.ebike.rider.domain.riding.RideFormat
import com.luopingtech.ebike.rider.domain.riding.RidePhase
import com.luopingtech.ebike.rider.feature.riding.RidingUiState
import com.luopingtech.ebike.rider.platform.MapProviderKind
import com.luopingtech.ebike.rider.ui.map.RiderMapSpec
import com.luopingtech.ebike.rider.ui.map.RiderMapView
import com.luopingtech.ebike.rider.ui.platform.RiderBackHandler
import com.luopingtech.ebike.rider.ui.theme.RiderTheme

/**
 * 确认开锁页 —— 对齐 UniApp `precycling.vue`：全屏地图 + 区 tip + 底栏
 *（可骑里程 / 起步价 / 确认开锁）。[RidePhase.Unlocking] 叠开锁进度面板。
 */
@Composable
fun PreCyclingScreen(
    app: RiderApp,
    state: RidingUiState,
    onUnlock: () -> Unit,
    onChangeVehicle: () -> Unit,
    onBack: () -> Unit,
    onOpenBilling: () -> Unit = {},
    modifier: Modifier = Modifier,
) {
    val vehicle = state.vehicle
    val unlocking = state.phase == RidePhase.Unlocking

    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)

    RiderBackHandler(onBack = { if (!unlocking) onBack() })

    LaunchedEffect(state.phase, state.carId) {
        if (state.phase == RidePhase.Confirming || state.phase == RidePhase.Unlocking) {
            app.ridingFeature.loadFences()
        }
    }

    val blockedReason = when {
        unlocking -> null
        vehicle == null && !state.busy -> {
            state.session.lastError.takeIf { it.isNotBlank() }
                ?: t(Str.RideVehicleUnavailable)
        }
        vehicle == null -> null
        vehicle.outOfServiceArea -> t(Str.RideOutOfServiceCannotUnlock)
        !vehicle.available -> t(Str.RideVehicleUnavailable)
        else -> null
    }

    val providerLabel = when (app.mapCapability.kind) {
        MapProviderKind.TENCENT -> t(Str.TencentMap)
        MapProviderKind.SIMULATOR -> t(Str.SimulatorMap)
        else -> app.mapCapability.kind.name
    }

    val unlockDisabled = unlocking ||
        state.busy ||
        vehicle == null ||
        blockedReason != null ||
        vehicle.outOfServiceArea

    Box(modifier = modifier.fillMaxSize()) {
        RiderMapView(
            spec = RiderMapSpec(
                pins = vehiclePins(state),
                providerLabel = providerLabel,
                clusterOverview = false,
                fencePolygons = state.fences,
                showStatusOverlay = false,
                followLat = vehicle?.takeIf { it.hasLocation }?.lat
                    ?: state.myLocation?.latitude,
                followLng = vehicle?.takeIf { it.hasLocation }?.lng
                    ?: state.myLocation?.longitude,
                followNonce = if (vehicle?.hasLocation == true) 1 else 0,
            ),
            modifier = Modifier.fillMaxSize(),
        )

        if (!unlocking) {
            TextButton(
                onClick = onBack,
                modifier = Modifier
                    .align(Alignment.TopStart)
                    .safeDrawingPadding()
                    .padding(4.dp),
            ) {
                Text(t(Str.Back))
            }
        }

        Column(
            modifier = Modifier
                .align(Alignment.BottomCenter)
                .fillMaxWidth()
                .safeDrawingPadding()
                .padding(horizontal = 12.dp, vertical = 12.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            if (!unlocking) {
                Surface(
                    color = RiderTheme.colors.chipBackground,
                    shape = MaterialTheme.shapes.small,
                    modifier = Modifier.fillMaxWidth(),
                ) {
                    Text(
                        text = t(Str.RideAreaTip),
                        style = MaterialTheme.typography.bodySmall,
                        color = RiderTheme.colors.textPrimary,
                        textAlign = TextAlign.Center,
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 12.dp, vertical = 8.dp),
                    )
                }
            }

            if (unlocking) {
                UnlockingSheet(
                    carId = state.carId,
                    formatProgress = { percent -> t(Str.RideUnlockProgress, percent) },
                    safetyTip = t(Str.RideUnlockSafetyTip),
                )
            } else {
                Surface(
                    tonalElevation = 3.dp,
                    shape = MaterialTheme.shapes.large,
                    modifier = Modifier.fillMaxWidth(),
                ) {
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 20.dp, vertical = 18.dp),
                        verticalArrangement = Arrangement.spacedBy(14.dp),
                    ) {
                        if (blockedReason != null) {
                            Text(
                                text = blockedReason,
                                style = MaterialTheme.typography.bodyMedium,
                                color = MaterialTheme.colorScheme.error,
                            )
                            Button(onClick = onChangeVehicle, modifier = Modifier.fillMaxWidth()) {
                                Text(t(Str.RideChangeVehicle))
                            }
                        } else {
                            Row(
                                modifier = Modifier.fillMaxWidth(),
                                horizontalArrangement = Arrangement.SpaceBetween,
                            ) {
                                Column(modifier = Modifier.weight(1f)) {
                                    Row(verticalAlignment = Alignment.Bottom) {
                                        Text(
                                            text = vehicle?.restMileage?.toString() ?: PLACEHOLDER,
                                            style = MaterialTheme.typography.headlineMedium.copy(
                                                fontWeight = FontWeight.Bold,
                                                fontSize = 32.sp,
                                            ),
                                            color = RiderTheme.colors.textPrimary,
                                        )
                                        Text(
                                            text = " km",
                                            style = MaterialTheme.typography.bodyMedium,
                                            color = RiderTheme.colors.textSecondary,
                                            modifier = Modifier.padding(bottom = 4.dp, start = 2.dp),
                                        )
                                    }
                                    Text(
                                        text = t(Str.RideableDistance),
                                        style = MaterialTheme.typography.bodySmall,
                                        color = RiderTheme.colors.textTertiary,
                                    )
                                    Text(
                                        text = "NO.${state.carId.ifBlank { PLACEHOLDER }}",
                                        style = MaterialTheme.typography.labelMedium,
                                        color = RiderTheme.colors.textSecondary,
                                        modifier = Modifier.padding(top = 4.dp),
                                    )
                                }
                                Column(
                                    modifier = Modifier.weight(1f),
                                    horizontalAlignment = Alignment.End,
                                ) {
                                    Row(verticalAlignment = Alignment.Bottom) {
                                        Text(
                                            text = vehicle?.startPriceFen?.let { RideFormat.yuan(it) }
                                                ?: PLACEHOLDER,
                                            style = MaterialTheme.typography.headlineMedium.copy(
                                                fontWeight = FontWeight.Bold,
                                                fontSize = 32.sp,
                                            ),
                                            color = RiderTheme.colors.textPrimary,
                                        )
                                        Text(
                                            text = " ${t(Str.RideYuan)}",
                                            style = MaterialTheme.typography.bodyMedium,
                                            color = RiderTheme.colors.textSecondary,
                                            modifier = Modifier.padding(bottom = 4.dp, start = 2.dp),
                                        )
                                    }
                                    Text(
                                        text = if (vehicle != null && vehicle.startPriceMinutes > 0) {
                                            t(Str.StartingPriceWithin, vehicle.startPriceMinutes)
                                        } else {
                                            t(Str.StartingPrice)
                                        },
                                        style = MaterialTheme.typography.bodySmall,
                                        color = RiderTheme.colors.textTertiary,
                                    )
                                    Text(
                                        text = t(Str.BillingRules),
                                        style = MaterialTheme.typography.labelMedium,
                                        color = RiderTheme.colors.primary,
                                        modifier = Modifier
                                            .padding(top = 4.dp)
                                            .clickable(onClick = onOpenBilling),
                                    )
                                }
                            }

                            state.session.lastError.takeIf { it.isNotBlank() }?.let { error ->
                                Text(
                                    text = error,
                                    style = MaterialTheme.typography.bodySmall,
                                    color = MaterialTheme.colorScheme.error,
                                )
                            }

                            Button(
                                onClick = onUnlock,
                                enabled = !unlockDisabled,
                                modifier = Modifier.fillMaxWidth(),
                            ) {
                                Text(t(Str.RideUnlock))
                            }

                            if (state.session.unlockAttempts >= CHANGE_VEHICLE_HINT_ATTEMPTS) {
                                OutlinedButton(
                                    onClick = onChangeVehicle,
                                    modifier = Modifier.fillMaxWidth(),
                                ) {
                                    Text(t(Str.RideChangeVehicle))
                                }
                            }
                        }
                    }
                }
            }
        }

        if (state.busy && !unlocking) {
            Surface(
                modifier = Modifier.fillMaxSize(),
                color = RiderTheme.colors.pageBackground.copy(alpha = 0.55f),
            ) {
                Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    Text(
                        text = state.busyLabel?.let { t(it) }.orEmpty().ifBlank { t(Str.Loading) },
                        color = RiderTheme.colors.textSecondary,
                    )
                }
            }
        }
    }
}

private fun vehiclePins(state: RidingUiState): List<MapPin> {
    val vehicle = state.vehicle ?: return emptyList()
    if (!vehicle.hasLocation) return emptyList()
    return listOf(
        MapPin(
            id = vehicle.carId.ifBlank { state.carId },
            lat = vehicle.lat,
            lng = vehicle.lng,
            title = vehicle.carId.ifBlank { state.carId },
            restBattery = vehicle.restBattery ?: 0,
        ),
    )
}

private const val PLACEHOLDER = "--"
private const val CHANGE_VEHICLE_HINT_ATTEMPTS = 2
