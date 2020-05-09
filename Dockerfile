FROM golang:1.13-alpine

ENV LANG "C.UTF-8"
ENV APP_ROOT /go

WORKDIR $APP_ROOT
COPY . $APP_ROOT

RUN apk update \
    && apk add --update git \
    && rm -rf /var/cache/apk/*

# go library
RUN go get -u github.com/spf13/cobra/cobra
