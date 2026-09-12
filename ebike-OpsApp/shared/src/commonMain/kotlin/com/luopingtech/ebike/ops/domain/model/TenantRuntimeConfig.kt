package com.luopingtech.ebike.ops.domain.model

/**
 * Runtime tenant branding / QR hosts from `/business/tenant/selectAppConfig`.
 * Distinct from build-time [com.luopingtech.ebike.ops.core.config.TenantConfig].
 */
data class TenantRuntimeConfig(
    val tenantId: String = "",
    val tenantName: String = "",
    val alias: String = "",
    val clientDomain: String = "",
    val businessDomain: String = "",
    /** Host fragments allowed in vehicle QR URLs (legacy `qrCodes` / `QR_CHECK`). */
    val qrHosts: List<String> = emptyList(),
) {
    val qrHostPattern: String
        get() = qrHosts.filter { it.isNotBlank() }.joinToString("|")
}
