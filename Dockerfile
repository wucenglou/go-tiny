FROM golang:alpine AS builder

WORKDIR /go/src/github.com/wucenglou/go-tiny
COPY . .

RUN go env -w GO111MODULE=on \
    && go env -w GOPROXY=https://goproxy.cn,direct \
    && go env -w CGO_ENABLED=0 \
    && go env \
    && go mod tidy \
    && go build -o server .

FROM alpine:latest

# 设置时区
ENV TZ=Asia/Shanghai
# RUN apk update && apk add --no-cache tzdata openntpd \
#     && ln -sf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /go/src/github.com/wucenglou/go-tiny

COPY --from=builder /go/src/github.com/wucenglou/go-tiny/server ./server
COPY --from=builder /go/src/github.com/wucenglou/go-tiny/config.yaml ./config.yaml

# 挂载目录：如果使用了sqlite数据库，容器命令示例：docker run -d -v /宿主机路径/gva.db:/go/src/github.com/flipped-aurora/gin-vue-admin/server/gva.db -p 8888:8888 --name gva-server-v1 gva-server:1.0
# VOLUME ["/go/src/github.com/flipped-aurora/gin-vue-admin/server"]

# 创建数据卷来存放配置文件
VOLUME ["/go/src/github.com/wucenglou/go-tiny"]

EXPOSE 8084

CMD ["./server", "-c", "config.yaml"]