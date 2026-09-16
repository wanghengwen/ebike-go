package com.luopingtech.ebike.rider.core.h5

/**
 * WebView 当前页 origin 必须等于租户 [h5.baseUrl] 的 origin。
 * `file://` / `content://` 留给本地占位页或离线包。
 */
object H5OriginPolicy {
    fun originOf(url: String): String {
        val noHash = url.substringBefore('#').trim()
        if (noHash.isBlank()) return ""
        val schemeEnd = noHash.indexOf("://")
        if (schemeEnd < 0) return ""
        val scheme = noHash.substring(0, schemeEnd).lowercase()
        if (scheme == "file" || scheme == "content") return scheme
        val hostPort = noHash.substring(schemeEnd + 3).substringBefore('/').substringBefore('?')
        if (hostPort.isBlank()) return ""
        return "$scheme://$hostPort"
    }

    fun isLocalDocument(url: String): Boolean {
        val lower = url.trim().lowercase()
        return lower.startsWith("file:") || lower.startsWith("content:")
    }

    fun isAllowed(pageUrl: String?, allowedBase: String): Boolean {
        val page = pageUrl?.trim().orEmpty()
        if (page.isEmpty() || page.equals("about:blank", ignoreCase = true)) return false
        if (isLocalDocument(page)) return true
        val allowed = allowedBase.trim()
        if (allowed.isEmpty()) return false
        val pageOrigin = originOf(page)
        val allowedOrigin = originOf(allowed)
        return pageOrigin.isNotBlank() && allowedOrigin.isNotBlank() && pageOrigin == allowedOrigin
    }
}
