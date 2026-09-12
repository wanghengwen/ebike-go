package com.luopingtech.ebike.ops.feature.report

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.report.FaultReportRepositoryImpl
import com.luopingtech.ebike.ops.data.vehicle.VehicleRepositoryImpl
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking

class FaultReportFeatureTest {
    private fun feature() = FaultReportFeature(
        repository = FaultReportRepositoryImpl(demoMode = true),
        serviceAreaIdProvider = { "1001" },
        vehicleRepository = VehicleRepositoryImpl(demoMode = true),
    )

    @Test
    fun demo_submitRequiresPhotoAndStopChoice() = runBlocking {
        val f = feature()
        f.openSubmit()
        f.ensureRepairTypes()
        f.setCarId("D1001-001")
        f.lookupVehicleIfReady()
        f.setFixReason("noise")
        f.toggleType("3")
        f.setIzStop(false)
        f.submit()
        assertEquals(
            Strings.t(Str.PhotoUploadFailed, Strings.t(Str.PhotoRequired)),
            f.state.value.errorMessage,
        )

        f.addDemoPhoto()
        f.submit()
        assertEquals(ReportPage.Hub, f.state.value.page)
        assertEquals(Strings.t(Str.FaultReportSuccess, "D1001-001"), f.state.value.message)
    }

    @Test
    fun demo_carIdLookupReloadsTypesByModel() = runBlocking {
        val f = feature()
        f.openSubmit()
        f.setCarId("D1001")
        f.lookupVehicleIfReady()
        assertEquals("", f.state.value.vehicleModel)
        assertTrue(f.state.value.repairTypes.isEmpty())

        f.setCarId("D1001-001")
        f.lookupVehicleIfReady()
        assertEquals("1", f.state.value.vehicleModel)
        assertEquals(5, f.state.value.repairTypes.size)
        assertTrue(f.state.value.repairTypes.none { it.id == "6" })

        f.setCarId("D1001-003")
        f.lookupVehicleIfReady()
        assertEquals("2", f.state.value.vehicleModel)
        assertTrue(f.state.value.repairTypes.any { it.id == "6" })
        assertEquals(Strings.t(Str.FaultModelTypesLoaded, "2"), f.state.value.message)
    }

    @Test
    fun demo_typedCarIdClippedToMax10() = runBlocking {
        val f = feature()
        f.openSubmit()
        f.setCarId("123456789012345")
        assertEquals("1234567890", f.state.value.carId)
    }

    @Test
    fun demo_historyAfterSubmit() = runBlocking {
        val repo = FaultReportRepositoryImpl(demoMode = true)
        val types = (repo.loadRepairTypes() as OpsResult.Ok).value
        repo.submit(
            carId = "D1001-009",
            fixReason = "soft brake",
            types = listOf(types.first()),
            photoUrls = listOf("demo://photo/1"),
            izStop = true,
        )
        val history = repo.myReports("1001")
        assertTrue(history is OpsResult.Ok)
        assertTrue((history as OpsResult.Ok).value.any { it.carId == "D1001-009" })
    }

    @Test
    fun demo_openHistoryLoadsRecords() = runBlocking {
        val f = feature()
        f.openHistory()
        assertEquals(ReportPage.History, f.state.value.page)
        assertTrue(f.state.value.records.isNotEmpty())
    }
}
