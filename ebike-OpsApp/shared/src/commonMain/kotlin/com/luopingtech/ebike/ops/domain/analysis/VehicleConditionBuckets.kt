package com.luopingtech.ebike.ops.domain.analysis

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.domain.model.Vehicle
import com.luopingtech.ebike.ops.domain.vehicle.VehicleOperationStates

/**
 * 车况分布分档，对齐遗留 Flutter `vehicle_condition_distribution_bloc`
 *（不要抄运营大屏粗分档，也不要抄原生刷新重分档里的单位 bug）。
 *
 * - 排除 `operationState` 含下架(1) 的车
 * - 电量：`(low, high]`，标签写闭区间但实现是左开右闭
 * - 闲置：须 `lockTime > unlockTime` 且空闲秒数 **严格大于** 1 小时
 * - 百分比分母 = 排除下架后的全量车数；`(count * 100 / total).truncate()`
 */
object VehicleConditionBuckets {
    data class Bucket(
        val index: Int,
        val labelKey: Str,
        val count: Int,
        val percent: Int,
        /** ARGB，与 Flutter 条色一致。 */
        val colorArgb: Long,
        val vehicles: List<Vehicle>,
    )

    data class Result(
        val total: Int,
        val buckets: List<Bucket>,
    )

    private data class Spec(
        val labelKey: Str,
        val colorArgb: Long,
    )

    private val batterySpecs: List<Spec> = listOf(
        Spec(Str.VcdBattery90, 0xFF67BC00),
        Spec(Str.VcdBattery80, 0xFF7AC000),
        Spec(Str.VcdBattery70, 0xFFA8C800),
        Spec(Str.VcdBattery60, 0xFFBBCA00),
        Spec(Str.VcdBattery50, 0xFFD6CF00),
        Spec(Str.VcdBattery40, 0xFFE5D100),
        Spec(Str.VcdBattery30, 0xFFFBCE00),
        Spec(Str.VcdBattery20, 0xFFFC9A00),
        Spec(Str.VcdBattery10, 0xFFFE6D00),
        Spec(Str.VcdBattery0, 0xFFFC3200),
    )

    /** 闲置条统一蓝渐变起点色（UI 可用终点色做渐变）。 */
    private val idleColor: Long = 0xFF188DF0

    private val idleSpecs: List<Spec> = listOf(
        Spec(Str.VcdIdle1to3, idleColor),
        Spec(Str.VcdIdle3to6, idleColor),
        Spec(Str.VcdIdle6to12, idleColor),
        Spec(Str.VcdIdle12to24, idleColor),
        Spec(Str.VcdIdle24to48, idleColor),
        Spec(Str.VcdIdle48to72, idleColor),
        Spec(Str.VcdIdleOver3Days, idleColor),
    )

    val idleGradientEndArgb: Long = 0xFF7EBDF5

    fun battery(vehicles: List<Vehicle>): Result {
        val eligible = vehicles.filterNot { it.operationStates.contains(VehicleOperationStates.OFF) }
        val bins = List(batterySpecs.size) { mutableListOf<Vehicle>() }
        eligible.forEach { v ->
            bins[batteryIndex(v.restBattery)].add(v)
        }
        return toResult(eligible.size, batterySpecs, bins)
    }

    fun idle(vehicles: List<Vehicle>, nowEpochMs: Long): Result {
        val eligible = vehicles.filterNot { it.operationStates.contains(VehicleOperationStates.OFF) }
        val bins = List(idleSpecs.size) { mutableListOf<Vehicle>() }
        eligible.forEach { v ->
            idleIndex(v.lockTimeMs, v.unlockTimeMs, nowEpochMs)?.let { bins[it].add(v) }
        }
        return toResult(eligible.size, idleSpecs, bins)
    }

    fun batteryIndex(restBattery: Int): Int = when {
        restBattery > 90 -> 0
        restBattery > 80 -> 1
        restBattery > 70 -> 2
        restBattery > 60 -> 3
        restBattery > 50 -> 4
        restBattery > 40 -> 5
        restBattery > 30 -> 6
        restBattery > 20 -> 7
        restBattery > 10 -> 8
        else -> 9
    }

    /**
     * @return 档位 index，或 `null` 表示不计入任何闲置档（骑行中 / ≤1h）。
     */
    fun idleIndex(lockTimeMs: Long, unlockTimeMs: Long, nowEpochMs: Long): Int? {
        if (lockTimeMs <= unlockTimeMs) return null
        val idleSec = (nowEpochMs - lockTimeMs) / 1000
        return when {
            idleSec > 3600L * 72 -> 6
            idleSec > 3600L * 48 -> 5
            idleSec > 3600L * 24 -> 4
            idleSec > 3600L * 12 -> 3
            idleSec > 3600L * 6 -> 2
            idleSec > 3600L * 3 -> 1
            idleSec > 3600L -> 0
            else -> null
        }
    }

    private fun toResult(
        total: Int,
        specs: List<Spec>,
        bins: List<List<Vehicle>>,
    ): Result {
        val buckets = specs.mapIndexed { index, spec ->
            val list = bins[index]
            val count = list.size
            val percent = if (total <= 0) 0 else (count * 100) / total
            Bucket(
                index = index,
                labelKey = spec.labelKey,
                count = count,
                percent = percent,
                colorArgb = spec.colorArgb,
                vehicles = list,
            )
        }
        return Result(total = total, buckets = buckets)
    }
}
