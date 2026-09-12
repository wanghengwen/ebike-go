package com.luopingtech.ebike.ops.ui.text

import kotlinx.cinterop.BetaInteropApi
import platform.Foundation.NSString
import platform.Foundation.create
import platform.Foundation.localizedCompare

@OptIn(BetaInteropApi::class)
actual fun labelComparator(): Comparator<String> =
    Comparator { a, b -> NSString.create(string = a).localizedCompare(b).toInt() }
