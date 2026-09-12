package com.luopingtech.ebike.ops.ui.fence

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.domain.model.FenceKind
import com.luopingtech.ebike.ops.domain.model.FencePolygon
import com.luopingtech.ebike.ops.ui.analysis.VcdTopBar
import com.luopingtech.ebike.ops.ui.theme.OpsTheme

@Composable
fun FenceBrowseScreen(
    app: OpsApp,
    onClose: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val home by app.homeFeature.state.collectAsState()
    val state by app.fenceBrowseFeature.state.collectAsState()

    LaunchedEffect(home.currentArea?.id) {
        app.fenceBrowseFeature.load(home.currentArea)
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(Color(0xFFF6F7F9)),
    ) {
        VcdTopBar(title = t(Str.ParkingFence), onBack = {
            app.fenceBrowseFeature.clear()
            onClose()
        })
        Column(
            modifier = Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState())
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            Text(
                text = t(Str.FenceBrowseHint),
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
            if (state.loading) {
                CircularProgressIndicator(color = OpsTheme.colors.primary)
            } else if (state.errorMessage != null) {
                Text(text = state.errorMessage.orEmpty(), color = MaterialTheme.colorScheme.error)
            } else {
                FenceSection(
                    title = t(Str.FenceSectionCount, t(Str.FenceParkingSection), state.bundle.parkings.size),
                    fences = state.bundle.parkings,
                    translate = { key, args -> t(key, *args) },
                )
                FenceSection(
                    title = t(Str.FenceSectionCount, t(Str.FenceNoParkingSection), state.bundle.noParkings.size),
                    fences = state.bundle.noParkings,
                    translate = { key, args -> t(key, *args) },
                )
                FenceSection(
                    title = t(Str.FenceSectionCount, t(Str.FenceServiceSection), state.bundle.serviceAreas.size),
                    fences = state.bundle.serviceAreas,
                    translate = { key, args -> t(key, *args) },
                )
            }
        }
    }
}

@Composable
private fun FenceSection(
    title: String,
    fences: List<FencePolygon>,
    translate: (Str, Array<out Any?>) -> String,
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(Color.White, MaterialTheme.shapes.medium)
            .padding(12.dp),
        verticalArrangement = Arrangement.spacedBy(0.dp),
    ) {
        Text(
            text = title,
            style = MaterialTheme.typography.titleSmall,
            fontWeight = FontWeight.SemiBold,
            modifier = Modifier.padding(bottom = 8.dp),
        )
        if (fences.isEmpty()) {
            Text(
                text = translate(Str.AdminEmptyList, emptyArray()),
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
        } else {
            fences.forEachIndexed { index, fence ->
                if (index > 0) {
                    HorizontalDivider(
                        modifier = Modifier.padding(vertical = 8.dp),
                        color = Color(0xFFE8E8E8),
                    )
                }
                FenceRow(fence = fence, translate = translate)
            }
        }
    }
}

@Composable
private fun FenceRow(
    fence: FencePolygon,
    translate: (Str, Array<out Any?>) -> String,
) {
    val statusText = when (fence.izEnable) {
        true -> translate(Str.StationStatusOperating, emptyArray())
        false -> translate(Str.StationStatusStopped, emptyArray())
        null -> null
    }
    val statusColor = when (fence.izEnable) {
        true -> Color(0xFF00BE59)
        false -> Color(0xFFFF2222)
        null -> MaterialTheme.colorScheme.onSurfaceVariant
    }
    val meta = fenceMetaLine(fence, translate)

    Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(
                text = fence.name,
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.Medium,
                modifier = Modifier.weight(1f, fill = false).padding(end = 8.dp),
            )
            if (statusText != null) {
                Text(
                    text = statusText,
                    style = MaterialTheme.typography.bodyMedium,
                    fontWeight = FontWeight.SemiBold,
                    color = statusColor,
                )
            }
        }
        if (meta.isNotBlank()) {
            Text(
                text = meta,
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }
        if (fence.address.isNotBlank()) {
            Text(
                text = fence.address,
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }
    }
}

private fun fenceMetaLine(
    fence: FencePolygon,
    translate: (Str, Array<out Any?>) -> String,
): String {
    val parts = mutableListOf<String>()
    when (fence.kind) {
        FenceKind.Parking -> {
            if (fence.maxParkingNumber > 0 || fence.currentParkingNumber > 0) {
                parts += translate(
                    Str.FenceParkingOccupancy,
                    arrayOf(fence.currentParkingNumber, fence.maxParkingNumber),
                )
            } else if (fence.carCount > 0) {
                parts += translate(Str.FenceCarCount, arrayOf(fence.carCount))
            }
        }
        FenceKind.ServiceArea -> {
            if (fence.carCount > 0) {
                parts += translate(Str.FenceCarCount, arrayOf(fence.carCount))
            }
        }
        else -> Unit
    }
    if (fence.points.isNotEmpty()) {
        parts += translate(Str.FencePointCount, arrayOf(fence.points.size))
    }
    return parts.joinToString(" · ")
}
