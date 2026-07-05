# KnowSync Gateway Docker 镜像
# 基于编译好的二进制文件构建轻量级镜像

FROM alpine:latest

# 构建参数
ARG BINARY_NAME=gateway-latest-linux-amd64
ARG SERVICE_NAME=gateway

# 安装基础依赖
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建应用目录
WORKDIR /app

# 复制二进制文件
COPY bin/${BINARY_NAME} /app/${SERVICE_NAME}

# 复制配置文件（如果存在）
COPY configs/ /app/configs/

# 设置权限
RUN chmod +x /app/${SERVICE_NAME}

# 暴露端口（Gateway 通常使用 80/443 或 8080）
EXPOSE 8080
EXPOSE 80
EXPOSE 443

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD pgrep ${SERVICE_NAME} || exit 1

# 启动命令
ENTRYPOINT ["sh", "-c", "umask 000 && exec /app/gateway"]
