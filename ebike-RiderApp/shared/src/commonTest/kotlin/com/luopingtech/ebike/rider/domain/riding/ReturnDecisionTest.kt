package com.luopingtech.ebike.rider.domain.riding

import com.luopingtech.ebike.rider.core.i18n.Str
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNull
import kotlin.test.assertTrue

/** 对齐 UniApp `features/bike/returnTypes.ts` 的判定矩阵。 */
class ReturnDecisionTest {

    @Test
    fun `only the 11xx band counts as a normal return`() {
        assertEquals(ReturnKind.Normal, ReturnDecision.kind(ReturnTypeCodes.NORMAL))
        assertEquals(ReturnKind.Normal, ReturnDecision.kind(ReturnTypeCodes.OUT_OF_PARK))
        assertEquals(ReturnKind.Normal, ReturnDecision.kind(1100))
        assertEquals(ReturnKind.Normal, ReturnDecision.kind(1199))

        assertEquals(ReturnKind.Unnormal, ReturnDecision.kind(1099))
        assertEquals(ReturnKind.Unnormal, ReturnDecision.kind(1200))
        assertEquals(ReturnKind.Unnormal, ReturnDecision.kind(ReturnTypeCodes.OUT_OF_SPOT))
        assertEquals(ReturnKind.Unnormal, ReturnDecision.kind(ReturnTypeCodes.NO_PARKING))
    }

    @Test
    fun `the 11xx check does not fall for string prefix tricks`() {
        // 旧版用 String(type).startsWith('11')，11、110、11000 都会误判成正常还车。
        assertEquals(ReturnKind.Unnormal, ReturnDecision.kind(11))
        assertEquals(ReturnKind.Unnormal, ReturnDecision.kind(110))
        assertEquals(ReturnKind.Unnormal, ReturnDecision.kind(11000))
        assertEquals(ReturnKind.Unnormal, ReturnDecision.kind(-1101))
        assertEquals(ReturnKind.Unnormal, ReturnDecision.kind(0))
    }

    @Test
    fun `normal return skips the reminder when the tenant disabled it`() {
        val flow = ReturnDecision.flow(
            canReturn = true,
            returnTypeCode = ReturnTypeCodes.NORMAL,
            civilizationRemind = false,
        )
        assertEquals(ReturnFlow.Normal, flow)
    }

    @Test
    fun `normal return shows the reminder when the tenant enabled it`() {
        val flow = ReturnDecision.flow(
            canReturn = true,
            returnTypeCode = ReturnTypeCodes.NORMAL,
            civilizationRemind = true,
        )
        assertEquals(ReturnFlow.Civilization, flow)
    }

    @Test
    fun `returnable but out of spot asks the user to accept the fee`() {
        val flow = ReturnDecision.flow(
            canReturn = true,
            returnTypeCode = ReturnTypeCodes.OUT_OF_SPOT,
            civilizationRemind = false,
        )
        assertEquals(ReturnFlow.PenaltySheet, flow)
    }

    @Test
    fun `the civilization switch does not leak into the penalty branch`() {
        // 认罚弹层本身就在解释代价，再叠一层文明提醒是旧版没有的行为。
        val flow = ReturnDecision.flow(
            canReturn = true,
            returnTypeCode = ReturnTypeCodes.OUT_OF_SPOT,
            civilizationRemind = true,
        )
        assertEquals(ReturnFlow.PenaltySheet, flow)
    }

    @Test
    fun `codes with a custom guide page route to the guide`() {
        val guided = listOf(
            ReturnTypeCodes.DIRECTIONAL,
            ReturnTypeCodes.RFID,
            ReturnTypeCodes.KICKSTAND,
            ReturnTypeCodes.CAMERA_MISS,
            ReturnTypeCodes.HELMET_MISS_A,
            ReturnTypeCodes.HELMET_MISS_B,
            ReturnTypeCodes.NO_PARKING_HELMET,
            ReturnTypeCodes.OUT_OF_SERVICE_HELMET,
        )
        for (code in guided) {
            assertEquals(
                ReturnFlow.Guide,
                ReturnDecision.flow(false, code, civilizationRemind = false),
                "code=$code",
            )
        }
    }

    @Test
    fun `blocking codes explain themselves in a sheet`() {
        val blocked = listOf(
            ReturnTypeCodes.OUT_OF_SPOT,
            ReturnTypeCodes.NO_PARKING,
            ReturnTypeCodes.OUT_OF_SERVICE,
            ReturnTypeCodes.FULL_PILE,
        )
        for (code in blocked) {
            assertEquals(
                ReturnFlow.BlockSheet,
                ReturnDecision.flow(false, code, civilizationRemind = false),
                "code=$code",
            )
        }
    }

    @Test
    fun `guide wins over the block sheet when a code qualifies for both`() {
        // RFID 既在引导表里也不可还；旧版先跳引导页。
        assertEquals(
            ReturnFlow.Guide,
            ReturnDecision.flow(false, ReturnTypeCodes.RFID, civilizationRemind = false),
        )
    }

    @Test
    fun `unknown blocking codes stay silent like the legacy app`() {
        assertEquals(
            ReturnFlow.Silent,
            ReturnDecision.flow(false, 9999, civilizationRemind = false),
        )
        assertEquals(
            ReturnFlow.Silent,
            ReturnDecision.flow(false, ReturnTypeCodes.FORBID_ZONE, civilizationRemind = false),
        )
    }

    @Test
    fun `cannot-return in the normal band still avoids a silent submit`() {
        // canReturn=false + 1101 是后端自相矛盾的响应；不能落进 Normal 直接提交。
        val flow = ReturnDecision.flow(false, ReturnTypeCodes.NORMAL, civilizationRemind = false)
        assertTrue(flow != ReturnFlow.Normal)
        assertEquals(ReturnFlow.Silent, flow)
    }

    @Test
    fun `guide page types match the legacy customizedReturn values`() {
        assertEquals(4, ReturnDecision.guidePageType(ReturnTypeCodes.DIRECTIONAL))
        assertEquals(5, ReturnDecision.guidePageType(ReturnTypeCodes.RFID))
        assertEquals(6, ReturnDecision.guidePageType(ReturnTypeCodes.HELMET_MISS_A))
        assertEquals(6, ReturnDecision.guidePageType(ReturnTypeCodes.HELMET_MISS_B))
        assertEquals(6, ReturnDecision.guidePageType(ReturnTypeCodes.NO_PARKING_HELMET))
        assertEquals(6, ReturnDecision.guidePageType(ReturnTypeCodes.OUT_OF_SERVICE_HELMET))
        assertEquals(8, ReturnDecision.guidePageType(ReturnTypeCodes.KICKSTAND))
        assertEquals(11, ReturnDecision.guidePageType(ReturnTypeCodes.CAMERA_MISS))

        assertNull(ReturnDecision.guidePageType(ReturnTypeCodes.NORMAL))
        assertNull(ReturnDecision.guidePageType(ReturnTypeCodes.OUT_OF_SPOT))
        assertNull(ReturnDecision.guidePageType(0))
    }

    @Test
    fun `apply type buckets follow the code band`() {
        assertEquals(1, ReturnDecision.applyType(ReturnTypeCodes.OUT_OF_SPOT))
        assertEquals(1, ReturnDecision.applyType(ReturnTypeCodes.CAMERA_MISS))
        assertEquals(2, ReturnDecision.applyType(ReturnTypeCodes.NO_PARKING))
        assertEquals(2, ReturnDecision.applyType(ReturnTypeCodes.NO_PARKING_HELMET))
        assertEquals(3, ReturnDecision.applyType(ReturnTypeCodes.NORMAL))
        assertEquals(3, ReturnDecision.applyType(ReturnTypeCodes.OUT_OF_SERVICE))
        assertEquals(3, ReturnDecision.applyType(ReturnTypeCodes.FULL_PILE))
    }

    @Test
    fun `camera guide maps to apply type 7`() {
        assertEquals(7, ReturnDecision.guideApplyType(ReturnGuidePageType.CAMERA))
        assertEquals(4, ReturnDecision.guideApplyType(ReturnGuidePageType.DIRECTIONAL))
        assertEquals(5, ReturnDecision.guideApplyType(ReturnGuidePageType.RFID))
        assertEquals(6, ReturnDecision.guideApplyType(ReturnGuidePageType.HELMET))
        assertEquals(8, ReturnDecision.guideApplyType(ReturnGuidePageType.KICKSTAND))
    }

    @Test
    fun `camera codes get their own distinct reasons`() {
        // 三个摄像头码在旧版里文案各不相同，最容易在移植时被合并掉。
        assertEquals(Str.ReturnReasonCameraDirection, ReturnDecision.meta(ReturnTypeCodes.CAMERA_DIRECTION).reason)
        assertEquals(Str.ReturnReasonCameraPoint, ReturnDecision.meta(ReturnTypeCodes.CAMERA_POINT).reason)
        assertEquals(Str.ReturnReasonCameraMiss, ReturnDecision.meta(ReturnTypeCodes.CAMERA_MISS).reason)
    }

    @Test
    fun `helmet codes all point at the helmet reason`() {
        val helmetCodes = listOf(
            ReturnTypeCodes.HELMET_MISS_A,
            ReturnTypeCodes.HELMET_MISS_B,
            ReturnTypeCodes.NO_PARKING_HELMET,
            ReturnTypeCodes.OUT_OF_SERVICE_HELMET,
        )
        for (code in helmetCodes) {
            assertEquals(Str.ReturnReasonHelmetMiss, ReturnDecision.meta(code).reason, "code=$code")
        }
    }

    @Test
    fun `out of service adds the ride-back tip`() {
        val meta = ReturnDecision.meta(ReturnTypeCodes.OUT_OF_SERVICE)
        assertEquals(Str.ReturnReasonOutOfService, meta.reason)
        assertEquals(Str.ReturnReasonOutOfServiceTips, meta.tips)
    }

    @Test
    fun `full pile is the only centered dialog`() {
        assertTrue(ReturnDecision.meta(ReturnTypeCodes.FULL_PILE).dialog)
        assertEquals(Str.ReturnReasonFullPileTitle, ReturnDecision.meta(ReturnTypeCodes.FULL_PILE).title)
        assertTrue(!ReturnDecision.meta(ReturnTypeCodes.NO_PARKING).dialog)
    }

    @Test
    fun `unknown codes fall back to the generic reason`() {
        assertEquals(Str.ReturnReasonOutOfPark, ReturnDecision.meta(424242).reason)
    }

    @Test
    fun `return by net types keep their wire values`() {
        assertEquals(0, ReturnByNetType.NORMAL)
        assertEquals(1, ReturnByNetType.ACCEPT_PENALTY)
        assertEquals(3, ReturnByNetType.BLE_REPORT)
    }
}
