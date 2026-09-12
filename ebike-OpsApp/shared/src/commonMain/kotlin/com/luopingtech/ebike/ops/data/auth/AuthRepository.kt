package com.luopingtech.ebike.ops.data.auth

import com.luopingtech.ebike.ops.core.network.NetworkSession
import com.luopingtech.ebike.ops.core.network.RequestAuth
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.core.util.Ids
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.domain.auth.PasswordRules
import com.luopingtech.ebike.ops.domain.model.BusinessTenant
import com.luopingtech.ebike.ops.domain.model.TenantRuntimeConfig
import com.luopingtech.ebike.ops.domain.model.UserSession
import com.luopingtech.ebike.ops.domain.permission.OpsPermissions
import com.luopingtech.ebike.ops.platform.SecureStore
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

/**
 * Auth repository. Demo mode (blank api.baseUrl) accepts local credentials;
 * SMS / account `hq` surfaces a multi-business picker.
 */
interface AuthRepository {
    val sessionInvalidMessage: StateFlow<String?>

    suspend fun login(account: String, password: String): OpsResult<UserSession>
    suspend fun sendSmsCode(phone: String, scene: Int = SmsScene.LOGIN): OpsResult<Unit>
    suspend fun loginWithSms(phone: String, messageCode: String): OpsResult<UserSession>
    suspend fun selectBusiness(tenantId: String): OpsResult<UserSession>
    fun pendingBusinesses(): List<BusinessTenant>
    fun clearPendingBusiness()

    suspend fun forgetPassword(phone: String, code: String, newPassword: String): OpsResult<Unit>
    suspend fun setPassword(newPassword: String): OpsResult<UserSession>
    suspend fun updatePassword(oldPassword: String, newPassword: String): OpsResult<Unit>
    fun runtimeConfig(): TenantRuntimeConfig?
    fun needsSetPassword(): Boolean

    fun loginAreaCode(): CallingCode
    fun setLoginAreaCode(code: CallingCode)

    suspend fun refreshIfNeeded(): OpsResult<UserSession>
    suspend fun logout()
    fun currentSession(): UserSession?
    fun markSessionInvalid(message: String)
    fun consumeSessionInvalid()
}

class AuthRepositoryImpl(
    private val secureStore: SecureStore,
    private val demoMode: Boolean,
    private val authApi: AuthApi? = null,
    private val requestAuth: RequestAuth? = null,
) : AuthRepository {
    private var session: UserSession? = restoreSession()
    private var pending: PendingBusinessAuth? = null
    private var runtimeConfig: TenantRuntimeConfig? = restoreRuntimeConfig()
    private val _sessionInvalidMessage = MutableStateFlow<String?>(null)
    override val sessionInvalidMessage: StateFlow<String?> = _sessionInvalidMessage.asStateFlow()

    override fun runtimeConfig(): TenantRuntimeConfig? = runtimeConfig

    override fun needsSetPassword(): Boolean = session?.hasPassword == false

    override fun loginAreaCode(): CallingCode =
        CallingCodeCatalog.find(
            secureStore.getString(SecureStore.KEY_LOGIN_AREA_REGION),
            secureStore.getString(SecureStore.KEY_LOGIN_AREA_CODE),
        )

    override fun setLoginAreaCode(code: CallingCode) {
        secureStore.putString(SecureStore.KEY_LOGIN_AREA_CODE, code.dialCode)
        secureStore.putString(SecureStore.KEY_LOGIN_AREA_REGION, code.regionCode)
    }

    override suspend fun login(account: String, password: String): OpsResult<UserSession> {
        if (account.isBlank() || password.isBlank()) {
            return OpsResult.Err(OpsError.business("AUTH_INVALID", "account or password empty"))
        }
        prepareFreshLogin()
        if (demoMode || authApi == null) {
            if (account.equals("hq", ignoreCase = true)) {
                pending = PendingBusinessAuth(
                    phone = account,
                    businesses = DEMO_BUSINESSES,
                    businessSecret = DEMO_BUSINESS_SECRET,
                )
                return OpsResult.Err(selectBusinessError())
            }
            return saveDemoSession(account)
        }

        return when (val tokenResult = authApi.loginWithPassword(account, password)) {
            is OpsResult.Err -> tokenResult
            is OpsResult.Ok -> resolveBusinessAfterHqToken(tokenResult.value, fallbackAccount = account)
        }
    }

    override suspend fun sendSmsCode(phone: String, scene: Int): OpsResult<Unit> {
        if (phone.isBlank()) {
            return OpsResult.Err(OpsError.business("AUTH_INVALID", "phone empty"))
        }
        if (demoMode || authApi == null) {
            return OpsResult.Ok(Unit)
        }
        return authApi.sendSmsCode(phone, scene)
    }

    override suspend fun forgetPassword(
        phone: String,
        code: String,
        newPassword: String,
    ): OpsResult<Unit> {
        PasswordRules.validationError(newPassword)?.let {
            return OpsResult.Err(OpsError.business("AUTH_PWD", it))
        }
        if (phone.isBlank() || code.isBlank()) {
            return OpsResult.Err(OpsError.business("AUTH_INVALID", "phone or code empty"))
        }
        if (demoMode || authApi == null) {
            return if (code.trim().length >= 4) {
                OpsResult.Ok(Unit)
            } else {
                OpsResult.Err(OpsError.business("AUTH_SMS", "invalid sms code"))
            }
        }
        return authApi.forgetPassword(phone, code, newPassword.trim())
    }

    override suspend fun setPassword(newPassword: String): OpsResult<UserSession> {
        PasswordRules.validationError(newPassword)?.let {
            return OpsResult.Err(OpsError.business("AUTH_PWD", it))
        }
        val current = session
            ?: return OpsResult.Err(OpsError.unauthorized("not signed in"))
        if (demoMode || authApi == null) {
            val next = current.copy(hasPassword = true)
            persistTokens(next)
            session = next
            return OpsResult.Ok(withArea(next))
        }
        val networkSession = NetworkSession(accessToken = current.accessToken)
        return when (val result = authApi.setPassword(networkSession, newPassword.trim())) {
            is OpsResult.Err -> result
            is OpsResult.Ok -> {
                val next = current.copy(hasPassword = true)
                persistTokens(next)
                session = next
                OpsResult.Ok(withArea(next))
            }
        }
    }

    override suspend fun updatePassword(oldPassword: String, newPassword: String): OpsResult<Unit> {
        PasswordRules.validationError(newPassword)?.let {
            return OpsResult.Err(OpsError.business("AUTH_PWD", it))
        }
        if (oldPassword.isBlank()) {
            return OpsResult.Err(OpsError.business("AUTH_INVALID", "old password empty"))
        }
        val current = session
            ?: return OpsResult.Err(OpsError.unauthorized("not signed in"))
        if (demoMode || authApi == null) {
            return OpsResult.Ok(Unit)
        }
        val networkSession = NetworkSession(accessToken = current.accessToken)
        return authApi.updatePassword(networkSession, oldPassword, newPassword.trim())
    }

    override suspend fun loginWithSms(phone: String, messageCode: String): OpsResult<UserSession> {
        if (phone.isBlank() || messageCode.isBlank()) {
            return OpsResult.Err(OpsError.business("AUTH_INVALID", "phone or code empty"))
        }
        prepareFreshLogin()
        if (demoMode || authApi == null) {
            if (messageCode.trim().length < 4) {
                return OpsResult.Err(OpsError.business("AUTH_SMS", "invalid sms code"))
            }
            pending = PendingBusinessAuth(
                phone = phone,
                businesses = DEMO_BUSINESSES,
                businessSecret = DEMO_BUSINESS_SECRET,
            )
            return OpsResult.Err(selectBusinessError())
        }

        return when (val tokenResult = authApi.loginWithSms(phone, messageCode)) {
            is OpsResult.Err -> tokenResult
            is OpsResult.Ok -> resolveBusinessAfterHqToken(tokenResult.value, fallbackAccount = phone)
        }
    }

    override suspend fun selectBusiness(tenantId: String): OpsResult<UserSession> {
        val pendingAuth = pending
            ?: return OpsResult.Err(OpsError.business("AUTH_NO_PENDING", "no business list"))
        val selected = pendingAuth.businesses.firstOrNull { it.tenantId == tenantId }
            ?: return OpsResult.Err(OpsError.business("AUTH_BUSINESS", "unknown business"))
        if (pendingAuth.businessSecret.isBlank()) {
            return OpsResult.Err(OpsError.business("AUTH_SECRET", "business secret empty"))
        }

        _sessionInvalidMessage.value = null
        if (demoMode || authApi == null) {
            secureStore.putString(SecureStore.KEY_TENANT_ID, selected.tenantId)
            requestAuth?.oauthBasicTenantId = selected.tenantId
            requestAuth?.oauthBasicSecret = pendingAuth.businessSecret
            clearPendingBusiness()
            return saveDemoSession(
                account = selected.displayLabel.ifBlank { pendingAuth.phone },
                tenantIdOverride = selected.tenantId,
            )
        }

        requestAuth?.oauthBasicTenantId = selected.tenantId
        requestAuth?.oauthBasicSecret = pendingAuth.businessSecret
        secureStore.putString(SecureStore.KEY_TENANT_ID, selected.tenantId)

        return when (
            val tokenResult = authApi.loginWithPhoneSecret(pendingAuth.phone, pendingAuth.businessSecret)
        ) {
            is OpsResult.Err -> {
                requestAuth?.clearOauthBasicOverride()
                tokenResult
            }
            is OpsResult.Ok -> {
                clearPendingBusiness()
                finalizeFromToken(
                    token = tokenResult.value,
                    fallbackAccount = pendingAuth.phone,
                    tenantIdOverride = selected.tenantId,
                )
            }
        }
    }

    override fun pendingBusinesses(): List<BusinessTenant> = pending?.businesses.orEmpty()

    override fun clearPendingBusiness() {
        pending = null
    }

    override suspend fun refreshIfNeeded(): OpsResult<UserSession> {
        val current = session ?: return OpsResult.Err(OpsError.unauthorized())
        if (demoMode || authApi == null) return OpsResult.Ok(withArea(current))
        val refresh = current.refreshToken
        if (refresh.isBlank()) return OpsResult.Err(OpsError.unauthorized("missing refresh token"))

        return when (val result = authApi.refreshToken(refresh)) {
            is OpsResult.Err -> result
            is OpsResult.Ok -> {
                val token = result.value
                val next = current.copy(
                    accessToken = token.accessToken.ifBlank { current.accessToken },
                    refreshToken = token.refreshToken.ifBlank { current.refreshToken },
                    displayName = token.nickname.ifBlank { current.displayName },
                )
                persistTokens(next)
                session = next
                OpsResult.Ok(withArea(next))
            }
        }
    }

    override suspend fun logout() {
        val current = session
        if (!demoMode && authApi != null && current != null && current.accessToken.isNotBlank()) {
            authApi.logout(NetworkSession(accessToken = current.accessToken))
        }
        clearLocalSession()
    }

    override fun currentSession(): UserSession? = session?.let(::withArea)

    override fun markSessionInvalid(message: String) {
        clearLocalSession()
        _sessionInvalidMessage.value = message.ifBlank { "session expired" }
    }

    override fun consumeSessionInvalid() {
        _sessionInvalidMessage.value = null
    }


    /**
     * After HQ oauth token (legacy NewLoginViewModel):
     * 1. getUserByToken → izRoot / tenantId
     * 2. queryList (izFilterRootTenant=true) → secret (+ optional list)
     * 3. department (!izRoot): phone_secret with user's own tenantId (no picker)
     * 4. HQ (izRoot): picker if >1, auto if 1, error if 0
     */
    private suspend fun resolveBusinessAfterHqToken(
        token: OauthTokenDto,
        fallbackAccount: String,
    ): OpsResult<UserSession> {
        val access = token.accessToken.trim()
        if (access.isBlank()) {
            return OpsResult.Err(OpsError.business("AUTH_EMPTY_TOKEN", "accessToken empty"))
        }
        // Legacy saves oauth token before getUserByToken / queryList.
        persistHqOauthTokens(access, token.refreshToken.trim())
        val phone = token.phone.ifBlank { fallbackAccount }
        val networkSession = NetworkSession(accessToken = access)
        val user = when (val profile = authApi!!.fetchUserInfo(networkSession)) {
            is OpsResult.Err -> {
                clearLocalSession()
                return OpsResult.Err(
                    OpsError.business(
                        profile.error.code.ifBlank { "AUTH_USER" },
                        profile.error.message.ifBlank { Strings.t(Str.AuthUserInfoFailed) },
                    ),
                )
            }
            is OpsResult.Ok -> profile.value
        }
        return when (val tenants = authApi.listTenants(phone, networkSession)) {
            is OpsResult.Err -> {
                clearLocalSession()
                tenants
            }
            is OpsResult.Ok -> {
                val list = tenants.value.tenants.filter { it.tenantId.isNotBlank() }
                val secret = tenants.value.secret
                if (secret.isBlank()) {
                    clearLocalSession()
                    return OpsResult.Err(
                        OpsError.business("AUTH_SECRET", Strings.t(Str.AuthNoBusinessTenant)),
                    )
                }
                // Department / non-root: never show the full HQ tenant catalog.
                if (!user.izRoot) {
                    val ownId = user.tenantId.trim()
                    if (ownId.isBlank()) {
                        clearLocalSession()
                        return OpsResult.Err(
                            OpsError.business("AUTH_NO_TENANT", Strings.t(Str.AuthNoBusinessTenant)),
                        )
                    }
                    val own = list.firstOrNull { it.tenantId == ownId }
                        ?: BusinessTenant(
                            tenantId = ownId,
                            tenantName = user.tenantName.ifBlank { user.nickname },
                        )
                    pending = PendingBusinessAuth(
                        phone = phone,
                        businesses = listOf(own),
                        businessSecret = secret,
                    )
                    return selectBusiness(ownId)
                }
                when {
                    list.isEmpty() -> {
                        clearLocalSession()
                        OpsResult.Err(
                            OpsError.business("AUTH_NO_TENANT", Strings.t(Str.AuthNoBusinessTenant)),
                        )
                    }
                    list.size == 1 -> {
                        pending = PendingBusinessAuth(
                            phone = phone,
                            businesses = list,
                            businessSecret = secret,
                        )
                        selectBusiness(list.first().tenantId)
                    }
                    else -> {
                        pending = PendingBusinessAuth(
                            phone = phone,
                            businesses = list,
                            businessSecret = secret,
                        )
                        OpsResult.Err(selectBusinessError())
                    }
                }
            }
        }
    }

    /** Drop previous session so HQ oauth uses config tenant Basic + no stale Bearer. */
    private fun prepareFreshLogin() {
        _sessionInvalidMessage.value = null
        clearLocalSession()
    }

    private fun persistHqOauthTokens(accessToken: String, refreshToken: String) {
        secureStore.putString(SecureStore.KEY_ACCESS_TOKEN, accessToken)
        if (refreshToken.isNotBlank()) {
            secureStore.putString(SecureStore.KEY_REFRESH_TOKEN, refreshToken)
        }
    }

    private suspend fun finalizeFromToken(
        token: OauthTokenDto,
        fallbackAccount: String,
        tenantIdOverride: String? = null,
    ): OpsResult<UserSession> {
        if (token.accessToken.isBlank()) {
            return OpsResult.Err(OpsError.business("AUTH_EMPTY_TOKEN", "accessToken empty"))
        }
        var next = UserSession(
            userId = token.pin.ifBlank { token.phone },
            displayName = token.nickname.ifBlank { token.phone.ifBlank { fallbackAccount } },
            phone = token.phone.ifBlank { fallbackAccount },
            accessToken = token.accessToken,
            refreshToken = token.refreshToken,
            tenantId = tenantIdOverride
                ?.takeIf { it.isNotBlank() }
                ?: secureStore.getString(SecureStore.KEY_TENANT_ID).orEmpty(),
            serviceAreaId = secureStore.getString(SecureStore.KEY_SERVICE_AREA_ID).orEmpty(),
            serviceAreaName = secureStore.getString(SecureStore.KEY_SERVICE_AREA_NAME).orEmpty(),
        )
        persistTokens(next)

        val api = authApi ?: return OpsResult.Ok(withArea(next)).also { session = next }
        val networkSession = NetworkSession(
            accessToken = next.accessToken.trim(),
        )
        when (val profile = api.fetchUserInfo(networkSession)) {
            is OpsResult.Ok -> {
                val user = profile.value
                next = next.copy(
                    userId = user.id.ifBlank { user.userId.ifBlank { user.pin.ifBlank { next.userId } } },
                    displayName = user.nickname
                        .ifBlank { user.realName }
                        .ifBlank { user.userName }
                        .ifBlank { user.phone }
                        .ifBlank { next.displayName },
                    phone = user.phone.ifBlank { next.phone },
                    tenantId = tenantIdOverride?.takeIf { it.isNotBlank() }
                        ?: user.tenantId.ifBlank { next.tenantId },
                    serviceAreaId = user.serviceAreaId.ifBlank { next.serviceAreaId },
                    permissionCodes = user.codes.filter { it.isNotBlank() }.distinct(),
                    hasPassword = user.hasPassword,
                    roleName = user.roleName.trim(),
                )
                persistTokens(next)
            }
            is OpsResult.Err -> {
                clearLocalSession()
                return OpsResult.Err(
                    OpsError.business(
                        profile.error.code.ifBlank { "AUTH_USER" },
                        profile.error.message.ifBlank { Strings.t(Str.AuthUserInfoFailed) },
                    ),
                )
            }
        }
        when (val cfg = refreshRuntimeConfig(networkSession)) {
            is OpsResult.Err -> {
                clearLocalSession()
                return OpsResult.Err(
                    OpsError.business(
                        cfg.error.code.ifBlank { "AUTH_CONFIG" },
                        cfg.error.message.ifBlank { Strings.t(Str.AuthConfigFailed) },
                    ),
                )
            }
            is OpsResult.Ok -> Unit
        }
        session = next
        return OpsResult.Ok(withArea(next))
    }

    private fun clearLocalSession() {
        session = null
        runtimeConfig = null
        clearPendingBusiness()
        requestAuth?.clearOauthBasicOverride()
        secureStore.remove(SecureStore.KEY_ACCESS_TOKEN)
        secureStore.remove(SecureStore.KEY_REFRESH_TOKEN)
        // Fall back to config.tenantId (HQ) on next oauth Basic / body tenantId.
        secureStore.remove(SecureStore.KEY_TENANT_ID)
        secureStore.remove(KEY_DISPLAY_NAME)
        secureStore.remove(KEY_USER_ID)
        secureStore.remove(KEY_PHONE)
        secureStore.remove(KEY_PERMISSION_CODES)
        secureStore.remove(KEY_HAS_PASSWORD)
        secureStore.remove(KEY_ROLE_NAME)
        secureStore.remove(KEY_QR_HOSTS)
        secureStore.remove(KEY_RUNTIME_TENANT_NAME)
    }

    private fun saveDemoSession(
        account: String,
        tenantIdOverride: String? = null,
    ): OpsResult<UserSession> {
        val codes = OpsPermissions.forDemoAccount(account).rawCodes.toList()
        val next = UserSession(
            userId = if (account.equals("limited", ignoreCase = true)) "demo-limited" else "demo-user",
            displayName = account,
            phone = "13800138000",
            accessToken = "demo-token",
            refreshToken = "demo-refresh",
            tenantId = tenantIdOverride
                ?.takeIf { it.isNotBlank() }
                ?: secureStore.getString(SecureStore.KEY_TENANT_ID).orEmpty(),
            serviceAreaId = secureStore.getString(SecureStore.KEY_SERVICE_AREA_ID).orEmpty(),
            serviceAreaName = secureStore.getString(SecureStore.KEY_SERVICE_AREA_NAME).orEmpty(),
            permissionCodes = codes,
            hasPassword = !account.equals("nopwd", ignoreCase = true),
            roleName = "管理员",
        )
        persistTokens(next)
        runtimeConfig = DEMO_RUNTIME_CONFIG
        persistRuntimeConfig(runtimeConfig!!)
        session = next
        return OpsResult.Ok(withArea(next))
    }

    private fun withArea(session: UserSession): UserSession = session.copy(
        serviceAreaId = secureStore.getString(SecureStore.KEY_SERVICE_AREA_ID)
            ?.takeIf { it.isNotBlank() }
            ?: session.serviceAreaId,
        serviceAreaName = secureStore.getString(SecureStore.KEY_SERVICE_AREA_NAME)
            ?.takeIf { it.isNotBlank() }
            ?: session.serviceAreaName,
    )

    private fun persistTokens(session: UserSession) {
        secureStore.putString(SecureStore.KEY_ACCESS_TOKEN, session.accessToken)
        secureStore.putString(SecureStore.KEY_REFRESH_TOKEN, session.refreshToken)
        if (session.tenantId.isNotBlank()) {
            secureStore.putString(SecureStore.KEY_TENANT_ID, session.tenantId)
        }
        secureStore.putString(KEY_DISPLAY_NAME, session.displayName)
        secureStore.putString(KEY_USER_ID, session.userId)
        secureStore.putString(KEY_PHONE, session.phone)
        secureStore.putString(KEY_PERMISSION_CODES, encodeCodes(session.permissionCodes))
        secureStore.putString(KEY_HAS_PASSWORD, if (session.hasPassword) "1" else "0")
        secureStore.putString(KEY_ROLE_NAME, session.roleName)
        if (session.serviceAreaId.isNotBlank()) {
            secureStore.putString(SecureStore.KEY_SERVICE_AREA_ID, session.serviceAreaId)
        }
        if (session.serviceAreaName.isNotBlank()) {
            secureStore.putString(SecureStore.KEY_SERVICE_AREA_NAME, session.serviceAreaName)
        }
    }

    private suspend fun refreshRuntimeConfig(networkSession: NetworkSession): OpsResult<Unit> {
        val api = authApi ?: return OpsResult.Ok(Unit)
        return when (val result = api.selectAppConfig(networkSession)) {
            is OpsResult.Ok -> {
                runtimeConfig = result.value
                persistRuntimeConfig(result.value)
                OpsResult.Ok(Unit)
            }
            is OpsResult.Err -> result
        }
    }

    private fun persistRuntimeConfig(config: TenantRuntimeConfig) {
        secureStore.putString(KEY_QR_HOSTS, config.qrHosts.joinToString("|"))
        secureStore.putString(KEY_RUNTIME_TENANT_NAME, config.tenantName)
    }

    private fun restoreRuntimeConfig(): TenantRuntimeConfig? {
        val hosts = secureStore.getString(KEY_QR_HOSTS)?.split('|')
            ?.map { it.trim() }
            ?.filter { it.isNotBlank() }
            .orEmpty()
        if (hosts.isEmpty()) return null
        return TenantRuntimeConfig(
            tenantName = secureStore.getString(KEY_RUNTIME_TENANT_NAME).orEmpty(),
            qrHosts = hosts,
        )
    }

    private fun restoreSession(): UserSession? {
        val access = secureStore.getString(SecureStore.KEY_ACCESS_TOKEN) ?: return null
        if (access.isBlank()) return null
        return UserSession(
            userId = secureStore.getString(KEY_USER_ID).orEmpty(),
            displayName = secureStore.getString(KEY_DISPLAY_NAME).orEmpty(),
            phone = secureStore.getString(KEY_PHONE).orEmpty(),
            accessToken = access,
            refreshToken = secureStore.getString(SecureStore.KEY_REFRESH_TOKEN).orEmpty(),
            tenantId = secureStore.getString(SecureStore.KEY_TENANT_ID).orEmpty(),
            serviceAreaId = secureStore.getString(SecureStore.KEY_SERVICE_AREA_ID).orEmpty(),
            serviceAreaName = secureStore.getString(SecureStore.KEY_SERVICE_AREA_NAME).orEmpty(),
            permissionCodes = decodeCodes(secureStore.getString(KEY_PERMISSION_CODES)),
            hasPassword = secureStore.getString(KEY_HAS_PASSWORD) != "0",
            roleName = secureStore.getString(KEY_ROLE_NAME).orEmpty(),
        )
    }

    private data class PendingBusinessAuth(
        val phone: String,
        val businesses: List<BusinessTenant>,
        val businessSecret: String,
    )

    companion object {
        const val CODE_SELECT_BUSINESS = "AUTH_SELECT_BUSINESS"

        private const val KEY_DISPLAY_NAME = "user_display_name"
        private const val KEY_USER_ID = "user_id"
        private const val KEY_PHONE = "user_phone"
        private const val KEY_PERMISSION_CODES = "user_permission_codes"
        private const val KEY_HAS_PASSWORD = "user_has_password"
        private const val KEY_ROLE_NAME = "user_role_name"
        private const val KEY_QR_HOSTS = "runtime_qr_hosts"
        private const val KEY_RUNTIME_TENANT_NAME = "runtime_tenant_name"
        private const val DEMO_BUSINESS_SECRET = "demo-business-secret"

        private val DEMO_BUSINESSES = listOf(
            BusinessTenant(
                tenantId = "1001",
                tenantName = "Demo East",
                companyName = "Demo East Co",
            ),
            BusinessTenant(
                tenantId = "1002",
                tenantName = "Demo West",
                companyName = "Demo West Co",
            ),
        )

        private val DEMO_RUNTIME_CONFIG = TenantRuntimeConfig(
            tenantId = "demo",
            tenantName = "Demo Ops",
            alias = "demo",
            qrHosts = listOf("ops.example", "qr.demo"),
        )

        private fun selectBusinessError(): OpsError =
            OpsError.business(CODE_SELECT_BUSINESS, "select business")

        fun deviceId(secureStore: SecureStore): String {
            val existing = secureStore.getString(SecureStore.KEY_DEVICE_ID)
            if (!existing.isNullOrBlank()) return existing
            val created = Ids.uuidV4()
            secureStore.putString(SecureStore.KEY_DEVICE_ID, created)
            return created
        }

        fun encodeCodes(codes: List<String>): String =
            codes.filter { it.isNotBlank() }.distinct().joinToString(",")

        fun decodeCodes(raw: String?): List<String> =
            raw.orEmpty().split(',').map { it.trim() }.filter { it.isNotBlank() }.distinct()
    }
}
