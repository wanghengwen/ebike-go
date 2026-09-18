package com.luopingtech.ebike.ops.domain.model

/** 创建停车区（对齐 createParking 常用字段）。 */
data class FenceParkingCreate(
    val serviceId: String,
    val name: String,
    val points: List<GeoLatLng>,
    val centerLat: Double,
    val centerLng: Double,
    val maxParkingNumber: Int = 50,
    val izEnable: Boolean = true,
    val bufferDistance: Int = 10,
    val direction: Int = 0,
    val tbeacon: Boolean = false,
    val rfid: Boolean = false,
    val directional: Boolean = false,
    val kickstand: Boolean = false,
    val camera: Boolean = false,
    val izFullPileNoStop: Boolean = false,
    val openingHoursBegin: String = "00:00:00",
    val openingHoursEnd: String = "23:59:00",
)

/** 更新停车区（对齐 updateParking）。 */
data class FenceParkingUpdate(
    val id: String,
    val serviceId: String,
    val name: String,
    val points: List<GeoLatLng>,
    val centerLat: Double,
    val centerLng: Double,
    val maxParkingNumber: Int = 50,
    val izEnable: Boolean = true,
    val bufferDistance: Int = 10,
    val direction: Int = 0,
    val tbeacon: Boolean = false,
    val rfid: Boolean = false,
    val directional: Boolean = false,
    val kickstand: Boolean = false,
    val camera: Boolean = false,
    val izFullPileNoStop: Boolean = false,
    val openingHoursBegin: String = "00:00:00",
    val openingHoursEnd: String = "23:59:00",
)

/** 创建禁停区最小字段（对齐 createNoParking）。 */
data class FenceNoParkingCreate(
    val serviceId: String,
    val name: String,
    val points: List<GeoLatLng>,
    val centerLat: Double,
    val centerLng: Double,
)

/** 更新禁停区（对齐 updateNoParking 最小字段）。 */
data class FenceNoParkingUpdate(
    val id: String,
    val serviceId: String,
    val name: String,
    val points: List<GeoLatLng>,
    val centerLat: Double,
    val centerLng: Double,
)
