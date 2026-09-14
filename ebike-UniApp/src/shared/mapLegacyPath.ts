/** Map legacy wechat `/pages` + `/pagesSub` paths to UniApp routes. */
const LEGACY_PAGE_MAP: Record<string, string> = {
  '/pages/userInfo/userInfo': '/pages/account/profile',
  '/pages/home/home': '/pages/home/home',
  '/pages/map/map': '/pages/map/map',
  '/pages/quickLogin/quickLogin': '/pages/auth/quick-login',
  '/pages/phoneLogin/phoneLogin': '/pages/auth/quick-login',
  '/pages/phoneLogin/codeLogin': '/pages/auth/quick-login',
  '/pages/verified/verified': '/pages/auth/verified',
  '/pagesSub/order/order': '/pages-sub/account/orders/orders',
  '/pagesSub/wallet/wallet': '/pages-sub/pay/wallet/wallet',
  '/pagesSub/charge/charge': '/pages-sub/pay/recharge/recharge',
  '/pagesSub/withdraw/withdraw': '/pages-sub/pay/withdraw/withdraw',
  '/pagesSub/customerService/customerService': '/pages-sub/support/customer-service/customer-service',
  '/pagesSub/accountRules/accountRules': '/pages-sub/account/billing-rules/billing-rules',
  '/pagesSub/cardCenter/cardCenter': '/pages-sub/account/card-shop/shop',
  '/pagesSub/myCard/myCard': '/pages-sub/account/cards/cards',
  '/pagesSub/myCard/myCardExpired': '/pages-sub/account/cards/cards-expired',
  '/pagesSub/rules/rules': '/pages-sub/account/card-rules/card-rules',
  '/pagesSub/voucher/voucher': '/pages-sub/account/voucher/voucher',
  '/pagesSub/msgList/msgList': '/pages-sub/support/messages/list',
  '/pagesSub/help/help': '/pages-sub/support/help/help',
  '/pagesSub/setting/setting': '/pages/account/settings',
  '/pagesSub/about/about': '/pages-sub/account/about/about',
  '/pagesSub/protocol/protocol': '/pages-sub/account/protocol/protocol',
  '/pagesSub/protocolsCustom/protocolsCustom': '/pages-sub/account/protocol/detail',
  '/pagesSub/invite/invite': '/pages-sub/account/invite/invite',
  '/pagesSub/activity/activity': '/pages-sub/account/activity/activity',
  '/pagesSub/creditScore/creditScore': '/pages-sub/account/credit/credit',
  '/pagesSub/parkSearch/parkSearch': '/pages-sub/ride/park-search/park-search',
  '/pagesSub/repair/repair': '/pages-sub/support/repair/repair',
  '/pagesSub/selectInvoiceType/selectInvoiceType': '/pages-sub/pay/invoice/list',
  '/pagesSub/issueInvoice/issueInvoice': '/pages-sub/pay/invoice/create',
  '/pagesSub/invoiceDetail/invoiceDetail': '/pages-sub/pay/invoice/apply',
  '/pagesSub/invoiceHistory/invoiceHistory': '/pages-sub/pay/invoice/history',
  '/pagesSub/invoiceHistoryDetail/invoiceHistoryDetail': '/pages-sub/pay/invoice/detail',
  '/pagesSub/accountSecurity/accountSecurity': '/pages-sub/account/security/security',
  '/pagesSub/cancelAccount/cancelAccount': '/pages-sub/account/cancel/cancel',
  '/pagesSub/costDetail/costDetail': '/pages-sub/pay/cost-detail/cost-detail',
  '/pagesSub/tripMap/tripMap': '/pages-sub/ride/trip-map/trip-map',
  '/pagesSub/objection/objection': '/pages-sub/support/objection/objection',
}

export function mapLegacyPagePath(url: string): string {
  if (!url) return url
  const [path, query = ''] = url.split('?')
  let next = LEGACY_PAGE_MAP[path] || path
  if (next.startsWith('/pagesSub/')) {
    next = next.replace(/^\/pagesSub\//, '/pages-sub/')
  }
  return query ? `${next}?${query}` : next
}
