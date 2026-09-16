package com.luopingtech.ebike.rider.ui.ride

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.luopingtech.ebike.rider.ui.theme.RiderTheme
import kotlin.math.max
import kotlinx.coroutines.delay

/**
 * 开锁中底部面板（对齐 UniApp `CarTipSheet` type=0）。
 *
 * 进度假走到 98% 后停住，等真实开锁结果把相位切走；不假装到 100%，避免「进度满了但车还没开」。
 */
@Composable
internal fun UnlockingSheet(
    carId: String,
    formatProgress: (Int) -> String,
    safetyTip: String,
    modifier: Modifier = Modifier,
) {
    var progress by remember { mutableIntStateOf(0) }
    LaunchedEffect(Unit) {
        val stepMs = max(20L, 2_000L / 98L)
        while (progress < 98) {
            delay(stepMs)
            progress += 1
        }
    }

    Surface(
        tonalElevation = 4.dp,
        shape = MaterialTheme.shapes.large,
        modifier = modifier.fillMaxWidth(),
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 20.dp, vertical = 16.dp),
            verticalArrangement = Arrangement.spacedBy(10.dp),
        ) {
            if (carId.isNotBlank()) {
                Text(
                    text = "NO.$carId",
                    style = MaterialTheme.typography.titleSmall,
                    color = RiderTheme.colors.textPrimary,
                )
            }
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(10.dp),
            ) {
                CircularProgressIndicator(
                    modifier = Modifier.size(22.dp),
                    strokeWidth = 2.dp,
                    color = RiderTheme.colors.primary,
                )
                Text(
                    text = formatProgress(progress),
                    style = MaterialTheme.typography.titleMedium,
                    color = RiderTheme.colors.textPrimary,
                )
            }
            Text(
                text = safetyTip,
                style = MaterialTheme.typography.bodySmall,
                color = RiderTheme.colors.textTertiary,
            )
            // 占位：租户 cyclingCfg.cartipUnlockingNew 未下发到原生时仍保持与 UniApp 相近的面板高度。
            Surface(
                color = RiderTheme.colors.chipBackground,
                shape = MaterialTheme.shapes.medium,
                modifier = Modifier
                    .fillMaxWidth()
                    .height(120.dp),
            ) {}
        }
    }
}
