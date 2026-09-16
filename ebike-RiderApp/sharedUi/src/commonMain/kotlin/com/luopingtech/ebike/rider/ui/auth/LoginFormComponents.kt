package com.luopingtech.ebike.rider.ui.auth

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Checkbox
import androidx.compose.material3.CheckboxDefaults
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.rider.ui.theme.RiderTheme

@Composable
fun LoginWelcomeHeader(
    welcome: String,
    appName: String,
    modifier: Modifier = Modifier,
) {
    val colors = RiderTheme.colors
    Column(modifier = modifier.fillMaxWidth()) {
        Text(
            text = welcome,
            color = colors.primary,
            fontSize = 32.sp,
            fontWeight = FontWeight.Normal,
            lineHeight = 40.sp,
        )
        Text(
            text = appName,
            color = Color.Black,
            fontSize = 32.sp,
            fontWeight = FontWeight.Normal,
            lineHeight = 40.sp,
        )
    }
}

@Composable
fun UnderlineTextField(
    value: String,
    onValueChange: (String) -> Unit,
    hint: String,
    modifier: Modifier = Modifier,
    keyboardType: KeyboardType = KeyboardType.Text,
    trailing: (@Composable () -> Unit)? = null,
) {
    val colors = RiderTheme.colors
    Column(modifier = modifier.fillMaxWidth()) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .height(40.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Box(modifier = Modifier.weight(1f)) {
                if (value.isEmpty()) {
                    Text(
                        text = hint,
                        color = colors.textTertiary,
                        fontSize = 20.sp,
                    )
                }
                BasicTextField(
                    value = value,
                    onValueChange = onValueChange,
                    modifier = Modifier.fillMaxWidth(),
                    singleLine = true,
                    textStyle = TextStyle(
                        color = colors.textPrimary,
                        fontSize = 20.sp,
                        fontWeight = FontWeight.Normal,
                    ),
                    cursorBrush = SolidColor(colors.primary),
                    keyboardOptions = KeyboardOptions(keyboardType = keyboardType),
                )
            }
            if (trailing != null) {
                Spacer(modifier = Modifier.width(8.dp))
                trailing()
            }
        }
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(1.dp)
                .background(colors.divider),
        )
    }
}

@Composable
fun LoginPrimaryButton(
    text: String,
    enabled: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val colors = RiderTheme.colors
    Button(
        onClick = onClick,
        enabled = enabled,
        modifier = modifier
            .fillMaxWidth()
            .height(48.dp),
        shape = RoundedCornerShape(4.dp),
        contentPadding = PaddingValues(0.dp),
        colors = ButtonDefaults.buttonColors(
            containerColor = colors.primary,
            contentColor = colors.onPrimary,
            disabledContainerColor = colors.primary.copy(alpha = 0.3f),
            disabledContentColor = colors.onPrimary,
        ),
        elevation = ButtonDefaults.buttonElevation(
            defaultElevation = 0.dp,
            pressedElevation = 0.dp,
            disabledElevation = 0.dp,
            focusedElevation = 0.dp,
            hoveredElevation = 0.dp,
        ),
    ) {
        Text(
            text = text,
            fontSize = 16.sp,
            fontWeight = FontWeight.Normal,
            textAlign = TextAlign.Center,
        )
    }
}

@Composable
fun GetSmsCodeAction(
    label: String,
    enabled: Boolean,
    onClick: () -> Unit,
) {
    val colors = RiderTheme.colors
    val bg = if (enabled) colors.primary else colors.disabled
    Box(
        modifier = Modifier
            .width(100.dp)
            .height(36.dp)
            .background(bg, RoundedCornerShape(4.dp))
            .clickable(
                enabled = enabled,
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            ),
        contentAlignment = Alignment.Center,
    ) {
        Text(
            text = label,
            color = colors.onPrimary,
            fontSize = 14.sp,
            textAlign = TextAlign.Center,
        )
    }
}

@Composable
fun ProtocolAgreeRow(
    agreed: Boolean,
    prefix: String,
    userAgreement: String,
    joiner: String,
    privacyPolicy: String,
    onAgreedChange: (Boolean) -> Unit,
    onUserAgreement: () -> Unit,
    onPrivacyPolicy: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val colors = RiderTheme.colors
    // 不要再给 Row 套 clickable：Checkbox 自己会消费点击，父子各 toggle 一次等于没勾上。
    Row(
        modifier = modifier.fillMaxWidth(),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Checkbox(
            checked = agreed,
            onCheckedChange = onAgreedChange,
            modifier = Modifier.size(20.dp),
            colors = CheckboxDefaults.colors(
                checkedColor = colors.primary,
                uncheckedColor = colors.divider,
            ),
        )
        Spacer(modifier = Modifier.width(4.dp))
        Text(
            text = prefix,
            color = colors.textTertiary,
            fontSize = 12.sp,
            modifier = Modifier.clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = { onAgreedChange(!agreed) },
            ),
        )
        Text(
            text = userAgreement,
            color = colors.primary,
            fontSize = 12.sp,
            modifier = Modifier.clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onUserAgreement,
            ),
        )
        Text(text = joiner, color = colors.textTertiary, fontSize = 12.sp, modifier = Modifier.padding(horizontal = 2.dp))
        Text(
            text = privacyPolicy,
            color = colors.primary,
            fontSize = 12.sp,
            modifier = Modifier.clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onPrivacyPolicy,
            ),
        )
    }
}
