# Ebike Fence (Go Edition)

该服务专注于两轮车地理围栏系统，主要负责最高并发和性能要求的核心业务接口（例如：用车边界校验 `ridingCar`、还车距离判定 `returnCar`、围栏关系查询 `getFenceRelation`）。


## 目录结构

- `/cmd` - 程序启动入口（包含 `main.go`）
- `/conf` - 配置文件及 Nacos 映射配置（`application.yml`）
- `/internal/api` - Gin 路由、DTO 结构体和控制器定义
- `/internal/domain` - 核心业务领域服务（围栏计算、车辆部位与工单检查状态机）
- `/internal/infrastructure` - 基础设施层（MySQL 仓储、RPC 调用）
- `/internal/middleware` - 中间件（日志拦截，透传请求上下文本）
- `/internal/pkg` - 公共工具类（错误处理、Redis、Config 等）

## 1. 编译构建 (Windows / WSL)

本项目针对 Kubernetes 环境部署，需要编译为 `linux/amd64` 架构的二进制文件。

**在 Windows PowerShell 中编译：**
1. 创建构建输出目录：
   ```powershell
   mkdir build
   ```
2. 执行交叉编译命令，将二进制生成到 `build` 目录下：
   ```powershell
   $env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"; go build -o build/ebike-fence-go ./cmd
   ```

**在 WSL / Linux 环境中编译：**
1. 创建构建输出目录：
   ```bash
   mkdir -p build
   ```
2. 执行编译命令：
   ```bash
   GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o build/ebike-fence-go ./cmd
   ```

## 2. Docker 打包与自动发布

如果项目中包含自动部署脚本 `deploy.sh`，则可以通过该脚本自动完成**镜像打包、推送以及配置传输**。或者，在终端中手动执行以下命令：

```bash
export version="v1.0.0"
docker build -t registry.cn-shanghai.aliyuncs.com/ebike_luoping/ebike-fence-go:"$version" . && \
docker push registry.cn-shanghai.aliyuncs.com/ebike_luoping/ebike-fence-go:"$version" 
```

> **注意：**
> 在打包前，请确保 `deployment.yaml` 中的 `image` 版本号已经与你想推送的版本号（如 `v1.0.0`）保持一致。

## 3. Kubernetes 部署

将最新的 `deployment.yaml` 传送到服务器后，在服务器内部执行应用部署。

1. 登录目标服务器：

2. 部署到 Kubernetes 集群：
   ```bash
   kubectl apply -f /tmp/deployment.yaml
   ```

3. **查看部署状态：**
   ```bash
   kubectl get pods -n prod -l app=ebike-fence-go
   kubectl logs -f -n prod -l app=ebike-fence-go
   ```

## 4. 配置中心

服务启动时读取本地 `conf/application.yml` 进行引导，核心配置（Redis、MySQL 数据源、系统基础配置等）统一通过 Nacos 配置中心自动拉取。
* **Nacos 地址**: 在 `application.yml` 或系统环境变量中配置。
* **命名空间**: `prod`
* **Data ID**: `ebike-fence-go.yaml`
