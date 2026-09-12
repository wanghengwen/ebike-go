package com.luopingtech.ebike.ops.data.analysis

import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.analysis.StationAnalyzeDetail
import com.luopingtech.ebike.ops.domain.analysis.StationAnalyzePage
import com.luopingtech.ebike.ops.domain.analysis.StationOptStateFilter
import com.luopingtech.ebike.ops.domain.analysis.StationSortOrder
import com.luopingtech.ebike.ops.domain.analysis.StationTag
import com.luopingtech.ebike.ops.domain.model.ServiceArea

interface StationAnalysisRepository {
    suspend fun list(
        area: ServiceArea,
        pageNum: Int,
        pageSize: Int = 10,
        optState: StationOptStateFilter,
        order: StationSortOrder,
        tagIds: List<String>,
    ): OpsResult<StationAnalyzePage>

    suspend fun analyzeOne(
        parkingId: String,
        serviceId: String,
    ): OpsResult<StationAnalyzeDetail>

    suspend fun tags(): OpsResult<List<StationTag>>
}

class StationAnalysisRepositoryImpl(
    private val demoMode: Boolean,
    private val api: StationAnalysisApi? = null,
) : StationAnalysisRepository {
    override suspend fun list(
        area: ServiceArea,
        pageNum: Int,
        pageSize: Int,
        optState: StationOptStateFilter,
        order: StationSortOrder,
        tagIds: List<String>,
    ): OpsResult<StationAnalyzePage> {
        if (demoMode || api == null) {
            return OpsResult.Ok(demoPage(pageNum, pageSize, optState, order, tagIds))
        }
        return api.list(
            serviceId = area.id,
            pageNum = pageNum,
            pageSize = pageSize,
            optState = optState,
            order = order,
            tagIds = tagIds,
        )
    }

    override suspend fun analyzeOne(
        parkingId: String,
        serviceId: String,
    ): OpsResult<StationAnalyzeDetail> {
        if (demoMode || api == null) {
            return OpsResult.Ok(demoDetail(parkingId, serviceId))
        }
        return api.analyzeOne(parkingId, serviceId)
    }

    override suspend fun tags(): OpsResult<List<StationTag>> {
        if (demoMode || api == null) {
            return OpsResult.Ok(demoTags())
        }
        return when (val r = api.tags()) {
            is OpsResult.Ok -> if (r.value.isEmpty()) OpsResult.Ok(demoTags()) else r
            is OpsResult.Err -> OpsResult.Ok(demoTags())
        }
    }

    companion object {
        fun demoTags(): List<StationTag> = listOf(
            StationTag("1", "学校"),
            StationTag("2", "小区"),
            StationTag("3", "商超"),
            StationTag("4", "公交站"),
            StationTag("5", "大学"),
            StationTag("6", "小学"),
        )

        fun demoPage(
            pageNum: Int,
            pageSize: Int,
            optState: StationOptStateFilter,
            order: StationSortOrder,
            tagIds: List<String>,
        ): StationAnalyzePage {
            val all = listOf(
                demoItem("101", "智富大厦", true, listOf("1"), 12, 3, 2, 1),
                demoItem("102", "荣安府西", true, listOf("2"), 8, 5, 0, 2),
                demoItem("103", "中央公园东门", false, listOf("3"), 0, 1, 4, 0),
                demoItem("104", "地铁A口", true, listOf("4"), 20, 0, 1, 3),
                demoItem("105", "校园南门", true, listOf("5", "1"), 15, 7, 6, 4),
            ).filter { item ->
                when (optState) {
                    StationOptStateFilter.All -> true
                    StationOptStateFilter.Operating -> item.operating
                    StationOptStateFilter.Stopped -> !item.operating
                }
            }.filter { item ->
                tagIds.isEmpty() || item.tags.any { it.id in tagIds }
            }.let { list ->
                when (order) {
                    StationSortOrder.None -> list
                    StationSortOrder.CanRentDesc -> list.sortedByDescending { it.canRent }
                    StationSortOrder.CanRentAsc -> list.sortedBy { it.canRent }
                    StationSortOrder.IdleDesc -> list.sortedByDescending { it.idle }
                    StationSortOrder.IdleAsc -> list.sortedBy { it.idle }
                    StationSortOrder.SiteOutDesc -> list.sortedByDescending { it.siteOut }
                    StationSortOrder.SiteOutAsc -> list.sortedBy { it.siteOut }
                    StationSortOrder.DdMissDesc -> list.sortedByDescending { it.ddMissOrder }
                    StationSortOrder.DdMissAsc -> list.sortedBy { it.ddMissOrder }
                }
            }
            val from = ((pageNum - 1) * pageSize).coerceAtLeast(0)
            val slice = if (from >= all.size) emptyList() else all.drop(from).take(pageSize)
            return StationAnalyzePage(items = slice, total = all.size, pageNum = pageNum, pageSize = pageSize)
        }

        fun demoDetail(parkingId: String, serviceId: String): StationAnalyzeDetail {
            fun hours(base: Int) = List(24) { i -> (base + (i % 5)).coerceAtLeast(0) }
            return StationAnalyzeDetail(
                parkingId = parkingId,
                serviceId = serviceId,
                canRent = 12,
                booking = 2,
                operation = 1,
                canRentHours = hours(8),
                bookingHours = hours(1),
                operationHours = hours(0),
                alarmHours = hours(0),
                faultHours = hours(0),
                siteOutHours = hours(2),
                ddMissHours = hours(1),
                idleBuckets = listOf(3, 2, 1, 0, 1, 0),
            )
        }

        private fun demoItem(
            id: String,
            name: String,
            operating: Boolean,
            tagIds: List<String>,
            canRent: Int,
            idle: Int,
            siteOut: Int,
            ddMiss: Int,
        ) = com.luopingtech.ebike.ops.domain.analysis.StationAnalyzeItem(
            parkingId = id,
            serviceId = "demo",
            name = name,
            operating = operating,
            tags = demoTags().filter { it.id in tagIds },
            canRent = canRent,
            idle = idle,
            siteOut = siteOut,
            ddMissOrder = ddMiss,
            booking = 1,
            operation = 0,
        )
    }
}
