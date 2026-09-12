package com.luopingtech.ebike.ops.domain.auth

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings

/**
 * Legacy RegexHelper.isValidPwd:
 * `^(?![0-9]+$)(?![a-zA-Z]+$)[0-9A-Za-z]{8,16}$`
 */
object PasswordRules {
    const val MIN_LENGTH = 8
    const val MAX_LENGTH = 16

    private val legacyPattern = Regex("^(?![0-9]+$)(?![a-zA-Z]+$)[0-9A-Za-z]{$MIN_LENGTH,$MAX_LENGTH}$")

    fun isValid(password: String): Boolean = validationError(password) == null

    fun validationError(password: String): String? {
        val trimmed = password.trim()
        return when {
            trimmed.length < MIN_LENGTH -> Strings.t(Str.PasswordTooShort)
            trimmed.length > MAX_LENGTH -> Strings.t(Str.PasswordTooLong)
            !legacyPattern.matches(trimmed) -> Strings.t(Str.PasswordNeedLettersAndDigits)
            else -> null
        }
    }
}
