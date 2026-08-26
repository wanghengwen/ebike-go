# fastid

`fastid` 是 Go 语言实现的机器 ID 分配服务。它为分布式唯一 ID 生成器（如 Snowflake 算法）提供机器 ID（WorkId）的在线分配、去重、状态心跳管理和持久化。

## 目录结构

- `/main.go` - 程序启动入口
- `/conf` - 本地引导配置（Nacos 地址等，运行时可由 Nacos / 环境变量覆盖）
- `/internal/api/controller` - Gin 路由与控制器
- `/internal/service` - 业务逻辑层
- `/internal/repository` - 数据访问层
- `/internal/model` - 实体定义
- `/internal/pkg/config` - 配置加载与 Nacos 注册


---

## ⚙️ 配置文件说明

本地运行时会读取 `conf/application.yml`。在 K8S 容器中部署时，配置项可通过环境变量覆盖；也可通过 `CONFIG_PATH` 指定配置文件路径。

### 1. 配置文件模版 (`conf/application.yml`)
```yaml
server:
  port: 8080                          # 服务端口

nacos:
  server-addr: "127.0.0.1:8848"       # Nacos 服务器地址
  namespace: ""                       # Nacos 命名空间
  group: "DEFAULT_GROUP"              # Nacos 分组
  app-name: "luopingtech-fastid"      # Nacos 注册的服务名，用于拼接 DataId (默认拉取 luopingtech-fastid.yml)
  discovery:
    enabled: false                    # 是否注册到 Nacos 发现中心
  config:
    enabled: false                    # 是否从 Nacos 拉取数据库及密钥配置

database:                             # 本地数据库配置（在 Nacos 禁用时作为 Fallback）
  host: "127.0.0.1"
  port: 3306
  name: "fastid_db"
  username: "fastid"
  password: "testfastid321#"
  params: "charset=utf8mb4&parseTime=True&loc=Local"

fastid:
  secrets:                            # API 鉴权密钥白名单
    - "xxxxxxxxxxxxxxxxxxx"
```


---

## 🛠️ 本地运行

在项目根目录下，确保本地 Go 版本为 1.24+：

1. **下载依赖**：
   ```bash
   go mod download
   ```
2. **运行服务**：
   ```bash
   go run .
   ```
3. **执行测试**：
   ```bash
   go test ./...
   ```

---

## 🐳 Docker 镜像构建（`build_image.sh` 方式）

为了能在 Linux/WSL 环境下快速编译并打包出极小的生产级 Docker 镜像，项目提供了 `build_image.sh` 编译打包脚本。

### 1. 前提条件
- 运行环境为 **Linux** 或 **Windows WSL** 终端。
- 宿主机已安装 **Go**（版本 >= 1.24）。
- 宿主机已安装并启动 **Docker** 服务。

### 2. 构建步骤
在项目根目录下执行以下命令：

```bash
# 1. 赋予脚本执行权限
chmod +x build_image.sh

# 2. 运行构建脚本
./build_image.sh
```

### 3. 构建流程说明
该脚本会自动执行以下步骤：
1. **交叉编译**：调用宿主机的 Go 编译器，针对 Linux AMD64 进行静态链接编译，生成二进制可执行文件 `register`。
2. **环境检查**：检查宿主机 Docker daemon 是否在运行，并检测是否需要 `sudo` 权限。
3. **镜像打包**：将生成的二进制文件及本地 `bootstrap.yaml` 复制到微型 `alpine:3.20` 镜像中，并自动初始化 `Asia/Shanghai` 生产时区。
4. **输出结果**：生成大小仅约 30MB 的轻量镜像 `luopingtech-fastid-register:latest`。

---

## 🚀 Kubernetes 部署

1. **修改镜像 Tag 并推送镜像**：
   ```bash
   docker tag luopingtech-fastid-register:latest YOUR_REGISTRY/luopingtech-fastid-register:latest
   docker push YOUR_REGISTRY/luopingtech-fastid-register:latest
   ```
2. **修改部署配置**：
   打开项目根目录下的 `k8s_deploy.yaml`，将镜像地址、Nacos 地址及命名空间等替换为你的实际环境参数。
3. **应用部署**：
   ```bash
   kubectl apply -f k8s_deploy.yaml
   ```
