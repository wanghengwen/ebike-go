package com.luopingtech.ebike.ops.ui.analysis

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.feature.analysis.VcdPage
import com.luopingtech.ebike.ops.ui.map.OpsMapSpec
import com.luopingtech.ebike.ops.ui.map.OpsMapView
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import com.luopingtech.ebike.ops.ui.theme.OpsToolIcon
import kotlinx.coroutines.launch

/**
 * 车况分布地图 —— 对齐遗留原生页：顶栏 + 档位下拉 + 左侧 刷新/聚合/围栏/定位。
 */
@Composable
fun VehicleConditionDistributionMapScreen(
    app: OpsApp,
    onClose: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val home by app.homeFeature.state.collectAsState()
    val state by app.vehicleConditionDistributionFeature.state.collectAsState()
    val detailMap by app.vehicleDetailMapFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    var clusterOverview by remember { mutableStateOf(true) }
    var showFence by remember { mutableStateOf(false) }
    var clusterIds by remember { mutableStateOf<List<String>?>(null) }
    var bucketMenuOpen by remember { mutableStateOf(false) }
    var fitNonce by remember { mutableIntStateOf(0) }

    LaunchedEffect(home.currentArea?.id) {
        val sid = home.currentArea?.id.orEmpty()
        if (sid.isNotBlank()) {
            app.vehicleDetailMapFeature.loadFence(sid)
        }
    }

    LaunchedEffect(state.page) {
        if (state.page != VcdPage.Map) {
            onClose()
        }
    }

    val pins = state.mapPins
    val selected = state.selectedCarId?.let { app.vehicleConditionDistributionFeature.findMapVehicle(it) }
    val fencePolygons = if (showFence) detailMap.fence?.all.orEmpty() else emptyList()
    val currentBucket = state.mapBuckets.getOrNull(state.mapBucketIndex)
    val currentLabel = currentBucket?.let { t(it.labelKey) }.orEmpty()

    fun selectCar(id: String) {
        app.vehicleConditionDistributionFeature.selectVehicle(id)
        clusterIds = null
        app.vehicleFeature.selectVehicle(id)
    }

    fun closeAll() {
        app.vehicleConditionDistributionFeature.closeMap()
        onClose()
    }

    Column(modifier = Modifier.fillMaxSize().background(Color.White)) {
        VcdTopBar(
            title = t(Str.VehicleDistribution),
            onBack = ::closeAll,
        )

        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(44.dp)
                .background(Color.White)
                .clickable(
                    interactionSource = remember { MutableInteractionSource() },
                    indication = null,
                    onClick = { bucketMenuOpen = !bucketMenuOpen },
                ),
            contentAlignment = Alignment.Center,
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(
                    text = currentLabel,
                    color = Color(0xFF333333),
                    fontSize = 15.sp,
                    fontWeight = FontWeight.Medium,
                )
                Text(
                    text = " ▾",
                    color = Color(0xFF999999),
                    fontSize = 12.sp,
                )
            }
        }
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(1.dp)
                .background(Color(0x14000000)),
        )

        Box(modifier = Modifier.fillMaxSize()) {
            OpsMapView(
                spec = OpsMapSpec(
                    pins = pins,
                    selectedCarId = state.selectedCarId,
                    providerLabel = t(Str.VehicleDistribution),
                    onSelectCarId = { selectCar(it) },
                    onSelectCluster = { ids ->
                        if (ids.size <= 1) {
                            ids.firstOrNull()?.let { selectCar(it) }
                        } else {
                            clusterOverview = false
                            clusterIds = ids
                            app.vehicleConditionDistributionFeature.selectVehicle(null)
                        }
                    },
                    clusterOverview = clusterOverview,
                    fencePolygons = fencePolygons,
                    fitNonce = fitNonce,
                    showStatusOverlay = false,
                ),
                modifier = Modifier.fillMaxSize(),
            )

            VcdMapToolsRail(
                refreshLabel = t(Str.Refresh),
                clusterLabel = t(Str.MapToolCluster),
                fenceLabel = t(Str.MapToolFence),
                locateLabel = t(Str.MapToolLocate),
                clusterSelected = clusterOverview,
                fenceSelected = showFence,
                onRefresh = {
                    scope.launch {
                        app.vehicleConditionDistributionFeature.load(home.currentArea)
                        home.currentArea?.id?.takeIf { it.isNotBlank() }?.let {
                            app.vehicleDetailMapFeature.loadFence(it)
                        }
                    }
                },
                onCluster = { clusterOverview = !clusterOverview },
                onFence = { showFence = !showFence },
                onLocate = { fitNonce += 1 },
                modifier = Modifier
                    .align(Alignment.CenterStart)
                    .padding(start = 12.dp),
            )

            if (bucketMenuOpen) {
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .heightIn(max = 320.dp)
                        .background(Color.White)
                        .verticalScroll(rememberScrollState())
                        .align(Alignment.TopCenter),
                ) {
                    state.mapBuckets.forEach { bucket ->
                        val selectedBucket = bucket.index == state.mapBucketIndex
                        Text(
                            text = "${t(bucket.labelKey)}  (${bucket.count})",
                            color = if (selectedBucket) OpsTheme.colors.primary else Color(0xFF333333),
                            fontSize = 14.sp,
                            fontWeight = if (selectedBucket) FontWeight.SemiBold else FontWeight.Normal,
                            modifier = Modifier
                                .fillMaxWidth()
                                .clickable {
                                    app.vehicleConditionDistributionFeature.selectMapBucket(bucket.index)
                                    bucketMenuOpen = false
                                }
                                .padding(horizontal = 20.dp, vertical = 14.dp),
                        )
                    }
                }
            }

            clusterIds?.takeIf { it.size > 1 }?.let { ids ->
                Surface(
                    tonalElevation = 2.dp,
                    modifier = Modifier
                        .align(Alignment.BottomCenter)
                        .fillMaxWidth()
                        .padding(12.dp),
                ) {
                    Column(
                        modifier = Modifier.padding(12.dp),
                        verticalArrangement = Arrangement.spacedBy(6.dp),
                    ) {
                        Text(t(Str.ClusterVehicles, ids.size), fontWeight = FontWeight.SemiBold)
                        ids.take(8).forEach { id ->
                            TextButton(onClick = { selectCar(id) }) { Text(id) }
                        }
                        TextButton(onClick = { clusterIds = null }) { Text(t(Str.Close)) }
                    }
                }
            }

            selected?.let { vehicle ->
                Surface(
                    tonalElevation = 3.dp,
                    shadowElevation = 4.dp,
                    modifier = Modifier
                        .align(Alignment.BottomCenter)
                        .fillMaxWidth()
                        .padding(12.dp),
                    shape = RoundedCornerShape(8.dp),
                ) {
                    Column(
                        modifier = Modifier.padding(14.dp),
                        verticalArrangement = Arrangement.spacedBy(4.dp),
                    ) {
                        Text(
                            "${vehicle.carId} · ${vehicle.batteryLabel} · ${vehicle.ridingLabel}",
                            fontWeight = FontWeight.SemiBold,
                            fontSize = 15.sp,
                        )
                        Text(
                            vehicle.siteLabel,
                            color = Color(0xFF888888),
                            fontSize = 12.sp,
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun VcdMapToolsRail(
    refreshLabel: String,
    clusterLabel: String,
    fenceLabel: String,
    locateLabel: String,
    clusterSelected: Boolean,
    fenceSelected: Boolean,
    onRefresh: () -> Unit,
    onCluster: () -> Unit,
    onFence: () -> Unit,
    onLocate: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Column(
        modifier = modifier
            .width(40.dp)
            .shadow(4.dp, RoundedCornerShape(20.dp))
            .background(Color.White, RoundedCornerShape(20.dp))
            .padding(vertical = 16.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        VcdMapToolItem(symbol = "↻", label = refreshLabel, selected = false, onClick = onRefresh)
        VcdMapToolItem(symbol = "▦", label = clusterLabel, selected = clusterSelected, onClick = onCluster)
        VcdMapToolItem(symbol = "▣", label = fenceLabel, selected = fenceSelected, onClick = onFence)
        VcdMapToolItem(symbol = "⌖", label = locateLabel, selected = false, onClick = onLocate)
    }
}

@Composable
private fun VcdMapToolItem(
    symbol: String,
    label: String,
    selected: Boolean,
    onClick: () -> Unit,
) {
    val color = if (selected) OpsTheme.colors.primary else OpsToolIcon
    Column(
        modifier = Modifier
            .width(40.dp)
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            ),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text(text = symbol, color = color, fontSize = 16.sp)
        Text(text = label, color = color, fontSize = 11.sp)
    }
}
