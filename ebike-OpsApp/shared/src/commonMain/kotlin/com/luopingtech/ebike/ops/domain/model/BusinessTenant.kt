package com.luopingtech.ebike.ops.domain.model

/**
 * Operator / 分部 from `/business/tenant/queryList`.
 * Top-level `secret` is stored separately for `phone_secret` exchange.
 */
data class BusinessTenant(
    val tenantId: String,
    val tenantName: String = "",
    val companyName: String = "",
    val aliasName: String = "",
) {
    val displayLabel: String
        get() = tenantName.ifBlank { companyName }.ifBlank { aliasName }.ifBlank { tenantId }
}

data class BusinessTenantList(
    val tenants: List<BusinessTenant> = emptyList(),
    /** Shared secret for phone_secret oauth — not per-item. */
    val secret: String = "",
)
