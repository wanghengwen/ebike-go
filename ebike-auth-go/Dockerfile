FROM alpine:3.20

# 修改 Alpine 源为阿里云镜像源，大幅提升国内 apk 安装速度
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 确保时区为上海，并安装CA证书
RUN apk --no-cache add tzdata ca-certificates && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 拷贝由外部手动编译好的Go二进制文件 (编译命令例如: GOOS=linux GOARCH=amd64 go build -o build/ebike-auth-go main.go)
COPY build/ebike-auth-go .

# 拷贝静态配置文件和公钥文件 (在 K8s 中更推荐通过 ConfigMap 挂载)
COPY conf/application.yml /app/conf/application.yml
COPY public-keys-apple.json .
COPY public-keys-google.json .

# 赋予二进制执行权限（可选防错机制）
RUN chmod +x ./ebike-auth-go

# 暴露端口 (可通过外部环境变量 PORT 覆盖)
EXPOSE 8080

# 启动命令
CMD ["./ebike-auth-go"]
