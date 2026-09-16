package com.luopingtech.ebike.rider.platform

import android.app.Activity
import android.content.Intent
import androidx.activity.ComponentActivity
import androidx.activity.result.contract.ActivityResultContracts
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.core.i18n.Strings
import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.scan.ScanQrActivity
import kotlin.coroutines.cancellation.CancellationException
import kotlinx.coroutines.suspendCancellableCoroutine
import kotlin.coroutines.resume

/**
 * Launches [ScanQrActivity] once and resumes with the raw QR / barcode payload.
 * Must be constructed while the host [ComponentActivity] is at least CREATED.
 */
class ActivityCodeScanner(
    private val activity: ComponentActivity,
) : CodeScanner {
    private var pending: kotlinx.coroutines.CancellableContinuation<RiderResult<String>>? = null

    private val launcher = activity.registerForActivityResult(
        ActivityResultContracts.StartActivityForResult(),
    ) { result ->
        val cont = pending
        pending = null
        if (cont == null || !cont.isActive) return@registerForActivityResult
        when (result.resultCode) {
            Activity.RESULT_OK -> {
                val raw = result.data?.getStringExtra(ScanQrActivity.EXTRA_RAW).orEmpty().trim()
                if (raw.isEmpty()) {
                    cont.resume(
                        RiderResult.Err(RiderError.business("SCAN_EMPTY", Strings.t(Str.ScanEmpty))),
                    )
                } else {
                    cont.resume(RiderResult.Ok(raw))
                }
            }
            Activity.RESULT_CANCELED -> cont.resume(
                RiderResult.Err(RiderError.cancelled(Strings.t(Str.ScanCancelled))),
            )
            else -> cont.resume(
                RiderResult.Err(RiderError.business("SCAN_FAILED", Strings.t(Str.ScanFailed))),
            )
        }
    }

    override suspend fun scanOnce(): RiderResult<String> =
        suspendCancellableCoroutine { cont ->
            if (pending != null) {
                cont.resume(
                    RiderResult.Err(RiderError.business("SCAN_BUSY", Strings.t(Str.ScanBusy))),
                )
                return@suspendCancellableCoroutine
            }
            pending = cont
            cont.invokeOnCancellation {
                if (pending === cont) pending = null
            }
            try {
                launcher.launch(Intent(activity, ScanQrActivity::class.java))
            } catch (t: Throwable) {
                pending = null
                if (t is CancellationException) throw t
                cont.resume(
                    RiderResult.Err(
                        RiderError.network(Strings.t(Str.ScanCannotStart, t.message), t),
                    ),
                )
            }
        }
}
