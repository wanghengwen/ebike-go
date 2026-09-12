package com.luopingtech.ebike.ops.ui.icons

import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.painter.ColorPainter
import androidx.compose.ui.graphics.painter.Painter

/** iOS 真图待补；占位色块保证 commonMain UI 能编过。 */
@Composable
actual fun painterResource(icon: OpsIcon): Painter = ColorPainter(Color(0xFFCCCCCC))
