package com.luopingtech.ebike.rider.platform

import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult

/**
 * 运营商一键登录。外部备案未完成，P1 只留接口 + 默认不支持实现，不要接真实 SDK。
 */
interface QuickLoginProvider {
    val isAvailable: Boolean
    suspend fun requestToken(): RiderResult<String>
}

object UnsupportedQuickLogin : QuickLoginProvider {
    override val isAvailable: Boolean = false

    override suspend fun requestToken(): RiderResult<String> =
        RiderResult.Err(RiderError.unsupported("quick-login"))
}
