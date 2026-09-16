package com.luopingtech.ebike.rider.ui.map

import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.viewinterop.UIKitView
import com.luopingtech.ebike.rider.RiderApp
import com.luopingtech.ebike.rider.domain.map.MapClusterer
import com.luopingtech.ebike.rider.domain.map.MapProjection
import com.luopingtech.ebike.rider.domain.model.FencePolygon
import com.luopingtech.ebike.rider.domain.model.MapPin
import com.luopingtech.ebike.rider.platform.MapProviderKind
import kotlinx.cinterop.ExperimentalForeignApi
import platform.UIKit.UIView

/**
 * 宿主原生地图的注入点。
 *
 * 腾讯地图 iOS SDK 是闭源二进制、还要按 bundleId 绑 Key，跟 Android 把
 * `TencentMapView` 留在 `androidApp` 一样，它只能待在 `iosApp`（CocoaPods）里；
 * `sharedUi` 只认这三个协议。宿主没注入时自动落到 [MapKitRiderMapRenderer]。
 *
 * 契约刻意只用 ObjC 友好的扁平类型（String / Double / Int / List），
 * 这样 Swift 端可以直接实现协议，不需要碰 Kotlin 的 data class 与 lambda 默认值。
 */
interface IosHostMapView {
    /** 挂到 Compose 树上的原生视图。 */
    fun view(): UIView

    /** 数据或相机指令变化时调用；实现方按 nonce 自行判重。 */
    fun apply(update: IosHostMapUpdate)

    /** 离开界面时释放 SDK 资源（腾讯的 QMapView 需要显式停。） */
    fun dispose()
}

interface IosHostMapFactory {
    /** provider 是否真的可用（Key 校验通过、SDK 已初始化）。 */
    fun isReady(): Boolean

    fun createPinsView(
        onSelectCarId: (String) -> Unit,
        onSelectCluster: (List<String>) -> Unit,
    ): IosHostMapView

    /** 还车散点图；不支持时返回 null，共享层会降级成车点图。 */
    fun createScatterView(): IosHostMapView?
}

/**
 * 坐标点。刻意不用 `List<Double>`：Kotlin 的 `List<Double>` 导出到 Swift 会变成
 * `[KotlinDouble]`，每个点都要 `.doubleValue` 拆箱，宿主代码会很难看。
 */
class IosHostLatLng(val lat: Double, val lng: Double)

class IosHostMapPin(
    val id: String,
    val lat: Double,
    val lng: Double,
    val title: String,
    val subtitle: String,
    val restBattery: Int,
    /** 1 = 骑行中；-1 = 未知。 */
    val ridingState: Int,
    /** >1 表示聚合点，成员在 [memberIds]。 */
    val memberCount: Int,
    val memberIds: List<String>,
    /** 还车散点图用：异常点。 */
    val abnormal: Boolean = false,
)

class IosHostMapFence(
    val id: String,
    val name: String,
    /** 0 = 服务区，1 = 停车点，2 = 禁停区。 */
    val kind: Int,
    val points: List<IosHostLatLng>,
)

class IosHostMapUpdate(
    val pins: List<IosHostMapPin>,
    val selectedCarId: String?,
    val clusterOverview: Boolean,
    val fences: List<IosHostMapFence>,
    val trackPoints: List<IosHostLatLng>,
    val fitNonce: Int,
    val zoomInNonce: Int,
    val zoomOutNonce: Int,
    val followNonce: Int,
    /** false 时 [followLat] / [followLng] 无意义（可空 Double 在 Swift 也是装箱的）。 */
    val hasFollowTarget: Boolean,
    val followLat: Double,
    val followLng: Double,
    val satellite: Boolean,
)

/** Swift 在启动时写这里，Kotlin 侧只读。 */
object IosMapHost {
    var factory: IosHostMapFactory? = null
}

/**
 * 用宿主原生地图（腾讯）渲染。视图生命周期跟着 Compose 走，
 * 数据同步全部下推给宿主实现——它才知道自家 SDK 哪些调用是幂等的。
 */
@OptIn(ExperimentalForeignApi::class)
class HostedRiderMapRenderer(
    private val factory: IosHostMapFactory,
) : RiderMapRenderer {

    @Composable
    override fun Pins(spec: RiderMapSpec, modifier: Modifier) {
        val hosted = remember(factory) {
            factory.createPinsView(
                onSelectCarId = { spec.onSelectCarId(it) },
                onSelectCluster = { spec.onSelectCluster(it) },
            )
        }
        DisposableEffect(hosted) { onDispose { hosted.dispose() } }
        UIKitView(
            factory = { hosted.view() },
            modifier = modifier,
            update = { hosted.apply(spec.toHostUpdate()) },
        )
    }

}

/** 腾讯就绪就用腾讯，否则苹果自带 MapKit 降级。 */
fun riderMapRendererForIos(app: RiderApp): RiderMapRenderer {
    val hosted = IosMapHost.factory
    val wantsTencent = app.mapCapability.kind == MapProviderKind.TENCENT
    return if (hosted != null && hosted.isReady() && wantsTencent) {
        HostedRiderMapRenderer(hosted)
    } else {
        MapKitRiderMapRenderer
    }
}

private fun RiderMapSpec.toHostUpdate(): IosHostMapUpdate {
    // 聚合放在 Kotlin 侧：腾讯 iOS SDK 没有内建聚合，而这套阈值已经在 Android
    // 和画布实现上验证过，没必要让宿主再实现一遍（也就不会两端手感不一致）。
    val markers = if (!clusterOverview || pins.size <= 4) {
        pins
    } else {
        val bounds = MapProjection.boundsOf(pins)
        MapClusterer.cluster(pins, MapClusterer.suggestedCellDegrees(bounds, targetCells = 5))
    }
    return IosHostMapUpdate(
        pins = markers.map { it.toHostPin() },
        selectedCarId = selectedCarId,
        clusterOverview = clusterOverview,
        fences = fencePolygons.map { it.toHostFence() },
        trackPoints = trackPoints.map { IosHostLatLng(it.lat, it.lng) },
        fitNonce = fitNonce,
        zoomInNonce = zoomInNonce,
        zoomOutNonce = zoomOutNonce,
        followNonce = followNonce,
        hasFollowTarget = followLat != null && followLng != null,
        followLat = followLat ?: 0.0,
        followLng = followLng ?: 0.0,
        satellite = mapTypeSatellite,
    )
}

private fun MapPin.toHostPin() = IosHostMapPin(
    id = id,
    lat = lat,
    lng = lng,
    title = title,
    subtitle = subtitle,
    restBattery = restBattery,
    ridingState = ridingState ?: -1,
    memberCount = memberCount,
    memberIds = memberIds,
)

private fun FencePolygon.toHostFence() = IosHostMapFence(
    id = id,
    name = name,
    kind = kind.ordinal,
    points = points.map { IosHostLatLng(it.lat, it.lng) },
)
