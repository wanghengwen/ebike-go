package com.luopingtech.ebike.ops.ui.vehicle

import androidx.activity.compose.BackHandler
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
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
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.domain.model.Vehicle
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale

/**
 * Legacy VehicleDetailInfoActivity — full device info list from「更多信息」.
 */
@Composable
fun VehicleMoreInfoScreen(
    app: OpsApp,
    vehicle: Vehicle,
    address: String?,
    onClose: () -> Unit,
    modifier: Modifier = Modifier,
) {
    BackHandler(onBack = onClose)
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)

    fun onOff(open: Boolean): String = if (open) t(Str.AccOpened) else t(Str.AccClosed)
    // Legacy: batteryLock/backWheelLock 0 = open(开启)
    fun lockOpen(code: Int?): String = onOff(code == 0)
    fun defendOn(code: Int?): String = onOff(code == 1)
    val helmetWear = when {
        vehicle.helmetBind == 0 -> t(Str.HelmetUnsupported)
        vehicle.helmetState == 0 -> t(Str.HelmetNotWorn)
        vehicle.helmetState == 1 -> t(Str.HelmetWorn)
        else -> t(Str.HelmetUnsupported)
    }
    val batteryText = "${vehicle.restBattery}%(${
        vehicle.voltageMv?.takeIf { it > 0 }?.let { "${it / 1000}v" } ?: "0v"
    })"
    val locateTime = formatLocateTime(vehicle.timestamp)

    val rows = listOf(
        t(Str.VehicleTagCarId) to vehicle.carId.ifBlank { "-" },
        t(Str.OrderQueryImeiLabel) to vehicle.imei.ifBlank { "-" },
        t(Str.DeviceVersion) to vehicle.version.ifBlank { "-" },
        t(Str.VehicleListColStatus) to vehicle.ridingLabel,
        t(Str.VehicleListColBattery) to batteryText,
        t(Str.AccStateLabel) to onOff(vehicle.acc == 1),
        t(Str.DefendState) to defendOn(vehicle.defend),
        t(Str.BackWheelLockState) to lockOpen(vehicle.backWheelLock),
        t(Str.BatteryBoxState) to lockOpen(vehicle.batteryLock),
        t(Str.EcuOnlineState) to if (vehicle.isOnline) t(Str.Online) else t(Str.Offline),
        t(Str.EcuSignalState) to (vehicle.gsmSignal?.toString() ?: "-"),
        t(Str.HelmetWearState) to helmetWear,
        t(Str.HelmetCode) to vehicle.helmetMac.ifBlank { "--" },
        t(Str.InertialNavState) to onOff(vehicle.headingAngle > 0),
        t(Str.LastLocatePlace) to (address?.ifBlank { null } ?: "-"),
        t(Str.LastLocateTime) to locateTime,
        t(Str.BelongStation) to vehicle.forParkName.ifBlank { "-" },
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
                    text = t(Str.VehicleInfoTitle),
                    color = Color.White,
                    fontSize = 18.sp,
                    fontWeight = FontWeight.Medium,
                    modifier = Modifier.align(Alignment.Center),
                )
            }
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .verticalScroll(rememberScrollState()),
            ) {
                rows.forEach { (label, value) ->
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 16.dp, vertical = 12.dp),
                        verticalAlignment = Alignment.Top,
                    ) {
                        Text(
                            text = label,
                            color = Color(0xFF242936),
                            fontSize = 14.sp,
                            modifier = Modifier.padding(end = 12.dp),
                        )
                        Text(
                            text = value,
                            color = Color(0xFF242936),
                            fontSize = 14.sp,
                            fontWeight = FontWeight.Bold,
                            textAlign = TextAlign.End,
                            modifier = Modifier.weight(1f),
                        )
                    }
                    HorizontalDivider(color = Color(0xFFD9DCE6), thickness = 0.5.dp)
                }
            }
        }
    }
}

private fun formatLocateTime(timestamp: Long?): String {
    val ms = timestamp?.takeIf { it > 0 } ?: return "-"
    val normalized = if (ms < 10_000_000_000L) ms * 1000L else ms
    return SimpleDateFormat("yyyy-MM-dd HH:mm:ss", Locale.getDefault()).format(Date(normalized))
}
