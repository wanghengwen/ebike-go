package com.luopingtech.ebike.rider.core.network

/**
 * C 端业务码，对齐 UniApp `shared/request.ts` 的 WHITE_CODES：
 *
 * - `00005` / `00013`：排队刷新一次再重试
 * - `00015`：强制重登
 * - `00006`：游客命中不清 session；只有当时已有 token 才清
 */
object ApiCodes {
    const val SUCCESS = "200"

    /** Access token expired — refresh then retry. */
    const val TOKEN_EXPIRED = "00005"

    /** Authorization missing / invalid. Guest hits must not clear a missing session. */
    const val ACCESS_UNAUTHORIZED = "00006"

    /** Refresh token failed — force re-login. */
    const val REFRESH_TOKEN_ERROR = "00012"

    /** Token invalid — refresh once (UniApp queues 00013 with 00005). */
    const val TOKEN_INVALID = "00013"

    /** Logged in elsewhere. */
    const val OTHER_DEVICE_LOGIN = "00015"

    /** Auth failed. */
    const val AUTHENTICATION_FAILED = "00009"

    /** Permissions changed. */
    const val PERMISSION_CHANGED = "00027"

    fun shouldRefresh(code: String): Boolean =
        code == TOKEN_EXPIRED || code == TOKEN_INVALID

    fun isGuestUnauthorized(code: String): Boolean = code == ACCESS_UNAUTHORIZED

    fun shouldForceRelogin(code: String): Boolean = code in setOf(
        OTHER_DEVICE_LOGIN,
        AUTHENTICATION_FAILED,
        PERMISSION_CHANGED,
        REFRESH_TOKEN_ERROR,
    )
}
