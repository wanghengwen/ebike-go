# 开源合规检查清单

本项目随 `ebike-go` 以 [Elastic License 2.0](../../LICENSE) **源码开放（Source Available）** 发布。
遗留 `APP/` 工程存在多处硬编码密钥与闭源依赖，**不能直接搬运**。本清单是 P0 的必过项。

对外描述统一用「Source available under Elastic License 2.0」，不要写成「MIT / 开源无限制」。

---

## 一、密钥与凭据

遗留工程中已确认存在的敏感项，迁移时一律外置，且**新仓库不继承遗留 git 历史**：

| 敏感项 | 遗留位置 | 处置 |
|--------|---------|------|
| `businessSecret`、`signSecret` | 遗留 Android flavor Gradle 脚本、iOS `config.json` | 移入 `config/{tenant}_release.json`，由 `.gitignore` 排除 |
| 腾讯地图 / Google Maps / 高德 Key | 遗留 flavor 脚本、`AndroidManifest.xml` 的 `AMAP_KEY`、iOS `Info-*.plist` | 同上，使用者自备 |
| 极光推送 AppKey | 遗留 flavor 脚本与 iOS 配置 | 同上，默认不启用推送 |
| 签名库口令 | Android Gradle 脚本内明文 | 移入 `local.properties` 或 CI Secret，模板为 `local.properties.example` |
| 云 MQTT 接入点 | iOS `MQTTManager.swift` 内硬编码 | 死代码，直接不迁移 |
| OSS 热更新地址 | 遗留双端硬编码的 `version.json` 地址 | 不迁移，发布流程重新设计 |
| 生产 API 基址 | 遗留 Gradle 与 iOS 配置内硬编码 | 不作为默认值硬编码，从租户配置读取 |

**已进入过版本库的密钥视为已泄露，必须轮换。** 开源前完成一次轮换，并在服务端确认旧密钥失效。

检查项：

- [ ] 全仓库检索 `secret`、`Secret`、`KEY`、`password`、`token`、`.jks`、`.p12`、`.mobileprovision` 无命中
- [ ] `config/` 下仅 `_defaults.json` 与 `demo_release.json` 入库，且不含任何真实密钥
- [ ] `.gitignore` 覆盖 `local.properties`、`*.jks`、`*.keystore`、`*.p12`、`*.mobileprovision`、`config/*_release.json`（demo 除外）
- [ ] CI 使用 Secret 注入，构建日志不回显配置内容
- [ ] 首次提交前用 `git log -p` 抽查，确认无密钥进入历史

---

## 二、闭源与第三方 SDK

以下依赖**不随仓库分发**，仓库内只保留接口与可运行的替代实现：

| 依赖 | 性质 | 仓库内提供 | 使用者需自备 |
|------|------|-----------|-------------|
| iOS `XCBLETool.framework` | 私有闭源 BLE SDK | `BleTransport` 接口 | 私有 Pod / 二进制 |
| Android `LruBle`（`com.niubi.sdk.bluetooth.library`） | 私有闭源 BLE SDK | `BleTransport` 接口 | 私有 Maven |
| 腾讯地图 SDK | 商业 SDK，需 Key | `MapProvider` 抽象 | SDK 与 Key |
| Google Maps SDK | 商业 SDK，需 Key | 同上 | SDK 与 Key |
| 极光推送 | 商业 SDK | `PushRegistrar` 空实现 | SDK 与 AppKey |
| 友盟统计 / APM | 商业 SDK | 不集成 | 按需自行接入 |

要点：

- 遗留 `Merchant-Android/localRepository` 与 `Merchant-iOS/localPod` 下的私有二进制**不要入库**
- `ble-simulator` 模块提供可运行的模拟实现，保证开源版本在没有真实 SDK 时也能编译、跑通任务流程与联调界面
- 依赖库仍受各自许可证约束，`LICENSE` 只覆盖本仓库自有代码
- 在 README 明确「私有 BLE SDK 需自备」，避免使用者误以为开箱可控车

检查项：

- [ ] 仓库内无 `.framework`、`.aar`、`.jar` 形式的私有二进制
- [ ] 移除私有 Maven / Pod 源地址中的内网域名与凭据
- [ ] 关闭真实 SDK 时 `./gradlew build` 与 iOS 构建均通过
- [ ] 第三方 SDK 清单与许可证在 `docs/` 中列明

---

## 三、品牌与业务数据

| 项 | 遗留现状 | 处置 |
|----|---------|------|
| 应用显示名 | 遗留双端使用租户商标名 | 改为中性名称，商标不入库 |
| 包名 / Bundle ID | 遗留使用租户品牌前缀 | 统一为 `com.luopingtech.ebike.ops` |
| 租户图标与素材 | 指向租户 OSS 资源 | 替换为中性占位资源 |
| 多租户配置 | 遗留按租户建 Gradle flavor | 仅保留 demo 租户，参照 `ebike-UniApp` 的 `demo_release.json` 做法 |
| 测试账号与内网地址 | 散落在代码与文档 | 清理 |
| 客服电话、协议链接 | 指向具体租户 | 移入配置 |

检查项：

- [ ] 全仓库检索 `renren`、`luoping`、内网 IP 段、具体客户名，确认仅出现在必要的兼容说明处
- [ ] 截图与文档中无真实用户数据、车牌、手机号、坐标
- [ ] 示例数据全部为构造数据

---

## 四、代码与文档质量门槛

遗留工程有几处会影响开源观感的问题，一并处理：

- [ ] 无开发者本机绝对路径残留（遗留 Gradle 曾写死个人 home 目录）
- [ ] Debug 构建不打印完整请求体（遗留为 BODY 级日志）
- [ ] 不带入明文 HTTP 放行（`usesCleartextTraffic`、`NSAllowsArbitraryLoads`）
- [ ] 不带入死代码：MQTT 桥接、百度定位、高德导航跳转、Compose 空壳配置
- [ ] README 说明清楚「这是运维端，不是用户骑行端」，避免使用者误用
- [ ] 文档不暴露内部人名、工位路径、内部系统名

---

## 五、发布前流程

1. 完成上述四节全部检查项
2. 密钥轮换并确认旧密钥失效
3. 在干净环境克隆仓库，仅按 README 步骤构建，验证不依赖任何未公开资源
4. 用模拟 BLE 实现跑通登录、地图、扫码、任务流程冒烟
5. 补齐 `LICENSE` 顶部版权主体为公司法人全称（EL2.0 条款正文不改）
6. 记录第三方 SDK 与许可证清单
