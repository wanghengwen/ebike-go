package com.luopingtech.ebike.rider.domain.home

import com.luopingtech.ebike.rider.core.config.BrandingConfig
import com.luopingtech.ebike.rider.core.config.H5ScreenKind

/**
 * 把运营下发的 `jumpPage.linkUrl`（含老微信 `/pagesSub/...`）映射到 Rider H5 入口。
 * 对齐 UniApp `mapLegacyPath.ts` + `openUserClickAction`。
 */
object HomeNavJump {
    private val LEGACY: Map<String, H5ScreenKind> = mapOf(
        "/pagesSub/order/order" to H5ScreenKind.Orders,
        "/pages-sub/account/orders/orders" to H5ScreenKind.Orders,
        "/pagesSub/wallet/wallet" to H5ScreenKind.Wallet,
        "/pages-sub/pay/wallet/wallet" to H5ScreenKind.Wallet,
        "/pagesSub/charge/charge" to H5ScreenKind.Recharge,
        "/pages-sub/pay/recharge/recharge" to H5ScreenKind.Recharge,
        "/pagesSub/customerService/customerService" to H5ScreenKind.CustomerService,
        "/pages-sub/support/customer-service/customer-service" to H5ScreenKind.CustomerService,
        "/pagesSub/help/help" to H5ScreenKind.Help,
        "/pages-sub/support/help/help" to H5ScreenKind.Help,
        "/pagesSub/accountRules/accountRules" to H5ScreenKind.BillingRules,
        "/pages-sub/account/billing-rules/billing-rules" to H5ScreenKind.BillingRules,
        "/pagesSub/msgList/msgList" to H5ScreenKind.Messages,
        "/pages-sub/support/messages/list" to H5ScreenKind.Messages,
        "/pages/userInfo/userInfo" to H5ScreenKind.Profile,
        "/pages/account/profile" to H5ScreenKind.Profile,
        "/pagesSub/repair/repair" to H5ScreenKind.Repair,
        "/pages-sub/support/repair/repair" to H5ScreenKind.Repair,
        "/pagesSub/protocol/protocol" to H5ScreenKind.Protocol,
        "/pages-sub/account/protocol/protocol" to H5ScreenKind.Protocol,
        "/pagesSub/activity/activity" to H5ScreenKind.Activity,
        "/pages-sub/account/activity/activity" to H5ScreenKind.Activity,
        "/pagesSub/myCard/myCard" to H5ScreenKind.Cards,
        "/pages-sub/account/cards/cards" to H5ScreenKind.Cards,
        "/pagesSub/cardCenter/cardCenter" to H5ScreenKind.CardShop,
        "/pages-sub/account/card-shop/shop" to H5ScreenKind.CardShop,
        "/pagesSub/setting/setting" to H5ScreenKind.Settings,
        "/pages/account/settings" to H5ScreenKind.Settings,
        "/pagesSub/invite/invite" to H5ScreenKind.Invite,
        "/pages-sub/account/invite/invite" to H5ScreenKind.Invite,
        "/pagesSub/creditScore/creditScore" to H5ScreenKind.Credit,
        "/pages-sub/account/credit/credit" to H5ScreenKind.Credit,
    )

    /** 与 UniApp `iconCfg` / `mapCfg` 同源，接口未配置时的兜底图（非凡皮肤）。 */
    private const val FALLBACK_ICON_ORDER =
        "https://feifan-frontend.oss-cn-shenzhen.aliyuncs.com/miniapp/project/feifan/new_userinfo_route.png"
    private const val FALLBACK_ICON_WALLET =
        "https://feifan-frontend.oss-cn-shenzhen.aliyuncs.com/miniapp/project/feifan/icon_wallet.png"
    private const val FALLBACK_ICON_SERVICE =
        "https://feifan-frontend.oss-cn-shenzhen.aliyuncs.com/miniapp/project/feifan/customerService.png"
    private const val FALLBACK_ICON_BILLING =
        "https://feifan-frontend.oss-cn-shenzhen.aliyuncs.com/miniapp/project/feifan/billingRules.png"

    fun resolve(item: HomeNavItem): HomeNavTarget {
        // chainType: 0 内部页，1 H5 URL，2 其他小程序（原生暂不支持）
        when (item.chainType) {
            2, 3 -> return HomeNavTarget.Unsupported
            1 -> {
                val url = item.linkUrl.trim()
                if (url.startsWith("http://", ignoreCase = true) ||
                    url.startsWith("https://", ignoreCase = true)
                ) {
                    // 外链 H5：交给容器打开任意 hash 不行，需 webview；暂用 Unsupported toast
                    return HomeNavTarget.Unsupported
                }
            }
        }
        val path = normalizePath(item.linkUrl)
        if (path.isBlank()) return HomeNavTarget.Unsupported
        LEGACY[path]?.let { return HomeNavTarget.Kind(it) }
        val uni = if (path.startsWith("/pagesSub/")) {
            path.replaceFirst("/pagesSub/", "/pages-sub/")
        } else {
            path
        }
        LEGACY[uni]?.let { return HomeNavTarget.Kind(it) }
        if (uni.startsWith("/pages") || uni.startsWith("/pages-sub")) {
            return HomeNavTarget.Hash(uni)
        }
        return HomeNavTarget.Unsupported
    }

    private fun normalizePath(raw: String): String {
        var url = raw.trim()
        if (url.startsWith("miniapp:/")) {
            url = url.substringAfter(":/")
        }
        return url.substringBefore('?').substringBefore('#')
    }

    /**
     * 接口为空 / 失败时的默认四宫格。
     * 人人骑行 / 伴我出行现网 `getHomeNavByServiceId` 常返回 `data:[]`，不能只靠接口。
     * 图标优先租户 `branding.homeIcon*`，否则用非凡皮肤 CDN。
     */
    fun defaultItems(branding: BrandingConfig = BrandingConfig()): List<HomeNavItem> = listOf(
        HomeNavItem(
            id = "default-orders",
            name = "行程",
            iconUrl = branding.homeIconTrips.ifBlank { FALLBACK_ICON_ORDER },
            chainType = 0,
            linkUrl = "/pagesSub/order/order",
        ),
        HomeNavItem(
            id = "default-wallet",
            name = "钱包",
            iconUrl = branding.homeIconWallet.ifBlank { FALLBACK_ICON_WALLET },
            chainType = 0,
            linkUrl = "/pagesSub/wallet/wallet",
        ),
        HomeNavItem(
            id = "default-cs",
            name = "客服",
            iconUrl = branding.homeIconService.ifBlank { FALLBACK_ICON_SERVICE },
            chainType = 0,
            linkUrl = "/pagesSub/customerService/customerService",
        ),
        HomeNavItem(
            id = "default-billing",
            name = "收费说明",
            iconUrl = branding.homeIconBilling.ifBlank { FALLBACK_ICON_BILLING },
            chainType = 0,
            linkUrl = "/pagesSub/accountRules/accountRules",
        ),
    )
}
