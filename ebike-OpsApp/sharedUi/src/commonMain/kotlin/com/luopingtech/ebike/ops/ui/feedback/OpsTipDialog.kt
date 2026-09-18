package com.luopingtech.ebike.ops.ui.feedback

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.zIndex

/**
 * Legacy Kongzue WaitDialog / TipDialog 风格：居中深色浮层，无按钮。
 * - Waiting：转圈 + 文案（操作车辆中）
 * - Success：对勾 + 文案（约 1s）
 * - Error：叉号 + 文案（约 2s）
 */
sealed class OpsTipState {
    data object Hidden : OpsTipState()
    data class Waiting(val message: String) : OpsTipState()
    data class Success(val message: String) : OpsTipState()
    data class Error(val message: String) : OpsTipState()
}

@Composable
fun OpsTipDialogHost(
    state: OpsTipState,
    modifier: Modifier = Modifier,
) {
    if (state is OpsTipState.Hidden) return
    val (icon, message, tint) = when (state) {
        is OpsTipState.Waiting -> Triple(null, state.message, Color.White)
        is OpsTipState.Success -> Triple("✓", state.message, Color(0xFF4CD964))
        is OpsTipState.Error -> Triple("✕", state.message, Color(0xFFFF3B30))
        OpsTipState.Hidden -> return
    }
    Box(
        modifier = modifier
            .fillMaxSize()
            .zIndex(100f),
        contentAlignment = Alignment.Center,
    ) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            modifier = Modifier
                .widthIn(min = 120.dp, max = 220.dp)
                .background(Color(0xCC242936), RoundedCornerShape(12.dp))
                .padding(horizontal = 20.dp, vertical = 22.dp),
        ) {
            when {
                state is OpsTipState.Waiting -> {
                    CircularProgressIndicator(
                        modifier = Modifier.size(36.dp),
                        color = Color.White,
                        strokeWidth = 3.dp,
                    )
                }
                icon != null -> {
                    Text(
                        text = icon,
                        color = tint,
                        fontSize = 32.sp,
                        fontWeight = FontWeight.Bold,
                    )
                }
            }
            Spacer(modifier = Modifier.height(12.dp))
            Text(
                text = message,
                color = Color.White,
                fontSize = 15.sp,
                textAlign = TextAlign.Center,
                fontWeight = FontWeight.Medium,
            )
        }
    }
}
