package com.luopingtech.ebike.ops.data.tracking

import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.platform.GeoPoint

interface EmployeeTrackRepository {
    suspend fun upload(userPin: String, points: List<GeoPoint>): OpsResult<Unit>
    fun demoUploadCount(): Int
}

class EmployeeTrackRepositoryImpl(
    private val demoMode: Boolean,
    private val api: EmployeeTrackApi? = null,
) : EmployeeTrackRepository {
    private var demoUploads: Int = 0

    override fun demoUploadCount(): Int = demoUploads

    override suspend fun upload(userPin: String, points: List<GeoPoint>): OpsResult<Unit> {
        if (userPin.isBlank()) {
            return OpsResult.Err(OpsError.business("TRACK_PIN", "userPin empty"))
        }
        if (points.isEmpty()) {
            return OpsResult.Err(OpsError.business("TRACK_EMPTY", "pointList empty"))
        }
        if (demoMode || api == null) {
            demoUploads++
            return OpsResult.Ok(Unit)
        }
        return api.upload(userPin, points)
    }
}
