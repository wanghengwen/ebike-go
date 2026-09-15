package com.luopingtech.ebike.ops.platform

import kotlinx.cinterop.ExperimentalForeignApi
import kotlinx.coroutines.suspendCancellableCoroutine
import platform.AVFoundation.AVAuthorizationStatusAuthorized
import platform.AVFoundation.AVAuthorizationStatusNotDetermined
import platform.AVFoundation.AVCaptureConnection
import platform.AVFoundation.AVCaptureDevice
import platform.AVFoundation.AVCaptureDeviceInput
import platform.AVFoundation.AVCaptureMetadataOutput
import platform.AVFoundation.AVCaptureMetadataOutputObjectsDelegateProtocol
import platform.AVFoundation.AVCaptureOutput
import platform.AVFoundation.AVCaptureSession
import platform.AVFoundation.AVCaptureSessionPresetHigh
import platform.AVFoundation.AVCaptureTorchModeOff
import platform.AVFoundation.AVCaptureTorchModeOn
import platform.AVFoundation.AVMediaTypeVideo
import platform.AVFoundation.AVMetadataMachineReadableCodeObject
import platform.AVFoundation.AVMetadataObjectTypeCode128Code
import platform.AVFoundation.AVMetadataObjectTypeCode39Code
import platform.AVFoundation.AVMetadataObjectTypeDataMatrixCode
import platform.AVFoundation.AVMetadataObjectTypeEAN13Code
import platform.AVFoundation.AVMetadataObjectTypeQRCode
import platform.AVFoundation.authorizationStatusForMediaType
import platform.AVFoundation.hasTorch
import platform.AVFoundation.isTorchModeSupported
import platform.AVFoundation.requestAccessForMediaType
import platform.AVFoundation.torchMode
import platform.darwin.NSObject
import platform.darwin.dispatch_async
import platform.darwin.dispatch_get_global_queue
import kotlin.coroutines.resume

/**
 * 一个可复用的 AVFoundation 扫码会话，对应 Android 的 CameraX + ML Kit 那一层。
 *
 * 之所以放在 `shared` 而不是 `sharedUi`：`CodeScanner`（feature 层用的一次性扫码）
 * 和 `OpsScanPreview`（界面里的取景框）要共用同一套相机配置与码制清单，
 * 差别只在谁来显示 [session] 的预览层。
 *
 * 码制跟遗留 Android 端对齐：车码 / 电池 SN 是 QR，中控 IMEI 贴纸是 Code128。
 */
@OptIn(ExperimentalForeignApi::class)
class IosScanSession {

    val session = AVCaptureSession()

    private val output = AVCaptureMetadataOutput()
    private val queue = dispatch_get_global_queue(0, 0u)
    private var device: AVCaptureDevice? = null
    private var configured = false

    /** 解码回调。取景框弹了手输框 / 正在提交时置空，避免同一个码反复回调。 */
    var onCode: ((String) -> Unit)? = null

    private val metadataDelegate = object : NSObject(), AVCaptureMetadataOutputObjectsDelegateProtocol {
        override fun captureOutput(
            output: AVCaptureOutput,
            didOutputMetadataObjects: List<*>,
            fromConnection: AVCaptureConnection,
        ) {
            val code = didOutputMetadataObjects
                .filterIsInstance<AVMetadataMachineReadableCodeObject>()
                .firstNotNullOfOrNull { it.stringValue?.takeIf(String::isNotBlank) }
                ?: return
            // 解码在后台队列上跑，消费者全是界面状态，统一在这里切回主线程。
            onMain { onCode?.invoke(code) }
        }
    }

    fun configure(): Boolean {
        if (configured) return true
        val camera = AVCaptureDevice.defaultDeviceWithMediaType(AVMediaTypeVideo) ?: return false
        val input = AVCaptureDeviceInput.deviceInputWithDevice(camera, null) ?: return false
        session.beginConfiguration()
        session.setSessionPreset(AVCaptureSessionPresetHigh)
        if (session.canAddInput(input)) session.addInput(input)
        if (session.canAddOutput(output)) {
            session.addOutput(output)
            output.setMetadataObjectsDelegate(metadataDelegate, queue)
            // metadataObjectTypes 只有在 output 已挂到 session 之后才认，顺序不能换。
            output.setMetadataObjectTypes(
                listOf(
                    AVMetadataObjectTypeQRCode,
                    AVMetadataObjectTypeCode128Code,
                    AVMetadataObjectTypeCode39Code,
                    AVMetadataObjectTypeEAN13Code,
                    AVMetadataObjectTypeDataMatrixCode,
                ),
            )
        }
        session.commitConfiguration()
        device = camera
        configured = true
        return true
    }

    fun start() {
        if (!configure()) return
        if (session.isRunning()) return
        // startRunning 是阻塞调用，放主线程会卡住首帧上屏。
        dispatch_async(queue) { session.startRunning() }
    }

    fun stop() {
        if (!session.isRunning()) return
        dispatch_async(queue) { session.stopRunning() }
    }

    fun setTorch(on: Boolean) {
        val camera = device ?: return
        if (!camera.hasTorch) return
        val mode = if (on) AVCaptureTorchModeOn else AVCaptureTorchModeOff
        if (!camera.isTorchModeSupported(mode)) return
        if (camera.lockForConfiguration(null)) {
            camera.torchMode = mode
            camera.unlockForConfiguration()
        }
    }

    companion object {
        suspend fun requestCameraAccess(): Boolean {
            return when (AVCaptureDevice.authorizationStatusForMediaType(AVMediaTypeVideo)) {
                AVAuthorizationStatusAuthorized -> true
                AVAuthorizationStatusNotDetermined -> suspendCancellableCoroutine { cont ->
                    AVCaptureDevice.requestAccessForMediaType(AVMediaTypeVideo) { granted ->
                        cont.resume(granted)
                    }
                }
                else -> false
            }
        }
    }
}
