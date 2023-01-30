FROM golang:1.19-alpine AS golang

ENV V2RAY_PLUGIN_VERSION v5.1.0
ENV GO111MODULE on
ENV GOPROXY="https://goproxy.cn,direct"

# Build v2ray-plugin
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories \
    && apk add --no-cache git build-base \
    && mkdir -p /go/src/github.com/teddysun \
    && cd /go/src/github.com/teddysun \
    && git clone https://github.com/teddysun/v2ray-plugin.git \
    && cd v2ray-plugin \
    && git checkout "$V2RAY_PLUGIN_VERSION" \
    && go get -d \
    && go build

FROM alpine:3.17

LABEL maintainer="Acris Liu <acrisliu@gmail.com>"

ENV SHADOWSOCKS_LIBEV_VERSION v3.3.5

# Build shadowsocks-libev
RUN set -ex \
    # Install dependencies
    && apk add --no-cache --virtual .build-deps \
    autoconf \
    automake \
    build-base \
    libev-dev \
    libtool \
    linux-headers \
    udns-dev \
    libsodium-dev \
    mbedtls-dev \
    pcre-dev \
    tar \
    udns-dev \
    c-ares-dev \
    git \
    # Build shadowsocks-libev
    && mkdir -p /tmp/build-shadowsocks-libev \
    && cd /tmp/build-shadowsocks-libev \
    && git clone https://github.com/shadowsocks/shadowsocks-libev.git \
    && cd shadowsocks-libev \
    && git checkout "$SHADOWSOCKS_LIBEV_VERSION" \
    && git submodule update --init --recursive \
    && ./autogen.sh \
    && ./configure --disable-documentation \
    && make install \
    && ssRunDeps="$( \
    scanelf --needed --nobanner /usr/local/bin/ss-server \
    | awk '{ gsub(/,/, "\nso:", $2); print "so:" $2 }' \
    | xargs -r apk info --installed \
    | sort -u \
    )" \
    && apk add --no-cache --virtual .ss-rundeps $ssRunDeps \
    && cd / \
    && rm -rf /tmp/build-shadowsocks-libev \
    # Delete dependencies
    && apk del .build-deps

# Copy v2ray-plugin
COPY --from=golang /go/src/github.com/teddysun/v2ray-plugin/v2ray-plugin /usr/local/bin

COPY ./config.json /tmp/config.json

# Shadowsocks environment variables
ENV CONFIG_FILE = /tmp/config.json

# Run as nobody
USER nobody

# Start shadowsocks-libev client
CMD exec ss-local -c /tmp/config.json -l 8080