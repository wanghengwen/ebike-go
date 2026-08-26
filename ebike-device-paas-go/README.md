# ebike-device-paas (Go Edition)

原 Java `ebike-device-paas` 服务的 Go 重构版本。该服务是设备域的 PaaS 聚合层，
负责设备信息查询、设备控制、蓝牙指令上报、状态变更与轨迹代理。

迁移采用**查询类优先 + Shadow 比对**策略，详见 [docs/PLAN.md](docs/PLAN.md)。

## 目录结构

```
internal/
  pkg/logger     zap 日志
  pkg/config     配置加载（yaml + env；Nacos 待接入）
  pkg/jsondiff   JSON 语义比对（含易变字段忽略 + 浮点容差）
  pkg/shadow     影子流量比对 Java（仅查询类）
  middleware     trace / tenant 上下文
  api            统一 Result + 错误码 + 路由
  protocol       DeviceProtocol 定长编解码（核心，1:1 移植 Java）
tools/logextract 生产日志提取器 + Lombok toString 解析器
```

## 构建与测试

```bash
go build ./...
go test ./...
```

交叉编译（K8s linux/amd64）：

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o build/ebike-device-paas-go .
```

## 从生产日志提取测试数据

```bash
go run ./tools/logextract -in ../ebike-device-paas/prod_*.log -out testdata -max 30
```

每个接口生成若干 `testdata/<endpoint>/NNN.json`，含 `url / req / rep / clazz / cusTime`。
`req` 为合法 JSON，`rep.data` 由 Java Lombok toString 转换为 JSON。

## 运行（本地 / 影子模式）

```bash
# 普通模式
go run .

# 影子模式：处理后镜像到 Java 并比对响应
DRY_RUN=true JAVA_SERVICE_URL=http://old-paas:8080 go run .
```
