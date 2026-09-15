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
            modifier = modifier,
            clusterOverview = spec.clusterOverview,
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
