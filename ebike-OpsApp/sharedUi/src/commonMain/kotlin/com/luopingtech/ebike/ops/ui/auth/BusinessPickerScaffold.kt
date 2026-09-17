package com.luopingtech.ebike.ops.ui.auth

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.domain.model.BusinessTenant
import com.luopingtech.ebike.ops.ui.icons.OpsBackChevron
import com.luopingtech.ebike.ops.ui.text.labelComparator
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import kotlinx.coroutines.launch

private val IndexLetters = listOf("#") + ('A'..'Z').map { it.toString() }

@Composable
fun BusinessPickerScaffold(
    title: String,
    searchHint: String,
    query: String,
    onQueryChange: (String) -> Unit,
    businesses: List<BusinessTenant>,
    loading: Boolean,
    errorMessage: String?,
    backLabel: String,
    onBack: () -> Unit,
    onSelect: (String) -> Unit,
) {
    val colors = OpsTheme.colors
    val collator = remember { labelComparator() }
    val sorted = remember(businesses, query) {
        val q = query.trim()
        businesses
            .filter {
                q.isEmpty() ||
                    it.displayLabel.contains(q, ignoreCase = true) ||
                    it.tenantName.contains(q, ignoreCase = true) ||
                    it.companyName.contains(q, ignoreCase = true)
            }
            .sortedWith { a, b -> collator.compare(a.displayLabel, b.displayLabel) }
    }
    val sections = remember(sorted) { groupByIndexLetter(sorted) }
    val listState = rememberLazyListState()
    val scope = rememberCoroutineScope()

    com.luopingtech.ebike.ops.ui.navigation.OpsBackHandler(onBack = onBack)

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(colors.pageBackground),
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .background(colors.primary)
                .statusBarsPadding(),
        ) {
            Column(modifier = Modifier.fillMaxWidth()) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(48.dp),
                ) {
                    OpsBackChevron(
                        onClick = onBack,
                        modifier = Modifier.align(Alignment.CenterStart),
                    )
                    Text(
                        text = title,
                        color = colors.onPrimary,
                        fontSize = 18.sp,
                        fontWeight = FontWeight.Medium,
                        modifier = Modifier.align(Alignment.Center),
                    )
                    Text(
                        text = backLabel,
                        color = Color.Transparent,
                        modifier = Modifier.align(Alignment.CenterEnd).padding(end = 16.dp),
                    )
                }
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(68.dp)
                        .padding(horizontal = 16.dp, vertical = 16.dp)
                        .background(Color.White.copy(alpha = 0.22f), RoundedCornerShape(4.dp))
                        .padding(horizontal = 12.dp),
                    contentAlignment = Alignment.CenterStart,
                ) {
                    if (query.isEmpty()) {
                        Text(
                            text = searchHint,
                            color = Color.White.copy(alpha = 0.7f),
                            fontSize = 14.sp,
                        )
                    }
                    BasicTextField(
                        value = query,
                        onValueChange = { if (it.length <= 9) onQueryChange(it) },
                        singleLine = true,
                        textStyle = TextStyle(color = Color.White, fontSize = 14.sp),
                        cursorBrush = SolidColor(Color.White),
                        modifier = Modifier.fillMaxWidth(),
                    )
                }
            }
        }

        Row(
            modifier = Modifier
                .weight(1f)
                .fillMaxWidth(),
        ) {
            LazyColumn(
                state = listState,
                modifier = Modifier
                    .weight(1f)
                    .fillMaxHeight(),
            ) {
                sections.forEach { (letter, items) ->
                    item(key = "h-$letter") {
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(24.dp)
                                .background(Color(0xFFF5F7FA))
                                .padding(start = 15.dp),
                            contentAlignment = Alignment.CenterStart,
                        ) {
                            Text(text = letter, color = Color(0xFF999999), fontSize = 14.sp)
                        }
                    }
                    items(items, key = { it.tenantId }) { biz ->
                        Column(
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(40.dp)
                                .clickable(enabled = !loading) { onSelect(biz.tenantId) }
                                .padding(start = 15.dp),
                        ) {
                            Box(
                                modifier = Modifier
                                    .weight(1f)
                                    .fillMaxWidth(),
                                contentAlignment = Alignment.CenterStart,
                            ) {
                                Text(
                                    text = biz.displayLabel,
                                    color = Color.Black,
                                    fontSize = 17.sp,
                                    maxLines = 1,
                                    overflow = TextOverflow.Ellipsis,
                                )
                            }
                            Box(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .height(0.5.dp)
                                    .background(Color(0x80727272)),
                            )
                        }
                    }
                }
            }
            Column(
                modifier = Modifier
                    .width(25.dp)
                    .fillMaxHeight()
                    .padding(vertical = 8.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
            ) {
                IndexLetters.forEach { letter ->
                    val enabled = sections.any { it.first == letter }
                    Text(
                        text = letter,
                        color = if (enabled) colors.primary else Color(0xFF999DA1),
                        fontSize = 10.sp,
                        modifier = Modifier
                            .padding(vertical = 1.dp)
                            .clickable(
                                enabled = enabled,
                                interactionSource = remember { MutableInteractionSource() },
                                indication = null,
                                onClick = {
                                    val index = flatIndexOfSection(sections, letter)
                                    if (index >= 0) {
                                        scope.launch { listState.animateScrollToItem(index) }
                                    }
                                },
                            ),
                    )
                }
            }
        }

        errorMessage?.let {
            Text(
                text = it,
                color = Color(0xFFE53935),
                fontSize = 13.sp,
                modifier = Modifier.padding(16.dp),
            )
        }
    }
}

private fun indexLetterOf(name: String): String {
    val first = name.trim().firstOrNull() ?: return "#"
    val upper = first.uppercaseChar()
    return if (upper in 'A'..'Z') upper.toString() else "#"
}

private fun groupByIndexLetter(
    tenants: List<BusinessTenant>,
): List<Pair<String, List<BusinessTenant>>> {
    val map = linkedMapOf<String, MutableList<BusinessTenant>>()
    IndexLetters.forEach { map[it] = mutableListOf() }
    tenants.forEach { t ->
        map.getOrPut(indexLetterOf(t.displayLabel)) { mutableListOf() }.add(t)
    }
    return IndexLetters.mapNotNull { letter ->
        val items = map[letter].orEmpty()
        if (items.isEmpty()) null else letter to items
    }
}

private fun flatIndexOfSection(
    sections: List<Pair<String, List<BusinessTenant>>>,
    letter: String,
): Int {
    var index = 0
    sections.forEach { (section, items) ->
        if (section == letter) return index
        index += 1 + items.size
    }
    return -1
}

