package com.luopingtech.ebike.ops.ui.scan

import androidx.activity.compose.BackHandler
import androidx.compose.foundation.background
import androidx.compose.foundation.border
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
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.Text
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
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

enum class ScanInputMode { Detail, Unlock, Lock }

/**
 * Legacy VehicleNumberUnlockActivity：白底输入编号页。
 * - 提示「请输入车辆号」
 * - 数字输入框 22sp / 高 48dp / 最多 10 位
 * - 按钮文案随模式：进入详情 / 确认开锁 / 确认关锁
 */
@Composable
fun VehicleNumberInputScreen(
    app: OpsApp,
    mode: ScanInputMode,
    onClose: () -> Unit,
    onOpenVehicleDetail: (carId: String) -> Unit,
    modifier: Modifier = Modifier,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val colors = OpsTheme.colors
    val scope = rememberCoroutineScope()
    val scanState by app.scanFeature.state.collectAsState()
    var input by remember { mutableStateOf("") }
    var busy by remember { mutableStateOf(false) }
    val focusRequester = remember { FocusRequester() }

    BackHandler(onBack = onClose)

    LaunchedEffect(Unit) {
        delay(600)
        runCatching { focusRequester.requestFocus() }
    }

    val digitsOnly = input.filter { it.isDigit() }.take(10)
    if (digitsOnly != input) input = digitsOnly
    val valid = digitsOnly.length in 1..10
    val filledStyle = digitsOnly.length >= 7

    val actionLabel = when (mode) {
        ScanInputMode.Detail -> t(Str.EnterTheDetails)
        ScanInputMode.Unlock -> t(Str.SureUnlock)
        ScanInputMode.Lock -> t(Str.SureLock)
    }

    fun submit() {
        if (!valid || busy) return
        scope.launch {
            busy = true
            when (val r = app.scanFeature.resolveManual(digitsOnly)) {
                is OpsResult.Err -> busy = false
                is OpsResult.Ok -> when (mode) {
                    ScanInputMode.Detail -> {
                        busy = false
                        onOpenVehicleDetail(r.value.carId)
                    }
                    ScanInputMode.Unlock -> {
                        app.scanFeature.unlock()
                        delay(1_200)
                        app.scanFeature.clear()
                        busy = false
                        onClose()
                    }
                    ScanInputMode.Lock -> {
                        app.scanFeature.lock()
                        delay(1_200)
                        app.scanFeature.clear()
                        busy = false
                        onClose()
                    }
                }
            }
        }
    }

    Column(
        modifier = modifier
            .fillMaxSize()
            .background(Color.White)
            .statusBarsPadding(),
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(48.dp)
                .background(colors.primary),
        ) {
            Text(
                text = "\u2039",
                color = Color.White,
                fontSize = 28.sp,
                modifier = Modifier
                    .align(Alignment.CenterStart)
                    .clickable(
                        interactionSource = remember { MutableInteractionSource() },
                        indication = null,
                        onClick = onClose,
                    )
                    .padding(horizontal = 16.dp),
            )
            Text(
                text = t(Str.InputCarNumber),
                color = Color.White,
                fontSize = 18.sp,
                fontWeight = FontWeight.Medium,
                modifier = Modifier.align(Alignment.Center),
            )
        }

        Text(
            text = t(Str.PleaseEnterVehicleNumber),
            color = Color(0xFF666666),
            fontSize = 16.sp,
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp)
                .padding(top = 96.dp),
        )

        BasicTextField(
            value = digitsOnly,
            onValueChange = { raw ->
                input = raw.filter { it.isDigit() }.take(10)
            },
            singleLine = true,
            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
            textStyle = TextStyle(
                color = Color.Black,
                fontSize = 22.sp,
            ),
            cursorBrush = SolidColor(colors.primary),
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp)
                .padding(top = 24.dp)
                .height(48.dp)
                .border(
                    width = 1.dp,
                    color = if (filledStyle) Color(0xFFCCCCCC) else colors.primary,
                    shape = RoundedCornerShape(4.dp),
                )
                .background(
                    color = if (filledStyle) Color.White else Color.Transparent,
                    shape = RoundedCornerShape(4.dp),
                )
                .padding(horizontal = 8.dp, vertical = 8.dp)
                .focusRequester(focusRequester),
            decorationBox = { inner ->
                Box(contentAlignment = Alignment.CenterStart) {
                    if (digitsOnly.isEmpty()) {
                        Text(
                            text = t(Str.PleaseEnterVehicleNumber),
                            color = Color(0xFF999999),
                            fontSize = 22.sp,
                        )
                    }
                    inner()
                }
            },
        )

        Box(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp)
                .padding(top = 96.dp)
                .height(48.dp)
                .background(
                    color = if (valid && !busy) colors.primary else colors.primary.copy(alpha = 0.35f),
                    shape = RoundedCornerShape(4.dp),
                )
                .clickable(
                    enabled = valid && !busy,
                    interactionSource = remember { MutableInteractionSource() },
                    indication = null,
                    onClick = { submit() },
                ),
            contentAlignment = Alignment.Center,
        ) {
            Text(
                text = if (busy) t(Str.LoadingEllipsis) else actionLabel,
                color = Color.White,
                fontSize = 16.sp,
                fontWeight = FontWeight.Medium,
            )
        }

        scanState.errorMessage?.let {
            Text(
                text = it,
                color = Color(0xFFE02020),
                fontSize = 13.sp,
                modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp),
            )
        }
        scanState.message?.let {
            Text(
                text = it,
                color = colors.primary,
                fontSize = 13.sp,
                modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp),
            )
        }
        Spacer(modifier = Modifier.weight(1f))
    }
}
