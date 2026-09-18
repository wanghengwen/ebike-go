package com.luopingtech.ebike.rider.platform

import android.Manifest
import android.content.Intent
import android.content.pm.PackageManager
import android.net.Uri
import android.provider.MediaStore
import android.util.Log
import androidx.activity.ComponentActivity
import androidx.activity.result.contract.ActivityResultContracts
import androidx.core.content.ContextCompat
import androidx.core.content.FileProvider
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.core.i18n.Strings
import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import java.io.File
import kotlin.coroutines.cancellation.CancellationException
import kotlin.coroutines.resume
import kotlinx.coroutines.CancellableContinuation
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.suspendCancellableCoroutine
import kotlinx.coroutines.withContext

/**
 * 相机 / 相册的 Android 实现，与 [ActivityCodeScanner] 同构：
 * 三个 launcher 必须在宿主 Activity 到 CREATED 之前注册，所以要在 `onCreate` 里构造。
 *
 * 把权限询问也收在这里，调用侧（`commonMain` 里的界面）只看最终结果。
 */
class ActivityPhotoCapture(
    private val activity: ComponentActivity,
) : PhotoCapture {
    private var pending: CancellableContinuation<RiderResult<String>>? = null

    /** TakePicture 只回 true/false，拍成功后要用发起时那个 URI，所以得存着。 */
    private var pendingUri: Uri? = null

    /** 权限授予后要接着拍，前缀在等待期间不能丢。 */
    private var pendingPrefix: String = "photo"

    private val takePicture = activity.registerForActivityResult(
        ActivityResultContracts.TakePicture(),
    ) { ok ->
        val uri = pendingUri
        pendingUri = null
        Log.i(TAG, "takePicture result ok=$ok uri=$uri")
        if (ok && uri != null) {
            finish(RiderResult.Ok(uri.toString()))
        } else {
            finish(RiderResult.Err(RiderError.cancelled(Strings.t(Str.NoPhotoTaken))))
        }
    }

    private val pickImage = activity.registerForActivityResult(
        ActivityResultContracts.GetContent(),
    ) { uri ->
        Log.i(TAG, "pickImage result uri=$uri")
        if (uri != null) {
            finish(RiderResult.Ok(uri.toString()))
        } else {
            finish(RiderResult.Err(RiderError.cancelled(Strings.t(Str.NoPhotoTaken))))
        }
    }

    private val requestCamera = activity.registerForActivityResult(
        ActivityResultContracts.RequestPermission(),
    ) { granted ->
        Log.i(TAG, "camera permission granted=$granted")
        if (granted) {
            launchCamera(pendingPrefix)
        } else {
            finish(RiderResult.Err(RiderError.unsupported(Strings.t(Str.CameraPermissionRequired))))
        }
    }

    override suspend fun takePhoto(prefix: String): RiderResult<String> = withContext(Dispatchers.Main) {
        await {
            pendingPrefix = prefix
            val granted = ContextCompat.checkSelfPermission(
                activity,
                Manifest.permission.CAMERA,
            ) == PackageManager.PERMISSION_GRANTED
            if (granted) launchCamera(prefix) else requestCamera.launch(Manifest.permission.CAMERA)
        }
    }

    override suspend fun pickFromGallery(): RiderResult<String> = withContext(Dispatchers.Main) {
        await {
            pickImage.launch("image/*")
        }
    }

    private fun launchCamera(prefix: String) {
        val uri = cachePhotoUri(prefix)
        if (uri == null) {
            Log.e(TAG, "cachePhotoUri failed prefix=$prefix pkg=${activity.packageName}")
            finish(
                RiderResult.Err(
                    RiderError.business(
                        "PHOTO_FILE",
                        "无法创建照片文件（FileProvider）",
                    ),
                ),
            )
            return
        }
        val capture = Intent(MediaStore.ACTION_IMAGE_CAPTURE)
        val resolved = capture.resolveActivity(activity.packageManager)
        Log.i(TAG, "launchCamera uri=$uri resolve=$resolved")
        if (resolved == null) {
            finish(
                RiderResult.Err(
                    RiderError.business(
                        "NO_CAMERA_APP",
                        "未找到可用相机应用，请检查系统相机或 Manifest queries",
                    ),
                ),
            )
            return
        }
        pendingUri = uri
        try {
            takePicture.launch(uri)
        } catch (t: Throwable) {
            pendingUri = null
            Log.e(TAG, "takePicture.launch failed", t)
            finish(RiderResult.Err(RiderError.network(t.message ?: "takePicture.launch", t)))
        }
    }

    /** FileProvider 的 authority 与 `AndroidManifest.xml` 里声明的一致。 */
    private fun cachePhotoUri(prefix: String): Uri? = runCatching {
        val dir = File(activity.cacheDir, "photos").apply { mkdirs() }
        val file = File(dir, "${prefix}_${System.currentTimeMillis()}.jpg")
        FileProvider.getUriForFile(activity, "${activity.packageName}.fileprovider", file)
    }.onFailure { Log.e(TAG, "FileProvider failed", it) }.getOrNull()

    private fun finish(result: RiderResult<String>) {
        val cont = pending
        pending = null
        if (cont != null && cont.isActive) cont.resume(result)
    }

    private suspend fun await(start: () -> Unit): RiderResult<String> =
        suspendCancellableCoroutine { cont ->
            if (pending != null) {
                cont.resume(RiderResult.Err(RiderError.business("PHOTO_BUSY", Strings.t(Str.NoPhotoTaken))))
                return@suspendCancellableCoroutine
            }
            pending = cont
            cont.invokeOnCancellation {
                if (pending === cont) pending = null
            }
            try {
                start()
            } catch (t: Throwable) {
                pending = null
                if (t is CancellationException) throw t
                cont.resume(RiderResult.Err(RiderError.network(t.message.orEmpty(), t)))
            }
        }

    private companion object {
        const val TAG = "ActivityPhotoCapture"
    }
}
