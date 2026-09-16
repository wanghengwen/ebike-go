package com.luopingtech.ebike.rider.domain.riding

import com.luopingtech.ebike.rider.core.i18n.Str
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertNull

/** 对齐 UniApp `features/bike/ridingFenceTips.ts`。 */
class RidingFenceTipsTest {

    @Test
    fun `no fence code means no banner`() {
        assertNull(RidingFenceTips.of(fenceTypeCode = 0, dispatchCostFen = 0, canReturn = true))
    }

    @Test
    fun `unknown fence code means no banner`() {
        assertNull(RidingFenceTips.of(fenceTypeCode = 7777, dispatchCostFen = 500, canReturn = true))
    }

    @Test
    fun `inside the service area is a plain info banner`() {
        val tip = RidingFenceTips.of(ReturnTypeCodes.NORMAL, dispatchCostFen = 0, canReturn = true)
        assertNotNull(tip)
        assertEquals(Str.FenceTipRideInArea, tip.text)
        assertEquals(FenceTipSeverity.Info, tip.severity)
    }

    @Test
    fun `out of spot with a payable fee spells out the amount`() {
        val tip = RidingFenceTips.of(ReturnTypeCodes.OUT_OF_SPOT, dispatchCostFen = 250, canReturn = true)
        assertNotNull(tip)
        assertEquals(Str.FenceTipOutOfSpotWithFee, tip.text)
        assertEquals("2.50", tip.textArg)
        assertEquals(Str.FenceTipOutOfSpotShortWithFee, tip.shortTip)
        assertEquals(FenceTipSeverity.Warning, tip.severity)
    }

    @Test
    fun `a fee is only quoted when the user can actually return here`() {
        // 不可还的时候报价没有意义，旧版也不报。
        val tip = RidingFenceTips.of(ReturnTypeCodes.OUT_OF_SPOT, dispatchCostFen = 250, canReturn = false)
        assertNotNull(tip)
        assertEquals(Str.FenceTipOutOfSpot, tip.text)
        assertEquals("", tip.textArg)
    }

    @Test
    fun `a zero fee is not quoted either`() {
        val tip = RidingFenceTips.of(ReturnTypeCodes.OUT_OF_SPOT, dispatchCostFen = 0, canReturn = true)
        assertNotNull(tip)
        assertEquals(Str.FenceTipOutOfSpot, tip.text)
    }

    @Test
    fun `rfid shares the out-of-spot banner`() {
        val rfid = RidingFenceTips.of(ReturnTypeCodes.RFID, dispatchCostFen = 100, canReturn = true)
        val spot = RidingFenceTips.of(ReturnTypeCodes.OUT_OF_SPOT, dispatchCostFen = 100, canReturn = true)
        assertEquals(spot, rfid)
    }

    @Test
    fun `no-parking and out-of-service are warnings`() {
        val noParking = RidingFenceTips.of(ReturnTypeCodes.NO_PARKING, 0, canReturn = false)
        assertNotNull(noParking)
        assertEquals(Str.FenceTipNoParking, noParking.text)
        assertEquals(FenceTipSeverity.Warning, noParking.severity)

        val outOfService = RidingFenceTips.of(ReturnTypeCodes.OUT_OF_SERVICE, 0, canReturn = false)
        assertNotNull(outOfService)
        assertEquals(Str.FenceTipOutOfService, outOfService.text)
        assertEquals(FenceTipSeverity.Warning, outOfService.severity)
    }

    @Test
    fun `the helmet variant of out-of-service shows the same banner`() {
        val plain = RidingFenceTips.of(ReturnTypeCodes.OUT_OF_SERVICE, 0, canReturn = false)
        val helmet = RidingFenceTips.of(ReturnTypeCodes.OUT_OF_SERVICE_HELMET, 0, canReturn = false)
        assertEquals(plain, helmet)
    }

}
