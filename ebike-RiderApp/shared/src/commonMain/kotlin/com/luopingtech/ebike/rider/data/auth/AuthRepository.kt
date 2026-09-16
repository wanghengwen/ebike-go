package com.luopingtech.ebike.rider.data.auth

import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.core.i18n.Strings
import com.luopingtech.ebike.rider.core.network.ApiCodes
import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.domain.model.UserSession
import com.luopingtech.ebike.rider.platform.SecureStore
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

interface AuthRepository {
    val sessionInvalidMessage: StateFlow<String?>

    suspend fun sendSmsCode(phone: String, scene: Int = SmsScene.LOGIN): RiderResult<Unit>
    suspend fun loginWithSms(phone: String, messageCode: String): RiderResult<UserSession>
    suspend fun refreshAccessToken(): RiderResult<UserSession>
    suspend fun refreshProfile(): RiderResult<UserSession>
    suspend fun logout()
    fun currentSession(): UserSession?
    fun markSessionInvalid(message: String)
    fun consumeSessionInvalid()
}

class AuthRepositoryImpl(
    private val secureStore: SecureStore,
    private val demoMode: Boolean,
    private val authApi: AuthRemote? = null,
    private val userApiProvider: () -> UserRemote? = { null },
    private val defaultTenantId: String = "",
) : AuthRepository {
    private var session: UserSession? = run {
        val restored = restoreSession()
        // 从 demo 切到真实网关时，清掉本地的 demo-token，否则会跳过登录直接进首页。
        if (!demoMode && restored != null && isDemoToken(restored.accessToken)) {
            wipePersistedSession()
            null
        } else {
            restored
        }
    }
    private val _sessionInvalidMessage = MutableStateFlow<String?>(null)
    override val sessionInvalidMessage: StateFlow<String?> = _sessionInvalidMessage.asStateFlow()

    override suspend fun sendSmsCode(phone: String, scene: Int): RiderResult<Unit> {
        if (phone.isBlank()) {
            return RiderResult.Err(RiderError.business("AUTH_INVALID", Strings.t(Str.PhoneInvalid)))
        }
        if (demoMode || authApi == null) return RiderResult.Ok(Unit)
        return authApi.sendSmsCode(phone, scene)
    }

    override suspend fun loginWithSms(phone: String, messageCode: String): RiderResult<UserSession> {
        if (phone.isBlank()) {
            return RiderResult.Err(RiderError.business("AUTH_INVALID", Strings.t(Str.PhoneInvalid)))
        }
        prepareFreshLogin()
        if (demoMode || authApi == null) {
            return saveDemoSession(phone)
        }
        return when (val tokenResult = authApi.loginWithSms(phone, messageCode)) {
            is RiderResult.Err -> tokenResult
            is RiderResult.Ok -> finalizeFromToken(tokenResult.value, fallbackPhone = phone)
        }
    }

    override suspend fun refreshAccessToken(): RiderResult<UserSession> {
        val current = session ?: return RiderResult.Err(RiderError.unauthorized())
        if (demoMode || authApi == null) return RiderResult.Ok(current)
        val refresh = current.refreshToken
        if (refresh.isBlank()) {
            markSessionInvalid(Strings.t(Str.SessionExpired))
            return RiderResult.Err(RiderError.unauthorized("missing refresh token"))
        }
        return when (val result = authApi.refreshToken(refresh)) {
            is RiderResult.Err -> {
                markSessionInvalid(result.error.message.ifBlank { Strings.t(Str.SessionExpired) })
                result
            }
            is RiderResult.Ok -> {
                val token = result.value
                val next = current.copy(
                    accessToken = token.accessToken.ifBlank { current.accessToken },
                    refreshToken = token.refreshToken.ifBlank { current.refreshToken },
                    displayName = token.nickname.ifBlank { current.displayName },
                    phone = token.phone.ifBlank { current.phone },
                    pin = token.pin.ifBlank { current.pin },
                    userId = token.pin.ifBlank { current.userId },
                )
                persistTokens(next)
                session = next
                RiderResult.Ok(next)
            }
        }
    }

    override suspend fun refreshProfile(): RiderResult<UserSession> {
        val current = session ?: return RiderResult.Err(RiderError.unauthorized())
        if (demoMode) return RiderResult.Ok(current)
        val userApi = userApiProvider() ?: return RiderResult.Ok(current)
        var next = current
        when (val profile = userApi.fetchPersonInfo()) {
            is RiderResult.Ok -> {
                val p = profile.value
                next = next.copy(
                    userId = p.pin.ifBlank { next.userId },
                    pin = p.pin.ifBlank { next.pin },
                    displayName = p.nickname.ifBlank { p.phone.ifBlank { next.displayName } },
                    phone = p.phone.ifBlank { next.phone },
                    avatar = p.avatar.ifBlank { next.avatar },
                    balance = p.balance,
                    serviceAreaId = p.serviceId.ifBlank { next.serviceAreaId },
                )
                persistTokens(next)
                session = next
            }
            is RiderResult.Err -> {
                if (session == null || isFatalAuth(profile.error)) {
                    return profile
                }
            }
        }
        when (val account = userApi.fetchUserAccount()) {
            is RiderResult.Ok -> {
                val wallet = account.value.userWallet
                val balance = wallet?.balance?.takeIf { it != 0L } ?: account.value.balance
                if (balance != 0L || next.balance == 0L) {
                    next = next.copy(balance = if (balance != 0L) balance else next.balance)
                    persistTokens(next)
                    session = next
                }
            }
            is RiderResult.Err -> {
                if (session == null || isFatalAuth(account.error)) {
                    return account
                }
            }
        }
        return RiderResult.Ok(session ?: next)
    }

    override suspend fun logout() {
        clearLocalSession()
    }

    override fun currentSession(): UserSession? = session

    override fun markSessionInvalid(message: String) {
        clearLocalSession()
        _sessionInvalidMessage.value = message.ifBlank { Strings.t(Str.SessionExpired) }
    }

    override fun consumeSessionInvalid() {
        _sessionInvalidMessage.value = null
    }

    private fun prepareFreshLogin() {
        _sessionInvalidMessage.value = null
        clearLocalSession()
    }

    private suspend fun finalizeFromToken(
        token: OauthTokenDto,
        fallbackPhone: String,
    ): RiderResult<UserSession> {
        if (token.accessToken.isBlank()) {
            return RiderResult.Err(RiderError.business("AUTH_EMPTY_TOKEN", "accessToken empty"))
        }
        val phone = token.phone.ifBlank { PhoneNormalizer.toLoginPhone(fallbackPhone) }
        var next = UserSession(
            userId = token.pin.ifBlank { phone },
            pin = token.pin,
            displayName = token.nickname.ifBlank { phone },
            phone = phone,
            accessToken = token.accessToken,
            refreshToken = token.refreshToken,
            tenantId = storedTenantId(),
            avatar = token.avatar,
        )
        persistTokens(next)
        session = next
        when (val profile = refreshProfile()) {
            is RiderResult.Ok -> next = profile.value
            is RiderResult.Err -> {
                if (session == null || isFatalAuth(profile.error)) return profile
            }
        }
        return RiderResult.Ok(session ?: next)
    }

    private fun saveDemoSession(phone: String): RiderResult<UserSession> {
        val normalized = PhoneNormalizer.toLoginPhone(phone)
        val next = UserSession(
            userId = "demo-user",
            pin = "demo-pin",
            displayName = normalized,
            phone = normalized,
            accessToken = DEMO_ACCESS,
            refreshToken = DEMO_REFRESH,
            tenantId = storedTenantId(),
        )
        persistTokens(next)
        session = next
        return RiderResult.Ok(next)
    }

    private fun persistTokens(session: UserSession) {
        secureStore.putString(SecureStore.KEY_ACCESS_TOKEN, session.accessToken)
        secureStore.putString(SecureStore.KEY_REFRESH_TOKEN, session.refreshToken)
        if (session.tenantId.isNotBlank()) {
            secureStore.putString(SecureStore.KEY_TENANT_ID, session.tenantId)
        }
        secureStore.putString(KEY_DISPLAY_NAME, session.displayName)
        secureStore.putString(KEY_USER_ID, session.userId)
        secureStore.putString(KEY_USER_PIN, session.pin)
        secureStore.putString(KEY_PHONE, session.phone)
        secureStore.putString(KEY_AVATAR, session.avatar)
        secureStore.putString(KEY_BALANCE, session.balance.toString())
        if (session.serviceAreaId.isNotBlank()) {
            secureStore.putString(SecureStore.KEY_SERVICE_AREA_ID, session.serviceAreaId)
        }
    }

    private fun restoreSession(): UserSession? {
        val access = secureStore.getString(SecureStore.KEY_ACCESS_TOKEN) ?: return null
        if (access.isBlank()) return null
        return UserSession(
            userId = secureStore.getString(KEY_USER_ID).orEmpty(),
            pin = secureStore.getString(KEY_USER_PIN).orEmpty(),
            displayName = secureStore.getString(KEY_DISPLAY_NAME).orEmpty(),
            phone = secureStore.getString(KEY_PHONE).orEmpty(),
            accessToken = access,
            refreshToken = secureStore.getString(SecureStore.KEY_REFRESH_TOKEN).orEmpty(),
            tenantId = storedTenantId(),
            serviceAreaId = secureStore.getString(SecureStore.KEY_SERVICE_AREA_ID).orEmpty(),
            avatar = secureStore.getString(KEY_AVATAR).orEmpty(),
            balance = secureStore.getString(KEY_BALANCE)?.toLongOrNull() ?: 0L,
        )
    }

    private fun clearLocalSession() {
        session = null
        wipePersistedSession()
    }

    private fun wipePersistedSession() {
        secureStore.remove(SecureStore.KEY_ACCESS_TOKEN)
        secureStore.remove(SecureStore.KEY_REFRESH_TOKEN)
        secureStore.remove(KEY_DISPLAY_NAME)
        secureStore.remove(KEY_USER_ID)
        secureStore.remove(KEY_USER_PIN)
        secureStore.remove(KEY_PHONE)
        secureStore.remove(KEY_AVATAR)
        secureStore.remove(KEY_BALANCE)
    }

    private fun storedTenantId(): String =
        secureStore.getString(SecureStore.KEY_TENANT_ID)?.takeIf { it.isNotBlank() }
            ?: defaultTenantId

    private fun isFatalAuth(error: RiderError): Boolean =
        error.code == "UNAUTHORIZED" ||
            ApiCodes.shouldForceRelogin(error.code) ||
            ApiCodes.isGuestUnauthorized(error.code)

    private fun isDemoToken(accessToken: String): Boolean =
        accessToken == DEMO_ACCESS || accessToken.startsWith("demo-")

    companion object {
        const val DEMO_ACCESS: String = "demo-token"
        const val DEMO_REFRESH: String = "demo-refresh"

        private const val KEY_DISPLAY_NAME = "user_display_name"
        private const val KEY_USER_ID = "user_id"
        private const val KEY_USER_PIN = "user_pin"
        private const val KEY_PHONE = "user_phone"
        private const val KEY_AVATAR = "user_avatar"
        private const val KEY_BALANCE = "user_balance"
    }
}
