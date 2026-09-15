package com.luopingtech.ebike.ops.ui.map

import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.platform.MapProviderKind

/**
 * 把共享界面的地图契约接到腾讯地图 SDK 上。
 *
 * 闭源 SDK 只出现在宿主，`sharedUi` 只认 [OpsMapRenderer]；
 * 没有 Key / 未就绪时宿主不注入，共享层自动落到 [SimulatorMapRenderer]。
 */
object TencentOpsMapRenderer : OpsMapRenderer {
    @Composable
    override fun Pins(spec: OpsMapSpec, modifier: Modifier) {
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

    @Composable
    override fun Scatter(spec: OpsScatterMapSpec, modifier: Modifier) {
        ReturnCarScatterMapView(
            points = spec.points,
            showNormal = spec.showNormal,
            showAbnormal = spec.showAbnormal,
            fencePolygons = spec.fencePolygons,
            fitNonce = spec.fitNonce,
            modifier = modifier,
        )
    }
}

/** 厂商地图就绪时用真实实现，否则用共享层的画布实现。 */
fun opsMapRendererFor(app: OpsApp): OpsMapRenderer =
    if (app.mapCapability.kind == MapProviderKind.TENCENT && app.mapCapability.isReady) {
        TencentOpsMapRenderer
    } else {
        SimulatorMapRenderer
    }
