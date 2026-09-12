package com.luopingtech.ebike.ops.domain.analysis

enum class OfflineOpsTab {
    ChangeBattery,
    MoveCar,
    Inspection,
    Repair,
}

data class MetricNowOld(
    val now: Double = 0.0,
    val old: Double = 0.0,
)

data class OfflineOpsAnalyze(
    val taskTime: MetricNowOld = MetricNowOld(),
    val distance: MetricNowOld = MetricNowOld(),
    val order: MetricNowOld = MetricNowOld(),
    val battery: MetricNowOld = MetricNowOld(),
    val orderCountOf12: MetricNowOld = MetricNowOld(),
    val validCount: Double = 0.0,
    val invalidCount: Double = 0.0,
    val movePie: List<PieSlice> = emptyList(),
    val repairSourcePie: List<PieSlice> = emptyList(),
    val repairPartPie: List<PieSlice> = emptyList(),
)

data class PieSlice(
    val name: String,
    val value: Double,
)

data class OfflineOpsTrendPoint(
    val label: String,
    val count: Int,
    val effective: Int = 0,
    val voided: Int = 0,
)

data class OfflineOpsRankRow(
    val opName: String,
    val effectiveNum: Int,
    val voidNum: Int,
    val countNum: Int,
)

data class OfflineOpsDashboard(
    val analyze: OfflineOpsAnalyze = OfflineOpsAnalyze(),
    /** 与统计分析同区间的趋势日点，用于「总计有效/无效」加总。 */
    val statsTrendDays: List<OfflineOpsTrendPoint> = emptyList(),
    val trendPoints: List<OfflineOpsTrendPoint> = emptyList(),
    val ranks: List<OfflineOpsRankRow> = emptyList(),
    val updatedAtText: String = "",
) {
    val totalEffective: Int get() = statsTrendDays.sumOf { it.effective }
    val totalVoid: Int get() = statsTrendDays.sumOf { it.voided }
}
