package com.luopingtech.ebike.rider.scan

import android.content.Intent
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.core.i18n.Strings
import java.util.concurrent.atomic.AtomicBoolean

/**
 * 全屏扫一次即回传。业务页内嵌请直接用 [RiderCameraScanPreview]。
 */
class ScanQrActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            MaterialTheme {
                Surface(modifier = Modifier.fillMaxSize()) {
                    val delivered = remember { AtomicBoolean(false) }
                    var finished by remember { mutableStateOf(false) }
                    Box(modifier = Modifier.fillMaxSize()) {
                        RiderCameraScanPreview(
                            onCode = { raw ->
                                if (delivered.compareAndSet(false, true)) {
                                    finished = true
                                    setResult(RESULT_OK, Intent().putExtra(EXTRA_RAW, raw))
                                    finish()
                                }
                            },
                            enabled = !finished,
                            modifier = Modifier.fillMaxSize(),
                        )
                        Button(
                            onClick = {
                                setResult(RESULT_CANCELED)
                                finish()
                            },
                            modifier = Modifier
                                .align(Alignment.BottomCenter)
                                .fillMaxWidth()
                                .padding(24.dp),
                        ) {
                            Text(Strings.t(Str.Cancel))
                        }
                    }
                }
            }
        }
    }

    companion object {
        const val EXTRA_RAW: String = "rider_scan_raw"
    }
}
