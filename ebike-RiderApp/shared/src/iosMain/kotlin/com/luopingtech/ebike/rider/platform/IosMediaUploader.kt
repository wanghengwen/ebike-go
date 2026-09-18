package com.luopingtech.ebike.rider.platform

import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.core.i18n.Strings
import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.core.util.Ids
import com.luopingtech.ebike.rider.data.media.FileUploadApi
import kotlinx.cinterop.ExperimentalForeignApi
import kotlinx.cinterop.addressOf
import kotlinx.cinterop.useContents
import kotlinx.cinterop.usePinned
import platform.CoreGraphics.CGRectMake
import platform.CoreGraphics.CGSizeMake
import platform.Foundation.NSData
import platform.Foundation.NSURL
import platform.Foundation.dataWithContentsOfURL
import platform.UIKit.UIGraphicsBeginImageContextWithOptions
import platform.UIKit.UIGraphicsEndImageContext
import platform.UIKit.UIGraphicsGetImageFromCurrentImageContext
import platform.UIKit.UIImage
import platform.UIKit.UIImageJPEGRepresentation
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

    @OptIn(ExperimentalForeignApi::class)
    private fun readBytes(path: String): ByteArray? {
        val url = if (path.startsWith("file://")) {
            NSURL.URLWithString(path)
        } else {
            NSURL.fileURLWithPath(path)
        } ?: return null
        val raw = NSData.dataWithContentsOfURL(url) ?: return null
        val data = compressJpeg(raw)
        val length = data.length.toInt()
        if (length == 0) return null
        val bytes = ByteArray(length)
        bytes.usePinned { pinned ->
            memcpy(pinned.addressOf(0), data.bytes, data.length)
        }
        return bytes
    }

    /**
     * 与 `AndroidMediaUploader` 同参数：长边 [MAX_EDGE]、体积 [MAX_BYTES]。
     * 相机原图直传会被网关按 413 拒掉，两端必须压成一样大小，否则同一工单的图在两端清晰度不同。
     */
    @OptIn(ExperimentalForeignApi::class)
    private fun compressJpeg(raw: NSData): NSData {
        val image = UIImage.imageWithData(raw) ?: return raw
        val resized = resize(image) ?: return raw
        var quality = 0.85
        var out = UIImageJPEGRepresentation(resized, quality) ?: return raw
        while (out.length.toInt() > MAX_BYTES && quality > 0.4) {
            quality -= 0.15
            out = UIImageJPEGRepresentation(resized, quality) ?: return raw
        }
        return if (out.length < raw.length) out else raw
    }

    /** `drawInRect` 按 `UIImage.imageOrientation` 渲染，顺带把竖拍方向烧进像素。 */
    @OptIn(ExperimentalForeignApi::class)
    private fun resize(image: UIImage): UIImage? {
        val (width, height) = image.size.useContents { width to height }
        if (width <= 0.0 || height <= 0.0) return null
        val longest = maxOf(width, height)
        val ratio = if (longest <= MAX_EDGE) 1.0 else MAX_EDGE / longest
        val targetWidth = width * ratio
        val targetHeight = height * ratio
        UIGraphicsBeginImageContextWithOptions(CGSizeMake(targetWidth, targetHeight), false, 1.0)
        image.drawInRect(CGRectMake(0.0, 0.0, targetWidth, targetHeight))
        val out = UIGraphicsGetImageFromCurrentImageContext()
        UIGraphicsEndImageContext()
        return out ?: image
    }

    private fun fileNameOf(path: String): String {
        val tail = path.substringAfterLast('/').substringBefore('?')
        return if (tail.contains('.')) tail else "photo_${Ids.uuidV4()}.jpg"
    }

    private companion object {
        const val MAX_EDGE = 1600.0
        const val MAX_BYTES = 800 * 1024
    }
}
