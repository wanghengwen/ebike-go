package com.luopingtech.ebike.ops.ui.icons

/**
 * 跨端图标键。commonMain 里不许碰 `R.drawable` / `@DrawableRes`；
 * Android 在 actual 里映射到资源，iOS 先给占位色块，有 Mac 再补真图。
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

    // —— 任务中心卡片 ——
    TaskChangeBattery,
    TaskMoveBike,
    TaskInspection,
    TaskRepair,
}
