package com.luopingtech.ebike.ops.feature.warehouse

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.warehouse.WarehouseRepositoryImpl
import com.luopingtech.ebike.ops.domain.model.WarehouseOperationType
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking

class WarehouseFeatureTest {
    @Test
    fun demo_scanThenOut() = runBlocking {
        val feature = WarehouseFeature(
            repository = WarehouseRepositoryImpl(demoMode = true),
            pinProvider = { "demo-user" },
        )
        feature.openOperate(WarehouseOperationType.Out, withCode = true)
        feature.resolveCode("BAT-9001")
        assertEquals(1, feature.state.value.scanned.size)
        feature.submitOperate()
        assertEquals(WarehousePage.Hub, feature.state.value.page)
        assertEquals(
            Strings.t(Str.OpSuccess, WarehouseOperationType.Out.label),
            feature.state.value.message,
        )
    }

    @Test
    fun demo_withoutCodeIn_andRecords() = runBlocking {
        val feature = WarehouseFeature(
            repository = WarehouseRepositoryImpl(demoMode = true),
            pinProvider = { "demo-user" },
        )
        feature.openOperate(WarehouseOperationType.In, withCode = false)
        feature.ensureComponentNames(existCode = 0)
        feature.setSelectedComponentName(Strings.t(Str.RepairTypeBattery))
        feature.setQuantity(3)
        feature.submitOperate()
        assertEquals(
            Strings.t(Str.OpSuccess, WarehouseOperationType.In.label),
            feature.state.value.message,
        )

        feature.openRecords()
        assertTrue(feature.state.value.records.isNotEmpty())
        val id = feature.state.value.records.first().id
        feature.openDetail(id)
        assertTrue(feature.state.value.details.isNotEmpty())
    }

    @Test
    fun demo_queryQuantity() = runBlocking {
        val repo = WarehouseRepositoryImpl(demoMode = true)
        val result = repo.queryQuantity(Strings.t(Str.RepairTypeBattery))
        assertTrue(result is OpsResult.Ok)
        assertTrue((result as OpsResult.Ok).value.stockQuantity > 0)
    }
}
