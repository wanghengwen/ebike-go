package com.luopingtech.ebike.ops.ui.map

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.viewinterop.AndroidView
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
import com.luopingtech.ebike.ops.domain.analysis.ReturnCarPoint
import com.luopingtech.ebike.ops.domain.model.FencePolygon
import com.tencent.map.sdk.utilities.visualization.datamodels.ScatterLatLng
import com.tencent.map.sdk.utilities.visualization.scatterplot.DotScatterPlotOverlayProvider
import com.tencent.tencentmap.mapsdk.maps.CameraUpdateFactory
import com.tencent.tencentmap.mapsdk.maps.MapView
import com.tencent.tencentmap.mapsdk.maps.TencentMap
import com.tencent.tencentmap.mapsdk.maps.model.LatLng
import com.tencent.tencentmap.mapsdk.maps.model.LatLngBounds
import com.tencent.tencentmap.mapsdk.maps.model.Polygon
import com.tencent.tencentmap.mapsdk.maps.model.PolygonOptions
import com.tencent.tencentmap.mapsdk.maps.model.VectorOverlay

/**
 * 还车分布专用腾讯地图：围栏 + DotScatter 海量散点（勿用 Marker 硬铺）。
 */
@Composable
fun ReturnCarScatterMapView(
    points: List<ReturnCarPoint>,
    showNormal: Boolean,
    showAbnormal: Boolean,
    fencePolygons: List<FencePolygon> = emptyList(),
    fitNonce: Int = 0,
    modifier: Modifier = Modifier,
) {
    val context = LocalContext.current
    val lifecycleOwner = LocalLifecycleOwner.current
    var mapLoaded by remember { mutableStateOf(false) }
    var tencentMap by remember { mutableStateOf<TencentMap?>(null) }
    val polygons = remember { mutableListOf<Polygon>() }
    var normalOverlay by remember { mutableStateOf<VectorOverlay?>(null) }
    var abnormalOverlay by remember { mutableStateOf<VectorOverlay?>(null) }
    var cameraFittedFor by remember { mutableStateOf<String?>(null) }

    val mapView = remember { MapView(context) }

    DisposableEffect(lifecycleOwner, mapView) {
        val observer = LifecycleEventObserver { _, event ->
            when (event) {
                Lifecycle.Event.ON_START -> mapView.onStart()
                Lifecycle.Event.ON_RESUME -> mapView.onResume()
                Lifecycle.Event.ON_PAUSE -> mapView.onPause()
                Lifecycle.Event.ON_STOP -> mapView.onStop()
                Lifecycle.Event.ON_DESTROY -> mapView.onDestroy()
                else -> Unit
            }
        }
        lifecycleOwner.lifecycle.addObserver(observer)
        if (lifecycleOwner.lifecycle.currentState.isAtLeast(Lifecycle.State.STARTED)) {
            mapView.onStart()
        }
        if (lifecycleOwner.lifecycle.currentState.isAtLeast(Lifecycle.State.RESUMED)) {
            mapView.onResume()
        }
        onDispose {
            lifecycleOwner.lifecycle.removeObserver(observer)
            runCatching {
                polygons.forEach { it.remove() }
                polygons.clear()
                normalOverlay?.remove()
                abnormalOverlay?.remove()
                normalOverlay = null
                abnormalOverlay = null
                mapView.onPause()
                mapView.onStop()
                mapView.onDestroy()
            }
        }
    }

    Box(modifier = modifier) {
        AndroidView(
            factory = {
                mapView.apply {
                    tencentMap = map
                    map.uiSettings.apply {
                        isCompassEnabled = true
                        isMyLocationButtonEnabled = false
                    }
                    map.addOnMapLoadedCallback { mapLoaded = true }
                }
            },
            modifier = Modifier.fillMaxSize(),
        )

        LaunchedEffect(points, showNormal, showAbnormal, mapLoaded, tencentMap, fencePolygons, fitNonce) {
            val map = tencentMap ?: return@LaunchedEffect
            if (!mapLoaded) return@LaunchedEffect
            try {
                polygons.forEach { it.remove() }
                polygons.clear()
                normalOverlay?.remove()
                abnormalOverlay?.remove()
                normalOverlay = null
                abnormalOverlay = null

                fencePolygons.forEach { fence ->
                    if (fence.points.size < 3) return@forEach
                    val opts = PolygonOptions()
                    fence.points.forEach { p -> opts.add(LatLng(p.lat, p.lng)) }
                    opts.fillColor(0x223AA0E8)
                    opts.strokeColor(0xFF3AA0E8.toInt())
                    opts.strokeWidth(3f)
                    polygons.add(map.addPolygon(opts))
                }

                val normalPts = if (showNormal) points.filter { !it.abnormal } else emptyList()
                val abnormalPts = if (showAbnormal) points.filter { it.abnormal } else emptyList()

                if (normalPts.isNotEmpty()) {
                    val nodes = normalPts.map { ScatterLatLng(LatLng(it.lat, it.lng)) }
                    val provider = DotScatterPlotOverlayProvider().apply {
                        data(nodes)
                        radius(10)
                        opacity(1f)
                        colors(intArrayOf(0xFF4CAF50.toInt()))
                        visibility(true)
                    }
                    normalOverlay = map.addVectorOverlay(provider)
                }
                if (abnormalPts.isNotEmpty()) {
                    val nodes = abnormalPts.map { ScatterLatLng(LatLng(it.lat, it.lng)) }
                    val provider = DotScatterPlotOverlayProvider().apply {
                        data(nodes)
                        radius(10)
                        opacity(1f)
                        colors(intArrayOf(0xFFE53935.toInt()))
                        visibility(true)
                    }
                    abnormalOverlay = map.addVectorOverlay(provider)
                }

                val fitPts = (if (showNormal) normalPts else emptyList()) +
                    (if (showAbnormal) abnormalPts else emptyList())
                val fitKey = "${fitPts.size}|f${fencePolygons.size}|n$fitNonce|" +
                    fitPts.take(20).joinToString { "${it.lat},${it.lng}" }
                if (fitKey != cameraFittedFor) {
                    val builder = LatLngBounds.Builder()
                    var has = false
                    fitPts.forEach {
                        builder.include(LatLng(it.lat, it.lng))
                        has = true
                    }
                    fencePolygons.flatMap { it.points }.forEach {
                        builder.include(LatLng(it.lat, it.lng))
                        has = true
                    }
                    if (has) {
                        map.moveCamera(CameraUpdateFactory.newLatLngBounds(builder.build(), 80))
                        cameraFittedFor = fitKey
                    }
                }
            } catch (_: Throwable) {
                // 鉴权失败或 utilities 不兼容时保持空图，由外层提示
            }
        }
    }
}
