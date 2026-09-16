package com.luopingtech.ebike.rider.ui.media

import androidx.compose.foundation.Image
import androidx.compose.foundation.layout.Box
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.ImageBitmap
import androidx.compose.ui.layout.ContentScale

/**
 * 远程 PNG/JPEG。对齐 UniApp `<image :src="url">`，首页快捷入口 / 扫码图标用。
 * 解码走各端 expect（Android BitmapFactory / iOS Skia）。
 */
@Composable
fun RiderNetworkImage(
    url: String,
    contentDescription: String?,
    modifier: Modifier = Modifier,
    contentScale: ContentScale = ContentScale.Fit,
) {
    var bitmap by remember(url) { mutableStateOf<ImageBitmap?>(null) }
    LaunchedEffect(url) {
        bitmap = if (url.isBlank()) null else RiderRemoteBitmap.load(url)
    }
    val bmp = bitmap
    if (bmp != null) {
        Image(
            bitmap = bmp,
            contentDescription = contentDescription,
            modifier = modifier,
            contentScale = contentScale,
        )
    } else {
        Box(modifier = modifier)
    }
}

internal expect object RiderRemoteBitmap {
    suspend fun load(url: String): ImageBitmap?
}
