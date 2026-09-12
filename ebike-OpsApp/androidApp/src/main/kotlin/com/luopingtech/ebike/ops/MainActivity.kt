package com.luopingtech.ebike.ops

import android.Manifest
import android.content.pm.PackageManager
import android.net.Uri
import android.os.Build
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.compose.setContent
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.Image
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.combinedClickable
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
import androidx.compose.material3.FilterChip
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Surface
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
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.core.content.ContextCompat
import com.luopingtech.ebike.ops.ui.analysis.AnalysisCardItem
import com.luopingtech.ebike.ops.ui.analysis.AnalysisScaffold
import com.luopingtech.ebike.ops.ui.analysis.VehicleConditionDistributionMapScreen
import com.luopingtech.ebike.ops.ui.analysis.VehicleConditionDistributionScreen
import com.luopingtech.ebike.ops.ui.auth.BusinessPickerScaffold
import com.luopingtech.ebike.ops.ui.auth.GetSmsCodeAction
import com.luopingtech.ebike.ops.ui.auth.LoginModeTabs
import com.luopingtech.ebike.ops.ui.auth.LoginPrimaryButton
import com.luopingtech.ebike.ops.ui.auth.LoginWelcomeHeader
import com.luopingtech.ebike.ops.ui.auth.UnderlineTextField
import com.luopingtech.ebike.ops.ui.home.AreaGateScaffold
import com.luopingtech.ebike.ops.ui.home.HomeAreaTitleBar
import com.luopingtech.ebike.ops.ui.home.HomeFilterHandle
import com.luopingtech.ebike.ops.ui.home.HomeMapToolsRail
import com.luopingtech.ebike.ops.ui.home.HomeStatItem
import com.luopingtech.ebike.ops.ui.home.HomeStatisticsPanel
import com.luopingtech.ebike.ops.ui.home.homeStatColor
import com.luopingtech.ebike.ops.ui.icons.OpsIcon
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import com.luopingtech.ebike.ops.ui.workbench.WORKBENCH_COMMON_MAX
import com.luopingtech.ebike.ops.ui.workbench.WorkbenchEditBadge
import com.luopingtech.ebike.ops.ui.workbench.WorkbenchModuleItem
import com.luopingtech.ebike.ops.ui.workbench.WorkbenchModuleSection
import com.luopingtech.ebike.ops.ui.workbench.WorkbenchScaffold
import com.luopingtech.ebike.ops.platform.SecureStore
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.material3.NavigationBarItemDefaults
import androidx.compose.ui.zIndex
import com.luopingtech.ebike.ops.core.config.H5ScreenKind
import com.luopingtech.ebike.ops.core.config.H5ScreenUrls
import com.luopingtech.ebike.ops.core.i18n.OpsLanguage
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.control.ControlChannel
import com.luopingtech.ebike.ops.domain.control.VehicleAction
import com.luopingtech.ebike.ops.domain.model.BusinessTenant
import com.luopingtech.ebike.ops.domain.model.MapPin
import com.luopingtech.ebike.ops.domain.model.OpsTask
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.model.Vehicle
import com.luopingtech.ebike.ops.domain.permission.OpsPermissionCodes
import com.luopingtech.ebike.ops.domain.permission.OpsPermissions
import com.luopingtech.ebike.ops.domain.vehicle.VehicleAlarmFilter
import com.luopingtech.ebike.ops.domain.vehicle.VehicleAlarmFilterLogic
import com.luopingtech.ebike.ops.domain.vehicle.VehicleMapFilter
import com.luopingtech.ebike.ops.domain.vehicle.VehicleMapFilterLogic
import com.luopingtech.ebike.ops.feature.home.HomeUiState
import com.luopingtech.ebike.ops.feature.task.ClaimableTaskFeature
import com.luopingtech.ebike.ops.platform.ActivityCodeScanner
import com.luopingtech.ebike.ops.platform.ActivityPhotoCapture
import com.luopingtech.ebike.ops.platform.AndroidLocationTracker
import com.luopingtech.ebike.ops.scan.OpsCameraScanPreview
import com.luopingtech.ebike.ops.ui.h5.H5Screen
import com.luopingtech.ebike.ops.ui.map.SimulatorMapView
import com.luopingtech.ebike.ops.ui.map.TencentMapView
import com.luopingtech.ebike.ops.ui.production.ProductionScreen
import com.luopingtech.ebike.ops.ui.report.FaultReportScreen
import com.luopingtech.ebike.ops.ui.task.BatchMoveCarSection
import com.luopingtech.ebike.ops.ui.task.FreeMoveCarSection
import com.luopingtech.ebike.ops.ui.task.MoveCarTaskSection
import com.luopingtech.ebike.ops.ui.task.TaskAuditResultSection
import com.luopingtech.ebike.ops.ui.task.TaskCenterCardItem
import com.luopingtech.ebike.ops.ui.task.TaskMapKind
import com.luopingtech.ebike.ops.ui.task.TaskMapScreen
import com.luopingtech.ebike.ops.ui.task.TaskScaffold
import com.luopingtech.ebike.ops.ui.relocation.RelocationScreen
import com.luopingtech.ebike.ops.ui.tools.UnlockedVehiclesScreen
import com.luopingtech.ebike.ops.ui.sneak.SneakReportScreen
import com.luopingtech.ebike.ops.ui.workorder.WorkOrderScreen
import com.luopingtech.ebike.ops.domain.model.WorkOrderKind
import com.luopingtech.ebike.ops.ui.vehicle.VehicleDetailSection
import com.luopingtech.ebike.ops.ui.warehouse.WarehouseScreen
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

private enum class MainTab { Map, Tasks, Analysis, Workbench }

private enum class ScanMode { Detail, Unlock, Lock }

private enum class TaskKind {
    Hub,
    ChangeBattery,
    MoveCar,
    FreeMoveCar,
    BatchMoveCar,
    Inspection,
    Repair,
}

class MainActivity : ComponentActivity() {
    private lateinit var activityCodeScanner: ActivityCodeScanner
    private lateinit var activityPhotoCapture: ActivityPhotoCapture

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        val host = application as OpsApplication
        val app = host.opsApp
        activityCodeScanner = ActivityCodeScanner(this)
        host.codeScannerBridge.bind(activityCodeScanner)
        activityPhotoCapture = ActivityPhotoCapture(this)
        host.photoCaptureBridge.bind(activityPhotoCapture)
        setContent {
            OpsTheme(branding = app.config.branding) {
                Surface(modifier = Modifier.fillMaxSize()) {
                    RootNav(app)
                }
            }
        }
    }

    override fun onDestroy() {
        val host = application as OpsApplication
        host.codeScannerBridge.unbind(activityCodeScanner)
        host.photoCaptureBridge.unbind(activityPhotoCapture)
        super.onDestroy()
    }
}

@Composable
private fun RootNav(app: OpsApp) {
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
        homeState.loadingAreas && homeState.serviceAreas.isEmpty() -> LoadingScreen(t(Str.LoadingServiceAreas))
        homeState.currentArea == null || pickingArea -> {
            AreaGateScreen(
                app = app,
                homeState = homeState,
                allowCancel = homeState.currentArea != null && pickingArea,
                onCancel = { pickingArea = false },
                onSelected = { pickingArea = false },
            )
        }
        else -> MainShell(
            app = app,
            homeState = homeState,
            onChangeArea = { pickingArea = true },
        )
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
private fun AreaGateScreen(
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

@Composable
private fun MainShell(
    app: OpsApp,
    homeState: HomeUiState,
    onChangeArea: () -> Unit,
) {
    val permissions = remember(homeState.session?.permissionCodes, app.isDemoMode) {
        when {
            app.isDemoMode && homeState.session?.permissionCodes.isNullOrEmpty() ->
                OpsPermissions.demoFull()
            else -> OpsPermissions.fromCodes(homeState.session?.permissionCodes.orEmpty())
        }
    }
    var tab by remember { mutableStateOf(MainTab.Map) }
    var taskKind by remember { mutableStateOf(TaskKind.Hub) }
    var warehouseOpen by remember { mutableStateOf(false) }
    var vehicleConditionDistOpen by remember { mutableStateOf(false) }
    var vehicleConditionDistMapOpen by remember { mutableStateOf(false) }
    var productionOpen by remember { mutableStateOf(false) }
    var faultReportOpen by remember { mutableStateOf(false) }
    var sneakReportOpen by remember { mutableStateOf(false) }
    var unlockedVehiclesOpen by remember { mutableStateOf(false) }
    var relocationOpen by remember { mutableStateOf(false) }
    var inspectionOrderOpen by remember { mutableStateOf(false) }
    var repairOrderOpen by remember { mutableStateOf(false) }
    var h5Screen by remember { mutableStateOf<H5ScreenKind?>(null) }
    var taskMapOpen by remember { mutableStateOf(false) }
    var scanOpen by remember { mutableStateOf(false) }
    var trackPermissionHint by remember { mutableStateOf<String?>(null) }
    val context = LocalContext.current

    val backgroundLocationLauncher = rememberLauncherForActivityResult(
        ActivityResultContracts.RequestPermission(),
    ) { /* optional; FGS still works with foreground-only */ }
    val locationPermissionLauncher = rememberLauncherForActivityResult(
        ActivityResultContracts.RequestMultiplePermissions(),
    ) { result ->
        val androidTracker = app.locationTracker as? AndroidLocationTracker
        val locationOk = app.isDemoMode ||
            androidTracker == null ||
            androidTracker.hasPermission() ||
            result[Manifest.permission.ACCESS_FINE_LOCATION] == true ||
            result[Manifest.permission.ACCESS_COARSE_LOCATION] == true
        if (locationOk) {
            trackPermissionHint = null
            app.trackUploadFeature.setEnabled(true)
            // Android 10+: always-allow (background) for kill/background continuity.
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q &&
                !app.isDemoMode &&
                ContextCompat.checkSelfPermission(
                    context,
                    Manifest.permission.ACCESS_BACKGROUND_LOCATION,
                ) != PackageManager.PERMISSION_GRANTED
            ) {
                backgroundLocationLauncher.launch(Manifest.permission.ACCESS_BACKGROUND_LOCATION)
            }
        } else {
            trackPermissionHint = app.i18n.t(Str.TrackNeedLocation)
        }
    }

    // Legacy: Splash auth → location → auto track while Main is alive.
    // Persist + START_STICKY + boot receiver cover process death / reboot resume.
    LaunchedEffect(homeState.session?.userId) {
        if (homeState.session == null) return@LaunchedEffect
        if (app.trackUploadFeature.state.value.enabled) return@LaunchedEffect
        val shouldResume = app.trackUploadFeature.wasEnabledPersisted() || true
        if (!shouldResume) return@LaunchedEffect
        val androidTracker = app.locationTracker as? AndroidLocationTracker
        val needLocation = androidTracker != null && !androidTracker.hasPermission() && !app.isDemoMode
        val needNotify = Build.VERSION.SDK_INT >= 33 &&
            ContextCompat.checkSelfPermission(
                context,
                Manifest.permission.POST_NOTIFICATIONS,
            ) != PackageManager.PERMISSION_GRANTED
        when {
            app.isDemoMode || androidTracker == null || (!needLocation && !needNotify) -> {
                app.trackUploadFeature.setEnabled(true)
            }
            else -> {
                val perms = buildList {
                    if (needLocation) {
                        add(Manifest.permission.ACCESS_FINE_LOCATION)
                        add(Manifest.permission.ACCESS_COARSE_LOCATION)
                    }
                    if (needNotify) add(Manifest.permission.POST_NOTIFICATIONS)
                }.toTypedArray()
                locationPermissionLauncher.launch(perms)
            }
        }
    }

    if (vehicleConditionDistMapOpen) {
        VehicleConditionDistributionMapScreen(
            app = app,
            onClose = { vehicleConditionDistMapOpen = false },
        )
        return
    }
    if (vehicleConditionDistOpen) {
        VehicleConditionDistributionScreen(
            app = app,
            onClose = { vehicleConditionDistOpen = false },
            onOpenMap = { vehicleConditionDistMapOpen = true },
        )
        return
    }
    if (warehouseOpen) {
        WarehouseScreen(
            app = app,
            permissions = permissions,
            onClose = { warehouseOpen = false },
        )
        return
    }
    if (productionOpen) {
        ProductionScreen(
            app = app,
            permissions = permissions,
            onClose = { productionOpen = false },
            onLocateOnMap = {
                productionOpen = false
                tab = MainTab.Map
            },
        )
        return
    }
    if (faultReportOpen) {
        FaultReportScreen(
            app = app,
            onClose = { faultReportOpen = false },
        )
        return
    }
    if (sneakReportOpen) {
        SneakReportScreen(
            app = app,
            onClose = { sneakReportOpen = false },
        )
        return
    }
    if (unlockedVehiclesOpen) {
        UnlockedVehiclesScreen(
            app = app,
            currentArea = homeState.currentArea,
            onClose = { unlockedVehiclesOpen = false },
        )
        return
    }
    if (relocationOpen) {
        RelocationScreen(
            app = app,
            currentArea = homeState.currentArea,
            onClose = { relocationOpen = false },
        )
        return
    }
    if (inspectionOrderOpen) {
        WorkOrderScreen(
            app = app,
            kind = WorkOrderKind.Inspection,
            currentArea = homeState.currentArea,
            onClose = { inspectionOrderOpen = false },
        )
        return
    }
    if (repairOrderOpen) {
        WorkOrderScreen(
            app = app,
            kind = WorkOrderKind.Repair,
            currentArea = homeState.currentArea,
            onClose = { repairOrderOpen = false },
        )
        return
    }
    if (h5Screen != null) {
        H5Screen(
            app = app,
            kind = h5Screen!!,
            onClose = { h5Screen = null },
        )
        return
    }
    if (taskMapOpen) {
        com.luopingtech.ebike.ops.ui.task.TaskMapScreen(
            app = app,
            permissions = permissions,
            initialKind = when {
                permissions.showChangeBattery -> com.luopingtech.ebike.ops.ui.task.TaskMapKind.ChangeBattery
                else -> com.luopingtech.ebike.ops.ui.task.TaskMapKind.MoveCar
            },
            onClose = { taskMapOpen = false },
        )
        return
    }
    if (scanOpen) {
        ScanOverlay(
            app = app,
            permissions = permissions,
            onClose = {
                app.scanFeature.clear()
                scanOpen = false
            },
        )
        return
    }

    LaunchedEffect(permissions.showMap, permissions.showTaskCenter, tab) {
        val allowed = buildList {
            if (permissions.showMap) add(MainTab.Map)
            if (permissions.showTaskCenter) add(MainTab.Tasks)
            add(MainTab.Analysis)
            add(MainTab.Workbench)
        }
        if (tab !in allowed) {
            tab = allowed.first()
        }
    }

    Scaffold(
        containerColor = Color.White,
        bottomBar = {
            NavigationBar(
                containerColor = Color.White,
                tonalElevation = 0.dp,
            ) {
                val itemColors = NavigationBarItemDefaults.colors(
                    selectedIconColor = OpsTheme.colors.primary,
                    selectedTextColor = OpsTheme.colors.primary,
                    indicatorColor = Color.Transparent,
                    unselectedIconColor = Color(0xFF666666),
                    unselectedTextColor = Color(0xFF666666),
                )
                if (permissions.showMap) {
                    NavigationBarItem(
                        selected = tab == MainTab.Map,
                        onClick = { tab = MainTab.Map },
                        colors = itemColors,
                        icon = { Text("◉", fontSize = 16.sp) },
                        label = { Text(app.i18n.t(Str.TabMap), fontSize = 11.sp) },
                    )
                }
                if (permissions.showTaskCenter) {
                    NavigationBarItem(
                        selected = tab == MainTab.Tasks,
                        onClick = {
                            tab = MainTab.Tasks
                            taskKind = TaskKind.Hub
                        },
                        colors = itemColors,
                        icon = { Text("☑", fontSize = 16.sp) },
                        label = { Text(app.i18n.t(Str.TabTasks), fontSize = 11.sp) },
                    )
                }
                if (permissions.showScan) {
                    NavigationBarItem(
                        selected = false,
                        onClick = { scanOpen = true },
                        colors = itemColors,
                        icon = {
                            Box(
                                modifier = Modifier
                                    .size(42.dp)
                                    .background(OpsTheme.colors.primary, CircleShape),
                                contentAlignment = Alignment.Center,
                            ) {
                                Text("⬚", color = Color.White, fontSize = 18.sp)
                            }
                        },
                        label = { Text("") },
                    )
                }
                NavigationBarItem(
                    selected = tab == MainTab.Analysis,
                    onClick = { tab = MainTab.Analysis },
                    colors = itemColors,
                    icon = { Text("📈", fontSize = 14.sp) },
                    label = { Text(app.i18n.t(Str.TabAnalysis), fontSize = 11.sp) },
                )
                NavigationBarItem(
                    selected = tab == MainTab.Workbench,
                    onClick = { tab = MainTab.Workbench },
                    colors = itemColors,
                    icon = { Text("▦", fontSize = 16.sp) },
                    label = { Text(app.i18n.t(Str.TabWorkbench), fontSize = 11.sp) },
                )
            }
        },
    ) { padding ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding),
        ) {
            when (tab) {
                MainTab.Map -> MapTab(
                    app = app,
                    homeState = homeState,
                    permissions = permissions,
                    onChangeArea = onChangeArea,
                )
                MainTab.Tasks -> TasksTab(
                    app = app,
                    homeState = homeState,
                    permissions = permissions,
                    taskKind = taskKind,
                    onTaskKind = { taskKind = it },
                    onChangeArea = onChangeArea,
                    onOpenTaskMap = { taskMapOpen = true },
                )
                MainTab.Analysis -> AnalysisTab(
                    app = app,
                    homeState = homeState,
                    permissions = permissions,
                    onChangeArea = onChangeArea,
                    onOpenVehicleDist = {
                        app.vehicleConditionDistributionFeature.open()
                        vehicleConditionDistOpen = true
                    },
                )
                MainTab.Workbench -> WorkbenchTab(
                    app = app,
                    homeState = homeState,
                    permissions = permissions,
                    trackPermissionHint = trackPermissionHint,
                    onOpenWarehouse = { warehouseOpen = true },
                    onOpenProduction = { productionOpen = true },
                    onOpenFaultReport = { faultReportOpen = true },
                    onOpenSneakReport = { sneakReportOpen = true },
                    onOpenUnlockedVehicles = { unlockedVehiclesOpen = true },
                    onOpenRelocation = { relocationOpen = true },
                    onOpenInspectionOrder = { inspectionOrderOpen = true },
                    onOpenRepairOrder = { repairOrderOpen = true },
                    onOpenH5 = { h5Screen = it },
                    onChangeArea = onChangeArea,
                )
            }
        }
    }
}

@Composable
private fun MapTab(
    app: OpsApp,
    homeState: HomeUiState,
    permissions: OpsPermissions,
    onChangeArea: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val scope = rememberCoroutineScope()
    var filter by remember { mutableStateOf(VehicleMapFilter.All) }
    var selectedAlarms by remember { mutableStateOf(setOf<Int>()) }
    var alarmPanelOpen by remember { mutableStateOf(false) }
    var cardMessage by remember { mutableStateOf<String?>(null) }
    var detailOpen by remember { mutableStateOf(false) }
    var showFence by remember { mutableStateOf(false) }
    var mapTypeSatellite by remember { mutableStateOf(false) }
    var clusterOverview by remember { mutableStateOf(true) }
    var clusterExpandIds by remember { mutableStateOf<List<String>?>(null) }
    val selected = homeState.vehicles.firstOrNull { it.carId == homeState.selectedCarId }
    val detailMap by app.vehicleDetailMapFeature.state.collectAsState()
    val vehicles = homeState.vehicles
    val counts = remember(vehicles) {
        VehicleMapFilterLogic.counts(vehicles)
    }
    val alarmCounts = remember(vehicles) {
        VehicleAlarmFilterLogic.counts(vehicles)
    }
    val filtered = remember(vehicles, filter, selectedAlarms) {
        val alarmActive = selectedAlarms.isNotEmpty()
        vehicles.filter {
            (!alarmActive || !VehicleAlarmFilterLogic.isSoldOut(it.operationStates)) &&
                VehicleMapFilterLogic.matches(
                    ridingState = it.ridingState,
                    operationStates = it.operationStates,
                    restBattery = it.restBattery,
                    filter = filter,
                ) && VehicleAlarmFilterLogic.matchesAlarms(
                alarmStates = it.alarmStates,
                isOnline = it.isOnline,
                selected = selectedAlarms,
            )
        }
    }
    val stateFilters = listOf(
        VehicleMapFilter.Warehouse,
        VehicleMapFilter.Ready,
        VehicleMapFilter.Booking,
        VehicleMapFilter.Riding,
        VehicleMapFilter.TempParking,
        VehicleMapFilter.LowBattery,
        VehicleMapFilter.Repairing,
        VehicleMapFilter.Moving,
    )
    val statItems = stateFilters.map { item ->
        HomeStatItem(
            filter = item,
            label = when (item) {
                VehicleMapFilter.Warehouse -> t(Str.FilterWarehouse)
                VehicleMapFilter.Ready -> t(Str.FilterReady)
                VehicleMapFilter.Booking -> t(Str.FilterBooking)
                VehicleMapFilter.Riding -> t(Str.FilterRiding)
                VehicleMapFilter.TempParking -> t(Str.FilterTempParking)
                VehicleMapFilter.LowBattery -> t(Str.FilterLowBattery)
                VehicleMapFilter.Repairing -> t(Str.FilterRepairing)
                VehicleMapFilter.Moving -> t(Str.FilterMoving)
                VehicleMapFilter.All -> t(Str.FilterAll)
            },
            count = counts[item] ?: 0,
            valueColor = homeStatColor(item),
        )
    }

    Box(modifier = Modifier.fillMaxSize()) {
        MapSurface(
            app = app,
            pins = filtered.map {
                MapPin(
                    id = it.carId,
                    lat = it.lat,
                    lng = it.lng,
                    title = it.carId,
                    subtitle = it.batteryLabel,
                    restBattery = it.restBattery,
                    ridingState = it.ridingState,
                    memberCount = 1,
                    memberIds = listOf(it.carId),
                )
            },
            selectedCarId = homeState.selectedCarId,
            mapReady = homeState.mapReady,
            mapProviderKind = homeState.mapProviderKind,
            clusterOverview = clusterOverview,
            fencePolygons = if (showFence || (detailOpen && detailMap.showFence)) {
                detailMap.fence?.all.orEmpty()
            } else {
                emptyList()
            },
            trackPoints = if (detailOpen && detailMap.showTrack) detailMap.track else emptyList(),
            onPinClick = {
                detailOpen = false
                clusterExpandIds = null
                cardMessage = null
                app.homeFeature.selectVehicle(it)
            },
            onClusterClick = { ids ->
                detailOpen = false
                if (ids.size <= 1) {
                    clusterExpandIds = null
                    app.homeFeature.selectVehicle(ids.firstOrNull())
                } else {
                    clusterOverview = false
                    clusterExpandIds = ids
                }
            },
            modifier = Modifier.fillMaxSize(),
        )

        HomeAreaTitleBar(
            areaName = homeState.currentArea?.name ?: t(Str.NoServiceArea),
            onClick = onChangeArea,
            modifier = Modifier
                .align(Alignment.TopCenter)
                .statusBarsPadding()
                .zIndex(2f),
        )

        HomeMapToolsRail(
            refreshLabel = t(Str.Refresh),
            detailLabel = t(Str.Detail),
            fenceLabel = t(Str.MapToolFence),
            switchLabel = t(Str.MapToolSwitch),
            detailSelected = detailOpen,
            fenceSelected = showFence,
            switchSelected = mapTypeSatellite,
            onRefresh = { scope.launch { app.homeFeature.reloadVehicles() } },
            onDetail = {
                val opening = !detailOpen
                detailOpen = opening
                if (opening && selected != null) {
                    scope.launch { app.homeFeature.refreshSelectedDetail() }
                }
            },
            onFence = {
                showFence = !showFence
                if (showFence && selected != null) {
                    scope.launch { app.homeFeature.refreshSelectedDetail() }
                }
            },
            onSwitch = { mapTypeSatellite = !mapTypeSatellite },
            modifier = Modifier
                .align(Alignment.CenterStart)
                .padding(start = 12.dp)
                .zIndex(2f),
        )

        HomeFilterHandle(
            label = t(Str.FilterHandle),
            open = alarmPanelOpen,
            onClick = { alarmPanelOpen = !alarmPanelOpen },
            modifier = Modifier
                .align(Alignment.CenterEnd)
                .zIndex(2f),
        )

        if (alarmPanelOpen) {
            Surface(
                modifier = Modifier
                    .align(Alignment.CenterEnd)
                    .fillMaxWidth(0.82f)
                    .fillMaxSize()
                    .zIndex(3f),
                color = Color.White,
                shadowElevation = 8.dp,
            ) {
                Column(modifier = Modifier.padding(12.dp)) {
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Text(t(Str.Alarms), fontWeight = FontWeight.Medium)
                        TextButton(onClick = { alarmPanelOpen = false }) { Text(t(Str.Close)) }
                    }
                    AlarmFilterPanel(
                        app = app,
                        selected = selectedAlarms,
                        counts = alarmCounts,
                        onToggle = { code ->
                            selectedAlarms = if (code in selectedAlarms) {
                                selectedAlarms - code
                            } else {
                                selectedAlarms + code
                            }
                        },
                        onClear = { selectedAlarms = emptySet() },
                    )
                }
            }
        }

        clusterExpandIds?.takeIf { it.size > 1 }?.let { ids ->
            Surface(
                tonalElevation = 2.dp,
                modifier = Modifier
                    .align(Alignment.Center)
                    .padding(horizontal = 24.dp)
                    .zIndex(2f),
            ) {
                Column(
                    modifier = Modifier.padding(12.dp),
                    verticalArrangement = Arrangement.spacedBy(4.dp),
                ) {
                    Text(
                        t(Str.ClusterVehicles, ids.size),
                        style = MaterialTheme.typography.titleSmall,
                    )
                    ids.take(10).forEach { id ->
                        TextButton(
                            onClick = {
                                clusterExpandIds = null
                                app.homeFeature.selectVehicle(id)
                            },
                        ) { Text(id) }
                    }
                    TextButton(onClick = { clusterExpandIds = null }) { Text(t(Str.Close)) }
                }
            }
        }

        Column(
            modifier = Modifier
                .align(Alignment.BottomCenter)
                .fillMaxWidth()
                .zIndex(2f),
        ) {
            if (selected != null || cardMessage != null) {
                SelectedVehicleCard(
                    app = app,
                    vehicle = selected,
                    message = cardMessage,
                    detailOpen = detailOpen,
                    serviceAreaId = homeState.currentArea?.id,
                    showChangeBattery = permissions.showChangeBattery,
                    showDetails = permissions.canScanDetails || permissions.showMap,
                    canBindBattery = permissions.canBindBatterySn,
                    onRing = {
                        scope.launch {
                            val carId = selected?.carId ?: return@launch
                            cardMessage = when (
                                val r = app.vehicleControl.execute(
                                    carId,
                                    VehicleAction.Ring,
                                    ControlChannel.BlePreferred,
                                )
                            ) {
                                is OpsResult.Ok -> t(Str.RingOk, carId)
                                is OpsResult.Err -> t(Str.RingFailed, r.error.message)
                            }
                        }
                    },
                    onOpenBattery = {
                        scope.launch {
                            val carId = selected?.carId ?: return@launch
                            cardMessage = when (
                                val r = app.vehicleControl.execute(
                                    carId,
                                    VehicleAction.OpenBatteryBox,
                                    ControlChannel.BlePreferred,
                                )
                            ) {
                                is OpsResult.Ok -> t(Str.OpenBoxOkShort, carId)
                                is OpsResult.Err -> t(Str.OpenBoxFailedShort, r.error.message)
                            }
                        }
                    },
                    onFinishSwap = {
                        scope.launch {
                            val carId = selected?.carId ?: return@launch
                            cardMessage = when (
                                val r = app.vehicleControl.execute(
                                    carId,
                                    VehicleAction.CloseBatteryBox,
                                    ControlChannel.BlePreferred,
                                )
                            ) {
                                is OpsResult.Ok -> {
                                    app.homeFeature.reloadVehicles()
                                    t(Str.FinishSwapOk, carId)
                                }
                                is OpsResult.Err -> t(Str.FinishSwapFailed, r.error.message)
                            }
                        }
                    },
                    onDetails = {
                        val opening = !detailOpen
                        detailOpen = opening
                        if (opening) {
                            scope.launch { app.homeFeature.refreshSelectedDetail() }
                        }
                    },
                )
            }
            homeState.errorMessage?.let {
                Text(
                    text = it,
                    color = MaterialTheme.colorScheme.error,
                    modifier = Modifier
                        .fillMaxWidth()
                        .background(Color.White.copy(alpha = 0.9f))
                        .padding(horizontal = 16.dp, vertical = 4.dp),
                )
            }
            HomeStatisticsPanel(
                items = statItems,
                selected = filter,
                onSelect = { filter = it },
            )
        }
    }
}

@Composable
private fun AnalysisTab(
    app: OpsApp,
    homeState: HomeUiState,
    permissions: OpsPermissions,
    onChangeArea: () -> Unit,
    onOpenVehicleDist: () -> Unit,
) {
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val context = LocalContext.current
    fun comingSoon() {
        android.widget.Toast.makeText(context, t(Str.FeatureComingSoon), android.widget.Toast.LENGTH_SHORT).show()
    }
    val cards = buildList {
        fun addCard(visible: Boolean, id: String, title: String, icon: OpsIcon, onClick: () -> Unit) {
            if (visible || app.isDemoMode) {
                add(
                    AnalysisCardItem(
                        id = id,
                        title = title,
                        icon = icon,
                        onClick = onClick,
                    ),
                )
            }
        }
        addCard(permissions.showAnalysisOfflineOps, "offline", t(Str.OfflineOperation), OpsIcon.OfflineOperation, ::comingSoon)
        addCard(permissions.showAnalysisStation, "station", t(Str.StationMonitor), OpsIcon.AnalysisStation, ::comingSoon)
        addCard(permissions.showAnalysisReturnCar, "return", t(Str.ReturnCarAnalysis), OpsIcon.AnalysisReturnBike, ::comingSoon)
        addCard(permissions.showAnalysisVehicleDist, "vdist", t(Str.VehicleDistribution), OpsIcon.AnalysisVehicleDistribution, onOpenVehicleDist)
    }.distinctBy { it.id }

    AnalysisScaffold(
        areaName = homeState.currentArea?.name?.takeIf { it.isNotBlank() }
            ?: t(Str.SelectServiceArea),
        pageTitle = t(Str.TabAnalysis),
        cards = cards,
        onChangeArea = onChangeArea,
    )
}

@Composable
private fun SelectedVehicleCard(
    app: OpsApp,
    vehicle: Vehicle?,
    message: String?,
    detailOpen: Boolean,
    serviceAreaId: String? = null,
    showChangeBattery: Boolean,
    showDetails: Boolean,
    canBindBattery: Boolean = false,
    onRing: () -> Unit,
    onOpenBattery: () -> Unit,
    onFinishSwap: () -> Unit,
    onDetails: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    Surface(
        tonalElevation = 2.dp,
        modifier = Modifier.fillMaxWidth(),
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            if (vehicle == null) {
                Text(
                    text = t(Str.MapPickVehicle),
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            } else {
                Text(
                    text = "${vehicle.carId} · ${vehicle.batteryLabel} · ${vehicle.ridingLabel}" +
                        if (vehicle.isOnline) " · ${t(Str.Online)}" else " · ${t(Str.Offline)}",
                    style = MaterialTheme.typography.titleSmall,
                )
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    Button(onClick = onRing, modifier = Modifier.weight(1f)) { Text(t(Str.Ring)) }
                    if (showChangeBattery) {
                        Button(onClick = onOpenBattery, modifier = Modifier.weight(1f)) {
                            Text(t(Str.OpenBatteryBox))
                        }
                        Button(onClick = onFinishSwap, modifier = Modifier.weight(1f)) {
                            Text(t(Str.FinishChangeBattery))
                        }
                    }
                    if (showDetails) {
                        TextButton(onClick = onDetails) {
                            Text(if (detailOpen) t(Str.Collapse) else t(Str.Detail))
                        }
                    }
                }
                if (detailOpen) {
                    VehicleDetailSection(
                        app = app,
                        vehicle = vehicle,
                        serviceAreaId = serviceAreaId,
                        canBindBattery = canBindBattery,
                    )
                }
            }
            message?.let { Text(text = it, style = MaterialTheme.typography.bodySmall) }
        }
    }
}

@Composable
private fun AlarmFilterPanel(
    app: OpsApp,
    selected: Set<Int>,
    counts: Map<VehicleAlarmFilter, Int>,
    onToggle: (Int) -> Unit,
    onClear: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    Surface(
        tonalElevation = 2.dp,
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 12.dp),
    ) {
        Column(
            modifier = Modifier.padding(12.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Text(t(Str.Alarms), style = MaterialTheme.typography.titleSmall)
                TextButton(onClick = onClear, enabled = selected.isNotEmpty()) {
                    Text(t(Str.Clear))
                }
            }
            Text(
                text = t(Str.AlarmFilterHint),
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
            VehicleAlarmFilter.entries.chunked(3).forEach { row ->
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(6.dp),
                ) {
                    row.forEach { item ->
                        FilterChip(
                            selected = item.code in selected,
                            onClick = { onToggle(item.code) },
                            label = {
                                Text("${item.label} ${counts[item] ?: 0}")
                            },
                            modifier = Modifier.weight(1f),
                        )
                    }
                    repeat(3 - row.size) {
                        Spacer(modifier = Modifier.weight(1f))
                    }
                }
            }
        }
    }
}

@Composable
private fun TasksTab(
    app: OpsApp,
    homeState: HomeUiState,
    permissions: OpsPermissions,
    taskKind: TaskKind,
    onTaskKind: (TaskKind) -> Unit,
    onChangeArea: () -> Unit,
    onOpenTaskMap: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val battery by app.changeBatteryTaskFeature.state.collectAsState()
    val move by app.moveCarTaskFeature.state.collectAsState()
    val inspection by app.inspectionTaskFeature.state.collectAsState()
    val repair by app.repairTaskFeature.state.collectAsState()
    var batchParentTaskId by remember { mutableStateOf<String?>(null) }

    LaunchedEffect(homeState.currentArea?.id) {
        val area = homeState.currentArea ?: return@LaunchedEffect
        if (permissions.showChangeBattery) app.changeBatteryTaskFeature.load(area)
        if (permissions.showMoveCar) app.moveCarTaskFeature.load(area)
        if (permissions.showInspection) app.inspectionTaskFeature.load(area)
        if (permissions.showRepair) app.repairTaskFeature.load(area)
    }

    when (taskKind) {
        TaskKind.Hub -> {
            val cards = buildList {
                if (permissions.showChangeBattery || app.isDemoMode) {
                    add(
                        TaskCenterCardItem(
                            id = "battery",
                            title = t(Str.ChangeBatteryShort),
                            icon = OpsIcon.TaskChangeBattery,
                            badgeText = badgeTotalCount(battery.tasks.size),
                            onClick = { onTaskKind(TaskKind.ChangeBattery) },
                        ),
                    )
                }
                if (permissions.showMoveCar || app.isDemoMode) {
                    add(
                        TaskCenterCardItem(
                            id = "move",
                            title = t(Str.MoveCarShort),
                            icon = OpsIcon.TaskMoveBike,
                            badgeText = badgeTotalCount(move.tasks.size),
                            onClick = { onTaskKind(TaskKind.MoveCar) },
                        ),
                    )
                }
                if (permissions.showInspection || app.isDemoMode) {
                    add(
                        TaskCenterCardItem(
                            id = "inspection",
                            title = t(Str.InspectionShort),
                            icon = OpsIcon.TaskInspection,
                            badgeText = badgeClaimed(inspection.tasks),
                            onClick = { onTaskKind(TaskKind.Inspection) },
                        ),
                    )
                }
                if (permissions.showRepair || app.isDemoMode) {
                    add(
                        TaskCenterCardItem(
                            id = "repair",
                            title = t(Str.RepairShort),
                            icon = OpsIcon.TaskRepair,
                            badgeText = badgeClaimed(repair.tasks),
                            onClick = { onTaskKind(TaskKind.Repair) },
                        ),
                    )
                }
            }
            TaskScaffold(
                areaName = homeState.currentArea?.name?.takeIf { it.isNotBlank() }
                    ?: t(Str.SelectServiceArea),
                pageTitle = t(Str.TaskCenter),
                cards = cards,
                onChangeArea = onChangeArea,
                emptyHint = t(Str.NoTaskPermission),
            )
        }
        TaskKind.ChangeBattery -> TaskDetailScaffold(
            app = app,
            title = t(Str.ChangeBatteryTaskTitle),
            onBack = { onTaskKind(TaskKind.Hub) },
            trailing = {
                TextButton(onClick = onOpenTaskMap) { Text(t(Str.TaskMap)) }
            },
        ) {
            ChangeBatteryTaskSection(app = app, currentArea = homeState.currentArea)
        }
        TaskKind.MoveCar -> TaskDetailScaffold(
            app = app,
            title = t(Str.MoveCarTaskTitle),
            onBack = { onTaskKind(TaskKind.Hub) },
            trailing = {
                TextButton(onClick = onOpenTaskMap) { Text(t(Str.TaskMap)) }
            },
        ) {
            MoveCarTaskSection(
                app = app,
                currentArea = homeState.currentArea,
                selectedCard = { task -> TaskSelectedCard(app = app, task = task) },
                onOpenBatchMove = { task ->
                    batchParentTaskId = task.batchParentId
                    onTaskKind(TaskKind.BatchMoveCar)
                },
                onOpenFreeMove = { onTaskKind(TaskKind.FreeMoveCar) },
            )
        }
        TaskKind.FreeMoveCar -> TaskDetailScaffold(
            app = app,
            title = t(Str.FreeMoveTaskTitle),
            onBack = { onTaskKind(TaskKind.MoveCar) },
        ) {
            FreeMoveCarSection(app = app, currentArea = homeState.currentArea)
        }
        TaskKind.BatchMoveCar -> TaskDetailScaffold(
            app = app,
            title = t(Str.BatchMoveTaskTitle),
            onBack = {
                app.batchMoveCarFeature.clear()
                batchParentTaskId = null
                onTaskKind(TaskKind.MoveCar)
            },
        ) {
            BatchMoveCarSection(app = app, parentTaskId = batchParentTaskId)
        }
        TaskKind.Inspection -> TaskDetailScaffold(
            app = app,
            title = t(Str.InspectionTaskTitle),
            onBack = { onTaskKind(TaskKind.Hub) },
        ) {
            ClaimableTaskSection(
                app = app,
                title = t(Str.InspectionShort),
                hint = t(Str.ClaimStartFinishHint),
                feature = app.inspectionTaskFeature,
                currentArea = homeState.currentArea,
            )
        }
        TaskKind.Repair -> TaskDetailScaffold(
            app = app,
            title = t(Str.RepairTaskTitle),
            onBack = { onTaskKind(TaskKind.Hub) },
        ) {
            ClaimableTaskSection(
                app = app,
                title = t(Str.RepairShort),
                hint = t(Str.ClaimStartFinishHint),
                feature = app.repairTaskFeature,
                currentArea = homeState.currentArea,
                showDragBack = true,
            )
        }
    }
}

/** Legacy: hide tip when total is 0. */
private fun badgeTotalCount(total: Int): String? =
    if (total <= 0) null else total.toString()

/** Legacy inspection/repair tip: received/total; hide when total is 0. */
private fun badgeClaimed(tasks: List<com.luopingtech.ebike.ops.domain.model.OpsTask>): String? {
    if (tasks.isEmpty()) return null
    val claimed = tasks.count { it.state == 1 }
    return "$claimed/${tasks.size}"
}

@Composable
private fun TaskDetailScaffold(
    app: OpsApp,
    title: String,
    onBack: () -> Unit,
    trailing: (@Composable () -> Unit)? = null,
    content: @Composable () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState())
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            TextButton(onClick = onBack) { Text("← ${t(Str.Back)}") }
            Text(
                text = title,
                style = MaterialTheme.typography.titleLarge,
                modifier = Modifier.weight(1f),
            )
            trailing?.invoke()
        }
        content()
    }
}

@Composable
private fun ScanOverlay(
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

    suspend fun handleRaw(raw: String) {
        val trimmed = raw.trim()
        if (trimmed.isEmpty() || busy || trimmed == lastRaw) return
        lastRaw = trimmed
        busy = true
        when (app.scanFeature.resolveManual(trimmed)) {
            is OpsResult.Err -> {
                busy = false
                // 识别错了允许马上再扫；同码也允许重试。error 已写入 scanFeature.state。
                lastRaw = ""
            }
            is OpsResult.Ok -> when (mode) {
                ScanMode.Detail -> {
                    // 详情留在本页展示 VehicleDetailSection；分析继续开着以便扫下一辆。
                    busy = false
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
        OpsCameraScanPreview(
            onCode = { raw -> scope.launch { handleRaw(raw) } },
            torchOn = torchOn,
            enabled = !busy && !showManual,
            modifier = Modifier
                .fillMaxWidth()
                .weight(1f),
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
                    iconRes = R.drawable.ic_scan_manual,
                    label = t(Str.InputCarNumber),
                    onClick = { showManual = !showManual },
                )
                ScanActionIcon(
                    iconRes = if (torchOn) R.drawable.ic_torch_on else R.drawable.ic_torch_off,
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

        // 详情单独占一块可滚动区域，不再跟预览抢 weight。
        if (mode == ScanMode.Detail && scanState.vehicle != null && !showManual) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .weight(1f)
                    .verticalScroll(rememberScrollState())
                    .background(MaterialTheme.colorScheme.background)
                    .padding(12.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                val vehicle = scanState.vehicle!!
                Text(
                    text = t(
                        Str.ParsedVehicleSummary,
                        vehicle.carId,
                        vehicle.batteryLabel,
                        vehicle.ridingLabel,
                    ),
                    style = MaterialTheme.typography.bodyMedium,
                )
                VehicleDetailSection(
                    app = app,
                    vehicle = vehicle,
                    serviceAreaId = homeState.currentArea?.id,
                    canBindBattery = permissions.canBindBatterySn,
                )
            }
        }
    }
}

@Composable
private fun ScanActionIcon(
    @androidx.annotation.DrawableRes iconRes: Int,
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
            painter = painterResource(iconRes),
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

@Composable
private fun WorkbenchTab(
    app: OpsApp,
    homeState: HomeUiState,
    permissions: OpsPermissions,
    trackPermissionHint: String?,
    onOpenWarehouse: () -> Unit,
    onOpenProduction: () -> Unit,
    onOpenFaultReport: () -> Unit,
    onOpenSneakReport: () -> Unit,
    onOpenUnlockedVehicles: () -> Unit,
    onOpenRelocation: () -> Unit,
    onOpenInspectionOrder: () -> Unit,
    onOpenRepairOrder: () -> Unit,
    onOpenH5: (H5ScreenKind) -> Unit,
    onChangeArea: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val authState by app.authFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    var settingsOpen by remember { mutableStateOf(false) }
    var showUpdatePwd by remember { mutableStateOf(false) }
    var showTrackSettings by remember { mutableStateOf(false) }
    var oldPwd by remember { mutableStateOf("") }
    var newPwd by remember { mutableStateOf("") }

    val context = LocalContext.current
    fun comingSoon() {
        android.widget.Toast.makeText(context, t(Str.FeatureComingSoon), android.widget.Toast.LENGTH_SHORT).show()
    }

    val opsItems = buildList {
        if (!permissions.showMaintainModule && !app.isDemoMode) return@buildList
        fun addModule(
            visible: Boolean,
            id: String,
            title: String,
            icon: OpsIcon,
            onClick: () -> Unit,
        ) {
            if (visible) add(WorkbenchModuleItem(id, title, icon, onClick))
        }
        addModule(permissions.showVehicleList, "vlist", t(Str.VehicleList), OpsIcon.VehicleList) { comingSoon() }
        addModule(permissions.showChangeBatteryTool, "batt", t(Str.ChangeBatteryTool), OpsIcon.ReplaceBattery) { comingSoon() }
        addModule(permissions.showMoveCarTool, "move", t(Str.MoveCarTool), OpsIcon.MoveVehicle) { comingSoon() }
        addModule(permissions.showFaultReport, "fault", t(Str.FaultReport), OpsIcon.Repair) {
            app.faultReportFeature.openHub()
            onOpenFaultReport()
        }
        addModule(permissions.showUnlockedVehicles, "unlocked", t(Str.UnlockedVehicles), OpsIcon.UnlockedVehicle, onOpenUnlockedVehicles)
        addModule(permissions.showBluetoothRadar, "ble", t(Str.BluetoothRadar), OpsIcon.BluetoothRadar) { comingSoon() }
        addModule(permissions.showOpsSetting, "ops-set", t(Str.OpsSettingTool), OpsIcon.OperationSetting) { comingSoon() }
        addModule(permissions.showMyTask, "my-task", t(Str.MyTaskTool), OpsIcon.MyTask) { comingSoon() }
        addModule(permissions.showVehicleTag, "vtag", t(Str.VehicleTagTool), OpsIcon.VehicleTag) { comingSoon() }
        addModule(permissions.showRelocation, "reloc", t(Str.Relocation), OpsIcon.Relocation, onOpenRelocation)
    }

    val operationItems = buildList {
        if (!permissions.showOperationModule && !app.isDemoMode) return@buildList
        fun addModule(
            visible: Boolean,
            id: String,
            title: String,
            icon: OpsIcon,
            onClick: () -> Unit,
        ) {
            if (visible) add(WorkbenchModuleItem(id, title, icon, onClick))
        }
        addModule(permissions.showParkingFence, "fence", t(Str.ParkingFence), OpsIcon.ParkingArea) { comingSoon() }
        addModule(
            permissions.has(OpsPermissionCodes.ORDER_QUERY),
            "h5-order",
            t(Str.OrderQueryScreen),
            OpsIcon.OrderQuery,
        ) {
            if (H5ScreenUrls.isConfigured(app.config, H5ScreenKind.Order)) {
                onOpenH5(H5ScreenKind.Order)
            } else {
                comingSoon()
            }
        }
        addModule(permissions.showOperationScreen, "h5-op", t(Str.OperationScreen), OpsIcon.OperationScreen) {
            if (H5ScreenUrls.isConfigured(app.config, H5ScreenKind.Operation)) {
                onOpenH5(H5ScreenKind.Operation)
            } else {
                comingSoon()
            }
        }
        addModule(permissions.showRevenueScreen, "h5-rev", t(Str.RevenueScreen), OpsIcon.RevenueScreen) {
            if (H5ScreenUrls.isConfigured(app.config, H5ScreenKind.Revenue)) {
                onOpenH5(H5ScreenKind.Revenue)
            } else {
                comingSoon()
            }
        }
        addModule(permissions.showStaffManage, "staff", t(Str.StaffManage), OpsIcon.EmployeeManager) { comingSoon() }
        addModule(permissions.showProfessionAudit, "prof", t(Str.ProfessionAudit), OpsIcon.ProfessionAudit) { comingSoon() }
        addModule(permissions.showObjectionOrder, "obj", t(Str.ObjectionOrder), OpsIcon.ObjectionOrder) { comingSoon() }
        addModule(permissions.showBlacklist, "black", t(Str.Blacklist), OpsIcon.BlackList) { comingSoon() }
        addModule(permissions.showIdBindAudit, "idbind", t(Str.IdBindAudit), OpsIcon.IdAudit) { comingSoon() }
        addModule(permissions.showOperationLog, "oplog", t(Str.OperationLog), OpsIcon.OperationLog) { comingSoon() }
    }

    val productionItems = buildList {
        if (permissions.showProductionDetect) {
            add(
                WorkbenchModuleItem(
                    id = "prod-detect",
                    title = t(Str.ProductionDetect),
                    icon = OpsIcon.VehicleInspection,
                    onClick = {
                        app.productionFeature.openHub()
                        onOpenProduction()
                    },
                ),
            )
        }
        if (permissions.showProductionBind) {
            add(
                WorkbenchModuleItem(
                    id = "prod-bind",
                    title = t(Str.ProductionBind),
                    icon = OpsIcon.CenterControlBind,
                    onClick = {
                        app.productionFeature.openHub()
                        onOpenProduction()
                    },
                ),
            )
        }
        if (permissions.showProductionShelves) {
            add(
                WorkbenchModuleItem(
                    id = "prod-shelves",
                    title = t(Str.ProductionShelves),
                    icon = OpsIcon.PutPullShelves,
                    onClick = {
                        app.productionFeature.openHub()
                        onOpenProduction()
                    },
                ),
            )
        }
    }

    val warehouseItems = buildList {
        if (permissions.showWarehouseIn) {
            add(
                WorkbenchModuleItem(
                    id = "wh-in",
                    title = t(Str.WarehouseIn),
                    icon = OpsIcon.InWarehouse,
                    onClick = {
                        app.warehouseFeature.openHub()
                        onOpenWarehouse()
                    },
                ),
            )
        }
        if (permissions.showWarehouseOut) {
            add(
                WorkbenchModuleItem(
                    id = "wh-out",
                    title = t(Str.WarehouseOut),
                    icon = OpsIcon.OutWarehouse,
                    onClick = {
                        app.warehouseFeature.openHub()
                        onOpenWarehouse()
                    },
                ),
            )
        }
        if (permissions.showWarehouseRecord) {
            add(
                WorkbenchModuleItem(
                    id = "wh-rec",
                    title = t(Str.WarehouseRecords),
                    icon = OpsIcon.WarehouseRecord,
                    onClick = {
                        app.warehouseFeature.openHub()
                        onOpenWarehouse()
                    },
                ),
            )
        }
    }

    val catalogItems = opsItems + operationItems + productionItems + warehouseItems
    val catalogById = catalogItems.associateBy { it.id }

    fun loadCommonIds(): List<String> =
        app.secureStore.getString(SecureStore.KEY_COMMON_MODULE_IDS)
            ?.split('|')
            ?.map { it.trim() }
            ?.filter { it.isNotBlank() }
            ?.distinct()
            .orEmpty()

    fun persistCommonIds(ids: List<String>) {
        app.secureStore.putString(
            SecureStore.KEY_COMMON_MODULE_IDS,
            ids.take(WORKBENCH_COMMON_MAX).joinToString("|"),
        )
    }

    var commonIds by remember {
        mutableStateOf(loadCommonIds().filter { catalogById.containsKey(it) }.take(WORKBENCH_COMMON_MAX))
    }
    var editingCommon by remember { mutableStateOf(false) }
    var draftCommonIds by remember { mutableStateOf(commonIds) }

    LaunchedEffect(catalogById.keys.joinToString()) {
        val filtered = commonIds.filter { catalogById.containsKey(it) }.take(WORKBENCH_COMMON_MAX)
        if (filtered != commonIds) {
            commonIds = filtered
            persistCommonIds(filtered)
        }
        if (!editingCommon) {
            draftCommonIds = filtered
        }
    }

    fun reorderIds(ids: List<String>, from: Int, to: Int): List<String> {
        if (from !in ids.indices || to !in ids.indices || from == to) return ids
        val mutable = ids.toMutableList()
        val item = mutable.removeAt(from)
        mutable.add(to, item)
        return mutable
    }

    val activeCommonIds = if (editingCommon) draftCommonIds else commonIds
    val commonItems = activeCommonIds.mapNotNull { id ->
        val base = catalogById[id] ?: return@mapNotNull null
        WorkbenchModuleItem(
            id = base.id,
            title = base.title,
            icon = base.icon,
            onClick = {
                if (!editingCommon) base.onClick()
            },
            editBadge = if (editingCommon) WorkbenchEditBadge.Remove else WorkbenchEditBadge.None,
            onBadgeClick = {
                draftCommonIds = draftCommonIds.filterNot { it == id }
            },
        )
    }

    fun decorateSection(items: List<WorkbenchModuleItem>): List<WorkbenchModuleItem> {
        if (!editingCommon) return items
        return items.map { item ->
            val alreadyPinned = draftCommonIds.contains(item.id)
            WorkbenchModuleItem(
                id = item.id,
                title = item.title,
                icon = item.icon,
                onClick = {},
                editBadge = if (alreadyPinned) WorkbenchEditBadge.None else WorkbenchEditBadge.Add,
                onBadgeClick = {
                    if (draftCommonIds.size >= WORKBENCH_COMMON_MAX) {
                        android.widget.Toast.makeText(
                            context,
                            t(Str.MaxCommonModules),
                            android.widget.Toast.LENGTH_SHORT,
                        ).show()
                    } else if (!draftCommonIds.contains(item.id)) {
                        draftCommonIds = draftCommonIds + item.id
                    }
                },
            )
        }
    }

    val sections = listOf(
        WorkbenchModuleSection(title = t(Str.OpsModule), items = decorateSection(opsItems)),
        WorkbenchModuleSection(title = t(Str.OperationModule), items = decorateSection(operationItems)),
        WorkbenchModuleSection(title = t(Str.ProductionModule), items = decorateSection(productionItems)),
        WorkbenchModuleSection(title = t(Str.WarehouseModule), items = decorateSection(warehouseItems)),
    )

    val companyName = authState.runtimeConfig?.tenantName
        ?.takeIf { it.isNotBlank() }
        ?: app.config.app.displayName.ifBlank { app.config.name }
    val userName = homeState.session?.displayName
        ?.takeIf { it.isNotBlank() }
        ?: homeState.session?.phone.orEmpty()
    val roleLabel = homeState.session?.roleName
        ?.takeIf { it.isNotBlank() }
        ?: t(Str.AdminRole)

    WorkbenchScaffold(
        areaName = homeState.currentArea?.name?.takeIf { it.isNotBlank() }
            ?: t(Str.SelectServiceArea),
        companyName = companyName,
        userName = userName,
        roleLabel = roleLabel,
        commonTitle = t(Str.CommonModules),
        editLabel = t(Str.Edit),
        settingsTitle = t(Str.Settings),
        backLabel = t(Str.Back),
        saveLabel = t(Str.Save),
        editCommonTitle = t(Str.EditCommonModules),
        dragHint = t(Str.DragCommonHint),
        commonItems = commonItems,
        sections = sections,
        settingsOpen = settingsOpen,
        editingCommon = editingCommon,
        onOpenSettings = { settingsOpen = true },
        onCloseSettings = { settingsOpen = false },
        onChangeArea = onChangeArea,
        onStartEditCommon = {
            draftCommonIds = commonIds
            editingCommon = true
        },
        onCancelEditCommon = {
            draftCommonIds = commonIds
            editingCommon = false
        },
        onSaveEditCommon = {
            commonIds = draftCommonIds.take(WORKBENCH_COMMON_MAX)
            persistCommonIds(commonIds)
            editingCommon = false
            android.widget.Toast.makeText(context, t(Str.Save), android.widget.Toast.LENGTH_SHORT).show()
        },
        onReorderCommon = { from, to ->
            if (editingCommon) {
                draftCommonIds = reorderIds(draftCommonIds, from, to)
            } else {
                commonIds = reorderIds(commonIds, from, to)
                persistCommonIds(commonIds)
            }
        },
        settingsContent = {
            TextButton(onClick = { showTrackSettings = !showTrackSettings }) {
                Text(if (showTrackSettings) t(Str.CollapseTrackUpload) else t(Str.TrackUpload))
            }
            if (showTrackSettings) {
                TrackUploadStatus(
                    app = app,
                    permissionHint = trackPermissionHint,
                )
            }
            TextButton(onClick = { showUpdatePwd = !showUpdatePwd }) {
                Text(if (showUpdatePwd) t(Str.CollapseChangePassword) else t(Str.ChangePassword))
            }
            if (showUpdatePwd) {
                OutlinedTextField(
                    value = oldPwd,
                    onValueChange = { oldPwd = it },
                    label = { Text(t(Str.OldPassword)) },
                    modifier = Modifier.fillMaxWidth(),
                    singleLine = true,
                )
                OutlinedTextField(
                    value = newPwd,
                    onValueChange = { newPwd = it },
                    label = { Text(t(Str.NewPasswordHint)) },
                    modifier = Modifier.fillMaxWidth(),
                    singleLine = true,
                )
                Button(
                    onClick = {
                        scope.launch { app.authFeature.updatePassword(oldPwd, newPwd) }
                    },
                    enabled = !authState.loading,
                    modifier = Modifier.fillMaxWidth(),
                ) {
                    Text(if (authState.loading) t(Str.Submitting) else t(Str.ConfirmChange))
                }
            }
            Text(t(Str.Language))
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
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
            Text(
                text = if (app.bleTransport.isAvailable) {
                    t(Str.BleModeSimulator)
                } else {
                    t(Str.BleModeNativeMissing)
                },
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
            authState.errorMessage?.let {
                Text(text = it, color = MaterialTheme.colorScheme.error)
            }
            authState.infoMessage?.let {
                Text(text = it, color = MaterialTheme.colorScheme.primary)
            }
            Button(
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
                modifier = Modifier.fillMaxWidth(),
            ) {
                Text(t(Str.Logout))
            }
            if (app.isDemoMode) {
                Text(
                    text = "shared ${OpsApp.LIBRARY_VERSION} · demo",
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            }
        },
    )
}

@Composable
private fun TrackUploadStatus(
    app: OpsApp,
    permissionHint: String?,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val trackState by app.trackUploadFeature.state.collectAsState()
    Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
        Text(
            text = buildString {
                append(t(Str.TrackUpload))
                append(" · ")
                append(if (trackState.enabled) t(Str.TrackOn) else t(Str.TrackOff))
                if (trackState.collecting) {
                    append(" · ")
                    append(t(Str.TrackCollecting))
                }
                append(" · ")
                append(t(Str.TrackPointsUploaded, trackState.uploadCount))
            },
            style = MaterialTheme.typography.bodySmall,
        )
        Text(
            text = t(Str.TrackAutoHint),
            style = MaterialTheme.typography.labelSmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
        trackState.lastMessage?.let {
            Text(it, color = MaterialTheme.colorScheme.primary, style = MaterialTheme.typography.labelSmall)
        }
        (permissionHint ?: trackState.errorMessage)?.let {
            Text(it, color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.labelSmall)
        }
        if (trackState.enabled) {
            TextButton(onClick = { app.trackUploadFeature.setEnabled(false) }) {
                Text(t(Str.PauseTrack))
            }
        } else {
            TextButton(onClick = { app.trackUploadFeature.setEnabled(true) }) {
                Text(t(Str.ResumeTrack))
            }
        }
    }
}

@Composable
private fun TaskSelectedCard(app: OpsApp, task: OpsTask?) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    Surface(
        tonalElevation = 1.dp,
        modifier = Modifier.fillMaxWidth(),
    ) {
        Column(
            modifier = Modifier.padding(12.dp),
            verticalArrangement = Arrangement.spacedBy(4.dp),
        ) {
            if (task == null) {
                Text(
                    text = t(Str.MapPickVehicle),
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            } else {
                Text(
                    text = "${task.carId} · ${task.stateLabel}",
                    style = MaterialTheme.typography.titleSmall,
                )
                Text(
                    text = buildString {
                        append(t(Str.BatteryPercent, task.restBattery))
                        if (task.imei.isNotBlank()) append(" · IMEI ${task.imei}")
                        append(" · id=${task.id}")
                    },
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
                Text(
                    text = task.address.ifBlank { task.areaName.ifBlank { t(Str.NoAddress) } },
                    style = MaterialTheme.typography.bodyMedium,
                )
            }
        }
    }
}

@Composable
private fun ChangeBatteryTaskSection(app: OpsApp, currentArea: ServiceArea?) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val taskState by app.changeBatteryTaskFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    val presets = remember(taskState.rangeMin, taskState.rangeMax) {
        listOf(taskState.rangeMin, 20, 30, taskState.rangeMax)
            .filter { it in taskState.rangeMin..taskState.rangeMax }
            .distinct()
            .sorted()
    }

    Text(
        text = t(Str.ChangeBatteryHint, taskState.rangeMin, taskState.rangeMax),
        style = MaterialTheme.typography.bodySmall,
        color = MaterialTheme.colorScheme.onSurfaceVariant,
    )
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        presets.forEach { value ->
            FilterChip(
                selected = taskState.maxBattery == value,
                onClick = {
                    scope.launch {
                        app.changeBatteryTaskFeature.applyMaxBattery(currentArea, value)
                    }
                },
                label = { Text("≤$value%") },
            )
        }
    }
    Button(
        onClick = { scope.launch { app.changeBatteryTaskFeature.load(currentArea) } },
        enabled = !taskState.loading,
        modifier = Modifier.fillMaxWidth(),
    ) { Text(if (taskState.loading) t(Str.LoadingEllipsis) else t(Str.Refresh)) }

    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        Button(
            onClick = {
                scope.launch {
                    when (val scan = app.codeScanner.scanOnce()) {
                        is OpsResult.Ok -> {
                            val hosts = app.authFeature.state.value.runtimeConfig?.qrHosts.orEmpty()
                            app.changeBatteryTaskFeature.selectByScanRaw(scan.value, hosts)
                        }
                        is OpsResult.Err -> {
                            // surface via feature message channel
                            app.changeBatteryTaskFeature.selectTask(null)
                        }
                    }
                }
            },
            enabled = !taskState.loading && taskState.tasks.isNotEmpty(),
            modifier = Modifier.weight(1f),
        ) { Text(t(Str.ScanSelectCar)) }
        if (app.isDemoMode) {
            Button(
                onClick = {
                    val carId = taskState.tasks.firstOrNull()?.carId ?: return@Button
                    app.changeBatteryTaskFeature.selectByScanRaw(carId)
                },
                enabled = taskState.tasks.isNotEmpty(),
                modifier = Modifier.weight(1f),
            ) { Text(t(Str.PickFirstCar)) }
        }
    }

    TaskSelectedCard(app = app, task = taskState.selected)

    var taskQuery by remember { mutableStateOf("") }
    OutlinedTextField(
        value = taskQuery,
        onValueChange = { taskQuery = it },
        label = { Text(t(Str.ListSearchHint)) },
        modifier = Modifier.fillMaxWidth(),
        singleLine = true,
    )
    val filteredTasks = remember(taskState.tasks, taskQuery) {
        val q = taskQuery.trim()
        if (q.isEmpty()) taskState.tasks
        else taskState.tasks.filter {
            it.carId.contains(q, ignoreCase = true) || it.imei.contains(q, ignoreCase = true)
        }
    }
    filteredTasks.forEach { task ->
        val selected = task.id == taskState.selectedTaskId
        Text(
            text = "${if (selected) "●" else "○"} ${task.carId} · ${task.restBattery}% · ${task.stateLabel}",
            modifier = Modifier
                .fillMaxWidth()
                .clickable { app.changeBatteryTaskFeature.selectTask(task.id) }
                .padding(vertical = 4.dp),
            color = if (selected) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.onSurface,
        )
    }
    Row(horizontalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.fillMaxWidth()) {
        Button(
            onClick = { scope.launch { app.changeBatteryTaskFeature.ringSelected() } },
            enabled = !taskState.loading && taskState.selected != null,
            modifier = Modifier.weight(1f),
        ) { Text(t(Str.Ring)) }
        Button(
            onClick = { scope.launch { app.changeBatteryTaskFeature.openBox() } },
            enabled = !taskState.loading && taskState.selected?.isActionable == true,
            modifier = Modifier.weight(1f),
        ) { Text(t(Str.OpenBatteryBox)) }
        Button(
            onClick = { scope.launch { app.changeBatteryTaskFeature.closeBox() } },
            enabled = !taskState.loading && taskState.selected != null,
            modifier = Modifier.weight(1f),
        ) { Text(t(Str.FinishChangeBattery)) }
    }
    taskState.message?.let { Text(it) }
    taskState.errorMessage?.let { Text(it, color = MaterialTheme.colorScheme.error) }
}

@Composable
private fun ClaimableTaskSection(
    app: OpsApp,
    title: String,
    hint: String,
    feature: ClaimableTaskFeature,
    currentArea: ServiceArea?,
    showDragBack: Boolean = false,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val taskState by feature.state.collectAsState()
    val scope = rememberCoroutineScope()
    var photoHint by remember { mutableStateOf<String?>(null) }

    /** 相机与相册只差取图那一步，权限与取消都由 [PhotoCapture] 实现方吞掉。 */
    fun addPhoto(take: suspend () -> OpsResult<String>) {
        scope.launch {
            when (val shot = take()) {
                is OpsResult.Ok -> {
                    feature.addPhotoUrl(shot.value)
                    photoHint = null
                }
                is OpsResult.Err -> photoHint = shot.error.message
            }
        }
    }

    Text(text = hint, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
    Button(
        onClick = { scope.launch { feature.load(currentArea) } },
        enabled = !taskState.loading,
        modifier = Modifier.fillMaxWidth(),
    ) { Text(if (taskState.loading) t(Str.LoadingEllipsis) else t(Str.Refresh)) }

    Button(
        onClick = {
            scope.launch {
                when (val scan = app.codeScanner.scanOnce()) {
                    is OpsResult.Ok -> {
                        val hosts = app.authFeature.state.value.runtimeConfig?.qrHosts.orEmpty()
                        feature.selectByScanRaw(scan.value, hosts)
                    }
                    is OpsResult.Err -> Unit
                }
            }
        },
        enabled = !taskState.loading && taskState.tasks.isNotEmpty(),
        modifier = Modifier.fillMaxWidth(),
    ) { Text(t(Str.ScanSelectCar)) }

    TaskSelectedCard(app = app, task = taskState.selected)

    var taskQuery by remember { mutableStateOf("") }
    OutlinedTextField(
        value = taskQuery,
        onValueChange = { taskQuery = it },
        label = { Text(t(Str.ListSearchHint)) },
        modifier = Modifier.fillMaxWidth(),
        singleLine = true,
    )
    val filteredTasks = remember(taskState.tasks, taskQuery) {
        val q = taskQuery.trim()
        if (q.isEmpty()) taskState.tasks
        else taskState.tasks.filter {
            it.carId.contains(q, ignoreCase = true) || it.imei.contains(q, ignoreCase = true)
        }
    }
    filteredTasks.forEach { task ->
        val selected = task.id == taskState.selectedTaskId
        Text(
            text = "${if (selected) "●" else "○"} ${task.carId} · ${task.stateLabel}",
            modifier = Modifier
                .fillMaxWidth()
                .clickable { feature.selectTask(task.id) }
                .padding(vertical = 4.dp),
            color = if (selected) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.onSurface,
        )
    }
    Row(horizontalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.fillMaxWidth()) {
        Button(
            onClick = { scope.launch { feature.claimSelected() } },
            enabled = !taskState.loading && taskState.selected?.state == 0,
            modifier = Modifier.weight(1f),
        ) { Text(t(Str.Claim)) }
        Button(
            onClick = { scope.launch { feature.startSelected() } },
            enabled = !taskState.loading &&
                taskState.selected != null &&
                taskState.selected!!.state != 2,
            modifier = Modifier.weight(1f),
        ) { Text(t(Str.Start)) }
        Button(
            onClick = { scope.launch { feature.finishSelected() } },
            enabled = !taskState.loading &&
                taskState.selected != null &&
                taskState.selected!!.state != 2 &&
                !taskState.selected!!.isDragBacking &&
                (
                    !taskState.needPhotograph ||
                        (taskState.photoUrls.isNotEmpty() && taskState.remark.isNotBlank())
                    ),
            modifier = Modifier.weight(1f),
        ) {
            Text(
                when {
                    taskState.loading -> t(Str.LoadingEllipsis)
                    taskState.needPhotograph -> t(Str.SubmitPhotoFinish)
                    else -> t(Str.Finish)
                },
            )
        }
    }
    TaskAuditResultSection(app = app, task = taskState.selected)
    if (showDragBack) {
        val selected = taskState.selected
        Text(
            text = t(Str.DragBackHint),
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
        OutlinedTextField(
            value = taskState.dragReason,
            onValueChange = { feature.setDragReason(it) },
            label = { Text(t(Str.DragBackReason)) },
            modifier = Modifier.fillMaxWidth(),
            enabled = selected != null && selected.dragState != 2 && selected.dragState != 3,
            singleLine = true,
        )
        OutlinedTextField(
            value = taskState.dragAddress,
            onValueChange = { feature.setDragAddress(it) },
            label = { Text(t(Str.DragBackAddress)) },
            modifier = Modifier.fillMaxWidth(),
            enabled = selected != null && selected.dragState != 2 && selected.dragState != 3,
            minLines = 2,
        )
        Button(
            onClick = { scope.launch { feature.createDragSelected() } },
            enabled = !taskState.loading && selected != null &&
                selected.dragState != 2 && selected.dragState != 3 &&
                taskState.dragReason.isNotBlank() && taskState.dragAddress.isNotBlank(),
            modifier = Modifier.fillMaxWidth(),
        ) {
            Text(
                when (selected?.dragState) {
                    2 -> t(Str.DragBackInProgress)
                    3 -> t(Str.DragBackDone)
                    else -> t(Str.DragBack)
                },
            )
        }
    }

    if (taskState.selected != null) {
        Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
            Text(
                text = when {
                    taskState.needPhotograph ->
                        t(Str.PhotoAuditCount, taskState.photoUrls.size)
                    taskState.photoUrls.isNotEmpty() ->
                        "$title · ${taskState.photoUrls.size}"
                    else -> t(Str.PhotoOptionalHint)
                },
                style = MaterialTheme.typography.bodySmall,
                color = if (taskState.needPhotograph) {
                    MaterialTheme.colorScheme.tertiary
                } else {
                    MaterialTheme.colorScheme.onSurfaceVariant
                },
            )
            if (taskState.needPhotograph || taskState.photoUrls.isNotEmpty()) {
                OutlinedTextField(
                    value = taskState.remark,
                    onValueChange = { feature.setRemark(it) },
                    label = {
                        Text(
                            if (taskState.needPhotograph) t(Str.PhotoRemarkHint) else t(Str.PhotoRemarkOptional),
                        )
                    },
                    modifier = Modifier.fillMaxWidth(),
                    minLines = 2,
                )
            }
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                Button(
                    onClick = { addPhoto { app.photoCapture.takePhoto(title.lowercase()) } },
                    modifier = Modifier.weight(1f),
                ) { Text(t(Str.Camera)) }
                Button(
                    onClick = { addPhoto { app.photoCapture.pickFromGallery() } },
                    modifier = Modifier.weight(1f),
                ) { Text(t(Str.Album)) }
                if (app.isDemoMode) {
                    Button(
                        onClick = { feature.addDemoPhoto() },
                        modifier = Modifier.weight(1f),
                    ) { Text(t(Str.DemoPhoto)) }
                }
            }
            if (taskState.photoUrls.isNotEmpty()) {
                Button(
                    onClick = { feature.clearPhotos() },
                    modifier = Modifier.fillMaxWidth(),
                ) { Text(t(Str.ClearPhotos)) }
            }
            photoHint?.let {
                Text(it, color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.bodySmall)
            }
            taskState.photoUrls.forEach { url ->
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                ) {
                    Text(
                        text = url.substringAfterLast('/').ifBlank { url }.take(48),
                        style = MaterialTheme.typography.bodySmall,
                        modifier = Modifier.weight(1f),
                    )
                    TextButton(onClick = { feature.removePhoto(url) }) {
                        Text(t(Str.Remove))
                    }
                }
            }
        }
    }

    taskState.message?.let { Text(it) }
    taskState.errorMessage?.let { Text(it, color = MaterialTheme.colorScheme.error) }
}

@Composable
private fun MapSurface(
    app: OpsApp,
    pins: List<MapPin>,
    selectedCarId: String?,
    mapReady: Boolean,
    mapProviderKind: String,
    onPinClick: (String) -> Unit,
    onClusterClick: (List<String>) -> Unit,
    modifier: Modifier = Modifier,
    clusterOverview: Boolean = true,
    fencePolygons: List<com.luopingtech.ebike.ops.domain.model.FencePolygon> = emptyList(),
    trackPoints: List<com.luopingtech.ebike.ops.domain.model.TrackPoint> = emptyList(),
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val isTencent = mapProviderKind.equals("tencent", ignoreCase = true) && mapReady
    when {
        isTencent -> {
            TencentMapView(
                pins = pins,
                selectedCarId = selectedCarId,
                onSelectCarId = onPinClick,
                onSelectCluster = onClusterClick,
                modifier = modifier,
                clusterOverview = clusterOverview,
                fencePolygons = fencePolygons,
                trackPoints = trackPoints,
            )
        }
        mapProviderKind.equals("none", ignoreCase = true) && !mapReady -> {
            Box(
                modifier = modifier
                    .background(MaterialTheme.colorScheme.surfaceVariant)
                    .border(1.dp, MaterialTheme.colorScheme.outlineVariant)
                    .padding(12.dp),
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    text = t(Str.MapNotEnabled),
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    textAlign = TextAlign.Center,
                )
            }
        }
        else -> {
            val label = when {
                mapProviderKind.equals("simulator", ignoreCase = true) -> t(Str.SimulatorMap)
                mapProviderKind.equals("tencent", ignoreCase = true) -> t(Str.MapSimNoTencentKey)
                mapProviderKind.equals("google", ignoreCase = true) -> t(Str.MapSimNoGoogle)
                else -> "${t(Str.SimulatorMap)} · $mapProviderKind"
            }
            SimulatorMapView(
                pins = pins,
                selectedCarId = selectedCarId,
                providerLabel = label,
                onSelectCarId = onPinClick,
                onSelectCluster = onClusterClick,
                modifier = modifier,
            )
        }
    }
}

@Composable
private fun VehicleRow(
    app: OpsApp,
    vehicle: Vehicle,
    selected: Boolean,
    onClick: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    Text(
        text = "${if (selected) "●" else "○"} ${vehicle.carId} · ${vehicle.batteryLabel} · ${vehicle.ridingLabel}" +
            if (vehicle.isOnline) " · ${t(Str.Online)}" else " · ${t(Str.Offline)}",
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .padding(vertical = 4.dp),
        color = if (selected) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.onSurface,
        style = MaterialTheme.typography.bodyMedium,
    )
}
