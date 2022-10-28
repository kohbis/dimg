FROM golang:1.15.2-buster as builder
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GO111MODULE=on
WORKDIR /go/src/github.com/kohbis/dimg/
COPY . .
RUN go build

FROM scratch as runner
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/src/github.com/kohbis/dimg/dimg /usr/local/bin/dimg
ENTRYPOINT ["dimg"]

