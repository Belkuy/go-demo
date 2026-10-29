# ---- 构建阶段 ----
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 先拷贝 go.mod，利用层缓存加速依赖下载
COPY go.mod ./
RUN go mod download

COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w -X main.version=$(date +%Y%m%d-%H%M)" -o /go-demo .

# ---- 运行阶段 ----
FROM alpine:3.20

# 非 root 用户运行，安全基线
RUN adduser -D -u 10001 appuser
USER appuser

COPY --from=builder /go-demo /usr/local/bin/go-demo

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/go-demo"]
