ARG GO_VERSION

# builder image
FROM golang:$GO_VERSION-alpine as builder

ENV GO111MODULE=on
RUN apk add -U --no-cache curl build-base git
WORKDIR /go/src/github.com/izumin5210/gex
COPY . .

RUN GO111MODULE=on go install ./cmd/gex

# test image
FROM golang:$GO_VERSION-alpine

RUN apk add -U --no-cache curl build-base git
COPY --from=builder /go/bin/gex /go/bin/

WORKDIR /go/src/myapp

ARG GO111MODULE
ENV GO111MODULE $GO111MODULE

COPY ./tests/e2e/testdata/ ./
