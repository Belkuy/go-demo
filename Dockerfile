# =====================================================================
# go-demo 多阶段构建 Dockerfile
#
# 思路：用带 Go 工具链的大镜像（约 300MB）编译，只把二进制拷进
# 极简的运行镜像（alpine），最终镜像约 21MB，且不含编译器和源码。
# =====================================================================

# ---------- 第一阶段：构建 ----------
# 起个别名 builder，供后面的 COPY --from=builder 引用
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 先只拷贝 go.mod 再下载依赖：
# Docker 按层缓存，只要 go.mod 没变，这一层就命中缓存，不用重新下载依赖
COPY go.mod ./
RUN go mod download

# 再拷贝源码并编译
# CGO_ENABLED=0     : 纯静态编译，避免依赖 alpine 里的 glibc
# GOOS=linux        : 明确目标平台，防止在 Mac 上构建出 darwin 二进制
# -s -w             : 去掉符号表和调试信息，减小体积
# -X main.version=… : 把版本号注入 main.go 里的 version 变量，无需改代码
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w -X main.version=$(date +%Y%m%d-%H%M)" -o /go-demo .

# ---------- 第二阶段：运行 ----------
# 运行基础镜像只有约 4MB
FROM alpine:3.20

# 创建 uid=10001 的普通用户，避免以 root 身份运行容器（安全基线要求）
RUN adduser -D -u 10001 appuser
USER appuser

# 从构建阶段只拷贝编译好的二进制文件，源码和 Go 工具链不会进入最终镜像
COPY --from=builder /go-demo /usr/local/bin/go-demo

# 声明服务端口（仅作文档提示，实际映射由 docker run -p 决定）
EXPOSE 8080

# 用 exec 形式启动，进程 PID 为 1，能正确接收 SIGTERM 实现优雅停止
ENTRYPOINT ["/usr/local/bin/go-demo"]
