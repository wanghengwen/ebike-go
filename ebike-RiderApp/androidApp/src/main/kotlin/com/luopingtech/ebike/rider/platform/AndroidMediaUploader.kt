package com.luopingtech.ebike.rider.platform

import android.content.Context
import android.net.Uri
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.core.i18n.Strings
import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.core.util.Ids
import com.luopingtech.ebike.rider.data.media.FileUploadApi
import java.io.File

/**
 * Reads content:// / file:// URIs and uploads via [FileUploadApi].
 * Passes through already-remote http(s) URLs.
 */
class AndroidMediaUploader(
    private val context: Context,
    private val api: FileUploadApi,
) : MediaUploader {
    override suspend fun upload(localUris: List<String>): RiderResult<List<String>> {
        if (localUris.isEmpty()) {
            return RiderResult.Err(RiderError.business("MEDIA_EMPTY", Strings.t(Str.PhotoRequired)))
        }
        val out = ArrayList<String>(localUris.size)
        for (uri in localUris) {
            val trimmed = uri.trim()
            when {
                trimmed.startsWith("http://", ignoreCase = true) ||
                    trimmed.startsWith("https://", ignoreCase = true) -> out.add(trimmed)
                trimmed.startsWith("demo://") ->
                    out.add("https://demo.cdn.rider/${trimmed.removePrefix("demo://")}")
                else -> {
                    val bytes = readBytes(trimmed)
                        ?: return RiderResult.Err(
                            RiderError.business(
                                "MEDIA_READ",
                                Strings.t(Str.MediaReadFailed, trimmed),
                            ),
                        )
                    when (val uploaded = api.uploadBytes(bytes, fileNameOf(trimmed))) {
                        is RiderResult.Ok -> out.add(uploaded.value)
                        is RiderResult.Err -> return uploaded
                    }
                }
            }
        }
        return RiderResult.Ok(out)
    }

    private fun readBytes(uriString: String): ByteArray? {
        return try {
            when {
                uriString.startsWith("content://") || uriString.startsWith("file://") -> {
                    context.contentResolver.openInputStream(Uri.parse(uriString))?.use { it.readBytes() }
                }
                else -> {
                    val file = File(uriString)
                    if (file.exists()) file.readBytes() else null
                }
            }
        } catch (_: Throwable) {
            null
        }
    }

    private fun fileNameOf(uriString: String): String {
        val tail = uriString.substringAfterLast('/').substringBefore('?').ifBlank { "" }
        return if (tail.contains('.')) tail else "photo_${Ids.uuidV4()}.jpg"
    }
}
