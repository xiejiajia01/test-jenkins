# 使用多阶段构建
# 第一阶段：构建
FROM c-xie-create-registry.cn-beijing.cr.aliyuncs.com/xiejaijia/docker:golang1.21 AS builder

# 设置工作目录
WORKDIR /app

# 设置 GOPROXY 环境变量
ENV GOPROXY=https://proxy.golang.org,direct

# 安装必要的系统依赖
RUN apk add --no-cache gcc musl-dev

# 复制go.mod和go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# 第二阶段：运行
FROM alpine:latest

# 安装ca-certificates，用于HTTPS请求
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 从builder阶段复制编译好的应用
COPY --from=builder /app/main .

# 暴露端口
EXPOSE 8001

# 运行应用
CMD ["./main"]
