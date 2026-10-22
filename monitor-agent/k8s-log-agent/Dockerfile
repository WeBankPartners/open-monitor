FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -o /agent ./cmd/k8s-log-agent
FROM alpine:3.21
ARG ALPINE_MIRROR=mirrors.aliyun.com
RUN sed -i "s#dl-cdn.alpinelinux.org#${ALPINE_MIRROR}#g" /etc/apk/repositories \
    && apk add --no-cache ca-certificates tzdata
COPY --from=build /agent /usr/local/bin/k8s-log-agent
ENV TZ=Asia/Shanghai
USER 10001:10001
EXPOSE 9105
ENTRYPOINT ["/usr/local/bin/k8s-log-agent"]
