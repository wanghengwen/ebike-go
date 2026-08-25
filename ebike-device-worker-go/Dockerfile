FROM alpine:3.20

# 修改 Alpine 源为阿里云镜像源，大幅提升国内 apk 安装速度
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 确保时区为上海，并安装 CA 证书与 curl（K8s preStop 注销 Nacos）
RUN apk --no-cache add tzdata ca-certificates curl && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 拷贝由外部手动编译好的 Go 二进制文件
# 提示：在宿主机打包镜像前，需要先使用以下命令进行编译：
#   CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/ebike-device-worker-go .
COPY build/ebike-device-worker-go .

# K8s 生产配置（对拍实例可改 CMD 为 application-k8s-shadow.yaml）
COPY conf/application-k8s.yaml conf/application-k8s.yaml
#COPY conf/application-k8s-shadow.yaml conf/application-k8s-shadow.yaml

RUN chmod +x ./ebike-device-worker-go

# 与 Java 版一致，对外暴露 8080
EXPOSE 8080

CMD ["./ebike-device-worker-go", "-config", "conf/application-k8s.yaml"]
