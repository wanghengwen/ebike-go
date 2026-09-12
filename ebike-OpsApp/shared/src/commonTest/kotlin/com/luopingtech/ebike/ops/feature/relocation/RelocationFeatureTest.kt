package com.luopingtech.ebike.ops.feature.relocation

import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.relocation.RelocationRepository
import com.luopingtech.ebike.ops.domain.model.RelocationDevice
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.scan.ScanTarget
import com.luopingtech.ebike.ops.platform.GeoPoint
import com.luopingtech.ebike.ops.platform.LocationTracker
import com.luopingtech.ebike.ops.platform.SimulatorLocationTracker
import kotlinx.coroutines.runBlocking
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class RelocationFeatureTest {
    private val area = ServiceArea(id = "1001", name = "Demo")

    @Test
    fun rejectsOnlineVehicle() = runBlocking {
        val feature = RelocationFeature(
            repository = object : RelocationRepository {
                override suspend fun scanDevice(serviceId: String, target: ScanTarget) =
                    OpsResult.Ok(
                        RelocationDevice(
                            carId = "D1",
                            imei = "8601",
                            isOnline = 1,
                        ),
                    )

                override suspend fun reportLocation(
                    imeiList: List<String>,
                    latitude: Double,
                    longitude: Double,
                ) = OpsResult.Ok(Unit)
            },
            locationTracker = SimulatorLocationTracker(),
        )
        assertTrue(feature.addByRaw("D1", area).isErr)
    }

    @Test
    fun confirmsAndRemovesSelected() = runBlocking {
        val reported = mutableListOf<List<String>>()
        val feature = RelocationFeature(
            repository = object : RelocationRepository {
                override suspend fun scanDevice(serviceId: String, target: ScanTarget) =
                    OpsResult.Ok(
                        RelocationDevice(carId = "D2", imei = "8602", isOnline = 0),
                    )

                override suspend fun reportLocation(
                    imeiList: List<String>,
                    latitude: Double,
                    longitude: Double,
                ): OpsResult<Unit> {
                    reported += imeiList
                    return OpsResult.Ok(Unit)
                }
            },
            locationTracker = object : LocationTracker {
                override suspend fun currentLocation(): OpsResult<GeoPoint> =
                    OpsResult.Ok(GeoPoint(28.1, 112.9))

                override fun track() = kotlinx.coroutines.flow.flowOf<GeoPoint>()
            },
        )
        assertTrue(feature.addByRaw("D2", area).isOk)
        assertTrue(feature.confirmLocation().isOk)
        assertEquals(listOf(listOf("8602")), reported)
        assertTrue(feature.state.value.devices.isEmpty())
    }
}
