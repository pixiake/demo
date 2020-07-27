FROM golang:1.14.6-alpine3.12 AS ks-upgrade
RUN apk --no-cache add git ca-certificates gcc libc-dev
ADD . /go/src/github.com/kubesphere/ks-upgrade
RUN go get -d github.com/kubesphere/ks-upgrade/...
WORKDIR /go/src/github.com/kubesphere/ks-upgrade
RUN go build -v -o ks-upgrade


FROM alpine:3.12

COPY --from=ks-upgrade /go/src/github.com/kubesphere/ks-upgrade/ks-upgrade /
WORKDIR /
