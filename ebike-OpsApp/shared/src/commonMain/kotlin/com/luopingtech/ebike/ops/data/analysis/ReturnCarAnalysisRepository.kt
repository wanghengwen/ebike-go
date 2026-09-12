package com.luopingtech.ebike.ops.data.analysis

import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsPeriod
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsTimeRanges
import com.luopingtech.ebike.ops.domain.analysis.ReturnCarAnalyzeResult
import com.luopingtech.ebike.ops.domain.analysis.ReturnCarPoint
import com.luopingtech.ebike.ops.domain.model.ServiceArea

interface ReturnCarAnalysisRepository {
    suspend fun load(area: ServiceArea, period: OfflineOpsPeriod): OpsResult<ReturnCarAnalyzeResult>
}

class ReturnCarAnalysisRepositoryImpl(
    private val demoMode: Boolean,
    private val api: ReturnCarAnalysisApi? = null,
) : ReturnCarAnalysisRepository {
    override suspend fun load(
        area: ServiceArea,
        period: OfflineOpsPeriod,
    ): OpsResult<ReturnCarAnalyzeResult> {
        if (demoMode || api == null) {
            return OpsResult.Ok(demoResult())
        }
        val range = OfflineOpsTimeRanges.period(period)
        return api.analyze(area.id, range)
    }

    companion object {
        fun demoResult(): ReturnCarAnalyzeResult {
            // 南京附近散点，演示正常(绿)/异常(红)
            val baseLat = 32.0603
            val baseLng = 118.7969
            val points = buildList {
                for (i in 0 until 40) {
                    val lat = baseLat + (i % 8) * 0.008 - 0.02
                    val lng = baseLng + (i / 8) * 0.01 - 0.02
                    add(ReturnCarPoint(lat = lat, lng = lng, abnormal = i % 5 == 0))
                }
            }
            val abnormal = points.count { it.abnormal }
            return ReturnCarAnalyzeResult(
                total = points.size,
                normalCount = points.size - abnormal,
                abnormalCount = abnormal,
                points = points,
            )
        }
    }
}
