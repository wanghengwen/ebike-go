FROM alpine:3.20

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk --no-cache add ca-certificates tzdata curl
ENV TZ=Asia/Shanghai

WORKDIR /app

# Build locally before docker build:
#   GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o build/ebike-open-paas-go .
COPY build/ebike-open-paas-go /app/ebike-open-paas-go
COPY conf/application-k8s.yml /app/conf/application.yaml

RUN chmod +x /app/ebike-open-paas-go

EXPOSE 8080

ENTRYPOINT ["./ebike-open-paas-go"]
