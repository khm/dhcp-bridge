FROM golang:1.8.3-alpine
MAINTAINER Kurt H Maier <khm@pnnl.gov>

COPY . /go/src/khm/dhcp-bridge
RUN apk add --no-cache --update alpine-sdk; \
    go get github.com/krolaw/dhcp4;

RUN cd /go/src/khm/dhcp-bridge; \
    cd pdhcp; \
    go build; \
    cd ../qdhcp; \
    go build;

FROM alpine:3.4

RUN apk add --update ca-certificates openssl

COPY --from=0 /go/src/khm/dhcp-bridge/qdhcp/qdhcp /usr/bin/qdhcp
COPY --from=0 /go/src/khm/dhcp-bridge/pdhcp/pdhcp /usr/bin/pdhcp

WORKDIR /
