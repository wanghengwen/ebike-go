package com.luopingtech.ebike.ops.data.staff

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.TeamWorker

interface ServiceUserRepository {
    suspend fun listTeamWorkers(serviceId: String): OpsResult<List<TeamWorker>>
}

class ServiceUserRepositoryImpl(
    private val demoMode: Boolean,
    private val api: ServiceUserApi? = null,
) : ServiceUserRepository {
    override suspend fun listTeamWorkers(serviceId: String): OpsResult<List<TeamWorker>> {
        if (demoMode || api == null) {
            return OpsResult.Ok(demoWorkers(serviceId))
        }
        return api.listByServiceId(serviceId)
    }

    companion object {
        fun demoWorkers(serviceId: String): List<TeamWorker> = listOf(
            TeamWorker(name = Strings.t(Str.DemoTeamWorkerA), phone = "1380000${serviceId.takeLast(4).padStart(4, '0')}"),
            TeamWorker(name = Strings.t(Str.DemoTeamWorkerB), phone = "1390000${serviceId.takeLast(4).padStart(4, '0')}"),
            TeamWorker(name = Strings.t(Str.DemoTeamWorkerC), phone = "1370000${serviceId.takeLast(4).padStart(4, '0')}"),
        )
    }
}
