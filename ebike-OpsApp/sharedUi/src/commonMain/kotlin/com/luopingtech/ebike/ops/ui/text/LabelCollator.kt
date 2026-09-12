package com.luopingtech.ebike.ops.ui.text

/**
 * 商户名排序要按中文读音而不是码位，A-Z 索引条才对得上。
 * 各端拿系统的排序规则，别自己塞拼音表。
 */
expect fun labelComparator(): Comparator<String>
