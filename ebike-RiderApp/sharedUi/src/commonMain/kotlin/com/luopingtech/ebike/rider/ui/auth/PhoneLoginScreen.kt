package com.luopingtech.ebike.rider.ui.auth

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.rider.RiderApp
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.ui.theme.RiderTheme
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

@Composable
fun PhoneLoginScreen(
    app: RiderApp,
    modifier: Modifier = Modifier,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?): String {
        language
        return app.i18n.t(key, *args)
    }
    val authState by app.authFeature.state.collectAsState()
    val scope = rememberCoroutineScope()
    var smsCode by remember { mutableStateOf("") }

    LaunchedEffect(authState.countdown) {
        if (authState.countdown > 0) {
            delay(1000)
            app.authFeature.tickCountdown()
        }
    }

    val phone = authState.phone
    val canSend = !authState.loading && phone.isNotBlank() && authState.countdown == 0 && authState.agreedProtocol
    val canLogin = !authState.loading && phone.isNotBlank() && smsCode.isNotBlank() && authState.agreedProtocol
    val smsLabel = when {
        authState.countdown > 0 -> "${authState.countdown}s"
        authState.smsSent -> t(Str.ResendSmsCode)
        else -> t(Str.GetSmsCode)
    }
    val appName = app.config.app.displayName.ifBlank { t(Str.AppName) }

    Column(
        modifier = modifier
            .fillMaxSize()
            .background(Color.White)
            .statusBarsPadding()
            .padding(horizontal = 32.dp),
    ) {
        Spacer(modifier = Modifier.height(99.dp))
        LoginWelcomeHeader(
            welcome = t(Str.WelcomeToUse),
            appName = appName,
        )
        Spacer(modifier = Modifier.height(48.dp))
        UnderlineTextField(
            value = phone,
            onValueChange = { app.authFeature.setPhone(it) },
            hint = t(Str.PhoneHint),
            keyboardType = KeyboardType.Phone,
        )
        Spacer(modifier = Modifier.height(28.dp))
        UnderlineTextField(
            value = smsCode,
            onValueChange = { smsCode = it },
            hint = t(Str.SmsCodeHint),
            keyboardType = KeyboardType.Number,
            trailing = {
                GetSmsCodeAction(
                    label = smsLabel,
                    enabled = canSend,
                    onClick = { scope.launch { app.authFeature.sendSms(phone) } },
                )
            },
        )
        Spacer(modifier = Modifier.height(24.dp))
        ProtocolAgreeRow(
            agreed = authState.agreedProtocol,
            prefix = t(Str.AgreeProtocolPrefix),
            userAgreement = t(Str.UserAgreement),
            joiner = t(Str.ProtocolAnd),
            privacyPolicy = t(Str.PrivacyPolicy),
            onAgreedChange = { app.authFeature.setAgreedProtocol(it) },
            onUserAgreement = { /* 协议 H5 未接：先勾选即可登录 */ },
            onPrivacyPolicy = { /* 协议 H5 未接：先勾选即可登录 */ },
        )
        Spacer(modifier = Modifier.height(32.dp))
        LoginPrimaryButton(
            text = if (authState.loading) t(Str.LoggingIn) else t(Str.Login),
            enabled = canLogin && authState.agreedProtocol,
            onClick = {
                scope.launch { app.authFeature.loginWithSms(phone, smsCode) }
            },
        )
        authState.error?.let {
            Spacer(modifier = Modifier.height(16.dp))
            Text(text = it, color = MaterialTheme.colorScheme.error, fontSize = 14.sp)
        }
        authState.info?.let {
            Spacer(modifier = Modifier.height(16.dp))
            Text(text = it, color = RiderTheme.colors.primary, fontSize = 14.sp)
        }
        Spacer(modifier = Modifier.weight(1f))
        Text(
            text = if (app.isDemoMode) {
                t(Str.DemoLoginHint)
            } else {
                t(
                    Str.LiveLoginHint,
                    app.config.app.displayName.ifBlank { app.config.alias },
                    app.config.api.baseUrl,
                )
            },
            color = RiderTheme.colors.textTertiary,
            fontSize = 11.sp,
            modifier = Modifier
                .fillMaxWidth()
                .padding(bottom = 16.dp),
            textAlign = TextAlign.Center,
        )
    }
}
