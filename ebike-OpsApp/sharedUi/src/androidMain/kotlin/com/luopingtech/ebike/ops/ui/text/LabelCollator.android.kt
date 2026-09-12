package com.luopingtech.ebike.ops.ui.text

import java.text.Collator
import java.util.Locale

actual fun labelComparator(): Comparator<String> {
    val collator = Collator.getInstance(Locale.CHINA)
    return Comparator { a, b -> collator.compare(a, b) }
}
