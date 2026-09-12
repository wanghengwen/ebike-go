# 模块处置清单

遗留 Android 工程 `APP/Merchant-Android/ManagerApp` 下全部 `*Activity.kt` 逐个归类的结果，
作为范围冻结（P0）的评审底稿。iOS 约 137 个 ViewController 与之大致一一对应，Flutter 约 20 个路由单列。

统计口径：`*Activity.kt` 共 119 个文件，排除 9 个基类与脚手架
（`BaseActivity`、`BaseToolbarActivity`、`BaseScanActivity`、`BaseCertificationActivity`、
`BaseCertificationDetailsActivity`、`BaseTaskDetailActivity`、`SwipeBackBaseActivity`、
`TestActivity`、`XFlutterActivity`），得到业务屏幕 **110** 个。

| 处置 | 数量 | 占比 |
|------|------|------|
| App 保留（现场作业） | 56 | 51% |
| 移出到 `pc` / 移动 H5 | 54 | 49% |

判定标准：需要**相机、蓝牙、现场定位或单手操作**的留在 App；
以**查询、审批、配置、报表**为主的移出。

---

## 一、保留在 App（56 屏）

### 账号 · 启动 · 租户切换（14）

`SplashActivity`、`LoginActivity`、`NewLoginActivity`、`VerificationCodeActivity`、
`ForgetPwdActivity`、`SetPwdActivity`、`UpdatePwdActivity`、`SelectBusinessActivity`、
`ChangeServiceAreaActivity`、`MainActivity`、`FunctionEntryActivity`、`SettingActivity`、
`UserAreaCodeActivity`、`UserLanguageSwitchingActivity`（语言切换已落在工作台设置：`OpsI18n` 中/英）

目标模块 `feature/auth`、`feature/tenant`。`LoginActivity` 与 `NewLoginActivity` 为新旧两版，合并为一个。

### 首页地图 · 车辆查找（8）

`VehicleDistributionMapActivity`、`VehicleLocationActivity`、`TaskMapActivity`、
`CarListActivity`、`SimpleCarListActivity`、`CarDetailActivity`、
`VehicleDetailInfoActivity`、`VehicleDetailNoteActivity`

目标模块 `feature/home`、`feature/vehicle`。三个地图页共用一套地图容器与聚合层；
`CarListActivity` 与 `SimpleCarListActivity` 合并为一个带模式参数的列表。

### 扫码 · 车控 · 蓝牙（6）

`ScanActivity`、`VehicleSimpleScanActivity`、`VehicleNumberUnlockActivity`、
`BluetoothIdentificationActivity`、`NewBluetoothRadarActivity`（真 BLE，闭源阻塞）、
`UnLockedVehiclesActivity`（已落地 `UnlockedVehicleFeature` / 网络关锁，权限 `1228`）

目标模块 `feature/scan`、`feature/control`。车控决策（蓝牙优先 → 网络兜底）下沉 `shared`，
两端只保留 SDK 适配。

### 任务执行（15）

换电与挪车：`ReplaceBatteryDescActivity`、`ReplaceBatteryViewActivity`、`MoveCarActivity`、
`MoveCarViewActivity`、`BatchMoveCarActivity`（已落地 `BatchMoveCarFeature` / man_made_move_list）、
`GroupMoveCarActivity`（真 BLE 扎堆，私有 SDK 未开放，见 `docs/BLE.md`）、`DragBackTaskActivity`

巡检与维修：`InspectionTaskDetailActivity`、`SingleInspectionOrderActivity`、`SingleRepairOrderActivity`

任务中心：`TaskCenterActivity`、`TaskDetailActivity`、`VehicleTaskActivity`、
`ChangeBatteryTaskDetailActivity`、`TaskAuditResultActivity`

目标模块 `feature/task`。**这是合并收益最大的一块**：四类任务在遗留工程里各有
「列表 + 详情 + 记录」三套实现，Flutter 侧还重复了一遍。统一为
「任务列表（类型参数化）+ 任务详情（步骤驱动）+ 完成结果」三个页面，15 屏可压到 5–6 屏。

### 仓库出入库（5）

`WarehouseMainActivity`、`WarehouseRecordActivity`、`InOrOutWarehouseWithQrCodeActivity`、
`InOrOutWarehouseWithoutQrCodeActivity`、`InOrOutWarehouseDetailsActivity`

目标模块 `feature/warehouse`。有码与无码出入库合并为一个页面的两种输入模式。

### 生产现场作业（3）

`VehicleDetectActivity`、`CenterControlBindActivity`、`VehiclePutPullShelvesDetailActivity`

目标模块 `feature/production`。依赖扫码与蓝牙，属产线现场操作，不能移出。

### 报修上报（5）

`FaultReportActivity`、`VehicleRepairActivity`、`MyReportActivity`、
`ReportActivity`、`ReportDetailActivity`

目标模块 `feature/report`。只保留**上报侧**，审核侧移出。

---

## 二、移出到 `pc` / 移动 H5（54 屏）

> **已复核（对照 `D:\workspace\ebike\pc`，Vue 2.6 + Ant Design Vue + ECharts 5）**：
> 本节各处「`pc` 已有 xxx」的说法逐条查过，基本属实，两处除外——
> `vehicle/vehicleHeatmap` 源码在但 `router.config.js` 里整段被注释；
> `powerChangeStatistics` / `repairStatistics` / `inspectionStatistics` / `moveCarStatistics`
> 只有 `locales` 菜单文案，`views` 下没有对应页面。
>
> **是否改为迁进 `webH5`：否（除挪车员分析 / 智能调度待定）。** 理由见 [四、H5 处置](#四h5-处置)。

### 订单 · 资金 · 会员（11）

`OrderSearchActivity`、`OrderHistoryActivity`、`MoreOrderActivity`、`ModifyAmountActivity`、
`SearchOrderCarDetailActivity`、`SearchOrderUserDetailActivity`、`UserListInfoActivity`、
`DepositMembershipActivity`、`RideCardActivity`、`WalletBalanceActivity`、`GuaranteeDetailActivity`

`pc` 已有 `order/orderInfo`、`user/ridingOrder`、`user/userOrder`、`user/rechargeOrder`、
`user/ridingCardOrder`、`report/*`。改价与退款属高风险操作，在 PC 上做更合适。

### 审核 · 认证 · 异议工单 · 拍照审核（12）

`CertificationAuditActivity`、`CertificationAuditDetailsActivity`、
`ChangeBindCertificationActivity`、`ChangeBindDetailsActivity`、
`ObjectionOrderActivity`、`ObjectionOrderDetailActivity`、`ObjectionOrderProcessActivity`、
`ReportReviewActivity`、`ReportReviewDetailsActivity`、`ReportReviewListActivity`、
`TakePhotosReviewActivity`、`PhotographAuditActivity`

`pc` 已有 `user/professionCert`、`user/changeBinding`、`user/userWorkSheet`。

> **已落地**：`TakePhotosReviewActivity` / `PhotographAuditActivity` 的拍照 + **备注（remark）**
> 已接到挪车 / 自主挪车 / 批量挪车 / 巡检 / 维修的 finish 路径（`needPhotograph` 时必填）。
> 自主挪车拍照审核额外支持可选 **协同人**（`listByServiceIds` → finish `teamWorker`，与旧版 PHOTOGRAPH_MOVE_CAR 一致）。
> 若由现场主管在车旁复核，则应保留在 App；若由后台审核岗处理，则移出。P0 需业务方明确，默认按移出。

### 员工 · 角色 · 权限配置（5）

`AddEmployeeActivity`、`EmployeeAccountActivity`、`RoleTableActivity`、
`PermissionConfigActivity`、`HomePermissionActivity`

`pc` 已有 `permissions` 模块。

### 轨迹查看 · 员工操作记录（3）

`EmployeeTrackActivity`、`RealTimeTrajectoryActivity`、`EmployeeOperateLogActivity`

轨迹**上报**是 App 的后台能力并保留；轨迹**查看**是管理视角，移出。

> **已落地（Android）**：`TrackUploadFeature` + `AndroidLocationTracker`；开启时拉起
> `TrackLocationService`（`foregroundServiceType=location`），退到后台仍可采点上报。
> Demo 模式继续用 `SimulatorLocationTracker`，不启 Service。

> **待定项**：`RealTimeTrajectoryActivity` 若用于现场调度时查看附近同事位置，可考虑保留精简版。

### 操作日志 · 追溯（5）

`OperationLogActivity`、`OperationLogDetailActivity`、`OperationLogListActivity`、
`VehicleScanLogActivity`、`VehicleSwitchLockLogActivity`

`pc` 已有 `operationLog`。

### 运维阈值 · 车辆标签 · 系统配置（4）

`OperationSettingActivity`、`OperationThresholdActivity`、
`ThresholdSettingActivity`、`VehicleTagActivity`

`pc` 已有 `maintain/taskConfig`、`maintain/carTags`、`operating/systemConfig`。
配置类功能低频且字段多，手机端体验本就不好。

### 围栏编辑（2）

`FenceActivity`、`FenceEditActivity`

`pc` 已有 `operating/fences`、`fenceParking`、`fenceForbidArea`、`fenceServiceArea`。
多边形绘制在大屏上体验明显更好。蓝牙道钉相关的现场标定若确有需要，单独做一个轻量工具页，不搬整个编辑器。

### 统计分析（变更：数据分析四卡改 CMP 重写）

`MaintenanceStatisticsActivity` 仍可交 `pc`（运维统计报表）。

**底栏「数据分析」四卡**（权限 `1217` / `1240` / `1243` / `1251`）**改判：留在 App，用
Compose Multiplatform 重写**，不再交 `pc` / `webH5`：

| 子功能 | 权限 | 遗留形态 | CMP 落点 |
|--------|------|----------|----------|
| 线下运维 | `1217` | Flutter `/offline_operation` | `sharedUi` 看板（Canvas 折线/饼图，**已落地**） |
| 站点监控 | `1240` | Flutter `/site_analysis` + `/echart` | `sharedUi` 列表 + 详情图（Canvas，**已落地**） |
| 还车分布 | `1243` | 原生 `ReturnCarAnalysisActivity` | **宿主地图** + DotScatter 散点（**已落地**） |
| 车况分布 | `1251` | Flutter 列表 + `VehicleDistributionMapActivity` | `sharedUi` 分档列表 + 宿主聚合地图（**优先落地**） |

说明：列表/看板进 `sharedUi`；地图（散点 / 聚合 / 围栏）留 Android/iOS 宿主，与首页地图同一套腾讯封装。
`pc` 的 `maintain/taskStatistics`、`stationMonitorList` 等可继续服务桌面，但**不再当作 App 入口的替代**；
站点监控 App 与 PC 口径本就不一致，见下文 H5 表。

### 黑名单（2）

`BlackListActivity`、`AddBlackListActivity`

`pc` 已有 `user/blackList`。

### 自动报修配置（2）

`AutoRepairActivity`、`AutoRepairDetailsActivity`

`pc` 已有 `operating/repairConfig`。

### 数据删除向导（6）

`DDStepOneActivity`、`DDStepTwoActivity`、`DDEnterPasswordActivity`、
`DDSelectServiceActivity`、`DDHistoryRecordActivity`、`DDeleteSuccessActivity`

低频合规功能，六步向导，无必要占用 App 体积与维护成本。

---

## 三、Flutter 路由处置（约 20）

| 路由 | 处置 |
|------|------|
| `/task_list_change_battery`、`/task_list_move_car`、`/task_list_repair`、`/task_list_inspection` | 迁入 `feature/task`，合并为参数化列表 |
| `/task_mine_task`、`/*_task_fetch`、`/*_task_detail`、`/*_task_Record` | 迁入 `feature/task` |
| `/move_car_scan`、`/move_car_input`、`/relocation_scan`、`/relocation_map` | 迁入 `feature/task`、`feature/scan` |
| `/warehouse_record_list`、`/warehouse_record_detail` | 迁入 `feature/warehouse` |
| `/vehicle_condition_distribution` | **迁入** `feature/analysis`（CMP：分档列表 + 宿主地图；接口仍 `/business/paas/device/list`） |
| `/offline_operation`、`/site_analysis`、`/echart` | **迁入** `feature/analysis`（CMP 看板；排期见 `MIGRATION.md` P4） |
| `/move_bike_analysis`、`/intelligent_scheduling` | 移出到 `pc`（或另立项；PC 人员维度挪车分析仍有真空） |
| `/car_audit_list`、`/car_audit_detail` | 移出到 `pc` |
| `/station_edit`、`/fence_detail` | 移出到 `pc`（围栏编辑） |
| `/overload` | P0 确认归属 |

---

## 四、H5 处置

`APP/Merchant-H5`（运营与营收大屏）**源码迁入本项目 `webH5/`**，并从 Vue 2.6 + Vant 2 + ECharts 4 + vue-cli
重写为 Vue 3 + TypeScript + Vant 4 + ECharts 6 + Vite，技术栈与 `ebike-UniApp` 对齐。

> 本条替代了原先「源码不迁入、继续由服务器托管」的判断。改判的理由是：H5 与 App 之间靠一张 URL query
> 参数表耦合（见 `webH5/README.md`），放在同一仓库才能让「改协议时两边一起改」落在一个 PR 里。

- 部署形态不变：仍是独立的静态站点，**不参与 Gradle 构建**
- OpsApp 侧不变：`shared/.../core/config/H5ScreenUrls.kt` + `androidApp/.../ui/h5/H5Screen.kt` 用 WebView 打开，
  地址来自租户配置 `h5.operationUrl` / `h5.revenueUrl`（空则隐藏入口）
- 权限码与旧版一致：`0221`（运营）、`LargeRevenueScreen`（营收）
- 遗留 `common/api.js` 中约 40 个标注「未调用」的 `/mieba/v2/big_screen/*` 路径已在迁移中删除
- 遗留的加盟商选择弹窗（`xcAgent.vue`）、`v-hasCode` 指令、6 个从未被模板引用的图表组件确认为死代码，未迁入；
  完整清单见 `webH5/README.md`

遗留 `APP/Merchant-H5/` 目录在新 H5 上线并回归通过后删除。

### 第二节的 54 屏是否改迁进 `webH5`

复核结论：**不迁**，`webH5` 维持「两块大屏」的定位。三条硬约束：

1. **鉴权模型不匹配。** `webH5` 没有登录页，token 由 App 通过 URL query 注入
   （见 `webH5/README.md` 的契约表）。审批、权限配置、黑名单、数据删除这类管理端功能
   需要独立登录才成立，做下去等于再造一个 `pc`。
2. **设备能力做不到。** 车辆标签依赖扫码（`VehicleTagActivity` 继承 `BaseScanActivity`）、
   拍照审核依赖相机 + Aliyun 上传、围栏编辑依赖地图 SDK + 定位、
   订单与用户详情里还有 `switchLockCar` / `endTrip` 车控要取当前位置。
3. **不是零 Kotlin 改动。** `H5ScreenUrls.kt` 的 `H5ScreenKind` 只有 `Operation` / `Revenue`
   两个值，每加一个 H5 页都要动枚举、租户配置 `h5.*Url`、入口菜单和权限码。

四个需要单独记的判断：

| 对象 | 判断 |
|------|------|
| Flutter `/vehicle_condition_distribution` | **改判：App 内 CMP 重写。** 与运营大屏虽同用 `/business/paas/device/list`，但分档口径不同（本页电量 10 档 / 闲置 7 档；大屏为粗分档）。列表 +「点某档 → 地图看车」整页迁入 OpsApp，不再依赖 Flutter MethodChannel，也不塞进 `webH5` |
| Flutter `/move_bike_analysis`、`/intelligent_scheduling` | **`pc` 唯一真实缺口，待定。** 接口 `dispatch/business/workman/{moving_statistics,moving_analysis,moved_user}` 在 `pc` 中搜不到；`pc` 的挪车分析走 `move_car/task/analyze_data`，是任务维度而非人员维度。两页纯查询、无设备依赖、无地图，正好落在 `webH5` 现有能力上 |
| Flutter `/echart`、`/site_analysis` | **改判：App 内 CMP 重写**（与数据分析四卡策略一致）。`pc` 的站点页口径不同，不能当 App 替代；图表栈选定前可暂占位 |
| Flutter `/offline_operation` | **改判：App 内 CMP 重写**。`pc` `maintain/taskStatistics` 接口同源，桌面可继续用；App 入口走 Kotlin 看板，避免「列表 Kotlin / 看板 H5」割裂 |

两处 `pc` 真空区，同样不放 `webH5`：

- **轨迹查看 3 屏**（`EmployeeTrackActivity` / `RealTimeTrajectoryActivity` / `EmployeeOperateLogActivity`）——
  `pc` 的 `router.config.js` 里确无轨迹页，但画 polyline 需要地图 SDK，
  `webH5/src/components/charts/echartsCore.ts` 只注册了 bar/line/pie/tree。
- **数据删除向导 6 屏**——`pc` 也没有，但这是带密码确认的不可逆合规操作，
  不应放进 token 明文走 URL 的 H5。应补到 `pc`。
