# 多阶段构建Dockerfile
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的系统依赖
RUN apk add --no-cache git ca-certificates tzdata

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o higress-agent ./cmd/agent

# 运行阶段
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 创建非root用户
RUN addgroup -g 1001 -S higress && \
    adduser -u 1001 -S higress -G higress

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/higress-agent .

# 复制配置文件
COPY --from=builder /app/configs ./configs

# 创建必要的目录
RUN mkdir -p /app/logs /app/data/knowledge && \
    chown -R higress:higress /app

# 切换到非root用户
USER higress

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

# 设置环境变量
ENV GIN_MODE=release
ENV TZ=UTC

# 启动应用
CMD ["./higress-agent"] 