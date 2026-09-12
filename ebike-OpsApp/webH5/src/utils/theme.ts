import { reactive } from 'vue'
import { storage, StorageKey } from './storage'

const DEFAULT_THEME_COLOR = 'FF9C80'
const DEFAULT_LIGHT_TXT_COLOR = 'FFFFFF'

export interface Theme {
  themeColor: string
  lightTxtColor: string
}

export const theme = reactive<Theme>({
  themeColor: `#${DEFAULT_THEME_COLOR}`,
  lightTxtColor: `#${DEFAULT_LIGHT_TXT_COLOR}`,
})

/** App 传入的是不带 `#` 的十六进制串，容错处理一下两种写法。 */
function normalize(value: string | null | undefined, fallback: string): string {
  const hex = (value ?? '').trim().replace(/^#/, '')
  return /^[0-9a-fA-F]{3,8}$/.test(hex) ? `#${hex}` : `#${fallback}`
}

/**
 * 主题色来自 URL query（`themeColor` / `lightTxtColor`），缺省时读上一次缓存。
 * 同步写入 CSS 变量，模板里用 `var(--ops-theme-color)` 即可，无需再依赖 `$STYLE`。
 */
export function applyTheme(themeColor?: string | null, lightTxtColor?: string | null): void {
  if (themeColor) storage.set(StorageKey.ThemeColor, themeColor.replace(/^#/, ''))
  if (lightTxtColor) storage.set(StorageKey.LightTxtColor, lightTxtColor.replace(/^#/, ''))

  theme.themeColor = normalize(
    themeColor ?? storage.get<string>(StorageKey.ThemeColor),
    DEFAULT_THEME_COLOR,
  )
  theme.lightTxtColor = normalize(
    lightTxtColor ?? storage.get<string>(StorageKey.LightTxtColor),
    DEFAULT_LIGHT_TXT_COLOR,
  )

  const root = document.documentElement.style
  root.setProperty('--ops-theme-color', theme.themeColor)
  root.setProperty('--ops-light-txt-color', theme.lightTxtColor)
}
