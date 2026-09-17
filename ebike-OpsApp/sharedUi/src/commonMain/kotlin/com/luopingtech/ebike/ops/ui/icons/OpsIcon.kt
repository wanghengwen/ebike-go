package com.luopingtech.ebike.ops.ui.icons

/**
 * 跨端图标键。commonMain 里不许碰 `R.drawable` / `@DrawableRes`；
 * 两端的 actual 各自把它映射到同一套 drawable 文件名：Android 走 `res/drawable-xxhdpi`，
 * iOS 走 app bundle 里的同名位图（见 iosApp/OpsAppHost/Media）。
 */
enum class OpsIcon {
    // —— 工作台脚手架自用 ——
    TenantLogo,
    ArrowDown,
    Setting,
    MineModuleAddGray,
    MineModuleEditAdd,
    MineModuleEditDelete,

    // —— 任务 / 分析脚手架自用 ——
    BgIconTaskCenter,
    BgAnalysis,

    // —— 工作台模块 ——
    VehicleList,
    ReplaceBattery,
    MoveVehicle,
    Repair,
    UnlockedVehicle,
    BluetoothRadar,
    OperationSetting,
    MyTask,
    VehicleTag,
    Relocation,
    ParkingArea,
    OrderQuery,
    OperationScreen,
    RevenueScreen,
    EmployeeManager,
    ProfessionAudit,
    ObjectionOrder,
    BlackList,
    IdAudit,
    OperationLog,
    VehicleInspection,
    CenterControlBind,
    PutPullShelves,
    InWarehouse,
    OutWarehouse,
    WarehouseRecord,

    // —— 分析卡片 ——
    OfflineOperation,
    AnalysisStation,
    AnalysisReturnBike,
    AnalysisVehicleDistribution,

    // —— 扫码屏 ——
    ScanManual,
    TorchOn,
    TorchOff,

    // —— 任务中心卡片 ——
    TaskChangeBattery,
    TaskMoveBike,
    TaskInspection,
    TaskRepair,

    // —— 首页右侧筛选 ——
    FilterDrawerOpen,
    FilterDrawerClose,

    // —— 报修 / 标记 ——
    VehicleRepairV1,
    VehicleRepairV2,
    VehicleRepairV3,
    PhotoReplaceHolder,
    PhotoDelete,
    VehicleAddIc,

    // —— 通用导航 / 切换 ——
    ChevronLeft,
    CommonSwitch,

    // —— 围栏浏览 / 地图侧栏 ——
    MapMore,
    MapMoreSelected,
    MapLocation,
    MapLocationSelected,
    MapExplain,
    MapSatellite,
    MapSatelliteSelected,
    IconParking,
    IconNoParking,
    ArrowDownBlack,
    ArrowUpBlue,
    CommonAble,
    CommonUnable,
    CommonDelete,
    BtnClose,
    FenceUndoAble,
    FenceUndoDisable,
    SelectMapPoint,
    MapCenterPoint,
    FenceAngleAble,
    FenceAngleDisable,
    FenceEditParams,
    FenceModifySize,
    FenceTabPatch,
    FenceTabPoint,

    // —— 首页地图工具 ——
    HomeRefresh,
    HomeDetail,
    HomeDetailSelected,
    HomeFence,
    HomeFenceSelected,
    HomeSwitch,
    HomeSwitchSelected,

    // —— 底部 Tab ——
    TabMap,
    TabMapSelected,
    TabTask,
    TabTaskSelected,
    TabScan,
    TabAnalysis,
    TabAnalysisSelected,
    TabMine,
    TabMineSelected,

    // —— 车辆检测（legacy VehicleDetect）——
    DetectScan,
    DetectSwitchAcc,
    DetectSwitchDefend,
    DetectSwitchBattery,
    DetectSwitchHelmet,
    DetectSwitchWheel,
    DetectLocation,
    DetectOverload,
    DetectArrow,
    DetectRing,
    DetectRefresh,
    DetectReboot,
    BindClear,
}
