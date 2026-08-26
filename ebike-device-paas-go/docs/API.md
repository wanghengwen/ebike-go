# ebike-device-paas-go 对外接口与服务

设备域 PaaS 聚合层：对上游（订单、运营、商家端、调度等）提供车辆查询、中控指令、蓝牙回调、状态变更与轨迹代理。

| 项 | 值 |
|---|---|
| Nacos 服务名 | `ebike-device-paas` |
| 默认端口 | `8080` |
| 协议 | HTTP JSON，业务接口一律 **POST** |
| 响应 HTTP 状态 | 业务失败也返回 **200**，以 `Result.code` 区分 |
| Feign 调用方 | `@FeignClient(name = "ebike-device-paas")` |

业务接口合计 **86** 个，另加 **3** 个运维探活/注销接口。

---

## 1. 对外服务能力

上游通过 7 个 Feign API 消费本服务：

| Feign 接口 | contextId | 职责 | 接口数 |
|---|---|---|---|
| `DeviceInfoApi` | DeviceInfo | 车辆缓存查询、列表/分页、大屏与运营统计、区间筛选 | 32 |
| `EcuQueryApi` | EcuQuery | 中控实时查询（内部参数 / 设备信息 / 道钉 / 头盔 / 电量 / GPS） | 6 |
| `EcuCommandApi` | EcuCommand | 中控指令下发（锁车、设防、升级、语音等） | 18 |
| `EcuJobQueryApi` | EcuJobQuery | 异步指令 job 结果查询 | 13 |
| `DeviceTrajectoryApi` | DeviceTrajectory | 实时/历史轨迹、距离、监控指标、轨迹落库 | 6 |
| `StateChangeApi` | StateChange | 骑行/告警/定位/扫码位置写入 | 4 |
| `EcuBleCommandReportApi` | EcuBleCommandReport | App 蓝牙控制结果回调 | 7 |

本服务**不对外暴露 gRPC / MQ 消费接口**。对外只有上述 HTTP；Kafka 仅作为本服务向下游投递 C34 快照的出口（见第 6 节）。

---

## 2. 调用约定

### 2.1 统一响应包络

```json
{
  "success": true,
  "code": "0",
  "msg": "成功",
  "data": {}
}
```

`data` 始终输出（失败时可为 `null`）。指令类失败仍可能带 `CommandResult`。

### 2.2 公共请求字段 `commandContext`

绝大多数请求体嵌套：

```json
{
  "commandContext": {
    "traceId": "…",
    "tenantId": "…",
    "pin": "…",
    "userId": "…",
    "platform": "wechat"
  }
}
```

| 字段 | 说明 |
|---|---|
| `tenantId` | 租户；可仅放 Header，服务会回填到 context |
| `traceId` | 链路追踪，会注入下游网关请求 |
| `platform` | `ios` / `android` / `wechat` / `pc` / `other`；微信寻车铃限流仅 `wechat` 生效 |
| `pin` / `userId` | 操作人，用于审计与限流 |

### 2.3 指令结果 `CommandResult`

```json
{
  "jobId": null,
  "async": false,
  "ecuCode": "0",
  "payload": null,
  "result": {}
}
```

`ecuCode` 为 `null` 或 `"0"` 视为成功。部分失败码：

| ecuCode | 对外 code | 含义 |
|---|---|---|
| `132` | `17002` | 车辆移动中，不能还车 |
| `138` | `17002` | 操作过于频繁 |
| 其它非 0 | `17002` | ECU execution failed |

EcuQuery 在线查询失败对外码为 `17003`（与 JobQuery/Command 的 `17002` 不同）。

### 2.4 业务错误码

| code | 含义 |
|---|---|
| `0` | 成功 |
| `00002` | 请求体不可解析 |
| `00026` | 限流（微信寻车铃） |
| `17002` | ECU 执行失败 |
| `17003` | ECU 信息查询失败 |
| `17004` | imei/carId 为空 |
| `17005` | imei/carId 绑定错误 |
| `17006` | 对应设备不存在 |
| `17012` | 设备离线 / 网关错误 |
| `17013` | 设备未返回 |
| `17014` | 订单轨迹不存在 |
| `17015` | 轨迹时间范围非法 |
| `17016` | 生成假点数量超限 |
| `17017` | 生成假点锁占用 |
| `170018` | 车辆未上架 |

---

## 3. 接口清单

以下路径均为 `POST`，Content-Type `application/json`。`data` 列为 `Result.data` 的类型。

### 3.1 DeviceInfoApi — 车辆查询与统计

数据主要来自 Redis 设备影子（DeviceProtocol 定长串 + GEO/ZSET）。只读接口可 shadow；标 **写** 的除外。

#### 基础查询

| 路径 | 说明 | 主要入参 | data |
|---|---|---|---|
| `/device/paas/device/detail` | 车辆详情（最热接口） | `imei` 或 `carId` | `DeviceDetailCo`（约 75 字段：位置、电门/设防、头盔、告警、电量等） |
| `/device/paas/device/list` | 按服务区 + 上报时间筛车辆 | `serviceIdList`, `reportTime`, 可选 `imeiList`/`carIdList` | `DeviceListCo[]` |
| `/device/paas/device/listByImeiList` | 按 imei 批量查 | `imeiList` | `DeviceListCo[]` |
| `/device/paas/device/listByCarIdList` | 按 carId 批量查 | `carIdList` | `DeviceListCo[]` |
| `/device/paas/device/filterList` | 带条件列表（不分页） | `serviceId` + 骑行/运维/告警/电量/头盔等过滤 | `DevicePageCo[]` |
| `/device/paas/device/page` | 平台端分页（含高德逆地理地址） | 同 filterList + `pageNum`/`pageSize` | `PageDTO<DevicePageCo>` |
| `/device/paas/device/pageBus` | 商家端分页（多状态并集/交集） | `serviceId`, `ridingStates[]`, `izStateAnd`, `batteryIds` 等 | `PageDTO<DevicePageCo>` |
| `/device/paas/device/eBikeLocation` | 附近可用车（GEO 半径） | `serviceId`, `lat`, `lng`, `radius`, `limit`(默认 20) | `DeviceLocationCo[]` |
| `/device/paas/device/queryCarImeiBind` | carId ↔ imei 绑定 | `carIdList` 或 `imeiList` | `CarImeiCo[]` |
| `/device/paas/device/queryDeviceCacheGps` | 缓存经纬度 | `imeiList` | `DeviceGpsCo[]` |
| `/device/paas/getDeviceRealGpsList` | 批量实时 GPS（15s 新鲜度） | `imeiList` | `DeviceRealGpsCo[]` |

#### 大屏 / 运营地图 / 统计

| 路径 | 说明 | 主要入参 | data |
|---|---|---|---|
| `/device/paas/device/queryDeviceScreen` | 大屏 V1：骑行计数 + 空闲/电量分桶 + GPS | `serviceIdList`, 可选 `ridingState` | `DeviceScreenCo` |
| `/device/paas/device/queryDeviceScreen/v2` | 大屏 V2（分桶口径不同，含 total/offline） | 同上 | `DeviceScreenCoV2` |
| `/device/paas/device/car_count` | 按服务区车辆数，按 total 降序 | `serviceIdList` | `DeviceScreenCoV2[]` |
| `/device/paas/device/queryDeviceOpeMap` | 运营地图聚合（运维/告警/电量阈值 + GPS） | `serviceIdList` + 过滤条件 | `DeviceOpeMapCo` |
| `/device/paas/device/queryServiceStatisticsList` | 按服务区状态统计 | `ids`（服务区 id 列表，保序） | `CarStateServiceStatisticsCo[]` |
| `/device/paas/device/getCarNumByServiceId` | 各服务区上架车辆数 | `serviceIdList` | `ServiceCarNumCo[]` |
| `/device/paas/device/carStatisticsByService` | 服务区 + 停车点级统计 | `serviceIdList` | `CarStatisticsByServiceCo` |
| `/device/paas/device/onlineNumByTenantId` | 当前租户上架车辆总数 | 仅 `commandContext` | `Integer` |
| `/device/paas/device/carStatistics` | 跨租户按服务区统计（约 1h 调一次） | 仅 `commandContext` | `CarServiceStatisticsCo[]` |
| `/device/paas/device/getRackCarNumAll` | 跨租户上架数按服务区 | 仅 `commandContext` | `RackCarNumCo[]` |
| `/device/paas/device/carParkingStatistics` | 跨租户停车点统计（约 6h 调一次） | 仅 `commandContext` | `CarParkingStatisticsCo[]` |

#### 区间筛选与零散查询

| 路径 | 说明 | 主要入参 | data |
|---|---|---|---|
| `/device/paas/device/queryDeviceByBattery` | 按剩余电量区间 | `serviceId`, `minRestBattery`, `maxRestBattery` | `DeviceByBatteryCo[]` |
| `/device/paas/device/queryDeviceByTotalMiles` | 按里程区间 | `serviceId`, `minTotalMiles`, `maxTotalMiles` | `DeviceByTotalMilesCo[]` |
| `/device/paas/device/queryDeviceByNoOrderTime` | 按无单时长 | `serviceId`, `minNoOrderTime`, `maxNoOrderTime` | `DeviceByNoOrderTimeCo[]` |
| `/device/paas/device/queryDeviceByStaticTime` | 按静置时长 | `serviceId`, `minStaticTime`, `maxStaticTime` | `DeviceByStaticTimeCo[]` |
| `/device/paas/device/getBlueToothToken` | 蓝牙控制 token | `imei` | `BlueToothTokenCo`（缺省 `168428805`） |
| `/device/paas/device/querySaddleOverloadContact` | 鞍座超载触点 | `imei` | `SaddleOverloadContactCo` |
| `/device/paas/device/queryCameraState` | 摄像头状态 | `imei` | `CameraCacheCo`（缺失为 null） |
| `/device/paas/device/queryDeviceMapFake` | 查询地图假点 | `serviceId` | `DeviceMapFakeCo` |
| `/device/paas/device/oneClickReturnNotify` | 一键还车失败状态 | `carId` | `PartAnalysisResultCO` |

#### 写操作

| 路径 | 说明 | 主要入参 | data |
|---|---|---|---|
| `/device/paas/device/removeDeviceTotalMiles` | **写** 清空里程 ZSET | `imei` | `Integer` |
| `/device/paas/device/genDeviceMapFake` | **写** 生成地图假点（数量上限 5×10⁵，3s 锁） | `serviceId`, `amount`, `fakeAmount` | `Integer` |

---

### 3.2 EcuQueryApi — 中控实时查询

除 `getDeviceRealGpsList` 读 Redis 外，其余转发 `ebike-device-openapi`（会真正下发设备查询命令）。

| 路径 | 说明 | 主要入参 | data |
|---|---|---|---|
| `/device/paas/innerParam` | 查询内部参数 | `imei`/`carId`, 可选 `async`/`payload` | `CommandResult<InnerParamQryCo>` |
| `/device/paas/deviceInfo` | 查询设备基础信息（缺省 payload `{"type":"c34"}`） | 同上 + `isCameraEnable` | `CommandResult<DeviceInfoCo>` |
| `/device/paas/blueTBeacon` | 蓝牙道钉信息 | 同上 | `CommandResult<BlueTBeaconInfoCo>` |
| `/device/paas/queryBleHelmetInfo` | 蓝牙智能头盔信息 | 同上 | `CommandResult<BleHelmetInfoCo>` |
| `/device/paas/queryRealRestBattery` | 实时电量（Redis soc 快路径，过期则算电量） | `imei`/`carId` | `RealRestBatteryCo` |

---

### 3.3 EcuCommandApi — 中控指令下发

全部转发 openapi，并可能写 Redis（风控标记、临时骑行标签等）。请求体透传指令字段，剔除 `commandContext` 后注入 `traceId`。

| 路径 | 说明 | 关键字段 | 响应策略 |
|---|---|---|---|
| `/device/paas/lock` | 电门开关 | `acc`, `izRiskControl`(缺省 true) | `toDeviceResult`（17002） |
| `/device/paas/helmetLock` | 头盔锁 | `sw`（显式为 0 时 idx=22） | 同上 |
| `/device/paas/rearWheelLock` | 后轮锁 | `sw` | 同上 |
| `/device/paas/defend` | 设防 | `defend` | 同上 |
| `/device/paas/setInnerParam` | 设置内部参数 | `freqMove`/`freqNorm` → snake_case | 同上 |
| `/device/paas/updateInnerFence` | 更新内置围栏 | 围栏参数 | 同上 |
| `/device/paas/deviceUpgrade` | 固件升级 | 升级包信息 | 同上 |
| `/device/paas/voiceUpgrade` | 语音包升级 | 升级包信息 | 同上 |
| `/device/paas/voice` | 语音命令（微信 idx=9 寻车铃限流） | `idx`；租户 250 且 idx∈{32,34} 且 ecuCode=105 视为成功 | 同上 |
| `/device/paas/batteryCompartment` | 电池仓锁 | `sw` | 同上 |
| `/device/paas/bluetooth` | 设置蓝牙参数 | 蓝牙参数 | 同上 |
| `/device/paas/scanBleHelmet` | 扫描绑定蓝牙头盔 | — | 同上 |
| `/device/paas/restart` | 设备重启 | — | 同上 |
| `/device/paas/dashboard` | 下发仪表盘 | — | 同上 |
| `/device/paas/transmission` | 透传 | — | **始终 success** |
| `/device/paas/tempUnLock` | 临时通电（关围栏；`izAccOn` 才下 lock） | `izAccOn` 等 | **始终 success** |
| `/device/paas/replyStopMove` | 一键还车停车站点回复 | 站点信息 | **始终 success** |
| `/device/paas/triggerTempState` | 临时状态触发（先清鞍座触点） | — | **始终 success** |

`voice` 微信寻车铃：5 分钟窗口累计 >30 次禁用 24h，单分钟 ≥10 次直接限流，对外 `00026`。

---

### 3.4 EcuJobQueryApi — 异步 job 结果

统一入参 `{ "jobId": "…", "commandContext": {} }`，转发 `GET /ebike/cmd/get_job_result?jobId=`。

| 路径 | 对应指令 | data |
|---|---|---|
| `/device/paas/jobQuery/lock` | 电门 | `CommandResult` |
| `/device/paas/jobQuery/defend` | 设防 | `CommandResult<DefendCo>` |
| `/device/paas/jobQuery/setInnerParam` | 内部参数 | `CommandResult` |
| `/device/paas/jobQuery/updateInnerFence` | 内置围栏 | `CommandResult` |
| `/device/paas/jobQuery/voiceUpgrade` | 语音升级 | `CommandResult` |
| `/device/paas/jobQuery/deviceUpgrade` | 固件升级 | `CommandResult` |
| `/device/paas/jobQuery/voice` | 语音 | `CommandResult` |
| `/device/paas/jobQuery/batteryCompartment` | 电池仓 | `CommandResult` |
| `/device/paas/jobQuery/bluetooth` | 蓝牙 | `CommandResult<BluetoothCo>` |
| `/device/paas/jobQuery/innerParam` | 查内部参数 | `CommandResult<InnerParamQryCo>` |
| `/device/paas/jobQuery/deviceInfo` | 查设备信息 | `CommandResult<DeviceInfoCo>` |
| `/device/paas/jobQuery/blueTBeacon` | 查道钉 | `CommandResult<BlueTBeaconInfoCo>` |
| `/device/paas/jobQuery/getJobSuc` | 是否成功 | `JobSucCo`（`{ "jobSuc": true/false }`） |

失败码映射同 Command（`17002`），不是 Query 的 `17003`。

---

### 3.5 DeviceTrajectoryApi — 轨迹与监控

转发 `ebike-device-worker`，请求自动带 `type=coordinate.type` 与 `traceId`。

| 路径 | 说明 | 主要入参 | data |
|---|---|---|---|
| `/device/trajectory/realTime` | 实时轨迹 | `imei`, `startTime`, `endTime` | `DeviceTrajectoryCo[]` |
| `/device/trajectory/history` | 订单历史轨迹 | `orderId`, `startTime`, `endTime` | `DeviceTrajectoryCo[]` |
| `/device/trajectory/history/batch` | 批量历史轨迹 | `orderBatch[]` | `Map<orderId, DeviceTrajectoryCo[]>` |
| `/device/trajectory/distance` | 轨迹距离（异常降级为 0） | 同 realTime | `TrajectoryDistanceCo` |
| `/device/trajectory/getMetric` | 监控指标（航向/速度/GPS/GSM/电压） | `imei`(可为 carId), `startTime`, `endTime`, `type` | `Map`（course/speed/gps/gsm/voltage） |
| `/device/trajectory/saveDb` | **写** 轨迹落库 | `orderId`, `imei`, `startTime`, `endTime` | `Boolean` |

`getMetric`：非 `86` 前缀按 carId 解析 imei；imei 须 `86` 开头且长度 15，否则 `17005`。

---

### 3.6 StateChangeApi — 设备状态写入

写 Redis（`SETRANGE` 局部更新 device_info + 刷新服务区 ZSET）或投 Kafka C34。成功时 `ridingState`/`alarmState` 的 data 为 `null`，定位类为 `1`。

| 路径 | 说明 | 主要入参 | 副作用 |
|---|---|---|---|
| `/device/ridingState/change` | 骑行/运维状态 | `carId`, `imei`, `serviceId`, `ridingState`, `operationState[]` | Redis 局部写 + 刷新 `service_gfence` |
| `/device/alarmState/change` | 告警状态 | `carId`, `imei`, `serviceId`, `alarmType[]` | 同上 |
| `/device/location/change` | 批量改定位 | `imeiList`, `lng`, `lat` | Kafka C34（GPS，GCJ02→WGS84） |
| `/device/scanLocation/change` | 最后扫码位置 | `carId`, `lng`, `lat` | Redis 写 `scanLng`/`scanLat` |

---

### 3.7 EcuBleCommandReportApi — 蓝牙控制回调

App 蓝牙操作后的结果上报。缺设备 `17006`，未上架 `170018`。成功 data 为 `null`。

| 路径 | 说明 | 关键字段 | 副作用 |
|---|---|---|---|
| `/device/paas/ble/lock/report` | 电门结果 | `acc`, `izRiskControl`(缺省 true) | 风控 Redis + C34 acc 快照 |
| `/device/paas/ble/defend/report` | 设防结果 | `defend` | 删风控标记 + C34 defend 快照 |
| `/device/paas/ble/helmetLock/report` | 头盔锁结果 | `sw` | Redis 写 `helmetLock` |
| `/device/paas/ble/rearWheelLock/report` | 后轮锁结果 | `sw` | Redis 写 `backWheelLock` |
| `/device/paas/ble/batteryCompartment/report` | 电池仓结果 | `sw` | Redis 写 `batteryLock` |
| `/device/paas/ble/deviceInfo/report` | 蓝牙查询设备状态 | gsm/voltage/gps/defend/acc/头盔等 | Kafka 全量 C34 |
| `/device/paas/ble/rfid/report` | RFID 上报 | `rfidAck`, `rfidCarId`, `rfidTimestamp` | Redis 写 RFID 字段 |

公共字段：`imei`, `carId`, `lng`, `lat`, `commandContext`。

---

### 3.8 运维 / 探活

对齐 Spring Actuator，供 K8s 探针与 preStop 使用。

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/actuator/health/liveness` | 存活探针，`{"status":"UP"}` |
| GET | `/actuator/health/readiness` | 就绪探针，`{"status":"UP"}` |
| POST | `/actuator/deregisterService` | 从 Nacos 注销；**仅允许 localhost** |

---

## 4. 请求 / 响应示例

### 车辆详情

```http
POST /device/paas/device/detail
Content-Type: application/json

{"carId":"10001","commandContext":{"tenantId":"1","traceId":"t-1"}}
```

```json
{
  "success": true,
  "code": "0",
  "msg": "成功",
  "data": {
    "imei": "86xxxxxxxxxxxxxxx",
    "carId": "10001",
    "lat": 31.23,
    "lng": 121.47,
    "acc": 0,
    "defend": 1,
    "isOnline": 1,
    "restBattery": 80,
    "ridingState": 0
  }
}
```

### 开电门

```http
POST /device/paas/lock
Content-Type: application/json

{"imei":"86xxxxxxxxxxxxxxx","acc":1,"izRiskControl":true,"commandContext":{"tenantId":"1","traceId":"t-1"}}
```

---

## 5. 本服务依赖的下游（调用链）

PaaS 自身聚合，不直接连 IOT 平台。下游失败时部分查询会降级（不过滤 / 空列表 / 地址置空），指令类则把网关错误映射为对外码。

```
调用方 (Feign: ebike-device-paas)
        │
        ▼
ebike-device-paas-go
        ├── Redis          设备影子 / GEO / ZSET / token / 限流
        ├── Kafka          产出 C34（topic = ecu.kafka.parent-topic，如 saas_0）
        ├── ebike-device-openapi   指令下发与 job 查询、电量计算
        ├── ebike-device-worker    轨迹 / 监控
        ├── ebike-management       角色可见车辆、租户 IOT 凭证
        ├── ebike-fence            服务区 / 还车跨区 / 隐藏车 / 多边形
        └── map-service            批量逆地理（仅 device/page）
```

| 下游 | 主要路径 | 用于 |
|---|---|---|
| **ebike-device-openapi** | `/ebike/cmd/lock` 等指令、`/ebike/cmd/query_*`、`GET /ebike/cmd/get_job_result`、`/voltage-plan/get-rest-battery` | 指令与中控查询 |
| **ebike-device-worker** | `/ebike/gps/getTrajectory`、`getOrderTrajectory`、`getBatchOrderTrajectory`、`getTrajectoryDistance`、`getMetric`、`saveOrderTrajectory` | 轨迹 |
| **ebike-management** | `POST /user/carInfos`、`POST /tenant/tenantIotPlatform/queryAll` | 角色过滤、AES 认证头、C34 `appId` |
| **ebike-fence** | `/serviceArea/getList`、`getById`、`getAllService`；`/config/backcar/getConfigByServiceId`、`/config/usecar/getConfigByServiceId` | 服务区与还车配置 |
| **map-service** | `POST /map/batchRegeo`（Header `secret`） | 分页地址 |

转发 openapi / worker 的 `/ebike/.*` 会带 AES 头：`authKey` / `authSecret` / `appId`（凭证来自 management 租户 IOT 平台缓存）。

---

## 6. 对外事件：Kafka C34

不是入站接口，但是本服务对外产出的设备状态事件。

| 项 | 说明 |
|---|---|
| Topic | `ecu.kafka.parent-topic`（Nacos，生产示例 `saas_0`） |
| 触发 | `location/change`、BLE `lock/defend/deviceInfo` 上报 |
| 消息 | 无 key，RoundRobin 分区；包络 `{payload, result, imei, deviceDataType:"cmd", code:0, appId}` |

上游消费方按既有 C34 约定解析。

---

## 7. 注册与发现

| 项 | 说明 |
|---|---|
| Nacos 服务名 | `ebike-device-paas` |
| Group | `xyy`（默认） |
| 注册开关 | `nacos.registerEnabled` / 环境变量 `NACOS_REGISTER_ENABLED` |
| K8s Service | `ebike-device-paas-go`（与 Nacos 名不同，仅集群内直连） |

上游 Java 继续用 Feign `name = "ebike-device-paas"` 即可，无需改客户端。

灰度路由（本服务内部，对调用方透明）：

| 配置 | 行为 |
|---|---|
| `liveList` | 本机 Go 处理 |
| `recordList` | 反代 Java，打 `[RECORD]` |
| 其余 + `dryRun=true` | Go+Java 双跑，客户端收 Java，打 `[SHADOW MATCH/DIFF]` |

---

## 8. 接口数量汇总

| 分组 | 数量 |
|---|---|
| DeviceInfo（含 getDeviceRealGpsList） | 33 |
| EcuQuery（不含 GPS 列表） | 5 |
| EcuCommand | 18 |
| EcuJobQuery | 13 |
| DeviceTrajectory | 6 |
| StateChange | 4 |
| EcuBleCommandReport | 7 |
| **业务合计** | **86** |
| Actuator | 3 |
| **HTTP 合计** | **89** |
