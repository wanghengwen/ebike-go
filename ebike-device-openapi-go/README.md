# Ebike Device OpenAPI (Go Edition)

本项目是原 Java `ebike-device-openapi` 服务的 Go 语言重构版本。
该服务主要负责两轮车设备的网关接入、数据解码（如 Bin66、Bin68、Bin70、Bin72 等 Xiaoan 协议）、并转发给下游的 Kafka 消费集群。
通过迁移到 Go 语言，我们极大降低了云端服务器的 CPU 与内存负载，提高了高并发情况下的吞吐能力，并完美兼容原有的各种边缘业务逻辑（包括特殊配置、加解密机制与老版本协议 BUG 容错）。

## 目录结构

- `/cmd` 或 `/main.go` - 程序启动入口
- `/internal/api` - Gin 路由与控制器定义 (包括解码、登录、登出、应答下发等接口)
- `/internal/decode` - 核心协议解码器实现 (Xiaoan 系列协议)
- `/internal/middleware` - 过滤器中间件 (跨云路由拦截等)
- `/internal/pkg` - 公共工具与配置模块 (Nacos, Kafka, Redis, Crypto, Utils)
- `/internal/service` - 业务逻辑层与并发锁控制

## 1. 编译构建 (Windows / WSL)

本项目针对 Kubernetes 环境部署，需要编译为 `linux/amd64` 架构的二进制文件。

**在 Windows PowerShell 中编译：**
1. 创建构建输出目录：
   ```powershell
   mkdir build
   ```
2. 执行交叉编译命令，将二进制生成到 `build` 目录下：
   ```powershell
   $env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"; go build -o build/ebike-device-openapi-go .
   ```

**在 WSL / Linux 环境中编译：**
1. 创建构建输出目录：
   ```bash
   mkdir -p build
   ```
2. 执行编译命令：
   ```bash
   GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o build/ebike-device-openapi-go .
   ```

## 2. Docker 打包与自动发布

本项目提供了一个自动化的部署脚本 `deploy.sh`，它会自动获取当前目录名作为项目名，并完成 **镜像打包、镜像推送、以及配置文件的远端传输**。

你可以直接运行 `deploy.sh`，或者在 WSL 中执行以下一行指令：

```bash
export version="v1.0.2"
docker build -t registry.cn-shanghai.aliyuncs.com/ebike_luoping/`basename "$PWD"`:"$version" . && \
docker push registry.cn-shanghai.aliyuncs.com/ebike_luoping/`basename "$PWD"`:"$version" && \
scp -i ~/luoping-ebike.pem deployment.yaml root@ecu.luopingtech.com:/tmp
```

> **注意：**
> 在打包前，请确保 `deployment.yaml` 中的 `image` 版本号已经与你准备推送的版本号（如 `v1.0.2`）保持一致。

## 3. Kubernetes 部署

在通过上一步（或 `deploy.sh`）将最新的 `deployment.yaml` 传送到 ECU 服务器的 `/tmp` 目录后，按照以下步骤使其生效：

1. 登录目标服务器：
   ```bash
   ssh -i ~/luoping-ebike.pem root@ecu.luopingtech.com
   ```

2. 部署到 Kubernetes 集群：
   ```bash
   kubectl apply -f /tmp/deployment.yaml
   ```

3. **查看部署状态：**
   ```bash
   kubectl get pods -n prod -l app=ebike-device-openapi
   kubectl logs -f -n prod -l app=ebike-device-openapi
   ```

## 配置项与环境变量

服务启动时优先从环境变量获取基础配置，核心业务配置会通过 Nacos 下发。

* `NACOS_SERVER_ADDR`: Nacos 注册中心地址 (如 `192.168.2.20:8848`)。
* `NACOS_NAMESPACE`: Nacos 的命名空间ID，默认为 `prod`。
* `GIN_MODE`: Gin 运行模式，推荐设为 `release` 提升性能。
