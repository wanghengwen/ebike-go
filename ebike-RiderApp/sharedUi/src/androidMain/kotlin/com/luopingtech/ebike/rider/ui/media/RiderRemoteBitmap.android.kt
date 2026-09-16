package com.luopingtech.ebike.rider.ui.media

import android.graphics.BitmapFactory
import androidx.compose.ui.graphics.ImageBitmap
import androidx.compose.ui.graphics.asImageBitmap
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import kotlinx.coroutines.withContext
import java.net.HttpURLConnection
import java.net.URL
import java.util.concurrent.ConcurrentHashMap

internal actual object RiderRemoteBitmap {
    private val cache = ConcurrentHashMap<String, ImageBitmap>()
    private val mutex = Mutex()

    actual suspend fun load(url: String): ImageBitmap? {
        if (url.isBlank()) return null
        cache[url]?.let { return it }
        return mutex.withLock {
            cache[url]?.let { return it }
            val loaded = withContext(Dispatchers.IO) { fetch(url) }
            if (loaded != null) cache[url] = loaded
            loaded
        }
    }

    private fun fetch(url: String): ImageBitmap? = runCatching {
        val conn = (URL(url).openConnection() as HttpURLConnection).apply {
            connectTimeout = 8_000
            readTimeout = 8_000
            instanceFollowRedirects = true
            requestMethod = "GET"
            setRequestProperty("Accept", "image/*,*/*")
        }
        try {
            if (conn.responseCode !in 200..299) return null
            // 网络流不能直接 decodeStream（无 mark/reset，常解出 null）
            val bytes = conn.inputStream.use { it.readBytes() }
            if (bytes.isEmpty()) return null
            BitmapFactory.decodeByteArray(bytes, 0, bytes.size)?.asImageBitmap()
        } finally {
            conn.disconnect()
        }
    }.getOrNull()
}
