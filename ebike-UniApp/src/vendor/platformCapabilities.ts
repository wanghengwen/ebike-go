/**
 * Platform capability matrix for app-plus / mp-weixin / h5.
 * Used as a single place to document native wiring progress.
 */
export const platformCapabilities = {
  location: true,
  scan: true,
  ble: true,
  wxpay: true,
  alipay: true,
  map: true,
} as const
