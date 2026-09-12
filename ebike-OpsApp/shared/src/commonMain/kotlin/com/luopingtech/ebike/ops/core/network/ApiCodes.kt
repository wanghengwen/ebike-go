package com.luopingtech.ebike.ops.core.network

/**
 * Business codes from the legacy merchant API that affect auth session handling.
 */
object ApiCodes {
    const val SUCCESS = "200"

    /** Access token expired — refresh then retry. */
    const val TOKEN_EXPIRED = "00005"

    /** Authorization missing / invalid shape — treat like expired for refresh once. */
    const val ACCESS_UNAUTHORIZED = "00006"

    /** Refresh token failed — force re-login. */
    const val REFRESH_TOKEN_ERROR = "00012"

    /** Token invalid — force re-login. */
    const val TOKEN_INVALID = "00013"

    /** Logged in elsewhere. */
    const val OTHER_DEVICE_LOGIN = "00015"

    /** Auth failed. */
    const val AUTHENTICATION_FAILED = "00009"

    /** Permissions changed. */
    const val PERMISSION_CHANGED = "00027"

    fun shouldRefresh(code: String): Boolean =
        code == TOKEN_EXPIRED || code == ACCESS_UNAUTHORIZED

    fun shouldForceRelogin(code: String): Boolean = code in setOf(
        TOKEN_INVALID,
        OTHER_DEVICE_LOGIN,
        AUTHENTICATION_FAILED,
        PERMISSION_CHANGED,
        REFRESH_TOKEN_ERROR,
    )
}
