FROM alpine:3.20

# 修改 Alpine 源为阿里云镜像源，大幅提升国内 apk 安装速度
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 确保时区为上海，并安装CA证书
RUN apk --no-cache add tzdata ca-certificates && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 拷贝由外部手动编译好的 Go 二进制文件 
# 提示：在宿主机打包镜像前，需要先使用以下命令进行编译：
# CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/ebike-gateway-go main.go
COPY build/ebike-gateway-go .

# 拷贝基础配置文件
COPY application.yml .

# 赋予二进制执行权限（防错机制）
RUN chmod +x ./ebike-gateway-go

# 暴露端口 (可通过外部环境变量 PORT 覆盖，对应 k8s containerPort: 8080)
EXPOSE 8080

# 启动命令
CMD ["./ebike-gateway-go"]
