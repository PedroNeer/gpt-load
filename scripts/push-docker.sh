#!/bin/bash

# Docker 镜像推送脚本
# 使用方法: ./scripts/push-docker.sh [dockerhub|github|both] [version]

set -e

# 配置
DOCKERHUB_USERNAME="你的DockerHub用户名"
GITHUB_USERNAME="pedroneer"
IMAGE_NAME="gpt-load"
DEFAULT_VERSION="latest"

# 获取参数
REGISTRY=${1:-"both"}
VERSION=${2:-$DEFAULT_VERSION}

echo "🚀 开始推送 Docker 镜像..."
echo "镜像名称: $IMAGE_NAME"
echo "版本: $VERSION"
echo "目标仓库: $REGISTRY"

# 构建镜像
echo "📦 构建 Docker 镜像..."
docker build -t $IMAGE_NAME:$VERSION .

# 推送到 Docker Hub
if [ "$REGISTRY" = "dockerhub" ] || [ "$REGISTRY" = "both" ]; then
    echo "🐳 推送到 Docker Hub..."

    # 检查是否已登录
    if ! docker info | grep -q "Username"; then
        echo "请先登录 Docker Hub: docker login"
        exit 1
    fi

    # 打标签
    docker tag $IMAGE_NAME:$VERSION $DOCKERHUB_USERNAME/$IMAGE_NAME:$VERSION

    # 推送
    docker push $DOCKERHUB_USERNAME/$IMAGE_NAME:$VERSION

    echo "✅ 成功推送到 Docker Hub: $DOCKERHUB_USERNAME/$IMAGE_NAME:$VERSION"
fi

# 推送到 GitHub Container Registry
if [ "$REGISTRY" = "github" ] || [ "$REGISTRY" = "both" ]; then
    echo "📦 推送到 GitHub Container Registry..."

    # 检查是否已登录
    # if ! docker info | grep -q "ghcr.io"; then
    #     echo "请先登录 GitHub Container Registry:"
    #     echo "echo 你的token | docker login ghcr.io -u $GITHUB_USERNAME --password-stdin"
    #     exit 1
    # fi

    # 打标签
    docker tag $IMAGE_NAME:$VERSION ghcr.io/$GITHUB_USERNAME/$IMAGE_NAME:$VERSION

    # 推送
    docker push ghcr.io/$GITHUB_USERNAME/$IMAGE_NAME:$VERSION

    echo "✅ 成功推送到 GitHub Container Registry: ghcr.io/$GITHUB_USERNAME/$IMAGE_NAME:$VERSION"
fi

echo "🎉 镜像推送完成！"

# 显示使用说明
echo ""
echo "📖 使用说明："
echo "1. 在 Colima 中拉取镜像："
if [ "$REGISTRY" = "dockerhub" ] || [ "$REGISTRY" = "both" ]; then
    echo "   docker pull $DOCKERHUB_USERNAME/$IMAGE_NAME:$VERSION"
fi
if [ "$REGISTRY" = "github" ] || [ "$REGISTRY" = "both" ]; then
    echo "   docker pull ghcr.io/$GITHUB_USERNAME/$IMAGE_NAME:$VERSION"
fi
echo ""
echo "2. 运行容器："
if [ "$REGISTRY" = "dockerhub" ] || [ "$REGISTRY" = "both" ]; then
    echo "   docker run -d -p 8080:8080 --name gpt-load $DOCKERHUB_USERNAME/$IMAGE_NAME:$VERSION"
fi
if [ "$REGISTRY" = "github" ] || [ "$REGISTRY" = "both" ]; then
    echo "   docker run -d -p 8080:8080 --name gpt-load ghcr.io/$GITHUB_USERNAME/$IMAGE_NAME:$VERSION"
fi
