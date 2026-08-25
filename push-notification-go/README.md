# Push-Notification (Go Edition)

该服务主要负责多渠道（短信、语音）消息触达，整合了阿里云和创蓝等多家外部短信/语音通道。通过迁移到 Go 语言并采用协程池模型，大幅降低了系统在发送消息高峰期等待第三方 API 响应时的内存和线程调度开销，提升了整体系统的发送吞吐量和资源利用率。


## 目录结构

- `/build` - 包含项目的 `Dockerfile` 构建脚本
- `/conf` - 存放本地默认引导配置文件 `application.yml`
- `/internal/cache` - 基于 `atomic.Value` 的本地内存缓存层，支持高并发无锁读取及定时热刷新
- `/internal/handler` - Gin 路由接口及全局 Middleware 中间件
- `/internal/model` - 数据库 GORM 数据模型及 DTO
- `/internal/pkg` - 公共工具与基础设施抽象（Nacos, MySQL, Config, Errcode 等）
- `/internal/pool` - 全局协程工作池，负责异步削峰
- `/internal/sender` - 渠道投递适配器策略实现（阿里云、创蓝等）
- `/internal/service` - 核心业务编排层
- `/main.go` - 程序启动主入口

## 1. 编译构建

本项目可以编译为跨平台的二进制文件，最终部署于 Linux 容器环境。

**在 Windows PowerShell 中交叉编译（Linux/amd64）：**
1. 创建构建输出目录：
   ```powershell
   mkdir build/bin
   ```
2. 执行编译命令：
   ```powershell
   $env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"; go build -o build/bin/push-notification-go ./main.go
   ```

**在 Linux/macOS 环境中编译：**
```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o build/bin/push-notification-go ./main.go
```

## 2. Docker 打包部署

使用 `build` 目录下的 `Dockerfile` 可以轻松将应用容器化。建议由 CI/CD 流水线统一打包。

若需手动打包并推送到阿里云镜像仓库：

```bash
export VERSION="v1.0.0"
docker build -t registry.cn-shanghai.aliyuncs.com/ebike_luoping/push-notification-go:$VERSION -f build/Dockerfile .
docker push registry.cn-shanghai.aliyuncs.com/ebike_luoping/push-notification-go:$VERSION
```

*(部署时请修改 `deployment.yaml` 为对应的镜像版本，并用 `kubectl apply -f deployment.yaml` 在 Kubernetes 集群中更新)*

## 3. 配置中心与优先级加载机制

系统配置采用了**三层优先级设计**，优先级从低到高如下：

1. **底层本地默认配置**：从 `conf/application.yml` 加载默认参数，保证应用基本可启动。
2. **Nacos 云端配置（核心）**：
   - 启动时自动从 Nacos (`group: xyy_ops`) 拉取 `mysql.yaml` 并解析映射数据库链接。
   - 随后从 Nacos (`group: xyy`) 拉取 `push-notification.yml` 覆盖 secrets 秘钥和线程池参数。
   - 配置通过 `ListenConfig` 支持热更新。
3. **环境变量覆盖（最高优）**：可通过宿主机注入 `MYSQL_DSN`, `NACOS_SERVER_ADDR`, `XYY_SECRETS` 等系统级环境变量进行最高优先级覆盖，便于在容器编排时进行快速干预。
