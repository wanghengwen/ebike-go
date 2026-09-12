package com.luopingtech.ebike.ops.core.config

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.domain.model.UserSession
import com.luopingtech.ebike.ops.domain.permission.OpsPermissionCodes
import io.ktor.http.encodeURLParameter

/**
 * 一个 H5 屏 = 一条 SPA hash 路由 + 一个标题 + 一个权限码。
 *
 * 加一屏只需要在这里加一行，工作台入口和 [H5Screen] 的标题都从这张表生成，
 * 不用再动 `MainActivity`。路由要和 `webH5/src/router/index.ts` 里注册的 path 一致。
 */
enum class H5ScreenKind(
    val route: String,
    val titleKey: Str,
    /**
     * 满足其一即放行，语义同遗留 `v-hasCode` 的 `||`。
     * 同一个功能在 App 和 PC 后台常常各有一个码，两边发码习惯又不统一，
     * 所以这里收一组而不是一个。
     */
    val permissionCodes: Set<String>,
    /**
     * 两块大屏各有一个整串 URL 的租户配置字段，是历史遗留；其余屏返回 null，
     * 走 [H5ScreensConfig.baseUrl] 拼 [route]。
     */
    internal val legacyUrl: (H5ScreensConfig) -> String? = { null },
) {
    Operation(
        "#/newOperationScreen",
        Str.OperationScreen,
        setOf(OpsPermissionCodes.OPERATION_DATA),
        legacyUrl = { it.operationUrl },
    ),
    Revenue(
        "#/newRevenueHome",
        Str.RevenueScreen,
        setOf(OpsPermissionCodes.REVENUE_DATA),
        legacyUrl = { it.revenueUrl },
    ),
}

/**
 * Resolve tenant-configured H5 URLs, optionally with auth query
 * (legacy iOS BigScreenViewController contract).
 */
object H5ScreenUrls {
    /**
     * 两块大屏保留各自的整串配置字段，是为了让已经在跑的租户配置不用改。
     * 其余屏一律由 [H5ScreensConfig.baseUrl] 拼 [H5ScreenKind.route] 得到，
     * 否则每加一屏都要给所有租户加一个配置字段。
     */
    fun rawUrl(config: TenantConfig, kind: H5ScreenKind): String {
        val explicit = kind.legacyUrl(config.h5)?.trim().orEmpty()
        if (explicit.isNotBlank()) return explicit
        val base = resolveBase(config)
        return if (base.isBlank()) "" else base + kind.route
    }

    /**
     * 没显式配 `baseUrl` 时从两块大屏的地址里反推，取 `#` 之前的部分。
     * 老租户配置只填了 operationUrl / revenueUrl，这样它们不用改就能开管理屏。
     */
    private fun resolveBase(config: TenantConfig): String {
        val explicit = config.h5.baseUrl.trim()
        if (explicit.isNotBlank()) return explicit.substringBefore('#')
        val fallback = config.h5.operationUrl.trim().ifBlank { config.h5.revenueUrl.trim() }
        return fallback.substringBefore('#')
    }

    fun isConfigured(config: TenantConfig, kind: H5ScreenKind): Boolean =
        rawUrl(config, kind).isNotBlank()

    fun resolve(
        config: TenantConfig,
        kind: H5ScreenKind,
        session: UserSession?,
        deviceId: String,
        platform: String = "android",
    ): String? {
        val base = rawUrl(config, kind)
        if (base.isBlank()) return null
        if (!config.h5.appendAuthQuery) return base
        val params = linkedMapOf(
            "apiHost" to config.api.baseUrl.trim().trimEnd('/'),
            "refreshToken" to (session?.refreshToken.orEmpty()),
            "accessToken" to (session?.accessToken.orEmpty()),
            "tenantId" to config.tenantId.ifBlank { session?.tenantId.orEmpty() },
            "sign" to config.auth.signSecret,
            "tenantSecret" to config.auth.businessSecret,
            "opMan" to (session?.phone?.ifBlank { session.displayName }.orEmpty()),
            "deviceId" to deviceId,
            "platform" to platform,
            "themeColor" to run {
                val raw = config.branding.primaryColor.trim()
                if (raw.isNotBlank()) raw else BrandPalette.cssHexRgb(BrandPalette.from(config.branding).primary)
            },
            "lightTxtColor" to config.branding.lightTextColor.trim().ifBlank { "#282828" },
        )
        return appendQuery(base, params)
    }

    /**
     * Append query after `#/route` when present (SPA hash routing), else after path.
     */
    fun appendQuery(url: String, params: Map<String, String>): String {
        val filtered = params.filterValues { it.isNotBlank() }
        if (filtered.isEmpty()) return url
        val query = filtered.entries.joinToString("&") { (k, v) ->
            "${k.encodeURLParameter()}=${v.encodeURLParameter()}"
        }
        val hashIdx = url.indexOf('#')
        if (hashIdx >= 0) {
            val before = url.substring(0, hashIdx)
            val hash = url.substring(hashIdx)
            val sep = if (hash.contains('?')) "&" else "?"
            return before + hash + sep + query
        }
        val sep = if (url.contains('?')) "&" else "?"
        return url + sep + query
    }
}
