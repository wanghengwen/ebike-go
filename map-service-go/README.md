# map-service-go

`map-service-go` 是一个高性能的地图网关微服务，由原 Java Spring Boot 版本平滑迁移而来。它负责对接高德地图（AMap）与腾讯地图（TMap）的相关接口，并提供逆地理编码缓存、GeoHash 计算、路线规划算法还原以及统一鉴权和日志追踪功能。

---

## 🚀 核心功能

1. **多地图源网关适配**：
   - 动态路由与工厂模式，兼容高德和腾讯地图 API。
   - 支持经纬度转换、逆地理编码（Regeo）、批量逆地理编码、地理编码以及路线规划（Navigate）。
2. **高效 Redis 缓存**：
   - 使用自定义算法高精度生成 GeoHash（按指定位宽合并经纬度）。
   - 实现逆地理编码查询的 Redis 缓存拦截（`re_{geohash}`，TTL 10 天），加速常用位置的返回并减少第三方 API 调用开销。
3. **K8S & Nacos 动态集成**：
   - 使用 Nacos 动态热更新高德、腾讯的 API 密钥及鉴权白名单。
   - 支持基于 K8S `POD_IP` 环境变量的微服务注册与主动下线。
4. **高并发生产优化**：
   - **连接池与超时**：统一配置 Resty 客户端的 `http.Transport`（默认最大并发连接 200，Dial 连接超时 10s，空闲超时 90s，单次请求超时 10s）。
   - **线程安全**：全局动态配置使用 `sync.RWMutex` 读写锁进行并发安全保护。
   - **优雅停机**：监听 `SIGINT` 和 `SIGTERM` 信号，自动执行 Nacos 注销，并等待存量请求处理完成（最长 10 秒缓冲）。
5. **链路追踪日志**：
   - 支持从 HTTP Header 或 Body 中智能提取 `tenantId` 和 `traceId`，通过 Logrus Hook 自动注入结构化 JSON 日志中，以对齐 Java 端的 SLF4J MDC 追踪体系。

---

## 🛠️ 本地运行

要在本地启动服务进行开发或测试，可以直接使用 Go 命令行工具：

```bash
# 本地直接运行服务（默认监听端口: 8080）
go run main.go
```

> [!NOTE]
> 服务启动时会根据环境变量/配置自动连接 Nacos 和 Redis。若需要模拟 Kubernetes 环境的容器 IP 注册，可提前设置 `POD_IP` 环境变量。

---

## 📦 编译指南

为了保证构建出的二进制文件体积最小、无冗余符号调试信息，并且兼容 Linux 生产容器环境，请在项目根目录下执行以下交叉编译命令：

```bash
# 创建存放预编译产物的构建目录
mkdir build

# 交叉编译 Linux (AMD64) 平台的 Go 二进制文件，移除符号表和调试信息 (-s -w)
env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o build/map-service main.go
```

编译完成后，产物将生成在 `build/map-service` 路径。

---

## 🚢 部署指南

### 1. 镜像构建、推送与 YAML 传输一键命令

在项目根目录下（即 `/mnt/d/workspace/ebike-v3-upgrade/map-service-go`），执行以下命令进行镜像构建、推送以及部署配置文件的传输：

```bash
# 1. 设置镜像版本号
export version="v1.0.2"

# 2. 构建镜像、推送至阿里云私有镜像仓并 scp 传输部署配置文件
docker build -t registry.cn-shanghai.aliyuncs.com/ebike_luoping/`basename "$PWD"`:"$version" . && \
docker push registry.cn-shanghai.aliyuncs.com/ebike_luoping/`basename "$PWD"`:"$version" && \
scp -i ~/luoping-ebike.pem deployment.yaml root@ecu.luopingtech.com:/tmp
```

> [!NOTE]
> - `basename "$PWD"` 会动态获取当前目录名 `map-service-go` 作为镜像名称。
> - 镜像仓地址为阿里云华东 2（上海）私有仓库：`registry.cn-shanghai.aliyuncs.com/ebike_luoping/map-service-go`。
> - `scp` 命令将本地的 `deployment.yaml` 拷贝到远程服务器 `ecu.luopingtech.com` 的 `/tmp` 目录，以便在服务器上执行 Kubernetes 资源应用。

### 2. 服务器上应用部署

通过 SSH 登录到远程服务器，并使用上一步传输的 `deployment.yaml` 发布/更新 Kubernetes 中的部署：

```bash
# 登录目标服务器
ssh -i ~/luoping-ebike.pem root@ecu.luopingtech.com

# 应用最新的 deployment 配置
kubectl apply -f /tmp/deployment.yaml
```

`deployment.yaml` 包含了如下生产配置：
- **存活/就绪探针**：基于 Gin 暴露的 `/actuator/health/readiness` 探针检查。
- **Downward API 注入**：将物理容器的 Pod IP 自动映射为 `POD_IP` 环境变量，供 Go 服务做 Nacos 动态地址注册。
- **内存优化**：得益于 Go 语言优秀的内存控制，其资源 Limit 仅需配额限制在 `512Mi` 内即可顺畅平稳运行（原 Java 服务为 2G）。
