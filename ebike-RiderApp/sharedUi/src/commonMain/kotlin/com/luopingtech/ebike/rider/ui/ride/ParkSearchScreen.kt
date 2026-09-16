package com.luopingtech.ebike.rider.ui.ride

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.luopingtech.ebike.rider.RiderApp
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.domain.model.MapPin
import com.luopingtech.ebike.rider.domain.riding.NearParking
import com.luopingtech.ebike.rider.feature.riding.RidingUiState
import com.luopingtech.ebike.rider.platform.MapProviderKind
import com.luopingtech.ebike.rider.ui.map.RiderMapSpec
import com.luopingtech.ebike.rider.ui.map.RiderMapView
import com.luopingtech.ebike.rider.ui.theme.RiderTheme

/**
 * 找附近停车点（对应 UniApp 的「找 P 点」）。
 *
 * 三种结果要分开说，否则用户分不清是「附近真没有」还是「这儿根本不在运营区」：
 * 服务区外、区内但无空位、有目标点。
 */
@Composable
fun ParkSearchScreen(
    app: RiderApp,
    state: RidingUiState,
    onRefresh: () -> Unit,
    onBack: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val parking = state.nearParking

    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)

    LaunchedEffect(Unit) { onRefresh() }

    val providerLabel = when (app.mapCapability.kind) {
        MapProviderKind.TENCENT -> t(Str.TencentMap)
        MapProviderKind.SIMULATOR -> t(Str.SimulatorMap)
        else -> app.mapCapability.kind.name
    }

    RideScaffold(
        title = t(Str.ParkSearchTitle),
        onBack = onBack,
        backLabel = t(Str.Back),
        busy = state.busy,
        busyLabel = state.busyLabel?.let { t(it) }.orEmpty(),
        modifier = modifier,
    ) {
        Box(modifier = Modifier.fillMaxSize()) {
            RiderMapView(
                spec = RiderMapSpec(
                    pins = parkingPins(parking, t(Str.ParkSearchTitle)),
                    providerLabel = providerLabel,
                    clusterOverview = false,
                    fencePolygons = state.fences,
                    showStatusOverlay = false,
                ),
                modifier = Modifier.fillMaxSize(),
            )

            Surface(
                tonalElevation = 2.dp,
                shape = MaterialTheme.shapes.medium,
                modifier = Modifier
                    .align(Alignment.BottomCenter)
                    .fillMaxWidth()
                    .padding(16.dp),
            ) {
                Column(
                    modifier = Modifier.padding(16.dp),
                    verticalArrangement = Arrangement.spacedBy(6.dp),
                ) {
                    when {
                        parking == null -> Text(
                            text = t(Str.Loading),
                            style = MaterialTheme.typography.bodyMedium,
                            color = RiderTheme.colors.textSecondary,
                        )

                        parking.outOfService -> {
                            Text(
                                text = t(Str.ParkSearchOutOfService),
                                style = MaterialTheme.typography.bodyMedium,
                                color = MaterialTheme.colorScheme.error,
                            )
                            Text(
                                text = t(Str.ParkSearchReselect),
                                style = MaterialTheme.typography.bodySmall,
                                color = RiderTheme.colors.textTertiary,
                            )
                        }

                        parking.count <= 0 -> Text(
                            text = t(Str.ParkSearchEmpty),
                            style = MaterialTheme.typography.bodyMedium,
                            color = RiderTheme.colors.textSecondary,
                        )

                        else -> {
                            Text(
                                text = t(Str.ParkSearchNearCount, parking.count),
                                style = MaterialTheme.typography.titleSmall,
                                color = RiderTheme.colors.textPrimary,
                            )
                            parking.name.takeIf { it.isNotBlank() }?.let {
                                Text(
                                    text = it,
                                    style = MaterialTheme.typography.bodySmall,
                                    color = RiderTheme.colors.textSecondary,
                                )
                            }
                            parking.distanceMeters?.let { meters ->
                                Text(
                                    text = t(Str.ParkSearchDistance, meters.toInt()),
                                    style = MaterialTheme.typography.bodySmall,
                                    color = RiderTheme.colors.textTertiary,
                                )
                            }
                        }
                    }
                    Button(
                        onClick = onRefresh,
                        enabled = !state.busy,
                        modifier = Modifier.fillMaxWidth(),
                    ) {
                        Text(t(Str.ReturnRefreshLocation))
                    }
                }
            }
        }
    }
}

private fun parkingPins(parking: NearParking?, title: String): List<MapPin> {
    if (parking == null || !parking.hasTarget) return emptyList()
    return listOf(
        MapPin(
            id = "parking",
            lat = parking.lat ?: 0.0,
            lng = parking.lng ?: 0.0,
            title = parking.name.ifBlank { title },
        ),
    )
}
