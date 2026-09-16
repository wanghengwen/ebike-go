package com.luopingtech.ebike.rider.scan

import android.Manifest
import android.content.pm.PackageManager
import android.util.Size
import android.view.ViewGroup
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.camera.core.Camera
import androidx.camera.core.CameraSelector
import androidx.camera.core.ImageAnalysis
import androidx.camera.core.Preview
import androidx.camera.core.resolutionselector.ResolutionSelector
import androidx.camera.core.resolutionselector.ResolutionStrategy
import androidx.camera.lifecycle.ProcessCameraProvider
import androidx.camera.view.PreviewView
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clipToBounds
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalLifecycleOwner
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.AndroidView
import androidx.core.content.ContextCompat
import com.google.mlkit.vision.barcode.BarcodeScanning
import com.google.mlkit.vision.common.InputImage
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.core.i18n.Strings
import java.util.concurrent.Executors
import java.util.concurrent.atomic.AtomicBoolean

/**
 * 页内嵌扫码预览。CameraX + ML Kit，生命周期跟 [LocalLifecycleOwner]。
 *
 * [onCode] 每识别到一次非空 raw 回调一次；调用方负责去重与业务节流。
 * [torchOn] 控制闪光灯；[enabled]=false 时跳过分析（处理上一码期间）。
 */
@Composable
fun RiderCameraScanPreview(
    onCode: (String) -> Unit,
    modifier: Modifier = Modifier,
    torchOn: Boolean = false,
    enabled: Boolean = true,
) {
    val context = LocalContext.current
    var permissionGranted by remember {
        mutableStateOf(
            ContextCompat.checkSelfPermission(context, Manifest.permission.CAMERA) ==
                PackageManager.PERMISSION_GRANTED,
        )
    }
    val permissionLauncher = rememberLauncherForActivityResult(
        ActivityResultContracts.RequestPermission(),
    ) { granted -> permissionGranted = granted }

    if (!permissionGranted) {
        LaunchedEffect(Unit) {
            permissionLauncher.launch(Manifest.permission.CAMERA)
        }
        Box(
            modifier = modifier.background(Color(0xFF242936)),
            contentAlignment = Alignment.Center,
        ) {
            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                CircularProgressIndicator(color = Color.White)
                TextButton(onClick = { permissionLauncher.launch(Manifest.permission.CAMERA) }) {
                    Text(Strings.t(Str.CameraPermissionRequired), color = Color.White)
                }
            }
        }
        return
    }

    val lifecycleOwner = LocalLifecycleOwner.current
    var status by remember { mutableStateOf<String?>(null) }
    var camera by remember { mutableStateOf<Camera?>(null) }
    var previewReady by remember { mutableStateOf(false) }
    val analysisEnabled = remember { AtomicBoolean(enabled) }
    val latestOnCode by rememberUpdatedState(onCode)

    LaunchedEffect(enabled) { analysisEnabled.set(enabled) }
    LaunchedEffect(torchOn, camera) {
        runCatching { camera?.cameraControl?.enableTorch(torchOn) }
    }

    val previewView = remember {
        PreviewView(context).apply {
            layoutParams = ViewGroup.LayoutParams(
                ViewGroup.LayoutParams.MATCH_PARENT,
                ViewGroup.LayoutParams.MATCH_PARENT,
            )
            scaleType = PreviewView.ScaleType.FILL_CENTER
            // PERFORMANCE 默认走 SurfaceView，会穿透 Compose 盖住下方文字；
            // COMPATIBLE 用 TextureView，遵守布局层级，提示文案才不会叠进预览。
            implementationMode = PreviewView.ImplementationMode.COMPATIBLE
        }
    }

    DisposableEffect(lifecycleOwner, previewView) {
        val mainExecutor = ContextCompat.getMainExecutor(context)
        val analysisExecutor = Executors.newSingleThreadExecutor()
        val scanner = BarcodeScanning.getClient()
        val future = ProcessCameraProvider.getInstance(context)
        var provider: ProcessCameraProvider? = null

        future.addListener(
            {
                try {
                    val p = future.get()
                    provider = p
                    val preview = Preview.Builder().build().also {
                        it.surfaceProvider = previewView.surfaceProvider
                    }
                    val resolutionSelector = ResolutionSelector.Builder()
                        .setResolutionStrategy(
                            ResolutionStrategy(
                                Size(1280, 720),
                                ResolutionStrategy.FALLBACK_RULE_CLOSEST_HIGHER_THEN_LOWER,
                            ),
                        )
                        .build()
                    val analysis = ImageAnalysis.Builder()
                        .setResolutionSelector(resolutionSelector)
                        .setBackpressureStrategy(ImageAnalysis.STRATEGY_KEEP_ONLY_LATEST)
                        .build()
                    analysis.setAnalyzer(analysisExecutor) { imageProxy ->
                        if (!analysisEnabled.get()) {
                            imageProxy.close()
                            return@setAnalyzer
                        }
                        val mediaImage = imageProxy.image
                        if (mediaImage == null) {
                            imageProxy.close()
                            return@setAnalyzer
                        }
                        val image = InputImage.fromMediaImage(
                            mediaImage,
                            imageProxy.imageInfo.rotationDegrees,
                        )
                        scanner.process(image)
                            .addOnSuccessListener { barcodes ->
                                if (!analysisEnabled.get()) return@addOnSuccessListener
                                val raw = barcodes
                                    .asSequence()
                                    .mapNotNull { it.rawValue?.trim()?.takeIf(String::isNotEmpty) }
                                    .firstOrNull()
                                if (raw != null) latestOnCode(raw)
                            }
                            .addOnCompleteListener { imageProxy.close() }
                    }
                    p.unbindAll()
                    camera = p.bindToLifecycle(
                        lifecycleOwner,
                        CameraSelector.DEFAULT_BACK_CAMERA,
                        preview,
                        analysis,
                    )
                    previewReady = true
                    // 成功时不在预览上叠提示文案，避免换电/挪车等页「字压在画面上」。
                    status = null
                } catch (t: Throwable) {
                    camera = null
                    previewReady = false
                    status = Strings.t(Str.CameraError, t.message)
                }
            },
            mainExecutor,
        )

        onDispose {
            runCatching { provider?.unbindAll() }
            camera = null
            previewReady = false
            scanner.close()
            analysisExecutor.shutdown()
        }
    }

    // clipToBounds：TextureView 偶发画出 Compose 测量边界时，不让画面盖住下方「手电筒」等文案。
    Box(modifier = modifier.clipToBounds().background(Color.Black)) {
        AndroidView(
            factory = { previewView },
            modifier = Modifier
                .fillMaxSize()
                .clipToBounds(),
        )
        if (!previewReady && status == null) {
            CircularProgressIndicator(
                modifier = Modifier.align(Alignment.Center),
                color = Color.White,
            )
        }
        // 仅错误时叠字；对准提示由业务页自己放在预览外。
        status?.let { msg ->
            Text(
                text = msg,
                color = Color.White.copy(alpha = 0.9f),
                style = MaterialTheme.typography.labelSmall,
                modifier = Modifier
                    .align(Alignment.BottomCenter)
                    .padding(8.dp)
                    .background(Color.Black.copy(alpha = 0.55f))
                    .padding(horizontal = 8.dp, vertical = 4.dp),
            )
        }
    }
}
