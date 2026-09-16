package com.luopingtech.ebike.rider.core.config

import com.luopingtech.ebike.rider.core.i18n.RiderLanguage
import io.ktor.http.encodeURLParameter

/**
 * C 端 H5 长尾路由表（对齐 UniApp `pages.json` 中非骑行主循环的页面）。
 *
 * 也可走 [H5ScreenUrls.resolveHash] 打开任意 hash。URL **只**拼 lang / tenantId / themeColor，
 * 绝不拼 `signSecret` / `businessSecret` / token。
 */
enum class H5ScreenKind(
    val route: String,
    val titleZh: String,
    val titleEn: String,
) {
    Profile("#/pages/account/profile", "个人中心", "Profile"),
    Settings("#/pages/account/settings", "设置", "Settings"),
    Wallet("#/pages-sub/pay/wallet/wallet", "我的钱包", "Wallet"),
    Recharge("#/pages-sub/pay/recharge/recharge", "充值", "Top up"),
    Withdraw("#/pages-sub/pay/withdraw/withdraw", "提现", "Withdraw"),
    DepositRefund("#/pages-sub/pay/deposit-refund/deposit-refund", "退押金", "Deposit refund"),
    WalletRecords("#/pages-sub/account/wallet-records/wallet-records", "账户明细", "Wallet records"),
    InvoiceList("#/pages-sub/pay/invoice/list", "开具发票", "Invoices"),
    InvoiceCreate("#/pages-sub/pay/invoice/create", "按订单开票", "Create invoice"),
    InvoiceApply("#/pages-sub/pay/invoice/apply", "开票详情", "Invoice apply"),
    InvoiceHistory("#/pages-sub/pay/invoice/history", "开票历史", "Invoice history"),
    InvoiceDetail("#/pages-sub/pay/invoice/detail", "发票详情", "Invoice detail"),
    CostDetail("#/pages-sub/pay/cost-detail/cost-detail", "费用明细", "Cost detail"),
    Pay("#/pages/pay/pay", "骑行结束", "Trip pay"),
    Orders("#/pages-sub/account/orders/orders", "我的行程", "Trips"),
    OrderDetail("#/pages-sub/account/orders/order-detail", "行程详情", "Trip detail"),
    Cards("#/pages-sub/account/cards/cards", "我的卡券", "Cards"),
    CardsExpired("#/pages-sub/account/cards/cards-expired", "失效卡券", "Expired cards"),
    CardShop("#/pages-sub/account/card-shop/shop", "卡券商城", "Card shop"),
    CardRules("#/pages-sub/account/card-rules/card-rules", "购卡须知", "Card rules"),
    Voucher("#/pages-sub/account/voucher/voucher", "兑换券", "Vouchers"),
    Activity("#/pages-sub/account/activity/activity", "活动中心", "Activities"),
    Invite("#/pages-sub/account/invite/invite", "邀请有礼", "Invite"),
    InviteRecords("#/pages-sub/account/invite/records", "邀请记录", "Invite records"),
    InviteShare("#/pages-sub/account/invite/share", "邀您免费骑", "Invite share"),
    InviteRules("#/pages-sub/account/invite/rules", "活动规则", "Invite rules"),
    BillingRules("#/pages-sub/account/billing-rules/billing-rules", "计费规则", "Billing rules"),
    Qualification("#/pages-sub/account/qualification/qualification", "用车资格", "Qualification"),
    Credit("#/pages-sub/account/credit/credit", "信用分", "Credit"),
    PayScore("#/pages-sub/account/pay-score/pay-score", "微信支付分", "Pay score"),
    Career("#/pages-sub/account/career/career", "职业认证", "Career"),
    Security("#/pages-sub/account/security/security", "账户与安全", "Security"),
    ChangePhone("#/pages-sub/account/change-phone/change-phone", "修改手机号", "Change phone"),
    Cancel("#/pages-sub/account/cancel/cancel", "注销账号", "Delete account"),
    About("#/pages-sub/account/about/about", "关于我们", "About"),
    Protocol("#/pages-sub/account/protocol/protocol", "条款与隐私政策", "Legal"),
    ProtocolDetail("#/pages-sub/account/protocol/detail", "协议详情", "Legal detail"),
    Help("#/pages-sub/support/help/help", "客户服务", "Help"),
    Faq("#/pages-sub/support/faq/faq", "常见问题", "FAQ"),
    FaqDetail("#/pages-sub/support/faq/faq-detail", "问题详情", "FAQ detail"),
    Messages("#/pages-sub/support/messages/list", "消息中心", "Messages"),
    MessageDetail("#/pages-sub/support/messages/detail", "消息详情", "Message"),
    CustomerService("#/pages-sub/support/customer-service/customer-service", "联系客服", "Support"),
    Repair("#/pages-sub/support/repair/repair", "车辆报修", "Repair"),
    RepairList("#/pages-sub/support/repair/repair-list", "报修记录", "Repair list"),
    RepairProgress("#/pages-sub/support/repair/repair-progress", "报修进度", "Repair progress"),
    Objection("#/pages-sub/support/objection/objection", "费用异议", "Objection"),
    Violation("#/pages-sub/support/violation/violation", "违规举报", "Report"),
    ApplyStation("#/pages-sub/support/apply-station/apply-station", "申请还车站", "Apply station"),
    ;

    fun title(language: RiderLanguage): String =
        if (language == RiderLanguage.EN) titleEn else titleZh
}

object H5ScreenUrls {
    private val SECRET_QUERY_KEYS = setOf(
        "sign",
        "tenantsecret",
        "signsecret",
        "businesssecret",
        "accesstoken",
        "refreshtoken",
        "token",
    )

    fun rawUrl(config: TenantConfig, kind: H5ScreenKind): String {
        val base = resolveBase(config)
        return if (base.isBlank()) "" else base + kind.route
    }

    fun isConfigured(config: TenantConfig, kind: H5ScreenKind? = null): Boolean {
        val base = resolveBase(config)
        if (base.isBlank()) return false
        if (kind == null) return true
        return rawUrl(config, kind).isNotBlank()
    }

    fun resolve(
        config: TenantConfig,
        kind: H5ScreenKind,
        language: RiderLanguage,
    ): String? {
        val base = rawUrl(config, kind)
        if (base.isBlank()) return null
        return appendQuery(base, displayParams(config, language))
    }

    /**
     * 打开任意 UniApp hash / 路径，例如 `#/pages-sub/pay/wallet/wallet` 或 `/pages/account/settings`。
     */
    fun resolveHash(
        config: TenantConfig,
        route: String,
        language: RiderLanguage,
    ): String? {
        val base = resolveBase(config)
        if (base.isBlank()) return null
        val hash = normalizeHash(route)
        if (hash.isBlank()) return null
        return appendQuery(base + hash, displayParams(config, language))
    }

    fun resolveBase(config: TenantConfig): String =
        config.h5.baseUrl.trim().substringBefore('#')

    fun displayParams(config: TenantConfig, language: RiderLanguage): Map<String, String> {
        val rawColor = config.branding.primaryColor.trim()
        val theme = rawColor.ifBlank {
            BrandPalette.cssHexRgb(BrandPalette.from(config.branding).primary)
        }
        return linkedMapOf(
            "lang" to language.tag,
            "tenantId" to config.tenantId.trim(),
            "themeColor" to theme,
        )
    }

    /**
     * SPA hash 路由：query 接在 `#/route` 后面，避免被文档基址吃掉。
     */
    fun appendQuery(url: String, params: Map<String, String>): String {
        val filtered = params
            .filterKeys { it.lowercase() !in SECRET_QUERY_KEYS }
            .filterValues { it.isNotBlank() }
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

    fun normalizeHash(route: String): String {
        val trimmed = route.trim()
        if (trimmed.isBlank()) return ""
        if (trimmed.startsWith("#")) return trimmed
        val path = if (trimmed.startsWith("/")) trimmed else "/$trimmed"
        return "#$path"
    }
}
