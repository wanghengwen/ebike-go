# ebike-open-paas-go

面向第三方的小安 PaaS v1.1.13 兼容适配层（薄契约网关）。

- 对外：`https://paas.luopingtech.com` → `/ebike/v1/*` + `POST /ebike/api/device`，Header `xc-access-token` + `agentId`
- 对内：转发 `ebike-device-paas` / `openapi` / `worker` / `anvelink-console` / `management` / `map-service`
- 回调：订阅 CRUD 落 Redis；本服务以独立 Kafka consumer group 直连 `saas_0` / `gray_saas_0`，映射后异步 POST 客户 URL（不依赖 `ebike-device-consume`）

## 目录结构

```
internal/
  api           Gin 路由与 Xiaoan 包络
  client        上游 HTTP（paas / openapi / worker / console / management / map）
  contract      对外错误码与响应形
  dispatcher    客户回调队列与重试
  event         saas_0 解析、payload、notify 码表、派生状态机
  middleware    鉴权 / 限流 / access log
  pkg/config    yaml + Nacos + env
  pkg/kafka     sarama 消费（独立 group）
  repository    回调订阅 / 内存订阅快照 / 限流
  service       查询与指令编排
```

## 本地运行

```bash
# 可选：跳过 Nacos
set NACOS_DISABLED=true
go run .
```

- 本地默认：`conf/application.yaml`（Kafka 默认关闭）
- K8s 引导：`conf/application-k8s.yml`（上游 ClusterIP + Kafka group；可 `CONFIG_PATH=...`）
- Nacos 应用配置模板：`conf/ebike-open-paas.yml` → DataId `ebike-open-paas.yml` / group `xyy`（agents、aes-key、限流、回调、notify）
- Redis / Kafka brokers·topics 由 Nacos `{group}_ops` 的 `redis.yaml` / `kafka.yaml` 合并

## 构建镜像

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o build/ebike-open-paas-go .
docker build -t registry.cn-shanghai.aliyuncs.com/ebike_luoping/ebike-open-paas-go:v1.0.0 .
```

## Kubernetes 部署

`deployment.yaml` 含 Deployment、Service、Ingress（`paas.luopingtech.com`）。

```bash
# 推镜像后更新 deployment.yaml 的 image tag，再 apply
kubectl apply -f deployment.yaml -n prod
kubectl get pods,svc,ingress -n prod -l app=ebike-open-paas-go
```

## Nacos 要点

| 项 | 值 |
|---|---|
| 服务名 | `ebike-open-paas` |
| DataId | `ebike-open-paas.yml`（group `xyy`；模板见 `conf/ebike-open-paas.yml`） |
| Redis / Kafka | `redis.yaml` / `kafka.yaml` @ `{group}_ops` |
| Redis | 与 device-paas 同库（读影子 + 写回调订阅） |
| 公网域名 | `paas.luopingtech.com`（Ingress） |

`open.agents` 示例：

```yaml
open:
  agents:
    - agentId: "87"
      token: "YOUR_XC_ACCESS_TOKEN"
      tenantId: "1"
      enabled: true
```

## 事件回调（Kafka）

- 订阅 topic：`saas_0` + `gray_saas_0`（与 worker 写出一致）
- `groupId` 必须含 `open-paas`（配置校验拒绝与 worker/consume 撞 group）
- 内存订阅快照：未注册租户热路径直接丢弃
- 可直接映射：PING / GPS / BMS / Alarm→Notify / Login·Logout→Notify
- Redis 状态机派生：出入围栏 (17/18)、SOC 阈值 (21/22)
- 投递按 URL 分片：某客户端点卡死只填满自己那条通道，不牵连其他租户
- 投递带 `X-Event-Id` / `X-Event-Attempt`，客户可据此去重（重试会重发完全相同的 body）
- 回调 URL 只要求 HTTPS（注册与 302 跳转均校验）；私网/集群内地址允许
- 连续失败熔断：默认同 URL 连续失败 20 次后屏蔽 10 分钟（不删订阅）；半开试探；重新注册立即复位
- 4xx（除 408/429）不重试
- 诊断：`GET /internal/callback/stats`，需 `X-Internal-Token`

环境变量（见 `deployment.yaml`）：`KAFKA_ENABLED` / `KAFKA_GROUP_ID` / `KAFKA_START_OFFSET`。

## 运维接口

| 路径 | 说明 |
|---|---|
| `GET /actuator/health/liveness` | 仅表示进程存活；依赖故障不重启 Pod |
| `GET /actuator/health/readiness` | Redis 或 Kafka 消费异常时返回 503，摘除流量 |
| `POST /actuator/deregisterService` | 仅接受回环地址（按 socket 对端判定，不看 Host 头），preStop 用 |
| `GET /internal/callback/stats` | 需 `X-Internal-Token`；`INTERNAL_TOKEN` 未设置则一律拒绝 |

Ingress 已开启 `force-ssl-redirect` 并在边缘屏蔽 `/internal/*`、`/actuator/*`；
`xc-access-token` 每次请求都在 Header 上，明文 HTTP 会泄露长期凭据，因此 TLS 证书
（`paas-luopingtech-com-tls`）与 `INTERNAL_TOKEN` Secret 需在 apply 前建好。

## 测试工具 `paas-test`

第三方联调 CLI：`cmd/paas-test`。完整步骤见 **[docs/paas-test.md](docs/paas-test.md)**。

```bash
go build -o build/paas-test ./cmd/paas-test
set PAAS_AGENT_ID=87
set PAAS_ACCESS_TOKEN=...
paas-test smoke
paas-test callback listen --listen :8089 --public-url https://hook.example.com/paas
```

## 未支持 / 弱支持

| 接口 | 状态 |
|---|---|
| `batteryPowerSwitch` | 不支持（无协议命令） |
| `lbs2gps` | 不支持（无基站定位服务） |
| `sms` | 不支持（无设备短信控车通道） |
| `allDevices` | 从 anvelink-console 按租户分页（含未绑车辆设备） |
| `batteryInfo` / `bmsInfo` | 弱：影子仅有 SN/SOC 等，其余字段省略而非补 0 |
| `currentBmsInfo` | 弱：仅 c34 能给出的 SOC/电压，`current`/`temperature`/`remain` 省略 |
| `GPSPoints` | 单次窗口上限 7 天，超出返回 105 |
| event=5 UART | 不支持（上报链路未解码），注册时直接拒绝 |

详见 `docs/CONTRACT.md`。
