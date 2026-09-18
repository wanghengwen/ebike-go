package com.luopingtech.ebike.rider.pay

import android.content.Context
import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.data.pay.WeChatPayParams
import com.luopingtech.ebike.rider.platform.WeChatPayLauncher
import com.tencent.mm.opensdk.constants.ConstantsAPI
import com.tencent.mm.opensdk.modelbase.BaseResp
import com.tencent.mm.opensdk.modelpay.PayReq
import com.tencent.mm.opensdk.openapi.IWXAPI
import com.tencent.mm.opensdk.openapi.WXAPIFactory
import kotlinx.coroutines.CancellableContinuation
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.suspendCancellableCoroutine
import kotlinx.coroutines.withContext
import kotlin.concurrent.Volatile
import kotlin.coroutines.resume

/**
 * 微信开放平台回调桥。WXPayEntryActivity 与 [AndroidWeChatPay] 共用。
 */
object WeChatPayBridge {
    @Volatile
    var appId: String = ""
        private set

    @Volatile
    private var pending: CancellableContinuation<RiderResult<Unit>>? = null

    fun rememberAppId(id: String) {
        if (id.isNotBlank()) appId = id
    }

    fun createApi(context: Context): IWXAPI {
        val api = WXAPIFactory.createWXAPI(context.applicationContext, appId, false)
        if (appId.isNotBlank()) {
            api.registerApp(appId)
        }
        return api
    }

    fun complete(resp: BaseResp) {
        if (resp.type != ConstantsAPI.COMMAND_PAY_BY_WX) return
        val cont = pending ?: return
        pending = null
        if (!cont.isActive) return
        val result = when (resp.errCode) {
            BaseResp.ErrCode.ERR_OK -> RiderResult.Ok(Unit)
            BaseResp.ErrCode.ERR_USER_CANCEL ->
                RiderResult.Err(RiderError.cancelled(resp.errStr.orEmpty().ifBlank { "cancelled" }))
            else -> RiderResult.Err(
                RiderError.business(
                    code = resp.errCode.toString(),
                    message = resp.errStr.orEmpty().ifBlank { "wechat pay failed" },
                ),
            )
        }
        cont.resume(result)
    }

    suspend fun awaitPay(send: () -> Boolean): RiderResult<Unit> =
        suspendCancellableCoroutine { cont ->
            pending?.cancel()
            pending = cont
            cont.invokeOnCancellation {
                if (pending === cont) pending = null
            }
            val sent = try {
                send()
            } catch (t: Throwable) {
                if (t is CancellationException) throw t
                pending = null
                cont.resume(RiderResult.Err(RiderError.network(t.message ?: "sendReq", t)))
                return@suspendCancellableCoroutine
            }
            if (!sent) {
                pending = null
                cont.resume(RiderResult.Err(RiderError.business("SEND_FAIL", "wechat sendReq failed")))
            }
        }
}

class AndroidWeChatPay(
    context: Context,
    private val appId: () -> String,
) : WeChatPayLauncher {
    private val appContext = context.applicationContext

    override suspend fun pay(params: WeChatPayParams): RiderResult<Unit> = withContext(Dispatchers.Main) {
        val id = params.appId.ifBlank { appId() }
        if (id.isBlank()) {
            return@withContext RiderResult.Err(RiderError.unsupported("WeChatPay appId"))
        }
        WeChatPayBridge.rememberAppId(id)
        val api = WeChatPayBridge.createApi(appContext)
        if (!api.isWXAppInstalled) {
            return@withContext RiderResult.Err(RiderError.unsupported("WeChat not installed"))
        }
        val req = PayReq().apply {
            appId = id
            partnerId = params.partnerId
            prepayId = params.prepayId
            packageValue = params.packageValue.ifBlank { "Sign=WXPay" }
            nonceStr = params.nonceStr
            timeStamp = params.timeStamp
            sign = params.sign
        }
        WeChatPayBridge.awaitPay { api.sendReq(req) }
    }
}
