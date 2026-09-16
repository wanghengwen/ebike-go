package com.luopingtech.ebike.rider.ui.media

import androidx.compose.ui.graphics.ImageBitmap
import androidx.compose.ui.graphics.toComposeImageBitmap
import kotlinx.cinterop.ExperimentalForeignApi
import kotlinx.cinterop.addressOf
import kotlinx.cinterop.usePinned
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import org.jetbrains.skia.Image
import platform.Foundation.NSData
import platform.Foundation.NSURL
import platform.Foundation.dataWithContentsOfURL
import platform.posix.memcpy

internal actual object RiderRemoteBitmap {
    @OptIn(ExperimentalForeignApi::class)
    actual suspend fun load(url: String): ImageBitmap? = withContext(Dispatchers.Default) {
        runCatching {
            val nsUrl = NSURL.URLWithString(url) ?: return@runCatching null
            val data = NSData.dataWithContentsOfURL(nsUrl) ?: return@runCatching null
            val bytes = ByteArray(data.length.toInt())
            bytes.usePinned { pinned ->
                memcpy(pinned.addressOf(0), data.bytes, data.length)
            }
            Image.makeFromEncoded(bytes).toComposeImageBitmap()
        }.getOrNull()
    }
}
