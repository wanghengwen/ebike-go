package com.luopingtech.ebike.ops.domain.model

/**
 * Field co-operator for free-move photo audit (legacy PhotographAudit 「协同人」).
 * Finish API expects [name]+[phone]; selection key is typically [phone].
 */
data class TeamWorker(
    val name: String,
    val phone: String,
) {
    val selectionKey: String
        get() = phone.ifBlank { name }

    val label: String
        get() = if (phone.isBlank()) name else "$name · $phone"
}
