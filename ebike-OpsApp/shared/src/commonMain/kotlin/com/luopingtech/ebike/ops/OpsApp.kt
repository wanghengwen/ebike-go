package com.luopingtech.ebike.ops

import com.luopingtech.ebike.ops.core.config.H5ScreenKind
import com.luopingtech.ebike.ops.core.config.H5ScreenUrls
import com.luopingtech.ebike.ops.core.config.TenantConfig
import com.luopingtech.ebike.ops.core.i18n.OpsI18n
import com.luopingtech.ebike.ops.core.logging.OpsLogger
import com.luopingtech.ebike.ops.core.logging.StdoutLogger
import com.luopingtech.ebike.ops.core.network.HttpClientFactory
import com.luopingtech.ebike.ops.core.network.NetworkSession
import com.luopingtech.ebike.ops.core.network.RequestAuth
import com.luopingtech.ebike.ops.core.network.SignedApiClient
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.auth.AuthApi
import com.luopingtech.ebike.ops.data.auth.AuthRepository
import com.luopingtech.ebike.ops.data.auth.AuthRepositoryImpl
import com.luopingtech.ebike.ops.data.auth.CallingCodeCatalog
import com.luopingtech.ebike.ops.data.control.DemoNetworkVehicleControl
import com.luopingtech.ebike.ops.data.control.RemoteNetworkVehicleControl
import com.luopingtech.ebike.ops.data.tenant.ServiceAreaApi
import com.luopingtech.ebike.ops.data.tenant.ServiceAreaRepository
import com.luopingtech.ebike.ops.data.tenant.ServiceAreaRepositoryImpl
import com.luopingtech.ebike.ops.data.task.ChangeBatteryTaskApi
import com.luopingtech.ebike.ops.data.task.ChangeBatteryTaskRepository
import com.luopingtech.ebike.ops.data.task.ChangeBatteryTaskRepositoryImpl
import com.luopingtech.ebike.ops.data.task.InspectionTaskApi
import com.luopingtech.ebike.ops.data.task.InspectionTaskRepository
import com.luopingtech.ebike.ops.data.task.InspectionTaskRepositoryImpl
import com.luopingtech.ebike.ops.data.task.MoveCarTaskApi
import com.luopingtech.ebike.ops.data.task.MoveCarTaskRepository
import com.luopingtech.ebike.ops.data.task.MoveCarTaskRepositoryImpl
import com.luopingtech.ebike.ops.data.task.RepairTaskApi
import com.luopingtech.ebike.ops.data.task.RepairTaskRepository
import com.luopingtech.ebike.ops.data.task.RepairTaskRepositoryImpl
import com.luopingtech.ebike.ops.data.task.TaskAuditApi
import com.luopingtech.ebike.ops.data.task.TaskAuditRepository
import com.luopingtech.ebike.ops.data.task.TaskAuditRepositoryImpl
import com.luopingtech.ebike.ops.data.production.ProductionApi
import com.luopingtech.ebike.ops.data.production.ProductionRepository
import com.luopingtech.ebike.ops.data.production.ProductionRepositoryImpl
import com.luopingtech.ebike.ops.data.sneak.SneakReportApi
import com.luopingtech.ebike.ops.data.sneak.SneakReportRepository
import com.luopingtech.ebike.ops.data.sneak.SneakReportRepositoryImpl
import com.luopingtech.ebike.ops.data.report.FaultReportApi
import com.luopingtech.ebike.ops.data.report.FaultReportRepository
import com.luopingtech.ebike.ops.data.report.FaultReportRepositoryImpl
import com.luopingtech.ebike.ops.data.relocation.RelocationApi
import com.luopingtech.ebike.ops.data.relocation.RelocationRepository
import com.luopingtech.ebike.ops.data.relocation.RelocationRepositoryImpl
import com.luopingtech.ebike.ops.data.workorder.InspectionOrderApi
import com.luopingtech.ebike.ops.data.workorder.InspectionOrderRepository
import com.luopingtech.ebike.ops.data.workorder.InspectionOrderRepositoryImpl
import com.luopingtech.ebike.ops.data.workorder.RepairOrderApi
import com.luopingtech.ebike.ops.data.workorder.RepairOrderRepository
import com.luopingtech.ebike.ops.data.workorder.RepairOrderRepositoryImpl
import com.luopingtech.ebike.ops.data.tracking.EmployeeTrackApi
import com.luopingtech.ebike.ops.data.tracking.EmployeeTrackRepository
import com.luopingtech.ebike.ops.data.tracking.EmployeeTrackRepositoryImpl
import com.luopingtech.ebike.ops.data.tools.UnlockedVehicleApi
import com.luopingtech.ebike.ops.data.tools.UnlockedVehicleRepository
import com.luopingtech.ebike.ops.data.tools.UnlockedVehicleRepositoryImpl
import com.luopingtech.ebike.ops.data.vehicle.VehicleApi
import com.luopingtech.ebike.ops.data.vehicle.VehicleRepository
import com.luopingtech.ebike.ops.data.vehicle.VehicleRepositoryImpl
import com.luopingtech.ebike.ops.data.media.FileUploadApi
import com.luopingtech.ebike.ops.data.movecar.BatchMoveCarApi
import com.luopingtech.ebike.ops.data.movecar.BatchMoveCarRepository
import com.luopingtech.ebike.ops.data.movecar.BatchMoveCarRepositoryImpl
import com.luopingtech.ebike.ops.data.movecar.FreeMoveCarApi
import com.luopingtech.ebike.ops.data.movecar.FreeMoveCarRepository
import com.luopingtech.ebike.ops.data.movecar.FreeMoveCarRepositoryImpl
import com.luopingtech.ebike.ops.data.staff.ServiceUserApi
import com.luopingtech.ebike.ops.data.staff.ServiceUserRepository
import com.luopingtech.ebike.ops.data.staff.ServiceUserRepositoryImpl
import com.luopingtech.ebike.ops.data.warehouse.WarehouseApi
import com.luopingtech.ebike.ops.data.warehouse.WarehouseRepository
import com.luopingtech.ebike.ops.data.warehouse.WarehouseRepositoryImpl
import com.luopingtech.ebike.ops.domain.control.NetworkVehicleControl
import com.luopingtech.ebike.ops.domain.control.VehicleControlPolicy
import com.luopingtech.ebike.ops.domain.permission.OpsPermissions
import com.luopingtech.ebike.ops.feature.auth.AuthFeature
import com.luopingtech.ebike.ops.feature.home.HomeFeature
import com.luopingtech.ebike.ops.feature.movecar.BatchMoveCarFeature
import com.luopingtech.ebike.ops.feature.movecar.FreeMoveCarFeature
import com.luopingtech.ebike.ops.feature.production.ProductionFeature
import com.luopingtech.ebike.ops.feature.sneak.SneakReportFeature
import com.luopingtech.ebike.ops.feature.report.FaultReportFeature
import com.luopingtech.ebike.ops.feature.relocation.RelocationFeature
import com.luopingtech.ebike.ops.feature.scan.ScanFeature
import com.luopingtech.ebike.ops.feature.workorder.WorkOrderFeature
import com.luopingtech.ebike.ops.domain.model.WorkOrderKind
import com.luopingtech.ebike.ops.feature.task.ChangeBatteryTaskFeature
import com.luopingtech.ebike.ops.feature.task.ClaimableTaskFeature
import com.luopingtech.ebike.ops.feature.task.InspectionTaskFeature
import com.luopingtech.ebike.ops.feature.task.MoveCarTaskFeature
import com.luopingtech.ebike.ops.feature.task.RepairTaskFeature
import com.luopingtech.ebike.ops.feature.task.TaskAuditFeature
import com.luopingtech.ebike.ops.feature.tenant.ServiceAreaFeature
import com.luopingtech.ebike.ops.feature.tools.FieldChangeBatteryFeature
import com.luopingtech.ebike.ops.feature.tools.UnlockedVehicleFeature
import com.luopingtech.ebike.ops.feature.tracking.TrackUploadFeature
import com.luopingtech.ebike.ops.data.fence.FenceApi
import com.luopingtech.ebike.ops.data.fence.FenceRepository
import com.luopingtech.ebike.ops.data.fence.FenceRepositoryImpl
import com.luopingtech.ebike.ops.data.geo.GeoApi
import com.luopingtech.ebike.ops.data.order.OrderApi
import com.luopingtech.ebike.ops.data.order.OrderRepository
import com.luopingtech.ebike.ops.data.order.OrderRepositoryImpl
import com.luopingtech.ebike.ops.feature.order.OrderQueryFeature
import com.luopingtech.ebike.ops.data.trajectory.TrajectoryApi
import com.luopingtech.ebike.ops.data.trajectory.TrajectoryRepository
import com.luopingtech.ebike.ops.data.trajectory.TrajectoryRepositoryImpl
import com.luopingtech.ebike.ops.feature.vehicle.VehicleDetailMapFeature
import com.luopingtech.ebike.ops.feature.vehicle.VehicleFeature
import com.luopingtech.ebike.ops.feature.analysis.VehicleConditionDistributionFeature
import com.luopingtech.ebike.ops.feature.analysis.OfflineOpsFeature
import com.luopingtech.ebike.ops.feature.analysis.StationAnalysisFeature
import com.luopingtech.ebike.ops.feature.analysis.ReturnCarAnalysisFeature
import com.luopingtech.ebike.ops.feature.analysis.TaskStatisticsFeature
import com.luopingtech.ebike.ops.feature.admin.BlacklistFeature
import com.luopingtech.ebike.ops.feature.admin.IdBindAuditFeature
import com.luopingtech.ebike.ops.feature.admin.ObjectionOrderFeature
import com.luopingtech.ebike.ops.feature.admin.OperationLogFeature
import com.luopingtech.ebike.ops.feature.admin.ProfessionAuditFeature
import com.luopingtech.ebike.ops.feature.fence.FenceBrowseFeature
import com.luopingtech.ebike.ops.feature.staff.StaffDirectoryFeature
import com.luopingtech.ebike.ops.feature.tag.VehicleTagFeature
import com.luopingtech.ebike.ops.feature.tools.BluetoothRadarFeature
import com.luopingtech.ebike.ops.feature.tools.OpsSettingFeature
import com.luopingtech.ebike.ops.feature.warehouse.WarehouseFeature
import com.luopingtech.ebike.ops.data.admin.AdminApi
import com.luopingtech.ebike.ops.data.admin.AdminRepository
import com.luopingtech.ebike.ops.data.admin.AdminRepositoryImpl
import com.luopingtech.ebike.ops.data.tag.VehicleTagApi
import com.luopingtech.ebike.ops.data.tag.VehicleTagRepository
import com.luopingtech.ebike.ops.data.tag.VehicleTagRepositoryImpl
import com.luopingtech.ebike.ops.data.tools.OpsSettingApi
import com.luopingtech.ebike.ops.data.tools.OpsSettingRepository
import com.luopingtech.ebike.ops.data.tools.OpsSettingRepositoryImpl
import com.luopingtech.ebike.ops.data.analysis.OfflineOpsApi
import com.luopingtech.ebike.ops.data.analysis.OfflineOpsRepository
import com.luopingtech.ebike.ops.data.analysis.OfflineOpsRepositoryImpl
import com.luopingtech.ebike.ops.data.analysis.StationAnalysisApi
import com.luopingtech.ebike.ops.data.analysis.StationAnalysisRepository
import com.luopingtech.ebike.ops.data.analysis.StationAnalysisRepositoryImpl
import com.luopingtech.ebike.ops.data.analysis.ReturnCarAnalysisApi
import com.luopingtech.ebike.ops.data.analysis.ReturnCarAnalysisRepository
import com.luopingtech.ebike.ops.data.analysis.ReturnCarAnalysisRepositoryImpl
import com.luopingtech.ebike.ops.data.analysis.TaskStatisticsRepository
import com.luopingtech.ebike.ops.data.analysis.TaskStatisticsRepositoryImpl
import com.luopingtech.ebike.ops.platform.BleTransport
import com.luopingtech.ebike.ops.platform.BleTransportFactory
import com.luopingtech.ebike.ops.platform.BindableMediaUploader
import com.luopingtech.ebike.ops.platform.CodeScanner
import com.luopingtech.ebike.ops.platform.DemoMediaUploader
import com.luopingtech.ebike.ops.platform.DemoReverseGeocoder
import com.luopingtech.ebike.ops.platform.DeviceInfo
import com.luopingtech.ebike.ops.platform.DisabledPushRegistrar
import com.luopingtech.ebike.ops.platform.LocationTracker
import com.luopingtech.ebike.ops.platform.MapCapability
import com.luopingtech.ebike.ops.platform.MapCapabilityFactory
import com.luopingtech.ebike.ops.platform.MediaUploader
import com.luopingtech.ebike.ops.platform.PushRegistrar
import com.luopingtech.ebike.ops.platform.ReverseGeocoder
import com.luopingtech.ebike.ops.platform.SecureStore
import com.luopingtech.ebike.ops.platform.SimulatorLocationTracker
import com.luopingtech.ebike.ops.platform.UnsupportedCodeScanner
import com.luopingtech.ebike.ops.platform.UnsupportedLocationTracker
import com.luopingtech.ebike.ops.platform.PhotoCapture
import com.luopingtech.ebike.ops.platform.UnsupportedPhotoCapture
import com.luopingtech.ebike.ops.platform.UnsupportedReverseGeocoder
import com.luopingtech.ebike.ops.platform.createDeviceInfo
import com.luopingtech.ebike.ops.platform.createSecureStore
import io.ktor.client.HttpClient
import kotlinx.coroutines.CoroutineScope
import kotlinx.serialization.json.Json

/**
 * Composition root for the shared layer. Host apps construct this once at startup.
 */
class OpsApp(
    val config: TenantConfig,
    val logger: OpsLogger = StdoutLogger,
    val secureStore: SecureStore = createSecureStore(),
    val deviceInfo: DeviceInfo = createDeviceInfo(appVersion = LIBRARY_VERSION),
    bleTransportOverride: BleTransport? = null,
    val locationTracker: LocationTracker = if (config.api.baseUrl.isBlank()) {
        SimulatorLocationTracker()
    } else {
        UnsupportedLocationTracker()
    },
    val reverseGeocoder: ReverseGeocoder = if (config.api.baseUrl.isBlank()) {
        DemoReverseGeocoder()
    } else {
        UnsupportedReverseGeocoder()
    },
    val codeScanner: CodeScanner = UnsupportedCodeScanner(),
    val photoCapture: PhotoCapture = UnsupportedPhotoCapture(),
    val mapCapability: MapCapability = MapCapabilityFactory.fromConfig(config),
    val pushRegistrar: PushRegistrar = DisabledPushRegistrar(),
    mediaUploader: MediaUploader? = null,
    networkVehicleControl: NetworkVehicleControl? = null,
    /** Override system language detection (e.g. Android Locale). */
    systemLanguage: String? = null,
) {
    val bleTransport: BleTransport = BleTransportFactory.fromConfig(config, bleTransportOverride)

    val i18n: OpsI18n = OpsI18n.fromStore(secureStore, systemLanguage)

    private val httpClientFactory = HttpClientFactory(config, logger)
    val httpClient: HttpClient = httpClientFactory.create()
    val json: Json get() = httpClientFactory.json
    val requestAuth: RequestAuth = RequestAuth(config, secureStore)

    private val demoMode: Boolean = config.api.baseUrl.isBlank()
    private val deviceIdProvider: () -> String = { AuthRepositoryImpl.deviceId(secureStore) }
    /** Always follow the selected business tenant; never freeze config.tenantId at construction. */
    private val tenantIdProvider: () -> String = {
        secureStore.getString(SecureStore.KEY_TENANT_ID)?.takeIf { it.isNotBlank() }
            ?: config.tenantId
    }

    val mediaUploader: MediaUploader = mediaUploader
        ?: if (demoMode) DemoMediaUploader() else BindableMediaUploader()

    private val authApi: AuthApi? = if (demoMode) {
        null
    } else {
        AuthApi(
            client = httpClient,
            requestAuth = requestAuth,
            json = httpClientFactory.json,
            tenantIdProvider = tenantIdProvider,
            signSecretProvider = { config.auth.signSecret },
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
            areaCodeProvider = {
                CallingCodeCatalog.find(
                    secureStore.getString(SecureStore.KEY_LOGIN_AREA_REGION),
                    secureStore.getString(SecureStore.KEY_LOGIN_AREA_CODE),
                ).dialCode
            },
        )
    }

    val authRepository: AuthRepository = AuthRepositoryImpl(
        secureStore = secureStore,
        demoMode = demoMode,
        authApi = authApi,
        requestAuth = requestAuth,
    )
    val authFeature: AuthFeature = AuthFeature(authRepository)

    private val signedApiClient: SignedApiClient? = if (demoMode) {
        null
    } else {
        SignedApiClient(
            client = httpClient,
            requestAuth = requestAuth,
            json = httpClientFactory.json,
            sessionProvider = {
                val session = authRepository.currentSession()
                NetworkSession(accessToken = session?.accessToken.orEmpty())
            },
            refreshAccessToken = {
                when (val result = authRepository.refreshIfNeeded()) {
                    is OpsResult.Ok -> OpsResult.Ok(result.value.accessToken)
                    is OpsResult.Err -> result
                }
            },
            onSessionInvalid = { _, message ->
                authRepository.markSessionInvalid(message)
            },
        )
    }

    private val serviceAreaApi: ServiceAreaApi? = signedApiClient?.let { client ->
        ServiceAreaApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }

    val serviceAreaRepository: ServiceAreaRepository = ServiceAreaRepositoryImpl(
        secureStore = secureStore,
        demoMode = demoMode,
        api = serviceAreaApi,
        json = httpClientFactory.json,
    )
    val serviceAreaFeature: ServiceAreaFeature = ServiceAreaFeature(serviceAreaRepository)

    private val vehicleApi: VehicleApi? = signedApiClient?.let { client ->
        VehicleApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }

    val vehicleRepository: VehicleRepository = VehicleRepositoryImpl(
        demoMode = demoMode,
        api = vehicleApi,
    )
    val vehicleFeature: VehicleFeature = VehicleFeature(vehicleRepository)
    val vehicleConditionDistributionFeature: VehicleConditionDistributionFeature =
        VehicleConditionDistributionFeature(vehicleRepository)

    private val offlineOpsApi: OfflineOpsApi? = signedApiClient?.let { client ->
        OfflineOpsApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }
    val offlineOpsRepository: OfflineOpsRepository = OfflineOpsRepositoryImpl(
        demoMode = demoMode,
        api = offlineOpsApi,
    )
    val offlineOpsFeature: OfflineOpsFeature = OfflineOpsFeature(offlineOpsRepository)

    private val stationAnalysisApi: StationAnalysisApi? = signedApiClient?.let { client ->
        StationAnalysisApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }
    val stationAnalysisRepository: StationAnalysisRepository = StationAnalysisRepositoryImpl(
        demoMode = demoMode,
        api = stationAnalysisApi,
    )
    val stationAnalysisFeature: StationAnalysisFeature = StationAnalysisFeature(stationAnalysisRepository)

    private val returnCarAnalysisApi: ReturnCarAnalysisApi? = signedApiClient?.let { client ->
        ReturnCarAnalysisApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }
    val returnCarAnalysisRepository: ReturnCarAnalysisRepository = ReturnCarAnalysisRepositoryImpl(
        demoMode = demoMode,
        api = returnCarAnalysisApi,
    )
    val returnCarAnalysisFeature: ReturnCarAnalysisFeature =
        ReturnCarAnalysisFeature(returnCarAnalysisRepository)

    private val opsSettingApi: OpsSettingApi? = signedApiClient?.let { client ->
        OpsSettingApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }
    val opsSettingRepository: OpsSettingRepository = OpsSettingRepositoryImpl(
        demoMode = demoMode,
        api = opsSettingApi,
    )
    val opsSettingFeature: OpsSettingFeature = OpsSettingFeature(opsSettingRepository)

    private val vehicleTagApi: VehicleTagApi? = signedApiClient?.let { client ->
        VehicleTagApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
            pinProvider = { authRepository.currentSession()?.userId.orEmpty() },
        )
    }
    val vehicleTagRepository: VehicleTagRepository = VehicleTagRepositoryImpl(
        demoMode = demoMode,
        api = vehicleTagApi,
    )
    val vehicleTagFeature: VehicleTagFeature = VehicleTagFeature(
        repository = vehicleTagRepository,
        vehicleRepository = vehicleRepository,
        qrHostsProvider = { authRepository.runtimeConfig()?.qrHosts.orEmpty() },
    )

    private val adminApi: AdminApi? = signedApiClient?.let { client ->
        AdminApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }
    val adminRepository: AdminRepository = AdminRepositoryImpl(
        demoMode = demoMode,
        api = adminApi,
    )
    val professionAuditFeature: ProfessionAuditFeature = ProfessionAuditFeature(adminRepository)
    val objectionOrderFeature: ObjectionOrderFeature = ObjectionOrderFeature(adminRepository)
    val blacklistFeature: BlacklistFeature = BlacklistFeature(adminRepository)
    val idBindAuditFeature: IdBindAuditFeature = IdBindAuditFeature(adminRepository)
    val operationLogFeature: OperationLogFeature = OperationLogFeature(adminRepository)

    private val fenceApi: FenceApi? = signedApiClient?.let { client ->
        FenceApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }
    val fenceRepository: FenceRepository = FenceRepositoryImpl(demoMode = demoMode, api = fenceApi)
    val fenceBrowseFeature: FenceBrowseFeature = FenceBrowseFeature(repository = fenceRepository)

    private val orderApi: OrderApi? = signedApiClient?.let { client ->
        OrderApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }
    val orderRepository: OrderRepository =
        OrderRepositoryImpl(demoMode = demoMode, api = orderApi)

    val orderQueryFeature: OrderQueryFeature = OrderQueryFeature(
        repository = orderRepository,
        permissionsProvider = { sessionPermissions() },
    )

    private val geoApi: GeoApi? = signedApiClient?.let { client ->
        GeoApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }

    private val trajectoryApi: TrajectoryApi? = signedApiClient?.let { client ->
        TrajectoryApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }
    val trajectoryRepository: TrajectoryRepository =
        TrajectoryRepositoryImpl(demoMode = demoMode, api = trajectoryApi)

    val vehicleDetailMapFeature: VehicleDetailMapFeature = VehicleDetailMapFeature(
        fenceRepository = fenceRepository,
        orderRepository = orderRepository,
        trajectoryRepository = trajectoryRepository,
        reverseGeocoder = reverseGeocoder,
        geoApi = geoApi,
        demoMode = demoMode,
    )

    val homeFeature: HomeFeature = HomeFeature(
        authFeature = authFeature,
        serviceAreaFeature = serviceAreaFeature,
        vehicleFeature = vehicleFeature,
        mapReady = mapCapability.isReady,
        mapProviderKind = mapCapability.kind,
    )

    private val resolvedNetworkControl: NetworkVehicleControl =
        networkVehicleControl
            ?: if (demoMode || signedApiClient == null) {
                DemoNetworkVehicleControl()
            } else {
                RemoteNetworkVehicleControl(
                    signedApi = signedApiClient,
                    tenantIdProvider = tenantIdProvider,
                    deviceInfo = deviceInfo,
                    deviceIdProvider = deviceIdProvider,
                )
            }

    val vehicleControl: VehicleControlPolicy =
        VehicleControlPolicy(ble = bleTransport, network = resolvedNetworkControl)

    val bluetoothRadarFeature: BluetoothRadarFeature = BluetoothRadarFeature(
        bleTransport = bleTransport,
        control = vehicleControl,
    )

    private val unlockedVehicleApi: UnlockedVehicleApi? = signedApiClient?.let { client ->
        UnlockedVehicleApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }

    val unlockedVehicleRepository: UnlockedVehicleRepository = UnlockedVehicleRepositoryImpl(
        demoMode = demoMode,
        api = unlockedVehicleApi,
    )

    val unlockedVehicleFeature: UnlockedVehicleFeature = UnlockedVehicleFeature(
        repository = unlockedVehicleRepository,
        control = vehicleControl,
        canFilterStaffProvider = {
            val codes = authRepository.currentSession()?.permissionCodes.orEmpty()
            val perms = if (demoMode && codes.isEmpty()) {
                OpsPermissions.demoFull()
            } else {
                OpsPermissions.fromCodes(codes)
            }
            perms.canFilterUnlockedStaff
        },
        selfPhoneProvider = { authRepository.currentSession()?.phone.orEmpty() },
    )

    private val relocationApi: RelocationApi? = signedApiClient?.let { client ->
        RelocationApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }
    val relocationRepository: RelocationRepository = RelocationRepositoryImpl(
        demoMode = demoMode,
        api = relocationApi,
    )
    val relocationFeature: RelocationFeature = RelocationFeature(
        repository = relocationRepository,
        locationTracker = locationTracker,
        qrHostsProvider = { authRepository.runtimeConfig()?.qrHosts.orEmpty() },
    )

    private fun sessionPermissions(): OpsPermissions {
        val codes = authRepository.currentSession()?.permissionCodes.orEmpty()
        return if (demoMode && codes.isEmpty()) {
            OpsPermissions.demoFull()
        } else {
            OpsPermissions.fromCodes(codes)
        }
    }

    private val inspectionOrderApi: InspectionOrderApi? = signedApiClient?.let { client ->
        InspectionOrderApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }
    val inspectionOrderRepository: InspectionOrderRepository = InspectionOrderRepositoryImpl(
        demoMode = demoMode,
        api = inspectionOrderApi,
    )
    val inspectionOrderFeature: WorkOrderFeature = WorkOrderFeature(
        kind = WorkOrderKind.Inspection,
        inspectionRepository = inspectionOrderRepository,
        canTakeProvider = { sessionPermissions().canTakeInspectionOrder },
    )

    private val repairOrderApi: RepairOrderApi? = signedApiClient?.let { client ->
        RepairOrderApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }
    val repairOrderRepository: RepairOrderRepository = RepairOrderRepositoryImpl(
        demoMode = demoMode,
        api = repairOrderApi,
    )
    val repairOrderFeature: WorkOrderFeature = WorkOrderFeature(
        kind = WorkOrderKind.Repair,
        repairRepository = repairOrderRepository,
        canTakeProvider = { sessionPermissions().canTakeRepairOrder },
    )

    val scanFeature: ScanFeature = ScanFeature(
        repository = vehicleRepository,
        control = vehicleControl,
        codeScanner = codeScanner,
        qrHostsProvider = { authRepository.runtimeConfig()?.qrHosts.orEmpty() },
        serviceAreaIdProvider = { serviceAreaRepository.currentArea()?.id.orEmpty() },
    )

    val fieldChangeBatteryFeature: FieldChangeBatteryFeature = FieldChangeBatteryFeature(
        repository = vehicleRepository,
        control = vehicleControl,
        qrHostsProvider = { authRepository.runtimeConfig()?.qrHosts.orEmpty() },
    )

    private val changeBatteryTaskApi: ChangeBatteryTaskApi? = signedApiClient?.let { client ->
        ChangeBatteryTaskApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }

    val changeBatteryTaskRepository: ChangeBatteryTaskRepository =
        ChangeBatteryTaskRepositoryImpl(
            demoMode = demoMode,
            api = changeBatteryTaskApi,
        )

    val changeBatteryTaskFeature: ChangeBatteryTaskFeature = ChangeBatteryTaskFeature(
        repository = changeBatteryTaskRepository,
        control = vehicleControl,
        pinProvider = { authRepository.currentSession()?.userId.orEmpty() },
    )

    private val moveCarTaskApi: MoveCarTaskApi? = signedApiClient?.let { client ->
        MoveCarTaskApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }

    val moveCarTaskRepository: MoveCarTaskRepository = MoveCarTaskRepositoryImpl(
        demoMode = demoMode,
        api = moveCarTaskApi,
    )

    val moveCarTaskFeature: MoveCarTaskFeature = MoveCarTaskFeature(
        repository = moveCarTaskRepository,
        control = vehicleControl,
        pinProvider = { authRepository.currentSession()?.userId.orEmpty() },
        locationTracker = locationTracker,
        mediaUploader = this.mediaUploader,
    )

    private val taskAuditApi: TaskAuditApi? = signedApiClient?.let { client ->
        TaskAuditApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }
    val taskAuditRepository: TaskAuditRepository = TaskAuditRepositoryImpl(
        demoMode = demoMode,
        api = taskAuditApi,
    )
    val taskAuditFeature: TaskAuditFeature = TaskAuditFeature(taskAuditRepository)

    private val freeMoveCarApi: FreeMoveCarApi? = signedApiClient?.let { client ->
        FreeMoveCarApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }

    val freeMoveCarRepository: FreeMoveCarRepository = FreeMoveCarRepositoryImpl(
        demoMode = demoMode,
        api = freeMoveCarApi,
    )

    private val serviceUserApi: ServiceUserApi? = signedApiClient?.let { client ->
        ServiceUserApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }

    val serviceUserRepository: ServiceUserRepository = ServiceUserRepositoryImpl(
        demoMode = demoMode,
        api = serviceUserApi,
    )
    val staffDirectoryFeature: StaffDirectoryFeature = StaffDirectoryFeature(serviceUserRepository)

    val freeMoveCarFeature: FreeMoveCarFeature = FreeMoveCarFeature(
        repository = freeMoveCarRepository,
        serviceUserRepository = serviceUserRepository,
        mediaUploader = this.mediaUploader,
        phoneProvider = { authRepository.currentSession()?.phone.orEmpty() },
    )

    private val batchMoveCarApi: BatchMoveCarApi? = signedApiClient?.let { client ->
        BatchMoveCarApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }

    val batchMoveCarRepository: BatchMoveCarRepository = BatchMoveCarRepositoryImpl(
        demoMode = demoMode,
        api = batchMoveCarApi,
    )

    val batchMoveCarFeature: BatchMoveCarFeature = BatchMoveCarFeature(
        repository = batchMoveCarRepository,
        mediaUploader = this.mediaUploader,
        pinProvider = { authRepository.currentSession()?.userId.orEmpty() },
    )

    private val inspectionTaskApi: InspectionTaskApi? = signedApiClient?.let { client ->
        InspectionTaskApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }

    val inspectionTaskRepository: InspectionTaskRepository = InspectionTaskRepositoryImpl(
        demoMode = demoMode,
        api = inspectionTaskApi,
    )

    val inspectionTaskFeature: ClaimableTaskFeature = InspectionTaskFeature(
        repository = inspectionTaskRepository,
        pinProvider = { authRepository.currentSession()?.userId.orEmpty() },
        mediaUploader = this.mediaUploader,
    )

    private val repairTaskApi: RepairTaskApi? = signedApiClient?.let { client ->
        RepairTaskApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }

    val repairTaskRepository: RepairTaskRepository = RepairTaskRepositoryImpl(
        demoMode = demoMode,
        api = repairTaskApi,
    )

    val repairTaskFeature: ClaimableTaskFeature = RepairTaskFeature(
        repository = repairTaskRepository,
        pinProvider = { authRepository.currentSession()?.userId.orEmpty() },
        mediaUploader = this.mediaUploader,
    )

    val taskStatisticsRepository: TaskStatisticsRepository = TaskStatisticsRepositoryImpl(
        demoMode = demoMode,
        changeBatteryApi = changeBatteryTaskApi,
        moveCarApi = moveCarTaskApi,
        inspectionApi = inspectionTaskApi,
        repairApi = repairTaskApi,
    )
    val taskStatisticsFeature: TaskStatisticsFeature = TaskStatisticsFeature(
        repository = taskStatisticsRepository,
        opPinProvider = { authRepository.currentSession()?.userId.orEmpty() },
    )

    private val warehouseApi: WarehouseApi? = signedApiClient?.let { client ->
        WarehouseApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }

    val warehouseRepository: WarehouseRepository = WarehouseRepositoryImpl(
        demoMode = demoMode,
        api = warehouseApi,
    )

    val warehouseFeature: WarehouseFeature = WarehouseFeature(
        repository = warehouseRepository,
        pinProvider = { authRepository.currentSession()?.userId.orEmpty() },
    )

    private val productionApi: ProductionApi? = signedApiClient?.let { client ->
        ProductionApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }

    val productionRepository: ProductionRepository = ProductionRepositoryImpl(
        demoMode = demoMode,
        api = productionApi,
    )

    val productionFeature: ProductionFeature = ProductionFeature(
        repository = productionRepository,
        control = vehicleControl,
        serviceAreaIdProvider = {
            serviceAreaRepository.currentArea()?.id
                ?: authRepository.currentSession()?.serviceAreaId.orEmpty()
        },
        vehicleRepository = vehicleRepository,
        bleAvailableProvider = { bleTransport.isAvailable },
    )

    private val faultReportApi: FaultReportApi? = signedApiClient?.let { client ->
        FaultReportApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }

    val faultReportRepository: FaultReportRepository = FaultReportRepositoryImpl(
        demoMode = demoMode,
        api = faultReportApi,
    )

    val faultReportFeature: FaultReportFeature = FaultReportFeature(
        repository = faultReportRepository,
        serviceAreaIdProvider = {
            serviceAreaRepository.currentArea()?.id
                ?: authRepository.currentSession()?.serviceAreaId.orEmpty()
        },
        vehicleRepository = vehicleRepository,
        mediaUploader = this.mediaUploader,
    )

    private val sneakReportApi: SneakReportApi? = signedApiClient?.let { client ->
        SneakReportApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }
    val sneakReportRepository: SneakReportRepository = SneakReportRepositoryImpl(
        demoMode = demoMode,
        api = sneakReportApi,
    )
    val sneakReportFeature: SneakReportFeature = SneakReportFeature(
        repository = sneakReportRepository,
        orderRepository = orderRepository,
        serviceAreaIdProvider = {
            serviceAreaRepository.currentArea()?.id
                ?: authRepository.currentSession()?.serviceAreaId.orEmpty()
        },
        reportManPinProvider = { authRepository.currentSession()?.userId.orEmpty() },
        reportManPhoneProvider = { authRepository.currentSession()?.phone.orEmpty() },
        mediaUploader = this.mediaUploader,
    )

    val fileUploadApi: FileUploadApi? = if (demoMode) {
        null
    } else {
        FileUploadApi(
            client = httpClient,
            requestAuth = requestAuth,
            json = httpClientFactory.json,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
            pinProvider = { authRepository.currentSession()?.userId.orEmpty() },
            sessionProvider = {
                NetworkSession(accessToken = authRepository.currentSession()?.accessToken.orEmpty())
            },
        )
    }

    private val employeeTrackApi: EmployeeTrackApi? = signedApiClient?.let { client ->
        EmployeeTrackApi(
            signedApi = client,
            tenantIdProvider = tenantIdProvider,
            deviceInfo = deviceInfo,
            deviceIdProvider = deviceIdProvider,
        )
    }

    val employeeTrackRepository: EmployeeTrackRepository = EmployeeTrackRepositoryImpl(
        demoMode = demoMode,
        api = employeeTrackApi,
    )

    val trackUploadFeature: TrackUploadFeature = TrackUploadFeature(
        repository = employeeTrackRepository,
        locationTracker = locationTracker,
        userPinProvider = { authRepository.currentSession()?.userId.orEmpty() },
        secureStore = secureStore,
    )

    val isDemoMode: Boolean get() = demoMode

    /** Resolve tenant-configured H5 dashboard URL (null if blank / disabled). */
    fun resolveH5ScreenUrl(kind: H5ScreenKind): String? =
        H5ScreenUrls.resolve(
            config = config,
            kind = kind,
            session = authFeature.state.value.session,
            deviceId = deviceIdProvider(),
            platform = deviceInfo.platform.ifBlank { "android" },
        )

    fun bind(scope: CoroutineScope) {
        authFeature.start(scope)
        homeFeature.start(scope)
        trackUploadFeature.bind(scope)
    }

    fun close() {
        httpClient.close()
    }

    companion object {
        const val LIBRARY_VERSION: String = "0.1.0-scaffold"

        fun demo(
            logger: OpsLogger = StdoutLogger,
            bleTransport: BleTransport? = null,
            secureStore: SecureStore = createSecureStore(),
            locationTracker: LocationTracker = SimulatorLocationTracker(),
        ): OpsApp = OpsApp(
            config = TenantConfig.demo(),
            logger = logger,
            bleTransportOverride = bleTransport,
            secureStore = secureStore,
            locationTracker = locationTracker,
        )

        fun create(
            config: TenantConfig,
            logger: OpsLogger = StdoutLogger,
            secureStore: SecureStore = createSecureStore(),
            deviceInfo: DeviceInfo = createDeviceInfo(appVersion = LIBRARY_VERSION),
            bleTransport: BleTransport? = null,
            codeScanner: CodeScanner = UnsupportedCodeScanner(),
            photoCapture: PhotoCapture = UnsupportedPhotoCapture(),
            locationTracker: LocationTracker? = null,
            reverseGeocoder: ReverseGeocoder? = null,
            systemLanguage: String? = null,
        ): OpsApp = OpsApp(
            config = config,
            logger = logger,
            secureStore = secureStore,
            deviceInfo = deviceInfo,
            bleTransportOverride = bleTransport,
            codeScanner = codeScanner,
            photoCapture = photoCapture,
            locationTracker = locationTracker
                ?: if (config.api.baseUrl.isBlank()) {
                    SimulatorLocationTracker()
                } else {
                    UnsupportedLocationTracker()
                },
            reverseGeocoder = reverseGeocoder
                ?: if (config.api.baseUrl.isBlank()) {
                    DemoReverseGeocoder()
                } else {
                    UnsupportedReverseGeocoder()
                },
            systemLanguage = systemLanguage,
        )
    }
}
