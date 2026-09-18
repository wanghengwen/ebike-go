package com.luopingtech.ebike.ops.feature.auth

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.auth.AuthRepository
import com.luopingtech.ebike.ops.data.auth.AuthRepositoryImpl
import com.luopingtech.ebike.ops.data.auth.CallingCode
import com.luopingtech.ebike.ops.data.auth.CallingCodeCatalog
import com.luopingtech.ebike.ops.data.auth.SmsScene
import com.luopingtech.ebike.ops.domain.model.BusinessTenant
import com.luopingtech.ebike.ops.domain.model.TenantRuntimeConfig
import com.luopingtech.ebike.ops.domain.model.UserSession
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

data class AuthUiState(
    val loading: Boolean = false,
    val session: UserSession? = null,
    val pendingBusinesses: List<BusinessTenant> = emptyList(),
    val needSetPassword: Boolean = false,
    val runtimeConfig: TenantRuntimeConfig? = null,
    val errorMessage: String? = null,
    val infoMessage: String? = null,
    val smsSent: Boolean = false,
    val loginArea: CallingCode = CallingCodeCatalog.DEFAULT,
)

/**
 * Auth feature facade for UI. No Android/iOS types here.
 */
class AuthFeature(
    private val repository: AuthRepository,
) {
    private val _state = MutableStateFlow(initialState())
    val state: StateFlow<AuthUiState> = _state.asStateFlow()

    fun start(scope: CoroutineScope) {
        scope.launch {
            repository.sessionInvalidMessage.collect { message ->
                if (message != null) {
                    _state.value = AuthUiState(
                        errorMessage = message,
                        loginArea = repository.loginAreaCode(),
                    )
                    repository.consumeSessionInvalid()
                }
            }
        }
    }

    fun setLoginAreaCode(code: CallingCode) {
        repository.setLoginAreaCode(code)
        _state.value = _state.value.copy(loginArea = code)
    }

    suspend fun login(account: String, password: String) {
        _state.value = _state.value.copy(
            loading = true,
            errorMessage = null,
            infoMessage = null,
            smsSent = false,
        )
        applyLoginResult(repository.login(account, password))
    }

    suspend fun sendSmsCode(phone: String, scene: Int = SmsScene.LOGIN) {
        _state.value = _state.value.copy(loading = true, errorMessage = null, smsSent = false)
        when (val result = repository.sendSmsCode(phone, scene)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    smsSent = true,
                    errorMessage = null,
                )
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    smsSent = false,
                    errorMessage = result.error.message,
                )
            }
        }
    }

    suspend fun loginWithSms(phone: String, messageCode: String) {
        _state.value = _state.value.copy(loading = true, errorMessage = null, infoMessage = null)
        applyLoginResult(repository.loginWithSms(phone, messageCode))
    }

    suspend fun selectBusiness(tenantId: String) {
        _state.value = _state.value.copy(loading = true, errorMessage = null)
        applyLoginResult(repository.selectBusiness(tenantId))
    }

    fun clearPendingBusiness() {
        repository.clearPendingBusiness()
        _state.value = _state.value.copy(pendingBusinesses = emptyList(), errorMessage = null)
    }

    suspend fun forgetPassword(phone: String, code: String, newPassword: String) {
        _state.value = _state.value.copy(loading = true, errorMessage = null, infoMessage = null)
        when (val result = repository.forgetPassword(phone, code, newPassword)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    infoMessage = Strings.t(Str.PasswordResetOk),
                    smsSent = false,
                )
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    errorMessage = result.error.message,
                )
            }
        }
    }

    suspend fun setPassword(newPassword: String) {
        _state.value = _state.value.copy(loading = true, errorMessage = null)
        when (val result = repository.setPassword(newPassword)) {
            is OpsResult.Ok -> {
                _state.value = AuthUiState(
                    loading = false,
                    session = result.value,
                    needSetPassword = false,
                    runtimeConfig = repository.runtimeConfig(),
                    loginArea = repository.loginAreaCode(),
                )
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    errorMessage = result.error.message,
                )
            }
        }
    }

    /** Legacy: leaving set-password without completing clears the session. */
    suspend fun cancelSetPassword() {
        repository.logout()
        _state.value = AuthUiState(loginArea = repository.loginAreaCode())
    }

    suspend fun updatePassword(oldPassword: String, newPassword: String): OpsResult<Unit> {
        _state.value = _state.value.copy(loading = true, errorMessage = null, infoMessage = null)
        val result = repository.updatePassword(oldPassword, newPassword)
        when (result) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(loading = false)
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    errorMessage = result.error.message,
                )
            }
        }
        return result
    }

    suspend fun logout() {
        repository.logout()
        _state.value = AuthUiState(loginArea = repository.loginAreaCode())
    }

    suspend fun listTenantsForSwitch(): OpsResult<List<BusinessTenant>> =
        repository.listTenantsForSwitch()

    fun canSwitchBusiness(): Boolean = repository.canSwitchBusiness()

    suspend fun switchBusiness(tenantId: String): OpsResult<UserSession> {
        _state.value = _state.value.copy(loading = true, errorMessage = null, infoMessage = null)
        return when (val result = repository.switchBusiness(tenantId)) {
            is OpsResult.Ok -> {
                _state.value = AuthUiState(
                    loading = false,
                    session = result.value,
                    needSetPassword = !result.value.hasPassword,
                    runtimeConfig = repository.runtimeConfig(),
                    loginArea = repository.loginAreaCode(),
                )
                result
            }
            is OpsResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    errorMessage = result.error.message,
                )
                result
            }
        }
    }

    fun refreshSessionFromStore() {
        _state.value = initialState()
    }

    private fun initialState(): AuthUiState {
        val session = repository.currentSession()
        return AuthUiState(
            session = session,
            pendingBusinesses = repository.pendingBusinesses(),
            needSetPassword = repository.needsSetPassword(),
            runtimeConfig = repository.runtimeConfig(),
            loginArea = repository.loginAreaCode(),
        )
    }

    private fun applyLoginResult(result: OpsResult<UserSession>) {
        val loginArea = repository.loginAreaCode()
        when (result) {
            is OpsResult.Ok -> {
                _state.value = AuthUiState(
                    loading = false,
                    session = result.value,
                    needSetPassword = !result.value.hasPassword,
                    runtimeConfig = repository.runtimeConfig(),
                    loginArea = loginArea,
                )
            }
            is OpsResult.Err -> {
                val pending = if (result.error.code == AuthRepositoryImpl.CODE_SELECT_BUSINESS) {
                    repository.pendingBusinesses()
                } else {
                    emptyList()
                }
                _state.value = AuthUiState(
                    loading = false,
                    session = null,
                    pendingBusinesses = pending,
                    errorMessage = if (pending.isNotEmpty()) null else result.error.message,
                    loginArea = loginArea,
                )
            }
        }
    }
}
