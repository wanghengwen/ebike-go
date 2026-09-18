package com.luopingtech.ebike.ops.feature.fence

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.fence.FenceRepository
import com.luopingtech.ebike.ops.domain.model.FenceKind
import com.luopingtech.ebike.ops.domain.model.FenceNoParkingCreate
import com.luopingtech.ebike.ops.domain.model.FenceNoParkingUpdate
import com.luopingtech.ebike.ops.domain.model.FenceParkingCreate
import com.luopingtech.ebike.ops.domain.model.FenceParkingUpdate
import com.luopingtech.ebike.ops.domain.model.FencePolygon
import com.luopingtech.ebike.ops.domain.model.GeoLatLng
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.platform.ReverseGeocoder
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

/**
 * 对齐 FenceEditActivity 选点新建 / 编辑。
 * - 停车新建：先画，首次保存再填站点参数（对齐原版）
 * - 贴片：屏幕矩形四角 → fromScreenLocation
 * - 选点：准星屏幕坐标 → fromScreenLocation；加点后按方位角重排
 */
enum class FenceDrawMode { Patch, Point }

data class FenceEditUiState(
    val active: Boolean = false,
    val editingId: String? = null,
    val kind: FenceKind = FenceKind.Parking,
    val drawMode: FenceDrawMode = FenceDrawMode.Patch,
    val draftPoints: List<GeoLatLng> = emptyList(),
    /** 切 Tab 时各自保留（对齐原版两套 clip 层）。 */
    val patchDraftPoints: List<GeoLatLng> = emptyList(),
    val pointDraftPoints: List<GeoLatLng> = emptyList(),
    val name: String = "",
    val maxParkingNumber: String = "50",
    val izEnable: Boolean = true,
    val bufferDistance: String = "10",
    val direction: String = "-1",
    val tbeacon: Boolean = false,
    val rfid: Boolean = false,
    val directional: Boolean = false,
    val kickstand: Boolean = false,
    val camera: Boolean = false,
    val izFullPileNoStop: Boolean = false,
    val openingHoursBegin: String = "00:00:00",
    val openingHoursEnd: String = "23:59:00",
    /**
     * 停车参数是否已确认可提交。
     * 新建：首次保存前弹窗确认；编辑：进入即就绪。
     */
    val parkingParamsConfirmed: Boolean = false,
    val saving: Boolean = false,
    val message: String? = null,
    val errorMessage: String? = null,
) {
    val isEdit: Boolean get() = !editingId.isNullOrBlank()
}

class FenceEditFeature(
    private val repository: FenceRepository,
    private val reverseGeocoder: ReverseGeocoder,
) {
    private val _state = MutableStateFlow(FenceEditUiState())
    val state: StateFlow<FenceEditUiState> = _state.asStateFlow()

    /** 最近一次保存成功的围栏中心，供浏览页相机跟随。 */
    var lastSavedCenter: GeoLatLng? = null
        private set

    fun clear() {
        _state.value = FenceEditUiState()
    }

    fun start(kind: FenceKind) {
        _state.value = FenceEditUiState(
            active = true,
            kind = kind,
            // 新建默认贴片；画布立刻可用（参数在首次保存时填）
            drawMode = FenceDrawMode.Patch,
            parkingParamsConfirmed = kind != FenceKind.Parking,
        )
    }

    fun startEdit(fence: FencePolygon) {
        if (fence.kind != FenceKind.Parking && fence.kind != FenceKind.NoParking) return
        if (fence.id.isBlank() || fence.points.size < 3) return
        val patchLike = isPatchMode(fence.points)
        _state.value = FenceEditUiState(
            active = true,
            editingId = fence.id,
            kind = fence.kind,
            drawMode = if (patchLike) FenceDrawMode.Patch else FenceDrawMode.Point,
            draftPoints = fence.points,
            name = fence.name,
            maxParkingNumber = fence.maxParkingNumber.takeIf { it > 0 }?.toString() ?: "50",
            izEnable = fence.izEnable != false,
            parkingParamsConfirmed = true,
        )
    }

    /**
     * 对齐 FenceManager.isPatchMode：
     * 非 4 点 → 选点；4 点且对边长度相等 → 贴片。
     * 兼容首尾闭合多一点；放宽对边容差，避免保存后再编辑被误判成选点（从而冒出 Pick Point）。
     */
    fun isPatchMode(points: List<GeoLatLng>): Boolean {
        val ring = normalizeRing(points)
        if (ring.size != 4) return false
        val d01 = distanceMeters(ring[0], ring[1])
        val d12 = distanceMeters(ring[1], ring[2])
        val d23 = distanceMeters(ring[2], ring[3])
        val d30 = distanceMeters(ring[3], ring[0])
        // 相对 2% 或至少 1.5m，覆盖投影/序列化误差
        fun nearly(a: Double, b: Double): Boolean {
            val tol = maxOf(1.5, maxOf(a, b) * 0.02)
            return kotlin.math.abs(a - b) <= tol
        }
        return nearly(d01, d23) && nearly(d12, d30)
    }

    /** 去掉闭合环重复的末点。 */
    private fun normalizeRing(points: List<GeoLatLng>): List<GeoLatLng> {
        if (points.size < 2) return points
        val first = points.first()
        val last = points.last()
        return if (kotlin.math.abs(first.lat - last.lat) < 1e-7 &&
            kotlin.math.abs(first.lng - last.lng) < 1e-7
        ) {
            points.dropLast(1)
        } else {
            points
        }
    }

    private fun distanceMeters(a: GeoLatLng, b: GeoLatLng): Double {
        val r = 6371000.0
        val p1 = Math.toRadians(a.lat)
        val p2 = Math.toRadians(b.lat)
        val dLat = Math.toRadians(b.lat - a.lat)
        val dLng = Math.toRadians(b.lng - a.lng)
        val h = kotlin.math.sin(dLat / 2) * kotlin.math.sin(dLat / 2) +
            kotlin.math.cos(p1) * kotlin.math.cos(p2) *
            kotlin.math.sin(dLng / 2) * kotlin.math.sin(dLng / 2)
        return 2 * r * kotlin.math.asin(kotlin.math.sqrt(h))
    }

    fun setDrawMode(mode: FenceDrawMode) {
        if (_state.value.isEdit) return
        if (_state.value.drawMode == mode) return
        val s = _state.value
        // 把当前草稿存回对应槽，再切到另一槽
        val saved = when (s.drawMode) {
            FenceDrawMode.Patch -> s.copy(patchDraftPoints = s.draftPoints)
            FenceDrawMode.Point -> s.copy(pointDraftPoints = s.draftPoints)
        }
        val restored = when (mode) {
            FenceDrawMode.Patch -> saved.patchDraftPoints
            FenceDrawMode.Point -> saved.pointDraftPoints
        }
        _state.value = saved.copy(
            drawMode = mode,
            draftPoints = restored,
            errorMessage = null,
            message = null,
        )
    }

    fun replacePoints(points: List<GeoLatLng>) {
        val s = _state.value
        _state.value = when (s.drawMode) {
            FenceDrawMode.Patch -> s.copy(
                draftPoints = points,
                patchDraftPoints = points,
                errorMessage = null,
                message = null,
            )
            FenceDrawMode.Point -> s.copy(
                draftPoints = points,
                pointDraftPoints = points,
                errorMessage = null,
                message = null,
            )
        }
    }

    fun movePoint(index: Int, lat: Double, lng: Double) {
        val pts = _state.value.draftPoints.toMutableList()
        if (index !in pts.indices) return
        pts[index] = GeoLatLng(lat, lng)
        replacePoints(pts)
    }

    fun confirmParkingParams(): Boolean {
        val s = _state.value
        if (s.kind != FenceKind.Parking) return true
        val cap = s.maxParkingNumber.toIntOrNull() ?: 0
        if (cap <= 0) {
            _state.value = s.copy(errorMessage = Strings.t(Str.FenceNeedCapacity))
            return false
        }
        val buf = s.bufferDistance.toIntOrNull()
        if (buf == null || buf !in 0..100) {
            _state.value = s.copy(errorMessage = Strings.t(Str.FenceNeedBuffer))
            return false
        }
        _state.value = s.copy(parkingParamsConfirmed = true, errorMessage = null)
        return true
    }

    fun addPoint(lat: Double, lng: Double) {
        if (!_state.value.active || _state.value.saving) return
        val next = orderlyPointList(_state.value.draftPoints + GeoLatLng(lat, lng))
        replacePoints(next)
    }

    /** 对齐 orderlyPointList：按相对中心方位角排序。 */
    fun orderlyPointList(list: List<GeoLatLng>): List<GeoLatLng> {
        if (list.size <= 2) return list
        val cx = list.map { it.lat }.average()
        val cy = list.map { it.lng }.average()
        return list.sortedBy { p ->
            kotlin.math.atan2(p.lng - cy, p.lat - cx)
        }
    }

    fun undoPoint() {
        val pts = _state.value.draftPoints
        if (pts.isEmpty()) return
        replacePoints(pts.dropLast(1))
    }

    fun setName(value: String) {
        _state.value = _state.value.copy(name = value)
    }

    fun setMaxParkingNumber(value: String) {
        _state.value = _state.value.copy(maxParkingNumber = value.filter { it.isDigit() }.take(5))
    }

    fun setIzEnable(value: Boolean) {
        _state.value = _state.value.copy(izEnable = value)
    }

    fun setBufferDistance(value: String) {
        _state.value = _state.value.copy(bufferDistance = value.filter { it.isDigit() }.take(3))
    }

    fun setDirection(value: String) {
        // 允许 -1（未设置）与 0–360
        val filtered = buildString {
            value.forEachIndexed { i, ch ->
                if (ch == '-' && i == 0 && isEmpty()) append(ch)
                else if (ch.isDigit()) append(ch)
            }
        }.take(4)
        _state.value = _state.value.copy(direction = filtered.ifBlank { "-1" })
    }

    fun setTbeacon(value: Boolean) {
        _state.value = _state.value.copy(tbeacon = value)
    }

    fun setRfid(value: Boolean) {
        _state.value = _state.value.copy(rfid = value)
    }

    fun setDirectional(value: Boolean) {
        _state.value = _state.value.copy(directional = value)
    }

    fun setKickstand(value: Boolean) {
        _state.value = _state.value.copy(kickstand = value)
    }

    fun setCamera(value: Boolean) {
        _state.value = _state.value.copy(camera = value)
    }

    fun setIzFullPileNoStop(value: Boolean) {
        _state.value = _state.value.copy(izFullPileNoStop = value)
    }

    fun setOpeningHoursBegin(value: String) {
        _state.value = _state.value.copy(openingHoursBegin = value)
    }

    fun setOpeningHoursEnd(value: String) {
        _state.value = _state.value.copy(openingHoursEnd = value)
    }

    suspend fun prepareNameFromGeocode() {
        val pts = _state.value.draftPoints
        if (pts.isEmpty() || _state.value.name.isNotBlank()) return
        val c = centroid(pts) ?: return
        when (val r = reverseGeocoder.addressOf(c.lat, c.lng)) {
            is OpsResult.Ok -> {
                val addr = r.value.trim()
                if (addr.isNotBlank()) {
                    _state.value = _state.value.copy(name = addr)
                }
            }
            is OpsResult.Err -> Unit
        }
    }

    suspend fun save(area: ServiceArea?): OpsResult<Unit> {
        if (area == null) {
            val msg = Strings.t(Str.SelectServiceAreaFirst)
            _state.value = _state.value.copy(errorMessage = msg)
            return OpsResult.Err(com.luopingtech.ebike.ops.core.result.OpsError.business("FENCE", msg))
        }
        val pts = _state.value.draftPoints
        if (pts.size < 3) {
            val msg = Strings.t(Str.FenceNeedThreePoints)
            _state.value = _state.value.copy(errorMessage = msg)
            return OpsResult.Err(com.luopingtech.ebike.ops.core.result.OpsError.business("FENCE", msg))
        }
        var name = _state.value.name.trim()
        if (name.isBlank()) {
            prepareNameFromGeocode()
            name = _state.value.name.trim().ifBlank { Strings.t(Str.FenceUnknownPlace) }
            _state.value = _state.value.copy(name = name)
        }
        val center = centroid(pts)!!
        val editingId = _state.value.editingId
        val isEdit = !editingId.isNullOrBlank()
        val kind = _state.value.kind
        val s = _state.value
        _state.value = s.copy(saving = true, errorMessage = null, message = null)
        val result = when (kind) {
            FenceKind.NoParking -> {
                if (isEdit && editingId != null) {
                    repository.updateNoParking(
                        FenceNoParkingUpdate(
                            id = editingId,
                            serviceId = area.id,
                            name = name,
                            points = pts,
                            centerLat = center.lat,
                            centerLng = center.lng,
                        ),
                    )
                } else {
                    repository.createNoParking(
                        FenceNoParkingCreate(
                            serviceId = area.id,
                            name = name,
                            points = pts,
                            centerLat = center.lat,
                            centerLng = center.lng,
                        ),
                    )
                }
            }
            else -> {
                val maxNum = s.maxParkingNumber.toIntOrNull() ?: 0
                if (maxNum <= 0) {
                    val msg = Strings.t(Str.FenceNeedCapacity)
                    _state.value = s.copy(saving = false, errorMessage = msg)
                    return OpsResult.Err(com.luopingtech.ebike.ops.core.result.OpsError.business("FENCE", msg))
                }
                val buf = s.bufferDistance.toIntOrNull() ?: 10
                val dir = s.direction.toIntOrNull() ?: -1
                if (isEdit && editingId != null) {
                    repository.updateParking(
                        FenceParkingUpdate(
                            id = editingId,
                            serviceId = area.id,
                            name = name,
                            points = pts,
                            centerLat = center.lat,
                            centerLng = center.lng,
                            maxParkingNumber = maxNum,
                            izEnable = s.izEnable,
                            bufferDistance = buf,
                            direction = dir,
                            tbeacon = s.tbeacon,
                            rfid = s.rfid,
                            directional = s.directional,
                            kickstand = s.kickstand,
                            camera = s.camera,
                            izFullPileNoStop = s.izFullPileNoStop,
                            openingHoursBegin = s.openingHoursBegin,
                            openingHoursEnd = s.openingHoursEnd,
                        ),
                    )
                } else {
                    repository.createParking(
                        FenceParkingCreate(
                            serviceId = area.id,
                            name = name,
                            points = pts,
                            centerLat = center.lat,
                            centerLng = center.lng,
                            maxParkingNumber = maxNum,
                            izEnable = s.izEnable,
                            bufferDistance = buf,
                            direction = dir,
                            tbeacon = s.tbeacon,
                            rfid = s.rfid,
                            directional = s.directional,
                            kickstand = s.kickstand,
                            camera = s.camera,
                            izFullPileNoStop = s.izFullPileNoStop,
                            openingHoursBegin = s.openingHoursBegin,
                            openingHoursEnd = s.openingHoursEnd,
                        ),
                    )
                }
            }
        }
        return when (result) {
            is OpsResult.Ok -> {
                lastSavedCenter = center
                if (isEdit) {
                    _state.value = FenceEditUiState(message = Strings.t(Str.FenceUpdateOk))
                } else {
                    // 对齐原版：新建成功留在画布，可继续画下一个；参数沿用（再画可改）
                    _state.value = FenceEditUiState(
                        active = true,
                        kind = kind,
                        drawMode = FenceDrawMode.Patch,
                        parkingParamsConfirmed = kind != FenceKind.Parking,
                        maxParkingNumber = s.maxParkingNumber,
                        izEnable = s.izEnable,
                        bufferDistance = s.bufferDistance,
                        direction = s.direction,
                        tbeacon = s.tbeacon,
                        rfid = s.rfid,
                        directional = s.directional,
                        kickstand = s.kickstand,
                        camera = s.camera,
                        izFullPileNoStop = s.izFullPileNoStop,
                        openingHoursBegin = s.openingHoursBegin,
                        openingHoursEnd = s.openingHoursEnd,
                        message = Strings.t(Str.FenceCreateOk),
                    )
                }
                result
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(saving = false, errorMessage = result.error.message)
                result
            }
        }
    }

    private fun centroid(points: List<GeoLatLng>): GeoLatLng? {
        if (points.isEmpty()) return null
        return GeoLatLng(
            lat = points.map { it.lat }.average(),
            lng = points.map { it.lng }.average(),
        )
    }
}
