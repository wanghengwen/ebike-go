package com.luopingtech.ebike.rider.core.i18n

import platform.Foundation.NSLocale
import platform.Foundation.preferredLanguages

actual fun platformLanguageTag(): String? =
    NSLocale.preferredLanguages.firstOrNull() as? String
