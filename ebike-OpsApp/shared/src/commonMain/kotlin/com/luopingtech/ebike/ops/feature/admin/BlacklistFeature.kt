package com.luopingtech.ebike.ops.feature.admin

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.admin.AdminRepository
import com.luopingtech.ebike.ops.domain.admin.BlacklistItem
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.order.OrderSearchClassifier
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class BlacklistUiState(
    val loading: Boolean = false,
    val items: List<BlacklistItem> = emptyList(),
    val keyword: String = "",
    val message: String? = null,
    val errorMessage: String? = null,
)

/**
 * 对齐原版 BlackListViewModel.checkInput：手机号 → 身份证 → 姓名。
 */
data class BlacklistSearchQuery(
    val phone: String? = null,
    val authName: String? = null,
    val authNo: String? = null,
)

class BlacklistFeature(
    private val repository: AdminRepository,
) {
    private val _state = MutableStateFlow(BlacklistUiState())
    val state: StateFlow<BlacklistUiState> = _state.asStateFlow()

    fun clear() {
        _state.value = BlacklistUiState()
    }

    fun setKeyword(value: String) {
        _state.value = _state.value.copy(keyword = value)
    }

    suspend fun load(area: ServiceArea?) {
        search(area, keywordOverride = _state.value.keyword)
    }

    suspend fun search(area: ServiceArea?, keywordOverride: String? = null) {
        if (area == null) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.SelectServiceAreaFirst))
            return
        }
        val keyword = (keywordOverride ?: _state.value.keyword).filterNot { it.isWhitespace() }
        if (keywordOverride != null) {
            _state.value = _state.value.copy(keyword = keywordOverride)
        }
        val query = classifyKeyword(keyword)
        _state.value = _state.value.copy(loading = true, errorMessage = null)
        when (
            val result = repository.blacklistPage(
                serviceId = area.id,
                phone = query.phone,
                authName = query.authName,
                authNo = query.authNo,
            )
        ) {
            is OpsResult.Ok -> _state.value = _state.value.copy(loading = false, items = result.value)
            is OpsResult.Err -> _state.value = _state.value.copy(loading = false, errorMessage = result.error.message)
        }
    }

    suspend fun cancel(area: ServiceArea?, id: String) {
        if (area == null) return
        _state.value = _state.value.copy(loading = true, message = null, errorMessage = null)
        when (val result = repository.blacklistCancel(id)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(message = Strings.t(Str.OperationOk))
                search(area)
            }
            is OpsResult.Err -> _state.value = _state.value.copy(loading = false, errorMessage = result.error.message)
        }
    }

    companion object {
        fun classifyKeyword(raw: String?): BlacklistSearchQuery {
            val it = raw?.filterNot { ch -> ch.isWhitespace() }.orEmpty()
            if (it.isEmpty()) return BlacklistSearchQuery()
            return when {
                OrderSearchClassifier.isPhone(it) -> BlacklistSearchQuery(phone = it)
                isLikelyIdCard(it) -> BlacklistSearchQuery(authNo = it)
                else -> BlacklistSearchQuery(authName = it)
            }
        }

        /** 轻量对齐原版身份证长度校验（15/18），避免把姓名误判成证件号。 */
        private fun isLikelyIdCard(value: String): Boolean {
            if (value.length != 15 && value.length != 18) return false
            val body = if (value.length == 18) value.dropLast(1) else value
            if (!body.all { it.isDigit() }) return false
            if (value.length == 18) {
                val last = value.last()
                if (!(last.isDigit() || last == 'x' || last == 'X')) return false
            }
            return true
        }
    }
}
