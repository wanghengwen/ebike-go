package com.luopingtech.ebike.rider.ui.scan

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.core.i18n.Strings

interface RiderScanPreview {
    @Composable
    fun Preview(
        modifier: Modifier,
        torchOn: Boolean,
        enabled: Boolean,
        onCode: (String) -> Unit,
    )
}

object UnavailableScanPreview : RiderScanPreview {
    @Composable
    override fun Preview(
        modifier: Modifier,
        torchOn: Boolean,
        enabled: Boolean,
        onCode: (String) -> Unit,
    ) {
        Box(
            modifier = modifier.background(Color(0xFF1B1F2A)),
            contentAlignment = Alignment.Center,
        ) {
            Text(
                text = Strings.t(Str.ScannerUnavailable),
                color = Color(0xFFCCCCCC),
                style = MaterialTheme.typography.bodyMedium,
                textAlign = TextAlign.Center,
                modifier = Modifier.padding(24.dp),
            )
        }
    }
}

val LocalRiderScanPreview = staticCompositionLocalOf<RiderScanPreview> { UnavailableScanPreview }
