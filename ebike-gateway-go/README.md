# ebike-gateway-go

`ebike-gateway-go` 是使用 Go 语言（基于 Gin 框架和 `httputil.ReverseProxy`）开发的高性能 API 网关，用于**替换并合并**原先 Java 版的 `ebike-gateway-business` 与 `ebike-gateway-client`。

当前版本：**1.0.20**

通过单套二进制 + `GATEWAY_MODE` 环境变量，在同一 Docker 镜像下切换 business / client 两种网关角色。

---

## 核心能力

| 能力 | 说明 |
|------|------|
| 动态路由 | Nacos 热更新 `ebike-gateway-{mode}-route.yml` |
| JWT 鉴权 | 路由级 Security filter，设备 Redis 校验 |
| 请求签名 | SHA256 + `_t`/`_s`，与 Java 交叉测试验证 |
| RBAC | business 模式 API 权限校验（10s 刷新） |
| 流量镜像 | client 模式可选 Mirror 到 Go 影子服务 |
| Prometheus | `GET /actuator/prometheus` |
| traceId | 写入 Go context，日志与下游 header 传播 |
| 配置热更新 | `redis.yaml` / `gateway-secret.yaml` 变更自动重连 Redis |

---

## 编译

```bash
cd d:/workspace/ebike/ebike-gateway-go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o build/ebike-gateway-go main.go
```

---

## Docker 镜像

```bash
docker build -t registry.cn-shanghai.aliyuncs.com/ebike_luoping/ebike-gateway-go:v1.0.20 .
```

---

## K8s 部署

```bash
kubectl apply -f deploy-business.yaml
kubectl apply -f deploy-client.yaml
```

Pod 已配置 Prometheus scrape 注解（`prometheus.io/scrape=true`）。

---

## 环境变量

| 变量 | 示例 | 说明 |
|------|------|------|
| `GATEWAY_MODE` | `business` / `client` | **核心开关**，启动时校验，非法值拒绝启动 |
| `PORT` | `8080` | 监听端口 |
| `NACOS_SERVER_ADDR` | `192.168.2.20:8848` | Nacos 地址 |
| `NACOS_NAMESPACE` | `prod` | 命名空间 |
| `NACOS_GROUP` | `xyy` | 配置分组 |
| `NACOS_REGISTER_ENABLED` | `false` | 网关无需注册到 Nacos |
| `ENABLE_API_VERIFY` | `false` | business 模式 API 权限开关（与 Java 默认一致；生产见 `deploy-business.yaml`） |
| `EXCLUDE_APIS` | 空 | 仅 `ENABLE_API_VERIFY=true` 时需要；逗号分隔 RBAC 排除 API |
| `MANAGEMENT_URL` | `http://ebike-management:8080` | 仅开启 API 权限校验时使用 |
| `REDIS_REQUIRED` | `true` | client 模式 Redis 是否必须；business 始终必须 |
| `LOG_EXCLUDE_PATHS` | `/callback/pay/score/` | client 模式跳过详细 accessLog 的路径片段 |
| `MIRROR_ENABLED` | `true` | client 影子流量开关 |
| `MIRROR_TARGET` | `http://...` | 影子服务地址 |

---

## 监控

```bash
# 健康检查
curl http://localhost:8080/actuator/health

# Prometheus 指标
curl http://localhost:8080/actuator/prometheus
```

主要指标：`ebike_gateway_http_requests_total`、`ebike_gateway_http_request_duration_seconds`、`ebike_gateway_upstream_request_duration_seconds`

---

## 测试

```bash
# 全量测试（签名模块首次会自动 mvn package 构建 Java 参考 jar）
go test ./...

# 仅签名 Java 交叉验证
go test ./sign/... -v
```

---

## 中间件链

```
Recovery → Prometheus → CORS → MatchRoute → ReadBody → Trace → Log → Sign → Security → [Mirror] → ReverseProxy
```

---

## 错误响应格式

与 Java `Result` 一致：

```json
{"success":false,"code":"00008","msg":"无效签名","data":null}
```

---

## 变更记录

详见 [ChangeLog.md](./ChangeLog.md)。
