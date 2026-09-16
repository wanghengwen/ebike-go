export type TenantConfig = {
  name: string
  alias: string
  description?: string
  introduce?: string
  platformTenantId?: string
  platformSecret?: string
  platformSign?: string
  api: {
    baseUrl: string
    logApi?: string
  }
  qrDomain?: string[]
  appId?: string
  /**
   * Auth knobs for WeChat mini-program login.
   * wechatGrantType: e.g. wechat_miniapp
   */
  auth?: {
    wechatGrantType?: string
    wechatPartnerCode?: string
  }
  /** Payment channel: BAOFU_WXLITE | WXLITE | UMS_WXLITE | UNION_WXLITE */
  pay?: {
    channelType?: string
    payBaseApi?: string
  }
  networkTimeout?: {
    request?: number
    connectSocket?: number
    uploadFile?: number
    downloadFile?: number
  }
  /**
   * Brand / skin / static feature knobs — mirrors legacy customSetting.
   * Face / civilization / helmet / ads are NOT here; they come from runtime APIs.
   */
  customSetting?: {
    buttonGreenColor?: string
    buttonWhiteColor?: string
    buttonDisabledColor?: string
    logo?: string
    launchBg?: string
    /** Find-bike bell icon (riding page) */
    carBell?: string
    /** Tenant-only static flags (e.g. bwcx) */
    autoRefundBalance?: boolean
    insufficientBalanceAutoJump?: boolean
    tempHideOrderPrice?: boolean
    hideCancelAccount?: boolean
    showRidePanelTips?: boolean
    documentCfg?: Record<string, string>
    mapCfg?: Record<string, string>
    iconCfg?: Record<string, string>
    cyclingCfg?: {
      bikeReturn?: string
      tempRecoverPowerEnable?: boolean
      tempRecoverPowerTime?: number
      tempRecoverPowerTitle?: string
      tempRecoverPowerDesc?: string
      tempRecoverPowerTips?: string
      tempRecoverPowerBtn?: string
      wearHelmetTip?: string
      [k: string]: unknown
    }
    redEnvelopeCfg?: Record<string, string>
  }
  globalStyle?: {
    navigationBarTitleText?: string
    navigationBarBackgroundColor?: string
    backgroundColor?: string
  }
  platform?: {
    'mp-weixin'?: {
      appid?: string
      pay?: { channelType?: string; payBaseApi?: string }
      plugins?: Record<string, unknown>
      setting?: Record<string, unknown>
      permission?: Record<string, unknown>
    }
    h5?: {
      router?: { mode?: string; base?: string }
      sdkConfigs?: Record<string, unknown>
      [k: string]: unknown
    }
  }
  returnReturnCarStatistics?: boolean
  /** Force BLE return/unlock path like legacy commonConfig.onlyBluetooth */
  onlyBluetooth?: boolean
  logLevel?: string | number
  appPlus?: {
    androidPackage?: string
    iosBundleId?: string
    /** Amap key for App map module (merged into manifest sdkConfigs.maps) */
    amapKey?: string
  }
}
