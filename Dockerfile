# 使用官方的 Go 运行时作为基础镜像
FROM golang:1.18-alpine as builder

# 设置工作目录
WORKDIR /app

# 将当前目录的内容复制到容器的工作目录中
COPY . .

# 设置环境变量以避免警告信息
ENV CGO_ENABLED=0

# 安装依赖
RUN go mod download

# 构建应用程序
RUN go build -o go-tiny .

# 使用 Alpine Linux 作为最终运行时的基础镜像
FROM alpine:latest

# 安装必要的依赖
RUN apk add --no-cache ca-certificates

# 设置工作目录
WORKDIR /app

# 将构建的应用程序复制到最终的镜像中
COPY --from=builder /app/go-tiny /app/go-tiny

# 创建一个数据卷来存放配置文件
VOLUME ["/app/config"]

# 设置端口
EXPOSE 8084

# 启动命令
CMD ["./go-tiny"]