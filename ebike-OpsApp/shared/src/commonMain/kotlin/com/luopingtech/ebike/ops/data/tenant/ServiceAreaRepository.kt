package com.luopingtech.ebike.ops.data.tenant

import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import com.luopingtech.ebike.ops.platform.SecureStore
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json

interface ServiceAreaRepository {
    suspend fun loadAreas(): OpsResult<List<ServiceArea>>
    fun currentArea(): ServiceArea?
    fun selectArea(area: ServiceArea)
    fun clearSelection()
}

class ServiceAreaRepositoryImpl(
    private val secureStore: SecureStore,
    private val demoMode: Boolean,
    private val api: ServiceAreaApi? = null,
    private val json: Json = Json { ignoreUnknownKeys = true },
) : ServiceAreaRepository {
    private var cachedAreas: List<ServiceArea> = emptyList()
    private var selected: ServiceArea? = restoreSelected()

    override suspend fun loadAreas(): OpsResult<List<ServiceArea>> {
        if (demoMode || api == null) {
            cachedAreas = DEMO_AREAS
            // Do not auto-select — host shows area gate like legacy ChangeServiceAreaActivity.
            return OpsResult.Ok(cachedAreas)
        }
        return when (val result = api.listByToken()) {
            is OpsResult.Err -> result
            is OpsResult.Ok -> {
                cachedAreas = result.value
                val currentId = selected?.id ?: secureStore.getString(SecureStore.KEY_SERVICE_AREA_ID)
                val matched = cachedAreas.firstOrNull { it.id == currentId }
                if (matched != null) {
                    selectArea(matched)
                }
                // No match → leave unselected so UI can force picker.
                OpsResult.Ok(cachedAreas)
            }
        }
    }

    override fun currentArea(): ServiceArea? = selected

    override fun selectArea(area: ServiceArea) {
        selected = area
        secureStore.putString(SecureStore.KEY_SERVICE_AREA_ID, area.id)
        secureStore.putString(SecureStore.KEY_SERVICE_AREA_NAME, area.name)
        secureStore.putString(KEY_AREA_JSON, json.encodeToString(ServiceAreaDto(
            id = area.id.toLongOrNull() ?: 0L,
            name = area.name,
            centerLat = area.centerLat,
            centerLng = area.centerLng,
            agentId = area.agentId,
            minDistance = area.minDistance,
        )))
    }

    override fun clearSelection() {
        selected = null
        secureStore.remove(SecureStore.KEY_SERVICE_AREA_ID)
        secureStore.remove(SecureStore.KEY_SERVICE_AREA_NAME)
        secureStore.remove(KEY_AREA_JSON)
    }

    private fun restoreSelected(): ServiceArea? {
        val raw = secureStore.getString(KEY_AREA_JSON) ?: return null
        return runCatching { json.decodeFromString(ServiceAreaDto.serializer(), raw).toDomain() }
            .getOrNull()
            ?: run {
                val id = secureStore.getString(SecureStore.KEY_SERVICE_AREA_ID) ?: return null
                val name = secureStore.getString(SecureStore.KEY_SERVICE_AREA_NAME).orEmpty()
                ServiceArea(id = id, name = name)
            }
    }

    companion object {
        private const val KEY_AREA_JSON = "service_area_json"

        val DEMO_AREAS = listOf(
            ServiceArea(id = "1001", name = "Demo 城东服务区", centerLat = 28.22, centerLng = 112.94),
            ServiceArea(id = "1002", name = "Demo 城西服务区", centerLat = 28.20, centerLng = 112.88),
            ServiceArea(id = "1003", name = "Demo 高新区", centerLat = 28.25, centerLng = 112.90),
        )
    }
}
