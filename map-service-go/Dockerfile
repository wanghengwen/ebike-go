FROM alpine:3.20

# Use Aliyun mirror for alpine packages to speed up builds
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# Install tzdata and curl (preStop deregister hook)
RUN apk --no-cache add tzdata curl
ENV TZ=Asia/Shanghai

WORKDIR /app

# Copy the pre-compiled binary from the build directory
COPY build/map-service /app/map-service

# Create log directory
RUN mkdir -p /logs

EXPOSE 8080

ENTRYPOINT ["/app/map-service"]
