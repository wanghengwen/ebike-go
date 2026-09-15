package com.luopingtech.ebike.ops.platform

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.core.util.Ids
import com.luopingtech.ebike.ops.data.media.FileUploadApi
import kotlinx.cinterop.ExperimentalForeignApi
import kotlinx.cinterop.addressOf
import kotlinx.cinterop.usePinned
import platform.Foundation.NSData
import platform.Foundation.NSURL
import platform.Foundation.dataWithContentsOfURL
import platform.posix.memcpy

/**
 * `AndroidMediaUploader` 的 iOS 对应实现：把本地文件读成字节交给 [FileUploadApi]。
 *
 * 传进来的路径由 [IosPhotoCapture] 产出，形如 `file:///.../Caches/ops-photos/xxx.jpg`；
 * 已经是 http(s) 的直接放行，`demo://` 走 demo CDN 前缀（与 Android 完全一致，
 * 否则同一条上报在两端会得到不同的图片地址）。
 */
class IosMediaUploader(
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

    @OptIn(ExperimentalForeignApi::class)
    private fun readBytes(path: String): ByteArray? {
        val url = if (path.startsWith("file://")) {
            NSURL.URLWithString(path)
        } else {
            NSURL.fileURLWithPath(path)
        } ?: return null
        val data = NSData.dataWithContentsOfURL(url) ?: return null
        val length = data.length.toInt()
        if (length == 0) return null
        val bytes = ByteArray(length)
        bytes.usePinned { pinned ->
            memcpy(pinned.addressOf(0), data.bytes, data.length)
        }
        return bytes
    }

    private fun fileNameOf(path: String): String {
        val tail = path.substringAfterLast('/').substringBefore('?')
        return if (tail.contains('.')) tail else "photo_${Ids.uuidV4()}.jpg"
    }
}
