FROM alpine:3.20

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

# Copy the pre-compiled Go binary from the build directory
# Ensure you have compiled it for linux/amd64 before running docker build!
COPY ./build/ebike-device-openapi-go /app/ebike-device-openapi-go

# Local bootstrap config (base layer below Nacos + env). The k8s profile is
# baked in as the default CONFIG_PATH (conf/application.yaml); Nacos overrides it.
COPY conf/application-k8s.yml /app/conf/application.yaml

# Add execute permissions just in case
RUN chmod +x /app/ebike-device-openapi-go

EXPOSE 8080

ENTRYPOINT ["./ebike-device-openapi-go"]
