#!/bin/bash

# 设置部署版本号 (由外部参数传入)
if [ -z "$1" ]; then
    echo "错误：请传入版本号参数！"
    echo "用法：./deploy.sh <version> (例如：./deploy.sh v1.0.2)"
    exit 1
fi
export version="$1"

# 获取当前项目目录名称 (例如: ebike-device-openapi-go)
PROJECT_NAME=$(basename "$PWD")
IMAGE_NAME="registry.cn-shanghai.aliyuncs.com/ebike_luoping/${PROJECT_NAME}:${version}"

echo "======================================================"
echo "跨平台编译 Go 二进制文件 (linux/amd64)..."
echo "======================================================"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/ebike-device-openapi-go main.go
if [ $? -ne 0 ]; then
    echo "Go 编译失败，停止执行。"
    exit 1
fi

echo "======================================================"
echo "同步更新 deployment.yaml 中的镜像版本号为: $version"
echo "======================================================"
sed -i -E "s|(image:[ ]*registry\.cn-shanghai\.aliyuncs\.com/ebike_luoping/${PROJECT_NAME}:).*|\1${version}|g" deployment.yaml

echo "======================================================"
echo "开始构建并打包 Docker 镜像: $IMAGE_NAME"
echo "======================================================"

# 1. 构建镜像
docker build -t "$IMAGE_NAME" .
if [ $? -ne 0 ]; then
    echo "Docker 构建失败，停止执行。"
    exit 1
fi

echo "======================================================"
echo "开始推送镜像到阿里云..."
echo "======================================================"

# 2. 推送镜像
docker push "$IMAGE_NAME"
if [ $? -ne 0 ]; then
    echo "Docker 推送失败，停止执行。"
    exit 1
fi

echo "======================================================"
echo "上传 deployment.yaml 到目标服务器 ECU..."
echo "======================================================"

# 3. 复制 deployment.yaml 到目标服务器 (使用指定的 ssh key)
scp -i ~/luoping-ebike.pem deployment.yaml root@ecu.luopingtech.com:/tmp

if [ $? -eq 0 ]; then
    echo "======================================================"
    echo "打包及传输完成！"
    echo "请登录 root@ecu.luopingtech.com 执行: kubectl apply -f /tmp/deployment.yaml"
    echo "======================================================"
else
    echo "SCP 传输失败，请检查网络或密钥权限。"
    exit 1
fi
