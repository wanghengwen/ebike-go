package com.luopingtech.ebike.ops.core.time

import kotlin.time.Clock
import kotlin.time.ExperimentalTime

/**
 * 当前 epoch 毫秒，签名头 `_t` 用它。
 *
 * 原先只为这一个调用引了 kotlinx-datetime；0.7 起 Clock / Instant 都搬进了 stdlib 的
 * `kotlin.time`，那个依赖就没必要留着了。
 */
@OptIn(ExperimentalTime::class)
fun nowEpochMillis(): Long = Clock.System.now().toEpochMilliseconds()
