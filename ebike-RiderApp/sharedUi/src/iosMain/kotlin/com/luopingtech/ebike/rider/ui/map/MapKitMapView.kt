package com.luopingtech.ebike.rider.ui.map

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.UIKitView
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.core.i18n.Strings
import com.luopingtech.ebike.rider.domain.model.FencePolygon
import com.luopingtech.ebike.rider.domain.model.MapPin
import com.luopingtech.ebike.rider.domain.model.TrackPoint
import kotlinx.cinterop.BetaInteropApi
import kotlinx.cinterop.CValue
import kotlinx.cinterop.ExperimentalForeignApi
import kotlinx.cinterop.allocArray
import kotlinx.cinterop.cValue
import kotlinx.cinterop.get
import kotlinx.cinterop.memScoped
import kotlinx.cinterop.useContents
import platform.CoreLocation.CLLocationCoordinate2D
import platform.CoreLocation.CLLocationCoordinate2DMake
import platform.MapKit.MKAnnotationProtocol
import platform.MapKit.MKAnnotationView
import platform.MapKit.MKClusterAnnotation
import platform.MapKit.MKCoordinateRegion
import platform.MapKit.MKMapTypeHybrid
import platform.MapKit.MKMapTypeStandard
import platform.MapKit.MKMapView
import platform.MapKit.MKMapViewDelegateProtocol
import platform.MapKit.MKMarkerAnnotationView
import platform.MapKit.MKOverlayProtocol
import platform.MapKit.MKOverlayRenderer
import platform.MapKit.MKPolygon
import platform.MapKit.MKPolygonRenderer
import platform.MapKit.MKPolyline
import platform.MapKit.MKPolylineRenderer
import platform.MapKit.MKUserLocation
import platform.MapKit.addOverlay
import platform.MapKit.overlays
import platform.MapKit.removeOverlays
import platform.UIKit.UIColor
import platform.UIKit.systemBlueColor
import platform.UIKit.systemGreenColor
import platform.UIKit.systemOrangeColor
import platform.UIKit.systemPurpleColor
import platform.UIKit.systemRedColor
import platform.UIKit.systemTealColor
import platform.darwin.NSObject

/**
 * 苹果自带地图（MapKit）版的车点图，是 iOS 上没有腾讯 SDK 时的真实地图实现。
 *
 * 聚合直接用 MapKit 内建的 `clusteringIdentifier`，而不是共享层的 [MapClusterer]：
 * 系统聚合跟缩放联动，手感和腾讯那边的一致；点中聚合点时把成员车号原样交回业务，
 * 跟 Android 的 `onSelectCluster` 是同一条路。
 *
 * 注意坐标系：MapKit 收 WGS-84，腾讯 SDK 收 GCJ-02。后端下发的是 GCJ-02，
 * 所以这里画出来会有几十米的整体偏移；等接上腾讯 SDK 后由它自己处理，
 * MapKit 只作为兜底，不为此再引一层坐标转换。
 */
@OptIn(ExperimentalForeignApi::class)
object MapKitRiderMapRenderer : RiderMapRenderer {

    @Composable
    override fun Pins(spec: RiderMapSpec, modifier: Modifier) {
        MapKitMapView(spec = spec, modifier = modifier)
    }
}

@OptIn(ExperimentalForeignApi::class)
@Composable
private fun MapKitMapView(spec: RiderMapSpec, modifier: Modifier) {
    val delegate = remember { RiderMapDelegate() }
    val sync = remember { MapSyncState() }
    // 回调每次重组都可能换实例，delegate 只持最新的一份。
    delegate.onSelectCar = spec.onSelectCarId
    delegate.onSelectCluster = spec.onSelectCluster
    delegate.clusterOverview = spec.clusterOverview

    Box(modifier = modifier) {
        UIKitView(
            factory = {
                MKMapView().apply {
                    setDelegate(delegate)
                    setShowsCompass(true)
                    setShowsScale(true)
                    setPitchEnabled(false)
                }
            },
            modifier = Modifier.fillMaxSize(),
            update = { map -> sync.apply(map, spec) },
        )
        if (spec.showStatusOverlay && spec.pins.isNotEmpty()) {
            Text(
                // 名字要说实话：spec.providerLabel 是共享层按配置算的（demo 下是「模拟地图」），
                // 这里真正在渲染的是 MapKit，跟 Android 用「腾讯地图 · …」是同一种自报。
                // 无车点时不显示「暂无车辆坐标」——首页空态是常态。
                text = Strings.t(Str.AppleMapPins, spec.pins.size, spec.pins.size),
                style = MaterialTheme.typography.labelMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.align(Alignment.TopStart).padding(10.dp),
            )
        }
    }
}

/**
 * MKMapView 是命令式的，Compose 每次重组都会走 update；这里记住上一次的输入，
 * 只在真的变了才动地图 —— 否则相机会被反复重置，用户拖不动图。
 */
@OptIn(ExperimentalForeignApi::class)
private class MapSyncState {
    private var pinsKey: Int? = null
    private var overlayKey: Int? = null
    private var selectedCarId: String? = null
    private var cameraFittedFor: String? = null
    private var cameraSelectedCarId: String? = null
    private var pendingFitRegion: CValue<MKCoordinateRegion>? = null
    private var fitNonce = -1
    private var zoomInNonce = 0
    private var zoomOutNonce = 0
    private var followNonce = 0
    private var satellite: Boolean? = null

    fun apply(map: MKMapView, spec: RiderMapSpec) {
        applyMapType(map, spec.mapTypeSatellite)
        applyPins(map, spec)
        applyOverlays(map, spec.fencePolygons, spec.trackPoints)
        applyCamera(map, spec)
    }

    private fun applyMapType(map: MKMapView, wantSatellite: Boolean) {
        if (satellite == wantSatellite) return
        satellite = wantSatellite
        // Hybrid 而不是 Satellite：卫星图上没有路名，现场找车反而更难。
        map.setMapType(if (wantSatellite) MKMapTypeHybrid else MKMapTypeStandard)
    }

    private fun applyPins(map: MKMapView, spec: RiderMapSpec) {
        val key = spec.pins.hashCode()
        if (pinsKey != key) {
            pinsKey = key
            map.removeAnnotations(map.annotations)
            map.addAnnotations(spec.pins.map { RiderPinAnnotation(it) })
        }
        if (selectedCarId != spec.selectedCarId) {
            selectedCarId = spec.selectedCarId
            val target = spec.selectedCarId
            val hit = map.annotations
                .filterIsInstance<RiderPinAnnotation>()
                .firstOrNull { it.pin.matches(target) }
            if (hit != null) map.selectAnnotation(hit, animated = true)
        }
    }

    private fun applyOverlays(
        map: MKMapView,
        fences: List<FencePolygon>,
        tracks: List<TrackPoint>,
    ) {
        val key = fences.hashCode() * 31 + tracks.hashCode()
        if (overlayKey == key) return
        overlayKey = key
        map.removeOverlays(map.overlays)
        fences.forEach { fence ->
            polygonOf(fence)?.let { map.addOverlay(it) }
        }
        if (tracks.size >= 2) {
            polylineOf(tracks)?.let { map.addOverlay(it) }
        }
    }

    private fun applyCamera(map: MKMapView, spec: RiderMapSpec) {
        // 车点是异步到的：首帧 pins 还是空，只盯 fitNonce 的话相机会一直停在
        // MKMapView 的默认视野（北美）。跟 Android 一样用「坐标集 + 围栏 + 轨迹 + nonce」
        // 做键，内容变了就重新贴一次视野。
        val valid = spec.pins.filter { it.lat != 0.0 || it.lng != 0.0 }
        val fitKey = valid.joinToString("|") { "${it.id}:${it.lat},${it.lng}" } +
            "|f${spec.fencePolygons.size}|t${spec.trackPoints.size}|n${spec.fitNonce}|" +
            (if (spec.clusterOverview) "home" else "ride") +
            if (spec.followLat != null && spec.followLng != null) {
                "|c${spec.followLat},${spec.followLng}"
            } else {
                ""
            }
        if (valid.isNotEmpty() && fitKey != cameraFittedFor) {
            cameraFittedFor = fitKey
            cameraSelectedCarId = spec.selectedCarId
            if (spec.clusterOverview) {
                // 首页对齐 UniApp scale=17：定焦定位点，不 fit 全部附近车。
                val lat = spec.followLat ?: valid.map { it.lat }.average()
                val lng = spec.followLng ?: valid.map { it.lng }.average()
                applyRegionWhenReady(
                    map,
                    regionOf(lat, lng, HOME_SPAN, HOME_SPAN),
                    animated = false,
                )
            } else if (valid.size == 1 && spec.fencePolygons.isEmpty() && spec.trackPoints.isEmpty()) {
                val only = valid.first()
                applyRegionWhenReady(
                    map,
                    regionOf(only.lat, only.lng, SINGLE_PIN_SPAN, SINGLE_PIN_SPAN),
                    animated = false,
                )
            } else {
                val coords = valid.map { it.lat to it.lng } +
                    spec.trackPoints.map { it.lat to it.lng } +
                    spec.fencePolygons.flatMap { fence -> fence.points.map { it.lat to it.lng } }
                regionFor(coords)?.let { applyRegionWhenReady(map, it, animated = false) }
            }
        } else if (cameraSelectedCarId != spec.selectedCarId) {
            cameraSelectedCarId = spec.selectedCarId
            val selected = spec.selectedCarId?.let { carId ->
                valid.firstOrNull { !it.isCluster && it.matches(carId) }
            }
            if (selected != null) {
                applyRegionWhenReady(
                    map,
                    regionOf(selected.lat, selected.lng, SELECTED_SPAN, SELECTED_SPAN),
                    animated = true,
                )
            }
        }
        flushPendingFitIfPossible(map)
        if (zoomInNonce != spec.zoomInNonce) {
            zoomInNonce = spec.zoomInNonce
            scaleSpan(map, 0.5)
        }
        if (zoomOutNonce != spec.zoomOutNonce) {
            zoomOutNonce = spec.zoomOutNonce
            scaleSpan(map, 2.0)
        }
        if (followNonce != spec.followNonce) {
            followNonce = spec.followNonce
            val lat = spec.followLat
            val lng = spec.followLng
            if (lat != null && lng != null) {
                applyRegionWhenReady(
                    map,
                    regionOf(lat, lng, HOME_SPAN, HOME_SPAN),
                    animated = true,
                )
            }
        }
    }

    private fun applyRegionWhenReady(
        map: MKMapView,
        region: CValue<MKCoordinateRegion>,
        animated: Boolean,
    ) {
        val ready = map.bounds.useContents { size.width > 1.0 && size.height > 1.0 }
        if (ready) {
            pendingFitRegion = null
            map.setRegion(region, animated = animated)
        } else {
            pendingFitRegion = region
        }
    }

    private fun flushPendingFitIfPossible(map: MKMapView) {
        val pending = pendingFitRegion ?: return
        val ready = map.bounds.useContents { size.width > 1.0 && size.height > 1.0 }
        if (!ready) return
        pendingFitRegion = null
        map.setRegion(pending, animated = false)
    }

    private fun scaleSpan(map: MKMapView, factor: Double) {
        val next = map.region.useContents {
            regionOf(
                centerLat = center.latitude,
                centerLng = center.longitude,
                latDelta = (span.latitudeDelta * factor).coerceIn(MIN_SPAN, MAX_SPAN),
                lngDelta = (span.longitudeDelta * factor).coerceIn(MIN_SPAN, MAX_SPAN),
            )
        }
        map.setRegion(next, animated = true)
    }

    private companion object {
        const val MIN_SPAN = 0.0008
        const val MAX_SPAN = 60.0
        /** 对齐 UniApp 首页 scale=17。 */
        const val HOME_SPAN = 0.003
        /** 对齐 Android 单点 zoom 16。 */
        const val SINGLE_PIN_SPAN = 0.006
        const val SELECTED_SPAN = 0.003
    }
}

@OptIn(ExperimentalForeignApi::class)
private class RiderMapDelegate : NSObject(), MKMapViewDelegateProtocol {
    var onSelectCar: (String) -> Unit = {}
    var onSelectCluster: (List<String>) -> Unit = {}
    var clusterOverview: Boolean = true

    override fun mapView(
        mapView: MKMapView,
        viewForAnnotation: MKAnnotationProtocol,
    ): MKAnnotationView? {
        if (viewForAnnotation is MKUserLocation) return null
        val view = mapView.dequeueReusableAnnotationViewWithIdentifier(REUSE_ID)
            as? MKMarkerAnnotationView
            ?: MKMarkerAnnotationView(annotation = viewForAnnotation, reuseIdentifier = REUSE_ID)
        view.annotation = viewForAnnotation
        when (viewForAnnotation) {
            is MKClusterAnnotation -> {
                view.clusteringIdentifier = null
                view.markerTintColor = UIColor.systemPurpleColor
                view.glyphText = viewForAnnotation.memberAnnotations.size.toString()
            }
            is RiderPinAnnotation -> {
                // 只有总览态才聚合；车辆详情、轨迹这些单点场景聚合反而挡住信息。
                view.clusteringIdentifier = if (clusterOverview) REUSE_ID else null
                view.markerTintColor = viewForAnnotation.pin.tintColor()
                view.glyphText = viewForAnnotation.pin.glyph()
            }
        }
        return view
    }

    override fun mapView(mapView: MKMapView, didSelectAnnotationView: MKAnnotationView) {
        when (val annotation = didSelectAnnotationView.annotation) {
            is MKClusterAnnotation -> {
                val ids = annotation.memberAnnotations
                    .filterIsInstance<RiderPinAnnotation>()
                    .flatMap { it.pin.carIds() }
                if (ids.isNotEmpty()) onSelectCluster(ids)
            }
            is RiderPinAnnotation -> {
                val pin = annotation.pin
                if (pin.isCluster) onSelectCluster(pin.carIds()) else onSelectCar(pin.id)
            }
            else -> Unit
        }
    }

    override fun mapView(
        mapView: MKMapView,
        rendererForOverlay: MKOverlayProtocol,
    ): MKOverlayRenderer = when (rendererForOverlay) {
        is MKPolygon -> MKPolygonRenderer(polygon = rendererForOverlay).apply {
            strokeColor = UIColor.systemBlueColor
            lineWidth = 2.0
            fillColor = UIColor.systemBlueColor.colorWithAlphaComponent(0.12)
        }
        is MKPolyline -> MKPolylineRenderer(polyline = rendererForOverlay).apply {
            strokeColor = UIColor.systemOrangeColor
            lineWidth = 4.0
        }
        else -> MKOverlayRenderer(overlay = rendererForOverlay)
    }
}

/** ObjC 子类的 companion 不能有字段，所以复用标识只能放文件级。 */
private const val REUSE_ID = "ops-pin"

@OptIn(ExperimentalForeignApi::class, BetaInteropApi::class)
private class RiderPinAnnotation(val pin: MapPin) : NSObject(), MKAnnotationProtocol {
    private val point = CLLocationCoordinate2DMake(pin.lat, pin.lng)

    override fun coordinate(): CValue<CLLocationCoordinate2D> = point
    override fun title(): String? = pin.title
    override fun subtitle(): String? = pin.subtitle.ifBlank { null }
}

private fun MapPin.carIds(): List<String> = memberIds.ifEmpty { listOf(id) }

private fun MapPin.matches(carId: String?): Boolean =
    carId != null && (id == carId || memberIds.contains(carId))

@OptIn(ExperimentalForeignApi::class)
private fun MapPin.tintColor(): UIColor = when {
    restBattery in 1..30 -> UIColor.systemOrangeColor
    ridingState == 1 -> UIColor.systemBlueColor
    else -> UIColor.systemTealColor
}

private fun MapPin.glyph(): String? = if (isCluster) memberCount.toString() else null

@OptIn(ExperimentalForeignApi::class)
private fun polygonOf(fence: FencePolygon): MKPolygon? {
    if (fence.points.size < 3) return null
    return memScoped {
        val buffer = allocArray<CLLocationCoordinate2D>(fence.points.size)
        fence.points.forEachIndexed { index, point ->
            buffer[index].latitude = point.lat
            buffer[index].longitude = point.lng
        }
        MKPolygon.polygonWithCoordinates(buffer, fence.points.size.toULong())
    }
}

@OptIn(ExperimentalForeignApi::class)
private fun polylineOf(points: List<TrackPoint>): MKPolyline? {
    if (points.size < 2) return null
    return memScoped {
        val buffer = allocArray<CLLocationCoordinate2D>(points.size)
        points.forEachIndexed { index, point ->
            buffer[index].latitude = point.lat
            buffer[index].longitude = point.lng
        }
        MKPolyline.polylineWithCoordinates(buffer, points.size.toULong())
    }
}

@OptIn(ExperimentalForeignApi::class)
private fun regionOf(
    centerLat: Double,
    centerLng: Double,
    latDelta: Double,
    lngDelta: Double,
): CValue<MKCoordinateRegion> = cValue {
    center.latitude = centerLat
    center.longitude = centerLng
    span.latitudeDelta = latDelta
    span.longitudeDelta = lngDelta
}

/** 把一批坐标包成刚好装得下的视野；空集合返回 null，让相机保持原状。 */
@OptIn(ExperimentalForeignApi::class)
private fun regionFor(coords: List<Pair<Double, Double>>): CValue<MKCoordinateRegion>? {
    val valid = coords.filter { (lat, lng) -> lat != 0.0 || lng != 0.0 }
    if (valid.isEmpty()) return null
    val minLat = valid.minOf { it.first }
    val maxLat = valid.maxOf { it.first }
    val minLng = valid.minOf { it.second }
    val maxLng = valid.maxOf { it.second }
    return regionOf(
        centerLat = (minLat + maxLat) / 2,
        centerLng = (minLng + maxLng) / 2,
        // 1.4 倍留白，跟腾讯那边 fitBounds 的边距观感对齐。
        latDelta = ((maxLat - minLat) * 1.4).coerceAtLeast(0.004),
        lngDelta = ((maxLng - minLng) * 1.4).coerceAtLeast(0.004),
    )
}