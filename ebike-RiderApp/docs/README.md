# ebike-RiderApp 文档

C 端（骑行用户）跨平台 App。与 `ebike-OpsApp`（运维端）同构：Kotlin Multiplatform +
Compose Multiplatform，`shared` 放业务与数据，`sharedUi` 放跨端界面，两个宿主
（`androidApp` / `iosApp`）只负责装配平台能力。

## 当前进度

**P4 完成**：骑行主循环（扫码 → 确认 → 解锁 → 骑行 → 临停 → 还车 → 结费），BLE 优先 + 网络降级，
杀进程可恢复。**P6 完成**：H5 长尾容器 + `nativeHost` 桥。
支付（P5）仍未接：结费屏只读展示费用，桥上 `pay` 返回 `UNSUPPORTED`。

| 项 | 状态 |
| --- | --- |
| `shared` / `sharedUi` 模块与 `verifyCommonMain` | 绿 |
| P0/P1（core、签名、登录 Session） | 绿，见下方 |
| Android 腾讯地图渲染器可注入；无 Key → `SimulatorMap` | 绿 |
| Android 扫码预览 / `ScanQrActivity` / 定位 / 拍照 / 上传 | 绿 |
| iOS MapKit 降级 + 扫码 / 定位 / 拍照源码 | 已落地（Windows 上 metadata 绿；真机 link 需 Mac） |
| `HomeMapScreen`（`Idle` 相位首页，可进 Debug） | 绿 |
| 上传路径 `/client/file/upload` | 绿（对齐 UniApp） |
| `ScanCodeParser` + 单测 | 绿 |
| `:shared:testAndroidHostTest` | 绿（含 BLE 黄金帧） |
| BLE 帧层 `domain/ble`（照搬 UniApp 怪语义） | 绿，见 [ble-protocol.md](ble-protocol.md) |
| `BleSessionFeature` + `BleTokenApi` + Simulator | 绿 |
| Android `BluetoothGatt` / iOS `CoreBluetooth` | 源码齐全；真机未联调 |
| `:androidApp:assembleDebug` | 绿，出 APK |
| `ebike-OpsApp` 未被本阶段修改 | 绿（只读参考拷贝） |
| P4 `RideStateMachine` 纯函数 + 杀进程恢复 | 绿 |
| P4 `RidingFeature`（BLE 优先 / 网络降级 / 静音解锁） | 绿 |
| P4 骑行域 9 屏 + `HomeMapScreen` 接开锁 | 绿 |
| P4 Android 前台定位「骑行中」/ iOS 后台定位 | 绿（iOS 真机待 Mac 验） |
| P4 demo 假解锁 → 骑行 → 假还车 → 结费 | 绿，无后端可走通 |
| P5 支付 | **未做**，结费屏只读 |
| P6 `H5Screen` / `NativeHostBridge` / 49 屏路由 | 绿；demo 无 `h5.baseUrl` 显示占位 |

P1 仍保留：`PhoneLoginScreen`、`RiderAppRoot`、Demo 登录、签名探针（Debug 屏入口）。

已落地的模块（P2 **粗体**，P3 另见 BLE 段，P4 见下节）：

- `shared/core` — 配置、签名、网络、i18n、结果类型
- `shared/platform` — 安全存储、设备信息、**ReverseGeocoder**、Capabilities（地图元数据 / 扫码 / 拍照 / 上传）
- `shared/domain` — **`MapPin` / 地图投影**、**`ScanCodeParser`**、`BleFrame` / `BleFrameParser` / `BleCommands`
- `shared/domain/riding` — `RideStateMachine` / `RideSession`、`ReturnDecision`、`RidingFenceTips`、`RideFormat`
- `shared/domain/tracking` — `TrackPointBuffer`（OpsApp 只读拷）
- `shared/data` — auth；**`FileUploadApi`（/client/file/upload）**、**`NearbyVehicleApi`**、`BleTokenApi`
- `shared/data/riding` — `RidingApi` / `BleRideApi` / `ApplyReturnApi` + demo 三件套
- `shared/data/fence` — `FenceApi`（服务区 / 停车点 / 附近围栏）
- `shared/feature/ble` — `BleSessionFeature`（历史、超时、`code==0` 成功）
- `shared/feature/riding` — `RidingFeature` 编排、`RideSessionStore` 持久化
- `sharedUi` — 登录；**`RiderMap` / `SimulatorMapView` / `HomeMapScreen`**、**`RiderScanPreview` CompositionLocal**
- `sharedUi/ui/ride` — `RideFlowHost` + 骑行 9 屏 + `RidePromptHost`
- `androidApp` — **TencentMapView、CameraX 扫码、Activity 拍照、AndroidLocationTracker、AndroidMediaUploader**、`TrackLocationService`
- `sharedUi/iosMain` — **MapKit 降级、IosHostMap 契约、IosScanPreview**

## P4 骑行主循环

### 相位机 `domain/riding/RideStateMachine`

`Idle → Scanning → Confirming → Unlocking → Riding → TempLocked → Returning → Settling`，
`Settling` 走完回 `Idle`。**纯函数**：没有协程、没有时钟，时间戳一律由事件带进来，所以每条转移都能单测钉死。

UniApp 把这套状态散在 `tempData.ride.status`、`getRideInfo().ridingState` 和各页面 `ref` 里，
漏改一处就出现「界面在骑行、store 说已结束」。这里收成单一 `RideSession`，只走 `reduce`。

两条刻意的设计：

1. **未定义的（相位, 事件）组合原样返回，不抛异常。** BLE 回调、`getRideInfo` 轮询和用户点击会并发到达，
   晚到的 ack 落在已经翻页的相位上属于常态，不是 bug。
2. **`Reset` 在骑行中被忽略。** 否则用户按「返回」就回首页，而车还开着。

### 杀进程恢复

`RideSessionStore` 把 `RideSession` 序列化进 SecureStore；冷启动 `KillRestore` 灌回，
再由 `getRideInfo` 的 `ServerSync` 校正。恢复原则是**只信稳定相位**：

| 存档相位 | 恢复成 | 为什么 |
| --- | --- | --- |
| `Riding` / `TempLocked` / `Settling` | 原样 | 稳定态，随后 `ServerSync` 校正 |
| `Unlocking` | `Confirming`（保留车牌） | 请求已发、响应未确认，本地判不出车开没开；订单若已建，`ServerSync` 会抬回 `Riding` |
| `Returning` | `Riding` | 车还在手上，重走还车 |
| `Scanning` | `Idle` | 相机会话不可能跨进程 |
| `Confirming` | 有车牌就留着，否则 `Idle` | — |

`ServerSync` 里服务端是权威：`ridingState` 为 3 → 临停，4/5/6 → 骑行中；
读不到进行中订单而本地还在骑行相位，说明车已被别处结束（运维后台，或上次还车成功但响应丢了）→ 推进结费。

### 解锁时序（照搬 UniApp `unlockByBle`）

```
ridePermission → BLE 0x2c(mute=true) → rideReport 建单 → PLAY_VOICE(START)
                                              │
                                        建单失败 → BLE lock 回滚
```

**静音下发、建单成功才播语音**：顺序反了会出现「车响了但没订单」—— 用户以为骑上了，实际白骑，账也对不上。
建单失败必须锁回去，同理。

`UnlockPolicy` 三档：`BlePreferred`（默认，地库电梯口只有蓝牙能用）、`NetworkPreferred`（车机离线时仍自动顶到蓝牙）、
`BleOnly`（租户 `features.onlyBluetooth`，跳过必然超时的那轮远程指令）。BLE 失败降级网络，`BleOnly` 除外。

### API path（对齐 UniApp）

| 用途 | path |
| --- | --- |
| 扫码 / 车辆详情 / 骑行详情 | `client/rent/scan`、`client/rent/getCarInfo`、`client/rent/getRideInfo` |
| 蓝牙链路 | `client/rent/blue/ridePermission`、`blue/ride`、`blue/tempParking`、`blue/endParking`、`blue/return` |
| 网络链路 | `client/rent/network/ride`、`network/return`、`client/rent/tempParking`、`endParking` |
| 还车判定 | `client/rent/returnPermission`、`client/system/getbackCarConfig[ByCarId]` |
| 围栏 | `client/fence/serviceArea/getByLocation`、`getNearFence`、`client/fence/parking/nearParkingNum` |
| 申诉还车 | `client/returnBikeAudit/izCapable`、`createReturnBikeAudit` |
| 其他 | `client/helmet/unlock`、`client/paas/device/carSearchVoice`、`client/rent/tempUnlock`、`unFrozenOrder` |

### 屏

`RideFlowHost` 是骑行域唯一路由，按 `RidePhase` + 局部 overlay 切屏：
`ScanScreen`、`ManualIdScreen`、`PreCyclingScreen`、`RidingScreen`、`ParkSearchScreen`、
`ReturnGuideScreen`、`ApplyReturnScreen`、`TripMapScreen`、`SettlementScreen`。
弹层统一走 `RidePromptHost`（文明停车、罚金、头盔、超载、余额不足…），一律 `AlertDialog`，不用实验期的 `ModalBottomSheet`。

`HomeMapScreen` 不再自己发 BLE 指令，改为 `onStartRide` 把车交给 `RideFlowHost`；顶栏另加「输入车牌号」。
`RiderAppRoot` 只把登录后的首页从 `HomeMapScreen` 换成 `RideFlowHost`，登录 / Debug / H5 分支未动。

### 前台定位

Android `TrackLocationService` 通知文案「骑行中」，channel 名与正文都走 `strings.xml`（含 `values-en`）。
开关条件是 `RideSession.needsForegroundLocation`：`Unlocking` 起、`Returning` 止。

iOS 要**两处一起改**才生效：`iosApp/project.yml` 的 `UIBackgroundModes: [location]`，
以及 `CoreLocationTracker` 里的 `allowsBackgroundLocationUpdates = true`。
少了 plist 那条，赋值会直接抛异常（Apple 的显式约束），所以代码里先探测 plist 再赋值。
只申请 `whenInUse`：配合 background mode，骑行中系统显示蓝色状态条，熄屏也能持续定位，不必要求 `always`。

### Demo（无后端）

`config/demo_release.json` 下走 `DemoRidingRemote` / `DemoBleRideRemote` / `DemoApplyReturnRemote` / `DemoFenceRemote`，
可以假解锁进骑行、假还车进结费。

两个 demo remote **共用一本 `DemoRideBook`**：BLE 建单写的是同一本账，否则蓝牙开锁后骑行页费用一直是 0 元
（BLE 建了单，`getRideInfo` 那边不知道）。

## P4 已知限制

1. **支付未接（P5）**：`SettlementScreen` 只读展示费用明细，「去支付」是占位；未接微信 / 支付宝 SDK。
2. **`ReturnGuideScreen` / `TripMapScreen` 简化**：前者按 `pageType` 给文字指引，未做 UniApp 的图示与实景对比图；
   后者画轨迹折线 + 汇总，未做分段配速。
3. **真车未联调**：解锁时序、`returnPermission` 各 `returnType` 分支都按 TS 逐字移植，测试用 fake remote；
   围栏判定的实际边界要真车 + 真租户配置验。
4. **轨迹只上报不落库**：`TrackPointBuffer` 按 OpsApp 节流规则缓冲，行程结束后不持久化，`TripMapScreen` 只看内存里这一趟。
5. **`ridingState` 语义靠猜**：3 = 临停、4/5/6 = 骑行中来自 UniApp 分支代码，后端未给枚举文档。

## P3 已知限制

1. **真车 notify 未抓包**：0x41 / 0x42 / 0x54 字段偏移按 TS 逐字移植，单测用合成 GPS + CRC 失败路径；要锁死固件布局需真车 hex。
2. **`resolveAck` 怪语义保留**：`len>1` 应答几乎总是 `code==0`，上层当成功。不要「修好」。
3. **默认 token 十进制**：字节是 `0A 0A 05 05` = `168428805`。规划稿 `168496389` 是 `0x0A0B0D05`，不是默认帧。
4. **厂商 BLE 栈**：小米 / OPPO / vivo / 鸿蒙可能延迟 CCCD、丢首包 notify、或不广播 `5841`。
5. **iOS**：Windows 上只保证源码齐全；真机 link / 系统蓝牙弹窗要在 Mac 上验。`NSBluetoothAlwaysUsageDescription` 已在 `iosApp/project.yml`。

## P2 已知限制

1. **Demo 租户无腾讯 Key**：`config/demo_release.json` 的 `mapProvider=simulator`，地图走 Canvas 模拟器，不崩。
2. **Android 模拟器无腾讯 x86 slice**：`androidApp` 的 `ndk.abiFilters` 仅 `armeabi-v7a` / `arm64-v8a`；Intel 模拟器请用 ARM 镜像或真机验腾讯地图。
3. **iOS 腾讯地图**：真机经 `TencentMapFactory` + QMapKit；模拟器无 SDK slice，自动 MapKit。Key / Bundle ID 由 `sync_tenant.sh` 注入。
4. **MapKit 坐标系**：后端 GCJ-02，MapKit 用 WGS-84，兜底地图会有整体偏移（文档化，不为此引转换层）。
5. **附近车辆 API**：`client/ebike/nearby` 为预留 path；demo 模式用 `DemoNearbyVehicleRemote` 固定三辆车。
6. **骑行闭环**已在 P4 落地（见上）；**支付**仍是 P5。H5 长尾见下方 P6。BLE 协议见 [ble-protocol.md](ble-protocol.md)。

### 腾讯地图 Key 登记

- Android：`local.properties` 的 `rider.map.tencentKey`，或租户 JSON 的 `map.tencentKey`；构建时写入 `TencentMapSDK` meta-data 与 `BuildConfig.TENCENT_MAP_KEY`。
- iOS：`Info.plist` 的 `RiderTencentMapKey`（由 `iosApp/sync_tenant.sh` / xcconfig 注入，与 OpsApp 同模式）。

## P1 未做（仍有效）

- 运营商一键登录真实 SDK
- 真机对后端登录（需租户 `api.baseUrl` / `businessSecret`）
- 协议页现走 H5 长尾（需配置 `h5.baseUrl`）

## 与 OpsApp 的差异

- 包名 `com.luopingtech.ebike.rider`，类型前缀 `Rider*`
- 上传走 **`/client/file/upload`**，不是运维 `/business/ebike-management/file/upload`
- C 端首页是 **`RideFlowHost`**（`Idle` 相位渲染 `HomeMapScreen`），不是运维四 Tab
- 未拷贝 `ReturnCarScatterMapView` 等运维专用地图

## P6 H5 长尾 + nativeHost

长尾页（钱包、帮助、设置、行程、发票等 49 个 UniApp hash）在 WebView 里跑 UniApp `build:h5` 产物。
骑行主循环仍走原生；H5 **不能**自己签请求。

### 桥方法

| 方法 | 行为 |
| --- | --- |
| `request` | 原生 `SignedApiClient` 代理；path 必须以 `/client/` 开头，拒绝 `/oauth/token` |
| `getProfile` | 展示字段（userId / pin / phone / …），**不含 token** |
| `pay` | 固定 `UNSUPPORTED`（P5 未做） |
| `scanCode` / `capturePhoto` / `currentLocation` | 复用 Rider 已有端口 |
| `openNavigation` | 系统地图 |
| `navigate` / `close` / `setTitle` / `toast` | 容器控制；登录跳转回原生登录页 |
| `getLanguage` / `setLanguage` | 与 `RiderI18n` + SecureStore 同步 |

Android：`window.__riderNative`（`addJavascriptInterface`）。iOS：`webkit.messageHandlers.riderNative`。
文档基址（`#` 前）变化才 `loadUrl`，避免 SPA hash 把用户打回入口（OpsApp 同坑）。

### 安全

1. `signSecret` / `businessSecret` **永不**进 H5 URL 或 `getProfile`
2. 网络只走桥；origin 必须等于租户 `h5.baseUrl`（`file://` 仅本地页）
3. URL 最多拼 `lang` / `tenantId` / `themeColor`

### 样例入口

登录后首页顶栏「我的钱包 / 帮助」，Debug 屏同样两个按钮。
`h5.baseUrl` 为空（demo 默认）时显示占位说明，不加载 WebView。
配好后例如：`https://your-cdn/h5/index.html`，打开 `#/pages-sub/pay/wallet/wallet`。

### 语言回写

H5 设置页切换语言会调 `nativeHost.setLanguage`。下次打开 H5，`App.vue` 的 `syncFromNativeHost()` 用 `getLanguage()` 覆盖 vue-i18n；URL 也带 `lang=`。

## 与 UniApp 对齐

| 点 | UniApp | Rider |
| --- | --- | --- |
| 上传 | `POST /client/file/upload` multipart | 同 |
| 附近车 | （各租户可能不同） | `client/ebike/nearby` 桩 + demo |
| BLE 帧 / token | `features/ble` + `/client/paas/device/getBlueToothToken` | 同语义；见 [ble-protocol.md](ble-protocol.md) |
| 解锁时序 | `unlockByBle`：静音 → 建单 → 语音 | 同；建单失败回滚上锁（见 P4 段） |
| 骑行状态 | `tempData.ride.status` + `ridingState` 双份 | 收成单一 `RideSession`，只走 `reduce` |
| 还车判定 | `returnPermission.returnType` 分支 | `ReturnDecision.flow` 单测覆盖 |

登录 / 刷新 / deviceId 等同 P1，见 git 历史或上一版文档段落。

## 构建

### 前置

复制 `local.properties.example` 为 `local.properties`，至少填 `sdk.dir`。
`rider.tenant` / `rider.mode` 默认 `demo` / `release` —— demo 不含密钥，`api.baseUrl` 为空即 demo 模式。

### 校验

```
./gradlew verifyCommonMain
./gradlew :shared:testAndroidHostTest
./gradlew :androidApp:assembleDebug
```

### Android

产物：`androidApp/build/outputs/apk/debug/androidApp-debug.apk`。

### iOS

需要 macOS。Windows 上 `compileIosMainKotlinMetadata` 可绿，完整 link 见 [../iosApp/README.md](../iosApp/README.md)。

模拟器构建钉 `ARCHS=arm64`（Compose 1.11+ 无 Intel slice）。`_build_ios_release_rider.sh` 已处理。

## 签名

见 [signing.md](signing.md)。
