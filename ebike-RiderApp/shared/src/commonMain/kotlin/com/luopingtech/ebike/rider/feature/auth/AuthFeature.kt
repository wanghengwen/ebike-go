package com.luopingtech.ebike.rider.feature.auth

import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.core.i18n.Strings
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.data.auth.AuthRepository
import com.luopingtech.ebike.rider.data.auth.SmsScene
import com.luopingtech.ebike.rider.domain.model.UserSession
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

data class AuthUiState(
    val loading: Boolean = false,
    val session: UserSession? = null,
    val needLogin: Boolean = true,
    val smsSent: Boolean = false,
    val countdown: Int = 0,
    val error: String? = null,
    val info: String? = null,
    val phone: String = "",
    val agreedProtocol: Boolean = false,
)

/**
 * Auth feature facade for UI. No Android/iOS types here.
 */
class AuthFeature(
    private val repository: AuthRepository,
) {
    private val _state = MutableStateFlow(initialState())
    val state: StateFlow<AuthUiState> = _state.asStateFlow()
    private var started = false

    fun start(scope: CoroutineScope) {
        if (started) return
        started = true
        scope.launch {
            repository.sessionInvalidMessage.collect { message ->
                if (message != null) {
                    _state.value = _state.value.copy(
                        loading = false,
                        session = null,
                        needLogin = true,
                        error = message,
                        info = null,
                        smsSent = false,
                        countdown = 0,
                    )
                    repository.consumeSessionInvalid()
                }
            }
        }
    }

    fun setPhone(phone: String) {
        _state.value = _state.value.copy(phone = phone, error = null)
    }

    fun setAgreedProtocol(agreed: Boolean) {
        _state.value = _state.value.copy(agreedProtocol = agreed, error = null)
    }

    fun tickCountdown() {
        val n = _state.value.countdown
        if (n > 0) {
            _state.value = _state.value.copy(countdown = n - 1)
        }
    }

    suspend fun sendSms(phone: String, scene: Int = SmsScene.LOGIN) {
        val trimmed = phone.trim()
        if (trimmed.isBlank()) {
            _state.value = _state.value.copy(error = Strings.t(Str.PhoneInvalid), smsSent = false)
            return
        }
        if (!_state.value.agreedProtocol) {
            _state.value = _state.value.copy(error = Strings.t(Str.AgreeProtocolRequired))
            return
        }
        _state.value = _state.value.copy(
            loading = true,
            error = null,
            info = null,
            smsSent = false,
            phone = trimmed,
        )
        when (val result = repository.sendSmsCode(trimmed, scene)) {
            is RiderResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    smsSent = true,
                    countdown = SMS_COUNTDOWN_SECONDS,
                    info = Strings.t(Str.SmsCodeSentTo, trimmed),
                    error = null,
                )
            }
            is RiderResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    smsSent = false,
                    error = result.error.message,
                )
            }
        }
    }

    suspend fun loginWithSms(phone: String, messageCode: String) {
        val trimmed = phone.trim()
        if (!_state.value.agreedProtocol) {
            _state.value = _state.value.copy(error = Strings.t(Str.AgreeProtocolRequired))
            return
        }
        _state.value = _state.value.copy(
            loading = true,
            error = null,
            info = null,
            phone = trimmed,
        )
        applySessionResult(repository.loginWithSms(trimmed, messageCode))
    }

    suspend fun logout() {
        repository.logout()
        _state.value = AuthUiState(
            needLogin = true,
            phone = _state.value.phone,
            agreedProtocol = _state.value.agreedProtocol,
        )
    }

    suspend fun restoreSession() {
        val existing = repository.currentSession()
        if (existing == null) {
            _state.value = _state.value.copy(
                loading = false,
                session = null,
                needLogin = true,
            )
            return
        }
        _state.value = _state.value.copy(
            loading = true,
            session = existing,
            needLogin = false,
            error = null,
        )
        when (val result = repository.refreshProfile()) {
            is RiderResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    session = result.value,
                    needLogin = false,
                )
            }
            is RiderResult.Err -> {
                val still = repository.currentSession()
                _state.value = _state.value.copy(
                    loading = false,
                    session = still,
                    needLogin = still == null,
                    error = if (still == null) result.error.message else null,
                )
            }
        }
    }

    private fun initialState(): AuthUiState {
        val session = repository.currentSession()
        return AuthUiState(
            session = session,
            needLogin = session == null,
        )
    }

    private fun applySessionResult(result: RiderResult<UserSession>) {
        when (result) {
            is RiderResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    session = result.value,
                    needLogin = false,
                    error = null,
                    info = null,
                    smsSent = false,
                    countdown = 0,
                )
            }
            is RiderResult.Err -> {
                _state.value = _state.value.copy(
                    loading = false,
                    session = null,
                    needLogin = true,
                    error = result.error.message,
                )
            }
        }
    }

    companion object {
        const val SMS_COUNTDOWN_SECONDS: Int = 60
    }
}
