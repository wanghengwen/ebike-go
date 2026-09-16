package com.luopingtech.ebike.rider.core.i18n

/**
 * Device language tag reported by the platform, e.g. `zh-Hans-CN` / `en-US`.
 *
 * Used as the first-launch default when the user has not picked a language yet.
 * Returns null when the platform cannot report one.
 */
expect fun platformLanguageTag(): String?
