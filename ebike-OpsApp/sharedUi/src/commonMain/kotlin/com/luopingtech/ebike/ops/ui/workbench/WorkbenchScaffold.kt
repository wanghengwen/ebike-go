package com.luopingtech.ebike.ops.ui.workbench

import androidx.compose.foundation.Canvas
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.gestures.detectDragGesturesAfterLongPress
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
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.IntSize
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.zIndex
import com.luopingtech.ebike.ops.ui.icons.OpsIcon
import com.luopingtech.ebike.ops.ui.icons.painterResource
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import kotlin.math.roundToInt

enum class WorkbenchEditBadge {
    None,
    Add,
    Remove,
}

data class WorkbenchModuleItem(
    val id: String,
    val title: String,
    val icon: OpsIcon,
    val onClick: () -> Unit,
    val editBadge: WorkbenchEditBadge = WorkbenchEditBadge.None,
    val onBadgeClick: (() -> Unit)? = null,
)

data class WorkbenchModuleSection(
    val title: String,
    val items: List<WorkbenchModuleItem>,
)

private val PageGray = Color(0xFFF2F2F2)
private val EditBlue = Color(0xFF1180F9)
private val RoleStroke = Color(0x30FFFFFF)
private val HeaderOnPrimary = Color.White
const val WORKBENCH_COMMON_MAX = 4

/**
 * Legacy MineFragment / fragment_new_mine:
 * theme toolbar → curved profile → gray scroll, white 8dp cards, 4-col grid.
 */
@Composable
fun WorkbenchScaffold(
    areaName: String,
    companyName: String,
    userName: String,
    roleLabel: String,
    commonTitle: String,
    editLabel: String,
    settingsTitle: String,
    backLabel: String,
    saveLabel: String,
    editCommonTitle: String,
    dragHint: String,
    commonItems: List<WorkbenchModuleItem>,
    sections: List<WorkbenchModuleSection>,
    settingsContent: (@Composable () -> Unit)? = null,
    settingsOpen: Boolean = false,
    editingCommon: Boolean = false,
    onOpenSettings: () -> Unit,
    onCloseSettings: () -> Unit,
    onChangeArea: () -> Unit,
    onStartEditCommon: () -> Unit,
    onCancelEditCommon: () -> Unit,
    onSaveEditCommon: () -> Unit,
    onReorderCommon: (from: Int, to: Int) -> Unit,
    avatar: OpsIcon = OpsIcon.TenantLogo,
) {
    if (settingsOpen && settingsContent != null) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .background(PageGray),
        ) {
            WorkbenchToolbar(
                title = settingsTitle,
                showBack = true,
                backLabel = backLabel,
                trailingLabel = null,
                onBack = onCloseSettings,
                onTrailing = null,
                onSettings = null,
                onTitleClick = null,
            )
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .verticalScroll(rememberScrollState())
                    .padding(16.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                settingsContent()
            }
        }
        return
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(PageGray),
    ) {
        if (editingCommon) {
            WorkbenchToolbar(
                title = editCommonTitle,
                showBack = true,
                backLabel = backLabel,
                trailingLabel = saveLabel,
                onBack = onCancelEditCommon,
                onTrailing = onSaveEditCommon,
                onSettings = null,
                onTitleClick = null,
            )
        } else {
            WorkbenchToolbar(
                title = areaName,
                showBack = false,
                backLabel = backLabel,
                trailingLabel = null,
                onBack = null,
                onTrailing = null,
                onSettings = onOpenSettings,
                onTitleClick = onChangeArea,
            )
        }
        Column(
            modifier = Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState()),
        ) {
            if (!editingCommon) {
                WorkbenchProfileHeader(
                    companyName = companyName,
                    userName = userName,
                    roleLabel = roleLabel,
                    avatar = avatar,
                )
            }
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .then(if (editingCommon) Modifier else Modifier.offset(y = (-60).dp))
                    .padding(horizontal = 12.dp)
                    .padding(top = if (editingCommon) 12.dp else 0.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                if (editingCommon) {
                    Text(
                        text = dragHint,
                        color = Color(0xFF666666),
                        fontSize = 14.sp,
                        textAlign = TextAlign.Center,
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(top = 8.dp, bottom = 4.dp),
                    )
                }
                WorkbenchModuleCard(
                    title = commonTitle,
                    items = commonItems,
                    emptyIcon = OpsIcon.MineModuleAddGray,
                    editLabel = if (!editingCommon && commonItems.isNotEmpty()) editLabel else null,
                    onEditClick = if (!editingCommon && commonItems.isNotEmpty()) onStartEditCommon else null,
                    onEmptyClick = if (!editingCommon) onStartEditCommon else null,
                    titleIndentExtra = false,
                    dragEnabled = commonItems.size > 1,
                    onReorder = onReorderCommon,
                )
                sections.forEach { section ->
                    if (section.items.isNotEmpty()) {
                        WorkbenchModuleCard(
                            title = section.title,
                            items = section.items,
                            titleIndentExtra = true,
                            dragEnabled = false,
                            onReorder = { _, _ -> },
                        )
                    }
                }
                Spacer(modifier = Modifier.height(24.dp))
            }
        }
    }
}

@Composable
private fun WorkbenchToolbar(
    title: String,
    showBack: Boolean,
    backLabel: String,
    trailingLabel: String?,
    onBack: (() -> Unit)?,
    onTrailing: (() -> Unit)?,
    onSettings: (() -> Unit)?,
    onTitleClick: (() -> Unit)?,
) {
    if (showBack && onBack != null) {
        com.luopingtech.ebike.ops.ui.navigation.OpsBackHandler(onBack = onBack)
    }
    val colors = OpsTheme.colors
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .background(colors.primary)
            .statusBarsPadding()
            .height(50.dp)
            .padding(horizontal = 12.dp),
    ) {
        if (showBack && onBack != null) {
            Text(
                text = backLabel,
                color = HeaderOnPrimary,
                fontSize = 16.sp,
                modifier = Modifier
                    .align(Alignment.CenterStart)
                    .clickable(
                        interactionSource = remember { MutableInteractionSource() },
                        indication = null,
                        onClick = onBack,
                    ),
            )
        }
        Row(
            modifier = Modifier
                .align(Alignment.Center)
                .then(
                    if (onTitleClick != null) {
                        Modifier.clickable(
                            interactionSource = remember { MutableInteractionSource() },
                            indication = null,
                            onClick = onTitleClick,
                        )
                    } else {
                        Modifier
                    },
                ),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(
                text = title,
                color = HeaderOnPrimary,
                fontSize = 18.sp,
                fontWeight = FontWeight.Medium,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            if (onTitleClick != null) {
                Image(
                    painter = painterResource(OpsIcon.ArrowDown),
                    contentDescription = null,
                    modifier = Modifier
                        .padding(start = 4.dp)
                        .size(12.dp),
                    contentScale = ContentScale.Fit,
                )
            }
        }
        when {
            !trailingLabel.isNullOrBlank() && onTrailing != null -> {
                Text(
                    text = trailingLabel,
                    color = HeaderOnPrimary,
                    fontSize = 16.sp,
                    modifier = Modifier
                        .align(Alignment.CenterEnd)
                        .clickable(
                            interactionSource = remember { MutableInteractionSource() },
                            indication = null,
                            onClick = onTrailing,
                        ),
                )
            }
            onSettings != null -> {
                Image(
                    painter = painterResource(OpsIcon.Setting),
                    contentDescription = "settings",
                    modifier = Modifier
                        .align(Alignment.CenterEnd)
                        .size(20.dp)
                        .clickable(
                            interactionSource = remember { MutableInteractionSource() },
                            indication = null,
                            onClick = onSettings,
                        ),
                    contentScale = ContentScale.Fit,
                )
            }
        }
    }
}

@Composable
private fun WorkbenchProfileHeader(
    companyName: String,
    userName: String,
    roleLabel: String,
    avatar: OpsIcon,
) {
    val colors = OpsTheme.colors
    val density = LocalDensity.current
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .height(157.dp),
    ) {
        Canvas(modifier = Modifier.fillMaxSize()) {
            val w = size.width
            val h = size.height
            val topBand = with(density) { 50.dp.toPx() }
            val ovalExtra = with(density) { 200.dp.toPx() }
            drawRect(color = colors.primary, size = Size(w, topBand))
            drawOval(
                color = colors.primary,
                topLeft = Offset(-ovalExtra, 0f),
                size = Size(w + ovalExtra * 2f, h),
            )
        }
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(start = 12.dp, top = 9.dp, end = 12.dp, bottom = 78.dp),
            verticalAlignment = Alignment.Top,
        ) {
            Image(
                painter = painterResource(avatar),
                contentDescription = null,
                modifier = Modifier
                    .size(70.dp)
                    .clip(CircleShape)
                    .background(Color.White, CircleShape),
                contentScale = ContentScale.Crop,
            )
            Column(
                modifier = Modifier.padding(start = 12.dp, top = 5.dp),
            ) {
                Text(
                    text = companyName.ifBlank { "-" },
                    color = HeaderOnPrimary,
                    fontSize = 20.sp,
                    fontWeight = FontWeight.Normal,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
                Spacer(modifier = Modifier.height(4.dp))
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text(
                        text = userName.ifBlank { "-" },
                        color = HeaderOnPrimary,
                        fontSize = 14.sp,
                    )
                    if (roleLabel.isNotBlank()) {
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(
                            text = roleLabel,
                            color = HeaderOnPrimary,
                            fontSize = 10.sp,
                            modifier = Modifier
                                .border(1.dp, RoleStroke, RoundedCornerShape(12.dp))
                                .padding(horizontal = 10.dp, vertical = 3.dp),
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun WorkbenchModuleCard(
    title: String,
    items: List<WorkbenchModuleItem>,
    emptyIcon: OpsIcon? = null,
    editLabel: String? = null,
    onEditClick: (() -> Unit)? = null,
    onEmptyClick: (() -> Unit)? = null,
    titleIndentExtra: Boolean = false,
    dragEnabled: Boolean,
    onReorder: (from: Int, to: Int) -> Unit,
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(Color.White, RoundedCornerShape(8.dp)),
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .padding(
                    start = if (titleIndentExtra) 28.dp else 12.dp,
                    end = 16.dp,
                    top = 16.dp,
                ),
        ) {
            Text(
                text = title,
                color = Color.Black,
                fontSize = 16.sp,
                fontWeight = FontWeight.Bold,
                modifier = Modifier.align(Alignment.CenterStart),
            )
            if (!editLabel.isNullOrBlank() && onEditClick != null) {
                Text(
                    text = editLabel,
                    color = EditBlue,
                    fontSize = 14.sp,
                    fontWeight = FontWeight.Bold,
                    modifier = Modifier
                        .align(Alignment.CenterEnd)
                        .clickable(
                            interactionSource = remember { MutableInteractionSource() },
                            indication = null,
                            onClick = onEditClick,
                        ),
                )
            }
        }
        if (items.isEmpty() && emptyIcon != null) {
            Image(
                painter = painterResource(emptyIcon),
                contentDescription = null,
                modifier = Modifier
                    .padding(start = 12.dp, top = 12.dp, bottom = 12.dp)
                    .size(56.dp)
                    .then(
                        if (onEmptyClick != null) {
                            Modifier.clickable(
                                interactionSource = remember { MutableInteractionSource() },
                                indication = null,
                                onClick = onEmptyClick,
                            )
                        } else {
                            Modifier
                        },
                    ),
            )
        } else {
            WorkbenchModuleGrid(
                items = items,
                dragEnabled = dragEnabled,
                onReorder = onReorder,
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 12.dp)
                    .padding(top = 12.dp, bottom = 12.dp),
            )
        }
    }
}

@Composable
private fun WorkbenchModuleGrid(
    items: List<WorkbenchModuleItem>,
    dragEnabled: Boolean,
    onReorder: (from: Int, to: Int) -> Unit,
    modifier: Modifier = Modifier,
) {
    var gridSize by remember { mutableStateOf(IntSize.Zero) }
    var draggingIndex by remember { mutableStateOf<Int?>(null) }
    var dragOffset by remember { mutableStateOf(Offset.Zero) }
    val rows = remember(items.size) { (items.size + 3) / 4 }
    val density = LocalDensity.current
    val minCellHeightPx = with(density) { 78.dp.toPx() }

    BoxWithConstraints(
        modifier = modifier.onSizeChanged { gridSize = it },
    ) {
        val cellW = maxWidth / 4
        val cellWPx = with(density) { cellW.toPx() }
        val cellHPx = if (rows > 0 && gridSize.height > 0) {
            (gridSize.height / rows.toFloat()).coerceAtLeast(minCellHeightPx)
        } else {
            minCellHeightPx
        }

        Column(modifier = Modifier.fillMaxWidth()) {
            items.chunked(4).forEachIndexed { rowIndex, row ->
                Row(modifier = Modifier.fillMaxWidth()) {
                    row.forEachIndexed { colIndex, item ->
                        val index = rowIndex * 4 + colIndex
                        val isDragging = draggingIndex == index
                        WorkbenchGridCell(
                            item = item,
                            clickEnabled = draggingIndex == null,
                            modifier = Modifier
                                .width(cellW)
                                .zIndex(if (isDragging) 2f else 0f)
                                .graphicsLayer {
                                    if (isDragging) {
                                        translationX = dragOffset.x
                                        translationY = dragOffset.y
                                        shadowElevation = 10f
                                        scaleX = 1.05f
                                        scaleY = 1.05f
                                    }
                                }
                                .then(
                                    if (dragEnabled) {
                                        Modifier.pointerInput(items.map { it.id }, index) {
                                            detectDragGesturesAfterLongPress(
                                                onDragStart = {
                                                    draggingIndex = index
                                                    dragOffset = Offset.Zero
                                                },
                                                onDragCancel = {
                                                    draggingIndex = null
                                                    dragOffset = Offset.Zero
                                                },
                                                onDragEnd = {
                                                    val from = draggingIndex
                                                    val offset = dragOffset
                                                    draggingIndex = null
                                                    dragOffset = Offset.Zero
                                                    if (from == null || cellWPx <= 0f) return@detectDragGesturesAfterLongPress
                                                    val startX = (from % 4) * cellWPx + cellWPx / 2f
                                                    val startY = (from / 4) * cellHPx + cellHPx / 2f
                                                    val endX = startX + offset.x
                                                    val endY = startY + offset.y
                                                    val toCol = (endX / cellWPx).roundToInt()
                                                        .coerceIn(0, 3)
                                                    val toRow = (endY / cellHPx).roundToInt()
                                                        .coerceIn(0, (rows - 1).coerceAtLeast(0))
                                                    val to = (toRow * 4 + toCol).coerceIn(0, items.lastIndex)
                                                    if (to != from) onReorder(from, to)
                                                },
                                                onDrag = { change, amount ->
                                                    change.consume()
                                                    dragOffset += amount
                                                },
                                            )
                                        }
                                    } else {
                                        Modifier
                                    },
                                ),
                        )
                    }
                    repeat(4 - row.size) {
                        Spacer(modifier = Modifier.width(cellW))
                    }
                }
            }
        }
    }
}

@Composable
private fun WorkbenchGridCell(
    item: WorkbenchModuleItem,
    clickEnabled: Boolean,
    modifier: Modifier = Modifier,
) {
    Column(
        modifier = modifier
            .padding(bottom = 12.dp)
            .then(
                if (clickEnabled) {
                    Modifier.clickable(
                        interactionSource = remember { MutableInteractionSource() },
                        indication = null,
                        onClick = item.onClick,
                    )
                } else {
                    Modifier
                },
            ),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .padding(top = 10.dp),
            contentAlignment = Alignment.TopCenter,
        ) {
            Image(
                painter = painterResource(item.icon),
                contentDescription = item.title,
                modifier = Modifier.size(30.dp),
                contentScale = ContentScale.Fit,
            )
            when (item.editBadge) {
                WorkbenchEditBadge.Add -> {
                    Image(
                        painter = painterResource(OpsIcon.MineModuleEditAdd),
                        contentDescription = null,
                        modifier = Modifier
                            .align(Alignment.TopEnd)
                            .padding(end = 8.dp)
                            .size(24.dp)
                            .clickable(
                                interactionSource = remember { MutableInteractionSource() },
                                indication = null,
                                onClick = { item.onBadgeClick?.invoke() },
                            ),
                    )
                }
                WorkbenchEditBadge.Remove -> {
                    Image(
                        painter = painterResource(OpsIcon.MineModuleEditDelete),
                        contentDescription = null,
                        modifier = Modifier
                            .align(Alignment.TopEnd)
                            .padding(end = 8.dp)
                            .size(24.dp)
                            .clickable(
                                interactionSource = remember { MutableInteractionSource() },
                                indication = null,
                                onClick = { item.onBadgeClick?.invoke() },
                            ),
                    )
                }
                WorkbenchEditBadge.None -> Unit
            }
        }
        Text(
            text = item.title,
            color = Color.Black,
            fontSize = 14.sp,
            textAlign = TextAlign.Center,
            modifier = Modifier.padding(horizontal = 2.dp),
        )
    }
}
