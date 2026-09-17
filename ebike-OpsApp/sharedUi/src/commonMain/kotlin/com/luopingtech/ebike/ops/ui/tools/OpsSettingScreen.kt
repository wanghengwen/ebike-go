package com.luopingtech.ebike.ops.ui.tools

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Switch
import androidx.compose.material3.SwitchDefaults
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.ui.analysis.VcdTopBar
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import kotlinx.coroutines.launch

/**
 * 对齐 OperationSettingActivity：列表行进阈值子页；开关即时保存。
 */
@Composable
fun OpsSettingScreen(
    app: OpsApp,
    onClose: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val home by app.homeFeature.state.collectAsState()
    val state by app.opsSettingFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    val feature = app.opsSettingFeature
    val colors = OpsTheme.colors
    var editingThreshold by remember { mutableStateOf(false) }
    var draftThreshold by remember { mutableStateOf("") }

    LaunchedEffect(home.currentArea?.id) {
        feature.load(home.currentArea)
    }

    if (editingThreshold) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.White),
        ) {
            VcdTopBar(
                title = t(Str.SwapThreshold),
                onBack = { editingThreshold = false },
            )
            OutlinedTextField(
                value = draftThreshold,
                onValueChange = { draftThreshold = it.filter { ch -> ch.isDigit() }.take(3) },
                placeholder = { Text(t(Str.ThresholdInputHint), color = Color(0xFF999999)) },
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp)
                    .padding(top = 32.dp),
                singleLine = true,
            )
            Box(modifier = Modifier.weight(1f))
            Button(
                onClick = {
                    draftThreshold.toIntOrNull()?.let { feature.setThreshold(it) }
                    scope.launch {
                        feature.save(home.currentArea)
                        editingThreshold = false
                    }
                },
                enabled = !state.saving && draftThreshold.isNotBlank(),
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 32.dp, vertical = 24.dp),
                colors = ButtonDefaults.buttonColors(
                    containerColor = colors.primary,
                    contentColor = colors.onPrimary,
                ),
            ) {
                Text(if (state.saving) t(Str.Submitting) else t(Str.Confirm))
            }
        }
        return
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(Color(0xFFF8F8F8)),
    ) {
        VcdTopBar(title = t(Str.OpsSettingTitle), onBack = {
            feature.clear()
            onClose()
        })
        if (state.loading) {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                CircularProgressIndicator(color = colors.primary)
            }
        } else {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .verticalScroll(rememberScrollState()),
            ) {
                Text(
                    text = t(Str.SwapSettingSection),
                    color = Color(0xFF999999),
                    fontSize = 12.sp,
                    modifier = Modifier.padding(start = 16.dp, top = 16.dp, bottom = 8.dp),
                )
                OpsSettingRow(
                    title = t(Str.SwapThreshold),
                    trailing = "${state.threshold}%",
                    showArrow = true,
                    introduce = t(Str.SwapThresholdHint),
                    onClick = {
                        draftThreshold = state.threshold.toString()
                        editingThreshold = true
                    },
                )
                OpsSettingRow(
                    title = t(Str.AutoSwapTask),
                    introduce = t(Str.AutoSwapHint),
                    showDivider = false,
                    switchChecked = state.autoSwap,
                    onSwitch = { enabled ->
                        feature.setAutoSwap(enabled)
                        scope.launch { feature.save(home.currentArea) }
                    },
                )
                state.message?.let {
                    Text(
                        it,
                        color = colors.primary,
                        modifier = Modifier.padding(16.dp),
                    )
                }
                state.errorMessage?.let {
                    Text(
                        it,
                        color = Color(0xFFE02020),
                        modifier = Modifier.padding(16.dp),
                    )
                }
            }
        }
    }
}

@Composable
private fun OpsSettingRow(
    title: String,
    introduce: String,
    trailing: String? = null,
    showArrow: Boolean = false,
    showDivider: Boolean = true,
    switchChecked: Boolean? = null,
    onClick: (() -> Unit)? = null,
    onSwitch: ((Boolean) -> Unit)? = null,
) {
    val colors = OpsTheme.colors
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(Color.White)
            .then(if (onClick != null) Modifier.clickable(onClick = onClick) else Modifier)
            .padding(horizontal = 16.dp)
            .padding(top = 16.dp),
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(
                text = title,
                color = Color(0xFF242936),
                fontSize = 16.sp,
                fontWeight = FontWeight.Bold,
                modifier = Modifier.weight(1f),
            )
            if (trailing != null) {
                Text(
                    text = trailing,
                    color = Color(0xFF999999),
                    fontSize = 14.sp,
                    modifier = Modifier.padding(end = 4.dp),
                )
            }
            if (showArrow) {
                Text("›", color = Color(0xFF999999), fontSize = 22.sp)
            }
            if (switchChecked != null && onSwitch != null) {
                Switch(
                    checked = switchChecked,
                    onCheckedChange = onSwitch,
                    colors = SwitchDefaults.colors(checkedTrackColor = colors.primary),
                )
            }
        }
        Text(
            text = introduce,
            color = Color(0xFF666666),
            fontSize = 13.sp,
            modifier = Modifier.padding(top = 8.dp, bottom = 16.dp),
        )
        if (showDivider) {
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(1.dp)
                    .background(Color(0xFFE5E5E5)),
            )
        }
    }
}
