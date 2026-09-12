package com.luopingtech.ebike.ops.data.tracking

import com.luopingtech.ebike.ops.core.network.CommonRequestBody
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.platform.DeviceInfo
import com.luopingtech.ebike.ops.platform.GeoPoint
import kotlinx.serialization.json.add
import kotlinx.serialization.json.buildJsonArray
import kotlinx.serialization.json.put

class EmployeeTrackApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) {
    suspend fun upload(
        userPin: String,
        points: List<GeoPoint>,
    ): OpsResult<Unit> {
        val last = points.lastOrNull()
        val body = CommonRequestBody.toJsonString(
            source = "/business/ebike-management/employeeTrack/add",
            tenantId = tenantIdProvider(),
            deviceInfo = deviceInfo,
            deviceId = deviceIdProvider(),
        ) {
            put("userPin", userPin)
            put(
                "pointList",
                buildJsonArray {
                    points.forEach { point ->
                        add(
                            buildJsonArray {
                                add(point.latitude)
                                add(point.longitude)
                            },
                        )
                    }
                },
            )
            if (last != null) {
                put("latitude", last.latitude.toString())
                put("longitude", last.longitude.toString())
            }
        }
        return signedApi.postUnit("business/ebike-management/employeeTrack/add", body)
    }
}
