package com.luopingtech.ebike.ops.domain.model

data class BindVehicleInfo(
    val id: String = "",
    val carId: String = "",
    val carNo: String = "",
    val brand: String = "",
    val model: String = "",
    val imei: String = "",
    val helmet: String = "",
    val serviceId: String = "",
    val bindTime: String = "",
)

data class ShelfCheckResult(
    val carId: String,
    val carNo: String = "",
    val brand: String = "",
    val serviceId: String = "",
    val serviceName: String = "",
)

enum class DetectStepKind {
    Unlock,
    Lock,
    Ring,
    OpenBox,
    CloseBox,
}

data class DetectStepState(
    val kind: DetectStepKind,
    val label: String,
    val status: DetectStepStatus = DetectStepStatus.Idle,
    val detail: String = "",
)

enum class DetectStepStatus {
    Idle,
    Running,
    Ok,
    Failed,
}
