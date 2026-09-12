package com.luopingtech.ebike.ops.feature.tracking

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.data.tracking.EmployeeTrackRepositoryImpl
import com.luopingtech.ebike.ops.domain.tracking.TrackUploadPolicy
import com.luopingtech.ebike.ops.platform.GeoPoint
import com.luopingtech.ebike.ops.platform.SimulatorLocationTracker
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking

class TrackUploadFeatureTest {
    @Test
    fun manualTicks_flushUploadInDemo() = runBlocking {
        val repo = EmployeeTrackRepositoryImpl(demoMode = true)
        val feature = TrackUploadFeature(
            repository = repo,
            locationTracker = SimulatorLocationTracker(intervalMs = 60_000L),
            userPinProvider = { "demo-user" },
            policy = TrackUploadPolicy(flushEveryTicks = 3, minDistanceMeters = 1.0),
        )
        feature.setEnabled(true)
        feature.onLocation(GeoPoint(30.0, 104.0))
        feature.onLocation(GeoPoint(30.001, 104.0))
        feature.onLocation(GeoPoint(30.002, 104.0))
        assertEquals(1, feature.state.value.uploadCount)
        assertEquals(1, repo.demoUploadCount())
        assertEquals(Strings.t(Str.TrackPointsUploaded, 3), feature.state.value.lastMessage)
        feature.setEnabled(false)
    }

    @Test
    fun rejectsEnableWithoutPin() {
        val feature = TrackUploadFeature(
            repository = EmployeeTrackRepositoryImpl(demoMode = true),
            locationTracker = SimulatorLocationTracker(),
            userPinProvider = { "" },
        )
        feature.setEnabled(true)
        assertEquals(false, feature.state.value.enabled)
        assertEquals(Strings.t(Str.TrackNeedLogin), feature.state.value.errorMessage)
    }
}
