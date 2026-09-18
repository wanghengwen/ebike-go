package com.luopingtech.ebike.rider.platform

import android.content.Context
import android.graphics.Bitmap
import android.graphics.BitmapFactory
import android.graphics.Matrix
import android.media.ExifInterface
import android.net.Uri
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.core.i18n.Strings
import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.core.util.Ids
import com.luopingtech.ebike.rider.data.media.FileUploadApi
import java.io.ByteArrayInputStream
import java.io.ByteArrayOutputStream
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
                    when (val uploaded = api.uploadBytes(compressJpeg(bytes), fileNameOf(trimmed))) {
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

    /**
     * 相机原图动辄 3~8MB，网关（nginx client_max_body_size）会直接回 413。
     * 长边压到 [MAX_EDGE]、体积压到 [MAX_BYTES] 以内；解码失败或原图更小就原样上传。
     */
    private fun compressJpeg(raw: ByteArray): ByteArray = try {
        val bounds = BitmapFactory.Options().apply { inJustDecodeBounds = true }
        BitmapFactory.decodeByteArray(raw, 0, raw.size, bounds)
        val longest = maxOf(bounds.outWidth, bounds.outHeight)
        if (longest <= 0) {
            raw
        } else {
            var sample = 1
            while (longest / sample > MAX_EDGE * 2) sample *= 2
            val decoded = BitmapFactory.decodeByteArray(
                raw,
                0,
                raw.size,
                BitmapFactory.Options().apply { inSampleSize = sample },
            )
            if (decoded == null) {
                raw
            } else {
                // 重新编码会丢 EXIF，先把方向烧进像素，否则竖拍照片上传后是躺着的。
                val upright = applyExifOrientation(scaleDown(decoded), raw)
                var quality = 85
                var out = encode(upright, quality)
                while (out.size > MAX_BYTES && quality > 40) {
                    quality -= 15
                    out = encode(upright, quality)
                }
                upright.recycle()
                if (out.size < raw.size) out else raw
            }
        }
    } catch (_: Throwable) {
        // OutOfMemoryError 也在内：压不动就交给服务端判死，别让上传直接崩。
        raw
    }

    private fun scaleDown(source: Bitmap): Bitmap {
        val longest = maxOf(source.width, source.height)
        if (longest <= MAX_EDGE) return source
        val ratio = MAX_EDGE.toFloat() / longest
        val scaled = Bitmap.createScaledBitmap(
            source,
            (source.width * ratio).toInt().coerceAtLeast(1),
            (source.height * ratio).toInt().coerceAtLeast(1),
            true,
        )
        if (scaled !== source) source.recycle()
        return scaled
    }

    private fun applyExifOrientation(source: Bitmap, raw: ByteArray): Bitmap {
        val degrees = try {
            val exif = ExifInterface(ByteArrayInputStream(raw))
            when (exif.getAttributeInt(ExifInterface.TAG_ORIENTATION, ExifInterface.ORIENTATION_NORMAL)) {
                ExifInterface.ORIENTATION_ROTATE_90 -> 90f
                ExifInterface.ORIENTATION_ROTATE_180 -> 180f
                ExifInterface.ORIENTATION_ROTATE_270 -> 270f
                else -> 0f
            }
        } catch (_: Throwable) {
            0f
        }
        if (degrees == 0f) return source
        val matrix = Matrix().apply { postRotate(degrees) }
        val rotated = Bitmap.createBitmap(source, 0, 0, source.width, source.height, matrix, true)
        if (rotated !== source) source.recycle()
        return rotated
    }

    private fun encode(bitmap: Bitmap, quality: Int): ByteArray =
        ByteArrayOutputStream().use { sink ->
            bitmap.compress(Bitmap.CompressFormat.JPEG, quality, sink)
            sink.toByteArray()
        }

    private companion object {
        const val MAX_EDGE = 1600
        const val MAX_BYTES = 800 * 1024
    }
}
