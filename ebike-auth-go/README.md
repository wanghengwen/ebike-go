# Ebike Auth Go

`ebike-auth-go` 是共享电单车项目下鉴权网关服务。



## 目录结构

- `/main.go` - 程序启动入口
- `/conf` - 本地引导配置（Nacos 地址等，运行时由 Nacos 覆盖业务配置）
- `/internal/auth` - 鉴权业务与 HTTP 路由（登录、登出、JWT、第三方登录等）
- `/internal/pkg` - 公共模块（config、logger、redis、rpc、utils）

## 🌟 核心特性

* **单体双模架构**：不再需要分开维护两套工程。通过注入环境变量 `AUTH_SERVICE_MODE=business` 或 `AUTH_SERVICE_MODE=client`，单个 Go 二进制程序即可任意接管对应的管理端或 C 端用户体系。
* **极致性能与轻量化**：
  * **容器体积**：大幅降低至 ~20MB（基于 Alpine 镜像构建）。
  * **内存开销**：Go 版本常驻内存仅需 ~20MB，K8s `limits` 轻松下调至 `256Mi`。
  * **零冷启动**：启动耗时从原本的几十秒缩减到毫秒级，有效提升 K8s HPA（水平自动扩缩容）和异常重启的响应速度。
* **架构高度兼容**：
  * 无缝对接原有的 **Nacos 配置中心**与服务发现体系。
  * JWT 颁发逻辑、Payload 内容及公私钥加解密规则。
  * `CommandContext`（含租户信息、追踪链路等）实现与微服务体系无缝衔接传递。

---

## 🛠️ 本地编译构建

在提交代码至流水线之前，你可以通过以下命令在本地交叉编译用于生产环境的 Linux 二进制文件。

```bash
# 1. 确保拉取所有依赖
go mod tidy

# 2. 编译 Linux amd64 二进制可执行文件 (使用 -s -w 剔除符号表以缩减体积)
# 注意：生成的产物需要放到 build 目录下，因为 Dockerfile 默认从 build/ 目录拷贝
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o build/ebike-auth-go main.go

# 玉环公共电单车客户定制版（默认开启玉岛行登录扩展，仍可用 YUHUAN_YUDAOXING_CUSTOM 覆盖）
# GOOS=linux GOARCH=amd64 go build -tags=yuhuan_yudaoxing -ldflags="-s -w" -o build/ebike-auth-go main.go
```

## 🐳 Docker 镜像打包

本项目采用轻量级的单阶段 Alpine 构建（代码构建过程交由外部宿主机或 CI 管道提前完成），大大加速打包过程。

在项目根目录下执行以下命令：

```bash
# 打包镜像并打上指定的远端仓库 tag
docker build -t registry.cn-shanghai.aliyuncs.com/ebike_luoping/ebike-auth-go:v1.0.0 .

# 将镜像推送到阿里云 ACR
docker push registry.cn-shanghai.aliyuncs.com/ebike_luoping/ebike-auth-go:v1.0.0
```

## ☸️ Kubernetes (K8s) 部署指南

项目中已经内置了 `deploy-business.yaml` 和 `deploy-client.yaml` 两个配套的声明式配置文件，你可以直接应用于目标 `prod` 命名空间。

### 环境变量说明
无论是以何种模式部署，通过 K8s Env 传入以下核心变量即可控制 Go 服务的行为：
* `AUTH_SERVICE_MODE`: 必填参数，设置为 `business` 或 `client`。
* `PORT`: 服务监听端口，通常固定为 `8080`。
* `SPRING_CLOUD_NACOS_*`: 虽然是 Go 程序，但代码层做了反射兼容，以直接复用原 Java 体系的 Nacos 环境变量配置。
* `WEIXIN_PUBLIC_APP_ID` / `WEIXIN_PUBLIC_APP_SECRET`: 微信公众平台 JSAPI 签名（对应 `ebike-auth-client` `test` 分支 `/oauth/jsapi/signature`）。也可在 Nacos 配置：

```yaml
weixin:
  publicplatform:
    appid: wxXXXXXXXX
    appSecret: xxxxxxxx
feign:
  weixin:
    url: https://api.weixin.qq.com
```

### Client 模式接口

| 接口 | 说明 |
|------|------|
| `POST /oauth/token` | 登录签发 Token |
| `POST /oauth/logout` | 登出 |
| `POST /oauth/phoneLogout` | 手机号登出 |
| `POST /oauth/jsapi/signature` | 微信 H5 JS-SDK 签名（`wx.config` / 扫码） |

### 部署执行

**部署管理后台鉴权网关（Business）**：
```bash
kubectl apply -f deploy-business.yaml
```

**部署 C端用户小程序鉴权网关（Client）**：
```bash
kubectl apply -f deploy-client.yaml
```

部署完成后，可通过以下命令监控 Pod 状态：
```bash
kubectl get pods -n prod -l "app in (ebike-auth-business, ebike-auth-client)"
```

如果看到所有探针（Readiness & Liveness）成功通过，即代表 Go 版鉴权网关已正式接管整个微服务体系的门户认证任务！
