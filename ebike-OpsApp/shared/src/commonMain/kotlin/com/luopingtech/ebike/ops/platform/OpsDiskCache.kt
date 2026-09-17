package com.luopingtech.ebike.ops.platform

/**
 * 对齐遗留 SettingActivity 清除缓存：清理应用磁盘缓存目录。
 * @return true 表示清理成功（或无可清内容）
 */
expect fun clearOpsDiskCache(): Boolean
