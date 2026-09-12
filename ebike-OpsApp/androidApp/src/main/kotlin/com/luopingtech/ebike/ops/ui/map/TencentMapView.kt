package com.luopingtech.ebike.ops.ui.map

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.AndroidView
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.domain.map.MapClusterer
import com.luopingtech.ebike.ops.domain.map.MapProjection
import com.luopingtech.ebike.ops.domain.model.FencePolygon
import com.luopingtech.ebike.ops.domain.model.MapPin
import com.luopingtech.ebike.ops.domain.model.TrackPoint
import com.tencent.tencentmap.mapsdk.maps.CameraUpdateFactory
import com.tencent.tencentmap.mapsdk.maps.MapView
import com.tencent.tencentmap.mapsdk.maps.TencentMap
import com.tencent.tencentmap.mapsdk.maps.model.BitmapDescriptorFactory
import com.tencent.tencentmap.mapsdk.maps.model.LatLng
import com.tencent.tencentmap.mapsdk.maps.model.LatLngBounds
import com.tencent.tencentmap.mapsdk.maps.model.Marker
import com.tencent.tencentmap.mapsdk.maps.model.MarkerOptions
import com.tencent.tencentmap.mapsdk.maps.model.Polygon
import com.tencent.tencentmap.mapsdk.maps.model.PolygonOptions
import com.tencent.tencentmap.mapsdk.maps.model.Polyline
import com.tencent.tencentmap.mapsdk.maps.model.PolylineOptions

/**
 * Tencent [MapView] showing vehicle pins (aligned with legacy TencentMapView).
 */
@Composable
fun TencentMapView(
    pins: List<MapPin>,
    selectedCarId: String?,
    onSelectCarId: (String) -> Unit,
    onSelectCluster: (List<String>) -> Unit = {},
    modifier: Modifier = Modifier,
    clusterOverview: Boolean = true,
    fencePolygons: List<FencePolygon> = emptyList(),
    trackPoints: List<TrackPoint> = emptyList(),
) {
    val context = LocalContext.current
    val lifecycleOwner = LocalLifecycleOwner.current
    var mapLoaded by remember { mutableStateOf(false) }
    var mapError by remember { mutableStateOf<String?>(null) }
    var tencentMap by remember { mutableStateOf<TencentMap?>(null) }
    val markers = remember { mutableListOf<Marker>() }
    val polygons = remember { mutableListOf<Polygon>() }
    var polyline by remember { mutableStateOf<Polyline?>(null) }
    var cameraFittedFor by remember { mutableStateOf<String?>(null) }
    val selectCarIdUpdated by rememberUpdatedState(onSelectCarId)
    val selectClusterUpdated by rememberUpdatedState(onSelectCluster)

    val displayPins = remember(pins, clusterOverview) {
        if (!clusterOverview || pins.size <= 4) {
            pins
        } else {
            val bounds = MapProjection.boundsOf(pins)
            val cell = MapClusterer.suggestedCellDegrees(bounds, targetCells = 5)
            MapClusterer.cluster(pins, cell)
        }
    }

    val mapView = remember {
        MapView(context)
    }

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
                markers.forEach { it.remove() }
                markers.clear()
                polygons.forEach { it.remove() }
                polygons.clear()
                polyline?.remove()
                polyline = null
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
                    val map = map
                    tencentMap = map
                    map.uiSettings.apply {
                        isCompassEnabled = true
                        isMyLocationButtonEnabled = false
                    }
                    map.addOnMapLoadedCallback {
                        mapLoaded = true
                    }
                }
            },
            modifier = Modifier.fillMaxSize(),
        )

        LaunchedEffect(tencentMap, selectCarIdUpdated, selectClusterUpdated) {
            val map = tencentMap ?: return@LaunchedEffect
            map.setOnMarkerClickListener { marker ->
                val tag = marker.tag as? MarkerTag ?: return@setOnMarkerClickListener false
                if (tag.isCluster) {
                    // Legacy BaseHomeMapFragment: zoom+1 under 16, else expand cluster list.
                    val zoom = map.cameraPosition.zoom
                    if (zoom < 16f) {
                        map.animateCamera(
                            CameraUpdateFactory.newLatLngZoom(marker.position, zoom + 1.2f),
                        )
                    } else {
                        selectClusterUpdated(tag.memberIds)
                    }
                } else {
                    selectCarIdUpdated(tag.id)
                }
                true
            }
        }

        LaunchedEffect(displayPins, selectedCarId, mapLoaded, tencentMap, fencePolygons, trackPoints) {
            val map = tencentMap ?: return@LaunchedEffect
            if (!mapLoaded) return@LaunchedEffect
            try {
                markers.forEach { it.remove() }
                markers.clear()
                polygons.forEach { it.remove() }
                polygons.clear()
                polyline?.remove()
                polyline = null

                fencePolygons.forEach { fence ->
                    if (fence.points.size < 3) return@forEach
                    val opts = PolygonOptions()
                    fence.points.forEach { p -> opts.add(LatLng(p.lat, p.lng)) }
                    opts.fillColor(0x223AA0E8)
                    opts.strokeColor(0xFF3AA0E8.toInt())
                    opts.strokeWidth(3f)
                    polygons.add(map.addPolygon(opts))
                }

                if (trackPoints.size >= 2) {
                    val line = PolylineOptions()
                    trackPoints.forEach { p -> line.add(LatLng(p.lat, p.lng)) }
                    line.color(0xFFE67E22.toInt())
                    line.width(8f)
                    polyline = map.addPolyline(line)
                }

                val valid = displayPins.filter { it.lat != 0.0 || it.lng != 0.0 }
                valid.forEach { pin ->
                    val selected = !pin.isCluster && pin.memberIds.contains(selectedCarId)
                    val hue = when {
                        pin.isCluster -> BitmapDescriptorFactory.HUE_VIOLET
                        selected -> BitmapDescriptorFactory.HUE_AZURE
                        pin.restBattery in 1..30 -> BitmapDescriptorFactory.HUE_ORANGE
                        pin.ridingState == 1 -> BitmapDescriptorFactory.HUE_GREEN
                        else -> BitmapDescriptorFactory.HUE_BLUE
                    }
                    val title = if (pin.isCluster) {
                        Strings.t(Str.ClusterVehicles, pin.memberCount)
                    } else {
                        pin.title
                    }
                    val marker = map.addMarker(
                        MarkerOptions(LatLng(pin.lat, pin.lng))
                            .title(title)
                            .snippet(pin.subtitle)
                            .icon(BitmapDescriptorFactory.defaultMarker(hue))
                            .zIndex(if (selected) 10f else 1f),
                    )
                    marker.tag = MarkerTag(
                        id = pin.id,
                        isCluster = pin.isCluster,
                        memberIds = pin.memberIds.ifEmpty { listOf(pin.id) },
                    )
                    markers.add(marker)
                    if (selected) marker.showInfoWindow()
                }

                val fitKey = valid.joinToString("|") { "${it.id}:${it.lat},${it.lng}" } +
                    "|f${fencePolygons.size}|t${trackPoints.size}"
                if (valid.isNotEmpty() && fitKey != cameraFittedFor) {
                    if (valid.size == 1 && fencePolygons.isEmpty() && trackPoints.isEmpty()) {
                        map.moveCamera(
                            CameraUpdateFactory.newLatLngZoom(
                                LatLng(valid.first().lat, valid.first().lng),
                                15f,
                            ),
                        )
                    } else {
                        val builder = LatLngBounds.Builder()
                        valid.forEach { builder.include(LatLng(it.lat, it.lng)) }
                        trackPoints.forEach { builder.include(LatLng(it.lat, it.lng)) }
                        fencePolygons.flatMap { it.points }.forEach {
                            builder.include(LatLng(it.lat, it.lng))
                        }
                        map.moveCamera(
                            CameraUpdateFactory.newLatLngBounds(builder.build(), 80),
                        )
                    }
                    cameraFittedFor = fitKey
                } else if (selectedCarId != null) {
                    val selected = valid.firstOrNull {
                        !it.isCluster && it.memberIds.contains(selectedCarId)
                    }
                    if (selected != null) {
                        map.animateCamera(
                            CameraUpdateFactory.newLatLngZoom(
                                LatLng(selected.lat, selected.lng),
                                16f,
                            ),
                        )
                    }
                }
                mapError = null
            } catch (t: Throwable) {
                mapError = t.message ?: Strings.t(Str.TencentMapRenderFailed)
            }
        }

        Text(
            text = when {
                mapError != null -> mapError!!
                !mapLoaded -> Strings.t(Str.TencentMapLoading)
                pins.isEmpty() -> Strings.t(Str.TencentMapNoPins)
                else -> Strings.t(Str.TencentMapPins, displayPins.size, pins.size)
            },
            modifier = Modifier
                .align(Alignment.TopStart)
                .padding(10.dp),
            style = MaterialTheme.typography.labelMedium,
            color = if (mapError != null) {
                MaterialTheme.colorScheme.error
            } else {
                MaterialTheme.colorScheme.onSurface
            },
        )
    }
}

private data class MarkerTag(
    val id: String,
    val isCluster: Boolean,
    val memberIds: List<String>,
)
