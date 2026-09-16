package com.luopingtech.ebike.rider.domain.home

/**
 * 首页底部动态导航项（对齐 UniApp `getHomeNav` / `bHomeNavCO`）。
 * 名称与图标来自运营配置，不是本地写死。
 */
data class HomeNavItem(
    val id: String = "",
    val name: String = "",
    val iconUrl: String = "",
    val chainType: Int = 0,
    val linkUrl: String = "",
    val linkTitle: String = "",
    val appId: String = "",
)

/**
 * 点击后在原生侧打开的目标：优先 [H5ScreenKind]，否则任意 hash。
 */
sealed class HomeNavTarget {
    data class Kind(val kind: com.luopingtech.ebike.rider.core.config.H5ScreenKind) : HomeNavTarget()
    data class Hash(val route: String) : HomeNavTarget()
    data object Unsupported : HomeNavTarget()
}
