package com.luopingtech.ebike.ops.domain.report

/**
 * Legacy [VehicleRepairPartEnum] body-part ids (left / right columns around the bike diagram).
 * API repairConfig/list items whose id matches these are shown as body parts; others go to「其他类型」.
 */
object RepairBodyParts {
    data class Part(val id: String, val titleZh: String)

    val left: List<Part> = listOf(
        Part("69139954125511966", "刹车"),
        Part("69139954125511968", "龙头"),
        Part("69139954125511970", "车筐"),
        Part("69139954125511972", "车灯"),
        Part("69139954125511974", "线路"),
        Part("69139954125511976", "前轮"),
        Part("69139954125511978", "踏板"),
        Part("69139954125511980", "脚撑"),
    )

    val right: List<Part> = listOf(
        Part("69139954125511967", "油门"),
        Part("69139954125511969", "车座"),
        Part("69139954125511971", "二维码"),
        Part("69139954125511973", "挡泥板"),
        Part("69139954125511975", "电机"),
        Part("69139954125511977", "后轮"),
        Part("69139954125511979", "加私锁"),
        Part("69139954125511981", "电池"),
    )

    val allIds: Set<String> = (left + right).map { it.id }.toSet()
}
