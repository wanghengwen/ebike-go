# ebike-go

共享电单车 / 两轮出行运营平台的 **Go 微服务集合**。


> **开源状态：渐进式源码开放（Source Available）**  
> 代码可查阅、可自建部署用于贵司自有业务；**不得**将本项目作为对外商业托管 / SaaS / 代运营平台向第三方提供服务。详见 [许可证](#许可证)。

---

## 你能用它做什么

- 自建或改造共享电单车 / 助力车运营后台与设备接入能力

| 场景 | 是否允许（Elastic License 2.0） |
|------|--------------------------------|
| 公司内部自用：跑自己的运营、自有品牌业务 | ✅ |
| 修改源码、二次开发后自用 | ✅ |
| 将修改后的源码再分发给同样遵守本协议的接收方 | ✅ |
| 对外售卖「基于本项目的托管运营 / SaaS / 代管服务」 | ❌ |
| 把本项目作为核心能力，向不特定第三方提供竞品级商业平台服务 | ❌ |

商业合作、OEM、区域授权等需求，请联系版权方单独签订商业许可。

---

## 架构概览

平台大致分为四层：

```text
                    ┌─────────────────────────────┐
  管理端 / 小程序    │  ebike-gateway-go           │
  App / OpenAPI     │  (business / client 双模)    │
                    └─────────────┬───────────────┘
                                  │
                    ┌─────────────▼───────────────┐
                    │  ebike-auth-go              │
                    │  identity-auth-go           │
                    │  ebike-service-*-go         │
                    │  ebike-fence-go / analyze…  │
                    └─────────────┬───────────────┘
                                  │
        ┌─────────────────────────┼─────────────────────────┐
        │                         │                         │
┌───────▼────────┐     ┌──────────▼──────────┐    ┌─────────▼────────┐
│ device-openapi │     │ device-gateway      │    │ Kafka / Redis    │
│ (协议接入)      │     │ (TCP 设备长连接)     │    │ Nacos / MySQL    │
└───────┬────────┘     └──────────┬──────────┘    └──────────────────┘
        │                         │
        └────────────┬────────────┘
                     │
          ┌──────────▼──────────┐
          │ device-consume      │
          │ device-worker       │
          │ device-paas         │
          └─────────────────────┘
                     │
              ECU / 物联网设备
```

基础设施常见组合：

- **服务发现 / 配置**：Nacos（兼容原 Java 侧环境变量与配置习惯）
- **缓存 / 会话**：Redis
- **消息**：Kafka
- **存储**：MySQL（部分分析链路可接 Elasticsearch）
- **运行时**：Docker + Kubernetes
- **HTTP 框架**：Gin


---

## 仓库结构（Go 服务）

当前工作区为多服务并列目录，GitHub 上的 `ebike-go` 可按同一布局组织为多模块仓库（每个服务独立 `go.mod`）。

| 目录 | 角色 | 说明 |
|------|------|------|
| `ebike-gateway-go` | API 网关 | 合并原 `gateway-business` / `gateway-client`；路由、签名、JWT、RBAC、Prometheus |
| `ebike-auth-go` | 鉴权网关 | 合并原 auth business/client；JWT / OAuth 风格登录、多端会话、微信等扩展 |
| `identity-auth-go` | 实名认证 | 身份核验相关能力（对接管理侧业务） |
| `ebike-service-client-go` | C 端 BFF / 聚合 | 用户、订单、支付、营销、运营等 C 端聚合 |
| `ebike-service-business-go` | 管理端 BFF / 聚合 | 运营后台聚合与网关逻辑 Go 化 |
| `ebike-fence-go` | 地理围栏 | 用车/还车边界、距离判定等高并发几何计算 |
| `ebike-analyze-go` | 分析查询 | 运营分析、检索相关接口（ES / MySQL 等） |
| `ebike-device-gateway-go` | 设备 TCP 网关 | 车辆/ECU 长连接接入、指令下发、连接诊断 |
| `ebike-device-openapi-go` | 设备协议入口 | 第三方/厂商协议解码并投递 Kafka |
| `ebike-device-consume-go` | 设备消息消费 | 消费上行数据，写 Redis / 推业务事件 |
| `ebike-device-worker-go` | 设备异步 Worker | Kafka 消费与后台任务 |
| `ebike-device-paas-go` | 设备 PaaS | 设备查询、控制、状态/轨迹等聚合（含 Shadow 比对） |
| `push-notification-go` | 消息推送 | 短信、推送等多通道异步投递 |
| `docs/` | 协议与交付文档 | MQTT 协议、DDL、部署相关说明等 |


---


## 环境要求

- **Go** 1.24+（多数模块声明为 1.25.x 附近，以各子目录 `go.mod` 为准）
- Docker（可选，用于镜像构建）
- 可访问的 Nacos / Redis / MySQL / Kafka（按你要启动的服务组合准备）
- Kubernetes（推荐生产部署方式）

---

## 快速开始（单服务）

各子项目均是独立 Go module，进入目录后即可构建。以网关为例：

```bash
cd ebike-gateway-go
go mod tidy

# 交叉编译 Linux amd64（常见 K8s / Alpine 场景）
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o build/ebike-gateway-go .

# 本地测试
go test ./...
```

鉴权服务双模式启动示例：

```bash
cd ebike-auth-go
export AUTH_SERVICE_MODE=client   # 或 business
export PORT=8080
# 配置 SPRING_CLOUD_NACOS_* 或本项目兼容的 Nacos 环境变量
go run .
```

设备网关压测工具（若该服务提供）：

```bash
cd ebike-device-gateway-go
go build -o build/loadtest ./cmd/loadtest
./build/loadtest -c 1000 -h 127.0.0.1 -p 9880
```

镜像与 K8s 清单一般在对应服务目录下的 `Dockerfile`、`deploy*.yaml` / `deployment.yaml`。**推送到公共仓库前请改掉示例镜像地址与命名空间。**

更细的环境变量、路由与监控说明见各子目录 `README.md`。

---



## 贡献

欢迎 Issue 与 PR（Bug 修复、文档、可观测性、无密钥的示例配置）。

在提交前请确认：

- 未包含密钥、证书、生产配置缓存  
- 通过目标服务的 `go test`（若有）  
- 变更说明写清兼容性影响（尤其是协议与签名）

贡献代码默认按本仓库 [Elastic License 2.0](./LICENSE) 授权给项目版权方；大范围功能贡献可能需另行签署 CLA（按你们实际政策补充）。

---

## 许可证

本项目 **源代码** 在 [**Elastic License 2.0**](./LICENSE) 下提供。

### 为什么选这个协议

需求是：**客户可把代码用于自家公司业务，但不能拿去对外提供商业服务**。

| 协议 | 是否适合 | 说明 |
|------|----------|------|
| MIT / Apache-2.0 / BSD | ❌ | 经典开源，**允许**任意商用与 SaaS |
| GPL / AGPL | ❌ | 限制再分发方式，但**不禁止**用软件对外提供商业服务 |
| SSPL | ⚠️ 可选偏强 | 强调“提供服务须开源管理栈”，争议大、社区工具支持差 |
| **Elastic License 2.0（推荐）** | ✅ | **明确禁止**把软件作为托管/管理型服务向第三方提供；仍允许自用、修改、分发源码 |
| PolyForm Shield | ✅ 备选 | 禁止用于提供与版权方**竞争**的产品/服务，表述更偏竞业 |

**结论：采用 Elastic License 2.0（EL2.0）。**

请注意：

- EL2.0 属于 **源码开放（Source Available）**，**不是** OSI 定义的 “Open Source”。GitHub 上请用 “Source available under Elastic License 2.0” 描述，避免写成 “MIT/开源无限制”。
- 依赖库（Gin、Kafka 客户端、Nacos SDK 等）仍受**各自原许可证**约束；本文件仅覆盖本仓库自有代码。
- 许可证**不能替代**合同：面向签约客户时，仍建议在采购/交付合同中写清使用范围、禁止再 SaaS、审计与违约责任；需要对外商业化再卖的客户，走**双许可（商业许可证）**。
- 请将 `LICENSE` 顶部 `Copyright (c) 2026 ebike-go contributors` 改为贵司法人全称（正文 EL2.0 条款保持不变）。

### 权限摘要（非法律意见）

**允许**

- 阅读、下载、修改源码  
- 部署运行于贵司自有或贵司租用基础设施上，服务**贵司自有业务**  
- 在满足 EL2.0 的前提下向他人提供副本及修改版（须附带协议文本）

**禁止**

- 将本软件（或其主要功能）作为 **hosted / managed service** 向第三方用户提供  
- 绕过许可证相关机制或删除版权与许可声明  

正式条款以 [LICENSE](./LICENSE) 全文为准。如需英文对外说明，可使用：

```text
Licensed under the Elastic License 2.0 (the "License");
you may not use this file except in compliance with the License.
You may not provide this software to third parties as a hosted or managed service.
```

---

## 免责声明

软件按 “现状” 提供，作者与版权方不对部署后的业务中断、数据丢失或合规结果承担责任。生产使用前请自行完成安全审计、压测与灾备方案。

---

## 联系

- 仓库：计划发布为 GitHub 项目 **`ebike-go`**
- 商业授权 / 合作：请在仓库 Issues 中说明需求
