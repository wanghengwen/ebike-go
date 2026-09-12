package com.luopingtech.ebike.ops.feature.tracking

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.tracking.EmployeeTrackRepository
import com.luopingtech.ebike.ops.domain.tracking.TrackBufferEvent
import com.luopingtech.ebike.ops.domain.tracking.TrackPointBuffer
import com.luopingtech.ebike.ops.domain.tracking.TrackUploadPolicy
import com.luopingtech.ebike.ops.platform.GeoPoint
import com.luopingtech.ebike.ops.platform.LocationTracker
import com.luopingtech.ebike.ops.platform.SecureStore
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.catch
import kotlinx.coroutines.launch

data class TrackUploadUiState(
    val enabled: Boolean = false,
    val collecting: Boolean = false,
    val lastFix: GeoPoint? = null,
    val bufferedPoints: Int = 0,
    val tickCount: Int = 0,
    val uploadCount: Int = 0,
    val lastMessage: String? = null,
    val errorMessage: String? = null,
)

/**
 * Collects [LocationTracker.track], buffers like legacy MainViewModel, uploads batches.
 * Host should request location permission before [setEnabled] true on real devices.
 */
class TrackUploadFeature(
    private val repository: EmployeeTrackRepository,
    private val locationTracker: LocationTracker,
    private val userPinProvider: () -> String,
    private val policy: TrackUploadPolicy = TrackUploadPolicy(),
    private val secureStore: SecureStore? = null,
) {
    private val buffer = TrackPointBuffer(policy)
    private val _state = MutableStateFlow(TrackUploadUiState())
    val state: StateFlow<TrackUploadUiState> = _state.asStateFlow()

    private var scope: CoroutineScope? = null
    private var collectJob: Job? = null

    fun bind(scope: CoroutineScope) {
        this.scope = scope
    }

    /** Whether user last left track upload on (survives process death). */
    fun wasEnabledPersisted(): Boolean =
        secureStore?.getString(SecureStore.KEY_TRACK_UPLOAD_ENABLED) == "1"

    fun setEnabled(enabled: Boolean) {
        if (!enabled) {
            stopCollecting()
            buffer.reset()
            secureStore?.putString(SecureStore.KEY_TRACK_UPLOAD_ENABLED, "0")
            _state.value = _state.value.copy(
                enabled = false,
                collecting = false,
                bufferedPoints = 0,
                tickCount = 0,
                lastMessage = Strings.t(Str.TrackDisabled),
                errorMessage = null,
            )
            return
        }
        val pin = userPinProvider()
        if (pin.isBlank()) {
            _state.value = _state.value.copy(
                enabled = false,
                errorMessage = Strings.t(Str.TrackNeedLogin),
            )
            return
        }
        secureStore?.putString(SecureStore.KEY_TRACK_UPLOAD_ENABLED, "1")
        _state.value = _state.value.copy(
            enabled = true,
            errorMessage = null,
            lastMessage = Strings.t(Str.TrackEnabled),
        )
        startCollecting()
    }

    fun toggle() {
        setEnabled(!_state.value.enabled)
    }

    /** Test / manual injection — same path as live tracker ticks. */
    suspend fun onLocation(point: GeoPoint) {
        _state.value = _state.value.copy(lastFix = point, errorMessage = null)
        when (val event = buffer.onTick(point)) {
            is TrackBufferEvent.None -> {
                _state.value = _state.value.copy(
                    bufferedPoints = buffer.size,
                    tickCount = buffer.ticks,
                )
            }
            is TrackBufferEvent.Flush -> flush(event.points)
        }
    }

    private fun startCollecting() {
        val activeScope = scope ?: return
        if (collectJob?.isActive == true) return
        locationTracker.startTracking()
        collectJob = activeScope.launch {
            _state.value = _state.value.copy(collecting = true)
            locationTracker.track()
                .catch { e ->
                    _state.value = _state.value.copy(
                        errorMessage = e.message ?: "location stream failed",
                        collecting = false,
                    )
                }
                .collect { point -> onLocation(point) }
        }
    }

    private fun stopCollecting() {
        collectJob?.cancel()
        collectJob = null
        locationTracker.stopTracking()
        _state.value = _state.value.copy(collecting = false)
    }

    private suspend fun flush(points: List<GeoPoint>) {
        val pin = userPinProvider()
        if (pin.isBlank() || points.isEmpty()) {
            buffer.clearAfterSuccess()
            _state.value = _state.value.copy(
                bufferedPoints = 0,
                tickCount = 0,
            )
            return
        }
        when (val result = repository.upload(pin, points)) {
            is OpsResult.Ok -> {
                buffer.clearAfterSuccess()
                _state.value = _state.value.copy(
                    bufferedPoints = 0,
                    tickCount = 0,
                    uploadCount = _state.value.uploadCount + 1,
                    lastMessage = Strings.t(Str.TrackPointsUploaded, points.size),
                    errorMessage = null,
                )
            }
            is OpsResult.Err -> {
                // Keep buffer + tickCount so next tick retries (legacy behavior).
                _state.value = _state.value.copy(
                    bufferedPoints = buffer.size,
                    tickCount = buffer.ticks,
                    errorMessage = result.error.message,
                )
            }
        }
    }
}
