package com.luopingtech.ebike.ops.platform

import android.Manifest
import android.content.pm.PackageManager
import android.net.Uri
import androidx.activity.ComponentActivity
import androidx.activity.result.contract.ActivityResultContracts
import androidx.core.content.ContextCompat
import androidx.core.content.FileProvider
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import java.io.File
import kotlin.coroutines.cancellation.CancellationException
import kotlin.coroutines.resume
import kotlinx.coroutines.CancellableContinuation
import kotlinx.coroutines.suspendCancellableCoroutine

/**
 * 相机 / 相册的 Android 实现，与 [ActivityCodeScanner] 同构：
 * 三个 launcher 必须在宿主 Activity 到 CREATED 之前注册，所以要在 `onCreate` 里构造。
 *
 * 把权限询问也收在这里，调用侧（`commonMain` 里的界面）只看最终结果。
 */
class ActivityPhotoCapture(
    private val activity: ComponentActivity,
) : PhotoCapture {
    private var pending: CancellableContinuation<OpsResult<String>>? = null

    /** TakePicture 只回 true/false，拍成功后要用发起时那个 URI，所以得存着。 */
    private var pendingUri: Uri? = null

    /** 权限授予后要接着拍，前缀在等待期间不能丢。 */
    private var pendingPrefix: String = "photo"

    private val takePicture = activity.registerForActivityResult(
        ActivityResultContracts.TakePicture(),
    ) { ok ->
        val uri = pendingUri
        pendingUri = null
        if (ok && uri != null) {
            finish(OpsResult.Ok(uri.toString()))
        } else {
            finish(OpsResult.Err(OpsError.cancelled(Strings.t(Str.NoPhotoTaken))))
        }
    }

    private val pickImage = activity.registerForActivityResult(
        ActivityResultContracts.GetContent(),
    ) { uri ->
        if (uri != null) {
            finish(OpsResult.Ok(uri.toString()))
        } else {
            finish(OpsResult.Err(OpsError.cancelled(Strings.t(Str.NoPhotoTaken))))
        }
    }

    private val requestCamera = activity.registerForActivityResult(
        ActivityResultContracts.RequestPermission(),
    ) { granted ->
        if (granted) {
            launchCamera(pendingPrefix)
        } else {
            finish(OpsResult.Err(OpsError.unsupported(Strings.t(Str.CameraPermissionRequired))))
        }
    }

    override suspend fun takePhoto(prefix: String): OpsResult<String> = await {
        pendingPrefix = prefix
        val granted = ContextCompat.checkSelfPermission(
            activity,
            Manifest.permission.CAMERA,
        ) == PackageManager.PERMISSION_GRANTED
        if (granted) launchCamera(prefix) else requestCamera.launch(Manifest.permission.CAMERA)
    }

    override suspend fun pickFromGallery(): OpsResult<String> = await {
        pickImage.launch("image/*")
    }

    private fun launchCamera(prefix: String) {
        val uri = cachePhotoUri(prefix)
        if (uri == null) {
            finish(OpsResult.Err(OpsError.business("PHOTO_FILE", Strings.t(Str.NoPhotoTaken))))
            return
        }
        pendingUri = uri
        takePicture.launch(uri)
    }

    /** FileProvider 的 authority 与 `AndroidManifest.xml` 里声明的一致。 */
    private fun cachePhotoUri(prefix: String): Uri? = runCatching {
        val dir = File(activity.cacheDir, "photos").apply { mkdirs() }
        val file = File(dir, "${prefix}_${System.currentTimeMillis()}.jpg")
        FileProvider.getUriForFile(activity, "${activity.packageName}.fileprovider", file)
    }.getOrNull()

    private fun finish(result: OpsResult<String>) {
        val cont = pending
        pending = null
        if (cont != null && cont.isActive) cont.resume(result)
    }

    private suspend fun await(start: () -> Unit): OpsResult<String> =
        suspendCancellableCoroutine { cont ->
            if (pending != null) {
                cont.resume(OpsResult.Err(OpsError.business("PHOTO_BUSY", Strings.t(Str.NoPhotoTaken))))
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
                cont.resume(OpsResult.Err(OpsError.network(t.message.orEmpty(), t)))
            }
        }
}
