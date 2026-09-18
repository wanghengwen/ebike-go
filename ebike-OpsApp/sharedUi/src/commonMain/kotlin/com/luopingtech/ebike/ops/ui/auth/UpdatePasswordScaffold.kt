package com.luopingtech.ebike.ops.ui.auth

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.domain.auth.PasswordRules
import com.luopingtech.ebike.ops.ui.theme.OpsTheme

/**
 * Legacy [UpdatePwdActivity] / activity_update_password.xml
 */
@Composable
fun UpdatePasswordScaffold(
    title: String,
    originalTitle: String,
    originalHint: String,
    originalValue: String,
    onOriginalChange: (String) -> Unit,
    originalVisible: Boolean,
    onToggleOriginalVisible: () -> Unit,
    newTitle: String,
    newHint: String,
    newValue: String,
    onNewChange: (String) -> Unit,
    newVisible: Boolean,
    onToggleNewVisible: () -> Unit,
    confirmLabel: String,
    loading: Boolean,
    errorMessage: String?,
    backLabel: String,
    onBack: () -> Unit,
    onConfirm: () -> Unit,
) {
    val colors = OpsTheme.colors
    val canSubmit = PasswordRules.canSubmitUpdatePwd(originalValue, newValue) && !loading

    com.luopingtech.ebike.ops.ui.navigation.OpsBackHandler(onBack = onBack)

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.White),
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .background(colors.primary)
                .statusBarsPadding()
                .height(50.dp)
                .padding(horizontal = 12.dp),
        ) {
            Text(
                text = backLabel,
                color = colors.onPrimary,
                fontSize = 16.sp,
                modifier = Modifier
                    .align(Alignment.CenterStart)
                    .clickable(
                        interactionSource = remember { MutableInteractionSource() },
                        indication = null,
                        onClick = onBack,
                    ),
            )
            Text(
                text = title,
                color = colors.onPrimary,
                fontSize = 18.sp,
                fontWeight = FontWeight.Medium,
                modifier = Modifier.align(Alignment.Center),
            )
        }
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 32.dp),
        ) {
            Spacer(modifier = Modifier.height(28.dp))
            Text(
                text = originalTitle,
                color = Color(0xFF242936),
                fontSize = 15.sp,
            )
            Spacer(modifier = Modifier.height(8.dp))
            UnderlineTextField(
                value = originalValue,
                onValueChange = { onOriginalChange(PasswordRules.filterInput(it)) },
                hint = originalHint,
                password = true,
                passwordVisible = originalVisible,
                onTogglePasswordVisible = onToggleOriginalVisible,
                keyboardType = KeyboardType.Password,
            )
            Spacer(modifier = Modifier.height(28.dp))
            Text(
                text = newTitle,
                color = Color(0xFF242936),
                fontSize = 15.sp,
            )
            Spacer(modifier = Modifier.height(8.dp))
            UnderlineTextField(
                value = newValue,
                onValueChange = { onNewChange(PasswordRules.filterInput(it)) },
                hint = newHint,
                password = true,
                passwordVisible = newVisible,
                onTogglePasswordVisible = onToggleNewVisible,
                keyboardType = KeyboardType.Password,
            )
            Spacer(modifier = Modifier.height(144.dp))
            LoginPrimaryButton(
                text = confirmLabel,
                enabled = canSubmit,
                onClick = onConfirm,
            )
            errorMessage?.let {
                Spacer(modifier = Modifier.height(12.dp))
                Text(text = it, color = Color(0xFFE53935), fontSize = 13.sp)
            }
        }
    }
}
