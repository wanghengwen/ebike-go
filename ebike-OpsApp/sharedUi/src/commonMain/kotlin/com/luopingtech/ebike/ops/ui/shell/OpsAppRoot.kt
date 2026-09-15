package com.luopingtech.ebike.ops.ui.shell

import com.luopingtech.ebike.ops.OpsApp
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
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
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
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.ui.auth.BusinessPickerScaffold
import com.luopingtech.ebike.ops.ui.auth.GetSmsCodeAction
import com.luopingtech.ebike.ops.ui.auth.LoginModeTabs
import com.luopingtech.ebike.ops.ui.auth.LoginPrimaryButton
import com.luopingtech.ebike.ops.ui.auth.LoginWelcomeHeader
import com.luopingtech.ebike.ops.ui.auth.UnderlineTextField
import com.luopingtech.ebike.ops.ui.home.AreaGateScaffold
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import com.luopingtech.ebike.ops.platform.SecureStore
import androidx.compose.ui.graphics.Color
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.domain.model.BusinessTenant
import com.luopingtech.ebike.ops.feature.home.HomeUiState
import com.luopingtech.ebike.ops.feature.home.SignedInHomeGate
import com.luopingtech.ebike.ops.feature.home.resolveSignedInHomeGate
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

/**
 * 整个 APP 的界面入口：登录 → 设密码 → 选服务区 → 主壳，全部在共享层。
 *
 * 宿主（Activity / UIViewController）唯一要做的事是把平台能力用 CompositionLocal
 * 插进来——地图渲染、相机取景、轻提示、轨迹权限——然后调这一个函数。
 */
@Composable
fun OpsAppRoot(app: OpsApp) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val homeState by app.homeFeature.state.collectAsState()
    val authState by app.authFeature.state.collectAsState()
    var pickingArea by remember { mutableStateOf(false) }

    LaunchedEffect(authState.session?.accessToken) {
        if (authState.session != null) {
            app.homeFeature.ensureAreasLoaded()
        }
    }

    when {
        authState.session == null -> LoginScreen(app)
        authState.needSetPassword -> SetPasswordScreen(app)
        else -> when (
            resolveSignedInHomeGate(
                pickingArea = pickingArea,
                currentArea = homeState.currentArea,
                // Session / SecureStore / 仓库任一有记录，启动都直接进主页，避免闪选区页。
                persistedAreaId = authState.session?.serviceAreaId
                    ?.takeIf { it.isNotBlank() }
                    ?: app.serviceAreaRepository.currentArea()?.id,
                areasLoaded = homeState.areasLoaded,
            )
        ) {
            SignedInHomeGate.LoadingAreas -> LoadingScreen(t(Str.LoadingServiceAreas))
            SignedInHomeGate.AreaPicker -> AreaGateScreen(
                app = app,
                homeState = homeState,
                allowCancel = homeState.currentArea != null && pickingArea,
                onCancel = { pickingArea = false },
                onSelected = { pickingArea = false },
            )
            SignedInHomeGate.Home -> MainShell(
                app = app,
                homeState = homeState,
                onChangeArea = { pickingArea = true },
            )
        }
    }
}

@Composable
private fun LoadingScreen(message: String) {
    Box(
        modifier = Modifier.fillMaxSize(),
        contentAlignment = Alignment.Center,
    ) {
        Text(message)
    }
}

@Composable
private fun LoginScreen(app: OpsApp) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val authState by app.authFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    var mode by remember { mutableStateOf("password") } // password | sms | forget
    var account by remember { mutableStateOf("") }
    var password by remember { mutableStateOf("") }
    var smsCode by remember { mutableStateOf("") }
    var newPassword by remember { mutableStateOf("") }
    var passwordVisible by remember { mutableStateOf(false) }
    var newPasswordVisible by remember { mutableStateOf(false) }
    var smsCooldown by remember { mutableStateOf(0) }
    var pickingAreaCode by remember { mutableStateOf(false) }

    LaunchedEffect(authState.smsSent) {
        if (authState.smsSent) {
            smsCooldown = 60
        }
    }
    LaunchedEffect(smsCooldown) {
        if (smsCooldown > 0) {
            delay(1000)
            smsCooldown -= 1
        }
    }

    if (authState.pendingBusinesses.isNotEmpty()) {
        BusinessPickerScreen(
            app = app,
            loading = authState.loading,
            businesses = authState.pendingBusinesses,
            errorMessage = authState.errorMessage,
            onSelect = { tenantId ->
                scope.launch { app.authFeature.selectBusiness(tenantId) }
            },
            onBack = { app.authFeature.clearPendingBusiness() },
        )
        return
    }

    if (pickingAreaCode) {
        AreaCodePickerScreen(
            app = app,
            selected = authState.loginArea,
            onSelect = {
                app.authFeature.setLoginAreaCode(it)
                pickingAreaCode = false
            },
            onBack = { pickingAreaCode = false },
        )
        return
    }

    val appName = app.config.app.displayName.ifBlank { t(Str.AppName) }
    val canSendSms = !authState.loading && account.isNotBlank() && smsCooldown == 0
    val canPasswordLogin = !authState.loading && account.isNotBlank() && password.isNotBlank()
    val canSmsLogin = !authState.loading && account.isNotBlank() && smsCode.isNotBlank()
    val canReset = !authState.loading &&
        account.isNotBlank() &&
        smsCode.isNotBlank() &&
        newPassword.isNotBlank()

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.White)
            .statusBarsPadding()
            .padding(horizontal = 32.dp),
    ) {
        // Legacy Android: marginTop 99dp below status / content top.
        Spacer(modifier = Modifier.height(99.dp))

        if (mode == "forget") {
            Text(
                text = t(Str.ResetPasswordTitle),
                color = OpsTheme.colors.primary,
                fontSize = 32.sp,
            )
            Spacer(modifier = Modifier.height(28.dp))
        } else {
            LoginWelcomeHeader(
                welcome = t(Str.WelcomeToUse),
                appName = appName,
                onWelcomeLongClick = { pickingAreaCode = true },
            )
            Spacer(modifier = Modifier.height(20.dp))
            LoginModeTabs(
                passwordLabel = t(Str.PasswordLogin),
                smsLabel = t(Str.SmsLogin),
                selectedPassword = mode == "password",
                onSelectPassword = { mode = "password" },
                onSelectSms = { mode = "sms" },
            )
            Spacer(modifier = Modifier.height(28.dp))
        }

        UnderlineTextField(
            value = account,
            onValueChange = { account = it },
            hint = if (app.isDemoMode && mode == "password") {
                t(Str.Account)
            } else {
                t(Str.PhoneHint)
            },
            keyboardType = if (app.isDemoMode && mode == "password") {
                KeyboardType.Text
            } else {
                KeyboardType.Phone
            },
        )

        when (mode) {
            "password" -> {
                Spacer(modifier = Modifier.height(28.dp))
                UnderlineTextField(
                    value = password,
                    onValueChange = { password = it },
                    hint = t(Str.PasswordHint),
                    password = true,
                    passwordVisible = passwordVisible,
                    onTogglePasswordVisible = { passwordVisible = !passwordVisible },
                )
                Spacer(modifier = Modifier.height(44.dp))
                LoginPrimaryButton(
                    text = if (authState.loading) t(Str.LoggingIn) else t(Str.Login),
                    enabled = canPasswordLogin,
                    onClick = { scope.launch { app.authFeature.login(account, password) } },
                )
                Spacer(modifier = Modifier.height(24.dp))
                Text(
                    text = t(Str.ForgotPassword),
                    color = OpsTheme.colors.primary,
                    fontSize = 14.sp,
                    modifier = Modifier
                        .align(Alignment.CenterHorizontally)
                        .clickable(
                            interactionSource = remember { MutableInteractionSource() },
                            indication = null,
                            onClick = { mode = "forget" },
                        ),
                )
            }
            "sms" -> {
                Spacer(modifier = Modifier.height(28.dp))
                UnderlineTextField(
                    value = smsCode,
                    onValueChange = { smsCode = it },
                    hint = t(Str.SmsCodeHint),
                    keyboardType = KeyboardType.Number,
                    trailing = {
                        GetSmsCodeAction(
                            label = if (smsCooldown > 0) "${smsCooldown}s" else t(Str.GetSmsCode),
                            enabled = canSendSms,
                            onClick = { scope.launch { app.authFeature.sendSmsCode(account) } },
                        )
                    },
                )
                Spacer(modifier = Modifier.height(44.dp))
                LoginPrimaryButton(
                    text = if (authState.loading) t(Str.LoggingIn) else t(Str.Login),
                    enabled = canSmsLogin,
                    onClick = {
                        scope.launch { app.authFeature.loginWithSms(account, smsCode) }
                    },
                )
            }
            else -> {
                Spacer(modifier = Modifier.height(28.dp))
                UnderlineTextField(
                    value = smsCode,
                    onValueChange = { smsCode = it },
                    hint = t(Str.SmsCodeHint),
                    keyboardType = KeyboardType.Number,
                    trailing = {
                        GetSmsCodeAction(
                            label = if (smsCooldown > 0) "${smsCooldown}s" else t(Str.GetSmsCode),
                            enabled = canSendSms,
                            onClick = {
                                scope.launch {
                                    app.authFeature.sendSmsCode(
                                        account,
                                        com.luopingtech.ebike.ops.data.auth.SmsScene.FORGET_PASSWORD,
                                    )
                                }
                            },
                        )
                    },
                )
                Spacer(modifier = Modifier.height(28.dp))
                UnderlineTextField(
                    value = newPassword,
                    onValueChange = { newPassword = it },
                    hint = t(Str.NewPasswordHint),
                    password = true,
                    passwordVisible = newPasswordVisible,
                    onTogglePasswordVisible = { newPasswordVisible = !newPasswordVisible },
                )
                Spacer(modifier = Modifier.height(44.dp))
                LoginPrimaryButton(
                    text = if (authState.loading) t(Str.Submitting) else t(Str.SubmitReset),
                    enabled = canReset,
                    onClick = {
                        scope.launch {
                            app.authFeature.forgetPassword(account, smsCode, newPassword)
                        }
                    },
                )
                Spacer(modifier = Modifier.height(24.dp))
                Text(
                    text = t(Str.BackToLogin),
                    color = OpsTheme.colors.primary,
                    fontSize = 14.sp,
                    modifier = Modifier
                        .align(Alignment.CenterHorizontally)
                        .clickable(
                            interactionSource = remember { MutableInteractionSource() },
                            indication = null,
                            onClick = { mode = "password" },
                        ),
                )
            }
        }

        authState.errorMessage?.let {
            Spacer(modifier = Modifier.height(16.dp))
            Text(text = it, color = MaterialTheme.colorScheme.error, fontSize = 14.sp)
        }
        authState.infoMessage?.let {
            Spacer(modifier = Modifier.height(16.dp))
            Text(text = it, color = OpsTheme.colors.primary, fontSize = 14.sp)
        }

        Spacer(modifier = Modifier.weight(1f))
        if (app.isDemoMode) {
            Text(
                text = t(Str.DemoAccountsHint),
                color = OpsTheme.colors.textTertiary,
                fontSize = 11.sp,
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(bottom = 16.dp),
                textAlign = TextAlign.Center,
            )
        }
    }
}


@Composable
private fun AreaCodePickerScreen(
    app: OpsApp,
    selected: com.luopingtech.ebike.ops.data.auth.CallingCode,
    onSelect: (com.luopingtech.ebike.ops.data.auth.CallingCode) -> Unit,
    onBack: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    var query by remember { mutableStateOf("") }
    val filtered = remember(query) {
        com.luopingtech.ebike.ops.data.auth.CallingCodeCatalog.filter(query)
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(24.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(t(Str.AreaCodeTitle), style = MaterialTheme.typography.titleMedium)
            TextButton(onClick = onBack) { Text(t(Str.Back)) }
        }
        OutlinedTextField(
            value = query,
            onValueChange = { query = it },
            label = { Text(t(Str.AreaCodeSearch)) },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true,
        )
        Column(
            modifier = Modifier
                .weight(1f)
                .verticalScroll(rememberScrollState()),
            verticalArrangement = Arrangement.spacedBy(2.dp),
        ) {
            filtered.forEach { code ->
                val isSelected = code.regionCode == selected.regionCode &&
                    code.dialCode == selected.dialCode
                Text(
                    text = "${if (isSelected) "●" else "○"} ${code.displayLabel}",
                    modifier = Modifier
                        .fillMaxWidth()
                        .clickable { onSelect(code) }
                        .padding(vertical = 10.dp),
                    color = if (isSelected) {
                        MaterialTheme.colorScheme.primary
                    } else {
                        MaterialTheme.colorScheme.onSurface
                    },
                )
            }
        }
    }
}

@Composable
private fun SetPasswordScreen(app: OpsApp) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val authState by app.authFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    var password by remember { mutableStateOf("") }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(24.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Spacer(modifier = Modifier.height(48.dp))
        Text(t(Str.SetPasswordTitle), style = MaterialTheme.typography.headlineSmall)
        Text(
            text = t(Str.LoginPrompt),
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
        OutlinedTextField(
            value = password,
            onValueChange = { password = it },
            label = { Text(t(Str.NewPasswordHint)) },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true,
        )
        Button(
            onClick = { scope.launch { app.authFeature.setPassword(password) } },
            enabled = !authState.loading,
            modifier = Modifier.fillMaxWidth(),
        ) {
            Text(if (authState.loading) t(Str.Submitting) else t(Str.Confirm))
        }
        TextButton(
            onClick = { scope.launch { app.authFeature.cancelSetPassword() } },
            enabled = !authState.loading,
        ) {
            Text(t(Str.CancelAndLogout))
        }
        authState.errorMessage?.let {
            Text(text = it, color = MaterialTheme.colorScheme.error)
        }
    }
}


@Composable
private fun BusinessPickerScreen(
    app: OpsApp,
    loading: Boolean,
    businesses: List<BusinessTenant>,
    errorMessage: String?,
    onSelect: (String) -> Unit,
    onBack: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    var query by remember { mutableStateOf("") }
    BusinessPickerScaffold(
        title = t(Str.SelectBusiness),
        searchHint = t(Str.Search),
        query = query,
        onQueryChange = { query = it },
        businesses = businesses,
        loading = loading,
        errorMessage = errorMessage,
        backLabel = t(Str.BackToLogin),
        onBack = onBack,
        onSelect = onSelect,
    )
}

@Composable
internal fun AreaGateScreen(
    app: OpsApp,
    homeState: HomeUiState,
    allowCancel: Boolean,
    onCancel: () -> Unit,
    onSelected: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val scope = rememberCoroutineScope()
    AreaGateScaffold(
        title = t(Str.SwitchServiceArea),
        currentLocationTitle = t(Str.CurrentLocation),
        selectAreaTitle = t(Str.SelectServiceArea),
        currentAreaLabel = homeState.currentArea?.name.orEmpty().ifBlank { "-" },
        areas = homeState.serviceAreas,
        selectedAreaId = homeState.currentArea?.id,
        loading = homeState.loadingAreas,
        loadingText = t(Str.Loading),
        errorMessage = homeState.errorMessage,
        allowCancel = allowCancel,
        backLabel = t(Str.Back),
        onBack = onCancel,
        onSelect = { area ->
            scope.launch {
                app.homeFeature.selectArea(area)
                app.authFeature.refreshSessionFromStore()
                onSelected()
            }
        },
    )
}
