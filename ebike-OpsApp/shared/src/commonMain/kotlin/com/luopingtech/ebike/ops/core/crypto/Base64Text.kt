package com.luopingtech.ebike.ops.core.crypto

import kotlin.io.encoding.Base64
import kotlin.io.encoding.ExperimentalEncodingApi

@OptIn(ExperimentalEncodingApi::class)
object Base64Text {
    fun encodeUtf8(input: String): String =
        Base64.Default.encode(input.encodeToByteArray())
}
