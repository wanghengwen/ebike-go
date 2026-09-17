package com.luopingtech.ebike.ops.ui.home

import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.geometry.CornerRadius
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.domain.model.Vehicle
import com.luopingtech.ebike.ops.domain.vehicle.VehicleAlarmFilter
import com.luopingtech.ebike.ops.domain.vehicle.VehicleAlarmStates
import com.luopingtech.ebike.ops.domain.vehicle.VehicleOperationStates
import com.luopingtech.ebike.ops.domain.vehicle.VehicleRidingStates
import com.luopingtech.ebike.ops.ui.theme.OpsStatBlue

/** Legacy HomeVehiclePop action colors. */
private val RingAction = Color(0xFFF2A626)
private val UnlockAction = Color(0xFF04B78A)
private val LockAction = Color(0xFF3F51B5)
private val DetailAction = Color(0xFF828DB3)
private val LabelColor = Color(0xFF333333)
private val BatteryFillGreen = Color(0xFF1DBA4F)
private val BatteryFillOrange = Color(0xFFFFAD00)
private val BatteryFillRed = Color(0xFFE02020)
private val ChipOkBg = Color(0xFFD1F5D7)
private val ChipOkFg = Color(0xFF1DBA4F)
private val ChipWarnBg = Color(0x4DFFAD00)
private val ChipWarnFg = Color(0xFFFFAD00)
private val ChipErrorBg = Color(0x26E02020)
private val ChipErrorFg = Color(0xFFE02020)

enum class HomePopChipTone { Ok, Warn, Error }

/**
 * Legacy [HomeVehiclePop] / [FFMapEbikeAlertView]：地图点车顶部操作卡。
 * 车辆编号 / 设备号 / 电量 / 状态标签 + 响铃 / 开锁 / 关锁 / 详情。
 */
@Composable
fun HomeVehiclePopupCard(
    vehicle: Vehicle?,
    message: String?,
    detailOpen: Boolean,
    showUnlock: Boolean,
    showDetails: Boolean,
    carNumberLabel: String,
    deviceNoLabel: String,
    statusLabel: String,
    ringLabel: String,
    unlockLabel: String,
    lockLabel: String,
    detailLabel: String,
    pickHint: String,
    statusChips: List<Pair<HomePopChipTone, String>>,
    detailContent: @Composable (() -> Unit)? = null,
    onRing: () -> Unit,
    onUnlock: () -> Unit,
    onLock: () -> Unit,
    onDetails: () -> Unit,
    onCopyCarId: (() -> Unit)? = null,
    onCopyImei: (() -> Unit)? = null,
    cardModifier: Modifier = Modifier,
) {
    Column(
        modifier = cardModifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp)
            .shadow(6.dp, RoundedCornerShape(16.dp))
            .background(Color.White, RoundedCornerShape(16.dp))
            .padding(start = 20.dp, end = 16.dp, top = 18.dp, bottom = 16.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        if (vehicle == null) {
            Text(text = pickHint, color = LabelColor.copy(alpha = 0.7f), fontSize = 14.sp)
        } else {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.Top,
            ) {
                Column(modifier = Modifier.weight(1f)) {
                    HomePopInfoRow(
                        label = carNumberLabel,
                        value = vehicle.carId,
                        onClick = onCopyCarId,
                    )
                    Spacer(Modifier.height(6.dp))
                    HomePopInfoRow(
                        label = deviceNoLabel,
                        value = vehicle.imei.ifBlank { "-" },
                        onClick = onCopyImei,
                    )
                }
                HomePopBattery(
                    percent = vehicle.restBattery,
                    label = vehicle.batteryLabel,
                )
            }

            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Text(text = statusLabel, color = LabelColor, fontSize = 14.sp)
                Spacer(Modifier.width(4.dp))
                Row(
                    modifier = Modifier.weight(1f),
                    horizontalArrangement = Arrangement.spacedBy(6.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    statusChips.forEach { (tone, label) ->
                        val (bg, fg) = when (tone) {
                            HomePopChipTone.Ok -> ChipOkBg to ChipOkFg
                            HomePopChipTone.Warn -> ChipWarnBg to ChipWarnFg
                            HomePopChipTone.Error -> ChipErrorBg to ChipErrorFg
                        }
                        Box(
                            modifier = Modifier
                                .background(bg, RoundedCornerShape(12.dp))
                                .padding(horizontal = 10.dp, vertical = 2.dp),
                        ) {
                            Text(text = label, color = fg, fontSize = 12.sp)
                        }
                    }
                }
            }

            Spacer(Modifier.height(4.dp))
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceEvenly,
            ) {
                HomePopActionButton(
                    label = ringLabel,
                    kind = HomePopActionKind.Ring,
                    color = RingAction,
                    onClick = onRing,
                )
                if (showUnlock) {
                    HomePopActionButton(
                        label = unlockLabel,
                        kind = HomePopActionKind.Unlock,
                        color = UnlockAction,
                        onClick = onUnlock,
                    )
                    HomePopActionButton(
                        label = lockLabel,
                        kind = HomePopActionKind.Lock,
                        color = LockAction,
                        onClick = onLock,
                    )
                }
                if (showDetails) {
                    HomePopActionButton(
                        label = detailLabel,
                        kind = HomePopActionKind.Detail,
                        color = DetailAction,
                        onClick = onDetails,
                    )
                }
            }

            if (detailOpen && detailContent != null) {
                Spacer(Modifier.height(4.dp))
                detailContent()
            }
        }
        message?.let {
            Text(text = it, color = LabelColor.copy(alpha = 0.75f), fontSize = 12.sp)
        }
    }
}

@Composable
private fun HomePopInfoRow(label: String, value: String, onClick: (() -> Unit)? = null) {
    Row(
        verticalAlignment = Alignment.CenterVertically,
        modifier = if (onClick != null) {
            Modifier.clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            )
        } else {
            Modifier
        },
    ) {
        Text(text = label, color = LabelColor, fontSize = 14.sp)
        Text(
            text = value,
            color = OpsStatBlue,
            fontSize = 16.sp,
            fontWeight = FontWeight.Bold,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
            modifier = Modifier.widthIn(max = 200.dp),
        )
    }
}

@Composable
private fun HomePopBattery(percent: Int, label: String) {
    val fill = when {
        percent >= 60 -> BatteryFillGreen
        percent >= 30 -> BatteryFillOrange
        else -> BatteryFillRed
    }
    val clamped = percent.coerceIn(0, 100) / 100f
    Row(
        verticalAlignment = Alignment.Bottom,
        horizontalArrangement = Arrangement.spacedBy(4.dp),
    ) {
        Row(verticalAlignment = Alignment.CenterVertically) {
            Box(
                modifier = Modifier
                    .padding(bottom = 4.dp)
                    .width(28.dp)
                    .height(14.dp)
                    .clip(RoundedCornerShape(2.dp))
                    .background(Color(0xFFDDDDDD)),
            ) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth(clamped)
                        .height(14.dp)
                        .clip(RoundedCornerShape(2.dp))
                        .background(fill),
                )
            }
            Box(
                modifier = Modifier
                    .padding(bottom = 4.dp)
                    .width(3.dp)
                    .height(6.dp)
                    .background(Color(0xFFBBBBBB), RoundedCornerShape(1.dp)),
            )
        }
        Text(
            text = label,
            color = LabelColor,
            fontSize = 22.sp,
            fontWeight = FontWeight.Bold,
        )
    }
}

@Composable
private fun HomePopActionButton(
    label: String,
    kind: HomePopActionKind,
    color: Color,
    onClick: () -> Unit,
) {
    Column(
        modifier = Modifier
            .width(72.dp)
            .height(56.dp)
            .clip(RoundedCornerShape(8.dp))
            .background(color)
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            ),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
    ) {
        HomePopActionIcon(kind = kind)
        Spacer(Modifier.height(2.dp))
        Text(
            text = label,
            color = Color.White,
            fontSize = 12.sp,
            fontWeight = FontWeight.SemiBold,
        )
    }
}

private enum class HomePopActionKind { Ring, Unlock, Lock, Detail }

@Composable
private fun HomePopActionIcon(kind: HomePopActionKind) {
    Canvas(modifier = Modifier.size(18.dp)) {
        val stroke = Stroke(width = 1.6.dp.toPx())
        when (kind) {
            HomePopActionKind.Ring -> {
                drawArc(
                    color = Color.White,
                    startAngle = 200f,
                    sweepAngle = 140f,
                    useCenter = false,
                    topLeft = Offset(size.width * 0.22f, size.height * 0.12f),
                    size = Size(size.width * 0.56f, size.height * 0.62f),
                    style = stroke,
                )
                drawLine(
                    color = Color.White,
                    start = Offset(size.width * 0.22f, size.height * 0.62f),
                    end = Offset(size.width * 0.78f, size.height * 0.62f),
                    strokeWidth = 1.6.dp.toPx(),
                )
                drawCircle(
                    color = Color.White,
                    radius = 1.4.dp.toPx(),
                    center = Offset(size.width * 0.5f, size.height * 0.78f),
                )
            }
            HomePopActionKind.Unlock -> {
                drawRoundRect(
                    color = Color.White,
                    topLeft = Offset(size.width * 0.22f, size.height * 0.42f),
                    size = Size(size.width * 0.56f, size.height * 0.4f),
                    cornerRadius = CornerRadius(2.dp.toPx()),
                    style = stroke,
                )
                drawArc(
                    color = Color.White,
                    startAngle = 200f,
                    sweepAngle = 140f,
                    useCenter = false,
                    topLeft = Offset(size.width * 0.34f, size.height * 0.08f),
                    size = Size(size.width * 0.4f, size.height * 0.42f),
                    style = stroke,
                )
            }
            HomePopActionKind.Lock -> {
                drawRoundRect(
                    color = Color.White,
                    topLeft = Offset(size.width * 0.22f, size.height * 0.42f),
                    size = Size(size.width * 0.56f, size.height * 0.4f),
                    cornerRadius = CornerRadius(2.dp.toPx()),
                    style = stroke,
                )
                drawArc(
                    color = Color.White,
                    startAngle = 180f,
                    sweepAngle = 180f,
                    useCenter = false,
                    topLeft = Offset(size.width * 0.3f, size.height * 0.12f),
                    size = Size(size.width * 0.4f, size.height * 0.4f),
                    style = stroke,
                )
                drawLine(
                    color = Color.White,
                    start = Offset(size.width * 0.3f, size.height * 0.32f),
                    end = Offset(size.width * 0.3f, size.height * 0.42f),
                    strokeWidth = 1.6.dp.toPx(),
                )
                drawLine(
                    color = Color.White,
                    start = Offset(size.width * 0.7f, size.height * 0.32f),
                    end = Offset(size.width * 0.7f, size.height * 0.42f),
                    strokeWidth = 1.6.dp.toPx(),
                )
            }
            HomePopActionKind.Detail -> {
                drawCircle(
                    color = Color.White,
                    radius = size.minDimension * 0.42f,
                    style = stroke,
                )
                drawCircle(
                    color = Color.White,
                    radius = 1.2.dp.toPx(),
                    center = Offset(size.width * 0.5f, size.height * 0.32f),
                )
                drawLine(
                    color = Color.White,
                    start = Offset(size.width * 0.5f, size.height * 0.46f),
                    end = Offset(size.width * 0.5f, size.height * 0.72f),
                    strokeWidth = 1.8.dp.toPx(),
                )
            }
        }
    }
}

fun buildHomeVehicleStatusChips(
    vehicle: Vehicle,
    canUseLabel: String,
    helmetNotClosedLabel: String,
    offlineLabel: String,
    soldOutLabel: String,
    lowBatteryLabel: String,
    repairingLabel: String,
    movingLabel: String,
    ridingLabel: String,
    tempParkingLabel: String,
    bookingLabel: String,
): List<Pair<HomePopChipTone, String>> = buildList {
    // Legacy: helmetLock 0 = open → 头盔未关
    if (vehicle.helmetLock == 0) {
        add(HomePopChipTone.Error to helmetNotClosedLabel)
    }
    vehicle.alarmStates.forEach { code ->
        val label = VehicleAlarmFilter.entries.firstOrNull { it.code == code }?.label
            ?: when (code) {
                VehicleAlarmStates.OFFLINE -> offlineLabel
                else -> null
            }
        if (!label.isNullOrBlank()) {
            add(HomePopChipTone.Error to label)
        }
    }
    vehicle.operationStates.forEach { code ->
        when (code) {
            VehicleOperationStates.OFF -> add(HomePopChipTone.Error to soldOutLabel)
            VehicleOperationStates.LOW_BATTERY -> add(HomePopChipTone.Error to lowBatteryLabel)
            VehicleOperationStates.REPAIRING -> add(HomePopChipTone.Error to repairingLabel)
            VehicleOperationStates.MOVING_CAR -> add(HomePopChipTone.Warn to movingLabel)
        }
    }
    when (vehicle.ridingState) {
        VehicleRidingStates.RIDEABLE -> add(HomePopChipTone.Ok to canUseLabel)
        VehicleRidingStates.RIDING -> add(HomePopChipTone.Ok to ridingLabel)
        VehicleRidingStates.TEMP_PARKING -> add(HomePopChipTone.Warn to tempParkingLabel)
        VehicleRidingStates.RESERVE -> add(HomePopChipTone.Warn to bookingLabel)
        else -> if (vehicle.ridingLabel.isNotBlank()) {
            add(HomePopChipTone.Ok to vehicle.ridingLabel)
        }
    }
    if (!vehicle.isOnline && vehicle.alarmStates.none { it == VehicleAlarmStates.OFFLINE }) {
        add(HomePopChipTone.Error to offlineLabel)
    }
}.distinctBy { it.second }.take(4)
