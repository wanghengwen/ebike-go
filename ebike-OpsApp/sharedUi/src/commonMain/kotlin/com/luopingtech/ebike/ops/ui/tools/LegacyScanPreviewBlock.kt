package com.luopingtech.ebike.ops.ui.tools

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.clipToBounds
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

/**
 * 对齐原版 VehicleTag / RelocationScan：300dp 深色区、240×240 预览、手电筒。
 */
@Composable
fun LegacyScanPreviewBlock(
    torchOn: Boolean,
    torchLabel: String,
    scanEnabled: Boolean,
    onTorchChange: (Boolean) -> Unit,
    scanPreview: @Composable (
        modifier: Modifier,
        torchOn: Boolean,
        enabled: Boolean,
        onCode: (String) -> Unit,
    ) -> Unit,
    onCode: (String) -> Unit,
) {
    val scanBg = Color(0xFF242936)
    Column(modifier = Modifier.fillMaxWidth().background(scanBg)) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(300.dp)
                .padding(top = 14.dp),
            contentAlignment = Alignment.TopCenter,
        ) {
            Box(
                modifier = Modifier
                    .width(240.dp)
                    .aspectRatio(1f)
                    .clip(RoundedCornerShape(8.dp))
                    .clipToBounds(),
            ) {
                scanPreview(
                    Modifier.fillMaxSize(),
                    torchOn,
                    scanEnabled,
                    onCode,
                )
                ScanCornerBrackets(color = Color.White)
            }
        }
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(vertical = 12.dp),
            horizontalArrangement = Arrangement.Center,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Row(
                verticalAlignment = Alignment.CenterVertically,
                modifier = Modifier.clickable { onTorchChange(!torchOn) },
            ) {
                CheckboxMark(checked = torchOn)
                Spacer(modifier = Modifier.width(6.dp))
                Text(torchLabel, color = Color(0xFFD3D3D3), fontSize = 16.sp)
            }
        }
    }
}
