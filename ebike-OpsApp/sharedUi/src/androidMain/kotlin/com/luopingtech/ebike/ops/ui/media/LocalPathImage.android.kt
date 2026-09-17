package com.luopingtech.ebike.ops.ui.media

import android.graphics.BitmapFactory
import android.net.Uri
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.asImageBitmap
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import java.io.File

@Composable
actual fun LocalPathImage(
    path: String,
    modifier: Modifier,
    contentDescription: String?,
) {
    val context = LocalContext.current
    val bitmap = remember(path) {
        runCatching {
            when {
                path.startsWith("content://") -> {
                    context.contentResolver.openInputStream(Uri.parse(path))?.use {
                        BitmapFactory.decodeStream(it)
                    }
                }
                path.startsWith("file://") -> {
                    BitmapFactory.decodeFile(Uri.parse(path).path)
                }
                path.startsWith("demo://") -> null
                else -> {
                    val file = File(path)
                    if (file.exists()) BitmapFactory.decodeFile(file.absolutePath) else null
                }
            }
        }.getOrNull()
    }
    if (bitmap != null) {
        Image(
            bitmap = bitmap.asImageBitmap(),
            contentDescription = contentDescription,
            modifier = modifier,
            contentScale = ContentScale.Crop,
        )
    } else {
        Box(modifier = modifier.background(Color(0xFFE8E8E8)))
    }
}
