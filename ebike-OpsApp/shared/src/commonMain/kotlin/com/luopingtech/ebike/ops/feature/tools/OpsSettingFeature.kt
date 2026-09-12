package com.luopingtech.ebike.ops.feature.tools

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.tools.OpsSettingRepository
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.domain.tools.OpsSettingConfig
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class OpsSettingUiState(
    val loading: Boolean = false,
    val saving: Boolean = false,
    val threshold: Int = 20,
    val autoSwap: Boolean = false,
    val message: String? = null,
    val errorMessage: String? = null,
)

class OpsSettingFeature(
    private val repository: OpsSettingRepository,
) {
    private val _state = MutableStateFlow(OpsSettingUiState())
    val state: StateFlow<OpsSettingUiState> = _state.asStateFlow()

    fun clear() {
        _state.value = OpsSettingUiState()
    }

    fun setThreshold(value: Int) {
        _state.value = _state.value.copy(threshold = value.coerceIn(0, 100), errorMessage = null)
    }

    fun setAutoSwap(enabled: Boolean) {
        _state.value = _state.value.copy(autoSwap = enabled, errorMessage = null)
    }

    suspend fun load(area: ServiceArea?) {
        if (area == null) {
            _state.value = _state.value.copy(
                loading = false,
                errorMessage = Strings.t(Str.SelectServiceAreaFirst),
            )
            return
        }
        _state.value = _state.value.copy(loading = true, errorMessage = null)
        when (val result = repository.load(area.id)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    threshold = result.value.swapBatteryThreshold,
                    autoSwap = result.value.izAutoSwapBattery,
                )
            }
            is OpsResult.Err -> _state.value = _state.value.copy(loading = false, errorMessage = result.error.message)
        }
    }

    suspend fun save(area: ServiceArea?) {
        if (area == null) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.SelectServiceAreaFirst))
            return
        }
        val snap = _state.value
        _state.value = snap.copy(saving = true, errorMessage = null, message = null)
        val config = OpsSettingConfig(snap.threshold, snap.autoSwap)
        when (val result = repository.save(area.id, config)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(
                saving = false,
                message = Strings.t(Str.OpsSettingSaved),
            )
            is OpsResult.Err -> _state.value = _state.value.copy(saving = false, errorMessage = result.error.message)
        }
    }
}
