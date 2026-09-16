package com.luopingtech.ebike.rider.core.i18n

import java.util.Locale

actual fun platformLanguageTag(): String? =
    Locale.getDefault().toLanguageTag().takeIf { it.isNotBlank() && it != "und" }
