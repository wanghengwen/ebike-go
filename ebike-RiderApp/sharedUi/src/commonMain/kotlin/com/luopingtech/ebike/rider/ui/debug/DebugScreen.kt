package com.luopingtech.ebike.rider.ui.debug

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.safeDrawingPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.luopingtech.ebike.rider.ProbeResult
import com.luopingtech.ebike.rider.RiderApp
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.ui.permission.LocalRiderBlePermissionGate
import com.luopingtech.ebike.rider.ui.theme.RiderTheme
import kotlinx.coroutines.launch

/**
 * P0 的唯一一屏。它不是给用户看的，是给「配置有没有进来 / 签名有没有算对 /
 * 请求有没有发出去」这三个问题一个当场可看的答案 —— 两端宿主（Android Activity 与
 * iOS ViewController）都渲染它，所以任一端起不来都会在这里露出来。
 */
@Composable
fun DebugScreen(
    app: RiderApp,
    modifier: Modifier = Modifier,
    onLogout: (() -> Unit)? = null,
    onOpenWallet: (() -> Unit)? = null,
    onOpenHelp: (() -> Unit)? = null,
) {
    var probe by remember { mutableStateOf<ProbeResult?>(null) }
    var running by remember { mutableStateOf(false) }
    var bleLine by remember { mutableStateOf<String?>(null) }
    var bleRunning by remember { mutableStateOf(false) }
    val scope = rememberCoroutineScope()
    val authState by app.authFeature.state.collectAsState()
    val blePermissionGate = LocalRiderBlePermissionGate.current

    Column(
        modifier = modifier
            .fillMaxSize()
            // iOS 宿主是 .ignoresSafeArea() 的整屏 Compose，不留这层 inset 标题会压在状态栏下面。
            .safeDrawingPadding()
            .verticalScroll(rememberScrollState())
            .padding(20.dp),
    ) {
        Text(
            text = app.config.app.displayName.ifBlank { "RiderApp" },
            style = MaterialTheme.typography.headlineSmall,
            fontWeight = FontWeight.SemiBold,
            color = RiderTheme.colors.textPrimary,
        )
        Text(
            text = "P6 H5 桥 · shared ${RiderApp.LIBRARY_VERSION}",
            style = MaterialTheme.typography.bodySmall,
            color = RiderTheme.colors.textTertiary,
        )

        Spacer(Modifier.height(16.dp))
        HorizontalDivider(color = RiderTheme.colors.divider)
        Spacer(Modifier.height(12.dp))

        Field("tenant.alias", app.config.alias.ifBlank { "(空)" })
        Field("tenantId", app.config.tenantId.ifBlank { "(空)" })
        Field("api.baseUrl", app.config.api.baseUrl.ifBlank { "(空 → demo)" })
        Field("h5.baseUrl", app.config.h5.baseUrl.ifBlank { "(空 → 占位说明)" })
        Field("signSecret", if (app.config.auth.signSecret.isNotBlank()) "已配置" else "(空)")
        Field("deviceId", app.deviceId)
        Field("platform", "${app.deviceInfo.platform} ${app.deviceInfo.osVersion}")
        Field("language", app.i18n.language.tag)
        val session = authState.session
        if (session != null) {
            Field("phone", session.phone.ifBlank { "(空)" })
            Field("pin", session.pin.ifBlank { "(空)" })
        }

        if (onLogout != null || onOpenWallet != null || onOpenHelp != null) {
            Spacer(Modifier.height(12.dp))
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                onOpenWallet?.let {
                    Button(onClick = it) { Text(app.i18n.t(Str.MyWallet)) }
                }
                onOpenHelp?.let {
                    Button(onClick = it) { Text(app.i18n.t(Str.HelpCenter)) }
                }
            }
            if (onLogout != null) {
                Spacer(Modifier.height(8.dp))
                Button(onClick = onLogout) {
                    Text(app.i18n.t(Str.Logout))
                }
            }
        }

        Spacer(Modifier.height(20.dp))
        HorizontalDivider(color = RiderTheme.colors.divider)
        Spacer(Modifier.height(12.dp))

        Text(
            text = "签名探针",
            style = MaterialTheme.typography.titleMedium,
            color = RiderTheme.colors.textPrimary,
        )
        Text(
            text = RiderApp.DEFAULT_PROBE_PATH,
            style = MaterialTheme.typography.bodySmall,
            fontFamily = FontFamily.Monospace,
            color = RiderTheme.colors.textTertiary,
        )

        Spacer(Modifier.height(12.dp))
        Row(verticalAlignment = Alignment.CenterVertically) {
            Button(
                enabled = !running,
                onClick = {
                    running = true
                    scope.launch {
                        probe = app.probe()
                        running = false
                    }
                },
            ) {
                Text(if (probe == null) "发一次探针请求" else "重发")
            }
            if (running) {
                Spacer(Modifier.width(12.dp))
                CircularProgressIndicator(modifier = Modifier.height(20.dp).width(20.dp))
            }
        }

        Spacer(Modifier.height(12.dp))
        probe?.let { ProbeReport(it) }

        Spacer(Modifier.height(20.dp))
        HorizontalDivider(color = RiderTheme.colors.divider)
        Spacer(Modifier.height(12.dp))
        Text(
            text = app.i18n.t(Str.BleUnlock),
            style = MaterialTheme.typography.titleMedium,
            color = RiderTheme.colors.textPrimary,
        )
        Text(
            text = "demo IMEI 867567046128534 · ${if (app.isDemoMode) "SimulatorBle" else app.bleTransport::class.simpleName}",
            style = MaterialTheme.typography.bodySmall,
            fontFamily = FontFamily.Monospace,
            color = RiderTheme.colors.textTertiary,
        )
        Spacer(Modifier.height(12.dp))
        Row(verticalAlignment = Alignment.CenterVertically) {
            Button(
                enabled = !bleRunning,
                onClick = {
                    blePermissionGate.ensure { msg ->
                        if (msg != null) {
                            bleLine = msg
                            return@ensure
                        }
                        bleRunning = true
                        scope.launch {
                            bleLine = when (val result = app.bleSession.unlock(DEMO_BLE_IMEI, mute = true)) {
                                is RiderResult.Ok ->
                                    app.i18n.t(Str.BleUnlockOk) + " code=${result.value.code} hex=${result.value.hex}"
                                is RiderResult.Err ->
                                    app.i18n.t(Str.BleUnlockFailed, result.error.message)
                            }
                            bleRunning = false
                        }
                    }
                },
            ) {
                Text(app.i18n.t(Str.BleUnlock))
            }
            if (bleRunning) {
                Spacer(Modifier.width(12.dp))
                CircularProgressIndicator(modifier = Modifier.height(20.dp).width(20.dp))
            }
        }
        bleLine?.let {
            Spacer(Modifier.height(8.dp))
            Field("BLE", it)
        }
    }
}

@Composable
private fun ProbeReport(probe: ProbeResult) {
    Column(verticalArrangement = Arrangement.spacedBy(2.dp)) {
        Field("_t", probe.timestamp.ifBlank { "(未发出)" })
        Field("_s", probe.signDigest.ifBlank { "(未发出)" })
        Field(
            label = "HTTP",
            value = when {
                probe.isDemo -> "— (demo mode)"
                probe.httpStatus != null -> probe.httpStatus.toString()
                else -> "无响应"
            },
        )
        Field("业务码", probe.businessCode.ifBlank { "—" })
        if (probe.message.isNotBlank()) Field("msg", probe.message)
        probe.error?.let { Field("error", it) }
        if (probe.responsePreview.isNotBlank()) Field("响应", probe.responsePreview)
        Spacer(Modifier.height(6.dp))
        Field("请求体", probe.requestBody)
    }
}

private const val DEMO_BLE_IMEI = "867567046128534"

@Composable
private fun Field(label: String, value: String) {
    Row(modifier = Modifier.fillMaxWidth().padding(vertical = 3.dp)) {
        Text(
            text = label,
            modifier = Modifier.width(96.dp),
            style = MaterialTheme.typography.bodySmall,
            color = RiderTheme.colors.textTertiary,
        )
        Text(
            text = value,
            modifier = Modifier.weight(1f),
            style = MaterialTheme.typography.bodySmall,
            fontFamily = FontFamily.Monospace,
            color = RiderTheme.colors.textPrimary,
        )
    }
}
