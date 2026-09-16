package com.luopingtech.ebike.rider.ui.ride

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.OutlinedTextFieldDefaults
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.clipToBounds
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.rider.RiderApp
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.domain.scan.ScanCodeParser
import com.luopingtech.ebike.rider.domain.scan.ScanTarget
import com.luopingtech.ebike.rider.ui.feedback.LocalRiderToast
import com.luopingtech.ebike.rider.ui.platform.RiderBackHandler
import com.luopingtech.ebike.rider.ui.scan.LocalRiderScanPreview
import com.luopingtech.ebike.rider.ui.theme.RiderTheme

/**
 * 用车入口：顶栏返回（在预览外，不被 Camera 盖住）+ 中部预览 + 底部手输贴底。
 */
@Composable
fun ScanScreen(
    app: RiderApp,
    onScanned: (carId: String, imei: String) -> Unit,
    onBack: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val toast = LocalRiderToast.current
    val preview = LocalRiderScanPreview.current
    var torchOn by remember { mutableStateOf(false) }
    var handled by remember { mutableStateOf(false) }
    var input by remember { mutableStateOf("") }
    val brand = RiderTheme.colors.primary
    val chrome = Color(0xFF111111)

    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)

    val normalized = input.trim().uppercase()
    val valid = normalized.length >= MIN_LENGTH &&
        normalized.length <= MAX_LENGTH &&
        normalized.all { it.isLetterOrDigit() }

    fun submitManual() {
        if (!valid || handled) return
        handled = true
        onScanned(normalized, "")
    }

    fun onRawCode(raw: String) {
        if (handled) return
        when (val target = ScanCodeParser.parse(raw)) {
            is ScanTarget.CarId -> {
                handled = true
                onScanned(target.value, "")
            }
            is ScanTarget.Imei -> {
                handled = true
                onScanned(target.value, target.value)
            }
            null -> toast(t(Str.ScanUnrecognized))
        }
    }

    // 系统返回 / 手势：回首页，不要 finish Activity
    RiderBackHandler(enabled = !handled, onBack = onBack)

    Column(
        modifier = modifier
            .fillMaxSize()
            .background(chrome),
    ) {
        // 顶栏：小圆形返回，避免 TextButton + 大号「＜」占半屏
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .statusBarsPadding()
                .background(chrome)
                .padding(horizontal = 8.dp, vertical = 4.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Box(
                modifier = Modifier
                    .size(36.dp)
                    .clip(CircleShape)
                    .clickable(onClick = onBack),
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    text = "‹",
                    color = Color.White,
                    fontSize = 26.sp,
                    fontWeight = FontWeight.Light,
                    textAlign = TextAlign.Center,
                )
            }
        }

        Box(
            modifier = Modifier
                .fillMaxWidth()
                .weight(1f)
                .clipToBounds()
                .background(chrome),
        ) {
            preview.Preview(
                modifier = Modifier.fillMaxSize(),
                torchOn = torchOn,
                enabled = !handled,
                onCode = { onRawCode(it) },
            )

            // 取景提示 + 手电筒：贴预览区底边，不压进下方表单
            Column(
                modifier = Modifier
                    .align(Alignment.BottomCenter)
                    .padding(bottom = 12.dp, start = 24.dp, end = 24.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
            ) {
                Text(
                    text = t(Str.ScanHint),
                    style = MaterialTheme.typography.bodyMedium,
                    color = Color(0xFFEEEEEE),
                    textAlign = TextAlign.Center,
                )
                Spacer(Modifier.height(10.dp))
                Column(
                    modifier = Modifier.clickable { torchOn = !torchOn },
                    horizontalAlignment = Alignment.CenterHorizontally,
                ) {
                    Box(
                        modifier = Modifier
                            .size(40.dp)
                            .clip(CircleShape)
                            .background(Color.White.copy(alpha = 0.2f)),
                        contentAlignment = Alignment.Center,
                    ) {
                        Text(
                            text = if (torchOn) "ON" else "✦",
                            color = Color.White,
                            style = MaterialTheme.typography.labelMedium,
                        )
                    }
                    Spacer(Modifier.height(4.dp))
                    Text(
                        text = if (torchOn) t(Str.TorchOff) else t(Str.TorchOn),
                        color = Color.White,
                        style = MaterialTheme.typography.labelMedium,
                    )
                }
            }
        }

        Surface(
            modifier = Modifier
                .fillMaxWidth()
                .navigationBarsPadding(),
            color = Color.White,
            shape = RoundedCornerShape(topStart = 16.dp, topEnd = 16.dp),
            shadowElevation = 6.dp,
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 24.dp, vertical = 20.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
            ) {
                Text(
                    text = t(Str.ManualIdHint),
                    style = MaterialTheme.typography.titleMedium,
                    color = RiderTheme.colors.textPrimary,
                    textAlign = TextAlign.Center,
                    fontWeight = FontWeight.Medium,
                    modifier = Modifier.padding(bottom = 16.dp),
                )

                OutlinedTextField(
                    value = input,
                    onValueChange = { raw ->
                        input = raw.filter { it.isLetterOrDigit() }.take(MAX_LENGTH)
                    },
                    singleLine = true,
                    placeholder = {
                        Text(
                            t(Str.ManualIdPlaceholder),
                            color = RiderTheme.colors.textTertiary,
                        )
                    },
                    isError = input.isNotBlank() && !valid,
                    supportingText = {
                        if (input.isNotBlank() && !valid) {
                            Text(t(Str.ManualIdTooShort, MIN_LENGTH))
                        }
                    },
                    shape = RoundedCornerShape(28.dp),
                    colors = OutlinedTextFieldDefaults.colors(
                        focusedBorderColor = brand,
                        unfocusedBorderColor = Color(0xFFDDDDDD),
                    ),
                    keyboardOptions = KeyboardOptions(
                        keyboardType = KeyboardType.Number,
                        imeAction = ImeAction.Done,
                    ),
                    keyboardActions = KeyboardActions(onDone = { submitManual() }),
                    modifier = Modifier.fillMaxWidth(),
                )

                Spacer(Modifier.height(16.dp))

                Button(
                    onClick = { submitManual() },
                    enabled = valid && !handled,
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(48.dp),
                    shape = RoundedCornerShape(24.dp),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = brand,
                        disabledContainerColor = Color(0xFFCCCCCC),
                        disabledContentColor = Color.White,
                    ),
                ) {
                    Text(
                        text = t(Str.ManualIdSubmit),
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.SemiBold,
                    )
                }
            }
        }
    }
}

private const val MIN_LENGTH = 7
private const val MAX_LENGTH = 20
