package com.luopingtech.ebike.ops.ui.tag

import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
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
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Checkbox
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
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
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.ui.analysis.VcdTopBar
import com.luopingtech.ebike.ops.ui.icons.OpsIcon
import com.luopingtech.ebike.ops.ui.icons.painterResource
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import com.luopingtech.ebike.ops.ui.tools.LegacyScanPreviewBlock
import kotlinx.coroutines.launch

/**
 * 对齐 VehicleTagActivity：页内连续扫码 + 手电筒 + 待标记列表 + 底部提交。
 * 系统返回由 [VcdTopBar] 统一拦截（等价遗留 finish()）。
 */
@Composable
fun VehicleTagScreen(
    app: OpsApp,
    onClose: () -> Unit,
    scanPreview: @Composable (
        modifier: Modifier,
        torchOn: Boolean,
        enabled: Boolean,
        onCode: (String) -> Unit,
    ) -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val home by app.homeFeature.state.collectAsState()
    val state by app.vehicleTagFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    val feature = app.vehicleTagFeature
    val colors = OpsTheme.colors
    var torchOn by remember { mutableStateOf(false) }
    var scanEnabled by remember { mutableStateOf(true) }
    var lastCode by remember { mutableStateOf<String?>(null) }

    fun closePage() {
        feature.clear()
        onClose()
    }

    LaunchedEffect(home.currentArea?.id) {
        feature.load(home.currentArea)
    }

    fun onScanned(raw: String) {
        if (!scanEnabled || state.loading) return
        if (raw == lastCode) return
        lastCode = raw
        scanEnabled = false
        scope.launch {
            feature.applyScan(raw)
            feature.addToPending(home.currentArea)
            scanEnabled = true
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.White),
    ) {
        VcdTopBar(title = t(Str.VehicleTagTool), onBack = ::closePage)
        LegacyScanPreviewBlock(
            torchOn = torchOn,
            torchLabel = t(Str.Torch),
            scanEnabled = scanEnabled && !state.loading,
            onTorchChange = { torchOn = it },
            scanPreview = scanPreview,
            onCode = ::onScanned,
        )
        // Legacy VehicleTagActivity shows scan + list immediately; type list loads in background.
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp)
                .padding(top = 24.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(10.dp),
        ) {
            OutlinedTextField(
                value = state.carId,
                onValueChange = { feature.setCarId(it) },
                placeholder = { Text(t(Str.VehicleTagCarId)) },
                modifier = Modifier
                    .weight(1f)
                    .height(56.dp),
                singleLine = true,
            )
            Button(
                onClick = { scope.launch { feature.addToPending(home.currentArea) } },
                enabled = !state.loading && state.carId.isNotBlank(),
                modifier = Modifier.height(40.dp),
                shape = RoundedCornerShape(4.dp),
                colors = ButtonDefaults.buttonColors(
                    containerColor = colors.primary,
                    contentColor = colors.onPrimary,
                ),
            ) { Text(t(Str.VehicleTagAdd), fontSize = 16.sp) }
        }

        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp)
                .padding(top = 24.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(t(Str.VehicleTagCarId), color = Color(0xFF666666), fontSize = 14.sp)
            Text(
                t(Str.VehicleTagActionCol),
                color = Color(0xFF666666),
                fontSize = 14.sp,
                modifier = Modifier.padding(start = 120.dp),
            )
            Spacer(modifier = Modifier.weight(1f))
            Checkbox(
                checked = state.pending.isNotEmpty() && state.pending.all { it.selected },
                onCheckedChange = { feature.selectAllPending(it) },
            )
        }
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .padding(top = 10.dp)
                .height(1.dp)
                .background(Color(0xFF999999)),
        )

        LazyColumn(modifier = Modifier.weight(1f)) {
            items(state.pending, key = { it.carId }) { item ->
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(40.dp)
                        .padding(horizontal = 16.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text(item.carId, color = Color(0xFF333333), fontSize = 16.sp)
                    Text(
                        t(Str.VehicleTagDelete),
                        color = Color(0xFFE02020),
                        fontSize = 16.sp,
                        modifier = Modifier
                            .padding(start = 120.dp)
                            .clickable { feature.removePending(item.carId) },
                    )
                    Spacer(modifier = Modifier.weight(1f))
                    Checkbox(
                        checked = item.selected,
                        onCheckedChange = { feature.togglePending(item.carId) },
                    )
                }
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(1.dp)
                        .background(Color(0xFFCCCCCC)),
                )
            }
        }

        Button(
            onClick = { feature.openTypePicker(true) },
            enabled = !state.submitting && state.pending.any { it.selected },
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 32.dp, vertical = 16.dp)
                .height(48.dp),
            colors = ButtonDefaults.buttonColors(
                containerColor = colors.primary,
                contentColor = colors.onPrimary,
                disabledContainerColor = Color(0xFFCCCCCC),
            ),
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Image(
                    painter = painterResource(OpsIcon.VehicleAddIc),
                    contentDescription = null,
                    modifier = Modifier.size(20.dp),
                    contentScale = ContentScale.Fit,
                )
                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    if (state.submitting) t(Str.LoadingEllipsis) else t(Str.VehicleTagSubmit),
                    fontWeight = FontWeight.Bold,
                    fontSize = 16.sp,
                )
            }
        }
        state.message?.let {
            Text(it, color = colors.primary, modifier = Modifier.padding(horizontal = 16.dp))
        }
        // Only action errors (add/submit); type-list failures stay silent like ManagerApp.
        state.errorMessage?.let {
            Text(it, color = Color(0xFFE02020), modifier = Modifier.padding(horizontal = 16.dp))
        }
    }

    if (state.showTypePicker) {
        Dialog(onDismissRequest = { feature.openTypePicker(false) }) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(Color.White, RoundedCornerShape(8.dp))
                    .padding(16.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                Text(t(Str.VehicleTagPickType), fontSize = 16.sp, fontWeight = FontWeight.Medium)
                state.types.forEach { type ->
                    Text(
                        type.name,
                        modifier = Modifier
                            .fillMaxWidth()
                            .clickable {
                                scope.launch {
                                    feature.submitSelected(home.currentArea, type.id)
                                }
                            }
                            .padding(vertical = 12.dp),
                        fontSize = 16.sp,
                    )
                }
                TextButton(
                    onClick = { feature.openTypePicker(false) },
                    modifier = Modifier.align(Alignment.End),
                ) { Text(t(Str.Cancel)) }
            }
        }
    }
}
