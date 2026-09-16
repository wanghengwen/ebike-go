package com.luopingtech.ebike.rider.ui.h5

/**
 * UniApp H5 在 WebView 里深链打开时，history 里常残留 launch / home / profile。
 * 从原生打开钱包等长尾页再 back 到这些无关壳页，应关闭容器；
 * 若本次入口本身就是个人中心，退回到入口应留下，由下一次后退再关。
 */
internal object H5BackPolicy {
    private val BOOTSTRAP_PREFIXES = listOf(
        "pages/launch/",
        "pages/home/",
        "pages/map/",
        "pages/account/profile",
    )

    fun hashPath(url: String?): String {
        if (url.isNullOrBlank()) return ""
        return url.substringAfter('#', missingDelimiterValue = "")
            .substringBefore('?')
            .trimStart('/')
    }

    fun isBootstrapRoute(urlOrHash: String?): Boolean {
        val path = hashPath(urlOrHash).ifBlank {
            urlOrHash?.substringBefore('?')?.trimStart('/')?.substringAfter("h5/") ?: ""
        }
        if (path.isBlank()) return true
        return BOOTSTRAP_PREFIXES.any { path.startsWith(it) }
    }

    fun atEntry(currentUrl: String?, entryUrl: String): Boolean {
        val current = hashPath(currentUrl)
        val entry = hashPath(entryUrl)
        if (entry.isBlank()) return true
        if (current.isBlank()) return true
        return current == entry
    }

    /**
     * WebView.goBack 落地后是否关容器。
     * 退回本次入口 → 否（留在个人中心等入口页）；
     * 落到无关壳页 → 是。
     */
    fun shouldCloseAfterBack(currentUrl: String?, entryUrl: String): Boolean {
        if (atEntry(currentUrl, entryUrl)) return false
        return isBootstrapRoute(currentUrl)
    }
}
