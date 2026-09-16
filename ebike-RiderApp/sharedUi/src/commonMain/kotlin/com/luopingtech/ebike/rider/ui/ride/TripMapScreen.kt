package com.luopingtech.ebike.rider.ui.ride

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.luopingtech.ebike.rider.RiderApp
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.domain.model.TrackPoint
import com.luopingtech.ebike.rider.domain.riding.RideFormat
import com.luopingtech.ebike.rider.feature.riding.RidingUiState
import com.luopingtech.ebike.rider.platform.MapProviderKind
import com.luopingtech.ebike.rider.ui.map.RiderMapSpec
import com.luopingtech.ebike.rider.ui.map.RiderMapView
import com.luopingtech.ebike.rider.ui.theme.RiderTheme

/**
 * 本次行程轨迹（简化版）。
 *
 * 轨迹是 `RidingFeature` 用 `TrackPointBuffer` 在本地抽稀攒出来的，**不**是服务端回放 ——
 * 服务端轨迹来自车机 GPS，前端再传一份只会打架。所以刚开锁时这里会是空的。
 */
@Composable
fun TripMapScreen(
    app: RiderApp,
    state: RidingUiState,
    onBack: () -> Unit,
    modifier: Modifier = Modifier,
) {
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)

    val providerLabel = when (app.mapCapability.kind) {
        MapProviderKind.TENCENT -> t(Str.TencentMap)
        MapProviderKind.SIMULATOR -> t(Str.SimulatorMap)
        else -> app.mapCapability.kind.name
    }

    RideScaffold(
        title = t(Str.TripMapTitle),
        onBack = onBack,
        backLabel = t(Str.Back),
        modifier = modifier,
    ) {
        Box(modifier = Modifier.fillMaxSize()) {
            RiderMapView(
                spec = RiderMapSpec(
                    providerLabel = providerLabel,
                    clusterOverview = false,
                    fencePolygons = state.fences,
                    trackPoints = state.track.map { TrackPoint(lat = it.lat, lng = it.lng) },
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
                    verticalArrangement = Arrangement.spacedBy(4.dp),
                ) {
                    if (state.track.isEmpty()) {
                        Text(
                            text = t(Str.TripMapEmpty),
                            style = MaterialTheme.typography.bodyMedium,
                            color = RiderTheme.colors.textSecondary,
                        )
                    }
                    RideDetailRow(
                        label = t(Str.RideDuration),
                        value = RideFormat.duration(state.elapsedSeconds),
                    )
                    RideDetailRow(
                        label = t(Str.RideDistance),
                        value = RideFormat.distance(state.rideInfo?.rideDistanceMeters ?: 0),
                    )
                    RideDetailRow(
                        label = t(Str.RideCost),
                        value = "${RideFormat.yuan(state.rideInfo?.costFeeFen ?: 0)} ${t(Str.RideYuan)}",
                        emphasize = true,
                    )
                }
            }
        }
    }
}
