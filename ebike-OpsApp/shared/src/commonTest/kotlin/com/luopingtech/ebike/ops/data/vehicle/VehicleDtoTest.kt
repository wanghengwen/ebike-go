package com.luopingtech.ebike.ops.data.vehicle

import com.luopingtech.ebike.ops.core.i18n.OpsI18n
import com.luopingtech.ebike.ops.core.i18n.OpsLanguage
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.domain.scan.ScanTarget
import com.luopingtech.ebike.ops.domain.vehicle.VehicleAlarmStates
import kotlin.test.BeforeTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking
import kotlinx.serialization.json.Json

class VehicleDtoTest {
    private val json = Json { ignoreUnknownKeys = true }

    @BeforeTest
    fun installStrings() {
        Strings.install(OpsI18n.fallback(OpsLanguage.ZH_CN))
    }

    @Test
    fun mapsDetailFields() {
        val dto = json.decodeFromString(
            VehicleDto.serializer(),
            """
            {
              "carId":"D1001-003",
              "imei":"860000000000003",
              "restBattery":17,
              "voltage":42800,
              "batterySn":"BSN-1001-003",
              "batteryLock":1,
              "helmetLock":0,
              "forParkName":"边缘站",
              "noParkName":"禁停样例",
              "isOutofServAera":true,
              "totalMiles":2100,
              "serviceName":"Demo",
              "model":"2",
              "izHaveOverload":true,
              "gsmSignal":-71,
              "isOnline":0,
              "alarmState":[7,3],
              "operationState":[4],
              "lockTime":"1710000000000",
              "unlockTime":1709000000000
            }
            """.trimIndent(),
        )
        val v = dto.toDomain()
        assertEquals("D1001-003", v.carId)
        assertEquals(42_800, v.voltageMv)
        assertEquals("42.8V", v.voltageLabel)
        assertEquals("BSN-1001-003", v.batterySn)
        assertEquals(1, v.batteryLock)
        assertEquals(0, v.helmetLock)
        assertEquals(true, v.isOutOfServiceArea)
        assertEquals(2100.0, v.totalMiles)
        assertEquals(true, v.izHaveOverload)
        assertEquals(-71, v.gsmSignal)
        assertEquals("-71dbm", v.signalLabel)
        assertTrue(v.siteLabel.contains("边缘站"))
        assertTrue(v.siteLabel.contains("超区"))
        assertTrue(v.alarmStates.contains(VehicleAlarmStates.OFFLINE))
        assertEquals(1_710_000_000_000L, v.lockTimeMs)
        assertEquals(1_709_000_000_000L, v.unlockTimeMs)
    }

    @Test
    fun acceptsLegacyIntFlags() {
        val dto = json.decodeFromString(
            VehicleDto.serializer(),
            """
            {
              "carId":"D1001-009",
              "imei":"860000000000009",
              "isOutofServAera":0,
              "isFenceEnable":1,
              "izHaveOverload":0,
              "isOnline":1
            }
            """.trimIndent(),
        )
        val v = dto.toDomain()
        assertEquals(false, v.isOutOfServiceArea)
        assertEquals(true, v.isFenceEnable)
        assertEquals(false, v.izHaveOverload)
        assertEquals(true, v.isOnline)
    }
}

class VehicleRepositoryDetailTest {
    @BeforeTest
    fun installStrings() {
        Strings.install(OpsI18n.fallback(OpsLanguage.ZH_CN))
    }

    @Test
    fun demoDetail_hasVoltageAndBatterySn() = runBlocking {
        val repo = VehicleRepositoryImpl(demoMode = true)
        val vehicle = repo.getDetail("D1001-001").getOrNull()!!
        assertEquals(54_200, vehicle.voltageMv)
        assertEquals("BSN-1001-001", vehicle.batterySn)
        assertEquals("54.2V", vehicle.voltageLabel)

        val out = repo.findByScanTarget(ScanTarget.CarId("D1001-003")).getOrNull()!!
        assertEquals(true, out.isOutOfServiceArea)
        assertTrue(out.forParkName.isNotBlank())
    }
}
