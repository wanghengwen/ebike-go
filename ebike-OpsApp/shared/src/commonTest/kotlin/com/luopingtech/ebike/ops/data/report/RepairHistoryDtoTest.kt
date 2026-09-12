package com.luopingtech.ebike.ops.data.report

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import kotlinx.serialization.json.Json

class RepairHistoryDtoTest {
    private val json = Json { ignoreUnknownKeys = true }

    @Test
    fun mapsFixTaskPageItemToFaultReportRecord() {
        val raw = """
            {
              "list": [{
                "id": 9001,
                "carId": "D1001-009",
                "fixReason": "brake noise",
                "nameExtraInfo": ["刹车", "轮毂"],
                "photo": ["https://cdn.example/a.jpg"],
                "izStop": true,
                "createdAt": "2026-09-11 12:00:00",
                "state": 0,
                "checkResult": 0
              }],
              "count": 1
            }
        """.trimIndent()
        val page = json.decodeFromString(RepairHistoryPageDto.serializer(), raw)
        assertEquals(1, page.list.size)
        val record = page.list.first().toDomain()
        assertEquals("9001", record.id)
        assertEquals("D1001-009", record.carId)
        assertEquals("brake noise", record.fixReason)
        assertEquals(listOf("刹车", "轮毂"), record.typeNames)
        assertEquals(listOf("https://cdn.example/a.jpg"), record.photoUrls)
        assertTrue(record.izStop)
        assertEquals("2026-09-11 12:00:00", record.createdAt)
        assertTrue(record.statusLabel.contains(Strings.t(Str.TaskStatePending)))
        assertTrue(record.statusLabel.contains(Strings.t(Str.WorkOrderStopped)))
    }
}
