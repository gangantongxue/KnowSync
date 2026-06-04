# KnowSync Chat Server Docker 镜像
# 基于编译好的二进制文件构建轻量级镜像

FROM alpine:latest

# 构建参数
ARG BINARY_NAME=chat-server-latest-linux-amd64
ARG SERVICE_NAME=chat-server

# 安装基础依赖
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建应用目录
WORKDIR /app

# 复制二进制文件
COPY bin/${BINARY_NAME} /app/${SERVICE_NAME}

# 复制配置文件
COPY configs/ /app/configs/

# 设置权限
RUN chmod +x /app/${SERVICE_NAME}

# 暴露 gRPC 端口和 WebSocket 端口
EXPOSE 50054 50055

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD pgrep ${SERVICE_NAME} || exit 1

# 启动命令
ENTRYPOINT ["/app/chat-server"]
