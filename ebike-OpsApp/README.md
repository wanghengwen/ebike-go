# ebike-OpsApp

共享电单车 **运维 / 商户端 App**（Kotlin Multiplatform + 双端原生 UI）。

重写目标是替换 `APP/` 下的遗留工程（`Merchant-Android` / `Merchant-iOS` / `Merchant-Flutter` / `Merchant-H5`），
把业务逻辑收敛到一份 Kotlin 共享层，UI 保持 Android / iOS 各自原生。

> 许可与仓库根一致：[Elastic License 2.0](../LICENSE)，属**源码开放（Source Available）**，不是 OSI 定义的开源。
> 允许自建自用与修改，禁止作为托管 / 代运营服务对外提供。

---

## 与遗留工程的关系

| 维度 | 遗留 `APP/` | 本项目 |
|------|------------|--------|
| 技术栈 | Android 原生 + iOS 原生 + Flutter 模块 + Vue2 H5 | Android 原生 + iOS 原生 + KMP `shared` + Vue3 H5（仅两块大屏） |
| 业务逻辑 | 双端各写一遍，Flutter 再切一部分 | 单点收敛到 `shared/commonMain` |
| 功能范围 | 约 110 屏，含大量管理 / 报表 / 审核 | 仅现场作业，约 56 屏（合并后目标 40 上下） |
| 管理类功能 | 与 `pc` 后台重复实现 | 交回 `pc` 管理后台与移动自适应 H5 |
| iOS 完整度 | 明显落后于 Android | 与 Android 功能对等为硬指标 |

范围收敛的逐模块依据见 [`docs/MODULE-INVENTORY.md`](docs/MODULE-INVENTORY.md)。

---

## 功能范围（现场作业端）

保留在 App 内的能力，判定标准是**需要相机、蓝牙、现场定位或单手操作**：

- **账号与租户**：登录、验证码、改密、切换商户与服务区、语言与区号、设置
- **地图与找车**：车点聚合、车辆分布、车辆定位、车辆列表与详情
- **扫码与车控**：扫车码 / 中控 IMEI、车牌号开锁、开关锁、响铃、开关电池仓
- **蓝牙**：蓝牙识别、蓝牙雷达、附近批量挪车
- **任务执行**：换电 / 挪车 / 巡检 / 维修任务的领取、执行、扫码完成、拖回
- **仓库**：有码与无码出入库、明细与记录
- **生产现场**：车辆检测、中控绑定解绑、上下架
- **报修**：现场故障上报与拍照、我的上报记录
- **后台定位**：运维员位置与轨迹上报

明确**不做**（走 `pc` 后台或移动 H5）：订单资金、审核认证、异议工单、员工与权限配置、
操作日志、围栏编辑、运维阈值与标签配置、统计分析大屏、数据删除向导。

---

## 架构

```text
androidApp (Kotlin, Compose/XML)      iosApp (Swift, SwiftUI/UIKit)
  地图 · BLE · 定位 · 相机               地图 · BLE · 定位 · 相机
            └────────────┬────────────────┘
                         ▼
                shared (Kotlin Multiplatform)
                  commonMain
                    core/      网络(Ktor) · 签名 · 配置 · 日志 · Result
                    domain/    模型 · 用例 · 任务状态机 · 车控策略
                    data/      API · DTO · Repository · Mapper
                    platform/  expect 契约：BLE · 定位 · 扫码 · 地图 · 存储
                    feature/   auth · home · vehicle · task · warehouse …
                  androidMain / iosMain
                    actual 实现，封装各端原生 SDK
```

另有 `webH5/`：运营大屏与营收大屏（Vue 3 + TypeScript + Vite），独立的静态站点，
不参与 Gradle 构建，App 通过 WebView 按租户配置的 URL 打开。见 [`webH5/README.md`](webH5/README.md)。

设计原则、`expect/actual` 契约清单与模块依赖见 [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)。

核心约束三条：

1. **业务规则不写在 UI 层**，全部在 `shared` 或后端，将来新增平台只需重写 UI。
2. **平台能力只定契约，不强求共享实现**。BLE、地图、后台定位允许各端用各自 SDK。
3. **闭源 SDK 与密钥不入库**，通过接口注入与本地配置提供，见下节。

---

## 开源边界

本仓库**不包含**以下内容，需使用者自备：

| 类别 | 说明 | 仓库内提供的替代 |
|------|------|-----------------|
| 私有蓝牙 SDK | 遗留 iOS `XCBLETool.framework`、Android `LruBle` 均为闭源 | `BleTransport` 接口 + 模拟实现，可跑通除真实车控外的全流程 |
| 地图 SDK 与 Key | 腾讯地图、Google Maps | `MapProvider` 抽象，Key 从本地配置读取 |
| 推送与统计 SDK | 极光、友盟等 | 接口留空实现，默认关闭 |
| 平台密钥 | `businessSecret`、`signSecret`、地图 Key、签名库口令 | `config/` 下 demo 配置与 `*.example` 模板 |
| 品牌资源 | 遗留「人人运维」商标、图标、OSS 素材 | 中性占位资源 |

完整的开源前检查项见 [`docs/OPEN-SOURCE.md`](docs/OPEN-SOURCE.md)。

---

## 环境与构建

- **JDK 17 或 21**（不要用 25：当前 Gradle Kotlin DSL 无法解析）
- Android SDK（`compileSdk 36`，`targetSdk 34`，`minSdk 24`）
- 复制 `local.properties.example` → `local.properties`，填写 `sdk.dir` 与 `org.gradle.java.home`

```bash
cd ebike-OpsApp
# Windows
gradlew.bat :shared:testDebugUnitTest :androidApp:assembleDebug

# macOS / Linux（需自备 gradlew；或用 Android Studio 打开本目录）
./gradlew :shared:testDebugUnitTest :androidApp:assembleDebug
```

Debug APK：`androidApp/build/outputs/apk/debug/androidApp-debug.apk`。

当前脚手架验证：demo 登录（`api.baseUrl` 为空时本地会话）、真实 `/oauth/token`（配置 `assets/tenant.json`）、
`SimulatorBleTransport` + `VehicleControlPolicy` 模拟响铃。
蓝牙真实 SDK、地图尚未接入。

联调远程登录：复制 `androidApp/src/main/assets/tenant.json.example` 为 `tenant.json`，填入 `api.baseUrl`、
`tenantId`、`auth.businessSecret`、`auth.signSecret`（该文件已 gitignore）。

iOS 宿主见 [`iosApp/README.md`](iosApp/README.md)（需在 macOS / Xcode 上创建工程并嵌入 `Shared.framework`）。

---

## 租户配置

沿用 `ebike-UniApp` 的做法：`config/{tenant}_{mode}.json`，仅 `demo_release.json` 与 `_defaults.json` 入库，
含密钥的正式租户配置由 `.gitignore` 排除。详见 [`config/README.md`](config/README.md)。

---

## 迁移规划

分阶段计划、里程碑判据、工作量粗估与风险登记表见 [`docs/MIGRATION.md`](docs/MIGRATION.md)。

当前状态：**P1 进行中** — 登录、Token 刷新、服务区、车辆列表（`/business/paas/device/list`）、地图 pin 占位已落地；真实地图 SDK / BLE 仍未接入。

---

## 文档

| 文档 | 内容 |
|------|------|
| [`docs/MIGRATION.md`](docs/MIGRATION.md) | 迁移规划主文档：阶段、判据、工期、风险 |
| [`docs/MODULE-INVENTORY.md`](docs/MODULE-INVENTORY.md) | 110 屏逐模块处置清单与新旧映射 |
| [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) | KMP 分层、平台契约、目录结构、技术决策 |
| [`docs/OPEN-SOURCE.md`](docs/OPEN-SOURCE.md) | 开源合规检查清单 |
| [`webH5/README.md`](webH5/README.md) | 运营 / 营收大屏 H5：与 App 的 URL 契约、开发方式、与遗留实现的行为差异 |
