package com.luopingtech.ebike.rider.wxapi

import android.app.Activity
import android.content.Intent
import android.os.Bundle
import com.luopingtech.ebike.rider.pay.WeChatPayBridge
import com.tencent.mm.opensdk.modelbase.BaseReq
import com.tencent.mm.opensdk.modelbase.BaseResp
import com.tencent.mm.opensdk.openapi.IWXAPIEventHandler

/**
 * 微信要求回调类名为 `{applicationId}.wxapi.WXPayEntryActivity`。
 * 源码包名固定为 namespace `com.luopingtech.ebike.rider.wxapi`；租户 applicationId
 * 与 namespace 不一致时，开放平台登记的包名需与此一致，否则支付结果回不来（下单仍可走）。
 */
class WXPayEntryActivity : Activity(), IWXAPIEventHandler {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        handle(intent)
    }

    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        setIntent(intent)
        handle(intent)
    }

    private fun handle(intent: Intent?) {
        val api = WeChatPayBridge.createApi(this)
        api.handleIntent(intent, this)
    }

    override fun onReq(req: BaseReq?) = Unit

    override fun onResp(resp: BaseResp?) {
        if (resp != null) {
            WeChatPayBridge.complete(resp)
        }
        finish()
    }
}
