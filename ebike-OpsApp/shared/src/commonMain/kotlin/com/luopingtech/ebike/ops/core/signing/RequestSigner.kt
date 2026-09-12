package com.luopingtech.ebike.ops.core.signing

import com.luopingtech.ebike.ops.core.crypto.Sha256

/**
 * Request signing aligned with the legacy merchant app:
 *
 * - GET / form: sorted `k=v&` pairs + `_t={timestamp}{signSecret}`, then SHA-256 hex
 * - POST JSON: `{body}_t={timestamp}{signSecret}`, then SHA-256 hex
 *
 * Header names match the legacy interceptor: `_t`, `_s`, `Authorization`.
 */
object RequestSigner {
    const val HEADER_TIMESTAMP = "_t"
    const val HEADER_SIGN = "_s"
    const val HEADER_AUTHORIZATION = "Authorization"

    fun signGet(queryParams: Map<String, String>, timestamp: String, signSecret: String): String {
        val sorted = queryParams.entries.sortedWith(
            compareBy<Map.Entry<String, String>> { it.key }.thenBy { it.value },
        )
        val builder = StringBuilder()
        sorted.forEach { (key, value) ->
            builder.append(key).append('=').append(value).append('&')
        }
        builder.append("_t=").append(timestamp).append(signSecret)
        return Sha256.hex(builder.toString())
    }

    fun signPostJson(body: String, timestamp: String, signSecret: String): String {
        val payload = body + "_t=" + timestamp + signSecret
        return Sha256.hex(payload)
    }

    /**
     * Form-style sign used by legacy `PostSignProvider.provideSign(Map, time)` for
     * tenant/queryList: sorted `k=v&` + `time={ts}&secret={secret}`, then SHA-256.
     */
    fun signFormParams(
        params: Map<String, String>,
        timestamp: String,
        secret: String,
    ): String {
        val builder = StringBuilder()
        params.entries.sortedBy { it.key }.forEach { (key, value) ->
            builder.append(key).append('=').append(value).append('&')
        }
        builder.append("time=").append(timestamp).append("&secret=").append(secret)
        return Sha256.hex(builder.toString())
    }
}
