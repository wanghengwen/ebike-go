package com.luopingtech.ebike.rider.ui.map

import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import com.luopingtech.ebike.rider.RiderApp
import com.luopingtech.ebike.rider.platform.MapProviderKind
import com.luopingtech.ebike.rider.ui.map.RiderMapSpec
import com.luopingtech.ebike.rider.ui.map.RiderMapRenderer
import com.luopingtech.ebike.rider.ui.map.SimulatorMapRenderer

object TencentRiderMapRenderer : RiderMapRenderer {
    @Composable
    override fun Pins(spec: RiderMapSpec, modifier: Modifier) {
        TencentMapView(
            pins = spec.pins,
            selectedCarId = spec.selectedCarId,
            onSelectCarId = spec.onSelectCarId,
            onSelectCluster = spec.onSelectCluster,
            modifier = modifier,
            clusterOverview = spec.clusterOverview,
            fencePolygons = spec.fencePolygons,
            trackPoints = spec.trackPoints,
            fitNonce = spec.fitNonce,
            zoomInNonce = spec.zoomInNonce,
            zoomOutNonce = spec.zoomOutNonce,
            followNonce = spec.followNonce,
            followLat = spec.followLat,
            followLng = spec.followLng,
            mapTypeSatellite = spec.mapTypeSatellite,
            showStatusOverlay = spec.showStatusOverlay,
        )
    }
}

fun riderMapRendererFor(app: RiderApp): RiderMapRenderer =
    if (app.mapCapability.kind == MapProviderKind.TENCENT && app.mapCapability.isReady) {
        TencentRiderMapRenderer
    } else {
        SimulatorMapRenderer
    }
