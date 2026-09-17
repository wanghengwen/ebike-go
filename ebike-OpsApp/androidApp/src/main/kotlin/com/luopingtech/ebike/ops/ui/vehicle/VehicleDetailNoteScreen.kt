package com.luopingtech.ebike.ops.ui.vehicle

import androidx.activity.compose.BackHandler
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.ui.theme.OpsTheme

/**
 * Legacy VehicleDetailNoteActivity — map marker / track legend.
 */
@Composable
fun VehicleDetailNoteScreen(
    app: OpsApp,
    onClose: () -> Unit,
    modifier: Modifier = Modifier,
) {
    BackHandler(onBack = onClose)
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)

    val trackNotes = listOf(
        "车辆定位" to "显示该车辆的实时定位位置，会存在少量定位偏差。",
        "最近行程轨迹" to "骑行中则显示骑行中轨迹；非骑行中则显示上一单骑行轨迹。",
        "车的行程起点" to "根据车辆定位，显示骑行轨迹的起点。",
        "车的行程终点" to "根据车辆定位，显示骑行轨迹的终点。",
        "用车人的起点" to "显示最近一单，用车时的手机 GPS 定位。",
        "还车人的终点" to "显示最近一单，还车时的手机 GPS 定位。",
    )
    val markerNotes = listOf(
        "车辆下架", "临停", "骑行中", "异常离线", "电瓶移除",
        "异常移动", "换电中", "维修", "拖车", "丢失",
        "一日无单", "超长订单", "超区", "禁停", "低电量 SOC",
        "被预约", "短时订单", "调度中", "占用",
    )

    Surface(modifier = modifier.fillMaxSize(), color = Color.White) {
        Column(modifier = Modifier.fillMaxSize()) {
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(OpsTheme.colors.primary)
                    .statusBarsPadding()
                    .height(48.dp),
            ) {
                Text(
                    text = "‹",
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
                    text = t(Str.FunctionDescription),
                    color = Color.White,
                    fontSize = 18.sp,
                    fontWeight = FontWeight.Medium,
                    modifier = Modifier.align(Alignment.Center),
                )
            }
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .verticalScroll(rememberScrollState())
                    .padding(bottom = 24.dp),
            ) {
                trackNotes.forEach { (title, tip) ->
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 20.dp, vertical = 12.dp),
                    ) {
                        Text(title, fontWeight = FontWeight.Bold, fontSize = 15.sp, color = Color(0xFF242936))
                        Text(
                            tip,
                            fontSize = 13.sp,
                            color = Color(0xFF7C87B1),
                            modifier = Modifier.padding(top = 4.dp),
                        )
                    }
                    HorizontalDivider(color = Color(0xFFEEF0F6))
                }
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(12.dp)
                        .background(Color(0xFFF4F6FF)),
                )
                Text(
                    text = t(Str.ScooterMarkAnnotation),
                    fontSize = 18.sp,
                    fontWeight = FontWeight.Bold,
                    color = Color.Black,
                    modifier = Modifier.padding(start = 20.dp, top = 16.dp, bottom = 6.dp),
                )
                markerNotes.forEachIndexed { index, title ->
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 20.dp, vertical = 10.dp),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Box(
                                modifier = Modifier
                                    .width(28.dp)
                                    .height(28.dp)
                                    .background(OpsTheme.colors.primary.copy(alpha = 0.15f)),
                                contentAlignment = Alignment.Center,
                            ) {
                                Text("${index + 1}", fontSize = 12.sp, color = OpsTheme.colors.primary)
                            }
                            Text(
                                title,
                                fontSize = 14.sp,
                                color = Color(0xFF333333),
                                modifier = Modifier.padding(start = 12.dp),
                            )
                        }
                        Text(
                            t(Str.PriorityDisplay, "${index + 1}"),
                            fontSize = 12.sp,
                            color = Color(0xFF999999),
                        )
                    }
                    HorizontalDivider(color = Color(0xFFEEF0F6))
                }
            }
        }
    }
}
