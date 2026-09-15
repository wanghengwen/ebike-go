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
}
