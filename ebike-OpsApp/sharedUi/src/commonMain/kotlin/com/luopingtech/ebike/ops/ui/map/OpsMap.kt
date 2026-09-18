package com.luopingtech.ebike.ops.ui.map

import androidx.compose.runtime.Composable
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.Modifier
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.domain.analysis.ReturnCarPoint
import com.luopingtech.ebike.ops.domain.model.FencePolygon
import com.luopingtech.ebike.ops.domain.model.MapPin
import com.luopingtech.ebike.ops.domain.model.TrackPoint

/**
 * 车点地图的渲染参数。全部是共享层类型，所以这份契约在 commonMain 就能表达，
 * 厂商 SDK 留在宿主，由宿主提供 [OpsMapRenderer]。
 */
data class OpsMapSpec(
    val pins: List<MapPin> = emptyList(),
    val selectedCarId: String? = null,
    /** 无厂商地图时画布左上角显示的来源标签。 */
    val providerLabel: String = "",
    val onSelectCarId: (String) -> Unit = {},
    val onSelectCluster: (List<String>) -> Unit = {},
    /** 点空白地图（一般业务用；围栏选点不走这个）。 */
    val onMapTap: ((lat: Double, lng: Double) -> Unit)? = null,
    /** 相机停止后回传中心经纬度。 */
    val onCameraIdle: ((lat: Double, lng: Double) -> Unit)? = null,
    /**
     * 相机移动中回传中心（对齐原版 obsMapIdle=false → onGestureChange(ON_MOVE)）。
     * 选点编辑用它持续 latLng→screen，避免区域钉在屏幕上跟着拖。
     */
    val onCameraMove: ((lat: Double, lng: Double) -> Unit)? = null,
    /**
     * 递增后把 [screenPickX]/[screenPickY]（相对地图视图像素）转成经纬度，
     * 通过 [onScreenToLatLng] 回传（对齐原版准星 centerPoint → fromScreenLocation）。
     */
    val screenToLatLngNonce: Int = 0,
    val screenPickX: Float = 0f,
    val screenPickY: Float = 0f,
    val onScreenToLatLng: ((lat: Double, lng: Double) -> Unit)? = null,
    /**
     * 批量屏幕坐标 → 经纬度（贴片四角）。递增 [batchScreenToLatLngNonce] 触发。
     */
    val batchScreenToLatLngNonce: Int = 0,
    val batchScreenPoints: List<Pair<Float, Float>> = emptyList(),
    val onBatchScreenToLatLng: ((List<Pair<Double, Double>>) -> Unit)? = null,
    /**
     * 批量经纬度 → 屏幕坐标（选点顶点重绘，对齐 screenLocation）。
     * 递增 [latLngToScreenNonce] 触发；结果为相对地图视图像素。
     */
    val latLngToScreenNonce: Int = 0,
    val latLngToScreenPoints: List<Pair<Double, Double>> = emptyList(),
    val onLatLngToScreen: ((List<Pair<Float, Float>>) -> Unit)? = null,
    /** false 时不因 pins 变化自动 fit（围栏绘制页选点会跳动）。 */
    val autoFitOnPins: Boolean = true,
    /**
     * 选中车辆后是否飞到该车（首页地图对齐原版不跟飞，详情页等可开）。
     */
    val animateToSelection: Boolean = true,
    val clusterOverview: Boolean = true,
    val fencePolygons: List<FencePolygon> = emptyList(),
    val trackPoints: List<TrackPoint> = emptyList(),
    /** 递增后强制重新 fit 视野（看全部）。 */
    val fitNonce: Int = 0,
    /** 递增后放大一级。 */
    val zoomInNonce: Int = 0,
    /** 递增后缩小一级。 */
    val zoomOutNonce: Int = 0,
    /** 递增后相机跟到 [followLat]/[followLng]（定位到我）。 */
    val followNonce: Int = 0,
    val followLat: Double? = null,
    val followLng: Double? = null,
    /** 跟飞缩放，对齐原版 FENCE_DEFAULT_ZOOM=17。 */
    val followZoom: Float = 17f,
    val mapTypeSatellite: Boolean = false,
    val showStatusOverlay: Boolean = true,
)

/**
 * 还车分布散点图的渲染参数。[fallbackPins] 供没有厂商散点能力时降级成普通车点。
 */
data class OpsScatterMapSpec(
    val points: List<ReturnCarPoint> = emptyList(),
    val showNormal: Boolean = true,
    val showAbnormal: Boolean = true,
    val fencePolygons: List<FencePolygon> = emptyList(),
    val fitNonce: Int = 0,
    val fallbackPins: List<MapPin> = emptyList(),
    val providerLabel: String = "",
)

/**
 * 地图渲染器契约。宿主注入真实实现（腾讯地图 / Google Maps / MapKit），
 * 缺省用 [SimulatorMapRenderer]，纯 Compose 画布，双端都能跑。
 */
interface OpsMapRenderer {
    @Composable
    fun Pins(spec: OpsMapSpec, modifier: Modifier)

    @Composable
    fun Scatter(spec: OpsScatterMapSpec, modifier: Modifier) {
        Pins(
            spec = OpsMapSpec(
                pins = spec.fallbackPins,
                providerLabel = spec.providerLabel,
                clusterOverview = false,
                fencePolygons = spec.fencePolygons,
                fitNonce = spec.fitNonce,
            ),
            modifier = modifier,
        )
    }
}

/** 无厂商 SDK 时的缺省渲染器：把经纬度投影到 Canvas。 */
object SimulatorMapRenderer : OpsMapRenderer {
    @Composable
    override fun Pins(spec: OpsMapSpec, modifier: Modifier) {
        SimulatorMapView(
            pins = spec.pins,
            selectedCarId = spec.selectedCarId,
            providerLabel = spec.providerLabel.ifBlank { Strings.t(Str.SimulatorMap) },
            onSelectCarId = spec.onSelectCarId,
            onSelectCluster = spec.onSelectCluster,
            onMapTap = spec.onMapTap,
            trackPoints = spec.trackPoints,
            modifier = modifier,
            clusterOverview = spec.clusterOverview,
            onCameraIdle = spec.onCameraIdle,
            onCameraMove = spec.onCameraMove,
            followNonce = spec.followNonce,
            followLat = spec.followLat,
            followLng = spec.followLng,
            followZoom = spec.followZoom,
            screenToLatLngNonce = spec.screenToLatLngNonce,
            screenPickX = spec.screenPickX,
            screenPickY = spec.screenPickY,
            onScreenToLatLng = spec.onScreenToLatLng,
            batchScreenToLatLngNonce = spec.batchScreenToLatLngNonce,
            batchScreenPoints = spec.batchScreenPoints,
            onBatchScreenToLatLng = spec.onBatchScreenToLatLng,
            latLngToScreenNonce = spec.latLngToScreenNonce,
            latLngToScreenPoints = spec.latLngToScreenPoints,
            onLatLngToScreen = spec.onLatLngToScreen,
            autoFitOnPins = spec.autoFitOnPins,
        )
    }
}

val LocalOpsMapRenderer = staticCompositionLocalOf<OpsMapRenderer> { SimulatorMapRenderer }

/** 业务界面统一用这个画车点地图，不关心背后是哪家 SDK。 */
@Composable
fun OpsMapView(spec: OpsMapSpec, modifier: Modifier = Modifier) {
    LocalOpsMapRenderer.current.Pins(spec, modifier)
}

/** 还车分布散点图。厂商不支持时自动降级为车点图。 */
@Composable
fun OpsScatterMapView(spec: OpsScatterMapSpec, modifier: Modifier = Modifier) {
    LocalOpsMapRenderer.current.Scatter(spec, modifier)
}
