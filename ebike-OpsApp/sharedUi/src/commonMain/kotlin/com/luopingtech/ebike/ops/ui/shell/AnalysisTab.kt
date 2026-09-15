package com.luopingtech.ebike.ops.ui.shell

import com.luopingtech.ebike.ops.OpsApp
import androidx.compose.runtime.Composable
import com.luopingtech.ebike.ops.ui.analysis.AnalysisCardItem
import com.luopingtech.ebike.ops.ui.analysis.AnalysisScaffold
import com.luopingtech.ebike.ops.ui.feedback.LocalOpsToast
import com.luopingtech.ebike.ops.ui.icons.OpsIcon
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.domain.permission.OpsPermissions
import com.luopingtech.ebike.ops.feature.home.HomeUiState

@Composable
internal fun AnalysisTab(
    app: OpsApp,
    homeState: HomeUiState,
    permissions: OpsPermissions,
    onChangeArea: () -> Unit,
    onOpenOfflineOps: () -> Unit,
    onOpenStationAnalysis: () -> Unit,
    onOpenReturnCarAnalysis: () -> Unit,
    onOpenVehicleDist: () -> Unit,
) {
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val toast = LocalOpsToast.current
    fun comingSoon() = toast(t(Str.FeatureComingSoon))
    val cards = buildList {
        fun addCard(visible: Boolean, id: String, title: String, icon: OpsIcon, onClick: () -> Unit) {
            if (visible || app.isDemoMode) {
                add(
                    AnalysisCardItem(
                        id = id,
                        title = title,
                        icon = icon,
                        onClick = onClick,
                    ),
                )
            }
        }
        addCard(permissions.showAnalysisOfflineOps, "offline", t(Str.OfflineOperation), OpsIcon.OfflineOperation, onOpenOfflineOps)
        addCard(permissions.showAnalysisStation, "station", t(Str.StationMonitor), OpsIcon.AnalysisStation, onOpenStationAnalysis)
        addCard(permissions.showAnalysisReturnCar, "return", t(Str.ReturnCarAnalysis), OpsIcon.AnalysisReturnBike, onOpenReturnCarAnalysis)
        addCard(permissions.showAnalysisVehicleDist, "vdist", t(Str.VehicleDistribution), OpsIcon.AnalysisVehicleDistribution, onOpenVehicleDist)
    }.distinctBy { it.id }

    AnalysisScaffold(
        areaName = homeState.currentArea?.name?.takeIf { it.isNotBlank() }
            ?: t(Str.SelectServiceArea),
        pageTitle = t(Str.TabAnalysis),
        cards = cards,
        onChangeArea = onChangeArea,
    )
}
