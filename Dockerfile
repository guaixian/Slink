# 前端请在宿主机先构建，再打包镜像（本文件不在容器内跑 Node）：
#   cd SlinkWeb && pnpm install && pnpm run build
# 需要存在 SlinkWeb/dist（含 index.html 等）。

# 构建阶段：用 Alpine 替代 Debian，更轻量且漏洞更少
FROM golang:1.26-alpine AS backend

# 安装必要构建工具
RUN apk add --no-cache git ca-certificates tzdata \
    && update-ca-certificates

WORKDIR /src

# SQLite 使用 github.com/glebarez/sqlite（纯 Go / modernc），与 CGO_ENABLED=0 兼容
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOTOOLCHAIN=auto

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 验证前端构建产物
RUN test -d SlinkWeb/dist && test -f SlinkWeb/dist/index.html || (echo "缺少 SlinkWeb/dist，请先执行: cd SlinkWeb && pnpm run build" >&2 && exit 1)

# 纯 Go 静态编译，加上 -trimpath 进一步减小体积并提高可移植性
RUN go build -ldflags="-s -w -extldflags=-static" -trimpath -o /slink .

# 运行阶段：用 Alpine 替代 Debian slim，漏洞少 90%+
FROM alpine:latest

# 1. 升级所有系统包，修复已知漏洞（这步解决你截图里的 CVE 问题）
# 2. 安装必要运行时依赖（ca-certificates 用于 HTTPS 请求）
# 3. 创建非 root 用户（安全最佳实践）
RUN apk --no-cache upgrade \
    && apk add --no-cache ca-certificates tzdata \
    && addgroup -g 1000 slink \
    && adduser -D -h /app -s /sbin/nologin -G slink -u 1000 slink \
    && rm -rf /var/cache/apk/*

WORKDIR /app

# 创建必要目录并设置权限
RUN mkdir -p /app/config /app/data /app/static \
    && chown -R slink:slink /app

# 复制编译好的二进制和前端资源
COPY --from=backend /slink /app/slink
COPY --from=backend /src/SlinkWeb/dist /app/SlinkWeb/dist

USER slink
EXPOSE 8080
ENV GIN_MODE=release
ENV SLINK_ADDR=:8080
ENV DB_NAME=/app/data/data.db

# 健康检查（可选但推荐）
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["/app/slink"]