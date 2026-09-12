package com.luopingtech.ebike.ops.platform

import android.content.Context
import android.net.Uri
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.core.util.Ids
import com.luopingtech.ebike.ops.data.media.FileUploadApi
import java.io.File

/**
 * Reads content:// / file:// URIs and uploads via [FileUploadApi].
 * Passes through already-remote http(s) URLs.
 */
class AndroidMediaUploader(
    private val context: Context,
    private val api: FileUploadApi,
) : MediaUploader {
    override suspend fun upload(localUris: List<String>): OpsResult<List<String>> {
        if (localUris.isEmpty()) {
            return OpsResult.Err(OpsError.business("MEDIA_EMPTY", Strings.t(Str.PhotoRequired)))
        }
        val out = ArrayList<String>(localUris.size)
        for (uri in localUris) {
            val trimmed = uri.trim()
            when {
                trimmed.startsWith("http://", ignoreCase = true) ||
                    trimmed.startsWith("https://", ignoreCase = true) -> out.add(trimmed)
                trimmed.startsWith("demo://") ->
                    out.add("https://demo.cdn.ops/${trimmed.removePrefix("demo://")}")
                else -> {
                    val bytes = readBytes(trimmed)
                        ?: return OpsResult.Err(
                            OpsError.business(
                                "MEDIA_READ",
                                Strings.t(Str.MediaReadFailed, trimmed),
                            ),
                        )
                    when (val uploaded = api.uploadBytes(bytes, fileNameOf(trimmed))) {
                        is OpsResult.Ok -> out.add(uploaded.value)
                        is OpsResult.Err -> return uploaded
                    }
                }
            }
        }
        return OpsResult.Ok(out)
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
