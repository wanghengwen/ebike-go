package com.luopingtech.ebike.rider.ui.map

import androidx.compose.runtime.Composable
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.Modifier
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.core.i18n.Strings
import com.luopingtech.ebike.rider.domain.model.FencePolygon
import com.luopingtech.ebike.rider.domain.model.MapPin
import com.luopingtech.ebike.rider.domain.model.TrackPoint

data class RiderMapSpec(
    val pins: List<MapPin> = emptyList(),
    val selectedCarId: String? = null,
    val providerLabel: String = "",
    val onSelectCarId: (String) -> Unit = {},
    val onSelectCluster: (List<String>) -> Unit = {},
    val clusterOverview: Boolean = true,
    val fencePolygons: List<FencePolygon> = emptyList(),
    val trackPoints: List<TrackPoint> = emptyList(),
    val fitNonce: Int = 0,
    val zoomInNonce: Int = 0,
    val zoomOutNonce: Int = 0,
    val followNonce: Int = 0,
    val followLat: Double? = null,
    val followLng: Double? = null,
    val mapTypeSatellite: Boolean = false,
    val showStatusOverlay: Boolean = true,
)

interface RiderMapRenderer {
    @Composable
    fun Pins(spec: RiderMapSpec, modifier: Modifier)
}

object SimulatorMapRenderer : RiderMapRenderer {
    @Composable
    override fun Pins(spec: RiderMapSpec, modifier: Modifier) {
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

val LocalRiderMapRenderer = staticCompositionLocalOf<RiderMapRenderer> { SimulatorMapRenderer }

@Composable
fun RiderMapView(spec: RiderMapSpec, modifier: Modifier = Modifier) {
    LocalRiderMapRenderer.current.Pins(spec, modifier)
}
