package com.luopingtech.ebike.ops.ui.auth

import androidx.compose.foundation.Canvas
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.IntrinsicSize
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
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.text.input.VisualTransformation
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.ui.theme.OpsTheme

@OptIn(ExperimentalFoundationApi::class)
@Composable
fun LoginWelcomeHeader(
    welcome: String,
    appName: String,
    modifier: Modifier = Modifier,
    onWelcomeLongClick: (() -> Unit)? = null,
) {
    val colors = OpsTheme.colors
    Column(modifier = modifier.fillMaxWidth()) {
        Text(
            text = welcome,
            color = colors.primary,
            fontSize = 32.sp,
            fontWeight = FontWeight.Normal,
            lineHeight = 40.sp,
            modifier = if (onWelcomeLongClick != null) {
                Modifier.combinedClickable(
                    interactionSource = remember { MutableInteractionSource() },
                    indication = null,
                    onClick = {},
                    onLongClick = onWelcomeLongClick,
                )
            } else {
                Modifier
            },
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
fun LoginModeTabs(
    passwordLabel: String,
    smsLabel: String,
    selectedPassword: Boolean,
    onSelectPassword: () -> Unit,
    onSelectSms: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Row(
        modifier = modifier
            .fillMaxWidth()
            .height(40.dp),
        verticalAlignment = Alignment.Bottom,
    ) {
        LoginModeTab(
            label = passwordLabel,
            selected = selectedPassword,
            onClick = onSelectPassword,
        )
        Spacer(modifier = Modifier.width(32.dp))
        LoginModeTab(
            label = smsLabel,
            selected = !selectedPassword,
            onClick = onSelectSms,
        )
    }
}

@Composable
private fun LoginModeTab(
    label: String,
    selected: Boolean,
    onClick: () -> Unit,
) {
    val colors = OpsTheme.colors
    Column(
        modifier = Modifier
            .width(IntrinsicSize.Max)
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            ),
        horizontalAlignment = Alignment.Start,
    ) {
        Text(
            text = label,
            color = if (selected) colors.primary else colors.tabUnselected,
            fontSize = 18.sp,
            fontWeight = if (selected) FontWeight.Bold else FontWeight.Normal,
        )
        Spacer(modifier = Modifier.height(6.dp))
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(3.dp)
                .background(
                    color = if (selected) colors.primary else Color.Transparent,
                    shape = RoundedCornerShape(1.5.dp),
                ),
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
    password: Boolean = false,
    passwordVisible: Boolean = true,
    onTogglePasswordVisible: (() -> Unit)? = null,
    trailing: (@Composable () -> Unit)? = null,
) {
    val colors = OpsTheme.colors
    val transformation = if (password && !passwordVisible) {
        PasswordVisualTransformation()
    } else {
        VisualTransformation.None
    }
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
                    visualTransformation = transformation,
                )
            }
            if (password && onTogglePasswordVisible != null) {
                PasswordEyeIcon(
                    visible = passwordVisible,
                    onClick = onTogglePasswordVisible,
                    modifier = Modifier.padding(start = 12.dp),
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
private fun PasswordEyeIcon(
    visible: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val color = Color(0xFFB0B0B0)
    Canvas(
        modifier = modifier
            .size(24.dp)
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            ),
    ) {
        val stroke = Stroke(width = 1.6.dp.toPx(), cap = StrokeCap.Round)
        val cx = size.width / 2f
        val cy = size.height / 2f
        val path = Path().apply {
            moveTo(size.width * 0.12f, cy)
            quadraticBezierTo(cx, size.height * 0.18f, size.width * 0.88f, cy)
            quadraticBezierTo(cx, size.height * 0.82f, size.width * 0.12f, cy)
            close()
        }
        drawPath(path, color = color, style = stroke)
        drawCircle(
            color = color,
            radius = size.minDimension * 0.14f,
            center = Offset(cx, cy),
            style = stroke,
        )
        if (!visible) {
            drawLine(
                color = color,
                start = Offset(size.width * 0.18f, size.height * 0.78f),
                end = Offset(size.width * 0.82f, size.height * 0.22f),
                strokeWidth = 1.6.dp.toPx(),
                cap = StrokeCap.Round,
            )
        }
    }
}

@Composable
fun LoginPrimaryButton(
    text: String,
    enabled: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val colors = OpsTheme.colors
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
    val colors = OpsTheme.colors
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
