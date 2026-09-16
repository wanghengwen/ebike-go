package com.luopingtech.ebike.rider.platform

import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import kotlinx.cinterop.BetaInteropApi
import kotlinx.cinterop.CValue
import kotlinx.cinterop.ExperimentalForeignApi
import kotlinx.cinterop.useContents
import kotlinx.coroutines.suspendCancellableCoroutine
import platform.AVFoundation.AVCaptureVideoPreviewLayer
import platform.AVFoundation.AVLayerVideoGravityResizeAspectFill
import platform.CoreGraphics.CGRect
import platform.CoreGraphics.CGRectMake
import platform.UIKit.UIAction
import platform.UIKit.UIApplication
import platform.UIKit.UIButton
import platform.UIKit.UIButtonTypeSystem
import platform.UIKit.UIColor
import platform.UIKit.UIControlEventTouchUpInside
import platform.UIKit.UIControlStateNormal
import platform.UIKit.UIModalPresentationFullScreen
import platform.UIKit.UIViewController
import platform.UIKit.UIWindow
import kotlin.coroutines.resume

/**
 * 一次性扫码，对应 Android 的 `ActivityCodeScanner`（启 ScanQrActivity 拿结果）。
 *
 * 界面里的取景框走 `RiderScanPreview`，这里是给 feature 层用的：全屏模态 + 一个取消按钮，
 * 拿到第一个码就自己关掉。UI 刻意做到最薄，扫码的业务外壳（手输、校验、提示）
 * 都在共享层的扫码屏里，这条路径只服务「代码里想扫一次」的调用。
 */
@OptIn(ExperimentalForeignApi::class)
class IosCodeScanner(
    private val cancelTitle: String = "取消",
) : CodeScanner {

    override suspend fun scanOnce(): RiderResult<String> {
        if (!IosScanSession.requestCameraAccess()) {
            return RiderResult.Err(RiderError.unsupported("Camera permission denied"))
        }
        return suspendCancellableCoroutine { cont ->
            onMain {
                val host = topViewController()
                if (host == null) {
                    cont.resume(RiderResult.Err(RiderError.unsupported("No presenting view controller")))
                    return@onMain
                }
                val scanSession = IosScanSession()
                if (!scanSession.configure()) {
                    cont.resume(RiderResult.Err(RiderError.unsupported("Camera unavailable")))
                    return@onMain
                }
                var settled = false
                val controller = ScanModalController(scanSession, cancelTitle) { code ->
                    if (settled) return@ScanModalController
                    settled = true
                    cont.resume(
                        code?.let { RiderResult.Ok(it) }
                            ?: RiderResult.Err(RiderError.business("SCAN_CANCELLED", cancelTitle)),
                    )
                }
                controller.setModalPresentationStyle(UIModalPresentationFullScreen)
                host.presentViewController(controller, animated = true, completion = null)
            }
        }
    }
}

@OptIn(ExperimentalForeignApi::class, BetaInteropApi::class)
private class ScanModalController(
    private val scanSession: IosScanSession,
    private val cancelTitle: String,
    private val onFinish: (String?) -> Unit,
) : UIViewController(nibName = null, bundle = null) {

    private var previewLayer: AVCaptureVideoPreviewLayer? = null
    private var cancelButton: UIButton? = null

    override fun viewDidLoad() {
        super.viewDidLoad()
        view.setBackgroundColor(UIColor.blackColor)

        val layer = AVCaptureVideoPreviewLayer(session = scanSession.session)
        layer.setVideoGravity(AVLayerVideoGravityResizeAspectFill)
        view.layer.addSublayer(layer)
        previewLayer = layer

        val button = UIButton.buttonWithType(UIButtonTypeSystem)
        button.setTitle(cancelTitle, forState = UIControlStateNormal)
        button.setTitleColor(UIColor.whiteColor, forState = UIControlStateNormal)
        // UIAction 而不是 target/action：省掉一个 @ObjCAction 选择器，语义一样。
        button.addAction(
            UIAction.actionWithHandler { close(null) },
            forControlEvents = UIControlEventTouchUpInside,
        )
        view.addSubview(button)
        cancelButton = button

        scanSession.onCode = { code -> onMain { close(code) } }
    }

    override fun viewDidLayoutSubviews() {
        super.viewDidLayoutSubviews()
        previewLayer?.setFrame(view.bounds)
        cancelButton?.setFrame(cancelFrame(view.bounds))
    }

    override fun viewWillAppear(animated: Boolean) {
        super.viewWillAppear(animated)
        scanSession.start()
    }

    override fun viewWillDisappear(animated: Boolean) {
        super.viewWillDisappear(animated)
        scanSession.onCode = null
        scanSession.stop()
    }

    private fun close(code: String?) {
        scanSession.onCode = null
        dismissViewControllerAnimated(true) { onFinish(code) }
    }

    private fun cancelFrame(bounds: CValue<CGRect>): CValue<CGRect> = bounds.useContents {
        CGRectMake(x = 16.0, y = size.height - 96.0, width = 120.0, height = 44.0)
    }
}

/** 最上层可用于 present 的控制器；Compose 宿主里就是 ComposeUIViewController。 */
@OptIn(ExperimentalForeignApi::class)
internal fun topViewController(): UIViewController? {
    val application = UIApplication.sharedApplication
    val root = application.windows
        .filterIsInstance<UIWindow>()
        .firstOrNull { it.isKeyWindow() }
        ?.rootViewController
        ?: application.keyWindow?.rootViewController
        ?: return null
    var current: UIViewController = root
    while (true) {
        current = current.presentedViewController ?: return current
    }
}
