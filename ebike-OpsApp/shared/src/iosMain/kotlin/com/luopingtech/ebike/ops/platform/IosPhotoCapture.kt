package com.luopingtech.ebike.ops.platform

import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import kotlinx.cinterop.BetaInteropApi
import kotlinx.cinterop.ExperimentalForeignApi
import kotlinx.cinterop.ObjCSignatureOverride
import kotlinx.coroutines.suspendCancellableCoroutine
import platform.Foundation.NSCachesDirectory
import platform.Foundation.NSDate
import platform.Foundation.NSSearchPathForDirectoriesInDomains
import platform.Foundation.NSUserDomainMask
import platform.Foundation.timeIntervalSince1970
import platform.Foundation.writeToFile
import platform.UIKit.UIImage
import platform.UIKit.UIImageJPEGRepresentation
import platform.UIKit.UIImagePickerController
import platform.UIKit.UIImagePickerControllerDelegateProtocol
import platform.UIKit.UIImagePickerControllerOriginalImage
import platform.UIKit.UIImagePickerControllerSourceType
import platform.UIKit.UIModalPresentationFullScreen
import platform.UIKit.UINavigationControllerDelegateProtocol
import platform.darwin.NSObject
import kotlin.coroutines.resume

/**
 * 拍照 / 相册，对应 Android 的 `ActivityPhotoCapture`。
 *
 * 返回值跟 Android 保持同一套字符串契约：本地文件路径，交给 `MediaUploader` 上传。
 * iOS 这边把 `UIImage` 落成 caches 里的 JPEG，故障上报那类界面拿到的就是普通路径，
 * 不需要知道背后是 `content://` 还是 `file://`。
 */
@OptIn(ExperimentalForeignApi::class)
class IosPhotoCapture(
    private val jpegQuality: Double = 0.85,
) : PhotoCapture {

    override suspend fun takePhoto(prefix: String): OpsResult<String> {
        val camera = UIImagePickerControllerSourceType.UIImagePickerControllerSourceTypeCamera
        if (!UIImagePickerController.isSourceTypeAvailable(camera)) {
            return OpsResult.Err(OpsError.unsupported("Camera unavailable"))
        }
        if (!IosScanSession.requestCameraAccess()) {
            return OpsResult.Err(OpsError.unsupported("Camera permission denied"))
        }
        return present(camera, prefix)
    }

    override suspend fun pickFromGallery(): OpsResult<String> = present(
        UIImagePickerControllerSourceType.UIImagePickerControllerSourceTypePhotoLibrary,
        "gallery",
    )

    private suspend fun present(
        source: UIImagePickerControllerSourceType,
        prefix: String,
    ): OpsResult<String> = suspendCancellableCoroutine { cont ->
        onMain {
            val host = topViewController()
            if (host == null) {
                cont.resume(OpsResult.Err(OpsError.unsupported("No presenting view controller")))
                return@onMain
            }
            val picker = UIImagePickerController()
            picker.setSourceType(source)
            var settled = false
            // delegate 必须被强引用，UIImagePickerController 只弱持有它。
            val handler = PickerDelegate { image ->
                if (settled) return@PickerDelegate
                settled = true
                val result = image
                    ?.let { saveJpeg(it, prefix) }
                    ?.let { OpsResult.Ok(it) }
                    ?: OpsResult.Err(OpsError.business("PHOTO_CANCELLED", "cancelled"))
                cont.resume(result)
            }
            picker.setDelegate(handler)
            retainedDelegates += handler
            picker.setModalPresentationStyle(UIModalPresentationFullScreen)
            host.presentViewController(picker, animated = true, completion = null)
        }
    }

    private fun saveJpeg(image: UIImage, prefix: String): String? {
        val data = UIImageJPEGRepresentation(image, jpegQuality) ?: return null
        val caches = NSSearchPathForDirectoriesInDomains(
            NSCachesDirectory,
            NSUserDomainMask,
            true,
        ).firstOrNull() as? String ?: return null
        val stamp = (NSDate().timeIntervalSince1970 * 1000).toLong()
        val path = "$caches/$prefix-$stamp.jpg"
        return if (data.writeToFile(path, atomically = true)) path else null
    }

    private companion object {
        /** 图片选择器结束前 delegate 不能被回收，选完就移除。 */
        val retainedDelegates = mutableListOf<NSObject>()
    }
}

@OptIn(ExperimentalForeignApi::class, BetaInteropApi::class)
private class PickerDelegate(
    private val onResult: (UIImage?) -> Unit,
) : NSObject(), UIImagePickerControllerDelegateProtocol, UINavigationControllerDelegateProtocol {

    @ObjCSignatureOverride
    override fun imagePickerController(
        picker: UIImagePickerController,
        didFinishPickingMediaWithInfo: Map<Any?, *>,
    ) {
        val image = didFinishPickingMediaWithInfo[UIImagePickerControllerOriginalImage] as? UIImage
        picker.dismissViewControllerAnimated(true) { onResult(image) }
    }

    override fun imagePickerControllerDidCancel(picker: UIImagePickerController) {
        picker.dismissViewControllerAnimated(true) { onResult(null) }
    }
}
