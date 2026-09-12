package com.luopingtech.ebike.ops.ui.task

import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
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
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.ui.icons.OpsIcon
import com.luopingtech.ebike.ops.ui.icons.painterResource
import com.luopingtech.ebike.ops.ui.theme.OpsTheme

private val PageBg = Color(0xFFF2F8FD)
private val TitleColor = Color(0xFF333333)
private val TipBg = Color(0xFFFE6868)

data class TaskCenterCardItem(
    val id: String,
    val title: String,
    val icon: OpsIcon,
    /** Null or blank → badge hidden (legacy: totalCount == 0). */
    val badgeText: String? = null,
    val onClick: () -> Unit,
)

/**
 * Legacy TaskCenterFragment:
 * #F2F8FD page, transparent area toolbar, right decoration,
 * title「任务中心」, 2-column white 68dp cards with left illustration + red tip.
 */
@Composable
fun TaskScaffold(
    areaName: String,
    pageTitle: String,
    cards: List<TaskCenterCardItem>,
    onChangeArea: () -> Unit,
    emptyHint: String? = null,
) {
    val colors = OpsTheme.colors
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(PageBg),
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .statusBarsPadding(),
        ) {
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(50.dp)
                    .padding(horizontal = 12.dp),
            ) {
                Row(
                    modifier = Modifier
                        .align(Alignment.Center)
                        .clickable(
                            interactionSource = remember { MutableInteractionSource() },
                            indication = null,
                            onClick = onChangeArea,
                        ),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text(
                        text = areaName,
                        color = colors.textPrimary,
                        fontSize = 18.sp,
                        fontWeight = FontWeight.Medium,
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis,
                    )
                    Text(text = " ▾", color = colors.textPrimary, fontSize = 12.sp)
                }
            }
            Box(modifier = Modifier.fillMaxSize()) {
                Image(
                    painter = painterResource(OpsIcon.BgIconTaskCenter),
                    contentDescription = null,
                    modifier = Modifier
                        .align(Alignment.TopEnd)
                        .padding(end = 24.dp)
                        .size(width = 108.dp, height = 115.dp),
                    contentScale = ContentScale.Fit,
                )
                Column(
                    modifier = Modifier
                        .fillMaxSize()
                        .verticalScroll(rememberScrollState()),
                ) {
                    Text(
                        text = pageTitle,
                        color = TitleColor,
                        fontSize = 20.sp,
                        fontWeight = FontWeight.SemiBold,
                        modifier = Modifier.padding(start = 24.dp, top = 20.dp, bottom = 8.dp),
                    )
                    if (cards.isEmpty()) {
                        Text(
                            text = emptyHint.orEmpty(),
                            color = Color(0xFF999999),
                            fontSize = 14.sp,
                            modifier = Modifier.padding(horizontal = 24.dp, vertical = 24.dp),
                        )
                    } else {
                        cards.chunked(2).forEach { row ->
                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(horizontal = 12.dp),
                                horizontalArrangement = Arrangement.spacedBy(12.dp),
                            ) {
                                row.forEach { card ->
                                    TaskCenterCard(
                                        item = card,
                                        modifier = Modifier
                                            .weight(1f)
                                            .padding(top = 12.dp),
                                    )
                                }
                                if (row.size == 1) {
                                    Spacer(modifier = Modifier.weight(1f))
                                }
                            }
                        }
                    }
                    Spacer(modifier = Modifier.height(32.dp))
                }
            }
        }
    }
}

@Composable
private fun TaskCenterCard(
    item: TaskCenterCardItem,
    modifier: Modifier = Modifier,
) {
    Box(
        modifier = modifier
            .height(68.dp)
            .background(Color.White, RoundedCornerShape(8.dp))
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = item.onClick,
            ),
    ) {
        Row(
            modifier = Modifier.fillMaxSize(),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Image(
                painter = painterResource(item.icon),
                contentDescription = item.title,
                modifier = Modifier.size(width = 78.dp, height = 68.dp),
                contentScale = ContentScale.Fit,
            )
            Text(
                text = item.title,
                color = TitleColor,
                fontSize = 16.sp,
                fontWeight = FontWeight.SemiBold,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
                modifier = Modifier
                    .weight(1f)
                    .padding(end = 8.dp),
            )
        }
        val badge = item.badgeText?.takeIf { it.isNotBlank() }
        if (badge != null) {
            Box(
                modifier = Modifier
                    .align(Alignment.TopEnd)
                    .width(47.dp)
                    .height(18.dp)
                    .background(
                        TipBg,
                        RoundedCornerShape(topStart = 0.dp, topEnd = 8.dp, bottomEnd = 0.dp, bottomStart = 8.dp),
                    ),
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    text = badge,
                    color = Color.White,
                    fontSize = 10.sp,
                    fontWeight = FontWeight.Medium,
                    textAlign = TextAlign.Center,
                    maxLines = 1,
                )
            }
        }
    }
}
