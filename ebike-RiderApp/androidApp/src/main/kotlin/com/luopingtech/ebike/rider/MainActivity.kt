package com.luopingtech.ebike.rider

import android.os.Bundle
import android.widget.Toast
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.Surface
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.ui.Modifier
import com.luopingtech.ebike.rider.platform.ActivityCodeScanner
import com.luopingtech.ebike.rider.platform.ActivityPhotoCapture
import com.luopingtech.ebike.rider.platform.rememberBlePermissionGate
import com.luopingtech.ebike.rider.platform.rememberLocationPermissionGate
import com.luopingtech.ebike.rider.scan.AndroidScanPreview
import com.luopingtech.ebike.rider.ui.feedback.LocalRiderToast
import com.luopingtech.ebike.rider.ui.map.LocalRiderMapRenderer
import com.luopingtech.ebike.rider.ui.map.riderMapRendererFor
import com.luopingtech.ebike.rider.ui.permission.LocalRiderBlePermissionGate
import com.luopingtech.ebike.rider.ui.permission.LocalRiderLocationPermissionGate
import com.luopingtech.ebike.rider.ui.scan.LocalRiderScanPreview
import com.luopingtech.ebike.rider.ui.shell.RiderAppRoot
import com.luopingtech.ebike.rider.ui.theme.RiderTheme

class MainActivity : ComponentActivity() {
    private lateinit var activityCodeScanner: ActivityCodeScanner
    private lateinit var activityPhotoCapture: ActivityPhotoCapture

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        val host = application as RiderApplication
        val app = host.riderApp
        activityCodeScanner = ActivityCodeScanner(this)
        host.codeScannerBridge.bind(activityCodeScanner)
        activityPhotoCapture = ActivityPhotoCapture(this)
        host.photoCaptureBridge.bind(activityPhotoCapture)
        setContent {
            RiderTheme(branding = app.config.branding) {
                CompositionLocalProvider(
                    LocalRiderMapRenderer provides riderMapRendererFor(app),
                    LocalRiderScanPreview provides AndroidScanPreview,
                    LocalRiderLocationPermissionGate provides rememberLocationPermissionGate(app),
                    LocalRiderBlePermissionGate provides rememberBlePermissionGate(app),
                    LocalRiderToast provides { message: String ->
                        Toast.makeText(this, message, Toast.LENGTH_SHORT).show()
                    },
                ) {
                    Surface(modifier = Modifier.fillMaxSize()) {
                        RiderAppRoot(app)
                    }
                }
            }
        }
    }

    override fun onDestroy() {
        val host = application as RiderApplication
        host.codeScannerBridge.unbind(activityCodeScanner)
        host.photoCaptureBridge.unbind(activityPhotoCapture)
        super.onDestroy()
    }
}
