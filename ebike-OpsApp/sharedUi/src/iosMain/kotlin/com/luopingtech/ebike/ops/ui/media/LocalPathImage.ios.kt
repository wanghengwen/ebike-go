package com.luopingtech.ebike.ops.ui.media

import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.ImageBitmap
import androidx.compose.ui.graphics.painter.BitmapPainter
import androidx.compose.ui.graphics.toComposeImageBitmap
import androidx.compose.ui.layout.ContentScale
import kotlinx.cinterop.ExperimentalForeignApi
import kotlinx.cinterop.addressOf
import kotlinx.cinterop.usePinned
import org.jetbrains.skia.Image as SkiaImage
import platform.Foundation.NSData
import platform.Foundation.dataWithContentsOfFile
import platform.posix.memcpy

@Composable
actual fun LocalPathImage(
    path: String,
    modifier: Modifier,
    contentDescription: String?,
) {
    val bitmap = remember(path) { loadLocalBitmap(path) }
    if (bitmap != null) {
        Image(
            painter = BitmapPainter(bitmap),
            contentDescription = contentDescription,
            modifier = modifier,
            contentScale = ContentScale.Crop,
        )
    } else {
        Box(modifier = modifier.background(Color(0xFFE8E8E8)))
    }
}

@OptIn(ExperimentalForeignApi::class)
private fun loadLocalBitmap(path: String): ImageBitmap? {
    if (path.startsWith("demo://")) return null
    val filePath = when {
        path.startsWith("file://") -> path.removePrefix("file://")
        else -> path
    }
    val data = NSData.dataWithContentsOfFile(filePath) ?: return null
    val size = data.length.toInt()
    if (size == 0) return null
    val bytes = ByteArray(size)
    val source = data.bytes ?: return null
    bytes.usePinned { pinned -> memcpy(pinned.addressOf(0), source, data.length) }
    return runCatching { SkiaImage.makeFromEncoded(bytes).toComposeImageBitmap() }.getOrNull()
}
