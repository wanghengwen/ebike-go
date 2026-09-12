package com.luopingtech.ebike.ops.domain.order

/**
 * 搜索框输入分类。优先级与遗留 / H5 一致：手机号 → IMEI → 车号 → 姓名。
 */
object OrderSearchClassifier {
    /** 大陆号段，逐字符对齐遗留 `RegexUtil.CHINA_PATTERN`。 */
    private val chinaPhone =
        Regex("^((13[0-9])|(14[0,1,4-9])|(15[0-3,5-9])|(16[2,5,6,7])|(17[0-8])|(18[0-9])|(19[0-3,5-9]))\\d{8}$")
    private val hkPhone = Regex("^(5|6|8|9)\\d{7}$")
    private val digits = Regex("^[0-9]+$")

    fun isPhone(value: String): Boolean =
        chinaPhone.matches(value) || hkPhone.matches(value)

    fun isImei(value: String): Boolean =
        digits.matches(value) && value.length == 15

    /** 9 位纯数字，或 `08` 开头的 7/8 位（修了遗留 `matches` 死代码）。 */
    fun isCarId(value: String): Boolean {
        if (!digits.matches(value)) return false
        if (value.length == 9) return true
        return value.startsWith("08") && (value.length == 7 || value.length == 8)
    }

    fun classify(raw: String): OrderSearchKind {
        val value = raw.filterNot { it.isWhitespace() }
        return when {
            isPhone(value) -> OrderSearchKind.Phone
            isImei(value) -> OrderSearchKind.Imei
            isCarId(value) -> OrderSearchKind.CarId
            else -> OrderSearchKind.Name
        }
    }
}
