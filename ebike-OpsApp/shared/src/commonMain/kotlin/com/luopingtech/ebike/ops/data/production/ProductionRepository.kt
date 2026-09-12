package com.luopingtech.ebike.ops.data.production

import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.domain.model.BindVehicleInfo
import com.luopingtech.ebike.ops.domain.model.SaddleOverloadContact
import com.luopingtech.ebike.ops.domain.model.ShelfCheckResult

interface ProductionRepository {
    suspend fun getBind(carId: String): OpsResult<BindVehicleInfo>
    suspend fun bind(carId: String, imei: String?, helmet: String?): OpsResult<Unit>
    suspend fun unbind(carId: String, imei: String?, helmet: String?): OpsResult<Unit>
    suspend fun onlineCheck(carId: String): OpsResult<ShelfCheckResult>
    suspend fun offlineCheck(carId: String): OpsResult<ShelfCheckResult>
    suspend fun onlineByCarList(serviceId: String, carIds: List<String>): OpsResult<Unit>
    suspend fun offlineByCarList(carIds: List<String>): OpsResult<Unit>
    suspend fun triggerOverloadCheck(carId: String, imei: String, timeoutSec: Int = 60): OpsResult<Unit>
    suspend fun queryOverloadContact(imei: String): OpsResult<SaddleOverloadContact>
}

class ProductionRepositoryImpl(
    private val demoMode: Boolean,
    private val api: ProductionApi? = null,
) : ProductionRepository {
    private val demoBinds = mutableMapOf<String, BindVehicleInfo>()
    private val demoOverloadPolls = mutableMapOf<String, Int>()

    override suspend fun getBind(carId: String): OpsResult<BindVehicleInfo> {
        if (carId.isBlank()) return OpsResult.Err(OpsError.business("PROD_EMPTY", "carId empty"))
        if (demoMode || api == null) {
            return OpsResult.Ok(
                demoBinds[carId] ?: BindVehicleInfo(
                    id = "demo-$carId",
                    carId = carId,
                    carNo = carId.takeLast(6),
                    brand = "Demo",
                    model = "Ops",
                    imei = "",
                    helmet = "",
                    serviceId = "",
                    bindTime = "",
                ),
            )
        }
        return api.getBind(carId.trim())
    }

    override suspend fun bind(carId: String, imei: String?, helmet: String?): OpsResult<Unit> {
        if (carId.isBlank()) return OpsResult.Err(OpsError.business("PROD_EMPTY", "carId empty"))
        if (imei.isNullOrBlank()) return OpsResult.Err(OpsError.business("PROD_IMEI", "imei required"))
        if (demoMode || api == null) {
            demoBinds[carId] = BindVehicleInfo(
                id = "demo-$carId",
                carId = carId.trim(),
                carNo = carId.takeLast(6),
                brand = "Demo",
                model = "Ops",
                imei = imei.trim(),
                helmet = helmet.orEmpty().trim(),
                bindTime = "demo-now",
            )
            return OpsResult.Ok(Unit)
        }
        return api.bind(carId.trim(), imei, helmet)
    }

    override suspend fun unbind(carId: String, imei: String?, helmet: String?): OpsResult<Unit> {
        if (carId.isBlank()) return OpsResult.Err(OpsError.business("PROD_EMPTY", "carId empty"))
        if (demoMode || api == null) {
            demoBinds.remove(carId)
            return OpsResult.Ok(Unit)
        }
        return api.unbind(carId.trim(), imei, helmet)
    }

    override suspend fun onlineCheck(carId: String): OpsResult<ShelfCheckResult> {
        if (carId.isBlank()) return OpsResult.Err(OpsError.business("PROD_EMPTY", "carId empty"))
        if (demoMode || api == null) {
            return OpsResult.Ok(
                ShelfCheckResult(
                    carId = carId.trim(),
                    carNo = carId.takeLast(6),
                    brand = "Demo",
                ),
            )
        }
        return api.onlineCheck(carId.trim())
    }

    override suspend fun offlineCheck(carId: String): OpsResult<ShelfCheckResult> {
        if (carId.isBlank()) return OpsResult.Err(OpsError.business("PROD_EMPTY", "carId empty"))
        if (demoMode || api == null) {
            return OpsResult.Ok(
                ShelfCheckResult(
                    carId = carId.trim(),
                    carNo = carId.takeLast(6),
                    brand = "Demo",
                    serviceId = "1001",
                    serviceName = "Demo Area",
                ),
            )
        }
        return api.offlineCheck(carId.trim())
    }

    override suspend fun onlineByCarList(serviceId: String, carIds: List<String>): OpsResult<Unit> {
        if (serviceId.isBlank()) return OpsResult.Err(OpsError.business("PROD_AREA", "serviceId empty"))
        if (carIds.isEmpty()) return OpsResult.Err(OpsError.business("PROD_EMPTY", "carList empty"))
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.onlineByCarList(serviceId, carIds)
    }

    override suspend fun offlineByCarList(carIds: List<String>): OpsResult<Unit> {
        if (carIds.isEmpty()) return OpsResult.Err(OpsError.business("PROD_EMPTY", "carList empty"))
        if (demoMode || api == null) return OpsResult.Ok(Unit)
        return api.offlineByCarList(carIds)
    }

    override suspend fun triggerOverloadCheck(
        carId: String,
        imei: String,
        timeoutSec: Int,
    ): OpsResult<Unit> {
        if (carId.isBlank()) return OpsResult.Err(OpsError.business("PROD_EMPTY", "carId empty"))
        if (imei.isBlank()) return OpsResult.Err(OpsError.business("PROD_IMEI", "imei required"))
        if (demoMode || api == null) {
            demoOverloadPolls[imei.trim()] = 0
            return OpsResult.Ok(Unit)
        }
        return api.triggerOverloadCheck(carId.trim(), imei.trim(), timeoutSec)
    }

    override suspend fun queryOverloadContact(imei: String): OpsResult<SaddleOverloadContact> {
        if (imei.isBlank()) return OpsResult.Err(OpsError.business("PROD_IMEI", "imei required"))
        if (demoMode || api == null) {
            val key = imei.trim()
            val n = (demoOverloadPolls[key] ?: 0) + 1
            demoOverloadPolls[key] = n
            return OpsResult.Ok(
                SaddleOverloadContact(
                    frontSaddleContact = if (n >= 1) 1 else 0,
                    centSaddleContact = if (n >= 2) 1 else 0,
                    backSaddleContact = if (n >= 3) 1 else 0,
                ),
            )
        }
        return api.queryOverloadContact(imei.trim())
    }
}
