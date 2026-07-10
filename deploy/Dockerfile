# lychee Dockerfile — 多阶段构建（CGO + SQLite）
#   docker build -f deploy/Dockerfile -t lychee:latest .
FROM golang:1.26-alpine AS builder
WORKDIR /build
ENV CGO_ENABLED=1 GOOS=linux GOPROXY=https://goproxy.cn,direct
RUN apk add --no-cache gcc musl-dev sqlite-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o /out/lychee ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata sqlite-libs
WORKDIR /app
COPY --from=builder /out/lychee /app/lychee
COPY web/static /app/web/static
COPY web/templates /app/web/templates
COPY web/jet /app/web/jet
# 配置文件由 docker-compose volume 挂载
CMD ["/app/lychee", "-config", "/app/config/config.yaml"]
