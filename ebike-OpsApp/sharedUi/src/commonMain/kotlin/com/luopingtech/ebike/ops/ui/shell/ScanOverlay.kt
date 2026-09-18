package com.luopingtech.ebike.ops.ui.shell

import com.luopingtech.ebike.ops.OpsApp
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.border
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
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.OutlinedTextFieldDefaults
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
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.foundation.layout.size
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.permission.OpsPermissions
import com.luopingtech.ebike.ops.ui.icons.OpsIcon
import com.luopingtech.ebike.ops.ui.icons.painterResource
import com.luopingtech.ebike.ops.ui.scan.LocalOpsScanPreview
import com.luopingtech.ebike.ops.ui.vehicle.LocalOpenVehicleDetail
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

private enum class ScanMode { Detail, Unlock, Lock }

@Composable
internal fun ScanOverlay(
    app: OpsApp,
    permissions: OpsPermissions,
    onClose: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    var mode by remember {
        // 对齐遗留 ScanActivity.initPermissionView：有详情优先详情，否则开锁。
        mutableStateOf(
            if (!permissions.canScanDetails && permissions.canScanUnlock) {
                ScanMode.Unlock
            } else {
                ScanMode.Detail
            },
        )
    }
    val scope = rememberCoroutineScope()
    val scanState by app.scanFeature.state.collectAsState()
    val homeState by app.homeFeature.state.collectAsState()
    val openVehicleDetail = LocalOpenVehicleDetail.current

    var torchOn by remember { mutableStateOf(false) }
    var showManual by remember { mutableStateOf(false) }
    var manualInput by remember { mutableStateOf("") }
    // 对齐遗留 preScanResult：同一串不重复处理；切模式时清空。
    var lastRaw by remember { mutableStateOf("") }
    var busy by remember { mutableStateOf(false) }

    fun resetDedupe() {
        lastRaw = ""
        busy = false
    }

    LaunchedEffect(mode) {
        resetDedupe()
        app.scanFeature.clear()
        showManual = false
    }

    fun handleSystemBack() {
        if (showManual) {
            showManual = false
            return
        }
        onClose()
    }

    com.luopingtech.ebike.ops.ui.navigation.OpsBackHandler(onBack = ::handleSystemBack)

    suspend fun handleRaw(raw: String) {
        val trimmed = raw.trim()
        if (trimmed.isEmpty() || busy || trimmed == lastRaw) return
        lastRaw = trimmed
        busy = true
        when (val result = app.scanFeature.resolveManual(trimmed)) {
            is OpsResult.Err -> {
                busy = false
                // 识别错了允许马上再扫；同码也允许重试。error 已写入 scanFeature.state。
                lastRaw = ""
            }
            is OpsResult.Ok -> when (mode) {
                ScanMode.Detail -> {
                    // 对齐 ScanActivity：校验通过后打开 CarDetailActivity 并 finish 扫码页。
                    openVehicleDetail.open(result.value, homeState.currentArea?.id)
                    busy = false
                    onClose()
                }
                ScanMode.Unlock -> {
                    app.scanFeature.unlock()
                    delay(1_200)
                    app.scanFeature.clear()
                    resetDedupe()
                }
                ScanMode.Lock -> {
                    app.scanFeature.lock()
                    delay(1_200)
                    app.scanFeature.clear()
                    resetDedupe()
                }
            }
        }
    }

    val scanBg = Color(0xFF242936)
    val scanMuted = Color(0xFFCCCCCC)
    val tabStroke = Color(0xFF4B4B4B)

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(scanBg)
            .statusBarsPadding()
            .navigationBarsPadding(),
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 8.dp, vertical = 4.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(
                text = t(Str.ScanTitle),
                color = Color.White,
                style = MaterialTheme.typography.titleMedium,
            )
            TextButton(onClick = onClose) {
                Text(t(Str.Close), color = scanMuted)
            }
        }

        // 上半屏预览：只吃剩余高度，绝不和下方操作区叠层。
        LocalOpsScanPreview.current.Preview(
            modifier = Modifier
                .fillMaxWidth()
                .weight(1f),
            torchOn = torchOn,
            enabled = !busy && !showManual,
            onCode = { raw -> scope.launch { handleRaw(raw) } },
        )

        // 不透明底栏：SurfaceView/预览再怎么画，也盖不住这块。
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .background(scanBg)
                .padding(top = 4.dp),
        ) {
            Text(
                text = when (mode) {
                    ScanMode.Detail -> t(Str.ScanModeDetailHint)
                    ScanMode.Unlock -> t(Str.ScanModeUnlockHint)
                    ScanMode.Lock -> t(Str.ScanModeLockHint)
                },
                color = scanMuted,
                style = MaterialTheme.typography.bodyMedium,
                textAlign = TextAlign.Center,
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 20.dp, vertical = 10.dp),
            )

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 56.dp, vertical = 4.dp),
                horizontalArrangement = Arrangement.SpaceBetween,
            ) {
                ScanActionIcon(
                    icon = OpsIcon.ScanManual,
                    label = t(Str.InputCarNumber),
                    onClick = { showManual = !showManual },
                )
                ScanActionIcon(
                    icon = if (torchOn) OpsIcon.TorchOn else OpsIcon.TorchOff,
                    label = if (torchOn) t(Str.CloseTorch) else t(Str.OpenTorch),
                    onClick = { torchOn = !torchOn },
                )
            }

            if (showManual) {
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 16.dp, vertical = 4.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    OutlinedTextField(
                        value = manualInput,
                        onValueChange = { manualInput = it },
                        label = { Text(t(Str.VehicleIdImeiQr)) },
                        modifier = Modifier.fillMaxWidth(),
                        singleLine = true,
                        textStyle = TextStyle(color = Color.White, fontSize = 16.sp),
                        colors = OutlinedTextFieldDefaults.colors(
                            focusedTextColor = Color.White,
                            unfocusedTextColor = Color.White,
                            disabledTextColor = Color.White.copy(alpha = 0.6f),
                            cursorColor = Color.White,
                            focusedBorderColor = Color(0xFF4A90E2),
                            unfocusedBorderColor = Color(0xFF4A90E2),
                            focusedLabelColor = Color(0xFF7EC8F8),
                            unfocusedLabelColor = Color(0xFF7EC8F8),
                            focusedContainerColor = Color.Transparent,
                            unfocusedContainerColor = Color.Transparent,
                        ),
                    )
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.spacedBy(8.dp),
                    ) {
                        Button(
                            onClick = {
                                scope.launch {
                                    showManual = false
                                    handleRaw(manualInput)
                                }
                            },
                            enabled = !busy && manualInput.isNotBlank(),
                            modifier = Modifier.weight(1f),
                        ) { Text(if (busy) t(Str.LoadingEllipsis) else t(Str.Confirm)) }
                        TextButton(
                            onClick = { showManual = false },
                            modifier = Modifier.weight(1f),
                        ) { Text(t(Str.Cancel), color = scanMuted) }
                    }
                }
            }

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 32.dp, vertical = 16.dp)
                    .border(1.dp, tabStroke, RoundedCornerShape(8.dp))
                    .padding(4.dp),
                horizontalArrangement = Arrangement.spacedBy(4.dp),
            ) {
                if (permissions.canScanDetails) {
                    ScanModeTab(
                        label = t(Str.ScanDetail),
                        selected = mode == ScanMode.Detail,
                        onClick = { mode = ScanMode.Detail },
                        modifier = Modifier.weight(1f),
                    )
                }
                if (permissions.canScanUnlock) {
                    ScanModeTab(
                        label = t(Str.ScanUnlock),
                        selected = mode == ScanMode.Unlock,
                        onClick = { mode = ScanMode.Unlock },
                        modifier = Modifier.weight(1f),
                    )
                    ScanModeTab(
                        label = t(Str.ScanLock),
                        selected = mode == ScanMode.Lock,
                        onClick = { mode = ScanMode.Lock },
                        modifier = Modifier.weight(1f),
                    )
                }
            }

            scanState.message?.let {
                Text(
                    text = it,
                    color = Color.White,
                    style = MaterialTheme.typography.bodySmall,
                    modifier = Modifier.padding(horizontal = 16.dp, vertical = 2.dp),
                )
            }
            scanState.errorMessage?.let {
                Text(
                    text = it,
                    color = MaterialTheme.colorScheme.error,
                    style = MaterialTheme.typography.bodySmall,
                    modifier = Modifier.padding(horizontal = 16.dp, vertical = 4.dp),
                )
            }
        }
    }
}

@Composable
private fun ScanActionIcon(
    icon: OpsIcon,
    label: String,
    onClick: () -> Unit,
) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        modifier = Modifier
            .clickable(onClick = onClick)
            .padding(8.dp),
    ) {
        Image(
            painter = painterResource(icon),
            contentDescription = label,
            modifier = Modifier.size(40.dp),
        )
        Spacer(modifier = Modifier.height(8.dp))
        Text(
            text = label,
            color = Color(0xFFCCCCCC),
            style = MaterialTheme.typography.bodySmall,
        )
    }
}

@Composable
private fun ScanModeTab(
    label: String,
    selected: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Box(
        modifier = modifier
            .clickable(onClick = onClick)
            .background(
                color = if (selected) Color(0xFFB8D4E8) else Color.Transparent,
                shape = RoundedCornerShape(6.dp),
            )
            .padding(vertical = 10.dp),
        contentAlignment = Alignment.Center,
    ) {
        Text(
            text = label,
            color = if (selected) Color.Black else Color.White,
            style = MaterialTheme.typography.labelLarge,
            fontWeight = if (selected) FontWeight.SemiBold else FontWeight.Normal,
        )
    }
}
