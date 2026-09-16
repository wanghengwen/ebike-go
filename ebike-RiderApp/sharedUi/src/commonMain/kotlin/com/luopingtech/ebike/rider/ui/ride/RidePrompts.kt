package com.luopingtech.ebike.rider.ui.ride

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.luopingtech.ebike.rider.core.i18n.RiderI18n
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.domain.riding.ReturnDecision
import com.luopingtech.ebike.rider.domain.riding.RideFormat
import com.luopingtech.ebike.rider.feature.riding.RidePrompt
import com.luopingtech.ebike.rider.ui.theme.RiderTheme

/**
 * 骑行页的弹层矩阵，一处集中处理 [RidePrompt] 的每个分支。
 *
 * 「可关 / 不可关」是有讲究的：超载断电属于阻塞态，随手关掉只会让用户对着不动的车
 * 反复点还车；其余都允许放弃（放弃还车时 [onDismiss] 会把相位收回骑行）。
 */
@Composable
internal fun RidePromptHost(
    i18n: RiderI18n,
    prompt: RidePrompt,
    onDismiss: () -> Unit,
    onConfirmReturn: (forcePenalty: Boolean) -> Unit,
    onOpenGuide: (pageType: Int, returnTypeCode: Int) -> Unit,
    onFindParking: () -> Unit,
    onRecoverPower: () -> Unit,
    onRecharge: () -> Unit,
    onRefreshReturn: () -> Unit,
) {
    fun t(key: Str, vararg args: Any?) = i18n.t(key, *args)

    when (prompt) {
        RidePrompt.Civilization -> RidePromptDialog(
            title = t(Str.ReturnCivilizationTitle),
            body = t(Str.ReturnCivilizationBody),
            confirmLabel = t(Str.RideReturn),
            onConfirm = { onConfirmReturn(false) },
            onDismiss = onDismiss,
            dismissLabel = t(Str.Cancel),
        )

        is RidePrompt.Penalty -> RidePromptDialog(
            title = t(Str.ReturnPenaltyTitle),
            body = t(Str.ReturnPenaltyBody, RideFormat.yuan(prompt.penaltyFen)),
            detail = t(ReturnDecision.meta(prompt.returnTypeCode).reason),
            // 认罚 = returnByNet 的 returnType=1。
            confirmLabel = t(Str.ReturnPayDispatch),
            onConfirm = { onConfirmReturn(true) },
            onDismiss = onDismiss,
            dismissLabel = t(Str.ReturnGoNearPark),
            onDismissAction = onFindParking,
        )

        is RidePrompt.Blocked -> {
            val meta = ReturnDecision.meta(prompt.returnTypeCode)
            RidePromptDialog(
                title = meta.title?.let { t(it) } ?: t(Str.ReturnBlockedTitle),
                body = t(meta.reason),
                detail = meta.tips?.let { t(it) },
                confirmLabel = t(Str.ReturnGoNearPark),
                onConfirm = onFindParking,
                onDismiss = onDismiss,
                // 满桩只能换地方，没有「刷新定位再试」这个出路。
                dismissLabel = if (meta.dialog) t(Str.Close) else t(Str.ReturnRefreshLocation),
                onDismissAction = if (meta.dialog) onDismiss else onRefreshReturn,
            )
        }

        is RidePrompt.Guide -> RidePromptDialog(
            title = t(Str.ReturnGuideTitle),
            body = t(ReturnDecision.meta(prompt.returnTypeCode).reason),
            confirmLabel = t(Str.Confirm),
            onConfirm = { onOpenGuide(prompt.pageType, prompt.returnTypeCode) },
            onDismiss = onDismiss,
            dismissLabel = t(Str.Cancel),
        )

        RidePrompt.HelmetNotReturned -> RidePromptDialog(
            title = t(Str.RideHelmet),
            body = t(Str.RideHelmetNotReturned),
            confirmLabel = t(Str.Confirm),
            onConfirm = onDismiss,
            onDismiss = onDismiss,
        )

        is RidePrompt.Overload -> RidePromptDialog(
            title = t(Str.RideOverloadTitle),
            body = if (prompt.powerOff) t(Str.RideOverloadPowerOff) else t(Str.RideOverloadWarn),
            confirmLabel = t(Str.Confirm),
            onConfirm = onDismiss,
            // 已断电时不给关：关掉也骑不动，只会让用户以为是 APP 卡了。
            onDismiss = if (prompt.powerOff) null else onDismiss,
        )

        is RidePrompt.RecoverPower -> RidePromptDialog(
            title = t(Str.RideRecoverPower),
            body = t(Str.RideRecoverPowerConfirm, prompt.minutes),
            confirmLabel = t(Str.RideRecoverPower),
            onConfirm = onRecoverPower,
            onDismiss = onDismiss,
            dismissLabel = t(Str.Cancel),
        )

        is RidePrompt.InsufficientBalance -> RidePromptDialog(
            title = t(Str.RideBalanceInsufficient),
            body = prompt.message.ifBlank { t(Str.RideBalanceInsufficient) },
            confirmLabel = t(Str.RideGoRecharge),
            onConfirm = onRecharge,
            onDismiss = onDismiss,
            dismissLabel = t(Str.Cancel),
        )
    }
}

/**
 * [onDismiss] 为 null 表示阻塞态：点外部区域和返回键都不生效，只留主按钮。
 * [onDismissAction] 让次要按钮做别的事（比如「去找 P 点」）而不是单纯关闭。
 */
@Composable
private fun RidePromptDialog(
    title: String,
    body: String,
    confirmLabel: String,
    onConfirm: () -> Unit,
    onDismiss: (() -> Unit)?,
    detail: String? = null,
    dismissLabel: String = "",
    onDismissAction: (() -> Unit)? = null,
) {
    AlertDialog(
        onDismissRequest = { onDismiss?.invoke() },
        title = { Text(title) },
        text = {
            Column(modifier = Modifier.fillMaxWidth()) {
                Text(
                    text = body,
                    style = MaterialTheme.typography.bodyMedium,
                    color = RiderTheme.colors.textSecondary,
                )
                detail?.takeIf { it.isNotBlank() }?.let {
                    Text(
                        text = it,
                        style = MaterialTheme.typography.bodySmall,
                        color = RiderTheme.colors.textTertiary,
                        modifier = Modifier.padding(top = 6.dp),
                    )
                }
            }
        },
        confirmButton = { TextButton(onClick = onConfirm) { Text(confirmLabel) } },
        dismissButton = {
            val action = onDismissAction ?: onDismiss
            if (action != null && dismissLabel.isNotBlank()) {
                TextButton(onClick = action) { Text(dismissLabel) }
            }
        },
    )
}
