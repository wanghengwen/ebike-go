package com.luopingtech.ebike.ops.ui.home

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInHorizontally
import androidx.compose.animation.slideOutHorizontally
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.zIndex
import com.luopingtech.ebike.ops.domain.vehicle.VehicleAlarmFilter
import com.luopingtech.ebike.ops.ui.icons.OpsIcon
import com.luopingtech.ebike.ops.ui.icons.painterResource
import com.luopingtech.ebike.ops.ui.theme.OpsTheme

/**
 * 对齐遗留 [FilterSlidePop] + [PanelStatisticsView]：
 * - 总宽 330dp = 关闭条 30dp + 白面板 305dp
 * - 蒙层 / 关闭条 / 重置·确定 后关闭；点选仅改草稿，确定才生效
 * - Chip：112×40、圆角 4、选中主题色、未选中 #F4F4F4+#D7D7D7
 */
@Composable
fun HomeAlarmFilterSlide(
    visible: Boolean,
    applied: Set<Int>,
    title: String,
    resetLabel: String,
    sureLabel: String,
    sureWithCountLabel: (Int) -> String,
    onDismiss: () -> Unit,
    onApply: (Set<Int>) -> Unit,
    modifier: Modifier = Modifier,
) {
    var draft by remember { mutableStateOf(applied) }
    LaunchedEffect(visible, applied) {
        if (visible) draft = applied
    }

    AnimatedVisibility(
        visible = visible,
        enter = fadeIn() + slideInHorizontally(initialOffsetX = { it }),
        exit = fadeOut() + slideOutHorizontally(targetOffsetX = { it }),
        modifier = modifier
            .fillMaxSize()
            .zIndex(20f),
    ) {
        Box(modifier = Modifier.fillMaxSize()) {
            // Outside mask — legacy mask_color #66000000 + setOutSideDismiss(true)
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(Color(0x66000000))
                    .clickable(
                        interactionSource = remember { MutableInteractionSource() },
                        indication = null,
                        onClick = onDismiss,
                    ),
            )
            Row(
                modifier = Modifier
                    .align(Alignment.CenterEnd)
                    .fillMaxHeight()
                    .width(330.dp),
            ) {
                // Left 30dp close strip (also catches clicks like viewParentLayout)
                Box(
                    modifier = Modifier
                        .width(30.dp)
                        .fillMaxHeight()
                        .clickable(
                            interactionSource = remember { MutableInteractionSource() },
                            indication = null,
                            onClick = onDismiss,
                        ),
                    contentAlignment = Alignment.Center,
                ) {
                    Image(
                        painter = painterResource(OpsIcon.FilterDrawerClose),
                        contentDescription = null,
                        modifier = Modifier.width(30.dp),
                        contentScale = ContentScale.FillWidth,
                    )
                }
                Box(
                    modifier = Modifier
                        .width(305.dp)
                        .fillMaxHeight()
                        .background(Color.White)
                        .clickable(
                            interactionSource = remember { MutableInteractionSource() },
                            indication = null,
                            onClick = {},
                        ),
                ) {
                    Column(
                        modifier = Modifier
                            .fillMaxSize()
                            .verticalScroll(rememberScrollState())
                            .padding(bottom = 120.dp),
                    ) {
                        Text(
                            text = title,
                            color = Color.Black,
                            fontSize = 16.sp,
                            fontWeight = FontWeight.Bold,
                            modifier = Modifier.padding(start = 24.dp, top = 40.dp),
                        )
                        Column(
                            modifier = Modifier
                                .padding(start = 16.5.dp, top = 8.dp, end = 8.dp)
                                .width(288.dp),
                            verticalArrangement = Arrangement.spacedBy(0.dp),
                        ) {
                            VehicleAlarmFilter.entries.chunked(2).forEach { row ->
                                Row(
                                    modifier = Modifier.fillMaxWidth(),
                                    horizontalArrangement = Arrangement.Start,
                                ) {
                                    row.forEach { item ->
                                        AlarmFilterChip(
                                            label = item.label,
                                            selected = item.code in draft,
                                            onClick = {
                                                draft = if (item.code in draft) {
                                                    draft - item.code
                                                } else {
                                                    draft + item.code
                                                }
                                            },
                                        )
                                    }
                                }
                            }
                        }
                        Spacer(modifier = Modifier.height(24.dp))
                    }

                    // Bottom bar: reset / sure — VehicleFilterLayoutScrollView
                    FilterBottomBar(
                        resetLabel = resetLabel,
                        sureLabel = if (draft.isEmpty()) sureLabel else sureWithCountLabel(draft.size),
                        onReset = {
                            draft = emptySet()
                            onApply(emptySet())
                            onDismiss()
                        },
                        onSure = {
                            onApply(draft)
                            onDismiss()
                        },
                        modifier = Modifier.align(Alignment.BottomCenter),
                    )
                }
            }
        }
    }
}

@Composable
private fun AlarmFilterChip(
    label: String,
    selected: Boolean,
    onClick: () -> Unit,
) {
    val primary = OpsTheme.colors.primary
    val bg = if (selected) primary else Color(0xFFF4F4F4)
    val border = if (selected) primary else Color(0xFFD7D7D7)
    val fg = if (selected) Color.White else Color.Black
    Box(
        modifier = Modifier
            .padding(8.dp)
            .width(112.dp)
            .height(40.dp)
            .background(bg, RoundedCornerShape(4.dp))
            .border(1.dp, border, RoundedCornerShape(4.dp))
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            ),
        contentAlignment = Alignment.Center,
    ) {
        Text(
            text = label,
            color = fg,
            fontSize = 16.sp,
            textAlign = TextAlign.Center,
            maxLines = 1,
        )
    }
}

@Composable
private fun FilterBottomBar(
    resetLabel: String,
    sureLabel: String,
    onReset: () -> Unit,
    onSure: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val primary = OpsTheme.colors.primary
    Box(
        modifier = modifier
            .fillMaxWidth()
            .height(100.dp)
            .background(Color.White),
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .align(Alignment.BottomCenter)
                .padding(start = 27.dp, end = 27.dp, bottom = 32.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
        ) {
            Box(
                modifier = Modifier
                    .width(120.dp)
                    .height(48.dp)
                    .background(primary.copy(alpha = 0.1f), RoundedCornerShape(4.dp))
                    .clickable(
                        interactionSource = remember { MutableInteractionSource() },
                        indication = null,
                        onClick = onReset,
                    ),
                contentAlignment = Alignment.Center,
            ) {
                Text(resetLabel, color = primary, fontSize = 16.sp, fontWeight = FontWeight.Bold)
            }
            Box(
                modifier = Modifier
                    .width(120.dp)
                    .height(48.dp)
                    .background(primary, RoundedCornerShape(4.dp))
                    .clickable(
                        interactionSource = remember { MutableInteractionSource() },
                        indication = null,
                        onClick = onSure,
                    ),
                contentAlignment = Alignment.Center,
            ) {
                Text(sureLabel, color = Color.White, fontSize = 16.sp, fontWeight = FontWeight.Bold)
            }
        }
    }
}
