package com.luopingtech.ebike.ops.domain.warehouse

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings

/**
 * Warehouse record/component filter: UI shows localized [Str.FilterAll],
 * while remote pageList still treats the Chinese literal as the “all” sentinel.
 */
object WarehouseComponentFilter {
    /** Server contract token — do not localize outbound. */
    const val API_ALL: String = "全部"

    fun isAll(name: String): Boolean {
        val trimmed = name.trim()
        return trimmed.isEmpty() ||
            trimmed == API_ALL ||
            trimmed == Strings.t(Str.FilterAll)
    }

    /** Map API / stored names so chips show the active locale's “All”. */
    fun toDisplayList(names: List<String>): List<String> =
        names.map { if (it.trim() == API_ALL) Strings.t(Str.FilterAll) else it }

    /** Value to put in pageList `componentName` (omit when blank). */
    fun toApiFilter(name: String): String =
        if (isAll(name)) "" else name.trim()
}
