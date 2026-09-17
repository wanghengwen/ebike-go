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

    /** Legacy ContentInputEditText password DigitsKeyListener. */
    const val ACCEPTED_CHARS = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890"

    /** Legacy RegexHelper.isValidPwd */
    private val legacyPattern = Regex("^(?![0-9]+$)(?![a-zA-Z]+$)[0-9A-Za-z]{$MIN_LENGTH,$MAX_LENGTH}$")

    fun filterInput(raw: String): String =
        raw.filter { it in ACCEPTED_CHARS }.take(MAX_LENGTH)

    /** Same as RegexHelper.isValidPwd — no trim. */
    fun isValidPwd(pass: String): Boolean = pass.isNotEmpty() && legacyPattern.matches(pass)

    fun isValid(password: String): Boolean = isValidPwd(password)

    /**
     * Legacy UpdatePwdActivity.updateBtnInfo:
     * !empty && length in 8..16 && isValidPwd (after trim).
     */
    fun canSubmitUpdatePwd(originPwd: String, newPwd: String): Boolean {
        val origin = originPwd.trim()
        val next = newPwd.trim()
        return origin.isNotEmpty() && origin.length in MIN_LENGTH..MAX_LENGTH && isValidPwd(origin) &&
            next.isNotEmpty() && next.length in MIN_LENGTH..MAX_LENGTH && isValidPwd(next)
    }

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
