FROM golang:alpine AS builder

# 禁用 CGO
ENV CGO_ENABLED=0
RUN apk --no-cache add tzdata
# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码并构建应用
COPY . .
RUN go build -ldflags "-s -w" -o /app/deeplx-pro .

# 运行: 使用scratch作为基础镜像
FROM scratch as prod
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# 在build阶段, 复制时区配置到镜像的/etc/localtime
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译好的应用和资源
COPY --from=builder /app/deeplx-pro /app/deeplx-pro

# 暴露端口
EXPOSE 9000

CMD ["/app/deeplx-pro"]