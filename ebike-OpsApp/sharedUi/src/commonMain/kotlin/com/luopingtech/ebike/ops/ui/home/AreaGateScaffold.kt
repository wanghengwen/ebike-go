package com.luopingtech.ebike.ops.ui.home

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.ui.icons.OpsBackChevron
import com.luopingtech.ebike.ops.ui.theme.OpsTheme

@Composable
fun AreaGateScaffold(
    title: String,
    currentLocationTitle: String,
    selectAreaTitle: String,
    currentAreaLabel: String,
    areas: List<ServiceArea>,
    selectedAreaId: String?,
    loading: Boolean,
    loadingText: String,
    errorMessage: String?,
    allowCancel: Boolean,
    backLabel: String,
    onBack: () -> Unit,
    onSelect: (ServiceArea) -> Unit,
) {
    val colors = OpsTheme.colors
    com.luopingtech.ebike.ops.ui.navigation.OpsBackHandler(enabled = allowCancel, onBack = onBack)
    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(colors.pageBackground),
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .background(colors.primary)
                .statusBarsPadding()
                .height(48.dp),
        ) {
            if (allowCancel) {
                OpsBackChevron(
                    onClick = onBack,
                    modifier = Modifier.align(Alignment.CenterStart),
                )
            }
            Text(
                text = title,
                color = colors.onPrimary,
                fontSize = 18.sp,
                fontWeight = FontWeight.Medium,
                modifier = Modifier.align(Alignment.Center),
            )
            if (allowCancel) {
                Text(
                    text = backLabel,
                    color = Color.Transparent,
                    modifier = Modifier.align(Alignment.CenterEnd).padding(end = 16.dp),
                )
            }
        }

        Column(
            modifier = Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState())
                .padding(horizontal = 12.dp, vertical = 12.dp),
        ) {
            Text(
                text = currentLocationTitle,
                color = Color.Black,
                fontSize = 16.sp,
                fontWeight = FontWeight.Bold,
            )
            Spacer(modifier = Modifier.height(12.dp))
            Box(
                modifier = Modifier
                    .height(56.dp)
                    .background(colors.primary, RoundedCornerShape(4.dp))
                    .padding(horizontal = 20.dp),
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    text = currentAreaLabel.ifBlank { "-" },
                    color = colors.onPrimary,
                    fontSize = 14.sp,
                    textAlign = TextAlign.Center,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                )
            }

            Spacer(modifier = Modifier.height(16.dp))
            Text(
                text = selectAreaTitle,
                color = Color.Black,
                fontSize = 16.sp,
                fontWeight = FontWeight.Bold,
            )
            Spacer(modifier = Modifier.height(12.dp))

            if (loading && areas.isEmpty()) {
                Text(text = loadingText, color = colors.textPrimary, fontSize = 14.sp)
            }

            BoxWithConstraints(modifier = Modifier.fillMaxWidth()) {
                val gap = 8.dp
                val cellW = (maxWidth - gap * 2) / 3
                // Box stacks children; wrap rows in Column so every chunked row is visible.
                Column(modifier = Modifier.fillMaxWidth()) {
                    areas.chunked(3).forEach { row ->
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(bottom = gap),
                            horizontalArrangement = Arrangement.spacedBy(gap),
                        ) {
                            row.forEach { area ->
                                val selected = area.id == selectedAreaId
                                ServiceAreaGridCell(
                                    name = area.name,
                                    id = area.id,
                                    selected = selected,
                                    modifier = Modifier
                                        .width(cellW)
                                        .height(84.dp),
                                    onClick = { onSelect(area) },
                                )
                            }
                            repeat(3 - row.size) {
                                Spacer(modifier = Modifier.width(cellW))
                            }
                        }
                    }
                }
            }

            errorMessage?.let {
                Spacer(modifier = Modifier.height(12.dp))
                Text(text = it, color = Color(0xFFE53935), fontSize = 14.sp)
            }
        }
    }
}

@Composable
private fun ServiceAreaGridCell(
    name: String,
    id: String,
    selected: Boolean,
    modifier: Modifier = Modifier,
    onClick: () -> Unit,
) {
    val colors = OpsTheme.colors
    Box(
        modifier = modifier
            .background(
                color = if (selected) colors.primary else colors.chipBackground,
                shape = RoundedCornerShape(4.dp),
            )
            .then(
                if (selected) {
                    Modifier
                } else {
                    Modifier.border(1.dp, colors.divider, RoundedCornerShape(4.dp))
                },
            )
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            )
            .padding(6.dp),
        contentAlignment = Alignment.Center,
    ) {
        Text(
            text = "$name\n($id)",
            color = if (selected) colors.onPrimary else colors.textPrimary,
            fontSize = 12.sp,
            textAlign = TextAlign.Center,
            maxLines = 4,
            overflow = TextOverflow.Ellipsis,
            lineHeight = 16.sp,
        )
    }
}

