package com.luopingtech.ebike.ops.data.tools

import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.tools.OpsSettingConfig

interface OpsSettingRepository {
    suspend fun load(serviceId: String): OpsResult<OpsSettingConfig>
    suspend fun save(serviceId: String, config: OpsSettingConfig): OpsResult<Unit>
}

class OpsSettingRepositoryImpl(
    private val demoMode: Boolean,
    private val api: OpsSettingApi? = null,
) : OpsSettingRepository {
    private var demoConfig = OpsSettingConfig(swapBatteryThreshold = 20, izAutoSwapBattery = false)

    override suspend fun load(serviceId: String): OpsResult<OpsSettingConfig> {
        if (demoMode || api == null) {
            return OpsResult.Ok(demoConfig)
        }
        return api.load(serviceId)
    }

    override suspend fun save(serviceId: String, config: OpsSettingConfig): OpsResult<Unit> {
        if (demoMode || api == null) {
            demoConfig = config
            return OpsResult.Ok(Unit)
        }
        return api.save(serviceId, config.swapBatteryThreshold, config.izAutoSwapBattery)
    }
}
