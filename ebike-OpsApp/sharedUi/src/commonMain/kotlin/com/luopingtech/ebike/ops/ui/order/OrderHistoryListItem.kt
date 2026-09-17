package com.luopingtech.ebike.ops.ui.order

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.domain.order.OrderFormat
import com.luopingtech.ebike.ops.domain.order.OrderRecord

private val TextPrimary = Color(0xFF242936)
private val PhoneBlue = Color(0xFF295FCC)
private val PayBlue = Color(0xFF295FCC)
private val PayGreen = Color(0xFF1DBA4F)
private val PayOrange = Color(0xFFF2A626)
private val OriginGreen = Color(0xFF00B68A)
private val EndPink = Color(0xFFEF2E6E)
private val SelectedBg = Color(0xFFF4F6FF)
private val DividerColor = Color(0xFFD9DCE6)

/**
 * 对齐遗留 item_order_history.xml 列表项。
 */
@Composable
fun OrderHistoryListItem(
    order: OrderRecord,
    t: (Str, Array<out Any?>) -> String,
    selected: Boolean,
    startAddress: String,
    endAddress: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val ridingLike = OrderFormat.isRidingLike(order.carState)
    val (payText, payTone) = OrderFormat.legacyPayBanner(order)
    val payColor = when (payTone) {
        OrderFormat.PayTone.Blue -> PayBlue
        OrderFormat.PayTone.Green -> PayGreen
        OrderFormat.PayTone.Orange -> PayOrange
    }
    val durationText = if (ridingLike) "--" else OrderFormat.duration(order.ridingTimeRaw)
    val mileText = if (ridingLike) "--" else OrderFormat.distance(order.mile)

    Column(
        modifier = modifier
            .fillMaxWidth()
            .background(if (selected) SelectedBg else Color.White)
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            )
            .padding(top = 12.dp),
    ) {
        // 行1：订单编号 …… 支付状态
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(t(Str.OrderIdColon, emptyArray()), color = TextPrimary, fontSize = 14.sp)
            Spacer(Modifier.width(8.dp))
            Text(
                OrderFormat.displayOrderId(order),
                color = TextPrimary,
                fontSize = 14.sp,
                fontWeight = FontWeight.Bold,
                modifier = Modifier.weight(1f),
            )
            Text(payText, color = payColor, fontSize = 14.sp, fontWeight = FontWeight.Bold)
        }

        Spacer(Modifier.height(8.dp))

        // 行2：车辆编号 …… 时间
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(t(Str.VehicleCarNumberColon, emptyArray()), color = TextPrimary, fontSize = 14.sp)
            Spacer(Modifier.width(8.dp))
            Text(
                OrderFormat.orDash(order.carId),
                color = TextPrimary,
                fontSize = 14.sp,
                fontWeight = FontWeight.Bold,
                modifier = Modifier.weight(1f),
            )
            Text(t(Str.OrderDurationColon, emptyArray()), color = TextPrimary, fontSize = 14.sp)
            Text(durationText, color = TextPrimary, fontSize = 14.sp, fontWeight = FontWeight.Bold)
        }

        Spacer(Modifier.height(8.dp))

        // 行3：用户信息 …… 里程
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(t(Str.UserInfoColon, emptyArray()), color = TextPrimary, fontSize = 14.sp)
            Spacer(Modifier.width(8.dp))
            Text(
                OrderFormat.orDash(order.phone),
                color = PhoneBlue,
                fontSize = 14.sp,
                fontWeight = FontWeight.Bold,
                modifier = Modifier.weight(1f),
            )
            Text(t(Str.MileageColon, emptyArray()), color = TextPrimary, fontSize = 14.sp)
            Text(mileText, color = TextPrimary, fontSize = 14.sp, fontWeight = FontWeight.Bold)
        }

        Spacer(Modifier.height(20.dp))
        TripPointBlock(
            tag = t(Str.OriginSingle, emptyArray()),
            tagColor = OriginGreen,
            time = OrderFormat.orDash(order.startTime),
            address = startAddress,
        )
        Spacer(Modifier.height(20.dp))
        TripPointBlock(
            tag = t(Str.EndSingle, emptyArray()),
            tagColor = EndPink,
            time = OrderFormat.orDash(order.endTime),
            address = endAddress,
        )
        Spacer(Modifier.height(20.dp))
        HorizontalDivider(color = DividerColor, thickness = 1.dp)
    }
}

@Composable
private fun TripPointBlock(
    tag: String,
    tagColor: Color,
    time: String,
    address: String,
) {
    Column(modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp)) {
        Row(verticalAlignment = Alignment.CenterVertically) {
            Box(
                modifier = Modifier
                    .background(tagColor, RoundedCornerShape(4.dp))
                    .padding(horizontal = 4.dp, vertical = 4.dp)
                    .widthIn(min = 20.dp),
                contentAlignment = Alignment.Center,
            ) {
                Text(tag, color = Color.White, fontSize = 12.sp)
            }
            Spacer(Modifier.width(8.dp))
            Text(
                text = time,
                color = TextPrimary,
                fontSize = 14.sp,
                fontWeight = FontWeight.Bold,
            )
        }
        Spacer(Modifier.height(8.dp))
        Text(
            text = address,
            color = TextPrimary,
            fontSize = 14.sp,
            fontWeight = FontWeight.Bold,
            modifier = Modifier.padding(start = 36.dp),
        )
    }
}
