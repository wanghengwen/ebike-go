# 架构设计

## 设计原则

1. **业务规则不进 UI 层。** 校验、状态机、策略编排全部在 `shared/commonMain` 或后端。
   目的不只是复用，而是让将来新增平台（含鸿蒙）的成本收敛为「重写一层 UI」。
2. **平台能力只定契约。** BLE、地图、后台定位、扫码允许各端用各自 SDK 实现，
   `shared` 只描述「要做什么」和「按什么顺序做」。
3. **闭源依赖可插拔。** 私有 SDK 通过接口注入，仓库内提供模拟实现，保证开源版本可编译可运行。
4. **双端行为一致优先于代码量最小。** iOS 对等是硬指标，宁可多写一层 UI，也不接受两端逻辑分叉。

---

## 分层

```text
androidApp                                  iosApp
  UI（Compose / XML）                         UI（SwiftUI / UIKit）
  平台 SDK 适配：腾讯地图 · BLE · 定位 · 相机     平台 SDK 适配：腾讯地图 · BLE · 定位 · 相机
            └───────────────┬───────────────────────┘
                            ▼
                   shared（Kotlin Multiplatform）
  ┌────────────────────────────────────────────────────────────┐
  │ feature/    每个业务域的 UseCase 编排与状态（无 UI 依赖）      │
  ├────────────────────────────────────────────────────────────┤
  │ domain/     模型 · 任务状态机 · 车控策略 · 校验规则           │
  ├────────────────────────────────────────────────────────────┤
  │ data/       API 定义 · DTO · Mapper · Repository            │
  ├────────────────────────────────────────────────────────────┤
  │ core/       Ktor 客户端 · 签名 · 配置 · 日志 · Result 类型    │
  ├────────────────────────────────────────────────────────────┤
  │ platform/   expect 契约：BLE · 定位 · 扫码 · 地图 · 安全存储  │
  └────────────────────────────────────────────────────────────┘
```

`feature/` 层向 UI 暴露状态流与意图函数，不暴露 Repository。UI 层只做渲染与事件转发。

---

## 目录结构

```text
ebike-OpsApp/
  settings.gradle.kts
  build.gradle.kts
  gradle/libs.versions.toml          依赖版本集中管理
  shared/
    build.gradle.kts
    src/commonMain/kotlin/com/luopingtech/ebike/ops/
      core/        network · signing · config · logging · result
      domain/      model · usecase · task · control
      data/        api · dto · mapper · repository
      platform/    expect 声明
      feature/     auth · tenant · home · vehicle · scan · control
                   task · warehouse · production · report · tracking
    src/androidMain/kotlin/...        actual 实现（Android）
    src/iosMain/kotlin/...            actual 实现（iOS）
    src/commonTest/kotlin/...         签名 · 状态机 · 策略单测
  androidApp/
    src/main/kotlin/...               UI · 权限 · 地图 · BLE SDK 适配
  iosApp/
    iosApp/                           UI · 权限 · 地图 · BLE SDK 适配
    Podfile                           或 SPM
  ble-simulator/                      可入库的蓝牙模拟实现
  config/                             租户配置（仅 demo 入库）
  docs/
```

包名前缀 `com.luopingtech.ebike.ops`，与 `ebike-UniApp` 建议的 `com.luopingtech.ebike.*` 一致。

---

## 平台契约（expect / actual）

`shared/src/commonMain/.../platform/` 下声明，两端各自 `actual` 实现。

| 契约 | 职责 | Android 实现 | iOS 实现 | 仓库内可用替代 |
|------|------|-------------|---------|---------------|
| `BleTransport` | 连接、发指令、断开、扫描附近车辆 | 私有 `LruBle` | 私有 `XCBLETool` | `ble-simulator` 模拟实现 |
| `LocationTracker` | 单次定位、连续定位、后台轨迹上报 | 腾讯定位 + 前台 Service | `CLLocationManager` always | 系统定位 |
| `CodeScanner` | 二维码与条码识别 | CameraX / ZXing | `AVFoundation` | 系统相机 |
| `MapProvider` | 地图容器、Marker、聚合、围栏绘制 | 腾讯 / Google | 腾讯 / Google | 无（需自备 Key） |
| `SecureStore` | Token 与凭据存储 | EncryptedSharedPreferences | Keychain | 内存实现（仅测试） |
| `PushRegistrar` | 推送 Token 注册与回调 | 极光或 FCM | APNs | 空实现（默认关闭） |
| `DeviceInfo` | 型号、系统版本、App 版本 | 平台 API | 平台 API | — |

**车控策略留在 `commonMain`。** 遗留工程里「蓝牙优先 / 仅蓝牙 / 网络」三种策略在双端各写一遍，
是行为不一致的主要来源。新架构中策略是一个纯 Kotlin 状态机，`BleTransport` 与网络 API 都是它的依赖。

---

## 网络层

- 客户端：Ktor，`commonMain` 定义接口与 DTO
- 鉴权：沿用遗留的 `Authorization` + 时间戳 + 签名方案，双端已对齐，服务端无需改动
- 基址：从租户配置读取，不硬编码
- 路径：只接新版 `/business/paas/*`、`/business/fence/*`、`/business/ebike-operation/*`、
  `/business/ebike-management/*`；遗留 `/ebike/v2/*`（在 `ApiService.kt` 中已整体标注 `@Deprecated`）不迁移
- 签名实现必须有单测覆盖，作为 P1 的验收项

---

## 并发模型

`shared` 内统一协程 + `Flow`。遗留工程 Android 用 RxJava3、iOS 用 RxSwift，本项目不再引入 Rx。
iOS 侧通过 SKIE 或等价桥接把 `Flow` 与 `suspend` 暴露为 Swift 友好形式，避免 Swift 层手写回调转换。

---

## 技术决策记录

### 为什么选 KMP 共享逻辑 + 双端原生 UI

范围收敛后 App 约 56 屏（合并后 40 上下），其中相当比例是地图、扫码、蓝牙、相机等重原生交互。
原生能力的工作量是常数，不随 UI 框架变化，因此「接原生的摩擦最小」比「少写一套 UI」更值钱。
同时 Android 侧约 1200 个 `.kt` 的业务逻辑可以搬而不是翻写。

### 被否方案

| 方案 | 否决原因 |
|------|---------|
| KMP + Compose Multiplatform 全量共享 UI | 地图与 BLE 需要额外 interop 封装；iOS 侧生态与包体风险；对鸿蒙也无帮助。可在 P6 之后对单个列表页做局部试验 |
| Flutter 收敛为唯一 UI 层 | 需把 1200 个 `.kt` 业务逻辑翻写成 Dart；鸿蒙依赖非主线 OHOS 分支 |
| uni-app 主壳 + 原生插件 | 私有 BLE 与千级车点聚合需自写插件，收益被抵消；运维端不需要小程序形态 |
| 继续双原生不抽共享层 | 双端业务逻辑分叉是遗留工程 iOS 落后的根因，不解决问题 |

### 鸿蒙余地怎么留

不为鸿蒙改变当前选型，改为三条架构约束：

1. 业务规则只在 `shared` 或后端，UI 层无逻辑
2. 平台能力是接口，新增平台等于新增一份实现
3. 任务与审核字段走后端 schema 下发（P6），新平台只写渲染器

这样鸿蒙进场的成本是「一层 UI」，而不是「一个 App」。

---

## 不迁移的遗留实现

| 项 | 说明 |
|----|------|
| MQTT 桥接 | Android Manifest 声明了 Paho `MqttService` 但业务无引用；iOS `MQTTManager.swift` 的 Pod 依赖已移除。均为死代码 |
| 百度定位 | `BaiDuLocationService.kt` 存在但未接入 |
| 高德导航跳转 | `XMapView` 已按国内外只选腾讯 / Google，高德仅遗留引用与聚合算法参考 |
| `kotlin-android-extensions` | 已废弃，新工程用 ViewBinding 或 Compose |
| 明文 HTTP 放行 | 遗留 `usesCleartextTraffic=true` 与 `NSAllowsArbitraryLoads=true` 不带入新工程 |
| OSS 热更新检查 | 与新发布流程一并重新设计，不照搬 `version.json` 方案 |
