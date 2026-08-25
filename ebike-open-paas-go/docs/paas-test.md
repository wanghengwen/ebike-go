# 开放平台联调测试指南

用仓库自带的 `cmd/paas-test`，以**第三方客户**身份调用 `https://paas.luopingtech.com`（可改）。

配套：`xiaoan-paas-api.md`（接口原文）、`CONTRACT.md`（本服务契约差异）。

---

## 1. 前置条件

| 项 | 说明 |
|---|---|
| Nacos `open.agents` | 已配置测试用 `agentId` / `token` / `tenantId`，且 `enabled: true` |
| 设备归属 | 测试 IMEI 已在 anvelink 导入到该 agent 映射的 `tenantId`（`device_ebike_{imei}.tenantId`），否则查询会 `1001` |
| 网络 | 能访问 PaaS 公网域名；回调联调还需要一台公网可达的 HTTPS 入口 |
| 构建 | Go 1.25+ |

```bash
cd ebike-open-paas-go
go build -o build/paas-test.exe ./cmd/paas-test   # Windows
# go build -o build/paas-test ./cmd/paas-test     # Linux/macOS
```

---

## 2. 鉴权与环境变量

对外凭证是 **`agentId` + `xc-access-token`**，不是内部 `tenantId`。  
`tenantId` 只在 Nacos 里做映射，测试工具不填。

| 环境变量 | 必填 | 默认 | 含义 |
|---|---|---|---|
| `PAAS_BASE_URL` | 否 | `https://paas.luopingtech.com` | 服务根地址 |
| `PAAS_AGENT_ID` | 是 | — | Xiaoan `agentId` |
| `PAAS_ACCESS_TOKEN` | 是 | — | Header `xc-access-token` |
| `PAAS_IMEI` | 查询/指令建议 | — | 15 位设备号 |
| `PAAS_CALLBACK_URL` | 回调联调 | — | **公网 HTTPS** 回调地址 |
| `PAAS_LISTEN` | 回调联调 | `:8089` | 本机监听地址 |

PowerShell 示例：

```powershell
$env:PAAS_BASE_URL = "https://paas.luopingtech.com"
$env:PAAS_AGENT_ID = "87"
$env:PAAS_ACCESS_TOKEN = "你的token"
$env:PAAS_IMEI = "865067022403441"

.\build\paas-test.exe smoke
```

也可用参数覆盖（token 更推荐环境变量或 `--token-file`，避免进 shell 历史）：

```bash
paas-test --base-url https://paas.luopingtech.com --agent-id 87 --token-file ./token.txt smoke
```

---

## 3. 推荐测试顺序

```text
① smoke（契约/鉴权，不动真车）
② query（只读）
③ callback CRUD / listen（事件推送）
④ cmd（真车指令，必须 --yes）
```

### 3.1 smoke — 冒烟

不依赖具体设备也能跑大部分检查；若设置了 `PAAS_IMEI`，会额外打一次 `deviceInfo`。

```bash
paas-test smoke
paas-test smoke -v          # 打印原始 HTTP 响应
```

覆盖点：

- 错误 token → `UNAUTHORIZED_ERROR`
- `allDevices` 返回 Xiaoan 包络
- `sms` / `lbs2gps` / `batteryPowerSwitch` 明确业务失败（不是裸 404）
- 注册 `event=5`（UART）被拒绝
- 注册 `http://…` 回调被拒绝（仅 HTTPS）

全部 `PASS` 且退出码 0 即通过。

### 3.2 query — 只读查询

```bash
paas-test query allDevices --pageSize 20 --pageNumber 1
paas-test query deviceInfo
paas-test query address
paas-test query battery
paas-test query bms
paas-test query currentBms
paas-test query realtime
paas-test query gps --from 1700000000 --to 1700003600
# from/to 省略时：最近 1 小时
paas-test query gps
```

关注：

- HTTP 状态多为 **200**，业务看 `success` / `error.errormessage`
- `deviceInfo` 未上报字段应**省略**，不应一律填 `0`
- `GPSPoints` 时间窗超过约 7 天应失败（`RANGE_TOO_LARGE` / 相关错误）

### 3.3 cmd — 真车指令（危险）

**默认拒绝发送**；必须显式 `--yes`。

```bash
paas-test cmd lock --locked 1 --yes
paas-test cmd acc --acc 1 --yes
paas-test cmd defend --defend 1 --yes
paas-test cmd backWheel --locked 1 --yes
paas-test cmd batteryCompartment --locked 1 --yes
paas-test cmd reboot --yes
paas-test cmd mc --speed 80 --yes
paas-test cmd deviceVoice --idx 1 --yes
paas-test cmd bluetooth --yes
```

成功时看 `data.code`（0 成功；108/109 等为设备侧结果，见 API 文档）。

---

## 4. 回调联调（重点）

PaaS 注册回调要求：

1. URL 必须是 **`https://`**
2. 必须能被集群**公网访问**（出站 POST）
3. 客户侧接口为 **POST**，收到后返回 2xx

本机工具只监听 HTTP；HTTPS / 证书 / 公网入口由你在**别处的反向代理**完成。

### 4.1 拓扑

```text
设备上报 → Kafka saas_0 → ebike-open-paas-go
                              │
                              │ POST https://hook.example.com/paas
                              ▼
                     公网反代（TLS 终止）
                              │
                              │ HTTP 转发，path 对齐
                              ▼
                     paas-test listen :8089/paas
```

反代示例（示意）：把 `https://hook.example.com/paas` 转到 `http://<本机IP>:8089/paas`。  
**path 必须一致**，否则收不到。

### 4.2 一键监听 + 自动注册

```bash
# 1) 先保证反代已通：浏览器或 curl 访问公网 URL 的 /healthz（若你把 healthz 也反代出来）
#    或：本机 curl http://127.0.0.1:8089/healthz → ok

paas-test callback listen \
  --listen :8089 \
  --public-url https://hook.example.com/paas \
  --events 1,2,3,4 \
  --timeout 30m
```

行为：

1. 本机监听 `PAAS_LISTEN` / `--listen`
2. 用 `--public-url` 向 PaaS 注册指定 `events`（默认 1–4）
3. 收到推送时打印 `X-Event-Id`、`X-Event-Attempt` 与 JSON body
4. Ctrl+C 或超时后，**默认自动注销**订阅

常用开关：

| 参数 | 含义 |
|---|---|
| `--events 1,2,3,4` | 注册的事件类型 |
| `--timeout 10m` | 到期退出；`0` 表示一直等到 Ctrl+C |
| `--expect 1` | 收到 N 条后成功退出 |
| `--unregister-on-exit=false` | 退出后保留订阅 |
| `--skip-register` | 只听不注册（URL 已手动注册时） |

事件类型（本平台可投递）：

| event | 含义 |
|---|---|
| 1 | PING |
| 2 | GPS |
| 3 | Notify |
| 4 | BMS |
| 5 | UART — **不可用**，注册会被拒绝 |

### 4.3 仅 CRUD（不启监听）

```bash
paas-test callback register --event 2 --url https://hook.example.com/paas
paas-test callback list --event 2
paas-test callback unregister --event 2 --url https://hook.example.com/paas
```

### 4.4 自检清单

1. `curl http://127.0.0.1:8089/healthz` → `ok`
2. 从外网或手机流量访问 `https://hook.example.com/paas` 对应健康检查 / 能打到本机（按你的反代配置）
3. `callback list` 能看到刚注册的 URL
4. 等设备上报，或在平台侧制造 GPS/PING；`listen` 终端出现 `── callback #1`
5. 核对 body：`{"success":true,"data":{"imei","event","data"}}`
6. 重试场景下同一逻辑事件的 `X-Event-Id` 应稳定，`X-Event-Attempt` 递增

---

## 5. 常见失败

| 现象 | 可能原因 |
|---|---|
| `UNAUTHORIZED_ERROR` | token / agentId 错误，或 Nacos 未启用该 agent |
| `IMEI_ILLEGAL` | 不是 15 位数字 |
| 设备不归属 / `1001` | 见 [TROUBLESHOOTING-1001.md](./TROUBLESHOOTING-1001.md)：多为 agent.tenantId ≠ `device_ebike_*.tenantId`，或 registryDb 不是 12 |
| 注册回调 `scheme must be https` | 用了 `http://` 或非 https |
| `listen` 已注册但收不到 | 反代 path 不一致、防火墙未放行、Kafka/订阅快照未刷新、该租户无设备上报；或该 URL 因连续超时被**临时熔断**（订阅仍在，`/internal/callback/stats` 的 `suppressed`/`tripped` 会涨；冷却结束或再次 `POST /callback` 同 URL 可恢复） |
| event=5 注册失败 | 预期行为（无 UART 源） |
| `stopped after N redirects` / HTTPS 也 308 | 外层已终结 TLS，Ingress 仍开着 `ssl-redirect`/`force-ssl-redirect`。应设为 `false` 后 apply；见下方「Ingress 308」 |
| Windows 下 token 鉴权一直失败 | `set PAAS_ACCESS_TOKEN="xxx"` 会把引号写进变量；改用 `set PAAS_ACCESS_TOKEN=xxx`（无引号） |
| Ingress 访问 `/internal/*` 404 | 预期：公网已屏蔽内部诊断口 |

### Ingress 308 死循环（paas.luopingtech.com）

现象：`curl -v https://paas.luopingtech.com/` 已是 TLS，仍返回：

```text
HTTP/2 308
location: https://paas.luopingtech.com
```

原因：公网 IP（如 `47.116.40.243`）上的负载均衡/边缘 nginx 已用 `*.luopingtech.com` 终结 HTTPS，再以 **HTTP** 转到集群 Ingress。Ingress 若开启 `ssl-redirect`，会认为「当前是明文」而 308 到 https，客户端再连回来 → 死循环。

处理（任选其一，推荐 1）：

```bash
# 1) 关掉该 Ingress 的 SSL 重定向（HTTPS 由边缘负责）
kubectl -n prod annotate ingress ebike-open-paas-go \
  nginx.ingress.kubernetes.io/ssl-redirect=false \
  nginx.ingress.kubernetes.io/force-ssl-redirect=false --overwrite

# 或 apply 仓库里已改好的 deployment.yaml 中的 Ingress
kubectl apply -f deployment.yaml -n prod

# 2) 确认生效
kubectl -n prod get ingress ebike-open-paas-go -o yaml | grep -A2 ssl-redirect
curl -sI https://paas.luopingtech.com/ | head -n 5
# 期望：不再是 308，而是 200 / 业务响应
```

不要在「边缘已 HTTPS、后端是 HTTP」的拓扑上再开 `force-ssl-redirect`。

集群内看投递计数（需 `INTERNAL_TOKEN`，且一般不能走公网 Ingress）：

```bash
curl -H "X-Internal-Token: $INTERNAL_TOKEN" \
  http://ebike-open-paas-go.prod.svc.cluster.local:8080/internal/callback/stats
```

---

## 6. 命令速查

```text
paas-test smoke [-v]
paas-test query allDevices [--pageSize N] [--pageNumber N]
paas-test query deviceInfo|address|battery|bms|currentBms|realtime [--imei]
paas-test query gps [--imei] [--from sec] [--to sec]
paas-test cmd lock|acc|defend|backWheel|batteryCompartment --<flag> 0|1 --yes
paas-test cmd reboot|bluetooth --yes
paas-test cmd mc --speed 0..100 --yes
paas-test cmd deviceVoice --idx N --yes
paas-test callback register|list|unregister ...
paas-test callback listen --listen :8089 --public-url https://... [--events ...]
```

全局帮助：`paas-test --help`。
