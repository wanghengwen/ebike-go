package com.luopingtech.ebike.ops.ui.common

import android.content.Context
import android.net.Uri
import androidx.core.content.FileProvider
import java.io.File

/** Cache-backed FileProvider URI for TakePicture contracts. */
fun createCachePhotoUri(context: Context, prefix: String = "photo"): Uri? {
    return runCatching {
        val dir = File(context.cacheDir, "photos").apply { mkdirs() }
        val file = File(dir, "${prefix}_${System.currentTimeMillis()}.jpg")
        FileProvider.getUriForFile(
            context,
            "${context.packageName}.fileprovider",
            file,
        )
    }.getOrNull()
}
