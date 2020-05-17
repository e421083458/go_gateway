#!/bin/sh
export GO111MODULE=auto && export GOPROXY=https://goproxy.cn && go mod tidy
GOOS=linux GOARCH=amd64 go build -o ./bin/go_gateway
docker build -f dockerfile_dashboard -t go_gateteway_dashboard .
docker build -f dockerfile_server -t go_gateteway_server .