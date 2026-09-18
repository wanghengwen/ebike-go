package com.luopingtech.ebike.ops.ui.map

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.toArgb
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalDensity
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
import com.luopingtech.ebike.ops.domain.model.MapPinIcon
import com.luopingtech.ebike.ops.domain.model.TrackPoint
import com.luopingtech.ebike.ops.domain.order.ORDER_PLAYBACK_PIN_ID
import com.luopingtech.ebike.ops.R
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
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
    /** Theme color for cluster count bubbles (legacy DefaultOptionGenerator). */
    clusterFillColor: Color = OpsTheme.colors.primary,
    /** 递增后强制重新 fit 视野（看全部）。 */
    fitNonce: Int = 0,
    /** 递增后放大一级。 */
    zoomInNonce: Int = 0,
    /** 递增后缩小一级。 */
    zoomOutNonce: Int = 0,
    /** 递增后相机跟到 [followLat]/[followLng]（定位到我）。 */
    followNonce: Int = 0,
    followLat: Double? = null,
    followLng: Double? = null,
    followZoom: Float = 17f,
    mapTypeSatellite: Boolean = false,
    showStatusOverlay: Boolean = true,
    /** false 时不因 pins 变化自动 fit（扫码定位等仅用 followNonce 飞点）。 */
    autoFitOnPins: Boolean = true,
    /** 选中车辆后是否飞到该车（首页对齐原版不跟飞）。 */
    animateToSelection: Boolean = true,
    onMapTap: ((lat: Double, lng: Double) -> Unit)? = null,
    onCameraIdle: ((lat: Double, lng: Double) -> Unit)? = null,
    onCameraMove: ((lat: Double, lng: Double) -> Unit)? = null,
    screenToLatLngNonce: Int = 0,
    screenPickX: Float = 0f,
    screenPickY: Float = 0f,
    onScreenToLatLng: ((lat: Double, lng: Double) -> Unit)? = null,
    batchScreenToLatLngNonce: Int = 0,
    batchScreenPoints: List<Pair<Float, Float>> = emptyList(),
    onBatchScreenToLatLng: ((List<Pair<Double, Double>>) -> Unit)? = null,
    latLngToScreenNonce: Int = 0,
    latLngToScreenPoints: List<Pair<Double, Double>> = emptyList(),
    onLatLngToScreen: ((List<Pair<Float, Float>>) -> Unit)? = null,
) {
    val context = LocalContext.current
    val density = LocalDensity.current.density
    val clusterArgb = clusterFillColor.toArgb()
    val lifecycleOwner = LocalLifecycleOwner.current
    var mapLoaded by remember { mutableStateOf(false) }
    var mapError by remember { mutableStateOf<String?>(null) }
    var tencentMap by remember { mutableStateOf<TencentMap?>(null) }
    val markers = remember { mutableListOf<Marker>() }
    val polygons = remember { mutableListOf<Polygon>() }
    var polyline by remember { mutableStateOf<Polyline?>(null) }
    var cameraFittedFor by remember { mutableStateOf<String?>(null) }
    /** 首页 !autoFitOnPins 时只认 fitNonce 递增，避免刷车重 fit。 */
    var lastForcedFitNonce by remember { mutableStateOf(0) }
    val selectCarIdUpdated by rememberUpdatedState(onSelectCarId)
    val selectClusterUpdated by rememberUpdatedState(onSelectCluster)
    val mapTapUpdated by rememberUpdatedState(onMapTap)
    val cameraIdleUpdated by rememberUpdatedState(onCameraIdle)
    val cameraMoveUpdated by rememberUpdatedState(onCameraMove)
    val screenToLatLngUpdated by rememberUpdatedState(onScreenToLatLng)
    val batchScreenToLatLngUpdated by rememberUpdatedState(onBatchScreenToLatLng)
    val latLngToScreenUpdated by rememberUpdatedState(onLatLngToScreen)

    // 聚合只认「相机停下」后的视野（对齐 AMapClusterManagerV3：仅 onCameraIdle 才 assignClusters），
    // 缩放过程中不重算，否则会在聚合/单点之间来回闪。
    var cameraZoom by remember { mutableStateOf(12f) }
    var visibleBounds by remember {
        mutableStateOf<com.luopingtech.ebike.ops.domain.map.LatLngBounds?>(null)
    }
    var cameraCenter by remember { mutableStateOf(28.22 to 112.94) }
    val displayPins = remember(pins, clusterOverview, cameraZoom, visibleBounds, cameraCenter) {
        val lat = cameraCenter.first
        val lng = cameraCenter.second
        if (!clusterOverview) {
            val near = MapClusterer.filterNearCenter(pins, lat, lng)
            val inView = visibleBounds?.let { MapClusterer.pinsInBounds(near, it.padded(0.08)) } ?: near
            inView.map { it.copy(memberCount = 1, memberIds = listOf(it.id), showCluster = false) }
        } else if (pins.size <= 1) {
            // 对齐原版：聚合模式下单车也是「1」的数字气泡
            pins.map { it.copy(memberCount = 1, memberIds = listOf(it.id), showCluster = true) }
        } else {
            MapClusterer.clusterInViewport(
                pins = pins,
                cellDegrees = MapClusterer.cellDegreesForZoom(
                    zoom = cameraZoom,
                    latitude = pins.firstOrNull { it.lat != 0.0 }?.lat ?: lat,
                ),
                visible = visibleBounds,
            )
        }
    }
    /** 内容一致时不重建 marker，避免相机停下后无谓地删一遍再加一遍（视觉上就是闪一下）。 */
    val displayPinsKey = remember(displayPins) {
        displayPins.joinToString("|") { p ->
            "${p.id}@${p.lat},${p.lng}#${p.memberCount}${if (p.showCluster) "c" else ""}" +
                ":${p.icon}:${p.badgeDrawableName}"
        }
    }

    fun refreshVisibleRegion(map: TencentMap) {
        runCatching {
            val region = map.projection?.visibleRegion?.latLngBounds ?: return
            visibleBounds = com.luopingtech.ebike.ops.domain.map.LatLngBounds(
                minLat = region.southwest.latitude,
                maxLat = region.northeast.latitude,
                minLng = region.southwest.longitude,
                maxLng = region.northeast.longitude,
            )
            map.cameraPosition?.target?.let { cameraCenter = it.latitude to it.longitude }
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
                        cameraZoom = map.cameraPosition.zoom
                        refreshVisibleRegion(map)
                    }
                    map.setOnCameraChangeListener(object : TencentMap.OnCameraChangeListener {
                        override fun onCameraChange(cameraPosition: com.tencent.tencentmap.mapsdk.maps.model.CameraPosition?) {
                            cameraPosition?.let {
                                cameraMoveUpdated?.invoke(it.target.latitude, it.target.longitude)
                            }
                        }
                        override fun onCameraChangeFinished(cameraPosition: com.tencent.tencentmap.mapsdk.maps.model.CameraPosition?) {
                            cameraPosition?.let {
                                cameraZoom = it.zoom
                                cameraCenter = it.target.latitude to it.target.longitude
                            }
                            refreshVisibleRegion(map)
                        }
                    })
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
                            CameraUpdateFactory.newLatLngZoom(marker.position, zoom + 1f),
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

        LaunchedEffect(tencentMap, mapTapUpdated) {
            val map = tencentMap ?: return@LaunchedEffect
            val tap = mapTapUpdated
            if (tap == null) {
                map.setOnMapClickListener(null)
            } else {
                map.setOnMapClickListener { latLng ->
                    tap(latLng.latitude, latLng.longitude)
                }
            }
        }

        LaunchedEffect(tencentMap, mapLoaded, cameraIdleUpdated, cameraMoveUpdated) {
            val map = tencentMap ?: return@LaunchedEffect
            if (!mapLoaded) return@LaunchedEffect
            map.setOnCameraChangeListener(object : TencentMap.OnCameraChangeListener {
                override fun onCameraChange(cameraPosition: com.tencent.tencentmap.mapsdk.maps.model.CameraPosition?) {
                    // 对齐原版 onCameraIdle(false) → ON_MOVE → invalidate 重投影；聚合不在这里重算
                    cameraPosition?.let {
                        cameraMoveUpdated?.invoke(it.target.latitude, it.target.longitude)
                    }
                }
                override fun onCameraChangeFinished(cameraPosition: com.tencent.tencentmap.mapsdk.maps.model.CameraPosition?) {
                    cameraPosition?.let {
                        cameraZoom = it.zoom
                        cameraCenter = it.target.latitude to it.target.longitude
                        cameraIdleUpdated?.invoke(it.target.latitude, it.target.longitude)
                    }
                    refreshVisibleRegion(map)
                }
            })
            map.cameraPosition?.target?.let { target ->
                cameraIdleUpdated?.invoke(target.latitude, target.longitude)
            }
        }

        LaunchedEffect(tencentMap, mapLoaded, screenToLatLngNonce, screenPickX, screenPickY) {
            val map = tencentMap ?: return@LaunchedEffect
            if (!mapLoaded || screenToLatLngNonce <= 0) return@LaunchedEffect
            val cb = screenToLatLngUpdated ?: return@LaunchedEffect
            val projection = map.projection ?: return@LaunchedEffect
            val latLng = projection.fromScreenLocation(
                android.graphics.Point(screenPickX.toInt(), screenPickY.toInt()),
            ) ?: return@LaunchedEffect
            cb(latLng.latitude, latLng.longitude)
        }

        LaunchedEffect(tencentMap, mapLoaded, batchScreenToLatLngNonce, batchScreenPoints) {
            val map = tencentMap ?: return@LaunchedEffect
            if (!mapLoaded || batchScreenToLatLngNonce <= 0 || batchScreenPoints.isEmpty()) return@LaunchedEffect
            val cb = batchScreenToLatLngUpdated ?: return@LaunchedEffect
            val projection = map.projection ?: return@LaunchedEffect
            val mapped = batchScreenPoints.mapNotNull { (x, y) ->
                val latLng = projection.fromScreenLocation(
                    android.graphics.Point(x.toInt(), y.toInt()),
                ) ?: return@mapNotNull null
                latLng.latitude to latLng.longitude
            }
            if (mapped.size == batchScreenPoints.size) cb(mapped)
        }

        LaunchedEffect(tencentMap, mapLoaded, latLngToScreenNonce, latLngToScreenPoints) {
            val map = tencentMap ?: return@LaunchedEffect
            if (!mapLoaded || latLngToScreenNonce <= 0) return@LaunchedEffect
            val cb = latLngToScreenUpdated ?: return@LaunchedEffect
            val projection = map.projection ?: return@LaunchedEffect
            val mapped = latLngToScreenPoints.map { (lat, lng) ->
                val point = projection.toScreenLocation(LatLng(lat, lng))
                (point?.x?.toFloat() ?: 0f) to (point?.y?.toFloat() ?: 0f)
            }
            cb(mapped)
        }

        LaunchedEffect(tencentMap, mapTypeSatellite) {
            val map = tencentMap ?: return@LaunchedEffect
            map.mapType = if (mapTypeSatellite) {
                TencentMap.MAP_TYPE_SATELLITE
            } else {
                TencentMap.MAP_TYPE_NORMAL
            }
        }

        LaunchedEffect(tencentMap, mapLoaded, zoomInNonce) {
            val map = tencentMap ?: return@LaunchedEffect
            if (!mapLoaded || zoomInNonce <= 0) return@LaunchedEffect
            val next = (map.cameraPosition.zoom + 1f).coerceAtMost(20f)
            map.animateCamera(CameraUpdateFactory.zoomTo(next))
        }

        LaunchedEffect(tencentMap, mapLoaded, zoomOutNonce) {
            val map = tencentMap ?: return@LaunchedEffect
            if (!mapLoaded || zoomOutNonce <= 0) return@LaunchedEffect
            val next = (map.cameraPosition.zoom - 1f).coerceAtLeast(3f)
            map.animateCamera(CameraUpdateFactory.zoomTo(next))
        }

        LaunchedEffect(tencentMap, mapLoaded, followNonce, followLat, followLng, followZoom) {
            val map = tencentMap ?: return@LaunchedEffect
            val lat = followLat
            val lng = followLng
            if (!mapLoaded || followNonce <= 0 || lat == null || lng == null) return@LaunchedEffect
            map.animateCamera(
                CameraUpdateFactory.newLatLngZoom(LatLng(lat, lng), followZoom),
            )
        }

        // 围栏 / 轨迹只随自身数据重建，不跟着聚合结果一起删了重画
        LaunchedEffect(tencentMap, mapLoaded, fencePolygons, trackPoints) {
            val map = tencentMap ?: return@LaunchedEffect
            if (!mapLoaded) return@LaunchedEffect
            try {
                polygons.forEach { it.remove() }
                polygons.clear()
                polyline?.remove()
                polyline = null

                fencePolygons.forEach { fence ->
                    if (fence.points.size < 3) return@forEach
                    val opts = PolygonOptions()
                    fence.points.forEach { p -> opts.add(LatLng(p.lat, p.lng)) }
                    // Legacy ParkingFenceProvide styles.
                    when (fence.kind) {
                        com.luopingtech.ebike.ops.domain.model.FenceKind.ServiceArea -> {
                            opts.fillColor(0x0F295FCC)
                            opts.strokeColor(0xFF1180F9.toInt())
                            opts.strokeWidth(8f)
                            opts.pattern(listOf(30, 20))
                        }
                        com.luopingtech.ebike.ops.domain.model.FenceKind.Parking -> {
                            opts.fillColor(0x33242F57)
                            opts.strokeColor(0xFF636E95.toInt())
                            opts.strokeWidth(3f)
                        }
                        com.luopingtech.ebike.ops.domain.model.FenceKind.NoParking -> {
                            opts.fillColor(0x54E02020)
                            opts.strokeColor(0xFFFF0808.toInt())
                            opts.strokeWidth(3f)
                        }
                        else -> {
                            opts.fillColor(0x223AA0E8)
                            opts.strokeColor(0xFF3AA0E8.toInt())
                            opts.strokeWidth(3f)
                        }
                    }
                    polygons.add(map.addPolygon(opts))
                }

                if (trackPoints.size >= 2) {
                    val line = PolylineOptions()
                    trackPoints.forEach { p -> line.add(LatLng(p.lat, p.lng)) }
                    // 对齐 TrackDataHelp.handleDeviceTrackOption：0xff0BB774
                    line.color(0xFF0BB774.toInt())
                    line.width(8f)
                    polyline = map.addPolyline(line)
                }
                mapError = null
            } catch (t: Throwable) {
                mapError = t.message ?: Strings.t(Str.TencentMapRenderFailed)
            }
        }

        // marker 只在聚合结果真的变化时重建（对齐原版 onCameraIdle → ClusterRenderer.render）
        LaunchedEffect(tencentMap, mapLoaded, displayPinsKey, clusterArgb, density) {
            val map = tencentMap ?: return@LaunchedEffect
            if (!mapLoaded) return@LaunchedEffect
            try {
                markers.forEach { it.remove() }
                markers.clear()

                val valid = displayPins.filter { it.lat != 0.0 || it.lng != 0.0 }
                valid.forEach { pin ->
                    // Legacy MapConfig.MARK_ZOOM = 999; selected vehicle does not change icon/zIndex.
                    val options = MarkerOptions(LatLng(pin.lat, pin.lng)).zIndex(999f)
                    // 对齐 DefaultOptionGenerator：size>1 或聚合模式的单点都画数字气泡
                    if (pin.isClusterBubble) {
                        val bitmap = ClusterMarkerBitmap.obtain(
                            count = pin.memberCount.coerceAtLeast(pin.memberIds.size),
                            fillColorArgb = clusterArgb,
                            density = density,
                        )
                        options
                            .icon(BitmapDescriptorFactory.fromBitmap(bitmap))
                            .anchor(0.5f, 0.5f)
                    } else {
                        val vehicleRes = legacyVehicleDrawable(pin)
                        val badgeRes = BadgeMarkerBitmap.drawableId(context, pin.badgeDrawableName)
                        val combined = if (badgeRes != 0) {
                            BadgeMarkerBitmap.obtain(context, vehicleRes, badgeRes)
                        } else {
                            null
                        }
                        if (combined != null) {
                            options
                                .icon(BitmapDescriptorFactory.fromBitmap(combined))
                                .anchor(0.5f, 0.5f)
                        } else {
                            when (pin.icon) {
                                MapPinIcon.UserStart -> {
                                    options
                                        .icon(BitmapDescriptorFactory.fromResource(R.drawable.icon_user_start))
                                        .anchor(0.5f, 1f)
                                }
                                MapPinIcon.UserEnd -> {
                                    options
                                        .icon(BitmapDescriptorFactory.fromResource(R.drawable.icon_user_end))
                                        .anchor(0.5f, 1f)
                                }
                                MapPinIcon.TrackOrigin -> {
                                    options
                                        .icon(BitmapDescriptorFactory.fromResource(R.drawable.btn_trajectory_origin))
                                        .anchor(0.5f, 1f)
                                        .zIndex(1000f)
                                }
                                MapPinIcon.TrackEnd -> {
                                    options
                                        .icon(BitmapDescriptorFactory.fromResource(R.drawable.btn_trajectory_end))
                                        .anchor(0.5f, 1f)
                                        .zIndex(1000f)
                                }
                                MapPinIcon.Parking,
                                MapPinIcon.ParkingFunction,
                                MapPinIcon.ParkingHidden,
                                MapPinIcon.NoParking,
                                -> {
                                    options
                                        .icon(BitmapDescriptorFactory.fromResource(vehicleRes))
                                        .anchor(0.5f, 1f)
                                }
                                MapPinIcon.VehicleRiding -> {
                                    val bottomAnchor = pin.id == ORDER_PLAYBACK_PIN_ID
                                    options
                                        .icon(BitmapDescriptorFactory.fromResource(R.drawable.icon_vehicle_riding))
                                        .anchor(0.5f, if (bottomAnchor) 1f else 0.5f)
                                    if (bottomAnchor) options.zIndex(1001f)
                                }
                                else -> {
                                    options
                                        .icon(BitmapDescriptorFactory.fromResource(vehicleRes))
                                        .anchor(0.5f, 0.5f)
                                }
                            }
                        }
                    }
                    val marker = map.addMarker(options)
                    marker.tag = MarkerTag(
                        id = pin.id,
                        isCluster = pin.isCluster,
                        memberIds = pin.memberIds.ifEmpty { listOf(pin.id) },
                    )
                    markers.add(marker)
                }
                mapError = null
            } catch (t: Throwable) {
                mapError = t.message ?: Strings.t(Str.TencentMapRenderFailed)
            }
        }

        LaunchedEffect(tencentMap, mapLoaded, pins, fencePolygons, trackPoints, fitNonce, autoFitOnPins) {
            val map = tencentMap ?: return@LaunchedEffect
            if (!mapLoaded) return@LaunchedEffect
            try {
                val sourcePins = pins.filter {
                    (it.lat != 0.0 || it.lng != 0.0) && it.id != ORDER_PLAYBACK_PIN_ID
                }
                val fencePoints = fencePolygons.flatMap { it.points }
                // autoFitOnPins：随车点变化 fit
                // !autoFitOnPins：仅在 fitNonce 递增时 fit（首页服务区），刷车点不跳动
                val fitKey = sourcePins.joinToString("|") { "${it.id}:${it.lat},${it.lng}" } +
                    "|f${fencePolygons.size}|t${trackPoints.size}|n$fitNonce"
                val shouldFit = if (autoFitOnPins) {
                    sourcePins.isNotEmpty() && fitKey != cameraFittedFor
                } else {
                    fitNonce > lastForcedFitNonce &&
                        (sourcePins.isNotEmpty() || fencePoints.isNotEmpty() || trackPoints.isNotEmpty())
                }
                if (shouldFit) {
                    // 首页服务区 fit：优先围栏；否则车点 / 轨迹
                    val hasFence = fencePoints.isNotEmpty()
                    val hasTrack = trackPoints.isNotEmpty()
                    if (sourcePins.size == 1 && !hasFence && !hasTrack) {
                        map.moveCamera(
                            CameraUpdateFactory.newLatLngZoom(
                                LatLng(sourcePins.first().lat, sourcePins.first().lng),
                                15f,
                            ),
                        )
                    } else {
                        val builder = LatLngBounds.Builder()
                        if (hasFence) {
                            fencePoints.forEach { builder.include(LatLng(it.lat, it.lng)) }
                        } else {
                            sourcePins.forEach { builder.include(LatLng(it.lat, it.lng)) }
                            trackPoints.forEach { builder.include(LatLng(it.lat, it.lng)) }
                        }
                        map.moveCamera(
                            CameraUpdateFactory.newLatLngBounds(builder.build(), 80),
                        )
                    }
                    if (autoFitOnPins) {
                        cameraFittedFor = fitKey
                    } else {
                        lastForcedFitNonce = fitNonce
                    }
                    refreshVisibleRegion(map)
                }
                mapError = null
            } catch (t: Throwable) {
                mapError = t.message ?: Strings.t(Str.TencentMapRenderFailed)
            }
        }

        // 仅在选中车辆变化且允许跟飞时 zoom（首页 animateToSelection=false）
        LaunchedEffect(tencentMap, mapLoaded, selectedCarId, animateToSelection) {
            val map = tencentMap ?: return@LaunchedEffect
            if (!mapLoaded || !animateToSelection) return@LaunchedEffect
            val carId = selectedCarId?.takeIf { it.isNotBlank() } ?: return@LaunchedEffect
            val pin = pins.firstOrNull {
                (it.lat != 0.0 || it.lng != 0.0) &&
                    !it.isCluster &&
                    (it.id == carId || it.memberIds.contains(carId))
            } ?: return@LaunchedEffect
            map.animateCamera(
                CameraUpdateFactory.newLatLngZoom(LatLng(pin.lat, pin.lng), 16f),
            )
        }

        if (showStatusOverlay) {
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
        } else if (mapError != null) {
            Text(
                text = mapError!!,
                modifier = Modifier
                    .align(Alignment.Center)
                    .padding(16.dp),
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.error,
            )
        }
    }
}

private data class MarkerTag(
    val id: String,
    val isCluster: Boolean,
    val memberIds: List<String>,
)

/** Legacy MapOptionProvide.getVehicleIcon / ico_vehicle_* resource ids. */
private fun legacyVehicleDrawable(pin: MapPin): Int = when (pin.icon) {
    MapPinIcon.VehicleWarning -> R.drawable.ico_vehicle_warning
    MapPinIcon.VehicleError -> R.drawable.ico_vehicle_error
    MapPinIcon.VehicleReady -> R.drawable.icon_vehicle_ready
    MapPinIcon.VehicleRiding -> R.drawable.icon_vehicle_riding
    MapPinIcon.VehicleBooking -> R.drawable.icon_vehicle_booking
    MapPinIcon.VehicleTempParking -> R.drawable.icon_vehicle_temp_parking
    MapPinIcon.VehicleLowBattery -> R.drawable.icon_vehicle_low_battery
    MapPinIcon.VehicleMoving -> R.drawable.icon_vehicle_moving
    MapPinIcon.VehicleRepairing -> R.drawable.icon_vehicle_repairing
    MapPinIcon.VehicleHome -> R.drawable.icon_vehicle
    MapPinIcon.Parking -> R.drawable.icon_parking_normal_unselect
    MapPinIcon.ParkingFunction -> R.drawable.icon_parking_function_unselect
    MapPinIcon.ParkingHidden -> R.drawable.icon_parking_hiden_unselect
    MapPinIcon.NoParking -> R.drawable.icon_no_parking_unselect
    MapPinIcon.TrackOrigin -> R.drawable.btn_trajectory_origin
    MapPinIcon.TrackEnd -> R.drawable.btn_trajectory_end
    MapPinIcon.Default -> when {
        pin.restBattery in 0..29 -> R.drawable.ico_vehicle_error
        pin.restBattery in 30..59 -> R.drawable.ico_vehicle_warning
        pin.ridingState == 1 -> R.drawable.icon_vehicle_ready
        else -> R.drawable.ico_vehicle_normal
    }
    else -> R.drawable.ico_vehicle_normal
}
