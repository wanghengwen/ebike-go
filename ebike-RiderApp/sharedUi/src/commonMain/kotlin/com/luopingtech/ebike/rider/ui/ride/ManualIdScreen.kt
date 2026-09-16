package com.luopingtech.ebike.rider.ui.ride

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardCapitalization
import androidx.compose.ui.unit.dp
import com.luopingtech.ebike.rider.RiderApp
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.ui.theme.RiderTheme

/**
 * 手输车牌号（对应 UniApp `pages-sub/ride/enter-id`）。二维码磨花是最常见的取车失败原因，
 * 这条兜底路径不能少。
 */
@Composable
fun ManualIdScreen(
    app: RiderApp,
    onSubmit: (carId: String) -> Unit,
    onBack: () -> Unit,
    modifier: Modifier = Modifier,
) {
    var input by remember { mutableStateOf("") }

    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)

    // 车牌是数字 + 大写字母，全角空格和小写在扫码贴纸上不存在。
    val normalized = input.trim().uppercase()
    val valid = normalized.length in MIN_LENGTH..MAX_LENGTH && normalized.all { it.isLetterOrDigit() }

    fun submit() {
        if (valid) onSubmit(normalized)
    }

    RideScaffold(
        title = t(Str.ManualIdTitle),
        onBack = onBack,
        backLabel = t(Str.Back),
        modifier = modifier,
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = 20.dp),
        ) {
            Text(
                text = t(Str.ManualIdHint),
                style = MaterialTheme.typography.bodyMedium,
                color = RiderTheme.colors.textSecondary,
                modifier = Modifier.padding(vertical = 12.dp),
            )
            OutlinedTextField(
                value = input,
                onValueChange = { input = it.filter { ch -> ch.isLetterOrDigit() }.take(MAX_LENGTH) },
                singleLine = true,
                isError = input.isNotBlank() && !valid,
                supportingText = {
                    if (input.isNotBlank() && !valid) Text(t(Str.ManualIdInvalid))
                },
                keyboardOptions = KeyboardOptions(
                    capitalization = KeyboardCapitalization.Characters,
                    imeAction = ImeAction.Done,
                ),
                keyboardActions = KeyboardActions(onDone = { submit() }),
                modifier = Modifier.fillMaxWidth(),
            )
            Button(
                onClick = { submit() },
                enabled = valid,
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 16.dp),
            ) {
                Text(t(Str.ManualIdSubmit))
            }
        }
    }
}

private const val MIN_LENGTH = 4
private const val MAX_LENGTH = 20
