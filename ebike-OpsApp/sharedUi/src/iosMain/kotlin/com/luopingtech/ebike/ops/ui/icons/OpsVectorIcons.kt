package com.luopingtech.ebike.ops.ui.icons

import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.graphics.vector.PathParser
import androidx.compose.ui.unit.dp

/**
 * 扫码页那三个图标 Android 侧是 vector XML（`res/drawable/ic_*.xml`），
 * iOS 没有 vector drawable 的概念，这里用同一份 pathData 重建 [ImageVector]。
 *
 * path 字符串是从 XML 里原样搬过来的：改图形要两边一起改，否则同一个按钮
 * 在两端长得不一样。
 */
internal object OpsVectorIcons {

    fun of(icon: OpsIcon): ImageVector? = when (icon) {
        OpsIcon.ScanManual -> scanManual
        OpsIcon.TorchOn -> torchOn
        OpsIcon.TorchOff -> torchOff
        else -> null
    }

    private val LINE = Color(0xFFCCCCCC)
    private val DOT = Color(0xFF242936)

    private val scanManual: ImageVector by lazy {
        vector(
            name = "ic_scan_manual",
            paths = listOf(
                "M3,17.25V21h3.75L17.81,9.94l-3.75,-3.75L3,17.25zM20.71,7.04c0.39,-0.39 " +
                    "0.39,-1.02 0,-1.41l-2.34,-2.34c-0.39,-0.39 -1.02,-0.39 -1.41,0l-1.83,1.83 " +
                    "3.75,3.75 1.83,-1.83z" to LINE,
            ),
        )
    }

    private val torchOn: ImageVector by lazy {
        vector(
            name = "ic_torch_on",
            paths = listOf(
                "M11,0h2v2h-2z" to LINE,
                "M16.2,0.6l1.4,1.4 -1.4,1.4 -1.4,-1.4z" to LINE,
                "M7.8,0.6l1.4,1.4 -1.4,1.4 -1.4,-1.4z" to LINE,
                "M9,4h6v2H9z" to LINE,
                "M8,6h8l-1.2,3.5H9.2L8,6z" to LINE,
                "M9.5,9.5h5V21c0,0.55 -0.45,1 -1,1h-3c-0.55,0 -1,-0.45 -1,-1V9.5z" to LINE,
                "M12,14m-1.2,0a1.2,1.2 0,1 1,2.4,0a1.2,1.2 0,1 1,-2.4,0" to DOT,
            ),
        )
    }

    private val torchOff: ImageVector by lazy {
        vector(
            name = "ic_torch_off",
            paths = listOf(
                "M9,2h6v2H9z" to LINE,
                "M8,4h8l-1.2,3.5H9.2L8,4z" to LINE,
                "M9.5,7.5h5V21c0,0.55 -0.45,1 -1,1h-3c-0.55,0 -1,-0.45 -1,-1V7.5z" to LINE,
                "M12,12m-1.2,0a1.2,1.2 0,1 1,2.4,0a1.2,1.2 0,1 1,-2.4,0" to DOT,
            ),
        )
    }

    private fun vector(name: String, paths: List<Pair<String, Color>>): ImageVector {
        val builder = ImageVector.Builder(
            name = name,
            defaultWidth = 48.dp,
            defaultHeight = 48.dp,
            viewportWidth = 24f,
            viewportHeight = 24f,
        )
        paths.forEach { (data, color) ->
            builder.addPath(
                pathData = PathParser().parsePathString(data).toNodes(),
                fill = SolidColor(color),
            )
        }
        return builder.build()
    }
}
