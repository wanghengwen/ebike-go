package com.luopingtech.ebike.ops.ui.fence

import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.Checkbox
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableFloatStateOf
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.layout.boundsInRoot
import androidx.compose.ui.layout.onGloballyPositioned
import androidx.compose.ui.layout.positionInRoot
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.analysis.StationOptStateFilter
import com.luopingtech.ebike.ops.domain.model.FenceKind
import com.luopingtech.ebike.ops.domain.model.FencePolygon
import com.luopingtech.ebike.ops.domain.model.MapPin
import com.luopingtech.ebike.ops.domain.model.TrackPoint
import com.luopingtech.ebike.ops.domain.model.fenceMapPinIcon
import com.luopingtech.ebike.ops.feature.fence.FenceBrowseFeature
import com.luopingtech.ebike.ops.feature.fence.FenceBrowseMode
import com.luopingtech.ebike.ops.feature.fence.FenceBrowseUiState
import com.luopingtech.ebike.ops.feature.fence.FenceDrawMode
import com.luopingtech.ebike.ops.feature.fence.FenceEditFeature
import com.luopingtech.ebike.ops.feature.fence.FenceEditUiState
import com.luopingtech.ebike.ops.feature.fence.FenceListTab
import com.luopingtech.ebike.ops.domain.model.GeoLatLng
import com.luopingtech.ebike.ops.domain.model.metersBetween
import com.luopingtech.ebike.ops.ui.feedback.LocalOpsToast
import com.luopingtech.ebike.ops.ui.icons.OpsIcon
import com.luopingtech.ebike.ops.ui.icons.painterResource
import com.luopingtech.ebike.ops.ui.map.OpsMapSpec
import com.luopingtech.ebike.ops.ui.map.OpsMapView
import com.luopingtech.ebike.ops.ui.navigation.OpsBackHandler
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import kotlin.math.pow
import kotlin.math.roundToInt
import kotlinx.coroutines.launch

/**
 * 对齐遗留 FenceActivity：地�?列表浏览 + 选点新建 + 详情/列表启停删�?
 */
@Composable
fun FenceBrowseScreen(
    app: OpsApp,
    onClose: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val home by app.homeFeature.state.collectAsState()
    val state by app.fenceBrowseFeature.state.collectAsState()
    val edit by app.fenceEditFeature.state.collectAsState()
    val feature = app.fenceBrowseFeature
    val editFeature = app.fenceEditFeature
    val colors = OpsTheme.colors
    val toast = LocalOpsToast.current
    val scope = rememberCoroutineScope()

    var filterPanel by remember { mutableStateOf<FenceFilterPanel?>(null) }
    var fitNonce by remember { mutableIntStateOf(0) }
    var followNonce by remember { mutableIntStateOf(0) }
    var followLat by remember { mutableStateOf<Double?>(null) }
    var followLng by remember { mutableStateOf<Double?>(null) }
    var mapTypeSatellite by remember { mutableStateOf(false) }
    var pickTypeOpen by remember { mutableStateOf(false) }
    var actionError by remember { mutableStateOf<String?>(null) }

    fun closePage() {
        editFeature.clear()
        feature.clear()
        onClose()
    }

    fun reload(fit: Boolean = true, centerLat: Double? = null, centerLng: Double? = null) {
        scope.launch {
            feature.load(home.currentArea)
            if (centerLat != null && centerLng != null) {
                followLat = centerLat
                followLng = centerLng
                followNonce += 1
            } else if (fit) {
                fitNonce += 1
            }
        }
    }

    fun runAction(block: suspend () -> OpsResult<Unit>, okMessage: String? = null) {
        scope.launch {
            actionError = null
            when (val r = block()) {
                is OpsResult.Ok -> {
                    okMessage?.let(toast)
                    feature.setShowAllFences(true)
                    reload(fit = false)
                }
                is OpsResult.Err -> {
                    actionError = r.error.message
                    toast(r.error.message)
                }
            }
        }
    }

    OpsBackHandler(onBack = {
        when {
            edit.active -> {
                feature.setShowAllFences(true)
                editFeature.clear()
                reload(
                    fit = false,
                    centerLat = editFeature.lastSavedCenter?.lat,
                    centerLng = editFeature.lastSavedCenter?.lng,
                )
            }
            state.selectedFenceId != null -> feature.selectFence(null)
            pickTypeOpen -> pickTypeOpen = false
            filterPanel != null -> filterPanel = null
            state.editMode -> feature.setEditMode(false)
            else -> closePage()
        }
    })

    LaunchedEffect(home.currentArea?.id) {
        feature.load(home.currentArea)
        fitNonce += 1
    }

    LaunchedEffect(edit.message) {
        val msg = edit.message ?: return@LaunchedEffect
        toast(msg)
    }

    val focus = state.focusCenter
    val mapPolygons = remember(state.bundle, state.showAllFences) { state.mapPolygons() }
    val mapPins = remember(mapPolygons) {
        mapPolygons
            .filter { it.kind == FenceKind.Parking || it.kind == FenceKind.NoParking }
            .mapNotNull { fence ->
                val c = fence.centerOrCentroid() ?: return@mapNotNull null
                MapPin(
                    id = fence.id,
                    lat = c.lat,
                    lng = c.lng,
                    title = fence.name,
                    icon = fence.fenceMapPinIcon(),
                )
            }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.White),
    ) {
        if (edit.active) {
            val editCenter = followLat?.let { lat -> followLng?.let { lng -> GeoLatLng(lat, lng) } }
                ?: edit.draftPoints.takeIf { it.isNotEmpty() }?.let { pts ->
                    GeoLatLng(pts.map { it.lat }.average(), pts.map { it.lng }.average())
                }
            // 对齐原版：zoom�?4 时按中心 1000m 裁剪附近围栏（编辑默�?zoom 17�?
            val nearbyPolys = remember(state.bundle, edit.editingId, editCenter) {
                val parkings = state.bundle.parkings.filter { it.id != edit.editingId }
                val noParks = state.bundle.noParkings.filter { it.id != edit.editingId }
                val center = editCenter ?: return@remember emptyList()
                (parkings + noParks).filter { fence ->
                    val c = fence.centerOrCentroid() ?: return@filter false
                    metersBetween(center, c) <= 1000.0
                }
            }
            FenceDrawContent(
                edit = edit,
                editFeature = editFeature,
                backgroundPolygons = state.bundle.serviceAreas + nearbyPolys,
                primary = colors.primary,
                translate = { key, args -> t(key, *args) },
                mapTypeSatellite = mapTypeSatellite,
                onToggleSatellite = { mapTypeSatellite = !mapTypeSatellite },
                followNonce = followNonce,
                followLat = followLat ?: focus?.lat,
                followLng = followLng ?: focus?.lng,
                onLocate = {
                    scope.launch {
                        when (val loc = app.locationTracker.currentLocation()) {
                            is OpsResult.Ok -> {
                                followLat = loc.value.latitude
                                followLng = loc.value.longitude
                                followNonce += 1
                            }
                            is OpsResult.Err -> toast(loc.error.message)
                        }
                    }
                },
                onBack = {
                    feature.setShowAllFences(true)
                    editFeature.clear()
                    reload(
                        fit = false,
                        centerLat = editFeature.lastSavedCenter?.lat,
                        centerLng = editFeature.lastSavedCenter?.lng,
                    )
                },
                onSave = {
                    val wasEdit = edit.isEdit
                    scope.launch {
                        when (val r = editFeature.save(home.currentArea)) {
                            is OpsResult.Ok -> {
                                feature.setShowAllFences(true)
                                val c = editFeature.lastSavedCenter
                                feature.load(home.currentArea)
                                if (c != null) {
                                    followLat = c.lat
                                    followLng = c.lng
                                    followNonce += 1
                                }
                                if (wasEdit) {
                                    // cleared by save
                                }
                            }
                            is OpsResult.Err -> toast(r.error.message)
                        }
                    }
                },
            )
        } else {
            FenceBrowseTopBar(
                title = if (state.mode == FenceBrowseMode.Map) {
                    t(Str.FenceMapTitle)
                } else {
                    t(Str.FenceListTitle)
                },
                primary = colors.primary,
                listMode = state.mode == FenceBrowseMode.List,
                editMode = state.editMode,
                editLabel = t(Str.FenceEdit),
                cancelEditLabel = t(Str.FenceCancelEdit),
                onBack = ::closePage,
                onToggleMode = { feature.toggleMode() },
                onToggleEdit = { feature.setEditMode(!state.editMode) },
            )

            actionError?.let { err ->
                Text(
                    text = err,
                    color = MaterialTheme.colorScheme.error,
                    modifier = Modifier.padding(horizontal = 16.dp, vertical = 4.dp),
                    fontSize = 13.sp,
                )
            }

            if (state.loading && state.bundle.all.isEmpty()) {
                Column(
                    modifier = Modifier.fillMaxSize(),
                    verticalArrangement = Arrangement.Center,
                    horizontalAlignment = Alignment.CenterHorizontally,
                ) {
                    CircularProgressIndicator(color = colors.primary)
                }
            } else if (state.errorMessage != null && state.bundle.all.isEmpty()) {
                Text(
                    text = state.errorMessage.orEmpty(),
                    color = MaterialTheme.colorScheme.error,
                    modifier = Modifier.padding(24.dp),
                )
            } else if (state.mode == FenceBrowseMode.Map) {
                FenceMapBrowseContent(
                    state = state,
                    feature = feature,
                    mapPins = mapPins,
                    mapPolygons = mapPolygons,
                    focusLat = followLat ?: focus?.lat,
                    focusLng = followLng ?: focus?.lng,
                    fitNonce = fitNonce,
                    followNonce = followNonce,
                    mapTypeSatellite = mapTypeSatellite,
                    translate = { key, args -> t(key, *args) },
                    onFollow = {
                        followLat = focus?.lat
                        followLng = focus?.lng
                        followNonce += 1
                    },
                    onToggleSatellite = { mapTypeSatellite = !mapTypeSatellite },
                    onNewFence = { pickTypeOpen = true },
                    onEnable = { id -> runAction({ feature.enableFence(id) }) },
                    onDisable = { id -> runAction({ feature.disableFence(id) }) },
                    onDelete = { id, kind -> runAction({ feature.deleteFence(id, kind) }) },
                    onEdit = { fence ->
                        feature.selectFence(null)
                        editFeature.startEdit(fence)
                        fence.centerOrCentroid()?.let { c ->
                            followLat = c.lat
                            followLng = c.lng
                            followNonce += 1
                        }
                    },
                )
            } else {
                FenceListBrowseContent(
                    state = state,
                    feature = feature,
                    primary = colors.primary,
                    translate = { key, args -> t(key, *args) },
                    onOpenTags = { filterPanel = FenceFilterPanel.Tags },
                    onOpenStatus = { filterPanel = FenceFilterPanel.Status },
                    onEnableSelected = {
                        if (state.selectedIds.isEmpty()) {
                            toast(t(Str.FenceBatchNeedSelect))
                        } else {
                            runAction({ feature.enableSelected() })
                        }
                    },
                    onDisableSelected = {
                        if (state.selectedIds.isEmpty()) {
                            toast(t(Str.FenceBatchNeedSelect))
                        } else {
                            runAction({ feature.disableSelected() })
                        }
                    },
                    onDeleteSelected = {
                        if (state.selectedIds.isEmpty()) {
                            toast(t(Str.FenceBatchNeedSelect))
                        } else {
                            runAction({ feature.deleteSelected() })
                        }
                    },
                    onEnable = { id -> runAction({ feature.enableFence(id) }) },
                    onDisable = { id -> runAction({ feature.disableFence(id) }) },
                    onDelete = { id, kind -> runAction({ feature.deleteFence(id, kind) }) },
                    onEdit = { fence ->
                        feature.selectFence(null)
                        editFeature.startEdit(fence)
                        fence.centerOrCentroid()?.let { c ->
                            followLat = c.lat
                            followLng = c.lng
                            followNonce += 1
                        }
                    },
                )
            }
        }
    }

    if (pickTypeOpen) {
        FencePickTypeDialog(
            title = t(Str.FencePickTypeTitle),
            parkingLabel = t(Str.FenceCreateParking),
            noParkingLabel = t(Str.FenceCreateNoParking),
            cancelLabel = t(Str.Cancel),
            onParking = {
                pickTypeOpen = false
                editFeature.start(FenceKind.Parking)
                scope.launch {
                    val sa = state.bundle.serviceAreas.firstOrNull()?.centerOrCentroid()
                    when (val loc = app.locationTracker.currentLocation()) {
                        is OpsResult.Ok -> {
                            // 人对在服务区则跟定位，否则服务区中心（简化对齐）
                            if (sa != null && metersBetween(sa, GeoLatLng(loc.value.latitude, loc.value.longitude)) < 5000) {
                                followLat = loc.value.latitude
                                followLng = loc.value.longitude
                            } else if (sa != null) {
                                followLat = sa.lat
                                followLng = sa.lng
                            } else {
                                followLat = loc.value.latitude
                                followLng = loc.value.longitude
                            }
                            followNonce += 1
                        }
                        is OpsResult.Err -> {
                            if (sa != null) {
                                followLat = sa.lat
                                followLng = sa.lng
                                followNonce += 1
                            }
                        }
                    }
                }
            },
            onNoParking = {
                pickTypeOpen = false
                editFeature.start(FenceKind.NoParking)
                scope.launch {
                    val sa = state.bundle.serviceAreas.firstOrNull()?.centerOrCentroid()
                    when (val loc = app.locationTracker.currentLocation()) {
                        is OpsResult.Ok -> {
                            followLat = loc.value.latitude
                            followLng = loc.value.longitude
                            followNonce += 1
                        }
                        is OpsResult.Err -> {
                            if (sa != null) {
                                followLat = sa.lat
                                followLng = sa.lng
                                followNonce += 1
                            }
                        }
                    }
                }
            },
            onDismiss = { pickTypeOpen = false },
        )
    }

    if (filterPanel == FenceFilterPanel.Tags) {
        FenceTagFilterDialog(
            tags = state.availableTags,
            selected = state.selectedTagIds,
            title = t(Str.StationFilterAllTags),
            clearLabel = t(Str.StationFilterAllTags),
            confirmLabel = t(Str.Confirm),
            onToggle = { feature.toggleTag(it) },
            onClear = { feature.clearTags() },
            onDismiss = { filterPanel = null },
        )
    } else if (filterPanel == FenceFilterPanel.Status) {
        FenceStatusFilterDialog(
            selected = state.optState,
            allLabel = t(Str.StationFilterAllStatus),
            operatingLabel = t(Str.StationStatusOperating),
            stoppedLabel = t(Str.StationStatusStopped),
            onSelect = {
                feature.setOptState(it)
                filterPanel = null
            },
            onDismiss = { filterPanel = null },
        )
    }

    // Collect language so string lookups recompose on locale change.
    language
}

private enum class FenceFilterPanel { Tags, Status }

@Composable
private fun FenceDrawContent(
    edit: FenceEditUiState,
    editFeature: FenceEditFeature,
    backgroundPolygons: List<FencePolygon>,
    primary: Color,
    translate: (Str, Array<out Any?>) -> String,
    mapTypeSatellite: Boolean,
    onToggleSatellite: () -> Unit,
    followNonce: Int,
    followLat: Double?,
    followLng: Double?,
    onBack: () -> Unit,
    onSave: () -> Unit,
    onLocate: (() -> Unit)? = null,
) {
    var showParams by remember { mutableStateOf(false) }
    var pendingSaveAfterParams by remember { mutableStateOf(false) }
    var showNoParkingName by remember { mutableStateOf(false) }
    var showAngle by remember { mutableStateOf(false) }
    var showSize by remember { mutableStateOf(false) }
    var screenToLatLngNonce by remember { mutableIntStateOf(0) }
    var screenPickX by remember { mutableFloatStateOf(0f) }
    var screenPickY by remember { mutableFloatStateOf(0f) }
    var mapOriginInRoot by remember { mutableStateOf(Offset.Zero) }
    var crosshairCenterInRoot by remember { mutableStateOf(Offset.Zero) }
    var patchRect by remember { mutableStateOf(FencePatchRectState()) }
    var batchNonce by remember { mutableIntStateOf(0) }
    var batchPoints by remember { mutableStateOf<List<Pair<Float, Float>>>(emptyList()) }
    var pendingPatchSave by remember { mutableStateOf(false) }
    var forcedPatchSizePx by remember { mutableStateOf<Pair<Float, Float>?>(null) }
    var sizeWidthM by remember { mutableStateOf("20") }
    var sizeHeightM by remember { mutableStateOf("10") }
    var projectNonce by remember { mutableIntStateOf(0) }
    var screenVertexes by remember { mutableStateOf<List<Pair<Float, Float>>>(emptyList()) }
    var pendingMoveIndex by remember { mutableIntStateOf(-1) }
    var pendingAddPoint by remember { mutableStateOf(false) }
    var patchSeed by remember { mutableStateOf<FencePatchRectState?>(null) }
    var patchSeedNonce by remember { mutableIntStateOf(0) }
    var pendingPatchSeed by remember { mutableStateOf(false) }
    var editCameraReady by remember { mutableStateOf(!edit.isEdit) }
    var patchSeedDone by remember { mutableStateOf(false) }
    var allowPatchSeed by remember { mutableStateOf(false) }
    var cameraZoom by remember { mutableFloatStateOf(17f) }
    var cameraLat by remember { mutableStateOf(followLat ?: 0.0) }
    var cameraLng by remember { mutableStateOf(followLng ?: 0.0) }
    LaunchedEffect(followNonce) {
        if (followNonce > 0) cameraZoom = 17f
    }
    val metersPerPixel = remember(cameraZoom, cameraLat) {
        val z = cameraZoom.toDouble().coerceIn(3.0, 22.0)
        156_543.03392 * kotlin.math.abs(kotlin.math.cos(cameraLat * kotlin.math.PI / 180.0))
            .coerceAtLeast(0.2) / 2.0.pow(z)
    }

    // 进入编辑：对�?XMapCoverView.setMap �?delay �?initPoints；贴片等 follow 后再投影
    LaunchedEffect(edit.editingId, edit.isEdit, edit.drawMode) {
        patchSeed = null
        patchSeedNonce = 0
        pendingPatchSeed = false
        patchSeedDone = false
        allowPatchSeed = false
        editCameraReady = !edit.isEdit
        if (!edit.isEdit) return@LaunchedEffect
        // 先让 follow 动画跑起来，忽略期间�?idle
        kotlinx.coroutines.delay(450)
        allowPatchSeed = true
        editCameraReady = true
        if (edit.drawMode == FenceDrawMode.Patch && edit.draftPoints.size == 4) {
            pendingPatchSeed = true
        }
        projectNonce += 1
        // 兜底：若 idle 未到，再�?
        kotlinx.coroutines.delay(700)
        if (edit.drawMode == FenceDrawMode.Patch &&
            edit.draftPoints.size == 4 &&
            !patchSeedDone
        ) {
            pendingPatchSeed = true
            projectNonce += 1
        }
    }

    val isPatch = edit.drawMode == FenceDrawMode.Patch
    val title = if (edit.isEdit) {
        translate(Str.FenceEdit, emptyArray())
    } else if (edit.kind == FenceKind.NoParking) {
        translate(Str.FenceCreateNoParking, emptyArray())
    } else {
        translate(Str.FenceCreateParking, emptyArray())
    }
    val canSavePoint = !edit.saving && edit.draftPoints.size >= 3
    val canSavePatch = !edit.saving && patchRect.isReady && (!edit.isEdit || patchSeedDone)
    val canSave = if (isPatch) canSavePatch else canSavePoint
    val canUndo = !isPatch && edit.draftPoints.isNotEmpty() && !edit.saving
    val isParking = edit.kind == FenceKind.Parking
    val showDraw = true
    val showTabs = !edit.isEdit
    val edgeMeters = remember(edit.draftPoints) {
        val pts = edit.draftPoints
        if (pts.size < 2) {
            emptyList()
        } else if (pts.size == 2) {
            listOf(metersBetween(pts[0], pts[1]).roundToInt())
        } else {
            pts.indices.map { i ->
                metersBetween(pts[i], pts[(i + 1) % pts.size]).roundToInt()
            }
        }
    }

    fun requestPatchSave() {
        val corners = patchRect.corners()
        if (corners.size < 4) return
        batchPoints = corners.map { corner -> corner.first to corner.second }
        pendingPatchSave = true
        batchNonce += 1
    }

    fun doPersist() {
        if (isPatch) requestPatchSave() else onSave()
    }

    fun requestSave() {
        if (isParking && !edit.isEdit && !edit.parkingParamsConfirmed) {
            pendingSaveAfterParams = true
            showParams = true
            return
        }
        doPersist()
    }

    fun requestAddAtCrosshair() {
        pendingMoveIndex = -1
        pendingAddPoint = true
        screenPickX = crosshairCenterInRoot.x - mapOriginInRoot.x
        screenPickY = crosshairCenterInRoot.y - mapOriginInRoot.y
        screenToLatLngNonce += 1
    }

    // 点数变化时重投影；拖顶点过程中不投影
    LaunchedEffect(edit.draftPoints, isPatch, showDraw, pendingMoveIndex, editCameraReady) {
        if (!isPatch && showDraw && pendingMoveIndex < 0 && editCameraReady) {
            projectNonce += 1
        }
    }

    Column(modifier = Modifier.fillMaxSize()) {
        FenceBrowseTopBar(
            title = title,
            primary = primary,
            listMode = false,
            editMode = false,
            editLabel = "",
            cancelEditLabel = "",
            onBack = onBack,
            onToggleMode = {},
            onToggleEdit = null,
        )
        if (showTabs && showDraw) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(Color.White)
                    .padding(horizontal = 24.dp, vertical = 8.dp),
                horizontalArrangement = Arrangement.spacedBy(32.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                FenceDrawTabChip(
                    icon = OpsIcon.FenceTabPatch,
                    label = translate(Str.FenceTabPatch, emptyArray()),
                    selected = isPatch,
                    primary = primary,
                    onClick = { editFeature.setDrawMode(FenceDrawMode.Patch) },
                )
                FenceDrawTabChip(
                    icon = OpsIcon.FenceTabPoint,
                    label = translate(Str.FenceTabPoint, emptyArray()),
                    selected = !isPatch,
                    primary = primary,
                    onClick = { editFeature.setDrawMode(FenceDrawMode.Point) },
                )
            }
        }

        if (!showDraw) {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                CircularProgressIndicator(color = primary)
            }
        } else {
            Box(modifier = Modifier.weight(1f).fillMaxWidth()) {
                OpsMapView(
                    spec = OpsMapSpec(
                        pins = emptyList(),
                        clusterOverview = false,
                        fencePolygons = backgroundPolygons,
                        trackPoints = emptyList(),
                        onMapTap = null,
                        onCameraIdle = { lat, lng ->
                            cameraLat = lat
                            cameraLng = lng
                            // zoom �?followZoom/交互近似；无 SDK zoom 回调时用当前�?
                            if (isPatch && edit.isEdit && allowPatchSeed && !patchSeedDone &&
                                edit.draftPoints.size == 4
                            ) {
                                pendingPatchSeed = true
                                projectNonce += 1
                            } else if (edit.isEdit && !editCameraReady && !isPatch) {
                                editCameraReady = true
                                projectNonce += 1
                            } else if (!isPatch && pendingMoveIndex < 0) {
                                projectNonce += 1
                            }
                        },
                        onCameraMove = { lat, lng ->
                            cameraLat = lat
                            cameraLng = lng
                            if (!isPatch && showDraw && pendingMoveIndex < 0 && editCameraReady) {
                                projectNonce += 1
                            }
                        },
                        screenToLatLngNonce = screenToLatLngNonce,
                        screenPickX = screenPickX,
                        screenPickY = screenPickY,
                        onScreenToLatLng = { lat, lng ->
                            val moveIdx = pendingMoveIndex
                            if (moveIdx >= 0) {
                                editFeature.movePoint(moveIdx, lat, lng)
                            } else if (pendingAddPoint) {
                                pendingAddPoint = false
                                editFeature.addPoint(lat, lng)
                            }
                        },
                        batchScreenToLatLngNonce = batchNonce,
                        batchScreenPoints = batchPoints,
                        onBatchScreenToLatLng = { mapped ->
                            if (pendingPatchSave && mapped.size >= 4) {
                                pendingPatchSave = false
                                editFeature.replacePoints(
                                    mapped.map { (lat, lng) -> GeoLatLng(lat, lng) },
                                )
                                onSave()
                            } else {
                                pendingPatchSave = false
                            }
                        },
                        latLngToScreenNonce = projectNonce,
                        latLngToScreenPoints = edit.draftPoints.map { it.lat to it.lng },
                        onLatLngToScreen = { screens ->
                            screenVertexes = screens
                            if (pendingPatchSeed && screens.size >= 4) {
                                pendingPatchSeed = false
                                // 屏外/未投影成功的点一律丢弃，避免回填到海�?
                                val seeded = FencePatchRectState.fromScreenCorners(screens)
                                if (seeded != null) {
                                    patchSeed = seeded
                                    patchSeedNonce += 1
                                    patchRect = seeded
                                    patchSeedDone = true
                                }
                            }
                        },
                        autoFitOnPins = false,
                        followNonce = followNonce,
                        followLat = followLat,
                        followLng = followLng,
                        followZoom = 17f,
                        mapTypeSatellite = mapTypeSatellite,
                        showStatusOverlay = false,
                    ),
                    modifier = Modifier
                        .fillMaxSize()
                        .onGloballyPositioned { mapOriginInRoot = it.positionInRoot() },
                )
                Image(
                    painter = painterResource(
                        if (mapTypeSatellite) OpsIcon.MapSatelliteSelected else OpsIcon.MapSatellite,
                    ),
                    contentDescription = translate(Str.MapToolSatellite, emptyArray()),
                    modifier = Modifier
                        .align(Alignment.TopEnd)
                        .padding(12.dp)
                        .size(36.dp)
                        .clickable(onClick = onToggleSatellite),
                    contentScale = ContentScale.Fit,
                )
                if (onLocate != null) {
                    Image(
                        painter = painterResource(OpsIcon.MapLocation),
                        contentDescription = translate(Str.MapToolLocate, emptyArray()),
                        modifier = Modifier
                            .align(Alignment.BottomStart)
                            .padding(start = 12.dp, bottom = 24.dp)
                            .size(36.dp)
                            .clickable(onClick = onLocate),
                        contentScale = ContentScale.Fit,
                    )
                }
                if (isPatch) {
                    FencePatchOverlay(
                        isParking = isParking,
                        onRectChanged = { patchRect = it },
                        forcedSizePx = forcedPatchSizePx,
                        seedRect = patchSeed,
                        seedNonce = patchSeedNonce,
                        allowDefaultInit = true,
                        metersPerPixel = metersPerPixel,
                        modifier = Modifier.fillMaxSize(),
                    )
                } else {
                    FenceSelectPointOverlay(
                        isParking = isParking,
                        screenPoints = screenVertexes,
                        edgeMeters = edgeMeters,
                        onMovePointScreen = { index, x, y ->
                            pendingAddPoint = false
                            pendingMoveIndex = index
                            screenPickX = x
                            screenPickY = y
                            screenToLatLngNonce += 1
                        },
                        onMoveFinished = {
                            pendingMoveIndex = -1
                            projectNonce += 1
                        },
                        modifier = Modifier.fillMaxSize(),
                    )
                    // 对齐原版：新�?编辑都有准星 + Pick Point + 撤销
                    Text(
                        text = translate(Str.FenceDrawHint, emptyArray()) + " (${edit.draftPoints.size})",
                        fontSize = 12.sp,
                        modifier = Modifier
                            .align(Alignment.TopCenter)
                            .padding(8.dp)
                            .background(Color.White.copy(alpha = 0.9f), RoundedCornerShape(6.dp))
                            .padding(horizontal = 10.dp, vertical = 6.dp),
                    )
                    Image(
                        painter = painterResource(OpsIcon.MapCenterPoint),
                        contentDescription = null,
                        modifier = Modifier
                            .align(Alignment.TopCenter)
                            .padding(top = 184.dp)
                            .size(width = 24.dp, height = 50.dp)
                            .onGloballyPositioned { coords ->
                                crosshairCenterInRoot = coords.boundsInRoot().center
                            },
                        contentScale = ContentScale.Fit,
                    )
                    Image(
                        painter = painterResource(OpsIcon.SelectMapPoint),
                        contentDescription = translate(Str.FenceAddPoint, emptyArray()),
                        modifier = Modifier
                            .align(Alignment.BottomCenter)
                            .padding(bottom = 26.dp)
                            .size(width = 128.dp, height = 48.dp)
                            .clickable(enabled = !edit.saving, onClick = ::requestAddAtCrosshair),
                        contentScale = ContentScale.Fit,
                    )
                    Image(
                        painter = painterResource(
                            if (canUndo) OpsIcon.FenceUndoAble else OpsIcon.FenceUndoDisable,
                        ),
                        contentDescription = translate(Str.FenceUndoPoint, emptyArray()),
                        modifier = Modifier
                            .align(Alignment.BottomEnd)
                            .padding(end = 20.dp, bottom = 26.dp)
                            .size(56.dp)
                            .clickable(enabled = canUndo, onClick = editFeature::undoPoint),
                        contentScale = ContentScale.Fit,
                    )
                }
            }

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(Color.White)
                    .padding(start = 10.dp, end = 24.dp, top = 20.dp, bottom = 18.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Row(
                    modifier = Modifier.weight(1f),
                    horizontalArrangement = Arrangement.spacedBy(16.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    if (isParking) {
                        Column(
                            horizontalAlignment = Alignment.CenterHorizontally,
                            modifier = Modifier.clickable(enabled = if (isPatch) canSavePatch else canSavePoint) {
                                showAngle = true
                            },
                        ) {
                            Image(
                                painter = painterResource(
                                    if (canSave) OpsIcon.FenceAngleAble else OpsIcon.FenceAngleDisable,
                                ),
                                contentDescription = translate(Str.FenceDirectionLabel, emptyArray()),
                                modifier = Modifier.size(32.dp),
                                contentScale = ContentScale.Fit,
                            )
                            Text(
                                translate(Str.FenceDirectionLabel, emptyArray()),
                                fontSize = 11.sp,
                                color = Color(0xFF666666),
                                maxLines = 1,
                                overflow = TextOverflow.Ellipsis,
                            )
                        }
                    }
                    if (isPatch) {
                        Column(
                            horizontalAlignment = Alignment.CenterHorizontally,
                            modifier = Modifier.clickable { showSize = true },
                        ) {
                            Image(
                                painter = painterResource(OpsIcon.FenceModifySize),
                                contentDescription = translate(Str.FenceModifySize, emptyArray()),
                                modifier = Modifier.size(32.dp),
                                contentScale = ContentScale.Fit,
                            )
                            Text(
                                translate(Str.FenceModifySize, emptyArray()),
                                fontSize = 11.sp,
                                color = Color(0xFF666666),
                                maxLines = 1,
                                overflow = TextOverflow.Ellipsis,
                            )
                        }
                    }
                    if (isParking && (!isPatch || edit.isEdit)) {
                        Column(
                            horizontalAlignment = Alignment.CenterHorizontally,
                            modifier = Modifier.clickable { showParams = true },
                        ) {
                            Image(
                                painter = painterResource(OpsIcon.FenceEditParams),
                                contentDescription = translate(Str.FenceEditParams, emptyArray()),
                                modifier = Modifier.size(32.dp),
                                contentScale = ContentScale.Fit,
                            )
                            Text(
                                translate(Str.FenceEditParams, emptyArray()),
                                fontSize = 11.sp,
                                color = Color(0xFF666666),
                                maxLines = 1,
                                overflow = TextOverflow.Ellipsis,
                            )
                        }
                    }
                    // 禁停编辑：参数（名称�?
                    if (!isParking && edit.isEdit) {
                        Column(
                            horizontalAlignment = Alignment.CenterHorizontally,
                            modifier = Modifier.clickable { showNoParkingName = true },
                        ) {
                            Image(
                                painter = painterResource(OpsIcon.FenceEditParams),
                                contentDescription = translate(Str.FenceEditParams, emptyArray()),
                                modifier = Modifier.size(32.dp),
                                contentScale = ContentScale.Fit,
                            )
                            Text(
                                translate(Str.FenceEditParams, emptyArray()),
                                fontSize = 11.sp,
                                color = Color(0xFF666666),
                                maxLines = 1,
                                overflow = TextOverflow.Ellipsis,
                            )
                        }
                    }
                }
                Button(
                    onClick = ::requestSave,
                    enabled = canSave,
                    modifier = Modifier
                        .weight(1f)
                        .height(48.dp),
                ) {
                    Text(translate(Str.FenceSave, emptyArray()))
                }
            }
            edit.errorMessage?.let {
                Text(
                    it,
                    color = MaterialTheme.colorScheme.error,
                    fontSize = 13.sp,
                    modifier = Modifier
                        .fillMaxWidth()
                        .background(Color.White)
                        .padding(horizontal = 16.dp, vertical = 4.dp),
                )
            }
        }
    }

    if (showParams && isParking) {
        FenceParkingParamsDialog(
            edit = edit,
            editFeature = editFeature,
            translate = translate,
            isCreate = !edit.isEdit,
            onConfirm = {
                if (editFeature.confirmParkingParams()) {
                    showParams = false
                    if (pendingSaveAfterParams) {
                        pendingSaveAfterParams = false
                        doPersist()
                    }
                }
            },
            onDismiss = {
                pendingSaveAfterParams = false
                showParams = false
            },
        )
    }

    if (showNoParkingName) {
        Dialog(onDismissRequest = { showNoParkingName = false }) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(Color.White, RoundedCornerShape(12.dp))
                    .padding(20.dp),
            ) {
                Text(translate(Str.FenceEditParams, emptyArray()), fontWeight = FontWeight.Bold)
                // spacing
                OutlinedTextField(
                    value = edit.name,
                    onValueChange = editFeature::setName,
                    modifier = Modifier.fillMaxWidth(),
                    singleLine = true,
                )
                Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.End) {
                    TextButton(onClick = { showNoParkingName = false }) {
                        Text(translate(Str.Cancel, emptyArray()))
                    }
                    TextButton(onClick = { showNoParkingName = false }) {
                        Text(translate(Str.Confirm, emptyArray()))
                    }
                }
            }
        }
    }

    if (showAngle && isParking) {
        FenceDirectionDialog(
            value = edit.direction,
            title = translate(Str.FenceDirectionLabel, emptyArray()),
            confirmLabel = translate(Str.Confirm, emptyArray()),
            cancelLabel = translate(Str.Cancel, emptyArray()),
            onValueChange = editFeature::setDirection,
            onConfirm = { showAngle = false },
            onDismiss = { showAngle = false },
        )
    }

    if (showSize && isPatch) {
        Dialog(onDismissRequest = { showSize = false }) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(Color.White, RoundedCornerShape(10.dp))
                    .padding(16.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                Text(translate(Str.FenceModifySize, emptyArray()), fontWeight = FontWeight.SemiBold, fontSize = 16.sp)
                OutlinedTextField(
                    value = sizeWidthM,
                    onValueChange = { sizeWidthM = it.filter { ch -> ch.isDigit() }.take(3) },
                    label = { Text("W (m)") },
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth(),
                )
                OutlinedTextField(
                    value = sizeHeightM,
                    onValueChange = { sizeHeightM = it.filter { ch -> ch.isDigit() }.take(3) },
                    label = { Text("H (m)") },
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth(),
                )
                Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.End) {
                    TextButton(onClick = { showSize = false }) {
                        Text(translate(Str.Cancel, emptyArray()))
                    }
                    TextButton(onClick = {
                        val wM = sizeWidthM.toFloatOrNull()?.coerceIn(10f, 100f) ?: return@TextButton
                        val hM = sizeHeightM.toFloatOrNull()?.coerceIn(10f, 100f) ?: return@TextButton
                        val mpp = metersPerPixel.toFloat().coerceAtLeast(0.01f)
                        forcedPatchSizePx = ((wM + 1f) / mpp) to ((hM + 1f) / mpp)
                        showSize = false
                    }) {
                        Text(translate(Str.Confirm, emptyArray()))
                    }
                }
            }
        }
    }
}

@Composable
private fun FenceDrawTabChip(
    icon: OpsIcon,
    label: String,
    selected: Boolean,
    primary: Color,
    onClick: () -> Unit,
) {
    Row(
        verticalAlignment = Alignment.CenterVertically,
        modifier = Modifier.clickable(onClick = onClick),
    ) {
        Image(
            painter = painterResource(icon),
            contentDescription = label,
            modifier = Modifier
                .size(22.dp)
                .then(if (selected) Modifier else Modifier),
            contentScale = ContentScale.Fit,
        )
        Spacer(modifier = Modifier.width(6.dp))
        Text(
            text = label,
            color = if (selected) primary else Color(0xFF666666),
            fontWeight = if (selected) FontWeight.SemiBold else FontWeight.Normal,
            fontSize = 15.sp,
        )
    }
}

@Composable
private fun FenceParkingParamsDialog(
    edit: FenceEditUiState,
    editFeature: FenceEditFeature,
    translate: (Str, Array<out Any?>) -> String,
    isCreate: Boolean,
    onConfirm: () -> Unit,
    onDismiss: () -> Unit,
) {
    Dialog(onDismissRequest = onDismiss) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color.White, RoundedCornerShape(10.dp))
                .padding(16.dp)
                .verticalScroll(rememberScrollState()),
            verticalArrangement = Arrangement.spacedBy(10.dp),
        ) {
            Text(
                translate(Str.FenceEditParams, emptyArray()),
                fontWeight = FontWeight.SemiBold,
                fontSize = 16.sp,
            )
            if (!isCreate) {
                OutlinedTextField(
                    value = edit.name,
                    onValueChange = editFeature::setName,
                    label = { Text(translate(Str.SiteName, emptyArray())) },
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth(),
                )
            }
            OutlinedTextField(
                value = edit.maxParkingNumber,
                onValueChange = editFeature::setMaxParkingNumber,
                label = { Text(translate(Str.FenceCapacityLabel, emptyArray())) },
                singleLine = true,
                modifier = Modifier.fillMaxWidth(),
            )
            OutlinedTextField(
                value = edit.bufferDistance,
                onValueChange = editFeature::setBufferDistance,
                label = { Text(translate(Str.FenceBufferLabel, emptyArray())) },
                singleLine = true,
                modifier = Modifier.fillMaxWidth(),
            )
            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween,
            ) {
                Text(translate(Str.FenceEnable, emptyArray()), fontSize = 14.sp)
                Switch(checked = edit.izEnable, onCheckedChange = editFeature::setIzEnable)
            }
            Text(translate(Str.FenceReturnModeLabel, emptyArray()), fontWeight = FontWeight.Medium, fontSize = 14.sp)
            FenceParamSwitchRow(translate(Str.FenceReturnTbeacon, emptyArray()), edit.tbeacon, editFeature::setTbeacon)
            FenceParamSwitchRow(translate(Str.FenceReturnRfid, emptyArray()), edit.rfid, editFeature::setRfid)
            FenceParamSwitchRow(translate(Str.FenceReturnDirectional, emptyArray()), edit.directional, editFeature::setDirectional)
            FenceParamSwitchRow(translate(Str.FenceReturnKickstand, emptyArray()), edit.kickstand, editFeature::setKickstand)
            FenceParamSwitchRow(translate(Str.FenceReturnCamera, emptyArray()), edit.camera, editFeature::setCamera)
            FenceParamSwitchRow(translate(Str.FenceFullPileNoStop, emptyArray()), edit.izFullPileNoStop, editFeature::setIzFullPileNoStop)
            edit.errorMessage?.let {
                Text(it, color = MaterialTheme.colorScheme.error, fontSize = 13.sp)
            }
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.End,
            ) {
                TextButton(onClick = onDismiss) { Text(translate(Str.Cancel, emptyArray())) }
                TextButton(onClick = onConfirm) { Text(translate(Str.Confirm, emptyArray())) }
            }
        }
    }
}

@Composable
private fun FenceParamSwitchRow(
    label: String,
    checked: Boolean,
    onCheckedChange: (Boolean) -> Unit,
) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.SpaceBetween,
    ) {
        Text(label, fontSize = 14.sp)
        Switch(checked = checked, onCheckedChange = onCheckedChange)
    }
}

@Composable
private fun FenceDirectionDialog(
    value: String,
    title: String,
    confirmLabel: String,
    cancelLabel: String,
    onValueChange: (String) -> Unit,
    onConfirm: () -> Unit,
    onDismiss: () -> Unit,
) {
    Dialog(onDismissRequest = onDismiss) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color.White, RoundedCornerShape(10.dp))
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            Text(title, fontWeight = FontWeight.SemiBold, fontSize = 16.sp)
            OutlinedTextField(
                value = value,
                onValueChange = onValueChange,
                label = { Text(title) },
                singleLine = true,
                modifier = Modifier.fillMaxWidth(),
            )
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.End,
            ) {
                TextButton(onClick = onDismiss) { Text(cancelLabel) }
                TextButton(onClick = onConfirm) { Text(confirmLabel) }
            }
        }
    }
}

@Composable
private fun FenceMapBrowseContent(
    state: FenceBrowseUiState,
    feature: FenceBrowseFeature,
    mapPins: List<MapPin>,
    mapPolygons: List<FencePolygon>,
    focusLat: Double?,
    focusLng: Double?,
    fitNonce: Int,
    followNonce: Int,
    mapTypeSatellite: Boolean,
    translate: (Str, Array<out Any?>) -> String,
    onFollow: () -> Unit,
    onToggleSatellite: () -> Unit,
    onNewFence: () -> Unit,
    onEnable: (String) -> Unit,
    onDisable: (String) -> Unit,
    onDelete: (String, FenceKind) -> Unit,
    onEdit: (FencePolygon) -> Unit,
) {
    Box(modifier = Modifier.fillMaxSize()) {
        OpsMapView(
            spec = OpsMapSpec(
                pins = mapPins,
                selectedCarId = state.selectedFenceId,
                onSelectCarId = { id -> feature.selectFence(id) },
                clusterOverview = false,
                fencePolygons = mapPolygons,
                fitNonce = fitNonce,
                followNonce = followNonce,
                followLat = focusLat,
                followLng = focusLng,
                mapTypeSatellite = mapTypeSatellite,
                showStatusOverlay = false,
            ),
            modifier = Modifier.fillMaxSize(),
        )
        FenceMapSideTools(
            showAllLabel = translate(Str.FenceShowAll, emptyArray()),
            locateLabel = translate(Str.MapToolLocate, emptyArray()),
            satelliteLabel = translate(Str.MapToolSatellite, emptyArray()),
            showAllSelected = state.showAllFences,
            satelliteSelected = mapTypeSatellite,
            onToggleShowAll = { feature.setShowAllFences(!state.showAllFences) },
            onLocate = onFollow,
            onToggleSatellite = onToggleSatellite,
            modifier = Modifier
                .align(Alignment.BottomStart)
                .padding(start = 12.dp, bottom = 120.dp),
        )
        Column(
            modifier = Modifier
                .align(Alignment.BottomCenter)
                .fillMaxWidth()
                .background(Color.White.copy(alpha = 0.96f))
                .padding(horizontal = 16.dp, vertical = 10.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            FenceMapBottomBar(
                parkingLabel = translate(
                    Str.FenceSectionCount,
                    arrayOf<Any?>(translate(Str.FenceParkingSection, emptyArray()), state.parkingCount),
                ),
                noParkingLabel = translate(
                    Str.FenceSectionCount,
                    arrayOf<Any?>(translate(Str.FenceNoParkingSection, emptyArray()), state.noParkingCount),
                ),
            )
            Button(
                onClick = onNewFence,
                modifier = Modifier.fillMaxWidth(),
            ) {
                Text(translate(Str.FenceNew, emptyArray()))
            }
        }
        state.selectedFence()?.let { fence ->
            FenceDetailCard(
                fence = fence,
                translate = translate,
                onDismiss = { feature.selectFence(null) },
                onEnable = { onEnable(fence.id) },
                onDisable = { onDisable(fence.id) },
                onDelete = { onDelete(fence.id, fence.kind) },
                onEdit = { onEdit(fence) },
                modifier = Modifier
                    .align(Alignment.BottomCenter)
                    .padding(12.dp)
                    .padding(bottom = 110.dp),
            )
        }
    }
}

@Composable
private fun FenceListBrowseContent(
    state: FenceBrowseUiState,
    feature: FenceBrowseFeature,
    primary: Color,
    translate: (Str, Array<out Any?>) -> String,
    onOpenTags: () -> Unit,
    onOpenStatus: () -> Unit,
    onEnableSelected: () -> Unit,
    onDisableSelected: () -> Unit,
    onDeleteSelected: () -> Unit,
    onEnable: (String) -> Unit,
    onDisable: (String) -> Unit,
    onDelete: (String, FenceKind) -> Unit,
    onEdit: (FencePolygon) -> Unit,
) {
    Column(modifier = Modifier.fillMaxSize()) {
        FenceListTabs(
            parkingLabel = translate(
                Str.FenceSectionCount,
                arrayOf<Any?>(translate(Str.FenceParkingSection, emptyArray()), state.parkingCount),
            ),
            noParkingLabel = translate(
                Str.FenceSectionCount,
                arrayOf<Any?>(translate(Str.FenceNoParkingSection, emptyArray()), state.noParkingCount),
            ),
            selected = state.listTab,
            primary = primary,
            onSelect = { feature.setListTab(it) },
        )
        FenceFilterBar(
            tagLabel = if (state.selectedTagIds.isEmpty()) {
                translate(Str.StationFilterAllTags, emptyArray())
            } else {
                "${translate(Str.StationFilterAllTags, emptyArray())}(${state.selectedTagIds.size})"
            },
            statusLabel = when (state.optState) {
                StationOptStateFilter.All -> translate(Str.StationFilterAllStatus, emptyArray())
                StationOptStateFilter.Operating -> translate(Str.StationStatusOperating, emptyArray())
                StationOptStateFilter.Stopped -> translate(Str.StationStatusStopped, emptyArray())
            },
            onTag = onOpenTags,
            onStatus = onOpenStatus,
        )
        Text(
            text = translate(Str.FenceBrowseHint, emptyArray()),
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp),
        )
        val items = state.listItems()
        if (items.isEmpty()) {
            Text(
                text = translate(Str.AdminEmptyList, emptyArray()),
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.padding(24.dp),
            )
        } else {
            LazyColumn(
                modifier = Modifier
                    .weight(1f)
                    .fillMaxWidth()
                    .background(Color(0xFFF6F7F9)),
            ) {
                items(items, key = { it.id }) { fence ->
                    FenceListRow(
                        fence = fence,
                        editMode = state.editMode,
                        selected = fence.id in state.selectedIds,
                        clickable = !state.editMode,
                        translate = translate,
                        onToggleSelect = { feature.toggleSelected(fence.id) },
                        onClick = {
                            if (!state.editMode) {
                                feature.selectFence(fence.id)
                            }
                        },
                    )
                }
            }
        }
        if (state.editMode) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(Color.White)
                    .padding(12.dp),
                horizontalArrangement = Arrangement.SpaceEvenly,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                if (state.listTab == FenceListTab.Parking) {
                    FenceActionIconButton(
                        icon = OpsIcon.CommonAble,
                        label = translate(Str.FenceEnable, emptyArray()),
                        onClick = onEnableSelected,
                    )
                    FenceActionIconButton(
                        icon = OpsIcon.CommonUnable,
                        label = translate(Str.FenceDisable, emptyArray()),
                        onClick = onDisableSelected,
                    )
                }
                FenceActionIconButton(
                    icon = OpsIcon.CommonDelete,
                    label = translate(Str.FenceDelete, emptyArray()),
                    onClick = onDeleteSelected,
                )
            }
        }
        state.selectedFence()?.takeIf { !state.editMode }?.let { fence ->
            Dialog(onDismissRequest = { feature.selectFence(null) }) {
                FenceDetailCard(
                    fence = fence,
                    translate = translate,
                    onDismiss = { feature.selectFence(null) },
                    onEnable = { onEnable(fence.id) },
                    onDisable = { onDisable(fence.id) },
                    onDelete = { onDelete(fence.id, fence.kind) },
                    onEdit = { onEdit(fence) },
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(16.dp),
                )
            }
        }
    }
}

@Composable
private fun FencePickTypeDialog(
    title: String,
    parkingLabel: String,
    noParkingLabel: String,
    cancelLabel: String,
    onParking: () -> Unit,
    onNoParking: () -> Unit,
    onDismiss: () -> Unit,
) {
    Dialog(onDismissRequest = onDismiss) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color.White, RoundedCornerShape(10.dp))
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(4.dp),
        ) {
            Text(title, fontWeight = FontWeight.SemiBold, fontSize = 16.sp)
            Text(
                text = parkingLabel,
                modifier = Modifier
                    .fillMaxWidth()
                    .clickable(onClick = onParking)
                    .padding(vertical = 12.dp),
            )
            Text(
                text = noParkingLabel,
                modifier = Modifier
                    .fillMaxWidth()
                    .clickable(onClick = onNoParking)
                    .padding(vertical = 12.dp),
            )
            TextButton(onClick = onDismiss, modifier = Modifier.align(Alignment.End)) {
                Text(cancelLabel)
            }
        }
    }
}

@Composable
private fun FenceBrowseTopBar(
    title: String,
    primary: Color,
    listMode: Boolean,
    editMode: Boolean,
    editLabel: String,
    cancelEditLabel: String,
    onBack: () -> Unit,
    onToggleMode: () -> Unit,
    onToggleEdit: (() -> Unit)?,
) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .background(primary)
            .statusBarsPadding()
            .height(48.dp)
            .padding(horizontal = 8.dp),
    ) {
        Image(
            painter = painterResource(OpsIcon.ChevronLeft),
            contentDescription = null,
            modifier = Modifier
                .align(Alignment.CenterStart)
                .size(28.dp)
                .clickable(
                    interactionSource = remember { MutableInteractionSource() },
                    indication = null,
                    onClick = onBack,
                )
                .padding(4.dp),
            contentScale = ContentScale.Fit,
        )
        Row(
            modifier = Modifier
                .align(Alignment.Center)
                .clickable(
                    interactionSource = remember { MutableInteractionSource() },
                    indication = null,
                    onClick = onToggleMode,
                ),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(
                text = title,
                color = Color.White,
                fontSize = 17.sp,
                fontWeight = FontWeight.SemiBold,
            )
            Spacer(modifier = Modifier.width(6.dp))
            Image(
                painter = painterResource(OpsIcon.CommonSwitch),
                contentDescription = null,
                modifier = Modifier.size(18.dp),
                contentScale = ContentScale.Fit,
            )
        }
        if (listMode && onToggleEdit != null) {
            Text(
                text = if (editMode) cancelEditLabel else editLabel,
                color = Color.White,
                fontSize = 14.sp,
                modifier = Modifier
                    .align(Alignment.CenterEnd)
                    .clickable(onClick = onToggleEdit)
                    .padding(8.dp),
            )
        }
    }
}

@Composable
private fun FenceMapSideTools(
    showAllLabel: String,
    locateLabel: String,
    satelliteLabel: String,
    showAllSelected: Boolean,
    satelliteSelected: Boolean,
    onToggleShowAll: () -> Unit,
    onLocate: () -> Unit,
    onToggleSatellite: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Column(
        modifier = modifier
            .width(48.dp)
            .shadow(3.dp, RoundedCornerShape(8.dp))
            .background(Color.White, RoundedCornerShape(8.dp))
            .padding(vertical = 6.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(4.dp),
    ) {
        FenceSideTool(
            icon = if (showAllSelected) OpsIcon.MapMoreSelected else OpsIcon.MapMore,
            label = showAllLabel,
            onClick = onToggleShowAll,
        )
        FenceSideTool(
            icon = OpsIcon.MapLocation,
            label = locateLabel,
            onClick = onLocate,
        )
        FenceSideTool(
            icon = if (satelliteSelected) OpsIcon.MapSatelliteSelected else OpsIcon.MapSatellite,
            label = satelliteLabel,
            onClick = onToggleSatellite,
        )
    }
}

@Composable
private fun FenceSideTool(icon: OpsIcon, label: String, onClick: () -> Unit) {
    Column(
        modifier = Modifier
            .clickable(onClick = onClick)
            .padding(vertical = 6.dp, horizontal = 4.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Image(
            painter = painterResource(icon),
            contentDescription = label,
            modifier = Modifier.size(22.dp),
            contentScale = ContentScale.Fit,
        )
        Spacer(modifier = Modifier.height(2.dp))
        Text(
            text = label,
            color = Color(0xFF48506C),
            fontSize = 10.sp,
            fontWeight = FontWeight.Medium,
        )
    }
}

@Composable
private fun FenceMapBottomBar(
    parkingLabel: String,
    noParkingLabel: String,
    modifier: Modifier = Modifier,
) {
    Row(
        modifier = modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.SpaceEvenly,
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Row(verticalAlignment = Alignment.CenterVertically) {
            Image(
                painter = painterResource(OpsIcon.IconParking),
                contentDescription = null,
                modifier = Modifier.size(18.dp),
                contentScale = ContentScale.Fit,
            )
            Spacer(modifier = Modifier.width(6.dp))
            Text(parkingLabel, color = Color(0xFF48506C), fontSize = 14.sp)
        }
        Row(verticalAlignment = Alignment.CenterVertically) {
            Image(
                painter = painterResource(OpsIcon.IconNoParking),
                contentDescription = null,
                modifier = Modifier.size(18.dp),
                contentScale = ContentScale.Fit,
            )
            Spacer(modifier = Modifier.width(6.dp))
            Text(noParkingLabel, color = Color(0xFF48506C), fontSize = 14.sp)
        }
    }
}

@Composable
private fun FenceActionIconButton(
    icon: OpsIcon,
    label: String,
    onClick: () -> Unit,
) {
    Column(
        modifier = Modifier
            .clickable(onClick = onClick)
            .padding(horizontal = 8.dp, vertical = 4.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Image(
            painter = painterResource(icon),
            contentDescription = label,
            modifier = Modifier.size(24.dp),
            contentScale = ContentScale.Fit,
        )
        Spacer(modifier = Modifier.height(4.dp))
        Text(label, fontSize = 12.sp, color = Color(0xFF48506C))
    }
}

@Composable
private fun FenceListTabs(
    parkingLabel: String,
    noParkingLabel: String,
    selected: FenceListTab,
    primary: Color,
    onSelect: (FenceListTab) -> Unit,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .background(Color.White)
            .padding(horizontal = 16.dp),
    ) {
        FenceTabChip(
            label = parkingLabel,
            selected = selected == FenceListTab.Parking,
            primary = primary,
            onClick = { onSelect(FenceListTab.Parking) },
            modifier = Modifier.weight(1f),
        )
        FenceTabChip(
            label = noParkingLabel,
            selected = selected == FenceListTab.NoParking,
            primary = primary,
            onClick = { onSelect(FenceListTab.NoParking) },
            modifier = Modifier.weight(1f),
        )
    }
}

@Composable
private fun FenceTabChip(
    label: String,
    selected: Boolean,
    primary: Color,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Column(
        modifier = modifier
            .clickable(onClick = onClick)
            .padding(top = 12.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text(
            text = label,
            color = if (selected) primary else Color(0xFF666666),
            fontWeight = if (selected) FontWeight.SemiBold else FontWeight.Normal,
            fontSize = 15.sp,
        )
        Spacer(modifier = Modifier.height(10.dp))
        Box(
            modifier = Modifier
                .fillMaxWidth(0.45f)
                .height(2.dp)
                .background(if (selected) primary else Color.Transparent),
        )
    }
}

@Composable
private fun FenceFilterBar(
    tagLabel: String,
    statusLabel: String,
    onTag: () -> Unit,
    onStatus: () -> Unit,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .background(Color.White)
            .padding(horizontal = 16.dp, vertical = 10.dp),
        horizontalArrangement = Arrangement.spacedBy(24.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Row(
            verticalAlignment = Alignment.CenterVertically,
            modifier = Modifier.clickable(onClick = onTag),
        ) {
            Text(text = tagLabel, color = Color(0xFF48506C), fontSize = 14.sp)
            Spacer(modifier = Modifier.width(4.dp))
            Image(
                painter = painterResource(OpsIcon.ArrowDownBlack),
                contentDescription = null,
                modifier = Modifier.size(12.dp),
                contentScale = ContentScale.Fit,
            )
        }
        Row(
            verticalAlignment = Alignment.CenterVertically,
            modifier = Modifier.clickable(onClick = onStatus),
        ) {
            Text(text = statusLabel, color = Color(0xFF48506C), fontSize = 14.sp)
            Spacer(modifier = Modifier.width(4.dp))
            Image(
                painter = painterResource(OpsIcon.ArrowDownBlack),
                contentDescription = null,
                modifier = Modifier.size(12.dp),
                contentScale = ContentScale.Fit,
            )
        }
    }
    HorizontalDivider(color = Color(0xFFE8E8E8), thickness = 0.5.dp)
}

@Composable
private fun FenceListRow(
    fence: FencePolygon,
    editMode: Boolean,
    selected: Boolean,
    clickable: Boolean,
    translate: (Str, Array<out Any?>) -> String,
    onToggleSelect: () -> Unit,
    onClick: () -> Unit,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 12.dp, vertical = 6.dp)
            .background(Color.White, RoundedCornerShape(8.dp))
            .then(
                if (editMode) {
                    Modifier.clickable(onClick = onToggleSelect)
                } else if (clickable) {
                    Modifier.clickable(onClick = onClick)
                } else {
                    Modifier
                },
            )
            .padding(12.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        if (editMode) {
            Checkbox(
                checked = selected,
                onCheckedChange = { onToggleSelect() },
            )
            Spacer(modifier = Modifier.width(4.dp))
        }
        Column(
            modifier = Modifier.weight(1f),
            verticalArrangement = Arrangement.spacedBy(6.dp),
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Text(
                    text = fence.name,
                    fontWeight = FontWeight.Medium,
                    fontSize = 15.sp,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                    modifier = Modifier.weight(1f, fill = false).padding(end = 8.dp),
                )
                if (fence.kind == FenceKind.Parking) {
                    val statusText = when (fence.izEnable) {
                        true -> translate(Str.StationStatusOperating, emptyArray())
                        false -> translate(Str.StationStatusStopped, emptyArray())
                        null -> null
                    }
                    val statusColor = when (fence.izEnable) {
                        true -> Color(0xFF00BE59)
                        false -> Color(0xFFFF2222)
                        null -> Color(0xFF999999)
                    }
                    if (statusText != null) {
                        Text(text = statusText, color = statusColor, fontWeight = FontWeight.SemiBold, fontSize = 13.sp)
                    }
                }
            }
            if (fence.kind == FenceKind.Parking && fence.tags.isNotEmpty()) {
                Row(
                    modifier = Modifier.horizontalScroll(rememberScrollState()),
                    horizontalArrangement = Arrangement.spacedBy(6.dp),
                ) {
                    fence.tags.forEach { tag ->
                        Text(
                            text = tag.name,
                            color = Color(0xFF1180F9),
                            fontSize = 12.sp,
                            modifier = Modifier
                                .border(0.5.dp, Color(0xFF1180F9), RoundedCornerShape(4.dp))
                                .padding(horizontal = 6.dp, vertical = 2.dp),
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun FenceDetailCard(
    fence: FencePolygon,
    translate: (Str, Array<out Any?>) -> String,
    onDismiss: () -> Unit,
    onEnable: () -> Unit,
    onDisable: () -> Unit,
    onDelete: () -> Unit,
    onEdit: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Column(
        modifier = modifier
            .fillMaxWidth()
            .shadow(6.dp, RoundedCornerShape(10.dp))
            .background(Color.White, RoundedCornerShape(10.dp))
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(
                text = fence.name,
                fontWeight = FontWeight.SemiBold,
                fontSize = 16.sp,
                modifier = Modifier.weight(1f),
            )
            TextButton(onClick = onDismiss) {
                Image(
                    painter = painterResource(OpsIcon.BtnClose),
                    contentDescription = translate(Str.Close, emptyArray()),
                    modifier = Modifier.size(18.dp),
                    contentScale = ContentScale.Fit,
                )
            }
        }
        val typeLabel = when (fence.kind) {
            FenceKind.Parking -> translate(Str.FenceParkingSection, emptyArray())
            FenceKind.NoParking -> translate(Str.FenceTypeNoParking, emptyArray())
            else -> fence.kind.name
        }
        Text(typeLabel, color = Color(0xFF666666), fontSize = 13.sp)
        when (fence.izEnable) {
            true -> Text(translate(Str.StationStatusOperating, emptyArray()), color = Color(0xFF00BE59), fontSize = 13.sp)
            false -> Text(translate(Str.StationStatusStopped, emptyArray()), color = Color(0xFFFF2222), fontSize = 13.sp)
            null -> Unit
        }
        if (fence.address.isNotBlank()) {
            Text(fence.address, color = Color(0xFF666666), fontSize = 13.sp)
        }
        if (fence.kind == FenceKind.Parking) {
            if (fence.maxParkingNumber > 0 || fence.currentParkingNumber > 0) {
                Text(
                    translate(
                        Str.FenceParkingOccupancy,
                        arrayOf(fence.currentParkingNumber, fence.maxParkingNumber),
                    ),
                    color = Color(0xFF666666),
                    fontSize = 13.sp,
                )
            }
            if (fence.tags.isNotEmpty()) {
                Text(
                    fence.tags.joinToString(" · ") { it.name },
                    color = Color(0xFF1180F9),
                    fontSize = 13.sp,
                )
            }
        }
        fence.area?.let {
            Text(
                translate(Str.FenceAreaLabel, arrayOf(it)),
                color = Color(0xFF666666),
                fontSize = 13.sp,
            )
        }
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceEvenly,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            TextButton(onClick = onEdit) {
                Text(translate(Str.FenceEdit, emptyArray()))
            }
            if (fence.kind == FenceKind.Parking) {
                if (fence.izEnable == false) {
                    FenceActionIconButton(
                        icon = OpsIcon.CommonAble,
                        label = translate(Str.FenceEnable, emptyArray()),
                        onClick = onEnable,
                    )
                } else {
                    FenceActionIconButton(
                        icon = OpsIcon.CommonUnable,
                        label = translate(Str.FenceDisable, emptyArray()),
                        onClick = onDisable,
                    )
                }
            }
            FenceActionIconButton(
                icon = OpsIcon.CommonDelete,
                label = translate(Str.FenceDelete, emptyArray()),
                onClick = onDelete,
            )
        }
    }
}

@Composable
private fun FenceTagFilterDialog(
    tags: List<com.luopingtech.ebike.ops.domain.analysis.StationTag>,
    selected: Set<String>,
    title: String,
    clearLabel: String,
    confirmLabel: String,
    onToggle: (String) -> Unit,
    onClear: () -> Unit,
    onDismiss: () -> Unit,
) {
    Dialog(onDismissRequest = onDismiss) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color.White, RoundedCornerShape(10.dp))
                .padding(16.dp)
                .verticalScroll(rememberScrollState()),
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Text(title, fontWeight = FontWeight.SemiBold, fontSize = 16.sp)
            TextButton(onClick = onClear) { Text(clearLabel) }
            tags.forEach { tag ->
                val on = tag.id in selected
                Text(
                    text = tag.name,
                    color = if (on) OpsTheme.colors.primary else Color(0xFF333333),
                    modifier = Modifier
                        .fillMaxWidth()
                        .clickable { onToggle(tag.id) }
                        .padding(vertical = 8.dp),
                )
            }
            TextButton(onClick = onDismiss, modifier = Modifier.align(Alignment.End)) {
                Text(confirmLabel)
            }
        }
    }
}

@Composable
private fun FenceStatusFilterDialog(
    selected: StationOptStateFilter,
    allLabel: String,
    operatingLabel: String,
    stoppedLabel: String,
    onSelect: (StationOptStateFilter) -> Unit,
    onDismiss: () -> Unit,
) {
    Dialog(onDismissRequest = onDismiss) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color.White, RoundedCornerShape(10.dp))
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(4.dp),
        ) {
            listOf(
                StationOptStateFilter.All to allLabel,
                StationOptStateFilter.Operating to operatingLabel,
                StationOptStateFilter.Stopped to stoppedLabel,
            ).forEach { (value, label) ->
                Text(
                    text = label,
                    color = if (selected == value) OpsTheme.colors.primary else Color(0xFF333333),
                    fontWeight = if (selected == value) FontWeight.SemiBold else FontWeight.Normal,
                    modifier = Modifier
                        .fillMaxWidth()
                        .clickable { onSelect(value) }
                        .padding(vertical = 12.dp),
                )
            }
        }
    }
}
