package com.luopingtech.ebike.ops.ui.media

import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier

/** 显示本地拍照/相册路径（file:// / content:// / 绝对路径）。 */
@Composable
expect fun LocalPathImage(
    path: String,
    modifier: Modifier = Modifier,
    contentDescription: String? = null,
)
