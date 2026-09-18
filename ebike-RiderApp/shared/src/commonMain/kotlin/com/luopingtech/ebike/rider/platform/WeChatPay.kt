package com.luopingtech.ebike.rider.platform

import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.data.pay.WeChatPayParams
import kotlin.concurrent.Volatile

/**
 * 调起微信 APP 支付。凭证未配、未绑定宿主、或参数是小程序五字段时，调用方回退 H5。
 */
fun interface WeChatPayLauncher {
    suspend fun pay(params: WeChatPayParams): RiderResult<Unit>
}

object UnsupportedWeChatPay : WeChatPayLauncher {
    override suspend fun pay(params: WeChatPayParams): RiderResult<Unit> =
        RiderResult.Err(RiderError.unsupported("WeChatPay"))
}

class BindableWeChatPay : WeChatPayLauncher {
    @Volatile
    private var bound: WeChatPayLauncher? = null

    fun bind(launcher: WeChatPayLauncher) {
        bound = launcher
    }

    fun unbind(launcher: WeChatPayLauncher? = null) {
        if (launcher == null || bound === launcher) {
            bound = null
        }
    }

    val isBound: Boolean get() = bound != null

    override suspend fun pay(params: WeChatPayParams): RiderResult<Unit> {
        val current = bound
        return current?.pay(params)
            ?: RiderResult.Err(RiderError.unsupported("WeChatPay not bound"))
    }
}
