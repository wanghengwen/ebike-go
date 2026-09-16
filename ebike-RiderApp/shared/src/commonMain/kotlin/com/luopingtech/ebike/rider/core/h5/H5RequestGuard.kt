package com.luopingtech.ebike.rider.core.h5

// H5 只能走 /client 业务接口；登录换票 /oauth/token 与密钥接口一律拒绝。
object H5RequestGuard {
    fun normalizePath(raw: String): String {
        val trimmed = raw.trim()
        if (trimmed.isBlank()) return ""
        val path = when {
            trimmed.startsWith("http://", ignoreCase = true) ||
                trimmed.startsWith("https://", ignoreCase = true) ->
                extractPath(trimmed)
            trimmed.startsWith("/") -> trimmed
            else -> "/$trimmed"
        }
        return path.substringBefore('?').substringBefore('#')
    }

    fun isAllowed(raw: String): Boolean {
        val path = normalizePath(raw).lowercase()
        if (path.isBlank()) return false
        if (path.contains("/oauth/token") || path.startsWith("/oauth")) return false
        return path.startsWith("/client/")
    }

    private fun extractPath(absolute: String): String {
        val noHash = absolute.substringBefore('#')
        val schemeEnd = noHash.indexOf("://")
        if (schemeEnd < 0) return noHash
        val afterHost = noHash.substring(schemeEnd + 3).substringAfter('/', missingDelimiterValue = "")
        return if (afterHost.isBlank()) "/" else "/$afterHost"
    }
}
