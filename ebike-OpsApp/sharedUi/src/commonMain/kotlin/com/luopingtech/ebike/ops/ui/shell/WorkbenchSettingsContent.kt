package com.luopingtech.ebike.ops.ui.shell

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.FilterChip
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Switch
import androidx.compose.material3.SwitchDefaults
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
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.OpsLanguage
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.BusinessTenant
import com.luopingtech.ebike.ops.feature.auth.AuthUiState
import com.luopingtech.ebike.ops.platform.clearOpsDiskCache
import com.luopingtech.ebike.ops.ui.feedback.LocalOpsToast
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import kotlinx.coroutines.launch

/**
 * 对齐遗留 [SettingActivity] / activity_setting.xml 条目。
 */
@Composable
internal fun WorkbenchSettingsContent(
    app: OpsApp,
    authState: AuthUiState,
    onOpenSwitchBusiness: (List<BusinessTenant>) -> Unit,
    onOpenUpdatePassword: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val toast = LocalOpsToast.current
    val scope = rememberCoroutineScope()
    val primary = OpsTheme.colors.primary
    val session = authState.session
    val phone = session?.phone?.takeIf { it.isNotBlank() }
        ?: session?.displayName.orEmpty().ifBlank { "-" }
    val version = app.deviceInfo.appVersion.ifBlank { OpsApp.LIBRARY_VERSION }
    val tenantName = authState.runtimeConfig?.tenantName
        ?.takeIf { it.isNotBlank() }
        ?: app.config.app.displayName.ifBlank { app.config.name }.ifBlank { "-" }

    var performanceOn by remember { mutableStateOf(app.isMapPerformanceMode()) }
    var switchLoading by remember { mutableStateOf(false) }

    Column(verticalArrangement = Arrangement.spacedBy(0.dp)) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color.White, RoundedCornerShape(8.dp))
                .padding(vertical = 4.dp),
        ) {
            SettingInfoRow(label = t(Str.SettingUserInfo), value = phone)
            SettingDivider()
            SettingInfoRow(label = t(Str.SettingVersion), value = version)
            SettingDivider()
            SettingSwitchRow(
                label = t(Str.SettingPerformanceMode),
                checked = performanceOn,
                onCheckedChange = { on ->
                    performanceOn = on
                    app.setMapPerformanceMode(on)
                },
            )
            SettingDivider()
            SettingActionRow(
                label = t(Str.SettingClearCache),
                actionLabel = t(Str.SettingClear),
                primary = primary,
                onAction = {
                    val ok = clearOpsDiskCache()
                    toast(if (ok) t(Str.SettingClearCacheOk) else t(Str.SettingClearCacheFail))
                },
            )
            SettingDivider()
            SettingActionRow(
                label = t(Str.SettingUploadLog),
                actionLabel = t(Str.SettingUpload),
                primary = primary,
                onAction = { toast(t(Str.SettingUploadLogEmpty)) },
            )
            SettingDivider()
            SettingNavRow(
                label = t(Str.Language),
                value = if (language == OpsLanguage.ZH_CN) t(Str.LanguageZh) else t(Str.LanguageEn),
                onClick = {},
                showArrow = false,
            )
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp, vertical = 8.dp),
                horizontalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                FilterChip(
                    selected = language == OpsLanguage.ZH_CN,
                    onClick = { app.i18n.setLanguage(OpsLanguage.ZH_CN) },
                    label = { Text(t(Str.LanguageZh)) },
                )
                FilterChip(
                    selected = language == OpsLanguage.EN,
                    onClick = { app.i18n.setLanguage(OpsLanguage.EN) },
                    label = { Text(t(Str.LanguageEn)) },
                )
            }
            SettingDivider()
            SettingNavRow(
                label = t(Str.ChangePassword),
                value = "",
                onClick = onOpenUpdatePassword,
            )
            SettingDivider()
            // Legacy: if (AppConfig.getIsRoot()) VISIBLE else GONE — izRoot, not codes.
            if (session?.izRoot == true) {
                SettingNavRow(
                    label = t(Str.SettingSwitchBusiness),
                    value = tenantName,
                    valueColor = primary,
                    onClick = {
                        if (switchLoading) return@SettingNavRow
                        scope.launch {
                            switchLoading = true
                            when (val r = app.authFeature.listTenantsForSwitch()) {
                                is OpsResult.Ok -> {
                                    // Legacy: size == 1 → silent return, no toast
                                    if (r.value.size > 1) {
                                        onOpenSwitchBusiness(r.value)
                                    }
                                }
                                is OpsResult.Err -> toast(r.error.message)
                            }
                            switchLoading = false
                        }
                    },
                )
            }
        }

        authState.errorMessage?.let {
            Spacer(modifier = Modifier.height(8.dp))
            Text(text = it, color = Color(0xFFFF5252), fontSize = 13.sp)
        }
        authState.infoMessage?.let {
            Spacer(modifier = Modifier.height(8.dp))
            Text(text = it, color = primary, fontSize = 13.sp)
        }

        Spacer(modifier = Modifier.height(72.dp))
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(48.dp)
                .padding(horizontal = 8.dp)
                .border(1.dp, Color(0xFFFF5252), RoundedCornerShape(4.dp))
                .clickable(
                    interactionSource = remember { MutableInteractionSource() },
                    indication = null,
                    onClick = {
                        scope.launch {
                            app.trackUploadFeature.setEnabled(false)
                            app.warehouseFeature.clear()
                            app.productionFeature.clear()
                            app.faultReportFeature.clear()
                            app.repairTaskFeature.clear()
                            app.inspectionTaskFeature.clear()
                            app.moveCarTaskFeature.clear()
                            app.freeMoveCarFeature.clear()
                            app.changeBatteryTaskFeature.clear()
                            app.scanFeature.clear()
                            app.vehicleFeature.clear()
                            app.serviceAreaFeature.clear()
                            app.authFeature.logout()
                        }
                    },
                ),
            contentAlignment = Alignment.Center,
        ) {
            Text(
                text = t(Str.Logout),
                color = Color(0xFFFF2222),
                fontWeight = FontWeight.Bold,
                fontSize = 15.sp,
            )
        }

        if (app.isDemoMode) {
            Spacer(modifier = Modifier.height(12.dp))
            Text(
                text = "shared ${OpsApp.LIBRARY_VERSION} · demo",
                color = Color(0xFF999999),
                fontSize = 11.sp,
                modifier = Modifier.fillMaxWidth(),
                textAlign = TextAlign.Center,
            )
        }
    }
}

@Composable
private fun SettingDivider() {
    HorizontalDivider(
        modifier = Modifier.padding(horizontal = 16.dp),
        thickness = 0.5.dp,
        color = Color(0xFFE5E5E5),
    )
}

@Composable
private fun SettingInfoRow(label: String, value: String) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .height(50.dp)
            .padding(horizontal = 16.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(label, color = Color(0xFF242936), fontSize = 15.sp)
        Spacer(modifier = Modifier.weight(1f))
        Text(value, color = Color(0xFF242936), fontSize = 15.sp)
    }
}

@Composable
private fun SettingSwitchRow(
    label: String,
    checked: Boolean,
    onCheckedChange: (Boolean) -> Unit,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .height(50.dp)
            .padding(horizontal = 16.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(label, color = Color(0xFF242936), fontSize = 15.sp)
        Spacer(modifier = Modifier.weight(1f))
        Switch(
            checked = checked,
            onCheckedChange = onCheckedChange,
            colors = SwitchDefaults.colors(
                checkedTrackColor = OpsTheme.colors.primary,
            ),
        )
    }
}

@Composable
private fun SettingActionRow(
    label: String,
    actionLabel: String,
    primary: Color,
    onAction: () -> Unit,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .height(50.dp)
            .padding(horizontal = 16.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(label, color = Color(0xFF242936), fontSize = 15.sp)
        Spacer(modifier = Modifier.weight(1f))
        Box(
            modifier = Modifier
                .height(32.dp)
                .background(primary.copy(alpha = 0.12f), RoundedCornerShape(16.dp))
                .clickable(
                    interactionSource = remember { MutableInteractionSource() },
                    indication = null,
                    onClick = onAction,
                )
                .padding(horizontal = 12.dp),
            contentAlignment = Alignment.Center,
        ) {
            Text(actionLabel, color = primary, fontSize = 14.sp)
        }
    }
}

@Composable
private fun SettingNavRow(
    label: String,
    value: String,
    onClick: () -> Unit,
    valueColor: Color = Color(0xFF242936),
    showArrow: Boolean = true,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .height(50.dp)
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            )
            .padding(horizontal = 16.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(label, color = Color(0xFF242936), fontSize = 15.sp)
        Spacer(modifier = Modifier.weight(1f))
        if (value.isNotBlank()) {
            Text(
                text = value,
                color = valueColor,
                fontSize = 15.sp,
                maxLines = 1,
                modifier = Modifier.padding(end = if (showArrow) 8.dp else 0.dp),
            )
        }
        if (showArrow) {
            Text("›", color = Color(0xFFCCCCCC), fontSize = 20.sp)
        }
    }
}
