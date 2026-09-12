package com.luopingtech.ebike.ops.scan

import android.Manifest
import android.content.Intent
import android.content.pm.PackageManager
import android.os.Bundle
import android.util.Size
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.result.contract.ActivityResultContracts
import androidx.camera.core.CameraSelector
import androidx.camera.core.ImageAnalysis
import androidx.camera.core.Preview
import androidx.camera.core.resolutionselector.ResolutionSelector
import androidx.camera.core.resolutionselector.ResolutionStrategy
import androidx.camera.lifecycle.ProcessCameraProvider
import androidx.camera.view.PreviewView
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.AndroidView
import androidx.core.content.ContextCompat
import com.google.mlkit.vision.barcode.BarcodeScanning
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.google.mlkit.vision.common.InputImage
import java.util.concurrent.Executors
import java.util.concurrent.atomic.AtomicBoolean

/**
 * Full-screen QR / barcode capture. Returns [EXTRA_RAW] on success.
 */
class ScanQrActivity : ComponentActivity() {
    private val permissionLauncher = registerForActivityResult(
        ActivityResultContracts.RequestPermission(),
    ) { granted ->
        permissionGranted = granted
        if (!granted) {
            setResult(RESULT_CANCELED)
            finish()
        }
    }

    private var permissionGranted by mutableStateOf(false)

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        permissionGranted = ContextCompat.checkSelfPermission(
            this,
            Manifest.permission.CAMERA,
        ) == PackageManager.PERMISSION_GRANTED
        if (!permissionGranted) {
            permissionLauncher.launch(Manifest.permission.CAMERA)
        }

        setContent {
            MaterialTheme {
                Surface(modifier = Modifier.fillMaxSize()) {
                    if (permissionGranted) {
                        CameraScanScreen(
                            onCode = { raw ->
                                setResult(RESULT_OK, Intent().putExtra(EXTRA_RAW, raw))
                                finish()
                            },
                            onCancel = {
                                setResult(RESULT_CANCELED)
                                finish()
                            },
                        )
                    } else {
                        Box(
                            modifier = Modifier.fillMaxSize(),
                            contentAlignment = Alignment.Center,
                        ) {
                            Text(Strings.t(Str.CameraPermissionRequired))
                        }
                    }
                }
            }
        }
    }

    companion object {
        const val EXTRA_RAW: String = "ops_scan_raw"
    }
}

@Composable
private fun CameraScanScreen(
    onCode: (String) -> Unit,
    onCancel: () -> Unit,
) {
    val activity = androidx.compose.ui.platform.LocalContext.current as ComponentActivity
    val delivered = remember { AtomicBoolean(false) }
    var status by remember { mutableStateOf(Strings.t(Str.AimAtQr)) }

    Box(modifier = Modifier.fillMaxSize()) {
        AndroidView(
            modifier = Modifier.fillMaxSize(),
            factory = { ctx ->
                PreviewView(ctx).also { previewView ->
                    bindCamera(
                        activity = activity,
                        previewView = previewView,
                        onStatus = { status = it },
                        onCode = { raw ->
                            if (delivered.compareAndSet(false, true)) {
                                onCode(raw)
                            }
                        },
                    )
                }
            },
        )
        Text(
            text = status,
            modifier = Modifier
                .align(Alignment.TopCenter)
                .padding(24.dp),
            color = MaterialTheme.colorScheme.onPrimary,
            style = MaterialTheme.typography.titleMedium,
        )
        Button(
            onClick = onCancel,
            modifier = Modifier
                .align(Alignment.BottomCenter)
                .fillMaxWidth()
                .padding(24.dp),
        ) {
            Text(Strings.t(Str.Cancel))
        }
    }
}

private fun bindCamera(
    activity: ComponentActivity,
    previewView: PreviewView,
    onStatus: (String) -> Unit,
    onCode: (String) -> Unit,
) {
    val cameraProviderFuture = ProcessCameraProvider.getInstance(activity)
    val executor = ContextCompat.getMainExecutor(activity)
    val analysisExecutor = Executors.newSingleThreadExecutor()
    val scanner = BarcodeScanning.getClient()

    cameraProviderFuture.addListener(
        {
            val cameraProvider = cameraProviderFuture.get()
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
                        val raw = barcodes
                            .asSequence()
                            .mapNotNull { it.rawValue?.trim()?.takeIf(String::isNotEmpty) }
                            .firstOrNull()
                        if (raw != null) {
                            onStatus(Strings.t(Str.Recognized))
                            onCode(raw)
                        }
                    }
                    .addOnCompleteListener {
                        imageProxy.close()
                    }
            }
            try {
                cameraProvider.unbindAll()
                cameraProvider.bindToLifecycle(
                    activity,
                    CameraSelector.DEFAULT_BACK_CAMERA,
                    preview,
                    analysis,
                )
                onStatus(Strings.t(Str.AimAtQr))
            } catch (t: Throwable) {
                onStatus(Strings.t(Str.CameraError, t.message))
            }
        },
        executor,
    )
}
