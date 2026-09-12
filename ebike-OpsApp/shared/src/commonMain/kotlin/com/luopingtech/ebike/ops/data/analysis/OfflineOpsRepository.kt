package com.luopingtech.ebike.ops.data.analysis

import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.analysis.MetricNowOld
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsAnalyze
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsDashboard
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsPeriod
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsRankRow
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsTab
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsTimeRanges
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsTrendGrain
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsTrendPoint
import com.luopingtech.ebike.ops.domain.analysis.PieSlice
import com.luopingtech.ebike.ops.domain.model.ServiceArea

interface OfflineOpsRepository {
    suspend fun load(
        area: ServiceArea,
        tab: OfflineOpsTab,
        period: OfflineOpsPeriod,
        grain: OfflineOpsTrendGrain,
        mineOnly: Boolean = false,
        opPin: String? = null,
    ): OpsResult<OfflineOpsDashboard>
}

class OfflineOpsRepositoryImpl(
    private val demoMode: Boolean,
    private val api: OfflineOpsApi? = null,
) : OfflineOpsRepository {
    override suspend fun load(
        area: ServiceArea,
        tab: OfflineOpsTab,
        period: OfflineOpsPeriod,
        grain: OfflineOpsTrendGrain,
        mineOnly: Boolean,
        opPin: String?,
    ): OpsResult<OfflineOpsDashboard> {
        if (demoMode || api == null) {
            return OpsResult.Ok(demoDashboard(tab, grain, skipRank = mineOnly))
        }
        val statsRange = OfflineOpsTimeRanges.period(period)
        val trendRange = OfflineOpsTimeRanges.trend(grain)
        val pinFilter = opPin?.takeIf { mineOnly && it.isNotBlank() }
        val analyze = when (val r = api.analyze(tab, area.id, statsRange, pinFilter)) {
            is OpsResult.Ok -> r.value
            is OpsResult.Err -> return r
        }
        val statsTrend = when (val r = api.trend(tab, area.id, statsRange, pinFilter)) {
            is OpsResult.Ok -> r.value
            is OpsResult.Err -> return r
        }
        val trendRaw = when (val r = api.trend(tab, area.id, trendRange, pinFilter)) {
            is OpsResult.Ok -> r.value
            is OpsResult.Err -> return r
        }
        val ranks = if (mineOnly || tab == OfflineOpsTab.Repair) {
            emptyList()
        } else {
            when (val r = api.rank(tab, area.id, statsRange, pinFilter)) {
                is OpsResult.Ok -> r.value
                is OpsResult.Err -> return r
            }
        }
        return OpsResult.Ok(
            OfflineOpsDashboard(
                analyze = analyze,
                statsTrendDays = statsTrend,
                trendPoints = aggregateTrend(trendRaw, grain),
                ranks = ranks,
                updatedAtText = OfflineOpsTimeRanges.formatUpdateTime(),
            ),
        )
    }

    private fun aggregateTrend(
        days: List<OfflineOpsTrendPoint>,
        grain: OfflineOpsTrendGrain,
    ): List<OfflineOpsTrendPoint> {
        if (grain == OfflineOpsTrendGrain.Daily) {
            // 近 15 天标签；接口若少返回则按返回顺序补 label
            val labels = OfflineOpsTimeRanges.dayLabelsForDaily()
            if (days.size >= labels.size) {
                return days.takeLast(labels.size).mapIndexed { i, p ->
                    p.copy(label = labels.getOrElse(i) { p.label })
                }
            }
            return days.mapIndexed { i, p ->
                p.copy(label = labels.getOrElse(i) { p.label.ifBlank { "${i + 1}" } })
            }
        }
        // 周/月：按返回顺序每 7 / 约 30 天累加，修正 Flutter 漏计边界的问题
        val bucketSize = if (grain == OfflineOpsTrendGrain.Weekly) 7 else 30
        if (days.isEmpty()) return emptyList()
        val out = mutableListOf<OfflineOpsTrendPoint>()
        var i = 0
        while (i < days.size) {
            val chunk = days.subList(i, minOf(i + bucketSize, days.size))
            val label = when (grain) {
                OfflineOpsTrendGrain.Weekly -> {
                    val a = chunk.first().label.substringAfterLast('-', chunk.first().label)
                    val b = chunk.last().label.substringAfterLast('-', chunk.last().label)
                    "${chunk.first().label.removePrefix("20").replace('-', '.')} / ${chunk.last().label.removePrefix("20").replace('-', '.')}"
                        .ifBlank { "$a/$b" }
                }
                else -> {
                    val day = chunk.first().label
                    val month = day.substring(5, 7).trimStart('0').ifBlank { day }
                    "${month}月"
                }
            }
            out += OfflineOpsTrendPoint(
                label = label,
                count = chunk.sumOf { it.count },
                effective = chunk.sumOf { it.effective },
                voided = chunk.sumOf { it.voided },
            )
            i += bucketSize
        }
        return out
    }

    companion object {
        fun demoDashboard(
            tab: OfflineOpsTab,
            grain: OfflineOpsTrendGrain,
            skipRank: Boolean = false,
        ): OfflineOpsDashboard {
            val labels = when (grain) {
                OfflineOpsTrendGrain.Daily -> OfflineOpsTimeRanges.dayLabelsForDaily()
                OfflineOpsTrendGrain.Weekly -> listOf("8.11/8.17", "8.18/8.24", "8.25/8.31", "9.1/9.7", "9.8/9.12")
                OfflineOpsTrendGrain.Monthly -> listOf("7月", "8月", "9月")
            }
            val trend = labels.mapIndexed { i, label ->
                OfflineOpsTrendPoint(label = label, count = (i % 4) + 1, effective = i % 3, voided = i % 2)
            }
            val analyze = OfflineOpsAnalyze(
                taskTime = MetricNowOld(now = 125.0),
                distance = MetricNowOld(now = 2.4),
                order = MetricNowOld(now = 3.0),
                battery = MetricNowOld(now = 42.0),
                orderCountOf12 = MetricNowOld(now = 1.0),
                validCount = 8.0,
                invalidCount = 2.0,
                movePie = listOf(
                    PieSlice("自主挪车", 5.0),
                    PieSlice("算法挪车", 3.0),
                    PieSlice("人工任务", 2.0),
                    PieSlice("运维规则", 1.0),
                ),
                repairSourcePie = listOf(
                    PieSlice("人工任务", 4.0),
                    PieSlice("客户反馈", 0.0),
                    PieSlice("运维规则", 2.0),
                ),
                repairPartPie = listOf(
                    PieSlice("电机", 3.0),
                    PieSlice("车把", 2.0),
                    PieSlice("脚撑", 1.0),
                ),
            )
            val ranks = if (skipRank || tab == OfflineOpsTab.Repair) {
                emptyList()
            } else {
                listOf(
                    OfflineOpsRankRow("张三", 12, 1, 13),
                    OfflineOpsRankRow("李四", 9, 2, 11),
                    OfflineOpsRankRow("王五", 7, 0, 7),
                )
            }
            return OfflineOpsDashboard(
                analyze = analyze,
                statsTrendDays = listOf(
                    OfflineOpsTrendPoint("d", 5, effective = 4, voided = 1),
                    OfflineOpsTrendPoint("d", 3, effective = 2, voided = 1),
                ),
                trendPoints = trend,
                ranks = ranks,
                updatedAtText = OfflineOpsTimeRanges.formatUpdateTime(),
            )
        }
    }
}
