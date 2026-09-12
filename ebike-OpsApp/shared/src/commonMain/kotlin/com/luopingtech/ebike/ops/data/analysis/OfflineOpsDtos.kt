package com.luopingtech.ebike.ops.data.analysis

import com.luopingtech.ebike.ops.domain.analysis.MetricNowOld
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsAnalyze
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsRankRow
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsTrendPoint
import com.luopingtech.ebike.ops.domain.analysis.PieSlice
import kotlinx.serialization.Serializable

@Serializable
data class OfflineOpsMetricDto(
    val now: Double? = 0.0,
    val old: Double? = 0.0,
) {
    fun toDomain(): MetricNowOld = MetricNowOld(now = now ?: 0.0, old = old ?: 0.0)
}

@Serializable
data class OfflineOpsMacroDto(
    val taskTime: OfflineOpsMetricDto? = null,
    val distance: OfflineOpsMetricDto? = null,
    val order: OfflineOpsMetricDto? = null,
    val battery: OfflineOpsMetricDto? = null,
    val orderCountOf12: OfflineOpsMetricDto? = null,
)

@Serializable
data class OfflineOpsTaskSourceAnalyseDto(
    val ruleNum: Double? = 0.0,
    val selfNum: Double? = 0.0,
    val manualNum: Double? = 0.0,
    val algorithmNum: Double? = 0.0,
)

@Serializable
data class OfflineOpsTaskSourceDto(
    val customerTask: Double? = 0.0,
    val manTask: Double? = 0.0,
    val rule: Double? = 0.0,
)

@Serializable
data class OfflineOpsNamedCountDto(
    val name: String? = "",
    val count: Double? = 0.0,
)

@Serializable
data class OfflineOpsValidProportionDto(
    val valid: Double? = 0.0,
    val invalid: Double? = 0.0,
)

@Serializable
data class OfflineOpsPieDto(
    val taskSourceAnalyse: OfflineOpsTaskSourceAnalyseDto? = null,
    val taskSource: OfflineOpsTaskSourceDto? = null,
    val fixPart: List<OfflineOpsNamedCountDto>? = null,
    val validProportion: OfflineOpsValidProportionDto? = null,
)

@Serializable
data class OfflineOpsAnalyzeDto(
    val macroInfo: OfflineOpsMacroDto? = null,
    val pieInfo: OfflineOpsPieDto? = null,
) {
    fun toDomain(): OfflineOpsAnalyze {
        val macro = macroInfo
        val pie = pieInfo
        val move = pie?.taskSourceAnalyse
        val source = pie?.taskSource
        return OfflineOpsAnalyze(
            taskTime = macro?.taskTime?.toDomain() ?: MetricNowOld(),
            distance = macro?.distance?.toDomain() ?: MetricNowOld(),
            order = macro?.order?.toDomain() ?: MetricNowOld(),
            battery = macro?.battery?.toDomain() ?: MetricNowOld(),
            orderCountOf12 = macro?.orderCountOf12?.toDomain() ?: MetricNowOld(),
            validCount = pie?.validProportion?.valid ?: 0.0,
            invalidCount = pie?.validProportion?.invalid ?: 0.0,
            movePie = move?.let {
                listOf(
                    PieSlice("自主挪车", it.selfNum ?: 0.0),
                    PieSlice("算法挪车", it.algorithmNum ?: 0.0),
                    PieSlice("人工任务", it.manualNum ?: 0.0),
                    PieSlice("运维规则", it.ruleNum ?: 0.0),
                )
            }.orEmpty(),
            repairSourcePie = source?.let {
                listOf(
                    PieSlice("人工任务", it.manTask ?: 0.0),
                    PieSlice("客户反馈", it.customerTask ?: 0.0),
                    PieSlice("运维规则", it.rule ?: 0.0),
                )
            }.orEmpty(),
            repairPartPie = pie?.fixPart.orEmpty().map {
                PieSlice(it.name.orEmpty().ifBlank { "-" }, it.count ?: 0.0)
            },
        )
    }
}

@Serializable
data class OfflineOpsSeriesDto(
    val day: String? = null,
    val opName: String? = null,
    val effectiveNum: Int? = 0,
    val voidNum: Int? = 0,
    val countNum: Int? = 0,
) {
    fun toTrend(label: String = day.orEmpty()): OfflineOpsTrendPoint =
        OfflineOpsTrendPoint(
            label = label,
            count = countNum ?: 0,
            effective = effectiveNum ?: 0,
            voided = voidNum ?: 0,
        )

    fun toRank(): OfflineOpsRankRow =
        OfflineOpsRankRow(
            opName = opName.orEmpty().ifBlank { "--" },
            effectiveNum = effectiveNum ?: 0,
            voidNum = voidNum ?: 0,
            countNum = countNum ?: 0,
        )
}
