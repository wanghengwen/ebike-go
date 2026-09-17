package com.luopingtech.ebike.ops.ui.admin

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.domain.admin.BlacklistItem
import com.luopingtech.ebike.ops.ui.analysis.VcdTopBar
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import kotlinx.coroutines.launch

@Composable
fun BlacklistScreen(app: OpsApp, onClose: () -> Unit) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val home by app.homeFeature.state.collectAsState()
    val state by app.blacklistFeature.state.collectAsState()
    val feature = app.blacklistFeature
    val scope = rememberCoroutineScope()
    val primary = OpsTheme.colors.primary

    LaunchedEffect(home.currentArea?.id) { feature.load(home.currentArea) }

    Column(
        Modifier
            .fillMaxSize()
            .background(Color(0xFFF6F7F9)),
    ) {
        VcdTopBar(
            title = t(Str.Blacklist),
            onBack = {
                feature.clear()
                onClose()
            },
        )
        Column(
            Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState())
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            Row(
                Modifier
                    .fillMaxWidth()
                    .height(40.dp)
                    .border(1.dp, Color(0xFFDCDCDC), RoundedCornerShape(5.dp))
                    .background(Color.White, RoundedCornerShape(5.dp)),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                BasicTextField(
                    value = state.keyword,
                    onValueChange = { value ->
                        feature.setKeyword(value)
                        if (value.isBlank()) {
                            scope.launch { feature.search(home.currentArea, "") }
                        }
                    },
                    singleLine = true,
                    textStyle = TextStyle(color = Color(0xFF242936), fontSize = 14.sp),
                    cursorBrush = SolidColor(primary),
                    modifier = Modifier
                        .weight(1f)
                        .padding(horizontal = 8.dp),
                    decorationBox = { inner ->
                        if (state.keyword.isEmpty()) {
                            Text(t(Str.BlacklistSearchHint), color = Color(0xFF9FA7C7), fontSize = 14.sp)
                        }
                        inner()
                    },
                )
                Button(
                    onClick = { scope.launch { feature.search(home.currentArea) } },
                    enabled = state.keyword.isNotBlank() && !state.loading,
                    modifier = Modifier.height(40.dp),
                    shape = RoundedCornerShape(topEnd = 5.dp, bottomEnd = 5.dp),
                ) {
                    Text(t(Str.Search))
                }
            }

            if (state.loading) CircularProgressIndicator(color = primary)
            state.errorMessage?.let { Text(it, color = MaterialTheme.colorScheme.error) }
            state.message?.let { Text(it, color = primary) }

            if (!state.loading && state.items.isEmpty()) {
                Text(t(Str.AdminEmptyList))
            } else {
                state.items.forEach { item ->
                    BlacklistRow(item, cancelLabel = t(Str.BlacklistCancel)) {
                        scope.launch { feature.cancel(home.currentArea, item.id) }
                    }
                }
            }
        }
    }
}

@Composable
private fun BlacklistRow(item: BlacklistItem, cancelLabel: String, onCancel: () -> Unit) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(Color.White, MaterialTheme.shapes.medium)
            .padding(12.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        Text(item.authName.ifBlank { item.phone }, style = MaterialTheme.typography.titleSmall)
        Text(item.phone, style = MaterialTheme.typography.bodySmall)
        Text(item.reason.ifBlank { "-" }, style = MaterialTheme.typography.bodySmall)
        if (item.state == 1) {
            TextButton(onClick = onCancel) { Text(cancelLabel) }
        }
    }
}
