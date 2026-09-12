package com.luopingtech.ebike.ops.data.analysis

import com.luopingtech.ebike.ops.domain.analysis.ReturnCarAnalyzeResult
import com.luopingtech.ebike.ops.domain.analysis.ReturnCarPoint
import kotlinx.serialization.ExperimentalSerializationApi
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonNames

@Serializable
data class ReturnCarPointDto(
    val l: String = "",
    val p: Boolean = false,
) {
    fun toDomain(): ReturnCarPoint? {
        val parts = l.split(',')
        if (parts.size < 2) return null
        // 后端拼接 lng,lat（对齐 iOS，勿用遗留 Android 反序）
        val lng = parts[0].trim().toDoubleOrNull() ?: return null
        val lat = parts[1].trim().toDoubleOrNull() ?: return null
        if (lat == 0.0 && lng == 0.0) return null
        return ReturnCarPoint(lat = lat, lng = lng, abnormal = p)
    }
}

@OptIn(ExperimentalSerializationApi::class)
@Serializable
data class ReturnCarAnalyzeDto(
    val total: Int = 0,
    @JsonNames("pCount")
    val pcount: Int = 0,
    @JsonNames("nCount")
    val ncount: Int = 0,
    @JsonNames("oList")
    val olist: List<ReturnCarPointDto> = emptyList(),
) {
    fun toDomain(): ReturnCarAnalyzeResult = ReturnCarAnalyzeResult(
        total = total,
        normalCount = ncount,
        abnormalCount = pcount,
        points = olist.mapNotNull { it.toDomain() },
    )
}
