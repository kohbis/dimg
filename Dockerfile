FROM golang:1.15.2-alpine as builder
ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64
ENV GO111MODULE on
WORKDIR /go/src/github.com/kohbis/dimg/
COPY . .
RUN go build

FROM alpine:3.11
COPY --from=builder /go/src/github.com/kohbis/dimg/dimg /usr/local/bin/dimg
ENTRYPOINT ["dimg"]
